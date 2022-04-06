#[macro_use]
extern crate criterion;
extern crate flux;

use criterion::{black_box, Criterion};

fn analyze_mean(c: &mut Criterion) {
    let mut group = c.benchmark_group("analyze");
    group.bench_function("analyze_mean", |b| {
        let source = r#"
from(bucket:"test")
|> range(start: 2022-01-12T13:03:43.223Z, stop: 2022-03-22T11:03:12.223Z)
|> filter(fn: (r) => r._measurement == "test" and r.id == "MYID")
|> mean()"#;
        let pkg = flux::parse("".into(), source);
        b.iter(black_box(|| {
            flux::analyze(&pkg).unwrap();
        }));
    });
    group.finish();
}

criterion_group!(analyze, analyze_mean);
criterion_main!(analyze);
