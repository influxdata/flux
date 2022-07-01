package asttest

import (
	"regexp"

	"github.com/google/go-cmp/cmp"
)

//go:generate go run github.com/influxdata/flux/internal/cmd/cmpgen cmpopts.go

var CompareOptions = append(IgnoreBaseNodeOptions,
	cmp.Comparer(func(x, y *regexp.Regexp) bool { return x.String() == y.String() }),
)
