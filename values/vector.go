package values

import (
	arrow "github.com/mvn-trinhnguyen2-dn/flux/array"
	"github.com/mvn-trinhnguyen2-dn/flux/semantic"
)

type Vector interface {
	Value
	ElementType() semantic.MonoType
	Arr() arrow.Array
}
