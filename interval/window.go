package interval

import (
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/values"
)

const epoch = values.Time(0)

var epochYear, epochMonth int64

func init() {
	ts := epoch.Time()
	y, m, _ := ts.Date()
	epochYear = int64(y)
	epochMonth = int64(m - 1)
}

// TODO(nathanielc): Make the epoch a parameter to the window
// See https://github.com/influxdata/flux/issues/2093
//
// Window is a description of an infinite set of boundaries in time.
//
// Note the properties of this struct should remain private.
// Furthermore they should not be exposed via public getter methods.
// There should never be any need to access a window's properties in order to
// perform window calculations. The public interface should be sufficient.
type Window struct {
	// The ith window start is expressed via this equation:
	//   window_start_i = zero + every * i
	//   window_stop_i = zero + every * i + period
	every      values.Duration
	period     values.Duration
	zero       values.Time
	zeroMonths int64
}

// NewWindow creates a window which can be used to determine the boundaries for a given point.
// Window boundaries start at the epoch plus the offset.
// Each subsequent window starts at a multiple of the every duration.
// Each window's length is the start boundary plus the period.
// Every must not be a mix of months and nanoseconds in order to preserve constant time bounds lookup.
func NewWindow(every, period, offset values.Duration) (Window, error) {
	zero := epoch.Add(offset)
	w := Window{
		every:      every,
		period:     period,
		zero:       zero,
		zeroMonths: monthsSince(zero),
	}
	if err := w.isValid(); err != nil {
		return Window{}, err
	}
	return w, nil
}

// IsZero checks if the window's every duration is zero
func (w Window) IsZero() bool {
	return w.every.IsZero()
}

func (w Window) Every() values.Duration {
	return w.every
}

func (w Window) Period() values.Duration {
	return w.period
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
	if w.every.IsNegative() {
		return errors.New(codes.Invalid, "duration used as an interval cannot be negative")
	}
	return nil
}

// GetLatestBounds returns the bounds for the latest window bounds that contains the given time t.
// For underlapping windows that do not contain time t, the window directly before time t will be returned.
func (w Window) GetLatestBounds(t values.Time) Bounds {
	// Get the latest index that should contain the time t
	index := w.lastIndex(t)
	// Construct the bounds from the index
	start := w.zero.Add(w.every.Mul(index))
	b := Bounds{
		start: start,
		stop:  start.Add(w.period),
		index: index,
	}
	// If the period is negative its possible future bounds can still contain this point
	if w.period.IsNegative() {
		// swap start and stop since the period was negative
		b.start, b.stop = b.stop, b.start
		// If period is NOT mixed we can do a direct calculation
		// to determine how far into the future a bounds may be found.
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
				indexDelta := period / every
				index += int(indexDelta)
			}
		}
		// Now do a direct search
		next := w.NextBounds(b)
		for next.Contains(t) {
			b = next
			next = w.NextBounds(next)
		}
	}
	return b
}

// GetOverlappingBounds returns a slice of bounds that overlaps the input bounds.
// The returned set of bounds are ordered by decreasing time.
func (w Window) GetOverlappingBounds(start, stop values.Time) []Bounds {
	bounds := Bounds{
		start: start,
		stop:  stop,
	}
	if bounds.IsEmpty() {
		return []Bounds{}
	}

	// Estimate the number of windows by using a rough approximation.
	count := (bounds.Length().Duration() / w.every.Duration()) + (w.period.Duration() / w.every.Duration())
	bs := make([]Bounds, 0, count)

	curr := w.GetLatestBounds(stop)
	for curr.stop > start {
		if curr.Overlaps(bounds) {
			bs = append(bs, curr)
		}
		curr = w.PrevBounds(curr)
	}

	return bs
}

// NextBounds returns the next boundary in sequence from the given boundary.
func (w Window) NextBounds(b Bounds) Bounds {
	index := b.index + 1
	start := w.zero.Add(w.every.Mul(index))
	stop := start.Add(w.period)
	if w.period.IsNegative() {
		start, stop = stop, start
	}
	return Bounds{
		start: start,
		stop:  stop,
		index: index,
	}
}

// PrevBounds returns the previous boundary in sequence from the given boundary.
func (w Window) PrevBounds(b Bounds) Bounds {
	index := b.index - 1
	start := w.zero.Add(w.every.Mul(index))
	stop := start.Add(w.period)
	if w.period.IsNegative() {
		start, stop = stop, start
	}
	return Bounds{
		start: start,
		stop:  stop,
		index: index,
	}
}

