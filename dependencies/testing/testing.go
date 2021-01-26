package testing

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

type key int

const testingKey key = iota

// Inject will inject the testing dependencies into the context.
func Inject(ctx context.Context) context.Context {
	cfg := FrameworkConfig{}
	return cfg.Inject(ctx)
}

// FrameworkConfig is the testing framework configuration.
// This can be used to inject the testing framework as a dependency.
type FrameworkConfig struct{}

func (f FrameworkConfig) Inject(ctx context.Context) context.Context {
	return context.WithValue(ctx, testingKey, &testingFramework{})
}

// getTestingFramework will retrieve the testing framework from
// the context.Context.
func getTestingFramework(ctx context.Context) (*testingFramework, error) {
	tf := ctx.Value(testingKey)
	if tf == nil {
		return nil, errors.Newf(codes.Unimplemented, "testing framework not configured in this context")
	}
	return tf.(*testingFramework), nil
}

// MarkInvokedPlannerRule will mark that a planner rule was
// invoked and record that information in the testing dependencies.
//
// This method is a no-op if testing dependencies are not present.
func MarkInvokedPlannerRule(ctx context.Context, name string) {
	if tf, err := getTestingFramework(ctx); err == nil {
		if tf.got.plannerRules == nil {
			tf.got.plannerRules = make(map[string]int)
		}
		tf.got.plannerRules[name] += 1
	}
}

// Check will check that all testing expectations have been met.
// This is a no-op if testing dependencies are not present.
func Check(ctx context.Context) error {
	if tf, err := getTestingFramework(ctx); err == nil {
		return tf.Check()
	}
	return nil
}

// ExpectPlannerRule will mark that a planner rule is expected
// to be executed n number of times.
//
// This returns an error if testing dependencies have not been configured.
func ExpectPlannerRule(ctx context.Context, name string, n int) error {
	tf, err := getTestingFramework(ctx)
	if err != nil {
		return err
	}

	if tf.want.plannerRules == nil {
		tf.want.plannerRules = make(map[string]int)
	}
	tf.want.plannerRules[name] = n
	return nil
}

type testingFramework struct {
	want results
	got  results
}

func (tf *testingFramework) Check() error {
	return tf.want.Check(tf.got)
}

type results struct {
	plannerRules map[string]int
}

func (want results) Check(got results) error {
	for name, want := range want.plannerRules {
		got := got.plannerRules[name]
		if want != got {
			return errors.Newf(codes.Invalid, "planner rule invoked an unexpected number of times: %d (want) != %d (got)", want, got)
		}
	}
	return nil
}
