package geo_test

import (
	"context"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/experimental/geo"
	"github.com/influxdata/flux/values"
	"testing"
)

func TestContainsLatLon_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "no args",
			Raw:     `import "experimental/geo" geo.containsLatLon()`,
			WantErr: true, // missing required parameter(s)
		},
		{
			Name:    "missing lat lon",
			Raw:     `import "experimental/geo" geo.containsLatLon(circle: { lat: 40.5, lon: -74.5, radius: 15.0 })`,
			WantErr: true, // missing required parameter(s)
		},
		{
			Name:    "invalid args - polygon",
			Raw:     `import "experimental/geo" geo.containsLatLon(polygon: { points: [{ lat: 40.5, lon: -74.5 }] }, lat: 40.5, lon: -74.5)`,
			WantErr: true, // invalid polygon specification - must have at least 3 points
		},
		{
			Name:    "invalid args - more than one region",
			Raw:     `import "experimental/geo" geo.containsLatLon(box: { minLat: 40.5, minLon: -74.5, maxLat: 41.5, maxLon: -73.5 }, circle: { lat: 40.5, lon: -74.5, radius: 15.0 }, lat: 40.5, lon: -74.5)`,
			WantErr: true, // either ...
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

func TestContainsLatLon_Process(t *testing.T) {
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
	testCases := []struct {
		name     string
		box      *box
		circle   *circle
		polygon  *[]point
		lat  float64
		lon  float64
		want     bool
	}{
		{
			name:     "box contains",
			box:      &box{minLat: 40.5880775, maxLat: 40.8247008, minLon: -73.80014, maxLon: -73.4630336},
			lat:  40.710594,
			lon:  -73.652183,
			want:     true,
		},
		{
			name:     "circle contains",
			circle:   &circle{lat: 40.7090214, lon: -73.61846, radius: 15.0},
			lat:  40.710594,
			lon:  -73.652183,
			want:     true,
		},
		{
			name: "polygon contains",
			polygon: &[]point{
				{lat: 40.776527, lon: -73.338811},
				{lat: 40.788093, lon: -73.776396},
				{lat: 40.475939, lon: -73.751854},
				{lat: 40.576506, lon: -73.573634},
			},
			lat:  40.710594,
			lon:  -73.652183,
			want:     true,
		},
		{
			name:     "not contains",
			box:      &box{minLat: 40.5880775, maxLat: 40.8247008, minLon: -73.80014, maxLon: -73.4630336},
			lat:  40.690732,
			lon:  -74.046267,
			want:     false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		getGrid := geo.Functions["containsLatLon"]
		var owv values.Object
		if tc.box != nil {
			owv = values.NewObjectWithValues(map[string]values.Value{
				"box": values.NewObjectWithValues(map[string]values.Value{
					"minLat": values.NewFloat(tc.box.minLat),
					"minLon": values.NewFloat(tc.box.minLon),
					"maxLat": values.NewFloat(tc.box.maxLat),
					"maxLon": values.NewFloat(tc.box.maxLon),
				}),
				"lat":  values.NewFloat(tc.lat),
				"lon":  values.NewFloat(tc.lon),
			})
		} else if tc.circle != nil {
			owv = values.NewObjectWithValues(map[string]values.Value{
				"circle": values.NewObjectWithValues(map[string]values.Value{
					"lat":    values.NewFloat(tc.circle.lat),
					"lon":    values.NewFloat(tc.circle.lon),
					"radius": values.NewFloat(tc.circle.radius),
				}),
				"lat":  values.NewFloat(tc.lat),
				"lon":  values.NewFloat(tc.lon),
			})
		} else if tc.polygon != nil {
			array := values.NewArray(semantic.Object)
			for _, p := range *tc.polygon {
				array.Append(values.NewObjectWithValues(map[string]values.Value{
					"lat": values.NewFloat(p.lat),
					"lon": values.NewFloat(p.lon),
				}))
			}
			owv = values.NewObjectWithValues(map[string]values.Value{
				"polygon": values.NewObjectWithValues(map[string]values.Value{
					"points": array,
				}),
				"lat":  values.NewFloat(tc.lat),
				"lon":  values.NewFloat(tc.lon),
			})
		}
		result, err := getGrid.Call(context.Background(), owv)
		if err != nil {
			t.Error(err.Error())
		} else if tc.want != result.Bool() {
			t.Errorf("[%s] expected %v (%T), got %v (%T)", tc.name, tc.want, tc.want, result, result)
		}
	}
}
