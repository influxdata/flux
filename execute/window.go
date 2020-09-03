package execute

import (
	"time"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/values"
)

type Window struct {
	Every  Duration
	Period Duration
	Offset Duration
}

// NewWindow creates a window with the given parameters,
// and normalizes the offset to a small positive duration.
// It also validates that the durations are valid when
// used within a window.
func NewWindow(every, period, offset Duration, months bool) (Window, error) {
	if !months {
		// Normalize nanosecond offsets to a small positive duration
		offset = offset.Normalize(every)
	}
	w := Window{
		Every:  every,
		Period: period,
		Offset: offset,
	}
	if err := w.IsValid(); err != nil {
		return Window{}, err
	}
	return w, nil
}

type truncateFunc func(t Time, d Duration) Time

func (w *Window) getTruncateFunc(d Duration) (truncateFunc, error) {
	switch months, nsecs := d.Months(), d.Nanoseconds(); {
	case months != 0 && nsecs != 0:
		const docURL = "https://v2.docs.influxdata.com/v2.0/reference/flux/stdlib/built-in/transformations/window/#calendar-months-and-years"
		return nil, errors.New(codes.Invalid, "duration used as an interval cannot mix month and nanosecond units").
			WithDocURL(docURL)
	case months != 0:
		return truncateByMonths, nil
	case nsecs != 0:
		return truncateByNsecs, nil
	default:
		return nil, errors.New(codes.Invalid, "duration used as an interval cannot be zero")
	}
}

// truncate will truncate the time using the duration.
func (w *Window) truncate(t Time) Time {
	fn, err := w.getTruncateFunc(w.Every)
	if err != nil {
		panic(err)
	}
	return fn(t, w.Every)
}

// IsValid will check if this Window is valid and it will
// return an error if it isn't.
func (w Window) IsValid() error {
	_, err := w.getTruncateFunc(w.Every)
	return err
}

// GetEarliestBounds returns the bounds for the earliest window bounds
// that contains the given time t.  For underlapping windows that
// do not contain time t, the window directly after time t will be returned.
func (w Window) GetEarliestBounds(t Time) Bounds {
	// translate to not-offset coordinate
	t = t.Add(w.Offset.Mul(-1))

	stop := w.truncate(t).Add(w.Every)

	// translate to offset coordinate
	stop = stop.Add(w.Offset)

	start := stop.Add(w.Period.Mul(-1))
	return Bounds{
		Start: start,
		Stop:  stop,
	}
}

// GetOverlappingBounds returns a slice of bounds for each window
// that overlaps the input bounds b.
func (w Window) GetOverlappingBounds(b Bounds) []Bounds {
	if b.IsEmpty() {
		return []Bounds{}
	}

	// Estimate the number of windows by using a rough approximation.
	c := (b.Duration().Duration() / w.Every.Duration()) + (w.Period.Duration() / w.Every.Duration())
	bs := make([]Bounds, 0, c)

	bi := w.GetEarliestBounds(b.Start)
	for bi.Start < b.Stop {
		bs = append(bs, bi)
		bi.Start = bi.Start.Add(w.Every)
		bi.Stop = bi.Stop.Add(w.Every)
	}
	return bs
}

// truncateByNsecs will truncate the time to the given number
// of nanoseconds.
func truncateByNsecs(t Time, d Duration) Time {
	remainder := int64(t) % d.Nanoseconds()
	return t - Time(remainder)
}

// truncateByMonths will truncate the time to the given
// number of months.
func truncateByMonths(t Time, d Duration) Time {
	ts := t.Time()
	year, month, _ := ts.Date()

	// Determine the total number of months and truncate
	// the number of months by the duration amount.
	total := int64(year*12) + int64(month-1)
	remainder := total % d.Months()
	total -= remainder

	// Recreate a new time from the year and month combination.
	year, month = int(total/12), time.Month(total%12)+1
	ts = time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	return values.ConvertTime(ts)
}
