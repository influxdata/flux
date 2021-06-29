package execute

type ring struct {
	buf []interface{}
	i   int
	sz  int
}

func newRing(sz int) *ring {
	return &ring{
		buf: make([]interface{}, sz),
	}
}

func (r *ring) Next() interface{} {
	if r.sz == 0 {
		return nil
	}

	v := r.buf[r.i]
	r.buf[r.i] = nil
	r.i++
	if r.i >= len(r.buf) {
		r.i = 0
	}
	r.sz--
	return v
}

func (r *ring) Len() int {
	return r.sz
}

func (r *ring) Append(v interface{}) {
	if r.sz == len(r.buf) {
		// The number of elements in the ring is equal
		// to the buffer size so we need to dynamically
		// resize it.
		//
		// Do this by copying the existing buffer from
		// the current index, which may be in the middle,
		// and copy that section of the array first.
		// Then copy the previous buffer from the zero index
		// to the current index which is the end of the ring.
		buf := make([]interface{}, 0, r.sz*2)
		buf = append(buf, r.buf[r.i:]...)
		buf = append(buf, r.buf[:r.i]...)
		// Resize the length of the buffer to its capacity
		// now that we have finished the appends.
		r.i, r.buf = 0, buf[:cap(buf)]
	}

	i := (r.i + r.sz) % len(r.buf)
	r.buf[i] = v
	r.sz++
}
