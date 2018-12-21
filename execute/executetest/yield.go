package executetest

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
)

func NewYieldProcedureSpec(name string) plan.PhysicalProcedureSpec {
	return &YieldProcedureSpec{name: name}
}

const YieldKind = "yield-test"

type YieldProcedureSpec struct {
	plan.DefaultCost
	name string
}

func (YieldProcedureSpec) Kind() plan.ProcedureKind {
	return YieldKind
}

func (y YieldProcedureSpec) Copy() plan.ProcedureSpec {
	return YieldProcedureSpec{name: y.name}
}

func (y YieldProcedureSpec) YieldName() string {
	return y.name
}

// yieldTransformation copies the table as it is.
type yieldTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache
}

func NewYieldTransformation(d execute.Dataset, cache execute.TableBuilderCache) *yieldTransformation {
	return &yieldTransformation{
		d:     d,
		cache: cache,
	}
}

func (t *yieldTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *yieldTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, _ := t.cache.TableBuilder(tbl.Key())
	err := execute.AddTableCols(tbl, builder)
	if err != nil {
		return err
	}
	return execute.AppendTable(tbl, builder)
}

func (t *yieldTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *yieldTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *yieldTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
