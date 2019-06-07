package lang

import (
	"context"
	"log"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/spec"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
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
func Compile(q string, now time.Time, opts ...CompileOption) (*Program, error) {
	astPkg, err := flux.Parse(q)
	if err != nil {
		return nil, err
	}
	return CompileAST(astPkg, now, opts...)
}

// CompileAST evaluates a Flux AST and produces a flux.Program.
// now parameter must be non-zero, that is the default now time should be set before compiling.
func CompileAST(astPkg *ast.Package, now time.Time, opts ...CompileOption) (*Program, error) {
	semPkg, err := semantic.New(astPkg)
	if err != nil {
		return nil, err
	}
	return &Program{
		ReadyToStartProgram: &ReadyToStartProgram{
			opts: applyOptions(opts...),
		},
		Pkg: semPkg,
		Now: now,
	}, nil
}

// CompileTableObject evaluates a TableObject and produces a flux.Program.
// now parameter must be non-zero, that is the default now time should be set before compiling.
func CompileTableObject(ctx context.Context, to *flux.TableObject, now time.Time, opts ...CompileOption) (*ReadyToStartProgram, error) {
	o := applyOptions(opts...)
	sp := spec.FromTableObject(to, now)
	if o.verbose {
		log.Println("Query Spec: ", flux.Formatted(sp, flux.FmtJSON))
	}
	s, _ := opentracing.StartSpanFromContext(ctx, "plan")
	defer s.Finish()
	ps, err := buildPlan(sp, o)
	if err != nil {
		return nil, err
	}
	return &ReadyToStartProgram{
		opts:     o,
		PlanSpec: ps,
	}, nil
}

func WalkIR(astPkg *ast.Package, f func(o *flux.Operation) error) error {
	if sp, err := spec.FromAST(astPkg, time.Now()); err != nil {
		return err
	} else {
		return sp.Walk(f)
	}
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
	Query string `json:"query"`
}

func (c FluxCompiler) Compile(ctx context.Context) (flux.Program, error) {
	s, _ := opentracing.StartSpanFromContext(ctx, "compile")
	defer s.Finish()
	return Compile(c.Query, time.Now())
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
	s, _ := opentracing.StartSpanFromContext(ctx, "compile")
	defer s.Finish()
	return CompileAST(c.AST, now)
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
	return CompileTableObject(ctx, c.Tables, c.Now)
}

func (*TableObjectCompiler) CompilerType() flux.CompilerType {
	panic("TableObjectCompiler is not associated with a CompilerType")
}

type DependenciesAwareProgram interface {
	SetExecutorDependencies(execute.Dependencies)
	SetLogger(logger *zap.Logger)
}

// ReadyToStartProgram implements the flux.Program interface.
// It executes a compiled plan using an executor.
type ReadyToStartProgram struct {
	Dependencies execute.Dependencies
	Logger       *zap.Logger
	PlanSpec     *plan.Spec

	opts *compileOptions
}

func (p *ReadyToStartProgram) SetExecutorDependencies(deps execute.Dependencies) {
	p.Dependencies = deps
}

func (p *ReadyToStartProgram) SetLogger(logger *zap.Logger) {
	p.Logger = logger
}

func (p *ReadyToStartProgram) Start(ctx context.Context, alloc *memory.Allocator) (flux.Query, error) {
	s, _ := opentracing.StartSpanFromContext(ctx, "start")
	defer s.Finish()
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

func (p *ReadyToStartProgram) processResults(ctx context.Context, q *query, resultMap map[string]flux.Result) {
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

func (p *ReadyToStartProgram) readMetadata(q *query, metaCh <-chan flux.Metadata) {
	defer q.wg.Done()
	for md := range metaCh {
		q.stats.Metadata.AddAll(md)
	}
}

// Program is a flux.Program that is not ready to start, but contains a semantic graph.
// Upon start it evaluates the semantic graph and starts the real execution.
// As such, evaluation errors are returned by Start.
type Program struct {
	*ReadyToStartProgram

	Pkg *semantic.Package
	Now time.Time
}

func (p *Program) Start(ctx context.Context, alloc *memory.Allocator) (flux.Query, error) {
	if p.opts == nil {
		p.opts = defaultOptions()
	}
	if p.Now.IsZero() {
		p.Now = time.Now()
	}
	s, _ := opentracing.StartSpanFromContext(ctx, "eval")
	sideEffects, _, actualNow, err := flux.EvalWithNow(p.Pkg, p.Now)
	if err != nil {
		s.Finish()
		return nil, err
	}
	sp, err := spec.FromSideEffects(sideEffects, actualNow)
	s.Finish()
	if err != nil {
		return nil, errors.Wrap(err, "error in evaluating AST while starting program")
	}
	if p.opts.verbose {
		log.Println("Query Spec: ", flux.Formatted(sp, flux.FmtJSON))
	}
	s, _ = opentracing.StartSpanFromContext(ctx, "plan")
	ps, err := buildPlan(sp, p.opts)
	s.Finish()
	if err != nil {
		return nil, errors.Wrap(err, "error in building plan while starting program")
	}
	p.PlanSpec = ps
	return p.ReadyToStartProgram.Start(ctx, alloc)
}
