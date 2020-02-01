package flux

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/libflux/go/libflux"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	TablesParameter = "tables"
	tableKindKey    = "kind"
	tableParentsKey = "parents"
	tableSpecKey    = "spec"

	NowOption = "now"
	nowPkg    = "universe"
)

// Parse parses a Flux script and produces an ast.Package.
func Parse(flux string) (*ast.Package, error) {
	astPkg := parser.ParseSource(flux)
	if ast.Check(astPkg) > 0 {
		return nil, ast.GetError(astPkg)
	}

	return astPkg, nil
}

// Eval accepts a Flux script and evaluates it to produce a set of side effects (as a slice of values) and a scope.
func Eval(ctx context.Context, flux string, opts ...ScopeMutator) ([]interpreter.SideEffect, values.Scope, error) {
	h := parser.ParseToHandle([]byte(flux))
	return evalHandle(ctx, h, opts...)
}

// EvalAST accepts a Flux AST and evaluates it to produce a set of side effects (as a slice of values) and a scope.
func EvalAST(ctx context.Context, astPkg *ast.Package, opts ...ScopeMutator) ([]interpreter.SideEffect, values.Scope, error) {
	h, err := parser.ToHandle(astPkg)
	if err != nil {
		return nil, nil, err
	}
	return evalHandle(ctx, h, opts...)
}

func evalHandle(ctx context.Context, h *libflux.ASTPkg, opts ...ScopeMutator) ([]interpreter.SideEffect, values.Scope, error) {
	semPkg, err := semantic.AnalyzePackage(h)
	if err != nil {
		return nil, nil, err
	}

	itrp := interpreter.NewInterpreter(nil)
	// Create a scope for execution whose parent is a copy of the prelude and whose current scope is the package.
	// A copy of the prelude must be used since options can be mutated.
	scope := defaultRuntime.prelude.Copy().Nest(nil)

	for _, opt := range opts {
		opt(scope)
	}

	sideEffects, err := itrp.Eval(ctx, semPkg, scope, StdLib())
	if err != nil {
		return nil, nil, err
	}

	return sideEffects, scope, nil
}

// EvalOptions is like EvalAST, but only evaluates options.
func EvalOptions(ctx context.Context, astPkg *ast.Package, opts ...ScopeMutator) ([]interpreter.SideEffect, values.Scope, error) {
	return EvalAST(ctx, options(astPkg), opts...)
}

// options returns a shallow copy of the AST, trimmed to include only option statements.
func options(astPkg *ast.Package) *ast.Package {
	trimmed := &ast.Package{
		BaseNode: astPkg.BaseNode,
		Path:     astPkg.Path,
		Package:  astPkg.Package,
	}
	for _, f := range astPkg.Files {
		var body []ast.Statement
		for _, s := range f.Body {
			if opt, ok := s.(*ast.OptionStatement); ok {
				body = append(body, opt)
			}
		}
		if len(body) > 0 {
			trimmed.Files = append(trimmed.Files, &ast.File{
				Body:     body,
				BaseNode: f.BaseNode,
				Name:     f.Name,
				Package:  f.Package,
				Imports:  f.Imports,
			})
		}
	}

	return trimmed
}

// ScopeMutator is any function that mutates the scope of an identifier.
type ScopeMutator = func(values.Scope)

// SetOption returns a func that adds a var binding to a scope.
func SetOption(pkg, name string, v values.Value) ScopeMutator {
	return func(scope values.Scope) {
		scope.SetOption(pkg, name, v)
	}
}

// SetNowOption returns a ScopeMutator that sets the `now` option to the given time.
func SetNowOption(now time.Time) ScopeMutator {
	return SetOption(nowPkg, NowOption, generateNowFunc(now))
}

func generateNowFunc(now time.Time) values.Function {
	timeVal := values.NewTime(values.ConvertTime(now))
	ftype := semantic.MustLookupBuiltinType("universe", "now")
	call := func(ctx context.Context, args values.Object) (values.Value, error) {
		return timeVal, nil
	}
	sideEffect := false
	return values.NewFunction(NowOption, ftype, call, sideEffect)
}

