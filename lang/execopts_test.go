package lang

import (
	"context"
	"testing"
	"time"

	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

func TestExecutionOptions(t *testing.T) {
	src := `
		import "profiler"
		option profiler.enabledProfilers = [ "query", "operator" ]
		option now = () => ( 2020-10-15T00:00:00Z )
	`

	h, err := parser.ParseToHandle([]byte(src))
	if err != nil {
		t.Fatalf("failed to parse test case: %v", err)
	}

	prelude := []string{
		"universe",
		"influxdata/influxdb",
	}

	scope := values.NewScope()
	importer := runtime.StdLib()
	for _, p := range prelude {
		pkg, err := importer.ImportPackageObject(p)
		if err != nil {
			panic(err)
		}
		pkg.Range(scope.Set)
	}

	// Prepare a context with execution dependencies.
	ctx := context.TODO()
	deps := execute.DefaultExecutionDependencies()
	ctx = deps.Inject(ctx)

	// Pass lang.ExecutionOptions as the options configurator. It is
	// responsible for installing the configured options into the execution
	// dependencies. The goal of this test is to verify the option
	// configuration is called, and also that they are installed into the
	// execution environment.
	itrp := interpreter.NewInterpreter(nil, &ExecOptsConfig{})

	semPkg, err := runtime.AnalyzePackage(h)
	if err != nil {
		t.Fatalf("failed to evaluate test case: %v", err)
	}

	_, err = itrp.Eval(ctx, semPkg, scope, importer)
	if err != nil {
		t.Fatalf("failed to evaluate test case: %v", err)
	}

	// Verify Profilers was set.
	if deps.ExecutionOptions.Profilers == nil {
		t.Errorf("ProfilerNames was not configured")
	} else {
		if len(deps.ExecutionOptions.Profilers) != 2 {
			t.Errorf("ProfilerNames not did not contain expected number of elements")
		}
	}

	// Verify that the operator profiler was picked out.
	if deps.ExecutionOptions.OperatorProfiler == nil {
		t.Errorf("ProfilerNames was not configured")
	}

	// Verify that now was set.
	if deps.Now == nil {
		t.Errorf("Now was not configured")
	} else {
		expectedTime, _ := time.Parse(time.RFC3339, "2020-10-15T00:00:00Z")
		if *deps.Now != expectedTime {
			t.Errorf("now was set with the expected value, expected: %v got: %v", expectedTime, *deps.Now)
		}
	}
}
