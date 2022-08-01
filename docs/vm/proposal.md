## Summary of Existing Problems

I will go into some broad categories where the existing interpreter and compiler cause issues.
This will just be a summary of ideas that I will go into further detail on.

### Performance

Performance is a broad category and likely the most obvious.
The performance problems are most heavily felt by the compiler, but some performance problems from the compiler are caused by the interpreter sending poor information to the compiler.
The performance mostly affects native flux code in tight loops such as the ones invoked by `map()` and `filter()`.
The current performance of the functions invoked by these causes substantial problems because the current speed effectively makes these functions difficult to use.
Poor performance has many knock-on effects that follow.

With poor performance, we close off the option to use native flux functions for additional language functionality because it would negatively impact the end-user experience.
For example, at the moment, we would never consider changing derivative or moving average to use a more general function.
This is because it would drastically affect the performance of these functions.
This adds up over time for developer time, complexity, and inevitably correctness as the transformations require a non-trivial amount of work to implement.

Vectorization will likely alleviate some of these issues, but there will be many places where vectorization will be difficult that make a general purpose method a larger requirement.
At the same time, the vectorization is implemented in a similar ad-hoc manner as most of the compiler.
The ad-hoc nature of the compiler makes it more difficult to understand, harder to debug, and less consistent because we now have to consider whether code is going through the vectorized path or not.
Refactoring the compiler should complement the vectorization work by adding a more formal framework for the compiler and for optimizations.
This is in contrast to treating vectorization as a specific method of optimization that partially lives outside the existing compiler.

### Complexity

The current architecture is very complex.
There are many phases and each of these phases can be linked together in ways that are sometimes difficult to understand.
There is code that runs when the interpreter executes, code that gets run by the executor afterwards, and code that gets executed in a mini-engine by `tableFind()`.
There is an interpreter and a compiler which get used at different points in the execution.
These two have overlapping functionality with sometimes different implementations.
Types get exposed to the Go runtime in different ways depending on how a particular function was created.
The interpreter does not substitute types while the compiler does.

The complexity makes it substantially more dangerous to add a new feature or to fix an existing bug.
This leads to some of the more difficult bugs just being part of flux because they are too difficult to fix without a larger investment of time.

The complexity isn't easily understood as there isn't a common language or debug tools to understand specific concepts.
Bugs are generally resolved by putting in more code to prevent a specific bug rather than to understand and fix the underlying reason why it happened.
This leads to more complexity as we will create one-off fixes in different portions of the code.

### Debuggability

The current interpreter and compiler are very hard to debug for correctness.
It's difficult to see what a particular program will do without reasoning through it and seeing the execution.
This adds developer time when making changes to the execution or attempting an optimization.

### Consistency

The current architecture creates some consistency problems.
The consistency problems then cause problems in correctness or developer understandability.
Consistency problems come from multiple angles.

The split between the interpreter and the compiler creates some issues with consistency.
In particular, similar sections of code are executed by different sections of the codebase.
When debugging some code, it can be difficult to know exactly where and how that code was executed.

There's also some functions that work differently than others.
`tableFind()` is the easiest example as it creates a mini-execution engine before or during execution.
Running a script with the interpreter executes code for non-sources/transformations, but it doesn't for sources/transformations.

The amount that a developer has to remember about the execution creates a large barrier to understanding how execution flows.
This can make some code very difficult to write or reason with.

## Proposal

I propose that we make an active effort to define and build a flux compiler and VM.

This sounds like a daunting proposal, but I believe that the work to do this would yield performance increases.
I also believe it would help the system be more consistent which would make it less complex.
By defining specific concepts and building them into a compiler and VM, we would also get improved visibility into the details of execution.

The separation of concerns between a flux compiler and a VM will require us to clearly define how concepts work and interact with each other in a consistent manner rather than writing glue code in Go for a specific workflow.

### Compiler

The purpose of the compiler would be to replace and unify the executor, interpreter, and the compiler into a single entity.
Its purpose is to translate a program into a series of instructions that would be executed by the VM.

### VM