type CreateOperationSpec func(args Arguments, a *Administration) (OperationSpec, error)

// set of builtins
var (
	// list of packages included in the prelude.
	// Packages must be listed in import order
	prelude = []string{
		"universe",
		"influxdata/influxdb",
	}
)

type scopeSet struct {
	packages []*interpreter.Package
}

func (s *scopeSet) Lookup(name string) (values.Value, bool) {
	for _, pkg := range s.packages {
		if v, ok := pkg.Get(name); ok {
			if _, ok := v.(values.Package); ok {
				// prelude should not expose any imported packages
				return nil, false
			}
			return v, ok
		}
	}
	return nil, false
}
func (s *scopeSet) LocalLookup(name string) (values.Value, bool) {
	// scopeSet is always a top level scope
	return s.Lookup(name)
}

func (s *scopeSet) Set(name string, v values.Value) {
	panic("cannot mutate the universe block")
}
func (s *scopeSet) SetOption(pkg, name string, v values.Value) (bool, error) {
	for _, p := range s.packages {
		if _, ok := p.Get(name); ok || p.Name() == pkg {
			p.SetOption(name, v)
			return true, nil
		}
	}
	return false, nil
}

func (s *scopeSet) Nest(obj values.Object) values.Scope {
	return values.NewNestedScope(s, obj)
}

func (s *scopeSet) Pop() values.Scope {
	return nil
}

func (s *scopeSet) Size() int {
	var size int
	for _, pkg := range s.packages {
		size += pkg.Len()
	}
	return size
}

func (s *scopeSet) Range(f func(k string, v values.Value)) {
	for _, pkg := range s.packages {
		if pkg == nil {
			panic(`nil package in scope; try importing "github.com/influxdata/flux/builtin"`)
		}
		pkg.Range(func(k string, v values.Value) {
			if _, ok := v.(values.Package); ok {
				// prelude should not expose any imported packages
				return
			}
			f(k, v)
		})
	}
}

func (s *scopeSet) LocalRange(f func(k string, v values.Value)) {
	// scopeSet is always a top level scope
	s.Range(f)
}

func (s *scopeSet) SetReturn(v values.Value) {
	panic("cannot set return value on universe block")
}

func (s *scopeSet) Return() values.Value {
	return nil
}

func (s *scopeSet) Copy() values.Scope {
	packages := make([]*interpreter.Package, len(s.packages))
	for i, pkg := range s.packages {
		packages[i] = pkg.Copy()
	}
	return &scopeSet{
		packages: packages,
	}
}

// StdLib returns an importer for the Flux standard library.
func StdLib() interpreter.Importer {
	return defaultRuntime.Stdlib()
}

// Prelude returns a scope object representing the Flux universe block
func Prelude() values.Scope {
	return defaultRuntime.Prelude()
}

// RegisterPackage adds a builtin package
func RegisterPackage(pkg *ast.Package) {
	if err := defaultRuntime.RegisterPackage(pkg); err != nil {
		panic(err)
	}
}

// RegisterPackageValue adds a value for an identifier in a builtin package
func RegisterPackageValue(pkgpath, name string, value values.Value) {
	if err := defaultRuntime.RegisterPackageValue(pkgpath, name, value); err != nil {
		panic(err)
	}
}

// ReplacePackageValue replaces a value for an identifier in a builtin package
func ReplacePackageValue(pkgpath, name string, value values.Value) {
	if err := defaultRuntime.ReplacePackageValue(pkgpath, name, value); err != nil {
		panic(err)
	}
}

// MustValue panics if err is not nil, otherwise value is returned.
func MustValue(v values.Value, err error) values.Value {
	if err != nil {
		panic(err)
	}
	return v
}

