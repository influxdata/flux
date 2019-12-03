package universe

import (
	"context"

	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
)

// TODO(jsternberg): This is exposed so the tests have access
// to the new transformation. This method should not be used
// externally as the new pivot is not meant to be exposed publically
// at the moment and the name will probably change when it is exposed.
func NewPivotTransformation2(ctx context.Context, spec PivotProcedureSpec, id execute.DatasetID, alloc *memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	return newPivotTransformation2(ctx, spec, id, alloc)
}
