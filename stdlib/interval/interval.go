package interval

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interval"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func init() {
	typ := runtime.MustLookupBuiltinType("interval", "intervals")
	rt, _ := typ.ReturnType()
	arrTyp, _ := rt.ReturnType()
	runtime.RegisterPackageValue("interval", "intervals", values.NewFunction(
		"intervals",
		typ,
		func(ctx context.Context, args values.Object) (values.Value, error) {
			every, err := getDuration(args, "every")
			if err != nil {
				return nil, err
			}
			period, err := getDuration(args, "period")
			if err != nil {
				return nil, err
			}
			offset, err := getDuration(args, "offset")
			if err != nil {
				return nil, err
			}
			w, err := interval.NewWindow(every, period, offset)
			if err != nil {
				return nil, err
			}
			return values.NewFunction("intervals", rt, func(ctx context.Context, args values.Object) (values.Value, error) {
				start, err := getTimeable(ctx, args, "start")
				if err != nil {
					return nil, err
				}
				stop, err := getTimeable(ctx, args, "stop")
				if err != nil {
					return nil, err
				}

				bounds := w.GetOverlappingBounds(start, stop)
				elements := make([]values.Value, len(bounds))
				for i := range bounds {
					elements[i] = values.NewObjectWithValues(map[string]values.Value{
						"start": values.NewTime(bounds[i].Start()),
						"stop":  values.NewTime(bounds[i].Stop()),
					})
				}
				return values.NewArrayWithBacking(arrTyp, elements), nil

			},
				false,
			), nil
		},
		false,
	))
}

func getDuration(args values.Object, name string) (values.Duration, error) {
	v, ok := args.Get(name)
	if !ok {
		return values.Duration{}, errors.Newf(codes.Internal, "unexpected missing argument %q", name)
	}
	if v.Type().Nature() != semantic.Duration {
		return values.Duration{}, errors.Newf(codes.Internal, "unexpected argument %q type %v", name, v.Type())
	}
	return v.Duration(), nil
}
func getTimeable(ctx context.Context, args values.Object, name string) (values.Time, error) {
	t, ok := args.Get(name)
	if !ok {
		return 0, errors.Newf(codes.Internal, "unexpected missing argument %q", name)
	}
	switch t.Type().Nature() {
	case semantic.Time:
		return t.Time(), nil
	case semantic.Duration:
		deps := execute.GetExecutionDependencies(ctx)
		nowTime := *deps.Now
		return values.ConvertTime(nowTime).Add(t.Duration()), nil
	default:
		return 0, errors.Newf(codes.Internal, "unexpected type of argument %q, type: %v", name, t.Type())
	}

}
