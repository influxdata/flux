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

var mockCompiler *mock.Compiler

func init() {
	mockCompiler = new(mock.Compiler)
	mockCompiler.CompileFn = func(ctx context.Context) (flux.Program, error) {
		return &mock.Program{
			ExecuteFn: func(ctx context.Context, q *mock.Query, alloc *memory.Allocator) {
				q.ResultsCh <- &executetest.Result{}
			},
		}, nil
	}
}

func TestController_QuerySuccess(t *testing.T) {
	ctrl := control.New(control.Config{})
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
	ctrl := control.New(control.Config{})
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
	ctrl := control.New(control.Config{})
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
	ctrl := control.New(control.Config{})
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
	ctrl := control.New(control.Config{})
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
	ctrl := control.New(control.Config{})
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

func shutdown(t *testing.T, ctrl *control.Controller) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := ctrl.Shutdown(ctx); err != nil {
		t.Error(err)
	}
}
