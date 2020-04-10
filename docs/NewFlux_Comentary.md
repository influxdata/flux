# New Flux

Author: Nathaniel Cook

Recently Paul Dix and I added a document on the vision of a new Flux that is more usable.
The design there solves several key points of feedback we have received on Flux.


That document describes the what and some of the why. This document will discuss the how.
This document is not meant to be prescriptive rather just my personal commentary on how I see the vision of new Flux could be implemented.
It is meant to be a seed to start a good team discussion around how  we can implement the new Flux vision.


## What is new?

First before we talk about how we would implement the new vision, we should clarify what is new.

### Single Table vs Stream of Tables

The new Flux data model changes to be a single table with a homogeneous schema.
This is different from the current design which represents data as a stream (unbounded sequence) of tables with possibility differing schemas.

This change is fundamental to the new design, the most common pit fall with current Flux is working with this stream of tables that may or may not be the same structure.

### Influxdb.From is always pivoted

To go along with the single table and homogeneous schema the data returned from the `influxdb.from` function will be _pivoted_.
Specifically the _\_field_ and _\_value_ structure of the data has been removed and instead column names correspond with the name of the InfluxDB tag or field.
Like before, no distinction is made between tags and fields as columns in the data model.

A consequence of this change is that the `from` call now also requires the `measurement` parameter as the InfluxDB data model only guarantees a homogeneuos schema within the scope of a measurement.

### Syntax Sugar for From

An additional syntax is proposed that makes querying data from InfluxDB much easier to type.

The syntax is:

```
@bucket.measurement
```

The syntax is equivalent to:

```
influxdb.from(bucket:"bucket", measurement:"measurement")
```

### Syntax Sugar for Filtering

There is new syntax to make creating filter predicates easier. This syntax has two parts.

1. Declares which columns to retrieve from the database.
2. Declares a predicate for which data to retrieve from the database.


The syntax is:

```
{location,state, state == "CA"}
```

That syntax is equivalent to:

```
filter(fn: (r) => r.state == "CA")
```

This has the addition of providing explicit expectation of which columns should be returned from the database.
Currently Flux doesn't know how to describe that behavior and so it has no existing equivalent.

It is intended that this new syntax can occur anywhere within a normal Flux data pipeline.

For example this should also be valid:

```
influxdb.from(bucket:"bucket", measurement:"measurement"){location, state, state == "CA"}
```

In the above I have simply chosen to be more explicit about the `from` call instead of using the `@` shorthand notation.

### Aggregates and Selectors

With the change to a single table model, aggregates and selectors are now expressed differently.
There is a new function `aggregate` that also comes with its own syntax. More on the syntax in a bit.

The `aggregate` function has a `by` parameter which specifies the dimensions on which you would like to perform an aggregate function.
Similarly the `select` function has a `by` parameter. 


These functions now make it explicit how you want to perform an operation and that the grouping of that operation is scoped to just that operation.
No longer is there a persistent state on the table that indicates how the table (or stream of tables) are grouped.

Similarly the `aggregate` and `select` functions will have parameters for `windowing`, that enable you to define how you would like to group the data in time.

The `group` and `window` functions are not part of this new Flux vision.

### Syntax Sugar for Aggregates

Aggregates also come with their own new syntax. The intent of this syntax is to make it easy to describe which aggregate function to apply to which columns.
This syntax allows for that definition to be explicit and terse.

The syntax is:

```
 |> aggregate({min(degrees_bottom), max(degrees_bottom), mean(degrees_bottom)}, by:["location"])
```

The semantics of this syntax are to apply three different aggregate functions to the same column named `degrees_bottom`.
The output table will then have three columns named `min_degrees_bottom`, `max_degrees_bottom`, and `mean_degrees_bottom`.


>NOTE: Selectors do not need their own syntax as it is easy to specify selection in the existing syntax.

### Column Indexing

A consequence of data being represented as a single table the `tableFind` function is not needed. As such  a user can call `getColumn` directly on a table.
We can go one step further and use the `.` syntax to mean indexing a table for its column.

For example:

```
data = @foo.bar{location, state == "CA"}

data.location // a list of locations in the state of CA
```


## The How

