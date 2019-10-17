package values

import (
	"time"
)

type Time int64

// Duration is a vector representing the duration unit components.
type Duration struct {
	// nsecs is the number of nanoseconds for the duration.
	nsecs int64
}

const (
	fixedWidthTimeFmt = "2006-01-02T15:04:05.000000000Z"
)

func ConvertTime(t time.Time) Time {
	return Time(t.UnixNano())
}

// ConvertDuration takes a time.Duration and converts it into a Duration.
func ConvertDuration(v time.Duration) Duration {
	return Duration{nsecs: int64(v)}
}

func (t Time) Round(d Duration) Time {
	if d.nsecs <= 0 {
		return t
	}
	r := t.Remainder(d)
	if lessThanHalf(r, d) {
		return t - Time(r.nsecs)
	}
	return t + Time(d.nsecs-r.nsecs)
}

func (t Time) Truncate(d Duration) Time {
	if d.nsecs <= 0 {
		return t
	}
	r := t.Remainder(d)
	return t - Time(r.nsecs)
}

func (t Time) Add(d Duration) Time {
	return t + Time(d.nsecs)
}

// Sub takes another time and returns a duration giving the duration
// between the two times. A positive duration indicates that the receiver
// occurs after the other time.
func (t Time) Sub(other Time) Duration {
	return Duration{nsecs: int64(t - other)}
}

// Remainder divides t by d and returns the remainder.
func (t Time) Remainder(d Duration) (r Duration) {
	return Duration{nsecs: int64(t) % int64(d.nsecs)}
}

// lessThanHalf reports whether x+x < y but avoids overflow,
// assuming x and y are both positive (Duration is signed).
func lessThanHalf(x, y Duration) bool {
	return uint64(x.nsecs)+uint64(x.nsecs) < uint64(y.nsecs)
}

func (t Time) String() string {
	return t.Time().Format(fixedWidthTimeFmt)
}

func ParseTime(s string) (Time, error) {
	t, err := time.Parse(fixedWidthTimeFmt, s)
	if err != nil {
		return 0, err
	}
	return ConvertTime(t), nil
}

func (t Time) Time() time.Time {
	return time.Unix(0, int64(t)).UTC()
}

// Mul will multiply the Duration by a scalar.
// This multiplies each component of the vector.
func (d Duration) Mul(scale int) Duration {
	return Duration{
		nsecs: d.nsecs * int64(scale),
	}
}

// IsPositive returns true if this is a positive number.
// It returns false if the number is zero.
func (d Duration) IsPositive() bool {
	return d.nsecs > 0
}

// IsZero returns true if this is a zero duration.
func (d Duration) IsZero() bool {
	return d.nsecs == 0
}

// Normalize will normalize the duration within the interval.
// It will ensure that the output duration is the smallest positive
// duration that is the equivalent of the current duration.
func (d Duration) Normalize(interval Duration) Duration {
	offset, every := d.nsecs, interval.nsecs
	if offset < 0 {
		offset += every * ((offset / -every) + 1)
	} else if offset > every {
		offset -= every * (offset / every)
	}
	return Duration{nsecs: offset}
}

// Equal returns true if the two durations are equal.
func (d Duration) Equal(other Duration) bool {
	return d.nsecs == other.nsecs
}

// Duration will return the nanosecond equivalent
// of this duration. It will assume that months are
// the equivalent of 30 days.
//
// It is recommended not to use this method unless
// it is absolutely needed. This method will lose
// any precision that is present in the Duration
// and it should only be used for interfacing with
// outside code that is not month-aware.
func (d Duration) Duration() time.Duration {
	return time.Duration(d.nsecs)
}
func (d Duration) String() string {
	return time.Duration(d.nsecs).String()
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}
func (d *Duration) UnmarshalText(data []byte) error {
	dur, err := ParseDuration(string(data))
	if err != nil {
		return err
	}
	*d = dur
	return nil
}

func ParseDuration(s string) (Duration, error) {
	// TODO(jsternberg): This should use the real duration parsing
	// instead of time.ParseDuration.
	d, err := time.ParseDuration(s)
	if err != nil {
		return Duration{}, err
	}
	return Duration{nsecs: int64(d)}, nil
}
