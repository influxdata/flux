package lang

import (
	"context"
	"log"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/spec"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/values"
	"github.com/pkg/errors"
)

// CreateContextAwareValue is a function that creates a value with the provided flux.ExecutionContext.
type CreateContextAwareValue func(*flux.ExecutionContext) values.Value

var executionAwareValues = make(map[string]CreateContextAwareValue)

// RegisterContextAwareValue registers a function to create a context aware value.
func RegisterContextAwareValue(name string, v CreateContextAwareValue) bool {
	_, replace := executionAwareValues[name]
	executionAwareValues[name] = v
	return replace
}

// BindContextAwareValues binds the result of every CreateContextAwareValue function to the registered name
// in the provided scope using the provided execution context. The scope is intended to be the prelude, but it
// could be any other. Registered bindings could shadow others in the provided scope, if they were present.
func BindContextAwareValues(prelude interpreter.Scope, ec *flux.ExecutionContext) {
	for name, f := range executionAwareValues {
		prelude.Set(name, f(ec))
	}
}

// TableObjectCompiler compiles a TableObject into an executable flux.Program.
// It is not added to CompilerMappings and it is not serializable, because
// it is impossible to use it outside of the context of an ongoing execution of a program.
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

// CompileTableObject evaluates a TableObject and produces a flux.Program.
// `now` parameter must be non-zero, that is the default now time should be set before compiling.
func CompileTableObject(to *flux.TableObject, now time.Time, opts ...CompileOption) (*Program, error) {
	o := applyOptions(opts...)
	s := spec.FromTableObject(to, now)
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

// Program implements the flux.Program interface.
// It will execute a compiled plan using an executor.
type Program struct {
	PlanSpec *plan.Spec
	opts     *compileOptions
}

func (p *Program) Start(ec *flux.ExecutionContext) (flux.Query, error) {
	ctx, cancel := context.WithCancel(ec.Context)
	results := make(chan flux.Result)
	q := &query{
		results: results,
		alloc:   ec.Allocator,
		cancel:  cancel,
		stats: flux.Statistics{
			Metadata: make(flux.Metadata),
		},
	}

	e := execute.NewExecutor(ec.Dependencies, ec.Logger)
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

func (p *AstProgram) Start(ec *flux.ExecutionContext) (flux.Query, error) {
	if p.opts == nil {
		p.opts = defaultOptions()
	}

	if p.Now.IsZero() {
		p.Now = time.Now()
	}
	// override execution context Now with the one obtained during compilation.
	ec.Now = p.Now
	s, err := spec.FromAST(ec.Context, p.Ast, p.Now, func(scope interpreter.Scope) {
		BindContextAwareValues(scope, ec)
	})
	if err != nil {
		return nil, errors.Wrap(err, "error in evaluating AST while starting program")
	}
	if p.opts.verbose {
		log.Println("Query Spec: ", flux.Formatted(s, flux.FmtJSON))
	}
	ps, err := buildPlan(s, p.opts)
	if err != nil {
		return nil, errors.Wrap(err, "error in building plan while starting program")
	}
	p.PlanSpec = ps

	return p.Program.Start(ec)
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
