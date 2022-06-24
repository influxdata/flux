package execute

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/execute/groupkey"
	"github.com/mvn-trinhnguyen2-dn/flux/values"
)

func NewGroupKey(cols []flux.ColMeta, values []values.Value) flux.GroupKey {
	return groupkey.New(cols, values)
}
