# Flux - Influx data language

[![CircleCI](https://circleci.com/gh/influxdata/flux/tree/master.svg?style=svg)](https://circleci.com/gh/influxdata/flux/tree/master)


Flux is a lightweight scripting language for querying databases (like [InfluxDB](https://github.com/influxdata/influxdb)) and working with data.
It is part of InfluxDB 1.7 and 2.0, but can be run independently of those.
This repository contains the language definition and an implementation of the language core.

## Specification

A complete specification can be found in [SPEC.md](./docs/SPEC.md).
The specification contains many examples to start learning Flux.

## Requirements

Building Flux requires the following:

* Go 1.16 or greater with module support enabled
* Latest stable version of Rust and Cargo (can be installed with [rustup](https://rustup.rs/))
* Clang

## Getting Started

Flux is currently available in InfluxDB 1.7 and 2.0, or through the REPL that can be compiled from this repository.

To build Flux, first install the GNU `pkg-config` utility on your system, then ensure the `pkg-config` wrapper is also installed.

```
# First install GNU pkg-config.
# On Debian/Ubuntu
$ sudo apt-get install -y clang pkg-config

# Or on Mac OS X with Homebrew
$ brew install pkg-config

# Next, install the pkg-config wrapper utility
$ go get github.com/influxdata/pkg-config

# Optionally, add the GOBIN directory to your PATH
$ export PATH=${GOPATH}/bin:${PATH}
```

If GOBIN is in your PATH, ensure that `pkg-config` is configured correctly by using `which -a`.

```
$ which -a pkg-config
/home/user/go/bin/pkg-config
/usr/bin/pkg-config
```

To compile and use the REPL, use the following command to run the repl with suggestions enabled:

```
$ go build ./cmd/flux
$ ./flux --enable-suggestions
```

Alternatively, because the `pkg-config` wrapper may not work in all projects you may not want to add the wrapper `pkg-config` to your `PATH`. In this case you can set `PKG_CONFIG` and Go will use it. Eg, to build and install to ${GOPATH}/bin using `PKG_CONFIG`:

```
$ PKG_CONFIG=/home/user/go/bin/pkg-config go install ./cmd/flux
$ ${GOPATH}/bin/flux repl
```

From within the REPL, you can run any Flux expression.
You can also load a file directly into the REPL by typing `@` followed by the filename.

```
> @my_file_to_load.flux
```

## Basic Syntax

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

    // Values can be assigned to identifiers
    x = 5.0
    x + 3.0 // 8.0

    // Import libraries
    import "math"

    // Call functions always using keyword arguments
    math.pow(x: 5.0, y: 3.0) // 5^3 = 125

    // Functions are defined by assigning them to identifiers
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

## Contributing

Flux welcomes contributions to the language and the runtime.

If you are interested in contributing, please read the [contributing guide](https://github.com/influxdata/flux/blob/master/CONTRIBUTING.md) for more information.

### Development basics

If you modify any Rust code, you will need to force Go to rebuild the library.

```bash
$ go generate ./libflux/go/libflux
```

If you create or change any Flux functions, you will need to rebuild the stdlib and inform Go that it must rebuild libflux:
```bash
$ go generate ./stdlib ./libflux/go/libflux
```

Your new Flux's code should be formatted to coexist nicely with the existing codebase with go fmt.  For example, if you add code to stdlib/universe:
```bash
$ go fmt ./stdlib/universe/
```

Don't forget to add your tests and make sure they work. Here is an example showing how to run the tests for the stdlib/universe package:
```bash
$ go test ./stdlib/universe/
```

If you modify the flatbuffer files you need to run `make generate` to update the generated bindings. To ensure that you have the correct version of `flatc` you may use `./install_flatc.sh`.

```bash
$ ./install_flatc.sh
$ make generate
```
