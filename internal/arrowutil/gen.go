package arrowutil

import (
	"fmt"
	"strings"

	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
)

//go:generate -command tmpl ../../gotool.sh github.com/benbjohnson/tmpl
//go:generate tmpl -data=@types.tmpldata -o array_values.gen.go array_values.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o builder.gen.go builder.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o compare.gen.go compare.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o copy.gen.go copy.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o iterator.gen.go iterator.gen.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o iterator.gen_test.go iterator.gen_test.go.tmpl
//go:generate tmpl -data=@types.tmpldata -o filter.gen.go filter.gen.go.tmpl

var _ values.ArrayElementwiser = (*FloatArrayValue)(nil)

func (v FloatArrayValue) ElementwiseAdd(mem *memory.Allocator, other values.ArrayElementwiser) values.Array {
	ao := other.(FloatArrayValue)
	if v.arr.Len() != ao.arr.Len() {
		panic("cannot add arrays of different lengths")
	}
	b := NewFloatBuilder(mem)
	for i := 0; i < v.arr.Len(); i++ {
		if v.arr.IsValid(i) && ao.arr.IsValid(i) {
			b.Append(v.arr.Value(i) + ao.arr.Value(i))
		} else {
			b.AppendNull()
		}
	}
	return NewFloatArrayValue(b.NewFloatArray())
}

func (v FloatArrayValue) ElementwiseGT(mem *memory.Allocator, rhs values.Value) values.Array {
	rhsFlt := rhs.Float()
	b := NewBooleanBuilder(mem)
	for i := 0; i < v.arr.Len(); i++ {
		if v.arr.IsValid(i) {
			b.Append(v.arr.Value(i) > rhsFlt)
		} else {
			b.AppendNull()
		}
	}
	return NewBooleanArrayValue(b.NewBooleanArray())
}

func (v FloatArrayValue) GetArrowArray() *array.Float {
	return v.arr
}

func (v FloatArrayValue) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	for i := 0; i < v.arr.Len(); i++ {
		if i != 0 {
			sb.WriteString(", ")
		}
		if v.arr.IsValid(i) {
			sb.WriteString(fmt.Sprintf("%v", v.arr.Value(i)))
		} else {
			sb.WriteString("null")
		}
	}
	sb.WriteString("]")
	return sb.String()
}
