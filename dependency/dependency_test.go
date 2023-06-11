package dependency_test

import (
	"context"
	"testing"

	"github.com/InfluxCommunity/flux/dependency"
)

type Dependency struct {
	InjectFn func(ctx context.Context) context.Context
}

func (d *Dependency) Inject(ctx context.Context) context.Context {
	return d.InjectFn(ctx)
}

type CloseFunc func() error

func (f CloseFunc) Close() error {
	return f()
}

func TestOnFinish(t *testing.T) {
	closed := false
	_, span := dependency.Inject(context.Background(), &Dependency{
		InjectFn: func(ctx context.Context) context.Context {
			dependency.OnFinish(ctx, CloseFunc(func() error {
				closed = true
				return nil
			}))
			return ctx
		},
	})

	if closed {
		t.Fatal("finish hook should not have been executed, but was")
	}
	span.Finish()
	if !closed {
		t.Fatal("finish hook did not execute")
	}
}
