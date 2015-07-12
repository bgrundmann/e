package buf

// A Marker represents a position in a buffer relative to its surrounding text.
// A marker changes its offset from the beginning of the buffer automatically
// whenever text is inserted or deleted, so that it stays with the two characters on
// either side of it.
type Marker interface {
	Offset() int
	// Move the Marker to the given offset.  Panics if the given offset is invalid.
	Move(int) 
} 

type marker struct {
	buf *Buf
	off int
	id int
} 

// Return a new marker at off.  
func (buf *Buf) NewMarker(off int) Marker {
	m := &marker {
		buf: buf,
		off: off,
	} 
	m.id = buf.AddObserver(m)
	return m
} 

func (m *marker) Offset() int {
	return m.off
}

func (m *marker) Move(off int) {
	// FIXME: panic if offset is invalid.  Or maybe something else
	m.off = off
}

func (m *marker) OnBufInsert(off int, bytes []byte) {
	if off <= m.off {
		m.off += len(bytes)
	} 
} 

func (m *marker) OnBufDelete(off1, off2 int) {
	// TODO: think about what should happen if
	// m.off between off1 and off2
	if off2 <= m.off {
		m.off -= off2 - off1
	} 
} 


