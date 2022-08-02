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
	true,
)

// GroupTransformationGroup - Enable GroupTransformation interface for the group function
func GroupTransformationGroup() BoolFlag {
	return groupTransformationGroup
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

var narrowTransformationDifference = feature.MakeBoolFlag(
	"Narrow Transformation Difference",
	"narrowTransformationDifference",
	"Markus Westerlind",
	false,
)

// NarrowTransformationDifference - Enable the NarrowTransformation implementation of difference
func NarrowTransformationDifference() BoolFlag {
	return narrowTransformationDifference
}

var narrowTransformationFill = feature.MakeBoolFlag(
	"Narrow Transformation Fill",
	"narrowTransformationFill",
	"Sunil Kartikey",
	false,
)

// NarrowTransformationFill - Enable the NarrowTransformation implementation of Fill
func NarrowTransformationFill() BoolFlag {
	return narrowTransformationFill
}

var optimizeAggregateWindow = feature.MakeBoolFlag(
	"Optimize Aggregate Window",
	"optimizeAggregateWindow",
	"Jonathan Sternberg",
	true,
)

// OptimizeAggregateWindow - Enables a version of aggregateWindow written in Go
func OptimizeAggregateWindow() BoolFlag {
	return optimizeAggregateWindow
}

var labelPolymorphism = feature.MakeBoolFlag(
	"Label polymorphism",
	"labelPolymorphism",
	"Markus Westerlind",
	false,
)

// LabelPolymorphism - Enables label polymorphism in the type system
func LabelPolymorphism() BoolFlag {
	return labelPolymorphism
}

var optimizeSetTransformation = feature.MakeBoolFlag(
	"Optimize Set Transformation",
	"optimizeSetTransformation",
	"Jonathan Sternberg",
	false,
)

// OptimizeSetTransformation - Enables a version of set that is optimized
func OptimizeSetTransformation() BoolFlag {
	return optimizeSetTransformation
}

var unusedSymbolWarnings = feature.MakeBoolFlag(
	"Unused Symbol Warnings",
	"unusedSymbolWarnings",
	"Markus Westerlind",
	false,
)

// UnusedSymbolWarnings - Enables warnings for unused symbols
func UnusedSymbolWarnings() BoolFlag {
	return unusedSymbolWarnings
}

var experimentalTestingDiff = feature.MakeBoolFlag(
	"Experimental Testing Diff",
	"experimentalTestingDiff",
	"Jonathan Sternberg",
	false,
)

// ExperimentalTestingDiff - Switches testing.diff to use experimental.diff
func ExperimentalTestingDiff() BoolFlag {
	return experimentalTestingDiff
}

var removeRedundantSortNodes = feature.MakeBoolFlag(
	"Remove Redundant Sort Nodes",
	"removeRedundantSortNodes",
	"Chris Wolff",
	false,
)

// RemoveRedundantSortNodes - Planner will remove sort nodes when tables are already sorted
func RemoveRedundantSortNodes() BoolFlag {
	return removeRedundantSortNodes
}

var queryConcurrencyIncrease = feature.MakeIntFlag(
	"Query Concurrency Increase",
	"queryConcurrencyIncrease",
	"Jonathan Sternberg, Adrian Thurston",
	0,
)

// QueryConcurrencyIncrease - Additional dispatcher workers to allocate on top of the minimimum allowable computed by the engine
func QueryConcurrencyIncrease() IntFlag {
	return queryConcurrencyIncrease
}

var vectorizedConditionals = feature.MakeBoolFlag(
	"Vectorized Conditionals",
	"vectorizedConditionals",
	"Owen Nelson",
	false,
)

// VectorizedConditionals - Calls to map can be vectorized when conditional expressions appear in the function
func VectorizedConditionals() BoolFlag {
	return vectorizedConditionals
}

var vectorizedEqualityOps = feature.MakeBoolFlag(
	"Vectorized Equality Ops",
	"vectorizedEqualityOps",
	"Owen Nelson",
	false,
)

// VectorizedEqualityOps - Calls to map can be vectorized when conditional expressions appear in the function
func VectorizedEqualityOps() BoolFlag {
	return vectorizedEqualityOps
}

// Inject will inject the Flagger into the context.
func Inject(ctx context.Context, flagger Flagger) context.Context {
	return feature.Inject(ctx, flagger)
}

var all = []Flag{
	aggregateTransformationTransport,
	groupTransformationGroup,
	optimizeUnionTransformation,
	vectorizedMap,
	narrowTransformationDifference,
	narrowTransformationFill,
	optimizeAggregateWindow,
	labelPolymorphism,
	optimizeSetTransformation,
	unusedSymbolWarnings,
	experimentalTestingDiff,
	removeRedundantSortNodes,
	queryConcurrencyIncrease,
	vectorizedConditionals,
	vectorizedEqualityOps,
}

var byKey = map[string]Flag{
	"aggregateTransformationTransport": aggregateTransformationTransport,
	"groupTransformationGroup":         groupTransformationGroup,
	"optimizeUnionTransformation":      optimizeUnionTransformation,
	"vectorizedMap":                    vectorizedMap,
	"narrowTransformationDifference":   narrowTransformationDifference,
	"narrowTransformationFill":         narrowTransformationFill,
	"optimizeAggregateWindow":          optimizeAggregateWindow,
	"labelPolymorphism":                labelPolymorphism,
	"optimizeSetTransformation":        optimizeSetTransformation,
	"unusedSymbolWarnings":             unusedSymbolWarnings,
	"experimentalTestingDiff":          experimentalTestingDiff,
	"removeRedundantSortNodes":         removeRedundantSortNodes,
	"queryConcurrencyIncrease":         queryConcurrencyIncrease,
	"vectorizedConditionals":           vectorizedConditionals,
	"vectorizedEqualityOps":            vectorizedEqualityOps,
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