// FunctionValue creates a values.Value from the operation spec and signature.
// Name is the name of the function as it would be called.
// c is a function reference of type CreateOperationSpec
// sig is a function signature type that specifies the names and types of each argument for the function.
func FunctionValue(name string, c CreateOperationSpec, ft semantic.MonoType) (values.Value, error) {
	return functionValue(name, c, ft, false)
}

// FunctionValueWithSideEffect creates a values.Value from the operation spec and signature.
// Name is the name of the function as it would be called.
// c is a function reference of type CreateOperationSpec
// sig is a function signature type that specifies the names and types of each argument for the function.
func FunctionValueWithSideEffect(name string, c CreateOperationSpec, ft semantic.MonoType) (values.Value, error) {
	return functionValue(name, c, ft, true)
}

func functionValue(name string, c CreateOperationSpec, mt semantic.MonoType, sideEffects bool) (values.Value, error) {
	if c == nil {
		c = func(args Arguments, a *Administration) (OperationSpec, error) {
			return nil, errors.Newf(codes.Unimplemented, "function %q is not implemented", name)
		}
	}
	if mt.Nature() != semantic.Function {
		return nil, errors.Newf(codes.Invalid, "cannot implement function %q with value of type %v", name, mt)
	}
	return &function{
		t:             mt,
		name:          name,
		createOpSpec:  c,
		hasSideEffect: sideEffects,
	}, nil
}

// FinalizeBuiltIns must be called to complete registration.
// Future calls to RegisterFunction or RegisterPackageValue will panic.
func FinalizeBuiltIns() {
	if err := defaultRuntime.Finalize(); err != nil {
		panic(err)
	}
}

// TODO(algow): Needs to be refactored into the runtime finalize.
// func evalBuiltInPackages() error {
// 	order, err := packageOrder(prelude, builtinPackages)
// 	if err != nil {
// 		return err
// 	}
// 	for _, astPkg := range order {
// 		if ast.Check(astPkg) > 0 {
// 			err := ast.GetError(astPkg)
// 			return errors.Wrapf(err, codes.Inherit, "failed to parse builtin package %q", astPkg.Path)
// 		}
// TODO(algow): https://github.com/influxdata/flux/issues/2404
//semPkg, err := semantic.New(astPkg)
//if err != nil {
//	return errors.Wrapf(err, codes.Inherit, "failed to create semantic graph for builtin package %q", astPkg.Path)
//}
//
//pkg := stdlib.pkgs[astPkg.Path]
//if pkg == nil {
//	return errors.Wrapf(err, codes.Inherit, "package does not exist %q", astPkg.Path)
//}
//
//// Validate packages before evaluating them
//if err := validatePackageBuiltins(pkg, astPkg); err != nil {
//	return errors.Wrapf(err, codes.Inherit, "package has invalid builtins %q", astPkg.Path)
//}
//
//itrp := interpreter.NewInterpreter(pkg)
//if _, err := itrp.Eval(context.Background(), semPkg, preludeScope.Nest(pkg), stdlib); err != nil {
//	return errors.Wrapf(err, codes.Inherit, "failed to evaluate builtin package %q", astPkg.Path)
//}
// }
// return nil
// }

// TODO(algow): Needs to be refactored into the runtime finalize.
// validatePackageBuiltins ensures that all package builtins have both an AST builtin statement and a registered value.
func validatePackageBuiltins(pkg *interpreter.Package, astPkg *ast.Package) error {
	builtinStmts := make(map[string]*ast.BuiltinStatement)
	ast.Walk(ast.CreateVisitor(func(n ast.Node) {
		if bs, ok := n.(*ast.BuiltinStatement); ok {
			builtinStmts[bs.ID.Name] = bs
		}
	}), astPkg)

	missing := make([]string, 0, len(builtinStmts))
	extra := make([]string, 0, len(builtinStmts))

	for n := range builtinStmts {
		if _, ok := pkg.Get(n); !ok {
			missing = append(missing, n)
			continue
		}
		// TODO(nathanielc): Ensure that the value's type matches the type expression
	}
	pkg.Range(func(k string, v values.Value) {
		if _, ok := builtinStmts[k]; !ok {
			extra = append(extra, k)
			return
		}
	})
	if len(missing) > 0 || len(extra) > 0 {
		return errors.Newf(codes.Internal, "missing builtin values %v, extra builtin values %v", missing, extra)
	}
	return nil
}

