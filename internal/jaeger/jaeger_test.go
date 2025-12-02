package jaeger

import (
	"testing"

	"go.opentelemetry.io/otel/trace"
)

const (
	traceIDHex = "00000000000000000000000000000001"
)

func TestTracingInfo(t *testing.T) {
	traceID, err := trace.TraceIDFromHex(traceIDHex)
	if err != nil {
		t.Fatalf("failed to parse trace ID: %v", err)
	}
	spanID, err := trace.SpanIDFromHex("0000000000000001")
	if err != nil {
		t.Fatalf("failed to parse span ID: %v", err)
	}

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	span := mockSpan{sc: sc}

	gotTraceID, sampled, found := InfoFromSpan(span)
	if !found {
		t.Fatal("trace ID not found in span context")
	}
	if gotTraceID != traceIDHex {
		t.Fatalf("trace ID does not match actual=%v expected=%v", gotTraceID, traceIDHex)
	}
	if !sampled {
		t.Fatal("trace ID found but not sampled")
	}
}

func TestTracingInfo_NotSampled(t *testing.T) {
	traceID, err := trace.TraceIDFromHex(traceIDHex)
	if err != nil {
		t.Fatalf("failed to parse trace ID: %v", err)
	}
	spanID, err := trace.SpanIDFromHex("0000000000000001")
	if err != nil {
		t.Fatalf("failed to parse span ID: %v", err)
	}

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: 0, // Not sampled
	})

	span := mockSpan{sc: sc}

	gotTraceID, sampled, found := InfoFromSpan(span)
	if !found {
		t.Fatal("trace ID not found in span context")
	}
	if gotTraceID != traceIDHex {
		t.Fatalf("trace ID does not match actual=%v expected=%v", gotTraceID, traceIDHex)
	}
	if sampled {
		t.Fatal("trace ID found but should not be sampled")
	}
}

func TestTracingInfo_InvalidSpan(t *testing.T) {
	span := mockSpan{sc: trace.SpanContext{}}

	_, _, found := InfoFromSpan(span)
	if found {
		t.Fatal("expected trace ID not to be found for invalid span")
	}
}

// mockSpan implements trace.Span for testing purposes
type mockSpan struct {
	trace.Span
	sc trace.SpanContext
}

func (m mockSpan) SpanContext() trace.SpanContext {
	return m.sc
}
