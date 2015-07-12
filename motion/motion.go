package motion

import (
	"github.com/bgrundmann/e/buf"
)

// A Motion moves the cursor in a buffer 
type Motion interface {
	// When the motion starts the reader will be initialized to be at the
	// current cursor position.  Return false if the motion is impossible
	// (e.g. a failed search).  
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

// Reverse the given motion.
// Works by the reversing the read direction of the passed
// in reader before passing it to the original motion.  
func reverse(m Motion) Motion {
	return New(func (buf *buf.Buf, rd *buf.Reader) bool {
		rd.Reverse()
		return m.Move(buf, rd)
	})
} 

// RuneForward moves one rune forwards
var RuneForward = New(func (buf *buf.Buf, rd *buf.Reader) bool {
	_, _, err := rd.ReadRune()
	return err == nil
})

// RuneBackward moves one rune backwards
var RuneBackward = reverse(RuneForward)

// Move till the next occurence of the given rune forward.
// Returns false if there is no such character before EOF 
func RuneFindForward(needle rune) Motion {
	return New(func (buf *buf.Buf, rd *buf.Reader) bool {
		for {
			r, _, err := rd.ReadRune()
			if err != nil {
				return false
			} 
			if needle == r {
				return true
			} 
		} 
	})
} 

//// Move several motions one after the other.  
//func Sequence(motions ...Motion) Motion {
//	return New(func (buf *buf.Buf, rd *buf.Reader) bool {
//		for _, m := range motions {
//			if !m.Move(buf, rd) {
//				return false
//			} 
//		} 
//		return true
//	} 
//} 

var LineForward = New(func (buf *buf.Buf, rd *buf.Reader) bool {
	pos, err := buf.PositionFromOffset(rd.Offset())
	if err != nil {
		return false
	} 
	pos.Line++
	off, err := buf.PositionToOffset(pos)
	if err != nil {
		// not a valid position, probably because line has less characters 
		// than the current line.  Let's try again
		pos.Column = 1
		off, err = buf.PositionToOffset(pos)
		if err != nil {
			return false
		}
	}
	_, err = rd.Seek(int64(off), 0)
	return err == nil
})

var LineBackward = New(func (buf *buf.Buf, rd *buf.Reader) bool {
	pos, err := buf.PositionFromOffset(rd.Offset())
	if err != nil {
		return false
	} 
	pos.Line--
	if pos.Line < 1 {
		return false
	}
	off, err := buf.PositionToOffset(pos)
	if err != nil {
		// not a valid position, probably because line has less characters 
		// than the current line.  Let's try again
		pos.Column = 1
		off, err = buf.PositionToOffset(pos)
		if err != nil {
			return false
		}
	}
	_, err = rd.Seek(int64(off), 0)
	return err == nil
})
