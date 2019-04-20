package control_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/control"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/mock"
)

var (
	mockCompiler = &mock.Compiler{
		CompileFn: func(ctx context.Context) (flux.Program, error) {
			return &mock.Program{
				ExecuteFn: func(ctx context.Context, q *mock.Query, alloc *memory.Allocator) {
					q.ResultsCh <- &executetest.Result{}
				},
			}, nil
		},
	}
	config = control.Config{
		ConcurrencyQuota:         1,
		MemoryBytesQuotaPerQuery: 1024,
		QueueSize:                1,
	}
)

func TestController_QuerySuccess(t *testing.T) {
	ctrl, err := control.New(config)
	if err != nil {
		t.Fatal(err)
	}
	defer shutdown(t, ctrl)

	q, err := ctrl.Query(context.Background(), mockCompiler)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	for range q.Results() {
		// discard the results as we do not care.
	}
	q.Done()

	if err := q.Err(); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	stats := q.Statistics()
	if stats.CompileDuration == 0 {
		t.Error("expected compile duration to be above zero")
	}
	if stats.QueueDuration == 0 {
		t.Error("expected queue duration to be above zero")
	}
	if stats.ExecuteDuration == 0 {
		t.Error("expected execute duration to be above zero")
	}
	if stats.TotalDuration == 0 {
		t.Error("expected total duration to be above zero")
	}
}

func TestController_AfterShutdown(t *testing.T) {
	ctrl, err := control.New(config)
	if err != nil {
		t.Fatal(err)
	}
	shutdown(t, ctrl)

	// No point in continuing. The shutdown didn't work
	// even though there are no queries.
	if t.Failed() {
		return
	}

	if _, err := ctrl.Query(context.Background(), mockCompiler); err == nil {
		t.Error("expected error")
	} else if got, want := err.Error(), "query controller shutdown"; got != want {
		t.Errorf("unexpected error -want/+got\n\t- %q\n\t+ %q", want, got)
	}
}

func TestController_CompileError(t *testing.T) {
	ctrl, err := control.New(config)
	if err != nil {
		t.Fatal(err)
	}
	defer shutdown(t, ctrl)

	compiler := &mock.Compiler{
		CompileFn: func(ctx context.Context) (flux.Program, error) {
			return nil, errors.New("expected error")
		},
	}
	if _, err := ctrl.Query(context.Background(), compiler); err == nil {
		t.Error("expected error")
	} else if got, want := err.Error(), "compilation failed: expected error"; got != want {
		t.Errorf("unexpected error -want/+got\n\t- %q\n\t+ %q", want, got)
	}
}

func TestController_ExecuteError(t *testing.T) {
	ctrl, err := control.New(config)
	if err != nil {
		t.Fatal(err)
	}
	defer shutdown(t, ctrl)

	compiler := &mock.Compiler{
		CompileFn: func(ctx context.Context) (flux.Program, error) {
			return &mock.Program{
				StartFn: func(ctx context.Context, alloc *memory.Allocator) (*mock.Query, error) {
					return nil, errors.New("expected error")
				},
			}, nil
		},
	}

	q, err := ctrl.Query(context.Background(), compiler)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// There should be no results.
	numResults := 0
	for range q.Results() {
		numResults++
	}

	if numResults != 0 {
		t.Errorf("no results should have been returned, but %d were", numResults)
	}
	q.Done()

	if err := q.Err(); err == nil {
		t.Error("expected error")
	} else if got, want := err.Error(), "expected error"; got != want {
		t.Errorf("unexpected error -want/+got\n\t- %q\n\t+ %q", want, got)
	}
}

func TestController_ShutdownWithRunningQuery(t *testing.T) {
	ctrl, err := control.New(config)
	if err != nil {
		t.Fatal(err)
	}
	defer shutdown(t, ctrl)

	executing := make(chan struct{})
	compiler := &mock.Compiler{
		CompileFn: func(ctx context.Context) (flux.Program, error) {
			return &mock.Program{
				ExecuteFn: func(ctx context.Context, q *mock.Query, alloc *memory.Allocator) {
					close(executing)
					<-ctx.Done()

					// This should still be read even if we have been canceled.
					q.ResultsCh <- &executetest.Result{}
				},
			}, nil
		},
	}

	q, err := ctrl.Query(context.Background(), compiler)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range q.Results() {
			// discard the results
		}
		q.Done()
	}()

	// Wait until execution has started.
	<-executing

	// Shutdown should succeed and not timeout. The above blocked
	// query should be canceled and then shutdown should return.
	shutdown(t, ctrl)
	wg.Wait()
}

