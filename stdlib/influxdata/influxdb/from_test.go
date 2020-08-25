package influxdb_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb/internal/testutil"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/flux/values/valuestest"
)

func TestFrom_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "from no args",
			Raw:     `from()`,
			WantErr: true,
		},
		{
			Name:    "from unexpected arg",
			Raw:     `from(bucket:"telegraf", chicken:"what is this?")`,
			WantErr: true,
		},
		{
			Name: "from with database",
			Raw:  `from(bucket:"mybucket") |> range(start:-4h, stop:-2h) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "sum2"},
				},
			},
		},
		{
			Name: "from with host and token",
			Raw:  `from(bucket:"mybucket", host: "http://localhost:9999", token: "mytoken")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
							Host:   stringPtr("http://localhost:9999"),
							Token:  stringPtr("mytoken"),
						},
					},
				},
			},
		},
		{
			Name: "from with org",
			Raw:  `from(org: "influxdata", bucket:"mybucket")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Org:    &influxdb.NameOrID{Name: "influxdata"},
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
				},
			},
		},
		{
			Name: "from with org id and bucket id",
			Raw:  `from(orgID: "97aa81cc0e247dc4", bucketID: "1e01ac57da723035")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Org:    &influxdb.NameOrID{ID: "97aa81cc0e247dc4"},
							Bucket: influxdb.NameOrID{ID: "1e01ac57da723035"},
						},
					},
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

