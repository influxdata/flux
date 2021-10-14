package execute

import (
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
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

	Disposable
}

var _ Transport = (*groupTransformation)(nil)
var _ Transformation = (*groupTransformation)(nil)

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

	return g, g.d, nil
}

// Implement the Transport interface
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
		return g.Process(m.SrcDatasetID(), m.Table())
	}
	return nil
}

func (g *groupTransformation) Process(id DatasetID, tbl flux.Table) error {
	if err := tbl.Do(func(cr flux.ColReader) error {
		chunk := table.ChunkFromReader(cr)
		chunk.Retain()
		m := processChunkMsg{
			srcMessage: srcMessage(id),
			chunk:      chunk,
		}
		return g.ProcessMessage(&m)
	}); err != nil {
		return err
	}

	m := flushKeyMsg{
		srcMessage: srcMessage(id),
		key:        tbl.Key(),
	}
	return g.ProcessMessage(&m)
}

func (g *groupTransformation) Finish(id DatasetID, err error) {
	g.d.Finish(err)
	g.t.Dispose()
}

func (g *groupTransformation) RetractTable(id DatasetID, key flux.GroupKey) error {
	return nil
}

func (g *groupTransformation) UpdateWatermark(id DatasetID, t Time) error {
	return nil
}

func (g *groupTransformation) UpdateProcessingTime(id DatasetID, t Time) error {
	return nil
}
