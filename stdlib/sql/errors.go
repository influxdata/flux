package sql

import (
	"github.com/InfluxCommunity/flux/codes"
	"github.com/InfluxCommunity/flux/internal/errors"
)

// ErrorDriverDisabled indicates a given database driver is disabled.
var ErrorDriverDisabled = errors.New(codes.Unimplemented, "database driver disabled")
