package geo

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

// TODO(ales.pour@bonitoo.io): This is exposed so the tests have access to the functions.
var Functions = map[string]values.Function{
	"getGrid":       generateGetGridFunc(),
	"getLevel":      generateGetLevelFunc(),
	"s2CellIDToken": generateS2CellIDTokenFunc(),
	"s2CellLatLon":  generateS2CellLatLonFunc(),
	"stContains":    generateSTContainsFunc(),
	"stDistance":    generateSTDistanceFunc(),
	"stLength":      generateSTLengthFunc(),
}

func TestGeometryArguments(t *testing.T) {
	type queryTestCase struct {
		Name    string
		Raw     string
		Want    string
		WantErr bool
		ErrMsg  string
	}

	tests := []queryTestCase{
		{
			Name: "null args",
			Raw: `import "csv"
import "experimental/geo"
import "influxdata/influxdb/schema"

data = "
#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,0,2021-03-17T19:48:00.37990328Z,2021-04-16T19:48:00.37990328Z,2021-04-09T19:00:06Z,38.4014,iss_position_longitude,iss,tahoecity.prod
,,1,2021-03-17T19:48:00.37990328Z,2021-04-16T19:48:00.37990328Z,2021-04-09T19:00:06Z,-24.1713,iss_position_latitude,iss,tahoecity.prod
"

csv.from(csv: data)
	|> schema.fieldsAsCols()
	|> filter(fn: (r) => geo.ST_Contains(region: {lat: 37.7858229, lon: -122.4058124, radius: 20.0}, geometry: {lat: r.lat, lon: r.lon}))
`,
			WantErr: true, // lat lon cannot be null
			ErrMsg:  "cannot be null",
		},
	}

	run := func(fluxScript string, writer io.Writer) (interface{}, error) {
		program, err := lang.Compile(fluxScript, runtime.Default, time.Now())
		if err != nil {
			t.Error(err)
		}
		ctx := flux.NewDefaultDependencies().Inject(context.Background())
		query, err := program.Start(ctx, &memory.Allocator{})
		if err != nil {
			t.Error(err)
		}
		results := flux.NewResultIteratorFromQuery(query)
		defer results.Release()
		encoder := &flux.DelimitedMultiResultEncoder{
			Delimiter: []byte("\n"),
			Encoder:   csv.NewResultEncoder(csv.DefaultEncoderConfig()),
		}
		return encoder.Encode(writer, results)
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			_, err := run(tc.Raw, &buf)
			got := buf.String()
			if err != nil {
				if !tc.WantErr {
					t.Error(err)
				} else if tc.ErrMsg != "" && !strings.Contains(err.Error(), tc.ErrMsg) {
					t.Errorf("unexpected error: %v", err)
				}
			} else if tc.WantErr {
				t.Fatalf("error was expected")
			} else if diff := cmp.Diff(tc.Want, got); diff != "" {
				t.Errorf("unexpected result -want/+got:\n%s\n", diff)
			}
		})
	}
}
