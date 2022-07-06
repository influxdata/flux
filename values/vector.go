package values

import (
	arrow "github.com/influxdata/flux/array"
	"github.com/influxdata/flux/semantic"
)

type Vector interface {
	Value
	ElementType() semantic.MonoType
	Arr() arrow.Array
	IsRepeat() bool
}
