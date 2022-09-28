package executetest

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependencies/feature"
)

type dependencyList []flux.Dependency

func (d dependencyList) Inject(ctx context.Context) context.Context {
	for _, dep := range d {
		ctx = dep.Inject(ctx)
	}
	return ctx
}

func NewTestExecuteDependencies() flux.Dependency {
	return dependencyList{
		dependenciestest.Default(),
		NewDefaultTestFlagger(),
	}
}

// NewDefaultTestFlagger gives a flagger dependency for a test harnesses to use as a baseline.
//
// Likely this will be made redundant by <https://github.com/influxdata/flux/issues/4777>
// since testcases will then be able to manage their own feature selection.
func NewDefaultTestFlagger() feature.Dependency {
	return feature.Dependency{
		Flagger: TestFlagger(testFlags),
	}
}

var testFlags = map[string]interface{}{
	// "aggregateTransformationTransport": true,
	// "groupTransformationGroup":         true,
	// "optimizeUnionTransformation": true,
	"labelPolymorphism":         true,
	"vectorizedMap":             true,
	"vectorizedConditionals":    true,
	"vectorizedFloat":           true,
	"vectorizeLogicalOperators": true,
	"vectorizedEqualityOps":     true,
	"vectorizedUnaryOps":        true,
	"optimizeAggregateWindow":   true,
	"optimizeStateTracking":     true,
	"optimizeSetTransformation": true,
	"removeRedundantSortNodes":  true,
	"strictNullLogicalOps":      true,
}

type TestFlagger map[string]interface{}

func (t TestFlagger) FlagValue(ctx context.Context, flag feature.Flag) interface{} {
	v, ok := t[flag.Key()]
	if !ok {
		return flag.Default()
	}
	return v
}
