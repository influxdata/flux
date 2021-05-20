package influxdb_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb/internal/testutil"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/flux/values/valuestest"
)

func TestCardinality_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "cardinality no args",
			Raw:     `influxdb.cardinality()`,
			WantErr: true,
		},
		{
			Name:    "cardinality unexpected arg",
			Raw:     `influxdb.cardinality(bucket:"telegraf", chicken:"what is this?")`,
			WantErr: true,
		},
		{
			Name: "cardinality with bucket and range",
			Raw:  `influxdb.cardinality(bucket:"mybucket",start:-4h,stop:-2h) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "influxdata/influxdb.cardinality0",
						Spec: &influxdb.CardinalityOpSpec{
							Config: influxdb.Config{
								Bucket: influxdb.NameOrID{Name: "mybucket"},
							},
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
						},
					},
					{
						ID: "sum1",
						Spec: &universe.SumOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "influxdata/influxdb.cardinality0", Child: "sum1"},
				},
			},
		},
		{
			Name: "cardinality with host and token",
			Raw:  `influxdb.cardinality(bucket:"mybucket", host: "http://localhost:8086", token: "mytoken", start: -2h)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "influxdata/influxdb.cardinality0",
						Spec: &influxdb.CardinalityOpSpec{
							Config: influxdb.Config{
								Bucket: influxdb.NameOrID{Name: "mybucket"},
								Host:   "http://localhost:8086",
								Token:  "mytoken",
							},
							Start: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Now,
						},
					},
				},
			},
		},
		{
			Name: "cardinality with org",
			Raw:  `influxdb.cardinality(org: "influxdata", bucket:"mybucket", start: -2h)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "influxdata/influxdb.cardinality0",
						Spec: &influxdb.CardinalityOpSpec{
							Config: influxdb.Config{
								Org:    influxdb.NameOrID{Name: "influxdata"},
								Bucket: influxdb.NameOrID{Name: "mybucket"},
							},
							Start: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Now,
						},
					},
				},
			},
		},
		{
			Name: "cardinality with org id and bucket id",
			Raw:  `influxdb.cardinality(orgID: "97aa81cc0e247dc4", bucketID: "1e01ac57da723035", start: -2h)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "influxdata/influxdb.cardinality0",
						Spec: &influxdb.CardinalityOpSpec{
							Config: influxdb.Config{
								Org:    influxdb.NameOrID{ID: "97aa81cc0e247dc4"},
								Bucket: influxdb.NameOrID{ID: "1e01ac57da723035"},
							},
							Start: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Now,
						},
					},
				},
			},
		},
	}

	const prefix = "import \"influxdata/influxdb\"\n"
	for _, tc := range tests {
		tc := tc
		tc.Raw = prefix + tc.Raw
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestCardinality_Run(t *testing.T) {
	defaultTablesFn := func() []*executetest.Table {
		return []*executetest.Table{{
			ColMeta: []flux.ColMeta{
				{Label: "_value", Type: flux.TInt},
			},
			Data: [][]interface{}{
				{int64(3)},
			},
		}}
	}

	now := mustParseTime("2020-10-22T09:30:00Z")
	for _, tt := range []struct {
		name string
		spec *influxdb.CardinalityProcedureSpec
		want testutil.Want
	}{
		{
			name: "basic query",
			spec: &influxdb.CardinalityProcedureSpec{
				Config: influxdb.Config{
					Org:    influxdb.NameOrID{Name: "influxdata"},
					Bucket: influxdb.NameOrID{Name: "telegraf"},
					Token:  "mytoken",
				},
				Bounds: flux.Bounds{
					Start: flux.Time{
						IsRelative: true,
						Relative:   -time.Minute,
					},
					Stop: flux.Time{
						IsRelative: true,
					},
					Now: now,
				},
			},
			want: testutil.Want{
				Params: url.Values{
					"org": []string{"influxdata"},
				},
				Query: `package main


import influxdb "influxdata/influxdb"

influxdb.cardinality(bucket: "telegraf", start: 2020-10-22T09:29:00Z, stop: 2020-10-22T09:30:00Z)`,
				Tables: defaultTablesFn,
			},
		},
		{
			name: "basic query with org id and bucket id",
			spec: &influxdb.CardinalityProcedureSpec{
				Config: influxdb.Config{
					Org:    influxdb.NameOrID{ID: "97aa81cc0e247dc4"},
					Bucket: influxdb.NameOrID{ID: "1e01ac57da723035"},
					Token:  "mytoken",
				},
				Bounds: flux.Bounds{
					Start: flux.Time{
						IsRelative: true,
						Relative:   -time.Minute,
					},
					Stop: flux.Time{
						IsRelative: true,
					},
					Now: now,
				},
			},
			want: testutil.Want{
				Params: url.Values{
					"orgID": []string{"97aa81cc0e247dc4"},
				},
				Query: `package main


import influxdb "influxdata/influxdb"

influxdb.cardinality(bucketID: "1e01ac57da723035", start: 2020-10-22T09:29:00Z, stop: 2020-10-22T09:30:00Z)`,
				Tables: defaultTablesFn,
			},
		},
		{
			name: "basic query with absolute time range",
			spec: &influxdb.CardinalityProcedureSpec{
				Config: influxdb.Config{
					Org:    influxdb.NameOrID{Name: "influxdata"},
					Bucket: influxdb.NameOrID{Name: "telegraf"},
					Token:  "mytoken",
				},
				Bounds: flux.Bounds{
					Start: flux.Time{
						Absolute: mustParseTime("2018-05-30T09:00:00Z"),
					},
					Stop: flux.Time{
						Absolute: mustParseTime("2018-05-30T10:00:00Z"),
					},
					Now: now,
				},
			},
			want: testutil.Want{
				Params: url.Values{
					"org": []string{"influxdata"},
				},
				Query: `package main


import influxdb "influxdata/influxdb"

influxdb.cardinality(bucket: "telegraf", start: 2018-05-30T09:00:00Z, stop: 2018-05-30T10:00:00Z)`,
				Tables: defaultTablesFn,
			},
		},
		{
			name: "filter query",
			spec: &influxdb.CardinalityProcedureSpec{
				Config: influxdb.Config{
					Org:    influxdb.NameOrID{Name: "influxdata"},
					Bucket: influxdb.NameOrID{Name: "telegraf"},
					Token:  "mytoken",
				},
				Bounds: flux.Bounds{
					Start: flux.Time{
						IsRelative: true,
						Relative:   -time.Minute,
					},
					Stop: flux.Time{
						IsRelative: true,
					},
					Now: now,
				},
				PredicateSet: influxdb.PredicateSet{{
					ResolvedFunction: interpreter.ResolvedFunction{
						Fn:    executetest.FunctionExpression(t, `(r) => r._value >= 0.0`),
						Scope: valuestest.Scope(),
					},
				}},
			},
			want: testutil.Want{
				Params: url.Values{
					"org": []string{"influxdata"},
				},
				Query: `package main


import influxdb "influxdata/influxdb"

influxdb.cardinality(
	bucket: "telegraf",
	start: 2020-10-22T09:29:00Z,
	stop: 2020-10-22T09:30:00Z,
	predicate: (r) => {
		return r["_value"] >= 0.0
	},
)`,
				Tables: defaultTablesFn,
			},
		},
		{
			name: "filter query with import",
			spec: &influxdb.CardinalityProcedureSpec{
				Config: influxdb.Config{
					Org:    influxdb.NameOrID{Name: "influxdata"},
					Bucket: influxdb.NameOrID{Name: "telegraf"},
					Token:  "mytoken",
				},
				Bounds: flux.Bounds{
					Start: flux.Time{
						IsRelative: true,
						Relative:   -time.Minute,
					},
					Stop: flux.Time{
						IsRelative: true,
					},
					Now: now,
				},
				PredicateSet: influxdb.PredicateSet{{
					ResolvedFunction: interpreter.ResolvedFunction{
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
				}},
			},
			want: testutil.Want{
				Params: url.Values{
					"org": []string{"influxdata"},
				},
				Query: `package main


import influxdb "influxdata/influxdb"
import math "math"

influxdb.cardinality(
	bucket: "telegraf",
	start: 2020-10-22T09:29:00Z,
	stop: 2020-10-22T09:30:00Z,
	predicate: (r) => {
		return r["_value"] >= math["pi"]
	},
)`,
				Tables: defaultTablesFn,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			testutil.RunSourceTestHelper(t, tt.spec, tt.want)
		})
	}
}
