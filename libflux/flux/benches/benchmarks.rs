#[allow(soft_unstable)]
extern crate criterion;

use core::scanner::Scanner;
use criterion::{criterion_group, criterion_main, Criterion};

// run only this benchmark using `cargo bench` from the current directory
fn bench_scanner(c: &mut Criterion) {
    let shorter = "from(bucket:\"foo\") |> range(start: -1m)";
    let mut s = Scanner::new(shorter);

    c.bench_function("scan_short_text", |b| b.iter(|| s.scan()));

    let mut sc = Scanner::new(LONGER);

    c.bench_function("scan_long_text", |b| b.iter(|| sc.scan()));
}

criterion_group!(benches, bench_scanner);
criterion_main!(benches);

//copied from everything.flux

const LONGER: &'static str = include_str!("everything.flux");
