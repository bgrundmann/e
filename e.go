package main

import "bytes"
import "fmt"
import "strings"

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

// A text editors buffer.
// It implements Writer.  Any writes done that way are appended at the end of the buffer.
type Buf struct {
	bytes    bytes.Buffer
	sentinel piece
	len      int
}

// Init initializes a buffer and returns it.
func (b *Buf) Init() *Buf {
	b.sentinel.next = &b.sentinel
	b.sentinel.prev = &b.sentinel
	return b
}

// Len returns the length of the buffer in bytes.
func (b *Buf) Len() int {
	return b.len
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

func (b *Buf) String() string {
	s := make([]string, 8)
	b.eachpiece(func(p *piece) {
		s = append(s, string(b.bytes.Bytes()[p.off1:p.off2]))
	})
	return strings.Join(s, "")
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

func (b *Buf) Write(p []byte) (n int, err error) {
	b.Insert(b.len, p)
	return len(p), nil
}


func main() {
	var b Buf
	b.Init()
	b.Insert(0, []byte("World"))
	b.Insert(0, []byte("Hello"))
	b.eachpiece(func(p *piece) {
		fmt.Printf("%+v\n", p)
	})
	fmt.Printf("%s\n", &b)
	b.Insert(5, []byte(" "))
	fmt.Printf("%s\n", &b)
	b.eachpiece(func(p *piece) {
		fmt.Printf("%+v\n", p)
	})
	fmt.Fprintf(&b, "\nHaha!")
	fmt.Printf("%s\n", &b)
}
