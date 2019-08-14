// NOTE: These test cases directly match ast/json_test.go.
// Every test is preceded by the correspondent test case in golang.
use super::*;
use chrono::TimeZone;

/*
{
    name: "string interpolation",
    node: &ast.StringExpression{
        Parts: []ast.StringExpressionPart{
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
    let n = StringExpression {
        base: BaseNode::default(),
        parts: vec![
            StringExpressionPart::Text(TextPart {
                base: BaseNode::default(),
                value: "a = ".to_string(),
            }),
            StringExpressionPart::Expr(InterpolatedPart {
                base: BaseNode::default(),
                expression: Expression::Idt(Identifier {
                    base: BaseNode::default(),
                    name: "a".to_string(),
                }),
            }),
        ],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"StringExpression","parts":[{"type":"TextPart","value":"a = "},{"type":"InterpolatedPart","expression":{"type":"Identifier","name":"a"}}]}"#);
    let deserialized: StringExpression = serde_json::from_str(serialized.as_str()).unwrap();
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
        body: vec![Statement::Expr(ExpressionStatement {
            base: BaseNode::default(),
            expression: Expression::Str(StringLiteral {
                base: Default::default(),
                value: "hello".to_string(),
            }),
        })],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"File","package":null,"imports":[],"body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}"#);
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
            path: StringLiteral {
                base: BaseNode::default(),
                value: "path/bar".to_string(),
            },
        }],
        name: String::new(),
        body: vec![Statement::Expr(ExpressionStatement {
            base: BaseNode::default(),
            expression: Expression::Str(StringLiteral {
                base: Default::default(),
                value: "hello".to_string(),
            }),
        })],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"File","package":{"type":"PackageClause","name":{"type":"Identifier","name":"foo"}},"imports":[{"type":"ImportDeclaration","as":{"type":"Identifier","name":"b"},"path":{"type":"StringLiteral","value":"path/bar"}}],"body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}"#);
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
        body: vec![Statement::Expr(ExpressionStatement {
            base: BaseNode::default(),
            expression: Expression::Str(StringLiteral {
                base: Default::default(),
                value: "hello".to_string(),
            }),
        })],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"Block","body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}"#);
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
    let n = ExpressionStatement {
        base: BaseNode::default(),
        expression: Expression::Str(StringLiteral {
            base: BaseNode::default(),
            value: "hello".to_string(),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}"#
    );
    let deserialized: ExpressionStatement = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = ReturnStatement {
        base: BaseNode::default(),
        argument: Expression::Str(StringLiteral {
            base: BaseNode::default(),
            value: "hello".to_string(),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ReturnStatement","argument":{"type":"StringLiteral","value":"hello"}}"#
    );
    let deserialized: ReturnStatement = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = OptionStatement {
        base: BaseNode::default(),
        assignment: Assignment::Variable(VariableAssignment {
            base: BaseNode::default(),
            id: Identifier {
                base: BaseNode::default(),
                name: "task".to_string(),
            },
            init: Expression::Obj(Box::new(ObjectExpression {
                base: BaseNode::default(),
                with: None,
                properties: vec![
                    Property {
                        base: BaseNode::default(),
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode::default(),
                            name: "name".to_string(),
                        }),
                        value: Some(Expression::Str(StringLiteral {
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
                        value: Some(Expression::Dur(DurationLiteral {
                            base: Default::default(),
                            values: vec![Duration {
                                magnitude: 1,
                                unit: "h".to_string(),
                            }],
                        })),
                    },
                ],
            })),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"OptionStatement","assignment":{"type":"VariableAssignment","id":{"type":"Identifier","name":"task"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"name"},"value":{"type":"StringLiteral","value":"foo"}},{"type":"Property","key":{"type":"Identifier","name":"every"},"value":{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"}]}}]}}}"#);
    let deserialized: OptionStatement = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = BuiltinStatement {
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
    let deserialized: BuiltinStatement = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = TestStatement {
        base: BaseNode::default(),
        assignment: VariableAssignment {
            base: BaseNode::default(),
            id: Identifier {
                base: BaseNode::default(),
                name: "mean".to_string(),
            },
            init: Expression::Obj(Box::new(ObjectExpression {
                base: BaseNode::default(),
                with: None,
                properties: vec![
                    Property {
                        base: BaseNode::default(),
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode::default(),
                            name: "want".to_string(),
                        }),
                        value: Some(Expression::Int(IntegerLiteral {
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
                        value: Some(Expression::Int(IntegerLiteral {
                            base: Default::default(),
                            value: 0,
                        })),
                    },
                ],
            })),
        },
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"TestStatement","assignment":{"type":"VariableAssignment","id":{"type":"Identifier","name":"mean"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"want"},"value":{"type":"IntegerLiteral","value":"0"}},{"type":"Property","key":{"type":"Identifier","name":"got"},"value":{"type":"IntegerLiteral","value":"0"}}]}}}"#);
    let deserialized: TestStatement = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = OptionStatement {
        base: BaseNode::default(),
        assignment: Assignment::Member(MemberAssignment {
            base: BaseNode::default(),
            member: MemberExpression {
                base: BaseNode::default(),
                object: Expression::Idt(Identifier {
                    base: BaseNode::default(),
                    name: "alert".to_string(),
                }),
                property: PropertyKey::Identifier(Identifier {
                    base: BaseNode::default(),
                    name: "state".to_string(),
                }),
            },
            init: Expression::Str(StringLiteral {
                base: Default::default(),
                value: "Warning".to_string(),
            }),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"OptionStatement","assignment":{"type":"MemberAssignment","member":{"type":"MemberExpression","object":{"type":"Identifier","name":"alert"},"property":{"type":"Identifier","name":"state"}},"init":{"type":"StringLiteral","value":"Warning"}}}"#);
    let deserialized: OptionStatement = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = VariableAssignment {
        base: BaseNode::default(),
        id: Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        },
        init: Expression::Str(StringLiteral {
            base: BaseNode::default(),
            value: "hello".to_string(),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"VariableAssignment","id":{"type":"Identifier","name":"a"},"init":{"type":"StringLiteral","value":"hello"}}"#);
    let deserialized: VariableAssignment = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = CallExpression {
        base: BaseNode::default(),
        callee: Expression::Idt(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        arguments: vec![Expression::Str(StringLiteral {
            base: BaseNode::default(),
            value: "hello".to_string(),
        })],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"CallExpression","callee":{"type":"Identifier","name":"a"},"arguments":[{"type":"StringLiteral","value":"hello"}]}"#);
    let deserialized: CallExpression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = PipeExpression {
        base: BaseNode::default(),
        argument: Expression::Idt(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        call: CallExpression {
            base: BaseNode::default(),
            callee: Expression::Idt(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            arguments: vec![Expression::Str(StringLiteral {
                base: BaseNode::default(),
                value: "hello".to_string(),
            })],
        },
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"PipeExpression","argument":{"type":"Identifier","name":"a"},"call":{"type":"CallExpression","callee":{"type":"Identifier","name":"a"},"arguments":[{"type":"StringLiteral","value":"hello"}]}}"#);
    let deserialized: PipeExpression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = MemberExpression {
        base: BaseNode::default(),
        object: Expression::Idt(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        property: PropertyKey::Identifier(Identifier {
            base: BaseNode::default(),
            name: "b".to_string(),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"MemberExpression","object":{"type":"Identifier","name":"a"},"property":{"type":"Identifier","name":"b"}}"#);
    let deserialized: MemberExpression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = MemberExpression {
        base: BaseNode::default(),
        object: Expression::Idt(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        property: PropertyKey::StringLiteral(StringLiteral {
            base: BaseNode::default(),
            value: "b".to_string(),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"MemberExpression","object":{"type":"Identifier","name":"a"},"property":{"type":"StringLiteral","value":"b"}}"#);
    let deserialized: MemberExpression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = IndexExpression {
        base: BaseNode::default(),
        array: Expression::Idt(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        index: Expression::Int(IntegerLiteral {
            base: BaseNode::default(),
            value: 3,
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"IndexExpression","array":{"type":"Identifier","name":"a"},"index":{"type":"IntegerLiteral","value":"3"}}"#);
    let deserialized: IndexExpression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = FunctionExpression {
        base: BaseNode::default(),
        params: vec![Property {
            base: BaseNode::default(),
            key: PropertyKey::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            value: None,
        }],
        body: FunctionBody::Expr(Expression::Str(StringLiteral {
            base: BaseNode::default(),
            value: "hello".to_string(),
        })),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"FunctionExpression","params":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}],"body":{"type":"StringLiteral","value":"hello"}}"#);
    let deserialized: FunctionExpression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = BinaryExpression {
        base: BaseNode::default(),
        operator: OperatorKind::AdditionOperator,
        left: Expression::Str(StringLiteral {
            base: BaseNode::default(),
            value: "hello".to_string(),
        }),
        right: Expression::Str(StringLiteral {
            base: BaseNode::default(),
            value: "world".to_string(),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"BinaryExpression","operator":"+","left":{"type":"StringLiteral","value":"hello"},"right":{"type":"StringLiteral","value":"world"}}"#);
    let deserialized: BinaryExpression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = UnaryExpression {
        base: BaseNode::default(),
        operator: OperatorKind::NotOperator,
        argument: Expression::Bool(BooleanLiteral {
            base: BaseNode::default(),
            value: true,
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"UnaryExpression","operator":"not","argument":{"type":"BooleanLiteral","value":true}}"#);
    let deserialized: UnaryExpression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = LogicalExpression {
        base: BaseNode::default(),
        operator: LogicalOperatorKind::OrOperator,
        left: Expression::Bool(BooleanLiteral {
            base: BaseNode::default(),
            value: false,
        }),
        right: Expression::Bool(BooleanLiteral {
            base: BaseNode::default(),
            value: true,
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"LogicalExpression","operator":"or","left":{"type":"BooleanLiteral","value":false},"right":{"type":"BooleanLiteral","value":true}}"#);
    let deserialized: LogicalExpression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = ArrayExpression {
        base: BaseNode::default(),
        elements: vec![Expression::Str(StringLiteral {
            base: BaseNode::default(),
            value: "hello".to_string(),
        })],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ArrayExpression","elements":[{"type":"StringLiteral","value":"hello"}]}"#
    );
    let deserialized: ArrayExpression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = ObjectExpression {
        base: BaseNode::default(),
        with: None,
        properties: vec![Property {
            base: BaseNode::default(),
            key: PropertyKey::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            value: Some(Expression::Str(StringLiteral {
                base: BaseNode::default(),
                value: "hello".to_string(),
            })),
        }],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hello"}}]}"#);
    let deserialized: ObjectExpression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = ObjectExpression {
        base: BaseNode::default(),
        with: None,
        properties: vec![Property {
            base: BaseNode::default(),
            key: PropertyKey::StringLiteral(StringLiteral {
                base: BaseNode::default(),
                value: "a".to_string(),
            }),
            value: Some(Expression::Str(StringLiteral {
                base: BaseNode::default(),
                value: "hello".to_string(),
            })),
        }],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"StringLiteral","value":"a"},"value":{"type":"StringLiteral","value":"hello"}}]}"#);
    let deserialized: ObjectExpression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = ObjectExpression {
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
    assert_eq!(serialized, r#"{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}]}"#);
    let deserialized: ObjectExpression = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}

#[test]
fn test_json_object_expression_implicit_keys_and_with() {
    let n = ObjectExpression {
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
    assert_eq!(serialized, r#"{"type":"ObjectExpression","with":{"type":"Identifier","name":"a"},"properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}]}"#);
    let deserialized: ObjectExpression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = ConditionalExpression {
        base: BaseNode::default(),
        test: Expression::Bool(BooleanLiteral {
            base: BaseNode::default(),
            value: true,
        }),
        alternate: Expression::Str(StringLiteral {
            base: BaseNode::default(),
            value: "false".to_string(),
        }),
        consequent: Expression::Str(StringLiteral {
            base: BaseNode::default(),
            value: "true".to_string(),
        }),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"ConditionalExpression","test":{"type":"BooleanLiteral","value":true},"consequent":{"type":"StringLiteral","value":"true"},"alternate":{"type":"StringLiteral","value":"false"}}"#);
    let deserialized: ConditionalExpression = serde_json::from_str(serialized.as_str()).unwrap();
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
        value: Some(Expression::Str(StringLiteral {
            base: BaseNode::default(),
            value: "hello".to_string(),
        })),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hello"}}"#);
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
    let n = StringLiteral {
        base: BaseNode::default(),
        value: "hello".to_string(),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"StringLiteral","value":"hello"}"#);
    let deserialized: StringLiteral = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = BooleanLiteral {
        base: BaseNode::default(),
        value: true,
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"BooleanLiteral","value":true}"#);
    let deserialized: BooleanLiteral = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = FloatLiteral {
        base: BaseNode::default(),
        value: 42.1,
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"FloatLiteral","value":42.1}"#);
    let deserialized: FloatLiteral = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = IntegerLiteral {
        base: BaseNode::default(),
        value: 9223372036854775807,
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"IntegerLiteral","value":"9223372036854775807"}"#
    );
    let deserialized: IntegerLiteral = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = UnsignedIntegerLiteral {
        base: BaseNode::default(),
        value: 18446744073709551615,
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"UnsignedIntegerLiteral","value":"18446744073709551615"}"#
    );
    let deserialized: UnsignedIntegerLiteral = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = RegexpLiteral {
        base: BaseNode::default(),
        value: ".*".to_string(),
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"RegexpLiteral","value":".*"}"#);
    let deserialized: RegexpLiteral = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = DurationLiteral {
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
    assert_eq!(serialized, r#"{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"},{"magnitude":1,"unit":"h"}]}"#);
    let deserialized: DurationLiteral = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = DateTimeLiteral {
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
    let deserialized: DateTimeLiteral = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
