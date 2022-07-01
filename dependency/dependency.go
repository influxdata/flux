package dependency

import (
	"context"
	"io"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

// Interface is an interface that must be implemented by every injectable dependency.
// On Inject, the dependency is injected into the context and the resulting one is returned.
// Every dependency must provide a function to extract it from the context.
type Interface interface {
	Inject(ctx context.Context) context.Context
}

type List []Interface

func (l List) Inject(ctx context.Context) context.Context {
	for _, dep := range l {
		ctx = dep.Inject(ctx)
	}
	return ctx
}

func Inject(ctx context.Context, deps ...Interface) (context.Context, *Span) {
	span := &Span{}
	ctx = context.WithValue(ctx, spanKey, span)
	for _, dep := range deps {
		ctx = dep.Inject(ctx)
	}
	return ctx, span
}

func OnFinish(ctx context.Context, c io.Closer) {
	span := spanFromContext(ctx)
	span.onFinish(c)
}

type closeFunc func() error

func (fn closeFunc) Close() error {
	return fn()
}

func OnFinishFunc(ctx context.Context, fn func() error) {
	OnFinish(ctx, closeFunc(fn))
}

type contextKey int

const (
	spanKey contextKey = iota
)

type Span struct {
	closers []io.Closer
}

func spanFromContext(ctx context.Context) *Span {
	span := ctx.Value(spanKey)
	if span == nil {
		panic(errors.Newf(codes.Internal, "dependency injection requires a span but one does not exist"))
	}
	return span.(*Span)
}

func (s *Span) onFinish(closer io.Closer) {
	s.closers = append(s.closers, closer)
}

func (s *Span) Finish() {
	for _, closer := range s.closers {
		_ = closer.Close()
	}
	s.closers = nil
}
