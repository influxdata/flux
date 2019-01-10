package execute_test

import (
	"fmt"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
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
