use crate::ast;
use crate::semantic::fresh::Fresher;
use crate::semantic::nodes::*;
use crate::semantic::types::MonoType;
use std::result;

pub type SemanticError = String;
pub type Result<T> = result::Result<T, SemanticError>;

// analyze analyzes an AST package node and returns its semantic analysis.
// The function explicitly moves the ast::Package because it adds information to it.
// We follow here the principle that every compilation step should be isolated and should add meaning
// to the previous one. In other terms, once one analyzes an AST he should not use it anymore.
// If one wants to do so, he should explicitly pkg.clone() and incur consciously in the memory
// overhead involved.
pub fn analyze(pkg: ast::Package) -> Result<Package> {
    analyze_with(pkg, &mut Fresher::new())
}

// analyze_with runs analyze using the provided Fresher.
pub fn analyze_with(pkg: ast::Package, fresher: &mut Fresher) -> Result<Package> {
    analyze_package(pkg, fresher)
    // TODO(affo): run checks on the semantic graph.
}

fn analyze_package(pkg: ast::Package, fresher: &mut Fresher) -> Result<Package> {
    let files = pkg
        .files
        .into_iter()
        .map(|f| analyze_file(f, fresher))
        .collect::<Result<Vec<File>>>()?;
    Ok(Package {
        loc: pkg.base.location,
        package: pkg.package,
        files: files,
    })
}

fn analyze_file(file: ast::File, fresher: &mut Fresher) -> Result<File> {
    let package = analyze_package_clause(file.package, fresher)?;
    let imports = file
        .imports
        .into_iter()
        .map(|i| analyze_import_declaration(i, fresher))
        .collect::<Result<Vec<ImportDeclaration>>>()?;
    let body = file
        .body
        .into_iter()
        .map(|s| analyze_statement(s, fresher))
        .collect::<Result<Vec<Statement>>>()?;
    Ok(File {
        loc: file.base.location,
        package,
        imports,
        body,
    })
}

fn analyze_package_clause(
    pkg: Option<ast::PackageClause>,
    fresher: &mut Fresher,
) -> Result<Option<PackageClause>> {
    if pkg.is_none() {
        return Ok(None);
    }
    let pkg = pkg.unwrap();
    let name = analyze_identifier(pkg.name, fresher)?;
    Ok(Some(PackageClause {
        loc: pkg.base.location,
        name,
    }))
}

fn analyze_import_declaration(
    imp: ast::ImportDeclaration,
    fresher: &mut Fresher,
) -> Result<ImportDeclaration> {
    let alias = match imp.alias {
        None => None,
        Some(id) => Some(analyze_identifier(id, fresher)?),
    };
    let path = analyze_string_literal(imp.path, fresher)?;
    Ok(ImportDeclaration {
        loc: imp.base.location,
        alias,
        path,
    })
}

fn analyze_statement(stmt: ast::Statement, fresher: &mut Fresher) -> Result<Statement> {
    match stmt {
        ast::Statement::Option(s) => Ok(Statement::Option(analyze_option_statement(s, fresher)?)),
        ast::Statement::Builtin(s) => {
            Ok(Statement::Builtin(analyze_builtin_statement(s, fresher)?))
        }
        ast::Statement::Test(s) => Ok(Statement::Test(analyze_test_statement(s, fresher)?)),
        ast::Statement::Expr(s) => Ok(Statement::Expr(analyze_expression_statement(s, fresher)?)),
        ast::Statement::Return(s) => Ok(Statement::Return(analyze_return_statement(s, fresher)?)),
        // TODO(affo): we should fix this to include MemberAssignement.
        //  The error lies in AST: the Statement enum does not include that.
        //  This is not a problem when parsing, because we parse it only in the option assignment case,
        //  and we return an OptionStmt, which is a Statement.
        ast::Statement::Variable(s) => Ok(Statement::Variable(analyze_variable_assignment(
            s, fresher,
        )?)),
        ast::Statement::Bad(_) => {
            Err("BadStatement is not supported in semantic analysis".to_string())
        }
    }
}

fn analyze_assignment(assign: ast::Assignment, fresher: &mut Fresher) -> Result<Assignment> {
    match assign {
        ast::Assignment::Variable(a) => Ok(Assignment::Variable(analyze_variable_assignment(
            a, fresher,
        )?)),
        ast::Assignment::Member(a) => {
            Ok(Assignment::Member(analyze_member_assignment(a, fresher)?))
        }
    }
}

fn analyze_option_statement(stmt: ast::OptionStmt, fresher: &mut Fresher) -> Result<OptionStmt> {
    Ok(OptionStmt {
        loc: stmt.base.location,
        assignment: analyze_assignment(stmt.assignment, fresher)?,
    })
}

fn analyze_builtin_statement(stmt: ast::BuiltinStmt, fresher: &mut Fresher) -> Result<BuiltinStmt> {
    Ok(BuiltinStmt {
        loc: stmt.base.location,
        id: analyze_identifier(stmt.id, fresher)?,
    })
}

fn analyze_test_statement(stmt: ast::TestStmt, fresher: &mut Fresher) -> Result<TestStmt> {
    Ok(TestStmt {
        loc: stmt.base.location,
        assignment: analyze_variable_assignment(stmt.assignment, fresher)?,
    })
}

fn analyze_expression_statement(stmt: ast::ExprStmt, fresher: &mut Fresher) -> Result<ExprStmt> {
    Ok(ExprStmt {
        loc: stmt.base.location,
        expression: analyze_expression(stmt.expression, fresher)?,
    })
}

fn analyze_return_statement(stmt: ast::ReturnStmt, fresher: &mut Fresher) -> Result<ReturnStmt> {
    Ok(ReturnStmt {
        loc: stmt.base.location,
        argument: analyze_expression(stmt.argument, fresher)?,
    })
}

fn analyze_variable_assignment(
    stmt: ast::VariableAssgn,
    fresher: &mut Fresher,
) -> Result<VariableAssgn> {
    Ok(VariableAssgn::new(
        analyze_identifier(stmt.id, fresher)?,
        analyze_expression(stmt.init, fresher)?,
        stmt.base.location,
    ))
}

fn analyze_member_assignment(stmt: ast::MemberAssgn, fresher: &mut Fresher) -> Result<MemberAssgn> {
    Ok(MemberAssgn {
        loc: stmt.base.location,
        member: analyze_member_expression(stmt.member, fresher)?,
        init: analyze_expression(stmt.init, fresher)?,
    })
}

