package execkit

import (
	"reflect"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/function"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

// Transformation is a method of transforming a Table or Tables into another Table.
// The Transformation is kept at a bare-minimum to keep it simple.
// It contains one method which tells it to process the next message received from an upstream.
// The Message can then be typecast into the proper underlying message type.
//
// It is recommended to use one of the Transformation types that implement a specific type
// of transformation.
//
// For backwards compatibility, Transformation also implements execute.Transformation.
type Transformation interface {
	execute.Transformation
	execute.Transport
}

// TransformationSpec defines a spec for creating a transformation.
type TransformationSpec interface {
	plan.ProcedureSpec
	// CreateTransformation will construct a transformation
	// and dataset using the given dataset and administration object.
	CreateTransformation(id execute.DatasetID, a execute.Administration) (execute.Transformation, execute.Dataset, error)
}

type specFactory struct {
	t reflect.Type
}

func (f *specFactory) createOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	ptr := reflect.New(f.t)
	if err := function.ReadArgs(ptr.Interface(), args, a); err != nil {
		return nil, err
	}
	return &operationSpec{
		spec: ptr.Interface().(plan.ProcedureSpec),
	}, nil
}

type operationSpec struct {
	spec plan.ProcedureSpec
}

func (o *operationSpec) Kind() flux.OperationKind {
	return flux.OperationKind(o.spec.Kind())
}

func newProcedureSpec(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*operationSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return spec.spec, nil
}

func createTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(TransformationSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return s.CreateTransformation(id, a)
}

// RegisterTransformation will register a TransformationSpec
// in the pkgpath with name. The TransformationSpec is read from
// arguments using ReadArgs.
//
// The operation spec and procedure spec will be automatically generated
// for this transformation and the kind will be a concatenation of
// the pkgpath and name separated by a dot.
func RegisterTransformation(spec TransformationSpec) {
	kind := spec.Kind()
	pkgpath, name := inferFromKind(kind)
	factory := &specFactory{
		t: reflect.TypeOf(spec).Elem(),
	}
	signature := runtime.MustLookupBuiltinType(pkgpath, name)
	fn := flux.MustValue(flux.FunctionValue(name, factory.createOpSpec, signature))
	runtime.RegisterPackageValue(pkgpath, name, fn)
	plan.RegisterProcedureSpec(kind, newProcedureSpec, flux.OperationKind(kind))
	execute.RegisterTransformation(kind, createTransformation)
}

// defaultPkgPath is the default package imported into the scope.
const defaultPkgPath = "universe"

// inferFromKind will infer the package path and name of the function
// from the procedure kind. The procedure kind should be in the form
// of "pkgpath.name".
func inferFromKind(kind plan.ProcedureKind) (string, string) {
	parts := strings.SplitN(string(kind), ".", 2)
	if len(parts) == 1 {
		return defaultPkgPath, parts[0]
	}
	return parts[0], parts[1]
}
