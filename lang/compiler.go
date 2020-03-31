package lang

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/spec"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

const (
	FluxCompilerType = "flux"
	ASTCompilerType  = "ast"
)

// AddCompilerMappings adds the Flux specific compiler mappings.
func AddCompilerMappings(mappings flux.CompilerMappings) error {
	if err := mappings.Add(FluxCompilerType, func() flux.Compiler {
		return new(FluxCompiler)

	}); err != nil {
		return err
	}
	if err := mappings.Add(ASTCompilerType, func() flux.Compiler {
		return new(ASTCompiler)

	}); err != nil {
		return err
	}
	return nil
}

// CompileOption represents an option for compilation.
type CompileOption func(*compileOptions)

type compileOptions struct {
	verbose bool

	extern *ast.File

	planOptions struct {
		logical  []plan.LogicalOption
		physical []plan.PhysicalOption
	}

	executeOptions []execute.ExecutionOption
}

func defaultOptions() *compileOptions {
	o := new(compileOptions)
	return o
}

func applyOptions(opts ...CompileOption) *compileOptions {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// NOTE(affo): compileOptions can be used only when invoking Compile* functions.
// They can't be used when unmarshaling a Compiler and invoking its Compile method.
// In order to make an implementation of `flux.Compiler` in package `lang` use some
// `lang.CompileOptions`, you must `Inject` those in the context. E.g.:
// ```
// 	opts := []CompileOption{lang.Verbose(true), lang.WithExtern(nil)}
// 	ctx = opts.Inject(ctx)
// ```

func Verbose(v bool) CompileOption {
	return func(o *compileOptions) {
		o.verbose = v
	}
}
func WithExtern(extern *ast.File) CompileOption {
	return func(o *compileOptions) {
		o.extern = extern
	}
}
func WithLogPlanOpts(lopts ...plan.LogicalOption) CompileOption {
	return func(o *compileOptions) {
		o.planOptions.logical = append(o.planOptions.logical, lopts...)
	}
}
func WithPhysPlanOpts(popts ...plan.PhysicalOption) CompileOption {
	return func(o *compileOptions) {
		o.planOptions.physical = append(o.planOptions.physical, popts...)
	}
}
func WithExecuteOptions(eopts ...execute.ExecutionOption) CompileOption {
	return func(o *compileOptions) {
		o.executeOptions = append(o.executeOptions, eopts...)
	}
}

// Compile evaluates a Flux script producing a flux.Program.
// now parameter must be non-zero, that is the default now time should be set before compiling.
func Compile(q string, now time.Time, opts ...CompileOption) (*AstProgram, error) {
	astPkg, err := flux.Parse(q)
	if err != nil {
		return nil, err
	}
	return CompileAST(astPkg, now, opts...), nil
}

// CompileAST evaluates a Flux AST and produces a flux.Program.
// now parameter must be non-zero, that is the default now time should be set before compiling.
func CompileAST(astPkg *ast.Package, now time.Time, opts ...CompileOption) *AstProgram {
	return &AstProgram{
		Program: &Program{
			opts: applyOptions(opts...),
		},
		Ast: astPkg,
		Now: now,
	}
}

// CompileTableObject evaluates a TableObject and produces a flux.Program.
// now parameter must be non-zero, that is the default now time should be set before compiling.
func CompileTableObject(ctx context.Context, to *flux.TableObject, now time.Time, opts ...CompileOption) (*Program, error) {
	o := applyOptions(opts...)
	s, err := spec.FromTableObject(ctx, to, now)
	if err != nil {
		return nil, err
	}
	if o.verbose {
		log.Println("Query Spec: ", flux.Formatted(s, flux.FmtJSON))
	}
	ps, err := buildPlan(context.Background(), s, o)
	if err != nil {
		return nil, err
	}
	return &Program{
		opts:     o,
		PlanSpec: ps,
	}, nil
}

func buildPlan(ctx context.Context, spec *flux.Spec, opts *compileOptions) (*plan.Spec, error) {
	s, _ := opentracing.StartSpanFromContext(ctx, "plan")
	defer s.Finish()
	pb := plan.PlannerBuilder{}

	planOptions := opts.planOptions

	lopts := planOptions.logical
	popts := planOptions.physical

	pb.AddLogicalOptions(lopts...)
	pb.AddPhysicalOptions(popts...)

	ps, err := pb.Build().Plan(spec)
	if err != nil {
		return nil, err
	}
	return ps, nil
}

// FluxCompiler compiles a Flux script into a spec.
type FluxCompiler struct {
	Now    time.Time
	Extern *ast.File `json:"extern"`
	Query  string    `json:"query"`
}

func (c FluxCompiler) Compile(ctx context.Context) (flux.Program, error) {
	opts := getCompileOptions(ctx)
	// Ignore context, it will be provided upon Program Start.
	return Compile(c.Query, c.Now, append(opts, WithExtern(c.Extern))...)
}

func (c FluxCompiler) CompilerType() flux.CompilerType {
	return FluxCompilerType
}

// ASTCompiler implements Compiler by producing a Spec from an AST.
type ASTCompiler struct {
	AST *ast.Package `json:"ast"`
	Now time.Time
}

func (c ASTCompiler) Compile(ctx context.Context) (flux.Program, error) {
	now := c.Now
	if now.IsZero() {
		now = time.Now()
	}
	opts := getCompileOptions(ctx)
	// Ignore context, it will be provided upon Program Start.
	return CompileAST(c.AST, now, opts...), nil
}

func (ASTCompiler) CompilerType() flux.CompilerType {
	return ASTCompilerType
}

// PrependFile prepends a file onto the compiler's list of package files.
func (c *ASTCompiler) PrependFile(file *ast.File) {
	c.AST.Files = append([]*ast.File{file}, c.AST.Files...)
}

// TableObjectCompiler compiles a TableObject into an executable flux.Program.
// It is not added to CompilerMappings and it is not serializable, because
// it is impossible to use it outside of the context of an ongoing execution.
type TableObjectCompiler struct {
	Tables *flux.TableObject
	Now    time.Time
}

func (c *TableObjectCompiler) Compile(ctx context.Context) (flux.Program, error) {
	opts := getCompileOptions(ctx)
	// Ignore context, it will be provided upon Program Start.
	return CompileTableObject(ctx, c.Tables, c.Now, opts...)
}

func (*TableObjectCompiler) CompilerType() flux.CompilerType {
	panic("TableObjectCompiler is not associated with a CompilerType")
}

type LoggingProgram interface {
	SetLogger(logger *zap.Logger)
}

// Program implements the flux.Program interface.
// It will execute a compiled plan using an executor.
type Program struct {
	Logger   *zap.Logger
	PlanSpec *plan.Spec

	opts *compileOptions
}

func (p *Program) SetLogger(logger *zap.Logger) {
	p.Logger = logger
}

func (p *Program) Start(ctx context.Context, alloc *memory.Allocator) (flux.Query, error) {
	ctx, cancel := context.WithCancel(ctx)

	// This span gets closed by the query when it is done.
	s, cctx := opentracing.StartSpanFromContext(ctx, "execute")
	results := make(chan flux.Result)
	q := &query{
		results: results,
		alloc:   alloc,
		span:    s,
		cancel:  cancel,
		stats: flux.Statistics{
			Metadata: make(flux.Metadata),
		},
	}

	e := execute.NewExecutor(p.Logger, p.opts.executeOptions...)
	resultMap, md, err := e.Execute(cctx, p.PlanSpec, q.alloc)
	if err != nil {
		s.Finish()
		return nil, err
	}

	// There was no error so send the results downstream.
	q.wg.Add(1)
	go p.processResults(cctx, q, resultMap)

	// Begin reading from the metadata channel.
	q.wg.Add(1)
	go p.readMetadata(q, md)

	return q, nil
}

func (p *Program) processResults(ctx context.Context, q *query, resultMap map[string]flux.Result) {
	defer q.wg.Done()
	defer close(q.results)

	for _, res := range resultMap {
		select {
		case q.results <- res:
		case <-ctx.Done():
			q.err = ctx.Err()
			return
		}
	}
}

func (p *Program) readMetadata(q *query, metaCh <-chan flux.Metadata) {
	defer q.wg.Done()
	for md := range metaCh {
		q.stats.Metadata.AddAll(md)
	}
}

// AstProgram wraps a Program with an AST that will be evaluated upon Start.
// As such, the PlanSpec is populated after Start and evaluation errors are returned by Start.
type AstProgram struct {
	*Program

	Ast *ast.Package
	Now time.Time
}

func (p *AstProgram) getSpec(ctx context.Context, alloc *memory.Allocator) (*flux.Spec, values.Scope, error) {
	if p.opts == nil {
		p.opts = defaultOptions()
	}
	if p.Now.IsZero() {
		p.Now = time.Now()
	}
	if p.opts.extern != nil {
		p.Ast.Files = append([]*ast.File{p.opts.extern}, p.Ast.Files...)
	}
	// The program must inject execution dependencies to make it available
	// to function calls during the evaluation phase (see `tableFind`).
	deps := ExecutionDependencies{
		Allocator: alloc,
		Logger:    p.Logger,
	}
	ctx = deps.Inject(ctx)
	s, cctx := opentracing.StartSpanFromContext(ctx, "eval")
	sideEffects, scope, err := flux.EvalAST(cctx, p.Ast, flux.SetNowOption(p.Now))
	if err != nil {
		return nil, nil, err
	}
	s.Finish()

	s, cctx = opentracing.StartSpanFromContext(ctx, "compile")
	defer s.Finish()
	nowOpt, ok := scope.Lookup(flux.NowOption)
	if !ok {
		return nil, nil, fmt.Errorf("%q option not set", flux.NowOption)
	}
	nowTime, err := nowOpt.Function().Call(ctx, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, codes.Inherit, "error in evaluating AST while starting program")
	}
	p.Now = nowTime.Time().Time()
	sp, err := spec.FromEvaluation(cctx, sideEffects, p.Now)
	if err != nil {
		return nil, nil, errors.Wrap(err, codes.Inherit, "error in query specification while starting program")
	}
	return sp, scope, nil
}

