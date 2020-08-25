// NOTE: These test cases directly match ast/json_test.go.
// Every test is preceded by the correspondent test case in golang.
use super::*;
use chrono::TimeZone;

/// ast_with_every_kind_of_node returns an AST that contains
/// every kind of node, which can be useful for testing.
pub fn ast_with_every_kind_of_node() -> Package {
    let f = vec![
        crate::parser::parse_string(
            "test1",
            r#"
package mypkg
import "my_other_pkg"
import "yet_another_pkg"
option now = () => (2030-01-01T00:00:00Z)
option foo.bar = "baz"
builtin foo

# // bad stmt

test aggregate_window_empty = () => ({
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    fn: (table=<-) =>
        table
            |> range(start: 2018-05-22T19:53:26Z, stop: 2018-05-22T19:55:00Z)
            |> aggregateWindow(every: 30s, fn: sum),
})
"#,
        ),
        crate::parser::parse_string(
            "test2",
            r#"
a

arr = [0, 1, 2]
f = (i) => i
ff = (i=<-, j) => {
  k = i + j
  return k
}
b = z and y
b = z or y
o = {red: "red", "blue": 30}
empty_obj = {}
m = o.red
i = arr[0]
n = 10 - 5 + 10
n = 10 / 5 * 10
m = 13 % 3
p = 2^10
b = 10 < 30
b = 10 <= 30
b = 10 > 30
b = 10 >= 30
eq = 10 == 10
neq = 11 != 10
b = not false
e = exists o.red
tables |> f()
fncall = id(v: 20)
fncall2 = foo(v: 20, w: "bar")
fncall_short_form_arg(arg)
fncall_short_form_args(arg0, arg1)
v = if true then 70.0 else 140.0
ans = "the answer is ${v}"
paren = (1)

i = 1
f = 1.0
s = "foo"
d = 10s
b = true
dt = 2030-01-01T00:00:00Z
re =~ /foo/
re !~ /foo/
bad_expr = 3 * / 1
"#,
        ),
    ];
    Package {
        base: BaseNode {
            ..BaseNode::default()
        },
        path: String::from("./"),
        package: String::from("test"),
        files: f,
    }
}

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
    let n = Expression::StringExpr(Box::new(StringExpr {
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
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"StringExpression","parts":[{"type":"TextPart","value":"a = "},{"type":"InterpolatedPart","expression":{"type":"Identifier","name":"a"}}]}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Paren(Box::new(ParenExpr {
        base: BaseNode::default(),
        lparen: None,
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
        rparen: None,
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ParenExpression","expression":{"type":"StringExpression","parts":[{"type":"TextPart","value":"a = "},{"type":"InterpolatedPart","expression":{"type":"Identifier","name":"a"}}]}}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
        metadata: String::new(),
        body: vec![Statement::Expr(Box::new(ExprStmt {
            base: BaseNode::default(),
            expression: Expression::StringLit(StringLit {
                base: Default::default(),
                value: "hello".to_string(),
            }),
        }))],
        eof: None,
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
        metadata: String::from("parser-type=none"),
        body: vec![Statement::Expr(Box::new(ExprStmt {
            base: BaseNode::default(),
            expression: Expression::StringLit(StringLit {
                base: Default::default(),
                value: "hello".to_string(),
            }),
        }))],
        eof: None,
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"File","metadata":"parser-type=none","package":{"type":"PackageClause","name":{"name":"foo"}},"imports":[{"type":"ImportDeclaration","as":{"name":"b"},"path":{"value":"path/bar"}}],"body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}"#
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
    let n = FunctionBody::Block(Block {
        base: BaseNode::default(),
        lbrace: None,
        body: vec![Statement::Expr(Box::new(ExprStmt {
            base: BaseNode::default(),
            expression: Expression::StringLit(StringLit {
                base: Default::default(),
                value: "hello".to_string(),
            }),
        }))],
        rbrace: None,
    });
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"Block","body":[{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}]}"#
    );
    let deserialized: FunctionBody = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Statement::Expr(Box::new(ExprStmt {
        base: BaseNode::default(),
        expression: Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "hello".to_string(),
        }),
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ExpressionStatement","expression":{"type":"StringLiteral","value":"hello"}}"#
    );
    let deserialized: Statement = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}

#[test]
fn test_json_namedtype() {
    let n = MonoType::Basic(NamedType {
        base: BaseNode::default(),
        name: Identifier {
            base: BaseNode::default(),
            name: "int".to_string(),
        },
    });
    let serialized = serde_json::to_string(&n).unwrap();
    // {"type":"Identifier","name":...}
    assert_eq!(serialized, r#"{"type":"NamedType","name":{"name":"int"}}"#);
    let deserialized: MonoType = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
#[test]
fn test_json_tvartype() {
    let n = MonoType::Tvar(TvarType {
        base: BaseNode::default(),
        name: Identifier {
            base: BaseNode::default(),
            name: "A".to_string(),
        },
    });
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"TvarType","name":{"name":"A"}}"#);
    let deserialized: MonoType = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Statement::Return(Box::new(ReturnStmt {
        base: BaseNode::default(),
        argument: Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "hello".to_string(),
        }),
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ReturnStatement","argument":{"type":"StringLiteral","value":"hello"}}"#
    );
    let deserialized: Statement = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}

#[test]
fn test_json_type_expression() {
    let n = TypeExpression {
        base: BaseNode::default(),
        monotype: MonoType::Function(Box::new(FunctionType {
            base: BaseNode::default(),
            parameters: vec![
                ParameterType::Required {
                    base: BaseNode::default(),
                    name: Identifier {
                        base: BaseNode::default(),
                        name: "a".to_string(),
                    },
                    monotype: MonoType::Tvar(TvarType {
                        base: BaseNode::default(),
                        name: Identifier {
                            base: BaseNode::default(),
                            name: "T".to_string(),
                        },
                    }),
                },
                ParameterType::Required {
                    base: BaseNode::default(),
                    name: Identifier {
                        base: BaseNode::default(),
                        name: "b".to_string(),
                    },
                    monotype: MonoType::Tvar(TvarType {
                        base: BaseNode::default(),
                        name: Identifier {
                            base: BaseNode::default(),
                            name: "T".to_string(),
                        },
                    }),
                },
            ],
            monotype: MonoType::Tvar(TvarType {
                base: BaseNode::default(),
                name: Identifier {
                    base: BaseNode::default(),
                    name: "T".to_string(),
                },
            }),
        })),
        constraints: vec![TypeConstraint {
            base: BaseNode::default(),
            tvar: Identifier {
                base: BaseNode::default(),
                name: "T".to_string(),
            },
            kinds: vec![
                Identifier {
                    base: BaseNode::default(),
                    name: "Addable".to_string(),
                },
                Identifier {
                    base: BaseNode::default(),
                    name: "Divisible".to_string(),
                },
            ],
        }],
    };
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"monotype":{"type":"FunctionType","parameters":[{"type":"Required","name":{"name":"a"},"monotype":{"type":"TvarType","name":{"name":"T"}}},{"type":"Required","name":{"name":"b"},"monotype":{"type":"TvarType","name":{"name":"T"}}}],"monotype":{"type":"TvarType","name":{"name":"T"}}},"constraints":[{"tvar":{"name":"T"},"kinds":[{"name":"Addable"},{"name":"Divisible"}]}]}"#
    );
    let deserialized: TypeExpression = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}

#[test]
fn test_json_array() {
    let n = MonoType::Array(Box::new(ArrayType {
        base: BaseNode::default(),
        element: MonoType::Array(Box::new(ArrayType {
            base: BaseNode::default(),
            element: MonoType::Basic(NamedType {
                base: BaseNode::default(),
                name: Identifier {
                    base: BaseNode::default(),
                    name: "A".to_string(),
                },
            }),
        })),
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ArrayType","element":{"type":"ArrayType","element":{"type":"NamedType","name":{"name":"A"}}}}"#
    );
    let deserialized: MonoType = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
#[test]
fn test_json_record() {
    let n = MonoType::Record(RecordType {
        base: BaseNode::default(),
        tvar: Some(Identifier {
            base: BaseNode::default(),
            name: "A".to_string(),
        }),
        properties: vec![PropertyType {
            base: BaseNode::default(),
            name: Identifier {
                base: BaseNode::default(),
                name: "A".to_string(),
            },
            monotype: MonoType::Basic(NamedType {
                base: BaseNode::default(),
                name: Identifier {
                    base: BaseNode::default(),
                    name: "int".to_string(),
                },
            }),
        }],
    });
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"RecordType","tvar":{"name":"A"},"properties":[{"name":{"name":"A"},"monotype":{"type":"NamedType","name":{"name":"int"}}}]}"#
    );
    let deserialized: MonoType = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
#[test]
fn test_json_record_no_tvar_no_properties() {
    let n = MonoType::Record(RecordType {
        base: BaseNode::default(),
        tvar: None,
        properties: vec![],
    });
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"RecordType","properties":[]}"#);
    let deserialized: MonoType = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}
#[test]
fn test_json_record_no_tvar() {
    let n = MonoType::Record(RecordType {
        base: BaseNode::default(),
        tvar: None,
        properties: vec![PropertyType {
            base: BaseNode::default(),
            name: Identifier {
                base: BaseNode::default(),
                name: "A".to_string(),
            },
            monotype: MonoType::Basic(NamedType {
                base: BaseNode::default(),
                name: Identifier {
                    base: BaseNode::default(),
                    name: "int".to_string(),
                },
            }),
        }],
    });
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"RecordType","properties":[{"name":{"name":"A"},"monotype":{"type":"NamedType","name":{"name":"int"}}}]}"#
    );
    let deserialized: MonoType = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}

