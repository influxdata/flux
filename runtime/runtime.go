package runtime

import (
	"context"
	"encoding/json"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/libflux/go/libflux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// Default contains the preregistered packages and builtin values
// required to execute a flux script.
var Default = &runtime{}

// runtime contains the flux runtime for interpreting and
// executing queries.
type runtime struct {
	astPkgs   map[string]*ast.Package
	pkgs      map[string]*semantic.Package
	builtins  map[string]map[string]values.Value
	finalized bool
}

func (r *runtime) Parse(flux string) (flux.ASTHandle, error) {
	return Parse(flux)
}

func (r *runtime) JSONToHandle(json []byte) (flux.ASTHandle, error) {
	return libflux.ParseJSON(json)
}

func (r *runtime) MergePackages(dst, src flux.ASTHandle) error {
	return MergePackages(dst, src)
}

func (r *runtime) IsPreludePackage(pkg string) bool {
	for _, p := range prelude {
		if p == pkg {
			return true
		}
	}
	return false
}

func (r *runtime) LookupBuiltinType(pkg, name string) (semantic.MonoType, error) {
	return LookupBuiltinType(pkg, name)
}

func (r *runtime) RegisterPackage(pkg *ast.Package) error {
	if r.finalized {
		return errors.New(codes.Internal, "already finalized, cannot register builtin package")
	}

	if r.astPkgs == nil {
		r.astPkgs = make(map[string]*ast.Package)
	}

	if _, ok := r.astPkgs[pkg.Path]; ok {
		return errors.Newf(codes.Internal, "duplicate builtin package %q", pkg.Path)
	}

	if ast.Check(pkg) > 0 {
		err := ast.GetError(pkg)
		return errors.Wrapf(err, codes.Inherit, "failed to parse builtin package %q", pkg.Path)
	}
	r.astPkgs[pkg.Path] = pkg
	return nil
}

func (r *runtime) RegisterPackageValue(pkgpath, name string, value values.Value) error {
	return r.registerPackageValue(pkgpath, name, value, false)
}

func (r *runtime) ReplacePackageValue(pkgpath, name string, value values.Value) error {
	return r.registerPackageValue(pkgpath, name, value, true)
}

func (r *runtime) registerPackageValue(pkgpath, name string, value values.Value, replace bool) error {
	if r.finalized {
		return errors.Newf(codes.Internal, "already finalized, cannot register builtin package value")
	}

	if r.builtins == nil {
		r.builtins = make(map[string]map[string]values.Value)
	}

	pkg, ok := r.builtins[pkgpath]
	if !ok {
		pkg = make(map[string]values.Value)
		r.builtins[pkgpath] = pkg
	}

	if _, ok := pkg[name]; ok && !replace {
		return errors.Newf(codes.Internal, "duplicate builtin package value %q %q", pkgpath, name)
	} else if !ok && replace {
		return errors.Newf(codes.Internal, "missing builtin package value %q %q", pkgpath, name)
	}
	pkg[name] = value
	return nil
}

func (r *runtime) Prelude() values.Scope {
	if !r.finalized {
		panic("builtins not finalized")
	}
	importer := r.Stdlib()
	scope, err := r.newScopeFor("main", importer)
	if err != nil {
		panic(err)
	}
	return scope
}

func (r *runtime) Eval(ctx context.Context, astPkg flux.ASTHandle, es interpreter.ExecOptsConfig, opts ...flux.ScopeMutator) ([]interpreter.SideEffect, values.Scope, error) {
	semPkg, err := AnalyzePackage(astPkg)
	if err != nil {
		return nil, nil, err
	}

	// Construct the initial scope for this package.
	importer := &importer{r: r}
	scope, err := r.newScopeFor("main", importer)
	if err != nil {
		return nil, nil, err
	}

	// Mutate the scope with any additional options.
	for _, opt := range opts {
		opt(r, scope)
	}

	// Execute the interpreter over the package.
	itrp := interpreter.NewInterpreter(nil, es)
	sideEffects, err := itrp.Eval(ctx, semPkg, scope, importer)
	if err != nil {
		return nil, nil, err
	}
	return sideEffects, scope, nil
}

// newScopeFor constructs a new scope for the given package using the
// passed in importer.
func (r *runtime) newScopeFor(pkgpath string, imp interpreter.Importer) (values.Scope, error) {
	// Construct the prelude scope from the prelude paths.
	// If we are importing part of the prelude, we do not
	// include it as part of the prelude and will stop
	// including values as soon as we hit the prelude.
	// This allows us to import all previous paths when loading
	// the prelude, but avoid a circular import.
	preludeScope := values.NewScope()
	for _, path := range prelude {
		if path == pkgpath {
			break
		}

		p, err := imp.ImportPackageObject(path)
		if err != nil {
			return nil, err
		}
		p.Range(preludeScope.Set)
	}

	// Build an object with the initial set of identifiers
	// from the known builtin values.
	object := values.NewObjectWithValues(r.builtins[pkgpath])
	scope := values.NewNestedScope(preludeScope, object)
	return scope, nil
}

func (r *runtime) Stdlib() interpreter.Importer {
	if !r.finalized {
		panic("builtins not finalized")
	}
	return &importer{r: r}
}

func (r *runtime) compilePackages() error {
	pkgs := make(map[string]*semantic.Package)
	for _, pkg := range r.astPkgs {
		bs, err := json.Marshal(pkg)
		if err != nil {
			return err
		}
		hdl, err := r.JSONToHandle(bs)
		if err != nil {
			return err
		}
		root, err := AnalyzePackage(hdl)
		if err != nil {
			return err
		}
		pkgs[pkg.Path] = root
	}
	r.pkgs = pkgs
	r.astPkgs = nil
	return nil
}

func (r *runtime) Finalize() error {
	if r.finalized {
		return errors.New(codes.Internal, "already finalized")
	}
	r.finalized = true

	if err := r.compilePackages(); err != nil {
		return err
	}

	for path, pkg := range r.builtins {
		semPkg, ok := r.pkgs[path]
		if !ok {
			return errors.Newf(codes.Internal, "missing semantic package %s", path)
		}
		if err := validatePackageBuiltins(pkg, semPkg); err != nil {
			return err
		}
	}
	return nil
}

// validatePackageBuiltins ensures that all package builtins have both an AST builtin statement and a registered value.
func validatePackageBuiltins(pkg map[string]values.Value, semPkg *semantic.Package) error {
	builtinStmts := make(map[string]*semantic.BuiltinStatement)
	semantic.Walk(semantic.CreateVisitor(func(n semantic.Node) {
		if bs, ok := n.(*semantic.BuiltinStatement); ok {
			builtinStmts[bs.ID.Name] = bs
		}
	}), semPkg)

	missing := make([]string, 0, len(builtinStmts))
	extra := make([]string, 0, len(builtinStmts))

	for n := range builtinStmts {
		if _, ok := pkg[n]; !ok {
			missing = append(missing, n)
			continue
		}
		// TODO(nathanielc): Ensure that the value's type matches the type expression
	}
	for k := range pkg {
		if _, ok := builtinStmts[k]; !ok {
			extra = append(extra, k)
		}
	}
	if len(missing) > 0 || len(extra) > 0 {
		return errors.Newf(codes.Internal, "missing builtin values %v, extra builtin values %v", missing, extra)
	}
	return nil
}
