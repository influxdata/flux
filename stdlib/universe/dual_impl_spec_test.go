package universe

import (
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/mock"
	"github.com/influxdata/flux/plan"
)

type mockOperationSpec struct{}

func (m mockOperationSpec) Kind() flux.OperationKind {
	return "mock"
}

type mockAdministration struct{}

func (m mockAdministration) Now() time.Time {
	return time.Now()
}

type MockProcedureSpec struct {
	plan.DefaultCost
}

func (MockProcedureSpec) Kind() plan.ProcedureKind {
	return "mock"
}

func (MockProcedureSpec) Copy() plan.ProcedureSpec {
	return MockProcedureSpec{}
}

func NewMockProcedureSpec(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	return MockProcedureSpec{}, nil
}

func TestDualImplProcedureSpec(t *testing.T) {
	spec, err := newDualImplSpec(NewMockProcedureSpec)(mockOperationSpec{}, mockAdministration{})
	if err != nil {
		t.Fatal(err)
	}
	usedDeprecated := true
	fnNew := func(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
		if _, ok := spec.(*DualImplProcedureSpec); ok {
			t.Fatal("spec received in fnNew should not be DualImplProcedureSpec")
		}
		usedDeprecated = false
		return nil, nil, nil
	}
	fnDeprecated := func(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
		if _, ok := spec.(*DualImplProcedureSpec); ok {
			t.Fatal("spec received in fnDeprecated should not be DualImplProcedureSpec")
		}
		usedDeprecated = true
		return nil, nil, nil
	}
	fn := createDualImplTf(fnNew, fnDeprecated)
	fn(executetest.RandomDatasetID(), execute.DiscardingMode, spec, &mock.Administration{})
	if usedDeprecated {
		t.Fatal("Expected not to use the deprecated transformation, but it did.")
	}
	UseDeprecatedImpl(spec)
	fn(executetest.RandomDatasetID(), execute.DiscardingMode, spec, &mock.Administration{})
	if !usedDeprecated {
		t.Fatal("Expected to use the deprecated transformation, but it did not.")
	}
	fn(executetest.RandomDatasetID(), execute.DiscardingMode, MockProcedureSpec{}, &mock.Administration{})
	if usedDeprecated {
		t.Fatal("Expected not to use the deprecated transformation, but it did.")
	}
}
