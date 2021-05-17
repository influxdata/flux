package execute

import (
	"context"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestDispatcher_Stop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	d := newPoolDispatcher(10, zaptest.NewLogger(t))
	d.Start(100, ctx)

	for i := 0; i < 100; i++ {
		d.Schedule(func(ctx context.Context, throughput int) {
			<-ctx.Done()
			panic("expected")
		})
	}
	cancel()

	if err := d.Stop(); err == nil {
		t.Fatal("expected error")
	} else if got, want := err.Error(), "panic: expected"; got != want {
		t.Fatalf("unexpected error -want/+got:\n\t- %s\n\t+ %s", want, got)
	}
}

func TestDispatcher_MultipleStops(t *testing.T) {
	d := newPoolDispatcher(10, zaptest.NewLogger(t))
	d.Start(1, context.Background())

	// Stopping repeatedly should not deadlock.
	for i := 0; i < 10; i++ {
		_ = d.Stop()
	}
}
