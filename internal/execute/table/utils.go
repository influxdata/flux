package table

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/array"
	"github.com/mvn-trinhnguyen2-dn/flux/execute/table"
)

func Values(cr flux.ColReader, j int) array.Array {
	return table.Values(cr, j)
}
