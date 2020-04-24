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

// TestGauge will make sure that gauge metrics produce accurate flux Tables.
func TestGauge(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
		# TYPE go_memstats_gc_cpu_fraction gauge
		go_memstats_gc_cpu_fraction 0.02865414065048765
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

	results := testSourceDecoder(p, t)

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

	wantResults.Tbls = append(wantResults.Tbls, wantGauge)

	err := executetest.EqualResult(wantResults, results)

	if err != nil {
		t.Fatal(err, wantResults.Tables(), results.Tables())
	}

}

// TestCounter will make sure that counter metrics produce accurate flux Tables.
func TestCounter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
		# TYPE go_memstats_lookups_total counter
		go_memstats_lookups_total 0
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

	results := testSourceDecoder(p, t)

	wantResults := &executetest.Result{}

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

	wantResults.Tbls = append(wantResults.Tbls, wantCounter)

	err := executetest.EqualResult(wantResults, results)

	if err != nil {
		t.Fatal(err, wantResults.Tables(), results.Tables())
	}

}

// TestSummary will make sure that summary metrics produce accurate flux Tables.
func TestSummary(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
		# TYPE go_gc_duration_seconds summary
		go_gc_duration_seconds{quantile="0"} 1.5819e-05
		go_gc_duration_seconds_sum 53.148989515
		go_gc_duration_seconds_count 1.405614e+06
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

	results := testSourceDecoder(p, t)

	wantResults := &executetest.Result{}

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
				53.148989515,
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
				1.405614e+06,
				"prometheus",
				"go_gc_duration_seconds_count",
				ts.URL,
			},
		},
	}
	wantResults.Tbls = append(wantResults.Tbls, wantSummaryCount, wantSummarySum, wantSummary)

	err := executetest.EqualResult(wantResults, results)

	if err != nil {
		t.Fatal(err, wantResults.Tables(), results.Tables())
	}

}

// testSourceDecoder will mimic the SourceDecoder interface for testing purposes.
func testSourceDecoder(p *PrometheusIterator, t *testing.T) *executetest.Result {
	results := &executetest.Result{}
	runOnce := true

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
	return results
}
