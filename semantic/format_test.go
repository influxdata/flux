package semantic_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
)

func TestFormatted(t *testing.T) {
	type testcase struct {
		name string
		flux string
		want string
	}

	defs := []string{`r = {_measurement: "cpu", _field: "usage_system"}`,
		`i = 1`,
		`j = 1.0`,
	}
	tcs := []testcase{
		{
			name: "filter expression",
			flux: `r._measurement == "cpu" and r._field != "usage_system"`,
			want: `r._measurement == "cpu" and r._field != "usage_system"`,
		},
		{
			name: "arithmetic expression multiply/divide",
			flux: `i * 3 > 0 and j / 7.0 >= 0`,
			want: `i * 3 > 0 and j / 7.000000 >= 0`,
		},
		{
			name: "arithmetic expression plus/minus",
			flux: `i + 3 < 0 and i - 7 <= 37`,
			want: `i + 3 < 0 and i - 7 <= 37`,
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			semPkg, err := runtime.AnalyzeSource(strings.Join(defs, "\n") + tc.flux)
			if err != nil {
				t.Fatal(err)
			}
			semExpr := semPkg.Files[0].Body[len(defs)].(*semantic.ExpressionStatement).Expression
			got := fmt.Sprintf("%v", semantic.Formatted(semExpr))
			if tc.want != got {
				t.Fatalf("unexpected output: -want/+got:\n- %v\n+ %v", tc.want, got)
			}
		})
	}
}
