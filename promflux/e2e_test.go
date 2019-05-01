// NOTE: This e2e suite requires starting the DB setup as described in README.md!
package main

import (
	"bytes"
	"flag"
	"fmt"
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

var testVariantArgs = map[string][]string{
	"range":  []string{"1s", "15s", "1m", "5m", "15m", "1h"},
	"offset": []string{"1m", "5m", "10m"},
	// stddev/stdvar Needs support for population standard deviation mode in Flux.
	// https://github.com/influxdata/flux/issues/1010
	"simpleAggrOp": []string{"sum", "avg", "max", "min", "count", "stddev", "stdvar"},
	"topBottomOp":  []string{"topk", "bottomk"},
	"quantile": []string{
		// TODO: Should return -Inf.
		// "-0.5",
		"0.1",
		"0.5",
		"0.75",
		"0.95",
		"0.90",
		"0.99",
		"1",
		// TODO: Should return +Inf.
		// "1.5",
	},
	// TODO: "%" and "^" not supported yet by Flux.
	"arithBinOp":           []string{"+", "-", "*", "/", "%", "^"},
	"compBinOp":            []string{"==", "!=", "<", ">", "<=", ">="},
	"binOp":                []string{"+", "-", "*", "/", "%", "^", "==", "!=", "<", ">", "<=", ">="},
	"simpleMathFunc":       []string{"abs", "ceil", "floor", "exp", "sqrt", "ln", "log2", "log10", "round"},
	"extrapolatedRateFunc": []string{"delta", "rate", "increase"},
}

var queries = []struct {
	query       string
	variantArgs []string
}{
	{
		query: `demo_cpu_usage_seconds_total + -(1 + 1)`,
	},
	{
		query: `demo_cpu_usage_seconds_total`,
	},
	{
		query: `{__name__="demo_cpu_usage_seconds_total"}`,
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
		query:       `demo_cpu_usage_seconds_total offset {{.offset}}`,
		variantArgs: []string{"offset"},
	},
	{
		query:       `{{.simpleAggrOp}} (demo_cpu_usage_seconds_total)`,
		variantArgs: []string{"simpleAggrOp"},
	},
	{
		query:       `{{.simpleAggrOp}} by() (demo_cpu_usage_seconds_total)`,
		variantArgs: []string{"simpleAggrOp"},
	},
	{
		query:       `{{.simpleAggrOp}} by(instance) (demo_cpu_usage_seconds_total)`,
		variantArgs: []string{"simpleAggrOp"},
	},
	{
		query:       `{{.simpleAggrOp}} by(instance, mode) (demo_cpu_usage_seconds_total)`,
		variantArgs: []string{"simpleAggrOp"},
	},
	// TODO: grouping by non-existent columns is not supported in Flux.
	// https://github.com/influxdata/flux/issues/1117
	// {
	// 	query:       `{{.simpleAggrOp}} by(nonexistent) (demo_cpu_usage_seconds_total)`,
	// 	variantArgs: []string{"simpleAggrOp"},
	// },
	{
		query:       `{{.simpleAggrOp}} without() (demo_cpu_usage_seconds_total)`,
		variantArgs: []string{"simpleAggrOp"},
	},
	// TODO: Need to handle & test external injections of special column names more systematically.
	{
		query:       `{{.simpleAggrOp}} without(_value) (demo_cpu_usage_seconds_total)`,
		variantArgs: []string{"simpleAggrOp"},
	},
	{
		query:       `{{.simpleAggrOp}} without(instance) (demo_cpu_usage_seconds_total)`,
		variantArgs: []string{"simpleAggrOp"},
	},
	{
		query:       `{{.simpleAggrOp}} without(instance, mode) (demo_cpu_usage_seconds_total)`,
		variantArgs: []string{"simpleAggrOp"},
	},
	{
		query:       `{{.simpleAggrOp}} without(nonexistent) (demo_cpu_usage_seconds_total)`,
		variantArgs: []string{"simpleAggrOp"},
	},
	{
		query:       `{{.topBottomOp}} (3, demo_cpu_usage_seconds_total)`,
		variantArgs: []string{"topBottomOp"},
	},
	{
		query:       `{{.topBottomOp}} by(instance) (2, demo_cpu_usage_seconds_total)`,
		variantArgs: []string{"topBottomOp"},
	},
	{
		query:       `quantile({{.quantile}}, demo_cpu_usage_seconds_total)`,
		variantArgs: []string{"quantile"},
	},
	{
		query: `avg(max by(mode) (demo_cpu_usage_seconds_total))`,
	},
	{
		query: `count(go_memstats_heap_released_bytes_total)`,
	},
	// {
	// TODO: Missing root node.
	// 	query: `1 * 2 + 4 / 6 - 10`,
	//  variantArgs: []string{""},
	// },
	{
		query:       `demo_num_cpus + (1 {{.compBinOp}} bool 2)`,
		variantArgs: []string{"compBinOp"},
	},
	{
		query: `-demo_cpu_usage_seconds_total`,
	},
	{
		query:       `demo_cpu_usage_seconds_total {{.binOp}} 1.2345`,
		variantArgs: []string{"binOp"},
	},
	{
		query:       `demo_cpu_usage_seconds_total {{.compBinOp}} bool 1.2345`,
		variantArgs: []string{"compBinOp"},
	},
	{
		query:       `1.2345 {{.compBinOp}} bool demo_cpu_usage_seconds_total`,
		variantArgs: []string{"compBinOp"},
	},
	{
		query:       `0.12345 {{.binOp}} demo_cpu_usage_seconds_total`,
		variantArgs: []string{"binOp"},
	},
	{
		query:       `(1 * 2 + 4 / 6 - (10%7)^2) {{.binOp}} demo_cpu_usage_seconds_total`,
		variantArgs: []string{"binOp"},
	},
	{
		query:       `demo_cpu_usage_seconds_total {{.binOp}} (1 * 2 + 4 / 6 - 10)`,
		variantArgs: []string{"binOp"},
	},

	// TODO: https://github.com/influxdata/flux/issues/1040
	// {
	// 	query: `demo_num_cpus * Inf`,
	// },
	// {
	// 	query: `demo_num_cpus * -Inf`,
	// },
	// {
	// 	query: `demo_num_cpus * NaN`,
	// },

	{
		query:       `demo_cpu_usage_seconds_total {{.binOp}} on(instance, job, mode) demo_cpu_usage_seconds_total`,
		variantArgs: []string{"binOp"},
	},
	{
		query:       `sum by(instance, mode) (demo_cpu_usage_seconds_total) {{.binOp}} on(instance, mode) group_left(job) demo_cpu_usage_seconds_total`,
		variantArgs: []string{"binOp"},
	},
	{
		query:       `demo_cpu_usage_seconds_total {{.compBinOp}} bool on(instance, job, mode) demo_cpu_usage_seconds_total`,
		variantArgs: []string{"compBinOp"},
	},
	{
		// Check that __name__ is always dropped, even if it's part of the matching labels.
		query: `demo_cpu_usage_seconds_total / on(instance, job, mode, __name__) demo_cpu_usage_seconds_total`,
	},
	{
		query: `sum without(job) (demo_cpu_usage_seconds_total) / on(instance, mode) demo_cpu_usage_seconds_total`,
	},
	{
		query: `sum without(job) (demo_cpu_usage_seconds_total) / on(instance, mode) group_left demo_cpu_usage_seconds_total`,
	},
	{
		query: `sum without(job) (demo_cpu_usage_seconds_total) / on(instance, mode) group_left(job) demo_cpu_usage_seconds_total`,
	},
	// {
	// 	query: `demo_cpu_usage_seconds_total / on(instance, job) group_left demo_num_cpus`,
	// },
	// TODO: See https://github.com/influxdata/flux/issues/1118
	// {
	// 	query: `demo_cpu_usage_seconds_total / on(instance, mode, job, non_existent) demo_cpu_usage_seconds_total`,
	// },
	// TODO: Add non-explicit many-to-one / one-to-many that errors.
	// TODO: Add many-to-many that errors.
	{
		query:       `{{.simpleAggrOp}}_over_time(demo_cpu_usage_seconds_total[{{.range}}])`,
		variantArgs: []string{"simpleAggrOp", "range"},
	},
	{
		query:       `quantile_over_time({{.quantile}}, demo_cpu_usage_seconds_total[{{.range}}])`,
		variantArgs: []string{"quantile", "range"},
	},
	{
		query: `timestamp(demo_num_cpus)`,
	},
	{
		// Check that vector-vector binops preserve time fields required by aggregations.
		query: `sum(demo_cpu_usage_seconds_total / on(instance, job, mode) demo_cpu_usage_seconds_total)`,
	},
	{
		// Check that scalar-vector binops sets _time field to the window's _stop.
		query: `timestamp(demo_cpu_usage_seconds_total * 1)`,
	},
	{
		// Check that unary minus sets _time field to the window's _stop.
		query: `timestamp(-demo_cpu_usage_seconds_total)`,
	},
	{
		query:       `{{.simpleMathFunc}}(demo_cpu_usage_seconds_total)`,
		variantArgs: []string{"simpleMathFunc"},
	},
	{
		query:       `{{.simpleMathFunc}}(-demo_cpu_usage_seconds_total)`,
		variantArgs: []string{"simpleMathFunc"},
	},
	{
		query:       "{{.extrapolatedRateFunc}}(nonexistent_metric[5m])",
		variantArgs: []string{"extrapolatedRateFunc"},
	},
	{
		query:       "{{.extrapolatedRateFunc}}(demo_cpu_usage_seconds_total[{{.range}}])",
		variantArgs: []string{"extrapolatedRateFunc", "range"},
	},
	{
		query: "sum by(job, instance) (rate(demo_cpu_usage_seconds_total[1m]))",
	},
	{
		query: "time()",
	},
	// Binops involving non-const scalars.
	{
		query:       "1 {{.arithBinOp}} time()",
		variantArgs: []string{"arithBinOp"},
	},
	{
		query:       "time() {{.arithBinOp}} 1",
		variantArgs: []string{"arithBinOp"},
	},
	{
		query:       "time() {{.compBinOp}} bool 1",
		variantArgs: []string{"compBinOp"},
	},
	{
		query:       "1 {{.compBinOp}} bool time()",
		variantArgs: []string{"compBinOp"},
	},
	{
		query:       "time() {{.arithBinOp}} time()",
		variantArgs: []string{"arithBinOp"},
	},
	{
		query:       "time() {{.compBinOp}} bool time()",
		variantArgs: []string{"compBinOp"},
	},
	{
		query:       "time() {{.binOp}} demo_cpu_usage_seconds_total",
		variantArgs: []string{"binOp"},
	},
	{
		query:       "demo_cpu_usage_seconds_total {{.binOp}} time()",
		variantArgs: []string{"binOp"},
	},
}

func TestQueriesE2E(t *testing.T) {
	var runner e2eRunner

	flag.StringVar(&runner.influxURL, "influx-url", "http://localhost:9999/", "InfluxDB server URL.")
	flag.StringVar(&runner.influxBucket, "influx-bucket", "prometheus", "InfluxDB bucket name.")
	flag.StringVar(&runner.influxToken, "influx-token", "", "InfluxDB authentication token.")
	flag.StringVar(&runner.influxOrg, "influx-org", "prometheus", "The InfluxDB organization name.")
	flag.StringVar(&runner.promURL, "prometheus-url", "http://localhost:9090/", "Prometheus server URL.")
	queryStart := flag.Int64("query-start", 1550767830000, "Query start timestamp in milliseconds.")
	queryEnd := flag.Int64("query-end", 1550767900000, "Query end timestamp in milliseconds.")
	//queryStart := flag.Int64("query-start", 1550767200000, "Query start timestamp in milliseconds.")
	//queryEnd := flag.Int64("query-end", 1550770000000, "Query end timestamp in milliseconds.")
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
		runner.runQueryVariants(t, q.query, q.variantArgs, map[string]string{})
	}
}

