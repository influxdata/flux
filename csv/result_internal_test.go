package csv

import "sync/atomic"

func (d *tableDecoder) IsDone() bool {
	return d.empty || atomic.LoadInt32(&d.used) != 0
}
