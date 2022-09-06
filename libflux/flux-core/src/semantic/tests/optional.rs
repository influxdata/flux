use super::*;

#[test]
fn optional_type() {
    test_infer! {
        env: map![
            "f" => "(x: A?) => A",
        ],
        src: r#"
            x = f(x: 1)
        "#,
        exp: map![
            "x" => "int",
        ],
    }
}

#[test]
fn optional_can_be_passed_to_optional_argument() {
    test_infer! {
        env: map![
            "x" => "int?",
            "f" => "(?x: A) => A",
        ],
        src: r#"
            y = f(x)
            z = f(x: "")
        "#,
        exp: map![
            "y" => "int",
            "z" => "string",
        ],
    }
}

#[test]
fn pass_optional_to_mandatory_is_error() {
    test_error_msg! {
        env: map![
            "x" => "int?",
            "f" => "(x: int) => int",
        ],
        src: r#"
            y = f(x)
        "#,
        expect: expect![[r#"
            error: expected int but found int? (argument x)
              ┌─ main:2:19
              │
            2 │             y = f(x)
              │                   ^

        "#]],
    }
}

#[test]
fn optional_do_not_unify_with_mandatory() {
    test_error_msg! {
        env: map![
            "x" => "int?",
            "true" => "bool",
        ],
        src: r#"
            y = if true then x else 1
            z = if true then 1 else x
        "#,
        expect: expect![[r#"
            error: expected int? but found int
              ┌─ main:2:37
              │
            2 │             y = if true then x else 1
              │                                     ^

            error: expected int but found int?
              ┌─ main:3:37
              │
            3 │             z = if true then 1 else x
              │                                     ^

        "#]],
    }
}

#[test]
fn wrong_type_to_optional_parameter() {
    test_error_msg! {
        env: map![
            "f" => "(?x: int) => int",
        ],
        src: r#"
            y = f(x: "")
        "#,
        // CHeck that we do not regress the error message due to optionals
        expect: expect![[r#"
            error: expected int but found string (argument x)
              ┌─ main:2:22
              │
            2 │             y = f(x: "")
              │                      ^^

        "#]],
    }
}
