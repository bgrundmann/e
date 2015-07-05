package main

import "github.com/nsf/termbox-go"
import "io"
import "bufio"
import "os"

type View struct {
	buffer *Buf
	off    int
}

func (v *View) Init(b *Buf) {
	v.buffer = b
	v.off = 0
}

func (v *View) Display() {
	const coldef = termbox.ColorDefault
	termbox.Clear(coldef, coldef)
	w, h := termbox.Size()
	r := bufio.NewReader(v.buffer.NewReader(v.off))
	x := 0
	y := 0
	for {
		rune, _, err := r.ReadRune()
		if x >= w {
			x = 0
			y++
		}
		if y >= h || err == io.EOF {
			break
		}
		switch rune {
		case '\n':
			y++
			x = 0
		case '\t':
			for {
				termbox.SetCell(x, y, ' ', coldef, coldef)
				x++
				if x%4 == 0 || x >= w {
					break
				}
			}
		default:
			termbox.SetCell(x, y, rune, coldef, coldef)
			x++
		}
	}
	termbox.Flush()
}

// AppendFile appends the contents of file to buf.
func AppendFile(buf *Buf, filename string) error {
	f, err := os.Open("e.go")
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(buf, f)
	return err
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)
	var b Buf
	b.Init()
	var v View
	v.Init(&b)
	AppendFile(&b, "e.go")

mainloop:
	for {
		v.Display()
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break mainloop

			default:
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}