func (p *AstProgram) Start(ctx context.Context, alloc *memory.Allocator) (flux.Query, error) {
	sp, scope, err := p.getSpec(ctx, alloc)
	if err != nil {
		return nil, err
	}
	s, cctx := opentracing.StartSpanFromContext(ctx, "plan")
	if p.opts.verbose {
		log.Println("Query Spec: ", flux.Formatted(sp, flux.FmtJSON))
	}
	if err := p.updateOpts(scope); err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "error in reading options while starting program")
	}
	ps, err := buildPlan(cctx, sp, p.opts)
	if err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "error in building plan while starting program")
	}
	p.PlanSpec = ps
	s.Finish()

	s, cctx = opentracing.StartSpanFromContext(ctx, "start-program")
	defer s.Finish()
	return p.Program.Start(cctx, alloc)
}

func (p *AstProgram) updateOpts(scope values.Scope) error {
	n := getPlannerPkgName(p.Ast)
	if n == "" {
		// No import for 'planner'. Nothing to update.
		return nil
	}
	lo, po, err := getPlanOptions(scope, n)
	if err != nil {
		return err
	}
	if lo != nil {
		p.opts.planOptions.logical = append(p.opts.planOptions.logical, lo)
	}
	if po != nil {
		p.opts.planOptions.physical = append(p.opts.planOptions.physical, po)
	}
	return nil
}

