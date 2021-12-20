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

	Closer
}

var _ Transport = (*narrowStateTransformation)(nil)

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
	return NewTransformationFromTransport(tr), tr.d, nil
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
			if v, ok := v.(Closer); ok {
				if err := v.Close(); err != nil {
					return err
				}
			}
		}
		return nil
	case ProcessMsg:
		panic("unreachable")
	}
	return nil
}

// Finish is implemented to remain compatible with legacy upstreams.
func (n *narrowStateTransformation) Finish(id DatasetID, err error) {
	_ = n.d.Range(func(key flux.GroupKey, value interface{}) error {
		if v, ok := value.(Closer); ok {
			return v.Close()
		}
		return nil
	})
	err = Close(err, n.t)
	n.d.Finish(err)
}

func (n *narrowStateTransformation) OperationType() string {
	return OperationType(n.t)
}
