#[macro_use]
extern crate criterion;
extern crate flux;

use criterion::{black_box, Criterion};
use flux::formatter::format;

fn format_everything(c: &mut Criterion) {
    let flux = include_str!("./everything.flux");
    c.bench_function("format_everything.flux", |b| {
        b.iter(black_box(|| {
            format(flux).unwrap();
        }));
    });
}

criterion_group!(formatter, format_everything,);
criterion_main!(formatter);
