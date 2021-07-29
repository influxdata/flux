package table

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/execute/table"
)

func Values(cr flux.ColReader, j int) array.Interface {
	return table.Values(cr, j)
}
