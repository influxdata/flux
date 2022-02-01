package universe

import (
	"sync"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
)

type unionTransformation2 struct {
	d       *execute.TransportDataset
	parents int
	mu      sync.Mutex
}

func newUnionTransformation2(id execute.DatasetID, parents []execute.DatasetID, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	tr := &unionTransformation2{
		d:       execute.NewTransportDataset(id, mem),
		parents: len(parents),
	}
	return tr, tr.d, nil
}

func (u *unionTransformation2) ProcessMessage(m execute.Message) error {
	defer m.Ack()

	// It is possible to receive messages from different parents concurrently.
	// Lock this here to prevent a race condition.
	u.mu.Lock()
	defer u.mu.Unlock()

	// If this transformation was already marked as finished
	// from an error, do not accept new messages at all.
	if u.parents == 0 {
		return nil
	}

	switch m := m.(type) {
	case execute.FinishMsg:
		u.Finish(m.SrcDatasetID(), m.Error())
		return nil
	case execute.ProcessChunkMsg:
		return u.processChunk(m.TableChunk(), u.d)
	case execute.FlushKeyMsg:
		return nil
	case execute.ProcessMsg:
		return u.Process(m.SrcDatasetID(), m.Table())
	}
	return nil
}

func (u *unionTransformation2) processChunk(chunk table.Chunk, d *execute.TransportDataset) error {
	schema, ok := d.Lookup(chunk.Key())
	if !ok {
		d.Set(chunk.Key(), unionSchema{cols: chunk.Cols()})
	} else {
		normalized, err := u.normalizeSchema(schema.(unionSchema), chunk.Cols())
		if err != nil {
			return err
		}
		d.Set(chunk.Key(), normalized)
	}
	chunk.Retain()
	return d.Process(chunk)
}

func (u *unionTransformation2) normalizeSchema(schema unionSchema, cols []flux.ColMeta) (unionSchema, error) {
	for _, col := range cols {
		idx := execute.ColIdx(col.Label, schema.cols)
		if idx >= 0 {
			cur := schema.cols[idx]
			if cur.Type != col.Type {
				return unionSchema{}, errors.Newf(codes.FailedPrecondition, "schema collision detected: column \"%s\" is both of type %s and %s", col.Label, col.Type, cur.Type)
			}
		} else {
			if !schema.owned {
				cpy := make([]flux.ColMeta, len(schema.cols), len(schema.cols)+1)
				copy(cpy, schema.cols)
				schema.cols, schema.owned = cpy, true
			}
			schema.cols = append(schema.cols, col)
		}
	}
	return schema, nil
}

func (u *unionTransformation2) Process(id execute.DatasetID, tbl flux.Table) error {
	return tbl.Do(func(cr flux.ColReader) error {
		chunk := table.ChunkFromReader(cr)
		chunk.Retain()
		return u.processChunk(chunk, u.d)
	})
}

func (u *unionTransformation2) Finish(id execute.DatasetID, err error) {
	u.parents--
	if u.parents == 0 || err != nil {
		u.d.Finish(err)
		u.parents = 0
	}
}

func (u *unionTransformation2) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return nil
}
func (u *unionTransformation2) UpdateWatermark(id execute.DatasetID, t execute.Time) error {
	return nil
}
func (u *unionTransformation2) UpdateProcessingTime(id execute.DatasetID, t execute.Time) error {
	return nil
}

func (u *unionTransformation2) Close() error { return nil }

type unionSchema struct {
	cols  []flux.ColMeta
	owned bool
}
