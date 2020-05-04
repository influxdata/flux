package execute_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/internal/gen"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
)

func TestTablesEqual(t *testing.T) {

	testCases := []struct {
		skip    bool
		name    string
		data0   *executetest.Table // data from parent 0
		data1   *executetest.Table // data from parent 1
		want    bool
		wantErr bool
	}{
		{
			name: "simple equality",

			data0: &executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
					{execute.Time(3), 3.0},
				},
			},
			data1: &executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
					{execute.Time(3), 3.0},
				},
			},
			want: true,
		},
		{
			name: "left empty",
			data0: &executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{},
			},
			data1: &executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
					{execute.Time(3), 3.0},
				},
			},
			want: false,
		},
		{
			name: "right empty",
			data0: &executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
					{execute.Time(3), 3.0},
				},
			},
			data1: &executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{},
			},
			want: false,
		},
		{
			name: "left short",
			data0: &executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
				},
			},
			data1: &executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
					{execute.Time(3), 3.0},
				},
			},
			want: false,
		},
		{
			name: "right short",
			data0: &executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
					{execute.Time(3), 3.0},
				},
			},
			data1: &executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
				},
			},
			want: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip()
			}

			// this is used to normalize tables
			tc.data0.Key()
			tc.data1.Key()

			equal, err := execute.TablesEqual(tc.data0, tc.data1, executetest.UnlimitedAllocator)

			if tc.wantErr {
				if err == nil {
					t.Fatal(fmt.Errorf("case %s expected an error, got none", tc.name))
				} else {
					return
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if want, got := tc.want, equal; want != got {
				if tc.want {
					t.Errorf("%s: expected equal tables, got false", tc.name)
				} else {
					t.Errorf("%s: expected unequal tables, got true", tc.name)
				}
			}
		})
	}
}

type TableIterator struct {
	Tables []flux.Table
}

func (ti TableIterator) Do(f func(flux.Table) error) error {
	for _, tbl := range ti.Tables {
		if err := f(tbl); err != nil {
			return err
		}
	}
	return nil
}

func TestColListTable(t *testing.T) {
	executetest.RunTableTests(t, executetest.TableTest{
		NewFn: func(ctx context.Context, alloc *memory.Allocator) flux.TableIterator {
			b1 := execute.NewColListTableBuilder(execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "host", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("a"),
				},
			), alloc)
			_, _ = b1.AddCol(flux.ColMeta{Label: "_time", Type: flux.TTime})
			_, _ = b1.AddCol(flux.ColMeta{Label: "host", Type: flux.TString})
			_, _ = b1.AddCol(flux.ColMeta{Label: "_value", Type: flux.TFloat})
			_ = b1.AppendTimes(0, arrow.NewInt(
				[]int64{0, 10, 20, 30, 40, 50},
				nil,
			))
			_ = b1.AppendStrings(1, arrow.NewString(
				[]string{"a", "a", "a", "a", "a", "a"},
				nil,
			))
			_ = b1.AppendFloats(2, arrow.NewFloat(
				[]float64{4, 2, 8, 3, 4, 9},
				nil,
			))
			tbl1, _ := b1.Table()
			b1.Release()

			b2 := execute.NewColListTableBuilder(execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "host", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
				},
			), alloc)
			_, _ = b2.AddCol(flux.ColMeta{Label: "_time", Type: flux.TTime})
			_, _ = b2.AddCol(flux.ColMeta{Label: "host", Type: flux.TString})
			_, _ = b2.AddCol(flux.ColMeta{Label: "_value", Type: flux.TFloat})
			tbl2, _ := b2.Table()
			b2.Release()
			return TableIterator{
				Tables: []flux.Table{tbl1, tbl2},
			}
		},
		IsDone: func(tbl flux.Table) bool {
			return tbl.(*execute.ColListTable).IsDone()
		},
	})
}

