package flux

import (
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/libflux/go/libflux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/semantic/internal/fbsemantic"
	// "github.com/influxdata/flux/internal/errors"
)

//LookupBuiltInType returns the type of the builtin value for a given Flux package.
// pkg is path
// name is the func name

func LookupBuiltInType(pkg, name string) semantic.PolyType {
	byteArr := libflux.EnvStdlib()

	env := fbsemantic.GetRootAsTypeEnvironment(byteArr, 0)

	// perform a lookup on flatbuffers TypeEnvironment
	fb, err := lookup(env, pkg, name)
	if err != nil {
		return nil
	}

	// return fb polytype within semantic wrapper
	return semantic.PolyType{fb}
}

func lookup(env *fbsemantic.TypeEnvironment, pkg, name string) (*fbsemantic.PolyType, error) {
	// check for package
	typeAssign := new(fbsemantic.TypeAssignment)
	if pkgErr := foundPackage(env, typeAssign, pkg); pkgErr != nil {
		return nil, pkgErr
	}

	polytype := typeAssign.Ty(nil)
	// grab PolyType expr
	if polytype.ExprType() != fbsemantic.Row {
		// expr is monotype of type row
		return errors.Newf(codes.Internal, "")
	}

	// create row type
	table := new(flatbuffers.Table)
	if !polytype.Expr(table) {
		return errors.Newf(codes.Internal, "")
	}
	row := new(fbsemantic.MonoTypeRow)
	row.Init(table.Bytes, table.Pos)

	// check for package identifier in row props
	prop := new(fbsemantic.Prop)
	if propErr := foundProp(row, prop, name); propErr != nil {
		return nil, propErr
	}

	// when found return monotype
	return prop.V()
}

// iterate over type assignments and check for correct package
func foundPackage(env *fbsemantic.TypeEnvironment, obj *fbsemantic.TypeAssignment, pkg string) error {
	l := env.AssignmentsLength()
	for i := 0; i < l; i++ {
		if !env.Assignments(obj, i) {
			return errors.Newf(codes.Internal, "", i)
		} else {
			if string(obj.Id()) == pkg {
				return nil
			}
		}
	}
	return errors.Newf(codes.Internal, "")
}

// iterate over row properties and check for correct package identifier
func foundProp(row *fbsemantic.Row, obj *fbsemantic.Prop, name string) error {
	l := row.PropsLength()
	for i := 0; i < l; i++ {
		if !row.Props(obj, i) {
			return errors.Newf(codes.Internal, "", i)
		} else {
			if string(obj.Id()) == name {
				return nil
			}
		}
	}
	return errors.Newf(codes.Internal, "")
}
