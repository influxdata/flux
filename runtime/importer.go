package runtime

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var (
	// list of packages included in the prelude.
	// Packages must be listed in import order
	prelude = []string{
		"universe",
		"influxdata/influxdb",
	}
)

type importer struct {
	r    *runtime
	pkgs map[string]*interpreter.Package
}

func (imp *importer) Import(path string) (semantic.MonoType, error) {
	p, err := imp.ImportPackageObject(path)
	if err != nil {
		return semantic.MonoType{}, err
	}
	return p.Type(), nil
}

func (imp *importer) ImportPackageObject(path string) (*interpreter.Package, error) {
	// If this package has been imported previously, return the import now.
	if p, ok := imp.pkgs[path]; ok {
		if p == nil {
			return nil, errors.Newf(codes.Invalid, "detected cyclical import for package path %q", path)
		}
		return p, nil
	}

	// Mark down that we are currently evaluating this package
	// so that we can detect a circular import.
	if imp.pkgs == nil {
		imp.pkgs = make(map[string]*interpreter.Package)
	}
	imp.pkgs[path] = nil

	// If this package is part of the prelude, fill in a fake
	// empty package to resolve cyclical imports.
	for _, ppath := range prelude {
		if ppath == path {
			imp.pkgs[path] = interpreter.NewPackage(path)
			break
		}
	}

	// Find the package for the given import path.
	semPkg, ok := imp.r.pkgs[path]
	if !ok {
		return nil, errors.Newf(codes.Invalid, "invalid import path %s", path)
	}

	// Construct the prelude scope from the prelude paths.
	// If we are importing part of the prelude, we do not
	// include it as part of the prelude and will stop
	// including values as soon as we hit the prelude.
	// This allows us to import all previous paths when loading
	// the prelude, but avoid a circular import.
	scope, err := imp.r.newScopeFor(path, imp)
	if err != nil {
		return nil, err
	}

	// Run the interpreter on the package to construct the values
	// created by the package. Pass in the previously initialized
	// packages as importable packages as we evaluate these in order.
	itrp := interpreter.NewInterpreter(nil)
	if _, err := itrp.Eval(context.Background(), semPkg, scope, imp); err != nil {
		return nil, err
	}
	obj := newObjectFromScope(scope)
	imp.pkgs[path] = interpreter.NewPackageWithValues(itrp.PackageName(), obj)
	return imp.pkgs[path], nil
}

func newObjectFromScope(scope values.Scope) values.Object {
	obj, _ := values.BuildObject(func(set values.ObjectSetter) error {
		scope.LocalRange(func(k string, v values.Value) {
			// Packages should not expose the packages they import.
			if _, ok := v.(values.Package); ok {
				return
			}
			set(k, v)
		})
		return nil
	})
	return obj
}
