package execute

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/internal/execute/groupkey"
	"github.com/InfluxCommunity/flux/values"
)

func NewGroupKey(cols []flux.ColMeta, values []values.Value) flux.GroupKey {
	return groupkey.New(cols, values)
}
