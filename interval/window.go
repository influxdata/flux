package interval

import (
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/values"
)

// TODO(nathanielc): Make the epoch a parameter to the window
// See https://github.com/influxdata/flux/issues/2093
const epoch = values.Time(0)

// Window is a description of an infinte set of boundaries in time.
type Window struct {
	// The ith window is expressed via this equation:
	//   window_stop_i = zero + every * i
	//   window_start_i = zero + every * i - period
	every      values.Duration
	period     values.Duration
	zero       values.Time
	zeroMonths int64
}

// NewWindow creates a window which can be used to determine the boundaries for a given point.
// Window boundaries are defined to start at the epoch plus the offset.
// Each subsequent window starts at a multiple of the every duration.
// Each window's length is the start boundary plus the period.
// Every must not be a mix of months and nanoseconds in order to preserve constant time bounds lookup.
func NewWindow(every, period, offset values.Duration) (Window, error) {
	zero := epoch.Add(offset)
	w := Window{
		every: every,
		// treat period as the time before the stop time
		period:     period.Mul(-1),
		zero:       zero,
		zeroMonths: monthsSince(zero),
	}
	if err := w.isValid(); err != nil {
		return Window{}, err
	}
	return w, nil
}

func (w Window) isValid() error {
	if w.every.IsZero() {
		return errors.New(codes.Invalid, "duration used as an interval cannot be zero")
	}
	if w.every.IsMixed() {
		const docURL = "https://v2.docs.influxdata.com/v2.0/reference/flux/stdlib/built-in/transformations/window/#calendar-months-and-years"
		return errors.New(codes.Invalid, "duration used as an interval cannot mix month and nanosecond units").
			WithDocURL(docURL)
	}
	// TODO(nathanielc): what about negative every is that allowed?
	if w.every.IsNegative() {
		return errors.New(codes.Invalid, "duration used as an interval cannot be negative")
	}
	return nil
}

// GetEarliestBounds returns the bounds for the earliest window bounds
// that contains the given time t.  For underlapping windows that
// do not contain time t, the window directly before time t will be returned.
func (w Window) GetEarliestBounds(t values.Time) Bounds {
	index := w.lastIndex(t)
	// We have the last index were we know t will exist.
	// Its possible that previous bounds could contain t depending on the period.
	if !w.period.IsMixed() {
		// Since its not mixed we can adjust the index closer based
		// on how many windows a period can span
		var period, every int64
		if w.every.MonthsOnly() {
			every = w.every.Months()
			period = w.period.Months()
		} else {
			every = w.every.Nanoseconds()
			period = w.period.Nanoseconds()
		}
		if period > every {
			indexDelta := (period / every) - 1
			index -= int(indexDelta)
		}
	}
	// Now do a direct search for the earliest bounds
	var start, stop values.Time
	if w.period.IsPositive() {
		stop = w.zero.Add(w.every.Mul(index + 1))
		start = stop.Add(w.period)
	} else {
		stop = w.zero.Add(w.every.Mul(index))
		start = stop.Add(w.period)
	}
	b := Bounds{
		start: start,
		stop:  stop,
		index: index,
	}
	prev := w.PrevBounds(b)
	for prev.Contains(t) {
		b = prev
		prev = w.PrevBounds(b)
	}
	return b
}

// GetOverlappingBounds returns a slice of bounds that overlaps the input bounds.
func (w Window) GetOverlappingBounds(start, stop values.Time) []Bounds {
	bounds := Bounds{
		start: start,
		stop:  stop,
	}
	if bounds.IsEmpty() {
		return []Bounds{}
	}

	// Estimate the number of windows by using a rough approximation.
	c := (bounds.Length().Duration() / w.every.Duration()) + (w.period.Duration() / w.every.Duration())
	bs := make([]Bounds, 0, c)

	bi := w.GetEarliestBounds(start)
	for bi.start < stop {
		if bi.Overlaps(bounds) {
			bs = append(bs, bi)
		}
		bi = w.NextBounds(bi)
	}
	return bs
}

