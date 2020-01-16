package prometheus

import (
	// Go stdlib and other packages

	"context"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/matttproud/golang_protobuf_extensions/pbutil"

	// Flux packages
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"

	// Prometheus packages
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

const ScrapePrometheusKind = "scrapePrometheus"

type ScrapePrometheusOpSpec struct {
	URL string `json:"token,omitempty"`
}

func init() {
	scrapePrometheusSignature := semantic.MustLookupBuiltinType("experimental/prometheus", "scrape")
	flux.RegisterPackageValue("experimental/prometheus", "scrape", flux.MustValue(flux.FunctionValue(ScrapePrometheusKind, createScrapePrometheusOpSpec, scrapePrometheusSignature)))
	flux.RegisterOpSpec(ScrapePrometheusKind, newScrapePrometheusOp)
	plan.RegisterProcedureSpec(ScrapePrometheusKind, newScrapePrometheusProcedure, ScrapePrometheusKind)
	execute.RegisterSource(ScrapePrometheusKind, createScrapePrometheusSource)
}

func createScrapePrometheusOpSpec(args flux.Arguments, administration *flux.Administration) (flux.OperationSpec, error) {
	spec := new(ScrapePrometheusOpSpec)

	if url, err := args.GetRequiredString("url"); err != nil {
		return nil, err
	} else {
		spec.URL = url
	}
	return spec, nil
}

func newScrapePrometheusOp() flux.OperationSpec {
	return new(ScrapePrometheusOpSpec)
}

func (s *ScrapePrometheusOpSpec) Kind() flux.OperationKind {
	return ScrapePrometheusKind
}

type ScrapePrometheusProcedureSpec struct {
	plan.DefaultCost
	URL string
}

func newScrapePrometheusProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*ScrapePrometheusOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Invalid, "invalid spec type %T", qs)
	}

	return &ScrapePrometheusProcedureSpec{
		URL: spec.URL,
	}, nil
}

func (s *ScrapePrometheusProcedureSpec) Kind() plan.ProcedureKind {
	return ScrapePrometheusKind
}

func (s *ScrapePrometheusProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(ScrapePrometheusProcedureSpec)
	ns.URL = s.URL
	return ns
}

func createScrapePrometheusSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec, ok := prSpec.(*ScrapePrometheusProcedureSpec)
	if !ok {
		return nil, errors.Newf(codes.Invalid, "invalid spec type %T", prSpec)
	}
	c := execute.NewTableBuilderCache(a.Allocator())
	c.SetTriggerSpec(plan.DefaultTriggerSpec)
	PrometheusIterator := PrometheusIterator{
		id:             dsid,
		spec:           spec,
		administration: a,
		cache:          c,
	}

	return execute.CreateSourceFromDecoder(&PrometheusIterator, dsid, a)
}

type PrometheusIterator struct {
	NowFn          func() time.Time // Convert times
	id             execute.DatasetID
	administration execute.Administration
	cache          execute.TableBuilderCache
	spec           *ScrapePrometheusProcedureSpec

	metrics []Metric // Slice of metrics to convert to tables
	i       int
	url     string // Store user defined url
	resp    *http.Response
	now     time.Time
}

// Metric stores the fields that we need to construct Table
type Metric struct {
	Field     string                 // Prometheus metric name
	Tags      map[string]string      // key is tag name; val is tag value
	TypeVal   map[string]interface{} // key is metric type; val is metric value
	Timestamp time.Time
	Type      string // Prometheus metric type
}

// This implementation of Connect takes in a user defined url, validates the url
// and gets an http response. It then calls parse to parse the body into a list
// Metrics or returns and error if not given a valid prometheus metric endpoint.
func (p *PrometheusIterator) Connect(ctx context.Context) error {
	p.url = p.spec.URL // Attach url to Prometheus Iterator

	if p.NowFn != nil {
		p.now = p.NowFn()
	} else {
		p.now = time.Now()
	}

	u, err := url.Parse(p.url)
	if err != nil {
		return err
	}

	// Validate url
	deps := flux.GetDependencies(ctx)
	validator, err := deps.URLValidator()
	if err != nil {
		return err
	}
	if err := validator.Validate(u); err != nil {
		return err
	}

	// Get response
	resp, err := http.Get(p.url)
	if err != nil {
		return err
	}
	p.resp = resp
	defer resp.Body.Close()

	// Parse the response body into list of Metrics
	err = p.parse(resp.Body, resp.Header)
	if err != nil {
		return err
	}

	return nil
}