func TestFrom_Run(t *testing.T) {
	defaultTablesFn := func() []*executetest.Table {
		return []*executetest.Table{{
			KeyCols: []string{"_measurement", "_field"},
			ColMeta: []flux.ColMeta{
				{Label: "_time", Type: flux.TTime},
				{Label: "_measurement", Type: flux.TString},
				{Label: "_field", Type: flux.TString},
				{Label: "_value", Type: flux.TFloat},
			},
			Data: [][]interface{}{
				{execute.Time(0), "cpu", "usage_user", 2.0},
				{execute.Time(10), "cpu", "usage_user", 8.0},
				{execute.Time(20), "cpu", "usage_user", 5.0},
				{execute.Time(30), "cpu", "usage_user", 9.0},
				{execute.Time(40), "cpu", "usage_user", 3.0},
				{execute.Time(50), "cpu", "usage_user", 1.0},
			},
		}}
	}

	for _, tt := range []struct {
		name string
		spec *influxdb.FromRemoteProcedureSpec
		want testutil.Want
	}{
		{
			name: "basic query",
			spec: &influxdb.FromRemoteProcedureSpec{
				FromProcedureSpec: &influxdb.FromProcedureSpec{
					Org:    &influxdb.NameOrID{Name: "influxdata"},
					Bucket: influxdb.NameOrID{Name: "telegraf"},
					Token:  stringPtr("mytoken"),
				},
				Range: &universe.RangeProcedureSpec{
					Bounds: flux.Bounds{
						Start: flux.Time{
							IsRelative: true,
							Relative:   -time.Minute,
						},
						Stop: flux.Time{
							IsRelative: true,
						},
					},
				},
			},
			want: testutil.Want{
				Params: url.Values{
					"org": []string{"influxdata"},
				},
				Query: `package main


from(bucket: "telegraf")
	|> range(start: -1m)`,
				Tables: defaultTablesFn,
			},
		},
		{
			name: "basic query with org id and bucket id",
			spec: &influxdb.FromRemoteProcedureSpec{
				FromProcedureSpec: &influxdb.FromProcedureSpec{
					Org:    &influxdb.NameOrID{ID: "97aa81cc0e247dc4"},
					Bucket: influxdb.NameOrID{ID: "1e01ac57da723035"},
					Token:  stringPtr("mytoken"),
				},
				Range: &universe.RangeProcedureSpec{
					Bounds: flux.Bounds{
						Start: flux.Time{
							IsRelative: true,
							Relative:   -time.Minute,
						},
						Stop: flux.Time{
							IsRelative: true,
						},
					},
				},
			},
			want: testutil.Want{
				Params: url.Values{
					"orgID": []string{"97aa81cc0e247dc4"},
				},
				Query: `package main


from(bucketID: "1e01ac57da723035")
	|> range(start: -1m)`,
				Tables: defaultTablesFn,
			},
		},
		{
			name: "basic query with absolute time range",
			spec: &influxdb.FromRemoteProcedureSpec{
				FromProcedureSpec: &influxdb.FromProcedureSpec{
					Org:    &influxdb.NameOrID{Name: "influxdata"},
					Bucket: influxdb.NameOrID{Name: "telegraf"},
					Token:  stringPtr("mytoken"),
				},
				Range: &universe.RangeProcedureSpec{
					Bounds: flux.Bounds{
						Start: flux.Time{
							Absolute: mustParseTime("2018-05-30T09:00:00Z"),
						},
						Stop: flux.Time{
							Absolute: mustParseTime("2018-05-30T10:00:00Z"),
						},
					},
				},
			},
			want: testutil.Want{
				Params: url.Values{
					"org": []string{"influxdata"},
				},
				Query: `package main


from(bucket: "telegraf")
	|> range(start: 2018-05-30T09:00:00Z, stop: 2018-05-30T10:00:00Z)`,
				Tables: defaultTablesFn,
			},
		},
		{
			name: "filter query",
			spec: &influxdb.FromRemoteProcedureSpec{
				FromProcedureSpec: &influxdb.FromProcedureSpec{
					Org:    &influxdb.NameOrID{Name: "influxdata"},
					Bucket: influxdb.NameOrID{Name: "telegraf"},
					Token:  stringPtr("mytoken"),
				},
				Range: &universe.RangeProcedureSpec{
					Bounds: flux.Bounds{
						Start: flux.Time{
							IsRelative: true,
							Relative:   -time.Minute,
						},
						Stop: flux.Time{
							IsRelative: true,
						},
					},
				},
				Transformations: []plan.ProcedureSpec{
					&universe.FilterProcedureSpec{
						Fn: interpreter.ResolvedFunction{
							Fn:    executetest.FunctionExpression(t, `(r) => r._value >= 0.0`),
							Scope: valuestest.Scope(),
						},
					},
				},
			},
			want: testutil.Want{
				Params: url.Values{
					"org": []string{"influxdata"},
				},
				Query: `package main


from(bucket: "telegraf")
	|> range(start: -1m)
	|> filter(fn: (r) => {
		return r["_value"] >= 0.0
	})`,
				Tables: defaultTablesFn,
			},
		},
		{
			name: "filter query with keep empty",
			spec: &influxdb.FromRemoteProcedureSpec{
				FromProcedureSpec: &influxdb.FromProcedureSpec{
					Org:    &influxdb.NameOrID{Name: "influxdata"},
					Bucket: influxdb.NameOrID{Name: "telegraf"},
					Token:  stringPtr("mytoken"),
				},
				Range: &universe.RangeProcedureSpec{
					Bounds: flux.Bounds{
						Start: flux.Time{
							IsRelative: true,
							Relative:   -time.Minute,
						},
						Stop: flux.Time{
							IsRelative: true,
						},
					},
				},
				Transformations: []plan.ProcedureSpec{
					&universe.FilterProcedureSpec{
						Fn: interpreter.ResolvedFunction{
							Fn:    executetest.FunctionExpression(t, `(r) => r._value >= 0.0`),
							Scope: valuestest.Scope(),
						},
						KeepEmptyTables: true,
					},
				},
			},
			want: testutil.Want{
				Params: url.Values{
					"org": []string{"influxdata"},
				},
				Query: `package main


from(bucket: "telegraf")
	|> range(start: -1m)
	|> filter(fn: (r) => {
		return r["_value"] >= 0.0
	}, onEmpty: "keep")`,
				Tables: defaultTablesFn,
			},
		},
		{
			name: "filter query with import",
			spec: &influxdb.FromRemoteProcedureSpec{
				FromProcedureSpec: &influxdb.FromProcedureSpec{
					Org:    &influxdb.NameOrID{Name: "influxdata"},
					Bucket: influxdb.NameOrID{Name: "telegraf"},
					Token:  stringPtr("mytoken"),
				},
				Range: &universe.RangeProcedureSpec{
					Bounds: flux.Bounds{
						Start: flux.Time{
							IsRelative: true,
							Relative:   -time.Minute,
						},
						Stop: flux.Time{
							IsRelative: true,
						},
					},
				},
				Transformations: []plan.ProcedureSpec{
					&universe.FilterProcedureSpec{
						Fn: interpreter.ResolvedFunction{
							Fn: executetest.FunctionExpression(t, `
import "math"
(r) => r._value >= math.pi`,
							),
							Scope: func() values.Scope {
								imp := runtime.StdLib()
								// This is needed to prime the importer since universe
								// depends on math and the anti-cyclical import detection
								// doesn't work if you import math first.
								_, _ = imp.ImportPackageObject("universe")
								pkg, err := imp.ImportPackageObject("math")
								if err != nil {
									t.Fatal(err)
								}

								scope := values.NewScope()
								scope.Set("math", pkg)
								return scope
							}(),
						},
						KeepEmptyTables: true,
					},
				},
			},
			want: testutil.Want{
				Params: url.Values{
					"org": []string{"influxdata"},
				},
				Query: `package main


import math "math"

from(bucket: "telegraf")
	|> range(start: -1m)
	|> filter(fn: (r) => {
		return r["_value"] >= math["pi"]
	}, onEmpty: "keep")`,
				Tables: defaultTablesFn,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			testutil.RunSourceTestHelper(t, tt.spec, tt.want)
		})
	}
}

