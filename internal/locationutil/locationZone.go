package locationutil

import (
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/zoneinfo"
	"github.com/influxdata/flux/values"
)

// UTC is the UTC zone with no additional offset.
var UTC = LocationZone{}

type LocationZone struct {
	zone *zoneinfo.Location

	// Offset declares an additional offset that will be applied.
	Offset values.Duration
}

func LoadLocation(name string) (LocationZone, error) {
	if name == "UTC" {
		return UTC, nil
	}

	loc, err := zoneinfo.LoadLocation(name)
	if err != nil {
		return UTC, errors.Wrap(err, codes.Invalid)
	}
	return LocationZone{
		zone: loc,
	}, nil
}

func (l LocationZone) Equal(other LocationZone) bool {
	if l.zone == nil && other.zone == nil {
		return l.Offset == other.Offset
	} else if l.zone == nil || other.zone == nil {
		return false
	}
	return l.zone.String() == other.zone.String() && l.Offset == other.Offset
}
