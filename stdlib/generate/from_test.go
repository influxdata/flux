package generate_test

import (
	"testing"
	"time"

	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/execute/executetest"
	_ "github.com/InfluxCommunity/flux/fluxinit/static"
	"github.com/InfluxCommunity/flux/internal/operation"
	"github.com/InfluxCommunity/flux/interpreter"
	"github.com/InfluxCommunity/flux/querytest"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/stdlib/generate"
	"github.com/InfluxCommunity/flux/values/valuestest"
)

func TestFrom_NewQuery(t *testing.T) {
	// create importer
	importer := runtime.StdLib()
	pkg, err := importer.ImportPackageObject("generate")

	if err != nil {
		t.Fatal(err)
	}
	scope := valuestest.Scope()
	scope.Set("generate", pkg)

	tests := []querytest.NewQueryTestCase{
		{
			Name: "from with duration",
			Raw: ` import "generate"
					generate.from(start: 0h, stop: 1h, count: 10, fn: (n) => n)`,

			Want: &operation.Spec{
				Operations: []*operation.Node{
					{
						ID: "fromGenerator0",
						Spec: &generate.FromGeneratorOpSpec{
							Start: flux.Time{
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   1 * time.Hour,
								IsRelative: true,
							},
							Count: 10,
							Fn: interpreter.ResolvedFunction{
								Fn:    executetest.FunctionExpression(t, `(n) => n`),
								Scope: scope,
							},
						},
					},
				},
			},
		},
		{
			Name: "from with time",
			Raw: ` import "generate"
					generate.from(start: 2030-01-01T00:00:00Z, stop: 2030-01-01T00:00:01Z, count: 10, fn: (n) => n)`,

			Want: &operation.Spec{
				Operations: []*operation.Node{
					{
						ID: "fromGenerator0",
						Spec: &generate.FromGeneratorOpSpec{
							Start: flux.Time{
								IsRelative: false,
								Absolute:   time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
							},
							Stop: flux.Time{
								IsRelative: false,
								Absolute:   time.Date(2030, 1, 1, 0, 0, 1, 0, time.UTC),
							},
							Count: 10,
							Fn: interpreter.ResolvedFunction{
								Fn:    executetest.FunctionExpression(t, `(n) => n`),
								Scope: scope,
							},
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
