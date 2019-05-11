package promflux

import (
	"testing"
	"time"
)

var testVariantArgs = map[string][]string{
	"range":        []string{"1s", "15s", "1m", "5m", "15m", "1h"},
	"offset":       []string{"1m", "5m", "10m"},
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
	"arithBinOp":           []string{"+", "-", "*", "/", "%", "^"},
	"compBinOp":            []string{"==", "!=", "<", ">", "<=", ">="},
	"binOp":                []string{"+", "-", "*", "/", "%", "^", "==", "!=", "<", ">", "<=", ">="},
	"simpleMathFunc":       []string{"abs", "ceil", "floor", "exp", "sqrt", "ln", "log2", "log10", "round"},
	"extrapolatedRateFunc": []string{"delta", "rate", "increase"},
	"clampFunc":            []string{"clamp_min", "clamp_max"},
	"instantRateFunc":      []string{"idelta"},
	"dateFunc":             []string{"day_of_month", "day_of_week", "days_in_month", "hour", "minute", "month", "year"},
}

var queries = []struct {
	query       string
	variantArgs []string
	// Needed for subqueries, which will never return 100% identical results.
	skipComparison bool
}{
	// Vector selectors.
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

	// Aggregation operators.
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

	// Binary operators.
	{
		query:       `1 * 2 + 4 / 6 - 10`,
		variantArgs: []string{""},
	},
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

	// TODO: Check this systematically for every node type.
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

	// TODO: Blocked on new join() implementation. https://github.com/influxdata/flux/issues/1219
	// {
	// 	query:       `demo_cpu_usage_seconds_total {{.binOp}} on(instance, job, mode) demo_cpu_usage_seconds_total`,
	// 	variantArgs: []string{"binOp"},
	// },
	// {
	// 	query:       `sum by(instance, mode) (demo_cpu_usage_seconds_total) {{.binOp}} on(instance, mode) group_left(job) demo_cpu_usage_seconds_total`,
	// 	variantArgs: []string{"binOp"},
	// },
	// {
	// 	query:       `demo_cpu_usage_seconds_total {{.compBinOp}} bool on(instance, job, mode) demo_cpu_usage_seconds_total`,
	// 	variantArgs: []string{"compBinOp"},
	// },
	// {
	// 	// Check that __name__ is always dropped, even if it's part of the matching labels.
	// 	query: `demo_cpu_usage_seconds_total / on(instance, job, mode, __name__) demo_cpu_usage_seconds_total`,
	// },
	// {
	// 	query: `sum without(job) (demo_cpu_usage_seconds_total) / on(instance, mode) demo_cpu_usage_seconds_total`,
	// },
	// {
	// 	query: `sum without(job) (demo_cpu_usage_seconds_total) / on(instance, mode) group_left demo_cpu_usage_seconds_total`,
	// },
	// {
	// 	query: `sum without(job) (demo_cpu_usage_seconds_total) / on(instance, mode) group_left(job) demo_cpu_usage_seconds_total`,
	// },

	// {
	// 	query: `demo_cpu_usage_seconds_total / on(instance, job) group_left demo_num_cpus`,
	// },
	// TODO: See https://github.com/influxdata/flux/issues/1118
	// {
	// 	query: `demo_cpu_usage_seconds_total / on(instance, mode, job, non_existent) demo_cpu_usage_seconds_total`,
	// },
	// TODO: Add non-explicit many-to-one / one-to-many that errors.
	// TODO: Add many-to-many that errors.

	// NaN/Inf/-Inf support.
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

	// Unary expressions.
	{
		query: `demo_cpu_usage_seconds_total + -(1 + 1)`,
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
	// {
	// 	query:       "time() {{.binOp}} demo_cpu_usage_seconds_total",
	// 	variantArgs: []string{"binOp"},
	// },
	// {
	// 	query:       "demo_cpu_usage_seconds_total {{.binOp}} time()",
	// 	variantArgs: []string{"binOp"},
	// },

	// Functions.
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
	{
		query: `label_join(demo_num_cpus, "new_label", "-", "instance", "job")`,
	},
	{
		query: `label_join(demo_num_cpus, "job", "-", "instance", "job")`,
	},
	{
		query: `label_join(demo_num_cpus, "job", "-", "instance")`,
	},
	{
		query:       `{{.dateFunc}}()`,
		variantArgs: []string{"dateFunc"},
	},
	{
		query:       `{{.dateFunc}}(demo_batch_last_success_timestamp_seconds)`,
		variantArgs: []string{"dateFunc"},
	},
	{
		query:       `{{.instantRateFunc}}(demo_cpu_usage_seconds_total[{{.range}}])`,
		variantArgs: []string{"instantRateFunc", "range"},
	},
	{
		query:       `{{.clampFunc}}(demo_cpu_usage_seconds_total, 2)`,
		variantArgs: []string{"clampFunc"},
	},
	{
		query:       `resets(demo_cpu_usage_seconds_total[{{.range}}])`,
		variantArgs: []string{"range"},
	},
	{
		query:       `changes(demo_batch_last_success_timestamp_seconds[{{.range}}])`,
		variantArgs: []string{"range"},
	},
	{
		query: `vector(1)`,
	},
	{
		query: `vector(time())`,
	},

	// Subqueries. Comparisons are skipped since the implementation cannot guarantee completely identical results.
	{
		query:          `max_over_time((time() - max(demo_batch_last_success_timestamp_seconds) < 1000)[5m:10s] offset 5m)`,
		skipComparison: true,
	},
	{
		query:          `avg_over_time(rate(demo_cpu_usage_seconds_total[1m])[2m:10s])`,
		skipComparison: true,
	},
}

func TestQueries(t *testing.T) {
	var (
		start      = time.Unix(1550767830, 0).UTC()
		end        = time.Unix(1550767900, 0).UTC()
		resolution = 10 * time.Second
	)

	querier := newTestQuerier(t, "testdata/prom-data", testVariantArgs, start, end, resolution)
	defer querier.close()

	for _, q := range queries {
		querier.runQueryVariants(q.query, q.variantArgs, map[string]string{}, q.skipComparison)
	}
}
