package control

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/internal/pkg/syncutil"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/mock"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var mockCompiler *mock.Compiler

func init() {
	mockCompiler = new(mock.Compiler)
	mockCompiler.CompileFn = func(ctx context.Context) (*flux.Spec, error) {
		return flux.Compile(ctx, `from(bucket: "telegraf") |> range(start: -5m) |> mean()`, time.Now())
	}
}

func TestController_CompileQuery_Failure(t *testing.T) {
	compiler := &mock.Compiler{
		CompileFn: func(ctx context.Context) (*flux.Spec, error) {
			return nil, errors.New("expected")
		},
	}

	ctrl := New(Config{})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer func() {
		if err := ctrl.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
		cancel()
	}()

	// Run the query. It should return an error.
	if _, err := ctrl.Query(context.Background(), compiler); err == nil {
		t.Fatal("expected error")
	}

	// Verify the metrics say there are no queries.
	gauge, err := ctrl.metrics.all.GetMetricWithLabelValues()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	metric := &dto.Metric{}
	if err := gauge.Write(metric); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got, exp := int(metric.Gauge.GetValue()), 0; got != exp {
		t.Fatalf("unexpected metric value: exp=%d got=%d", exp, got)
	}
}

func TestController_PlanQuery_Failure(t *testing.T) {
	// Register a rule that destroys the integrity of the plan returned by the mock compiler.
	// The query should fail.
	config := Config{
		PPlannerOptions: []plan.PhysicalOption{
			plan.OnlyPhysicalRules(plantest.CreateCycleRule{Kind: universe.RangeKind}),
		},
	}
	ctrl := New(config)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer func() {
		if err := ctrl.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
		cancel()
	}()

	// Run the query. It should return an error.
	if _, err := ctrl.Query(context.Background(), mockCompiler); err == nil {
		t.Fatal("expected error")
	}

	// Verify the metrics say there are no queries.
	gauge, err := ctrl.metrics.all.GetMetricWithLabelValues()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	metric := &dto.Metric{}
	if err := gauge.Write(metric); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got, exp := int(metric.Gauge.GetValue()), 0; got != exp {
		t.Fatalf("unexpected metric value: exp=%d got=%d", exp, got)
	}
}

func TestController_EnqueueQuery_Failure(t *testing.T) {
	compiler := &mock.Compiler{
		CompileFn: func(ctx context.Context) (*flux.Spec, error) {
			// This returns an invalid spec so that enqueueing the query fails.
			// TODO(jsternberg): We should probably move the validation step to compilation
			// instead as it makes more sense. In that case, we would still need to verify
			// that enqueueing the query was successful in some way.
			return &flux.Spec{}, nil
		},
	}

	ctrl := New(Config{})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer func() {
		if err := ctrl.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
		cancel()
	}()

	// Run the query. It should return an error.
	if _, err := ctrl.Query(context.Background(), compiler); err == nil {
		t.Fatal("expected error")
	}

	// Verify the metrics say there are no queries.
	for name, gaugeVec := range map[string]*prometheus.GaugeVec{
		"all":      ctrl.metrics.all,
		"queueing": ctrl.metrics.queueing,
	} {
		gauge, err := gaugeVec.GetMetricWithLabelValues()
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		metric := &dto.Metric{}
		if err := gauge.Write(metric); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if got, exp := int(metric.Gauge.GetValue()), 0; got != exp {
			t.Fatalf("unexpected %s metric value: exp=%d got=%d", name, exp, got)
		}
	}
}

func TestController_ExecuteQuery_Failure(t *testing.T) {
	executor := mock.NewExecutor()
	executor.ExecuteFn = func(context.Context, *plan.PlanSpec, *memory.Allocator) (map[string]flux.Result, <-chan flux.Metadata, error) {
		return nil, mock.NoMetadata, errors.New("expected")
	}

	ctrl := New(Config{})
	ctrl.executor = executor

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer func() {
		if err := ctrl.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
		cancel()
	}()

	// Run a query and then wait for it to be ready.
	q, err := ctrl.Query(context.Background(), mockCompiler)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// We do not care about the results, just that the query is ready.
	<-q.Ready()

	if err := q.Err(); err == nil {
		t.Fatal("expected error")
	} else if got, want := err.Error(), "failed to execute query: expected"; got != want {
		t.Fatalf("unexpected error: exp=%s want=%s", want, got)
	}

	// Now finish the query by using Done.
	q.Done()

	// Verify the metrics say there are no queries.
	gauge, err := ctrl.metrics.all.GetMetricWithLabelValues()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	metric := &dto.Metric{}
	if err := gauge.Write(metric); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got, exp := int(metric.Gauge.GetValue()), 0; got != exp {
		t.Fatalf("unexpected metric value: exp=%d got=%d", exp, got)
	}
}

