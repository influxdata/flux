#[macro_use]
extern crate criterion;
extern crate core;

use std::ffi::CString;

use criterion::{black_box, Criterion};

use core::scanner::{rust::Scan, rust::Scanner as RustScanner, Scanner, TOK_EOF};

const FLUX: &'static str = r#"from(bucket: "benchtest")
    # Here's a random comment
    |> range(start: -10m)
    |> map(fn: (r) => ({r with square: r._value * r._value}))"#;

/// Create a Scanner with pre-determined text, and scan every token
/// until EOF.
fn scanner_scan(c: &mut Criterion) {
    let cdata = CString::new(FLUX).expect("CString::new failed");
    c.bench_function("scanner.scan", |b| {
        b.iter(black_box(|| {
            let mut s = Scanner::new(cdata.clone());
            loop {
                let token = s.scan();
                if token.tok == TOK_EOF {
                    break;
                }
            }
        }));
    });
}

/// Create a Scanner with pre-determined text, and scan every token
/// until EOF. NOTE: This benchmark can be removed when the rust scanner
/// replaces the current scanner.
fn rust_scanner_scan(c: &mut Criterion) {
    let cdata = CString::new(FLUX).expect("CString::new failed");
    c.bench_function("rustscanner.scan", |b| {
        b.iter(black_box(|| {
            let mut s = RustScanner::new(cdata.clone());
            loop {
                let token = s.scan();
                if token.tok == TOK_EOF {
                    break;
                }
            }
        }));
    });
}
criterion_group!(scanner, scanner_scan, rust_scanner_scan);
criterion_main!(scanner);
