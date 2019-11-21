// NOTE: These test cases directly match ast/json_test.go.
// Every test is preceded by the correspondent test case in golang.
use super::*;
use chrono::TimeZone;

/*
{
    name: "string interpolation",
    node: &ast.StringExpression{
        Parts: []ast.StringExprPart{
            &ast.TextPart{
                Value: "a = ",
            },
            &ast.InterpolatedPart{
                Expression: &ast.Identifier{
                    Name: "a",
                },
            },
        },
    },
    want: `{"type":"StringExpression","parts":[{"type":"TextPart","value":"a = "},{"type":"InterpolatedPart","expression":{"type":"Identifier","name":"a"}}]}`,
},
*/
#[test]
fn test_string_interpolation() {
    let n = StringExpr {
        base: BaseNode::default(),
        parts: vec![
            StringExprPart::Text(TextPart {
                base: BaseNode::default(),
                value: "a = ".to_string(),
            }),
            StringExprPart::Interpolated(InterpolatedPart {
                base: BaseNode::default(),
                expression: Expression::Identifier(Identifier {
                    base: BaseNode::default(),
                    name: "a".to_string(),
                }),
            }),
        ],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"StringExpression","parts":[{"type":"TextPart","value":"a = "},{"type":"InterpolatedPart","expression":{"type":"Identifier","name":"a"}}]}"#
    );
    let deserialized: StringExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "paren expression",
    node: &ast.ParenExpression{
        Expression: &ast.StringExpression{
            Parts: []ast.StringExprPart{
                &ast.TextPart{
                    Value: "a = ",
                },
                &ast.InterpolatedPart{
                    Expression: &ast.Identifier{
                        Name: "a",
                    },
                },
            },
        },
    },
    want: `{"type":"ParenExpression","expression":{"type":"StringExpression","parts":[{"type":"TextPart","value":"a = "},{"type":"InterpolatedPart","expression":{"type":"Identifier","name":"a"}}]}}`,
},
*/
#[test]
fn test_paren_expression() {
    let n = ParenExpr {
        base: BaseNode::default(),
        expression: Expression::StringExpr(Box::new(StringExpr {
            base: BaseNode::default(),
            parts: vec![
                StringExprPart::Text(TextPart {
                    base: BaseNode::default(),
                    value: "a = ".to_string(),
                }),
                StringExprPart::Interpolated(InterpolatedPart {
                    base: BaseNode::default(),
                    expression: Expression::Identifier(Identifier {
                        base: BaseNode::default(),
                        name: "a".to_string(),
                    }),
                }),
            ],
        })),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ParenExpression","expression":{"type":"StringExpression","parts":[{"type":"TextPart","value":"a = "},{"type":"InterpolatedPart","expression":{"type":"Identifier","name":"a"}}]}}"#
    );
    let deserialized: ParenExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "simple package",
    node: &ast.Package{
        Package: "foo",
    },
    want: `{"type":"Package","package":"foo","files":null}`,
},
*/
#[test] // NOTE: adapted for non-nullable files.
fn test_json_simple_package() {
    let n = Package {
        base: BaseNode::default(),
        path: String::new(),
        package: "foo".to_string(),
        files: Vec::new(),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"Package","package":"foo","files":[]}"#
    );
    let deserialized: Package = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "package path",
    node: &ast.Package{
        Path:    "bar/foo",
        Package: "foo",
    },
    want: `{"type":"Package","path":"bar/foo","package":"foo","files":null}`,
},
*/
#[test] // NOTE: adapted for non-nullable files.
fn test_json_package_path() {
    let n = Package {
        base: BaseNode::default(),
        path: "bar/foo".to_string(),
        package: "foo".to_string(),
        files: Vec::new(),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"Package","path":"bar/foo","package":"foo","files":[]}"#
    );
    let deserialized: Package = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "simple file",
    node: &ast.File{
        Body: []ast.Statement{
            &ast.ExpressionStatement{
                Expression: &ast.StringLiteral{Value: "hello"},
            },
        },
    },
    want: `{"type":"File","package":null,"imports":null,"body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}`,
},
*/
#[test] // NOTE: adapted for non-nullable imports.
fn test_json_simple_file() {
    let n = File {
        base: BaseNode::default(),
        package: Option::None,
        imports: Vec::new(),
        name: String::new(),
        body: vec![Statement::Expr(ExprStmt {
            base: BaseNode::default(),
            expression: Expression::StringLit(StringLit {
                base: Default::default(),
                value: "hello".to_string(),
            }),
        })],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"File","package":null,"imports":[],"body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}"#
    );
    let deserialized: File = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "file",
    node: &ast.File{
        Package: &ast.PackageClause{
            Name: &ast.Identifier{Name: "foo"},
        },
        Imports: []*ast.ImportDeclaration{{
            As:   &ast.Identifier{Name: "b"},
            Path: &ast.StringLiteral{Value: "path/bar"},
        }},
        Body: []ast.Statement{
            &ast.ExpressionStatement{
                Expression: &ast.StringLiteral{Value: "hello"},
            },
        },
    },
    want: `{"type":"File","package":{"type":"PackageClause","name":{"type":"Identifier","name":"foo"}},"imports":[{"type":"ImportDeclaration","as":{"type":"Identifier","name":"b"},"path":{"type":"StringLiteral","value":"path/bar"}}],"body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}`,
},
*/
#[test]
fn test_json_file() {
    let n = File {
        base: BaseNode::default(),
        package: Some(PackageClause {
            base: BaseNode::default(),
            name: Identifier {
                base: Default::default(),
                name: "foo".to_string(),
            },
        }),
        imports: vec![ImportDeclaration {
            base: BaseNode::default(),
            alias: Some(Identifier {
                base: Default::default(),
                name: "b".to_string(),
            }),
            path: StringLit {
                base: BaseNode::default(),
                value: "path/bar".to_string(),
            },
        }],
        name: String::new(),
        body: vec![Statement::Expr(ExprStmt {
            base: BaseNode::default(),
            expression: Expression::StringLit(StringLit {
                base: Default::default(),
                value: "hello".to_string(),
            }),
        })],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"File","package":{"type":"PackageClause","name":{"type":"Identifier","name":"foo"}},"imports":[{"type":"ImportDeclaration","as":{"type":"Identifier","name":"b"},"path":{"type":"StringLiteral","value":"path/bar"}}],"body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}"#
    );
    let deserialized: File = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "block",
    node: &ast.Block{
        Body: []ast.Statement{
            &ast.ExpressionStatement{
                Expression: &ast.StringLiteral{Value: "hello"},
            },
        },
    },
    want: `{"type":"Block","body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}`,
},
*/
#[test]
fn test_json_block() {
    let n = Block {
        base: BaseNode::default(),
        body: vec![Statement::Expr(ExprStmt {
            base: BaseNode::default(),
            expression: Expression::StringLit(StringLit {
                base: Default::default(),
                value: "hello".to_string(),
            }),
        })],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"Block","body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}"#
    );
    let deserialized: Block = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "expression statement",
    node: &ast.ExpressionStatement{
        Expression: &ast.StringLiteral{Value: "hello"},
    },
    want: `{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}`,
},
*/
#[test]
fn test_json_expression_statement() {
    let n = ExprStmt {
        base: BaseNode::default(),
        expression: Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "hello".to_string(),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}"#
    );
    let deserialized: ExprStmt = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "return statement",
    node: &ast.ReturnStatement{
        Argument: &ast.StringLiteral{Value: "hello"},
    },
    want: `{"type":"ReturnStatement","argument":{"type":"StringLiteral","value":"hello"}}`,
},
*/
#[test]
fn test_json_return_statement() {
    let n = ReturnStmt {
        base: BaseNode::default(),
        argument: Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "hello".to_string(),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ReturnStatement","argument":{"type":"StringLiteral","value":"hello"}}"#
    );
    let deserialized: ReturnStmt = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "option statement",
    node: &ast.OptionStatement{
        Assignment: &ast.VariableAssignment{
            ID: &ast.Identifier{Name: "task"},
            Init: &ast.ObjectExpression{
                Properties: []*ast.Property{
                    {
                        Key:   &ast.Identifier{Name: "name"},
                        Value: &ast.StringLiteral{Value: "foo"},
                    },
                    {
                        Key: &ast.Identifier{Name: "every"},
                        Value: &ast.DurationLiteral{
                            Values: []ast.Duration{
                                {
                                    Magnitude: 1,
                                    Unit:      "h",
                                },
                            },
                        },
                    },
                },
            },
        },
    },
    want: `{"type":"OptionStatement","assignment":{"type":"VariableAssignment","id":{"type":"Identifier","name":"task"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"name"},"value":{"type":"StringLiteral","value":"foo"}},{"type":"Property","key":{"type":"Identifier","name":"every"},"value":{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"}]}}]}}}`,
},
*/
#[test]
fn test_json_option_statement() {
    let n = OptionStmt {
        base: BaseNode::default(),
        assignment: Assignment::Variable(Box::new(VariableAssgn {
            base: BaseNode::default(),
            id: Identifier {
                base: BaseNode::default(),
                name: "task".to_string(),
            },
            init: Expression::Object(Box::new(ObjectExpr {
                base: BaseNode::default(),
                with: None,
                properties: vec![
                    Property {
                        base: BaseNode::default(),
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode::default(),
                            name: "name".to_string(),
                        }),
                        value: Some(Expression::StringLit(StringLit {
                            base: Default::default(),
                            value: "foo".to_string(),
                        })),
                    },
                    Property {
                        base: BaseNode::default(),
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode::default(),
                            name: "every".to_string(),
                        }),
                        value: Some(Expression::Duration(DurationLit {
                            base: Default::default(),
                            values: vec![Duration {
                                magnitude: 1,
                                unit: "h".to_string(),
                            }],
                        })),
                    },
                ],
            })),
        })),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"OptionStatement","assignment":{"type":"VariableAssignment","id":{"type":"Identifier","name":"task"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"name"},"value":{"type":"StringLiteral","value":"foo"}},{"type":"Property","key":{"type":"Identifier","name":"every"},"value":{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"}]}}]}}}"#
    );
    let deserialized: OptionStmt = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "builtin statement",
    node: &ast.BuiltinStatement{
        ID: &ast.Identifier{Name: "task"},
    },
    want: `{"type":"BuiltinStatement","id":{"type":"Identifier","name":"task"}}`,
},
*/
#[test]
fn test_json_builtin_statement() {
    let n = BuiltinStmt {
        base: BaseNode::default(),
        id: Identifier {
            base: BaseNode::default(),
            name: "task".to_string(),
        },
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"BuiltinStatement","id":{"type":"Identifier","name":"task"}}"#
    );
    let deserialized: BuiltinStmt = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "test statement",
    node: &ast.TestStatement{
        Assignment: &ast.VariableAssignment{
            ID: &ast.Identifier{Name: "mean"},
            Init: &ast.ObjectExpression{
                Properties: []*ast.Property{
                    {
                        Key: &ast.Identifier{
                            Name: "want",
                        },
                        Value: &ast.IntegerLiteral{
                            Value: 0,
                        },
                    },
                    {
                        Key: &ast.Identifier{
                            Name: "got",
                        },
                        Value: &ast.IntegerLiteral{
                            Value: 0,
                        },
                    },
                },
            },
        },
    },
    want: `{"type":"TestStatement","assignment":{"type":"VariableAssignment","id":{"type":"Identifier","name":"mean"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"want"},"value":{"type":"IntegerLiteral","value":"0"}},{"type":"Property","key":{"type":"Identifier","name":"got"},"value":{"type":"IntegerLiteral","value":"0"}}]}}}`,
},
*/
#[test]
fn test_json_test_statement() {
    let n = TestStmt {
        base: BaseNode::default(),
        assignment: VariableAssgn {
            base: BaseNode::default(),
            id: Identifier {
                base: BaseNode::default(),
                name: "mean".to_string(),
            },
            init: Expression::Object(Box::new(ObjectExpr {
                base: BaseNode::default(),
                with: None,
                properties: vec![
                    Property {
                        base: BaseNode::default(),
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode::default(),
                            name: "want".to_string(),
                        }),
                        value: Some(Expression::Integer(IntegerLit {
                            base: Default::default(),
                            value: 0,
                        })),
                    },
                    Property {
                        base: BaseNode::default(),
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode::default(),
                            name: "got".to_string(),
                        }),
                        value: Some(Expression::Integer(IntegerLit {
                            base: Default::default(),
                            value: 0,
                        })),
                    },
                ],
            })),
        },
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"TestStatement","assignment":{"type":"VariableAssignment","id":{"type":"Identifier","name":"mean"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"want"},"value":{"type":"IntegerLiteral","value":"0"}},{"type":"Property","key":{"type":"Identifier","name":"got"},"value":{"type":"IntegerLiteral","value":"0"}}]}}}"#
    );
    let deserialized: TestStmt = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "qualified option statement",
    node: &ast.OptionStatement{
        Assignment: &ast.MemberAssignment{
            Member: &ast.MemberExpression{
                Object: &ast.Identifier{
                    Name: "alert",
                },
                Property: &ast.Identifier{
                    Name: "state",
                },
            },
            Init: &ast.StringLiteral{
                Value: "Warning",
            },
        },
    },
    want: `{"type":"OptionStatement","assignment":{"type":"MemberAssignment","member":{"type":"MemberExpression","object":{"type":"Identifier","name":"alert"},"property":{"type":"Identifier","name":"state"}},"init":{"type":"StringLiteral","value":"Warning"}}}`,
},
*/
#[test]
fn test_json_qualified_option_statement() {
    let n = OptionStmt {
        base: BaseNode::default(),
        assignment: Assignment::Member(Box::new(MemberAssgn {
            base: BaseNode::default(),
            member: MemberExpr {
                base: BaseNode::default(),
                object: Expression::Identifier(Identifier {
                    base: BaseNode::default(),
                    name: "alert".to_string(),
                }),
                property: PropertyKey::Identifier(Identifier {
                    base: BaseNode::default(),
                    name: "state".to_string(),
                }),
            },
            init: Expression::StringLit(StringLit {
                base: Default::default(),
                value: "Warning".to_string(),
            }),
        })),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"OptionStatement","assignment":{"type":"MemberAssignment","member":{"type":"MemberExpression","object":{"type":"Identifier","name":"alert"},"property":{"type":"Identifier","name":"state"}},"init":{"type":"StringLiteral","value":"Warning"}}}"#
    );
    let deserialized: OptionStmt = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "variable assignment",
    node: &ast.VariableAssignment{
        ID:   &ast.Identifier{Name: "a"},
        Init: &ast.StringLiteral{Value: "hello"},
    },
    want: `{"type":"VariableAssignment","id":{"type":"Identifier","name":"a"},"init":{"type":"StringLiteral","value":"hello"}}`,
},
*/
#[test]
fn test_json_variable_assignment() {
    let n = VariableAssgn {
        base: BaseNode::default(),
        id: Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        },
        init: Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "hello".to_string(),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"VariableAssignment","id":{"type":"Identifier","name":"a"},"init":{"type":"StringLiteral","value":"hello"}}"#
    );
    let deserialized: VariableAssgn = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "call expression",
    node: &ast.CallExpression{
        Callee:    &ast.Identifier{Name: "a"},
        Arguments: []ast.Expression{&ast.StringLiteral{Value: "hello"}},
    },
    want: `{"type":"CallExpression","callee":{"type":"Identifier","name":"a"},"arguments":[{"type":"StringLiteral","value":"hello"}]}`,
},
*/
#[test]
fn test_json_call_expression() {
    let n = CallExpr {
        base: BaseNode::default(),
        callee: Expression::Identifier(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        arguments: vec![Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "hello".to_string(),
        })],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"CallExpression","callee":{"type":"Identifier","name":"a"},"arguments":[{"type":"StringLiteral","value":"hello"}]}"#
    );
    let deserialized: CallExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "pipe expression",
    node: &ast.PipeExpression{
        Argument: &ast.Identifier{Name: "a"},
        Call: &ast.CallExpression{
            Callee:    &ast.Identifier{Name: "a"},
            Arguments: []ast.Expression{&ast.StringLiteral{Value: "hello"}},
        },
    },
    want: `{"type":"PipeExpression","argument":{"type":"Identifier","name":"a"},"call":{"type":"CallExpression","callee":{"type":"Identifier","name":"a"},"arguments":[{"type":"StringLiteral","value":"hello"}]}}`,
},
*/
#[test]
fn test_json_pipe_expression() {
    let n = PipeExpr {
        base: BaseNode::default(),
        argument: Expression::Identifier(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        call: CallExpr {
            base: BaseNode::default(),
            callee: Expression::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            arguments: vec![Expression::StringLit(StringLit {
                base: BaseNode::default(),
                value: "hello".to_string(),
            })],
        },
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"PipeExpression","argument":{"type":"Identifier","name":"a"},"call":{"type":"CallExpression","callee":{"type":"Identifier","name":"a"},"arguments":[{"type":"StringLiteral","value":"hello"}]}}"#
    );
    let deserialized: PipeExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "member expression with identifier",
    node: &ast.MemberExpression{
        Object:   &ast.Identifier{Name: "a"},
        Property: &ast.Identifier{Name: "b"},
    },
    want: `{"type":"MemberExpression","object":{"type":"Identifier","name":"a"},"property":{"type":"Identifier","name":"b"}}`,
},
*/
#[test]
fn test_json_member_expression_with_identifier() {
    let n = MemberExpr {
        base: BaseNode::default(),
        object: Expression::Identifier(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        property: PropertyKey::Identifier(Identifier {
            base: BaseNode::default(),
            name: "b".to_string(),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"MemberExpression","object":{"type":"Identifier","name":"a"},"property":{"type":"Identifier","name":"b"}}"#
    );
    let deserialized: MemberExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "member expression with string literal",
    node: &ast.MemberExpression{
        Object:   &ast.Identifier{Name: "a"},
        Property: &ast.StringLiteral{Value: "b"},
    },
    want: `{"type":"MemberExpression","object":{"type":"Identifier","name":"a"},"property":{"type":"StringLiteral","value":"b"}}`,
},
*/
#[test]
fn test_json_member_expression_with_string_literal() {
    let n = MemberExpr {
        base: BaseNode::default(),
        object: Expression::Identifier(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        property: PropertyKey::StringLit(StringLit {
            base: BaseNode::default(),
            value: "b".to_string(),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"MemberExpression","object":{"type":"Identifier","name":"a"},"property":{"type":"StringLiteral","value":"b"}}"#
    );
    let deserialized: MemberExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "index expression",
    node: &ast.IndexExpression{
        Array: &ast.Identifier{Name: "a"},
        Index: &ast.IntegerLiteral{Value: 3},
    },
    want: `{"type":"IndexExpression","array":{"type":"Identifier","name":"a"},"index":{"type":"IntegerLiteral","value":"3"}}`,
},
*/
#[test]
fn test_json_index_expression() {
    let n = IndexExpr {
        base: BaseNode::default(),
        array: Expression::Identifier(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        index: Expression::Integer(IntegerLit {
            base: BaseNode::default(),
            value: 3,
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"IndexExpression","array":{"type":"Identifier","name":"a"},"index":{"type":"IntegerLiteral","value":"3"}}"#
    );
    let deserialized: IndexExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "arrow function expression",
    node: &ast.FunctionExpression{
        Params: []*ast.Property{{Key: &ast.Identifier{Name: "a"}}},
        Body:   &ast.StringLiteral{Value: "hello"},
    },
    want: `{"type":"FunctionExpression","params":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}],"body":{"type":"StringLiteral","value":"hello"}}`,
},
*/
#[test]
fn test_json_arrow_function_expression() {
    let n = FunctionExpr {
        base: BaseNode::default(),
        params: vec![Property {
            base: BaseNode::default(),
            key: PropertyKey::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            value: None,
        }],
        body: FunctionBody::Expr(Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "hello".to_string(),
        })),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"FunctionExpression","params":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}],"body":{"type":"StringLiteral","value":"hello"}}"#
    );
    let deserialized: FunctionExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "binary expression",
    node: &ast.BinaryExpression{
        Operator: ast.AdditionOperator,
        Left:     &ast.StringLiteral{Value: "hello"},
        Right:    &ast.StringLiteral{Value: "world"},
    },
    want: `{"type":"BinaryExpression","operator":"+","left":{"type":"StringLiteral","value":"hello"},"right":{"type":"StringLiteral","value":"world"}}`,
},
*/
#[test]
fn test_json_binary_expression() {
    let n = BinaryExpr {
        base: BaseNode::default(),
        operator: Operator::AdditionOperator,
        left: Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "hello".to_string(),
        }),
        right: Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "world".to_string(),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"BinaryExpression","operator":"+","left":{"type":"StringLiteral","value":"hello"},"right":{"type":"StringLiteral","value":"world"}}"#
    );
    let deserialized: BinaryExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "unary expression",
    node: &ast.UnaryExpression{
        Operator: ast.NotOperator,
        Argument: &ast.BooleanLiteral{Value: true},
    },
    want: `{"type":"UnaryExpression","operator":"not","argument":{"type":"BooleanLiteral","value":true}}`,
},
*/
#[test]
fn test_json_unary_expression() {
    let n = UnaryExpr {
        base: BaseNode::default(),
        operator: Operator::NotOperator,
        argument: Expression::Boolean(BooleanLit {
            base: BaseNode::default(),
            value: true,
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"UnaryExpression","operator":"not","argument":{"type":"BooleanLiteral","value":true}}"#
    );
    let deserialized: UnaryExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "logical expression",
    node: &ast.LogicalExpression{
        Operator: ast.OrOperator,
        Left:     &ast.BooleanLiteral{Value: false},
        Right:    &ast.BooleanLiteral{Value: true},
    },
    want: `{"type":"LogicalExpression","operator":"or","left":{"type":"BooleanLiteral","value":false},"right":{"type":"BooleanLiteral","value":true}}`,
},
*/
#[test]
fn test_json_logical_expression() {
    let n = LogicalExpr {
        base: BaseNode::default(),
        operator: LogicalOperator::OrOperator,
        left: Expression::Boolean(BooleanLit {
            base: BaseNode::default(),
            value: false,
        }),
        right: Expression::Boolean(BooleanLit {
            base: BaseNode::default(),
            value: true,
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"LogicalExpression","operator":"or","left":{"type":"BooleanLiteral","value":false},"right":{"type":"BooleanLiteral","value":true}}"#
    );
    let deserialized: LogicalExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "array expression",
    node: &ast.ArrayExpression{
        Elements: []ast.Expression{&ast.StringLiteral{Value: "hello"}},
    },
    want: `{"type":"ArrayExpression","elements":[{"type":"StringLiteral","value":"hello"}]}`,
},
*/
#[test]
fn test_json_array_expression() {
    let n = ArrayExpr {
        base: BaseNode::default(),
        elements: vec![Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "hello".to_string(),
        })],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ArrayExpression","elements":[{"type":"StringLiteral","value":"hello"}]}"#
    );
    let deserialized: ArrayExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "object expression",
    node: &ast.ObjectExpression{
        Properties: []*ast.Property{{
            Key:   &ast.Identifier{Name: "a"},
            Value: &ast.StringLiteral{Value: "hello"},
        }},
    },
    want: `{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hello"}}]}`,
},
*/
#[test]
fn test_json_object_expression() {
    let n = ObjectExpr {
        base: BaseNode::default(),
        with: None,
        properties: vec![Property {
            base: BaseNode::default(),
            key: PropertyKey::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            value: Some(Expression::StringLit(StringLit {
                base: BaseNode::default(),
                value: "hello".to_string(),
            })),
        }],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hello"}}]}"#
    );
    let deserialized: ObjectExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "object expression with string literal key",
    node: &ast.ObjectExpression{
        Properties: []*ast.Property{{
            Key:   &ast.StringLiteral{Value: "a"},
            Value: &ast.StringLiteral{Value: "hello"},
        }},
    },
    want: `{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"StringLiteral","value":"a"},"value":{"type":"StringLiteral","value":"hello"}}]}`,
},
*/
#[test]
fn test_json_object_expression_with_string_literal_key() {
    let n = ObjectExpr {
        base: BaseNode::default(),
        with: None,
        properties: vec![Property {
            base: BaseNode::default(),
            key: PropertyKey::StringLit(StringLit {
                base: BaseNode::default(),
                value: "a".to_string(),
            }),
            value: Some(Expression::StringLit(StringLit {
                base: BaseNode::default(),
                value: "hello".to_string(),
            })),
        }],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"StringLiteral","value":"a"},"value":{"type":"StringLiteral","value":"hello"}}]}"#
    );
    let deserialized: ObjectExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
    {
        name: "object expression implicit keys",
        node: &ast.ObjectExpression{
            Properties: []*ast.Property{{
                Key: &ast.Identifier{Name: "a"},
            }},
        },
        want: `{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}]}`,
    },
