package v1_test

import (
	"net/url"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb/internal/testutil"
	v1 "github.com/influxdata/flux/stdlib/influxdata/influxdb/v1"
)

func TestDatabases_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "databases no args",
			Raw: `import "influxdata/influxdb/v1"
v1.databases()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID:   "databases0",
						Spec: &v1.DatabasesOpSpec{},
					},
				},
			},
		},
		{
			Name: "databases unexpected arg",
			Raw: `import "influxdata/influxdb/v1"
v1.databases(chicken:"what is this?")`,
			WantErr: true,
		},
		{
			Name: "databases with host and token",
			Raw: `import "influxdata/influxdb/v1"
v1.databases(host: "http://localhost:9999", token: "mytoken")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "databases0",
						Spec: &v1.DatabasesOpSpec{
							Host:  stringPtr("http://localhost:9999"),
							Token: stringPtr("mytoken"),
						},
					},
				},
			},
		},
		{
			Name: "databases with org",
			Raw: `import "influxdata/influxdb/v1"
v1.databases(org: "influxdata")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "databases0",
						Spec: &v1.DatabasesOpSpec{
							Org: &influxdb.NameOrID{Name: "influxdata"},
						},
					},
				},
			},
		},
		{
			Name: "databases with org id",
			Raw: `import "influxdata/influxdb/v1"
v1.databases(orgID: "97aa81cc0e247dc4")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "databases0",
						Spec: &v1.DatabasesOpSpec{
							Org: &influxdb.NameOrID{ID: "97aa81cc0e247dc4"},
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

func TestDatabases_Run(t *testing.T) {
	defaultTablesFn := func() []*executetest.Table {
		return []*executetest.Table{{
			KeyCols: []string{"organizationID"},
			ColMeta: []flux.ColMeta{
				{Label: "organizationID", Type: flux.TString},
				{Label: "databaseName", Type: flux.TString},
				{Label: "retentionPolicy", Type: flux.TString},
				{Label: "retentionPeriod", Type: flux.TInt},
				{Label: "default", Type: flux.TBool},
				{Label: "bucketID", Type: flux.TString},
			},
			Data: [][]interface{}{
				{"97aa81cc0e247dc4", "telegraf", "autogen", int64(0), true, "1e01ac57da723035"},
			},
		}}
	}

	for _, tt := range []struct {
		name string
		spec *v1.DatabasesRemoteProcedureSpec
		want testutil.Want
	}{
		{
			name: "basic query",
			spec: &v1.DatabasesRemoteProcedureSpec{
				DatabasesProcedureSpec: &v1.DatabasesProcedureSpec{
					Org:   &influxdb.NameOrID{Name: "influxdata"},
					Token: stringPtr("mytoken"),
				},
			},
			want: testutil.Want{
				Params: url.Values{
					"org": []string{"influxdata"},
				},
				Ast: &ast.Package{
					Package: "main",
					Files: []*ast.File{{
						Name: "query.flux",
						Package: &ast.PackageClause{
							Name: &ast.Identifier{Name: "main"},
						},
						Imports: []*ast.ImportDeclaration{{
							Path: &ast.StringLiteral{Value: "influxdata/influxdb/v1"},
						}},
						Body: []ast.Statement{
							&ast.ExpressionStatement{
								Expression: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "v1"},
										Property: &ast.Identifier{Name: "databases"},
									},
								},
							},
						},
					}},
				},
				Tables: defaultTablesFn,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			testutil.RunSourceTestHelper(t, tt.spec, tt.want)
		})
	}
}

func TestDatabases_Run_Errors(t *testing.T) {
	testutil.RunSourceErrorTestHelper(t, &v1.DatabasesRemoteProcedureSpec{
		DatabasesProcedureSpec: &v1.DatabasesProcedureSpec{
			Org:   &influxdb.NameOrID{Name: "influxdata"},
			Token: stringPtr("mytoken"),
		},
	})
}

func TestDatabases_URLValidator(t *testing.T) {
	testutil.RunSourceURLValidatorTestHelper(t, &v1.DatabasesRemoteProcedureSpec{
		DatabasesProcedureSpec: &v1.DatabasesProcedureSpec{
			Org:   &influxdb.NameOrID{Name: "influxdata"},
			Token: stringPtr("mytoken"),
		},
	})
}

func TestDatabases_HTTPClient(t *testing.T) {
	testutil.RunSourceHTTPClientTestHelper(t, &v1.DatabasesRemoteProcedureSpec{
		DatabasesProcedureSpec: &v1.DatabasesProcedureSpec{
			Org:   &influxdb.NameOrID{Name: "influxdata"},
			Token: stringPtr("mytoken"),
		},
	})
}

func stringPtr(v string) *string {
	return &v
}
