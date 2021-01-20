package table

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute/table"
)

// Copy returns a buffered copy of the table and consumes the
// input table. If the input table is already buffered, it "consumes"
// the input and returns the same table.
//
// The buffered table can then be copied additional times using the
// BufferedTable.Copy method.
//
// This method should be used sparingly if at all. It will retain
// each of the buffers of data coming out of a table so the entire
// table is materialized in memory. For large datasets, this could
// potentially cause a problem. The allocator is meant to catch when
// this happens and prevent it.
func Copy(t flux.Table) (flux.BufferedTable, error) {
	return table.Copy(t)
}
