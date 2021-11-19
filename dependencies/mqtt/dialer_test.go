package mqtt_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux/dependencies/mqtt"
)

func TestGetNoDialer(t *testing.T) {
	ctx := context.Background()

	got := mqtt.GetDialer(ctx)
	if _, ok := got.(mqtt.ErrorDialer); !ok {
		t.Fatalf("expected error dialer, got:\n%T", got)
	}
}
