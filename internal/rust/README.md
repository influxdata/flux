# Rust implementation of a Flux parser

## Installing toolchain

Install Rust https://www.rust-lang.org/tools/install

Install  `wasm-pack`:

    $ cargo install wasm-pack

See [this](https://developer.mozilla.org/en-US/docs/WebAssembly/Rust_to_wasm) for a hello world example of using Rust with WASM and npm.

## Build WASM

Use `wasm-pack` to build an npm package from the compiled wasm code.

    $ cd internal/rust/parser
    $ wasm-pack build --dev --scope influxdata

### Link npm modules

This only needs to be done once to create symlinks that npm can use to consume the wasm npm package without publishing it.

    $ cd internal/rust/parser/pkg
    $ npm link
    $ cd ../../site
    $ npm install
    $ npm link @influxdata/parser

Once that is done the `parser` dependency in the simple npm site will reference the build artifacts from `wasm-pack`.

>NOTE: The `npm install` command will destroy the link. So if you run `npm install` again you must rerun the `npm link @influxdata/parser` command in the `site` directory.


## Run in Browser

A trivial web app has been created that loads the parser wasm module and call parse on it with a static Flux string and then console logs the parsed AST.

    $ npm run serve

Navigate to http://localhost:8080 to try it out.
This will watch the filesystm and rebuild on changes.
As such you should be able to run `wasm-pack` to get new changes and then refresh the browser to test.

## Test

Use `cargo`

    $ cd internal/rust/{parser,scanner,ast}
    $ cargo test


## Build Go binary

TODO There is nothing to build for Go yet.
