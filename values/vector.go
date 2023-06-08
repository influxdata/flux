package values

import (
	arrow "github.com/InfluxCommunity/flux/array"
	"github.com/InfluxCommunity/flux/semantic"
)

type Vector interface {
	Value
	ElementType() semantic.MonoType
	Arr() arrow.Array
	IsRepeat() bool
}
