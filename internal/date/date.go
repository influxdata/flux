package date

import (
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/zoneinfo"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func GetTimeInLocation(t values.Time, location string, offset values.Duration) (values.Value, error) {
	if location != "UTC" {
		loc, err := zoneinfo.LoadLocation(location)
		if err != nil {
			return nil, errors.New(codes.Invalid, "invalid location")
		}
		// zone & offset if present
		localTime := values.Time(loc.FromLocalClock(t.Time().UnixNano()))
		localTime = localTime.Add(offset)
		return values.NewTime(localTime), nil
	}
	// only offset is present
	if !offset.IsZero() {
		timeWithOffset := t.Add(offset)
		return values.NewTime(timeWithOffset), nil
	}
	return values.NewTime(t), nil
}

func GetLocationFromObjArgs(args values.Object) (string, values.Duration, error) {
	a := interpreter.NewArguments(args)
	return GetLocationFromFluxArgs(a)
}

func GetLocationFromFluxArgs(args interpreter.Arguments) (string, values.Duration, error) {
	location, err := args.GetRequiredObject("location")
	if err != nil {
		return "UTC", values.ConvertDurationNsecs(0), err
	}
	return GetLocation(location)
}

func GetLocation(location values.Object) (string, values.Duration, error) {
	var (
		name, offset values.Value
		ok           bool
	)

	name, ok = location.Get("zone")
	if !ok {
		return "UTC", values.ConvertDurationNsecs(0), errors.New(codes.Invalid, "zone property missing from location record")
	} else if got := name.Type().Nature(); got != semantic.String {
		return "UTC", values.ConvertDurationNsecs(0), errors.Newf(codes.Invalid, "zone property for location must be of type %s, got %s", semantic.String, got)
	}

	if offset, ok = location.Get("offset"); ok {
		if got := offset.Type().Nature(); got != semantic.Duration {
			return "UTC", values.ConvertDurationNsecs(0), errors.Newf(codes.Invalid, "offset property for location must be of type %s, got %s", semantic.Duration, got)
		}
	}

	if name.IsNull() {
		return "UTC", offset.Duration(), nil
	}
	return name.Str(), offset.Duration(), nil
}
