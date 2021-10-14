package execute_test

import (
	"context"
	"sort"
	"testing"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/execute/table/static"
	"github.com/influxdata/flux/mock"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestAggregateTransformation_ProcessChunk(t *testing.T) {
	// Ensure we allocate and free all memory correctly.
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	// Generate one table chunk using static.Table.
	// This will only produce one column reader, so we are
	// extracting that value from the nested iterators.
	gen := static.Table{
		static.Times("_time", 0, 10, 20),
		static.Floats("_value", 1, 2, 3),
	}

	isAggregated, shouldHaveState := false, false
	tr, _, err := execute.NewAggregateTransformation(
		executetest.RandomDatasetID(),
		&mock.AggregateTransformation{
			AggregateFn: func(chunk table.Chunk, state interface{}, _ memory.Allocator) (interface{}, bool, error) {
				// Memory should be allocated and should not have been improperly freed.
				// This accounts for 64 bytes (data) + 64 bytes (null bitmap) for each column
				// of which there are two. 64 bytes is the minimum that arrow will allocate
				// for a particular data buffer.
				mem.AssertSize(t, 256)

				if shouldHaveState {
					if state == nil {
						t.Error("should have state, but state was nil")
					} else if want, got := "mystate", state.(string); want != got {
						t.Errorf("unexpected state -want/+got:\n\t- %s\n\t+ %s", want, got)
					}
				} else {
					if state != nil {
						t.Error("should not have state, but state was not nil")
					}
				}
				isAggregated = true
				return "mystate", true, nil
			},
			ComputeFn: func(key flux.GroupKey, state interface{}, d *execute.TransportDataset, mem memory.Allocator) error {
				t.Error("did not expect to call compute")
				return nil
			},
		},
		mem,
	)
	if err != nil {
		t.Fatal(err)
	}

	// We can use a TransportDataset as a mock source
	// to send messages to the transformation we are testing.
	source := execute.NewTransportDataset(executetest.RandomDatasetID(), mem)
	source.AddTransformation(tr)

	tbl := gen.Table(mem)
	if err := tbl.Do(func(cr flux.ColReader) error {
		chunk := table.ChunkFromReader(cr)
		chunk.Retain()
		return source.Process(chunk)
	}); err != nil {
		t.Fatal(err)
	}

	if !isAggregated {
		t.Fatal("expected aggregate function to be invoked, but it was not")
	}
	// Memory should have been released since we did not retain the data.
	mem.AssertSize(t, 0)

	isAggregated, shouldHaveState = false, true
	tbl = gen.Table(mem)
	if err := tbl.Do(func(cr flux.ColReader) error {
		chunk := table.ChunkFromReader(cr)
		chunk.Retain()
		return source.Process(chunk)
	}); err != nil {
		t.Fatal(err)
	}

	if !isAggregated {
		t.Fatal("expected aggregate function to be invoked, but it was not")
	}
}

func TestAggregateTransformation_FlushKey(t *testing.T) {
	// Ensure we allocate and free all memory correctly.
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	var disposeCount int
	isComputed := false
	tr, _, err := execute.NewAggregateTransformation(
		executetest.RandomDatasetID(),
		&mock.AggregateTransformation{
			AggregateFn: func(chunk table.Chunk, state interface{}, _ memory.Allocator) (interface{}, bool, error) {
				return &mockState{
					value:        "mystate",
					disposeCount: &disposeCount,
				}, true, nil
			},
			ComputeFn: func(key flux.GroupKey, state interface{}, d *execute.TransportDataset, mem memory.Allocator) error {
				if state == nil {
					t.Error("invoked compute without state")
				} else if want, got := "mystate", state.(*mockState).value; want != got {
					t.Errorf("unexpected state -want/+got:\n\t- %s\n\t+ %s", want, got)
				}
				isComputed = true
				return nil
			},
		},
		mem,
	)
	if err != nil {
		t.Fatal(err)
	}

	// We can use a TransportDataset as a mock source
	// to send messages to the transformation we are testing.
	source := execute.NewTransportDataset(executetest.RandomDatasetID(), mem)
	source.AddTransformation(tr)

	// Generate one table chunk using static.Table.
	// This will only produce one column reader, so we are
	// extracting that value from the nested iterators.
	gen := static.Table{
		static.Times("_time", 0, 10, 20),
		static.Floats("_value", 1, 2, 3),
	}

	tbl := gen.Table(mem)
	if err := source.FlushKey(tbl.Key()); err != nil {
		t.Fatal(err)
	} else if isComputed {
		t.Fatal("did not expect compute to be called")
	}

	// Now process that table and attempt to send flush key again.
	// This time, it should work.
	if err := tbl.Do(func(cr flux.ColReader) error {
		chunk := table.ChunkFromReader(cr)
		chunk.Retain()
		return source.Process(chunk)
	}); err != nil {
		t.Fatal(err)
	}

	if err := source.FlushKey(tbl.Key()); err != nil {
		t.Fatal(err)
	} else if !isComputed {
		t.Fatal("expected compute to be called")
	}

	// The state should have been disposed.
	if want, got := 1, disposeCount; want != got {
		t.Errorf("unexpected dispose count -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
}

func TestAggregateTransformation_Process(t *testing.T) {
	// Ensure we allocate and free all memory correctly.
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	// Generate one table chunk using static.Table.
	// This will only produce one column reader, so we are
	// extracting that value from the nested iterators.
	gen := static.Table{
		static.Times("_time", 0, 10, 20),
		static.Floats("_value", 1, 2, 3),
	}

	isAggregated, isComputed := false, false
	tr, _, err := execute.NewAggregateTransformation(
		executetest.RandomDatasetID(),
		&mock.AggregateTransformation{
			AggregateFn: func(chunk table.Chunk, state interface{}, _ memory.Allocator) (interface{}, bool, error) {
				// Memory should be allocated and should not have been improperly freed.
				// This accounts for 64 bytes (data) + 64 bytes (null bitmap) for each column
				// of which there are two. 64 bytes is the minimum that arrow will allocate
				// for a particular data buffer.
				mem.AssertSize(t, 256)
				isAggregated = true
				return "mystate", true, nil
			},
			ComputeFn: func(key flux.GroupKey, state interface{}, d *execute.TransportDataset, mem memory.Allocator) error {
				if state == nil {
					t.Error("invoked compute without state")
				} else if want, got := "mystate", state.(string); want != got {
					t.Errorf("unexpected state -want/+got:\n\t- %s\n\t+ %s", want, got)
				}
				isComputed = true
				return nil
			},
		},
		mem,
	)
	if err != nil {
		t.Fatal(err)
	}

	tbl := gen.Table(mem)
	if err := tr.Process(executetest.RandomDatasetID(), tbl); err != nil {
		t.Fatal(err)
	}

	if !isAggregated {
		t.Fatal("expected aggregate function to be invoked, but it was not")
	}
	if !isComputed {
		t.Fatal("expected compute function to be invoked, but it was not")
	}
}

func TestAggregateTransformation_ProcessEmpty(t *testing.T) {
	// Ensure we allocate and free all memory correctly.
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	// Create an empty table. The table should still send at least
	// one chunk and the chunk should be empty.
	tbl := &table.BufferedTable{
		GroupKey: execute.NewGroupKey(nil, nil),
		Columns: []flux.ColMeta{
			{Label: "_time", Type: flux.TTime},
			{Label: "_value", Type: flux.TFloat},
		},
	}

	isAggregated, isComputed := false, false
	tr, _, err := execute.NewAggregateTransformation(
		executetest.RandomDatasetID(),
		&mock.AggregateTransformation{
			AggregateFn: func(chunk table.Chunk, state interface{}, _ memory.Allocator) (interface{}, bool, error) {
				// Empty table chunks use no memory.
				mem.AssertSize(t, 0)
				if chunk.Len() > 0 {
					t.Errorf("table was not empty, is %d", chunk.Len())
				}
				isAggregated = true
				return "mystate", true, nil
			},
			ComputeFn: func(key flux.GroupKey, state interface{}, d *execute.TransportDataset, mem memory.Allocator) error {
				if state == nil {
					t.Error("invoked compute without state")
				} else if want, got := "mystate", state.(string); want != got {
					t.Errorf("unexpected state -want/+got:\n\t- %s\n\t+ %s", want, got)
				}
				isComputed = true
				return nil
			},
		},
		mem,
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := tr.Process(executetest.RandomDatasetID(), tbl); err != nil {
		t.Fatal(err)
	}

	if !isAggregated {
		t.Fatal("expected aggregate function to be invoked, but it was not")
	}
	if !isComputed {
		t.Fatal("expected compute function to be invoked, but it was not")
	}
}

func TestAggregateTransformation_Finish(t *testing.T) {
	// Ensure we allocate and free all memory correctly.
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	var (
		disposeCount int
		isDisposed   bool
	)
	isAggregated, isComputed := false, false
	tr, _, err := execute.NewAggregateTransformation(
		executetest.RandomDatasetID(),
		&mock.AggregateTransformation{
			AggregateFn: func(chunk table.Chunk, state interface{}, _ memory.Allocator) (interface{}, bool, error) {
				isAggregated = true
				return &mockState{
					value:        "mystate",
					disposeCount: &disposeCount,
				}, true, nil
			},
			ComputeFn: func(key flux.GroupKey, state interface{}, d *execute.TransportDataset, mem memory.Allocator) error {
				if state == nil {
					t.Error("invoked compute without state")
				} else if want, got := "mystate", state.(*mockState).value; want != got {
					t.Errorf("unexpected state -want/+got:\n\t- %s\n\t+ %s", want, got)
				}
				isComputed = true
				return nil
			},
			DisposeFn: func() {
				isDisposed = true
			},
		},
		mem,
	)
	if err != nil {
		t.Fatal(err)
	}

	// We can use a TransportDataset as a mock source
	// to send messages to the transformation we are testing.
	source := execute.NewTransportDataset(executetest.RandomDatasetID(), mem)
	source.AddTransformation(tr)

	// Generate one table chunk using static.Table.
	// This will only produce one column reader, so we are
	// extracting that value from the nested iterators.
	gen := static.Table{
		static.Times("_time", 0, 10, 20),
		static.Floats("_value", 1, 2, 3),
	}

	// Process the table but do not flush the key.
	tbl := gen.Table(mem)
	if err := tbl.Do(func(cr flux.ColReader) error {
		chunk := table.ChunkFromReader(cr)
		chunk.Retain()
		return source.Process(chunk)
	}); err != nil {
		t.Fatal(err)
	}

	if !isAggregated {
		t.Fatal("expected aggregate function to be called")
	} else if isComputed {
		t.Fatal("did not expect compute function to be called")
	}

	source.Finish(nil)
	if !isComputed {
		t.Fatal("expected compute function to be called")
	}

	// The state should have been disposed.
	if want, got := 1, disposeCount; want != got {
		t.Errorf("unexpected dispose count -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// So should the transformation.
	if !isDisposed {
		t.Error("transformation was not disposed")
	}
}

func TestSimpleAggregate_Process(t *testing.T) {
	sumAgg := new(universe.SumAgg)
	countAgg := new(universe.CountAgg)
	testCases := []struct {
		name   string
		agg    execute.SimpleAggregate
		config execute.SimpleAggregateConfig
		data   []*executetest.Table
		want   []*executetest.Table
	}{
		{
			name:   "single",
			config: execute.DefaultSimpleAggregateConfig,
			agg:    sumAgg,
			data: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(100), execute.Time(0), 0.0},
					{execute.Time(0), execute.Time(100), execute.Time(10), 1.0},
					{execute.Time(0), execute.Time(100), execute.Time(20), 2.0},
					{execute.Time(0), execute.Time(100), execute.Time(30), 3.0},
					{execute.Time(0), execute.Time(100), execute.Time(40), 4.0},
					{execute.Time(0), execute.Time(100), execute.Time(50), 5.0},
					{execute.Time(0), execute.Time(100), execute.Time(60), 6.0},
					{execute.Time(0), execute.Time(100), execute.Time(70), 7.0},
					{execute.Time(0), execute.Time(100), execute.Time(80), 8.0},
					{execute.Time(0), execute.Time(100), execute.Time(90), 9.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(100), 45.0},
				},
			}},
		},
		{
			name: "single use start time",
			config: execute.SimpleAggregateConfig{
				Columns: []string{execute.DefaultValueColLabel},
			},
			agg: sumAgg,
			data: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(100), execute.Time(0), 0.0},
					{execute.Time(0), execute.Time(100), execute.Time(10), 1.0},
					{execute.Time(0), execute.Time(100), execute.Time(20), 2.0},
					{execute.Time(0), execute.Time(100), execute.Time(30), 3.0},
					{execute.Time(0), execute.Time(100), execute.Time(40), 4.0},
					{execute.Time(0), execute.Time(100), execute.Time(50), 5.0},
					{execute.Time(0), execute.Time(100), execute.Time(60), 6.0},
					{execute.Time(0), execute.Time(100), execute.Time(70), 7.0},
					{execute.Time(0), execute.Time(100), execute.Time(80), 8.0},
					{execute.Time(0), execute.Time(100), execute.Time(90), 9.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(100), 45.0},
				},
			}},
		},
		{
			name:   "multiple tables",
			config: execute.DefaultSimpleAggregateConfig,
			agg:    sumAgg,
			data: []*executetest.Table{
				{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), execute.Time(100), execute.Time(0), 0.0},
						{execute.Time(0), execute.Time(100), execute.Time(10), 1.0},
						{execute.Time(0), execute.Time(100), execute.Time(20), 2.0},
						{execute.Time(0), execute.Time(100), execute.Time(30), 3.0},
						{execute.Time(0), execute.Time(100), execute.Time(40), 4.0},
						{execute.Time(0), execute.Time(100), execute.Time(50), 5.0},
						{execute.Time(0), execute.Time(100), execute.Time(60), 6.0},
						{execute.Time(0), execute.Time(100), execute.Time(70), 7.0},
						{execute.Time(0), execute.Time(100), execute.Time(80), 8.0},
						{execute.Time(0), execute.Time(100), execute.Time(90), 9.0},
					},
				},
				{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(100), execute.Time(200), execute.Time(100), 10.0},
						{execute.Time(100), execute.Time(200), execute.Time(110), 11.0},
						{execute.Time(100), execute.Time(200), execute.Time(120), 12.0},
						{execute.Time(100), execute.Time(200), execute.Time(130), 13.0},
						{execute.Time(100), execute.Time(200), execute.Time(140), 14.0},
						{execute.Time(100), execute.Time(200), execute.Time(150), 15.0},
						{execute.Time(100), execute.Time(200), execute.Time(160), 16.0},
						{execute.Time(100), execute.Time(200), execute.Time(170), 17.0},
						{execute.Time(100), execute.Time(200), execute.Time(180), 18.0},
						{execute.Time(100), execute.Time(200), execute.Time(190), 19.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), execute.Time(100), 45.0},
					},
				},
				{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(100), execute.Time(200), 145.0},
					},
				},
			},
		},
		{
			name:   "empty table",
			config: execute.DefaultSimpleAggregateConfig,
			agg:    sumAgg,
			data: []*executetest.Table{
				{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					KeyValues: []interface{}{
						execute.Time(100),
						execute.Time(200),
					},
					Data: [][]interface{}{},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(100), execute.Time(200), nil},
					},
				},
			},
		},
		{
			name:   "table count all null",
			config: execute.DefaultSimpleAggregateConfig,
			agg:    countAgg,
			data: []*executetest.Table{
				{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					KeyValues: []interface{}{
						execute.Time(100),
						execute.Time(200),
					},
					Data: [][]interface{}{
						{execute.Time(100), execute.Time(200), execute.Time(70), nil},
						{execute.Time(100), execute.Time(200), execute.Time(80), nil},
						{execute.Time(100), execute.Time(200), execute.Time(90), nil},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(100), execute.Time(200), int64(3)},
					},
				},
			},
		},
		{
			name:   "table sum all null",
			config: execute.DefaultSimpleAggregateConfig,
			agg:    sumAgg,
			data: []*executetest.Table{
				{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					KeyValues: []interface{}{
						execute.Time(100),
						execute.Time(200),
					},
					Data: [][]interface{}{
						{execute.Time(100), execute.Time(200), execute.Time(70), nil},
						{execute.Time(100), execute.Time(200), execute.Time(80), nil},
						{execute.Time(100), execute.Time(200), execute.Time(90), nil},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(100), execute.Time(200), nil},
					},
				},
			},
		},
		{
			name:   "table some null",
			config: execute.DefaultSimpleAggregateConfig,
			agg:    sumAgg,
			data: []*executetest.Table{
				{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					KeyValues: []interface{}{
						execute.Time(100),
						execute.Time(200),
					},
					Data: [][]interface{}{
						{execute.Time(100), execute.Time(200), execute.Time(70), 10.0},
						{execute.Time(100), execute.Time(200), execute.Time(80), 20.0},
						{execute.Time(100), execute.Time(200), execute.Time(90), nil},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(100), execute.Time(200), 30.0},
					},
				},
			},
		},
		{
			name:   "multiple tables with keyed columns",
			config: execute.DefaultSimpleAggregateConfig,
			agg:    sumAgg,
			data: []*executetest.Table{
				{
					KeyCols: []string{"_start", "_stop", "t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "t1", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), execute.Time(100), "a", execute.Time(0), 0.0},
						{execute.Time(0), execute.Time(100), "a", execute.Time(10), 1.0},
						{execute.Time(0), execute.Time(100), "a", execute.Time(20), 2.0},
						{execute.Time(0), execute.Time(100), "a", execute.Time(30), 3.0},
						{execute.Time(0), execute.Time(100), "a", execute.Time(40), 4.0},
						{execute.Time(0), execute.Time(100), "a", execute.Time(50), 5.0},
						{execute.Time(0), execute.Time(100), "a", execute.Time(60), 6.0},
						{execute.Time(0), execute.Time(100), "a", execute.Time(70), 7.0},
						{execute.Time(0), execute.Time(100), "a", execute.Time(80), 8.0},
						{execute.Time(0), execute.Time(100), "a", execute.Time(90), 9.0},
					},
				},
				{
					KeyCols: []string{"_start", "_stop", "t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "t1", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), execute.Time(100), "b", execute.Time(0), 0.3},
						{execute.Time(0), execute.Time(100), "b", execute.Time(10), 1.3},
						{execute.Time(0), execute.Time(100), "b", execute.Time(20), 2.3},
						{execute.Time(0), execute.Time(100), "b", execute.Time(30), 3.3},
						{execute.Time(0), execute.Time(100), "b", execute.Time(40), 4.3},
						{execute.Time(0), execute.Time(100), "b", execute.Time(50), 5.3},
						{execute.Time(0), execute.Time(100), "b", execute.Time(60), 6.3},
						{execute.Time(0), execute.Time(100), "b", execute.Time(70), 7.3},
						{execute.Time(0), execute.Time(100), "b", execute.Time(80), 8.3},
						{execute.Time(0), execute.Time(100), "b", execute.Time(90), 9.3},
					},
				},
				{
					KeyCols: []string{"_start", "_stop", "t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "t1", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(100), execute.Time(200), "a", execute.Time(100), 10.0},
						{execute.Time(100), execute.Time(200), "a", execute.Time(110), 11.0},
						{execute.Time(100), execute.Time(200), "a", execute.Time(120), 12.0},
						{execute.Time(100), execute.Time(200), "a", execute.Time(130), 13.0},
						{execute.Time(100), execute.Time(200), "a", execute.Time(140), 14.0},
						{execute.Time(100), execute.Time(200), "a", execute.Time(150), 15.0},
						{execute.Time(100), execute.Time(200), "a", execute.Time(160), 16.0},
						{execute.Time(100), execute.Time(200), "a", execute.Time(170), 17.0},
						{execute.Time(100), execute.Time(200), "a", execute.Time(180), 18.0},
						{execute.Time(100), execute.Time(200), "a", execute.Time(190), 19.0},
					},
				},
				{
					KeyCols: []string{"_start", "_stop", "t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "t1", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(100), execute.Time(200), "b", execute.Time(100), 10.3},
						{execute.Time(100), execute.Time(200), "b", execute.Time(110), 11.3},
						{execute.Time(100), execute.Time(200), "b", execute.Time(120), 12.3},
						{execute.Time(100), execute.Time(200), "b", execute.Time(130), 13.3},
						{execute.Time(100), execute.Time(200), "b", execute.Time(140), 14.3},
						{execute.Time(100), execute.Time(200), "b", execute.Time(150), 15.3},
						{execute.Time(100), execute.Time(200), "b", execute.Time(160), 16.3},
						{execute.Time(100), execute.Time(200), "b", execute.Time(170), 17.3},
						{execute.Time(100), execute.Time(200), "b", execute.Time(180), 18.3},
						{execute.Time(100), execute.Time(200), "b", execute.Time(190), 19.3},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_start", "_stop", "t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "t1", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), execute.Time(100), "a", 45.0},
					},
				},
				{
					KeyCols: []string{"_start", "_stop", "t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "t1", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(100), execute.Time(200), "a", 145.0},
					},
				},
				{
					KeyCols: []string{"_start", "_stop", "t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "t1", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), execute.Time(100), "b", 48.0},
					},
				},
				{
					KeyCols: []string{"_start", "_stop", "t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "t1", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(100), execute.Time(200), "b", 148.0},
					},
				},
			},
		},
		{
			name: "multiple values",
			config: execute.SimpleAggregateConfig{
				Columns: []string{"x", "y"},
			},
			agg: sumAgg,
			data: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(100), execute.Time(0), 0.0, 0.0},
					{execute.Time(0), execute.Time(100), execute.Time(10), 1.0, -1.0},
					{execute.Time(0), execute.Time(100), execute.Time(20), 2.0, -2.0},
					{execute.Time(0), execute.Time(100), execute.Time(30), 3.0, -3.0},
					{execute.Time(0), execute.Time(100), execute.Time(40), 4.0, -4.0},
					{execute.Time(0), execute.Time(100), execute.Time(50), 5.0, -5.0},
					{execute.Time(0), execute.Time(100), execute.Time(60), 6.0, -6.0},
					{execute.Time(0), execute.Time(100), execute.Time(70), 7.0, -7.0},
					{execute.Time(0), execute.Time(100), execute.Time(80), 8.0, -8.0},
					{execute.Time(0), execute.Time(100), execute.Time(90), 9.0, -9.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(100), 45.0, -45.0},
				},
			}},
		},
		{
			name: "multiple values changing types",
			config: execute.SimpleAggregateConfig{
				Columns: []string{"x", "y"},
			},
			agg: countAgg,
			data: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(100), execute.Time(0), 0.0, 0.0},
					{execute.Time(0), execute.Time(100), execute.Time(10), 1.0, -1.0},
					{execute.Time(0), execute.Time(100), execute.Time(20), 2.0, -2.0},
					{execute.Time(0), execute.Time(100), execute.Time(30), 3.0, -3.0},
					{execute.Time(0), execute.Time(100), execute.Time(40), 4.0, -4.0},
					{execute.Time(0), execute.Time(100), execute.Time(50), 5.0, -5.0},
					{execute.Time(0), execute.Time(100), execute.Time(60), 6.0, -6.0},
					{execute.Time(0), execute.Time(100), execute.Time(70), 7.0, -7.0},
					{execute.Time(0), execute.Time(100), execute.Time(80), 8.0, -8.0},
					{execute.Time(0), execute.Time(100), execute.Time(90), 9.0, -9.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "x", Type: flux.TInt},
					{Label: "y", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(100), int64(10), int64(10)},
				},
			}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := executetest.NewTestExecuteDependencies().Inject(context.Background())
			agg, d, err := execute.NewSimpleAggregateTransformation(ctx, executetest.RandomDatasetID(), tc.agg, tc.config, memory.DefaultAllocator)
			if err != nil {
				t.Fatal(err)
			}

			store := executetest.NewDataStore()
			d.AddTransformation(store)
			d.SetTriggerSpec(plan.DefaultTriggerSpec)

			parentID := executetest.RandomDatasetID()
			for _, b := range tc.data {
				if err := agg.Process(parentID, b); err != nil {
					t.Fatal(err)
				}
			}
			agg.Finish(parentID, nil)

			got, err := executetest.TablesFromCache(store)
			if err != nil {
				t.Fatal(err)
			}

			executetest.NormalizeTables(got)
			executetest.NormalizeTables(tc.want)

			sort.Sort(executetest.SortedTables(got))
			sort.Sort(executetest.SortedTables(tc.want))

			if !cmp.Equal(tc.want, got, cmpopts.EquateNaNs()) {
				t.Errorf("unexpected tables -want/+got\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}

type mockState struct {
	value        string
	disposeCount *int
}

func (s *mockState) Dispose() {
	if s.disposeCount != nil {
		*s.disposeCount++
	}
}
