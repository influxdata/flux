package execute

import "sync/atomic"

func (t *ColListTable) IsDone() bool {
	return t.nrows == 0 || atomic.LoadInt32(&t.used) != 0
}
