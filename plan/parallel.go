package plan

import (
	"fmt"
)

// Physical attributes used in specifying parallelization.

const ParallelRunKey = "parallel-run"

// ParallelRunAttribute means the node executes in parallel when present. It accepts parallel data (a subset of
// the source) and produces parallel data.
type ParallelRunAttribute struct {
	Factor int
}

var _ PhysicalAttr = ParallelRunAttribute{}

func (ParallelRunAttribute) Key() string { return ParallelRunKey }

// SuccessorsMustRequire implements the PhysicalAttribute interface.
// if a node produces parallel data, then all successors must require parallel
// data, otherwise there will be a plan error.
func (ParallelRunAttribute) SuccessorsMustRequire() bool {
	return true
}

func (a ParallelRunAttribute) SatisfiedBy(attr PhysicalAttr) bool {
	other, ok := attr.(ParallelRunAttribute)
	if !ok {
		return false
	}
	return a == other
}

func (a ParallelRunAttribute) String() string {
	return fmt.Sprintf("%v{Factor: %d}", ParallelRunKey, a.Factor)
}

const ParallelMergeKey = "parallel-merge"

// ParallelMergeAttribute means that the node accepts parallel data, merges the streams, and produces non-parallel
// data that covers the entire data source.
type ParallelMergeAttribute struct {
	Factor int
}

var _ PhysicalAttr = ParallelMergeAttribute{}

func (ParallelMergeAttribute) Key() string { return ParallelMergeKey }

func (ParallelMergeAttribute) SuccessorsMustRequire() bool {
	return false
}

func (a ParallelMergeAttribute) SatisfiedBy(attr PhysicalAttr) bool {
	other, ok := attr.(ParallelMergeAttribute)
	if !ok {
		return false
	}
	return a == other
}

func (a ParallelMergeAttribute) PlanDetails() string {
	return fmt.Sprintf("ParallelMergeFactor: %v", a.Factor)
}

func (a ParallelMergeAttribute) String() string {
	return fmt.Sprintf("%v{Factor: %d}", ParallelMergeKey, a.Factor)
}
