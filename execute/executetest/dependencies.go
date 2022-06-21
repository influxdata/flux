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
		feature.Dependency{
			Flagger: TestFlagger(testFlags),
		},
	}
}

var testFlags = map[string]interface{}{
	// "aggregateTransformationTransport": true,
	// "groupTransformationGroup":         true,
	// "optimizeUnionTransformation": true,
	"vectorizedMap":             true,
	"vectorizeLogicalOperators": true,
	"optimizeAggregateWindow":   true,
	"narrowTransformationLimit": true,
	"optimizeStateTracking":     true,
	"optimizeSetTransformation": true,
}

type TestFlagger map[string]interface{}

func (t TestFlagger) FlagValue(ctx context.Context, flag feature.Flag) interface{} {
	v, ok := t[flag.Key()]
	if !ok {
		return flag.Default()
	}
	return v
}
