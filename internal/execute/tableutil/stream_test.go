package tableutil_test

import (
	"context"
	"testing"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/internal/execute/tableutil"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
)

type TableIterator []flux.Table

func (t TableIterator) Do(f func(flux.Table) error) error {
	for _, tbl := range t {
		if err := f(tbl); err != nil {
			return err
		}
	}
	return nil
}

func TestStream(t *testing.T) {
	executetest.RunTableTests(t,
		executetest.TableTest{
			NewFn: func(ctx context.Context, alloc *memory.Allocator) flux.TableIterator {
				// Only a single buffer.
				tbl1 := MustStreamContext(ctx, func(ctx context.Context, fn tableutil.SendFunc) error {
					key := execute.NewGroupKey(
						[]flux.ColMeta{
							{Label: "_measurement", Type: flux.TString},
							{Label: "_field", Type: flux.TString},
						},
						[]values.Value{
							values.NewString("m0"),
							values.NewString("f0"),
						},
					)
					cols := append(key.Cols(),
						flux.ColMeta{Label: "_time", Type: flux.TTime},
						flux.ColMeta{Label: "_value", Type: flux.TFloat},
					)
					vs := make([]array.Interface, len(cols))
					vs[0] = arrow.Repeat(key.Value(0), 3, alloc)
					vs[1] = arrow.Repeat(key.Value(1), 3, alloc)
					vs[2] = arrow.NewInt([]int64{0, 1, 2}, alloc)
					vs[3] = arrow.NewFloat([]float64{4, 8, 7}, alloc)
					fn(&arrow.TableBuffer{
						GroupKey: key,
						Columns:  cols,
						Values:   vs,
					})
					return nil
				})
				// Multiple buffers.
				tbl2 := MustStreamContext(ctx, func(ctx context.Context, fn tableutil.SendFunc) error {
					key := execute.NewGroupKey(
						[]flux.ColMeta{
							{Label: "_measurement", Type: flux.TString},
							{Label: "_field", Type: flux.TString},
						},
						[]values.Value{
							values.NewString("m1"),
							values.NewString("f0"),
						},
					)
					cols := append(key.Cols(),
						flux.ColMeta{Label: "_time", Type: flux.TTime},
						flux.ColMeta{Label: "_value", Type: flux.TFloat},
					)
					vs := make([]array.Interface, len(cols))
					vs[0] = arrow.Repeat(key.Value(0), 3, alloc)
					vs[1] = arrow.Repeat(key.Value(1), 3, alloc)
					vs[2] = arrow.NewInt([]int64{0, 1, 2}, alloc)
					vs[3] = arrow.NewFloat([]float64{4, 8, 7}, alloc)
					fn(&arrow.TableBuffer{
						GroupKey: key,
						Columns:  cols,
						Values:   vs,
					})

					vs = make([]array.Interface, len(cols))
					vs[0] = arrow.Repeat(key.Value(0), 5, alloc)
					vs[1] = arrow.Repeat(key.Value(1), 5, alloc)
					vs[2] = arrow.NewInt([]int64{3, 4, 5, 6, 7}, alloc)
					vs[3] = arrow.NewFloat([]float64{2, 9, 4, 6, 2}, alloc)
					fn(&arrow.TableBuffer{
						GroupKey: key,
						Columns:  cols,
						Values:   vs,
					})
					return nil
				})
				// Empty table.
				tbl3 := MustStreamContext(ctx, func(ctx context.Context, fn tableutil.SendFunc) error {
					key := execute.NewGroupKey(
						[]flux.ColMeta{
							{Label: "_measurement", Type: flux.TString},
							{Label: "_field", Type: flux.TString},
						},
						[]values.Value{
							values.NewString("m2"),
							values.NewString("f0"),
						},
					)
					cols := append(key.Cols(),
						flux.ColMeta{Label: "_time", Type: flux.TTime},
						flux.ColMeta{Label: "_value", Type: flux.TFloat},
					)
					vs := make([]array.Interface, len(cols))
					for i, col := range cols {
						vs[i] = arrow.NewBuilder(col.Type, alloc).NewArray()
					}
					fn(&arrow.TableBuffer{
						GroupKey: key,
						Columns:  cols,
						Values:   vs,
					})
					return nil
				})
				return TableIterator(
					[]flux.Table{tbl1, tbl2, tbl3},
				)
			},
			IsDone: func(tbl flux.Table) bool {
				return tbl.(interface{ IsDone() bool }).IsDone()
			},
		},
	)
}

func MustStreamContext(ctx context.Context, f func(ctx context.Context, fn tableutil.SendFunc) error) flux.Table {
	tbl, err := tableutil.StreamWithContext(ctx, f)
	if err != nil {
		panic(err)
	}
	return tbl
}
