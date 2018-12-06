package transformations

import (
	"errors"
	"fmt"
	"sync"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const AssertEqualsKind = "assertEquals"

type AssertEqualsOpSpec struct {
	Name string `json:"name"`
}

func (s *AssertEqualsOpSpec) Kind() flux.OperationKind {
	return AssertEqualsKind
}

func init() {
	assertEqualsSignature := semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"name": semantic.String,
			"got":  flux.TableObjectType,
			"want": flux.TableObjectType,
		},
		Required: semantic.LabelSet{"name", "got", "want"},
		Return:   flux.TableObjectType,
	}

	flux.RegisterFunction(AssertEqualsKind, createAssertEqualsOpSpec, assertEqualsSignature)
	flux.RegisterOpSpec(AssertEqualsKind, newAssertEqualsOp)
	plan.RegisterProcedureSpec(AssertEqualsKind, newAssertEqualsProcedure, AssertEqualsKind)
	execute.RegisterTransformation(AssertEqualsKind, createAssertEqualsTransformation)
}

func createAssertEqualsOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	t, err := args.GetRequiredObject("got")
	if err != nil {
		return nil, err
	}
	p, ok := t.(*flux.TableObject)
	if !ok {
		err = errors.New("got input to assertEquals is not a table object")
	}
	a.AddParent(p)

	t, err = args.GetRequiredObject("want")
	if err != nil {
		return nil, err
	}
	p, ok = t.(*flux.TableObject)
	if !ok {
		err = errors.New("want input to assertEquals is not a table object")
	}
	a.AddParent(p)

	if err != nil {
		return nil, err
	}

	var name string
	if name, err = args.GetRequiredString("name"); err != nil {
		return nil, err
	}

	return &AssertEqualsOpSpec{Name: name}, nil
}

func newAssertEqualsOp() flux.OperationSpec {
	return new(AssertEqualsOpSpec)
}

type AssertEqualsProcedureSpec struct {
	plan.DefaultCost
	Name string
}

func (s *AssertEqualsProcedureSpec) Kind() plan.ProcedureKind {
	return AssertEqualsKind
}

func (s *AssertEqualsProcedureSpec) Copy() plan.ProcedureSpec {
	return &AssertEqualsProcedureSpec{}
}

func newAssertEqualsProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*AssertEqualsOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &AssertEqualsProcedureSpec{Name: spec.Name}, nil
}

type AssertEqualsTransformation struct {
	mu sync.Mutex

	gotParent  *assertEqualsParentState
	wantParent *assertEqualsParentState

	d     execute.Dataset
	cache execute.TableBuilderCache

	name string
}

type assertEqualsParentState struct {
	id         execute.DatasetID
	mark       execute.Time
	processing execute.Time
	finished   bool
}

func createAssertEqualsTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	if len(a.Parents()) != 2 {
		return nil, nil, errors.New("assertEquals should have exactly 2 parents")
	}

	cache := execute.NewTableBuilderCache(a.Allocator())
	dataset := execute.NewDataset(id, mode, cache)
	pspec, ok := spec.(*AssertEqualsProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", pspec)
	}

	transform := NewAssertEqualsTransformation(dataset, cache, pspec, a.Parents()[0], a.Parents()[1])

	return transform, dataset, nil
}

func NewAssertEqualsTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *AssertEqualsProcedureSpec, gotID, wantID execute.DatasetID) *AssertEqualsTransformation {
	return &AssertEqualsTransformation{
		gotParent:  &assertEqualsParentState{id: gotID},
		wantParent: &assertEqualsParentState{id: wantID},
		d:          d,
		cache:      cache,
		name:       spec.Name,
	}
}

func (t *AssertEqualsTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	panic("not implemented")
}

func (t *AssertEqualsTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	var colMap = make([]int, 0, len(tbl.Cols()))
	var err error
	builder, created := t.cache.TableBuilder(tbl.Key())
	if created {
		colMap, err = execute.AddNewTableCols(tbl, builder, colMap)
		if err != nil {
			return err
		}
		if err := execute.AppendMappedTable(tbl, builder, colMap); err != nil {
			return err
		}
	} else {
		cacheTable, err := builder.Table()
		if err != nil {
			return err
		}
		if ok, err := execute.TablesEqual(cacheTable, tbl); !ok {
			return fmt.Errorf("test %s: tables not equal", t.name)
		} else if err != nil {
			return err
		}
	}

	return nil
}

func (t *AssertEqualsTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	min := mark
	if t.gotParent.id == id {
		t.gotParent.mark = mark
		if t.wantParent.mark < min {
			min = t.wantParent.mark
		}
	} else if t.wantParent.id == id {
		t.wantParent.mark = mark
		if t.gotParent.mark < min {
			min = t.gotParent.mark
		}
	} else {
		return fmt.Errorf("unexpected dataset id: %v", id)
	}

	return t.d.UpdateWatermark(min)
}

func (t *AssertEqualsTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	min := pt
	if t.gotParent.id == id {
		t.gotParent.processing = pt
		if t.wantParent.processing < min {
			min = t.wantParent.processing
		}
	} else if t.wantParent.id == id {
		t.wantParent.processing = pt
		if t.gotParent.processing < min {
			min = t.gotParent.processing
		}
	} else {
		return fmt.Errorf("unexpected dataset id: %v", id)
	}
	return t.d.UpdateProcessingTime(min)
}

func (t *AssertEqualsTransformation) Finish(id execute.DatasetID, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.gotParent.id == id {
		t.gotParent.finished = true
	} else if t.wantParent.id == id {
		t.wantParent.finished = true
	} else {
		t.d.Finish(fmt.Errorf("unexpected dataset id: %v", id))
	}

	if err != nil {
		t.d.Finish(err)
	}

	if t.gotParent.finished && t.wantParent.finished {
		t.d.Finish(nil)
	}
}
