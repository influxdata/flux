package execute_test

import (
	"testing"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/execute/table/static"
)

func TestProcessMsg(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	// Create a set of tables, wrap them in a process message,
	// and then ack it.
	ti := static.Table{
		static.StringKey("_measurement", "m0"),
		static.Times("_time", 0, 10, 20),
		static.Floats("_value", 0, 1, 2),
	}
	if err := ti.Do(func(tbl flux.Table) error {
		m := execute.NewProcessMsg(tbl)
		m.Ack()
		return nil
	}); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	mem.AssertSize(t, 0)

	// Create the same set of tables, but this time we are
	// going to use dup and acknowledge the duplications.
	// This should similarly work.
	if err := ti.Do(func(tbl flux.Table) error {
		m := execute.NewProcessMsg(tbl)
		dup1, dup2 := m.Dup(), m.Dup()
		m.Ack()
		dup1.Ack()
		dup2.Ack()
		return nil
	}); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	mem.AssertSize(t, 0)

	// It should be possible to use Do in both of the above situations.
	if err := ti.Do(func(tbl flux.Table) error {
		m := execute.NewProcessMsg(tbl)
		if err := m.Table().Do(func(cr flux.ColReader) error {
			return nil
		}); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		m.Ack()
		return nil
	}); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	mem.AssertSize(t, 0)

	if err := ti.Do(func(tbl flux.Table) error {
		m := execute.NewProcessMsg(tbl)
		dup1, dup2 := m.Dup(), m.Dup()
		m.Ack()

		if err := dup1.(execute.ProcessMsg).Table().Do(func(cr flux.ColReader) error {
			return nil
		}); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		dup1.Ack()

		if err := dup2.(execute.ProcessMsg).Table().Do(func(cr flux.ColReader) error {
			return nil
		}); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		dup2.Ack()
		return nil
	}); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	mem.AssertSize(t, 0)
}

func TestProcessChunkMsg(t *testing.T) {
	fromTableChunk := func(chunk table.Chunk) flux.Table {
		buffer := chunk.Buffer()
		buffer.Retain()
		return table.FromBuffer(&buffer)
	}
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	// Create a set of tables, wrap them in a process message,
	// and then ack it.
	ti := static.Table{
		static.StringKey("_measurement", "m0"),
		static.Times("_time", 0, 10, 20),
		static.Floats("_value", 0, 1, 2),
	}
	if err := ti.Do(func(tbl flux.Table) error {
		return tbl.Do(func(cr flux.ColReader) error {
			chunk := table.ChunkFromReader(cr)
			chunk.Retain()
			m := execute.NewProcessChunkMsg(chunk)

			cr.Retain()
			want := table.Iterator{table.FromBuffer(cr)}
			got := table.Iterator{fromTableChunk(m.TableChunk())}
			if diff := table.Diff(want, got); diff != "" {
				t.Errorf("unexpected table data -want/+got:\n%s", diff)
			}
			m.Ack()
			return nil
		})
	}); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	mem.AssertSize(t, 0)

	// Create the same set of tables, but this time we are
	// going to use dup and acknowledge the duplications.
	// This should similarly work.
	if err := ti.Do(func(tbl flux.Table) error {
		return tbl.Do(func(cr flux.ColReader) error {
			chunk := table.ChunkFromReader(cr)
			chunk.Retain()
			m := execute.NewProcessChunkMsg(chunk)
			dup1, dup2 := m.Dup(), m.Dup()
			m.Ack()

			// Dup contents should be identical.
			if diff := table.Diff(
				table.Iterator{fromTableChunk(dup1.(execute.ProcessChunkMsg).TableChunk())},
				table.Iterator{fromTableChunk(dup2.(execute.ProcessChunkMsg).TableChunk())},
			); diff != "" {
				t.Errorf("unexpected table data:\n%s", diff)
			}

			// Acknowledge the original message multiple times.
			// This should not affect the dup'd messages.
			for i := 0; i < 10; i++ {
				m.Ack()
			}
			mem.AssertSize(t, 0)

			if diff := table.Diff(
				table.Iterator{fromTableChunk(dup1.(execute.ProcessChunkMsg).TableChunk())},
				table.Iterator{fromTableChunk(dup2.(execute.ProcessChunkMsg).TableChunk())},
			); diff != "" {
				t.Errorf("unexpected table data:\n%s", diff)
			}
			dup1.Ack()
			dup2.Ack()
			return nil
		})
	}); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	mem.AssertSize(t, 0)
}
