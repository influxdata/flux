package execute

import (
	"fmt"
	"math"
	"time"

	"github.com/influxdata/flux/values"
)

type Time = values.Time
type Duration = values.Duration

const (
	MaxTime = math.MaxInt64
	MinTime = math.MinInt64
)

type Bounds struct {
	Start Time
	Stop  Time
}

var AllTime = Bounds{
	Start: MinTime,
	Stop:  MaxTime,
}

func (b Bounds) IsEmpty() bool {
	return b.Start >= b.Stop
}

func (b Bounds) String() string {
	return fmt.Sprintf("[%v, %v)", b.Start, b.Stop)
}

func (b Bounds) Contains(t Time) bool {
	return t >= b.Start && t < b.Stop
}

func (b Bounds) Overlaps(o Bounds) bool {
	return b.Contains(o.Start) || (b.Contains(o.Stop) && o.Stop > b.Start) || o.Contains(b.Start)
}

// Intersect returns the intersection of two bounds.
// It returns empty bounds if one of the input bounds are empty.
// TODO: there are several places that implement bounds and related utilities.
//  consider a central place for them?
func (b *Bounds) Intersect(o Bounds) Bounds {
	if b.IsEmpty() || o.IsEmpty() || !b.Overlaps(o) {
		return Bounds{
			Start: b.Start,
			Stop:  b.Start,
		}
	}
	i := Bounds{}

	i.Start = b.Start
	if o.Start > b.Start {
		i.Start = o.Start
	}

	i.Stop = b.Stop
	if o.Stop < b.Stop {
		i.Stop = o.Stop
	}

	return i
}

func (b Bounds) Equal(o Bounds) bool {
	return b == o
}

func (b Bounds) Shift(d Duration) Bounds {
	return Bounds{Start: b.Start.Add(d), Stop: b.Stop.Add(d)}
}

func (b Bounds) Duration() Duration {
	if b.IsEmpty() {
		return values.ConvertDurationNsecs(0)
	}
	return b.Stop.Sub(b.Start)
}

func Now() Time {
	return values.ConvertTime(time.Now())
}
