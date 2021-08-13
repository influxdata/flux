package execute_test

import (
	"testing"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/execute/table/static"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/mock"
	"github.com/influxdata/flux/values"
)

func TestNarrowTransformation_ProcessChunk(t *testing.T) {
	// Ensure we allocate and free all memory correctly.
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	// Generate one table chunk using static.Table.
	// This will only produce one column reader, so we are
	// extracting that value from the nested iterators.
	want := static.Table{
		static.Times("_time", 0, 10, 20),
		static.Floats("_value", 1, 2, 3),
	}

	isProcessed := false
	tr, _, err := execute.NewNarrowTransformation(
		executetest.RandomDatasetID(),
		&mock.NarrowTransformation{
			ProcessFn: func(chunk table.Chunk, d *execute.TransportDataset, _ memory.Allocator) error {
				// Memory should be allocated and should not have been improperly freed.
				// This accounts for 64 bytes (data) + 64 bytes (null bitmap) for each column
				// of which there are two. 64 bytes is the minimum that arrow will allocate
				// for a particular data buffer.
				mem.AssertSize(t, 256)

				// Compare the buffer contents to the table we wanted.
				// Because we should have produced only one table chunk,
				// we are comparing the entirety of the chunk to the entirety
				// of the wanted output.
				buffer := chunk.Buffer()
				got := table.Iterator{
					table.FromBuffer(&buffer),
				}

				if diff := table.Diff(want, got); diff != "" {
					t.Errorf("unexpected diff -want/+got:\n%s", diff)
				}
				isProcessed = true
				return nil
			},
		},
		mem,
	)
	if err != nil {
		t.Fatal(err)
	}

	// We can use a TransportDataset as a mock source
	// to send messages to the transformation we are testing.
	source := execute.NewTransportDataset(executetest.RandomDatasetID(), mem)
	source.AddTransformation(tr)

	tbl := want.Table(mem)
	if err := tbl.Do(func(cr flux.ColReader) error {
		chunk := table.ChunkFromReader(cr)
		chunk.Retain()
		return source.Process(chunk)
	}); err != nil {
		t.Fatal(err)
	}

	if !isProcessed {
		t.Error("message was never processed")
	}
}

func TestNarrowTransformation_FlushKey(t *testing.T) {
	want := execute.NewGroupKey(
		[]flux.ColMeta{
			{Label: "t0", Type: flux.TString},
		},
		[]values.Value{
			values.NewString("a"),
		},
	)

	isProcessed := false
	tr := &mock.Transport{
		ProcessMessageFn: func(m execute.Message) error {
			msg, ok := m.(execute.FlushKeyMsg)
			if !ok {
				t.Fatalf("expected flush key message, got %T", m)
			}

			if got := msg.Key(); !want.Equal(got) {
				t.Fatalf("unexpected group key -want/+got:\n%s", cmp.Diff(want, got))
			}
			isProcessed = true
			return nil
		},
	}

	d := execute.NewTransportDataset(executetest.RandomDatasetID(), memory.DefaultAllocator)
	d.AddTransformation(tr)
	if err := d.FlushKey(want); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !isProcessed {
		t.Error("message was never processed")
	}
}

func TestNarrowTransformation_Process(t *testing.T) {
	// Ensure we allocate and free all memory correctly.
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	// Generate one table chunk using static.Table.
	// This will only produce one column reader, so we are
	// extracting that value from the nested iterators.
	want := static.Table{
		static.Times("_time", 0, 10, 20),
		static.Floats("_value", 1, 2, 3),
	}

	isProcessed := false
	tr, _, err := execute.NewNarrowTransformation(
		executetest.RandomDatasetID(),
		&mock.NarrowTransformation{
			ProcessFn: func(chunk table.Chunk, d *execute.TransportDataset, _ memory.Allocator) error {
				// Memory should be allocated and should not have been improperly freed.
				// This accounts for 64 bytes (data) + 64 bytes (null bitmap) for each column
				// of which there are two. 64 bytes is the minimum that arrow will allocate
				// for a particular data buffer.
				mem.AssertSize(t, 256)

				// Compare the buffer contents to the table we wanted.
				// Because we should have produced only one table chunk,
				// we are comparing the entirety of the chunk to the entirety
				// of the wanted output.
				buffer := chunk.Buffer()
				got := table.Iterator{
					table.FromBuffer(&buffer),
				}

				if diff := table.Diff(want, got); diff != "" {
					t.Errorf("unexpected diff -want/+got:\n%s", diff)
				}
				isProcessed = true
				return nil
			},
		},
		mem,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Process the table and ensure it gets converted to a table chunk,
	// memory is still allocated for it, and the actual data is correct.
	//
	// Instead of using public methods, we simulate sending a process message.
	// We want to test a transport's ability to forward the process message,
	// but the only transport that has this capability is the consecutive transport.
	// As we don't want to add the dispatcher or concurrency to this test, we manually
	// construct the message and send it ourselves.
	m := execute.NewProcessMsg(want.Table(mem))
	if err := tr.(execute.Transport).ProcessMessage(m); err != nil {
		t.Fatal(err)
	}

	if !isProcessed {
		t.Error("message was never processed")
	}
}

func TestNarrowTransformation_Finish(t *testing.T) {
	isFinished := []bool{false, false}

	want := errors.New(codes.Internal, "expected")

	tr, d, err := execute.NewNarrowTransformation(
		executetest.RandomDatasetID(),
		&mock.NarrowTransformation{},
		memory.DefaultAllocator,
	)
	if err != nil {
		t.Fatal(err)
	}
	d.AddTransformation(
		&mock.Transport{
			ProcessMessageFn: func(m execute.Message) error {
				msg, ok := m.(execute.FinishMsg)
				if !ok {
					t.Fatalf("expected finish message, got %T", m)
				}

				if got := msg.Error(); !cmp.Equal(want, got) {
					t.Fatalf("unexpected error -want/+got:\n%s", cmp.Diff(want, got))
				}
				isFinished[0] = true
				return nil
			},
		},
	)
	d.AddTransformation(
		&mock.Transformation{
			FinishFn: func(id execute.DatasetID, err error) {
				if got := err; !cmp.Equal(want, got) {
					t.Fatalf("unexpected error -want/+got:\n%s", cmp.Diff(want, got))
				}
				isFinished[1] = true
			},
		},
	)

	// We want to check that finish is forwarded correctly.
	// We construct the finish message and then send it directly
	// to the process message method.
	if err := tr.(execute.Transport).ProcessMessage(
		execute.NewFinishMsg(want),
	); err != nil {
		t.Fatal(err)
	}

	d.Finish(want)

	if !isFinished[0] {
		t.Error("transport did not receive finish message")
	}
	if !isFinished[1] {
		t.Error("transformation did not receive finish message")
	}
}

// Ensure that we report the operation type of the type we wrap
// and ensure that we don't report ourselves as the operation type.
//
// This is to prevent opentracing from showing narrowTransformation
// as the operation.
func TestNarrowTransformation_OperationType(t *testing.T) {
	tr, _, err := execute.NewNarrowTransformation(
		executetest.RandomDatasetID(),
		&mock.NarrowTransformation{},
		memory.DefaultAllocator,
	)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := "*mock.NarrowTransformation", execute.OperationType(tr); want != got {
		t.Errorf("unexpected operation type -want/+got:\n\t- %s\n\t+ %s", want, got)
	}
}
