package date

import (
	"github.com/mvn-trinhnguyen2-dn/flux/codes"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/errors"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/zoneinfo"
	"github.com/mvn-trinhnguyen2-dn/flux/values"
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
