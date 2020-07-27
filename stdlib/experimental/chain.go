package experimental

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/lang/execdeps"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

func init() {
	runtime.RegisterPackageValue("experimental", "chain", MakeChainFunction())
}

func MakeChainFunction() values.Function {
	chainSignature := runtime.MustLookupBuiltinType("experimental", "chain")
	return values.NewFunction("chain", chainSignature, chainCall, false)
}

func chainCall(ctx context.Context, args values.Object) (values.Value, error) {
	arguments := interpreter.NewArguments(args)

	first, err := arguments.GetRequired("first")
	if err != nil {
		return nil, err
	}

	second, err := arguments.GetRequired("second")
	if err != nil {
		return nil, err
	}

	compiler := lang.TableObjectCompiler{
		Tables: first.(*flux.TableObject),
	}

	program, err := compiler.Compile(ctx)
	if err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "error in table object compilation")
	}

	if !execdeps.HaveExecutionDependencies(ctx) {
		return nil, errors.New(codes.Internal, "no execution context for chain to use")
	}
	deps := execdeps.GetExecutionDependencies(ctx)

	if program, ok := program.(lang.LoggingProgram); ok {
		program.SetLogger(deps.Logger)
	}
	query, err := program.Start(ctx, deps.Allocator)
	if err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "error in table object start")
	}

	for res := range query.Results() {
		if err := res.Tables().Do(func(table flux.Table) error {
			defer table.Done()
			return nil
		}); err != nil {
			return nil, err
		}
	}

	deps.Metadata.AddAll(query.Statistics().Metadata)

	return second, nil
}
