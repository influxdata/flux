package random

import (
	"fmt"
	"math/rand"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/pkg/errors"
)

func randomUInt64() values.Function {
	return values.NewFunction(
		"uint64",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{"max": semantic.Int},
			Required:   semantic.LabelSet{"max"},
			Return:     semantic.UInt,
		}),
		func(args values.Object) (values.Value, error) {
			v1, ok := args.Get("max")
			if !ok {
				return nil, errors.New("missing argument max")
			}

			if v1.Type().Nature() == semantic.Int {
				return values.NewUInt(uint64(rand.Int63n(v1.Int()))), nil
			}

			return nil, fmt.Errorf("cannot convert argument max of type %v to uint", v1.Type().Nature())
		}, false,
	)
}

func init() {
	flux.RegisterPackageValue("random", "uint64", randomUInt64())
}
