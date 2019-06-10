package flux_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/pkg/errors"
)

func mustGetSampleQuery() *semantic.Package {
	q := `import g "generate"
g.from(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:53:00Z, count: 10, fn: (n) => n)`
	astPkg, err := flux.Parse(q)
	if err != nil {
		panic(errors.Wrap(err, "cannot compile simple query"))
	}
	semPkg, err := semantic.New(astPkg)
	if err != nil {
		panic(errors.Wrap(err, "cannot compile simple query"))
	}
	return semPkg
}

func Test_EvalWithNow(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	tcs := []struct {
		name        string
		paramNow    time.Time
		optNow      time.Time
		expectedNow time.Time
	}{
		{
			name:        "no option set",
			paramNow:    now,
			expectedNow: now,
		},
		{
			name:        "option set",
			paramNow:    now,
			optNow:      tomorrow,
			expectedNow: tomorrow,
		},
		{
			name:        "zero param with opt",
			paramNow:    time.Time{},
			optNow:      tomorrow,
			expectedNow: tomorrow,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			q := mustGetSampleQuery()
			if !tc.optNow.IsZero() {
				nowOpt := &semantic.OptionStatement{
					Assignment: &semantic.NativeVariableAssignment{
						Identifier: &semantic.Identifier{
							Name: "now",
						},
						Init: &semantic.FunctionExpression{
							Block: &semantic.FunctionBlock{
								Body: &semantic.DateTimeLiteral{
									Value: tc.optNow,
								},
							},
						},
					},
				}
				q.Files[0].Body = append([]semantic.Statement{nowOpt}, q.Files[0].Body...)
			}
			_, _, gotNow, err := flux.EvalWithNow(q, tc.paramNow)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.expectedNow, gotNow); diff != "" {
				t.Fatalf("got unexpected now time: -want/+got:\n%s", diff)
			}
		})
	}
}

func Test_EvalWithNow_Zero(t *testing.T) {
	now := time.Now()
	q := mustGetSampleQuery()
	_, _, gotNow, err := flux.EvalWithNow(q, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if !now.Before(gotNow) {
		t.Fatalf("query executed with wrong now: %v", gotNow)
	}
}
