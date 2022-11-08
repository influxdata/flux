use super::*;

#[test]
fn arrow_function_called() {
    let mut p = Parser::new(
        r#"plusOne = (r) => r + 1
   plusOne(r:5)"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 2, column: 16",
                    source: "plusOne = (r) => r + 1\n   plusOne(r:5)",
                },
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [],
            body: [
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 23",
                                source: "plusOne = (r) => r + 1",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 8",
                                    source: "plusOne",
                                },
                            },
                            name: "plusOne",
                        },
                        init: Function(
                            FunctionExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 11",
                                        end: "line: 1, column: 23",
                                        source: "(r) => r + 1",
                                    },
                                },
                                lparen: [],
                                params: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 12",
                                                end: "line: 1, column: 13",
                                                source: "r",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 12",
                                                        end: "line: 1, column: 13",
                                                        source: "r",
                                                    },
                                                },
                                                name: "r",
                                            },
                                        ),
                                        separator: [],
                                        value: None,
                                        comma: [],
                                    },
                                ],
                                rparen: [],
                                arrow: [],
                                body: Expr(
                                    Binary(
                                        BinaryExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 18",
                                                    end: "line: 1, column: 23",
                                                    source: "r + 1",
                                                },
                                            },
                                            operator: AdditionOperator,
                                            left: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 18",
                                                            end: "line: 1, column: 19",
                                                            source: "r",
                                                        },
                                                    },
                                                    name: "r",
                                                },
                                            ),
                                            right: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 22",
                                                            end: "line: 1, column: 23",
                                                            source: "1",
                                                        },
                                                    },
                                                    value: 1,
                                                },
                                            ),
                                        },
                                    ),
                                ),
                            },
                        ),
                    },
                ),
                Expr(
                    ExprStmt {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 2, column: 4",
                                end: "line: 2, column: 16",
                                source: "plusOne(r:5)",
                            },
                        },
                        expression: Call(
                            CallExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 4",
                                        end: "line: 2, column: 16",
                                        source: "plusOne(r:5)",
                                    },
                                },
                                callee: Identifier(
                                    Identifier {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 4",
                                                end: "line: 2, column: 11",
                                                source: "plusOne",
                                            },
                                        },
                                        name: "plusOne",
                                    },
                                ),
                                lparen: [],
                                arguments: [
                                    Object(
                                        ObjectExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 2, column: 12",
                                                    end: "line: 2, column: 15",
                                                    source: "r:5",
                                                },
                                            },
                                            lbrace: [],
                                            with: None,
                                            properties: [
                                                Property {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 2, column: 12",
                                                            end: "line: 2, column: 15",
                                                            source: "r:5",
                                                        },
                                                    },
                                                    key: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 2, column: 12",
                                                                    end: "line: 2, column: 13",
                                                                    source: "r",
                                                                },
                                                            },
                                                            name: "r",
                                                        },
                                                    ),
                                                    separator: [],
                                                    value: Some(
                                                        Integer(
                                                            IntegerLit {
                                                                base: BaseNode {
                                                                    location: SourceLocation {
                                                                        start: "line: 2, column: 14",
                                                                        end: "line: 2, column: 15",
                                                                        source: "5",
                                                                    },
                                                                },
                                                                value: 5,
                                                            },
                                                        ),
                                                    ),
                                                    comma: [],
                                                },
                                            ],
                                            rbrace: [],
                                        },
                                    ),
                                ],
                                rparen: [],
                            },
                        ),
                    },
                ),
            ],
            eof: [],
        }
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn arrow_function_return_map() {
    let mut p = Parser::new(r#"toMap = (r) =>({r:r})"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 22",
                    source: "toMap = (r) =>({r:r})",
                },
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [],
            body: [
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 22",
                                source: "toMap = (r) =>({r:r})",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 6",
                                    source: "toMap",
                                },
                            },
                            name: "toMap",
                        },
                        init: Function(
                            FunctionExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 9",
                                        end: "line: 1, column: 22",
                                        source: "(r) =>({r:r})",
                                    },
                                },
                                lparen: [],
                                params: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 10",
                                                end: "line: 1, column: 11",
                                                source: "r",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 10",
                                                        end: "line: 1, column: 11",
                                                        source: "r",
                                                    },
                                                },
                                                name: "r",
                                            },
                                        ),
                                        separator: [],
                                        value: None,
                                        comma: [],
                                    },
                                ],
                                rparen: [],
                                arrow: [],
                                body: Expr(
                                    Paren(
                                        ParenExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 15",
                                                    end: "line: 1, column: 22",
                                                    source: "({r:r})",
                                                },
                                            },
                                            lparen: [],
                                            expression: Object(
                                                ObjectExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 16",
                                                            end: "line: 1, column: 21",
                                                            source: "{r:r}",
                                                        },
                                                    },
                                                    lbrace: [],
                                                    with: None,
                                                    properties: [
                                                        Property {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 17",
                                                                    end: "line: 1, column: 20",
                                                                    source: "r:r",
                                                                },
                                                            },
                                                            key: Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 1, column: 17",
                                                                            end: "line: 1, column: 18",
                                                                            source: "r",
                                                                        },
                                                                    },
                                                                    name: "r",
                                                                },
                                                            ),
                                                            separator: [],
                                                            value: Some(
                                                                Identifier(
                                                                    Identifier {
                                                                        base: BaseNode {
                                                                            location: SourceLocation {
                                                                                start: "line: 1, column: 19",
                                                                                end: "line: 1, column: 20",
                                                                                source: "r",
                                                                            },
                                                                        },
                                                                        name: "r",
                                                                    },
                                                                ),
                                                            ),
                                                            comma: [],
                                                        },
                                                    ],
                                                    rbrace: [],
                                                },
                                            ),
                                            rparen: [],
                                        },
                                    ),
                                ),
                            },
                        ),
                    },
                ),
            ],
            eof: [],
        }
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn arrow_function() {
    let mut p = Parser::new(r#"(x,y) => x == y"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 16",
                    source: "(x,y) => x == y",
                },
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [],
            body: [
                Expr(
                    ExprStmt {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 16",
                                source: "(x,y) => x == y",
                            },
                        },
                        expression: Function(
                            FunctionExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 1",
                                        end: "line: 1, column: 16",
                                        source: "(x,y) => x == y",
                                    },
                                },
                                lparen: [],
                                params: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 2",
                                                end: "line: 1, column: 3",
                                                source: "x",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 2",
                                                        end: "line: 1, column: 3",
                                                        source: "x",
                                                    },
                                                },
                                                name: "x",
                                            },
                                        ),
                                        separator: [],
                                        value: None,
                                        comma: [],
                                    },
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 4",
                                                end: "line: 1, column: 5",
                                                source: "y",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 4",
                                                        end: "line: 1, column: 5",
                                                        source: "y",
                                                    },
                                                },
                                                name: "y",
                                            },
                                        ),
                                        separator: [],
                                        value: None,
                                        comma: [],
                                    },
                                ],
                                rparen: [],
                                arrow: [],
                                body: Expr(
                                    Binary(
                                        BinaryExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 10",
                                                    end: "line: 1, column: 16",
                                                    source: "x == y",
                                                },
                                            },
                                            operator: EqualOperator,
                                            left: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 10",
                                                            end: "line: 1, column: 11",
                                                            source: "x",
                                                        },
                                                    },
                                                    name: "x",
                                                },
                                            ),
                                            right: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 15",
                                                            end: "line: 1, column: 16",
                                                            source: "y",
                                                        },
                                                    },
                                                    name: "y",
                                                },
                                            ),
                                        },
                                    ),
                                ),
                            },
                        ),
                    },
                ),
            ],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn arrow_function_with_default_arg() {
    let mut p = Parser::new(r#"addN = (r, n=5) => r + n"#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 25",
                    source: "addN = (r, n=5) => r + n",
                },
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [],
            body: [
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 25",
                                source: "addN = (r, n=5) => r + n",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 5",
                                    source: "addN",
                                },
                            },
                            name: "addN",
                        },
                        init: Function(
                            FunctionExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 8",
                                        end: "line: 1, column: 25",
                                        source: "(r, n=5) => r + n",
                                    },
                                },
                                lparen: [],
                                params: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 9",
                                                end: "line: 1, column: 10",
                                                source: "r",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 9",
                                                        end: "line: 1, column: 10",
                                                        source: "r",
                                                    },
                                                },
                                                name: "r",
                                            },
                                        ),
                                        separator: [],
                                        value: None,
                                        comma: [],
                                    },
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 12",
                                                end: "line: 1, column: 15",
                                                source: "n=5",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 12",
                                                        end: "line: 1, column: 13",
                                                        source: "n",
                                                    },
                                                },
                                                name: "n",
                                            },
                                        ),
                                        separator: [],
                                        value: Some(
                                            Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 14",
                                                            end: "line: 1, column: 15",
                                                            source: "5",
                                                        },
                                                    },
                                                    value: 5,
                                                },
                                            ),
                                        ),
                                        comma: [],
                                    },
                                ],
                                rparen: [],
                                arrow: [],
                                body: Expr(
                                    Binary(
                                        BinaryExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 20",
                                                    end: "line: 1, column: 25",
                                                    source: "r + n",
                                                },
                                            },
                                            operator: AdditionOperator,
                                            left: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 20",
                                                            end: "line: 1, column: 21",
                                                            source: "r",
                                                        },
                                                    },
                                                    name: "r",
                                                },
                                            ),
                                            right: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 24",
                                                            end: "line: 1, column: 25",
                                                            source: "n",
                                                        },
                                                    },
                                                    name: "n",
                                                },
                                            ),
                                        },
                                    ),
                                ),
                            },
                        ),
                    },
                ),
            ],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn arrow_function_called_in_binary_expression() {
    let mut p = Parser::new(
        r#"
            plusOne = (r) => r + 1
            plusOne(r:5) == 6 or die()"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 3, column: 39",
                    source: "\n            plusOne = (r) => r + 1\n            plusOne(r:5) == 6 or die()",
                },
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [],
            body: [
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 2, column: 13",
                                end: "line: 2, column: 35",
                                source: "plusOne = (r) => r + 1",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 2, column: 13",
                                    end: "line: 2, column: 20",
                                    source: "plusOne",
                                },
                            },
                            name: "plusOne",
                        },
                        init: Function(
                            FunctionExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 2, column: 23",
                                        end: "line: 2, column: 35",
                                        source: "(r) => r + 1",
                                    },
                                },
                                lparen: [],
                                params: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 2, column: 24",
                                                end: "line: 2, column: 25",
                                                source: "r",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 2, column: 24",
                                                        end: "line: 2, column: 25",
                                                        source: "r",
                                                    },
                                                },
                                                name: "r",
                                            },
                                        ),
                                        separator: [],
                                        value: None,
                                        comma: [],
                                    },
                                ],
                                rparen: [],
                                arrow: [],
                                body: Expr(
                                    Binary(
                                        BinaryExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 2, column: 30",
                                                    end: "line: 2, column: 35",
                                                    source: "r + 1",
                                                },
                                            },
                                            operator: AdditionOperator,
                                            left: Identifier(
                                                Identifier {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 2, column: 30",
                                                            end: "line: 2, column: 31",
                                                            source: "r",
                                                        },
                                                    },
                                                    name: "r",
                                                },
                                            ),
                                            right: Integer(
                                                IntegerLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 2, column: 34",
                                                            end: "line: 2, column: 35",
                                                            source: "1",
                                                        },
                                                    },
                                                    value: 1,
                                                },
                                            ),
                                        },
                                    ),
                                ),
                            },
                        ),
                    },
                ),
                Expr(
                    ExprStmt {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 3, column: 13",
                                end: "line: 3, column: 39",
                                source: "plusOne(r:5) == 6 or die()",
                            },
                        },
                        expression: Logical(
                            LogicalExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 3, column: 13",
                                        end: "line: 3, column: 39",
                                        source: "plusOne(r:5) == 6 or die()",
                                    },
                                },
                                operator: OrOperator,
                                left: Binary(
                                    BinaryExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 3, column: 13",
                                                end: "line: 3, column: 30",
                                                source: "plusOne(r:5) == 6",
                                            },
                                        },
                                        operator: EqualOperator,
                                        left: Call(
                                            CallExpr {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 3, column: 13",
                                                        end: "line: 3, column: 25",
                                                        source: "plusOne(r:5)",
                                                    },
                                                },
                                                callee: Identifier(
                                                    Identifier {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 3, column: 13",
                                                                end: "line: 3, column: 20",
                                                                source: "plusOne",
                                                            },
                                                        },
                                                        name: "plusOne",
                                                    },
                                                ),
                                                lparen: [],
                                                arguments: [
                                                    Object(
                                                        ObjectExpr {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 3, column: 21",
                                                                    end: "line: 3, column: 24",
                                                                    source: "r:5",
                                                                },
                                                            },
                                                            lbrace: [],
                                                            with: None,
                                                            properties: [
                                                                Property {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 3, column: 21",
                                                                            end: "line: 3, column: 24",
                                                                            source: "r:5",
                                                                        },
                                                                    },
                                                                    key: Identifier(
                                                                        Identifier {
                                                                            base: BaseNode {
                                                                                location: SourceLocation {
                                                                                    start: "line: 3, column: 21",
                                                                                    end: "line: 3, column: 22",
                                                                                    source: "r",
                                                                                },
                                                                            },
                                                                            name: "r",
                                                                        },
                                                                    ),
                                                                    separator: [],
                                                                    value: Some(
                                                                        Integer(
                                                                            IntegerLit {
                                                                                base: BaseNode {
                                                                                    location: SourceLocation {
                                                                                        start: "line: 3, column: 23",
                                                                                        end: "line: 3, column: 24",
                                                                                        source: "5",
                                                                                    },
                                                                                },
                                                                                value: 5,
                                                                            },
                                                                        ),
                                                                    ),
                                                                    comma: [],
                                                                },
                                                            ],
                                                            rbrace: [],
                                                        },
                                                    ),
                                                ],
                                                rparen: [],
                                            },
                                        ),
                                        right: Integer(
                                            IntegerLit {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 3, column: 29",
                                                        end: "line: 3, column: 30",
                                                        source: "6",
                                                    },
                                                },
                                                value: 6,
                                            },
                                        ),
                                    },
                                ),
                                right: Call(
                                    CallExpr {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 3, column: 34",
                                                end: "line: 3, column: 39",
                                                source: "die()",
                                            },
                                        },
                                        callee: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 3, column: 34",
                                                        end: "line: 3, column: 37",
                                                        source: "die",
                                                    },
                                                },
                                                name: "die",
                                            },
                                        ),
                                        lparen: [],
                                        arguments: [],
                                        rparen: [],
                                    },
                                ),
                            },
                        ),
                    },
                ),
            ],
            eof: [],
        }
    "#]].assert_debug_eq(&parsed);
}

