# Design overview

This document contains an overview of the components of the Flux engine ---a.k.a. Flux Virtual Machine (FVM)---
that cooperate in the process of executing a Flux query.

In a nutshell, the engine is responsible for executing a query. A query expresses data manipulation
pipelines through common operations, such as additions, subtractions, and conditionals, and data manipulation
operations, i.e. data extraction from sources, and transformations ---outputs to sinks are transformations with a side
effect. The results of the query are obtained from the side effects of data manipulation.
A query can be represented in various languages.
The engine is equipped with one executors per language. An executor can translate a
representation of a query in some language into another target language and delegate the execution to
the executor for the target language.
The translation process is carried out by compilers.
Once a proper target language has been reached (possibly thanks to multiple compilation processes),
the final representation of the query can be interpreted to produce results.
During interpretation, the engine is responsible to provide the actual implementation of the data manipulation operations
to the interpreter. Different engines can implement sources and transformations in different ways.

We now provide the high-level interfaces for the components described above:

The engine provides methods for registering sources and transformations, and for binding builtin values used for interpretation.
Sources and transformations constitute the actual implementation of the data manipulation operations of some function values in the scope.

```go
type CreateSource = fn (ExecutionContext) Source
type CreateTransformation = fn (ExecutionContext) Transformation
type CompilerType interface {
    From() Language
    To() Language
}

type Engine interface {
    BindValue(id string, Value)
    Prelude() Scope
    
    RegisterSource(id string, CreateSource)
    RegisterTransformation(id string, CreateTransformation)
    
    Query(Representation) (Query, error)
}

type Scope interface {} // ...
type Source interface {} // ...
type Transformation interface {} // ...
type Value interface {} // ...
```

A language is used to describe a query. It consists of words whose letters are taken from an alphabet.
Letters can be of any type and complexity, from characters, to nodes in an Abstract Syntax Tree (AST). Various languages
can be used to represent a query.

```go
type Language interface {
    Name() string
}
```

A query defines a unit of work for the engine. It comprises one or more pipelines.
It contains its representation, its results, and execution statistics.

```go
type Query interface {
    Results() Results
    Stats() map[string]interface{}
    Representation() Representation
}

type Representation interface {
    Visitor // a representation can be visited for compilation/execution.
    Lang() Language // a representation makes sense in a certain language.
    String() string // a visual representation of the query.
}

// The results for a Query.
type Results interface {
    ForEach(func (Result))
}

type Result interface {
    Name() string
    // ... methods to iterate other the contents of the result
}
```

The executor executes a query expressed in some language and fills its results, provided an execution context.
There can be multiple executors, one for each language supported by the engine.
The engine is responsible for providing the executor with an
execution context that embeds everything that is necessary for the execution of the query, included the registered values,
sources and transformations. 

```go
type Executor interface {
    Execute(Query, ExecutionContext) error // fills query results.
    Type() Language
}

// The ExecutionContext contains any information necessary for properly executing a query.
type ExecutionContext interface {
    Bindings() Scope
    GetSource(id string) Source
    GetTransformation(id string) Transformation
    
    // ... examples of other components of the execution context.
    Context() context.Context
    MemoryAllocator() Allocator
    Logger() Logger
}

// Scope is a set of bindings from identifiers to values.
type Scope = map[string]Value
```
   
A compiler translates a query representation in a language into another target language.

```go
type Compiler interface {
    Compile(Representation) (Representation, err)
    Type() CompilerType
}
```

As an example, let `e` be an engine that supports `String`, `AST`, and `SemanticGraph` queries:

```go
func (e *Engine) Query(r Representation) (Query, error) {
    var e Executor
    switch l := r.Language() {
        case "String":
            e = &StringExecutor{...}
        case "AST":
            e = &ASTExecutor{...}
        case "SemanticGraph":
            // the interpreter is an Executor!
            e = interpreter.NewInterpreter()
        default:
            return nil, fmt.Errorf("unknown language: %v", l)
    }
    
    q := ... // new query
    ec := ... // create execution context
    err := e.Execute(q, ec)
    return q, err
}
```

The `StringExecutor` does not actually know how to execute a script, but it uses a compiler to translate the script into its
AST representation and delegate to an `ASTExecutor` in turn:

```go
func (e *StringExecutor) Execute(q Query, ec ExecutionContext) error {
    if ln := q.Representation().Language().Name(); ln != "String" {
        return fmt.Errorf("String executor cannot execute %s", ln)
    }
    
    astc := &ASTCompiler{...}
    newR, err := astc.Compile(q.Representation())
    if err != nil {
        return fmt.Errorf("cannot compile")
    }
    q.SetRepresentation(newR) // substitute Representation
    aste, err:= &ASTExecutor{...}
    return se.Execute(q, ec) // forward the execution context!
}
```

The `ASTExecutor` does not actually know how to execute an AST, but it uses a compiler to translate the AST into a
semantic graph and delegate to a `SemanticExecutor` in turn:

```go
func (e *ASTExecutor) Execute(q Query, ec ExecutionContext) error {
    if ln := q.Representation().Language().Name(); ln != "AST" {
        return fmt.Errorf("AST executor cannot execute %s", ln)
    }
    
    sgc := &SemanticGraphCompiler{...}
    newR, err := sgc.Compile(q.Representation())
    if err != nil {
        return fmt.Errorf("cannot compile")
    }
    q.SetRepresentation(newR) // substitute new Representation
    // the interpreter is an Executor!
    itrp := interpreter.NewInterpreter()
    return itrp.Execute(q, ec) // forward the execution context!
}
```

The `Interpreter` is responsible for executing semantic graphs:

```go
func (i *Interpreter) Execute(q Query, ec ExecutionContext) error {
    if ln := q.Representation().Language().Name(); ln != "SemanticGraph" {
        return fmt.Errorf("Interpreter cannot execute %s", ln)
    }
    
    itrp := interpreter.NewInterpreter()
    universe := ec.Bindings()
    semPkg := q.Representation().(*semantic.Package)
    // pass the execution context!
    results, err := itrp.Eval(semPkg, universe, ec)
    if err != nil {
        return err
    }
    // ... add results the query (see below).
    return nil
}
```

While evaluating, the interpreter updates the scope of execution by visiting the semantic graph. It performs
additions, subtractions, comparisons, etc., variable assignments, imports, and data manipulation functions. 
The interpreter can further compile the semantic graph to an IR and to an
optimized IR through planning, as explained in [FVM documentation](VirtualMachine.md). 
When running data manipulation operations, the executor can access their actual implementation through the `ExecutionContext`.

When interpreting the function calls that generate a pipeline for data manipulation, the interpreter generates a specification
for the data manipulation pipeline and executes it when the results are needed ---this could be for `yield`ing results,
or for a `tableFind` call. Upon pipeline execution, the interpreter compiles the specification for the pipeline
into the FVM IR and, later, to an optimized FVM IR ---this is planning. The planner is, indeed, a compiler. It can make decisions
based on the specific sources that will be used to start the pipeline. For instance, the executor (the one that runs the
compilation to an optimized IR) obtains an `InfluxDBSource` by invoking `ExecutionContext.GetSource(<id>)`, where `<id>`
is the identifier in the IR ---i.e. the current query representation--- that matches that operation, 
and it instantiates a planner that "understands" `InfluxDBSource` capabilities and hints:

```go
// InfluxDBSource contains an instance of `Storage`.
type InfluxDBSource interface {
    Storage
    Run(ExecutionContext) error
}

// Storage provides an interface to the storage layer.
type Storage interface {
    // Read gets data from the underlying storage system.
    Read(context.Context, []Predicate, TimeRange, Grouping) (TableIterator, error)
    // Capabilities exposes the capabilities of the storage interface.
    Capabilities() []Capability
    // Hints provides hints about the characteristics of the data.
    Hints(context.Context, []Predicate, TimeRange, Grouping) Hints
}

// Predicate filters data.
type Predicate interface {}

// TimeRange is the beginning time and ending time
type TimeRange interface {
    Begin() int64
    End() int64
}

// Grouping are key groups
type Grouping interface {
    Keys() []string
}

// Hints provide insight into the size and shape of the data that would likely be returned
// from a storage read operation.
type Hints interface {
    Cardinality() // Count tag values
    ByteSize()   int64
    Blocks()     int64
}

// Capability represents a single capability of a storage interface.
type Capability interface{
    Name() string
}
```

If we are in InfluxDB's codebase, then the launcher will register `InfluxDBSource` in the engine at startup.
That source exposes the capabilities of the storage and directly uses its internal codebase to obtain data on `Run`.