*/
#[test]
fn test_json_object_expression_implicit_keys() {
    let n = ObjectExpr {
        base: BaseNode::default(),
        with: None,
        properties: vec![Property {
            base: BaseNode::default(),
            key: PropertyKey::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            value: None,
        }],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}]}"#
    );
    let deserialized: ObjectExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}

#[test]
fn test_json_object_expression_implicit_keys_and_with() {
    let n = ObjectExpr {
        base: BaseNode::default(),
        with: Some(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        properties: vec![Property {
            base: BaseNode::default(),
            key: PropertyKey::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            value: None,
        }],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ObjectExpression","with":{"type":"Identifier","name":"a"},"properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}]}"#
    );
    let deserialized: ObjectExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "conditional expression",
    node: &ast.ConditionalExpression{
        Test:       &ast.BooleanLiteral{Value: true},
        Alternate:  &ast.StringLiteral{Value: "false"},
        Consequent: &ast.StringLiteral{Value: "true"},
    },
    want: `{"type":"ConditionalExpression","test":{"type":"BooleanLiteral","value":true},"consequent":{"type":"StringLiteral","value":"true"},"alternate":{"type":"StringLiteral","value":"false"}}`,
},
*/
#[test]
fn test_json_conditional_expression() {
    let n = ConditionalExpr {
        base: BaseNode::default(),
        test: Expression::Boolean(BooleanLit {
            base: BaseNode::default(),
            value: true,
        }),
        alternate: Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "false".to_string(),
        }),
        consequent: Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "true".to_string(),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ConditionalExpression","test":{"type":"BooleanLiteral","value":true},"consequent":{"type":"StringLiteral","value":"true"},"alternate":{"type":"StringLiteral","value":"false"}}"#
    );
    let deserialized: ConditionalExpr = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "property",
    node: &ast.Property{
        Key:   &ast.Identifier{Name: "a"},
        Value: &ast.StringLiteral{Value: "hello"},
    },
    want: `{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hello"}}`,
},
*/
#[test]
fn test_json_property() {
    let n = Property {
        base: BaseNode::default(),
        key: PropertyKey::Identifier(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        value: Some(Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "hello".to_string(),
        })),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hello"}}"#
    );
    let deserialized: Property = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "identifier",
    node: &ast.Identifier{
        Name: "a",
    },
    want: `{"type":"Identifier","name":"a"}`,
},
*/
#[test]
fn test_json_identifier() {
    let n = Identifier {
        base: BaseNode::default(),
        name: "a".to_string(),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"Identifier","name":"a"}"#);
    let deserialized: Identifier = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "string literal",
    node: &ast.StringLiteral{
        Value: "hello",
    },
    want: `{"type":"StringLiteral","value":"hello"}`,
},
*/
#[test]
fn test_json_string_literal() {
    let n = StringLit {
        base: BaseNode::default(),
        value: "hello".to_string(),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"StringLiteral","value":"hello"}"#);
    let deserialized: StringLit = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "boolean literal",
    node: &ast.BooleanLiteral{
        Value: true,
    },
    want: `{"type":"BooleanLiteral","value":true}`,
},
*/
#[test]
fn test_json_boolean_literal() {
    let n = BooleanLit {
        base: BaseNode::default(),
        value: true,
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"BooleanLiteral","value":true}"#);
    let deserialized: BooleanLit = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "float literal",
    node: &ast.FloatLiteral{
        Value: 42.1,
    },
    want: `{"type":"FloatLiteral","value":42.1}`,
},
*/
#[test]
fn test_json_float_literal() {
    let n = FloatLit {
        base: BaseNode::default(),
        value: 42.1,
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"FloatLiteral","value":42.1}"#);
    let deserialized: FloatLit = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "integer literal",
    node: &ast.IntegerLiteral{
        Value: math.MaxInt64,
    },
    want: `{"type":"IntegerLiteral","value":"9223372036854775807"}`,
},
*/
#[test]
fn test_json_integer_literal() {
    let n = IntegerLit {
        base: BaseNode::default(),
        value: 9223372036854775807,
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"IntegerLiteral","value":"9223372036854775807"}"#
    );
    let deserialized: IntegerLit = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "unsigned integer literal",
    node: &ast.UnsignedIntegerLiteral{
        Value: math.MaxUint64,
    },
    want: `{"type":"UnsignedIntegerLiteral","value":"18446744073709551615"}`,
},
*/
#[test]
fn test_json_unsigned_integer_literal() {
    let n = UintLit {
        base: BaseNode::default(),
        value: 18446744073709551615,
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"UnsignedIntegerLiteral","value":"18446744073709551615"}"#
    );
    let deserialized: UintLit = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "regexp literal",
    node: &ast.RegexpLiteral{
        Value: regexp.MustCompile(`.*`),
    },
    want: `{"type":"RegexpLiteral","value":".*"}`,
},
*/
#[test]
fn test_json_regexp_literal() {
    let n = RegexpLit {
        base: BaseNode::default(),
        value: ".*".to_string(),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"RegexpLiteral","value":".*"}"#);
    let deserialized: RegexpLit = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "duration literal",
    node: &ast.DurationLiteral{
        Values: []ast.Duration{
            {
                Magnitude: 1,
                Unit:      "h",
            },
            {
                Magnitude: 1,
                Unit:      "h",
            },
        },
    },
    want: `{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"},{"magnitude":1,"unit":"h"}]}`,
},
*/
#[test]
fn test_json_duration_literal() {
    let n = DurationLit {
        base: BaseNode::default(),
        values: vec![
            Duration {
                magnitude: 1,
                unit: "h".to_string(),
            },
            Duration {
                magnitude: 1,
                unit: "h".to_string(),
            },
        ],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"},{"magnitude":1,"unit":"h"}]}"#
    );
    let deserialized: DurationLit = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
/*
{
    name: "datetime literal",
    node: &ast.DateTimeLiteral{
        Value: time.Date(2017, 8, 8, 8, 8, 8, 8, time.UTC),
    },
    want: `{"type":"DateTimeLiteral","value":"2017-08-08T08:08:08.000000008Z"}`,
},
*/
// NOTE(affo): the output has changed to 2017-08-08T08:08:08.000000008+00:00.
// There is no problem in doing that, because it is a RFC3339 compliant notation,
// and it will be parsed by the Go backend.
#[test]
fn test_json_datetime_literal() {
    let n = DateTimeLit {
        base: BaseNode::default(),
        value: FixedOffset::east(0)
            .ymd(2017, 8, 8)
            .and_hms_nano(8, 8, 8, 8),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"DateTimeLiteral","value":"2017-08-08T08:08:08.000000008+00:00"}"#
    );
    let deserialized: DateTimeLit = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}

