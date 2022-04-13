use super::*;

use crate::semantic::Feature;

#[test]
fn labels_simple() {
    test_infer! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        env: map![
            "fill" => "(<-tables: [{ A with B: C }], ?column: B, ?value: D) => [{ A with B: D }]
                where B: Label
                "
        ],
        src: r#"
            x = [{ a: 1 }] |> fill(column: "a", value: "x")
            y = [{ a: 1, b: ""}] |> fill(column: "b", value: 1.0)
        "#,
        exp: map![
            "x" => "[{ a: string }]",
            "y" => "[{ a: int, b: float }]",
        ],
    }
}

#[test]
fn labels_unbound() {
    test_infer! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        env: map![
            "f" => "(<-tables: [{ A with B: C }], ?value: D) => [{ A with B: D }]
                where B: Label
                "
        ],
        src: r#"
            x = [{ a: 1, b: 2.0 }] |> f(value: "x")
        "#,
        exp: map![
            "x" => "[{ a: string, b: float }]",
        ],
    }
}

#[test]
fn labels_dynamic_string() {
    test_error_msg! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        env: map![
            "fill" => "(<-tables: [{ A with B: C }], ?column: B, ?value: D) => [{ A with B: D }]
                where B: Label
                "
        ],
        src: r#"
            column = "" + "a"
            x = [{ a: 1 }] |> fill(column: column, value: "x")
        "#,
        expect: expect![[r#"
            error: string is not a label
              ┌─ main:3:31
              │
            3 │             x = [{ a: 1 }] |> fill(column: column, value: "x")
              │                               ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

        "#]],
    }
}

#[test]
fn undefined_field() {
    test_error_msg! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        env: map![
            "fill" => "(<-tables: [{ A with B: C }], ?column: B, ?value: D) => [{ A with B: D }]
                where B: Label
                "
        ],
        src: r#"
            x = [{ b: 1 }] |> fill(column: "a", value: "x")
        "#,
        expect: expect![[r#"
            error: record is missing label a
              ┌─ main:2:31
              │
            2 │             x = [{ b: 1 }] |> fill(column: "a", value: "x")
              │                               ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

        "#]],
    }
}

#[test]
fn merge_labels_to_string() {
    test_infer! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        src: r#"
            x = if 1 == 1 then "a" else "b"
            y = if 1 == 1 then "a" else "b" + "b"
            z = ["a", "b"]
        "#,
        exp: map![
            "x" => "string",
            "y" => "string",
            "z" => "[string]",
        ],
    }
}

#[test]
fn merge_labels_to_string_in_function() {
    test_infer! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        env: map![
            "same" => "(x: A, y: A) => A"
        ],
        src: r#"
            x = same(x: "a", y: "b")
            y = same(x: ["a"], y: ["b"])
        "#,
        exp: map![
            "x" => "string",
            "y" => "[string]",
        ],
    }
}

#[test]
fn optional_label() {
    test_error_msg! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        env: map![
            "columns" => "(table: A, ?column: C) => { C: string } where A: Record",
        ],
        src: r#"
            x = columns(table: { a: 1, b: "b" })
            y = x.abc
        "#,
        // TODO Improve this error
        expect: expect![[r#"
            error: A is not a label
              ┌─ main:2:17
              │
            2 │             x = columns(table: { a: 1, b: "b" })
              │                 ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

        "#]],
    }
}
