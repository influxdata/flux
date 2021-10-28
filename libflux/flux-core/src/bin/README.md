# Fluxdoc

We have a fluxdoc command that can lint Flux source code for correct documentation.

## Usage

First compile the stdlib to a temporary directory. Starting from the root of the Flux repo run these commands.

    cd libflux
    cargo run --bin fluxc -- stdlib -o ../stdlib-compiled -s ../stdlib

That will take about a minute and only needs to be run again if the types of Flux functions change.

Second run the fluxdoc lint command

    cargo run --bin fluxdoc --features=doc -- lint -s ../stdlib-compiled -d <path/to/flux/directory/to/lint>

A list of errors will be reported otherwise the output will be empty.

## Updating Exceptions List

In the file `libflux/flux-core/src/bin/fluxdoc.rs` at the end is a list of packages that are exceptions.
Once a package is passing lint it should be removed from that list.

