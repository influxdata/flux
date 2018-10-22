package tablebuilder

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/values"
)

// table is an implementation of flux.Table that is constructed
// by the Instance.
// TODO(jsternberg): The ColReader interface needs to be updated to
// use the array interface types. At the moment, we use staticarray
// values as a compatibility layer so we will have to cast away from
// that.
type table struct {
	key     flux.GroupKey
	colMeta []flux.ColMeta
	columns []array.Base
	sz      int
}

func (tbl *table) Key() flux.GroupKey {
	return tbl.key
}

func (tbl *table) Cols() []flux.ColMeta {
	return tbl.colMeta
}

func (tbl *table) Do(f func(flux.ColReader) error) error {
	return f(tbl)
}

func (tbl *table) RefCount(n int) {
	// TODO(jsternberg): Does this need to be implemented?
}

func (tbl *table) Empty() bool {
	return tbl.sz == 0
}

func (tbl *table) Len() int {
	return tbl.sz
}

func (tbl *table) Bools(j int) []bool {
	column := tbl.columns[j]
	return column.(interface {
		BoolValues() []bool
	}).BoolValues()
}

func (tbl *table) Ints(j int) []int64 {
	column := tbl.columns[j]
	return column.(array.Int).Int64Values()
}

func (tbl *table) UInts(j int) []uint64 {
	column := tbl.columns[j]
	return column.(array.UInt).Uint64Values()
}

func (tbl *table) Floats(j int) []float64 {
	column := tbl.columns[j]
	return column.(array.Float).Float64Values()
}

func (tbl *table) Strings(j int) []string {
	column := tbl.columns[j]
	return column.(interface {
		StringValues() []string
	}).StringValues()
}

func (tbl *table) Times(j int) []values.Time {
	column := tbl.columns[j]
	return column.(interface {
		TimeValues() []values.Time
	}).TimeValues()
}
