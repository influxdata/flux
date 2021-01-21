package execkit

import (
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/execute/table"
)

// AggregateTransformation implements a transformation that aggregates
// the results from multiple TableView values and then outputs a Table
// with the same group key.
//
// This is similar to NarrowTransformation that it does not modify the group key,
// but different because it will only output a table when the key is flushed.
type AggregateTransformation interface {
	// Aggregate will process the TableView with the state from the previous
	// time a table with this group key was invoked.
	// If this group key has never been invoked before, the
	// state will be nil.
	// The transformation should return the new state and a boolean
	// value of true if the state was created or modified.
	// If false is returned, the new state will be discarded and any
	// old state will be kept.
	// It is ok for the transformation to modify the state if it is
	// a pointer. This is both allowed and recommended.
	Aggregate(view table.View, state interface{}, mem memory.Allocator) (interface{}, bool, error)

	// Compute will signal the AggregateTransformation to compute
	// the aggregate for the given key from the provided state.
	Compute(key flux.GroupKey, state interface{}, d *Dataset, mem memory.Allocator) error
}

type aggregateTransformation struct {
	t AggregateTransformation
	d *Dataset
}

// NewAggregateTransformation constructs a Transformation and Dataset
// using the aggregateTransformation implementation.
func NewAggregateTransformation(id DatasetID, t AggregateTransformation, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	tr := &aggregateTransformation{
		t: t,
		d: NewDataset(id, mem),
	}
	return tr, tr.d, nil
}

// ProcessMessage will process the incoming message.
func (n *aggregateTransformation) ProcessMessage(m execute.Message) error {
	defer m.Ack()

	switch m := m.(type) {
	case execute.FinishMsg:
		n.Finish(m.SrcDatasetID(), m.Error())
		return nil
	case execute.ProcessViewMsg:
		// Load the state associated with this group key.
		// If there is no state, use nil as a placeholder.
		// We use LookupOrCreate because we assume that this
		// will succeed and LookupOrCreate is more efficient
		// than Lookup followed by Set.
		// If an error does occur, then this isn't going to
		// matter anyway.
		view := m.View()
		state, _ := n.d.Lookup(view.Key())
		if ns, ok, err := n.t.Aggregate(view, state, n.d.mem); err != nil {
			return err
		} else if ok {
			n.d.Set(view.Key(), ns)
		}
		return nil
	case execute.FlushKeyMsg:
		// When we are flushing a key, perform a lookup for any
		// state. If there is state, then send it to Compute.
		if state, ok := n.d.Delete(m.Key()); ok {
			if err := n.t.Compute(m.Key(), state, n.d, n.d.mem); err != nil {
				return err
			}
			return n.d.FlushKey(m.Key())
		}
		return nil
	case execute.ProcessMsg:
		return n.Process(m.SrcDatasetID(), m.Table())
	}
	return nil
}

// Process is implemented to remain compatible with legacy upstreams.
// It converts the incoming stream into a set of appropriate messages.
func (n *aggregateTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
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
func (n *aggregateTransformation) Finish(id execute.DatasetID, err error) {
	if err == nil {
		n.d.cache.Range(func(key flux.GroupKey, value interface{}) {
			if err != nil {
				return
			}
			err = n.t.Compute(key, value, n.d, n.d.mem)
		})
		n.d.cache.Clear()
	}

	if err != nil {
		_ = n.d.Abort(err)
		return
	}
	_ = n.d.Close()
}

func (n *aggregateTransformation) OperationType() string {
	return execute.OperationType(n.t)
}
func (n *aggregateTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return nil
}
func (n *aggregateTransformation) UpdateWatermark(id execute.DatasetID, t execute.Time) error {
	return nil
}
func (n *aggregateTransformation) UpdateProcessingTime(id execute.DatasetID, t execute.Time) error {
	return nil
}
