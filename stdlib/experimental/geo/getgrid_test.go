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

var pointT = semantic.NewObjectType([]semantic.PropertyType{
	{Key: []byte("lat"), Value: semantic.BasicFloat},
	{Key: []byte("lon"), Value: semantic.BasicFloat},
})

func TestGetGrid_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "no args",
			Raw:     `import "experimental/geo" geo.getGrid()`,
			WantErr: true, // missing required keyword argument
		},
		{
			Name:    "invalid args - invalid box",
			Raw:     `import "experimental/geo" geo.getGrid(region: { minLat: 40.5, minLon: -74.5 }, units: {distance: "km"})`,
			WantErr: true, // box must have minLat, minLon, maxLat, maxLon fields
		},
		{
			Name:    "invalid args - invalid circle",
			Raw:     `import "experimental/geo" geo.getGrid(region: { radius: 16.0 }, units: {distance: "km"})`,
			WantErr: true, // circle must have lat, lon, radius fields
		},
		{
			Name:    "invalid args - invalid polygon",
			Raw:     `import "experimental/geo" geo.getGrid(region: { points: [{ lat: 40.5, lon: -74.5 }] }, units: {distance: "km"})`,
			WantErr: true, // circle must have at least 3 points
		},
		{
			Name:    "invalid args - unknown region",
			Raw:     `import "experimental/geo" geo.getGrid(region: { lat: 40.5, lon: -74.5 }, units: {distance: "km"})`,
			WantErr: true, // cannot infer region type
		},
		{
			Name:    "invalid args - multitype region",
			Raw:     `import "experimental/geo" geo.getGrid(region: { minLat: 40.5, minLon: -74.5, maxLat: 41.5, maxLon: -73.5, lat: 41.0, lon: -74.0, radius: 15.0 }, units: {distance: "km"})`,
			WantErr: true, // infers multiple region types
		},
		{
			Name:    "invalid args - minSize > maxSize",
			Raw:     `import "experimental/geo" geo.getGrid(region: { minLat: 40.5, minLon: -74.5, maxLat: 41.5, maxLon: -73.5 }, minSize: 11, maxSize: 9, units: {distance: "km"})`,
			WantErr: true, // minSize > maxSize (11 > 9)
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

func TestGetGrid_Process(t *testing.T) {
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
		name     string
		box      *box
		circle   *circle
		polygon  *[]point
		minsize  int
		maxsize  int
		level    int
		maxLevel int
		units    *map[string]string
		want     values.Value
	}{
		{
			name:     "explicit level / box",
			box:      &box{minLat: 40.5880775, maxLat: 40.8247008, minLon: -73.80014, maxLon: -73.4630336},
			minsize:  24, // ignored when level > -1
			maxsize:  -1,
			level:    9,
			maxLevel: -1,
			want:     gridToValue(9, []string{"89c264", "89c26c", "89c274", "89c27c", "89c284", "89c28c", "89e82c", "89e9d4"}),
		},
		{
			name:     "explicit level / circle",
			circle:   &circle{lat: 40.7090214, lon: -73.61846, radius: 15.0},
			minsize:  24, // ignored when level > -1
			maxsize:  -1,
			level:    9,
			maxLevel: -1,
			want:     gridToValue(9, []string{"89c264", "89c26c", "89c274", "89c27c", "89c284", "89c28c", "89e82c", "89e9d4"}),
		},
		{
			name: "explicit level / polygon",
			polygon: &[]point{
				{lat: 40.776527, lon: -73.338811},
				{lat: 40.788093, lon: -73.776396},
				{lat: 40.475939, lon: -73.751854},
				{lat: 40.576506, lon: -73.573634},
			},
			minsize:  24, // ignored when level > -1
			maxsize:  -1,
			level:    9,
			maxLevel: -1,
			want:     gridToValue(9, []string{"89c264", "89c26c", "89c274", "89c27c", "89c284", "89c28c", "89e82c", "89e9d4"}),
		},
		{
			name:     "minSize",
			box:      &box{minLat: 40.5880775, maxLat: 40.8247008, minLon: -73.80014, maxLon: -73.4630336},
			minsize:  7,
			maxsize:  -1,
			level:    -1,
			maxLevel: 11,
			want:     gridToValue(9, []string{"89c264", "89c26c", "89c274", "89c27c", "89c284", "89c28c", "89e82c", "89e9d4"}),
		},
		{
			name:     "maxSize",
			box:      &box{minLat: 40.5880775, maxLat: 40.8247008, minLon: -73.80014, maxLon: -73.4630336},
			minsize:  7,
			maxsize:  10,
			level:    -1,
			maxLevel: 11,
			want:     gridToValue(9, []string{"89c264", "89c26c", "89c274", "89c27c", "89c284", "89c28c", "89e82c", "89e9d4"}),
		},
		{
			name:     "cannot satisfy minSize",
			box:      &box{minLat: 40.5880775, maxLat: 40.8247008, minLon: -73.80014, maxLon: -73.4630336},
			minsize:  1000,
			maxsize:  -1,
			level:    -1,
			maxLevel: -1,
			want:     gridToValue(-1, []string{}),
		},
		{
			name:     "cannot satisfy minSize but has fallback",
			box:      &box{minLat: 40.5880775, maxLat: 40.8247008, minLon: -73.80014, maxLon: -73.4630336},
			minsize:  1000,
			maxsize:  -1,
			level:    -1,
			maxLevel: 9, // used as fallback
			want:     gridToValue(9, []string{"89c264", "89c26c", "89c274", "89c27c", "89c284", "89c28c", "89e82c", "89e9d4"}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		getGrid := geo.Functions["getGrid"]
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
				"minSize":  values.NewInt(int64(tc.minsize)),
				"maxSize":  values.NewInt(int64(tc.maxsize)),
				"level":    values.NewInt(int64(tc.level)),
				"maxLevel": values.NewInt(int64(tc.maxLevel)),
				"units":    unitsToValue(*tc.units),
			})
		} else if tc.circle != nil {
			owv = values.NewObjectWithValues(map[string]values.Value{
				"region": values.NewObjectWithValues(map[string]values.Value{
					"lat":    values.NewFloat(tc.circle.lat),
					"lon":    values.NewFloat(tc.circle.lon),
					"radius": values.NewFloat(tc.circle.radius),
				}),
				"minSize":  values.NewInt(int64(tc.minsize)),
				"maxSize":  values.NewInt(int64(tc.maxsize)),
				"level":    values.NewInt(int64(tc.level)),
				"maxLevel": values.NewInt(int64(tc.maxLevel)),
				"units":    unitsToValue(*tc.units),
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
				"minSize":  values.NewInt(int64(tc.minsize)),
				"maxSize":  values.NewInt(int64(tc.maxsize)),
				"level":    values.NewInt(int64(tc.level)),
				"maxLevel": values.NewInt(int64(tc.maxLevel)),
				"units":    unitsToValue(*tc.units),
			})
		}
		result, err := getGrid.Call(context.Background(), owv)
		if err != nil {
			t.Error(err.Error())
		} else if !tc.want.Equal(result) { // !reflect.DeepEqual(tc.want, result)
			t.Errorf("[%s] expected %v (%T), got %v (%T)", tc.name, tc.want, tc.want, result, result)
		}
	}
}

//
// Helpers
//

func gridToValue(level int, set []string) values.Value {
	array := values.NewArray(semantic.NewArrayType(semantic.BasicString))
	for _, s := range set {
		array.Append(values.NewString(s))
	}
	return values.NewObjectWithValues(map[string]values.Value{
		"level": values.NewInt(int64(level)),
		"set":   array,
	})
}

func unitsToValue(units map[string]string) values.Value {
	return values.NewObjectWithValues(map[string]values.Value{
		"distance": values.NewString(units["distance"]),
	})
}
