// build.rs

use std::env;
use std::fs::copy;
use std::path::PathBuf;

// Bring in a dependency on an externally maintained `cc` package which manages
// invoking the C compiler.
extern crate cc;

fn main() {
    let out_path = PathBuf::from(env::var("OUT_DIR").unwrap());

    // The bindgen::Builder is the main entry point
    // to bindgen, and lets you build up options for
    // the resulting bindings.
    let bindings = bindgen::Builder::default()
        // The input header we would like to generate
        // bindings for.
        .header("src/scanner/scanner.h")
        // Finish the builder and generate the bindings.
        .generate()
        // Unwrap the Result and panic on failure.
        .expect("Unable to generate bindings");

    // Write the bindings to the $OUT_DIR/bindings.rs file.
    bindings
        .write_to_file(out_path.join("bindings.rs"))
        .expect("Couldn't write bindings!");

    let ctypes = bindgen::Builder::default()
        .header("include/influxdata/flux.h")
        .generate()
        .expect("Unable to generate c type bindings");

    ctypes
        .write_to_file(out_path.join("ctypes.rs"))
        .expect("Couldn't write c type bindings!");

    copy("scanner.c", out_path.join("scanner.c")).expect("Could not copy scanner.c");

    // Compile generated scanner
    cc::Build::new()
        .include("src/scanner")
        .file(out_path.join("scanner.c"))
        .compile("scanner");
}
