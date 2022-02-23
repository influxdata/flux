package plan

// Physical attributes used in specifying parallelization. The Run attribute
// means the node executes in parallel. It accepts parallel data (a subset of
// the source) and produces parallel data. The merge attribute means that the
// node accepts parallel data, merges the streams, and produces non-parallel
// data that covers the entire data source.
const ParallelRunKey = "parallel-run"
const ParallelMergeKey = "parallel-merge"

type ParallelRunAttribute struct {
	Factor int
}

// If a node produces parallel data, then all successors must require parallel
// data, otherwise there will be a plan error.
func (ParallelRunAttribute) SuccessorsMustRequire() bool {
	return true
}

type ParallelMergeAttribute struct {
	Factor int
}

func (ParallelMergeAttribute) SuccessorsMustRequire() bool {
	return false
}
