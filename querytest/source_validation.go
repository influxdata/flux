package querytest

import (
	"context"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/mock"
	"github.com/influxdata/flux/plan"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Some sources are located by a URL. e.g. sql.from, socket.from
// the URL/DSN supplied by the user need to be validated by a URLValidator{}
// before we can establish the connection.
// This struct (as well as the Run() method) acts as a test harness for that.
type SourceUrlValidationTestCases []struct {
	Name   string
	Spec   plan.ProcedureSpec
	V      url.Validator
	ErrMsg string
}

func (testCases *SourceUrlValidationTestCases) Run(t *testing.T, fn execute.CreateNewPlannerSource) {
	for _, tc := range *testCases {
		deps := dependenciestest.Default()
		if tc.V != nil {
			deps.Deps.URLValidator = tc.V
		}
		ctx := deps.Inject(context.Background())
		a := mock.AdministrationWithContext(ctx)
		t.Run(tc.Name, func(t *testing.T) {
			id := executetest.RandomDatasetID()
			_, err := fn(tc.Spec, id, a)
			if tc.ErrMsg != "" {
				if err == nil {
					t.Errorf("Expect an error with message \"%s\", but did not get one.", tc.ErrMsg)
				} else {
					assert.Contains(t, err.Error(), tc.ErrMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect to get an error, but got %v", err)
				}
			}
		})
	}
}