// NextBounds returns the next boundary in sequence from the given boundary.
func (w Window) NextBounds(b Bounds) Bounds {
	stop := w.zero.Add(w.every.Mul(b.index + 1))
	start := stop.Add(w.period)
	if w.period.IsPositive() {
		start, stop = stop, start
	}
	return Bounds{
		start: start,
		stop:  stop,
		index: b.index + 1,
	}
}

// PrevBounds returns the previous boundary in sequence from the given boundary.
func (w Window) PrevBounds(b Bounds) Bounds {
	stop := w.zero.Add(w.every.Mul(b.index - 1))
	start := stop.Add(w.period)
	if w.period.IsPositive() {
		start, stop = stop, start
	}
	return Bounds{
		start: start,
		stop:  stop,
		index: b.index - 1,
	}
}

// lastIndex will compute the index of the last bounds to contain t
func (w Window) lastIndex(t values.Time) int {
	// We treat both nanoseconds and months as the space of whole numbers (aka integers).
	// This keeps the math the same once we transform into the correct space.
	//    For months we operate in the number of months since the epoch
	//    For nanoseconds we operate in the number of nanoseconds since the epoch
	if w.every.MonthsOnly() {
		return lastIndex(w.zeroMonths, monthsSince(t), w.every.Months())
	}
	return lastIndex(int64(w.zero), int64(t), w.every.Nanoseconds())
}

// lastIndex computes the index where zero + every * index > target
// The zero, target and every values can be in any units so long as they are consistent and zero based.
func lastIndex(zero, target, every int64) int {
	// Given
	//   zero + every * index > target
	// Therefore
	//   index > (target - zero) / every
	// We want the most negative index where the above is true

	// Example: Postive Index
	// zero = 3 target = 14 every = 5
	// Number line with window stops marked:
	//    -2 -1 0 1 2 |3 4 5 6 7 |8 9 10 11 12 |13 14 15 16 17|
	//                0          1             2              3
	// We can see that the index we want is 3
	// (target - zero) /every
	//    = (14 - 3) / 5
	//    = 11 / 5
	//    = 2
	// The we have to adjust by 1 because the delta was positive
	// and we get 3

	// Example: Postive Index on boundary
	// zero = 3 target = 8 every = 5
	// Number line with window stops marked:
	//   -3 |-2 -1 0 1 2 |3 4 5 6 7 |8 9 10 11 12 |13 14 15
	//     -1            0          1             2
	// We can see that the index we want is 2, remember stop is exclusive
	// (target - zero) /every
	//    = (8 - 3) / 5
	//    = 5 / 5
	//    = 1
	// The we have to adjust by 1 because the delta was positive
	// and we get 2

	// Example: Negative Index
	// zero = 3 target = -9 every = 5
	// Number line with window stops marked:
	//    |-12 -11 -10 -9 -8 |-7 -6 -5 -4 -3 |-2 -1 0 1 2 |3 4 5 6 7
	//   -3                 -2              -1            0
	// We can see that the index we want is -2
	// (target - zero) /every
	//    = (-9 - 3) / 5
	//    = -12 / 5
	//    = -2
	// We do not have to adjust by because the delta was negative

	// Example: Negative Index on boundary
	// zero = 3 target = -7 every = 5
	// Number line with window stops marked:
	//    |-12 -11 -10 -9 -8 |-7 -6 -5 -4 -3 |-2 -1 0 1 2 |3 4 5 6 7
	//   -3                 -2              -1            0
	// We can see that the index we want is -1
	// (target - zero) /every
	//    = (-7 - 3) / 5
	//    = -10 / 5
	//    = -2
	// We have to adjust by 1 because the delta was negative and we landed on the boundary
	// and we get -1

	delta := target - zero
	index := delta / every

	if delta > 0 || (delta < 0 && delta%every == 0) {
		index += 1
	}
	return int(index)
}

// monthsSince converts a time into the number of months since the unix epoch
func monthsSince(t values.Time) int64 {
	ts := t.Time()
	year, month, _ := ts.Date()
	return int64(year)*12 + int64(month) - 1
}

//TODO
// Move into values package
// Add tests for NextBounds and PrevBounds
// Add tests very far away from the epoch
// GetLatestBounds instead as a more efficient API
