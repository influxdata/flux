package influxdb_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
)

func TestBuckets_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "buckets no args",
			Raw:  `buckets()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID:   "buckets0",
						Spec: &influxdb.BucketsOpSpec{},
					},
				},
			},
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
