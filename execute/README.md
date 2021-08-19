# Flux Engine Developer Guide

This guide covers the Flux Engine design and related concepts.
The engine includes:

* Execution engine
* Builtin sources and transformations
* Memory layout and resource management

## Concepts

The following concepts are important to understanding the Flux engine and the operations that can be executed within the engine.
This guide will reference these concepts with the assumption that the reader has reviewed the following section.

### Tables and Table Chunks

In Flux, a **Table** is a set of ordered rows grouped together by a common **group key**.
The **group key** is a set of key/value pairs.
The elements of the group key appear as columns in the table with the element key appearing as the column name and the element value appearing as the value, constant across all rows.
In practical terms, this means that for any columns that are in the group key, the values in that table will all be the same.

The table will also contain columns that are not referenced by the group key.
These columns are free to be any value as long as they are all the same type.

For a table, the following invariants should hold:

* All values inside a column that is part of the group key should have the same value.
* That value should match the group key's value.
* The values in a column should all be of the same type.

The last of these invariants is **loosely** held.
This means that while they should be true, individual transformations may not make an effort to enforce them.
That means that if a transformation would have its performance affected by enforcing the invariant and the invariant does not affect the transformation's correctness, the transformation may not enforce it when producing the output.
For an example, the `group()` transformation does not enforce the last invariant and relies on downstream transformations to enforce it.

The set of rows contained within a table are processed in chunks.
A table chunk is a subset of the rows within a table that are in-memory.

The rows within a table are **ordered**.
When table chunks are sent, the rows within the first are all ordered before the rows in the second.
Transformations may choose to change the order, or they may choose to have the order be meaningful.
For example, the `derivative()` function gives meaning to the order of the input rows while `sort()` will rearrange the order.

A table chunk is composed of a list of arrow arrays, each array corresponding to a column of the table.

### Arrow Arrays