fn analyze_expression(expr: ast::Expression, fresher: &mut Fresher) -> Result<Expression> {
    match expr {
        ast::Expression::Function(expr) => Ok(Expression::Function(Box::new(analyze_function_expression(*expr, fresher)?))),
        ast::Expression::Call(expr) => Ok(Expression::Call(Box::new(analyze_call_expression(*expr, fresher)?))),
        ast::Expression::Member(expr) => Ok(Expression::Member(Box::new(analyze_member_expression(*expr, fresher)?))),
        ast::Expression::Index(expr) => Ok(Expression::Index(Box::new(analyze_index_expression(*expr, fresher)?))),
        ast::Expression::PipeExpr(expr) => Ok(Expression::Call(Box::new(analyze_pipe_expression(*expr, fresher)?))),
        ast::Expression::Binary(expr) => Ok(Expression::Binary(Box::new(analyze_binary_expression(*expr, fresher)?))),
        ast::Expression::Unary(expr) => Ok(Expression::Unary(Box::new(analyze_unary_expression(*expr, fresher)?))),
        ast::Expression::Logical(expr) => Ok(Expression::Logical(Box::new(analyze_logical_expression(*expr, fresher)?))),
        ast::Expression::Conditional(expr) => Ok(Expression::Conditional(Box::new(analyze_conditional_expression(*expr, fresher)?))),
        ast::Expression::Object(expr) => Ok(Expression::Object(Box::new(analyze_object_expression(*expr, fresher)?))),
        ast::Expression::Array(expr) => Ok(Expression::Array(Box::new(analyze_array_expression(*expr, fresher)?))),
        ast::Expression::Identifier(expr) => Ok(Expression::Identifier(analyze_identifier_expression(expr, fresher)?)),
        ast::Expression::StringExpr(expr) => Ok(Expression::StringExpr(Box::new(analyze_string_expression(*expr, fresher)?))),
        ast::Expression::Paren(expr) => analyze_expression(expr.expression, fresher),
        ast::Expression::StringLit(lit) => Ok(Expression::StringLit(analyze_string_literal(lit, fresher)?)),
        ast::Expression::Boolean(lit) => Ok(Expression::Boolean(analyze_boolean_literal(lit, fresher)?)),
        ast::Expression::Float(lit) => Ok(Expression::Float(analyze_float_literal(lit, fresher)?)),
        ast::Expression::Integer(lit) => Ok(Expression::Integer(analyze_integer_literal(lit, fresher)?)),
        ast::Expression::Uint(lit) => Ok(Expression::Uint(analyze_unsigned_integer_literal(lit, fresher)?)),
        ast::Expression::Regexp(lit) => Ok(Expression::Regexp(analyze_regexp_literal(lit, fresher)?)),
        ast::Expression::Duration(lit) => Ok(Expression::Duration(analyze_duration_literal(lit, fresher)?)),
        ast::Expression::DateTime(lit) => Ok(Expression::DateTime(analyze_date_time_literal(lit, fresher)?)),
        ast::Expression::PipeLit(_) => Err("a pipe literal may only be used as a default value for an argument in a function definition".to_string()),
        ast::Expression::Bad(_) => Err("BadExpression is not supported in semantic analysis".to_string())
    }
}

fn analyze_function_expression(
    expr: ast::FunctionExpr,
    fresher: &mut Fresher,
) -> Result<FunctionExpr> {
    let params = analyze_function_params(expr.params, fresher)?;
    let body = analyze_function_body(expr.body, fresher)?;
    Ok(FunctionExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        params,
        body,
    })
}

fn analyze_function_params(
    props: Vec<ast::Property>,
    fresher: &mut Fresher,
) -> Result<Vec<FunctionParameter>> {
    // The iteration here is complex, cannot use iter().map()..., better to write it explicitly.
    let mut params: Vec<FunctionParameter> = Vec::new();
    let mut piped = false;
    for prop in props {
        let id = match prop.key {
            ast::PropertyKey::Identifier(id) => Ok(id),
            _ => Err("function params must be identifiers".to_string()),
        }?;
        let key = analyze_identifier(id, fresher)?;
        let mut default: Option<Expression> = None;
        let mut is_pipe = false;
        match prop.value {
            Some(expr) => match expr {
                ast::Expression::PipeLit(_) => {
                    if piped {
                        return Err("only a single argument may be piped".to_string());
                    } else {
                        piped = true;
                        is_pipe = true;
                    };
                }
                e => default = Some(analyze_expression(e, fresher)?),
            },
            None => (),
        };
        params.push(FunctionParameter {
            loc: prop.base.location,
            is_pipe,
            key,
            default,
        });
    }
    Ok(params)
}

fn analyze_function_body(body: ast::FunctionBody, fresher: &mut Fresher) -> Result<Block> {
    match body {
        ast::FunctionBody::Expr(e) => Ok(Block::Return(analyze_expression(e, fresher)?)),
        ast::FunctionBody::Block(block) => Ok(analyze_block(block, fresher)?),
    }
}

fn analyze_block(block: ast::Block, fresher: &mut Fresher) -> Result<Block> {
    let mut body = block.body.into_iter().rev();

    let block = if let Some(ast::Statement::Return(stmt)) = body.next() {
        Block::Return(analyze_expression(stmt.argument, fresher)?)
    } else {
        return Err("missing return statement in block".to_string());
    };

    body.try_fold(block, |acc, s| match s {
        ast::Statement::Variable(dec) => Ok(Block::Variable(
            analyze_variable_assignment(dec, fresher)?,
            Box::new(acc),
        )),
        ast::Statement::Expr(stmt) => Ok(Block::Expr(
            analyze_expression_statement(stmt, fresher)?,
            Box::new(acc),
        )),
        _ => Err(format!("invalid statement in function block {:#?}", s)),
    })
}

fn analyze_call_expression(expr: ast::CallExpr, fresher: &mut Fresher) -> Result<CallExpr> {
    let callee = analyze_expression(expr.callee, fresher)?;
    // TODO(affo): I'd prefer these checks to be in ast.Check().
    if expr.arguments.len() > 1 {
        return Err("arguments are more than one object expression".to_string());
    }
    let mut args = expr
        .arguments
        .into_iter()
        .map(|a| match a {
            ast::Expression::Object(obj) => analyze_object_expression(*obj, fresher),
            _ => Err("arguments not an object expression".to_string()),
        })
        .collect::<Result<Vec<ObjectExpr>>>()?;
    let arguments = match args.len() {
        0 => Ok(Vec::new()),
        1 => Ok(args.pop().expect("there must be 1 element").properties),
        _ => Err("arguments are more than one object expression".to_string()),
    }?;
    Ok(CallExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        callee,
        arguments,
        pipe: None,
    })
}

fn analyze_member_expression(expr: ast::MemberExpr, fresher: &mut Fresher) -> Result<MemberExpr> {
    let object = analyze_expression(expr.object, fresher)?;
    let property = match expr.property {
        ast::PropertyKey::Identifier(id) => id.name,
        ast::PropertyKey::StringLit(lit) => lit.value,
    };
    Ok(MemberExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        object,
        property,
    })
}

fn analyze_index_expression(expr: ast::IndexExpr, fresher: &mut Fresher) -> Result<IndexExpr> {
    let array = analyze_expression(expr.array, fresher)?;
    let index = analyze_expression(expr.index, fresher)?;
    Ok(IndexExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        array,
        index,
    })
}

fn analyze_pipe_expression(expr: ast::PipeExpr, fresher: &mut Fresher) -> Result<CallExpr> {
    let mut call = analyze_call_expression(expr.call, fresher)?;
    let pipe = analyze_expression(expr.argument, fresher)?;
    call.pipe = Some(pipe);
    Ok(call)
}

fn analyze_binary_expression(expr: ast::BinaryExpr, fresher: &mut Fresher) -> Result<BinaryExpr> {
    let left = analyze_expression(expr.left, fresher)?;
    let right = analyze_expression(expr.right, fresher)?;
    Ok(BinaryExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        operator: expr.operator,
        left,
        right,
    })
}

fn analyze_unary_expression(expr: ast::UnaryExpr, fresher: &mut Fresher) -> Result<UnaryExpr> {
    let argument = analyze_expression(expr.argument, fresher)?;
    Ok(UnaryExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        operator: expr.operator,
        argument,
    })
}