var _ = tableSpecKey // So that linter doesn't think tableSpecKey is unused, considering above TODO.

// IDer produces the mapping of table Objects to OperationIDs
type IDer interface {
	ID(*TableObject) OperationID
}

// IDerOpSpec is the interface any operation spec that needs
// access to OperationIDs in the query spec must implement.
type IDerOpSpec interface {
	IDer(ider IDer)
}

// TableObject represents the value returned by a transformation.
// As such, it holds the OperationSpec of the transformation it is associated with,
// and it is a values.Value (and, also, a values.Object).
// It can be compiled and executed as a flux.Program by using a lang.TableObjectCompiler.
type TableObject struct {
	// TODO(Josh): Remove args once the
	// OperationSpec interface has an Equal method.
	t       semantic.MonoType
	args    Arguments
	Kind    OperationKind
	Spec    OperationSpec
	Parents []*TableObject
}

func (t *TableObject) Operation(ider IDer) *Operation {
	if iderOpSpec, ok := t.Spec.(IDerOpSpec); ok {
		iderOpSpec.IDer(ider)
	}

	return &Operation{
		ID:   ider.ID(t),
		Spec: t.Spec,
	}
}
func (t *TableObject) IsNull() bool {
	return false
}
func (t *TableObject) String() string {
	str := new(strings.Builder)
	t.str(str, false)
	return str.String()
}
func (t *TableObject) str(b *strings.Builder, arrow bool) {
	multiParent := len(t.Parents) > 1
	if multiParent {
		b.WriteString("( ")
	}
	for _, v := range t.Parents {
		v.str(b, !multiParent)
		if multiParent {
			b.WriteString("; ")
		}
	}
	if multiParent {
		b.WriteString(" ) -> ")
	}
	b.WriteString(string(t.Kind))
	if arrow {
		b.WriteString(" -> ")
	}
}

func (t *TableObject) Type() semantic.MonoType {
	return t.t
}

func (t *TableObject) Str() string {
	panic(values.UnexpectedKind(semantic.Array, semantic.String))
}
func (t *TableObject) Bytes() []byte {
	panic(values.UnexpectedKind(semantic.Array, semantic.Bytes))
}
func (t *TableObject) Int() int64 {
	panic(values.UnexpectedKind(semantic.Array, semantic.Int))
}
func (t *TableObject) UInt() uint64 {
	panic(values.UnexpectedKind(semantic.Array, semantic.UInt))
}
func (t *TableObject) Float() float64 {
	panic(values.UnexpectedKind(semantic.Array, semantic.Float))
}
func (t *TableObject) Bool() bool {
	panic(values.UnexpectedKind(semantic.Array, semantic.Bool))
}
func (t *TableObject) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Array, semantic.Time))
}
func (t *TableObject) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Array, semantic.Duration))
}
func (t *TableObject) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Array, semantic.Regexp))
}
func (t *TableObject) Array() values.Array {
	return t
}
func (t *TableObject) Object() values.Object {
	panic(values.UnexpectedKind(semantic.Array, semantic.Object))
}
func (t *TableObject) Equal(rhs values.Value) bool {
	v, ok := rhs.(*TableObject)
	return ok && t == v
}
func (t *TableObject) Function() values.Function {
	panic(values.UnexpectedKind(semantic.Array, semantic.Function))
}

func (t *TableObject) Get(i int) values.Value {
	panic("cannot index into stream")
}
func (t *TableObject) Set(i int, v values.Value) {
	panic("cannot index into stream")
}
func (t *TableObject) Append(v values.Value) {
	panic("cannot append onto stream")
}
func (t *TableObject) Len() int {
	panic("length of stream not supported")
}
func (t *TableObject) Range(f func(i int, v values.Value)) {
	panic("cannot range over values in stream")
}
func (t *TableObject) Sort(f func(i, j values.Value) bool) {
	panic("cannot sort stream")
}