// lastIndex will compute the index of the last bounds to contain t
func (w Window) lastIndex(t values.Time) int {
	// We treat both nanoseconds and months as the space of whole numbers (aka integers).
	// This keeps the math the same once we transform into the correct space.
	//    For months, we operate in the number of months since the epoch.
	//    For nanoseconds, we operate in the number of nanoseconds since the epoch.
	if w.every.MonthsOnly() {
		target := monthsSince(t)
		// Check if the target day and time of the month is before the zero day and time of the month.
		// If it is, that means that in _months_ space we are really in the previous month.
		if isBeforeWithinMonth(t, w.zero) {
			target -= 1
		}
		return lastIndex(w.zeroMonths, target, w.every.Months())
	}
	return lastIndex(int64(w.zero), int64(t), w.every.Nanoseconds())
}

// lastIndex computes the index where zero + every * index ≤ target
// The zero, target and every values can be in any units so long as they are consistent and zero based.
func lastIndex(zero, target, every int64) int {
	// Given
	//   zero + every * index ≤ target
	// Therefore
	//   index ≤ (target - zero) / every
	// We want to find the most positive index where the above is true

	// Example: Positive Index
	// zero = 3 target = 14 every = 5
	// Number line with window starts marked:
	//    -2 -1 0 1 2 |3 4 5 6 7 |8 9 10 11 12 |13 14 15 16 17
	//                0          1             2
	// We can see that the index we want is 2
	// (target - zero) /every
	//    = (14 - 3) / 5
	//    = 11 / 5
	//    = 2
	// We do not adjust because the delta was positive

	// Example: Positive Index on boundary
	// zero = 3 target = 13 every = 5
	// Number line with window starts marked:
	//    -2 -1 0 1 2 |3 4 5 6 7 |8 9 10 11 12 |13 14 15 16 17
	//                0          1             2
	// We can see that the index we want is 2
	// (target - zero) /every
	//    = (13 - 3) / 5
	//    = 10 / 5
	//    = 2
	// We do not adjust because the delta was positive

	// Example: Negative Index
	// zero = 3 target = -9 every = 5
	// Number line with window starts marked:
	//    |-12 -11 -10 -9 -8 |-7 -6 -5 -4 -3 |-2 -1 0 1 2 |3 4 5 6 7
	//   -3                 -2              -1            0
	// We can see that the index we want is -3
	// (target - zero) /every
	//    = (-9 - 3) / 5
	//    = -12 / 5
	//    = -2
	// We have to adjust by 1 because the delta was negative
	// and we get -3

	// Example: Negative Index on boundary
	// zero = 3 target = -7 every = 5
	// Number line with window starts marked:
	//    |-12 -11 -10 -9 -8 |-7 -6 -5 -4 -3 |-2 -1 0 1 2 |3 4 5 6 7
	//   -3                 -2              -1            0
	// We can see that the index we want is -2
	// (target - zero) /every
	//    = (-7 - 3) / 5
	//    = -10 / 5
	//    = -2
	// This time we land right on the boundary, since we are lower inclusive
	// we do not need to adjust.

	delta := target - zero
	index := delta / every

	// For targets before the zero we need to adjust the index,
	// but only if we did not land right on the boundary.
	if delta < 0 && delta%every != 0 {
		index -= 1
	}
	return int(index)
}

// monthsSince converts a time into the number of months since the unix epoch
func monthsSince(t values.Time) int64 {
	ts := t.Time()
	year, month, _ := ts.Date()
	return (int64(year)-epochYear)*12 + int64(month-1) - epochMonth
}

// isBeforeWithinMonth reports whether a comes before b within the month.
// The year and month of a and b are not relevant.
func isBeforeWithinMonth(a, b values.Time) bool {
	at := a.Time()
	bt := b.Time()
	ad := at.Day()
	bd := bt.Day()
	if ad > bd {
		return false
	}
	if ad < bd {
		return true
	}

	ah, am, as := at.Clock()
	bh, bm, bs := bt.Clock()
	if ah > bh {
		return false
	}
	if ah < bh {
		return true
	}
	if am > bm {
		return false
	}
	if am < bm {
		return true
	}
	if as > bs {
		return false
	}
	if as < bs {
		return true
	}
	an := at.Nanosecond()
	bn := bt.Nanosecond()
	if an > bn {
		return false
	}
	if an < bn {
		return true
	}
	return false
}
