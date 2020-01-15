package executetest

import (
	"runtime/debug"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	uuid "github.com/satori/go.uuid"
)

func RandomDatasetID() execute.DatasetID {
	// uuid.NewV4 can return an error because of enthropy. We will stick with the previous
	// behavior of panicing on errors when creating new uuid's
	return execute.DatasetID(uuid.Must(uuid.NewV4()))
}

type Dataset struct {
	ID                    execute.DatasetID
	Retractions           []flux.GroupKey
	ProcessingTimeUpdates []execute.Time
	WatermarkUpdates      []execute.Time
	Finished              bool
	FinishedErr           error
}

func NewDataset(id execute.DatasetID) *Dataset {
	return &Dataset{
		ID: id,
	}
}

func (d *Dataset) AddTransformation(t execute.Transformation) {
	panic("not implemented")
}

func (d *Dataset) RetractTable(key flux.GroupKey) error {
	d.Retractions = append(d.Retractions, key)
	return nil
}

func (d *Dataset) UpdateProcessingTime(t execute.Time) error {
	d.ProcessingTimeUpdates = append(d.ProcessingTimeUpdates, t)
	return nil
}

func (d *Dataset) UpdateWatermark(mark execute.Time) error {
	d.WatermarkUpdates = append(d.WatermarkUpdates, mark)
	return nil
}

func (d *Dataset) Finish(err error) {
	if d.Finished {
		panic("finish has already been called")
	}
	d.Finished = true
	d.FinishedErr = err
}

func (d *Dataset) SetTriggerSpec(t plan.TriggerSpec) {
	panic("not implemented")
}

type datasetTransformation struct {
	d Dataset
}

func newDatasetTransformation(id execute.DatasetID) *datasetTransformation {
	return &datasetTransformation{
		d: Dataset{ID: id},
	}
}
func (d *datasetTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return d.d.RetractTable(key)
}
func (d *datasetTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	tbl.Done()
	return nil
}
func (d *datasetTransformation) UpdateWatermark(id execute.DatasetID, t execute.Time) error {
	return d.d.UpdateWatermark(t)
}
func (d *datasetTransformation) UpdateProcessingTime(id execute.DatasetID, t execute.Time) error {
	return d.d.UpdateProcessingTime(t)
}
func (d *datasetTransformation) Finish(id execute.DatasetID, err error) {
	d.d.Finish(err)
}

type NewTransformation func(execute.Dataset, execute.TableBuilderCache) execute.Transformation

func TransformationPassThroughTestHelper(t *testing.T, newTr NewTransformation) {
	t.Helper()

	now := execute.Now()
	d := NewDataset(RandomDatasetID())
	c := execute.NewTableBuilderCache(UnlimitedAllocator)
	c.SetTriggerSpec(plan.DefaultTriggerSpec)

	parentID := RandomDatasetID()
	tr := newTr(d, c)
	if err := tr.UpdateWatermark(parentID, now); err != nil {
		t.Fatal(err)
	}
	if err := tr.UpdateProcessingTime(parentID, now); err != nil {
		t.Fatal(err)
	}
	tr.Finish(parentID, nil)

	exp := &Dataset{
		ID:                    d.ID,
		ProcessingTimeUpdates: []execute.Time{now},
		WatermarkUpdates:      []execute.Time{now},
		Finished:              true,
		FinishedErr:           nil,
	}
	if !cmp.Equal(d, exp) {
		t.Errorf("unexpected dataset -want/+got\n%s", cmp.Diff(exp, d))
	}
}

func TransformationPassThroughTestHelper2(
	t *testing.T,
	create func(id execute.DatasetID, alloc *memory.Allocator) (execute.Transformation, execute.Dataset),
) {
	t.Helper()

	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			t.Fatalf("caught panic: %v", err)
		}
	}()

	now := execute.Now()
	exp := &Dataset{
		ID:                    RandomDatasetID(),
		ProcessingTimeUpdates: []execute.Time{now},
		WatermarkUpdates:      []execute.Time{now},
		Finished:              true,
		FinishedErr:           nil,
	}

	alloc := &memory.Allocator{}
	gotT := newDatasetTransformation(exp.ID)
	tx, d := create(RandomDatasetID(), alloc)
	d.SetTriggerSpec(plan.DefaultTriggerSpec)
	d.AddTransformation(gotT)

	parentID := RandomDatasetID()
	if err := tx.UpdateWatermark(parentID, now); err != nil {
		t.Fatal(err)
	}
	if err := tx.UpdateProcessingTime(parentID, now); err != nil {
		t.Fatal(err)
	}
	tx.Finish(parentID, nil)

	if !cmp.Equal(&gotT.d, exp) {
		t.Errorf("unexpected dataset -want/+got\n%s", cmp.Diff(exp, &gotT.d))
	}
}
