# Compilation and Execution

This document contains an overview of the components involved in the process of compiling and executing a Flux query.

In a nutshell, the process is separated in compilation and execution. The Flux compiler is responsible to turn the initial
representation of a query into an Intermediate Representation (IR) that makes sense for the execution step.
The Flux Virtual Machine ([FVM](VirtualMachine.md)) is in charge of interpreting the IR and execute it to output its results.
In figure:

![chain](compilation_execution.png)

 - The initial representation of the query can be either a raw string or an already parsed Abstract Syntax Tree.
 - The compilation process chains multiple compilation steps.
 `Compiler`s turn a query representation in a language into another one, until they reach the required IR.
 - The `Interpreter` takes the IR and interprets it.
 Some operations are trivial ---e.g., `1 + 2` or `obj.b` or `arr[0]`--- and do not require any further step to be executed.
 For data manipulation pipelines, instead, the `Intepreter` needs more.
 - The `Interpreter` does not know how to execute data pipelines.
 So, it takes their representation ---i.e., the `Spec`-- and optimizes it by applying subsequent _transformations_,
 in order to delegates execution to the `Engine`.
 Those transformations are logical and physical planning. We consider these transformations as further compilation
 steps, in that they turn a query representation into another.
 - Finally, the `Engine` executes the optimized representation and outputs the results.

Wrapping up, the actors that play in this process are `Compiler`s ---that turn a representation in some language into another one---
and `Executor`s ---that execute a representation and produce results.
Both the `Interpreter` and the `Engine` are `Executor`s:
 - The `Engine` actually executes data manipulation pipelines and produces results;
 - The `Interpreter` directly executes trivial operations and delegates pipeline execution to the `Engine`.
 In its essence, it executes a higher level representation for the query than the `Engine`'s.
 
In order to carry out execution, `Executor` need to know some information on a per-query basis
(for instance, its memory limits, the request context, etc.).
This information is embedded into the `ExecutionContext`.
 
These are the interfaces:

```go
// Language is a language used to express a Flux query.
type Language interface {
    LanguageName() String
}

// Representation contains the actual content of a Flux query expressed in some Language.
type Representation interface {
    Lang() Language
}

// Results are the results of a Query.
type Results interface {}
// Statistics are the statistics of a Query execution.
type Statistics interface {}

// Query is an executed Flux query.
// It provides its representation, results, error, and statistics.
type Query interface {
    Results() Results
    Error() error
    Stats() Statistics
    Representation() Representation
}

// The ExecutionContext contains the information necessary for properly executing a query.
type ExecutionContext interface { 
    // ... examples of components for the execution context.
    Context() context.Context
    MemoryAllocator() Allocator // (Allocator definition is not relevant here)
    Logger() Logger // (Logger definition is not relevant here)
}

// Executor executes a query Representation and returns a Query given an ExecutionContext.
type Executor interface {
    Execute(ExecutionContext, Representation) (Query, error)
    // each Executor can execute query representations in a target Language.
    ExecutorType() Language
}

// Compiler turns a Representation in a Language into another one.
type Compiler interface {
    Compile(Representation) (Representation, error)
    // each Compiler has a source and a target Language.
    CompilerType() (Language, Language)
}
```

Interfaces are decoupled thanks to referencing a generic query `Representation`, in order to favor `Executor`s and `Compiler`s composability:
 - each compilation step is a black box to the others. There is no dependency among compilation steps;
 - changing a `Compiler` implementation solely impacts a single compilation step;
 - a compilation step can be split in two (or more) compilation steps by passing through an intermediate representation.
 This comes in handy when we need to separate a complex process into more, simpler ones;
 - adding a compilation step is as easy as increasing the compilation chain by one. This comes in handy because the IR accepted by the FVM is,
 for now, a blur line: at the moment, it coincides with the semantic graph representation, but nothing prevents us to add more compilation steps in the future.
 - The process that the FVM runs is the same as the Flux compiler does, so it gets the same benefits as above.
 Indeed, the `Interpreter` passes through compilation steps and provides a lower-level IR to the `Engine` ---another `Executor`.

The contract is now moved from the interfaces to the concrete `Representation`s and `Language`s.
Suppose we have a compilation step from language `A` to `B`, but we change `B` to `B'`.
Then either the compiler for `A` changes, or we add a step of compilation from `B` to `B'`.
A bigger problem arises if `B` is the language accepted by the downstream `Executor`.
In that case either we change the `Executor`'s implementation, or, equivalently, we implement a new `Executor` that targets `B'` and swap the implementations.

The delegation of execution from the `Interpreter` to the `Engine` (through compilation steps) is crucial, in that
it allows the `Interpreter` to trigger execution in intermediate steps of interpretation for dynamic queries:

```
t = from(...)
    |> filter(...)
    |> group(...)
    |> tableFind(fn: (key) => ...)

record = t |> getRecord(idx: 0)

from(...)
 |> filter(fn: (r) => r._value = record._value * 2)
```

This query requires the `Interpreter` to compute the results of `from(...) |> filter(...) |> group(...)` to extract the table `t`,
in order to make it available to the rest of the computation.
When the `Interpreter` encounters the `tableFind` call, it must delegate the execution to the `Engine` to get those results.
Once obtained them, it can continue with its evaluation and, finally, delegate the execution to the `Engine`
for the final results.