package execute

import (
	"context"
	"sync"
	"testing"
	"time"

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

func TestDispatcher_ScheduleMany(t *testing.T) {
	// Continuously schedule jobs that schedule other jobs.
	// The schedule method should not block the dispatcher but
	// instead grow continously.
	d := newPoolDispatcher(10, zaptest.NewLogger(t))

	// This test should finish by the timeout.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	d.Start(1, ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		_ = d.Stop()
	}()

	// The default ring size is 100.
	// A size of 200 should trigger any deadlock from scheduling
	// too many times.
	for i := 0; i < 200; i++ {
		d.Schedule(func(ctx context.Context, throughput int) {
			for i := 0; i < throughput; i++ {
				// Attempt to schedule another job.
				// If the dispatcher doesn't expand correctly,
				// this should deadlock.
				d.Schedule(func(ctx context.Context, throughput int) {
					// Do nothing here.
				})
			}
		})
	}

	// Check for the context error.
	if err := ctx.Err(); err != nil {
		t.Errorf("timeout reached: %s", err)
	}
	cancel()
	wg.Wait()
}
