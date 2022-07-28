// Generated by tmpl
// https://github.com/benbjohnson/tmpl
//
// DO NOT EDIT!
// Source: conditional.gen.go.tmpl

package array

import (
	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

func IntConditional(t *Boolean, c, a *Int, mem memory.Allocator) (*Int, error) {
	n := t.Len()
	b := NewIntBuilder(mem)
	b.Resize(n)

	if !(c.Len() == n && a.Len() == n) {
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length")
	}

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy && c.IsValid(i) {
			b.Append(c.Value(i))
		} else if !truthy && a.IsValid(i) {
			b.Append(a.Value(i))
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewIntArray()
	b.Release()
	return arr, nil
}

func IntConditionalCConst(t *Boolean, c int64, a *Int, mem memory.Allocator) (*Int, error) {
	n := t.Len()
	b := NewIntBuilder(mem)
	b.Resize(n)

	if !(a.Len() == n) {
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length")
	}

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy {
			b.Append(c)
		} else if !truthy && a.IsValid(i) {
			b.Append(a.Value(i))
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewIntArray()
	b.Release()
	return arr, nil
}

func IntConditionalAConst(t *Boolean, c *Int, a int64, mem memory.Allocator) (*Int, error) {
	n := t.Len()
	b := NewIntBuilder(mem)
	b.Resize(n)

	if !(c.Len() == n) {
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length")
	}

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy && c.IsValid(i) {
			b.Append(c.Value(i))
		} else if !truthy {
			b.Append(a)
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewIntArray()
	b.Release()
	return arr, nil
}

func IntConditionalCConstAConst(t *Boolean, c, a int64, mem memory.Allocator) (*Int, error) {
	n := t.Len()
	b := NewIntBuilder(mem)
	b.Resize(n)

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy {
			b.Append(c)
		} else if !truthy {
			b.Append(a)
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewIntArray()
	b.Release()
	return arr, nil
}

func UintConditional(t *Boolean, c, a *Uint, mem memory.Allocator) (*Uint, error) {
	n := t.Len()
	b := NewUintBuilder(mem)
	b.Resize(n)

	if !(c.Len() == n && a.Len() == n) {
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length")
	}

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy && c.IsValid(i) {
			b.Append(c.Value(i))
		} else if !truthy && a.IsValid(i) {
			b.Append(a.Value(i))
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewUintArray()
	b.Release()
	return arr, nil
}

func UintConditionalCConst(t *Boolean, c uint64, a *Uint, mem memory.Allocator) (*Uint, error) {
	n := t.Len()
	b := NewUintBuilder(mem)
	b.Resize(n)

	if !(a.Len() == n) {
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length")
	}

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy {
			b.Append(c)
		} else if !truthy && a.IsValid(i) {
			b.Append(a.Value(i))
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewUintArray()
	b.Release()
	return arr, nil
}

func UintConditionalAConst(t *Boolean, c *Uint, a uint64, mem memory.Allocator) (*Uint, error) {
	n := t.Len()
	b := NewUintBuilder(mem)
	b.Resize(n)

	if !(c.Len() == n) {
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length")
	}

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy && c.IsValid(i) {
			b.Append(c.Value(i))
		} else if !truthy {
			b.Append(a)
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewUintArray()
	b.Release()
	return arr, nil
}

func UintConditionalCConstAConst(t *Boolean, c, a uint64, mem memory.Allocator) (*Uint, error) {
	n := t.Len()
	b := NewUintBuilder(mem)
	b.Resize(n)

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy {
			b.Append(c)
		} else if !truthy {
			b.Append(a)
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewUintArray()
	b.Release()
	return arr, nil
}

func FloatConditional(t *Boolean, c, a *Float, mem memory.Allocator) (*Float, error) {
	n := t.Len()
	b := NewFloatBuilder(mem)
	b.Resize(n)

	if !(c.Len() == n && a.Len() == n) {
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length")
	}

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy && c.IsValid(i) {
			b.Append(c.Value(i))
		} else if !truthy && a.IsValid(i) {
			b.Append(a.Value(i))
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewFloatArray()
	b.Release()
	return arr, nil
}

func FloatConditionalCConst(t *Boolean, c float64, a *Float, mem memory.Allocator) (*Float, error) {
	n := t.Len()
	b := NewFloatBuilder(mem)
	b.Resize(n)

	if !(a.Len() == n) {
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length")
	}

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy {
			b.Append(c)
		} else if !truthy && a.IsValid(i) {
			b.Append(a.Value(i))
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewFloatArray()
	b.Release()
	return arr, nil
}

func FloatConditionalAConst(t *Boolean, c *Float, a float64, mem memory.Allocator) (*Float, error) {
	n := t.Len()
	b := NewFloatBuilder(mem)
	b.Resize(n)

	if !(c.Len() == n) {
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length")
	}

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy && c.IsValid(i) {
			b.Append(c.Value(i))
		} else if !truthy {
			b.Append(a)
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewFloatArray()
	b.Release()
	return arr, nil
}

func FloatConditionalCConstAConst(t *Boolean, c, a float64, mem memory.Allocator) (*Float, error) {
	n := t.Len()
	b := NewFloatBuilder(mem)
	b.Resize(n)

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy {
			b.Append(c)
		} else if !truthy {
			b.Append(a)
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewFloatArray()
	b.Release()
	return arr, nil
}

func StringConditional(t *Boolean, c, a *String, mem memory.Allocator) (*String, error) {
	n := t.Len()
	b := NewStringBuilder(mem)
	b.Resize(n)

	if !(c.Len() == n && a.Len() == n) {
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length")
	}

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy && c.IsValid(i) {
			b.Append(c.Value(i))
		} else if !truthy && a.IsValid(i) {
			b.Append(a.Value(i))
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewStringArray()
	b.Release()
	return arr, nil
}

func StringConditionalCConst(t *Boolean, c string, a *String, mem memory.Allocator) (*String, error) {
	n := t.Len()
	b := NewStringBuilder(mem)
	b.Resize(n)

	if !(a.Len() == n) {
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length")
	}

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy {
			b.Append(c)
		} else if !truthy && a.IsValid(i) {
			b.Append(a.Value(i))
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewStringArray()
	b.Release()
	return arr, nil
}

func StringConditionalAConst(t *Boolean, c *String, a string, mem memory.Allocator) (*String, error) {
	n := t.Len()
	b := NewStringBuilder(mem)
	b.Resize(n)

	if !(c.Len() == n) {
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length")
	}

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy && c.IsValid(i) {
			b.Append(c.Value(i))
		} else if !truthy {
			b.Append(a)
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewStringArray()
	b.Release()
	return arr, nil
}

func StringConditionalCConstAConst(t *Boolean, c, a string, mem memory.Allocator) (*String, error) {
	n := t.Len()
	b := NewStringBuilder(mem)
	b.Resize(n)

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy {
			b.Append(c)
		} else if !truthy {
			b.Append(a)
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewStringArray()
	b.Release()
	return arr, nil
}

func BooleanConditional(t *Boolean, c, a *Boolean, mem memory.Allocator) (*Boolean, error) {
	n := t.Len()
	b := NewBooleanBuilder(mem)
	b.Resize(n)

	if !(c.Len() == n && a.Len() == n) {
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length")
	}

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy && c.IsValid(i) {
			b.Append(c.Value(i))
		} else if !truthy && a.IsValid(i) {
			b.Append(a.Value(i))
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewBooleanArray()
	b.Release()
	return arr, nil
}

func BooleanConditionalCConst(t *Boolean, c bool, a *Boolean, mem memory.Allocator) (*Boolean, error) {
	n := t.Len()
	b := NewBooleanBuilder(mem)
	b.Resize(n)

	if !(a.Len() == n) {
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length")
	}

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy {
			b.Append(c)
		} else if !truthy && a.IsValid(i) {
			b.Append(a.Value(i))
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewBooleanArray()
	b.Release()
	return arr, nil
}

func BooleanConditionalAConst(t *Boolean, c *Boolean, a bool, mem memory.Allocator) (*Boolean, error) {
	n := t.Len()
	b := NewBooleanBuilder(mem)
	b.Resize(n)

	if !(c.Len() == n) {
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length")
	}

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy && c.IsValid(i) {
			b.Append(c.Value(i))
		} else if !truthy {
			b.Append(a)
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewBooleanArray()
	b.Release()
	return arr, nil
}

func BooleanConditionalCConstAConst(t *Boolean, c, a bool, mem memory.Allocator) (*Boolean, error) {
	n := t.Len()
	b := NewBooleanBuilder(mem)
	b.Resize(n)

	for i := 0; i < n; i++ {
		// nulls are considered as false
		truthy := t.IsValid(i) && t.Value(i)
		if truthy {
			b.Append(c)
		} else if !truthy {
			b.Append(a)
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewBooleanArray()
	b.Release()
	return arr, nil
}
