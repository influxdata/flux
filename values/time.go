package values

import (
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/parser"
)

type Time int64

// Duration is a vector representing the duration unit components.
type Duration struct {
	// months is the number of months for the duration.
	// This must be a positive value.
	months int64

	// nsecs is the number of nanoseconds for the duration.
	// This must be a positive value.
	nsecs int64

	// negative indicates this duration is a negative value.
	negative bool
}

const (
	fixedWidthTimeFmt = "2006-01-02T15:04:05.000000000Z"
)

func ConvertTime(t time.Time) Time {
	return Time(t.UnixNano())
}

// ConvertDuration takes a time.Duration and converts it into a Duration.
func ConvertDuration(v time.Duration) Duration {
	negative := false
	if v < 0 {
		negative, v = true, -v
	}
	return Duration{
		negative: negative,
		nsecs:    int64(v),
	}
}

func (t Time) Round(d Duration) Time {
	if !d.IsPositive() {
		return t
	}
	r := t.Remainder(d)
	if lessThanHalf(r, d) {
		return t - Time(r.Duration())
	}
	return t + Time(d.Duration()-r.Duration())
}

func (t Time) Truncate(d Duration) Time {
	if !d.IsPositive() {
		return t
	}
	r := t.Remainder(d)
	return t - Time(r.Duration())
}

func (t Time) Add(d Duration) Time {
	return t + Time(d.Duration())
}

// Sub takes another time and returns a duration giving the duration
// between the two times. A positive duration indicates that the receiver
// occurs after the other time.
func (t Time) Sub(other Time) Duration {
	return ConvertDuration(time.Duration(t - other))
}

// Remainder divides t by d and returns the remainder.
func (t Time) Remainder(d Duration) (r Duration) {
	return ConvertDuration(time.Duration(t) % d.Duration())
}

// lessThanHalf reports whether x+x < y but avoids overflow,
// assuming x and y are both positive (Duration is signed).
func lessThanHalf(x, y Duration) bool {
	xnsecs, ynsecs := x.Duration(), y.Duration()
	return uint64(xnsecs)+uint64(xnsecs) < uint64(ynsecs)
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
	if scale < 0 {
		scale = -scale
		d.negative = !d.negative
	}
	d.months *= int64(scale)
	d.nsecs *= int64(scale)
	return d
}

// IsPositive returns true if this is a positive number.
// It returns false if the number is zero.
func (d Duration) IsPositive() bool {
	return !d.negative && !d.IsZero()
}

// IsZero returns true if this is a zero duration.
func (d Duration) IsZero() bool {
	return d.months == 0 && d.nsecs == 0
}

// Normalize will normalize the duration within the interval.
// It will ensure that the output duration is the smallest positive
// duration that is the equivalent of the current duration.
func (d Duration) Normalize(interval Duration) Duration {
	offset, every := int64(d.Duration()), int64(interval.Duration())
	if offset < 0 {
		offset += every * ((offset / -every) + 1)
	} else if offset > every {
		offset -= every * (offset / every)
	}
	return ConvertDuration(time.Duration(offset))
}

// Equal returns true if the two durations are equal.
func (d Duration) Equal(other Duration) bool {
	return d.negative == other.negative &&
		d.months == other.months &&
		d.nsecs == other.nsecs
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
	months := d.months * 30 * 24 * int64(time.Hour)
	dur := d.nsecs + months
	if d.negative {
		dur = -dur
	}
	return time.Duration(dur)
}

func (d Duration) String() string {
	if d.IsZero() {
		return "0ns"
	}

	var buf []byte
	writeInt := func(sb *strings.Builder, v int64, unit string) {
		buf = strconv.AppendInt(buf, v, 10)
		sb.Grow(len(buf) + len(unit))
		sb.Write(buf)
		sb.WriteString(unit)
		buf = buf[:0]
	}
	var sb strings.Builder
	if d.negative {
		sb.WriteByte('-')
	}

	if d.months > 0 {
		if years := d.months / 12; years > 0 {
			writeInt(&sb, years, "y")
		}
		if months := d.months % 12; months > 0 {
			writeInt(&sb, months, "mo")
		}
	}

	if d.nsecs > 0 {
		nsecs := d.nsecs % 1000
		d.nsecs /= 1000
		usecs := d.nsecs % 1000
		d.nsecs /= 1000
		msecs := d.nsecs % 1000
		d.nsecs /= 1000
		secs := d.nsecs % 60
		d.nsecs /= 60
		mins := d.nsecs % 60
		d.nsecs /= 60
		hours := d.nsecs % 24
		d.nsecs /= 24
		days := d.nsecs % 7
		d.nsecs /= 7
		weeks := d.nsecs

		if weeks > 0 {
			writeInt(&sb, weeks, "w")
		}
		if days > 0 {
			writeInt(&sb, days, "d")
		}
		if hours > 0 {
			writeInt(&sb, hours, "h")
		}
		if mins > 0 {
			writeInt(&sb, mins, "m")
		}
		if secs > 0 {
			writeInt(&sb, secs, "s")
		}
		if msecs > 0 {
			writeInt(&sb, msecs, "ms")
		}
		if usecs > 0 {
			writeInt(&sb, usecs, "us")
		}
		if nsecs > 0 {
			writeInt(&sb, nsecs, "ns")
		}
	}
	return sb.String()
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
	dur, err := parser.ParseSignedDuration(s)
	if err != nil {
		return Duration{}, err
	}
	return DurationFromAST(dur)
}

// DurationFromAST creates a duration value from the AST DurationLiteral.
func DurationFromAST(dur *ast.DurationLiteral) (d Duration, err error) {
	if len(dur.Values) == 0 {
		return d, nil
	}

	// Determine if this duration is negative. Every other value
	// must be consistent with this.
	d.negative = dur.Values[0].Magnitude < 0
	for _, du := range dur.Values {
		mag, unit := du.Magnitude, du.Unit
		if (mag >= 0 && d.negative) || (mag < 0 && !d.negative) {
			return Duration{}, errors.New(codes.Invalid, "duration magnitudes must be the same sign")
		}

		if d.negative {
			mag = -mag
		}

		switch unit {
		case "y":
			mag *= 12
			fallthrough
		case "mo":
			d.months += mag
		case "w":
			mag *= 7
			fallthrough
		case "d":
			mag *= 24
			fallthrough
		case "h":
			mag *= 60
			fallthrough
		case "m":
			mag *= 60
			fallthrough
		case "s":
			mag *= 1000
			fallthrough
		case "ms":
			mag *= 1000
			fallthrough
		case "us", "µs":
			mag *= 1000
			fallthrough
		case "ns":
			d.nsecs += mag
		}
	}
	return d, nil
}
