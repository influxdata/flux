package lang

import (
	"context"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/internal/spec"
	"github.com/influxdata/flux/plan"
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

func WalkIR(astPkg *ast.Package, f func(o *flux.Operation) error) error {
	if s, err := spec.FromAST(context.Background(), astPkg, time.Now()); err != nil {
		return err
	} else {
		return s.Walk(f)
	}
}

// FluxCompiler compiles a Flux script into a program.
type FluxCompiler struct {
	Query string `json:"query"`
}

func (c FluxCompiler) Compile(ctx context.Context) (flux.Program, error) {
	// Ignore context, it will be provided upon Program Start.
	return Compile(c.Query, time.Now())
}

func (c FluxCompiler) CompilerType() flux.CompilerType {
	return FluxCompilerType
}

// ASTCompiler implements Compiler by producing a program from an AST.
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
