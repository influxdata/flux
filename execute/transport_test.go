package execute_test

import (
	"testing"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/execute/table/static"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/mock"
	"github.com/influxdata/flux/values"
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

func TestWrapTransformationInTransport(t *testing.T) {
	want := static.TableGroup{
		static.StringKey("_m", "m0"),
		static.StringKeys("t0", "a", "b"),
		static.Times("_time", 0, 10, 20),
		static.Floats("_value", 0, 1, 2),
	}

	t.Run("FlushKey", func(t *testing.T) {
		var (
			got      table.Iterator
			finished bool
		)

		// Create transformation that stores any processed tables into the
		// list of got tables and marks when it has been finished.
		tr := execute.WrapTransformationInTransport(&mock.Transformation{
			ProcessFn: func(id execute.DatasetID, tbl flux.Table) error {
				buf, err := table.Copy(tbl)
				if err != nil {
					return err
				}
				got = append(got, buf)
				return nil
			},
			FinishFn: func(id execute.DatasetID, err error) {
				if err != nil {
					t.Error(err)
				}
				finished = true
			},
		}, memory.DefaultAllocator)

		processed := 0
		if err := want.Do(func(tbl flux.Table) error {
			if err := tbl.Do(func(cr flux.ColReader) error {
				chunk := table.ChunkFromReader(cr)
				chunk.Retain()
				m := execute.NewProcessChunkMsg(chunk)
				return tr.ProcessMessage(m)
			}); err != nil {
				return err
			}

			// At this point, the table has sent all chunks to the transformation,
			// but the table hasn't been processed yet because it hasn't been flushed.
			if got, want := len(got), processed; got != want {
				t.Errorf("wrong number of tables processed -want/+got:\n\t- %d\n\t+ %d", want, got)
			}

			// Flush the group key which will cause the above table to be processed.
			m := execute.NewFlushKeyMsg(tbl.Key())
			if err := tr.ProcessMessage(m); err != nil {
				return err
			}
			// Increment the number processed because the flush key message should increment it.
			processed++

			// We should have processed the table.
			if got, want := len(got), processed; got != want {
				t.Errorf("wrong number of tables processed -want/+got:\n\t- %d\n\t+ %d", want, got)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}

		// Include a group key that hasn't previously been seen.
		// This should not cause an error and nothing should happen.
		{
			key := execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "_nonexistant", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("a"),
				},
			)
			m := execute.NewFlushKeyMsg(key)
			if err := tr.ProcessMessage(m); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		}

		m := execute.NewFinishMsg(nil)
		if err := tr.ProcessMessage(m); err != nil {
			t.Fatal(err)
		}

		// Finish should have been invoked.
		if !finished {
			t.Error("finish message not received")
		}

		// Compare the output to ensure tables were processed correctly.
		if diff := table.Diff(want, got); diff != "" {
			t.Errorf("unexpected table data -want/+got:\n%s", diff)
		}
	})

	t.Run("Finish", func(t *testing.T) {
		var (
			got      table.Iterator
			finished bool
		)

		// Create transformation that stores any processed tables into the
		// list of got tables and marks when it has been finished.
		tr := execute.WrapTransformationInTransport(&mock.Transformation{
			ProcessFn: func(id execute.DatasetID, tbl flux.Table) error {
				buf, err := table.Copy(tbl)
				if err != nil {
					return err
				}
				got = append(got, buf)
				return nil
			},
			FinishFn: func(id execute.DatasetID, err error) {
				if err != nil {
					t.Error(err)
				}
				finished = true
			},
		}, memory.DefaultAllocator)

		processed := 0
		if err := want.Do(func(tbl flux.Table) error {
			if err := tbl.Do(func(cr flux.ColReader) error {
				chunk := table.ChunkFromReader(cr)
				chunk.Retain()
				m := execute.NewProcessChunkMsg(chunk)
				return tr.ProcessMessage(m)
			}); err != nil {
				return err
			}
			processed++

			// We increment the number of tables that should be ready to be processed,
			// but we check here to make sure that process has not been invoked.
			if got, want := len(got), 0; got != want {
				t.Errorf("wrong number of tables processed -want/+got:\n\t- %d\n\t+ %d", want, got)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}

		// Since there have been no flushes, none of the tables should have been processed.
		if got, want := len(got), 0; got != want {
			t.Errorf("wrong number of tables processed -want/+got:\n\t- %d\n\t+ %d", want, got)
		}

		m := execute.NewFinishMsg(nil)
		if err := tr.ProcessMessage(m); err != nil {
			t.Fatal(err)
		}

		// The transformation should have called finish.
		if !finished {
			t.Error("finish message not received")
		}

		// Pending tables should have been flushed.
		if got, want := len(got), processed; got != want {
			t.Errorf("wrong number of tables processed -want/+got:\n\t- %d\n\t+ %d", want, got)
		}

		// Compare the output to ensure tables were processed correctly.
		if diff := table.Diff(want, got); diff != "" {
			t.Errorf("unexpected table data -want/+got:\n%s", diff)
		}
	})

	t.Run("ProcessError", func(t *testing.T) {
		var finished bool

		// Create transformation that errors when it processes a table,
		// but also tracks if finish is called.
		tr := execute.WrapTransformationInTransport(&mock.Transformation{
			ProcessFn: func(id execute.DatasetID, tbl flux.Table) error {
				return errors.New(codes.Invalid, "expected")
			},
			FinishFn: func(id execute.DatasetID, err error) {
				if err == nil {
					t.Error("expected error")
				} else if want, got := "expected", err.Error(); want != got {
					t.Errorf("unexpected error -want/+got:\n\t- %s\n\t+ %s", want, got)
				}
				finished = true
			},
		}, memory.DefaultAllocator)

		// Send table chunks which should be buffered.
		// This should not error as it isn't sent to the transformation yet.
		if err := want.Do(func(tbl flux.Table) error {
			return tbl.Do(func(cr flux.ColReader) error {
				chunk := table.ChunkFromReader(cr)
				chunk.Retain()
				m := execute.NewProcessChunkMsg(chunk)
				return tr.ProcessMessage(m)
			})
		}); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		// Invoke finish on the transport.
		// The transport should be finished with an error.
		m := execute.NewFinishMsg(nil)
		if err := tr.ProcessMessage(m); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if !finished {
			t.Error("expected finish to be invoked with an error")
		}
	})
}