Now that we have an overview of the new features syntax, let's discuss how I expect they can be implemented.
Again this is just to seed the conversation and is by no means a prescription on how it must be done.


### Single Table vs Stream of Tables

The current Flux internals use the interpreter to create DAG of promises to data to construct a query spec.
That spec is then planned etc and passed down to the execution engine.

So long as we can continue to build a query spec from this new Flux API, we can continue to execute the queries.
My rough plan is that a set of new Flux functions will be added to a `table` package.
The `table` package will be very similar to the current `universe` package. It will contain all basic functions and new functions as well like `aggregate`.

This new `table` package will need to provide its own mechanism for producing a query spec.
The current execution engine can handle streams of tables with heterogenous schemas.
A homogenous schema is a subset of functionality in a heterogeneous schema and as such can run on the execution engine without issue.
We will still leverage the engines concept of groups to do aggregates etc but those details will be hidden from the user.

As a rough example to illustrate how this would work, the following new Flux:


```
@data.temperatures{location,state,degrees_bottom}
 |> aggregate({min(degrees_bottom)}, by:["location"])
```

would generate a spec similar to what this Flux would generate

```
from(bucket:"data")
 |> filter(fn:(r) => r._measurement == "temperatures")
 |> group(columns:["location"])
 |> min()
 |> group()
```


### Influxdb.From is always pivoted

The `influxdb.from` function will be updated to accept a `measurement` parameter.
If it exists the spec will produce the appropriate filter.
If it doesn't exist the behavior will not change.
Once we have versioning we can then break the API of the function and require the measurement parameter.


### Syntax Sugar

Syntax sugar will be implemented as a kind of macro.
By macro I mean a source to source transformation before any semantic analysis is performed.
This means that we will be updating the parser to recognize the new syntax.
At a later stage we will convert that AST into an AST that the semantic analyzer will understand.
I don't have a clear plan on whether that is AST to AST translations or an improved process of AST to semantic graph.




### Syntax Sugar for Aggregates

The syntax sugar for aggregates needs some specific discussion.
If the `{min(bottom_degrees}` syntax is a macro, what does it expand into?

First let me propose the type signature of the `aggregate` method.


```
aggregate : forall [t0,t1] where t0 : Row, t1 : Row (table: [t0], fn:(view: [t0]) -> t1, by: [string]) -> [t1]
```

Basically `aggregate` is a function that takes in a table, an aggregation function and a list of columns on which to group.
The result is a table of the records returned from the aggregation function.

My thought is that the aggregate syntax expands to the aggregation function.

So this `{min(bottom_degrees), max(bottom_degrees), mean(bottom_degrees)}` expands to:

```
(view) => ({
    min_bottom_degrees: min(view.bottom_degrees),
    max_bottom_degrees: max(view.bottom_degrees),
    mean_bottom_degrees: mean(view.bottom_degrees),
})
```

All together then the single aggregate function call expands like this:

```
 |> aggregate({min(degrees_bottom), max(degrees_bottom), mean(degrees_bottom)}, by:["location"])
 // exapands to 
 |> aggregate(
        (view) => ({
            min_bottom_degrees: min(view.bottom_degrees),
            max_bottom_degrees: max(view.bottom_degrees),
            mean_bottom_degrees: mean(view.bottom_degrees),
        }),
        by:["location"]
    )
```

Note we are assuming a positional parameter for the `view` function.

Also notice that the `min`, `max` and `mean` functions now have different type signatures.
Instead of being functions that operate on whole tables, they are functions that operate on list (i.e. columns)

Proposed type signature for the aggregate functions.

```
min : forall [t0] where t0 : Comparable (data: [t0]) -> t0
max : forall [t0] where t0 : Comparable (data: [t0]) -> t0
mean : forall [t0] where t0 : Numeric (data: [t0]) -> float
```

Finally, there is one piece missing from this design. The group columns got dropped.
In the example you would expect the output data to contain the `location` column, but as I have described the design it would be dropped.
Obviously we need a fix for that, we could play more macro tricks or something. Looking for ideas on this part of the problem.




## Summary

In summary, we see these changes as being additive to the existing language with the addition of some new core features like macros and a new `table` package for these new function types.
