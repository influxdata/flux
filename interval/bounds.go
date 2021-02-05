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
	// index keeps track of how many windows have been added or subtracted as additional
	// windows are added to or subtracted from the initial bounds. In essence, it tracks the
	// offset from the original bounds in order to keep operations more straightforward.
	// See the Window struct and the window tests for additional info.
	index int
}

// NewBounds create a new Bounds given start and stop values
func NewBounds(start, stop values.Time) Bounds {
	return Bounds{
		start: start,
		stop:  stop,
	}
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

// IsZero returns true if the start and stop values are both zero.
func (b Bounds) IsZero() bool {
	return b.start == 0 && b.stop == 0
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

// Intersect returns the intersection of two bounds.
// It returns empty bounds if one of the input bounds are empty.
// TODO: there are several places that implement bounds and related utilities.
//  consider a central place for them?
func (b Bounds) Intersect(o Bounds) Bounds {
	if b.IsEmpty() || o.IsEmpty() || !b.Overlaps(o) {
		return Bounds{
			start: b.start,
			stop:  b.stop,
		}
	}
	i := Bounds{}

	i.start = b.start
	if o.start > b.start {
		i.start = o.start
	}

	i.stop = b.stop
	if o.stop < b.stop {
		i.stop = o.stop
	}

	return i
}

// Union returns the smallest bounds which contain both input bounds.
// It returns empty bounds if one of the input bounds are empty.
func (b Bounds) Union(o Bounds) Bounds {
	if b.IsEmpty() || o.IsEmpty() {
		return Bounds{
			start: values.Time(0),
			stop:  values.Time(0),
		}
	}
	u := new(Bounds)

	u.start = b.start
	if o.start < b.start {
		u.start = o.start
	}

	u.stop = b.stop
	if o.stop > b.stop {
		u.stop = o.stop
	}

	return *u
}