The proposed VM is a straightforward execution.
I do not propose that the VM be advanced such as JIT compiling.
I think we would benefit from just a straight linear execution VM.
The purpose of the VM would be to just require the compiler to define how code interacts with the runtime in advance and to build the specific capabilities we need in a general way.

## Differences

It is important to list the primary differences and why these differences are significant.

### There is One Dispatcher, One Executor, and Multiple Plans

There is a single dispatcher that is responsible for running streams.
The executor, otherwise known as the VM, is also a singular entity.
The VM would construct plans during execution.
When the output of a plan is needed, that plan is passed to the planner and then executed through the dispatcher.
It is normal for multiple plans to be executed rather than have it be a special workflow.

At the current moment, each plan comes with its own executor and dispatcher.
These extra plans can run in many locations that are not consistent.
Consolidating the interaction of these different plans into a single framework makes it normal rather than special to have multiple plans in a query.

### Control Flow is Explicit

The control flow of a program is explicit and can be determined by looking only at the IR.
At the current moment, looking at the program won't tell you how it runs.
You would have to keep the knowledge about how table objects work and how it interacts with the interpreter.

As an example, here's a simple summary of looking at a script and determining the control flow.

1. Does the script invoke a function that interacts with external services such as `http.post()` or `mqtt.publish()`?
2. Does the script use a table stream in an expression statement in the main program at the top level?
3. Does the script use `yield()`?
4. Are there multiple yields? Are they named?
5. Is `tableFind()` used at the top level of a program?
6. Are any of the functions that invoke `tableFind()` at the top level?
7. What are those functions?
8. Is a stream function invoked inside something like `filter()` or `map()`?
9. Are `tableFind()` or any of the derivations inside `filter()` or `map()`?
10. Is `experimental.chain()` used?
11. What order do the streams run in? Is this even possible with the current architecture?

These probably aren't all the questions that could be asked about the workflow.
They are some of them.
It is difficult to know exactly when some section of code will run since code can be run during the interpreter phase or the execution phase.
There can also be multiple execution phases that are implicitly invoked by a function.

As an example of how this might look, we can look at the following pseudo-IR.

    define void @main() {
        %1 = func @gen.tables
        %2 = call stream %1
        yield %2, "_result"
        ret void
    }

We can see from the IR that this invokes the `gen.tables()` source which returns a stream.
We invoke the function.
We then yield the result which hands the stream over to be yielded to `_result`.

If we were to implement `tableFind`, it would be similar and interact with the VM in the same way.

    define void @main() {
        %1 = func @gen.tables
        %2 = call stream %1
        %3 = func @tableFind
        %4 = call %3, %2, ...
        print %4
        ret void
    }

While the `yield` would yield the stream to be returned to the user, invoking `tableFind` would cause the built plan to be optimized, executed, and the result would be returned.

This would also enable other applications that we presently have difficulty with.
If the above works, then we would be able to define the execution order more easily.
This would make it easier to create a more reliable version of `experimental.chain()`.

### More Consistent Naming

An underrated difference is that each of the components in this proposal would have a proper name that denotes what it does.
We would have a compiler which compiles a script into an intermediate representation.
We would have a VM that can run the intermediate representation.

We would not have both a compiler and an interpreter.
The compiler would not execute any code.
The interpreter and the executor would both be merged into the VM rather than separate concepts.
The planner would continue to exist and be invoked within the VM.

### Easier Program Analysis

Programs would be easier to analyze.
We would be able to quantify the cost of each instruction as all attributes of the program would be viewable by looking at the IR.
As we would be using a proper IR to represent the program, we would be able to use standard algorithms to analyze the control flow of the program.
Algorithms that manipulate the programs would be easier to write because each instruction would correspond to one action rather than potentially multiple operations.

For example, vectorization is a project that I believe would have been easier to implement if we had an IR.

It would also be easier to create more optimizations that rely on program analysis.
This is because we would have a control flow graph that encapsulates all operations being invoked by the VM.
We can also view the IR to manually verify the behavior is as we expect.

