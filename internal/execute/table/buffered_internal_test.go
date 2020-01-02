package table

import "sync/atomic"

// IsDone is used to allow the tests to access internal parts
// of the table structure for the table tests.
// This method can only be used by asserting that it exists
// through an anonymous interface. This should not be used
// outside of testing code because there is no guarantee
// on the safety of this method.
func (b *BufferedTable) IsDone() bool {
	return len(b.Buffers) == 0 || atomic.LoadInt32(&b.used) != 0
}
