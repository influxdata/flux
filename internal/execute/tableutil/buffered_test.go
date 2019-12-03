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

func TestBufferedTable(t *testing.T) {
	executetest.RunTableTests(t,
		executetest.TableTest{
			NewFn: func(ctx context.Context, alloc *memory.Allocator) flux.TableIterator {
				// Only a single buffer.
				tbl1 := tableutil.FromBuffer(
					func() flux.ColReader {
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
						return &arrow.TableBuffer{
							GroupKey: key,
							Columns:  cols,
							Values:   vs,
						}
					}(),
				)
				// Multiple buffers.
				tbl2 := func() flux.Table {
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

					table := &tableutil.BufferedTable{
						GroupKey: key,
						Columns:  cols,
					}
					vs := make([]array.Interface, len(cols))
					vs[0] = arrow.Repeat(key.Value(0), 3, alloc)
					vs[1] = arrow.Repeat(key.Value(1), 3, alloc)
					vs[2] = arrow.NewInt([]int64{0, 1, 2}, alloc)
					vs[3] = arrow.NewFloat([]float64{4, 8, 7}, alloc)
					table.Buffers = append(table.Buffers, &arrow.TableBuffer{
						GroupKey: key,
						Columns:  cols,
						Values:   vs,
					})

					vs = make([]array.Interface, len(cols))
					vs[0] = arrow.Repeat(key.Value(0), 5, alloc)
					vs[1] = arrow.Repeat(key.Value(1), 5, alloc)
					vs[2] = arrow.NewInt([]int64{3, 4, 5, 6, 7}, alloc)
					vs[3] = arrow.NewFloat([]float64{2, 9, 4, 6, 2}, alloc)
					table.Buffers = append(table.Buffers, &arrow.TableBuffer{
						GroupKey: key,
						Columns:  cols,
						Values:   vs,
					})
					return table
				}()
				// Empty table.
				tbl3 := func() flux.Table {
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
					return &tableutil.BufferedTable{
						GroupKey: key,
						Columns:  cols,
					}
				}()
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
