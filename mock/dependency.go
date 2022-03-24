package mock

import "context"

type Dependency struct {
	InjectFn func(ctx context.Context) context.Context
}

func (d Dependency) Inject(ctx context.Context) context.Context {
	return d.InjectFn(ctx)
}
