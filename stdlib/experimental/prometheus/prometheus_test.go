package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	flux "github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/mock"
	"github.com/influxdata/flux/values"
)

func TestPrometheusScrape(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
		# TYPE go_memstats_gc_cpu_fraction gauge
		go_memstats_gc_cpu_fraction 0.02865414065048765
		# TYPE go_memstats_lookups_total counter
		go_memstats_lookups_total 0
		# TYPE go_gc_duration_seconds summary
		go_gc_duration_seconds{quantile="0"} 1.5819e-05
		go_gc_duration_seconds_sum 52.148989515
		go_gc_duration_seconds_count 1.505614e+06
		# TYPE prometheus_http_request_duration_seconds histogram
		prometheus_tsdb_compaction_chunk_range_seconds_bucket{le="+Inf"} 1.7792863e+07
		`)
	}))
	defer ts.Close()

	spec := &ScrapePrometheusProcedureSpec{URL: ts.URL}
	admin := &mock.Administration{}
	c := execute.NewTableBuilderCache(admin.Allocator())
	timestamp := time.Now()
	p := &PrometheusIterator{
		NowFn:          func() time.Time { return timestamp },
		spec:           spec,
		administration: admin,
		cache:          c,
	}

	ctx := dependenciestest.Default().Inject(context.Background())
	err := p.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = p.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	results := &executetest.Result{}
	runOnce := true

	more, err := p.Fetch(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for runOnce || more {
		runOnce = false
		tbl, err := p.Decode(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// convert flux Table to result Table
		resTbl, err := executetest.ConvertTable(tbl)
		if err != nil {
			t.Fatal(err)
		}

		// add to executetest Result
		results.Tbls = append(results.Tbls, resTbl)
		more, err = p.Fetch(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}

	wantResults := &executetest.Result{}

	wantGauge := &executetest.Table{
		KeyCols: []string{"_measurement", "_field"},
		ColMeta: []flux.ColMeta{
			{Label: "_time", Type: flux.TTime},
			{Label: "_value", Type: flux.TFloat},
			{Label: "_measurement", Type: flux.TString},
			{Label: "_field", Type: flux.TString},
			{Label: "url", Type: flux.TString},
		},
		Data: [][]interface{}{
			{
				values.ConvertTime(timestamp),
				0.02865414065048765,
				"prometheus",
				"go_memstats_gc_cpu_fraction",
				ts.URL,
			},
		},
	}

	wantCounter := &executetest.Table{
		KeyCols: []string{"_measurement", "_field"},
		ColMeta: []flux.ColMeta{
			{Label: "_time", Type: flux.TTime},
			{Label: "_value", Type: flux.TFloat},
			{Label: "_measurement", Type: flux.TString},
			{Label: "_field", Type: flux.TString},
			{Label: "url", Type: flux.TString},
		},
		Data: [][]interface{}{
			{
				values.ConvertTime(timestamp),
				0.0,
				"prometheus",
				"go_memstats_lookups_total",
				ts.URL,
			},
		},
	}

	wantSummary := &executetest.Table{
		KeyCols: []string{"_measurement", "_field", "quantile"},
		ColMeta: []flux.ColMeta{
			{Label: "_time", Type: flux.TTime},
			{Label: "_value", Type: flux.TFloat},
			{Label: "_measurement", Type: flux.TString},
			{Label: "_field", Type: flux.TString},
			{Label: "url", Type: flux.TString},
			{Label: "quantile", Type: flux.TString},
		},
		Data: [][]interface{}{
			{
				values.ConvertTime(timestamp),
				1.5819e-05,
				"prometheus",
				"go_gc_duration_seconds",
				ts.URL,
				"0",
			},
		},
	}

	wantSummarySum := &executetest.Table{
		KeyCols: []string{"_measurement", "_field"},
		ColMeta: []flux.ColMeta{
			{Label: "_time", Type: flux.TTime},
			{Label: "_value", Type: flux.TFloat},
			{Label: "_measurement", Type: flux.TString},
			{Label: "_field", Type: flux.TString},
			{Label: "url", Type: flux.TString},
		},
		Data: [][]interface{}{
			{
				values.ConvertTime(timestamp),
				52.148989515,
				"prometheus",
				"go_gc_duration_seconds_sum",
				ts.URL,
			},
		},
	}

	wantSummaryCount := &executetest.Table{
		KeyCols: []string{"_measurement", "_field"},
		ColMeta: []flux.ColMeta{
			{Label: "_time", Type: flux.TTime},
			{Label: "_value", Type: flux.TFloat},
			{Label: "_measurement", Type: flux.TString},
			{Label: "_field", Type: flux.TString},
			{Label: "url", Type: flux.TString},
		},
		Data: [][]interface{}{
			{
				values.ConvertTime(timestamp),
				1.505614e+06,
				"prometheus",
				"go_gc_duration_seconds_count",
				ts.URL,
			},
		},
	}

	wantResults.Tbls = append(wantResults.Tbls, wantGauge, wantCounter, wantSummaryCount, wantSummarySum, wantSummary)
	fmt.Println(wantResults.Tbls)

	err = executetest.EqualResult(wantResults, results)

	if err != nil {
		t.Fatal(err, wantResults.Tables(), results.Tables())
	}

}
