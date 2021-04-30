package table

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute/table"
)

func Values(cr flux.ColReader, j int) array.Interface {
	return table.Values(cr, j)
}
