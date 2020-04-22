# Rust implementation of a Flux parser

## Development dependencies

- [Install the Rust toolchain](https://www.rust-lang.org/tools/install)

- Install `wasm-pack`:

  ```
  $ cargo install wasm-pack
  ```

- If you wish to test the WASM target locally, install [Node.js](https://nodejs.org/en/download/package-manager/) and [Yarn](https://yarnpkg.com/en/docs/install)

See [this](https://developer.mozilla.org/en-US/docs/WebAssembly/Rust_to_wasm) guide for a hello world example of using Rust with WASM.

## Building the WASM package

Use `wasm-pack` to build an npm package from the compiled WASM code.
You will need to use the `clang` compiler at least version 8.

### Linux

Use your distributions package manager to install clang.

    $ cd libflux
    $ CC=clang wasm-pack build --scope influxdata --dev

### MacOS

MacOS doesn't appear to have a functional version of clang that will work.
As such we have created a Dockerfile to abstract these dependencies.
To use it run the `build.sh` script which will run all the build command inside the docker container.

    $ ./libflux/build.sh --dev

> NOTE: The docker image uses a local volume mount at `./libflux/.cache` to cache Rust/WASM build artifacts to make builds faster inside the container.

## Testing the built WASM package locally

A trivial web app has been created that will load the WASM parser module, call it with a static Flux string, and then log the parsed AST.

Before running it for the first time, you'll have to follow these steps:

1. Change into the web app directory:

   ```
   $ cd libflux/site
   ```

2. Install its dependencies:

   ```
   $ yarn
   ```

3. Replace the published `@influxdata/flux-parser` dependency with a symlink to your locally built WASM package:

   ```
   $ cd ../pkg
   $ yarn link
   $ cd ../site
   $ yarn link @influxdata/flux
   ```

Now you should be able to run the web app:

    $ yarn serve

Navigate to http://localhost:8080 to try it out.
This will watch the filesystem and rebuild on changes.
As such you should be able to run `wasm-pack` to get new changes and then refresh the browser to test.

## Updating NodeJS Dependencies

On occasion a vulnerability is found in one of the nodejs dependencies of the WASM package.
To update the vulnerable dependencies do the following

1. Change into the web app directory:

   ```
   $ cd libflux/site
   ```

2. Install its dependencies locally so you can introspect them:

   ```
   $ yarn install
   ```

3. Check for outdated dependencies

   ```
   $ npm outdated
   ```

4. Update any dependecies listed as outdated. For example the `webpack` dependencies is outdate run:

   ```
   $ npm update webpack
   ```

5. Finally update the yarn.lock file:

   ```
   $ yarn install
   ```

Now you can commit the new `package.json` and `yarn.lock` files.

## Publishing the WASM package

1. Log into yarn (`yarn login`) with an account that has access to the [influxdata npm organization](https://www.npmjs.com/org/influxdata)

2. Bump the version in `libflux/Cargo.toml`

3. Run the publish script:

   ```
   $ ./libflux/publish.sh
   ```

   Note that this will create a build optimized for size using the Docker-based process.

## Test

Use `cargo`

    $ cd libflux
    $ cargo test


## Build Go binary

TODO There is nothing to build for Go yet.
