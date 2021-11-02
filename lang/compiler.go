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
	"github.com/influxdata/flux/internal/jaeger"
	"github.com/influxdata/flux/internal/spec"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/metadata"
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
	ps, err := buildPlan(ctx, s, o)
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

	ps, err := pb.Build().Plan(ctx, spec)
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

func wrapFileJSONInPkg(bs []byte) []byte {
	return []byte(fmt.Sprintf(
		`{"type":"Package","package":"main","files":[%s]}`,
		string(bs)))
}

func IsNonNullJSON(bs json.RawMessage) bool {
	if len(bs) == 0 {
		return false
	}
	if len(bs) == 4 && string(bs) == "null" {
		return false
	}
	return true
}

func (c FluxCompiler) Compile(ctx context.Context, runtime flux.Runtime) (flux.Program, error) {
	query := c.Query

	// Ignore context, it will be provided upon Program Start.
	if IsNonNullJSON(c.Extern) {
		hdl, err := runtime.JSONToHandle(wrapFileJSONInPkg(c.Extern))
		if err != nil {
			return nil, errors.Wrap(err, codes.Inherit, "extern json parse error")
		}
		return Compile(query, runtime, c.Now, WithExtern(hdl))
	}
	return Compile(query, runtime, c.Now)
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
	if IsNonNullJSON(c.Extern) {
		extHdl, err := runtime.JSONToHandle(wrapFileJSONInPkg(c.Extern))
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
		ctx:     cctx,
		results: results,
		alloc:   alloc,
		span:    s,
		cancel:  cancel,
		stats: flux.Statistics{
			Metadata: make(metadata.Metadata),
		},
	}

	if execute.HaveExecutionDependencies(ctx) {
		deps := execute.GetExecutionDependencies(ctx)
		q.stats.Metadata.AddAll(deps.Metadata)
	}

	if traceID, sampled, found := jaeger.InfoFromSpan(s); found {
		q.stats.Metadata.Add("tracing/id", traceID)
		q.stats.Metadata.Add("tracing/sampled", sampled)
	}

	q.stats.Metadata.Add("flux/query-plan",
		fmt.Sprintf("%v", plan.Formatted(p.PlanSpec, plan.WithDetails())))

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

func (p *Program) readMetadata(q *query, metaCh <-chan metadata.Metadata) {
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
	// A list of profilers that are profiling this query
	Profilers []execute.Profiler
	// The operator profiler that is profiling this query, if any.
	// Note this operator profiler is also cached in the Profilers array.
	tfProfiler *execute.OperatorProfiler
}

// Prepare the Ast for semantic analysis
func (p *AstProgram) GetAst() (flux.ASTHandle, error) {
	if p.Now.IsZero() {
		p.Now = time.Now()
	}
	if p.opts == nil {
		p.opts = defaultOptions()
	}
	if p.opts.extern != nil {
		extern := p.opts.extern
		if err := p.Runtime.MergePackages(extern, p.Ast); err != nil {
			return nil, err
		}
		p.Ast = extern
		p.opts.extern = nil
	}
	return p.Ast, nil
}

// The ExecOptsConfig structure implements the interpreter.ExecOptsConfig
// interface, which the interpreter uses to configure options relevant to the
// execution engine. The interpreter is able to invoke the execution engine via
// tableFind and others, and therefore must be able to install these options
// into the execution dependency state. We use an interface to break the import
// cycle implied by accessing the execution module from the interpreter.
type ExecOptsConfig struct {
}

func (eoc *ExecOptsConfig) ConfigureProfiler(ctx context.Context, profilerNames []string) {
	var tfProfiler *execute.OperatorProfiler
	dedupeMap := make(map[string]bool)
	profilers := make([]execute.Profiler, 0)
	for _, profilerName := range profilerNames {
		if createProfilerFn, exists := execute.AllProfilers[profilerName]; !exists {
			// profiler does not exist
			continue
		} else {
			if _, exists := dedupeMap[profilerName]; exists {
				// Ignore duplicates
				continue
			}
			dedupeMap[profilerName] = true
			profiler := createProfilerFn()
			if tfp, ok := profiler.(*execute.OperatorProfiler); ok {
				// The operator profiler needs to be in the context so transformations
				// and data sources can easily locate it when creating spans.
				// We cache the operator profiler here in addition to the Profilers
				// array to avoid the array look-up.

				tfProfiler = tfp
			}
			profilers = append(profilers, profiler)
		}
	}

	if execute.HaveExecutionDependencies(ctx) {
		deps := execute.GetExecutionDependencies(ctx)
		deps.ExecutionOptions.OperatorProfiler = tfProfiler
		deps.ExecutionOptions.Profilers = profilers
		deps.Inject(ctx)
	}
}

func (eoc *ExecOptsConfig) ConfigureNow(ctx context.Context, now time.Time) {
	// Stash in the execution dependencies. The deps use a pointer and we
	// overwrite the dest of the pointer. Overwritng the pointer would have no
	// effect as context changes are passed down only.
	deps := execute.GetExecutionDependencies(ctx)
	*deps.Now = now
	deps.Inject(ctx)
}

func (p *AstProgram) getSpec(ctx context.Context, alloc *memory.Allocator) (*flux.Spec, values.Scope, error) {
	ast, astErr := p.GetAst()
	if astErr != nil {
		return nil, nil, astErr
	}

	s, cctx := opentracing.StartSpanFromContext(ctx, "eval")

	// Set the now option to our own default and capture the option itself
	// to allow us to find it after the run. A user might overwrite the
	// now parameter with their own thing so we don't want to allow for
	// that interference. If `option now` is used to overwrite this,
	// the inner value pointed to by the option will be modified.
	// TODO(jsternberg): Personal note, I don't like how now interacts with
	// the runtime and flux code in so many places. We should evaluate how
	// now is used and see if we can improve how now interacts with the system.
	var nowOpt values.Value
	sideEffects, scope, err := p.Runtime.Eval(cctx, ast, &ExecOptsConfig{},
		flux.SetNowOption(p.Now),
		func(r flux.Runtime, scope values.Scope) {
			nowOpt, _ = scope.Lookup(interpreter.NowOption)
			if _, ok := nowOpt.(*values.Option); !ok {
				panic("now must be an option")
			}
		},
	)
	if err != nil {
		return nil, nil, err
	}
	s.Finish()

	s, cctx = opentracing.StartSpanFromContext(ctx, "compile")
	defer s.Finish()
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
	// The program must inject execution dependencies to make it available to
	// function calls during the evaluation phase (see `tableFind`).
	deps := execute.NewExecutionDependencies(alloc, &p.Now, p.Logger)
	ctx = deps.Inject(ctx)
	nextPlanNodeID := new(int)
	ctx = context.WithValue(ctx, plan.NextPlanNodeIDKey, nextPlanNodeID)

	// Evaluation.
	sp, scope, err := p.getSpec(ctx, alloc)
	if err != nil {
		return nil, err
	}

	// Planning.
	s, cctx := opentracing.StartSpanFromContext(ctx, "plan")
	if p.opts.verbose {
		log.Println("Query Spec: ", flux.Formatted(sp, flux.FmtJSON))
	}
	if err := p.updateOpts(scope); err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "error in reading options while starting program")
	}
	if err := p.updateProfilers(ctx, scope); err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "error in reading profiler settings while starting program")
	}
	ps, err := buildPlan(cctx, sp, p.opts)
	if err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "error in building plan while starting program")
	}
	p.PlanSpec = ps
	s.Finish()

	// Execution.
	s, cctx = opentracing.StartSpanFromContext(ctx, "start-program")
	defer s.Finish()
	return p.Program.Start(cctx, alloc)
}

