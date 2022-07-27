package universe_test

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/gen"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
)

func TestLimitOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"limit","kind":"limit","spec":{"n":10}}`)
	op := &flux.Operation{
		ID: "limit",
		Spec: &universe.LimitOpSpec{
			N: 10,
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestLimit_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.LimitProcedureSpec
		data func() []flux.Table
		want []*executetest.Table
	}{
		{
			name: "empty table",
			spec: &universe.LimitProcedureSpec{
				N: 1,
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: nil,
			}},
		},
		{
			name: "one table",
			spec: &universe.LimitProcedureSpec{
				N: 1,
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0},
						{execute.Time(2), 1.0},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
				},
			}},
		},
		{
			name: "one table n=0",
			spec: &universe.LimitProcedureSpec{
				N: 0,
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0},
						{execute.Time(2), 1.0},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: nil,
			}},
		},
		{
			name: "with null",
			spec: &universe.LimitProcedureSpec{
				N:      2,
				Offset: 1,
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), nil},
						{execute.Time(2), nil},
						{execute.Time(2), 1.0},
						{execute.Time(2), 1.0},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil},
					{execute.Time(2), 1.0},
				},
			}},
		},
		{
			name: "one table with offset single batch",
			spec: &universe.LimitProcedureSpec{
				N:      1,
				Offset: 1,
			},
			data: func() []flux.Table {
				return []flux.Table{executetest.MustCopyTable(&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0},
						{execute.Time(2), 1.0},
						{execute.Time(3), 0.0},
					},
				})}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 1.0},
				},
			}},
		},
		{
			name: "one table with offset multiple batches",
			spec: &universe.LimitProcedureSpec{
				N:      1,
				Offset: 1,
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0},
						{execute.Time(2), 1.0},
						{execute.Time(3), 0.0},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 1.0},
				},
			}},
		},
		{
			name: "multiple tables",
			spec: &universe.LimitProcedureSpec{
				N: 2,
			},
			data: func() []flux.Table {
				return []flux.Table{
					&executetest.Table{
						KeyCols: []string{"t1"},
						ColMeta: []flux.ColMeta{
							{Label: "t1", Type: flux.TString},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{"a", execute.Time(1), 3.0},
							{"a", execute.Time(2), 2.0},
							{"a", execute.Time(2), 1.0},
						},
					},
					&executetest.Table{
						KeyCols: []string{"t1"},
						ColMeta: []flux.ColMeta{
							{Label: "t1", Type: flux.TString},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{"b", execute.Time(3), 3.0},
							{"b", execute.Time(3), 2.0},
							{"b", execute.Time(4), 1.0},
						},
					},
				}
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "t1", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"a", execute.Time(1), 3.0},
						{"a", execute.Time(2), 2.0},
					},
				},
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "t1", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"b", execute.Time(3), 3.0},
						{"b", execute.Time(3), 2.0},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper2(
				t,
				tc.data(),
				tc.want,
				nil,
				func(id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset) {
					tr, ds, err := universe.NewLimitTransformation(tc.spec, id, alloc)
					if err != nil {
						t.Fatal(err)
					}
					return tr, ds
				},
			)
		})
	}
}

func TestProcess_Limit_MultiBuffer(t *testing.T) {
	key := execute.NewGroupKey(nil, nil)
	mem := &memory.ResourceAllocator{}
	b := table.NewBufferedBuilder(key, mem)
	{
		buf := arrow.TableBuffer{
			GroupKey: key,
			Columns: []flux.ColMeta{
				{Label: "_time", Type: flux.TTime},
				{Label: "_value", Type: flux.TInt},
			},
			Values: make([]array.Array, 2),
		}

		times := array.NewIntBuilder(mem)
		for ts := int64(0); ts < 40; ts += 10 {
			times.Append(ts)
		}
		buf.Values[0] = times.NewArray()

		values := array.NewIntBuilder(mem)
		for v := int64(0); v < 4; v++ {
			values.Append(v)
		}
		buf.Values[1] = values.NewArray()
		if err := b.AppendBuffer(&buf); err != nil {
			t.Fatal(err)
		}
	}

	{
		buf := arrow.TableBuffer{
			GroupKey: key,
			Columns: []flux.ColMeta{
				{Label: "_time", Type: flux.TTime},
				{Label: "_value", Type: flux.TInt},
			},
			Values: make([]array.Array, 2),
		}

		times := array.NewIntBuilder(mem)
		for ts := int64(40); ts < 80; ts += 10 {
			times.Append(ts)
		}
		buf.Values[0] = times.NewArray()

		values := array.NewIntBuilder(mem)
		for v := int64(4); v < 8; v++ {
			values.Append(v)
		}
		buf.Values[1] = values.NewArray()
		if err := b.AppendBuffer(&buf); err != nil {
			t.Fatal(err)
		}
	}

	in, err := b.Table()
	if err != nil {
		t.Fatal(err)
	}

	spec := &universe.LimitProcedureSpec{
		N:      4,
		Offset: 2,
	}
	tr, d, err := universe.NewLimitTransformation(spec, executetest.RandomDatasetID(), memory.DefaultAllocator)
	if err != nil {
		t.Fatal(err)
	}
	store := executetest.NewDataStore()
	d.AddTransformation(store)

	parentID := executetest.RandomDatasetID()
	if err := tr.Process(parentID, in); err != nil {
		t.Fatal(err)
	}
	tr.Finish(parentID, nil)

	got, err := executetest.TablesFromCache(store)
	if err != nil {
		t.Fatal(err)
	}

	want := []*executetest.Table{
		{
			ColMeta: []flux.ColMeta{
				{Label: "_time", Type: flux.TTime},
				{Label: "_value", Type: flux.TInt},
			},
			Data: [][]interface{}{
				{values.Time(20), int64(2)},
				{values.Time(30), int64(3)},
				{values.Time(40), int64(4)},
				{values.Time(50), int64(5)},
			},
		},
	}
	executetest.NormalizeTables(want)

	sort.Sort(executetest.SortedTables(got))
	sort.Sort(executetest.SortedTables(want))

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected tables -want/+got\n%s", cmp.Diff(want, got))
	}
}

func BenchmarkLimit_1N_1000(b *testing.B) {
	benchmarkLimit(b, 1, 1000)
}

func BenchmarkLimit_100N_1000(b *testing.B) {
	benchmarkLimit(b, 100, 1000)
}

func benchmarkLimit(b *testing.B, n, l int) {
	spec := &universe.LimitProcedureSpec{
		N: int64(n),
	}
	executetest.ProcessBenchmarkHelper(b,
		func(alloc memory.Allocator) (flux.TableIterator, error) {
			schema := gen.Schema{
				NumPoints: l,
				Alloc:     alloc,
				Tags: []gen.Tag{
					{Name: "_measurement", Cardinality: 1},
					{Name: "_field", Cardinality: 6},
					{Name: "t0", Cardinality: 100},
					{Name: "t1", Cardinality: 50},
				},
			}
			return gen.Input(context.Background(), schema)
		},
		func(id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset) {
			tr, d, err := universe.NewLimitTransformation(spec, id, alloc)
			if err != nil {
				b.Fatal(err)
			}
			return tr, d
		},
	)
}
