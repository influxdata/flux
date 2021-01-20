package table

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute/table"
)

// View is a view of a Table.
// The view is divided into a set of rows with a common
// set of columns known as the group key.
// The view does not provide a full view of the entire group key
// and a Table is not guaranteed to have rows ordered by the group key.
type View = table.View

// ViewFromBuffer will create a View from the TableBuffer.
func ViewFromBuffer(buf arrow.TableBuffer) View {
	return table.ViewFromBuffer(buf)
}

// ViewFromReader will create a View from the ColReader.
func ViewFromReader(cr flux.ColReader) View {
	return table.ViewFromReader(cr)
}