func (p *AstProgram) updateProfilers(ctx context.Context, scope values.Scope) error {
	if execute.HaveExecutionDependencies(ctx) {
		deps := execute.GetExecutionDependencies(ctx)
		p.tfProfiler = deps.ExecutionOptions.OperatorProfiler
		p.Profilers = deps.ExecutionOptions.Profilers
	}
	return nil
}

func (p *AstProgram) updateOpts(scope values.Scope) error {
	pkg, ok := getPackageFromScope("planner", scope)
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

func getPackageFromScope(pkgName string, scope values.Scope) (values.Package, bool) {
	found := false
	var foundPkg values.Package
	scope.Range(func(k string, v values.Value) {
		if found {
			return
		}
		if pkg, ok := v.(values.Package); ok {
			if pkg.Name() == pkgName {
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

	ls, err := getOptionValues(plannerPkg.Object(), "disableLogicalRules")
	if err != nil {
		return nil, nil, err
	}
	ps, err := getOptionValues(plannerPkg.Object(), "disablePhysicalRules")
	if err != nil {
		return nil, nil, err
	}
	return plan.RemoveLogicalRules(ls...), plan.RemovePhysicalRules(ps...), nil
}

func getOptionValues(pkg values.Object, optionName string) ([]string, error) {
	value, ok := pkg.Get(optionName)
	if !ok {
		// No value in package.
		return []string{}, nil
	}

	rules := value.Array()
	noRules := rules.Len()
	rs := make([]string, noRules)
	rules.Range(func(i int, v values.Value) {
		rs[i] = v.Str()
	})
	return rs, nil
}
