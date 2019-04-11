# Flux as a Library Example

This folder contains a [main.go](main.go) file that shows how flux can be used as a library in your programs.

## Components

The main components you need are:

- The interpreter;
- The scope (aka Prelude);
- The builtin library and your additional functions, if you want to define them;
- The language specification compiler;
- A querier.

## What does this example do?

This example takes all the components described above from the Flux repo and uses them to
compile, execute and print the results of this query.

```flux
import g "generate"
g.from(start: 1993-02-16T00:00:00Z, stop: 1993-02-16T00:03:00Z, count: 5, fn: (n) => 1)
```

The `generate.from` function is not defined by the `builtin` library but it is in the `generate` package,
you can define your own functions by registering them. You can see how that is done in `stdlib/generate/from.go` at `init`.

Since using `generate.from` function as a data source we are completely decoupled from InfluxDB, so, this example
does not require a working installation of InfluxDB.

To implement your own data source, have a look at the various `from.go` files in `stdlib` sub-folders.

## Usage

From this directory:

```
go test -v .
```