Consider how we have seen functions get invoked.
Invoking functions is a relatively expensive operation in Flux.
This is because we may invoke functions from a function object.
That function object may have a type signature that says, "I take x and y as an argument", but we don't necessarily know the physical order of those arguments.
This means there's a small amount of dynamicness to invoking a function that prevents a certain optimization.

Using the IR graph, we can trace the `call` invocation to reference the function object created within the same scope.
We could change the invocation to a special version of `call` that invokes the function directly with the arguments in the correct order.

    // %1 = func @add2
    // %2 = call int %1, x: 2
    %1 = call_direct int @add2, 2

With the additional information about which function we are invoking, we can instantiate the function with the proper type arguments ahead of time and prepare the arguments in the proper order to optimize function calls.

After we optimize this, we can also perform inlining.
We could decide to replace the call with the body of the `add2` function.

With our current infrastructure, we might be able to do this optimization.
It would be difficult to debug and visualize and hard to understand how different optimizations interact with each other.

Another useful optimization that specifically relates to vectorizing is how to deal with constants.
The following IR has a real cost to the `const` instruction which corresponds to our existing literal evaluators (like `floatEvaluator`).

    define int @add2(int %v) {
        %1 = const int 1
        %2 = add int %v, %1
        ret int %2
    }

This could be changed to:

    define int @add2(int %v) {
        %1 = addi int %v, 1
        ret int %1
    }

When we perform vectorization, it's easy to see the correlation to vectorization.
We could have an operation for vectorization that adds two vectors and one that adds a constant to a vector.

## Development Plan

The goal of the development plan would be to perform an "as much as possible" compiler and VM implementation that would be resilient to unimplemented operations.

An "as much as possible" implementation would be a compiler that attempts to materialize as many instructions as it can.
If it runs into an operation that is not implemented, it will not emit that function definition, but it will continue to process other function definitions that are contained in the body.
This effectively creates a bottom-up approach since the compiler will emit a partial IR that implements the script.

If the full program gets translated, the `main` function will exist.
For builtin functions like `map`, they will check to see if the function has a VM implementation and will use that one if it is present.
Otherwise, it will use the existing compiler.

For languages, I believe that Rust is a more suitable location to implement the compiler.
The compiler will translate semantic graph to IR, run optimization passes on the IR, then emit an IR suitable to be run by the VM.
The VM will be defined in Go because our runtime is defined in Go.

This plan allows us to gradually implement different instructions in the VM and continue to expand the scope of what can be run by the VM.
The eventual goal is the removal of the interpreter and compiler from Go.

At the end of development, the following will have been implemented:

* Defined an SSA IR with a textual output.
* Constructed function definitions.
* Implemented function objects and call invocations.
* Implemented closures as a part of function objects.
* Constructed VM from SSA IR.
* Ran VM program with function calls, closures, and plan execution operations.
* Implement yield and table find.

### Compiler

The compiler, written in Rust, would be responsible for implementing the following:

* Define an SSA IR with a textual output.
* Translate semantic graph to IR.
* Define function definitions that can be reified at runtime.

The output of this would be a system roughly equivalent to what we have today.
From there, we can use the debug output of the IR and use existing compiler optimization algorithms to implement optimization passes.
The optimization passes would take IR as input and produce IR as output.

### VM

The VM, written in Go, would be responsible for implementing the instructions from the compiler.
For the compiler to emit an instruction, it needs to be implemented in the VM.

The VM is also responsible for constructing plan nodes when they are executed and passing them to the executor to be executed.
The VM finishes execution when its main thread and all dispatcher threads have finished.

## Design

### Compiler

The compiler would evaluate the semantic graph from the type system and produce an SSA IR.
The IR would define various primitive operations that correspond to existing concepts.
Looking at the IR would give a good indication of what the compiled flux script will do and the SSA form will allow easy construction of a control-flow graph.
This will make it generally easier to perform introspection of the program along with making and debugging necessary transformations.

The IR format will be representable textually to aid with debugging.
There is no requirement that the underlying system interact with a textual IR.