// parse will take in an http header, and read the body of an http response. It looks for prometheus
// Metrics and calls either makeQuantiles, makeBuckets or getNameandValue depending on each Metric
// type. It produces a list of type Metric and stores them in p.metrics.
func (p *PrometheusIterator) parse(reader io.Reader, header http.Header) (err error) {
	var parser expfmt.TextParser

	mediatype, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		return err
	}

	metricFamilies := make(map[string]*dto.MetricFamily)
	if mediatype == "application/vnd.google.protobuf" &&
		params["encoding"] == "delimited" &&
		params["proto"] == "io.prometheus.client.MetricFamily" {
		for {
			mf := &dto.MetricFamily{}
			if _, err := pbutil.ReadDelimited(reader, mf); err != nil {
				if err == io.EOF {
					break
				}
				return errors.Newf(codes.Internal, "reading metric family protocol buffer failed: %s", err)
			}
			metricFamilies[mf.GetName()] = mf
		}
	} else {
		metricFamilies, err = parser.TextToMetricFamilies(reader)
		if err != nil {
			return errors.Newf(codes.Internal, "reading text format failed: %s", err)
		}
	}
	p.metrics = make([]Metric, 0)

	// Read metrics
	for field, family := range metricFamilies {
		for _, metr := range family.Metric {
			// Read tags
			tags := makeLabels(metr)
			switch family.GetType() {

			// Metric Type: Summary
			case dto.MetricType_SUMMARY:
				makeMetrics := p.makeQuantiles(metr, tags, field, family.GetType())
				p.metrics = append(p.metrics, makeMetrics...)

			// Metric Type: Histogram
			case dto.MetricType_HISTOGRAM:
				makeMetrics := p.makeBuckets(metr, tags, field, family.GetType())
				p.metrics = append(p.metrics, makeMetrics...)

			// Metric Type: Gague, Counter, Untyped
			default:
				typeValue := getNameAndValue(metr, field)

				if len(typeValue) > 0 {
					var t time.Time
					if metr.TimestampMs != nil && *metr.TimestampMs > 0 {
						t = time.Unix(0, *metr.TimestampMs*int64(time.Millisecond))
					} else {
						t = p.now
					}
					met := Metric{
						Timestamp: t,
						Tags:      tags,
						TypeVal:   typeValue,
						Field:     field,
						Type:      family.GetType().String(),
					}
					p.metrics = append(p.metrics, met)
				}
			}
		}
	}
	return nil
}

// This implementation of Fetch will iterate over p.metrics
func (p *PrometheusIterator) Fetch(ctx context.Context) (bool, error) {

	// Iterate over all Metrics
	if p.i < len(p.metrics) {
		// Grab the next metric in list
		return true, nil
	}

	// No more metrics to return
	return false, nil
}

// This implementation of Decode will create flux Tables for a give Metric. It retrieves one Metric
// from p.metrics and places it into a flux.Table
func (p *PrometheusIterator) Decode(ctx context.Context) (table flux.Table, err error) {
	met := p.metrics[p.i]

	// Unpacking TypeVal map
	var val interface{}
	for _, v := range met.TypeVal {
		val = v
	}

	groupKey := execute.NewGroupKeyBuilder(nil)
	groupKey.AddKeyValue("_measurement", values.New("prometheus"))
	groupKey.AddKeyValue("_field", values.New(met.Field))

	// Add all tag names to Group Key
	gkInt := 2
	for name, val := range met.Tags {
		gkInt++
		if groupKey.Len() < gkInt {
			groupKey.AddKeyValue(name, values.New(val))
		}
	}

	gk, err := groupKey.Build()
	if err != nil {
		return nil, err
	}

	builder := execute.NewColListTableBuilder(gk, p.administration.Allocator())

	builder.AddCol(flux.ColMeta{
		Label: "_time",
		Type:  flux.TTime,
	})
	builder.AddCol(flux.ColMeta{
		Label: "_value",
		Type:  flux.TFloat,
	})
	builder.AddCol(flux.ColMeta{
		Label: "_measurement", // data source
		Type:  flux.TString,
	})
	builder.AddCol(flux.ColMeta{
		Label: "_field", // prometheus metric name
		Type:  flux.TString,
	})
	builder.AddCol(flux.ColMeta{
		Label: "url",
		Type:  flux.TString,
	})

	// Add all tags to Col list
	for name := range met.Tags {
		if execute.ColIdx(name, builder.Cols()) == -1 {
			builder.AddCol(flux.ColMeta{
				Label: name,
				Type:  flux.TString,
			})
		}
	}

	builder.AppendTime(0, values.ConvertTime(met.Timestamp))
	builder.AppendValue(1, values.New(val))
	builder.AppendValue(2, values.New("prometheus"))
	builder.AppendValue(3, values.New(met.Field))
	builder.AppendValue(4, values.New(p.url))

	// Add tag values
	for name, tagVal := range met.Tags {
		builder.AppendValue(execute.ColIdx(name, builder.Cols()), values.New(tagVal))
	}

	// Grab the next metric in list
	p.i++

	return builder.Table()
}

