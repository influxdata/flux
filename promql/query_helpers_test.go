package promflux

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/influxdb/cmd/influxd/launcher"
	"github.com/influxdata/influxdb/query"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/tsdb"
	"github.com/prometheus/tsdb/wal"

	_ "github.com/influxdata/flux/builtin"
)

type testQuerier struct {
	t          *testing.T
	ctx        context.Context
	promDB     storage.Storage
	promEngine *promql.Engine
	influxDB   *launcher.TestLauncher

	variantArgs map[string][]string
	start       time.Time
	end         time.Time
	resolution  time.Duration
}

func newTestQuerier(
	t *testing.T,
	tsdbPath string,
	variantArgs map[string][]string,
	start time.Time,
	end time.Time,
	resolution time.Duration) *testQuerier {

	ctx := context.Background()

	l := launcher.RunTestLauncherOrFail(t, ctx)
	l.SetupOrFail(t)

	db, err := tsdb.Open(tsdbPath, nil, nil, &tsdb.Options{
		WALSegmentSize:    wal.DefaultSegmentSize,
		RetentionDuration: 99999 * 24 * 60 * 60 * model.Duration(time.Second),
		MinBlockDuration:  model.Duration(2 * time.Hour),
		MaxBlockDuration:  model.Duration(2 * time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	db.DisableCompactions()

	q := &testQuerier{
		t:      t,
		ctx:    ctx,
		promDB: tsdb.Adapter(db, 0),
		promEngine: promql.NewEngine(promql.EngineOpts{
			MaxConcurrent: 10,
			MaxSamples:    1e12,
			Timeout:       time.Hour,
		}),
		influxDB: l,

		start:      start,
		end:        end,
		resolution: resolution,
	}
	q.transferSamples()

	return q
}

func (q *testQuerier) close() {
	if err := q.promDB.Close(); err != nil {
		q.t.Fatal(err)
	}
	q.influxDB.ShutdownOrFail(q.t, q.ctx)
}

func (q *testQuerier) transferSamples() {
	pq, err := q.promDB.Querier(q.ctx, math.MinInt64, math.MaxInt64)
	if err != nil {
		log.Fatal(err)
	}
	defer pq.Close()

	matcher, err := labels.NewMatcher(labels.MatchRegexp, "__name__", ".*")
	if err != nil {
		q.t.Fatal("Error creating label matcher:", err)
	}
	ss, warnings, err := pq.Select(nil, matcher)
	if warnings != nil {
		log.Fatal(warnings)
	}
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	for ss.Next() {
		series := ss.At()
		labels := series.Labels()
		tags := make([]string, 0, len(labels))
		field := ""
		for _, l := range labels {
			if l.Name == "__name__" {
				field = l.Value
				continue
			}
			tags = append(tags, escapeInfluxDBChars(l.Name)+"="+escapeInfluxDBChars(l.Value))
		}
		if field == "" {
			q.t.Fatalf("no metric name found in series %v", labels)
		}
		it := series.Iterator()
		for it.Next() {
			ts, val := it.At()
			if math.IsNaN(val) {
				// TODO: InfluxDB does not support NaNs yet, skip these for now.
				continue
			}
			_, err := buf.WriteString(fmt.Sprintf("prometheus,%s %s=%s %d\n", strings.Join(tags, ","), field, strconv.FormatFloat(val, 'f', -1, 64), ts*1e6))
			if err != nil {
				q.t.Fatal(err)
			}
		}
		if it.Err() != nil {
			log.Fatal(ss.Err())
		}
	}

	if ss.Err() != nil {
		log.Fatal(ss.Err())
	}

	q.influxDB.WritePointsOrFail(q.t, buf.String())
}

func escapeInfluxDBChars(str string) string {
	specialChars := []string{`,`, `=`, ` `, `\`}
	for _, c := range specialChars {
		str = strings.Replace(str, c, `\`+c, -1)
	}
	return str
}

// tprintf replaces template arguments in a string with their instantiations from the provided map.
func tprintf(tmpl string, data map[string]string) string {
	t := template.Must(template.New("query").Parse(tmpl))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, data); err != nil {
		panic(err)
	}
	return buf.String()
}

// runQueryVariants runs a query with all possible combinations (variants) of query args.
func (q *testQuerier) runQueryVariants(query string, variantArgs []string, args map[string]string, skipComparison bool) {
	// Either this query had no variants defined to begin with or they have
	// been fully filled out in "args" from recursive parent calls.
	if len(variantArgs) == 0 {
		query := tprintf(query, args)
		q.runQuery(query, skipComparison)
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
			q.t.Fatalf("Unknown variant arg %q", vArg)
		}
		for _, variantVal := range vals {
			args[vArg] = variantVal
			q.runQueryVariants(query, filteredVArgs, args, skipComparison)
		}
	}
}

func (q *testQuerier) runQuery(qry string, skipComparison bool) {
	// Transpile PromQL into Flux.
	promqlNode, err := promql.ParseExpr(qry)
	if err != nil {
		q.t.Fatalf("Error parsing PromQL expression %q: %s", qry, err)
	}
	tr := &Transpiler{
		Bucket:     q.influxDB.Bucket.Name,
		Start:      q.start,
		End:        q.end,
		Resolution: q.resolution,
	}
	fluxFile, err := tr.Transpile(promqlNode)
	if err != nil {
		q.t.Fatalf("Error transpiling PromQL expression %q to Flux: %s", qry, err)
	}

	// Query Prometheus.
	rq, err := q.promEngine.NewRangeQuery(q.promDB, qry, q.start, q.end, q.resolution)
	if err != nil {
		q.t.Fatalf("Error creating Prometheus range query for %q: %s", qry, err)
	}
	defer rq.Close()
	promResult := rq.Exec(q.ctx)
	if promResult.Err != nil {
		q.t.Fatalf("Error querying Prometheus for %q: %s", qry, promResult.Err)
	}
	promMatrix, err := promResult.Matrix()
	if err != nil {
		q.t.Fatalf("Error converting Prometheus result for %q to Matrix: %s", qry, err)
	}

	// Query InfluxDB.
	req := &query.Request{
		Authorization:  q.influxDB.Auth,
		OrganizationID: q.influxDB.Org.ID,
		Compiler: lang.ASTCompiler{
			AST: &ast.Package{Package: "main", Files: []*ast.File{fluxFile}},
			Now: time.Now(),
		},
	}
	var influxMatrix promql.Value = promql.Matrix{}
	err = q.influxDB.QueryAndConsume(q.ctx, req, func(r flux.Result) error {
		influxMatrix = FluxResultToPromQLValue(r, promql.ValueTypeMatrix)
		return nil
	})
	if err != nil {
		q.t.Fatalf("Error querying InfluxDB for %q: %s\n\nFlux script:%s", qry, err, ast.Format(fluxFile))
	}

	if skipComparison {
		return
	}

	cmpOpts := cmp.Options{
		// Translate sample values into float64 so that cmpopts.EquateApprox() works.
		cmp.Transformer("", func(in model.SampleValue) float64 {
			return float64(in)
		}),
		// Allow comparison tolerances due to floating point inaccuracy.
		cmpopts.EquateApprox(0.0000000000001, 0),
		cmpopts.EquateNaNs(),
		cmpopts.EquateEmpty(),
	}
	if diff := cmp.Diff(promMatrix, influxMatrix, cmpOpts); diff != "" {
		q.t.Fatal(
			"FAILED! Prometheus and InfluxDB results differ:\n\n", diff,
			"\nPromQL query was:\n============================================\n", qry, "\n============================================\n\n",
			"\nFlux query was:\n============================================\n", ast.Format(fluxFile), "\n============================================\n\n",
			"\nFull results:",
			"\n=== InfluxDB results:\n", influxMatrix,
			"\n=== Prometheus results:\n", promMatrix,
		)
	}
}