#[test]
fn arrow_function_as_single_expression() {
    let mut p = Parser::new(r#"f = (r) => r["_measurement"] == "cpu""#);
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 1, column: 38",
                    source: "f = (r) => r[\"_measurement\"] == \"cpu\"",
                },
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [],
            body: [
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 1, column: 38",
                                source: "f = (r) => r[\"_measurement\"] == \"cpu\"",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "f",
                                },
                            },
                            name: "f",
                        },
                        init: Function(
                            FunctionExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 1, column: 38",
                                        source: "(r) => r[\"_measurement\"] == \"cpu\"",
                                    },
                                },
                                lparen: [],
                                params: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 7",
                                                source: "r",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 7",
                                                        source: "r",
                                                    },
                                                },
                                                name: "r",
                                            },
                                        ),
                                        separator: [],
                                        value: None,
                                        comma: [],
                                    },
                                ],
                                rparen: [],
                                arrow: [],
                                body: Expr(
                                    Binary(
                                        BinaryExpr {
                                            base: BaseNode {
                                                location: SourceLocation {
                                                    start: "line: 1, column: 12",
                                                    end: "line: 1, column: 38",
                                                    source: "r[\"_measurement\"] == \"cpu\"",
                                                },
                                            },
                                            operator: EqualOperator,
                                            left: Member(
                                                MemberExpr {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 12",
                                                            end: "line: 1, column: 29",
                                                            source: "r[\"_measurement\"]",
                                                        },
                                                    },
                                                    object: Identifier(
                                                        Identifier {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 12",
                                                                    end: "line: 1, column: 13",
                                                                    source: "r",
                                                                },
                                                            },
                                                            name: "r",
                                                        },
                                                    ),
                                                    lbrack: [],
                                                    property: StringLit(
                                                        StringLit {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 1, column: 14",
                                                                    end: "line: 1, column: 28",
                                                                    source: "\"_measurement\"",
                                                                },
                                                            },
                                                            value: "_measurement",
                                                        },
                                                    ),
                                                    rbrack: [],
                                                },
                                            ),
                                            right: StringLit(
                                                StringLit {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 1, column: 33",
                                                            end: "line: 1, column: 38",
                                                            source: "\"cpu\"",
                                                        },
                                                    },
                                                    value: "cpu",
                                                },
                                            ),
                                        },
                                    ),
                                ),
                            },
                        ),
                    },
                ),
            ],
            eof: [],
        }
    "#]]
    .assert_debug_eq(&parsed);
}