func getPlannerPkgName(pkg *ast.Package) string {
	for _, f := range pkg.Files {
		for _, imp := range f.Imports {
			if path := imp.Path.Value; path == "planner" {
				name := path
				if alias := imp.As; alias != nil {
					name = alias.Name
				}
				return name
			}
		}
	}
	return ""
}

func getPlanOptions(scope values.Scope, pkgName string) (plan.LogicalOption, plan.PhysicalOption, error) {
	// find the 'planner' package
	plannerPkg, ok := scope.Lookup(pkgName)
	if !ok {
		// No import for planner, this is useless.
		return nil, nil, nil
	}
	if plannerPkg.Type().Nature() != semantic.Object {
		// No import for planner, this is useless.
		return nil, nil, nil
	}

	ls, err := getRules(plannerPkg.Object(), "disableLogicalRules")
	if err != nil {
		return nil, nil, err
	}
	ps, err := getRules(plannerPkg.Object(), "disablePhysicalRules")
	if err != nil {
		return nil, nil, err
	}
	return plan.RemoveLogicalRules(ls...), plan.RemovePhysicalRules(ps...), nil
}

func getRules(plannerPkg values.Object, optionName string) ([]string, error) {
	value, ok := plannerPkg.Get(optionName)
	if !ok {
		// No value in package.
		return []string{}, nil
	}

	// TODO(affo): the rules are arrays of strings as defined in the 'planner' package.
	//  During evaluation, the interpreter should raise an error if the user tries to assign
	//  an option of a type to another. So we should be able to rely on the fact that the type
	//  for value is fixed. At the moment is it not so.
	//  So, we have to check and return an error to avoid a panic.
	//  See (https://github.com/influxdata/flux/issues/1829).
	if t := value.Type().Nature(); t != semantic.Array {
		return nil, fmt.Errorf("'planner.%s' must be an array of string, got %s", optionName, t.String())
	}
	rules := value.Array()
	if et := rules.Type().ElementType().Nature(); et != semantic.String {
		return nil, fmt.Errorf("'planner.%s' must be an array of string, got an array of %s", optionName, et.String())
	}
	noRules := rules.Len()
	rs := make([]string, noRules)
	rules.Range(func(i int, v values.Value) {
		rs[i] = v.Str()
	})
	return rs, nil
}

// WalkIR applies the function `f` to each operation in the compiled spec.
// WARNING: this function evaluates the AST using an unlimited allocator.
// In case of dynamic queries this could lead to unexpected memory usage.
func WalkIR(ctx context.Context, astPkg *ast.Package, f func(o *flux.Operation) error) error {
	p := CompileAST(astPkg, time.Now())
	if sp, _, err := p.getSpec(ctx, new(memory.Allocator)); err != nil {
		return err
	} else {
		return sp.Walk(f)
	}
}
