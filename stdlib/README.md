# Flux Standard Library


This directory contains the implementation of the Flux standard library.

## Code Organization


A Flux package is represented as a directory containing at least one file with a `.flux` extension.
The package name is declared within the file using the `package` clause.
Package names should be the same name as the directory.

Test files may placed in the same directory as the packing using the `_test.flux` suffix on the file name.
The test files must have a `_test` suffix for the package name and the prefix must match the name of the non test package.

Because the above mirrors the Go package structure it is common to also have `.go` file and `_test.go` files that mirror the `.flux` files.


A typical Flux package structure:


```
stdlib/strings/
├── flux_gen.go
├── flux_test_gen.go
├── replaceAll_test.flux
├── replace_test.flux
├── strings.flux
├── strings.go
├── strings_test.go
├── subset_test.flux
├── title_test.flux
├── toLower_test.flux
└── toUpper_test.flux
```


The files `flux_gen.go` and `flux_test_gen.go` are generated using the `builtin` command.
They contain the AST of the various `.flux` files in the package as Go structs.
All `*_test.flux` files are encoded into `flux_test_gen.go` the non test `.flux` files are encoded in `flux_gen.go`

> NOTE: The `flux_test_gen.go` file is not a Go test file as we want to include the Flux test code into the normal build.
This enables downstream projects that import Flux to run the test suite define in the standard library against their implementation.


## Third Party Contributions

We collect third part contributions into the `contrib` package.
See the [README](https://github.com/influxdata/flux/blob/master/stdlib/contrib/README.md) for details on how to contribute a third party package to Flux.
