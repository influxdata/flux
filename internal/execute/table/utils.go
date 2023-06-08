package table

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/array"
	"github.com/InfluxCommunity/flux/execute/table"
)

func Values(cr flux.ColReader, j int) array.Array {
	return table.Values(cr, j)
}
