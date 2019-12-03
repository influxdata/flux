package tableutil

import (
	"context"
	"sync/atomic"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

// SendFunc is used to send a flux.ColReader to a table stream so
// it can be read by the table consumer.
type SendFunc func(flux.ColReader)

// Stream will call StreamWithContext with a background context.
func Stream(f func(ctx context.Context, fn SendFunc) error) (flux.Table, error) {
	return StreamWithContext(context.Background(), f)
}

// StreamWithContext will create a table that streams column readers
// through the flux.Table. This method will return only after
// the function buffers the first column reader.
// This first column reader is used to identify the group key
// and columns for the entire table stream.
//
// Implementors using this *must* return at least one table.
// If the function returns without returning at least one table,
// then an error will be returned. If the first table that is returned
// is empty, then this will return an empty table and further buffers
// will not be used.
func StreamWithContext(ctx context.Context, f func(ctx context.Context, fn SendFunc) error) (flux.Table, error) {
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan streamBuffer)

	// Create the send method.
	send := func(cr flux.ColReader) {
		select {
		case ch <- streamBuffer{cr: cr}:
		case <-ctx.Done():
			// We could not send the column reader because this was cancelled.
			cr.Release()
		}
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer close(ch)
		if err := f(ctx, send); err != nil {
			ch <- streamBuffer{err: err}
		}
	}()

	select {
	case sp := <-ch:
		cr, err := sp.cr, sp.err
		if cr == nil {
			if err == nil {
				err = errors.New(codes.Internal, "empty table stream")
			}
			cancel()
			return nil, err
		}

		// Retrieve the group key and columns from the column reader.
		key, cols := cr.Key(), cr.Cols()

		// If the table is empty, signal to the context
		// that we are not expecting more tables just
		// in case the implementor does something wrong.
		// We also release the column reader since we don't need
		// it anymore and set it to nil.
		empty := cr.Len() == 0
		if empty {
			cancel()
			cr.Release()
			cr = nil
		}
		return &streamTable{
			first:  cr,
			key:    key,
			cols:   cols,
			cancel: cancel,
			ch:     ch,
			done:   done,
			empty:  empty,
		}, nil
	case <-ctx.Done():
		cancel()
		return nil, ctx.Err()
	}
}

// streamBuffer is a column reader or error sent
// from the streaming function.
type streamBuffer struct {
	cr  flux.ColReader
	err error
}

// streamTable is an implementation of flux.Table
// that will stream buffers from a column reader.
type streamTable struct {
	used   int32
	first  flux.ColReader
	key    flux.GroupKey
	cols   []flux.ColMeta
	cancel func()
	ch     <-chan streamBuffer
	done   <-chan struct{}
	empty  bool
}

func (s *streamTable) Key() flux.GroupKey {
	return s.key
}

func (s *streamTable) Cols() []flux.ColMeta {
	return s.cols
}

func (s *streamTable) Do(f func(flux.ColReader) error) error {
	if !atomic.CompareAndSwapInt32(&s.used, 0, 1) {
		return errors.New(codes.Internal, "table already read")
	}

	// Ensure that we always call cancel to free any resources from
	// the context after we have completely read the channel.
	defer s.cancel()

	// If the table is empty, return immediately.
	// We already released the column reader.
	if s.empty {
		return nil
	}

	// Act on the first column reader that was read.
	if err := f(s.first); err != nil {
		s.first.Release()
		s.first = nil
		return nil
	}
	s.first.Release()
	s.first = nil

	for sp := range s.ch {
		cr, err := sp.cr, sp.err
		if err != nil {
			return err
		}
		if err := f(cr); err != nil {
			cr.Release()
			return err
		}
		cr.Release()
	}
	// Allow the stream function to exit.
	<-s.done
	return nil
}

func (s *streamTable) Done() {
	if atomic.CompareAndSwapInt32(&s.used, 0, 1) {
		if s.first != nil {
			s.first.Release()
			s.first = nil
		}
		s.cancel()
	}
	// Wait for the stream function to exit before we return.
	<-s.done
}

func (s *streamTable) Empty() bool {
	return s.empty
}

// IsDone is used to allow the tests to access internal parts
// of the table structure for the table tests.
// This method can only be used by asserting that it exists
// through an anonymous interface. This should not be used
// outside of testing code because there is no guarantee
// on the safety of this method.
func (s *streamTable) IsDone() bool {
	return s.empty || atomic.LoadInt32(&s.used) != 0
}
