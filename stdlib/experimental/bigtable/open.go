package bigtable

import (
	"context"
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	project  = "project"
	instance = "instance"
)

var open = values.NewFunction(
	"open",
	semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{project: semantic.String, instance: semantic.String},
		Required:   semantic.LabelSet{project, instance},
		Return:     semantic.Object,
	}),
	func(ctx context.Context, deps dependencies.Interface, args values.Object) (values.Value, error) {
		p, ok := args.Get(project)
		if !ok {
			return nil, fmt.Errorf("missing argument %q", project)
		}

		i, ok := args.Get(instance)
		if !ok {
			return nil, fmt.Errorf("missing argument %q", instance)
		}

		if p.Type().Nature() == semantic.String && i.Type().Nature() == semantic.String {
			return args, nil
		}

		return nil, fmt.Errorf("cannot create connection")
	}, false,
)

func init() {
	flux.RegisterPackageValue("experimental/bigtable", "open", open)
}
