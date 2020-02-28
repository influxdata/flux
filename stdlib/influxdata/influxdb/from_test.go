package influxdb_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/csv"
	urldeps "github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
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
	type want struct {
		params url.Values
		ast    *ast.Package
		tables func() []*executetest.Table
	}

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
		want want
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
			want: want{
				params: url.Values{
					"org": []string{"influxdata"},
				},
				ast: &ast.Package{
					Package: "main",
					Files: []*ast.File{{
						Name: "query.flux",
						Package: &ast.PackageClause{
							Name: &ast.Identifier{Name: "main"},
						},
						Body: []ast.Statement{
							&ast.ExpressionStatement{
								Expression: &ast.PipeExpression{
									Argument: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "from"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key:   &ast.Identifier{Name: "bucket"},
														Value: &ast.StringLiteral{Value: "telegraf"},
													},
												},
											},
										},
									},
									Call: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "range"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key: &ast.Identifier{Name: "start"},
														Value: &ast.DurationLiteral{Values: []ast.Duration{
															{Magnitude: -1, Unit: "m"},
														}},
													},
												},
											},
										},
									},
								},
							},
						},
					}},
				},
				tables: defaultTablesFn,
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
			want: want{
				params: url.Values{
					"orgID": []string{"97aa81cc0e247dc4"},
				},
				ast: &ast.Package{
					Package: "main",
					Files: []*ast.File{{
						Name: "query.flux",
						Package: &ast.PackageClause{
							Name: &ast.Identifier{Name: "main"},
						},
						Body: []ast.Statement{
							&ast.ExpressionStatement{
								Expression: &ast.PipeExpression{
									Argument: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "from"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key:   &ast.Identifier{Name: "bucketID"},
														Value: &ast.StringLiteral{Value: "1e01ac57da723035"},
													},
												},
											},
										},
									},
									Call: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "range"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key: &ast.Identifier{Name: "start"},
														Value: &ast.DurationLiteral{Values: []ast.Duration{
															{Magnitude: -1, Unit: "m"},
														}},
													},
												},
											},
										},
									},
								},
							},
						},
					}},
				},
				tables: defaultTablesFn,
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
			want: want{
				params: url.Values{
					"org": []string{"influxdata"},
				},
				ast: &ast.Package{
					Package: "main",
					Files: []*ast.File{{
						Name: "query.flux",
						Package: &ast.PackageClause{
							Name: &ast.Identifier{Name: "main"},
						},
						Body: []ast.Statement{
							&ast.ExpressionStatement{
								Expression: &ast.PipeExpression{
									Argument: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "from"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key:   &ast.Identifier{Name: "bucket"},
														Value: &ast.StringLiteral{Value: "telegraf"},
													},
												},
											},
										},
									},
									Call: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "range"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key: &ast.Identifier{Name: "start"},
														Value: &ast.DateTimeLiteral{
															Value: mustParseTime("2018-05-30T09:00:00Z"),
														},
													},
													{
														Key: &ast.Identifier{Name: "stop"},
														Value: &ast.DateTimeLiteral{
															Value: mustParseTime("2018-05-30T10:00:00Z"),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					}},
				},
				tables: defaultTablesFn,
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
			want: want{
				params: url.Values{
					"org": []string{"influxdata"},
				},
				ast: &ast.Package{
					Package: "main",
					Files: []*ast.File{{
						Name: "query.flux",
						Package: &ast.PackageClause{
							Name: &ast.Identifier{Name: "main"},
						},
						Body: []ast.Statement{
							&ast.ExpressionStatement{
								Expression: &ast.PipeExpression{
									Argument: &ast.PipeExpression{
										Argument: &ast.CallExpression{
											Callee: &ast.Identifier{Name: "from"},
											Arguments: []ast.Expression{
												&ast.ObjectExpression{
													Properties: []*ast.Property{
														{
															Key:   &ast.Identifier{Name: "bucket"},
															Value: &ast.StringLiteral{Value: "telegraf"},
														},
													},
												},
											},
										},
										Call: &ast.CallExpression{
											Callee: &ast.Identifier{Name: "range"},
											Arguments: []ast.Expression{
												&ast.ObjectExpression{
													Properties: []*ast.Property{
														{
															Key: &ast.Identifier{Name: "start"},
															Value: &ast.DurationLiteral{Values: []ast.Duration{
																{Magnitude: -1, Unit: "m"},
															}},
														},
													},
												},
											},
										},
									},
									Call: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "filter"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key: &ast.Identifier{Name: "fn"},
														Value: &ast.FunctionExpression{
															Params: []*ast.Property{{
																Key: &ast.Identifier{Name: "r"},
															}},
															Body: &ast.Block{
																Body: []ast.Statement{
																	&ast.ReturnStatement{
																		Argument: &ast.BinaryExpression{
																			Operator: ast.GreaterThanEqualOperator,
																			Left: &ast.MemberExpression{
																				Object:   &ast.Identifier{Name: "r"},
																				Property: &ast.StringLiteral{Value: "_value"},
																			},
																			Right: &ast.FloatLiteral{Value: 0.0},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					}},
				},
				tables: defaultTablesFn,
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
			want: want{
				params: url.Values{
					"org": []string{"influxdata"},
				},
				ast: &ast.Package{
					Package: "main",
					Files: []*ast.File{{
						Name: "query.flux",
						Package: &ast.PackageClause{
							Name: &ast.Identifier{Name: "main"},
						},
						Body: []ast.Statement{
							&ast.ExpressionStatement{
								Expression: &ast.PipeExpression{
									Argument: &ast.PipeExpression{
										Argument: &ast.CallExpression{
											Callee: &ast.Identifier{Name: "from"},
											Arguments: []ast.Expression{
												&ast.ObjectExpression{
													Properties: []*ast.Property{
														{
															Key:   &ast.Identifier{Name: "bucket"},
															Value: &ast.StringLiteral{Value: "telegraf"},
														},
													},
												},
											},
										},
										Call: &ast.CallExpression{
											Callee: &ast.Identifier{Name: "range"},
											Arguments: []ast.Expression{
												&ast.ObjectExpression{
													Properties: []*ast.Property{
														{
															Key: &ast.Identifier{Name: "start"},
															Value: &ast.DurationLiteral{Values: []ast.Duration{
																{Magnitude: -1, Unit: "m"},
															}},
														},
													},
												},
											},
										},
									},
									Call: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "filter"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key: &ast.Identifier{Name: "fn"},
														Value: &ast.FunctionExpression{
															Params: []*ast.Property{{
																Key: &ast.Identifier{Name: "r"},
															}},
															Body: &ast.Block{
																Body: []ast.Statement{
																	&ast.ReturnStatement{
																		Argument: &ast.BinaryExpression{
																			Operator: ast.GreaterThanEqualOperator,
																			Left: &ast.MemberExpression{
																				Object:   &ast.Identifier{Name: "r"},
																				Property: &ast.StringLiteral{Value: "_value"},
																			},
																			Right: &ast.FloatLiteral{Value: 0.0},
																		},
																	},
																},
															},
														},
													},
													{
														Key:   &ast.Identifier{Name: "onEmpty"},
														Value: &ast.StringLiteral{Value: "keep"},
													},
												},
											},
										},
									},
								},
							},
						},
					}},
				},
				tables: defaultTablesFn,
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
			want: want{
				params: url.Values{
					"org": []string{"influxdata"},
				},
				ast: &ast.Package{
					Package: "main",
					Files: []*ast.File{{
						Name: "query.flux",
						Package: &ast.PackageClause{
							Name: &ast.Identifier{Name: "main"},
						},
						Imports: []*ast.ImportDeclaration{
							{
								Path: &ast.StringLiteral{Value: "math"},
								As:   &ast.Identifier{Name: "math"},
							},
						},
						Body: []ast.Statement{
							&ast.ExpressionStatement{
								Expression: &ast.PipeExpression{
									Argument: &ast.PipeExpression{
										Argument: &ast.CallExpression{
											Callee: &ast.Identifier{Name: "from"},
											Arguments: []ast.Expression{
												&ast.ObjectExpression{
													Properties: []*ast.Property{
														{
															Key:   &ast.Identifier{Name: "bucket"},
															Value: &ast.StringLiteral{Value: "telegraf"},
														},
													},
												},
											},
										},
										Call: &ast.CallExpression{
											Callee: &ast.Identifier{Name: "range"},
											Arguments: []ast.Expression{
												&ast.ObjectExpression{
													Properties: []*ast.Property{
														{
															Key: &ast.Identifier{Name: "start"},
															Value: &ast.DurationLiteral{Values: []ast.Duration{
																{Magnitude: -1, Unit: "m"},
															}},
														},
													},
												},
											},
										},
									},
									Call: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "filter"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key: &ast.Identifier{Name: "fn"},
														Value: &ast.FunctionExpression{
															Params: []*ast.Property{{
																Key: &ast.Identifier{Name: "r"},
															}},
															Body: &ast.Block{
																Body: []ast.Statement{
																	&ast.ReturnStatement{
																		Argument: &ast.BinaryExpression{
																			Operator: ast.GreaterThanEqualOperator,
																			Left: &ast.MemberExpression{
																				Object:   &ast.Identifier{Name: "r"},
																				Property: &ast.StringLiteral{Value: "_value"},
																			},
																			Right: &ast.MemberExpression{
																				Object:   &ast.Identifier{Name: "math"},
																				Property: &ast.StringLiteral{Value: "pi"},
																			},
																		},
																	},
																},
															},
														},
													},
													{
														Key:   &ast.Identifier{Name: "onEmpty"},
														Value: &ast.StringLiteral{Value: "keep"},
													},
												},
											},
										},
									},
								},
							},
						},
					}},
				},
				tables: defaultTablesFn,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if want, got := "/api/v2/query", r.URL.Path; want != got {
					t.Errorf("unexpected query path -want/+got:\n- %q\n+ %q", want, got)
				}
				if want, got := tt.want.params, r.URL.Query(); !cmp.Equal(want, got) {
					t.Errorf("unexpected query params -want/+got:\n%s", cmp.Diff(want, got))
				}
				if want, got := "application/json", r.Header.Get("Content-Type"); want != got {
					t.Errorf("unexpected query content type -want/+got:\n- %q\n+ %q", want, got)
					return
				}

				var req struct {
					AST     *ast.Package `json:"ast"`
					Dialect struct {
						Header         bool     `json:"header"`
						DateTimeFormat string   `json:"dateTimeFormat"`
						Annotations    []string `json:"annotations"`
					} `json:"dialect"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("client did not send json: %s", err)
					return
				}

				if want, got := tt.want.ast, req.AST; !cmp.Equal(want, got) {
					t.Errorf("unexpected ast in request body -want/+got:\n%s", cmp.Diff(want, got))
				}

				w.Header().Add("Content-Type", "text/csv")
				results := flux.NewSliceResultIterator([]flux.Result{
					&executetest.Result{
						Nm:   "_result",
						Tbls: tt.want.tables(),
					},
				})
				enc := csv.NewMultiResultEncoder(csv.ResultEncoderConfig{
					Annotations: req.Dialect.Annotations,
					NoHeader:    !req.Dialect.Header,
					Delimiter:   ',',
				})
				if _, err := enc.Encode(w, results); err != nil {
					t.Errorf("error encoding results: %s", err)
				}
			}))
			defer server.Close()

			spec := tt.spec.Copy().(*influxdb.FromRemoteProcedureSpec)
			spec.Host = stringPtr(server.URL)

			deps := flux.NewDefaultDependencies()
			ctx := deps.Inject(context.Background())
			store := executetest.NewDataStore()
			s, err := influxdb.CreateSource(ctx, spec)
			if err != nil {
				t.Fatal(err)
			}
			s.AddTransformation(store)
			s.Run(context.Background())

			if err := store.Err(); err != nil {
				t.Fatal(err)
			}

			got, err := executetest.TablesFromCache(store)
			if err != nil {
				t.Fatal(err)
			}
			executetest.NormalizeTables(got)

			want := tt.want.tables()
			executetest.NormalizeTables(want)

			if !cmp.Equal(want, got) {
				t.Errorf("unexpected tables returned from server -want/+got:\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestFrom_Run_Errors(t *testing.T) {
	for _, tt := range []struct {
		name string
		fn   func(w http.ResponseWriter)
		want error
	}{
		{
			name: "internal error",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = io.WriteString(w, `{"code":"internal error","message":"An internal error has occurred"}`)
			},
			want: errors.New(codes.Internal, "An internal error has occurred"),
		},
		{
			name: "not found",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = io.WriteString(w, `{"code":"not found","message":"bucket not found"}`)
			},
			want: errors.New(codes.NotFound, "bucket not found"),
		},
		{
			name: "invalid",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"code":"invalid","message":"query was invalid"}`)
			},
			want: errors.New(codes.Invalid, "query was invalid"),
		},
		{
			name: "unavailable",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = io.WriteString(w, `{"code":"unavailable","message":"service unavailable"}`)
			},
			want: errors.New(codes.Unavailable, "service unavailable"),
		},
		{
			name: "forbidden",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = io.WriteString(w, `{"code":"forbidden","message":"user does not have access to bucket"}`)
			},
			want: errors.New(codes.PermissionDenied, "user does not have access to bucket"),
		},
		{
			name: "unauthorized",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = io.WriteString(w, `{"code":"unauthorized","message":"credentials required"}`)
			},
			want: errors.New(codes.Unauthenticated, "credentials required"),
		},
		{
			name: "nested influxdb error",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"code":"invalid","message":"query was invalid","error":{"code":"not found","message":"resource not found"}}`)
			},
			want: errors.Wrap(
				errors.New(codes.NotFound, "resource not found"),
				codes.Invalid,
				"query was invalid",
			),
		},
		{
			name: "nested internal error",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"code":"invalid","message":"query was invalid","error":"internal error"}`)
			},
			want: errors.Wrap(
				errors.New(codes.Unknown, "internal error"),
				codes.Invalid,
				"query was invalid",
			),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.fn(w)
			}))
			defer server.Close()

			spec := &influxdb.FromRemoteProcedureSpec{
				FromProcedureSpec: &influxdb.FromProcedureSpec{
					Org:    &influxdb.NameOrID{Name: "influxdata"},
					Bucket: influxdb.NameOrID{Name: "telegraf"},
					Host:   stringPtr(server.URL),
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
			}

			deps := flux.NewDefaultDependencies()
			ctx := deps.Inject(context.Background())
			store := executetest.NewDataStore()
			s, err := influxdb.CreateSource(ctx, spec)
			if err != nil {
				t.Fatal(err)
			}
			s.AddTransformation(store)
			s.Run(context.Background())

			got := store.Err()
			if got == nil {
				t.Fatal("expected error")
			}
			want := tt.want

			if !cmp.Equal(want, got) {
				t.Errorf("unexpected error:\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestFrom_URLValidator(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("received unexpected request")
	}))
	defer server.Close()

	spec := &influxdb.FromRemoteProcedureSpec{
		FromProcedureSpec: &influxdb.FromProcedureSpec{
			Org:    &influxdb.NameOrID{Name: "influxdata"},
			Bucket: influxdb.NameOrID{Name: "telegraf"},
			Host:   stringPtr(server.URL),
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
	}

	deps := flux.NewDefaultDependencies()
	deps.Deps.URLValidator = urldeps.PrivateIPValidator{}
	ctx := deps.Inject(context.Background())
	if _, err := influxdb.CreateSource(ctx, spec); err == nil {
		t.Fatal("expected error")
	}
}

func TestFrom_HTTPClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(""))
	}))
	defer server.Close()

	spec := &influxdb.FromRemoteProcedureSpec{
		FromProcedureSpec: &influxdb.FromProcedureSpec{
			Org:    &influxdb.NameOrID{Name: "influxdata"},
			Bucket: influxdb.NameOrID{Name: "telegraf"},
			Host:   stringPtr(server.URL),
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
	}

	counter := &RequestCounter{}
	deps := flux.NewDefaultDependencies()
	deps.Deps.HTTPClient = &http.Client{
		Transport: counter,
	}
	ctx := deps.Inject(context.Background())
	store := executetest.NewDataStore()
	s, err := influxdb.CreateSource(ctx, spec)
	if err != nil {
		t.Fatal(err)
	}
	s.AddTransformation(store)
	s.Run(context.Background())

	if err := store.Err(); err != nil {
		t.Fatal(err)
	}

	if counter.Count == 0 {
		t.Error("custom http client was not used")
	}
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

type RequestCounter struct {
	Count int
}

func (r *RequestCounter) RoundTrip(req *http.Request) (*http.Response, error) {
	r.Count++
	return http.DefaultTransport.RoundTrip(req)
}
