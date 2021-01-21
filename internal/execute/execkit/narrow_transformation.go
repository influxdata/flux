package execkit

import (
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/execute/table"
)

// NarrowTransformation implements a transformation that processes
// a TableView and does not modify its group key.
type NarrowTransformation interface {
	// Process will process the TableView.
	Process(view table.View, d *Dataset, mem memory.Allocator) error
}

type narrowTransformation struct {
	t NarrowTransformation
	d *Dataset
}

// NewNarrowTransformation constructs a Transformation and Dataset
// using the NarrowTransformation implementation.
func NewNarrowTransformation(id DatasetID, t NarrowTransformation, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	tr := &narrowTransformation{
		t: t,
		d: NewDataset(id, mem),
	}
	return tr, tr.d, nil
}

// ProcessMessage will process the incoming message.
func (n *narrowTransformation) ProcessMessage(m execute.Message) error {
	defer m.Ack()

	switch m := m.(type) {
	case execute.FinishMsg:
		n.Finish(m.SrcDatasetID(), m.Error())
		return nil
	case execute.ProcessViewMsg:
		return n.t.Process(m.View(), n.d, n.d.mem)
	case execute.FlushKeyMsg:
		return n.d.FlushKey(m.Key())
	case execute.ProcessMsg:
		return n.Process(m.SrcDatasetID(), m.Table())
	}
	return nil
}

// Process is implemented to remain compatible with legacy upstreams.
// It converts the incoming stream into a set of appropriate messages.
func (n *narrowTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	if err := tbl.Do(func(cr flux.ColReader) error {
		view := table.ViewFromReader(cr)
		view.Retain()
		m := processViewMsg{
			srcMessage: srcMessage(id),
			view:       view,
		}
		return n.ProcessMessage(&m)
	}); err != nil {
		return err
	}

	m := flushKeyMsg{
		srcMessage: srcMessage(id),
		key:        tbl.Key(),
	}
	return n.ProcessMessage(&m)
}

// Finish is implemented to remain compatible with legacy upstreams.
func (n *narrowTransformation) Finish(id execute.DatasetID, err error) {
	if err != nil {
		_ = n.d.Abort(err)
		return
	}
	_ = n.d.Close()
}

func (n *narrowTransformation) OperationType() string {
	return execute.OperationType(n.t)
}
func (n *narrowTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return nil
}
func (n *narrowTransformation) UpdateWatermark(id execute.DatasetID, t execute.Time) error {
	return nil
}
func (n *narrowTransformation) UpdateProcessingTime(id execute.DatasetID, t execute.Time) error {
	return nil
}
