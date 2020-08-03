package geo_test

import (
	"context"
	"testing"

	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/experimental/geo"
	"github.com/influxdata/flux/values"
)

func TestSTLength_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "no args",
			Raw:     `import "experimental/geo" geo.ST_Length()`,
			WantErr: true, // missing required parameter(s)
		},
		{
			Name:    "invalid args - unsupported geometry",
			Raw:     `import "experimental/geo" geo.ST_Length(geometry: { x: 1.0, y: 0.0 })`,
			WantErr: true, // unsupported geometry type
		},
		{
			Name:    "invalid args - invalid linestring",
			Raw:     `import "experimental/geo" geo.ST_Length(geometry: { linestring: "" })`,
			WantErr: true, // invalid linestring
		},
		{
			Name:    "invalid args - invalid units",
			Raw:     `import "experimental/geo" geo.ST_Length(geometry: {lat: 40.5, lon: -74.5}, units: { distance: "yd" })`,
			WantErr: true, // unsupported unit
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestSTLength_Process(t *testing.T) {
	type point struct {
		lat float64
		lon float64
	}
	var defaultUnits = map[string]string{
		"distance": "km",
	}
	testCases := []struct {
		name  string
		point *point
		line  string
		units *map[string]string
		want  float64
	}{
		{
			name:  "point length",
			point: &point{40.710594, -73.652183},
			want:  0.0,
		},
		{
			name: "linestring length",
			line: "-73.936631 40.671659, -73.749177 40.706543, -73.880327 40.791333",
			want: 30.80,
		},
	}

	for _, tc := range testCases {
		tc := tc
		stLength := geo.Functions["stLength"]
		var owv values.Object
		if tc.units == nil {
			tc.units = &defaultUnits
		}
		if tc.point != nil {
			owv = values.NewObjectWithValues(map[string]values.Value{
				"geometry": values.NewObjectWithValues(map[string]values.Value{
					"lat": values.NewFloat(tc.point.lat),
					"lon": values.NewFloat(tc.point.lon),
				}),
				"units": unitsToValue(*tc.units),
			})
		} else if tc.line != "" {
			owv = values.NewObjectWithValues(map[string]values.Value{
				"geometry": values.NewObjectWithValues(map[string]values.Value{
					"linestring": values.NewString(tc.line),
				}),
				"units": unitsToValue(*tc.units),
			})
		}
		result, err := stLength.Call(context.Background(), owv)
		if err != nil {
			t.Error(err.Error())
		} else if tc.want != roundDistance(result) {
			t.Errorf("[%s] expected %v (%T), got %v (%T)", tc.name, tc.want, tc.want, roundDistance(result), result)
		}
	}
}
