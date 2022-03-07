package execute_test

import (
	"testing"

	arrowmem "github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/mock"
)

func TestTransportDataset_Process(t *testing.T) {
	isProcessed := false

	transport := &mock.Transport{
		ProcessMessageFn: func(m execute.Message) error {
			defer m.Ack()

			if want, got := m.Type(), execute.ProcessChunkType; want != got {
				t.Errorf("unexpected message type -want/+got:\n\t- %v\n\t+ %v", want, got)
			}

			chunk := m.(execute.ProcessChunkMsg).TableChunk()
			if want, got := []flux.ColMeta{
				{Label: "_time", Type: flux.TTime},
				{Label: "_value", Type: flux.TFloat},
			}, chunk.Cols(); !cmp.Equal(want, got) {
				t.Fatalf("unexpected table columns -want/+got:\n%s", cmp.Diff(want, got))
			}

			if want, got := chunk.Values(0).(*array.Int).Int64Values(), []int64{0, 10, 20}; !cmp.Equal(want, got) {
				t.Fatalf("unexpected time values -want/+got:\n%s", cmp.Diff(want, got))
			}

			if want, got := chunk.Values(1).(*array.Float).Float64Values(), []float64{1, 2, 3}; !cmp.Equal(want, got) {
				t.Fatalf("unexpected time values -want/+got:\n%s", cmp.Diff(want, got))
			}

			isProcessed = true
			return nil
		},
	}

	dataset := execute.NewTransportDataset(executetest.RandomDatasetID(), memory.DefaultAllocator)
	dataset.AddTransformation(transport)

	mem := arrowmem.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)
	alloc := &memory.Allocator{
		Allocator: mem,
	}
	buffer := arrow.TableBuffer{
		GroupKey: execute.NewGroupKey(nil, nil),
		Columns: []flux.ColMeta{
			{Label: "_time", Type: flux.TTime},
			{Label: "_value", Type: flux.TFloat},
		},
		Values: []array.Array{
			arrow.NewInt([]int64{0, 10, 20}, alloc),
			arrow.NewFloat([]float64{1, 2, 3}, alloc),
		},
	}
	chunk := table.ChunkFromBuffer(buffer)

	if err := dataset.Process(chunk); err != nil {
		t.Fatal(err)
	}

	if !isProcessed {
		t.Fatal("message was not processed")
	}
}

func TestTransportDataset_AddTransformation(t *testing.T) {
	isProcessed := false

	transformation := &mock.Transformation{
		ProcessFn: func(id execute.DatasetID, tbl flux.Table) error {
			isProcessed = true
			tbl.Done()
			return nil
		},
		FinishFn: func(id execute.DatasetID, err error) {
			t.Error(err)
		},
	}

	dataset := execute.NewTransportDataset(executetest.RandomDatasetID(), memory.DefaultAllocator)
	dataset.AddTransformation(transformation)

	mem := arrowmem.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)
	alloc := &memory.Allocator{
		Allocator: mem,
	}
	buffer := arrow.TableBuffer{
		GroupKey: execute.NewGroupKey(nil, nil),
		Columns: []flux.ColMeta{
			{Label: "_time", Type: flux.TTime},
			{Label: "_value", Type: flux.TFloat},
		},
		Values: []array.Array{
			arrow.NewInt([]int64{0, 10, 20}, alloc),
			arrow.NewFloat([]float64{1, 2, 3}, alloc),
		},
	}
	chunk := table.ChunkFromBuffer(buffer)

	if err := dataset.Process(chunk); err != nil {
		t.Fatal(err)
	}

	if isProcessed {
		t.Fatal("table processed before key flush")
	}

	if err := dataset.FlushKey(chunk.Key()); err != nil {
		t.Fatal(err)
	}

	if !isProcessed {
		t.Fatal("message was not processed")
	}
}

func TestTransportDataset_FlushKey(t *testing.T) {
	isProcessed := false

	transport := &mock.Transport{
		ProcessMessageFn: func(m execute.Message) error {
			defer m.Ack()

			if want, got := m.Type(), execute.FlushKeyType; want != got {
				t.Errorf("unexpected message type -want/+got:\n\t- %v\n\t+ %v", want, got)
			}
			isProcessed = true
			return nil
		},
	}

	dataset := execute.NewTransportDataset(executetest.RandomDatasetID(), memory.DefaultAllocator)
	dataset.AddTransformation(transport)

	key := execute.NewGroupKey(nil, nil)
	if err := dataset.FlushKey(key); err != nil {
		t.Fatal(err)
	}

	if !isProcessed {
		t.Fatal("message was not processed")
	}
}

func TestTransportDataset_MultipleDownstream(t *testing.T) {
	numProcessed := 0

	transport := &mock.Transport{
		ProcessMessageFn: func(m execute.Message) error {
			defer m.Ack()

			if want, got := m.Type(), execute.ProcessChunkType; want != got {
				t.Errorf("unexpected message type -want/+got:\n\t- %v\n\t+ %v", want, got)
			}

			chunk := m.(execute.ProcessChunkMsg).TableChunk()
			if want, got := []flux.ColMeta{
				{Label: "_time", Type: flux.TTime},
				{Label: "_value", Type: flux.TFloat},
			}, chunk.Cols(); !cmp.Equal(want, got) {
				t.Fatalf("unexpected table columns -want/+got:\n%s", cmp.Diff(want, got))
			}

			if want, got := chunk.Values(0).(*array.Int).Int64Values(), []int64{0, 10, 20}; !cmp.Equal(want, got) {
				t.Fatalf("unexpected time values -want/+got:\n%s", cmp.Diff(want, got))
			}

			if want, got := chunk.Values(1).(*array.Float).Float64Values(), []float64{1, 2, 3}; !cmp.Equal(want, got) {
				t.Fatalf("unexpected time values -want/+got:\n%s", cmp.Diff(want, got))
			}

			numProcessed++
			return nil
		},
	}

	dataset := execute.NewTransportDataset(executetest.RandomDatasetID(), memory.DefaultAllocator)
	dataset.AddTransformation(transport)
	dataset.AddTransformation(transport)

	mem := arrowmem.NewCheckedAllocator(memory.DefaultAllocator)
	defer mem.AssertSize(t, 0)
	alloc := &memory.Allocator{
		Allocator: mem,
	}
	buffer := arrow.TableBuffer{
		GroupKey: execute.NewGroupKey(nil, nil),
		Columns: []flux.ColMeta{
			{Label: "_time", Type: flux.TTime},
			{Label: "_value", Type: flux.TFloat},
		},
		Values: []array.Array{
			arrow.NewInt([]int64{0, 10, 20}, alloc),
			arrow.NewFloat([]float64{1, 2, 3}, alloc),
		},
	}
	chunk := table.ChunkFromBuffer(buffer)

	if err := dataset.Process(chunk); err != nil {
		t.Fatal(err)
	}

	if want, got := 2, numProcessed; want != got {
		t.Fatalf("unexpected number of messages -want/+got:\n\t- %d\n\t+ %d", want, got)
	}
}
