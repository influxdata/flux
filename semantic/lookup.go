package semantic

import (
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/libflux/go/libflux"
	"github.com/influxdata/flux/semantic/internal/fbsemantic"
)

// LookupBuiltInType returns the type of the builtin value for a given
// Flux stdlib package. Returns a empty MonoType if lookup fails.
func LookupBuiltInType(pkg, name string) (MonoType, error) {
	byteArr := libflux.EnvStdlib()
	env := fbsemantic.GetRootAsTypeEnvironment(byteArr, 0)
	var table flatbuffers.Table
	// perform a lookup of package identifier in flatbuffers TypeEnvironment
	prop, err := lookup(env, pkg, name)
	if err != nil {
		return MonoType{}, err
	}
	if !prop.V(&table) {
		return MonoType{}, errors.Newf(codes.Internal, "Prop value is not valid: %v", err)
	}
	monotype, err := NewMonoType(table, prop.VType())
	if err != nil {
		return MonoType{}, err
	}
	// return fb polytype within semantic wrapper
	return monotype, nil
}

// lookup is a helper function that performs a lookup of a package identifier in
// a flatbuffers TypeEnvironment. It first checks for the package using path string
// and then checks for the package identifier, the "name" string in this case, returning
// the corresponding MonoType if found
func lookup(env *fbsemantic.TypeEnvironment, pkg, name string) (*fbsemantic.Prop, error) {
	// Find package
	typeAssign := new(fbsemantic.TypeAssignment)
	if pkgErr := foundPackage(env, typeAssign, pkg); pkgErr != nil {
		return nil, pkgErr
	}

	// Grab PolyType expr; expr should always be of type MonoTypeRow
	polytype := typeAssign.Ty(nil)
	if polytype.ExprType() != fbsemantic.MonoTypeRow {
		return nil, errors.Newf(
			codes.Internal,
			"Expected PolyType expr to be fbsemantic.MonoTypeRow; found %v",
			polytype.ExprType(),
		)
	}
	var table flatbuffers.Table
	if !polytype.Expr(&table) {
		return nil, errors.New(codes.Internal, "PolyType expr is not valid")
	}

	// check for package identifier in Row Props
	row := new(fbsemantic.Row)
	row.Init(table.Bytes, table.Pos)
	prop := new(fbsemantic.Prop)
	if propErr := foundProp(row, prop, name); propErr != nil {
		return nil, propErr
	}

	return prop, nil
}

// foundPackage is a helper function that iterates over type assignments and checks
// for a given package returning an error if the package is not found
func foundPackage(env *fbsemantic.TypeEnvironment, obj *fbsemantic.TypeAssignment, pkg string) error {
	l := env.AssignmentsLength()
	for i := 0; i < l; i++ {
		if !env.Assignments(obj, i) {
			return errors.Newf(codes.Internal, "package %v not found; last position %v", pkg, i)
		} else {
			if string(obj.Id()) == pkg {
				return nil
			}
		}
	}
	return errors.Newf(codes.Internal, "package not found %v", pkg)
}

// foundProp is a helper function that iterates over row properties and checks
// for a given package identifier, returning an error if the identifier is not found
func foundProp(row *fbsemantic.Row, obj *fbsemantic.Prop, name string) error {
	l := row.PropsLength()
	for i := 0; i < l; i++ {
		if !row.Props(obj, i) {
			return errors.Newf(codes.Internal, "package identifier %v not found; last position %v", name, i)
		} else {
			if string(obj.K()) == name {
				return nil
			}
		}
	}
	return errors.Newf(codes.Internal, "package identifier not found %v", name)
}