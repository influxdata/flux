package universe

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
)

type DualImplProcedureSpec struct {
	plan.ProcedureSpec
	plan.DefaultCost
	UseDeprecated bool
}

func (s *DualImplProcedureSpec) Kind() plan.ProcedureKind {
	return s.ProcedureSpec.Kind()
}

func (s *DualImplProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(DualImplProcedureSpec)
	*ns = *s
	return ns
}

func UseDeprecatedImpl(spec plan.ProcedureSpec) {
	if dualImplSpec, ok := spec.(*DualImplProcedureSpec); ok {
		dualImplSpec.UseDeprecated = true
	}
}

func newDualImplSpec(fn plan.CreateProcedureSpec) plan.CreateProcedureSpec {
	return func(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
		spec, err := fn(qs, pa)
		if err != nil {
			return nil, err
		}
		return &DualImplProcedureSpec{
			ProcedureSpec: spec,
			UseDeprecated: false,
		}, nil
	}
}

func createDualImplTf(fnNew execute.CreateTransformation, fnDeprecated execute.CreateTransformation) execute.CreateTransformation {
	return func(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
		dualImplSpec, ok := spec.(*DualImplProcedureSpec)
		if !ok {
			return fnNew(id, mode, spec, a)
		}
		if !dualImplSpec.UseDeprecated {
			return fnNew(id, mode, dualImplSpec.ProcedureSpec, a)
		}
		return fnDeprecated(id, mode, dualImplSpec.ProcedureSpec, a)
	}
}
