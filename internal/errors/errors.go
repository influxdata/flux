package errors

import (
	"fmt"
	"strings"

	"github.com/influxdata/flux/codes"
)

// Error is the error struct of flux.
type Error struct {
	// Code is the code of the error as defined in the codes package.
	// This describes the type and category of the error. It is required.
	Code codes.Code

	// Msg contains a human-readable description and additional information
	// about the error itself. This is optional.
	Msg string

	// Err contains the error that was the cause of this error.
	// This is optional.
	Err error
}

// Error implement the error interface by outputting the Code and Err.
func (e *Error) Error() string {
	if e.Msg != "" && e.Err != nil {
		var b strings.Builder
		b.WriteString(e.Msg)
		b.WriteString(": ")
		b.WriteString(e.Err.Error())
		return b.String()
	} else if e.Msg != "" {
		return e.Msg
	} else if e.Err != nil {
		return e.Err.Error()
	}
	return e.Code.String()
}

// Unwrap will return the wrapped error.
func (e *Error) Unwrap() error {
	return e.Err
}

func New(code codes.Code, msg ...interface{}) error {
	return Wrap(nil, code, msg...)
}

func Newf(code codes.Code, fmtStr string, args ...interface{}) error {
	return Wrapf(nil, code, fmtStr, args...)
}

func Wrap(err error, code codes.Code, msg ...interface{}) error {
	var s string
	if len(msg) > 0 {
		s = fmt.Sprint(msg...)
	}
	return &Error{
		Code: code,
		Msg:  s,
		Err:  err,
	}
}

func Wrapf(err error, code codes.Code, format string, a ...interface{}) error {
	return &Error{
		Code: code,
		Msg:  fmt.Sprintf(format, a...),
		Err:  err,
	}
}