#[test]
fn test_json_functiontype_no_params() {
    let n = MonoType::Function(Box::new(FunctionType {
        base: BaseNode::default(),
        parameters: vec![],
        monotype: MonoType::Basic(NamedType {
            base: BaseNode::default(),
            name: Identifier {
                base: BaseNode::default(),
                name: "int".to_string(),
            },
        }),
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"FunctionType","parameters":[],"monotype":{"type":"NamedType","name":{"name":"int"}}}"#
    );
    let deserialized: MonoType = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}

#[test]
fn test_json_functiontype_required() {
    let n = MonoType::Function(Box::new(FunctionType {
        base: BaseNode::default(),
        parameters: vec![ParameterType::Required {
            base: BaseNode::default(),
            name: Identifier {
                base: BaseNode::default(),
                name: "B".to_string(),
            },
            monotype: MonoType::Basic(NamedType {
                base: BaseNode::default(),
                name: Identifier {
                    base: BaseNode::default(),
                    name: "string".to_string(),
                },
            }),
        }],
        monotype: MonoType::Basic(NamedType {
            base: BaseNode::default(),
            name: Identifier {
                base: BaseNode::default(),
                name: "uint".to_string(),
            },
        }),
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"FunctionType","parameters":[{"type":"Required","name":{"name":"B"},"monotype":{"type":"NamedType","name":{"name":"string"}}}],"monotype":{"type":"NamedType","name":{"name":"uint"}}}"#
    );
    let deserialized: MonoType = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}

#[test]
fn test_json_functiontype_optional() {
    let n = MonoType::Function(Box::new(FunctionType {
        base: BaseNode::default(),
        parameters: vec![ParameterType::Optional {
            base: BaseNode::default(),
            name: Identifier {
                base: BaseNode::default(),
                name: "A".to_string(),
            },
            monotype: MonoType::Basic(NamedType {
                base: BaseNode::default(),
                name: Identifier {
                    base: BaseNode::default(),
                    name: "int".to_string(),
                },
            }),
        }],
        monotype: MonoType::Basic(NamedType {
            base: BaseNode::default(),
            name: Identifier {
                base: BaseNode::default(),
                name: "int".to_string(),
            },
        }),
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"FunctionType","parameters":[{"type":"Optional","name":{"name":"A"},"monotype":{"type":"NamedType","name":{"name":"int"}}}],"monotype":{"type":"NamedType","name":{"name":"int"}}}"#
    );
    let deserialized: MonoType = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}

#[test]
fn test_json_functiontype_named_pipe() {
    let n = MonoType::Function(Box::new(FunctionType {
        base: BaseNode::default(),
        parameters: vec![ParameterType::Pipe {
            base: BaseNode::default(),
            name: Some(Identifier {
                base: BaseNode::default(),
                name: "A".to_string(),
            }),
            monotype: MonoType::Basic(NamedType {
                base: BaseNode::default(),
                name: Identifier {
                    base: BaseNode::default(),
                    name: "int".to_string(),
                },
            }),
        }],
        monotype: MonoType::Basic(NamedType {
            base: BaseNode::default(),
            name: Identifier {
                base: BaseNode::default(),
                name: "int".to_string(),
            },
        }),
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"FunctionType","parameters":[{"type":"Pipe","name":{"name":"A"},"monotype":{"type":"NamedType","name":{"name":"int"}}}],"monotype":{"type":"NamedType","name":{"name":"int"}}}"#
    );
    let deserialized: MonoType = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}

#[test]
fn test_json_functiontype_unnamed_pipe() {
    let n = MonoType::Function(Box::new(FunctionType {
        base: BaseNode::default(),
        parameters: vec![ParameterType::Pipe {
            base: BaseNode::default(),
            name: None,
            monotype: MonoType::Basic(NamedType {
                base: BaseNode::default(),
                name: Identifier {
                    base: BaseNode::default(),
                    name: "int".to_string(),
                },
            }),
        }],
        monotype: MonoType::Basic(NamedType {
            base: BaseNode::default(),
            name: Identifier {
                base: BaseNode::default(),
                name: "int".to_string(),
            },
        }),
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"FunctionType","parameters":[{"type":"Pipe","monotype":{"type":"NamedType","name":{"name":"int"}}}],"monotype":{"type":"NamedType","name":{"name":"int"}}}"#
    );
    let deserialized: MonoType = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Statement::Option(Box::new(OptionStmt {
        base: BaseNode::default(),
        assignment: Assignment::Variable(Box::new(VariableAssgn {
            base: BaseNode::default(),
            id: Identifier {
                base: BaseNode::default(),
                name: "task".to_string(),
            },
            init: Expression::Object(Box::new(ObjectExpr {
                base: BaseNode::default(),
                lbrace: None,
                with: None,
                properties: vec![
                    Property {
                        base: BaseNode::default(),
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode::default(),
                            name: "name".to_string(),
                        }),
                        separator: None,
                        value: Some(Expression::StringLit(StringLit {
                            base: Default::default(),
                            value: "foo".to_string(),
                        })),
                        comma: None,
                    },
                    Property {
                        base: BaseNode::default(),
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode::default(),
                            name: "every".to_string(),
                        }),
                        separator: None,
                        value: Some(Expression::Duration(DurationLit {
                            base: Default::default(),
                            values: vec![Duration {
                                magnitude: 1,
                                unit: "h".to_string(),
                            }],
                        })),
                        comma: None,
                    },
                ],
                rbrace: None,
            })),
        })),
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"OptionStatement","assignment":{"type":"VariableAssignment","id":{"name":"task"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"name"},"value":{"type":"StringLiteral","value":"foo"}},{"type":"Property","key":{"type":"Identifier","name":"every"},"value":{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"}]}}]}}}"#
    );
    let deserialized: Statement = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Statement::Builtin(Box::new(BuiltinStmt {
        base: BaseNode::default(),
        id: Identifier {
            base: BaseNode::default(),
            name: "task".to_string(),
        },
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"BuiltinStatement","id":{"name":"task"}}"#
    );
    let deserialized: Statement = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Statement::Test(Box::new(TestStmt {
        base: BaseNode::default(),
        assignment: VariableAssgn {
            base: BaseNode::default(),
            id: Identifier {
                base: BaseNode::default(),
                name: "mean".to_string(),
            },
            init: Expression::Object(Box::new(ObjectExpr {
                base: BaseNode::default(),
                lbrace: None,
                with: None,
                properties: vec![
                    Property {
                        base: BaseNode::default(),
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode::default(),
                            name: "want".to_string(),
                        }),
                        separator: None,
                        value: Some(Expression::Integer(IntegerLit {
                            base: Default::default(),
                            value: 0,
                        })),
                        comma: None,
                    },
                    Property {
                        base: BaseNode::default(),
                        key: PropertyKey::Identifier(Identifier {
                            base: BaseNode::default(),
                            name: "got".to_string(),
                        }),
                        separator: None,
                        value: Some(Expression::Integer(IntegerLit {
                            base: Default::default(),
                            value: 0,
                        })),
                        comma: None,
                    },
                ],
                rbrace: None,
            })),
        },
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"TestStatement","assignment":{"id":{"name":"mean"},"init":{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"want"},"value":{"type":"IntegerLiteral","value":"0"}},{"type":"Property","key":{"type":"Identifier","name":"got"},"value":{"type":"IntegerLiteral","value":"0"}}]}}}"#
    );
    let deserialized: Statement = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Statement::Option(Box::new(OptionStmt {
        base: BaseNode::default(),
        assignment: Assignment::Member(Box::new(MemberAssgn {
            base: BaseNode::default(),
            member: MemberExpr {
                base: BaseNode::default(),
                object: Expression::Identifier(Identifier {
                    base: BaseNode::default(),
                    name: "alert".to_string(),
                }),
                lbrack: None,
                property: PropertyKey::Identifier(Identifier {
                    base: BaseNode::default(),
                    name: "state".to_string(),
                }),
                rbrack: None,
            },
            init: Expression::StringLit(StringLit {
                base: Default::default(),
                value: "Warning".to_string(),
            }),
        })),
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"OptionStatement","assignment":{"type":"MemberAssignment","member":{"object":{"type":"Identifier","name":"alert"},"property":{"type":"Identifier","name":"state"}},"init":{"type":"StringLiteral","value":"Warning"}}}"#
    );
    let deserialized: Statement = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Statement::Variable(Box::new(VariableAssgn {
        base: BaseNode::default(),
        id: Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        },
        init: Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "hello".to_string(),
        }),
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"VariableAssignment","id":{"name":"a"},"init":{"type":"StringLiteral","value":"hello"}}"#
    );
    let deserialized: Statement = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Call(Box::new(CallExpr {
        base: BaseNode::default(),
        callee: Expression::Identifier(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        lparen: None,
        arguments: vec![Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "hello".to_string(),
        })],
        rparen: None,
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"CallExpression","callee":{"type":"Identifier","name":"a"},"arguments":[{"type":"StringLiteral","value":"hello"}]}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
fn test_json_call_expression_empty_arguments() {
    let n = Expression::Call(Box::new(CallExpr {
        base: BaseNode::default(),
        callee: Expression::Identifier(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        lparen: None,
        arguments: vec![],
        rparen: None,
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"CallExpression","callee":{"type":"Identifier","name":"a"}}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::PipeExpr(Box::new(PipeExpr {
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
            lparen: None,
            arguments: vec![Expression::StringLit(StringLit {
                base: BaseNode::default(),
                value: "hello".to_string(),
            })],
            rparen: None,
        },
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"PipeExpression","argument":{"type":"Identifier","name":"a"},"call":{"callee":{"type":"Identifier","name":"a"},"arguments":[{"type":"StringLiteral","value":"hello"}]}}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Member(Box::new(MemberExpr {
        base: BaseNode::default(),
        object: Expression::Identifier(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        lbrack: None,
        property: PropertyKey::Identifier(Identifier {
            base: BaseNode::default(),
            name: "b".to_string(),
        }),
        rbrack: None,
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"MemberExpression","object":{"type":"Identifier","name":"a"},"property":{"type":"Identifier","name":"b"}}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Member(Box::new(MemberExpr {
        base: BaseNode::default(),
        object: Expression::Identifier(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        lbrack: None,
        property: PropertyKey::StringLit(StringLit {
            base: BaseNode::default(),
            value: "b".to_string(),
        }),
        rbrack: None,
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"MemberExpression","object":{"type":"Identifier","name":"a"},"property":{"type":"StringLiteral","value":"b"}}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Index(Box::new(IndexExpr {
        base: BaseNode::default(),
        array: Expression::Identifier(Identifier {
            base: BaseNode::default(),
            name: "a".to_string(),
        }),
        lbrack: None,
        index: Expression::Integer(IntegerLit {
            base: BaseNode::default(),
            value: 3,
        }),
        rbrack: None,
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"IndexExpression","array":{"type":"Identifier","name":"a"},"index":{"type":"IntegerLiteral","value":"3"}}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Function(Box::new(FunctionExpr {
        base: BaseNode::default(),
        lparen: None,
        params: vec![Property {
            base: BaseNode::default(),
            key: PropertyKey::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            separator: None,
            value: None,
            comma: None,
        }],
        rparen: None,
        arrow: None,
        body: FunctionBody::Expr(Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "hello".to_string(),
        })),
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"FunctionExpression","params":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}],"body":{"type":"StringLiteral","value":"hello"}}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Binary(Box::new(BinaryExpr {
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
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"BinaryExpression","operator":"+","left":{"type":"StringLiteral","value":"hello"},"right":{"type":"StringLiteral","value":"world"}}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Unary(Box::new(UnaryExpr {
        base: BaseNode::default(),
        operator: Operator::NotOperator,
        argument: Expression::Boolean(BooleanLit {
            base: BaseNode::default(),
            value: true,
        }),
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"UnaryExpression","operator":"not","argument":{"type":"BooleanLiteral","value":true}}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Logical(Box::new(LogicalExpr {
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
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"LogicalExpression","operator":"or","left":{"type":"BooleanLiteral","value":false},"right":{"type":"BooleanLiteral","value":true}}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Array(Box::new(ArrayExpr {
        base: BaseNode::default(),
        lbrack: None,
        elements: vec![ArrayItem {
            expression: Expression::StringLit(StringLit {
                base: BaseNode::default(),
                value: "hello".to_string(),
            }),
            comma: None,
        }],
        rbrack: None,
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ArrayExpression","elements":[{"type":"StringLiteral","value":"hello"}]}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Object(Box::new(ObjectExpr {
        base: BaseNode::default(),
        lbrace: None,
        with: None,
        properties: vec![Property {
            base: BaseNode::default(),
            key: PropertyKey::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            separator: None,
            value: Some(Expression::StringLit(StringLit {
                base: BaseNode::default(),
                value: "hello".to_string(),
            })),
            comma: None,
        }],
        rbrace: None,
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":{"type":"StringLiteral","value":"hello"}}]}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Object(Box::new(ObjectExpr {
        base: BaseNode::default(),
        lbrace: None,
        with: None,
        properties: vec![Property {
            base: BaseNode::default(),
            key: PropertyKey::StringLit(StringLit {
                base: BaseNode::default(),
                value: "a".to_string(),
            }),
            separator: None,
            value: Some(Expression::StringLit(StringLit {
                base: BaseNode::default(),
                value: "hello".to_string(),
            })),
            comma: None,
        }],
        rbrace: None,
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"StringLiteral","value":"a"},"value":{"type":"StringLiteral","value":"hello"}}]}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Object(Box::new(ObjectExpr {
        base: BaseNode::default(),
        lbrace: None,
        with: None,
        properties: vec![Property {
            base: BaseNode::default(),
            key: PropertyKey::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            separator: None,
            value: None,
            comma: None,
        }],
        rbrace: None,
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ObjectExpression","properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}]}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}

#[test]
fn test_json_object_expression_implicit_keys_and_with() {
    let n = Expression::Object(Box::new(ObjectExpr {
        base: BaseNode::default(),
        lbrace: None,
        with: Some(WithSource {
            source: Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            },
            with: None,
        }),
        properties: vec![Property {
            base: BaseNode::default(),
            key: PropertyKey::Identifier(Identifier {
                base: BaseNode::default(),
                name: "a".to_string(),
            }),
            separator: None,
            value: None,
            comma: None,
        }],
        rbrace: None,
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ObjectExpression","with":{"name":"a"},"properties":[{"type":"Property","key":{"type":"Identifier","name":"a"},"value":null}]}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Conditional(Box::new(ConditionalExpr {
        base: BaseNode::default(),
        tk_if: None,
        test: Expression::Boolean(BooleanLit {
            base: BaseNode::default(),
            value: true,
        }),
        tk_then: None,
        alternate: Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "false".to_string(),
        }),
        tk_else: None,
        consequent: Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "true".to_string(),
        }),
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ConditionalExpression","test":{"type":"BooleanLiteral","value":true},"consequent":{"type":"StringLiteral","value":"true"},"alternate":{"type":"StringLiteral","value":"false"}}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
        separator: None,
        value: Some(Expression::StringLit(StringLit {
            base: BaseNode::default(),
            value: "hello".to_string(),
        })),
        comma: None,
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
    let n = Expression::Identifier(Identifier {
        base: BaseNode::default(),
        name: "a".to_string(),
    });
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"Identifier","name":"a"}"#);
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::StringLit(StringLit {
        base: BaseNode::default(),
        value: "hello".to_string(),
    });
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"StringLiteral","value":"hello"}"#);
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Boolean(BooleanLit {
        base: BaseNode::default(),
        value: true,
    });
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"BooleanLiteral","value":true}"#);
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Float(FloatLit {
        base: BaseNode::default(),
        value: 42.1,
    });
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"FloatLiteral","value":42.1}"#);
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Integer(IntegerLit {
        base: BaseNode::default(),
        value: 9223372036854775807,
    });
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"IntegerLiteral","value":"9223372036854775807"}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Uint(UintLit {
        base: BaseNode::default(),
        value: 18446744073709551615,
    });
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"UnsignedIntegerLiteral","value":"18446744073709551615"}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Regexp(RegexpLit {
        base: BaseNode::default(),
        value: ".*".to_string(),
    });
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(serialized, r#"{"type":"RegexpLiteral","value":".*"}"#);
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::Duration(DurationLit {
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
    });
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"DurationLiteral","values":[{"magnitude":1,"unit":"h"},{"magnitude":1,"unit":"h"}]}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
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
    let n = Expression::DateTime(DateTimeLit {
        base: BaseNode::default(),
        value: FixedOffset::east(0)
            .ymd(2017, 8, 8)
            .and_hms_nano(8, 8, 8, 8),
    });
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"DateTimeLiteral","value":"2017-08-08T08:08:08.000000008+00:00"}"#
    );
    let deserialized: Expression = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}

#[test]
fn test_object_expression_with_source_locations_and_errors() {
    let n = Expression::Object(Box::new(ObjectExpr {
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
            ..BaseNode::default()
        },
        lbrace: None,
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
                ..BaseNode::default()
            },
            key: PropertyKey::Identifier(Identifier {
                base: BaseNode {
                    location: SourceLocation {
                        file: Some("foo.flux".to_string()),
                        start: Position { line: 1, column: 2 },
                        end: Position { line: 1, column: 3 },
                        source: Some("a".to_string()),
                    },
                    ..BaseNode::default()
                },
                name: "a".to_string(),
            }),
            separator: None,
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
                    ..BaseNode::default()
                },
                value: "hello".to_string(),
            })),
            comma: None,
        }],
        rbrace: None,
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"ObjectExpression","location":{"file":"foo.flux","start":{"line":1,"column":1},"end":{"line":1,"column":13},"source":"{a: \"hello\"}"},"properties":[{"type":"Property","location":{"file":"foo.flux","start":{"line":1,"column":2},"end":{"line":1,"column":12},"source":"a: \"hello\""},"errors":[{"msg":"an error"}],"key":{"type":"Identifier","location":{"file":"foo.flux","start":{"line":1,"column":2},"end":{"line":1,"column":3},"source":"a"},"name":"a"},"value":{"type":"StringLiteral","location":{"file":"foo.flux","start":{"line":1,"column":5},"end":{"line":1,"column":12},"source":"\"hello\""},"errors":[{"msg":"an error"},{"msg":"another error"}],"value":"hello"}}]}"#
    );
    // TODO(affo): leaving proper error deserialization for the future.
    // let deserialized: ObjectExpr = serde_json::from_str(serialized.as_str()).unwrap();
    // assert_eq!(deserialized, n)
}

#[test]
fn test_json_bad_statement() {
    let n = Statement::Bad(Box::new(BadStmt {
        base: BaseNode::default(),
        text: String::from("this is bad"),
    }));
    let serialized = serde_json::to_string(&n).unwrap();
    assert_eq!(
        serialized,
        r#"{"type":"BadStatement","text":"this is bad"}"#
    );
    let deserialized: Statement = serde_json::from_str(serialized.as_str()).unwrap();
    assert_eq!(deserialized, n)
}

#[test]
fn test_ast_json_roundtrip() {
    let ast = ast_with_every_kind_of_node();
    let serialized = match serde_json::to_string(&ast) {
        Ok(str) => str,
        Err(e) => panic!(format!("error serializing JSON: {}", e)),
    };
    let roundtrip_ast: Package = match serde_json::from_str(serialized.as_str()) {
        Ok(ast) => {
            println!("successfully deserialized AST");
            ast
        }
        Err(e) => panic!(format!("error deserializing JSON: {}", e)),
    };
    assert_eq!(ast, roundtrip_ast);
}
