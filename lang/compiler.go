package lang

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/influxdata/flux"
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

	extern flux.ASTHandle

	planOptions struct {
		logical  []plan.LogicalOption
		physical []plan.PhysicalOption
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
func WithExtern(extern flux.ASTHandle) CompileOption {
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
func Compile(q string, runtime flux.Runtime, now time.Time, opts ...CompileOption) (*AstProgram, error) {
	astPkg, err := runtime.Parse(q)
	if err != nil {
		return nil, err
	}
	return CompileAST(astPkg, runtime, now, opts...), nil
}

// CompileAST evaluates a Flux handle to an AST and produces a flux.Program.
// now parameter must be non-zero, that is the default now time should be set before compiling.
func CompileAST(astPkg flux.ASTHandle, runtime flux.Runtime, now time.Time, opts ...CompileOption) *AstProgram {
	return &AstProgram{
		Program: &Program{
			Runtime: runtime,
			opts:    applyOptions(opts...),
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
	Extern json.RawMessage `json:"extern,omitempty"`
	Query  string          `json:"query"`
}

func (c FluxCompiler) Compile(ctx context.Context, runtime flux.Runtime) (flux.Program, error) {
	// Ignore context, it will be provided upon Program Start.
	if c.Extern != nil {
		hdl, err := runtime.JSONToHandle(c.Extern)
		if err != nil {
			return nil, err
		}
		return Compile(c.Query, runtime, c.Now, WithExtern(hdl))
	}
	return Compile(c.Query, runtime, c.Now)
}

func (c FluxCompiler) CompilerType() flux.CompilerType {
	return FluxCompilerType
}

// ASTCompiler implements Compiler by producing a Program from an AST.
type ASTCompiler struct {
	Extern json.RawMessage `json:"extern,omitempty"`
	AST    json.RawMessage `json:"ast"`
	Now    time.Time
}

func (c ASTCompiler) Compile(ctx context.Context, runtime flux.Runtime) (flux.Program, error) {
	now := c.Now
	if now.IsZero() {
		now = time.Now()
	}
	hdl, err := runtime.JSONToHandle(c.AST)
	if err != nil {
		return nil, err
	}
	if err := hdl.GetError(); err != nil {
		return nil, err
	}

	// Ignore context, it will be provided upon Program Start.
	if c.Extern != nil {
		extHdl, err := runtime.JSONToHandle(c.Extern)
		if err != nil {
			return nil, err
		}
		return CompileAST(hdl, runtime, now, WithExtern(extHdl)), nil
	}
	return CompileAST(hdl, runtime, now), nil
}

func (ASTCompiler) CompilerType() flux.CompilerType {
	return ASTCompilerType
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
	return CompileTableObject(ctx, c.Tables, c.Now)
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
	Runtime  flux.Runtime

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

	e := execute.NewExecutor(p.Logger)
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

	Ast flux.ASTHandle
	Now time.Time
}

func (p *AstProgram) getSpec(ctx context.Context, runtime flux.Runtime, alloc *memory.Allocator) (*flux.Spec, values.Scope, error) {
	if p.opts == nil {
		p.opts = defaultOptions()
	}
	if p.Now.IsZero() {
		p.Now = time.Now()
	}
	if p.opts.extern != nil {
		extern := p.opts.extern
		if err := runtime.MergePackages(extern, p.Ast); err != nil {
			return nil, nil, err
		}
		p.Ast = extern
		p.opts.extern = nil
	}
	// The program must inject execution dependencies to make it available
	// to function calls during the evaluation phase (see `tableFind`).
	deps := ExecutionDependencies{
		Allocator: alloc,
		Logger:    p.Logger,
	}
	ctx = deps.Inject(ctx)
	s, cctx := opentracing.StartSpanFromContext(ctx, "eval")
	sideEffects, scope, err := runtime.Eval(cctx, p.Ast, flux.SetNowOption(p.Now))
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
	sp, scope, err := p.getSpec(ctx, p.Runtime, alloc)
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
	pkg, ok := getPlannerPkg(scope)
	if !ok {
		return nil
	}
	lo, po, err := getPlanOptions(pkg)
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

func getPlannerPkg(scope values.Scope) (values.Package, bool) {
	found := false
	var foundPkg values.Package
	scope.Range(func(k string, v values.Value) {
		if found {
			return
		}
		if pkg, ok := v.(values.Package); ok {
			if pkg.Name() == "planner" {
				found = true
				foundPkg = pkg
			}
		}
	})
	return foundPkg, found
}

func getPlanOptions(plannerPkg values.Package) (plan.LogicalOption, plan.PhysicalOption, error) {
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
	et, err := rules.Type().ElemType()
	if err != nil {
		return nil, err
	}
	if et.Nature() != semantic.String {
		return nil, fmt.Errorf("'planner.%s' must be an array of string, got an array of %s", optionName, et.String())
	}
	noRules := rules.Len()
	rs := make([]string, noRules)
	rules.Range(func(i int, v values.Value) {
		rs[i] = v.Str()
	})
	return rs, nil
}
