package execute_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/gen"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
)

func TestTablesEqual(t *testing.T) {

	testCases := []struct {
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
			return table.Iterator{tbl1, tbl2}
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

func TestColListTableBuilder_AppendValues(t *testing.T) {
	for _, tt := range []struct {
		name   string
		typ    flux.ColType
		values array.Interface
		want   *executetest.Table
	}{
		{
			name: "Ints",
			typ:  flux.TInt,
			values: func() array.Interface {
				b := array.NewInt64Builder(memory.DefaultAllocator)
				b.Append(2)
				b.Append(3)
				b.AppendNull()
				b.Append(8)
				b.AppendNull()
				b.Append(6)
				return b.NewArray()
			}(),
			want: &executetest.Table{
				ColMeta: []flux.ColMeta{{
					Label: "_value",
					Type:  flux.TInt,
				}},
				Data: [][]interface{}{
					{int64(2)}, {int64(3)}, {nil}, {int64(8)}, {nil}, {int64(6)},
				},
			},
		},
		{
			name: "UInts",
			typ:  flux.TUInt,
			values: func() array.Interface {
				b := array.NewUint64Builder(memory.DefaultAllocator)
				b.Append(2)
				b.Append(3)
				b.AppendNull()
				b.Append(8)
				b.AppendNull()
				b.Append(6)
				return b.NewArray()
			}(),
			want: &executetest.Table{
				ColMeta: []flux.ColMeta{{
					Label: "_value",
					Type:  flux.TUInt,
				}},
				Data: [][]interface{}{
					{uint64(2)}, {uint64(3)}, {nil}, {uint64(8)}, {nil}, {uint64(6)},
				},
			},
		},
		{
			name: "Floats",
			typ:  flux.TFloat,
			values: func() array.Interface {
				b := array.NewFloat64Builder(memory.DefaultAllocator)
				b.Append(2)
				b.Append(3)
				b.AppendNull()
				b.Append(8)
				b.AppendNull()
				b.Append(6)
				return b.NewArray()
			}(),
			want: &executetest.Table{
				ColMeta: []flux.ColMeta{{
					Label: "_value",
					Type:  flux.TFloat,
				}},
				Data: [][]interface{}{
					{2.0}, {3.0}, {nil}, {8.0}, {nil}, {6.0},
				},
			},
		},
		{
			name: "Strings",
			typ:  flux.TString,
			values: func() array.Interface {
				b := arrow.NewStringBuilder(&memory.Allocator{})
				b.AppendString("a")
				b.AppendString("d")
				b.AppendNull()
				b.AppendString("b")
				b.AppendNull()
				b.AppendString("e")
				return b.NewArray()
			}(),
			want: &executetest.Table{
				ColMeta: []flux.ColMeta{{
					Label: "_value",
					Type:  flux.TString,
				}},
				Data: [][]interface{}{
					{"a"}, {"d"}, {nil}, {"b"}, {nil}, {"e"},
				},
			},
		},
		{
			name: "Bools",
			typ:  flux.TBool,
			values: func() array.Interface {
				b := array.NewBooleanBuilder(memory.DefaultAllocator)
				b.Append(true)
				b.Append(false)
				b.AppendNull()
				b.Append(false)
				b.AppendNull()
				b.Append(true)
				return b.NewArray()
			}(),
			want: &executetest.Table{
				ColMeta: []flux.ColMeta{{
					Label: "_value",
					Type:  flux.TBool,
				}},
				Data: [][]interface{}{
					{true}, {false}, {nil}, {false}, {nil}, {true},
				},
			},
		},
		{
			name: "Times",
			typ:  flux.TTime,
			values: func() array.Interface {
				b := array.NewInt64Builder(memory.DefaultAllocator)
				b.Append(2)
				b.Append(3)
				b.AppendNull()
				b.Append(8)
				b.AppendNull()
				b.Append(6)
				return b.NewArray()
			}(),
			want: &executetest.Table{
				ColMeta: []flux.ColMeta{{
					Label: "_value",
					Type:  flux.TTime,
				}},
				Data: [][]interface{}{
					{execute.Time(2)}, {execute.Time(3)}, {nil}, {execute.Time(8)}, {nil}, {execute.Time(6)},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			key := execute.NewGroupKey(nil, nil)
			b := execute.NewColListTableBuilder(key, &memory.Allocator{})
			if _, err := b.AddCol(flux.ColMeta{Label: "_value", Type: tt.typ}); err != nil {
				t.Fatal(err)
			}

			switch tt.typ {
			case flux.TBool:
				if err := b.AppendBools(0, tt.values.(*array.Boolean)); err != nil {
					t.Fatal(err)
				}
			case flux.TInt:
				if err := b.AppendInts(0, tt.values.(*array.Int64)); err != nil {
					t.Fatal(err)
				}
			case flux.TUInt:
				if err := b.AppendUInts(0, tt.values.(*array.Uint64)); err != nil {
					t.Fatal(err)
				}
			case flux.TFloat:
				if err := b.AppendFloats(0, tt.values.(*array.Float64)); err != nil {
					t.Fatal(err)
				}
			case flux.TString:
				if err := b.AppendStrings(0, tt.values.(*array.Binary)); err != nil {
					t.Fatal(err)
				}
			case flux.TTime:
				if err := b.AppendTimes(0, tt.values.(*array.Int64)); err != nil {
					t.Fatal(err)
				}
			default:
				execute.PanicUnknownType(tt.typ)
			}

			table, err := b.Table()
			if err != nil {
				t.Fatal(err)
			}
			got, err := executetest.ConvertTable(table)
			if err != nil {
				t.Fatal(err)
			}
			got.Normalize()
			tt.want.Normalize()

			if !cmp.Equal(tt.want, got) {
				t.Fatalf("unexpected output -want/+got:\n%s", cmp.Diff(tt.want, got))
			}
		})
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

func TestEmptyWindowTable(t *testing.T) {
	executetest.RunTableTests(t, executetest.TableTest{
		NewFn: func(ctx context.Context, alloc *memory.Allocator) flux.TableIterator {
			// Prime the allocator with an allocation to avoid
			// an error happening for no allocations.
			// No allocations is expected but the table tests check that
			// an allocation happened.
			alloc.Free(alloc.Allocate(1))

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
			cols := []flux.ColMeta{
				{Label: "_measurement", Type: flux.TString},
				{Label: "_field", Type: flux.TString},
				{Label: "_time", Type: flux.TTime},
				{Label: "_value", Type: flux.TInt},
			}
			return table.Iterator{
				execute.NewEmptyTable(key, cols),
			}
		},
		IsDone: func(tbl flux.Table) bool {
			return true
		},
	})
}
