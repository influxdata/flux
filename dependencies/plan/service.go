package plan

import "context"

type Service interface {
	Plan(context.Context, Spec) (Spec, error)
}

type Spec interface {
	CheckIntegrity() error
}
