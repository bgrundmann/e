// This package implements a text editors buffer using the piece table method
// ala Oberon.
package buf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type piece struct {
	off1 int
	off2 int
	prev *piece
	next *piece
}

func (p *piece) len() int {
	return p.off2 - p.off1
}

func (p *piece) link(p2 *piece) {
	p.next = p2
	p2.prev = p
}

// split piece into two pieces such that the first piece is n characters long
func (p *piece) split(n int) (*piece, *piece) {
	off2 := p.off1 + n
	return &piece{off1: p.off1, off2: off2}, &piece{off1: off2, off2: p.off2}
}

// BufferObserver is the interface that get's notified when a Buffer changes
// Both functions are called before the change has happened
type BufferObserver interface {
	OnBufDelete(off1, off2 int)
	OnBufInsert(off int, bytes []byte)
}

// A text editors buffer.
// It implements Writer.  Any writes done that way are appended at the end of the buffer.
type Buf struct {
	bytes              bytes.Buffer
	sentinel           piece
	len                int
	nextFreeObserverId int
	observers          map[int]BufferObserver
	lineCache          OneLineCache // position of most recently asked for line
	lines              int // number of lines in buffer or 0 if unknown
}

type OneLineCache struct {
	line int  // the line starting at 1 (if zero the cache is invalid)
	off int   // offset of the line
} 

// Init initializes a buffer and returns it.
func (b *Buf) Init() *Buf {
	b.sentinel.next = &b.sentinel
	b.sentinel.prev = &b.sentinel
	b.observers = make(map[int]BufferObserver)
	return b
}

// Len returns the length of the buffer in bytes.
func (b *Buf) Len() int {
	return b.len
}

// Delete the bytes between off1 (inclusive) and off2 (exclusive) in a Buf.
func (b *Buf) Delete(off1, off2 int) {
	if off1 > off2 || off1 < 0 || off2 > b.len {
		panic(fmt.Sprintf("Delete: Invalid offsets given %v-%v valid:0-%v", off1, off2, b.len))
	}
	if off1 == off2 {
		// deleting the empty string => noop
		return
	}
	b.lineCache.line = 0
	b.lines = 0
	for _, ob := range b.observers {
		ob.OnBufDelete(off1, off2)
	}

	o1, p1 := b.findPiece(off1)
	o2, p2 := b.findPiece(off2)

	var left *piece
	if off1 == o1 {
		// we are deleting all of p1
		left = p1.prev
	} else {
		len1 := off1 - o1
		prev := p1.prev
		left, _ = p1.split(len1)
		prev.link(left)
	}

	var right *piece
	if off2 == o2 {
		// we at the beginning of p2 and therefore won't delete
		// anything of it
		right = p2
	} else {
		len2 := off2 - o2
		next := p2.next
		_, right = p2.split(len2)
		right.link(next)
	}
	left.link(right)
	b.len -= off2 - off1
}

// Insert the bytes starting at off into a buf.
func (b *Buf) Insert(off int, s []byte) {
	if off < 0 || off > b.len {
		panic(fmt.Sprintf("Insert: invalid offset %v valid:0-%v", off, b.len))
	}
	if len(s) == 0 {
		// inserting the empty string => noop
		return
	}
	b.lineCache.line = 0
	b.lines = 0
	for _, ob := range b.observers {
		ob.OnBufInsert(off, s)
	}

	off1 := b.bytes.Len()
	n, err := b.bytes.Write(s)
	if err != nil {
		panic("bytes.Write returned an error but doc says it never does so")
	}
	np := &piece{
		off1: off1,
		off2: off1 + n,
	}
	o, p := b.findPiece(off)
	left := p.prev
	if off == o {
		// insert at beginning of piece
		np.link(p)
		left.link(np)
	} else {
		// split piece and insert in middle
		len1 := off - o
		p1, p2 := p.split(len1)
		p1.link(np)
		np.link(p2)
		left.link(p1)
	}
	b.len += n
}

func (b *Buf) eachpiece(f func(p *piece)) {
	for p := b.sentinel.next; p != &b.sentinel; p = p.next {
		f(p)
	}
}

// findPiece finds the piece with off1 >= off
func (b *Buf) findPiece(off int) (pieceStart int, piece *piece) {
	pieceStart = 0
	for piece = b.sentinel.next; piece != &b.sentinel; piece = piece.next {
		if pieceStart <= off && off < pieceStart+piece.len() {
			return
		}
		pieceStart += piece.len()
	}
	return
}

func (b *Buf) sliceOfPiece(p *piece) []byte {
	return b.bytes.Bytes()[p.off1:p.off2]
}

