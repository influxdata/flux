package execute

import (
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute/table"
)

// NarrowTransformation implements a transformation that processes
// a table.Chunk and does not modify its group key.
type NarrowTransformation interface {
	// Process will process the table.Chunk and send any output to the TransportDataset.
	Process(chunk table.Chunk, d *TransportDataset, mem memory.Allocator) error

	Disposable
}

var _ Transport = (*narrowTransformation)(nil)
var _ Transformation = (*narrowTransformation)(nil)

type narrowTransformation struct {
	t NarrowTransformation
	d *TransportDataset
}

// NewNarrowTransformation constructs a Transformation and Dataset
// using the NarrowTransformation implementation.
func NewNarrowTransformation(id DatasetID, t NarrowTransformation, mem memory.Allocator) (Transformation, Dataset, error) {
	tr := &narrowTransformation{
		t: t,
		d: NewTransportDataset(id, mem),
	}
	return tr, tr.d, nil
}

// ProcessMessage will process the incoming message.
func (n *narrowTransformation) ProcessMessage(m Message) error {
	defer m.Ack()

	switch m := m.(type) {
	case FinishMsg:
		n.Finish(m.SrcDatasetID(), m.Error())
		return nil
	case ProcessChunkMsg:
		return n.t.Process(m.TableChunk(), n.d, n.d.mem)
	case FlushKeyMsg:
		return n.d.FlushKey(m.Key())
	case ProcessMsg:
		return n.Process(m.SrcDatasetID(), m.Table())
	}
	return nil
}

// Process is implemented to remain compatible with legacy upstreams.
// It converts the incoming stream into a set of appropriate messages.
func (n *narrowTransformation) Process(id DatasetID, tbl flux.Table) error {
	if err := tbl.Do(func(cr flux.ColReader) error {
		chunk := table.ChunkFromReader(cr)
		chunk.Retain()
		m := processChunkMsg{
			srcMessage: srcMessage(id),
			chunk:      chunk,
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
func (n *narrowTransformation) Finish(id DatasetID, err error) {
	n.d.Finish(err)
	n.t.Dispose()
}

func (n *narrowTransformation) OperationType() string {
	return OperationType(n.t)
}
func (n *narrowTransformation) RetractTable(id DatasetID, key flux.GroupKey) error {
	return nil
}
func (n *narrowTransformation) UpdateWatermark(id DatasetID, t Time) error {
	return nil
}
func (n *narrowTransformation) UpdateProcessingTime(id DatasetID, t Time) error {
	return nil
}
