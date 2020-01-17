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

func TestGetGrid_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "no args",
			Raw:     `import "experimental/geo" geo.getGrid()`,
			WantErr: true, // missing required parameter(s)
		},
		{
			Name:    "invalid args - box",
			Raw:     `import "experimental/geo" geo.getGrid(box: { minLat: 40.5, minLon: -74.5 })`,
			WantErr: true, // invalid box specification - must have minLat, minLon, maxLat, maxLon fields
		},
		{
			Name:    "invalid args - minSize > maxSize",
			Raw:     `import "experimental/geo" geo.getGrid(box: { minLat: 40.5, minLon: -74.5, maxLat: 41.5, maxLon: -73.5 }, minSize: 11, maxSize: 9)`,
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
	testCases := []struct {
		name string
		box box
		precision int
		minsize int
		maxsize int
		want values.Object
	}{
		{
			name: "minSize #1",
			box: box{minLat: 40.49958463695424, maxLat: 40.91598930547667, minLon: -74.4267501831055, maxLon: -73.6027755737305},
			minsize: 9,
			maxsize: -1,
			precision: -1,
			want: gridValue(4, []string{"dr70", "dr72", "dr78", "dr5p", "dr5r", "dr5x", "dr5n", "dr5q", "dr5w"}),
		},
	{
			name: "maxSize #1",
			box: box{minLat: 40.49958463695424, maxLat: 40.91598930547667, minLon: -74.4267501831055, maxLon: -73.6027755737305},
			minsize: -1,
			maxsize: 20,
			precision: -1,
			want: gridValue(4, []string{"dr70", "dr72", "dr78", "dr5p", "dr5r", "dr5x", "dr5n", "dr5q", "dr5w"}),
		},
		{
			name: "maxSize #2",
			box: box{minLat: 39.49958463695424, maxLat: 41.91598930547667, minLon: -75.4267501831055, maxLon: -72.6027755737305},
			minsize: -1,
			maxsize: 9,
			precision: -1,
			want: gridValue(3, []string{"dr6", "dr7", "drk", "dr4", "dr5", "drh"}),
		},
		{
			name: "precision #1",
			box: box{minLat: 40.49958463695424, maxLat: 40.91598930547667, minLon: -74.4267501831055, maxLon: -73.6027755737305},
			minsize: -1,
			maxsize: -1,
			precision: 2,
			want: gridValue(2, []string{"dr"}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		getGrid := geo.Functions["getGrid"]
		result, err := getGrid.Call(context.Background(),
			values.NewObjectWithValues(map[string]values.Value{
				"box": values.NewObjectWithValues(map[string]values.Value{
					"minLat": values.NewFloat(tc.box.minLat),
					"minLon": values.NewFloat(tc.box.minLon),
					"maxLat": values.NewFloat(tc.box.maxLat),
					"maxLon": values.NewFloat(tc.box.maxLon),
				}),
				"minSize": values.NewInt(int64(tc.minsize)),
				"maxSize": values.NewInt(int64(tc.maxsize)),
				"precision": values.NewInt(int64(tc.precision)),
			}),
		)
		if err != nil {
			t.Error(err.Error())
		} else if /*!reflect.DeepEqual(tc.want, result) ||*/ !tc.want.Equal(result) {
			t.Errorf("expected %v (%T), got %v (%T)", tc.want, tc.want, result, result)
		}
	}
}

// Helpers

func gridValue(precision int, set []string) values.Object {
	array := values.NewArray(semantic.String)
	for _, s := range set {
		array.Append(values.NewString(s))
	}
	object := values.NewObject()
	object.Set("precision", values.NewInt(int64(precision)))
	object.Set("set", array)
	return object
}