func TestFrom_Run_Errors(t *testing.T) {
	testutil.RunSourceErrorTestHelper(t, &influxdb.FromRemoteProcedureSpec{
		FromProcedureSpec: &influxdb.FromProcedureSpec{
			Org:    &influxdb.NameOrID{Name: "influxdata"},
			Bucket: influxdb.NameOrID{Name: "telegraf"},
			Token:  stringPtr("mytoken"),
		},
		Range: &universe.RangeProcedureSpec{
			Bounds: flux.Bounds{
				Start: flux.Time{
					IsRelative: true,
					Relative:   -time.Minute,
				},
				Stop: flux.Time{
					IsRelative: true,
				},
			},
		},
	})
}

func TestFrom_URLValidator(t *testing.T) {
	testutil.RunSourceURLValidatorTestHelper(t, &influxdb.FromRemoteProcedureSpec{
		FromProcedureSpec: &influxdb.FromProcedureSpec{
			Org:    &influxdb.NameOrID{Name: "influxdata"},
			Bucket: influxdb.NameOrID{Name: "telegraf"},
			Token:  stringPtr("mytoken"),
		},
		Range: &universe.RangeProcedureSpec{
			Bounds: flux.Bounds{
				Start: flux.Time{
					IsRelative: true,
					Relative:   -time.Minute,
				},
				Stop: flux.Time{
					IsRelative: true,
				},
			},
		},
	})
}

func TestFrom_HTTPClient(t *testing.T) {
	testutil.RunSourceHTTPClientTestHelper(t, &influxdb.FromRemoteProcedureSpec{
		FromProcedureSpec: &influxdb.FromProcedureSpec{
			Org:    &influxdb.NameOrID{Name: "influxdata"},
			Bucket: influxdb.NameOrID{Name: "telegraf"},
			Token:  stringPtr("mytoken"),
		},
		Range: &universe.RangeProcedureSpec{
			Bounds: flux.Bounds{
				Start: flux.Time{
					IsRelative: true,
					Relative:   -time.Minute,
				},
				Stop: flux.Time{
					IsRelative: true,
				},
			},
		},
	})
}

func stringPtr(v string) *string {
	return &v
}

func mustParseTime(v string) time.Time {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		panic(err)
	}
	return t
}
