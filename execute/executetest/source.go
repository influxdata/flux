package executetest

import (
	"context"

	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	uuid "github.com/satori/go.uuid"
)

const FromTestKind = "from-test"

// FromProcedureSpec is a procedure spec AND an execution Node.
// It simulates the execution of a basic physical scan operation.
type FromProcedureSpec struct {
	data []*Table
	ts   []execute.Transformation
}

// NewFromProcedureSpec specifies a from-test procedure with source data
func NewFromProcedureSpec(data []*Table) *FromProcedureSpec {
	// Normalize data before anything can read it
	for _, tbl := range data {
		tbl.Normalize()
	}
	return &FromProcedureSpec{data: data}
}

func (src *FromProcedureSpec) Kind() plan.ProcedureKind {
	return FromTestKind
}

func (src *FromProcedureSpec) Copy() plan.ProcedureSpec {
	return src
}

func (src *FromProcedureSpec) Cost(inStats []plan.Statistics) (plan.Cost, plan.Statistics) {
	return plan.Cost{}, plan.Statistics{}
}

func (src *FromProcedureSpec) AddTransformation(t execute.Transformation) {
	src.ts = append(src.ts, t)
}

func (src *FromProcedureSpec) Run(ctx context.Context) {
	id := execute.DatasetID(uuid.NewV4())
	for _, t := range src.ts {
		var max execute.Time
		for _, tbl := range src.data {
			t.Process(id, tbl)
			stopIdx := execute.ColIdx(execute.DefaultStopColLabel, tbl.Cols())
			if stopIdx >= 0 {
				if s := tbl.Key().ValueTime(stopIdx); s > max {
					max = s
				}
			}
		}
		t.UpdateWatermark(id, max)
		t.Finish(id, nil)
	}
}

func CreateFromSource(spec plan.ProcedureSpec, id execute.DatasetID, a execute.Administration) (execute.Source, error) {
	return spec.(*FromProcedureSpec), nil
}

// AllocatingFromProcedureSpec is a procedure spec AND an execution node
// that allocates ByteCount bytes during execution.
type AllocatingFromProcedureSpec struct {
	ByteCount int

	alloc *memory.Allocator
	ts    []execute.Transformation
}

const AllocatingFromTestKind = "allocating-from-test"

func (AllocatingFromProcedureSpec) Kind() plan.ProcedureKind {
	return AllocatingFromTestKind
}

func (s *AllocatingFromProcedureSpec) Copy() plan.ProcedureSpec {
	return &AllocatingFromProcedureSpec{
		ByteCount: s.ByteCount,
		alloc:     s.alloc,
	}
}

func (AllocatingFromProcedureSpec) Cost(inStats []plan.Statistics) (cost plan.Cost, outStats plan.Statistics) {
	return plan.Cost{}, plan.Statistics{}
}

func CreateAllocatingFromSource(spec plan.ProcedureSpec, id execute.DatasetID, a execute.Administration) (execute.Source, error) {
	s := spec.(*AllocatingFromProcedureSpec)
	s.alloc = a.Allocator()
	return s, nil
}

func (s *AllocatingFromProcedureSpec) Run(ctx context.Context) {
	if err := s.alloc.Allocate(s.ByteCount); err != nil {
		panic(err)
	}
}

func (s *AllocatingFromProcedureSpec) AddTransformation(t execute.Transformation) {
	s.ts = append(s.ts, t)
}