type Administration struct {
	parents []*TableObject
}

func newAdministration() *Administration {
	return &Administration{
		parents: make([]*TableObject, 0, 8),
	}
}

// AddParentFromArgs reads the args for the `table` argument and adds the value as a parent.
func (a *Administration) AddParentFromArgs(args Arguments) error {
	parent, ok := args.Get(TablesParameter)
	if !ok {
		return errors.Newf(codes.Invalid, "could not find %s parameter", TablesParameter)
	}
	p, ok := parent.(*TableObject)
	if !ok {
		return errors.Newf(codes.Invalid, "argument is not a table object: got %T", parent)
	}
	a.AddParent(p)
	return nil
}

// AddParent instructs the evaluation Context that a new edge should be created from the parent to the current operation.
// Duplicate parents will be removed, so the caller need not concern itself with which parents have already been added.
func (a *Administration) AddParent(np *TableObject) {
	// Check for duplicates
	for _, v := range a.parents {
		if v == np {
			return
		}
	}
	a.parents = append(a.parents, np)
}

type function struct {
	name          string
	t             semantic.MonoType
	createOpSpec  CreateOperationSpec
	hasSideEffect bool
}

func (f *function) Type() semantic.MonoType {
	return f.t
}
func (f *function) IsNull() bool {
	return false
}
func (f *function) Str() string {
	panic(values.UnexpectedKind(semantic.Function, semantic.String))
}
func (f *function) Bytes() []byte {
	panic(values.UnexpectedKind(semantic.Function, semantic.Bytes))
}
func (f *function) Int() int64 {
	panic(values.UnexpectedKind(semantic.Function, semantic.Int))
}
func (f *function) UInt() uint64 {
	panic(values.UnexpectedKind(semantic.Function, semantic.UInt))
}
func (f *function) Float() float64 {
	panic(values.UnexpectedKind(semantic.Function, semantic.Float))
}
func (f *function) Bool() bool {
	panic(values.UnexpectedKind(semantic.Function, semantic.Bool))
}
func (f *function) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Function, semantic.Time))
}
func (f *function) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Function, semantic.Duration))
}
func (f *function) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Function, semantic.Regexp))
}
func (f *function) Array() values.Array {
	panic(values.UnexpectedKind(semantic.Function, semantic.Array))
}
func (f *function) Object() values.Object {
	panic(values.UnexpectedKind(semantic.Function, semantic.Object))
}
func (f *function) Function() values.Function {
	return f
}
func (f *function) Equal(rhs values.Value) bool {
	if f.Type() != rhs.Type() {
		return false
	}
	v, ok := rhs.(*function)
	return ok && (f == v)
}
func (f *function) HasSideEffect() bool {
	return f.hasSideEffect
}

func (f *function) Call(ctx context.Context, args values.Object) (values.Value, error) {
	return interpreter.DoFunctionCall(f.call, args)
}

func (f *function) call(args interpreter.Arguments) (values.Value, error) {
	returnType, err := f.t.ReturnType()
	if err != nil {
		return nil, err
	}

	a := newAdministration()
	arguments := Arguments{Arguments: args}
	spec, err := f.createOpSpec(arguments, a)
	if err != nil {
		return nil, err
	}

	t := &TableObject{
		t:       returnType,
		args:    arguments,
		Kind:    spec.Kind(),
		Spec:    spec,
		Parents: a.parents,
	}
	return t, nil
}
func (f *function) String() string {
	return fmt.Sprintf("%v", f.t)
}

type Arguments struct {
	interpreter.Arguments
}

func (a Arguments) GetTime(name string) (Time, bool, error) {
	v, ok := a.Get(name)
	if !ok {
		return Time{}, false, nil
	}
	qt, err := ToQueryTime(v)
	if err != nil {
		return Time{}, ok, err
	}
	return qt, ok, nil
}

