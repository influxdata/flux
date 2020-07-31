package cloudwatch

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type accumulator struct {
	metrics []telegraf.Metric
}

func (a *accumulator) AddFields(measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time) {
	tm := time.Now().UTC()
	if len(t) > 0 {
		tm = t[0]
	}
	m, _ := metric.New(measurement, tags, fields, tm)
	a.AddMetric(m)
}

// AddGauge is the same as AddFields, but will add the metric as a "Gauge" type
func (a *accumulator) AddGauge(measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time) {

}

// AddCounter is the same as AddFields, but will add the metric as a "Counter" type
func (a *accumulator) AddCounter(measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time) {

}

// AddSummary is the same as AddFields, but will add the metric as a "Summary" type
func (a *accumulator) AddSummary(measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time) {

}

// AddHistogram is the same as AddFields, but will add the metric as a "Histogram" type
func (a *accumulator) AddHistogram(measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time) {

}

// AddMetric adds an metric to the accumulator.
func (a *accumulator) AddMetric(metric telegraf.Metric) {
	a.metrics = append(a.metrics, metric)
}

// SetPrecision sets the timestamp rounding precision.  All metrics addeds
// added to the accumulator will have their timestamp rounded to the
// nearest multiple of precision.
func (a *accumulator) SetPrecision(precision time.Duration) {
}

// Report an error.
func (a *accumulator) AddError(err error) {

}

// Upgrade to a TrackingAccumulator with space for maxTracked
// metrics/batches.
func (a *accumulator) WithTracking(maxTracked int) telegraf.TrackingAccumulator {
	return nil
}