A sample of the textual output of an IR for a Flux program follows:

    // y = 2
    // addy = (x) => x + y
    // addy(x: 1)
    define (T: Addable) T @addy(%x: T, %y: T) {
    add2:
        %1 = add T %x, %y
        ret T %1
    }

    define void @main() {
        %1 = const int 2
        %2 = func @add2, y: %1
        %3 = const int 1
        %4 = call int %2, x: %3
        print %4
        ret void
    }

Important concepts that are included in this output:

* Constants
* Function calls
* Closures
* Generics

In the above, we translate the operations at the top level into a default "main" function.
When we see a function expression, we define another function with the generic templates and create a function object that references that function along with the values included in the closure.
This ensures the function holds a copy of the value wherever it may be invoked.
We then load the parameter that will be used for the function call and invoke the function retrieving the result.
In a typical flux script, this output would not be used for anything.
For debug purposes, we add a `print` instruction to print the output.
Then we end the main function by returning void.

A compiler and a VM that runs the above would encapsulate most of a minimum-viable product.
For a minimum-viable product, we would also need to execute a function that produces a stream and have it interact with yield and the dispatcher.

### VM

The purpose of the VM would be to run the above.
The VM would be responsible for handling the dispatcher.
It would be responsible for grouping sources and transformations into plans that would be passed to the planner and then executed by the dispatcher.
It would be responsible for executing each instruction that was chosen by the IR.
It would be responsible for determining storage locations for each register and determining the size of those virtual registers.

#### Function Invocation

A large part of the VM is determining how to execute functions.
In the above example, we could statically determine all types that will be used and we could instantiate a version of the `addy` function that is not generic.

That isn't always possible.
When we invoke a generic function, it might be required to instantiate a generic function into a version with the proper types at runtime.
This is reification and would be a responsibility for the `call` instruction in the VM.

## Development Phases

### Phase 1

Phase 1 would focus on the proof of concept for an SSA IR and the accompanying VM.
This would prove that you could create a simple IR definition and invoke it.
We would then run this in a VM.

The following will be true at the end of this phase:

* IR is generated for main function if all operations are supported.
* No IR is generated if a node cannot be converted to IR, but no error occurs.
* Generics are not implemented.
* Function expressions and call invocations are not implemented.
* Constants are implemented.
* Some subset of binary expressions are implemented.
* Meta `print` expression exists and will print the result of an expression statement to allow for debugging.

There is no production output from this phase.

### Phase 2

Phase 2 would focus on ensuring functions can be created and used.
This would prove you could create functions and call them.
It would also introduce generics into function definitions and reify them based on type arguments in the VM.

Similar to the above, a function that cannot be converted will just not be emitted.
Functions that are not emitted should still emit nested function definitions.
This is to allow a function invoked by a transformation like `map()` to still be emitted even if the main program cannot be.

The following will be true at the end of this phase:

* IR is generated for native flux functions if all operations are supported.
* Recursive function definitions are generated even if the function they are contained within is not.
* Generics are implemented.
* Function objects exist.
* Closures are not implemented.
* Indirect call instruction is implemented in the IR and capable of reifying functions.
* Imports and packages are not implemented.

### Phase 3

Phase 3 is focused on records, their data layout, passing them as arguments, and accessing them.

A primary requirement of this phase is that record literals should be cheap to allocate and use.
In the current implementation, record literals create a memory allocation.
In this version, records should not allocate memory outside the VM stack.

During this phase, we will also explore if the above requirements are unrealistic regarding memory allocations.
While there are situations where records would be faster when passed by value, there are others where they are constructed once and passed around many times.
We may want to explore these areas during this time.

The following will be true at the end of this phase:

* Record literals can be used and returned from functions.
* Records can be passed as an argument to a function.
* Records will be representable directly on the stack.
* Member expressions will be implemented in the IR and VM.

### Phase 4

Phase 4 is focused on implementing closures.

The following will be true at the end of this phase:

* Values that are closed over will be saved as part of the function object.
* Closure mechanics can be visibly seen from the IR.

### Phase 5

Phase 5 is focused on importing packages and using builtin functions that are not streams.

