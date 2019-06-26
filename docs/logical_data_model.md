# Flux Language Logical Data Model

The flux data model can be analyzed from two viewpoints: physical and logical.  The physical data model perspective considers the functional layout of data elements in physical memory, while the logical model considers the conceptual understanding necessary for a user of flux to write effective queries.  It is important to note that the physical data model may, at times, differ significantly from the logical model so long as it provides a correct interface to the underlying data. 

This document details the logical perspective, and should inform users what to expect as input and output from their queries.  

## Built-in Data Types In Flux
Flux provides a number of convenient built-in data types that have come to be expected in modern programming languages.  This section is meant to give a conceptual overview of the base types in order to provide background for the discussion around the flux query data model.  More info may be found on these types in the [flux language spec](https://github.com/influxdata/flux/blob/master/docs/SPEC.md). 

### Primitive Data Types

In total, flux has 8 data types that may be expressed as a literal constant: 
* Integer
* Unsigned Integer
* Float
* Boolean
* String
* Regular Expression
* Duration
* Time

### Function Types 
All functions in flux may be expressed anonymously and stored in a named variable.  For discussion in this document it's understood that all functions have a zero or more strongly-typed _input_ parameters, and exactly one strongly-typed return value.  
abs = (x) => { return if x < 0 then  -x else x }

### Array Types
Flux supports homogeneous-typed arrays of any type: 
```
a = [1,2,4,3,9]
```

### Object Types

Flux implements primitive objects as key-value containers: 
```
obj = {a: 1.0, b: 2.5, f: (x,y) => x + y/2}
```
While functions may be embedded in an object like any other type, they are not object methods: they are not scoped to implicitly access the other data members in the object instance.

The flux type system will implicitly determine the type of an object from its keys and values.  For example, the object: 
```
o = {a: 4.0, b: "xy", c: true, d: 3}
```
will have the type schema: 
```
{a: float, b: string, c: boolean, d: integer}
```
Two objects are considered to be compatible types if they share a common subset of name/type pairs, so that: 
```
{a: 1.0, b: "extra member"}
``` 
would be a valid assignment in any circumstance where an object of type `{a:float}` is required (e.g. as a function input parameter)

### Complex Types
Objects and arrays may be nested: 
```
o2 = {a: 4.0,  b: {x: 2, y: 3}, c: ["f", "c"], d: [{a:1, b:2}, {a:3, b: 4}]}
```
It is important to remember that all array elements must have a single identical type for all elements.  In the context of objects, all elements must share a common subset of key/value-type pairs.  This necessarily implies that all objects in an array must have identical key/value-type sets, with no one object being a super/sub-set of the others. 

### Record Types
A final background needed for understanding table types is the record type, which represents a single row of data from a flux table.  
A record is not a special, distinct type in flux, but conceptually we define it as an object whose key/value-type signature is restricted to the following simple types: 
* Integer
* Unsigned Integer
* Float
* Boolean
* String
* Time

## Flux Table Data Model
A flux table is, analogous to a relation in SQL.  It is a collection of data arranged into rows and columns, where each column has a single defined type.  A table may be arranged row-wise or column-wise: 

* A row-wise table collects data into horizontal arrays containing one value for each column on a row.  Multiple rows (all of equal length) are then held in a single collection to represent a table.  
* A column-wise table collects data into vertical arrays containing all values for a single column.  Multiple columns (all of equal length) are then collected to represent a single table.

In general, the arrangement of a table is hidden from the user: it's always possible to extract a single row or a single column from a table regardless of its arrangement.  It's more useful to understand that tables may also be _constructed_ from row collections or column collections. 

### Row Collections
A row collection is an array of records. The following collection:  
```
rowCollection = [{a: 1, b: 2}, {a: 4, b: 5}]
```
represents the table:

| a 	| b 	|
|---	|---	|
|  1 	|   2	|
|  4 	|   5	|

### Column Collections
A column collection is an object of name/value-array pairs: 
``` 
columnCollection = {a: [1,4], b: [2,5]}
```
represents the table:

| a 	| b 	|
|---	|---	|
|  1 	|   2	|
|  4 	|   5	|

## Sepcial Table Types
In simplest terms, a Flux table is a rectangle of data.  The more interesting concept in Flux is how we characterize these simple tables, and collect them into streams.

### Correlated Tables
A correlated table is a simple table in which all _rows_ have identical values among a set of _correlation columns_.  Column-wise, this would mean that one entire column is conceptually filled with an identical value.  

Correlated tables are useful for representing batches of data that qualitatively belong together.  For example, if we are collecting disk usage for our server hosts that are organized into regions, we might have a correlated table to represent a single host in a single region: 

| host 	    | region	| disk-percent |
|---	    |---	    | ---          |
|  HostA 	|   us-east	|  55          |

This may by too fine-grained for a regional analysis, so flux enables you to re-group your data into a new correlation: 

| host 	    | region	| disk-percent |
|---	    |---	    | ---          |
|  HostA 	|   us-east	|  55          |
|  HostA 	|   us-east	|  72          |
|  HostB 	|   us-east	|  99          |
|  HostC 	|   us-east	|  21          |

Note in the example above, we see a special case, that any table with a single row is by definition correlated. Another special case, discussed further below, is when the correlation columns are an empty set. 

### Streams and Series
A _stream_ is a single unbounded table.  Many applications, such as server monitoring or sensor readings, are continuously producing new data, so that conceptually a table representing such data has infinite length over time.  

A _series_ is a stream that is deterministically ordered according to the value of one or more columns.  A _time series_ is one in which the ordering column is filled with time values:  

| host 	    | region	| disk-percent | time |
|---	    |---	    | ---          | ---  |
|  HostA 	|   us-east	|  55          | 7:00 |
|  HostA 	|   us-east	|  65          | 7:05 |
|  HostA 	|   us-east	|  75          | 7:10 |
|  HostA 	|   us-east	|  70          | 7:15 |
|...	    |...	    | ...          | ...  |

A correlated time series is one in which the series table is correlated.  

While streams and series are indeed distinct, the majority of practical data streams found in real world data sets are series.  Therefore, the remainder of this document discusses series in isolation, though some of the concepts may also apply to unordered streams. 

### Correlated Series Sets
A _correlated series set_ is a set of correlated series.  They may only be created by re-grouping rows from an existing series or series set according to a values obtained from a common set of correlation columns.  Specifically, a correlated series set cannot be created by collecting multiple arbitrary series. 
A consequence of this requirement is that all Correlated Series in a set have the same column names and types.  Correlated series sets are further constrained by requiring that no two series in the set may have the same correlation column values.  If such a case exists, the two tables should be merged.  

Example: 

| host 	    | region	| disk-percent | time |
|---	    |---	    | ---          | ---  |
|  HostA 	|   us-east	|  55          | 7:00 |
|  HostA 	|   us-east	|  65          | 7:05 |
|...	    |...	    | ...          | ...  |

| host 	    | region	| disk-percent | time |
|---	    |---	    | ---          | ---  |
|  HostB 	|   us-east	|  75          | 7:00 |
|  HostB 	|   us-east	|  70          | 7:05 |
|...	    |...	    | ...          | ...  |

As a special case, a correlated series set may have a single table that is "correlated" by an empty set of columns (i.e., all rows have the empty set in common).

### Windowed Series
A windowed series is one that is one that is bounded by a maximum and minimum value on the series' ordering column(s). By convention, flux uses the column names `_start` and `_stop` to indicate bounds: 
 
 Example: 
 
 | host 	| region	| disk-percent | time | _start | _stop |
 |---	    |---	    | ---          | ---  | ---    | ---   |
 |  HostA 	|   us-east	|  55          | 7:00 | 7:00   | 7:30  |
 |  HostA 	|   us-east	|  65          | 7:05 | 7:00   | 7:30  |
 
 By representing the bounds as columns on the windowed series, it is by definition correlated, at least, by the table bounds `_start` and `_stop`.  
  
### Windowed Series Set
A windowed series set is similar to a correlated series set, but differ in how they are constructed.  Similar to a correlated series set, a windowed series set is created from an input series.
The difference is that to construct a windowed series, we apply the following algorithm: 
1. Determine one or more valid _start/_stop bounds pairs defined on the series' ordering column(s).  The pairs may overlap, such as `[7:00, 7:30], [7:15, 7:45]`.  Further, the pairs need not be uniform in distribution, such as `[3:00,4:00], [10:00, 10:01]`. 
2. Create an empty series for each bounds pair
3. For each row of an existing series, copy that row into any corresponding series for which its value columns are within the bounds.  

Note that the 'copy' operation may be virtual, but conceptually speaking, the same input row may appear more than once in the windowed output because windowed series are permitted to overlap.  
Finally, a Windowed series set is by definition a special type of correlated series set, and therefore has all of the same properties.  

### Mixed Series Set
A mixed series set is the most flexible type of data in flux, but it's also the hardest to work with, as we'll see in the next section.  A mixed series set is created by interleaving, in no specific order, 2 or more Correlated Series sets.  
That is, it is a sequence of tables in which the table schemas will arbitrarily change over the course of the sequence. In general, it will be the query designer's job to know that a series is mixed, and to write their query accordingly to handle surprise transitions in table schema.  

## Series Transformations
Flux accepts queries that transform an input correlated series set into a new correlated series set.  The changes may involve one or more of: 
1.  Adding/removing/renaming columns from all series
2.  Removing (filtering) rows from one or more series
3.  Updating one or more column values in one or more series
4.  Computing + Adding New Rows to a new or existing series
5.  Changing window bound sets
6.  Regrouping into a new correlated set

All transformations, however complex, may be expressed as a combination of these 6 operations.  All flux transformations are written in the scope of a single, correlated series.  The transformation will then be applied once to each series in the correlated series set.  

Within a series set, each transformation call may result in adding rows to either a new series that it creates, or else a series that was previously created as a result of calling the transformation on an earlier series in the set.  