package interpreter

import "github.com/influxdata/flux/values"

type Scope interface {
	Lookup(name string) (values.Value, bool)

	Set(name string, v values.Value)

	Nest() Scope

	Size() int

	Range(f func(k string, v values.Value))

	SetReturn(values.Value)

	Return() values.Value
}
