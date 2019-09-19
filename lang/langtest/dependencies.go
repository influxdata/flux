package langtest

import (
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
)

func DefaultExecutionDependencies() lang.ExecutionDependencies {
	return lang.ExecutionDependencies{
		Allocator: new(memory.Allocator),
	}
}
