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
            b = "b"
            z = [{ a: 1, b: ""}] |> fill(column: b, value: 1.0)
        "#,
        exp: map![
            "b" => "\"b\"",
            "x" => "[{ a: string }]",
            "y" => "[{ a: int, b: float }]",
            "z" => "[{ a: int, b: float }]",
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
            "x" => "[{A with a:int, B:string}] where B: Label",
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
            error: string is not Label (argument column)
              ┌─ main:3:44
              │
            3 │             x = [{ a: 1 }] |> fill(column: column, value: "x")
              │                                            ^^^^^^

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
fn attempt_to_use_label_polymorphism_without_feature() {
    test_error_msg! {
        env: map![
            "columns" => "(table: A, ?column: C) => { C: string } where A: Record, C: Label",
        ],
        src: r#"
            x = columns(table: { a: 1, b: "b" }, column: "abc")
            y = x.abc
        "#,
        expect: expect![[r#"
            error: string is not Label (argument column)
              ┌─ main:2:58
              │
            2 │             x = columns(table: { a: 1, b: "b" }, column: "abc")
              │                                                          ^^^^^

            error: record is missing label abc
              ┌─ main:3:17
              │
            3 │             y = x.abc
              │                 ^

        "#]],
    }
}
#[test]
fn columns() {
    test_infer! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        env: map![
            "stream" => "stream[{ a: int }]",
            "map" => "(<-tables: stream[A], fn: (r: A) => B) => stream[B]"
        ],
        imp: map![
            "experimental/universe" => package![
                "fill" => "(<-tables: stream[{A with C: B}], ?column: C, ?value: B, ?usePrevious: bool) => stream[{A with C: B}]
        where
        A: Record,
        C: Label",
                "columns" => "(<-tables: stream[A], ?column: C) => stream[{ C: string }] where A: Record, C: Label",
            ],
        ],
        src: r#"
            import "experimental/universe"

            x = stream
                |> universe.columns(column: "abc")
                |> map(fn: (r) => ({ x: r.abc }))
        "#,
        exp: map![
            "x" => "stream[{ x: string }]",
        ],
    }
}

#[test]
fn optional_label_defined() {
    test_infer! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        env: map![
            "columns" => r#"(table: A, ?column: C = "abc") => { C: string } where A: Record, C: Label"#,
        ],
        src: r#"
            x = columns(table: { a: 1, b: "b" })
            y = x.abc
        "#,
        exp: map![
            "x" => "{ abc: string }",
            "y" => "string",
        ],
    }
}

#[test]
fn label_types_are_preserved_in_exports() {
    test_infer! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        src: r#"
            builtin elapsed: (?timeColumn: T = "_time") => stream[{ A with T: time }]
                    where
                    A: Record,
                    T: Label
        "#,
        exp: map![
            "elapsed" => r#"(?timeColumn:A = "_time") => stream[{B with A:time}] where A: Label, B: Record"#
        ],
    }
}

#[test]
fn optional_label_undefined() {
    test_error_msg! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        env: map![
            "columns" => "(table: A, ?column: C) => { C: string } where A: Record, C: Label",
        ],
        src: r#"
            x = columns(table: { a: 1, b: "b" })
            y = x.abc
        "#,
        // TODO This fails because `column` is not specified but it ought to provide a better error
        expect: expect![[r#"
            error: record is missing label abc
              ┌─ main:3:17
              │
            3 │             y = x.abc
              │                 ^

        "#]],
    }
}

#[test]
fn default_arguments_do_not_try_to_treat_literals_as_strings_when_they_must_be_a_label() {
    test_infer! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        env: map![
            "max" => r#"(<-tables: stream[{ A with L: B }], ?column: L) => stream[{ A with L: B }]
                where A: Record,
                      B: Comparable,
                      L: Label"#,
        ],
        src: r#"
            f = (
                column="_value",
                tables=<-,
            ) =>
                tables
                    // `column` would be treated as `string` instead of a label when checking the
                    // default arguments
                    |> max(column: column)
        "#,
        exp: map![
            "f" => r#"(<-tables:stream[{C with A:B}], ?column:A) => stream[{C with A:B}] where A: Label, B: Comparable, C: Record"#,
        ],
    }
}

#[test]
fn constraints_propagate_fully() {
    test_infer! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        env: map![
            "aggregateWindow" => r#"(fn:(<-:stream[B], column:C) => stream[D], ?column:C) => stream[E]"#,
            "max" => r#"(<-tables: stream[{ A with L: B }], ?column: L) => stream[{ A with L: B }]
                where A: Record,
                      B: Comparable,
                      L: Label"#,
        ],
        src: r#"
            x = aggregateWindow(
                fn: max,
                column: "_value",
            )
        "#,
        exp: map![
            "x" => "stream[E]",
        ],
    }
}

#[test]
fn variables_used_in_label_position_must_have_label_kind() {
    test_error_msg! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        src: r#"
            builtin abc: (record: { A with T: time }, ?timeColumn: T = "_time") => int
        "#,
        expect: expect![[r#"
            error: variable T lacks the Label constraint
              ┌─ main:2:13
              │
            2 │             builtin abc: (record: { A with T: time }, ?timeColumn: T = "_time") => int
              │             ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

        "#]],
    }
}

#[test]
fn label_variables_do_not_get_inferred_to_string() {
    test_infer! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        env: map![
            "keep" => r#"(<-table: A, column: string) => { _value: string } where A: Record"#,
            "columns" => r#"(?column: C = "abc") => { C: string } where C: Label"#,
        ],
        src: r#"
            f = (column) =>
                columns(column: column)
                    // Inferring `column` as a `string` here should not force
                    // the type signature of `f` to only accept `string`
                    |> keep(column: column)
        "#,
        exp: map![
            "f" => "(column: A) => { _value: string } where A: Label",
        ],
    }
}
