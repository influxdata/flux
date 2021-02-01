package values

import "github.com/influxdata/flux/semantic"

var (
	trueValue Value = value{
		t: semantic.BasicBoolMonoType,
		v: true,
	}
	falseValue Value = value{
		t: semantic.BasicBoolMonoType,
		v: false,
	}
)

func NewBool(v bool) Value {
	if v {
		return trueValue
	}
	return falseValue
}
