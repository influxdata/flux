# Rust implementation of a Flux parser

## Installing toolchain

Install Rust https://www.rust-lang.org/tools/install

Install  `wasm-pack`:

    $ cargo install wasm-pack

See [this](https://developer.mozilla.org/en-US/docs/WebAssembly/Rust_to_wasm) for a hello world example of using Rust with WASM and npm.

## Build WASM

Use `wasm-pack` to build an npm package from the compiled wasm code.
You will need to use the `clang` compiler at least version 8.

### Linux

Use your distributions package manager to install clang.

    $ cd internal/rust/parser
    $ CC=clang wasm-pack build --dev --scope influxdata

### MacOS

MacOS doesn't appear to have a functional version of clang that will work.
As such we have created a Dockerfile to abstract these dependecies.
To use it run the `build.sh` script which will run all the build command inside the docker contianer.

    $ ./internal/rust/build.sh

> NOTE: The docker image uses a local volume mount at `./internal/rust/.cache` to cache rust/wasm build artifacts to make builds faster inside the container.

## Link npm modules

This only needs to be done once to create symlinks that npm can use to consume the wasm npm package without publishing it.

    $ cd internal/rust/parser/pkg
    $ npm link
    $ cd ../../site
    $ # edit package.json and remove the `dependencies` section
    $ npm install
    $ # edit package.json and re-add the `dependencies` section, use `git checkout package.json` to quickly revert.
    $ npm link @influxdata/parser

Once that is done the `parser` dependency in the simple npm site will reference the build artifacts from `wasm-pack`.

> NOTE: The `npm install` command will destroy the link.
So if you run `npm install` again you must rerun the `npm link @influxdata/parser` command in the `site` directory.

> NOTE: The `npm install` command will fail if the `@influxdata/parser` dependecy is listed because the depencies doesn't exist publicly.
This will prevent npm from installing the needed dev dependencies.
A quick hack is to delete the `dependencies` section from `internal/rust/site/package.json` and then run `npm install`.
Once that has passed you can re-add the dependcies and run `npm link @influxdata/parser`.


## Run in Browser

A trivial web app has been created that loads the parser wasm module and call parse on it with a static Flux string and then console logs the parsed AST.

    $ npm run serve

Navigate to http://localhost:8080 to try it out.
This will watch the filesystem and rebuild on changes.
As such you should be able to run `wasm-pack` to get new changes and then refresh the browser to test.

## Test

Use `cargo`

    $ cd internal/rust/{parser,scanner,ast}
    $ cargo test


## Build Go binary

TODO There is nothing to build for Go yet.
