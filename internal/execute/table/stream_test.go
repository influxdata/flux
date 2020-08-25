package table_test

import (
	"context"
	"testing"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
)

func TestStream(t *testing.T) {
	executetest.RunTableTests(t,
		executetest.TableTest{
			NewFn: func(ctx context.Context, alloc *memory.Allocator) flux.TableIterator {
				// Only a single buffer.
				key1 := execute.NewGroupKey(
					[]flux.ColMeta{
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					[]values.Value{
						values.NewString("m0"),
						values.NewString("f0"),
					},
				)
				cols1 := append(key1.Cols(),
					flux.ColMeta{Label: "_time", Type: flux.TTime},
					flux.ColMeta{Label: "_value", Type: flux.TFloat},
				)
				tbl1 := MustStreamContext(ctx, key1, cols1, func(ctx context.Context, w *table.StreamWriter) error {
					vs := make([]array.Interface, len(w.Cols()))
					vs[0] = arrow.Repeat(w.Key().Value(0), 3, alloc)
					vs[1] = arrow.Repeat(w.Key().Value(1), 3, alloc)
					vs[2] = arrow.NewInt([]int64{0, 1, 2}, alloc)
					vs[3] = arrow.NewFloat([]float64{4, 8, 7}, alloc)
					return w.Write(vs)
				})
				// Multiple buffers.
				key2 := execute.NewGroupKey(
					[]flux.ColMeta{
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					[]values.Value{
						values.NewString("m1"),
						values.NewString("f0"),
					},
				)
				cols2 := append(key2.Cols(),
					flux.ColMeta{Label: "_time", Type: flux.TTime},
					flux.ColMeta{Label: "_value", Type: flux.TFloat},
				)
				tbl2 := MustStreamContext(ctx, key2, cols2, func(ctx context.Context, w *table.StreamWriter) error {
					vs := make([]array.Interface, len(w.Cols()))
					vs[0] = arrow.Repeat(w.Key().Value(0), 3, alloc)
					vs[1] = arrow.Repeat(w.Key().Value(1), 3, alloc)
					vs[2] = arrow.NewInt([]int64{0, 1, 2}, alloc)
					vs[3] = arrow.NewFloat([]float64{4, 8, 7}, alloc)
					if err := w.Write(vs); err != nil {
						return err
					}

					vs = make([]array.Interface, len(w.Cols()))
					vs[0] = arrow.Repeat(w.Key().Value(0), 5, alloc)
					vs[1] = arrow.Repeat(w.Key().Value(1), 5, alloc)
					vs[2] = arrow.NewInt([]int64{3, 4, 5, 6, 7}, alloc)
					vs[3] = arrow.NewFloat([]float64{2, 9, 4, 6, 2}, alloc)
					return w.Write(vs)
				})
				// Empty table.
				key3 := execute.NewGroupKey(
					[]flux.ColMeta{
						{Label: "_measurement", Type: flux.TString},
						{Label: "_field", Type: flux.TString},
					},
					[]values.Value{
						values.NewString("m2"),
						values.NewString("f0"),
					},
				)
				cols3 := append(key3.Cols(),
					flux.ColMeta{Label: "_time", Type: flux.TTime},
					flux.ColMeta{Label: "_value", Type: flux.TFloat},
				)
				tbl3 := MustStreamContext(ctx, key3, cols3, func(ctx context.Context, w *table.StreamWriter) error {
					vs := make([]array.Interface, len(w.Cols()))
					for i, col := range w.Cols() {
						vs[i] = arrow.NewBuilder(col.Type, alloc).NewArray()
					}
					return w.Write(vs)
				})
				return table.Iterator{tbl1, tbl2, tbl3}
			},
			IsDone: func(tbl flux.Table) bool {
				return tbl.(interface{ IsDone() bool }).IsDone()
			},
		},
	)
}

func MustStreamContext(ctx context.Context, key flux.GroupKey, cols []flux.ColMeta, f func(ctx context.Context, w *table.StreamWriter) error) flux.Table {
	tbl, err := table.StreamWithContext(ctx, key, cols, f)
	if err != nil {
		panic(err)
	}
	return tbl
}
