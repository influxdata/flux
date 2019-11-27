package executetest

import (
	"math"
	"sort"
	"strings"
	"testing"

	"runtime/debug"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependencies/url"
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
	store := NewDataStore()
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

// DataStore will store the incoming tables from an upstream transformation or source.
type DataStore struct {
	tables *execute.GroupLookup
	err    error
}

func NewDataStore() *DataStore {
	return &DataStore{
		tables: execute.NewGroupLookup(),
	}
}

func (d *DataStore) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	d.tables.Delete(key)
	return nil
}

func (d *DataStore) Process(id execute.DatasetID, tbl flux.Table) error {
	tbl, err := execute.CopyTable(tbl)
	if err != nil {
		return err
	}
	d.tables.Set(tbl.Key(), tbl)
	return nil
}

func (d *DataStore) UpdateWatermark(id execute.DatasetID, t execute.Time) error {
	return nil
}

func (d *DataStore) UpdateProcessingTime(id execute.DatasetID, t execute.Time) error {
	return nil
}

func (d *DataStore) Finish(id execute.DatasetID, err error) {
	if err != nil {
		d.err = err
	}
}

func (d *DataStore) Table(key flux.GroupKey) (flux.Table, error) {
	data, ok := d.tables.Lookup(key)
	if !ok {
		return nil, errors.Newf(codes.Internal, "table with key %v not found", key)
	}
	return data.(flux.Table), nil
}

func (d *DataStore) Err() error { return d.err }

func (d *DataStore) ForEach(f func(key flux.GroupKey)) {
	d.tables.Range(func(key flux.GroupKey, _ interface{}) {
		f(key)
	})
}

func (d *DataStore) ForEachWithContext(f func(flux.GroupKey, execute.Trigger, execute.TableContext)) {
	d.tables.Range(func(key flux.GroupKey, _ interface{}) {
		f(key, nil, execute.TableContext{
			Key: key,
		})
	})
}

func (d *DataStore) DiscardTable(key flux.GroupKey) {
	d.tables.Delete(key)
}

func (d *DataStore) ExpireTable(key flux.GroupKey) {
	d.tables.Delete(key)
}

func (d *DataStore) SetTriggerSpec(t plan.TriggerSpec) {
}

func ProcessBenchmarkHelper(
	b *testing.B,
	genInput func(alloc *memory.Allocator) (flux.TableIterator, error),
	create func(id execute.DatasetID, alloc *memory.Allocator) (execute.Transformation, execute.Dataset),
) {
	b.Helper()

	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			b.Fatalf("caught panic: %v", err)
		}
	}()

	alloc := &memory.Allocator{}
	parentID := RandomDatasetID()
	tables, err := genInput(alloc)
	if err != nil {
		b.Fatalf("unexpected error: %s", err)
	}

	store := NewDevNullStore()
	tx, d := create(RandomDatasetID(), alloc)
	d.SetTriggerSpec(plan.DefaultTriggerSpec)
	d.AddTransformation(store)

	if err := tables.Do(func(table flux.Table) error {
		return tx.Process(parentID, table)
	}); err != nil {
		b.Fatalf("unexpected error: %s", err)
	}

	// We always return a fatal error on failure so
	// we only get here when the error is nil.
	tx.Finish(parentID, nil)
}

type devNullStore struct{}

func NewDevNullStore() execute.Transformation {
	return devNullStore{}
}

func (d devNullStore) RetractTable(id execute.DatasetID, key flux.GroupKey) error { return nil }
func (d devNullStore) Process(id execute.DatasetID, tbl flux.Table) error {
	return tbl.Do(func(flux.ColReader) error {
		return nil
	})
}
func (d devNullStore) UpdateWatermark(id execute.DatasetID, t execute.Time) error      { return nil }
func (d devNullStore) UpdateProcessingTime(id execute.DatasetID, t execute.Time) error { return nil }
func (d devNullStore) Finish(id execute.DatasetID, err error)                          {}

// Some transformations need to take a URL e.g. sql.to, kafka.to
// the URL/DSN supplied by the user need to be validated by a URLValidator{}
// before we can establish the connection.
// TfUrlValidationTestCase, TfUrlValidationTest (as well as the Run() method)
// acts as a test harness for that.

type TfUrlValidationTestCase struct {
	Name      string
	Spec      plan.ProcedureSpec
	Validator url.Validator
	WantErr   string
}

type TfUrlValidationTest struct {
	CreateFn CreateNewTransformationWithDeps
	Cases    []TfUrlValidationTestCase
}

// sql.createToSQLTransformation() and kafka.createToKafkaTransformation() converts plan.ProcedureSpec
// to their struct implementations ToSQLProcedureSpec and ToKafkaProcedureSpec respectively.
// This complicated the test harness requiring us to provide CreateNewTransformationWithDeps
// functions to do the plan.ProcedureSpec conversion and call the subsequent factory method
// namely: kafka.NewToKafkaTransformation() and sql.NewToSQLTransformation()
// See also: sql/to_test.go/TestToSql_NewTransformation and kafka/to_test.go/TestToKafka_NewTransformation
type CreateNewTransformationWithDeps func(d execute.Dataset, deps flux.Dependencies,
	cache execute.TableBuilderCache, spec plan.ProcedureSpec) (execute.Transformation, error)

func (test *TfUrlValidationTest) Run(t *testing.T) {
	for _, tc := range test.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			d := NewDataset(RandomDatasetID())
			c := execute.NewTableBuilderCache(UnlimitedAllocator)
			deps := dependenciestest.Default()
			if tc.Validator != nil {
				deps.Deps.URLValidator = tc.Validator
			}
			_, err := test.CreateFn(d, deps, c, tc.Spec)
			if err != nil {
				if tc.WantErr != "" {
					got := err.Error()
					if !strings.Contains(got, tc.WantErr) {
						t.Fatalf("unexpected result -want/+got:\n%s",
							cmp.Diff(got, tc.WantErr))
					}
					return
				} else {
					t.Fatal(err)
				}
			}
		})
	}
}
