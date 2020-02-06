package semantic

import (
	"fmt"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/libflux/go/libflux"
	"github.com/influxdata/flux/semantic/internal/fbsemantic"
)

var stdlibTypeEnvironment = TypeEnvMap(fbsemantic.GetRootAsTypeEnvironment(libflux.EnvStdlib(), 0))

type envKey struct {
	Package string
	Prop    string
}

// LookupBuiltinType returns the type of the builtin value for a given
// Flux stdlib package. Returns an error if lookup fails.
func LookupBuiltinType(pkg, name string) (MonoType, error) {
	key := envKey{
		Package: pkg,
		Prop:    name,
	}
	prop, ok := stdlibTypeEnvironment[key]
	if !ok {
		return MonoType{}, errors.Newf(codes.Internal, "Expected to find Prop for %v %v, but Prop was missing.", pkg, name)
	}
	var table flatbuffers.Table
	if !prop.V(&table) {
		return MonoType{}, errors.Newf(codes.Internal, "Prop value is not valid: pkg %v name %v", pkg, name)
	}
	monotype, err := NewMonoType(table, prop.VType())
	if err != nil {
		return MonoType{}, err
	}
	// return fb polytype within semantic wrapper
	return monotype, nil
}

// MustLookupBuiltinType validates that call to LookupBuiltInType was
// successful. If there is an error with lookup, then panic.
func MustLookupBuiltinType(pkg, name string) MonoType {
	mt, err := LookupBuiltinType(pkg, name)
	if err != nil {
		panic(err)
	}
	return mt
}

// TypeEnvMap creates a global map of the TypeEnvironment
func TypeEnvMap(env *fbsemantic.TypeEnvironment) map[envKey]*fbsemantic.Prop {
	envMap := make(map[envKey]*fbsemantic.Prop)
	var table flatbuffers.Table
	l := env.AssignmentsLength()

	for i := 0; i < l; i++ {
		newAssign := new(fbsemantic.TypeAssignment)
		_ = env.Assignments(newAssign, i) // this call assigns a value to newAssign
		assignId := string(newAssign.Id())
		polytype := newAssign.Ty(nil)
		if polytype.ExprType() != fbsemantic.MonoTypeRow {
			panic(fmt.Errorf("expected PolyType Expr of %v to be fbsemantic.MonoTypeRow; found fbsemantic.%v",
				assignId,
				fbsemantic.EnumNamesMonoType[polytype.ExprType()],
			))
		}
		if !polytype.Expr(&table) {
			panic(fmt.Errorf(
				"PolyType does not have a MonoType; something went wrong. Assignment: %v MonoType: %v",
				assignId,
				fbsemantic.EnumNamesMonoType[polytype.ExprType()],
			))
		}

		// initialize table before use in row
		row := new(fbsemantic.Row)
		row.Init(table.Bytes, table.Pos)
		propLen := row.PropsLength()

		for j := 0; j < propLen; j++ {
			newProp := new(fbsemantic.Prop)
			_ = row.Props(newProp, j) // this call assigns value to newProp
			propKey := string(newProp.K())
			key := envKey{
				Package: assignId,
				Prop:    propKey,
			}
			envMap[key] = newProp
		}

	}
	return envMap
}
