package universe_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/runtime"
	_ "github.com/influxdata/flux/stdlib"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

func init() {
	runtime.FinalizeBuiltIns()
}

func TestCovariance_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "simple covariance",
			Raw:  `from(bucket:"mybucket") |> covariance(columns:["a","b"],)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "covariance1",
						Spec: &universe.CovarianceOpSpec{
							ValueDst: execute.DefaultValueColLabel,
							Columns:  []string{"a", "b"},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "covariance1"},
				},
			},
		},
		{
			Name: "pearsonr",
			Raw:  `from(bucket:"mybucket")|>covariance(columns:["a","b"],pearsonr:true)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "covariance1",
						Spec: &universe.CovarianceOpSpec{
							ValueDst:           execute.DefaultValueColLabel,
							PearsonCorrelation: true,
							Columns:            []string{"a", "b"},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "covariance1"},
				},
			},
		},
		{
			Name: "global covariance",
			Raw:  `cov(x: from(bucket:"mybucket"), y:from(bucket:"mybucket"), on:["host"], pearsonr:true)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "from1",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "join2",
						Spec: &universe.JoinOpSpec{
							On: []string{"host"},
							TableNames: map[flux.OperationID]string{
								"from0": "x",
								"from1": "y",
							},
							Method: "inner",
						},
					},
					{
						ID: "covariance3",
						Spec: &universe.CovarianceOpSpec{
							ValueDst:           execute.DefaultValueColLabel,
							PearsonCorrelation: true,
							Columns:            []string{"_value_x", "_value_y"},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "join2"},
					{Parent: "from1", Child: "join2"},
					{Parent: "join2", Child: "covariance3"},
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

func TestCovarianceOperation_Marshaling(t *testing.T) {
	data := []byte(`{
		"id":"covariance",
		"kind":"covariance",
		"spec":{
			"pearsonr":true
		}
	}`)
	op := &flux.Operation{
		ID: "covariance",
		Spec: &universe.CovarianceOpSpec{
			PearsonCorrelation: true,
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestCovariance_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.CovarianceProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "variance",
			spec: &universe.CovarianceProcedureSpec{
				ValueLabel: execute.DefaultValueColLabel,
				Columns:    []string{"x", "y"},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(5), execute.Time(0), 1.0, 1.0},
					{execute.Time(0), execute.Time(5), execute.Time(1), 2.0, 2.0},
					{execute.Time(0), execute.Time(5), execute.Time(2), 3.0, 3.0},
					{execute.Time(0), execute.Time(5), execute.Time(3), 4.0, 4.0},
					{execute.Time(0), execute.Time(5), execute.Time(4), 5.0, 5.0},
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
					{execute.Time(0), execute.Time(5), 2.5},
				},
			}},
		},
		{
			name: "negative covariance",
			spec: &universe.CovarianceProcedureSpec{
				ValueLabel: execute.DefaultValueColLabel,
				Columns:    []string{"x", "y"},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(5), execute.Time(0), 1.0, 5.0},
					{execute.Time(0), execute.Time(5), execute.Time(1), 2.0, 4.0},
					{execute.Time(0), execute.Time(5), execute.Time(2), 3.0, 3.0},
					{execute.Time(0), execute.Time(5), execute.Time(3), 4.0, 2.0},
					{execute.Time(0), execute.Time(5), execute.Time(4), 5.0, 1.0},
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
					{execute.Time(0), execute.Time(5), -2.5},
				},
			}},
		},
		{
			name: "small covariance",
			spec: &universe.CovarianceProcedureSpec{
				ValueLabel: execute.DefaultValueColLabel,
				Columns:    []string{"x", "y"},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(5), execute.Time(0), 1.0, 1.0},
					{execute.Time(0), execute.Time(5), execute.Time(1), 2.0, 1.0},
					{execute.Time(0), execute.Time(5), execute.Time(2), 3.0, 1.0},
					{execute.Time(0), execute.Time(5), execute.Time(3), 4.0, 1.0},
					{execute.Time(0), execute.Time(5), execute.Time(4), 5.0, 2.0},
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
					{execute.Time(0), execute.Time(5), 0.5},
				},
			}},
		},
		{
			name: "pearson correlation",
			spec: &universe.CovarianceProcedureSpec{
				ValueLabel:         execute.DefaultValueColLabel,
				PearsonCorrelation: true,
				Columns:            []string{"x", "y"},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(5), execute.Time(0), 1.0, 1.0},
					{execute.Time(0), execute.Time(5), execute.Time(1), 2.0, 2.0},
					{execute.Time(0), execute.Time(5), execute.Time(2), 3.0, 3.0},
					{execute.Time(0), execute.Time(5), execute.Time(3), 4.0, 4.0},
					{execute.Time(0), execute.Time(5), execute.Time(4), 5.0, 5.0},
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
					{execute.Time(0), execute.Time(5), 1.0},
				},
			}},
		},
		{
			name: "pearson correlation opposite",
			spec: &universe.CovarianceProcedureSpec{
				ValueLabel:         execute.DefaultValueColLabel,
				PearsonCorrelation: true,
				Columns:            []string{"x", "y"},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(5), execute.Time(0), 1.0, 5.0},
					{execute.Time(0), execute.Time(5), execute.Time(1), 2.0, 4.0},
					{execute.Time(0), execute.Time(5), execute.Time(2), 3.0, 3.0},
					{execute.Time(0), execute.Time(5), execute.Time(3), 4.0, 2.0},
					{execute.Time(0), execute.Time(5), execute.Time(4), 5.0, 1.0},
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
					{execute.Time(0), execute.Time(5), -1.0},
				},
			}},
		},
		{
			name: "variance with nulls",
			spec: &universe.CovarianceProcedureSpec{
				ValueLabel: execute.DefaultValueColLabel,
				Columns:    []string{"x", "y"},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_start", "_stop"},
				ColMeta: []flux.ColMeta{
					{Label: "_start", Type: flux.TTime},
					{Label: "_stop", Type: flux.TTime},
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), execute.Time(9), execute.Time(0), 1.0, 1.0},
					{execute.Time(0), execute.Time(9), execute.Time(1), 2.0, 2.0},
					{execute.Time(0), execute.Time(9), execute.Time(2), nil, 3.0},
					{execute.Time(0), execute.Time(9), execute.Time(3), 4.0, 4.0},
					{execute.Time(0), execute.Time(9), execute.Time(4), 5.0, nil},
					{execute.Time(0), execute.Time(9), execute.Time(5), 6.0, 6.0},
					{execute.Time(0), execute.Time(9), execute.Time(6), nil, nil},
					{execute.Time(0), execute.Time(9), execute.Time(7), 8.0, 8.0},
					{execute.Time(0), execute.Time(9), execute.Time(8), 9.0, 9.0},
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
					{execute.Time(0), execute.Time(9), 10.4},
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
				nil,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return universe.NewCovarianceTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
