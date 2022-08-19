package execute

import (
	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute/table"
)

// NarrowStateTransformation is the same as a NarrowTransformation
// except that it retains state between processing buffers.
type NarrowStateTransformation[T any] interface {
	// Process will process the TableView.
	Process(chunk table.Chunk, state T, d *TransportDataset, mem memory.Allocator) (T, bool, error)

	Closer
}

var _ Transport = (*narrowStateTransformation[any])(nil)

type narrowStateTransformation[T any] struct {
	t NarrowStateTransformation[T]
	d *TransportDataset
}

// NewNarrowStateTransformation constructs a Transformation and Dataset
// using the NarrowStateTransformation implementation.
func NewNarrowStateTransformation[T any](id DatasetID, t NarrowStateTransformation[T], mem memory.Allocator) (Transformation, Dataset, error) {
	tr := &narrowStateTransformation[T]{
		t: t,
		d: NewTransportDataset(id, mem),
	}
	return NewTransformationFromTransport(tr), tr.d, nil
}

// ProcessMessage will process the incoming message.
func (n *narrowStateTransformation[T]) ProcessMessage(m Message) error {
	defer m.Ack()

	switch m := m.(type) {
	case FinishMsg:
		n.Finish(m.SrcDatasetID(), m.Error())
		return nil
	case ProcessChunkMsg:
		chunk := m.TableChunk()

		var state T
		if value, ok := n.d.Lookup(chunk.Key()); ok {
			state = value.(T)
		}

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
func (n *narrowStateTransformation[T]) Finish(id DatasetID, err error) {
	_ = n.d.Range(func(key flux.GroupKey, value interface{}) error {
		if v, ok := value.(Closer); ok {
			return v.Close()
		}
		return nil
	})
	err = Close(err, n.t)
	n.d.Finish(err)
}

func (n *narrowStateTransformation[T]) OperationType() string {
	return OperationType(n.t)
}
