package values

import (
	"time"
)

type Time int64
type Duration int64

const (
	fixedWidthTimeFmt = "2006-01-02T15:04:05.000000000Z"
)

func ConvertTime(t time.Time) Time {
	return Time(t.UnixNano())
}

// ConvertDuration takes a time.Duration and converts it into a Duration.
func ConvertDuration(v time.Duration) Duration {
	return Duration(v)
}

func (t Time) Round(d Duration) Time {
	if d <= 0 {
		return t
	}
	r := t.Remainder(d)
	if lessThanHalf(r, d) {
		return t - Time(r)
	}
	return t + Time(d-r)
}

func (t Time) Truncate(d Duration) Time {
	if d <= 0 {
		return t
	}
	r := t.Remainder(d)
	return t - Time(r)
}

func (t Time) Add(d Duration) Time {
	return t + Time(d)
}

// Sub takes another time and returns a duration giving the duration
// between the two times. A positive duration indicates that the receiver
// occurs after the other time.
func (t Time) Sub(other Time) Duration {
	return Duration(t - other)
}

// Remainder divides t by d and returns the remainder.
func (t Time) Remainder(d Duration) (r Duration) {
	return Duration(int64(t) % int64(d))
}

// lessThanHalf reports whether x+x < y but avoids overflow,
// assuming x and y are both positive (Duration is signed).
func lessThanHalf(x, y Duration) bool {
	return uint64(x)+uint64(x) < uint64(y)
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
	return d * Duration(scale)
}

// IsPositive returns true if this is a positive number.
// It returns false if the number is zero.
func (d Duration) IsPositive() bool {
	return d > 0
}

// IsZero returns true if this is a zero duration.
func (d Duration) IsZero() bool {
	return d == 0
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
	return time.Duration(d)
}
func (d Duration) String() string {
	return time.Duration(d).String()
}
func ParseDuration(s string) (Duration, error) {
	// TODO(jsternberg): This should use the real duration parsing
	// instead of time.ParseDuration.
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	return Duration(d), nil
}
