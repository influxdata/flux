package execkit

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/opentracing/opentracing-go"
)

type srcMessage execute.DatasetID

func (m srcMessage) SrcDatasetID() DatasetID {
	return DatasetID(m)
}
func (m srcMessage) Ack() {}
func (m srcMessage) SetTags(span opentracing.Span) {
	span.SetTag("dataset", DatasetID(m).String())
}

type finishMsg struct {
	srcMessage
	err error
}

func (m *finishMsg) Type() execute.MessageType {
	return execute.FinishType
}
func (m *finishMsg) Error() error {
	return m.err
}
func (m *finishMsg) Dup() execute.Message {
	return m
}
func (m *finishMsg) SetTags(span opentracing.Span) {
	m.srcMessage.SetTags(span)
	if m.err != nil {
		span.SetTag("error", m.err.Error())
	}
}

type processViewMsg struct {
	srcMessage
	view table.View
}

func (m *processViewMsg) Type() execute.MessageType {
	return execute.ProcessViewType
}
func (m *processViewMsg) View() table.View {
	return m.view
}
func (m *processViewMsg) Ack() {
	m.view.Release()
}
func (m *processViewMsg) Dup() execute.Message {
	m.view.Retain()
	return m
}
func (m *processViewMsg) SetTags(span opentracing.Span) {
	m.srcMessage.SetTags(span)
	span.SetTag("key", m.view.Key().String())
}

type flushKeyMsg struct {
	srcMessage
	key flux.GroupKey
}

func (m *flushKeyMsg) Type() execute.MessageType {
	return execute.FlushKeyType
}
func (m *flushKeyMsg) Key() flux.GroupKey {
	return m.key
}
func (m *flushKeyMsg) Dup() execute.Message {
	return m
}
func (m *flushKeyMsg) SetTags(span opentracing.Span) {
	m.srcMessage.SetTags(span)
	span.SetTag("key", m.key.String())
}

type watermarkKeyMsg struct {
	srcMessage
	columnName string
	watermark  int64
	key        flux.GroupKey
}

func (m *watermarkKeyMsg) Type() execute.MessageType {
	return execute.WatermarkKeyType
}
func (m *watermarkKeyMsg) ColumnName() string {
	return m.columnName
}
func (m *watermarkKeyMsg) Time() int64 {
	return m.watermark
}
func (m *watermarkKeyMsg) Key() flux.GroupKey {
	return m.key
}
func (m *watermarkKeyMsg) Dup() execute.Message {
	return m
}
func (m *watermarkKeyMsg) SetTags(span opentracing.Span) {
	m.srcMessage.SetTags(span)
	span.SetTag("column", m.columnName)
	span.SetTag("time", m.watermark)
	span.SetTag("key", m.key.String())
}