[Apache Arrow](https://arrow.apache.org/) is a language-independent columnar format.
The flux engine utilizes the Arrow library to represent the columnar data inside of table chunks.

### Execution Pipeline

The **execution pipeline** is the set of nodes passing data from a source to the result.
The execution pipeline is composed of **nodes**.
A node with no inputs is a **source** which produces data for nodes that are after, or downstream, of itself.
A node that takes one or more inputs and produces an output is a **transformation**.
The final node in the pipeline is the **result** which holds the results of the pipeline.

You can think of the execution pipeline in flux code.

    A |> B |> C

This would translate into three nodes where `A` is a source and `B` and `C` are transformations.
Flux code does not correspond directly with one function being one node.
Some functions will be combined into a single node while other functions may get rearranged into other nodes or split into multiple nodes.
It is the responsibility of the planner to convert flux code into an execution pipeline.

### Dataset

A dataset is used to hold node state and manage sending messages to downstream nodes.
Each transformation is associated with a single dataset.

When thinking about the execution pipeline, the dataset can be thought of as a distinct part of the node.

    A |> B |> B(d) |> C |> C(d)

### Message

A message is a signal that contains associated data that is being sent from one dataset to the downstream nodes.
Each message has a lifetime where it is created by the sender and then acknowledged by the receiver.
A message may hold onto memory and will release its reference to that memory after it has been acknowledged.
The data contained within a message may be retained and used in other ways.

The following messages presently exist:

|Name|Data|Description
|----|----|-----------
|ProcessChunk|table.Chunk|A table chunk is ready to be processed
|FlushKey|group key|Data associated with the given group key should be flushed from the dataset
|Finish|error (optional)|The upstream dataset will produce no more messages

The following messages exist, but are deprecated and should not be used by future transformations.

|Name|Data|Description
|----|----|-----------
|Process|flux.Table|A full table is ready to be processed
|UpdateWatermark|time|Data older than the given time will not be sent
|UpdateProcessingTime|time|Marks the present time
|RetractTable|key|Data associated with the given group key should be retracted

### Memory Types

Data memory references memory used for storing user data.
This is in contrast to process or program memory.
Since flux is integrated in co-tenant environments, it needs to handle arbitrary user data and avoid crashing the system or causing a noisy neighbor performance problem.
Flux does this by separating the concept of process memory and data memory.

Process memory is any memory required to execute code.
Process memory is anything in Go that would allocate stack or heap memory and is tracked by the garbage collector.
User code, by necessity, will use process memory.
The amount of process memory used by a transformation should be designed to be fairly consistent.
If a user has 10 rows or 10,000 rows, the process memory should not scale directly with the number of rows.
This is not a strict requirement.
There's no expectation that a transformation will use the same memory regardless of the number of rows, but just the requirement that it doesn't linearly scale with the number of rows or worse.

Data memory has different conditions.
Data memory is memory that is used to store user data and data memory is tracked by the `memory.Allocator`.
The flux engine places limits on the amount of data memory that can be used and will abort execution if more is used.
It is allowed and expected that some transformations will have bad memory footprints for certain inputs.

The primary method of storing data is through immutable arrow arrays.
There are also circumstances where mutable data is needed to implement an algorithm, but process memory is not an appropriate place to store that data (such as `distinct()` or `sort()`).
See the **Mutable Data** section for how to handle these circumstances.

### Side Effects

Side effects are metadata attached to a builtin function call to signify that the function performs actions outside the query execution.
At least one side effect is required for a query to be valid.
Examples of side effect functions are `yield()` and `to()`.
The main package will turn an expression statement into a side effect by adding an implicit `yield()` to the end of the expression.

## Execution Engine

The execution engine is central to executing flux queries.

### Pipeline Creation

Before the execution engine begins, it's important to understand the steps that happen before execution.
This is only a brief outline of those steps.

* Text is parsed into an AST.
* AST is converted into semantic graph.
* Type inference runs on the semantic graph and assigns types to all nodes.
* The interpreter uses a recursive-descent execution to evaluate the semantic nodes.
* Side effects are gathered from the interpreter execution and side effects that are linked to table objects are collected into a spec.
* The spec is converted into a plan spec (one-to-one mapping).
* The plan spec is passed into the planner which executes logical rules, converts logical nodes to physical nodes, and then runs physical rules.
* The plan spec is passed to the execution engine.

That last step is where the execution engine starts.
The execution engine starts with an already constructed plan and the execution engine's job is to execute that as faithfully as possible.

### Nodes and Transports

When initializing the execution engine, the plan contains a directional graph which can be converted into a pipeline.
Each node in the graph corresponds to a source or a transformation.
A source is a node that has no upstream (input) nodes.
A source will produce data and send that data to the downstream (output) nodes.
A transformation is a node that has one or more upstream (input) nodes.
A transformation will also have one or more downstream (output) nodes.
At the end of the pipeline, a result is waiting to read the output data.

Nodes are connected by transports.
A transport exists to send messages from one node to the downstream node.
Transformations implement the `Transport` interface using one of the transformation primitives mentioned in the next section.

### Dispatcher and Consecutive Transport

Each transformation node in the pipeline implements the `Transport` interface and execution is controlled by the dispatcher.
These transports are automatically wrapped by the consecutive transport.
The consecutive transport is a transport on every node that keeps a message queue and processes messages from that queue inside of dispatcher worker threads.

The practical effect of this is that invoking `ProcessMessage` on a consecutive transport will not immediately execute the action associated with that message.
Instead, the dispatcher will make decisions about which transport to run depending on the concurrency resource limit.
If the concurrency limit is only one, then only one transformation will execute at a time.
If the concurrency limit is more, we can have more than one transformation running concurrently.
In all situations, it is impossible for the same node to execute in multiple dispatcher workers at the same time.

The dispatcher is initialized with a throughput.
The throughput is used to determine how many messages will be processed by a single `Transport` before another `Transport` is given the worker thread.
The throughput _is not_ the concurrency.

## Transformations

Most transformations fall into one of the following broad categories.

* Narrow
* Group
* Aggregate

These three bases are the cornerstone of most transformations and understanding which one to choose will influence how you write your transformation.

### Choosing a Transformation Type

A **narrow** transformation is one where the group key of the input is not changed and corresponds 1:1 with the output table's group key.
A narrow transformation will also be able to produce its output as it receives its input without getting a final finish message to flush state.
This is the simplest transformation type to implement and should be preferred over others when possible.

A **group** transformation is similar to a narrow transformation, except the output group key changes and group keys can be transformed one-to-one, one-to-many, many-to-one, or many-to-many.
When a transformation will change the group key in some way but does not need a final finish message to send its output, this transformation should be preferred.

An **aggregate** transformation is a narrow transformation that is split into two phases: aggregation and computation.
First, an aggregate transformation should meet the same requirement as a narrow transformation in regards to the group key input and output.
The group key of the input is not changed and corresponds 1:1 with the output table's group key.
After that, processing is split into the aggregation phase which reads the data, performs some processing, and outputs an intermediate state.
When we receive the message that the group key can be flushed, we enter into the computation section of processing which turns the intermediate state into a materialized table.

### Narrow Transformation

A **narrow** transformation is one where the group key of the input is not changed and corresponds 1:1 with the output table's group key.
A narrow transformation will also be able to produce its output as it receives its input without getting a final finish message to flush state.
This is the simplest transformation type to implement and should be preferred over others when possible.

There are two subtypes of narrow transformations: stateless and stateful.

A narrow transformation that is stateless is implemented using the `NarrowTransformation` interface.

    type NarrowTransformation interface {
        Process(chunk table.Chunk, d *TransportDataset, mem memory.Allocator) error
    }

The `Process` method is implemented to take a table chunk, transform it using the memory allocator for new allocations, and then send it to be processed by the dataset.
A skeleton implementation is shown below:

    func (t *MyNarrowTransformation) Process(chunk table.Chunk, d *TransportDataset, mem memory.Allocator) error {
        out, err := t.process(chunk, mem)
        if err != nil {
            return err
        }
        return d.Process(out)
    }

    func (t *MyNarrowTransformation) process(chunk table.Chunk, mem memory.Allocator) (table.Chunk, error) {
        /* transformation-specific logic */
    }

Some examples where this version of the narrow transformation is used: [filter()](https://github.com/influxdata/flux/blob/master/stdlib/universe/filter.go) and [fill()](https://github.com/influxdata/flux/blob/master/stdlib/universe/fill.go).

An alternative is used when we need to maintain state between chunks.

    type NarrowStateTransformation interface {
        Process(chunk table.Chunk, state interface{}, d *TransportDataset, mem memory.Allocator) (interface{}, bool, error)
    }

The first time the group key is encountered, the state will be nil.
It should be initialized by the `Process` method.
The new state will be stored if it is returned along with the second return argument being true.
If the second return argument is false, the state will not be modified.
It is both ok and expected that the `interface{}` will be a pointer to a struct and will be modified in a mutable way.
The state is not expected to be immutable.

A skeleton implementation is shown below:

    type myState struct { ... }

    func (t *MyNarrowTransformation) Process(chunk table.Chunk, state interface{}, d *TransportDataset, mem memory.Allocator) error {
        state := t.loadState(state)
        out, err := t.process(chunk, state, mem)
        if err != nil {
            return nil, false, err
        }

        if err := d.Process(out); err != nil {
            return nil, false, err
        }
        return state, true, nil
    }

    func (t *MyNarrowTransformation) loadState(state interface{}) *myState {
        if state == nil {
            return &myState{}
        }
        return state.(*myState)
    }

    func (t *MyNarrowTransformation) process(chunk table.Chunk, state *myState, mem memory.Allocator) (table.Chunk, error) {
        /* transformation-specific logic */
    }

Some examples where this version of the narrow transformation is used: [derivative()](https://github.com/influxdata/flux/blob/master/stdlib/universe/derivative.go).

### Group Transformation

TODO: Write this section.

### Aggregate Transformation

TODO: Need to refactor the existing aggregate transformation into a more generic interface.

## Building Table Chunks

The above transformations involve taking input data, reading it, and producing a table.
We need to know how to create a table chunk to create a table.

A table chunk is composed of a group key, a list of columns, and an array of values to represent each column.
The following general steps are used to build every table:

* Determine the columns and group key of the output table.
* Determine the length of the table chunk.
* Construct the array for each column.
* Construct and send a [table.Chunk](https://pkg.go.dev/github.com/influxdata/flux/execute/table#Chunk) using [arrow.TableBuffer](https://pkg.go.dev/github.com/influxdata/flux/arrow#TableBuffer).

### Determine the columns and group key of the output table

To determine the columns and group key of the output table will depend entirely on the transformation that is being implemented.
Many transformations will not modify the group key.
For transformations that do not modify the group key, the [execute.NarrowTransformation](https://pkg.go.dev/github.com/influxdata/flux/execute#NarrowTransformation) transport implementation can greatly simplify the creation of those transformations.

### Determine the length of the table chunk

After the columns and group key have been determined, the length of the next table chunk should be determined.
Some transformations, like `map()`, will always output the same number of rows they receive in a chunk.
These are the easiest.
Others, like `filter()`, might reduce the length of the array and should determine the new length of the table chunk in advance.
There are also cases where a transformation might need to rearrange data from different buffers or could produce more data.
For these circumstances, chunk sizes should be limited to [table.BufferSize](https://pkg.go.dev/github.com/influxdata/flux/execute/table#pkg-constants).

**It is not required that a transformation determine the length of a table chunk before producing one**, but it is highly advised.
Memory reallocation during table chunk creation is a top contributor to slowdown.

### Construct the array for each column

We produce an array for each column in the table chunk using the [github.com/influxdata/flux/array](https://pkg.go.dev/github.com/influxdata/flux/array) package.
Each flux type corresponds to an array type according to the following table:

|Flux Type|Arrow Type
|---------|----------
|Int|Int
|UInt|Uint
|Float|Float
|String|String
|Bool|Boolean
|Time|Int

At its simplest, creating an array is done using the given skeleton.

    b := array.NewIntBuilder(mem)
    b.Resize(10)
    for i := 0; i < 10; i++ {
        b.Append(int64(i))
    }
    return b.NewArray()

Other techniques for building arrays efficiently are contained in the arrow arrays section below.

### Construct and send a table.Chunk using arrow.TableBuffer

We construct a [table.Chunk](https://pkg.go.dev/github.com/influxdata/flux/execute/table#Chunk) using [arrow.TableBuffer](https://pkg.go.dev/github.com/influxdata/flux/arrow#TableBuffer).

    buffer := arrow.TableBuffer{
        GroupKey: execute.NewGroupKey(...),
        Columns: []flux.ColMeta{...},
        Values: []array.Interface{...},
    }
    chunk := table.ChunkFromBuffer(buffer)

    if err := d.Process(chunk); err != nil {
        return err
    }

## Arrow Arrays

When constructing arrow arrays, there are some general guidelines that apply to every transformation.

### Preallocate Memory

Preallocate memory by determining the size of a chunk in advance and using `Resize()` to set the capacity of the builder.
For strings, it is also helpful to preallocate memory for the data using `ReserveData` if this can be easily known.
String appends are usually the biggest performance sink for efficiency.

### Limit Chunk Sizes

Limit chunk sizes to [table.BufferSize](https://pkg.go.dev/github.com/influxdata/flux/execute/table#pkg-constants).
The array values in a column are contained in contiguous data.
When the chunk size gets larger, the memory allocator has to find a spot in memory that fits that large size.
Larger chunk sizes are generally better for performance, but buffer sizes that are too large put too much pressure on the memory allocator and garbage collector.
The `table.BufferSize` constant is an agreed upon size to balance these concerns.
At the moment, this value is the same as the buffer size that comes from the influxdb storage engine.
If we find that another buffer size works better in the future, we can change this one constant.

### Prefer Column-Based Algorithms

Column-based algorithms are generally faster than row-based algorithms.
A column-based algorithm is one that lends itself to constructing each column individually instead of by row.
Consider the following two examples:

    b.Resize(10)
    switch b := b.(type) {
    case *array.FloatBuilder:
        for i := 0; i < 10; i++ {
            b.Append(rand.Float64())
        }
    case *array.IntBuilder:
        for i := 0; i < 10; i++ {
            b.Append(rand.Int64())
        }
    }

    b.Resize(10)
    for i := 0; i < 10; i++ {
        switch b := b.(type) {
        case *array.FloatBuilder:
            b.Append(rand.Float64())
        case *array.IntBuilder:
            b.Append(rand.Int64())
        }
    }

These are simple examples, but the first is column-based and the second is row-based.
The row-based one spends a lot of time checking the type of the builder before appending the next value.
This is a slow operation inside the for loop which itself is a hotspot for optimization.
The first determines the type of the column first and then constructs the entire column using a specialized type.
The first one only needs to pay the indirection cost of the interface once and can benefit from for loop optimizations.

One practical example is `filter()`.
The `filter()` function is row-based in that it evaluates each row independently to determine if it is filtered or not.
A naive implementation would construct the builder for each row, resize it to the maximum capacity, and then append each value to each column whenever the row passed the filter.
It would then finish by reallocating the arrays to the proper size.
A faster method is to use a bitset to keep track of which rows are filtered and which ones remain.
We can allocate a bitset that holds the rows in the current table chunk, run the filter on each row, and record whether it passes the filter or not.
We can then use that bitset to construct each column independently of each other.

### Utilize Slices

Arrow arrays are immutable and have built in support for slicing data.
A slice keeps the same reference to the underlying data, but limits the view of that data to a smaller section.
Slices should be used when the data naturally slices itself.
An example is the `limit()` transformation where the data naturally gets sliced and the extra memory that is retained may not matter very much.

Slices can have disadvantages though.
An algorithm that goes out of its way to use slices to conserve memory can end up using more memory and cpu cycles.
An example is `filter()`.
With `filter()`, we can have a situation where we filter out every even row and keep every odd row.
If we attempted to use slices for this, we would have a bunch of table chunks that were of length 1.
Table chunks of one increase the number of table chunks and increase likelihood that we will spend more time on overhead than data processing.
Since the slices reference the same underlying array, it also prevents us from releasing the memory used by data in the even rows which uses more memory in the overall query.
This last part still applies to transformations like `limit()` even though that transformation can benefit from slices in some circumstances.

In circumstances like the above, copies can be much more efficient.

### Copying Data

Data can be copied with the `Copy` utilities in the [github.com/influxdata/flux/internal/arrowutil](https://pkg.go.dev/github.com/influxdata/flux/internal/arrowutil) package.
There are many copy utilities in there, but the most useful is likely [CopyByIndexTo](https://pkg.go.dev/github.com/influxdata/flux/internal/arrowutil#CopyByIndexTo).
This method takes a list of indices from the source array and copies them into the destination builder.

### Dynamic Builders

Sometimes, we cannot avoid row-based algorithms and row-based algorithms are likely going to require dynamically appending values.
There are two useful methods in the [github.com/influxdata/flux/arrow](https://pkg.go.dev/github.com/influxdata/flux/arrow) for this.

The first is [arrow.NewBuilder](https://pkg.go.dev/github.com/influxdata/flux/arrow#NewBuilder).
This takes a column type and produces an appropriate builder for that column type.

The second is [arrow.AppendValue](https://pkg.go.dev/github.com/influxdata/flux/arrow#AppendValue).
This one takes a builder, usually constructed with `arrow.NewBuilder`, and appends the value to the arrow builder.

The most common usage of these is like this:

    b := arrow.NewBuilder(flux.ColumnType(v.Type()), mem)
    if err := arrow.AppendValue(b, v); err != nil {
        return nil, err
    }
    return b.NewArray(), nil
