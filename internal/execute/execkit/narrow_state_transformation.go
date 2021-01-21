package execkit

import (
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/execute/table"
)

// NarrowStateTransformation is the same as a NarrowTransformation
// except that it retains state between processing buffers.
type NarrowStateTransformation interface {
	// Process will process the TableView.
	Process(view table.View, state interface{}, d *Dataset, mem memory.Allocator) (interface{}, bool, error)
}

type narrowStateTransformation struct {
	t NarrowStateTransformation
}

// NewNarrowStateTransformation constructs a Transformation and Dataset
// using the NarrowStateTransformation implementation.
func NewNarrowStateTransformation(id DatasetID, t NarrowStateTransformation, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	tr := &narrowStateTransformation{t: t}
	return NewNarrowTransformation(id, tr, mem)
}

func (n *narrowStateTransformation) Process(view table.View, d *Dataset, mem memory.Allocator) error {
	state, _ := d.Lookup(view.Key())
	if ns, ok, err := n.t.Process(view, state, d, d.mem); err != nil {
		return err
	} else if ok {
		d.Set(view.Key(), ns)
	}
	return nil
}

func (n *narrowStateTransformation) OperationType() string {
	return execute.OperationType(n.t)
}
