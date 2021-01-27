# Flux Engine Developer Guide

This guide will cover how to create a flux source and transformation.
This guide will also cover practical tips and guidelines for how to make sources and transformations efficient.

This guide specifically focuses on how to write these sources and transformations in Go code as builtin operations.
When writing a new transformation, it is best to first see if you can write your transformation using existing builtins.
If it is possible to represent your source or transformation using existing builtins, it is preferred to do that instead under all circumstances.
If the only reason you are using a builtin is for performance, then see the section **Optimizing Native Operations**.

## Concepts

The following concepts are important to understanding the Flux engine and the operations that can be executed within the engine.
This guide will reference these concepts with the assumption that the reader knows what is contained in this section.

### Tables and Table Views

In Flux, a Table is the set of all data produced by a source.
As an example, if we read and filter on a specific measurement from InfluxDB, the Table will contain all of the values for all of the series in that measurement.
Tables are ephemeral.
Transformations do not handle the entire table at a single time and the data for a table is not kept in memory unless a transformation is specifically written to do that.
Instead, tables are split into table views.

A table view is a view of the table for a specific segment of in-memory data.
Any data within a table view is **in memory** and can be actively accessed.
Table views are also divided by how data is grouped.
Various different sources and transformations will group data in different ways.
InfluxDB groups data from the same series and different series are divided into separate views.
While many sources produce grouped data in close proximity to each other, table views may arrive to a transformation out of order.
Table views with the same group key may also arrive to a transformation with different schemas.

For those writing transformations outside of `execkit`, the name `Table` has an overloaded meaning.
In older Flux sources and transformations, a `flux.Table` represented a single group key and a stream of views, called `flux.ColReader`.
In `execkit`, a table refers to the entire collection of data across group keys.

### Dataset

A dataset is connector between transformations and a collection of state for the transformation.
Each transformation is associated with a single dataset and that dataset keeps the state for that transformation.
That dataset also keeps a list of the downstream datasets for data produced by the transformation.
A downstream dataset is any dataset that will receive data from the current transformation.

Upstream datasets are the datasets that pass information to the transformation.
Most transformations only have one direct upstream dataset but will transparently handle multiple direct downstream datasets.

Datasets communicate using messages and transformations handle what actions happen when a message is received.
The choice of a correct transformation type will influence how the transformation handles messages.

### Message

A message is a signal and associated data that is sent from one dataset to another.
A message is processed by the associated transformation when it is sent to the downstream dataset.

The following messages presently exist:

|Name|Data|Description
|----|----|-----------
|ProcessView|table.View|A table view is ready to be processed
|FlushKey|group key|Data associated with the given group key should be flushed from the dataset
|WatermarkKey|group key and time|Data older than the given time will not be sent for the given group key (unimplemented)
|Finish|error (optional)|The upstream dataset will produce no more messages
|Process|flux.Table|A full table is ready to be processed (deprecated)
|UpdateWatermark|time|Data older than the given time will not be sent (deprecated)
|UpdateProcessingTime|time|Marks the present time (deprecated)
|RetractTable|key|Data associated with the given group key should be retracted (deprecated)

Messages depend on the source implementing them.
Sources need to implement, at a minimum, `ProcessView` and `Finish`.
It is advised for them to implement other messages so that transformations can take smarter actions.

### Group Key

A group key is a collection of key/values that are common within a table view.
Any columns that are part of the group key will have the same value within the view.

### Data Memory

Data memory references memory used for storing user data.
This is in contrast to process or program memory.
Since flux is integrated in co-tenant environments, it needs to handle arbitrary user data and avoid crashing the system or causing a noisy neighbor performance problem.
Flux does this by separating the concept of process memory and data memory.

Process memory handles any memory required to execute code and allocate memory to support the execution of that code.
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

## Transformations

### Choosing a Transformation Type

Most transformations fall into one of the following broad categories.

* Narrow
* Group
* Aggregate

These three bases are the cornerstone of most transformations and understanding which one to choose will influence how you write your transformation.

#### Narrow Transformations

A narrow transformation is one that operates on one group key and does not modify the group key.
Narrow transformations will potentially produce new table views as they process incoming table views.

A narrow transformation is split into two categories: ones that have state and ones that do not.

For transformations that save state for each transformation and then produce a table when all of the views have been processed, it is recommended to use an aggregate transformation instead.
An aggregate transformation is a further specialized narrow transformation.

A narrow transformation is created by implementing the `execkit.NarrowTransformation` interface.

    type NarrowTransformation interface {
        // Process will process the TableView.
        Process(view table.View, d *Dataset, mem memory.Allocator) error
    }

