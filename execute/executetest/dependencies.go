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
			Flagger: testFlagger{},
		},
	}
}

var testFlags = map[string]interface{}{
	"narrowTransformationFilter":       true,
	"aggregateTransformationTransport": true,
	"optimizeDerivative":               true,
	"groupTransformationGroup":         true,
}

type testFlagger struct{}

func (t testFlagger) FlagValue(ctx context.Context, flag feature.Flag) interface{} {
	v, ok := testFlags[flag.Key()]
	if !ok {
		return flag.Default()
	}
	return v
}
