package prometheus_test

import (
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/prometheus"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/prometheus/prometheus/prompb"
)

func TestFromProm_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "prometheus.from no arg",
			Raw: `import "prometheus" 
			prometheus.from()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "prometheus.from0",
						Spec: &prometheus.FromOpSpec{
							Step:    10 * time.Second,
							Matcher: make([]*prompb.LabelMatcher, 0)},
					},
				},
			},
		},
		{
			Name: "from unexpected arg",
			Raw: `import "prometheus" 
			prometheus.from(chicken:"what is this?")`,
			WantErr: true,
		},
		{
			Name: "from unexpected step",
			Raw: `import "prometheus" 
			prometheus.from(name:"name", step:"60")`,
			WantErr: true,
		},
		{
			Name: "prometheus.from with URL and query",
			Raw: `import "prometheus" 
			prometheus.from(url:"http://localhost:9090",query:"go_gc_duration_seconds")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "prometheus.from0",
						Spec: &prometheus.FromOpSpec{
							URL:      "http://localhost:9090",
							Query:    "go_gc_duration_seconds",
							Step:     10 * time.Second,
							HasQuery: true,
							Matcher:  make([]*prompb.LabelMatcher, 0)},
					},
				},
			},
		},
		{
			Name: "prometheus.from with URL and a metrics name",
			Raw: `import "prometheus" 
			prometheus.from(url:"http://localhost:9090",name:"go_gc_duration_seconds")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "prometheus.from0",
						Spec: &prometheus.FromOpSpec{
							URL:  "http://localhost:9090",
							Step: 10 * time.Second,
							Matcher: append(make([]*prompb.LabelMatcher, 0),
								&prompb.LabelMatcher{
									Type: prompb.LabelMatcher_EQ, Name: "__name__",
									Value: "go_gc_duration_seconds"})},
					},
				},
			},
		},
		{
			Name: "prometheus.from with URL, a metrics name, a user and a password",
			Raw: `import "prometheus" 
			prometheus.from(url:"http://localhost:9090",name:"go_gc_duration_seconds", user:"read", password:"pwd")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "prometheus.from0",
						Spec: &prometheus.FromOpSpec{
							URL:      "http://localhost:9090",
							Step:     10 * time.Second,
							User:     "read",
							Password: "pwd",
							HasAuth:  true,
							Matcher: append(make([]*prompb.LabelMatcher, 0),
								&prompb.LabelMatcher{
									Type: prompb.LabelMatcher_EQ, Name: "__name__",
									Value: "go_gc_duration_seconds"})},
					},
				},
			},
		},
		{
			Name: "prometheus.from with URL, query, step, user and password",
			Raw: `import "prometheus" 
			prometheus.from(url:"http://localhost:9090",query:"go_gc_duration_seconds", step: 60s, user:"read", password:"pwd")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "prometheus.from0",
						Spec: &prometheus.FromOpSpec{
							URL:      "http://localhost:9090",
							Query:    "go_gc_duration_seconds",
							Step:     60 * time.Second,
							User:     "read",
							Password: "pwd",
							HasAuth:  true,
							HasQuery: true,
							Matcher:  make([]*prompb.LabelMatcher, 0)},
					},
				},
			},
		},
		{
			Name: "prometheus.from flux and PromQL query",
			Raw: `import "prometheus" 
			prometheus.from(url:"http://localhost:9090",query:"go_gc_duration_seconds")|> range(start:-4h, stop:-2h) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "prometheus.from0",
						Spec: &prometheus.FromOpSpec{
							URL:      "http://localhost:9090",
							Query:    "go_gc_duration_seconds",
							Step:     10 * time.Second,
							HasQuery: true,
							Matcher:  make([]*prompb.LabelMatcher, 0)},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "prometheus.from0", Child: "range1"},
					{Parent: "range1", Child: "sum2"},
				},
			},
		},
		{
			Name: "prometheus.from flux and remote_read query",
			Raw: `import "prometheus" 
			prometheus.from(url:"http://localhost:9090",name:"go_gc_duration_seconds", user:"read", password:"pwd")|> range(start:-4h, stop:-2h) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "prometheus.from0",
						Spec: &prometheus.FromOpSpec{
							URL:      "http://localhost:9090",
							Step:     10 * time.Second,
							User:     "read",
							Password: "pwd",
							HasAuth:  true,
							Matcher: append(make([]*prompb.LabelMatcher, 0),
								&prompb.LabelMatcher{
									Type: prompb.LabelMatcher_EQ, Name: "__name__",
									Value: "go_gc_duration_seconds"})},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "prometheus.from0", Child: "range1"},
					{Parent: "range1", Child: "sum2"},
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