The typical workflow for a narrow transformation is the following structure:

    func (t *Transformation) Process(view table.View, d *execkit.Dataset, mem memory.Allocator) error {
        buffer, err := t.createSchema(view)
        if err != nil {
            return err
        } else if err := t.processView(view, &buffer); err != nil {
            return err
        }
        return d.ProcessFromBuffer(buffer)
    }

If the narrow transformation needs to maintain state between buffers, such as things like derivatives and moving averages, the stateful version is useful.

    type NarrowStateTransformation interface {
        // Process will process the TableView.
        Process(view table.View, state interface{}, d *Dataset, mem memory.Allocator) (interface{}, bool, error)
    }

The first time the group key is encountered, the state will be nil.
It should be initialized by the `Process` method.
The new state will be stored if it is returned along with the second return argument being true.
If the second return argument is false, the state will not be modified.
It is both ok and expected that the `interface{}` will be a pointer to a struct and will be modified in a mutable way.
The state is not expected to be immutable.

#### Group Transformations

A group transformation is one that modifies the group key of the incoming transformation.
It is otherwise identical to a narrow transformation.
The primary difference is that because group transformations do not know when to flush data associated with a key, they will swallow that message and prevent downstreams from receiving it.

When implementing a group transformation, it is likely that `CopySchema` will not be used.
In general, the schema will be changed and the group key might be changed.
Other than the possibility of changing the group key, the code is identical to a narrow transformation.

#### Aggregate Transformations

An aggregate transformation is one that aggregates and computes a result for a single group key and does not modify the group key.
Aggregate transformations do not stream tables, but instead wait for all table views of a group key to be received first.

Aggregate transformations will save state between each view they receive and then invoke a computation method with that state when that group key is flushed or when the finish signal is received.
This type of transformation is used most frequently for something like `mean()`, but it can also be used for other transformations that cannot act until they receive all of their data such as `sort()`.

An aggregate transformation is created by implementing the `execkit.AggregateTransformation` interface.

    type AggregateTransformation interface {
        // Aggregate will process the TableView with the state from the previous
        // time a table with this group key was invoked.
        // If this group key has never been invoked before, the
        // state will be nil.
        // The transformation should return the new state and a boolean
        // value of true if the state was created or modified.
        // If false is returned, the new state will be discarded and any
        // old state will be kept.
        // It is ok for the transformation to modify the state if it is
        // a pointer. This is both allowed and recommended.
        Aggregate(view table.View, state interface{}, mem memory.Allocator) (interface{}, bool, error)

        // Compute will signal the AggregateTransformation to compute
        // the aggregate for the given key from the provided state.
        Compute(key flux.GroupKey, state interface{}, d *Dataset, mem memory.Allocator) error
    }

Similar to the narrow state transformation, the state is nil when it is first created and the new state is returned on each invocation of `Aggregate`.
The primary difference is that `Compute` is invoked when the aggregate has been signaled that no more data for the current key will be sent.
A common pattern for this transformation:

    func (t *Transformation) Aggregate(view table.View, state interface{}, mem memory.Allocator) (interface{}, bool, error) {
        var s *myState
        if state == nil {
            s = &myState{}
        }

        if err := t.processView(view, s); err != nil {
            return nil, false, err
        }
        return s, true, nil
    }

    func (t *Transformation) Compute(key flux.GroupKey, state interface{}, d *execkit.Dataset, mem memory.Allocator) error {
        buffer, err := t.createBufferFromState(key, state.(*myState))
        if err != nil {
            return err
        }
        return d.ProcessFromBuffer(buffer)
    }

## Arrow Operations

Arrow is a library and data format for data.
The purpose of arrow is to be an interoperable format to be used across multiple data science platforms so they may communicate.
Flux uses Arrow as its primary method of storing and interacting with data.
When producing sources or transformations in Flux, Arrow arrays are used to represent each column in a table.
For this reason, it is important to understand the capabilities of Arrow and some common operations employed by Flux transformations to most efficiently use the Arrow library and data format.

Arrow arrays are immutable after they are created.

### Constructing a Table View

When constructing a table view, a developer needs to organize a group key, table schema, and columnar values.
This is done by using the `arrow.TableBuffer` struct in the `github.com/influxdata/flux/arrow` package.
This is defined as:

    type TableBuffer struct {
        GroupKey flux.GroupKey
        Columns  []flux.ColMeta
        Values   []array.Interface
    }

The `Columns` field holds the schema for the table by specifying the column label and type for a certain index.
Each of the column names must be unique.

The `GroupKey` column holds the grouping for the current buffer.
Each table view has a group key that contains the list of columns where all of the values are the same.
If this is a narrow transformation, the key will be a copy of the previous key so it is not necessary to create a new one.
Otherwise, the `execute.NewGroupKey` function is available.

