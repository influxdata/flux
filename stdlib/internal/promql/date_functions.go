package promql

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/pkg/errors"
)

func generateDateFunction(name string, dateFn func(time.Time) int) values.Function {
	return values.NewFunction(
		name,
		runtime.MustLookupBuiltinType("internal/promql", name),
		func(ctx context.Context, args values.Object) (values.Value, error) {
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
	runtime.RegisterPackageValue("internal/promql", "promqlDayOfMonth", generateDateFunction("promqlDayOfMonth", dayOfMonthFn))
	runtime.RegisterPackageValue("internal/promql", "promqlDayOfWeek", generateDateFunction("promqlDayOfWeek", dayOfWeekFn))
	runtime.RegisterPackageValue("internal/promql", "promqlDaysInMonth", generateDateFunction("promqlDaysInMonth", daysInMonthFn))
	runtime.RegisterPackageValue("internal/promql", "promqlHour", generateDateFunction("promqlHour", hourFn))
	runtime.RegisterPackageValue("internal/promql", "promqlMinute", generateDateFunction("promqlMinute", minuteFn))
	runtime.RegisterPackageValue("internal/promql", "promqlMonth", generateDateFunction("promqlMonth", monthFn))
	runtime.RegisterPackageValue("internal/promql", "promqlYear", generateDateFunction("promqlYear", yearFn))
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
