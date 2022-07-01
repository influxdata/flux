package execute

import (
	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux/execute/table"
)

// NarrowTransformation implements a transformation that processes
// a table.Chunk and does not modify its group key.
type NarrowTransformation interface {
	// Process will process the table.Chunk and send any output to the TransportDataset.
	Process(chunk table.Chunk, d *TransportDataset, mem memory.Allocator) error

	Closer
}

var _ Transport = (*narrowTransformation)(nil)

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
	return NewTransformationFromTransport(tr), tr.d, nil
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
		panic("unreachable")
	}
	return nil
}

// Finish is implemented to remain compatible with legacy upstreams.
func (n *narrowTransformation) Finish(id DatasetID, err error) {
	err = Close(err, n.t)
	n.d.Finish(err)
}

func (n *narrowTransformation) OperationType() string {
	return OperationType(n.t)
}
