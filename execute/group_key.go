package execute

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/internal/execute/groupkey"
	"github.com/influxdata/flux/values"
)

func NewGroupKey(cols []flux.ColMeta, values []values.Value) flux.GroupKey {
	return groupkey.New(cols, values)
}