func TestController_ShutdownWithTimeout(t *testing.T) {
	ctrl, err := control.New(config)
	if err != nil {
		t.Fatal(err)
	}
	defer shutdown(t, ctrl)

	// This channel blocks program execution until we are done
	// with running the test.
	done := make(chan struct{})
	defer close(done)

	executing := make(chan struct{})
	compiler := &mock.Compiler{
		CompileFn: func(ctx context.Context) (flux.Program, error) {
			return &mock.Program{
				ExecuteFn: func(ctx context.Context, q *mock.Query, alloc *memory.Allocator) {
					// This should just block until the end of the test
					// when we perform cleanup.
					close(executing)
					<-done
				},
			}, nil
		},
	}

	q, err := ctrl.Query(context.Background(), compiler)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	go func() {
		for range q.Results() {
			// discard the results
		}
		q.Done()
	}()

	// Wait until execution has started.
	<-executing

	// The shutdown should not succeed.
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	if err := ctrl.Shutdown(ctx); err == nil {
		t.Error("expected error")
	} else if got, want := err.Error(), context.DeadlineExceeded.Error(); got != want {
		t.Errorf("unexpected error -want/+got\n\t- %q\n\t+ %q", want, got)
	}
	cancel()
}

func TestController_PerQueryMemoryLimit(t *testing.T) {
	ctrl, err := control.New(config)
	if err != nil {
		t.Fatal(err)
	}
	defer shutdown(t, ctrl)

	compiler := &mock.Compiler{
		CompileFn: func(ctx context.Context) (flux.Program, error) {
			return &mock.Program{
				ExecuteFn: func(ctx context.Context, q *mock.Query, alloc *memory.Allocator) {
					// This is emulating the behavior of exceeding the memory limit at runtime
					if err := alloc.Allocate(int(config.MemoryBytesQuotaPerQuery + 1)); err != nil {
						q.SetErr(err)
					}
				},
			}, nil
		},
	}

	q, err := ctrl.Query(context.Background(), compiler)
	if err != nil {
		t.Fatal(err)
	}

	for range q.Results() {
		// discard the results
	}
	q.Done()

	if q.Err() == nil {
		t.Fatal("expected error about memory limit exceeded")
	}
}

func TestController_ConcurrencyQuota(t *testing.T) {
	const (
		numQueries       = 3
		concurrencyQuota = 2
	)

	config := config
	config.ConcurrencyQuota = concurrencyQuota
	config.QueueSize = numQueries
	ctrl, err := control.New(config)
	if err != nil {
		t.Fatal(err)
	}
	defer shutdown(t, ctrl)

	executing := make(chan struct{}, numQueries)
	compiler := &mock.Compiler{
		CompileFn: func(ctx context.Context) (flux.Program, error) {
			return &mock.Program{
				ExecuteFn: func(ctx context.Context, q *mock.Query, alloc *memory.Allocator) {
					select {
					case <-q.Canceled:
					default:
						executing <- struct{}{}
						<-q.Canceled
					}
				},
			}, nil
		},
	}

	for i := 0; i < numQueries; i++ {
		q, err := ctrl.Query(context.Background(), compiler)
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			for range q.Results() {
				// discard the results
			}
			q.Done()
		}()
	}

	// Give 2 queries a chance to begin executing.  The remaining third query should stay queued.
	time.Sleep(250 * time.Millisecond)

	if err := ctrl.Shutdown(context.Background()); err != nil {
		t.Error(err)
	}

	// There is a chance that the remaining query managed to get executed after the executing queries
	// were canceled.  As a result, this test is somewhat flaky.

	close(executing)

	var count int
	for range executing {
		count++
	}

	if count != concurrencyQuota {
		t.Fatalf("expected exactly %v queries to execute, but got: %v", concurrencyQuota, count)
	}
}

func TestController_QueueSize(t *testing.T) {
	const (
		concurrencyQuota = 2
		queueSize        = 3
	)

	config := config
	config.ConcurrencyQuota = concurrencyQuota
	config.QueueSize = queueSize
	ctrl, err := control.New(config)
	if err != nil {
		t.Fatal(err)
	}
	defer shutdown(t, ctrl)

	// This channel blocks program execution until we are done
	// with running the test.
	done := make(chan struct{})
	defer close(done)

	executing := make(chan struct{}, config.ConcurrencyQuota)
	compiler := &mock.Compiler{
		CompileFn: func(ctx context.Context) (flux.Program, error) {
			return &mock.Program{
				ExecuteFn: func(ctx context.Context, q *mock.Query, alloc *memory.Allocator) {
					executing <- struct{}{}
					// Block until test is finished
					<-done
				},
			}, nil
		},
	}

	// Start as many queries as can be running at the same time
	for i := 0; i < concurrencyQuota; i++ {
		q, err := ctrl.Query(context.Background(), compiler)
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			for range q.Results() {
				// discard the results
			}
			q.Done()
		}()

		// Wait until it's executing
		<-executing
	}

	// Now fill up the queue
	for i := 0; i < queueSize; i++ {
		q, err := ctrl.Query(context.Background(), compiler)
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			for range q.Results() {
				// discard the results
			}
			q.Done()
		}()
	}

	_, err = ctrl.Query(context.Background(), compiler)
	if err == nil {
		t.Fatal("expected an error about queue length exceeded")
	}
}

func shutdown(t *testing.T, ctrl *control.Controller) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := ctrl.Shutdown(ctx); err != nil {
		t.Error(err)
	}
}