#[test]
fn arrow_function_as_block() {
    let mut p = Parser::new(
        r#"f = (r) => {
                m = r["_measurement"]
                return m == "cpu"
            }"#,
    );
    let parsed = p.parse_file("".to_string());
    expect![[r#"
        File {
            base: BaseNode {
                location: SourceLocation {
                    start: "line: 1, column: 1",
                    end: "line: 4, column: 14",
                    source: "f = (r) => {\n                m = r[\"_measurement\"]\n                return m == \"cpu\"\n            }",
                },
            },
            name: "",
            metadata: "parser-type=rust",
            package: None,
            imports: [],
            body: [
                Variable(
                    VariableAssgn {
                        base: BaseNode {
                            location: SourceLocation {
                                start: "line: 1, column: 1",
                                end: "line: 4, column: 14",
                                source: "f = (r) => {\n                m = r[\"_measurement\"]\n                return m == \"cpu\"\n            }",
                            },
                        },
                        id: Identifier {
                            base: BaseNode {
                                location: SourceLocation {
                                    start: "line: 1, column: 1",
                                    end: "line: 1, column: 2",
                                    source: "f",
                                },
                            },
                            name: "f",
                        },
                        init: Function(
                            FunctionExpr {
                                base: BaseNode {
                                    location: SourceLocation {
                                        start: "line: 1, column: 5",
                                        end: "line: 4, column: 14",
                                        source: "(r) => {\n                m = r[\"_measurement\"]\n                return m == \"cpu\"\n            }",
                                    },
                                },
                                lparen: [],
                                params: [
                                    Property {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 6",
                                                end: "line: 1, column: 7",
                                                source: "r",
                                            },
                                        },
                                        key: Identifier(
                                            Identifier {
                                                base: BaseNode {
                                                    location: SourceLocation {
                                                        start: "line: 1, column: 6",
                                                        end: "line: 1, column: 7",
                                                        source: "r",
                                                    },
                                                },
                                                name: "r",
                                            },
                                        ),
                                        separator: [],
                                        value: None,
                                        comma: [],
                                    },
                                ],
                                rparen: [],
                                arrow: [],
                                body: Block(
                                    Block {
                                        base: BaseNode {
                                            location: SourceLocation {
                                                start: "line: 1, column: 12",
                                                end: "line: 4, column: 14",
                                                source: "{\n                m = r[\"_measurement\"]\n                return m == \"cpu\"\n            }",
                                            },
                                        },
                                        lbrace: [],
                                        body: [
                                            Variable(
                                                VariableAssgn {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 2, column: 17",
                                                            end: "line: 2, column: 38",
                                                            source: "m = r[\"_measurement\"]",
                                                        },
                                                    },
                                                    id: Identifier {
                                                        base: BaseNode {
                                                            location: SourceLocation {
                                                                start: "line: 2, column: 17",
                                                                end: "line: 2, column: 18",
                                                                source: "m",
                                                            },
                                                        },
                                                        name: "m",
                                                    },
                                                    init: Member(
                                                        MemberExpr {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 2, column: 21",
                                                                    end: "line: 2, column: 38",
                                                                    source: "r[\"_measurement\"]",
                                                                },
                                                            },
                                                            object: Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 2, column: 21",
                                                                            end: "line: 2, column: 22",
                                                                            source: "r",
                                                                        },
                                                                    },
                                                                    name: "r",
                                                                },
                                                            ),
                                                            lbrack: [],
                                                            property: StringLit(
                                                                StringLit {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 2, column: 23",
                                                                            end: "line: 2, column: 37",
                                                                            source: "\"_measurement\"",
                                                                        },
                                                                    },
                                                                    value: "_measurement",
                                                                },
                                                            ),
                                                            rbrack: [],
                                                        },
                                                    ),
                                                },
                                            ),
                                            Return(
                                                ReturnStmt {
                                                    base: BaseNode {
                                                        location: SourceLocation {
                                                            start: "line: 3, column: 17",
                                                            end: "line: 3, column: 34",
                                                            source: "return m == \"cpu\"",
                                                        },
                                                    },
                                                    argument: Binary(
                                                        BinaryExpr {
                                                            base: BaseNode {
                                                                location: SourceLocation {
                                                                    start: "line: 3, column: 24",
                                                                    end: "line: 3, column: 34",
                                                                    source: "m == \"cpu\"",
                                                                },
                                                            },
                                                            operator: EqualOperator,
                                                            left: Identifier(
                                                                Identifier {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 3, column: 24",
                                                                            end: "line: 3, column: 25",
                                                                            source: "m",
                                                                        },
                                                                    },
                                                                    name: "m",
                                                                },
                                                            ),
                                                            right: StringLit(
                                                                StringLit {
                                                                    base: BaseNode {
                                                                        location: SourceLocation {
                                                                            start: "line: 3, column: 29",
                                                                            end: "line: 3, column: 34",
                                                                            source: "\"cpu\"",
                                                                        },
                                                                    },
                                                                    value: "cpu",
                                                                },
                                                            ),
                                                        },
                                                    ),
                                                },
                                            ),
                                        ],
                                        rbrace: [],
                                    },
                                ),
                            },
                        ),
                    },
                ),
            ],
            eof: [],
        }
    "#]].assert_debug_eq(&parsed);
}