func (b *Buf) String() string {
	s := make([]string, 8)
	b.eachpiece(func(p *piece) {
		s = append(s, string(b.sliceOfPiece(p)))
	})
	return strings.Join(s, "")
}

func (b *Buf) Write(p []byte) (n int, err error) {
	b.Insert(b.len, p)
	return len(p), nil
}

// A position in a file given by line and column (both starting at 1)
// Note that this is a position in the file.  In particular columns
// are counted in number of runes in the line NOT number of characters
// displayed on the screen (e.g. '\t' counts as 1 not 8, ...).
type Position struct {
	Line   int
	Column int
}

// Translate a offset into a position.  Errors if offset is not a valid
// position (that is either > length of the file or in the middle of a
// multibyte utf8 sequence).
func (b *Buf) PositionFromOffset(off int) (Position, error) {
	// TODO: This can obviously made more efficient by caching, etc...
	pos := Position{
		Line:   1,
		Column: 1,
	}
	rd := b.NewReader(0)
	for rd.Offset() != off {
		r, _, err := rd.ReadRune()
		if err != nil {
			return Position{}, err
		}
		if r == '\n' {
			pos.Line++
			pos.Column = 1
		} else {
			pos.Column++
		}
	}
	return pos, nil
}

// Translate a position into an offset. Errors if the given position
// is not a valid position.
func (b *Buf) PositionToOffset(p Position) (int, error) {
	off := b.Line(p.Line)
	rd := b.NewReader(off)
	// we are in the right line
	for runesToSkip := p.Column - 1; runesToSkip > 0; runesToSkip-- {
		r, _, err := rd.ReadRune()
		if err != nil {
			return 0, err
		}
		if r == '\n' {
			return 0, fmt.Errorf("Invalid position line %i contains less than %i columns", p.Line, p.Column)
		}
	}
	return rd.Offset(), nil
}

// Line returns the offset of the first character of Line n.  
// Note Line numbers start at 1.
// FIXME: Either add error code, or make it panic if line number > number
func (b *Buf) Line(n int) int {
	var startOfLine, linesToSkip int
	if b.lineCache.line != 0 && b.lineCache.line < n {
		startOfLine = b.lineCache.off
		linesToSkip = n - b.lineCache.line
	} else if (b.lineCache.line == n) {
		return b.lineCache.off
	} else {
		startOfLine = 0
		linesToSkip = n - 1
	} 
	rd := b.NewReader(startOfLine)
	for ; linesToSkip > 0; linesToSkip-- {
		for {
			rn, _, err := rd.ReadRune()
			if err != nil {
				return startOfLine
			}
			if rn == '\n' {
				startOfLine = rd.Offset()
				break
			}
		}
	}
	// we always update the cache if it is invalid or
	// if we asked for a line above the current line and we can't
	// easily reach that line from the beginning or
	// if it is more than a few lines past the the current line 
	if (b.lineCache.line == 0) || 
		(n < b.lineCache.line && n > 5) ||
		(n - b.lineCache.line > 5) {
		b.lineCache.line = n
		b.lineCache.off = startOfLine
	} 
	return startOfLine
}

// Lines returns the number of lines in the buffer
// The empty buffer has exactly one (empty) line.
func (b *Buf) Lines() int {
	if b.lines != 0 {
		return b.lines
	} else {
		r := b.NewReader(0)
		lines := 1
		for {
			rn, _, err := r.ReadRune()
			if err != nil {
				break
			}
			if rn == '\n' {
				lines++
			}
		}
		b.lines = lines
		return lines
	} 
}

// The type of a Reader on the buffer.
// Implements io.ReadSeeker and RuneScanner.
// It also implements reading in reverse direction.  At the moment only
// for runes.
type Reader struct {
	buf          *Buf
	piece        *piece
	offInPiece   int  // offset in the current piece
	off          int  // absolute offset in file
	reverse      bool // read in reverse direction
	lastRuneSize int  // -1 if last read was not a ReadRune
}

// NewReader creates a new reader starting at off.
func (b *Buf) NewReader(off int) *Reader {
	o, p := b.findPiece(off)
	return &Reader{
		buf:          b,
		piece:        p,
		offInPiece:   off - o,
		off:          off,
		reverse:      false,
		lastRuneSize: -1,
	}
}

// Reverse reverses direction of reading.
func (rd *Reader) Reverse() {
	rd.reverse = !rd.reverse
}

