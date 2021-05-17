package internal

import (
	"time"

	lp "github.com/influxdata/line-protocol"
)

// RowMetric is a Metric
type RowMetric struct {
	NameStr string
	Tags    []*lp.Tag
	Fields  []*lp.Field
	TS      time.Time
}

func (r RowMetric) Time() time.Time {
	return r.TS
}

func (r RowMetric) Name() string {
	return r.NameStr
}

func (r RowMetric) TagList() []*lp.Tag {
	return r.Tags
}

func (r RowMetric) FieldList() []*lp.Field {
	return r.Fields
}

var _ lp.Metric = &RowMetric{}
