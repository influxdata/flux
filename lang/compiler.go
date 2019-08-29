package lang

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/spec"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/opentracing/opentracing-go"
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

	planOptions planOptions
}

type planOptions struct {
	logical  []plan.LogicalOption
	physical []plan.PhysicalOption
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
func WithExtern(extern *ast.File) CompileOption {
	return func(o *compileOptions) {
		o.extern = extern
	}
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

// NOTE: compileOptions can be used only when invoking Compile* functions.
// They can't be used when unmarshaling a Compiler and invoking its Compile method.

func Verbose(v bool) CompileOption {
	return func(o *compileOptions) {
		o.verbose = v
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
func CompileTableObject(to *flux.TableObject, now time.Time, opts ...CompileOption) (*Program, error) {
	o := applyOptions(opts...)
	s, err := spec.FromTableObject(to, now)
	if err != nil {
		return nil, err
	}
	if o.verbose {
		log.Println("Query Spec: ", flux.Formatted(s, flux.FmtJSON))
	}
	ps, err := buildPlan(s, o)
	if err != nil {
		return nil, err
	}
	return &Program{
		opts:     o,
		PlanSpec: ps,
	}, nil
}

func buildPlan(spec *flux.Spec, opts *compileOptions) (*plan.Spec, error) {
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
	// Ignore context, it will be provided upon Program Start.
	return Compile(c.Query, c.Now, WithExtern(c.Extern))
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
	// Ignore context, it will be provided upon Program Start.
	return CompileAST(c.AST, now), nil
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
	// Ignore context, it will be provided upon Program Start.
	return CompileTableObject(c.Tables, c.Now)
}

func (*TableObjectCompiler) CompilerType() flux.CompilerType {
	panic("TableObjectCompiler is not associated with a CompilerType")
}

type DependenciesAwareProgram interface {
	SetExecutorDependencies(execute.Dependencies)
	SetLogger(logger *zap.Logger)
}

// Program implements the flux.Program interface.
// It will execute a compiled plan using an executor.
type Program struct {
	Dependencies execute.Dependencies
	Logger       *zap.Logger
	PlanSpec     *plan.Spec

	opts *compileOptions
}

func (p *Program) SetExecutorDependencies(deps execute.Dependencies) {
	p.Dependencies = deps
}

func (p *Program) SetLogger(logger *zap.Logger) {
	p.Logger = logger
}

func (p *Program) Start(ctx context.Context, alloc *memory.Allocator) (flux.Query, error) {
	ctx, cancel := context.WithCancel(ctx)
	results := make(chan flux.Result)
	q := &query{
		results: results,
		alloc:   alloc,
		cancel:  cancel,
		stats: flux.Statistics{
			Metadata: make(flux.Metadata),
		},
	}

	e := execute.NewExecutor(p.Dependencies, p.Logger)
	resultMap, md, err := e.Execute(ctx, p.PlanSpec, q.alloc)
	if err != nil {
		return nil, err
	}

	// There was no error so send the results downstream.
	q.wg.Add(1)
	go p.processResults(ctx, q, resultMap)

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

func (p *AstProgram) Start(ctx context.Context, alloc *memory.Allocator) (flux.Query, error) {
	if p.opts == nil {
		p.opts = defaultOptions()
	}

	if p.Now.IsZero() {
		p.Now = time.Now()
	}

	if p.opts.extern != nil {
		p.Ast.Files = append([]*ast.File{p.opts.extern}, p.Ast.Files...)
	}

	deps, ok := p.Dependencies[dependencies.InterpreterDepsKey]
	if !ok {
		// TODO(Adam): this should be more of a noop dependency package
		return nil, fmt.Errorf("no interpreter dependencies found")
	}
	depsI := deps.(dependencies.Interface)
	ses, scope, err := p.eval(ctx, depsI)
	if err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "error in evaluating AST while starting program")
	}
	s, err := spec.FromEvaluation(ses, p.Now)
	if err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "error in query specification while starting program")
	}
	if p.opts.verbose {
		log.Println("Query Spec: ", flux.Formatted(s, flux.FmtJSON))
	}
	if err := p.updateOpts(scope); err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "error in reading options while starting program")
	}
	ps, err := buildPlan(s, p.opts)
	if err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "error in building plan while starting program")
	}
	p.PlanSpec = ps

	return p.Program.Start(ctx, alloc)
}

