package universe

import (
	"context"

	"github.com/InfluxCommunity/flux/execute"
	"github.com/apache/arrow/go/v7/arrow/memory"
)

func NewMapTransformation(ctx context.Context, id execute.DatasetID, spec *MapProcedureSpec, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	return newMapTransformation(ctx, id, spec, mem)
}
