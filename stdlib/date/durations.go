package date

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/date"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/function"
	"github.com/influxdata/flux/internal/zoneinfo"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func init() {
	pkg := function.ForPackage("date")
	pkg.RegisterContext("_add", Add)
	pkg.RegisterContext("_sub", Sub)
	pkg.Register("scale", Scale)
}

func Add(ctx context.Context, args *function.Arguments) (values.Value, error) {
	d, err := args.GetRequired("d")
	if err != nil {
		return nil, err
	}

	t, err := args.GetRequired("to")
	if err != nil {
		return nil, err
	}

	loc, err := args.GetRequiredObject("location")
	if err != nil {
		return nil, err
	}
	return addDuration(ctx, t, d.Duration(), 1, loc)
}

func Sub(ctx context.Context, args *function.Arguments) (values.Value, error) {
	d, err := args.GetRequired("d")
	if err != nil {
		return nil, err
	}

	t, err := args.GetRequired("from")
	if err != nil {
		return nil, err
	}

	loc, err := args.GetRequiredObject("location")
	if err != nil {
		return nil, err
	}
	return addDuration(ctx, t, d.Duration(), -1, loc)
}

func addDuration(ctx context.Context, t values.Value, d values.Duration, scale int, loc values.Object) (values.Value, error) {
	deps := execute.GetExecutionDependencies(ctx)
	time, err := deps.ResolveTimeable(t)
	if err != nil {
		return nil, err
	}

	name, offset, err := date.GetLocation(loc)
	if err != nil {
		return nil, err
	}

	if !offset.IsZero() {
		// The offset for a timezone is the duration difference
		// between the local time and utc at the same clock time.
		// To convert to utc from our local clock, we apply this offset.
		time = time.Add(offset)
	}

	if scale != 1 {
		d = d.Mul(scale)
	}

	if name != "UTC" {
		loc, err := zoneinfo.LoadLocation(name)
		if err != nil {
			return nil, err
		}

		utc := values.Time(loc.FromLocalClock(int64(time)))
		utc = utc.Add(d)
		time = values.Time(loc.ToLocalClock(int64(utc)))
	} else {
		time = time.Add(d)
	}

	if !offset.IsZero() {
		// Need to reverse the location offset to
		// go back to the local clock.
		time = time.Add(offset.Mul(-1))
	}
	return values.NewTime(time), nil
}

func Scale(args *function.Arguments) (values.Value, error) {
	d, err := args.GetRequired("d")
	if err != nil {
		return nil, err
	} else if d.Type().Nature() != semantic.Duration {
		return nil, errors.Newf(codes.Invalid, "keyword argument %q should be of kind %v, but got %v", "scale", semantic.Duration, d.Type().Nature())
	}

	n, err := args.GetRequiredInt("n")
	if err != nil {
		return nil, err
	}

	if n == 1 {
		return d, nil
	}
	return values.NewDuration(d.Duration().Mul(int(n))), nil
}