The following will be true at the end of this phase:

* Imports function.
* Builtin values that do not interact with table streams will work.

At this point, the remaining instructions can be implemented separately and the VM can replace the compiler for most circumstances.
This includes any remaining instructions that have not been implemented from phase 1.

### Phase 6

Phase 6 is focused on integrating table streams into the VM.
Functions that return table streams will construct a plan and the VM will optimize the plan and then pass it to the executor/dispatcher.

The following will be true at the end of this phase:

* Sources and transformations are constructed by the VM.
* Constructed plans are passed through the planner and executed by the executor.
* The yield instruction is implemented and is capable of returning table stream data.
* Table find style functions do not work.

### Phase 7

Phase 7 is focused on ensuring table find style functions work properly within the VM.

The following will be true at the end of this phase:

* Table find style functions will be capable of using the same planner/executor as yielded table streams.
* Table find style functions work.

### Phase 8

Phase 8 is focused on polishing up the implementation and ensuring it is feature complete.
This phase can be started as early as after phase 5 and is primarily focused on implementing any remaining instructions from the semantic graph that haven't been implemented.

The following will be true at the end of this phase:

* The bytecode compiler and VM will be feature complete and capable of replacing the compiler and interpreter for all flux programs.

## Optimizations

The above development phases are focused on creating a feature complete compiler and VM.
The result should unify the compiler and interpreter and create a framework for further optimizations.
At this point in development, we would branch out and begin to implement the desired optimizations.

Some optimizations are purely changes to the IR while some require additional instructions so the VM can support the operations.

### Rolling Constants

Rolling constants will take constant operations and roll them into a single instruction.
As an example, we might take the following code:

    %1 = const int 2
    %2 = add int %x, %1

And we would change it to:

    %1 = addi int %x, 2

This would reduce the number of values on the stack and create a single instruction for adding a constant instead of one instruction to create the constant and another to use it.

Another aspect of rolling constants is to combine constants together.

    %1 = const int 2
    %2 = const int 4
    %3 = add int %1, %2

This becomes the following code pretty easily:

    %1 = const int 6

If this value is then used with a variable, we can also use the first described optimization to take a lot of potential instructions and collapse them into one instruction.

### Direct Calls

A direct call is a call instruction where we can statically know the location.
An indirect call requires us to perform a lookup on the function object for the order it expects arguments at runtime because we don't know how the function expects its arguments and flux uses named arguments.
A direct call allows us to perform this lookup at compile time instead.

As an example, consider the following:

    // add2 = (x) => x + 2
    // add2(x: 2)

    define int @add2(int %x) {
        %1 = const int 2
        %2 = add int %x, %1
        ret int %2
    }

    define void @main() {
        %1 = func @add2
        %2 = const int 2
        %3 = call int %1, x: %2
        print %3
    }

As we invoke `call` from the same function where we define the function object, we can statically adjust this IR to direclty call `add2` instead.

    define void @main() {
        %1 = const int 2
        %2 = call_direct int @add2, %1
    }

Note how in this second one we do not have to specify the keyword arguments.
This is because we already know the order that arguments should be supplied.

### Inlining

This optimization is more used to enable the other optimizations.
For example, say we have the following code:

    v = () => 2
    2 + v()

This might become the following IR:

    define int @v() {
        %1 = const int 2
        ret int %1
    }

    define void @main() {
        %1 = func @v
        %2 = const int 2
        %3 = call int %1
        %4 = add int %2, %3
        print int %4
    }

We first start off by rewriting the call to a direct call.

    define void @main() {
        %1 = call_direct int @v
        %2 = const int 2 
        %3 = add int %1, %2
        print int %3
    }

We recognize that the direct call might be less costly if we just inline it.

    define void @main() {
        %1 = const int 2
        %2 = const int 2
        %3 = add int %1, %2
        print int %3
    }

We now gain insight that we actually have two constants.

    define void @main() {
        %1 = const int 4
        print %1
    }

### Vectorization

We would translate the vectorization code into an IR by defining and implementing new versions of the instructions that operate on vectorized types.
