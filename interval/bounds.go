package interval

import (
	"fmt"
	"math"

	"github.com/influxdata/flux/values"
)

const (
	MaxTime = math.MaxInt64
	MinTime = math.MinInt64
)

type Bounds struct {
	start values.Time
	stop  values.Time
<<<<<<< HEAD
	// As additional windows are added to or subtracted from the initial bounds, index keeps track of how many
	// windows have been added or subtracted. In essence, it tracks the offset from the original bounds in order
	// to keep operations more straightforward. See the Window struct and the window tests for additional info.
=======
>>>>>>> 7087e407 (feat(interval): adds new interval package for consistent window handling)
	index int
}

func (b Bounds) Start() values.Time {
	return b.start
}

func (b Bounds) Stop() values.Time {
	return b.stop
}

func (b Bounds) IsEmpty() bool {
	return b.start >= b.stop
}

func (b Bounds) String() string {
	return fmt.Sprintf("[%v, %v)", b.start, b.stop)
}

func (b Bounds) Contains(t values.Time) bool {
	return t >= b.start && t < b.stop
}

func (b Bounds) Overlaps(o Bounds) bool {
	return b.Contains(o.start) || (b.Contains(o.stop) && o.stop > b.start) || o.Contains(b.start)
}

func (b Bounds) Equal(o Bounds) bool {
	return b == o
}

func (b Bounds) Length() values.Duration {
	if b.IsEmpty() {
		return values.ConvertDurationNsecs(0)
	}
	return b.stop.Sub(b.start)
}
