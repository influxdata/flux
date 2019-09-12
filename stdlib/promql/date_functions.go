package promql

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/pkg/errors"
)

func generateDateFunction(name string, dateFn func(time.Time) int) values.Function {
	return values.NewFunction(
		name,
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{"timestamp": semantic.Float},
			Required:   semantic.LabelSet{"timestamp"},
			Return:     semantic.Float,
		}),
		func(ctx context.Context, deps dependencies.Interface, args values.Object) (values.Value, error) {
			v, ok := args.Get("timestamp")
			if !ok {
				return nil, errors.New("missing argument timestamp")
			}

			if v.Type().Nature() != semantic.Float {
				return nil, fmt.Errorf("cannot convert argument of type %v to float", v.Type().Nature())
			}

			t := values.Time(v.Float() * 1e9).Time()
			return values.NewFloat(float64(dateFn(t))), nil
		}, false,
	)
}

func init() {
	flux.RegisterPackageValue("promql", "promqlDayOfMonth", generateDateFunction("promqlDayOfMonth", dayOfMonthFn))
	flux.RegisterPackageValue("promql", "promqlDayOfWeek", generateDateFunction("promqlDayOfWeek", dayOfWeekFn))
	flux.RegisterPackageValue("promql", "promqlDaysInMonth", generateDateFunction("promqlDaysInMonth", daysInMonthFn))
	flux.RegisterPackageValue("promql", "promqlHour", generateDateFunction("promqlHour", hourFn))
	flux.RegisterPackageValue("promql", "promqlMinute", generateDateFunction("promqlMinute", minuteFn))
	flux.RegisterPackageValue("promql", "promqlMonth", generateDateFunction("promqlMonth", monthFn))
	flux.RegisterPackageValue("promql", "promqlYear", generateDateFunction("promqlYear", yearFn))
}

func dayOfMonthFn(t time.Time) int {
	return t.Day()
}

func dayOfWeekFn(t time.Time) int {
	return int(t.Weekday())
}

func daysInMonthFn(t time.Time) int {
	return 32 - time.Date(t.Year(), t.Month(), 32, 0, 0, 0, 0, time.UTC).Day()
}

func hourFn(t time.Time) int {
	return t.Hour()
}

func minuteFn(t time.Time) int {
	return t.Minute()
}

func monthFn(t time.Time) int {
	return int(t.Month())
}

func yearFn(t time.Time) int {
	return t.Year()
}
