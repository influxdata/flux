// build.rs

use std::env;
use std::process::Command;

// Bring in a dependency on an externally maintained `cc` package which manages
// invoking the C compiler.
extern crate cc;

fn main() {
    // Run Ragel
    let out_dir = env::var("OUT_DIR").unwrap();
    Command::new("ragel")
        .args(&[
            "-C",
            "-o",
            &format!("{}/scanner.c", out_dir),
            "src/scanner.rl",
        ])
        .status()
        .unwrap();
    // Compile generated scanner
    cc::Build::new()
        .file(format!("{}/scanner.c", out_dir))
        .compile("hello");
}