#[test]
fn test_object_expression_with_source_locations_and_errors() {
    let n = ObjectExpr {
        base: BaseNode {
            location: SourceLocation {
                file: Some("foo.flux".to_string()),
                start: Position { line: 1, column: 1 },
                end: Position {
                    line: 1,
                    column: 13,
                },
                source: Some("{a: \"hello\"}".to_string()),
            },
            errors: vec![],
        },
        with: None,
        properties: vec![Property {
            base: BaseNode {
                location: SourceLocation {
                    file: Some("foo.flux".to_string()),
                    start: Position { line: 1, column: 2 },
                    end: Position {
                        line: 1,
                        column: 12,
                    },
                    source: Some("a: \"hello\"".to_string()),
                },
                errors: vec!["an error".to_string()],
            },
            key: PropertyKey::Identifier(Identifier {
                base: BaseNode {
                    location: SourceLocation {
                        file: Some("foo.flux".to_string()),
                        start: Position { line: 1, column: 2 },
                        end: Position { line: 1, column: 3 },
                        source: Some("a".to_string()),
                    },
                    errors: vec![],
                },
                name: "a".to_string(),
            }),
            value: Some(Expression::StringLit(StringLit {
                base: BaseNode {
                    location: SourceLocation {
                        file: Some("foo.flux".to_string()),
                        start: Position { line: 1, column: 5 },
                        end: Position {
                            line: 1,
                            column: 12,
                        },
                        source: Some("\"hello\"".to_string()),
                    },
                    errors: vec!["an error".to_string(), "another error".to_string()],
                },
                value: "hello".to_string(),
            })),
        }],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ObjectExpression","location":{"file":"foo.flux","start":{"line":1,"column":1},"end":{"line":1,"column":13},"source":"{a: \"hello\"}"},"properties":[{"type":"Property","location":{"file":"foo.flux","start":{"line":1,"column":2},"end":{"line":1,"column":12},"source":"a: \"hello\""},"errors":[{"msg":"an error"}],"key":{"type":"Identifier","location":{"file":"foo.flux","start":{"line":1,"column":2},"end":{"line":1,"column":3},"source":"a"},"name":"a"},"value":{"type":"StringLiteral","location":{"file":"foo.flux","start":{"line":1,"column":5},"end":{"line":1,"column":12},"source":"\"hello\""},"errors":[{"msg":"an error"},{"msg":"another error"}],"value":"hello"}}]}"#
    );
    // TODO(affo): leaving proper error deserialization for the future.
    // let deserialized: ObjectExpr = serde_json::from_str(serialized.as_str()).unwrap();
    // assert_eq!(deserialized, n)
}
