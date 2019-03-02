// NOTE: This e2e suite requires starting the DB setup as described in README.md!
package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os/user"
	"path"
	"testing"
	"text/template"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/flux/ast"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql"
)

var testOffsets = []map[string]interface{}{
	{"offset": "1m"},
	{"offset": "5m"},
	{"offset": "10m"},
}

var simpleAggrOps = []map[string]interface{}{
	{"op": "sum"},
	{"op": "avg"},
	{"op": "max"},
	{"op": "min"},
	{"op": "count"},
	// TODO: Needs support for population standard deviation mode in Flux.
	// {"op": "stddev"},
}

var topBottomOps = []map[string]interface{}{
	{"op": "topk"},
	{"op": "bottomk"},
}

var testQuantiles = []map[string]interface{}{
	// TODO: Should return -Inf.
	// {"quantile": "-0.5"},
	{"quantile": "0.1"},
	{"quantile": "0.5"},
	{"quantile": "0.75"},
	{"quantile": "0.95"},
	{"quantile": "0.90"},
	{"quantile": "0.99"},
	{"quantile": "1"},
	// TODO: Should return +Inf.
	// {"quantile": "1.5"},
}

var testArithBinops = []map[string]interface{}{
	{"op": "+"},
	{"op": "-"},
	{"op": "*"},
	{"op": "/"},
	// TODO: Not supported yet.
	// {"op": "%"},
	// {"op": "^"},
}

var testRanges = []map[string]interface{}{
	{"range": "1s"},
	{"range": "15s"},
	{"range": "1m"},
	{"range": "5m"},
	{"range": "15m"},
	{"range": "1h"},
}

var queries = []struct {
	query    string
	variants []map[string]interface{}
}{
	{
		query: `demo_cpu_usage_seconds_total`,
	},
	{
		query: `demo_cpu_usage_seconds_total{mode="idle"}`,
	},
	{
		query: `demo_cpu_usage_seconds_total{mode!="idle"}`,
	},
	{
		query: `demo_cpu_usage_seconds_total{instance=~"localhost:.*"}`,
	},
	{
		query: `demo_cpu_usage_seconds_total{instance=~"host"}`,
	},
	{
		query: `demo_cpu_usage_seconds_total{instance!~".*:10000"}`,
	},
	{
		query: `demo_cpu_usage_seconds_total{mode="idle", instance!="localhost:10000"}`,
	},
	{
		query: `{mode="idle", instance!="localhost:10000"}`,
	},
	{
		query: "nonexistent_metric_name",
	},
	{
		query:    `demo_cpu_usage_seconds_total offset {{.offset}}`,
		variants: testOffsets,
	},
	{
		query:    `{{.op}} (demo_cpu_usage_seconds_total)`,
		variants: simpleAggrOps,
	},
	{
		query:    `{{.op}} by(instance) (demo_cpu_usage_seconds_total)`,
		variants: simpleAggrOps,
	},
	{
		query:    `{{.op}} by(instance, mode) (demo_cpu_usage_seconds_total)`,
		variants: simpleAggrOps,
	},
	// TODO: grouping by non-existent columns is not supported in Flux.
	//
	// {
	// 	query:    `{{.op}} by(nonexistent) (demo_cpu_usage_seconds_total)`,
	// 	variants: simpleAggrOps,
	// },
	{
		query:    `{{.op}} without(instance) (demo_cpu_usage_seconds_total)`,
		variants: simpleAggrOps,
	},
	{
		query:    `{{.op}} without(instance, mode) (demo_cpu_usage_seconds_total)`,
		variants: simpleAggrOps,
	},
	{
		query:    `{{.op}} without(nonexistent) (demo_cpu_usage_seconds_total)`,
		variants: simpleAggrOps,
	},
	{
		query:    `{{.op}} (3, demo_cpu_usage_seconds_total)`,
		variants: topBottomOps,
	},
	{
		query:    `{{.op}} by(instance) (2, demo_cpu_usage_seconds_total)`,
		variants: topBottomOps,
	},
	{
		query:    `quantile({{.quantile}}, demo_cpu_usage_seconds_total)`,
		variants: testQuantiles,
	},
	{
		query: `avg(max by(mode) (demo_cpu_usage_seconds_total))`,
	},
	// {
	// 	query: `1 * 2 + 4 / 6 - 10`,
	// },
	{
		query:    `demo_cpu_usage_seconds_total {{.op}} 1.2345`,
		variants: testArithBinops,
	},
	{
		query:    `0.12345 {{.op}} demo_cpu_usage_seconds_total`,
		variants: testArithBinops,
	},
	// TODO: Flux drops parens when formatting out, loses associativity.
	// {
	// 	query:    `(1 * 2 + 4 / 6 - 10) {{.op}} demo_cpu_usage_seconds_total`,
	// 	variants: testArithBinops,
	// },
	// {
	// 	query:    `demo_cpu_usage_seconds_total {{.op}} (1 * 2 + 4 / 6 - 10)`,
	// 	variants: testArithBinops,
	// },
	{
		query:    `count_over_time(demo_cpu_usage_seconds_total[{{.range}}])`,
		variants: testRanges,
	},
}