func TestController_CancelQuery_Ready(t *testing.T) {
	executor := mock.NewExecutor()
	executor.ExecuteFn = func(context.Context, *plan.PlanSpec, *memory.Allocator) (map[string]flux.Result, <-chan flux.Metadata, error) {
		// Return an empty result.
		return map[string]flux.Result{}, mock.NoMetadata, nil
	}

	ctrl := New(Config{})
	ctrl.executor = executor

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer func() {
		if err := ctrl.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
		cancel()
	}()

	// Run a query and then wait for it to be ready.
	q, err := ctrl.Query(context.Background(), mockCompiler)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// We do not care about the results, just that the query is ready.
	<-q.Ready()

	// Cancel the query. This is after the executor has already run,
	// but before we finalize the query. This ensures that canceling
	// at this stage will report the canceled state.
	q.Cancel()

	if want, got := Canceled, q.(*Query).State(); want != got {
		t.Errorf("unexpected state: want=%s got=%s", want, got)
	}

	// Now finish the query by using Done.
	q.Done()

	// Verify the metrics say there are no queries.
	gauge, err := ctrl.metrics.all.GetMetricWithLabelValues()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	metric := &dto.Metric{}
	if err := gauge.Write(metric); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got, exp := int(metric.Gauge.GetValue()), 0; got != exp {
		t.Fatalf("unexpected metric value: exp=%d got=%d", exp, got)
	}
}

func TestController_CancelQuery_Execute(t *testing.T) {
	executing := make(chan struct{})
	defer close(executing)

	executor := mock.NewExecutor()
	executor.ExecuteFn = func(ctx context.Context, spec *plan.PlanSpec, a *memory.Allocator) (map[string]flux.Result, <-chan flux.Metadata, error) {
		executing <- struct{}{}
		<-ctx.Done()
		return nil, mock.NoMetadata, ctx.Err()
	}

	ctrl := New(Config{})
	ctrl.executor = executor

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := ctrl.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
		cancel()
	}()

	// Run a query and then wait for it to be ready.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	q, err := ctrl.Query(ctx, mockCompiler)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	<-executing
	q.Cancel()

	// We do not care about the results, just that the query is ready.
	// We should not receive any results as the cancellation should
	// have signaled to the executor to cancel the query.
	select {
	case <-q.Ready():
		// The execute function should have received the cancel signal and exited
		// with an error.
	case <-ctx.Done():
		t.Error("timeout while waiting for the query to be canceled")
	}
	cancel()

	// The state should be canceled.
	if want, got := Canceled, q.(*Query).State(); want != got {
		t.Errorf("unexpected state: want=%s got=%s", want, got)
	}

	// Now finish the query by using Done.
	q.Done()

	// The query should have been canceled and not a timeout.
	if want, got := context.Canceled, q.Err(); want != got {
		t.Errorf("unexpected error: want=%s got=%s", want, got)
	}

	// Verify the metrics say there are no queries.
	gauge, err := ctrl.metrics.all.GetMetricWithLabelValues()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	metric := &dto.Metric{}
	if err := gauge.Write(metric); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got, exp := int(metric.Gauge.GetValue()), 0; got != exp {
		t.Fatalf("unexpected metric value: exp=%d got=%d", exp, got)
	}
}

