package array

import (
	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

func IntAdd(l, r *Int, mem memory.Allocator) (*Int, error) {
	n := l.Len()
	if n != r.Len() {
		return nil, errors.Newf(codes.Invalid, "Vectors must have equal length for binary operations")
	}

	b := NewIntBuilder(mem)
	b.Resize(n)
	lValues := l.Int64Values()
	rValues := r.Int64Values()
	for i := 0; i < n; i++ {
		if l.IsValid(i) && r.IsValid(i) {
			b.Append(lValues[i] + rValues[i])
		} else {
			b.AppendNull()
		}
	}
	return b.NewIntArray(), nil
}
