package querytest

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/plan"
)

// Some transformations need to take a URL e.g. sql.to, kafka.to
// the URL/DSN supplied by the user need to be validated by a URLValidator{}
// before we can establish the connection.
// TfUrlValidationTestCase, TfUrlValidationTest (as well as the Run() method)
// acts as a test harness for that.

type TfUrlValidationTestCase struct {
	Name      string
	Spec      plan.ProcedureSpec
	Validator url.Validator
	WantErr   string
}

type TfUrlValidationTest struct {
	CreateFn CreateNewTransformationWithDeps
	Cases    []TfUrlValidationTestCase
}

// sql.createToSQLTransformation() and kafka.createToKafkaTransformation() converts plan.ProcedureSpec
// to their struct implementations ToSQLProcedureSpec and ToKafkaProcedureSpec respectively.
// This complicated the test harness requiring us to provide CreateNewTransformationWithDeps
// functions to do the plan.ProcedureSpec conversion and call the subsequent factory method
// namely: kafka.NewToKafkaTransformation() and sql.NewToSQLTransformation()
// See also: sql/to_test.go/TestToSql_NewTransformation and kafka/to_test.go/TestToKafka_NewTransformation
type CreateNewTransformationWithDeps func(d execute.Dataset, deps flux.Dependencies,
	cache execute.TableBuilderCache, spec plan.ProcedureSpec) (execute.Transformation, error)

func (test *TfUrlValidationTest) Run(t *testing.T) {
	for _, tc := range test.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			d := executetest.NewDataset(executetest.RandomDatasetID())
			c := execute.NewTableBuilderCache(executetest.UnlimitedAllocator)
			deps := dependenciestest.Default()
			if tc.Validator != nil {
				deps.Deps.URLValidator = tc.Validator
			}
			_, err := test.CreateFn(d, deps, c, tc.Spec)
			if err != nil {
				if tc.WantErr != "" {
					got := err.Error()
					if !strings.Contains(got, tc.WantErr) {
						t.Fatalf("unexpected result -want/+got:\n%s",
							cmp.Diff(got, tc.WantErr))
					}
					return
				} else {
					t.Fatal(err)
				}
			}
		})
	}
}