func (r *Reader) Read(dst []byte) (int, error) {
	if r.reverse {
		panic("Reader.Read in reverse direction not implemented")
	}
	offDst := 0
process_piece:
	if r.piece == &r.buf.sentinel { // no more bytes
		// return however much we copied
		return offDst, io.EOF
	}
	bytes := r.buf.sliceOfPiece(r.piece)
	n := copy(dst[offDst:], bytes[r.offInPiece:])
	offDst += n
	r.off += n
	if offDst == len(dst) { // no more space in buffer
		r.offInPiece += n
		r.lastRuneSize = -1 // invalidate calls to UnreadRune
		return offDst, nil
	} else { // we are done with the current piece
		// but there is still space in the buffer
		r.piece = r.piece.next
		r.offInPiece = 0
		goto process_piece
	}
}

func (rd *Reader) readRuneForward() (r rune, size int, err error) {
	bytes := rd.buf.sliceOfPiece(rd.piece)[rd.offInPiece:]
	// specialisation of the common case
	if len(bytes) > 0 && bytes[0] < 0x80 { // one byte utf-8 sequence
		r, size = rune(bytes[0]), 1
		rd.off += size
		rd.offInPiece += size
	} else if utf8.FullRune(bytes) { // multi byte but complete in current piece
		r, size = utf8.DecodeRune(bytes)
		rd.off += size
		rd.offInPiece += size
	} else { // need to read several bytes
		var buf [utf8.UTFMax]byte
		n, err := rd.Read(buf[:])
		if n == 0 {
			return 0, 0, io.EOF
		} else if err != nil && err != io.EOF {
			return 0, 0, err
		}
		r, size = utf8.DecodeRune(buf[:n])
	}
	return r, size, nil
}

func (rd *Reader) readRuneBackward() (r rune, size int, err error) {
	var bytes [4]byte
	size = 0
read_next_byte:
	if rd.off == 0 {
		if size == 0 {
			return 0, 0, io.EOF
		}
		// this means we wanted to read another byte
		// because we don't have a valid utf character
		// yet but there are not anymore...
		// TODO: handle that
		panic("partial utf8 at end of buffer not yet implemented")
	}
	if rd.offInPiece <= 0 {
		rd.piece = rd.piece.prev
		rd.offInPiece = rd.piece.off2
	}
	bytes[size] = rd.buf.sliceOfPiece(rd.piece)[rd.offInPiece-1]
	size++
	rd.offInPiece--
	rd.off--
	if rd.offInPiece <= 0 {
		rd.piece = rd.piece.prev
		rd.offInPiece = rd.piece.off2
	}
	if utf8.FullRune(bytes[:size]) {
		r, size = utf8.DecodeRune(bytes[:size])
		return r, size, nil
	}
	// not a full rune read another byte into the
	// buffer and try again
	goto read_next_byte
}

func (rd *Reader) ReadRune() (r rune, size int, err error) {
	if rd.reverse {
		r, size, err = rd.readRuneBackward()
	} else {
		r, size, err = rd.readRuneForward()
	}
	if err == nil {
		rd.lastRuneSize = size
	}
	return r, size, err
}

func (rd *Reader) UnreadRune() error {
	// TODO bgrundmann: This can be optimized for the common case
	if rd.lastRuneSize < 0 {
		return errors.New("Cannot call UnreadRune when previous operation wasn't ReadRune")
	}
	var offset int64
	if rd.reverse {
		offset = int64(rd.off + rd.lastRuneSize)
	} else {
		offset = int64(rd.off - rd.lastRuneSize)
	}
	_, err := rd.Seek(offset, 0)
	return err
}

// Return the current offset of the reader in the file.
// Equivalent to Seek(0, 1) but more readable
func (r *Reader) Offset() int {
	return r.off
}

func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	// TODO: Many special cases could written out.  For example
	// if position is in current piece.  Figure out if that is
	// worth it.
	var absoluteOff int
	switch whence {
	case 0: // relative to origin
		absoluteOff = int(offset)
	case 1: // relative to current offset
		absoluteOff = r.off + int(offset)
	case 2: // relative to end
		absoluteOff = r.buf.Len() + int(offset)
	default:
		panic("Invalid argument passed as whence to Seek")
	}
	if absoluteOff < 0 {
		return 0, errors.New("Invalid offset given to Seek")
	}
	o, p := r.buf.findPiece(absoluteOff)
	r.off = absoluteOff
	r.offInPiece = absoluteOff - o
	r.piece = p
	r.lastRuneSize = -1
	return int64(absoluteOff), nil
}

func (b *Buf) AddObserver(buf BufferObserver) int {
	n := b.nextFreeObserverId
	b.nextFreeObserverId++
	b.observers[n] = buf
	return n
}

func (b *Buf) RemoveObserver(id int) {
	delete(b.observers, id)
}