fn analyze_logical_expression(
    expr: ast::LogicalExpr,
    fresher: &mut Fresher,
) -> Result<LogicalExpr> {
    let left = analyze_expression(expr.left, fresher)?;
    let right = analyze_expression(expr.right, fresher)?;
    Ok(LogicalExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        operator: expr.operator,
        left,
        right,
    })
}

fn analyze_conditional_expression(
    expr: ast::ConditionalExpr,
    fresher: &mut Fresher,
) -> Result<ConditionalExpr> {
    let test = analyze_expression(expr.test, fresher)?;
    let consequent = analyze_expression(expr.consequent, fresher)?;
    let alternate = analyze_expression(expr.alternate, fresher)?;
    Ok(ConditionalExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        test,
        consequent,
        alternate,
    })
}

fn analyze_object_expression(expr: ast::ObjectExpr, fresher: &mut Fresher) -> Result<ObjectExpr> {
    let properties = expr
        .properties
        .into_iter()
        .map(|p| analyze_property(p, fresher))
        .collect::<Result<Vec<Property>>>()?;
    let with = match expr.with {
        Some(id) => Some(analyze_identifier_expression(id, fresher)?),
        None => None,
    };
    Ok(ObjectExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        with,
        properties,
    })
}

fn analyze_property(prop: ast::Property, fresher: &mut Fresher) -> Result<Property> {
    let key = match prop.key {
        ast::PropertyKey::Identifier(id) => analyze_identifier(id, fresher)?,
        ast::PropertyKey::StringLit(lit) => Identifier {
            loc: lit.base.location.clone(),
            name: analyze_string_literal(lit, fresher)?.value,
        },
    };
    let value = match prop.value {
        Some(expr) => analyze_expression(expr, fresher)?,
        None => Expression::Identifier(IdentifierExpr {
            loc: key.loc.clone(),
            typ: MonoType::Var(fresher.fresh()),
            name: key.name.clone(),
        }),
    };
    Ok(Property {
        loc: prop.base.location,
        key,
        value,
    })
}

fn analyze_array_expression(expr: ast::ArrayExpr, fresher: &mut Fresher) -> Result<ArrayExpr> {
    let elements = expr
        .elements
        .into_iter()
        .map(|e| analyze_expression(e, fresher))
        .collect::<Result<Vec<Expression>>>()?;
    Ok(ArrayExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        elements,
    })
}

fn analyze_identifier(id: ast::Identifier, _fresher: &mut Fresher) -> Result<Identifier> {
    Ok(Identifier {
        loc: id.base.location,
        name: id.name,
    })
}

fn analyze_identifier_expression(
    id: ast::Identifier,
    fresher: &mut Fresher,
) -> Result<IdentifierExpr> {
    Ok(IdentifierExpr {
        loc: id.base.location,
        typ: MonoType::Var(fresher.fresh()),
        name: id.name,
    })
}

fn analyze_string_expression(expr: ast::StringExpr, fresher: &mut Fresher) -> Result<StringExpr> {
    let parts = expr
        .parts
        .into_iter()
        .map(|p| analyze_string_expression_part(p, fresher))
        .collect::<Result<Vec<StringExprPart>>>()?;
    Ok(StringExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        parts,
    })
}

fn analyze_string_expression_part(
    expr: ast::StringExprPart,
    fresher: &mut Fresher,
) -> Result<StringExprPart> {
    match expr {
        ast::StringExprPart::Text(txt) => Ok(StringExprPart::Text(TextPart {
            loc: txt.base.location,
            value: txt.value,
        })),
        ast::StringExprPart::Interpolated(itp) => {
            Ok(StringExprPart::Interpolated(InterpolatedPart {
                loc: itp.base.location,
                expression: analyze_expression(itp.expression, fresher)?,
            }))
        }
    }
}

fn analyze_string_literal(lit: ast::StringLit, fresher: &mut Fresher) -> Result<StringLit> {
    Ok(StringLit {
        loc: lit.base.location,
        typ: MonoType::Var(fresher.fresh()),
        value: lit.value,
    })
}

fn analyze_boolean_literal(lit: ast::BooleanLit, fresher: &mut Fresher) -> Result<BooleanLit> {
    Ok(BooleanLit {
        loc: lit.base.location,
        typ: MonoType::Var(fresher.fresh()),
        value: lit.value,
    })
}

fn analyze_float_literal(lit: ast::FloatLit, fresher: &mut Fresher) -> Result<FloatLit> {
    Ok(FloatLit {
        loc: lit.base.location,
        typ: MonoType::Var(fresher.fresh()),
        value: lit.value,
    })
}

fn analyze_integer_literal(lit: ast::IntegerLit, fresher: &mut Fresher) -> Result<IntegerLit> {
    Ok(IntegerLit {
        loc: lit.base.location,
        typ: MonoType::Var(fresher.fresh()),
        value: lit.value,
    })
}

fn analyze_unsigned_integer_literal(lit: ast::UintLit, fresher: &mut Fresher) -> Result<UintLit> {
    Ok(UintLit {
        loc: lit.base.location,
        typ: MonoType::Var(fresher.fresh()),
        value: lit.value,
    })
}

fn analyze_regexp_literal(lit: ast::RegexpLit, fresher: &mut Fresher) -> Result<RegexpLit> {
    Ok(RegexpLit {
        loc: lit.base.location,
        typ: MonoType::Var(fresher.fresh()),
        value: lit.value,
    })
}

fn analyze_duration_literal(lit: ast::DurationLit, fresher: &mut Fresher) -> Result<DurationLit> {
    Ok(DurationLit {
        loc: lit.base.location,
        typ: MonoType::Var(fresher.fresh()),
        value: convert_duration(&lit.values)?,
    })
}

fn analyze_date_time_literal(lit: ast::DateTimeLit, fresher: &mut Fresher) -> Result<DateTimeLit> {
    Ok(DateTimeLit {
        loc: lit.base.location,
        typ: MonoType::Var(fresher.fresh()),
        value: lit.value,
    })
}

// In these tests we test the results of semantic analysis on some ASTs.
// NOTE: we do not care about locations.
// We create a default base node and clone it in various AST nodes.
#[cfg(test)]
mod tests {
    use super::*;
    use crate::semantic::types::{MonoType, Tvar};
    use pretty_assertions::assert_eq;

    // type_info() is used for the expected semantic graph.
    // The id for the Tvar does not matter, because that is not compared.
    fn type_info() -> MonoType {
        MonoType::Var(Tvar(0))
    }

    fn test_analyze(pkg: ast::Package) -> Result<Package> {
        analyze(pkg)
    }

