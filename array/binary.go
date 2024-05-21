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
		var elmL *bool
		if l.IsValid(i) {
			x := l.Value(i)
			elmL = &x
		}
		var elmR *bool
		if r.IsValid(i) {
			x := r.Value(i)
			elmR = &x
		}

		if elmL == nil && elmR == nil {
			// both sides are null
			b.AppendNull()
		} else if elmL == nil || elmR == nil {
			// one side is null, the other is false
			if (elmL != nil && !*elmL) || (elmR != nil && !*elmR) {
				b.Append(false)
			} else {
				b.AppendNull()
			}
		} else {
			// no nulls
			b.Append(*elmL && *elmR)
		}
	}
	a := b.NewBooleanArray()
	b.Release()
	return a, nil
}

func AndConst(fixed *bool, arr *Boolean, mem memory.Allocator) (*Boolean, error) {
	n := arr.Len()
	b := NewBooleanBuilder(mem)
	b.Resize(n)
	for i := 0; i < n; i++ {
		var elm *bool
		if arr.IsValid(i) {
			x := arr.Value(i)
			elm = &x
		}
		if fixed == nil && elm == nil {
			// both sides are null
			b.AppendNull()
		} else if fixed == nil || elm == nil {
			// one side is null, the other is false
			if (fixed != nil && !*fixed) || (elm != nil && !*elm) {
				b.Append(false)
			} else {
				b.AppendNull()
			}
		} else {
			// no nulls
			b.Append(*fixed && *elm)
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
		var elmL *bool
		if l.IsValid(i) {
			x := l.Value(i)
			elmL = &x
		}
		var elmR *bool
		if r.IsValid(i) {
			x := r.Value(i)
			elmR = &x
		}

		if elmL == nil && elmR == nil {
			// both sides are null
			b.AppendNull()
		} else if elmL == nil || elmR == nil {
			// one side is null, the other is true
			if (elmL != nil && *elmL) || (elmR != nil && *elmR) {
				b.Append(true)
			} else {
				b.AppendNull()
			}
		} else {
			// no nulls
			b.Append(*elmL || *elmR)
		}
	}
	a := b.NewBooleanArray()
	b.Release()
	return a, nil
}

func OrConst(fixed *bool, arr *Boolean, mem memory.Allocator) (*Boolean, error) {
	n := arr.Len()
	b := NewBooleanBuilder(mem)
	b.Resize(n)
	for i := 0; i < n; i++ {
		var elm *bool
		if arr.IsValid(i) {
			x := arr.Value(i)
			elm = &x
		}
		if fixed == nil && elm == nil {
			// both sides are null
			b.AppendNull()
		} else if fixed == nil || elm == nil {
			// one side is null, the other is true
			if (fixed != nil && *fixed) || (elm != nil && *elm) {
				b.Append(true)
			} else {
				b.AppendNull()
			}
		} else {
			// no nulls
			b.Append(*fixed || *elm)
		}
	}
	a := b.NewBooleanArray()
	b.Release()
	return a, nil
}

func StringAdd(l, r *String, mem memory.Allocator) (*String, error) {
	n := l.Len()
	if n != r.Len() {
		return nil, errors.Newf(codes.Invalid, "vectors must have equal length for binary operations")
	}
	b := NewStringBuilder(mem)
	b.Resize(n)
	for i := 0; i < n; i++ {
		if l.IsValid(i) && r.IsValid(i) {
			ls := l.Value(i)
			rs := r.Value(i)
			buf := make([]byte, len(ls)+len(rs))
			copy(buf, ls)
			copy(buf[len(ls):], rs)
			b.AppendBytes(buf)

		} else {
			b.AppendNull()
		}
	}
	a := b.NewStringArray()
	b.Release()
	return a, nil
}

func StringAddLConst(l string, r *String, mem memory.Allocator) (*String, error) {
	n := r.Len()
	b := NewStringBuilder(mem)
	b.Resize(n)
	for i := 0; i < n; i++ {
		if r.IsValid(i) {
			rs := r.Value(i)
			buf := make([]byte, len(l)+len(rs))
			copy(buf, l)
			copy(buf[len(l):], rs)
			b.AppendBytes(buf)

		} else {
			b.AppendNull()
		}
	}
	a := b.NewStringArray()
	b.Release()
	return a, nil
}

func StringAddRConst(l *String, r string, mem memory.Allocator) (*String, error) {
	n := l.Len()
	b := NewStringBuilder(mem)
	b.Resize(n)
	for i := 0; i < n; i++ {
		if l.IsValid(i) {
			ls := l.Value(i)
			buf := make([]byte, len(ls)+len(r))
			copy(buf, ls)
			copy(buf[len(ls):], r)
			b.AppendBytes(buf)

		} else {
			b.AppendNull()
		}
	}
	a := b.NewStringArray()
	b.Release()
	return a, nil
}
