package jaeger

import (
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/uber/jaeger-client-go"
)

const (
	traceIDDecimal = 10
	traceIDHex     = "000000000000000a"
)

func TestTracingInfo(t *testing.T) {
	span := mockSpan{
		MockSpan: mocktracer.New().StartSpan("test").(*mocktracer.MockSpan),
	}
	span.SpanContext.Sampled = true
	span.SpanContext.TraceID = traceIDDecimal
	span.ctx = mockSpanContext{span.SpanContext}

	traceID, sampled, found := InfoFromSpan(span)
	if !found {
		t.Fatal("trace ID not found in span context")
	}
	if traceID != traceIDHex {
		t.Fatalf("trace ID does not match actual=%v expected=%v", traceID, traceIDHex)
	}
	if !sampled {
		t.Fatal("trace ID found but not sampled")
	}
}

type mockSpan struct {
	ctx mockSpanContext
	*mocktracer.MockSpan
}

func (m mockSpan) Context() opentracing.SpanContext {
	return m.ctx
}

type mockSpanContext struct {
	mocktracer.MockSpanContext
}

func (m mockSpanContext) TraceID() jaeger.TraceID {
	return jaeger.TraceID{High: 0, Low: uint64(m.MockSpanContext.TraceID)}
}

func (m mockSpanContext) IsSampled() bool {
	return m.MockSpanContext.Sampled
}