    #[test]
    fn test_analyze_empty() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: Vec::new(),
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: Vec::new(),
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_package() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: Some(ast::PackageClause {
                    base: b.clone(),
                    name: ast::Identifier {
                        base: b.clone(),
                        name: "foo".to_string(),
                    },
                }),
                imports: Vec::new(),
                body: Vec::new(),
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: Some(PackageClause {
                    loc: b.location.clone(),
                    name: Identifier {
                        loc: b.location.clone(),
                        name: "foo".to_string(),
                    },
                }),
                imports: Vec::new(),
                body: Vec::new(),
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_imports() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: Some(ast::PackageClause {
                    base: b.clone(),
                    name: ast::Identifier {
                        base: b.clone(),
                        name: "foo".to_string(),
                    },
                }),
                imports: vec![
                    ast::ImportDeclaration {
                        base: b.clone(),
                        path: ast::StringLit {
                            base: b.clone(),
                            value: "path/foo".to_string(),
                        },
                        alias: None,
                    },
                    ast::ImportDeclaration {
                        base: b.clone(),
                        path: ast::StringLit {
                            base: b.clone(),
                            value: "path/bar".to_string(),
                        },
                        alias: Some(ast::Identifier {
                            base: b.clone(),
                            name: "b".to_string(),
                        }),
                    },
                ],
                body: Vec::new(),
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: Some(PackageClause {
                    loc: b.location.clone(),
                    name: Identifier {
                        loc: b.location.clone(),
                        name: "foo".to_string(),
                    },
                }),
                imports: vec![
                    ImportDeclaration {
                        loc: b.location.clone(),
                        path: StringLit {
                            loc: b.location.clone(),
                            typ: type_info(),
                            value: "path/foo".to_string(),
                        },
                        alias: None,
                    },
                    ImportDeclaration {
                        loc: b.location.clone(),
                        path: StringLit {
                            loc: b.location.clone(),
                            typ: type_info(),
                            value: "path/bar".to_string(),
                        },
                        alias: Some(Identifier {
                            loc: b.location.clone(),
                            name: "b".to_string(),
                        }),
                    },
                ],
                body: Vec::new(),
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_var_assignment() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![
                    ast::Statement::Variable(ast::VariableAssgn {
                        base: b.clone(),
                        id: ast::Identifier {
                            base: b.clone(),
                            name: "a".to_string(),
                        },
                        init: ast::Expression::Boolean(ast::BooleanLit {
                            base: b.clone(),
                            value: true,
                        }),
                    }),
                    ast::Statement::Expr(ast::ExprStmt {
                        base: b.clone(),
                        expression: ast::Expression::Identifier(ast::Identifier {
                            base: b.clone(),
                            name: "a".to_string(),
                        }),
                    }),
                ],
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![
                    Statement::Variable(VariableAssgn::new(
                        Identifier {
                            loc: b.location.clone(),
                            name: "a".to_string(),
                        },
                        Expression::Boolean(BooleanLit {
                            loc: b.location.clone(),
                            typ: type_info(),
                            value: true,
                        }),
                        b.location.clone(),
                    )),
                    Statement::Expr(ExprStmt {
                        loc: b.location.clone(),
                        expression: Expression::Identifier(IdentifierExpr {
                            loc: b.location.clone(),
                            typ: type_info(),
                            name: "a".to_string(),
                        }),
                    }),
                ],
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_object() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Object(Box::new(ast::ObjectExpr {
                        base: b.clone(),
                        with: None,
                        properties: vec![ast::Property {
                            base: b.clone(),
                            key: ast::PropertyKey::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "a".to_string(),
                            }),
                            value: Some(ast::Expression::Integer(ast::IntegerLit {
                                base: b.clone(),
                                value: 10,
                            })),
                        }],
                    })),
                })],
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![Statement::Expr(ExprStmt {
                    loc: b.location.clone(),
                    expression: Expression::Object(Box::new(ObjectExpr {
                        loc: b.location.clone(),
                        typ: type_info(),
                        with: None,
                        properties: vec![Property {
                            loc: b.location.clone(),
                            key: Identifier {
                                loc: b.location.clone(),
                                name: "a".to_string(),
                            },
                            value: Expression::Integer(IntegerLit {
                                loc: b.location.clone(),
                                typ: type_info(),
                                value: 10,
                            }),
                        }],
                    })),
                })],
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_object_with_string_key() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Object(Box::new(ast::ObjectExpr {
                        base: b.clone(),
                        with: None,
                        properties: vec![ast::Property {
                            base: b.clone(),
                            key: ast::PropertyKey::StringLit(ast::StringLit {
                                base: b.clone(),
                                value: "a".to_string(),
                            }),
                            value: Some(ast::Expression::Integer(ast::IntegerLit {
                                base: b.clone(),
                                value: 10,
                            })),
                        }],
                    })),
                })],
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![Statement::Expr(ExprStmt {
                    loc: b.location.clone(),
                    expression: Expression::Object(Box::new(ObjectExpr {
                        loc: b.location.clone(),
                        typ: type_info(),
                        with: None,
                        properties: vec![Property {
                            loc: b.location.clone(),
                            key: Identifier {
                                loc: b.location.clone(),
                                name: "a".to_string(),
                            },
                            value: Expression::Integer(IntegerLit {
                                loc: b.location.clone(),
                                typ: type_info(),
                                value: 10,
                            }),
                        }],
                    })),
                })],
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_object_with_mixed_keys() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Object(Box::new(ast::ObjectExpr {
                        base: b.clone(),
                        with: None,
                        properties: vec![
                            ast::Property {
                                base: b.clone(),
                                key: ast::PropertyKey::StringLit(ast::StringLit {
                                    base: b.clone(),
                                    value: "a".to_string(),
                                }),
                                value: Some(ast::Expression::Integer(ast::IntegerLit {
                                    base: b.clone(),
                                    value: 10,
                                })),
                            },
                            ast::Property {
                                base: b.clone(),
                                key: ast::PropertyKey::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "b".to_string(),
                                }),
                                value: Some(ast::Expression::Integer(ast::IntegerLit {
                                    base: b.clone(),
                                    value: 11,
                                })),
                            },
                        ],
                    })),
                })],
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![Statement::Expr(ExprStmt {
                    loc: b.location.clone(),
                    expression: Expression::Object(Box::new(ObjectExpr {
                        loc: b.location.clone(),
                        typ: type_info(),
                        with: None,
                        properties: vec![
                            Property {
                                loc: b.location.clone(),
                                key: Identifier {
                                    loc: b.location.clone(),
                                    name: "a".to_string(),
                                },
                                value: Expression::Integer(IntegerLit {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    value: 10,
                                }),
                            },
                            Property {
                                loc: b.location.clone(),
                                key: Identifier {
                                    loc: b.location.clone(),
                                    name: "b".to_string(),
                                },
                                value: Expression::Integer(IntegerLit {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    value: 11,
                                }),
                            },
                        ],
                    })),
                })],
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_object_with_implicit_keys() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Object(Box::new(ast::ObjectExpr {
                        base: b.clone(),
                        with: None,
                        properties: vec![
                            ast::Property {
                                base: b.clone(),
                                key: ast::PropertyKey::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "a".to_string(),
                                }),
                                value: None,
                            },
                            ast::Property {
                                base: b.clone(),
                                key: ast::PropertyKey::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "b".to_string(),
                                }),
                                value: None,
                            },
                        ],
                    })),
                })],
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![Statement::Expr(ExprStmt {
                    loc: b.location.clone(),
                    expression: Expression::Object(Box::new(ObjectExpr {
                        loc: b.location.clone(),
                        typ: type_info(),
                        with: None,
                        properties: vec![
                            Property {
                                loc: b.location.clone(),
                                key: Identifier {
                                    loc: b.location.clone(),
                                    name: "a".to_string(),
                                },
                                value: Expression::Identifier(IdentifierExpr {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    name: "a".to_string(),
                                }),
                            },
                            Property {
                                loc: b.location.clone(),
                                key: Identifier {
                                    loc: b.location.clone(),
                                    name: "b".to_string(),
                                },
                                value: Expression::Identifier(IdentifierExpr {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    name: "b".to_string(),
                                }),
                            },
                        ],
                    })),
                })],
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_options_declaration() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Option(ast::OptionStmt {
                    base: b.clone(),
                    assignment: ast::Assignment::Variable(ast::VariableAssgn {
                        base: b.clone(),
                        id: ast::Identifier {
                            base: b.clone(),
                            name: "task".to_string(),
                        },
                        init: ast::Expression::Object(Box::new(ast::ObjectExpr {
                            base: b.clone(),
                            with: None,
                            properties: vec![
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "name".to_string(),
                                    }),
                                    value: Some(ast::Expression::StringLit(ast::StringLit {
                                        base: b.clone(),
                                        value: "foo".to_string(),
                                    })),
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "every".to_string(),
                                    }),
                                    value: Some(ast::Expression::Duration(ast::DurationLit {
                                        base: b.clone(),
                                        values: vec![ast::Duration {
                                            magnitude: 1,
                                            unit: "h".to_string(),
                                        }],
                                    })),
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "delay".to_string(),
                                    }),
                                    value: Some(ast::Expression::Duration(ast::DurationLit {
                                        base: b.clone(),
                                        values: vec![ast::Duration {
                                            magnitude: 10,
                                            unit: "m".to_string(),
                                        }],
                                    })),
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "cron".to_string(),
                                    }),
                                    value: Some(ast::Expression::StringLit(ast::StringLit {
                                        base: b.clone(),
                                        value: "0 2 * * *".to_string(),
                                    })),
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "retry".to_string(),
                                    }),
                                    value: Some(ast::Expression::Integer(ast::IntegerLit {
                                        base: b.clone(),
                                        value: 5,
                                    })),
                                },
                            ],
                        })),
                    }),
                })],
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![Statement::Option(OptionStmt {
                    loc: b.location.clone(),
                    assignment: Assignment::Variable(VariableAssgn::new(
                        Identifier {
                            loc: b.location.clone(),
                            name: "task".to_string(),
                        },
                        Expression::Object(Box::new(ObjectExpr {
                            loc: b.location.clone(),
                            typ: type_info(),
                            with: None,
                            properties: vec![
                                Property {
                                    loc: b.location.clone(),
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "name".to_string(),
                                    },
                                    value: Expression::StringLit(StringLit {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        value: "foo".to_string(),
                                    }),
                                },
                                Property {
                                    loc: b.location.clone(),
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "every".to_string(),
                                    },
                                    value: Expression::Duration(DurationLit {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        value: chrono::Duration::hours(1),
                                    }),
                                },
                                Property {
                                    loc: b.location.clone(),
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "delay".to_string(),
                                    },
                                    value: Expression::Duration(DurationLit {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        value: chrono::Duration::minutes(10),
                                    }),
                                },
                                Property {
                                    loc: b.location.clone(),
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "cron".to_string(),
                                    },
                                    value: Expression::StringLit(StringLit {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        value: "0 2 * * *".to_string(),
                                    }),
                                },
                                Property {
                                    loc: b.location.clone(),
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "retry".to_string(),
                                    },
                                    value: Expression::Integer(IntegerLit {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        value: 5,
                                    }),
                                },
                            ],
                        })),
                        b.location.clone(),
                    )),
                })],
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_qualified_option_statement() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Option(ast::OptionStmt {
                    base: b.clone(),
                    assignment: ast::Assignment::Member(ast::MemberAssgn {
                        base: b.clone(),
                        member: ast::MemberExpr {
                            base: b.clone(),
                            object: ast::Expression::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "alert".to_string(),
                            }),
                            property: ast::PropertyKey::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "state".to_string(),
                            }),
                        },
                        init: ast::Expression::StringLit(ast::StringLit {
                            base: b.clone(),
                            value: "Warning".to_string(),
                        }),
                    }),
                })],
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![Statement::Option(OptionStmt {
                    loc: b.location.clone(),
                    assignment: Assignment::Member(MemberAssgn {
                        loc: b.location.clone(),
                        member: MemberExpr {
                            loc: b.location.clone(),
                            typ: type_info(),
                            object: Expression::Identifier(IdentifierExpr {
                                loc: b.location.clone(),
                                typ: type_info(),
                                name: "alert".to_string(),
                            }),
                            property: "state".to_string(),
                        },
                        init: Expression::StringLit(StringLit {
                            loc: b.location.clone(),
                            typ: type_info(),
                            value: "Warning".to_string(),
                        }),
                    }),
                })],
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_function() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![
                    ast::Statement::Variable(ast::VariableAssgn {
                        base: b.clone(),
                        id: ast::Identifier {
                            base: b.clone(),
                            name: "f".to_string(),
                        },
                        init: ast::Expression::Function(Box::new(ast::FunctionExpr {
                            base: b.clone(),
                            params: vec![
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "a".to_string(),
                                    }),
                                    value: None,
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "b".to_string(),
                                    }),
                                    value: None,
                                },
                            ],
                            body: ast::FunctionBody::Expr(ast::Expression::Binary(Box::new(
                                ast::BinaryExpr {
                                    base: b.clone(),
                                    operator: ast::Operator::AdditionOperator,
                                    left: ast::Expression::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "a".to_string(),
                                    }),
                                    right: ast::Expression::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "b".to_string(),
                                    }),
                                },
                            ))),
                        })),
                    }),
                    ast::Statement::Expr(ast::ExprStmt {
                        base: b.clone(),
                        expression: ast::Expression::Call(Box::new(ast::CallExpr {
                            base: b.clone(),
                            callee: ast::Expression::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "f".to_string(),
                            }),
                            arguments: vec![ast::Expression::Object(Box::new(ast::ObjectExpr {
                                base: b.clone(),
                                with: None,
                                properties: vec![
                                    ast::Property {
                                        base: b.clone(),
                                        key: ast::PropertyKey::Identifier(ast::Identifier {
                                            base: b.clone(),
                                            name: "a".to_string(),
                                        }),
                                        value: Some(ast::Expression::Integer(ast::IntegerLit {
                                            base: b.clone(),
                                            value: 2,
                                        })),
                                    },
                                    ast::Property {
                                        base: b.clone(),
                                        key: ast::PropertyKey::Identifier(ast::Identifier {
                                            base: b.clone(),
                                            name: "b".to_string(),
                                        }),
                                        value: Some(ast::Expression::Integer(ast::IntegerLit {
                                            base: b.clone(),
                                            value: 3,
                                        })),
                                    },
                                ],
                            }))],
                        })),
                    }),
                ],
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![
                    Statement::Variable(VariableAssgn::new(
                        Identifier {
                            loc: b.location.clone(),
                            name: "f".to_string(),
                        },
                        Expression::Function(Box::new(FunctionExpr {
                            loc: b.location.clone(),
                            typ: type_info(),
                            params: vec![
                                FunctionParameter {
                                    loc: b.location.clone(),
                                    is_pipe: false,
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "a".to_string(),
                                    },
                                    default: None,
                                },
                                FunctionParameter {
                                    loc: b.location.clone(),
                                    is_pipe: false,
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "b".to_string(),
                                    },
                                    default: None,
                                },
                            ],
                            body: Block::Return(Expression::Binary(Box::new(BinaryExpr {
                                loc: b.location.clone(),
                                typ: type_info(),
                                operator: ast::Operator::AdditionOperator,
                                left: Expression::Identifier(IdentifierExpr {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    name: "a".to_string(),
                                }),
                                right: Expression::Identifier(IdentifierExpr {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    name: "b".to_string(),
                                }),
                            }))),
                        })),
                        b.location.clone(),
                    )),
                    Statement::Expr(ExprStmt {
                        loc: b.location.clone(),
                        expression: Expression::Call(Box::new(CallExpr {
                            loc: b.location.clone(),
                            typ: type_info(),
                            pipe: None,
                            callee: Expression::Identifier(IdentifierExpr {
                                loc: b.location.clone(),
                                typ: type_info(),
                                name: "f".to_string(),
                            }),
                            arguments: vec![
                                Property {
                                    loc: b.location.clone(),
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "a".to_string(),
                                    },
                                    value: Expression::Integer(IntegerLit {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        value: 2,
                                    }),
                                },
                                Property {
                                    loc: b.location.clone(),
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "b".to_string(),
                                    },
                                    value: Expression::Integer(IntegerLit {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        value: 3,
                                    }),
                                },
                            ],
                        })),
                    }),
                ],
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_function_with_defaults() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![
                    ast::Statement::Variable(ast::VariableAssgn {
                        base: b.clone(),
                        id: ast::Identifier {
                            base: b.clone(),
                            name: "f".to_string(),
                        },
                        init: ast::Expression::Function(Box::new(ast::FunctionExpr {
                            base: b.clone(),
                            params: vec![
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "a".to_string(),
                                    }),
                                    value: Some(ast::Expression::Integer(ast::IntegerLit {
                                        base: b.clone(),
                                        value: 0,
                                    })),
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "b".to_string(),
                                    }),
                                    value: Some(ast::Expression::Integer(ast::IntegerLit {
                                        base: b.clone(),
                                        value: 0,
                                    })),
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "c".to_string(),
                                    }),
                                    value: None,
                                },
                            ],
                            body: ast::FunctionBody::Expr(ast::Expression::Binary(Box::new(
                                ast::BinaryExpr {
                                    base: b.clone(),
                                    operator: ast::Operator::AdditionOperator,
                                    left: ast::Expression::Binary(Box::new(ast::BinaryExpr {
                                        base: b.clone(),
                                        operator: ast::Operator::AdditionOperator,
                                        left: ast::Expression::Identifier(ast::Identifier {
                                            base: b.clone(),
                                            name: "a".to_string(),
                                        }),
                                        right: ast::Expression::Identifier(ast::Identifier {
                                            base: b.clone(),
                                            name: "b".to_string(),
                                        }),
                                    })),
                                    right: ast::Expression::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "c".to_string(),
                                    }),
                                },
                            ))),
                        })),
                    }),
                    ast::Statement::Expr(ast::ExprStmt {
                        base: b.clone(),
                        expression: ast::Expression::Call(Box::new(ast::CallExpr {
                            base: b.clone(),
                            callee: ast::Expression::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "f".to_string(),
                            }),
                            arguments: vec![ast::Expression::Object(Box::new(ast::ObjectExpr {
                                base: b.clone(),
                                with: None,
                                properties: vec![ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "c".to_string(),
                                    }),
                                    value: Some(ast::Expression::Integer(ast::IntegerLit {
                                        base: b.clone(),
                                        value: 42,
                                    })),
                                }],
                            }))],
                        })),
                    }),
                ],
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![
                    Statement::Variable(VariableAssgn::new(
                        Identifier {
                            loc: b.location.clone(),
                            name: "f".to_string(),
                        },
                        Expression::Function(Box::new(FunctionExpr {
                            loc: b.location.clone(),
                            typ: type_info(),
                            params: vec![
                                FunctionParameter {
                                    loc: b.location.clone(),
                                    is_pipe: false,
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "a".to_string(),
                                    },
                                    default: Some(Expression::Integer(IntegerLit {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        value: 0,
                                    })),
                                },
                                FunctionParameter {
                                    loc: b.location.clone(),
                                    is_pipe: false,
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "b".to_string(),
                                    },
                                    default: Some(Expression::Integer(IntegerLit {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        value: 0,
                                    })),
                                },
                                FunctionParameter {
                                    loc: b.location.clone(),
                                    is_pipe: false,
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "c".to_string(),
                                    },
                                    default: None,
                                },
                            ],
                            body: Block::Return(Expression::Binary(Box::new(BinaryExpr {
                                loc: b.location.clone(),
                                typ: type_info(),
                                operator: ast::Operator::AdditionOperator,
                                left: Expression::Binary(Box::new(BinaryExpr {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    operator: ast::Operator::AdditionOperator,
                                    left: Expression::Identifier(IdentifierExpr {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        name: "a".to_string(),
                                    }),
                                    right: Expression::Identifier(IdentifierExpr {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        name: "b".to_string(),
                                    }),
                                })),
                                right: Expression::Identifier(IdentifierExpr {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    name: "c".to_string(),
                                }),
                            }))),
                        })),
                        b.location.clone(),
                    )),
                    Statement::Expr(ExprStmt {
                        loc: b.location.clone(),
                        expression: Expression::Call(Box::new(CallExpr {
                            loc: b.location.clone(),
                            typ: type_info(),
                            pipe: None,
                            callee: Expression::Identifier(IdentifierExpr {
                                loc: b.location.clone(),
                                typ: type_info(),
                                name: "f".to_string(),
                            }),
                            arguments: vec![Property {
                                loc: b.location.clone(),
                                key: Identifier {
                                    loc: b.location.clone(),
                                    name: "c".to_string(),
                                },
                                value: Expression::Integer(IntegerLit {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    value: 42,
                                }),
                            }],
                        })),
                    }),
                ],
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_function_multiple_pipes() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Variable(ast::VariableAssgn {
                    base: b.clone(),
                    id: ast::Identifier {
                        base: b.clone(),
                        name: "f".to_string(),
                    },
                    init: ast::Expression::Function(Box::new(ast::FunctionExpr {
                        base: b.clone(),
                        params: vec![
                            ast::Property {
                                base: b.clone(),
                                key: ast::PropertyKey::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "a".to_string(),
                                }),
                                value: None,
                            },
                            ast::Property {
                                base: b.clone(),
                                key: ast::PropertyKey::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "piped1".to_string(),
                                }),
                                value: Some(ast::Expression::PipeLit(ast::PipeLit {
                                    base: b.clone(),
                                })),
                            },
                            ast::Property {
                                base: b.clone(),
                                key: ast::PropertyKey::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "piped2".to_string(),
                                }),
                                value: Some(ast::Expression::PipeLit(ast::PipeLit {
                                    base: b.clone(),
                                })),
                            },
                        ],
                        body: ast::FunctionBody::Expr(ast::Expression::Identifier(
                            ast::Identifier {
                                base: b.clone(),
                                name: "a".to_string(),
                            },
                        )),
                    })),
                })],
            }],
        };
        let got = test_analyze(pkg).err().unwrap().to_string();
        assert_eq!("only a single argument may be piped".to_string(), got);
    }

    #[test]
    fn test_analyze_call_multiple_object_arguments() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Call(Box::new(ast::CallExpr {
                        base: b.clone(),
                        callee: ast::Expression::Identifier(ast::Identifier {
                            base: b.clone(),
                            name: "f".to_string(),
                        }),
                        arguments: vec![
                            ast::Expression::Object(Box::new(ast::ObjectExpr {
                                base: b.clone(),
                                with: None,
                                properties: vec![ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "a".to_string(),
                                    }),
                                    value: Some(ast::Expression::Integer(ast::IntegerLit {
                                        base: b.clone(),
                                        value: 0,
                                    })),
                                }],
                            })),
                            ast::Expression::Object(Box::new(ast::ObjectExpr {
                                base: b.clone(),
                                with: None,
                                properties: vec![ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "b".to_string(),
                                    }),
                                    value: Some(ast::Expression::Integer(ast::IntegerLit {
                                        base: b.clone(),
                                        value: 1,
                                    })),
                                }],
                            })),
                        ],
                    })),
                })],
            }],
        };
        let got = test_analyze(pkg).err().unwrap().to_string();
        assert_eq!(
            "arguments are more than one object expression".to_string(),
            got
        );
    }

    #[test]
    fn test_analyze_pipe_expression() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![
                    ast::Statement::Variable(ast::VariableAssgn {
                        base: b.clone(),
                        id: ast::Identifier {
                            base: b.clone(),
                            name: "f".to_string(),
                        },
                        init: ast::Expression::Function(Box::new(ast::FunctionExpr {
                            base: b.clone(),
                            params: vec![
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "piped".to_string(),
                                    }),
                                    value: Some(ast::Expression::PipeLit(ast::PipeLit {
                                        base: b.clone(),
                                    })),
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "a".to_string(),
                                    }),
                                    value: None,
                                },
                            ],
                            body: ast::FunctionBody::Expr(ast::Expression::Binary(Box::new(
                                ast::BinaryExpr {
                                    base: b.clone(),
                                    operator: ast::Operator::AdditionOperator,
                                    left: ast::Expression::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "a".to_string(),
                                    }),
                                    right: ast::Expression::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "piped".to_string(),
                                    }),
                                },
                            ))),
                        })),
                    }),
                    ast::Statement::Expr(ast::ExprStmt {
                        base: b.clone(),
                        expression: ast::Expression::PipeExpr(Box::new(ast::PipeExpr {
                            base: b.clone(),
                            argument: ast::Expression::Integer(ast::IntegerLit {
                                base: b.clone(),
                                value: 3,
                            }),
                            call: ast::CallExpr {
                                base: b.clone(),
                                callee: ast::Expression::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "f".to_string(),
                                }),
                                arguments: vec![ast::Expression::Object(Box::new(
                                    ast::ObjectExpr {
                                        base: b.clone(),
                                        with: None,
                                        properties: vec![ast::Property {
                                            base: b.clone(),
                                            key: ast::PropertyKey::Identifier(ast::Identifier {
                                                base: b.clone(),
                                                name: "a".to_string(),
                                            }),
                                            value: Some(ast::Expression::Integer(
                                                ast::IntegerLit {
                                                    base: b.clone(),
                                                    value: 2,
                                                },
                                            )),
                                        }],
                                    },
                                ))],
                            },
                        })),
                    }),
                ],
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![
                    Statement::Variable(VariableAssgn::new(
                        Identifier {
                            loc: b.location.clone(),
                            name: "f".to_string(),
                        },
                        Expression::Function(Box::new(FunctionExpr {
                            loc: b.location.clone(),
                            typ: type_info(),
                            params: vec![
                                FunctionParameter {
                                    loc: b.location.clone(),
                                    is_pipe: true,
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "piped".to_string(),
                                    },
                                    default: None,
                                },
                                FunctionParameter {
                                    loc: b.location.clone(),
                                    is_pipe: false,
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: "a".to_string(),
                                    },
                                    default: None,
                                },
                            ],
                            body: Block::Return(Expression::Binary(Box::new(BinaryExpr {
                                loc: b.location.clone(),
                                typ: type_info(),
                                operator: ast::Operator::AdditionOperator,
                                left: Expression::Identifier(IdentifierExpr {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    name: "a".to_string(),
                                }),
                                right: Expression::Identifier(IdentifierExpr {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    name: "piped".to_string(),
                                }),
                            }))),
                        })),
                        b.location.clone(),
                    )),
                    Statement::Expr(ExprStmt {
                        loc: b.location.clone(),
                        expression: Expression::Call(Box::new(CallExpr {
                            loc: b.location.clone(),
                            typ: type_info(),
                            pipe: Some(Expression::Integer(IntegerLit {
                                loc: b.location.clone(),
                                typ: type_info(),
                                value: 3,
                            })),
                            callee: Expression::Identifier(IdentifierExpr {
                                loc: b.location.clone(),
                                typ: type_info(),
                                name: "f".to_string(),
                            }),
                            arguments: vec![Property {
                                loc: b.location.clone(),
                                key: Identifier {
                                    loc: b.location.clone(),
                                    name: "a".to_string(),
                                },
                                value: Expression::Integer(IntegerLit {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    value: 2,
                                }),
                            }],
                        })),
                    }),
                ],
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_function_expression_simple() {
        let b = ast::BaseNode::default();
        let f = FunctionExpr {
            loc: b.location.clone(),
            typ: type_info(),
            params: vec![
                FunctionParameter {
                    loc: b.location.clone(),
                    is_pipe: false,
                    key: Identifier {
                        loc: b.location.clone(),
                        name: "a".to_string(),
                    },
                    default: None,
                },
                FunctionParameter {
                    loc: b.location.clone(),
                    is_pipe: false,
                    key: Identifier {
                        loc: b.location.clone(),
                        name: "b".to_string(),
                    },
                    default: None,
                },
            ],
            body: Block::Return(Expression::Binary(Box::new(BinaryExpr {
                loc: b.location.clone(),
                typ: type_info(),
                operator: ast::Operator::AdditionOperator,
                left: Expression::Identifier(IdentifierExpr {
                    loc: b.location.clone(),
                    typ: type_info(),
                    name: "a".to_string(),
                }),
                right: Expression::Identifier(IdentifierExpr {
                    loc: b.location.clone(),
                    typ: type_info(),
                    name: "b".to_string(),
                }),
            }))),
        };
        assert_eq!(Vec::<&FunctionParameter>::new(), f.defaults());
        assert_eq!(None, f.pipe());
    }

    #[test]
    fn test_function_expression_defaults_and_pipes() {
        let b = ast::BaseNode::default();
        let piped = FunctionParameter {
            loc: b.location.clone(),
            is_pipe: true,
            key: Identifier {
                loc: b.location.clone(),
                name: "a".to_string(),
            },
            default: Some(Expression::Integer(IntegerLit {
                loc: b.location.clone(),
                typ: type_info(),
                value: 0,
            })),
        };
        let default1 = FunctionParameter {
            loc: b.location.clone(),
            is_pipe: false,
            key: Identifier {
                loc: b.location.clone(),
                name: "b".to_string(),
            },
            default: Some(Expression::Integer(IntegerLit {
                loc: b.location.clone(),
                typ: type_info(),
                value: 1,
            })),
        };
        let default2 = FunctionParameter {
            loc: b.location.clone(),
            is_pipe: false,
            key: Identifier {
                loc: b.location.clone(),
                name: "c".to_string(),
            },
            default: Some(Expression::Integer(IntegerLit {
                loc: b.location.clone(),
                typ: type_info(),
                value: 2,
            })),
        };
        let no_default = FunctionParameter {
            loc: b.location.clone(),
            is_pipe: false,
            key: Identifier {
                loc: b.location.clone(),
                name: "d".to_string(),
            },
            default: None,
        };
        let defaults = vec![&piped, &default1, &default2];
        let f = FunctionExpr {
            loc: b.location.clone(),
            typ: type_info(),
            params: vec![
                piped.clone(),
                default1.clone(),
                default2.clone(),
                no_default.clone(),
            ],
            body: Block::Return(Expression::Binary(Box::new(BinaryExpr {
                loc: b.location.clone(),
                typ: type_info(),
                operator: ast::Operator::AdditionOperator,
                left: Expression::Identifier(IdentifierExpr {
                    loc: b.location.clone(),
                    typ: type_info(),
                    name: "a".to_string(),
                }),
                right: Expression::Identifier(IdentifierExpr {
                    loc: b.location.clone(),
                    typ: type_info(),
                    name: "b".to_string(),
                }),
            }))),
        };
        assert_eq!(defaults, f.defaults());
        assert_eq!(Some(&piped), f.pipe());
    }

    #[test]
    fn test_analyze_index_expression() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Index(Box::new(ast::IndexExpr {
                        base: b.clone(),
                        array: ast::Expression::Identifier(ast::Identifier {
                            base: b.clone(),
                            name: "a".to_string(),
                        }),
                        index: ast::Expression::Integer(ast::IntegerLit {
                            base: b.clone(),
                            value: 3,
                        }),
                    })),
                })],
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![Statement::Expr(ExprStmt {
                    loc: b.location.clone(),
                    expression: Expression::Index(Box::new(IndexExpr {
                        loc: b.location.clone(),
                        typ: type_info(),
                        array: Expression::Identifier(IdentifierExpr {
                            loc: b.location.clone(),
                            typ: type_info(),
                            name: "a".to_string(),
                        }),
                        index: Expression::Integer(IntegerLit {
                            loc: b.location.clone(),
                            typ: type_info(),
                            value: 3,
                        }),
                    })),
                })],
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_nested_index_expression() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Index(Box::new(ast::IndexExpr {
                        base: b.clone(),
                        array: ast::Expression::Index(Box::new(ast::IndexExpr {
                            base: b.clone(),
                            array: ast::Expression::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "a".to_string(),
                            }),
                            index: ast::Expression::Integer(ast::IntegerLit {
                                base: b.clone(),
                                value: 3,
                            }),
                        })),
                        index: ast::Expression::Integer(ast::IntegerLit {
                            base: b.clone(),
                            value: 5,
                        }),
                    })),
                })],
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![Statement::Expr(ExprStmt {
                    loc: b.location.clone(),
                    expression: Expression::Index(Box::new(IndexExpr {
                        loc: b.location.clone(),
                        typ: type_info(),
                        array: Expression::Index(Box::new(IndexExpr {
                            loc: b.location.clone(),
                            typ: type_info(),
                            array: Expression::Identifier(IdentifierExpr {
                                loc: b.location.clone(),
                                typ: type_info(),
                                name: "a".to_string(),
                            }),
                            index: Expression::Integer(IntegerLit {
                                loc: b.location.clone(),
                                typ: type_info(),
                                value: 3,
                            }),
                        })),
                        index: Expression::Integer(IntegerLit {
                            loc: b.location.clone(),
                            typ: type_info(),
                            value: 5,
                        }),
                    })),
                })],
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_access_idexed_object_returned_from_function_call() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Index(Box::new(ast::IndexExpr {
                        base: b.clone(),
                        array: ast::Expression::Call(Box::new(ast::CallExpr {
                            base: b.clone(),
                            callee: ast::Expression::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "f".to_string(),
                            }),
                            arguments: vec![],
                        })),
                        index: ast::Expression::Integer(ast::IntegerLit {
                            base: b.clone(),
                            value: 3,
                        }),
                    })),
                })],
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![Statement::Expr(ExprStmt {
                    loc: b.location.clone(),
                    expression: Expression::Index(Box::new(IndexExpr {
                        loc: b.location.clone(),
                        typ: type_info(),
                        array: Expression::Call(Box::new(CallExpr {
                            loc: b.location.clone(),
                            typ: type_info(),
                            pipe: None,
                            callee: Expression::Identifier(IdentifierExpr {
                                loc: b.location.clone(),
                                typ: type_info(),
                                name: "f".to_string(),
                            }),
                            arguments: Vec::new(),
                        })),
                        index: Expression::Integer(IntegerLit {
                            loc: b.location.clone(),
                            typ: type_info(),
                            value: 3,
                        }),
                    })),
                })],
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_nested_member_expression() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Member(Box::new(ast::MemberExpr {
                        base: b.clone(),
                        object: ast::Expression::Member(Box::new(ast::MemberExpr {
                            base: b.clone(),
                            object: ast::Expression::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "a".to_string(),
                            }),
                            property: ast::PropertyKey::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "b".to_string(),
                            }),
                        })),
                        property: ast::PropertyKey::Identifier(ast::Identifier {
                            base: b.clone(),
                            name: "c".to_string(),
                        }),
                    })),
                })],
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![Statement::Expr(ExprStmt {
                    loc: b.location.clone(),
                    expression: Expression::Member(Box::new(MemberExpr {
                        loc: b.location.clone(),
                        typ: type_info(),
                        object: Expression::Member(Box::new(MemberExpr {
                            loc: b.location.clone(),
                            typ: type_info(),
                            object: Expression::Identifier(IdentifierExpr {
                                loc: b.location.clone(),
                                typ: type_info(),
                                name: "a".to_string(),
                            }),
                            property: "b".to_string(),
                        })),
                        property: "c".to_string(),
                    })),
                })],
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_analyze_member_with_call_expression() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Member(Box::new(ast::MemberExpr {
                        base: b.clone(),
                        object: ast::Expression::Call(Box::new(ast::CallExpr {
                            base: b.clone(),
                            callee: ast::Expression::Member(Box::new(ast::MemberExpr {
                                base: b.clone(),
                                object: ast::Expression::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "a".to_string(),
                                }),
                                property: ast::PropertyKey::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "b".to_string(),
                                }),
                            })),
                            arguments: vec![],
                        })),
                        property: ast::PropertyKey::Identifier(ast::Identifier {
                            base: b.clone(),
                            name: "c".to_string(),
                        }),
                    })),
                })],
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![Statement::Expr(ExprStmt {
                    loc: b.location.clone(),
                    expression: Expression::Member(Box::new(MemberExpr {
                        loc: b.location.clone(),
                        typ: type_info(),
                        object: Expression::Call(Box::new(CallExpr {
                            loc: b.location.clone(),
                            typ: type_info(),
                            pipe: None,
                            callee: Expression::Member(Box::new(MemberExpr {
                                loc: b.location.clone(),
                                typ: type_info(),
                                object: Expression::Identifier(IdentifierExpr {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    name: "a".to_string(),
                                }),
                                property: "b".to_string(),
                            })),
                            arguments: Vec::new(),
                        })),
                        property: "c".to_string(),
                    })),
                })],
            }],
        };
        let got = test_analyze(pkg).unwrap();
        assert_eq!(want, got);
    }
}
