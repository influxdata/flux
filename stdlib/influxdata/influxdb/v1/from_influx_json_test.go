package v1_test

import (
	"testing"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin" // We need to import the builtins for the tests to work.
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb/v1"
)

func TestFromInfluxJSON_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "no args",
			Raw:     `import "influxdata/influxdb/v1" vs.json()`,
			WantErr: true,
		},
		{
			Name:    "conflicting args",
			Raw:     `import "influxdata/influxdb/v1" v1.json(json:"d", file:"b")`,
			WantErr: true,
		},
		{
			Name:    "repeat arg",
			Raw:     `import "influxdata/influxdb/v1" v1.json(json:"telegraf", json:"oops")`,
			WantErr: true,
		},
		{
			Name: "text",
			Raw:  `import "influxdata/influxdb/v1" v1.json(json: "{results: []}")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "fromInfluxJSON0",
						Spec: &v1.FromInfluxJSONOpSpec{
							JSON: "{results: []}",
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}
