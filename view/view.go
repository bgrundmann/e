package view

import (
	"io"

	"github.com/nsf/termbox-go"
	"github.com/bgrundmann/e/buf"
	"github.com/bgrundmann/e/motion"
)

type View struct {
	buffer        *buf.Buf // views may share same buffer
	firstLine     int      // first visible line on screen
	width, height int      // size last time it was displayed
	cursor        buf.Marker
}

func (v *View) Init(b *buf.Buf) {
	v.buffer = b
	v.firstLine = 1
	// We initialize width and height with something
	// sensible here.  Will be updated on first display
	v.width = 80
	v.height = 25
	v.cursor = v.buffer.NewMarker(0)
}

func (v *View) PageDown() {
	lines := v.buffer.Lines()
	v.firstLine += v.height - 2 // like a little overlap
	if v.firstLine > lines-v.height+1 {
		v.firstLine = lines - v.height + 1
	}
}

func (v *View) PageUp() {
	v.firstLine -= v.height - 2 // like a little overlap
	if v.firstLine < 0 {
		v.firstLine = 0
	}
}

// MoveCursor moves the cursor by motion
func (v *View) MoveCursor(m motion.Motion) {
	rd := v.buffer.NewReader(v.cursor.Offset())
	if m.Move(v.buffer, rd) {
		pos, _ := rd.Seek(0, 1)
		v.cursor.Move(int(pos))
	}
}

func (v *View) Display() {
	// This implements simple wrapping
	const coldef = termbox.ColorDefault
	termbox.Clear(coldef, coldef)
	w, h := termbox.Size()
	v.width = w
	v.height = h
	off := v.buffer.Line(v.firstLine)
	r := v.buffer.NewReader(off)
	x := 0
	y := 0
	termbox.HideCursor()
	for {
		rune, n, err := r.ReadRune()
		if v.cursor.Offset() == off {
			termbox.SetCursor(x, y)
		}
		off += n
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