func TestQueriesE2E(t *testing.T) {
	var runner e2eRunner

	flag.StringVar(&runner.influxURL, "influx-url", "http://localhost:9999/", "InfluxDB server URL.")
	flag.StringVar(&runner.influxBucket, "influx-bucket", "prometheus", "InfluxDB bucket name.")
	flag.StringVar(&runner.influxToken, "influx-token", "", "InfluxDB authentication token.")
	flag.StringVar(&runner.influxOrg, "influx-org", "prometheus", "The InfluxDB organization name.")
	flag.StringVar(&runner.promURL, "prometheus-url", "http://localhost:9090/", "Prometheus server URL.")
	queryStart := flag.Int64("query-start", 1550781000000, "Query start timestamp in milliseconds.")
	queryEnd := flag.Int64("query-end", 1550781900000, "Query end timestamp in milliseconds.")
	flag.DurationVar(&runner.resolution, "query-resolution", 10*time.Second, "Query resolution in seconds.")

	flag.Parse()

	if runner.influxToken == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatalln("Error getting current user:", err)
		}
		tokenPath := path.Join(usr.HomeDir, ".influxdbv2/credentials")
		token, err := ioutil.ReadFile(tokenPath)
		if err != nil {
			log.Fatalf("Error reading auth token from %q (-influx-token was not set): %s", tokenPath, err)
		}
		runner.influxToken = string(token)
	}

	runner.start = time.Unix(0, *queryStart*1e6).UTC()
	runner.end = time.Unix(0, *queryEnd*1e6).UTC()

	// TODO: Bring up test setup in Go...
	//
	// dir, err := ioutil.TempDir("", "test-promql-flux-transpilation")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer os.RemoveAll(dir)

	// cmd := exec.Command(
	// 	"influxdb",
	// 	"--bolt-path=influx-data/influxd.bolt",
	// 	"engine-path=influx-data/engine",
	// 	"protos-path=influx-data/protos",
	// 	"reporting-disabled",
	// )
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// if err := cmd.Start(); err != nil {
	// 	t.Fatal(err)
	// }
	// if err != cmd.Start()

	for _, q := range queries {
		if len(q.variants) > 0 {
			for _, variant := range q.variants {
				query := tprintf(q.query, variant)
				runner.runQuery(t, query)
			}
		} else {
			runner.runQuery(t, q.query)
		}
	}
}

// tprintf passed template string is formatted usign its operands and returns the resulting string.
// Spaces are added between operands when neither is a string.
func tprintf(tmpl string, data map[string]interface{}) string {
	t := template.Must(template.New("query").Parse(tmpl))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, data); err != nil {
		panic(err)
	}
	return buf.String()
}

type e2eRunner struct {
	influxURL    string
	influxBucket string
	influxToken  string
	influxOrg    string

	promURL string

	start      time.Time
	end        time.Time
	resolution time.Duration
}

func (r *e2eRunner) runQuery(t *testing.T, query string) {
	// Transpile PromQL into Flux.
	promqlNode, err := promql.ParseExpr(query)
	if err != nil {
		t.Fatalf("Error parsing PromQL expression %q: %s", query, err)
	}
	tr := &transpiler{
		bucket:     r.influxBucket,
		start:      r.start,
		end:        r.end,
		resolution: r.resolution,
	}
	fluxNode, err := tr.transpile(promqlNode)
	if err != nil {
		t.Fatalf("Error transpiling PromQL expression %q to Flux: %s", query, err)
	}

	// Query both Prometheus and InfluxDB, expect same result.
	promMatrix, err := queryPrometheus(r.promURL, query, r.start, r.end, r.resolution)
	if err != nil {
		t.Fatalf("Error querying Prometheus for %q: %s", query, err)
	}
	influxResult, err := queryInfluxDB(r.influxURL, r.influxOrg, r.influxToken, r.influxBucket, ast.Format(fluxNode))
	if err != nil {
		t.Fatalf("Error querying InfluxDB for %q: %s", query, err)
	}
	// Make InfluxDB result comparable with the Prometheus result.
	influxMatrix, err := influxResultToPromMatrix(influxResult)
	if err != nil {
		t.Fatalf("Error processing InfluxDB results for %q: %s", query, err)
	}

	cmpOpts := cmp.Options{
		// Translate sample values into float64 so that cmpopts.EquateApprox() works.
		cmp.Transformer("", func(in model.SampleValue) float64 {
			return float64(in)
		}),
		// Allow comparison tolerances due to floating point inaccuracy.
		cmpopts.EquateApprox(0.000000000000001, 0),
	}
	if diff := cmp.Diff(promMatrix, influxMatrix, cmpOpts); diff != "" {
		t.Error(
			"FAILED! Prometheus and InfluxDB results differ:\n\n", diff,
			"\nPromQL query was:\n============================================\n", query, "\n============================================\n\n",
			"\nFlux query was:\n============================================\n", ast.Format(fluxNode), "\n============================================\n\n",
			"\nFull results:",
			"\n=== InfluxDB results:\n", influxMatrix,
			"\n=== Prometheus results:\n", promMatrix,
		)
	}
}
