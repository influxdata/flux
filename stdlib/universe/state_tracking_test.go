package universe_test

import (
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values/valuestest"
	"github.com/pkg/errors"
)

func TestStateTracking_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from range count",
			Raw:  `from(bucket:"mydb") |> range(start:-1h) |> stateCount(fn: (r) => true)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mydb",
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Hour,
								IsRelative: true,
							},
							Stop:        flux.Now,
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "stateTracking2",
						Spec: &universe.StateTrackingOpSpec{
							CountColumn:    "stateCount",
							DurationColumn: "",
							DurationUnit:   flux.Duration(time.Second),
							TimeColumn:     "_time",
							Fn: interpreter.ResolvedFunction{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
										},
										Body: &semantic.BooleanLiteral{Value: true},
									},
								},
								Scope: valuestest.NowScope(),
							},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "stateTracking2"},
				},
			},
		},
		{
			Name:    "from range count with time column",
			Raw:     `from(bucket:"mydb") |> range(start:-1h) |> stateCount(fn: (r) => true, timeColumn: "err")`,
			WantErr: true,
		},
		{
			Name: "from duration",
			Raw:  `from(bucket:"mydb") |> stateDuration(fn: (r) => true, timeColumn: "ts")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mydb",
						},
					},
					{
						ID: "stateTracking1",
						Spec: &universe.StateTrackingOpSpec{
							CountColumn:    "",
							DurationColumn: "stateDuration",
							DurationUnit:   flux.Duration(time.Second),
							TimeColumn:     "ts",
							Fn: interpreter.ResolvedFunction{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
										},
										Body: &semantic.BooleanLiteral{Value: true},
									},
								},
								Scope: valuestest.NowScope(),
							},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "stateTracking1"},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestStateTrackingOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"id","kind":"stateTracking","spec":{"countColumn":"c","durationColumn":"d","durationUnit":"1m","timeColumn":"t"}}`)
	op := &flux.Operation{
		ID: "id",
		Spec: &universe.StateTrackingOpSpec{
			CountColumn:    "c",
			DurationColumn: "d",
			DurationUnit:   flux.Duration(time.Minute),
			TimeColumn:     "t",
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestStateTracking_Process(t *testing.T) {
	gt5 := interpreter.ResolvedFunction{
		Fn: &semantic.FunctionExpression{
			Block: &semantic.FunctionBlock{
				Parameters: &semantic.FunctionParameters{
					List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
				},
				Body: &semantic.BinaryExpression{
					Operator: ast.GreaterThanOperator,
					Left: &semantic.MemberExpression{
						Object:   &semantic.IdentifierExpression{Name: "r"},
						Property: "_value",
					},
					Right: &semantic.FloatLiteral{Value: 5.0},
				},
			},
		},
		Scope: flux.Prelude(),
	}

	testCases := []struct {
		name    string
		spec    *universe.StateTrackingProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "only duration",
			spec: &universe.StateTrackingProcedureSpec{
				DurationColumn: "duration",
				DurationUnit:   1,
				Fn:             gt5,
				TimeCol:        "_time",
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
					{execute.Time(3), 6.0},
					{execute.Time(4), 7.0},
					{execute.Time(5), 8.0},
					{execute.Time(6), 1.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "duration", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, int64(-1)},
					{execute.Time(2), 1.0, int64(-1)},
					{execute.Time(3), 6.0, int64(0)},
					{execute.Time(4), 7.0, int64(1)},
					{execute.Time(5), 8.0, int64(2)},
					{execute.Time(6), 1.0, int64(-1)},
				},
			}},
		},
		{
			name: "only duration, null timestamps",
			spec: &universe.StateTrackingProcedureSpec{
				DurationColumn: "duration",
				DurationUnit:   1,
				Fn:             gt5,
				TimeCol:        "_time",
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
					{execute.Time(3), 6.0},
					{nil, 7.0},
					{execute.Time(5), 8.0},
					{nil, 1.0},
				},
			}},
			wantErr: errors.New("got a null timestamp"),
		},
		{
			name: "only duration, out of order timestamps",
			spec: &universe.StateTrackingProcedureSpec{
				DurationColumn: "duration",
				DurationUnit:   1,
				Fn:             gt5,
				TimeCol:        "_time",
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 1.0},
					{execute.Time(4), 7.0},
					{execute.Time(1), 2.0},
					{execute.Time(5), 8.0},
					{execute.Time(6), 1.0},
					{execute.Time(3), 6.0},
				},
			}},
			wantErr: errors.New("got an out-of-order timestamp"),
		},
		{
			name: "only count",
			spec: &universe.StateTrackingProcedureSpec{
				CountColumn: "count",
				Fn:          gt5,
				TimeCol:     "_time",
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
					{execute.Time(3), 6.0},
					{execute.Time(4), 7.0},
					{execute.Time(5), 8.0},
					{execute.Time(6), 1.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "count", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, int64(-1)},
					{execute.Time(2), 1.0, int64(-1)},
					{execute.Time(3), 6.0, int64(1)},
					{execute.Time(4), 7.0, int64(2)},
					{execute.Time(5), 8.0, int64(3)},
					{execute.Time(6), 1.0, int64(-1)},
				},
			}},
		},
		{
			name: "only count, out of order and null timestamps",
			spec: &universe.StateTrackingProcedureSpec{
				CountColumn: "count",
				Fn:          gt5,
				TimeCol:     "_time",
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(3), 6.0},
					{nil, 2.0},
					{execute.Time(5), 8.0},
					{nil, 1.0},
					{execute.Time(2), 7.0},
					{execute.Time(4), 10.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "count", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(3), 6.0, int64(1)},
					{nil, 2.0, int64(-1)},
					{execute.Time(5), 8.0, int64(1)},
					{nil, 1.0, int64(-1)},
					{execute.Time(2), 7.0, int64(1)},
					{execute.Time(4), 10.0, int64(2)},
				},
			}},
		},
		{
			name: "one table",
			spec: &universe.StateTrackingProcedureSpec{
				CountColumn:    "count",
				DurationColumn: "duration",
				DurationUnit:   1,
				Fn:             gt5,
				TimeCol:        "_time",
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
					{execute.Time(3), 6.0},
					{execute.Time(4), 7.0},
					{execute.Time(5), 8.0},
					{execute.Time(6), 1.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "count", Type: flux.TInt},
					{Label: "duration", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, int64(-1), int64(-1)},
					{execute.Time(2), 1.0, int64(-1), int64(-1)},
					{execute.Time(3), 6.0, int64(1), int64(0)},
					{execute.Time(4), 7.0, int64(2), int64(1)},
					{execute.Time(5), 8.0, int64(3), int64(2)},
					{execute.Time(6), 1.0, int64(-1), int64(-1)},
				},
			}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				tc.wantErr,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					tx, err := universe.NewStateTrackingTransformation(d, c, tc.spec)
					if err != nil {
						t.Fatal(err)
					}
					return tx
				},
			)
		})
	}
}
