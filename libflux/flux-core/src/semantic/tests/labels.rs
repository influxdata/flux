use super::*;

#[test]
fn labels_simple() {
    test_infer! {
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
