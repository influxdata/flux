package array

import (
	"github.com/InfluxCommunity/flux/codes"
	"github.com/InfluxCommunity/flux/internal/errors"
	"github.com/apache/arrow/go/v7/arrow/memory"
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
			lb := l.ValueBytes(i)
			rb := r.ValueBytes(i)
			buf := make([]byte, len(lb)+len(rb))
			copy(buf, lb)
			copy(buf[len(lb):], rb)
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
			rb := r.ValueBytes(i)
			buf := make([]byte, len(l)+len(rb))
			copy(buf, l)
			copy(buf[len(l):], rb)
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
			lb := l.ValueBytes(i)
			buf := make([]byte, len(lb)+len(r))
			copy(buf, lb)
			copy(buf[len(lb):], r)
			b.AppendBytes(buf)

		} else {
			b.AppendNull()
		}
	}
	a := b.NewStringArray()
	b.Release()
	return a, nil
}