// Start queries and then immediately cancel them to try and trigger
// a race condition while testing under the race detector.
func TestController_CancelQuery_Concurrent(t *testing.T) {
	executor := mock.NewExecutor()
	executor.ExecuteFn = func(ctx context.Context, spec *plan.PlanSpec, a *memory.Allocator) (map[string]flux.Result, <-chan flux.Metadata, error) {
		return map[string]flux.Result{}, mock.NoMetadata, nil
	}

	ctrl := New(Config{})
	ctrl.executor = executor

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := ctrl.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
		cancel()
	}()

	// Run a bunch of queries and cancel them. We don't really know
	// when they will be canceled, so we are mostly testing that
	// canceling at any point during the query does not cause a problem
	// like a data race.
	queries := make(chan flux.Query, 100)
	var creatorWaitGroup syncutil.WaitGroup
	for i := 0; i < 5; i++ {
		creatorWaitGroup.Do(func() error {
			for j := 0; j < 100; j++ {
				q, err := ctrl.Query(context.Background(), mockCompiler)
				if err != nil {
					return err
				}
				queries <- q
			}
			return nil
		})
	}

	var (
		cancelWaitGroup syncutil.WaitGroup
		doneWaitGroup   syncutil.WaitGroup
	)
	for i := 0; i < 5; i++ {
		cancelWaitGroup.Do(func() error {
			for q := range queries {
				// Cancel the query. This may or may not cancel
				// it as the results may have already been reported,
				// but we will attempt this anyway.
				q.Cancel()

				// Assign a variable so the closure
				// captures the current query rather than
				// the one from the for loop that will
				// change.
				query := q
				doneWaitGroup.Do(func() error {
					<-query.Ready()
					query.Done()
					return nil
				})
			}
			return nil
		})
	}

	if err := creatorWaitGroup.Wait(); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	close(queries)

	// All of the cancellations should finish.
	if err := cancelWaitGroup.Wait(); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// Wait for all of the queries to finish.
	if err := doneWaitGroup.Wait(); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestController_BlockedExecutor(t *testing.T) {
	done := make(chan struct{})

	executor := mock.NewExecutor()
	executor.ExecuteFn = func(context.Context, *plan.PlanSpec, *memory.Allocator) (map[string]flux.Result, <-chan flux.Metadata, error) {
		<-done
		return nil, mock.NoMetadata, nil
	}

	ctrl := New(Config{})
	ctrl.executor = executor

	cctx, ccancel := context.WithTimeout(context.Background(), time.Second)
	defer func() {
		if err := ctrl.Shutdown(cctx); err != nil {
			t.Fatal(err)
		}
		ccancel()
	}()

	// Run a query that will cause the controller to stall.
	q, err := ctrl.Query(context.Background(), mockCompiler)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer func() {
		close(done)
		<-q.Ready()
		q.Done()
	}()

	// Run another query. It should block in the Query call and then unblock when we cancel
	// the context.
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case <-timer.C:
		case <-done:
			timer.Stop()
		}
	}()

	if _, err := ctrl.Query(ctx, mockCompiler); err == nil {
		t.Fatal("expected error")
	} else if got, want := err, context.Canceled; got != want {
		t.Fatalf("unexpected error: got=%q want=%q", got, want)
	}
}

func TestController_CancelledContextPropagatesToExecutor(t *testing.T) {
	t.Parallel()

	executor := mock.NewExecutor()
	executor.ExecuteFn = func(ctx context.Context, _ *plan.PlanSpec, _ *memory.Allocator) (map[string]flux.Result, <-chan flux.Metadata, error) {
		<-ctx.Done() // Unblock only when context has been cancelled
		return nil, mock.NoMetadata, nil
	}

	ctrl := New(Config{})
	ctrl.executor = executor

	cctx, ccancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer func() {
		if err := ctrl.Shutdown(cctx); err != nil {
			t.Fatal(err)
		}
		ccancel()
	}()

	// Parent query context
	pctx, pcancel := context.WithCancel(context.Background())

	// done signals that ExecuteFn returned
	done := make(chan struct{})

	go func() {
		q, err := ctrl.Query(pctx, mockCompiler)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		// Ready will unblock when executor unblocks
		<-q.Ready()
		// TODO(jlapacik): query should expose error if cancelled during execution
		// if q.Err() == nil {
		//     t.Errorf("expected error; cancelled query context before execution finished")
		// }
		q.Done()
		close(done)
	}()

	waitCheckDelay := 500 * time.Millisecond

	select {
	case <-done:
		t.Fatalf("ExecuteFn returned before parent context was cancelled")
	case <-time.After(waitCheckDelay):
		// Okay.
	}

	pcancel()

	select {
	case <-done:
		// Okay.
	case <-time.After(waitCheckDelay):
		t.Fatalf("ExecuteFn didn't return after parent context canceled")
	}
}

