#[macro_use]
extern crate criterion;
extern crate fluxcore;

use criterion::{black_box, Criterion};

use fluxcore::scanner;

const FLUX: &'static str = r#"from(bucket: "benchtest")
    # Here's a random comment
    |> range(start: -10m)
    |> map(fn: (r) => ({r with square: r._value * r._value}))"#;

/// Create a Scanner with pre-determined text, and scan every token
/// until EOF.
fn scanner_scan(c: &mut Criterion) {
    c.bench_function("scanner.scan", |b| {
        b.iter(black_box(|| {
            let mut s = scanner::Scanner::new(FLUX);
            loop {
                let token = s.scan();
                if token.tok == scanner::TokenType::Eof {
                    break;
                }
            }
        }));
    });
}

criterion_group!(scanner, scanner_scan);
criterion_main!(scanner);
