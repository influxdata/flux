package universe

import (
	"context"

	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
)

// TODO(jsternberg): This is exposed so the tests have access
// to the new transformation. This method should not be used
// externally as the new pivot is not meant to be exposed publicly
// at the moment and the name will probably change when it is exposed.
func NewSortedPivotTransformation(ctx context.Context, spec SortedPivotProcedureSpec, id execute.DatasetID, alloc *memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	return newSortedPivotTransformation(ctx, spec, id, alloc)
}
