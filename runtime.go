package flux

import (
	"context"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// defaultRuntime contains the preregistered packages and builtin values
// required to execute a flux script.
var defaultRuntime = &runtime{}

// runtimeBuilder is used to construct the runtime before it is finalized.
type runtimeBuilder struct {
	pkgs     map[string]*ast.Package
	builtins map[string]map[string]values.Value
}

func (b *runtimeBuilder) RegisterPackage(pkg *ast.Package) error {
	if b == nil {
		return errors.New(codes.Internal, "already finalized, cannot register builtin package")
	}

	if b.pkgs == nil {
		b.pkgs = make(map[string]*ast.Package)
	}

	if _, ok := b.pkgs[pkg.Path]; ok {
		return errors.Newf(codes.Internal, "duplicate builtin package %q", pkg.Path)
	}
	b.pkgs[pkg.Path] = pkg
	return nil
}

func (b *runtimeBuilder) RegisterPackageValue(pkgpath, name string, value values.Value) error {
	return b.registerPackageValue(pkgpath, name, value, false)
}

func (b *runtimeBuilder) ReplacePackageValue(pkgpath, name string, value values.Value) error {
	return b.registerPackageValue(pkgpath, name, value, true)
}

func (b *runtimeBuilder) registerPackageValue(pkgpath, name string, value values.Value, replace bool) error {
	if b == nil {
		return errors.Newf(codes.Internal, "already finalized, cannot register builtin package value")
	}

	if b.builtins == nil {
		b.builtins = make(map[string]map[string]values.Value)
	}

	pkg, ok := b.builtins[pkgpath]
	if !ok {
		pkg = make(map[string]values.Value)
		b.builtins[pkgpath] = pkg
	}

	if _, ok := pkg[name]; ok && !replace {
		return errors.Newf(codes.Internal, "duplicate builtin package value %q %q", pkgpath, name)
	} else if !ok && replace {
		return errors.Newf(codes.Internal, "missing builtin package value %q %q", pkgpath, name)
	}
	pkg[name] = value
	return nil
}

// runtime contains the flux runtime for interpreting and
// executing queries.
type runtime struct {
	pkgs      map[string]*interpreter.Package
	prelude   *scopeSet
	rbuilder  *runtimeBuilder
	finalized bool
}

// builder returns the runtime builder for this runtime
// or constructs one if the runtime hasn't been finalized.
func (r *runtime) builder() *runtimeBuilder {
	if r.rbuilder == nil {
		if r.finalized {
			return nil
		}
		r.rbuilder = &runtimeBuilder{}
	}
	return r.rbuilder
}

func (r *runtime) RegisterPackage(pkg *ast.Package) error {
	return r.builder().RegisterPackage(pkg)
}

func (r *runtime) RegisterPackageValue(pkgpath, name string, value values.Value) error {
	return r.builder().RegisterPackageValue(pkgpath, name, value)
}

func (r *runtime) ReplacePackageValue(pkgpath, name string, value values.Value) error {
	return r.builder().ReplacePackageValue(pkgpath, name, value)
}

func (r *runtime) Prelude() values.Scope {
	if !r.finalized {
		panic("builtins not finalized")
	}
	return r.prelude.Nest(nil)
}

func (r *runtime) Stdlib() interpreter.Importer {
	importer := importer{pkgs: r.pkgs}
	return importer.Copy()
}

func (r *runtime) Finalize() error {
	if r.finalized {
		return errors.New(codes.Internal, "already finalized")
	}
	r.finalized = true

	b := r.builder()
	order, err := packageOrder(prelude, b.pkgs)
	if err != nil {
		return err
	}

	r.prelude = &scopeSet{
		packages: make([]*interpreter.Package, 0, len(prelude)),
	}
	r.pkgs = make(map[string]*interpreter.Package, len(order))
	for _, astPkg := range order {
		if ast.Check(astPkg) > 0 {
			err := ast.GetError(astPkg)
			return errors.Wrapf(err, codes.Inherit, "failed to parse builtin package %q", astPkg.Path)
		}
		pkgpath := astPkg.Path

		// Analyze the package using the semantic analyzer.
		ap, err := parser.ToHandle(astPkg)
		if err != nil {
			return err
		}

		root, err := semantic.AnalyzePackage(ap)
		if err != nil {
			return err
		}

		// Build an object with the initial set of identifiers
		// from the known builtin values.
		object, _ := values.BuildObject(func(set values.ObjectSetter) error {
			for k, v := range b.builtins[pkgpath] {
				set(k, v)
			}
			return nil
		})
		scope := r.prelude.Nest(object)

		// Run the interpreter on the package to construct the values
		// created by the package. Pass in the previously initialized
		// packages as importable packages as we evaluate these in order.
		importer := importer{pkgs: r.pkgs}
		itrp := interpreter.NewInterpreter(interpreter.NewPackage(""))
		if _, err := itrp.Eval(context.Background(), root, scope, &importer); err != nil {
			return err
		}
		obj, _ := values.BuildObject(func(set values.ObjectSetter) error {
			scope.LocalRange(set)
			return nil
		})
		r.pkgs[pkgpath] = interpreter.NewPackageWithValues(itrp.PackageName(), obj)
		for _, ppath := range prelude {
			if ppath == pkgpath {
				r.prelude.packages = append(r.prelude.packages, r.pkgs[pkgpath])
				break
			}
		}
	}

	r.rbuilder = nil
	return nil
}
