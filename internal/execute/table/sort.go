package table

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute/table"
)

// Sort will read a TableIterator and produce another TableIterator
// where the keys are sorted.
//
// This method will buffer all of the data since it needs to ensure
// all of the tables are read to avoid any deadlocks. Be careful
// using this method in performance sensitive areas.
func Sort(tables flux.TableIterator) (flux.TableIterator, error) {
	return table.Sort(tables)
}
