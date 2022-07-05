package influxdb

import (
	"time"

	"github.com/influxdata/flux/dependencies/influxdb"
)

// RowMetric is a Metric
type RowMetric struct {
	NameStr string
	Tags    []*influxdb.Tag
	Fields  []*influxdb.Field
	TS      time.Time
}

func (r RowMetric) Time() time.Time {
	return r.TS
}

func (r RowMetric) Name() string {
	return r.NameStr
}

func (r RowMetric) TagList() []*influxdb.Tag {
	return r.Tags
}

func (r RowMetric) FieldList() []*influxdb.Field {
	return r.Fields
}

var _ influxdb.Metric = &RowMetric{}
