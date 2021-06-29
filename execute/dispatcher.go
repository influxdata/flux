package execute

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Dispatcher schedules work for a query.
// Each transformation submits work to be done to the dispatcher.
// Then the dispatcher schedules to work based on the available resources.
type Dispatcher interface {
	// Schedule fn to be executed
	Schedule(fn ScheduleFunc)
}

// ScheduleFunc is a function that represents work to do.
// The throughput is the maximum number of messages to process for this scheduling.
type ScheduleFunc func(ctx context.Context, throughput int)

// poolDispatcher implements Dispatcher using a pool of goroutines.
type poolDispatcher struct {
	work   *ring
	ready  chan struct{}
	workMu sync.Mutex

	throughput int

	mu      sync.Mutex
	closed  bool
	closing chan struct{}
	wg      sync.WaitGroup
	err     error
	errC    chan error

	logger *zap.Logger
}

func newPoolDispatcher(throughput int, logger *zap.Logger) *poolDispatcher {
	return &poolDispatcher{
		throughput: throughput,
		work:       newRing(100),
		ready:      make(chan struct{}, 1),
		closing:    make(chan struct{}),
		errC:       make(chan error, 1),
		logger:     logger.With(zap.String("component", "dispatcher")),
	}
}

func (d *poolDispatcher) Schedule(fn ScheduleFunc) {
	d.workMu.Lock()
	defer d.workMu.Unlock()

	// Schedule the work and then report to the channel that there
	// is available work to unblock the worker scheduler thread.
	d.work.Append(fn)
	select {
	case d.ready <- struct{}{}:
		// The ready channel should have a buffer of 1.
		// Work being present is a binary yes or no.
		// If we say yes multiple times, we only need to read it once
		// in the outermost run loop.
	default:
	}
}

func (d *poolDispatcher) Start(n int, ctx context.Context) {
	d.wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer d.wg.Done()
			// Setup panic handling on the worker goroutines
			defer func() {
				if e := recover(); e != nil {
					err, ok := e.(error)
					if !ok {
						err = fmt.Errorf("%v", e)
					}

					if errors.Code(err) == codes.ResourceExhausted {
						d.setErr(err)
						return
					}

					err = errors.Wrap(err, codes.Internal, "panic")
					d.setErr(err)
					if entry := d.logger.Check(zapcore.InfoLevel, "Dispatcher panic"); entry != nil {
						entry.Stack = string(debug.Stack())
						entry.Write(zap.Error(err))
					}
				}
			}()
			d.run(ctx)
		}()
	}
}

// Err returns a channel with will produce an error if encountered.
func (d *poolDispatcher) Err() <-chan error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.errC
}

func (d *poolDispatcher) setErr(err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// TODO(nathanielc): Collect all error information.
	if d.err == nil {
		d.err = err
		d.errC <- err
	}
}

// Stop the dispatcher.
func (d *poolDispatcher) Stop() error {
	// Check if this is the first time invoking this method.
	d.mu.Lock()
	if !d.closed {
		// If not, mark the dispatcher as closed and signal to the current
		// workers that they should stop processing more work.
		d.closed = true
		close(d.closing)
	}
	d.mu.Unlock()

	// Wait for the existing workers to finish.
	d.wg.Wait()

	// Grab the error from within a lock.
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.err
}

// run is the logic executed by each worker goroutine in the pool.
func (d *poolDispatcher) run(ctx context.Context) {
	for {
		// This loop waits for any work to be present in the queue
		// or for the dispatcher to be closed or the context canceled.
		select {
		case <-ctx.Done():
			// Immediately return, do not process any more work
			return
		case <-d.closing:
			// We are done, nothing left to do.
			return
		case <-d.ready:
			// Work is in the queue. Continue to pull work
			// from the queue until there is none left or
			// we are supposed to stop for one of the other
			// reasons stated above.
			d.doWork(ctx)
		}
	}
}

// doWork will continue pulling work from the work queue
// and running the scheduled functions until the context is canceled,
// the dispatcher is closed, or there is no more work in the queue.
func (d *poolDispatcher) doWork(ctx context.Context) {
	for {
		var fn ScheduleFunc
		d.workMu.Lock()
		if next := d.work.Next(); next != nil {
			fn = next.(ScheduleFunc)
		}
		d.workMu.Unlock()

		if fn == nil {
			// No work anymore. Return to the top level loop
			// which will wait until new work has been appended.
			return
		}
		fn(ctx, d.throughput)

		// Check to see if the context was canceled or
		// the dispatcher was closed. This allows us to exit
		// even if we have not pulled off all of the available work.
		select {
		case <-ctx.Done():
			return
		case <-d.closing:
			return
		default:
		}
	}
}
