#![allow(non_snake_case)] // Tests are testing non-snake case functions.
#[macro_use]
extern crate criterion;
extern crate flux;

use criterion::Criterion;
use flux::parser::Parser;

fn strings_replaceAll(c: &mut Criterion) {
    let flux = include_str!("../../../../stdlib/strings/replaceAll_test.flux");
    c.bench_function("stdlib/strings/replaceAll_test.flux", |b| {
        b.iter(|| {
            let mut parser = Parser::new(flux);
            parser.parse_file("".to_string());
        })
    });
}

fn strings_title(c: &mut Criterion) {
    let flux = include_str!("../../../../stdlib/strings/title_test.flux");
    c.bench_function("stdlib/strings/title_test.flux", |b| {
        b.iter(|| {
            let mut parser = Parser::new(flux);
            parser.parse_file("".to_string());
        })
    });
}

fn strings_trim(c: &mut Criterion) {
    let flux = include_str!("../../../../stdlib/strings/trim_test.flux");
    c.bench_function("stdlib/strings/trim_test.flux", |b| {
        b.iter(|| {
            let mut parser = Parser::new(flux);
            parser.parse_file("".to_string());
        })
    });
}

fn strings_toUpper(c: &mut Criterion) {
    let flux = include_str!("../../../../stdlib/strings/toUpper_test.flux");
    c.bench_function("stdlib/strings/toUpper_test.flux", |b| {
        b.iter(|| {
            let mut parser = Parser::new(flux);
            parser.parse_file("".to_string());
        })
    });
}

fn strings_substring(c: &mut Criterion) {
    let flux = include_str!("../../../../stdlib/strings/substring_test.flux");
    c.bench_function("stdlib/strings/substring_test.flux", |b| {
        b.iter(|| {
            let mut parser = Parser::new(flux);
            parser.parse_file("".to_string());
        })
    });
}

fn strings_toLower(c: &mut Criterion) {
    let flux = include_str!("../../../../stdlib/strings/toLower_test.flux");
    c.bench_function("stdlib/strings/toLower_test.flux", |b| {
        b.iter(|| {
            let mut parser = Parser::new(flux);
            parser.parse_file("".to_string());
        })
    });
}

fn strings_replace(c: &mut Criterion) {
    let flux = include_str!("../../../../stdlib/strings/replace_test.flux");
    c.bench_function("stdlib/strings/replace_test.flux", |b| {
        b.iter(|| {
            let mut parser = Parser::new(flux);
            parser.parse_file("".to_string());
        })
    });
}

fn strings_length(c: &mut Criterion) {
    let flux = include_str!("../../../../stdlib/strings/length_test.flux");
    c.bench_function("stdlib/strings/length_test.flux", |b| {
        b.iter(|| {
            let mut parser = Parser::new(flux);
            parser.parse_file("".to_string());
        })
    });
}

fn strings_subset(c: &mut Criterion) {
    let flux = include_str!("../../../../stdlib/strings/subset_test.flux");
    c.bench_function("stdlib/strings/subset_test.flux", |b| {
        b.iter(|| {
            let mut parser = Parser::new(flux);
            parser.parse_file("".to_string());
        })
    });
}

criterion_group!(
    strings,
    strings_replaceAll,
    strings_title,
    strings_trim,
    strings_toUpper,
    strings_substring,
    strings_toLower,
    strings_replace,
    strings_length,
    strings_subset
);
criterion_main!(strings);
