package semantic_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
)

func TestSymbol(t *testing.T) {
	tcs := []struct {
		name    string
		fluxSrc string
		err     error
	}{
		{
			name: "isType",
			fluxSrc: `
				import "types"
				x = types.isType(v: 1, type: "string")
            `,
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			pkg, err := runtime.AnalyzeSource(ctx, tc.fluxSrc)
			if err != nil {
				t.Fatal(err)
			}

			call := pkg.Files[0].Body[0].(*semantic.NativeVariableAssignment).Init.(*semantic.CallExpression)
			property := call.Callee.(*semantic.MemberExpression).Property
			if property != semantic.NewSymbol("isType@types") {
				t.Fatalf("Expected the `isType` call to resolve to the module type: got:\n%v", property)
			}
		})
	}
}