// tprintf passed template string is formatted usign its operands and returns the resulting string.
// Spaces are added between operands when neither is a string.
func tprintf(tmpl string, data map[string]string) string {
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

// runQueryVariants runs a query with all possible combinations (variants) of query args.
func (r *e2eRunner) runQueryVariants(t *testing.T, query string, variantArgs []string, args map[string]string) {
	// Either this query had no variants defined to begin with or they have
	// been fully filled out in "args" from recursive parent calls.
	if len(variantArgs) == 0 {
		query := tprintf(query, args)
		r.runQuery(t, query)
		return
	}

	// Recursively iterate through the values for each variant arg dimension,
	// selecting one dimension (arg) to vary per recursion level and let the
	// other recursion levels iterate through the remaining dimensions until
	// all args are defined.
	for _, vArg := range variantArgs {
		filteredVArgs := make([]string, 0, len(variantArgs)-1)
		for _, va := range variantArgs {
			if va != vArg {
				filteredVArgs = append(filteredVArgs, va)
			}
		}

		vals := testVariantArgs[vArg]
		if len(vals) == 0 {
			t.Fatalf("Unknown variant arg %q", vArg)
		}
		for _, variantVal := range vals {
			args[vArg] = variantVal
			r.runQueryVariants(t, query, filteredVArgs, args)
		}
	}
}

func (r *e2eRunner) runQuery(t *testing.T, query string) {
	fmt.Println(query)
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
	fluxFile, err := tr.transpile(promqlNode)
	if err != nil {
		t.Fatalf("Error transpiling PromQL expression %q to Flux: %s", query, err)
	}

	// Query both Prometheus and InfluxDB, expect same result.
	promMatrix, err := queryPrometheus(r.promURL, query, r.start, r.end, r.resolution)
	if err != nil {
		t.Fatalf("Error querying Prometheus for %q: %s", query, err)
	}
	influxResult, err := queryInfluxDB(r.influxURL, r.influxOrg, r.influxToken, r.influxBucket, ast.Format(fluxFile))
	if err != nil {
		t.Fatalf("Error querying InfluxDB for %q: %s", query, err)
	}
	// Make InfluxDB result comparable with the Prometheus result.
	influxMatrix, err := influxResultToPromMatrix(influxResult)
	if err != nil {
		t.Fatalf("Error processing InfluxDB results for %q\n\n%s: %s", query, ast.Format(fluxFile), err)
	}

	cmpOpts := cmp.Options{
		// Translate sample values into float64 so that cmpopts.EquateApprox() works.
		cmp.Transformer("", func(in model.SampleValue) float64 {
			return float64(in)
		}),
		// Allow comparison tolerances due to floating point inaccuracy.
		cmpopts.EquateApprox(0.00000000000001, 0),
		cmpopts.EquateNaNs(),
	}
	if diff := cmp.Diff(promMatrix, influxMatrix, cmpOpts); diff != "" {
		t.Fatal(
			"FAILED! Prometheus and InfluxDB results differ:\n\n", diff,
			"\nPromQL query was:\n============================================\n", query, "\n============================================\n\n",
			"\nFlux query was:\n============================================\n", ast.Format(fluxFile), "\n============================================\n\n",
			"\nFull results:",
			"\n=== InfluxDB results:\n", influxMatrix,
			"\n=== Prometheus results:\n", promMatrix,
		)
	}
}
