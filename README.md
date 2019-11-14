# Flux - Influx data language

[![CircleCI](https://circleci.com/gh/influxdata/flux/tree/master.svg?style=svg)](https://circleci.com/gh/influxdata/flux/tree/master)


Flux is a lightweight scripting language for querying databases (like InfluxDB) and working with data. It's part of InfluxDB 1.7 and 2.0, but can be run independently of those.
This repo represents the language definition and an implementation of the language core.

## Specification

A complete specification can be found in [SPEC.md](./docs/SPEC.md).
The specification contains many examples to start learning Flux.

## Getting Started

Flux is currently available in InfluxDB 1.7 and 2.0 or through the REPL that can be compiled from this repository.

To compile the REPL, use the following command:

```
$ go build ./cmd/flux
$ ./flux repl
```

If you create or change any flux functions, you will need to rebuild the stdlib:
```
$ go generate ./stdlib
```

Your new Flux's code should be formatted to coexist nicely with the existing codebase with go fmt.  For example, if you add code to stdlib/universe:
```
$ go fmt ./stdlib/universe/
```

Don't forget to add your tests and make sure they work. Here is an example showing how to run the tests for the stdlib/universe package:
```
$ go test ./stdlib/universe/
```



>NOTE: The Flux REPL above does not contain the ability to connect to InfluxDB.
To connect to InfluxDB, please read the [InfluxDB 2.0](https://v2.docs.influxdata.com/v2.0/query-data/get-started/) query documentation or the [InfluxDB 1.7](http://docs.influxdata.com/flux/) documentation.

From within the REPL, you can run any Flux expression.
Additionally, you can also load a file directly into the REPL by typing `@` followed by the filename.

```
> @my_file_to_load.flux
```

### Basic Syntax

Here are a few examples of the language to get an idea of the syntax.


    // This line is a comment

    // Support for traditional math operators
    1 + 1

    // Several data types are built-in
    true                     // a boolean true value
    1                        // an int
    1.0                      // a float
    "this is a string"       // a string literal
    1h5m                     // a duration of time representing 1 hour and 5 minutes
    2018-10-10               // a time starting at midnight for the default timezone on Oct 10th 2018
    2018-10-10T10:05:00      // a time at 10:05 AM for the default timezone on Oct 10th 2018
    [1,1,2]                  // an array of integers
    {foo: "str", bar: false} // an object with two keys and their values

    // Values can be assigned to identifers
    x = 5.0
    x + 3.0 // 8.0

    // Import libraries
    import "math"

    // Call functions always using keyword arguments
    math.pow(x: 5.0, y: 3.0) // 5^3 = 125

    // Functions are defined by assigning them to identifers
    add = (a, b) => a + b

    // Call add using keyword arguments
    add(a: 5, b: 3) // 8

    // Functions are polymorphic
    add(a: 5.5, b: 2.5) // 8.0
    
    // And strongly typed
    add(a: 5, b: 2.5) // type error

    // Access data from a database and store it as an identifier
    // This is only possible within the influxdb repl (at the moment).
    import "influxdata/influxdb"
    data = influxdb.from(bucket:"telegraf/autogen")
    
    // When running inside of influxdb, the import isn't needed.
    data = from(bucket:"telegraf/autogen")

    // Chain more transformation functions to further specify the desired data
    cpu = data 
        // only get the last 5m of data
        |> range(start: -5m)
        // only get the "usage_user" data from the _measurement "cpu"
        |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_user")

    // Return the data to the client
    cpu |> yield()

    // Group an aggregate along different dimensions
    cpu
        // organize data into groups by host and region
        |> group(columns:["host","region"])
        // compute the mean of each group
        |> mean()
        // yield this result to the client
        |> yield()

    // Window an aggregate over time
    cpu
        // organize data into groups of 1 minute
        // compute the mean of each group
        |> aggregateWindow(every: 1m, fn: mean)
        // yield this result to the client
        |> yield()

    // Gather different data
    mem = data 
        // only get the last 5m of data
        |> range(start: -5m)
        // only get the "used_percent" data from the _measurement "mem"
        |> filter(fn: (r) => r._measurement == "mem" and r._field == "used_percent")


    // Join data to create wider tables and map a function over the result
    join(tables: {cpu:cpu, mem:mem}, on:["_time", "host"])
        // compute the ratio of cpu usage to mem used_percent
        |> map(fn:(r) => {_time: r._time, _value: r._value_cpu / r._value_mem)
        // again yield this result to the client
        |> yield()

The above examples give only a taste of what is possible with Flux.
See the complete [documentation](https://v2.docs.influxdata.com/v2.0/query-data/get-started/) for more complete examples and instructions for how to use Flux with InfluxDB 2.0.
