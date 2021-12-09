package locationutil

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/values"
)

type Location struct {
	Name   string
	Offset flux.Duration
}

func (l Location) IsUTC() bool {
	name := l.Name
	if name == "" {
		name = "UTC"
	}
	return name == "UTC" && l.Offset.IsZero()
}

func (l Location) Load() (LocationZone, error) {
	name := l.Name
	if name == "" {
		name = "UTC"
	}
	loc, err := LoadLocation(name)
	if err != nil {
		return LocationZone{}, err
	}
	loc.Offset = l.Offset
	return loc, nil
}

func (l Location) ToUTCTime(t values.Time) values.Value {
	// load location
	LocationZone, err := l.Load()
	if err != nil {
		return values.NewTime(t)
	}
	// zone & offset if present
	if LocationZone.zone != nil {
		localTime := values.Time(LocationZone.zone.FromLocalClock(t.Time().UnixNano()))
		localTime = localTime.Add(l.Offset)
		return values.NewTime(localTime)
	}
	// only offset is present
	if !l.Offset.IsZero() {
		timeWithOffset := t.Add(LocationZone.Offset)
		return values.NewTime(timeWithOffset)
	}
	return values.NewTime(t)
}
