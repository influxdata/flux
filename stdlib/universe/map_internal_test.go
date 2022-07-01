package universe

import (
	"context"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux/execute"
)

func NewMapTransformation2(ctx context.Context, id execute.DatasetID, spec *MapProcedureSpec, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	return newMapTransformation2(ctx, id, spec, mem)
}
