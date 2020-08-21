package executetest

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/mock"
	"github.com/influxdata/flux/plan"
	uuid "github.com/gofrs/uuid"
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
	// uuid.NewV4 can return an error because of enthropy. We will stick with the previous
	// behavior of panicing on errors when creating new uuid's
	id := execute.DatasetID(uuid.Must(uuid.NewV4()))

	if len(src.ts) == 0 {
		return
	} else if len(src.ts) == 1 {
		t := src.ts[0]

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
		return
	}

	buffers := make([]flux.BufferedTable, 0, len(src.data))
	for _, tbl := range src.data {
		bufTable, _ := execute.CopyTable(tbl)
		buffers = append(buffers, bufTable.(flux.BufferedTable))
	}

	// Ensure that the buffers are released after the source has finished.
	defer func() {
		for _, tbl := range buffers {
			tbl.Done()
		}
	}()

	for _, t := range src.ts {
		var max execute.Time
		for _, tbl := range buffers {
			t.Process(id, tbl.Copy())
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
	// Allocate the amount of memory as specified in the byte count.
	// This memory is not used or returned.
	_ = s.alloc.Allocate(s.ByteCount)
}

func (s *AllocatingFromProcedureSpec) AddTransformation(t execute.Transformation) {
	s.ts = append(s.ts, t)
}

// Some sources are located by a URL. e.g. sql.from, socket.from
// the URL/DSN supplied by the user need to be validated by a URLValidator{}
// before we can establish the connection.
// This struct (as well as the Run() method) acts as a test harness for that.
type SourceUrlValidationTestCases []struct {
	Name   string
	Spec   plan.ProcedureSpec
	V      url.Validator
	ErrMsg string
}

func (testCases *SourceUrlValidationTestCases) Run(t *testing.T, fn execute.CreateSource) {
	for _, tc := range *testCases {
		deps := dependenciestest.Default()
		if tc.V != nil {
			deps.Deps.URLValidator = tc.V
		}
		ctx := deps.Inject(context.Background())
		a := mock.AdministrationWithContext(ctx)
		t.Run(tc.Name, func(t *testing.T) {
			id := RandomDatasetID()
			_, err := fn(tc.Spec, id, a)
			if tc.ErrMsg != "" {
				if err == nil {
					t.Errorf("Expect an error with message \"%s\", but did not get one.", tc.ErrMsg)
				} else {
					if !strings.Contains(err.Error(), tc.ErrMsg) {
						t.Fatalf("unexpected result -want/+got:\n%s",
							cmp.Diff(err.Error(), tc.ErrMsg))
					}
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect to get an error, but got %v", err)
				}
			}
		})
	}
}
