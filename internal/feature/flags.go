// Code generated by the feature package; DO NOT EDIT.

package feature

import (
	"context"

	"github.com/influxdata/flux/internal/pkg/feature"
)

type (
	Flag       = feature.Flag
	Flagger    = feature.Flagger
	StringFlag = feature.StringFlag
	FloatFlag  = feature.FloatFlag
	IntFlag    = feature.IntFlag
	BoolFlag   = feature.BoolFlag
)

var aggregateTransformationTransport = feature.MakeBoolFlag(
	"Aggregate Transformation Transport",
	"aggregateTransformationTransport",
	"Jonathan Sternberg",
	false,
)

// AggregateTransformationTransport - Enable Transport interface for AggregateTransformation
func AggregateTransformationTransport() BoolFlag {
	return aggregateTransformationTransport
}

var groupTransformationGroup = feature.MakeBoolFlag(
	"Group Transformation Group",
	"groupTransformationGroup",
	"Sean Brickley",
	false,
)

// GroupTransformationGroup - Enable GroupTransformation interface for the group function
func GroupTransformationGroup() BoolFlag {
	return groupTransformationGroup
}

var queryConcurrencyLimit = feature.MakeIntFlag(
	"Query Concurrency Limit",
	"queryConcurrencyLimit",
	"Jonathan Sternberg",
	0,
)

// QueryConcurrencyLimit - Sets the query concurrency limit for the planner
func QueryConcurrencyLimit() IntFlag {
	return queryConcurrencyLimit
}

var optimizeUnionTransformation = feature.MakeBoolFlag(
	"Optimize Union Transformation",
	"optimizeUnionTransformation",
	"Jonathan Sternberg",
	false,
)

// OptimizeUnionTransformation - Optimize the union transformation
func OptimizeUnionTransformation() BoolFlag {
	return optimizeUnionTransformation
}

var mqttPoolDialer = feature.MakeBoolFlag(
	"MQTT Pool Dialer",
	"mqttPoolDialer",
	"Jonathan Sternberg",
	false,
)

// MqttPoolDialer - MQTT pool dialer
func MqttPoolDialer() BoolFlag {
	return mqttPoolDialer
}

var vectorizedMap = feature.MakeBoolFlag(
	"Vectorized Map",
	"vectorizedMap",
	"Jonathan Sternberg",
	false,
)

// VectorizedMap - Enables the version of map that supports vectorized functions
func VectorizedMap() BoolFlag {
	return vectorizedMap
}

// Inject will inject the Flagger into the context.
func Inject(ctx context.Context, flagger Flagger) context.Context {
	return feature.Inject(ctx, flagger)
}

var all = []Flag{
	aggregateTransformationTransport,
	groupTransformationGroup,
	queryConcurrencyLimit,
	optimizeUnionTransformation,
	mqttPoolDialer,
	vectorizedMap,
}

var byKey = map[string]Flag{
	"aggregateTransformationTransport": aggregateTransformationTransport,
	"groupTransformationGroup":         groupTransformationGroup,
	"queryConcurrencyLimit":            queryConcurrencyLimit,
	"optimizeUnionTransformation":      optimizeUnionTransformation,
	"mqttPoolDialer":                   mqttPoolDialer,
	"vectorizedMap":                    vectorizedMap,
}

// Flags returns all feature flags.
func Flags() []Flag {
	return all
}

// ByKey returns the Flag corresponding to the given key.
func ByKey(k string) (Flag, bool) {
	v, found := byKey[k]
	return v, found
}

type Metrics = feature.Metrics

// SetMetrics sets the metric store for feature flags.
func SetMetrics(m Metrics) {
	feature.SetMetrics(m)
}
