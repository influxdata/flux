package table

import (
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute/table"
)

const BufferSize = table.BufferSize

type BufferedBuilder = table.BufferedBuilder

func GetBufferedBuilder(key flux.GroupKey, cache *BuilderCache) (builder *BufferedBuilder, created bool) {
	return table.GetBufferedBuilder(key, cache)
}

func NewBufferedBuilder(key flux.GroupKey, mem memory.Allocator) *BufferedBuilder {
	return table.NewBufferedBuilder(key, mem)
}
