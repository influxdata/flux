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
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length") // FIXME: make message consistent with prior art
	}

	for i := 0; i < n; i++ {
		if t.IsValid(i) && c.IsValid(i) && a.IsValid(i) {
			// FIXME: is this?? Not sure when we need to append null.
			//  The standard conditional treats a null test as a false, but what about null consequent/alternate?
			if c.IsNull(i) || a.IsNull(i) {
				b.AppendNull()
			} else if t.IsNull(i) || !t.Value(i) {
				b.Append(a.Value(i)) // Falsy
			} else {
				b.Append(c.Value(i)) // Truthy
			}
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
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length") // FIXME: make message consistent with prior art
	}

	for i := 0; i < n; i++ {
		if t.IsValid(i) && a.IsValid(i) {
			// FIXME: is this?? Not sure when we need to append null.
			//  The standard conditional treats a null test as a false, but what about a null alternate?
			if a.IsNull(i) {
				b.AppendNull()
			} else if t.IsNull(i) || !t.Value(i) {
				b.Append(a.Value(i)) // Falsy
			} else {
				b.Append(c) // Truthy
			}
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
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length") // FIXME: make message consistent with prior art
	}

	for i := 0; i < n; i++ {
		if t.IsValid(i) && c.IsValid(i) {
			// FIXME: is this?? Not sure when we need to append null.
			//  The standard conditional treats a null test as a false, but what about a null consequent?
			if c.IsNull(i) {
				b.AppendNull()
			} else if t.IsNull(i) || !t.Value(i) {
				b.Append(a) // Falsy
			} else {
				b.Append(c.Value(i)) // Truthy
			}
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
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length") // FIXME: make message consistent with prior art
	}

	for i := 0; i < n; i++ {
		if t.IsValid(i) && c.IsValid(i) && a.IsValid(i) {
			// FIXME: is this?? Not sure when we need to append null.
			//  The standard conditional treats a null test as a false, but what about null consequent/alternate?
			if c.IsNull(i) || a.IsNull(i) {
				b.AppendNull()
			} else if t.IsNull(i) || !t.Value(i) {
				b.Append(a.Value(i)) // Falsy
			} else {
				b.Append(c.Value(i)) // Truthy
			}
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
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length") // FIXME: make message consistent with prior art
	}

	for i := 0; i < n; i++ {
		if t.IsValid(i) && a.IsValid(i) {
			// FIXME: is this?? Not sure when we need to append null.
			//  The standard conditional treats a null test as a false, but what about a null alternate?
			if a.IsNull(i) {
				b.AppendNull()
			} else if t.IsNull(i) || !t.Value(i) {
				b.Append(a.Value(i)) // Falsy
			} else {
				b.Append(c) // Truthy
			}
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
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length") // FIXME: make message consistent with prior art
	}

	for i := 0; i < n; i++ {
		if t.IsValid(i) && c.IsValid(i) {
			// FIXME: is this?? Not sure when we need to append null.
			//  The standard conditional treats a null test as a false, but what about a null consequent?
			if c.IsNull(i) {
				b.AppendNull()
			} else if t.IsNull(i) || !t.Value(i) {
				b.Append(a) // Falsy
			} else {
				b.Append(c.Value(i)) // Truthy
			}
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
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length") // FIXME: make message consistent with prior art
	}

	for i := 0; i < n; i++ {
		if t.IsValid(i) && c.IsValid(i) && a.IsValid(i) {
			// FIXME: is this?? Not sure when we need to append null.
			//  The standard conditional treats a null test as a false, but what about null consequent/alternate?
			if c.IsNull(i) || a.IsNull(i) {
				b.AppendNull()
			} else if t.IsNull(i) || !t.Value(i) {
				b.Append(a.Value(i)) // Falsy
			} else {
				b.Append(c.Value(i)) // Truthy
			}
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
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length") // FIXME: make message consistent with prior art
	}

	for i := 0; i < n; i++ {
		if t.IsValid(i) && a.IsValid(i) {
			// FIXME: is this?? Not sure when we need to append null.
			//  The standard conditional treats a null test as a false, but what about a null alternate?
			if a.IsNull(i) {
				b.AppendNull()
			} else if t.IsNull(i) || !t.Value(i) {
				b.Append(a.Value(i)) // Falsy
			} else {
				b.Append(c) // Truthy
			}
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
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length") // FIXME: make message consistent with prior art
	}

	for i := 0; i < n; i++ {
		if t.IsValid(i) && c.IsValid(i) {
			// FIXME: is this?? Not sure when we need to append null.
			//  The standard conditional treats a null test as a false, but what about a null consequent?
			if c.IsNull(i) {
				b.AppendNull()
			} else if t.IsNull(i) || !t.Value(i) {
				b.Append(a) // Falsy
			} else {
				b.Append(c.Value(i)) // Truthy
			}
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
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length") // FIXME: make message consistent with prior art
	}

	for i := 0; i < n; i++ {
		if t.IsValid(i) && c.IsValid(i) && a.IsValid(i) {
			// FIXME: is this?? Not sure when we need to append null.
			//  The standard conditional treats a null test as a false, but what about null consequent/alternate?
			if c.IsNull(i) || a.IsNull(i) {
				b.AppendNull()
			} else if t.IsNull(i) || !t.Value(i) {
				b.Append(a.Value(i)) // Falsy
			} else {
				b.Append(c.Value(i)) // Truthy
			}
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
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length") // FIXME: make message consistent with prior art
	}

	for i := 0; i < n; i++ {
		if t.IsValid(i) && a.IsValid(i) {
			// FIXME: is this?? Not sure when we need to append null.
			//  The standard conditional treats a null test as a false, but what about a null alternate?
			if a.IsNull(i) {
				b.AppendNull()
			} else if t.IsNull(i) || !t.Value(i) {
				b.Append(a.Value(i)) // Falsy
			} else {
				b.Append(c) // Truthy
			}
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
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length") // FIXME: make message consistent with prior art
	}

	for i := 0; i < n; i++ {
		if t.IsValid(i) && c.IsValid(i) {
			// FIXME: is this?? Not sure when we need to append null.
			//  The standard conditional treats a null test as a false, but what about a null consequent?
			if c.IsNull(i) {
				b.AppendNull()
			} else if t.IsNull(i) || !t.Value(i) {
				b.Append(a) // Falsy
			} else {
				b.Append(c.Value(i)) // Truthy
			}
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
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length") // FIXME: make message consistent with prior art
	}

	for i := 0; i < n; i++ {
		if t.IsValid(i) && c.IsValid(i) && a.IsValid(i) {
			// FIXME: is this?? Not sure when we need to append null.
			//  The standard conditional treats a null test as a false, but what about null consequent/alternate?
			if c.IsNull(i) || a.IsNull(i) {
				b.AppendNull()
			} else if t.IsNull(i) || !t.Value(i) {
				b.Append(a.Value(i)) // Falsy
			} else {
				b.Append(c.Value(i)) // Truthy
			}
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
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length") // FIXME: make message consistent with prior art
	}

	for i := 0; i < n; i++ {
		if t.IsValid(i) && a.IsValid(i) {
			// FIXME: is this?? Not sure when we need to append null.
			//  The standard conditional treats a null test as a false, but what about a null alternate?
			if a.IsNull(i) {
				b.AppendNull()
			} else if t.IsNull(i) || !t.Value(i) {
				b.Append(a.Value(i)) // Falsy
			} else {
				b.Append(c) // Truthy
			}
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
		return nil, errors.Newf(codes.Invalid, "vectors must be equal length") // FIXME: make message consistent with prior art
	}

	for i := 0; i < n; i++ {
		if t.IsValid(i) && c.IsValid(i) {
			// FIXME: is this?? Not sure when we need to append null.
			//  The standard conditional treats a null test as a false, but what about a null consequent?
			if c.IsNull(i) {
				b.AppendNull()
			} else if t.IsNull(i) || !t.Value(i) {
				b.Append(a) // Falsy
			} else {
				b.Append(c.Value(i)) // Truthy
			}
		} else {
			b.AppendNull()
		}
	}
	arr := b.NewBooleanArray()
	b.Release()
	return arr, nil
}
