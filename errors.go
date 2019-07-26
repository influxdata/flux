package flux

import (
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

type Error = errors.Error

// ErrorCode returns the error code for the given error.
// If the error is not a flux.Error, this will return
// Unknown for the code. If the error is a flux.Error
// and its code is Inherit, then this will return the
// wrapped error's code.
func ErrorCode(err error) codes.Code {
	return errors.Code(err)
}
