package sql

import (
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

// ErrorDriverDisabled indicates a given database driver is disabled.
var ErrorDriverDisabled = errors.New(codes.Unimplemented, "database driver disabled")
