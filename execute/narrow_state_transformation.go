package execute

import (
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute/table"
)

// NarrowStateTransformation is the same as a NarrowTransformation
// except that it retains state between processing buffers.
type NarrowStateTransformation interface {
	// Process will process the TableView.
	Process(chunk table.Chunk, state interface{}, d *TransportDataset, mem memory.Allocator) (interface{}, bool, error)

	Disposable
}

var _ Transport = (*narrowStateTransformation)(nil)
var _ Transformation = (*narrowStateTransformation)(nil)

type narrowStateTransformation struct {
	t NarrowStateTransformation
	d *TransportDataset
}

// NewNarrowStateTransformation constructs a Transformation and Dataset
// using the NarrowStateTransformation implementation.
func NewNarrowStateTransformation(id DatasetID, t NarrowStateTransformation, mem memory.Allocator) (Transformation, Dataset, error) {
	tr := &narrowStateTransformation{
		t: t,
		d: NewTransportDataset(id, mem),
	}
	return tr, tr.d, nil
}

// ProcessMessage will process the incoming message.
func (n *narrowStateTransformation) ProcessMessage(m Message) error {
	defer m.Ack()

	switch m := m.(type) {
	case FinishMsg:
		n.Finish(m.SrcDatasetID(), m.Error())
		return nil
	case ProcessChunkMsg:
		chunk := m.TableChunk()
		state, _ := n.d.Lookup(chunk.Key())
		if ns, ok, err := n.t.Process(chunk, state, n.d, n.d.mem); err != nil {
			return err
		} else if ok {
			n.d.Set(chunk.Key(), ns)
		}
		return nil
	case FlushKeyMsg:
		if err := n.d.FlushKey(m.Key()); err != nil {
			return err
		}
		if v, ok := n.d.Delete(m.Key()); ok {
			if v, ok := v.(Disposable); ok {
				v.Dispose()
			}
		}
		return nil
	case ProcessMsg:
		return n.Process(m.SrcDatasetID(), m.Table())
	}
	return nil
}

// Process is implemented to remain compatible with legacy upstreams.
// It converts the incoming stream into a set of appropriate messages.
func (n *narrowStateTransformation) Process(id DatasetID, tbl flux.Table) error {
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
func (n *narrowStateTransformation) Finish(id DatasetID, err error) {
	_ = n.d.Range(func(key flux.GroupKey, value interface{}) error {
		if v, ok := value.(Disposable); ok {
			v.Dispose()
		}
		return nil
	})
	n.d.Finish(err)
	n.t.Dispose()
}

func (n *narrowStateTransformation) OperationType() string {
	return OperationType(n.t)
}
func (n *narrowStateTransformation) RetractTable(id DatasetID, key flux.GroupKey) error {
	return nil
}
func (n *narrowStateTransformation) UpdateWatermark(id DatasetID, t Time) error {
	return nil
}
func (n *narrowStateTransformation) UpdateProcessingTime(id DatasetID, t Time) error {
	return nil
}
