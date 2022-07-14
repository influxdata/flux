package array

import (
	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

func And(l, r *Boolean, mem memory.Allocator) (*Boolean, error) {
	n := l.Len()
	if n != r.Len() {
		return nil, errors.Newf(codes.Invalid, "vectors must have equal length for binary operations")
	}

	b := NewBooleanBuilder(mem)
	b.Resize(n)
	for i := 0; i < n; i++ {
		if l.IsValid(i) && r.IsValid(i) {
			b.Append(l.Value(i) && r.Value(i))
		} else {
			b.AppendNull()
		}
	}
	a := b.NewBooleanArray()
	b.Release()
	return a, nil
}

func AndLConst(l bool, r *Boolean, mem memory.Allocator) (*Boolean, error) {
	n := r.Len()
	b := NewBooleanBuilder(mem)
	b.Resize(n)
	for i := 0; i < n; i++ {
		if r.IsValid(i) {
			b.Append(l && r.Value(i))
		} else {
			b.AppendNull()
		}
	}
	a := b.NewBooleanArray()
	b.Release()
	return a, nil
}

func AndRConst(l *Boolean, r bool, mem memory.Allocator) (*Boolean, error) {
	n := l.Len()
	b := NewBooleanBuilder(mem)
	b.Resize(n)
	for i := 0; i < n; i++ {
		if l.IsValid(i) {
			b.Append(l.Value(i) && r)
		} else {
			b.AppendNull()
		}
	}
	a := b.NewBooleanArray()
	b.Release()
	return a, nil
}

func Or(l, r *Boolean, mem memory.Allocator) (*Boolean, error) {
	n := l.Len()
	if n != r.Len() {
		return nil, errors.Newf(codes.Invalid, "vectors must have equal length for binary operations")
	}

	b := NewBooleanBuilder(mem)
	b.Resize(n)
	for i := 0; i < n; i++ {
		if l.IsValid(i) && r.IsValid(i) {
			b.Append(l.Value(i) || r.Value(i))
		} else {
			b.AppendNull()
		}
	}
	a := b.NewBooleanArray()
	b.Release()
	return a, nil
}

func OrLConst(l bool, r *Boolean, mem memory.Allocator) (*Boolean, error) {
	n := r.Len()
	b := NewBooleanBuilder(mem)
	b.Resize(n)
	for i := 0; i < n; i++ {
		if r.IsValid(i) {
			b.Append(l || r.Value(i))
		} else {
			b.AppendNull()
		}
	}
	a := b.NewBooleanArray()
	b.Release()
	return a, nil
}

func OrRConst(l *Boolean, r bool, mem memory.Allocator) (*Boolean, error) {
	n := l.Len()
	b := NewBooleanBuilder(mem)
	b.Resize(n)
	for i := 0; i < n; i++ {
		if l.IsValid(i) {
			b.Append(l.Value(i) || r)
		} else {
			b.AppendNull()
		}
	}
	a := b.NewBooleanArray()
	b.Release()
	return a, nil
}
