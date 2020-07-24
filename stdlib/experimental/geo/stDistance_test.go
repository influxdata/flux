package geo_test

import (
	"context"
	"math"
	"testing"

	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/experimental/geo"
	"github.com/influxdata/flux/values"
)

func TestSTDistance_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "no args",
			Raw:     `import "experimental/geo" geo.ST_Distance()`,
			WantErr: true, // missing required parameter(s)
		},
		{
			Name:    "missing geometry arg",
			Raw:     `import "experimental/geo" geo.ST_Distance(region: { lat: 40.5, lon: -74.5, radius: 15.0 })`,
			WantErr: true, // missing required parameter(s)
		},
		{
			Name:    "invalid args - invalid box",
			Raw:     `import "experimental/geo" geo.ST_Distance(region: { minLat: 40.5, minLon: -74.5, maxLat: 41.5 }, geometry: {lat: 40.5, lon: -74.5})`,
			WantErr: true, // missing maxLon
		},
		{
			Name:    "invalid args - invalid circle",
			Raw:     `import "experimental/geo" geo.ST_Distance(region: { lat: 40.5, radius: 15.0 }, geometry: {lat: 40.5, lon: -74.5})`,
			WantErr: true, // missing lon
		},
		{
			Name:    "invalid args - invalid polygon",
			Raw:     `import "experimental/geo" geo.ST_Distance(region: { points: [{ lat: 40.5, lon: -74.5 }] }, geometry: {lat: 40.5, lon: -74.5})`,
			WantErr: true, // polygon must have at least 3 points
		},
		{
			Name:    "invalid args - unsupported region",
			Raw:     `import "experimental/geo" geo.ST_Distance(region: { x: 1.0, y: 0.0 }, geometry: {lat: 40.5, lon: -74.5})`,
			WantErr: true, // cannot infer region type
		},
		{
			Name:    "invalid args - invalid units",
			Raw:     `import "experimental/geo" geo.ST_Distance(region: { lat: 40.5, lon: -74.5, radius: 15.0 }, geometry: {lat: 40.5, lon: -74.5}, units: { distance: "yd" })`,
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

func TestSTDistance_Process(t *testing.T) {
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
		point   *point
		box     *box
		circle  *circle
		polygon *[]point
		units   *map[string]string
		toPoint *point
		toLine  string
		want    float64
	}{
		{
			name:    "point-point distance",
			point:   &point{40.854984, -73.403617}, // somewhere in Long island
			toPoint: &point{40.6892, -74.0445},     // Statue of Liberty
			want:    57.03,
		},
		{
			name:    "point-point distance - mile units",
			point:   &point{40.854984, -73.403617}, // somewhere in Long island
			toPoint: &point{40.6892, -74.0445},     // Statue of Liberty
			units:   &map[string]string{"distance": "mile"},
			want:    35.44,
		},
		{
			name:    "point-point distance - m units",
			point:   &point{40.854984, -73.403617}, // somewhere in Long island
			toPoint: &point{40.6892, -74.0445},     // Statue of Liberty
			units:   &map[string]string{"distance": "m"},
			want:    57029.82,
		},
		{
			name:   "point-linestring distance",
			point:  &point{40.6892, -74.0445},                    // Statue of Liberty
			toLine: "-73.912119 40.751085, -73.940563 40.745249", // somewhere in Brooklyn
			want:   10.75,
		},
		{
			name:    "box-point distance",
			box:     &box{minLat: 40.5880775, maxLat: 40.8247008, minLon: -73.80014, maxLon: -73.4630336},
			toPoint: &point{40.6892, -74.0445}, // Statue of Liberty
			want:    20.60,
		},
		{
			name:   "box-linestring distance",
			box:    &box{minLat: 40.5880775, maxLat: 40.8247008, minLon: -73.80014, maxLon: -73.4630336},
			toLine: "-73.912119 40.751085, -73.940563 40.745249", // somewhere in Brooklyn
			want:   9.43,
		},
		{
			name:    "circle-point distance",
			circle:  &circle{lat: 40.7090214, lon: -73.61846, radius: 15.0}, // somewhere in Long Island
			toPoint: &point{40.6892, -74.0445},                              // Statue of Liberty
			want:    20.98,
		},
		{
			name:   "circle-linestring distance",
			circle: &circle{lat: 40.7090214, lon: -73.61846, radius: 15.0}, // somewhere in Long Island
			toLine: "-73.912119 40.751085, -73.940563 40.745249",           // somewhere in Brooklyn
			want:   10.18,
		},
		{
			name: "polygon-point distance",
			polygon: &[]point{ // Brooklyn
				{lat: 40.671659, lon: -73.936631},
				{lat: 40.706543, lon: -73.749177},
				{lat: 40.791333, lon: -73.880327},
			},
			toPoint: &point{40.6892, -74.0445}, // Statue of Liberty
			want:    9.30,
		},
		{
			name: "polygon-linestring distance",
			polygon: &[]point{ // Brooklyn
				{lat: 40.671659, lon: -73.936631},
				{lat: 40.706543, lon: -73.749177},
				{lat: 40.791333, lon: -73.880327},
			},
			toLine: "-73.68691 40.820317, -73.742812 40.773334", // somewhere in Long Island but not far
			want:   5.99,
		},
		{
			name:    "zero distance",
			circle:  &circle{lat: 40.7090214, lon: -73.61846, radius: 15.0}, // somewhere in Long Island
			toPoint: &point{40.718170, -73.635265},                          // about 1 km from the circle center
			want:    0.0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		stDistance := geo.Functions["stDistance"]
		var owv values.Object
		if tc.units == nil {
			tc.units = &defaultUnits
		}
		if tc.point != nil {
			if tc.toPoint != nil {
				owv = values.NewObjectWithValues(map[string]values.Value{
					"region": values.NewObjectWithValues(map[string]values.Value{
						"lat": values.NewFloat(tc.point.lat),
						"lon": values.NewFloat(tc.point.lon),
					}),
					"geometry": values.NewObjectWithValues(map[string]values.Value{
						"lat": values.NewFloat(tc.toPoint.lat),
						"lon": values.NewFloat(tc.toPoint.lon),
					}),
					"units": unitsToValue(*tc.units),
				})
			} else if tc.toLine != "" {
				owv = values.NewObjectWithValues(map[string]values.Value{
					"region": values.NewObjectWithValues(map[string]values.Value{
						"lat": values.NewFloat(tc.point.lat),
						"lon": values.NewFloat(tc.point.lon),
					}),
					"geometry": values.NewObjectWithValues(map[string]values.Value{
						"linestring": values.NewString(tc.toLine),
					}),
					"units": unitsToValue(*tc.units),
				})
			}
		} else if tc.box != nil {
			if tc.toPoint != nil {
				owv = values.NewObjectWithValues(map[string]values.Value{
					"region": values.NewObjectWithValues(map[string]values.Value{
						"minLat": values.NewFloat(tc.box.minLat),
						"minLon": values.NewFloat(tc.box.minLon),
						"maxLat": values.NewFloat(tc.box.maxLat),
						"maxLon": values.NewFloat(tc.box.maxLon),
					}),
					"geometry": values.NewObjectWithValues(map[string]values.Value{
						"lat": values.NewFloat(tc.toPoint.lat),
						"lon": values.NewFloat(tc.toPoint.lon),
					}),
					"units": unitsToValue(*tc.units),
				})
			} else if tc.toLine != "" {
				owv = values.NewObjectWithValues(map[string]values.Value{
					"region": values.NewObjectWithValues(map[string]values.Value{
						"minLat": values.NewFloat(tc.box.minLat),
						"minLon": values.NewFloat(tc.box.minLon),
						"maxLat": values.NewFloat(tc.box.maxLat),
						"maxLon": values.NewFloat(tc.box.maxLon),
					}),
					"geometry": values.NewObjectWithValues(map[string]values.Value{
						"linestring": values.NewString(tc.toLine),
					}),
					"units": unitsToValue(*tc.units),
				})
			}
		} else if tc.circle != nil {
			if tc.toPoint != nil {
				owv = values.NewObjectWithValues(map[string]values.Value{
					"region": values.NewObjectWithValues(map[string]values.Value{
						"lat":    values.NewFloat(tc.circle.lat),
						"lon":    values.NewFloat(tc.circle.lon),
						"radius": values.NewFloat(tc.circle.radius),
					}),
					"geometry": values.NewObjectWithValues(map[string]values.Value{
						"lat": values.NewFloat(tc.toPoint.lat),
						"lon": values.NewFloat(tc.toPoint.lon),
					}),
					"units": unitsToValue(*tc.units),
				})
			} else if tc.toLine != "" {
				owv = values.NewObjectWithValues(map[string]values.Value{
					"region": values.NewObjectWithValues(map[string]values.Value{
						"lat":    values.NewFloat(tc.circle.lat),
						"lon":    values.NewFloat(tc.circle.lon),
						"radius": values.NewFloat(tc.circle.radius),
					}),
					"geometry": values.NewObjectWithValues(map[string]values.Value{
						"linestring": values.NewString(tc.toLine),
					}),
					"units": unitsToValue(*tc.units),
				})
			}
		} else if tc.polygon != nil {
			array := values.NewArray(semantic.NewArrayType(pointT))
			for _, p := range *tc.polygon {
				array.Append(values.NewObjectWithValues(map[string]values.Value{
					"lat": values.NewFloat(p.lat),
					"lon": values.NewFloat(p.lon),
				}))
			}
			if tc.toPoint != nil {
				owv = values.NewObjectWithValues(map[string]values.Value{
					"region": values.NewObjectWithValues(map[string]values.Value{
						"points": array,
					}),
					"geometry": values.NewObjectWithValues(map[string]values.Value{
						"lat": values.NewFloat(tc.toPoint.lat),
						"lon": values.NewFloat(tc.toPoint.lon),
					}),
					"units": unitsToValue(*tc.units),
				})
			} else if tc.toLine != "" {
				owv = values.NewObjectWithValues(map[string]values.Value{
					"region": values.NewObjectWithValues(map[string]values.Value{
						"points": array,
					}),
					"geometry": values.NewObjectWithValues(map[string]values.Value{
						"linestring": values.NewString(tc.toLine),
					}),
					"units": unitsToValue(*tc.units),
				})
			}
		}
		result, err := stDistance.Call(context.Background(), owv)
		if err != nil {
			t.Error(err.Error())
		} else if tc.want != roundDistance(result) {
			t.Errorf("[%s] expected %v (%T), got %v (%T)", tc.name, tc.want, tc.want, roundDistance(result), result)
		}
	}
}

//
// Helpers
//

func roundDistance(value values.Value) float64 {
	return math.Round(value.Float()*100) / 100
}