func (p *AstProgram) eval(ctx context.Context, deps dependencies.Interface) ([]interpreter.SideEffect, values.Scope, error) {
	s, _ := opentracing.StartSpanFromContext(ctx, "eval")

	sideEffects, scope, err := flux.EvalAST(ctx, deps, p.Ast, flux.SetNowOption(p.Now))
	if err != nil {
		return nil, nil, err
	}

	s.Finish()
	s, _ = opentracing.StartSpanFromContext(ctx, "compile")
	defer s.Finish()

	nowOpt, ok := scope.Lookup(flux.NowOption)
	if !ok {
		return nil, nil, fmt.Errorf("%q option not set", flux.NowOption)
	}

	nowTime, err := nowOpt.Function().Call(ctx, deps, nil)
	if err != nil {
		return nil, nil, err
	}
	p.Now = nowTime.Time().Time()
	return sideEffects, scope, nil
}

func (p *AstProgram) updateOpts(scope values.Scope) error {
	planOpts, err := getPlanOptions(scope)
	if err != nil {
		return err
	}
	p.opts.planOptions.logical = append(p.opts.planOptions.logical, planOpts.logical...)
	p.opts.planOptions.physical = append(p.opts.planOptions.physical, planOpts.physical...)
	return nil
}

func getPlanOptions(scope values.Scope) (planOptions, error) {
	planOpt, ok := scope.Lookup("planner")
	if !ok {
		// No option specified.
		return planOptions{}, nil
	}
	if planOpt.Type().Nature() != semantic.Object {
		return planOptions{}, fmt.Errorf("option 'planner' must be an object, got %s", planOpt.Type().Nature().String())
	}
	var err error
	var wrongKeysErr = fmt.Errorf("the only available field for option 'planner' is 'disable'")
	if planOpt.Object().Len() > 3 {
		return planOptions{}, wrongKeysErr
	}
	planOpt.Object().Range(func(name string, v values.Value) {
		// TODO(affo): we could add 'enable' and 'only'.
		//  In order to do this we should add the possibility to enable rules by name in the planner, first.
		if name != "disable" {
			err = wrongKeysErr
		}
	})
	if err != nil {
		return planOptions{}, err
	}

	la, pa, err := getRuleArraysForKey("disable", planOpt)
	if err != nil {
		return planOptions{}, err
	}

	ls := make([]plan.LogicalOption, la.Array().Len())
	ps := make([]plan.PhysicalOption, pa.Array().Len())
	la.Array().Range(func(i int, v values.Value) {
		ls[i] = plan.RemoveLogicalRule(v.Str())
	})
	pa.Array().Range(func(i int, v values.Value) {
		ps[i] = plan.RemovePhysicalRule(v.Str())
	})
	return planOptions{
		logical:  ls,
		physical: ps,
	}, nil
}

func getRuleArraysForKey(key string, planOpt values.Value) (values.Array, values.Array, error) {
	lv := values.NewArray(semantic.String)
	pv := values.NewArray(semantic.String)
	obj, found := planOpt.Object().Get(key)
	if !found {
		// no rule for this key
		return lv, pv, nil
	}
	if obj.Type().Nature() != semantic.Object {
		return lv, pv, fmt.Errorf("'planner.%s' must be an object, got %s", key, obj.Type().Nature().String())
	}
	var wrongKeysErr = fmt.Errorf("the only available fields for option 'planner.%s' are 'logical' and 'physical'", key)
	var err error
	if obj.Object().Len() > 2 {
		return nil, nil, wrongKeysErr
	}
	obj.Object().Range(func(name string, v values.Value) {
		switch name {
		case "logical":
			if v.Type().Nature() != semantic.Array {
				err = fmt.Errorf("'planner.%s.logical' must be an array, got %s", key, v.Type().Nature().String())
				return
			}
			if v.Array().Type().ElementType().Nature() != semantic.String {
				err = fmt.Errorf("'planner.%s.logical' must contain strings, got %s", key, v.Array().Type().ElementType().Nature().String())
				return
			}
			lv = v.Array()
		case "physical":
			if v.Type().Nature() != semantic.Array {
				err = fmt.Errorf("'planner.%s.physical' must be an array, got %s", key, v.Type().Nature().String())
				return
			}
			if v.Array().Type().ElementType().Nature() != semantic.String {
				err = fmt.Errorf("'planner.%s.physical' must contain strings, got %s", key, v.Array().Type().ElementType().Nature().String())
				return
			}
			pv = v.Array()
		default:
			err = wrongKeysErr
		}
	})
	return lv, pv, err
}
