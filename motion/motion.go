package motion

import (
	"github.com/bgrundmann/e/buf"
)

// A Motion moves the cursor in a buffer 
type Motion interface {
	// When the motion starts the reader will be initialized to be at the
	// current cursor position.  Return false if the motion is impossible
	// (e.g. a failed search)
	Move(buf *buf.Buf, reader *buf.Reader) bool
} 

type motion func(*buf.Buf, *buf.Reader) bool

func (f motion) Move(buf *buf.Buf, reader *buf.Reader) bool {
	return f(buf, reader)
} 

// New creates a new motion from a function.
func New(move func(*buf.Buf, *buf.Reader) bool) Motion {
	return motion(move)
}

// RuneBackward moves one rune backwards
var RuneBackward = New(func (buf *buf.Buf, rd *buf.Reader) bool {
	rd.Reverse()
	_, _, err := rd.ReadRune()
	return err == nil
})

// RuneForward moves one rune forwards
var RuneForward = New(func (buf *buf.Buf, rd *buf.Reader) bool {
	_, _, err := rd.ReadRune()
	return err == nil
})