func TestController_Shutdown(t *testing.T) {
	// Create a wait group that finishes when it attempts to execute.
	// This is used to ensure that it is in the list of queries.
	var executeGroup sync.WaitGroup
	executeGroup.Add(10)
	executor := mock.NewExecutor()
	executor.ExecuteFn = func(ctx context.Context, p *plan.PlanSpec, a *memory.Allocator) (results map[string]flux.Result, metaCh <-chan flux.Metadata, e error) {
		executeGroup.Done()
		return nil, mock.NoMetadata, nil
	}
	ctrl := New(Config{})
	ctrl.executor = executor

	// Create a bunch of queries and never call Ready which should leave them in the controller.
	queries := make([]flux.Query, 0, 15)
	for i := 0; i < 10; i++ {
		q, err := ctrl.Query(context.Background(), mockCompiler)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
			continue
		}
		queries = append(queries, q)
	}

	if len(queries) != 10 {
		// Exit now since not all of the queries executed.
		return
	}

	// Run shutdown which should wait until the queries are finished.
	var wg syncutil.WaitGroup
	wg.Do(func() error {
		return ctrl.Shutdown(context.Background())
	})

	// Attempt to create new queries until one is rejected.
	// An initial query may not be rejected because the controller
	// has not yet started to shutdown.
	rejected := false
	for i := 0; i < 5; i++ {
		if q, err := ctrl.Query(context.Background(), mockCompiler); err != nil {
			rejected = true
			break
		} else {
			// It was not rejected, so add it to our expected queries.
			queries = append(queries, q)
			executeGroup.Add(1)
		}

		// Wait for 200 microseconds to allow the controller to shutdown.
		<-time.After(200 * time.Microsecond)
	}

	// A new query should be rejected.
	if !rejected {
		t.Error("expected a query to be rejected after controller shutdown")
	}

	// Ensure that all of the started queries have been executed.
	executeGroup.Wait()

	// There should be at least 10 active queries.
	if want, got := len(queries), len(ctrl.Queries()); want != got {
		t.Errorf("unexpected query count -want/+got\n\t- %d\n\t+ %d", want, got)
	}

	// Mark each of the queries as done.
	for _, q := range queries {
		q := q
		wg.Do(func() error {
			<-q.Ready()
			q.Done()
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// There should be no queries.
	if want, got := 0, len(ctrl.Queries()); want != got {
		t.Fatalf("unexpected query count -want/+got\n\t- %d\n\t+ %d", want, got)
	}
}

func TestController_Statistics(t *testing.T) {
	executor := mock.NewExecutor()
	executor.ExecuteFn = func(ctx context.Context, p *plan.PlanSpec, a *memory.Allocator) (results map[string]flux.Result, metadata <-chan flux.Metadata, e error) {
		// Create a metadata channel that we will use to simulate sending metadata
		// from the executor.
		metaCh := make(chan flux.Metadata, 2)
		go func() {
			defer close(metaCh)
			metaCh <- flux.Metadata{
				"influxdb/scanned-values": []interface{}{int64(60)},
				"influxdb/scanned-bytes":  []interface{}{int64(60 * 8)},
			}
			metaCh <- flux.Metadata{
				"influxdb/scanned-values": []interface{}{int64(34)},
				"influxdb/scanned-bytes":  []interface{}{int64(34 * 8)},
			}
		}()
		return nil, metaCh, nil
	}
	ctrl := New(Config{})
	ctrl.executor = executor

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer func() {
		if err := ctrl.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
		cancel()
	}()

	// Run the query. It should not return an error.
	q, err := ctrl.Query(context.Background(), mockCompiler)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	<-q.Ready()
	time.Sleep(time.Millisecond)
	q.Done()

	// Ensure this works without
	stats := q.Statistics()
	if stats.TotalDuration == 0 {
		t.Error("total duration should be greater than zero")
	}
	if want, got := (flux.Metadata{
		"influxdb/scanned-values": []interface{}{int64(60), int64(34)},
		"influxdb/scanned-bytes":  []interface{}{int64(60 * 8), int64(int64(34 * 8))},
	}), stats.Metadata; !cmp.Equal(want, got) {
		t.Errorf("unexpected metadata -want/+got\n%s", cmp.Diff(want, got))
	}
}
