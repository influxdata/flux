package lang

import (
	"context"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
)

const (
	FluxCompilerType = "flux"
	SpecCompilerType = "spec"
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
	return mappings.Add(SpecCompilerType, func() flux.Compiler {
		return new(SpecCompiler)
	})
}

// FluxCompiler compiles a Flux script into a spec.
type FluxCompiler struct {
	Query string `json:"query"`
}

func (c FluxCompiler) Compile(ctx context.Context) (flux.Program, error) {
	spec, err := flux.Compile(ctx, c.Query, time.Now())
	if err != nil {
		return nil, err
	}

	planner := (&plan.PlannerBuilder{}).Build()
	ps, err := planner.Plan(spec)
	if err != nil {
		return nil, err
	}

	return Program{
		ps: ps,
	}, err
}

func (c FluxCompiler) CompilerType() flux.CompilerType {
	return FluxCompilerType
}

// SpecCompiler implements Compiler by returning a known spec.
type SpecCompiler struct {
	Spec *flux.Spec `json:"spec"`
}

func (c SpecCompiler) Compile(ctx context.Context) (flux.Program, error) {
	planner := (&plan.PlannerBuilder{}).Build()
	ps, err := planner.Plan(c.Spec)
	if err != nil {
		return nil, err
	}

	return Program{
		ps: ps,
	}, err
}

func (c SpecCompiler) CompilerType() flux.CompilerType {
	return SpecCompilerType
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

	spec, err := flux.CompileAST(ctx, c.AST, now)
	if err != nil {
		return Program{}, err
	}

	planner := (&plan.PlannerBuilder{}).Build()
	ps, err := planner.Plan(spec)
	if err != nil {
		return Program{}, err
	}

	return Program{ps: ps}, err
}

func (ASTCompiler) CompilerType() flux.CompilerType {
	return ASTCompilerType
}

// PrependFile prepends a file onto the compiler's list of package files.
func (c *ASTCompiler) PrependFile(file *ast.File) {
	c.AST.Files = append([]*ast.File{file}, c.AST.Files...)
}

// Program implements the flux.Program interface
type Program struct {
	deps execute.Dependencies
	ps   *plan.Spec
}

func (p Program) Start(ctx context.Context, allocator *memory.Allocator) (flux.Query, error) {
	e := execute.NewExecutor(p.deps, nil)
	results, _, err := e.Execute(ctx, p.ps, allocator)
	if err != nil {
		return nil, err
	}

	ch := make(chan flux.Result)
	go func() {
		for _, r := range results {
			ch <- r
		}
		close(ch)
	}()

	return &Query{
		ch: ch,
	}, nil
}
