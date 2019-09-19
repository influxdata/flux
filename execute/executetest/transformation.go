package executetest

import (
	"math"
	"sort"
	"testing"

	"runtime/debug"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"gonum.org/v1/gonum/floats"
)

// Two floating point values are considered
// equal if they are within tol of each other.
const tol float64 = 1e-25

// The maximum number of floating point values that are allowed
// to lie between two float64s and still be considered equal.
const ulp uint = 2

// Comparison options for floating point values.
// NaNs are considered equal, and float64s must
// be sufficiently close to be considered equal.
var floatOptions = cmp.Options{
	cmpopts.EquateNaNs(),
	cmp.FilterValues(func(x, y float64) bool {
		return !math.IsNaN(x) && !math.IsNaN(y)
	}, cmp.Comparer(func(x, y float64) bool {
		// If sufficiently close, then move on.
		// This avoids situations close to zero.
		if floats.EqualWithinAbs(x, y, tol) {
			return true
		}
		// If not sufficiently close, both floats
		// must be within ulp steps of each other.
		if !floats.EqualWithinULP(x, y, ulp) {
			return false
		}
		return true
	})),
}

func ProcessTestHelper(
	t *testing.T,
	data []flux.Table,
	want []*Table,
	wantErr error,
	create func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation,
) {
	t.Helper()

	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			t.Fatalf("caught panic: %v", err)
		}
	}()

	d := NewDataset(RandomDatasetID())
	c := execute.NewTableBuilderCache(UnlimitedAllocator)
	c.SetTriggerSpec(plan.DefaultTriggerSpec)

	tx := create(d, c)

	parentID := RandomDatasetID()
	var gotErr error
	for _, b := range data {
		if err := tx.Process(parentID, b); err != nil {
			gotErr = err
			break
		}
	}

	tx.Finish(parentID, gotErr)
	if gotErr == nil {
		gotErr = d.FinishedErr
	}

	if gotErr == nil && wantErr != nil {
		t.Fatalf("expected error %s, got none", wantErr.Error())
	} else if gotErr != nil && wantErr == nil {
		t.Fatalf("expected no error, got %s", gotErr.Error())
	} else if gotErr != nil && wantErr != nil {
		if wantErr.Error() != gotErr.Error() {
			t.Fatalf("unexpected error -want/+got\n%s", cmp.Diff(wantErr.Error(), gotErr.Error()))
		} else {
			return
		}
	}

	got, err := TablesFromCache(c)
	if err != nil {
		t.Fatal(err)
	}

	NormalizeTables(got)
	NormalizeTables(want)

	sort.Sort(SortedTables(got))
	sort.Sort(SortedTables(want))

	if !cmp.Equal(want, got, floatOptions) {
		t.Errorf("unexpected tables -want/+got\n%s", cmp.Diff(want, got))
	}
}

func ProcessTestHelper2(
	t *testing.T,
	data []flux.Table,
	want []*Table,
	wantErr error,
	create func(id execute.DatasetID, alloc *memory.Allocator) (execute.Transformation, execute.Dataset),
) {
	t.Helper()

	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			t.Fatalf("caught panic: %v", err)
		}
	}()

	alloc := &memory.Allocator{}
	store := newDataStore()
	tx, d := create(RandomDatasetID(), alloc)
	d.SetTriggerSpec(plan.DefaultTriggerSpec)
	d.AddTransformation(store)

	parentID := RandomDatasetID()
	var gotErr error
	for _, b := range data {
		if err := tx.Process(parentID, b); err != nil {
			gotErr = err
			break
		}
	}

	tx.Finish(parentID, gotErr)
	if gotErr == nil {
		gotErr = store.err
	}

	if gotErr == nil && wantErr != nil {
		t.Fatalf("expected error %s, got none", wantErr.Error())
	} else if gotErr != nil && wantErr == nil {
		t.Fatalf("expected no error, got %s", gotErr.Error())
	} else if gotErr != nil && wantErr != nil {
		if wantErr.Error() != gotErr.Error() {
			t.Fatalf("unexpected error -want/+got\n%s", cmp.Diff(wantErr.Error(), gotErr.Error()))
		} else {
			return
		}
	}

	got, err := TablesFromCache(store)
	if err != nil {
		t.Fatal(err)
	}

	NormalizeTables(got)
	NormalizeTables(want)

	sort.Sort(SortedTables(got))
	sort.Sort(SortedTables(want))

	if !cmp.Equal(want, got, floatOptions) {
		t.Errorf("unexpected tables -want/+got\n%s", cmp.Diff(want, got))
	}
}

// dataStore will store the incoming tables from an upstream transformation or source.
type dataStore struct {
	tables *execute.GroupLookup
	err    error
}

func newDataStore() *dataStore {
	return &dataStore{
		tables: execute.NewGroupLookup(),
	}
}

func (d *dataStore) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	d.tables.Delete(key)
	return nil
}

func (d *dataStore) Process(id execute.DatasetID, tbl flux.Table) error {
	tbl, err := execute.CopyTable(tbl)
	if err != nil {
		return err
	}
	d.tables.Set(tbl.Key(), tbl)
	return nil
}

func (d *dataStore) UpdateWatermark(id execute.DatasetID, t execute.Time) error {
	return nil
}

func (d *dataStore) UpdateProcessingTime(id execute.DatasetID, t execute.Time) error {
	return nil
}

func (d *dataStore) Finish(id execute.DatasetID, err error) {
	if err != nil {
		d.err = err
	}
}

func (d *dataStore) Table(key flux.GroupKey) (flux.Table, error) {
	data, ok := d.tables.Lookup(key)
	if !ok {
		return nil, errors.Newf(codes.Internal, "table with key %v not found", key)
	}
	return data.(flux.Table), nil
}

func (d *dataStore) ForEach(f func(key flux.GroupKey)) {
	d.tables.Range(func(key flux.GroupKey, _ interface{}) {
		f(key)
	})
}

func (d *dataStore) ForEachWithContext(f func(flux.GroupKey, execute.Trigger, execute.TableContext)) {
	d.tables.Range(func(key flux.GroupKey, _ interface{}) {
		f(key, nil, execute.TableContext{
			Key: key,
		})
	})
}

func (d *dataStore) DiscardTable(key flux.GroupKey) {
	d.tables.Delete(key)
}

func (d *dataStore) ExpireTable(key flux.GroupKey) {
	d.tables.Delete(key)
}

func (d *dataStore) SetTriggerSpec(t plan.TriggerSpec) {
}
