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
    test_infer! {
        env: map![
            "fill" => "(<-tables: [{ A with B: C }], ?column: B, ?value: D) => [{ A with B: D }]
                where B: Label
                "
        ],
        src: r#"
            column = "" + "a"
            x = [{ a: 1 }] |> fill(column: column, value: "x")
        "#,
        exp: map![
            "x" => "string",
            "x" => "[{ a: string }]",
        ],
    }
}
