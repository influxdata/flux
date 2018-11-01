package inputs_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/functions/inputs"
	"github.com/influxdata/flux/querytest"
)

func TestFromJSON_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "from no args",
			Raw:     `fromJSON()`,
			WantErr: true,
		},
		{
			Name:    "from conflicting args",
			Raw:     `fromJSON(json:"d", file:"b")`,
			WantErr: true,
		},
		{
			Name:    "from repeat arg",
			Raw:     `from(json:"telegraf", json:"oops")`,
			WantErr: true,
		},
		{
			Name: "fromJSON text",
			Raw:  `fromJSON(json: "{results: []}")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "fromJSON0",
						Spec: &inputs.FromJSONOpSpec{
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