func (p *PrometheusIterator) Close() error {
	// nothing to close
	return nil
}

// makeBuckets will return a list of summary values of type Metric given the prometheus metric, tags,
// name and metric type
func (p *PrometheusIterator) makeQuantiles(m *dto.Metric, tags map[string]string, metricName string, metricType dto.MetricType) []Metric {
	var metrics []Metric
	typeValue := make(map[string]interface{})
	var t time.Time
	if m.TimestampMs != nil && *m.TimestampMs > 0 {
		t = time.Unix(0, *m.TimestampMs*int64(time.Millisecond))
	} else {
		t = p.now
	}

	countName := metricName + "_count"
	typeValue[countName] = float64(m.GetSummary().GetSampleCount())
	countMet := Metric{
		Timestamp: t,
		Tags:      tags,
		TypeVal:   typeValue,
		Field:     countName,
		Type:      "summary",
	}

	// Clear map for sum values
	typeValue = make(map[string]interface{})
	sumName := metricName + "_sum"
	typeValue[sumName] = float64(m.GetSummary().GetSampleSum())
	sumMet := Metric{
		Timestamp: t,
		Tags:      tags,
		TypeVal:   typeValue,
		Field:     sumName,
		Type:      "summary",
	}
	metrics = append(metrics, countMet, sumMet)

	for _, q := range m.GetSummary().Quantile {
		newTags := make(map[string]string)
		for k, v := range tags {
			newTags[k] = v
		}

		typeValue = make(map[string]interface{})
		if !math.IsNaN(q.GetValue()) {
			newTags["quantile"] = fmt.Sprint(q.GetQuantile())
			typeValue[metricName] = float64(q.GetValue())
			met := Metric{
				Timestamp: t,
				Tags:      newTags,
				TypeVal:   typeValue,
				Field:     metricName,
				Type:      "summary",
			}
			metrics = append(metrics, met)
		}
	}
	return metrics
}

// makeBuckets will return a list of histogram values of type Metric given the prometheus metric, taags,
// name and metric type
func (p *PrometheusIterator) makeBuckets(m *dto.Metric, tags map[string]string, metricName string, metricType dto.MetricType) []Metric {
	var metrics []Metric
	typeValue := make(map[string]interface{})

	var t time.Time
	if m.TimestampMs != nil && *m.TimestampMs > 0 {
		t = time.Unix(0, *m.TimestampMs*int64(time.Millisecond))
	} else {
		t = p.now
	}

	countName := metricName + "_count"
	typeValue[countName] = float64(m.GetHistogram().GetSampleCount())
	countMet := Metric{
		Timestamp: t,
		Tags:      tags,
		TypeVal:   typeValue,
		Field:     countName,
		Type:      "histogram",
	}

	typeValue[metricName+"_sum"] = float64(m.GetHistogram().GetSampleSum())

	sumName := metricName + "_sum"
	typeValue[sumName] = float64(m.GetSummary().GetSampleSum())
	sumMet := Metric{
		Timestamp: t,
		Tags:      tags,
		TypeVal:   typeValue,
		Field:     sumName,
		Type:      "histogram",
	}
	metrics = append(metrics, countMet, sumMet)

	for _, b := range m.GetHistogram().Bucket {
		newTags := make(map[string]string)
		for k, v := range tags {
			newTags[k] = v
		}

		typeValue = make(map[string]interface{})
		newTags["le"] = fmt.Sprint(b.GetUpperBound())
		typeValue[metricName] = float64(b.GetCumulativeCount())

		met := Metric{
			Timestamp: t,
			Tags:      newTags,
			TypeVal:   typeValue,
			Field:     metricName,
			Type:      "histogram",
		}
		metrics = append(metrics, met)
	}
	return metrics
}

// makeLabels will return all labels on a given metric
func makeLabels(m *dto.Metric) map[string]string {
	result := map[string]string{}
	for _, lp := range m.Label {
		result[lp.GetName()] = lp.GetValue()
	}
	return result
}

// getNameandValue will return the metric name and value for a given counter, gague and or untyped value
func getNameAndValue(m *dto.Metric, metricName string) map[string]interface{} {
	nameVal := make(map[string]interface{})
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			nameVal[metricName] = float64(m.GetGauge().GetValue())
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) {
			nameVal[metricName] = float64(m.GetCounter().GetValue())
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetUntyped().GetValue()) {
			nameVal[metricName] = float64(m.GetUntyped().GetValue())
		}
	}
	return nameVal
}
