package mock

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
)

type TableBuilderCache struct{}

func (tbc *TableBuilderCache) TableBuilder(key flux.GroupKey) (execute.TableBuilder, bool) {
	return nil, true
}

func (tbc *TableBuilderCache) ForEachBuilder(f func(flux.GroupKey, execute.TableBuilder)) {
	return
}
