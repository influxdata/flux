package table

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute/table"
)

// Diff will perform a diff between two table iterators.
// This will sort the tables within the table iterators and produce
// a diff of the full output.
func Diff(want, got flux.TableIterator, opts ...DiffOption) string {
	return table.Diff(want, got, opts...)
}

type DiffOption = table.DiffOption

func DiffContext(n int) DiffOption {
	return table.DiffContext(n)
}
