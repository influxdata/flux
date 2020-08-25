package geo_test

import (
	"context"
	"math"
	"strings"
	"testing"

	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/experimental/geo"
	"github.com/influxdata/flux/values"
)

func TestS2CellLatLon_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "no args",
			Raw:     `import "experimental/geo" geo.s2CellLatLon()`,
			WantErr: true, // missing required parameter(s)
		},
		{
			Name:    "invalid arg",
			Raw:     `import "experimental/geo" geo.s2CellLatLon(id: "hi")`,
			WantErr: true, // missing required parameter(s)
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

func TestS2CellILatLon_Process(t *testing.T) {
	testCases := []struct {
		name         string
		token        string
		want         values.Value
		wantErr      bool
		errSubstring string
	}{
		{
			name:  "level 9 token",
			token: "89c284", // level 9
			want:  roundPoint(pointToValue(40.812535546624574, -73.55941282728273)),
		},
		{
			name:  "level 7 token",
			token: "89c2c",
			want:  roundPoint(pointToValue(41.089524879437086, -73.8419063407762)),
		},
		{
			name:         "invalid token",
			token:        "%*^&(*^*%&$",
			wantErr:      true,
			errSubstring: "invalid token specified",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s2CellLatLon := geo.Functions["s2CellLatLon"]
		var owv values.Object
		if tc.token != "" {
			owv = values.NewObjectWithValues(map[string]values.Value{
				"token": values.NewString(tc.token),
			})
		}
		result, err := s2CellLatLon.Call(context.Background(), owv)
		if err != nil {
			if !tc.wantErr {
				t.Error(err.Error())
			}
			if tc.errSubstring != "" && !strings.Contains(err.Error(), tc.errSubstring) {
				t.Errorf("[%s] expected error with '%s', got '%v'", tc.name, tc.errSubstring, err)
			}
		} else if !tc.want.Equal(roundPoint(result)) { // !reflect.DeepEqual(tc.want, result)
			t.Errorf("[%s] expected %v (%T), got %v (%T)", tc.name, tc.want, tc.want, roundPoint(result), result)
		}
	}
}

//
// Helpers
//

func pointToValue(lat, lon float64) values.Value {
	return values.NewObjectWithValues(map[string]values.Value{
		"lat": values.NewFloat(lat),
		"lon": values.NewFloat(lon),
	})
}

func roundPoint(value values.Value) values.Value {
	lat, latOk := value.Object().Get("lat")
	if latOk {
		value.Object().Set("lat", values.NewFloat(math.Round(lat.Float()*1000000)/1000000))
	}
	lon, lonOk := value.Object().Get("lon")
	if lonOk {
		value.Object().Set("lon", values.NewFloat(math.Round(lon.Float()*1000000)/1000000))
	}
	return value
}
