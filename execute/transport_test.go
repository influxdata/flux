package execute_test

import (
	"testing"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
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
