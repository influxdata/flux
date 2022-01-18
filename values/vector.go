package values

import (
	"github.com/influxdata/flux/semantic"
)

type Vector interface {
	Value
	ElementType() semantic.MonoType
	Get(i int) Value
	Set(i int, value Value)
	Len() int
}
