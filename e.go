package main

import "bytes"
import "fmt"

type Piece struct {
	off1 int
	off2 int
	prev *Piece
	next *Piece
} 

type Buf struct {
	buf bytes.Buffer
	pieces *Piece
}

func (b *Buf) eachPiece(f func(p *Piece)) {
	for p := b.pieces; p != nil; p = p.next {
		f(p)
	} 
} 

func (b *Buf) appendString(s string) {
	off1 := b.buf.Len()
	n, _ := b.buf.WriteString(s)
	if b.pieces == nil {
		b.pieces = &Piece{
			off1: off1,
			off2: off1+n,
		} 
	} else {
		panic("Not yet implemented")
	} 
} 

func main() {
	var b Buf
	b.eachPiece(func(p *Piece) {
		fmt.Println("%v", p)
	})
	b.appendString("Hello World")
	b.eachPiece(func(p *Piece) {
		fmt.Println(p)
	})
}