func TestColListTable_AppendNil(t *testing.T) {
	key := execute.NewGroupKey(nil, nil)
	tb := execute.NewColListTableBuilder(key, &memory.Allocator{})

	// Add a column for the value.
	idx, _ := tb.AddCol(flux.ColMeta{
		Label: execute.DefaultValueColLabel,
		Type:  flux.TFloat,
	})

	// Add one normal value and add one nil value.
	_ = tb.AppendFloat(idx, 1.0)
	_ = tb.AppendNil(idx)

	// Build the table and then verify the arrow table.
	tbl, err := tb.Table()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := tbl.Do(func(cr flux.ColReader) error {
		vs := cr.Floats(idx)
		if got, want := vs.Len(), 2; got != want {
			t.Errorf("unexpected length -want/+got\n\t- %d\n\t+ %d", want, got)
			return nil
		}

		if vs.IsNull(0) {
			t.Error("first value should not be null")
		}
		if !vs.IsNull(1) {
			t.Error("second value should be null")
		}
		return nil
	}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestColListTable_SetNil(t *testing.T) {
	key := execute.NewGroupKey(nil, nil)
	tb := execute.NewColListTableBuilder(key, &memory.Allocator{})

	// Add a column for the value.
	idx, _ := tb.AddCol(flux.ColMeta{
		Label: execute.DefaultValueColLabel,
		Type:  flux.TFloat,
	})

	// Grow by two values, set the first to 1 and set the second to nil.
	_ = tb.GrowFloats(idx, 2)
	_ = tb.SetValue(0, idx, values.New(1.0))
	_ = tb.SetNil(1, idx)

	// Build the table and then verify the arrow table.
	tbl, err := tb.Table()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := tbl.Do(func(cr flux.ColReader) error {
		vs := cr.Floats(idx)
		if got, want := vs.Len(), 2; got != want {
			t.Errorf("unexpected length -want/+got\n\t- %d\n\t+ %d", want, got)
			return nil
		}

		if vs.IsNull(0) {
			t.Error("first value should not be null")
		}
		if !vs.IsNull(1) {
			t.Error("second value should be null")
		}
		return nil
	}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestCopyTable(t *testing.T) {
	alloc := &memory.Allocator{}

	input, err := gen.Input(context.Background(), gen.Schema{
		Tags: []gen.Tag{
			{Name: "t0", Cardinality: 1},
		},
		NumPoints: 100,
		Period:    time.Hour,
		Types: map[flux.ColType]int{
			flux.TFloat: 1,
		},
		Alloc: alloc,
	})
	if err != nil {
		t.Fatalf("unable to generate tables: %s", err)
	}

	var buffers []flux.BufferedTable
	if err := input.Do(func(table flux.Table) error {
		bt, err := execute.CopyTable(table)
		if err != nil {
			return err
		}
		buffers = append(buffers, bt)
		return nil
	}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Ensure we can copy the table and read a point from the
	// column reader without panicking.
	for _, buf := range buffers {
		cpy := buf.Copy()
		if err := cpy.Do(func(cr flux.ColReader) error {
			if cr.Len() == 0 {
				return nil
			}

			_ = execute.ValueForRow(cr, 0, 0)
			return nil
		}); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	}

	// The memory should not have been freed yet.
	if got := alloc.Allocated(); got == 0 {
		t.Errorf("expected memory to be consumed: got=%d", got)
	}

	// Mark each of the tables as done which should free the
	// remaining memory.
	for _, buf := range buffers {
		buf.Done()
	}

	// Ensure there has been no memory leak.
	if got, want := alloc.Allocated(), int64(0); got != want {
		t.Errorf("memory leak -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
}

func TestCopyTable_Empty(t *testing.T) {
	in := &executetest.Table{
		GroupKey: execute.NewGroupKey(
			[]flux.ColMeta{
				{Label: "t0", Type: flux.TString},
			},
			[]values.Value{
				values.NewString("v0"),
			},
		),
		ColMeta: []flux.ColMeta{
			{Label: "t0", Type: flux.TString},
			{Label: "_value", Type: flux.TFloat},
		},
	}

	cpy, err := execute.CopyTable(in)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	defer cpy.Done()
	if !cpy.Empty() {
		t.Fatal("expected copied table to be empty, but it wasn't")
	}
}
