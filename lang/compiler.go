package lang

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/values"
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

	planner := plan.PlannerBuilder{}.Build()
	ps, err := planner.Plan(spec)
	if err != nil {
		return nil, err
	}

	return &Program{
		PlanSpec: ps,
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
	planner := plan.PlannerBuilder{}.Build()
	ps, err := planner.Plan(c.Spec)
	if err != nil {
		return nil, err
	}

	return &Program{
		PlanSpec: ps,
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
		return nil, err
	}

	planner := plan.PlannerBuilder{}.Build()
	ps, err := planner.Plan(spec)
	if err != nil {
		return nil, err
	}

	return &Program{PlanSpec: ps}, err
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
}

func (c *TableObjectCompiler) Compile(ctx context.Context) (flux.Program, error) {
	ider := &ider{
		id:     0,
		lookup: make(map[*flux.TableObject]flux.OperationID),
	}
	spec := new(flux.Spec)
	visited := make(map[*flux.TableObject]bool)
	buildSpec(c.Tables, ider, spec, visited)

	specCompiler := &SpecCompiler{
		Spec: spec,
	}
	return specCompiler.Compile(ctx)
}

func (*TableObjectCompiler) CompilerType() flux.CompilerType {
	panic("TableObjectCompiler is not associated with a CompilerType")
}

// TODO(affo): this is duplicate code of the private types in /compile.go to avoid cyclic
//  dependencies between `flux` and `lang`.
type ider struct {
	id     int
	lookup map[*flux.TableObject]flux.OperationID
}

func (i *ider) nextID() int {
	next := i.id
	i.id++
	return next
}

func (i *ider) get(t *flux.TableObject) (flux.OperationID, bool) {
	tableID, ok := i.lookup[t]
	return tableID, ok
}

func (i *ider) set(t *flux.TableObject, id int) flux.OperationID {
	opID := flux.OperationID(fmt.Sprintf("%s%d", t.Kind, id))
	i.lookup[t] = opID
	return opID
}

func (i *ider) ID(t *flux.TableObject) flux.OperationID {
	tableID, ok := i.get(t)
	if !ok {
		tableID = i.set(t, i.nextID())
	}
	return tableID
}

// TODO(affo): duplicate code in /compile.go.
func buildSpec(t *flux.TableObject, ider flux.IDer, spec *flux.Spec, visited map[*flux.TableObject]bool) {
	// Traverse graph upwards to first unvisited node.
	// Note: parents are sorted based on parameter name, so the visit order is consistent.
	t.Parents.Range(func(i int, v values.Value) {
		p := v.(*flux.TableObject)
		if !visited[p] {
			// rescurse up parents
			buildSpec(p, ider, spec, visited)
		}
	})

	// Assign ID to table object after visiting all ancestors.
	tableID := ider.ID(t)

	// Link table object to all parents after assigning ID.
	t.Parents.Range(func(i int, v values.Value) {
		p := v.(*flux.TableObject)
		spec.Edges = append(spec.Edges, flux.Edge{
			Parent: ider.ID(p),
			Child:  tableID,
		})
	})

	visited[t] = true
	spec.Operations = append(spec.Operations, t.Operation(ider))
}

// Program implements the flux.Program interface.
// It will execute a compiled plan using an executor.
type Program struct {
	Dependencies execute.Dependencies
	PlanSpec     *plan.Spec
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
	q.wg.Add(1)
	go p.run(ctx, q)
	return q, nil
}

func (p *Program) run(ctx context.Context, q *query) {
	defer q.wg.Done()
	defer close(q.results)

	e := execute.NewExecutor(p.Dependencies, nil)
	results, md, err := e.Execute(ctx, p.PlanSpec, q.alloc)
	if err != nil {
		q.err = err
		return
	}

	// Begin reading from the metadata channel.
	q.wg.Add(1)
	go p.readMetadata(q, md)

	// There was no error so send the results downstream.
	for _, res := range results {
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