For narrow transformations, it is common for the above two to be the same as the incoming table view.
The `CopySchema` method exists for these situations.

    // A transformation like filter or sort.
    func (t *Transformation) Process(view table.View, d *execkit.Dataset, mem memory.Allocator) error {
        buffer := view.CopySchema()
        // Construct the new values and assign them to the buffer.
        out := table.ViewFromBuffer(buffer)
        return d.Process(out)
    }

The final attributes is `Values`.
This contains the arrays of values for each of the columns as represented by arrow arrays.
Each of the arrays must be the same length.

### Copying Data

Copying data for Arrow arrays comes in a two different ways.

1. Full or slice of data copies.
2. Random-access copies.

The first is always preferred if possible.

To copy the full array, the `Retain()` method may be invoked on the array.
As an alternative, the `arrow.Copy()` method may also be used to do the same thing.
There is no performance difference between these two as they are identical.

    import (
        "github.com/apache/arrow/go/arrow/array"
        "github.com/influxdata/flux/arrow"
    )
    func ExampleCopyValues(dst, src []array.Interface) {
        // Standard arrow.
        for i := range dst {
            dst[i] = src[i]
            dst[i].Retain()
        }

        // Using arrow.Copy.
        for i := range dst {
            dst[i] = arrow.Copy(src[i])
        }
    }

To take a slice of a contiguous set of values, the `arrow.Slice` may be used.

    import (
        "github.com/apache/arrow/go/arrow/array"
        "github.com/influxdata/flux/arrow"
    )
    func ExampleSliceValues(dst, src []array.Interface, i, j int64) {
        for i := range dst {
            dst[i] = arrow.Slice(src[i], i, j)
        }
    }

For this one, you must use `arrow.Slice` instead of the `array.NewSlice` method directly from arrow.
This is because the arrow version will generate the incorrect type for string arrays which can cause a panic when another transformation attempts to read it.

TODO: Write a function to aid with for random-access copies and then document it here.

### Filter Data

Similar to copying data, filtering can be seen in the same two ways.

1. Filter of a range of data.
2. Random-access filtering.

For filtering a range of data, a slice as described above can be used.

For random-access filtering, the `arrowutil.Filter` function works best.
To filter, you create a bitset using arrow and then invoke the `Filter` function.

    import (
        "github.com/apache/arrow/go/arrow/memory"
        "github.com/apache/arrow/go/arrow/bitutil"
        "github.com/influxdata/flux/internal/arrowutil"
    )

    func FilterNegativeValues(arr *array.Float64, mem memory.Allocator) *array.Float64 {
        bitset := memory.NewResizableBuffer(mem)
        defer bitset.Release()

        n := bitutil.BytesForBits(arr.Len())
        for i, l := 0, arr.Len(); i < l; i++ {
            val := arr.Value(i) >= 0.0
            bitutil.SetBitTo(bitset.Buf(), i, val)
        }
        return arrowutil.FilterFloat64s(arr, bitset.Bytes(), mem)
    }

### Empty Array

It is sometimes necessary to create an empty array.
To create an empty array, the `arrow.Empty` function may be used.

    import (
        "fmt"
        "github.com/influxdata/flux"
        "github.com/influxdata/flux/arrow"
    )

    func Process(...) {
        arr := arrow.Empty(flux.TFloat)
        fmt.Println(arr.Len())
    }

If your transformation creates a new builder and doesn't add any values, it will produce the same array as the above.
That means you should not do this:

    b := array.NewFloat64Builder(mem)
    // ... code that builds the array ...
    var arr array.Interface
    if b.Len() == 0 {
        // UNNECESSARY!
        arr = arrow.Empty(flux.TFloat)
    } else {
        arr = b.NewArray()
    }

Instead, this is the correct method.

    b := array.NewFloat64Builder(mem)
    // ... code that builds the array ...
    arr := b.NewArray()

### Mutable Data

## Optimizing Native Operations

Flux, by its nature, is meant to be used in performance sensitive areas.
If there is an area where the existing builtin transformations can't quite get the performance that is needed, but the operation is otherwise able to be represented by pure Flux code, it is advised to utilize a hybrid approach.
In the hybrid approach, we define the operation in pure Flux code, but we add a planner rule to the Flux engine that replaces the native Flux code with our builtin code.

TODO: write how to do this.

## Tips

### Prefer algorithms that operate on data in a columnar way

Flux keeps data in memory for a table view in a columnar format.
This means that data within a column is typically stored close to each other in RAM and can be accessed as a continugous array.
When accessing multiple columns, this isn't the case.
The OS may have more cache misses when accessing data and row-based algorithms will also have to make calls to the virtual table to determine what type of data it is reading.

