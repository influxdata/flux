package influxdb_test

import (
	"net/url"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb/internal/testutil"
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
		{
			Name:    "buckets unexpected arg",
			Raw:     `buckets(chicken:"what is this?")`,
			WantErr: true,
		},
		{
			Name: "buckets with host and token",
			Raw:  `buckets(host: "http://localhost:9999", token: "mytoken")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "buckets0",
						Spec: &influxdb.BucketsOpSpec{
							Host:  stringPtr("http://localhost:9999"),
							Token: stringPtr("mytoken"),
						},
					},
				},
			},
		},
		{
			Name: "buckets with org",
			Raw:  `buckets(org: "influxdata")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "buckets0",
						Spec: &influxdb.BucketsOpSpec{
							Org: &influxdb.NameOrID{Name: "influxdata"},
						},
					},
				},
			},
		},
		{
			Name: "buckets with org id",
			Raw:  `buckets(orgID: "97aa81cc0e247dc4")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "buckets0",
						Spec: &influxdb.BucketsOpSpec{
							Org: &influxdb.NameOrID{ID: "97aa81cc0e247dc4"},
						},
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

func TestBuckets_Run(t *testing.T) {
	defaultTablesFn := func() []*executetest.Table {
		return []*executetest.Table{{
			KeyCols: []string{"organizationID"},
			ColMeta: []flux.ColMeta{
				{Label: "organizationID", Type: flux.TString},
				{Label: "name", Type: flux.TString},
				{Label: "id", Type: flux.TString},
				{Label: "retentionPolicy", Type: flux.TString},
				{Label: "retentionPeriod", Type: flux.TInt},
			},
			Data: [][]interface{}{
				{"97aa81cc0e247dc4", "telegraf", "1e01ac57da723035", nil, int64(0)},
			},
		}}
	}

	for _, tt := range []struct {
		name string
		spec *influxdb.BucketsRemoteProcedureSpec
		want testutil.Want
	}{
		{
			name: "basic query",
			spec: &influxdb.BucketsRemoteProcedureSpec{
				BucketsProcedureSpec: &influxdb.BucketsProcedureSpec{
					Org:   &influxdb.NameOrID{Name: "influxdata"},
					Token: stringPtr("mytoken"),
				},
			},
			want: testutil.Want{
				Params: url.Values{
					"org": []string{"influxdata"},
				},
				Query: `package main


buckets()`,
				Tables: defaultTablesFn,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			testutil.RunSourceTestHelper(t, tt.spec, tt.want)
		})
	}
}

func TestBuckets_Run_Errors(t *testing.T) {
	testutil.RunSourceErrorTestHelper(t, &influxdb.BucketsRemoteProcedureSpec{
		BucketsProcedureSpec: &influxdb.BucketsProcedureSpec{
			Org:   &influxdb.NameOrID{Name: "influxdata"},
			Token: stringPtr("mytoken"),
		},
	})
}

func TestBuckets_URLValidator(t *testing.T) {
	testutil.RunSourceURLValidatorTestHelper(t, &influxdb.BucketsRemoteProcedureSpec{
		BucketsProcedureSpec: &influxdb.BucketsProcedureSpec{
			Org:   &influxdb.NameOrID{Name: "influxdata"},
			Token: stringPtr("mytoken"),
		},
	})
}

func TestBuckets_HTTPClient(t *testing.T) {
	testutil.RunSourceHTTPClientTestHelper(t, &influxdb.BucketsRemoteProcedureSpec{
		BucketsProcedureSpec: &influxdb.BucketsProcedureSpec{
			Org:   &influxdb.NameOrID{Name: "influxdata"},
			Token: stringPtr("mytoken"),
		},
	})
}
