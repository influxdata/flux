package asttest

import (
	"regexp"

	"github.com/google/go-cmp/cmp"
)

var CmpOptions = []cmp.Option{
	cmp.Comparer(func(x, y *regexp.Regexp) bool {
		if x == nil && y == nil {
			return true
		}
		if x == nil || y == nil {
			return false
		}
		return x.String() == y.String()
	}),
}
