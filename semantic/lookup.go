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

// LookupBuiltInType returns the type of the builtin value for a given
// Flux stdlib package. Returns an error if lookup fails.
func LookupBuiltInType(pkg, name string) (MonoType, error) {
	var table flatbuffers.Table
	prop := stdlibTypeEnvironment[pkg][name] // query environment map for prop

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

func TypeEnvMap(env *fbsemantic.TypeEnvironment) map[string]map[string]*fbsemantic.Prop {
	envMap := make(map[string]map[string]*fbsemantic.Prop)
	var table flatbuffers.Table
	l := env.AssignmentsLength()

	for i := 0; i < l; i++ {
		newAssign := new(fbsemantic.TypeAssignment)
		_ = env.Assignments(newAssign, i) // this call assigns a value to newAssign
		assignId := string(newAssign.Id())
		polytype := newAssign.Ty(nil)
		if polytype.ExprType() != fbsemantic.MonoTypeRow {
			panic(fmt.Errorf(
				"Expected PolyType Expr of %v to be fbsemantic.MonoTypeRow; found fbsemantic.%v",
				assignId,
				fbsemantic.EnumNamesMonoType[polytype.ExprType()],
			))
		}
		if !polytype.Expr(&table) {
			panic(fmt.Errorf("PolyType does not have a MonoType; something went wrong %v", string(polytype.ExprType())))
		}

		// initialize table before use in row
		row := new(fbsemantic.Row)
		row.Init(table.Bytes, table.Pos)
		propLen := row.PropsLength()
		propMap := make(map[string]*fbsemantic.Prop)

		for j := 0; j < propLen; j++ {
			newProp := new(fbsemantic.Prop)
			_ = row.Props(newProp, j) // this call assigns value to newProp
			propKey := string(newProp.K())
			propMap[propKey] = newProp
		}
		envMap[assignId] = propMap
	}
	return envMap
}
