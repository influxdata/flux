# Rust implementation of a Flux parser

## Installing toolchain

Install Rust https://www.rust-lang.org/tools/install

Install  `wasm-pack`:

    $ cargo install wasm-pack

See [this](https://developer.mozilla.org/en-US/docs/WebAssembly/Rust_to_wasm) for a hello world example of using Rust with WASM and npm.

## Build WASM

Use `wasm-pack` to build an npm package from the compiled wasm code.
You need clang 1.8

    $ cd internal/rust/parser
    $ CC=clang wasm-pack build --dev --scope influxdata

### Link npm modules

This only needs to be done once to create symlinks that npm can use to consume the wasm npm package without publishing it.

    $ cd internal/rust/parser/pkg
    $ npm link
    $ cd ../../site
    $ npm link @influxdata/parser

Once that is done the `parser` dependency in the simple npm site will reference the build artifacts from `wasm-pack`.


## Run in Browser

A trivial web app has been created that loads the parser wasm module and call parse on it with a static Flux string and then console logs the parsed AST.

    $ cd internal/rust/site
    $ npm install
    $ npm link @influxdata/parser
    $ npm run serve

## Test

Use `cargo`

    $ cd internal/rust/parser
    $ cargo test



## Build Go binary
