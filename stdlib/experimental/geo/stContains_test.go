package geo_test

import (
	"context"
	"testing"

	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/experimental/geo"
	"github.com/influxdata/flux/values"
)

func TestSTContains_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "no args",
			Raw:     `import "experimental/geo" geo.ST_Contains()`,
			WantErr: true, // missing required parameter(s)
		},
		{
			Name:    "missing geometry arg",
			Raw:     `import "experimental/geo" geo.ST_Contains(region: { lat: 40.5, lon: -74.5, radius: 15.0 })`,
			WantErr: true, // missing required parameter(s)
		},
		{
			Name:    "invalid args - invalid box",
			Raw:     `import "experimental/geo" geo.ST_Contains(region: { minLat: 40.5, minLon: -74.5, maxLat: 41.5 }, geometry: {lat: 40.5, lon: -74.5})`,
			WantErr: true, // missing maxLon
		},
		{
			Name:    "invalid args - invalid circle",
			Raw:     `import "experimental/geo" geo.ST_Contains(region: { lat: 40.5, radius: 15.0 }, geometry: {lat: 40.5, lon: -74.5})`,
			WantErr: true, // missing lon
		},
		{
			Name:    "invalid args - invalid polygon",
			Raw:     `import "experimental/geo" geo.ST_Contains(region: { points: [{ lat: 40.5, lon: -74.5 }] }, geometry: {lat: 40.5, lon: -74.5})`,
			WantErr: true, // polygon must have at least 3 points
		},
		{
			Name:    "invalid args - unsupported region",
			Raw:     `import "experimental/geo" geo.ST_Contains(region: { x: 1.0, y: 0.0 }, geometry: {lat: 40.5, lon: -74.5})`,
			WantErr: true, // cannot infer region type
		},
		{
			Name:    "invalid args - invalid units",
			Raw:     `import "experimental/geo" geo.ST_Contains(region: { lat: 40.5, lon: -74.5, radius: 15.0 }, geometry: {lat: 40.5, lon: -74.5}, units: { distance: "yd" })`,
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

func TestSTContains_Process(t *testing.T) {
	type box struct {
		minLat float64
		maxLat float64
		minLon float64
		maxLon float64
	}
	type circle struct {
		lat    float64
		lon    float64
		radius float64
	}
	type point struct {
		lat float64
		lon float64
	}
	var defaultUnits = map[string]string{
		"distance": "km",
	}
	testCases := []struct {
		name    string
		box     *box
		circle  *circle
		polygon *[]point
		units   *map[string]string
		lat     float64
		lon     float64
		want    bool
	}{
		{
			name: "box contains",
			box:  &box{minLat: 40.5880775, maxLat: 40.8247008, minLon: -73.80014, maxLon: -73.4630336},
			lat:  40.710594,
			lon:  -73.652183,
			want: true,
		},
		{
			name:   "circle contains",
			circle: &circle{lat: 40.7090214, lon: -73.61846, radius: 15.0},
			lat:    40.710594,
			lon:    -73.652183,
			want:   true,
		},
		{
			name:   "circle contains - m units",
			circle: &circle{lat: 40.7090214, lon: -73.61846, radius: 15000.0},
			lat:    40.710594,
			lon:    -73.652183,
			units:  &map[string]string{"distance": "m"},
			want:   true,
		},
		{
			name: "polygon contains",
			polygon: &[]point{
				{lat: 40.671659, lon: -73.936631},
				{lat: 40.706543, lon: -73.749177},
				{lat: 40.791333, lon: -73.880327},
			},
			lat:  40.702594,
			lon:  -73.909699,
			want: true,
		},
		{
			name: "box not contains",
			box:  &box{minLat: 40.5880775, maxLat: 40.8247008, minLon: -73.80014, maxLon: -73.4630336},
			lat:  40.690732,
			lon:  -74.046267,
			want: false,
		},
		{
			name: "polygon not contains",
			polygon: &[]point{
				{lat: 40.671659, lon: -73.936631},
				{lat: 40.706543, lon: -73.749177},
				{lat: 40.791333, lon: -73.880327},
			},
			lat:  40.6892,
			lon:  -74.0445,
			want: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		stContains := geo.Functions["stContains"]
		var owv values.Object
		if tc.units == nil {
			tc.units = &defaultUnits
		}
		if tc.box != nil {
			owv = values.NewObjectWithValues(map[string]values.Value{
				"region": values.NewObjectWithValues(map[string]values.Value{
					"minLat": values.NewFloat(tc.box.minLat),
					"minLon": values.NewFloat(tc.box.minLon),
					"maxLat": values.NewFloat(tc.box.maxLat),
					"maxLon": values.NewFloat(tc.box.maxLon),
				}),
				"geometry": values.NewObjectWithValues(map[string]values.Value{
					"lat": values.NewFloat(tc.lat),
					"lon": values.NewFloat(tc.lon),
				}),
				"units": unitsToValue(*tc.units),
			})
		} else if tc.circle != nil {
			owv = values.NewObjectWithValues(map[string]values.Value{
				"region": values.NewObjectWithValues(map[string]values.Value{
					"lat":    values.NewFloat(tc.circle.lat),
					"lon":    values.NewFloat(tc.circle.lon),
					"radius": values.NewFloat(tc.circle.radius),
				}),
				"geometry": values.NewObjectWithValues(map[string]values.Value{
					"lat": values.NewFloat(tc.lat),
					"lon": values.NewFloat(tc.lon),
				}),
				"units": unitsToValue(*tc.units),
			})
		} else if tc.polygon != nil {
			array := values.NewArray(semantic.NewArrayType(pointT))
			for _, p := range *tc.polygon {
				array.Append(values.NewObjectWithValues(map[string]values.Value{
					"lat": values.NewFloat(p.lat),
					"lon": values.NewFloat(p.lon),
				}))
			}
			owv = values.NewObjectWithValues(map[string]values.Value{
				"region": values.NewObjectWithValues(map[string]values.Value{
					"points": array,
				}),
				"geometry": values.NewObjectWithValues(map[string]values.Value{
					"lat": values.NewFloat(tc.lat),
					"lon": values.NewFloat(tc.lon),
				}),
				"units": unitsToValue(*tc.units),
			})
		}
		result, err := stContains.Call(context.Background(), owv)
		if err != nil {
			t.Error(err.Error())
		} else if tc.want != result.Bool() {
			t.Errorf("[%s] expected %v (%T), got %v (%T)", tc.name, tc.want, tc.want, result, result)
		}
	}
}