func (a Arguments) GetRequiredTime(name string) (Time, error) {
	qt, ok, err := a.GetTime(name)
	if err != nil {
		return Time{}, err
	}
	if !ok {
		return Time{}, errors.Newf(codes.Invalid, "missing required keyword argument %q", name)
	}
	return qt, nil
}

func (a Arguments) GetDuration(name string) (Duration, bool, error) {
	v, ok := a.Get(name)
	if !ok {
		return ConvertDuration(0), false, nil
	}
	return v.Duration(), true, nil
}

func (a Arguments) GetRequiredDuration(name string) (Duration, error) {
	d, ok, err := a.GetDuration(name)
	if err != nil {
		return ConvertDuration(0), err
	}
	if !ok {
		return ConvertDuration(0), errors.Newf(codes.Invalid, "missing required keyword argument %q", name)
	}
	return d, nil
}

func ToQueryTime(value values.Value) (Time, error) {
	switch value.Type().Nature() {
	case semantic.Time:
		return Time{
			Absolute: value.Time().Time(),
		}, nil
	case semantic.Duration:
		return Time{
			Relative:   value.Duration().Duration(),
			IsRelative: true,
		}, nil
	case semantic.Int:
		return Time{
			Absolute: time.Unix(value.Int(), 0),
		}, nil
	default:
		return Time{}, errors.Newf(codes.Invalid, "value is not a time, got %v", value.Type())
	}
}

type importer struct {
	pkgs map[string]*interpreter.Package
}

func (imp *importer) Copy() *importer {
	packages := make(map[string]*interpreter.Package, len(imp.pkgs))
	for k, v := range imp.pkgs {
		packages[k] = v.Copy()
	}
	return &importer{
		pkgs: packages,
	}
}

func (imp *importer) Import(path string) (semantic.MonoType, bool) {
	p, ok := imp.pkgs[path]
	if !ok {
		return semantic.MonoType{}, false
	}
	return p.Type(), true
}

func (imp *importer) ImportPackageObject(path string) (*interpreter.Package, bool) {
	p, ok := imp.pkgs[path]
	return p, ok
}

// packageOrder determines a safe order to process builtin packages such that all dependent packages are previously processed.
func packageOrder(prelude []string, pkgs map[string]*ast.Package) (order []*ast.Package, err error) {
	//TODO(nathanielc): Add import cycle detection, this is not needed until this code is promoted to work with third party imports

	// Always import prelude first so other packages need not explicitly import the prelude packages.
	for _, path := range prelude {
		pkg := pkgs[path]
		order, err = insertPkg(pkg, pkgs, order)
		if err != nil {
			return
		}
	}
	// Import all other packages
	for _, pkg := range pkgs {
		order, err = insertPkg(pkg, pkgs, order)
		if err != nil {
			return
		}
	}
	return
}

func insertPkg(pkg *ast.Package, pkgs map[string]*ast.Package, order []*ast.Package) (_ []*ast.Package, err error) {
	imports := findImports(pkg)
	for _, path := range imports {
		dep, ok := pkgs[path]
		if !ok {
			return nil, errors.Newf(codes.Invalid, "unknown builtin package %q", path)
		}
		order, err = insertPkg(dep, pkgs, order)
		if err != nil {
			return nil, err
		}
	}
	return appendPkg(pkg, order), nil
}

func appendPkg(pkg *ast.Package, pkgs []*ast.Package) []*ast.Package {
	if containsPkg(pkg.Path, pkgs) {
		return pkgs
	}
	return append(pkgs, pkg)
}

func containsPkg(path string, pkgs []*ast.Package) bool {
	for _, pkg := range pkgs {
		if pkg.Path == path {
			return true
		}
	}
	return false
}

func findImports(pkg *ast.Package) (imports []string) {
	for _, f := range pkg.Files {
		for _, i := range f.Imports {
			imports = append(imports, i.Path.Value)
		}
	}
	return
}
