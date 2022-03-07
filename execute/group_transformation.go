package execute

import (
	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux/execute/table"
)

// GroupTransformation is a transformation that can modify the group key.
// Other than modifying the group key, it is identical to a NarrowTransformation.
//
// The main difference between this and NarrowTransformation is that
// NarrowTransformation will pass the FlushKeyMsg to the Dataset
// and GroupTransformation will swallow this Message.
type GroupTransformation interface {
	Process(chunk table.Chunk, d *TransportDataset, mem memory.Allocator) error

	Closer
}

var _ Transport = (*groupTransformation)(nil)

type groupTransformation struct {
	t GroupTransformation
	d *TransportDataset
}

func (g *groupTransformation) OperationType() string {
	return OperationType(g.t)
}

func NewGroupTransformation(id DatasetID, t GroupTransformation, mem memory.Allocator) (Transformation, Dataset, error) {
	g := &groupTransformation{
		t: t,
		d: NewTransportDataset(id, mem),
	}

	return NewTransformationFromTransport(g), g.d, nil
}

// ProcessMessage will process the incoming message.
func (g *groupTransformation) ProcessMessage(m Message) error {
	defer m.Ack()

	switch m := m.(type) {
	case FinishMsg:
		g.Finish(m.SrcDatasetID(), m.Error())
		return nil
	case ProcessChunkMsg:
		return g.t.Process(m.TableChunk(), g.d, g.d.mem)
	case FlushKeyMsg:
		return nil
	case ProcessMsg:
		panic("unreachable")
	}
	return nil
}

func (g *groupTransformation) Finish(id DatasetID, err error) {
	err = Close(err, g.t)
	g.d.Finish(err)
}
