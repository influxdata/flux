package execute_test

import (
	"testing"

	"github.com/apache/arrow/go/v7/arrow/memory"
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
	"github.com/stretchr/testify/assert"
)

func TestNarrowStateTransformation_ProcessChunk(t *testing.T) {
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

	isProcessed, hasState := false, false
	tr, _, err := execute.NewNarrowStateTransformation(
		executetest.RandomDatasetID(),
		&mock.NarrowStateTransformation{
			ProcessFn: func(chunk table.Chunk, state interface{}, d *execute.TransportDataset, _ memory.Allocator) (interface{}, bool, error) {
				if state != nil {
					if want, got := int64(4), state.(int64); want != got {
						t.Errorf("unexpected state on second call -want/+got:\n\t- %d\n\t+ %d", want, got)
					}
					hasState = true
				}

				// Memory should be allocated and should not have been improperly freed.
				// This accounts for 64 bytes (data) + 64 bytes (null bitmap) for each column
				// of which there are two. 64 bytes is the minimum that arrow will allocate
				// for a particular data buffer.
				assert.Equal(t, 256, mem.CurrentAlloc(), "unexpected memory allocation.")

				// Compare the buffer contents to the table we wanted.
				// Because we should have produced only one table chunk,
				// we are comparing the entirety of the chunk to the entirety
				// of the wanted output.
				buffer := chunk.Buffer()
				buffer.Retain()
				got := table.Iterator{
					table.FromBuffer(&buffer),
				}

				if diff := table.Diff(want, got); diff != "" {
					t.Errorf("unexpected diff -want/+got:\n%s", diff)
				}
				isProcessed = true
				return int64(4), true, nil
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
		if err := source.Process(chunk); err != nil {
			return err
		}

		chunk.Retain()
		return source.Process(chunk)
	}); err != nil {
		t.Fatal(err)
	}

	if !isProcessed {
		t.Error("message was never processed")
	}
	if !hasState {
		t.Error("process was never invoked with state")
	}
}

func TestNarrowStateTransformation_FlushKey(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	want := static.Table{
		static.StringKey("t0", "a"),
		static.Times("_time", 0, 10, 20),
		static.Floats("_value", 1, 2, 3),
	}

	var disposeCount int
	tr, d, err := execute.NewNarrowStateTransformation(
		executetest.RandomDatasetID(),
		&mock.NarrowStateTransformation{
			ProcessFn: func(chunk table.Chunk, state interface{}, d *execute.TransportDataset, mem memory.Allocator) (interface{}, bool, error) {
				if state != nil {
					t.Error("process unexpectedly invoked with state")
				}
				chunk.Retain()
				if err := d.Process(chunk); err != nil {
					return nil, false, err
				}
				return &mockState{
					disposeCount: &disposeCount,
				}, true, nil
			},
		},
		mem,
	)
	if err != nil {
		t.Fatal(err)
	}

	isProcessed, isFlushed := false, false
	d.AddTransformation(
		&mock.Transport{
			ProcessMessageFn: func(m execute.Message) error {
				defer m.Ack()

				switch m := m.(type) {
				case execute.ProcessChunkMsg:
					isProcessed = true
				case execute.FlushKeyMsg:
					want := execute.NewGroupKey(
						[]flux.ColMeta{{Label: "t0", Type: flux.TString}},
						[]values.Value{values.NewString("a")},
					)

					if got := m.Key(); !want.Equal(got) {
						t.Errorf("unexpected group key -want/+got:\n%s", cmp.Diff(want, got))
					}
					isFlushed = true
				}
				return nil
			},
		},
	)

	// Flush key should flush the state so the second call to process
	// should not have any state.
	parentID := executetest.RandomDatasetID()
	for i := 0; i < 2; i++ {
		tbl := want.Table(mem)
		if err := tr.Process(parentID, tbl); err != nil {
			t.Fatal(err)
		}
	}

	if !isProcessed {
		t.Error("process message was never processed")
	}
	if !isFlushed {
		t.Error("flush key message was never processed")
	}

	// The state should have been disposed of twice.
	if want, got := 2, disposeCount; want != got {
		t.Errorf("unexpected dispose count -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
}

func TestNarrowStateTransformation_Process(t *testing.T) {
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
	tr, _, err := execute.NewNarrowStateTransformation(
		executetest.RandomDatasetID(),
		&mock.NarrowStateTransformation{
			ProcessFn: func(chunk table.Chunk, state interface{}, d *execute.TransportDataset, _ memory.Allocator) (interface{}, bool, error) {
				// Memory should be allocated and should not have been improperly freed.
				// This accounts for 64 bytes (data) + 64 bytes (null bitmap) for each column
				// of which there are two. 64 bytes is the minimum that arrow will allocate
				// for a particular data buffer.
				assert.Equal(t, 256, mem.CurrentAlloc(), "unexpected memory allocation.")

				// Compare the buffer contents to the table we wanted.
				// Because we should have produced only one table chunk,
				// we are comparing the entirety of the chunk to the entirety
				// of the wanted output.
				buffer := chunk.Buffer()
				buffer.Retain()
				got := table.Iterator{
					table.FromBuffer(&buffer),
				}

				if diff := table.Diff(want, got); diff != "" {
					t.Errorf("unexpected diff -want/+got:\n%s", diff)
				}
				isProcessed = true
				return int64(4), true, nil
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
	//
	// This test is identical to the version for narrow transformation
	// so we don't have anything special regarding state. The tests for
	// state are around the individual messages rather than this test and
	// this test is mostly for verifying that process is still equivalent to
	// process chunk and flush key.
	m := execute.NewProcessMsg(want.Table(mem))
	if err := tr.(execute.Transport).ProcessMessage(m); err != nil {
		t.Fatal(err)
	}

	if !isProcessed {
		t.Error("message was never processed")
	}
}

func TestNarrowStateTransformation_Finish(t *testing.T) {
	// Ensure we allocate and free all memory correctly.
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)

	isFinished := []bool{false, false}

	want := errors.New(codes.Internal, "expected")

	var (
		disposeCount int
		isDisposed   bool
	)
	tr, d, err := execute.NewNarrowStateTransformation(
		executetest.RandomDatasetID(),
		&mock.NarrowStateTransformation{
			ProcessFn: func(chunk table.Chunk, state interface{}, d *execute.TransportDataset, mem memory.Allocator) (interface{}, bool, error) {
				return &mockState{
					disposeCount: &disposeCount,
				}, true, nil
			},
			CloseFn: func() error {
				isDisposed = true
				return nil
			},
		},
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

	// We can use a TransportDataset as a mock source
	// to send messages to the transformation we are testing.
	source := execute.NewTransportDataset(executetest.RandomDatasetID(), mem)
	source.AddTransformation(tr)

	// Generate one table chunk using static.Table.
	// This will only produce one column reader, so we are
	// extracting that value from the nested iterators.
	gen := static.Table{
		static.Times("_time", 0, 10, 20),
		static.Floats("_value", 1, 2, 3),
	}

	// Process the table but do not flush the key.
	tbl := gen.Table(mem)
	if err := tbl.Do(func(cr flux.ColReader) error {
		chunk := table.ChunkFromReader(cr)
		chunk.Retain()
		return source.Process(chunk)
	}); err != nil {
		t.Fatal(err)
	}

	// We want to check that finish is forwarded correctly.
	source.Finish(want)

	if !isFinished[0] {
		t.Error("transport did not receive finish message")
	}
	if !isFinished[1] {
		t.Error("transformation did not receive finish message")
	}

	// The state should have been disposed.
	if want, got := 1, disposeCount; want != got {
		t.Errorf("unexpected dispose count -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	// So should the transformation.
	if !isDisposed {
		t.Error("transformation was not disposed")
	}
}

// Ensure that we report the operation type of the type we wrap
// and ensure that we don't report ourselves as the operation type.
//
// This is to prevent opentracing from showing narrowTransformation
// as the operation.
func TestNarrowStateTransformation_OperationType(t *testing.T) {
	tr, _, err := execute.NewNarrowStateTransformation(
		executetest.RandomDatasetID(),
		&mock.NarrowStateTransformation{},
		memory.DefaultAllocator,
	)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := "*mock.NarrowStateTransformation", execute.OperationType(tr); want != got {
		t.Errorf("unexpected operation type -want/+got:\n\t- %s\n\t+ %s", want, got)
	}
}
