package main

import "bytes"
import "fmt"
import "strings"

type Piece struct {
	off1 int
	off2 int
	prev *Piece
	next *Piece
}

func (p *Piece) len() int {
	return p.off2 - p.off1
}

func (p *Piece) link(p2 *Piece) {
	p.next = p2
	p2.prev = p
}

// split piece into two pieces such that the first piece is n characters long
func (p *Piece) split(n int) (*Piece, *Piece) {
	off2 := p.off1 + n
	return &Piece{off1: p.off1, off2: off2}, &Piece{off1: off2, off2: p.off2}
}

type Buf struct {
	bytes    bytes.Buffer
	sentinel Piece
	len      int
}

// Init initializes a buffer and returns it.
func (b *Buf) Init() *Buf {
	b.sentinel.next = &b.sentinel
	b.sentinel.prev = &b.sentinel
	return b
}

func (b *Buf) eachPiece(f func(p *Piece)) {
	for p := b.sentinel.next; p != &b.sentinel; p = p.next {
		f(p)
	}
}

// FindPiece finds the piece with off1 >= off
func (b *Buf) findPiece(off int) (pieceStart int, piece *Piece) {
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
	b.eachPiece(func(p *Piece) {
		s = append(s, string(b.bytes.Bytes()[p.off1:p.off2]))
	})
	return strings.Join(s, "")
}

func (b *Buf) Insert(off int, s []byte) {
	if off < 0 || off > b.len {
		panic(fmt.Sprintf("Insert given invalid offset %v valid:0-%v", off, b.len))
	}
	off1 := b.bytes.Len()
	n, err := b.bytes.Write(s)
	if err != nil {
		panic("bytes.Write returned an error but doc says it never does so")
	}
	np := &Piece{
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
	b.eachPiece(func(p *Piece) {
		fmt.Printf("%+v\n", p)
	})
	fmt.Printf("%s\n", &b)
	b.Insert(5, []byte(" "))
	fmt.Printf("%s\n", &b)
	b.eachPiece(func(p *Piece) {
		fmt.Printf("%+v\n", p)
	})
	fmt.Fprintf(&b, "\nHaha!")
	fmt.Printf("%s\n", &b)
}
