package execkit

import (
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/execute/table"
)

// GroupTransformation is a transformation that can modify the group key.
// Other than modifying the group key, it is identical to a NarrowTransformation.
//
// The main difference between this and NarrowTransformation is that
// NarrowTransformation will pass the FlushKeyMsg to the Dataset
// and GroupTransformation will swallow this Message.
type GroupTransformation interface {
	// Process will process the TableView.
	Process(view table.View, d *Dataset, mem memory.Allocator) error
}

type groupTransformation struct {
	narrowTransformation
}

// NewGroupTransformation constructs a Transformation and Dataset
// using the GroupTransformation implementation.
func NewGroupTransformation(id DatasetID, t GroupTransformation, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	tr := &groupTransformation{
		narrowTransformation: narrowTransformation{
			t: t,
			d: NewDataset(id, mem),
		},
	}
	return tr, tr.d, nil
}

// ProcessMessage will process the incoming message.
func (n *groupTransformation) ProcessMessage(m execute.Message) error {
	if _, ok := m.(execute.FlushKeyMsg); ok {
		m.Ack()
		return nil
	}
	return n.narrowTransformation.ProcessMessage(m)
}

// Process is implemented to remain compatible with legacy upstreams.
// It converts the incoming stream into a set of appropriate messages.
func (n *groupTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	return tbl.Do(func(cr flux.ColReader) error {
		view := table.ViewFromReader(cr)
		view.Retain()
		m := processViewMsg{
			srcMessage: srcMessage(id),
			view:       view,
		}
		return n.ProcessMessage(&m)
	})
}
