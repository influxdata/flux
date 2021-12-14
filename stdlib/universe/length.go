package universe

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// MakeLengthFunc create the "length()" function.
//
// Length will return the length of the given arr array.
func MakeLengthFunc() values.Function {
	return values.NewFunction(
		"length",
		runtime.MustLookupBuiltinType("universe", "length"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			a := interpreter.NewArguments(args)
			v, err := a.GetRequired("arr")
			if err != nil {
				return nil, err
			} else if got := v.Type().Nature(); got != semantic.Array {
				return nil, errors.Newf(codes.Invalid, "arr must be an array, got %s", got)
			}

			maybeArr := v.Array()
			switch maybeArr.(type) {
			// The type system currently conflates TableObject (aka "table stream")
			// with fully-realized arrays of records requiring it to implement
			// the Array interface. Since TableObject will currently panic
			// if the methods provided by this interface are invoked, short-circuit
			// by returning an error.
			// XXX: remove when array/stream are different types <https://github.com/influxdata/flux/issues/4343>
			case values.TableObject:
				return nil, errors.New(codes.Invalid, "arr must be an array, got table stream")
			default:
				l := maybeArr.Len()
				return values.NewInt(int64(l)), nil
			}
		}, false,
	)
}

func init() {
	runtime.RegisterPackageValue("universe", "length", MakeLengthFunc())
}
