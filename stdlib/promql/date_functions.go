package promql

import (
	"fmt"
	"time"

	"github.com/influxdata/flux"
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
		func(args values.Object) (values.Value, error) {
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
	flux.RegisterPackageValue("promql", "dayOfMonth", generateDateFunction("dayOfMonth", dayOfMonthFn))
	flux.RegisterPackageValue("promql", "dayOfWeek", generateDateFunction("dayOfWeek", dayOfWeekFn))
	flux.RegisterPackageValue("promql", "daysInMonth", generateDateFunction("daysInMonth", daysInMonthFn))
	flux.RegisterPackageValue("promql", "hour", generateDateFunction("hour", hourFn))
	flux.RegisterPackageValue("promql", "minute", generateDateFunction("minute", minuteFn))
	flux.RegisterPackageValue("promql", "month", generateDateFunction("month", monthFn))
	flux.RegisterPackageValue("promql", "year", generateDateFunction("year", yearFn))
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
