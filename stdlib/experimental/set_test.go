package experimental_test

import (
	"sort"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/experimental"
	"github.com/influxdata/flux/values"
)

// newOrderedObject creates a object with sorted keys order
func newOrderedObject(vs map[string]values.Value) values.Object {
	keys := make([]string, 0, len(vs))
	for key := range vs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	properties := make([]semantic.PropertyType, 0, len(keys))
	for _, key := range keys {
		value := vs[key]
		properties = append(properties, semantic.PropertyType{
			Key:   []byte(key),
			Value: value.Type(),
		})
	}

	object := values.NewObject(semantic.NewObjectType(properties))
	object.Range(func(name string, _ values.Value) {
		object.Set(name, vs[name])
	})
	return object
}

func TestSet_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *experimental.SetProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "new col",
			spec: &experimental.SetProcedureSpec{
				Object: newOrderedObject(map[string]values.Value{
					"a": values.NewInt(23),
					"b": values.NewString("hi"),
				}),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "a", Type: flux.TInt},
					{Label: "b", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, int64(23), "hi"},
					{execute.Time(2), 1.0, int64(23), "hi"},
				},
			}},
		},
		{
			name: "new col with null value",
			spec: &experimental.SetProcedureSpec{
				Object: newOrderedObject(map[string]values.Value{
					"t1": values.NewString("bob"),
				}),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), nil},
					{execute.Time(2), 1.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), nil, "bob"},
					{execute.Time(2), 1.0, "bob"},
				},
			}},
		},
		{
			name: "replace col",
			spec: &experimental.SetProcedureSpec{
				Object: newOrderedObject(map[string]values.Value{
					"t1": values.NewString("bob"),
				}),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "jim"},
					{execute.Time(2), 2.0, "sue"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "bob"},
					{execute.Time(2), 2.0, "bob"},
				},
			}},
		},
		{
			name: "replace col type",
			spec: &experimental.SetProcedureSpec{
				Object: newOrderedObject(map[string]values.Value{
					"t1": values.NewInt(4),
				}),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "jim"},
					{execute.Time(2), 2.0, "sue"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, int64(4)},
					{execute.Time(2), 2.0, int64(4)},
				},
			}},
		},
		{
			name: "replace col with null",
			spec: &experimental.SetProcedureSpec{
				Object: newOrderedObject(map[string]values.Value{
					"t1": values.NewString("bob"),
				}),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, nil},
					{execute.Time(2), 2.0, "sue"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "bob"},
					{execute.Time(2), 2.0, "bob"},
				},
			}},
		},
		{
			name: "replace key col",
			spec: &experimental.SetProcedureSpec{
				Object: newOrderedObject(map[string]values.Value{
					"t1": values.NewString("bob"),
				}),
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "alice", "a"},
					{execute.Time(2), 1.0, "alice", "b"},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "bob", "a"},
					{execute.Time(2), 1.0, "bob", "b"},
				},
			}},
		},
		{
			name: "replace key col type",
			spec: &experimental.SetProcedureSpec{
				Object: newOrderedObject(map[string]values.Value{
					"t1": values.NewFloat(4.5),
				}),
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "alice", "a"},
					{execute.Time(2), 1.0, "alice", "b"},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TFloat},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, 4.5, "a"},
					{execute.Time(2), 1.0, 4.5, "b"},
				},
			}},
		},
		{
			name: "replace common col, merging tables",
			spec: &experimental.SetProcedureSpec{
				Object: newOrderedObject(map[string]values.Value{
					"t1": values.NewString("bob"),
				}),
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "alice"},
						{execute.Time(2), 1.0, "alice"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(3), 3.0, "sue"},
						{execute.Time(4), 5.0, "sue"},
					},
				},
			},
			want: []*executetest.Table{{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "bob"},
					{execute.Time(2), 1.0, "bob"},
					{execute.Time(3), 3.0, "bob"},
					{execute.Time(4), 5.0, "bob"},
				},
			}},
		},
		{
			name: "new common col, multiple tables",
			spec: &experimental.SetProcedureSpec{
				Object: newOrderedObject(map[string]values.Value{
					"t2": values.NewString("bob"),
				}),
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "alice"},
						{execute.Time(2), 1.0, "alice"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(3), 3.0, "sue"},
						{execute.Time(4), 5.0, "sue"},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "alice", "bob"},
						{execute.Time(2), 1.0, "alice", "bob"},
					},
				},
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(3), 3.0, "sue", "bob"},
						{execute.Time(4), 5.0, "sue", "bob"},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				nil,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return experimental.NewSetTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
