package table

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute/table"
)

// Diff will perform a diff between two tables.
// If the tables are the same, the output will be an empty string.
// This will produce a fatal error if there was any problem reading
// either table.
func Diff(want, got flux.Table, opts ...DiffOption) string {
	return table.Diff(want, got, opts...)
}

// DiffIterator will perform a diff between two table iterators.
// This will sort the tables within the table iterators and produce
// a diff of the full output.
func DiffIterator(want, got flux.TableIterator, opts ...DiffOption) string {
	return table.DiffIterator(want, got, opts...)
}

type DiffOption = table.DiffOption

func DiffContext(n int) DiffOption {
	return table.DiffContext(n)
}
