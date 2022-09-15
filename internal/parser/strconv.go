package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

// ParseTime will parse a time literal from a string.
func ParseTime(lit string) (time.Time, error) {
	if !strings.Contains(lit, "T") {
		// This is a date.
		t, err := time.Parse("2006-01-02", lit)
		if err != nil {
			return time.Time{}, errors.New(codes.Invalid, "cannot parse date")
		}
		return t, nil
	}
	t, err := time.Parse(time.RFC3339Nano, lit)
	if err != nil {
		return time.Time{}, errors.New(codes.Invalid, "cannot parse date time")
	}
	return t, nil
}

// MustParseTime parses a time literal and panics in the case of an error.
func MustParseTime(lit string) time.Time {
	ts, err := ParseTime(lit)
	if err != nil {
		panic(err)
	}
	return ts
}

// ParseDuration will convert a string into components of the duration.
func ParseDuration(lit string) ([]ast.Duration, error) {
	var values []ast.Duration
	for len(lit) > 0 {
		n := 0
		for n < len(lit) {
			ch, size := utf8.DecodeRuneInString(lit[n:])
			if size == 0 {
				panic("invalid rune in duration")
			}

			if !unicode.IsDigit(ch) {
				break
			}
			n += size
		}

		if n == 0 {
			return nil, errors.Newf(codes.Invalid, "invalid duration %s", lit)
		}

		magnitude, err := strconv.ParseInt(lit[:n], 10, 64)
		if err != nil {
			return nil, err
		}
		lit = lit[n:]

		n = 0
		for n < len(lit) {
			ch, size := utf8.DecodeRuneInString(lit[n:])
			if size == 0 {
				panic("invalid rune in duration")
			}

			if !unicode.IsLetter(ch) {
				break
			}
			n += size
		}

		if n == 0 {
			return nil, errors.Newf(codes.Invalid, "duration is missing a unit: %s", lit)
		}

		unit := lit[:n]
		if unit == "µs" {
			unit = "us"
		}
		values = append(values, ast.Duration{
			Magnitude: magnitude,
			Unit:      unit,
		})
		lit = lit[n:]
	}
	return values, nil
}

// ParseString removes quotes and unescapes the string literal.
func ParseString(lit string) (string, error) {
	if len(lit) < 2 || lit[0] != '"' || lit[len(lit)-1] != '"' {
		return "", fmt.Errorf("invalid syntax")
	}
	return ParseText(lit[1 : len(lit)-1])
}

// ParseText parses a UTF-8 block of text with escaping rules.
func ParseText(lit string) (string, error) {
	var (
		builder    strings.Builder
		width, pos int
		err        error
	)
	builder.Grow(len(lit))
	for pos < len(lit) {
		width, err = writeNextUnescapedRune(lit[pos:], &builder)
		if err != nil {
			return "", err
		}
		pos += width
	}
	return builder.String(), nil
}

// writeNextUnescapedRune writes a rune to builder from s.
// The rune is the next decoded UTF-8 rune with escaping rules applied.
func writeNextUnescapedRune(s string, builder *strings.Builder) (width int, err error) {
	var r rune
	r, width = utf8.DecodeRuneInString(s)
	if r == '\\' {
		next, w := utf8.DecodeRuneInString(s[width:])
		width += w
		switch next {
		case 'n':
			r = '\n'
		case 'r':
			r = '\r'
		case 't':
			r = '\t'
		case '\\':
			r = '\\'
		case '"':
			r = '"'
		case '$':
			r = '$'
		case 'x':
			b, err := fromHexDigits(s[width:])
			if err != nil {
				return 0, err
			}
			builder.WriteByte(b)
			return width + 2, nil
		default:
			return 0, fmt.Errorf("invalid escape character %q", next)
		}
	}
	// sanity check before writing the rune
	if width > 0 {
		builder.WriteRune(r)
	}
	return
}

// fromHexDigits decodes a single byte from two hex digits from the string or an error
func fromHexDigits(s string) (byte, error) {
	// Decode two hex chars as a single byte
	if len(s) < 2 {
		return 0, errors.New(codes.Invalid, "expected 2 hex characters")
	}
	ch1, ok1 := fromHexChar(s[0])
	ch2, ok2 := fromHexChar(s[1])
	if !ok1 || !ok2 {
		return 0, fmt.Errorf("invalid byte value %q", s)
	}
	return ((ch1 << 4) | ch2), nil
}

// fromHexChar converts a hex character into its value and a success flag.
func fromHexChar(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

// ParseRegexp converts text surrounded by forward slashes into a regular expression.
func ParseRegexp(lit string) (*regexp.Regexp, error) {
	if len(lit) < 3 {
		return nil, fmt.Errorf("regexp must be at least 3 characters")
	}

	if lit[0] != '/' {
		return nil, fmt.Errorf("regexp literal must start with a slash")
	} else if lit[len(lit)-1] != '/' {
		return nil, fmt.Errorf("regexp literal must end with a slash")
	}

	expr := lit[1 : len(lit)-1]
	// Unescape regex literal
	var (
		builder    strings.Builder
		width, pos int
		err        error
	)
	builder.Grow(len(expr))
	for pos < len(expr) {
		width, err = writeNextUnescapedRegexRune(expr[pos:], &builder)
		if err != nil {
			return nil, err
		}
		pos += width
	}
	return regexp.Compile(builder.String())

}

// writeNextUnescapedRegexRune writes a rune to builder from s.
// The rune is the next decoded UTF-8 rune with regex escaping rules applied.
func writeNextUnescapedRegexRune(s string, builder *strings.Builder) (int, error) {
	r, width := utf8.DecodeRuneInString(s)
	if r == '\\' {
		next, w := utf8.DecodeRuneInString(s[width:])
		width += w
		switch next {
		case '/':
			builder.WriteRune('/')
			return width, nil
		case 'x':
			b, err := fromHexDigits(s[width:])
			if err != nil {
				return 0, err
			}
			builder.WriteByte(b)
			return width + 2, nil
		default:
			// Standard regexp escape characters may exist,
			// we leave them alone and let Go's regex parser validate them.
			builder.WriteRune('\\')
			builder.WriteRune(next)
			return width, nil
		}
	}
	builder.WriteRune(r)
	return width, nil
}
