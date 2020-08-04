use crate::ast;
use crate::semantic::fresh::Fresher;
use crate::semantic::nodes::*;
use crate::semantic::types;
use crate::semantic::types::MonoType;
use crate::semantic::types::MonoTypeMap;
use std::collections::HashMap;
use std::result;

pub type SemanticError = String;
pub type Result<T> = result::Result<T, SemanticError>;

/// convert_with converts an AST package node to its semantic representation using
/// the provided fresher.
///
/// Note: most external callers of this function will want to use the analyze()
/// function in the flux crate instead, which is aware of everything in the Flux stdlib and prelude.
///
/// The function explicitly moves the ast::Package because it adds information to it.
/// We follow here the principle that every compilation step should be isolated and should add meaning
/// to the previous one. In other terms, once one converts an AST he should not use it anymore.
/// If one wants to do so, he should explicitly pkg.clone() and incur consciously in the memory
/// overhead involved.
pub fn convert_with(pkg: ast::Package, fresher: &mut Fresher) -> Result<Package> {
    convert_package(pkg, fresher)
    // TODO(affo): run checks on the semantic graph.
}

fn convert_package(pkg: ast::Package, fresher: &mut Fresher) -> Result<Package> {
    let files = pkg
        .files
        .into_iter()
        .map(|f| convert_file(f, fresher))
        .collect::<Result<Vec<File>>>()?;
    Ok(Package {
        loc: pkg.base.location,
        package: pkg.package,
        files,
    })
}

pub fn convert_file(file: ast::File, fresher: &mut Fresher) -> Result<File> {
    let package = convert_package_clause(file.package, fresher)?;
    let imports = file
        .imports
        .into_iter()
        .map(|i| convert_import_declaration(i, fresher))
        .collect::<Result<Vec<ImportDeclaration>>>()?;
    let body = file
        .body
        .into_iter()
        .map(|s| convert_statement(s, fresher))
        .collect::<Result<Vec<Statement>>>()?;
    Ok(File {
        loc: file.base.location,
        package,
        imports,
        body,
    })
}

fn convert_package_clause(
    pkg: Option<ast::PackageClause>,
    fresher: &mut Fresher,
) -> Result<Option<PackageClause>> {
    if pkg.is_none() {
        return Ok(None);
    }
    let pkg = pkg.unwrap();
    let name = convert_identifier(pkg.name, fresher)?;
    Ok(Some(PackageClause {
        loc: pkg.base.location,
        name,
    }))
}

fn convert_import_declaration(
    imp: ast::ImportDeclaration,
    fresher: &mut Fresher,
) -> Result<ImportDeclaration> {
    let alias = match imp.alias {
        None => None,
        Some(id) => Some(convert_identifier(id, fresher)?),
    };
    let path = convert_string_literal(imp.path, fresher)?;
    Ok(ImportDeclaration {
        loc: imp.base.location,
        alias,
        path,
    })
}

fn convert_statement(stmt: ast::Statement, fresher: &mut Fresher) -> Result<Statement> {
    match stmt {
        ast::Statement::Option(s) => Ok(Statement::Option(Box::new(convert_option_statement(
            *s, fresher,
        )?))),
        ast::Statement::Builtin(s) => {
            Ok(Statement::Builtin(convert_builtin_statement(*s, fresher)?))
        }
        ast::Statement::Test(s) => Ok(Statement::Test(Box::new(convert_test_statement(
            *s, fresher,
        )?))),
        ast::Statement::Expr(s) => Ok(Statement::Expr(convert_expression_statement(*s, fresher)?)),
        ast::Statement::Return(s) => Ok(Statement::Return(convert_return_statement(*s, fresher)?)),
        // TODO(affo): we should fix this to include MemberAssignement.
        //  The error lies in AST: the Statement enum does not include that.
        //  This is not a problem when parsing, because we parse it only in the option assignment case,
        //  and we return an OptionStmt, which is a Statement.
        ast::Statement::Variable(s) => Ok(Statement::Variable(Box::new(
            convert_variable_assignment(*s, fresher)?,
        ))),
        ast::Statement::Bad(_) => {
            Err("BadStatement is not supported in semantic analysis".to_string())
        }
    }
}

fn convert_assignment(assign: ast::Assignment, fresher: &mut Fresher) -> Result<Assignment> {
    match assign {
        ast::Assignment::Variable(a) => Ok(Assignment::Variable(convert_variable_assignment(
            *a, fresher,
        )?)),
        ast::Assignment::Member(a) => {
            Ok(Assignment::Member(convert_member_assignment(*a, fresher)?))
        }
    }
}

fn convert_option_statement(stmt: ast::OptionStmt, fresher: &mut Fresher) -> Result<OptionStmt> {
    Ok(OptionStmt {
        loc: stmt.base.location,
        assignment: convert_assignment(stmt.assignment, fresher)?,
    })
}

fn convert_builtin_statement(stmt: ast::BuiltinStmt, fresher: &mut Fresher) -> Result<BuiltinStmt> {
    Ok(BuiltinStmt {
        loc: stmt.base.location,
        id: convert_identifier(stmt.id, fresher)?,
    })
}

#[allow(unused)]
fn convert_monotype(
    ty: ast::MonoType,
    tvars: &mut HashMap<String, types::Tvar>,
    f: &mut Fresher,
) -> Result<MonoType> {
    match ty {
        ast::MonoType::Tvar(tv) => {
            let tvar = tvars.entry(tv.name.name).or_insert_with(|| f.fresh());
            Ok(MonoType::Var(*tvar))
        }
        ast::MonoType::Basic(basic) => match basic.name.name.as_str() {
            "bool" => Ok(MonoType::Bool),
            "int" => Ok(MonoType::Int),
            "uint" => Ok(MonoType::Uint),
            "float" => Ok(MonoType::Float),
            "string" => Ok(MonoType::String),
            "duration" => Ok(MonoType::Duration),
            "time" => Ok(MonoType::Time),
            "regexp" => Ok(MonoType::Regexp),
            "bytes" => Ok(MonoType::Bytes),
            _ => Err("Bad parameter type.".to_string()),
        },
        ast::MonoType::Array(arr) => Ok(MonoType::Arr(Box::new(types::Array(convert_monotype(
            arr.element,
            tvars,
            f,
        )?)))),
        ast::MonoType::Function(func) => {
            let mut req = MonoTypeMap::new();
            let mut opt = MonoTypeMap::new();
            let mut _pipe = None;
            let mut dirty = false;
            for param in func.parameters {
                match param {
                    ast::ParameterType::Required { name, monotype, .. } => {
                        req.insert(name.name, convert_monotype(monotype, tvars, f)?);
                    }
                    ast::ParameterType::Optional { name, monotype, .. } => {
                        opt.insert(name.name, convert_monotype(monotype, tvars, f)?);
                    }
                    ast::ParameterType::Pipe { name, monotype, .. } => {
                        if !dirty {
                            _pipe = Some(types::Property {
                                k: match name {
                                    Some(n) => n.name,
                                    None => String::from("<-"),
                                },
                                v: convert_monotype(monotype, tvars, f)?,
                            });
                            dirty = true;
                        } else {
                            return Err("Bad parameter type.".to_string());
                        }
                    }
                }
            }
            Ok(MonoType::Fun(Box::new(types::Function {
                req,
                opt,
                pipe: _pipe,
                retn: convert_monotype(func.monotype, tvars, f)?,
            })))
        }
        ast::MonoType::Record(rec) => {
            let mut r = match rec.tvar {
                None => MonoType::Row(Box::new(types::Row::Empty)),
                Some(id) => {
                    let tv = ast::MonoType::Tvar(ast::TvarType {
                        base: id.clone().base,
                        name: id,
                    });
                    convert_monotype(tv, tvars, f)?
                }
            };
            for prop in rec.properties {
                let property = types::Property {
                    k: prop.name.name,
                    v: convert_monotype(prop.monotype, tvars, f)?,
                };
                r = MonoType::Row(Box::new(types::Row::Extension {
                    head: property,
                    tail: r,
                }))
            }
            Ok(r)
        }
    }
}

fn convert_test_statement(stmt: ast::TestStmt, fresher: &mut Fresher) -> Result<TestStmt> {
    Ok(TestStmt {
        loc: stmt.base.location,
        assignment: convert_variable_assignment(stmt.assignment, fresher)?,
    })
}

fn convert_expression_statement(stmt: ast::ExprStmt, fresher: &mut Fresher) -> Result<ExprStmt> {
    Ok(ExprStmt {
        loc: stmt.base.location,
        expression: convert_expression(stmt.expression, fresher)?,
    })
}

fn convert_return_statement(stmt: ast::ReturnStmt, fresher: &mut Fresher) -> Result<ReturnStmt> {
    Ok(ReturnStmt {
        loc: stmt.base.location,
        argument: convert_expression(stmt.argument, fresher)?,
    })
}

fn convert_variable_assignment(
    stmt: ast::VariableAssgn,
    fresher: &mut Fresher,
) -> Result<VariableAssgn> {
    Ok(VariableAssgn::new(
        convert_identifier(stmt.id, fresher)?,
        convert_expression(stmt.init, fresher)?,
        stmt.base.location,
    ))
}

fn convert_member_assignment(stmt: ast::MemberAssgn, fresher: &mut Fresher) -> Result<MemberAssgn> {
    Ok(MemberAssgn {
        loc: stmt.base.location,
        member: convert_member_expression(stmt.member, fresher)?,
        init: convert_expression(stmt.init, fresher)?,
    })
}

fn convert_expression(expr: ast::Expression, fresher: &mut Fresher) -> Result<Expression> {
    match expr {
        ast::Expression::Function(expr) => Ok(Expression::Function(Box::new(convert_function_expression(*expr, fresher)?))),
        ast::Expression::Call(expr) => Ok(Expression::Call(Box::new(convert_call_expression(*expr, fresher)?))),
        ast::Expression::Member(expr) => Ok(Expression::Member(Box::new(convert_member_expression(*expr, fresher)?))),
        ast::Expression::Index(expr) => Ok(Expression::Index(Box::new(convert_index_expression(*expr, fresher)?))),
        ast::Expression::PipeExpr(expr) => Ok(Expression::Call(Box::new(convert_pipe_expression(*expr, fresher)?))),
        ast::Expression::Binary(expr) => Ok(Expression::Binary(Box::new(convert_binary_expression(*expr, fresher)?))),
        ast::Expression::Unary(expr) => Ok(Expression::Unary(Box::new(convert_unary_expression(*expr, fresher)?))),
        ast::Expression::Logical(expr) => Ok(Expression::Logical(Box::new(convert_logical_expression(*expr, fresher)?))),
        ast::Expression::Conditional(expr) => Ok(Expression::Conditional(Box::new(convert_conditional_expression(*expr, fresher)?))),
        ast::Expression::Object(expr) => Ok(Expression::Object(Box::new(convert_object_expression(*expr, fresher)?))),
        ast::Expression::Array(expr) => Ok(Expression::Array(Box::new(convert_array_expression(*expr, fresher)?))),
        ast::Expression::Identifier(expr) => Ok(Expression::Identifier(convert_identifier_expression(expr, fresher)?)),
        ast::Expression::StringExpr(expr) => Ok(Expression::StringExpr(Box::new(convert_string_expression(*expr, fresher)?))),
        ast::Expression::Paren(expr) => convert_expression(expr.expression, fresher),
        ast::Expression::StringLit(lit) => Ok(Expression::StringLit(convert_string_literal(lit, fresher)?)),
        ast::Expression::Boolean(lit) => Ok(Expression::Boolean(convert_boolean_literal(lit, fresher)?)),
        ast::Expression::Float(lit) => Ok(Expression::Float(convert_float_literal(lit, fresher)?)),
        ast::Expression::Integer(lit) => Ok(Expression::Integer(convert_integer_literal(lit, fresher)?)),
        ast::Expression::Uint(lit) => Ok(Expression::Uint(convert_unsigned_integer_literal(lit, fresher)?)),
        ast::Expression::Regexp(lit) => Ok(Expression::Regexp(convert_regexp_literal(lit, fresher)?)),
        ast::Expression::Duration(lit) => Ok(Expression::Duration(convert_duration_literal(lit, fresher)?)),
        ast::Expression::DateTime(lit) => Ok(Expression::DateTime(convert_date_time_literal(lit, fresher)?)),
        ast::Expression::PipeLit(_) => Err("a pipe literal may only be used as a default value for an argument in a function definition".to_string()),
        ast::Expression::Bad(_) => Err("BadExpression is not supported in semantic analysis".to_string())
    }
}

fn convert_function_expression(
    expr: ast::FunctionExpr,
    fresher: &mut Fresher,
) -> Result<FunctionExpr> {
    let params = convert_function_params(expr.params, fresher)?;
    let body = convert_function_body(expr.body, fresher)?;
    Ok(FunctionExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        params,
        body,
    })
}

fn convert_function_params(
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
        let key = convert_identifier(id, fresher)?;
        let mut default: Option<Expression> = None;
        let mut is_pipe = false;
        if let Some(expr) = prop.value {
            match expr {
                ast::Expression::PipeLit(_) => {
                    if piped {
                        return Err("only a single argument may be piped".to_string());
                    } else {
                        piped = true;
                        is_pipe = true;
                    };
                }
                e => default = Some(convert_expression(e, fresher)?),
            }
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

fn convert_function_body(body: ast::FunctionBody, fresher: &mut Fresher) -> Result<Block> {
    match body {
        ast::FunctionBody::Expr(expr) => {
            let argument = convert_expression(expr, fresher)?;
            Ok(Block::Return(ReturnStmt {
                loc: argument.loc().clone(),
                argument,
            }))
        }
        ast::FunctionBody::Block(block) => Ok(convert_block(block, fresher)?),
    }
}

fn convert_block(block: ast::Block, fresher: &mut Fresher) -> Result<Block> {
    let mut body = block.body.into_iter().rev();

    let block = if let Some(ast::Statement::Return(stmt)) = body.next() {
        let argument = convert_expression(stmt.argument, fresher)?;
        Block::Return(ReturnStmt {
            loc: stmt.base.location.clone(),
            argument,
        })
    } else {
        return Err("missing return statement in block".to_string());
    };

    body.try_fold(block, |acc, s| match s {
        ast::Statement::Variable(dec) => Ok(Block::Variable(
            Box::new(convert_variable_assignment(*dec, fresher)?),
            Box::new(acc),
        )),
        ast::Statement::Expr(stmt) => Ok(Block::Expr(
            convert_expression_statement(*stmt, fresher)?,
            Box::new(acc),
        )),
        _ => Err(format!("invalid statement in function block {:#?}", s)),
    })
}

fn convert_call_expression(expr: ast::CallExpr, fresher: &mut Fresher) -> Result<CallExpr> {
    let callee = convert_expression(expr.callee, fresher)?;
    // TODO(affo): I'd prefer these checks to be in ast.Check().
    if expr.arguments.len() > 1 {
        return Err("arguments are more than one object expression".to_string());
    }
    let mut args = expr
        .arguments
        .into_iter()
        .map(|a| match a {
            ast::Expression::Object(obj) => convert_object_expression(*obj, fresher),
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

fn convert_member_expression(expr: ast::MemberExpr, fresher: &mut Fresher) -> Result<MemberExpr> {
    let object = convert_expression(expr.object, fresher)?;
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

fn convert_index_expression(expr: ast::IndexExpr, fresher: &mut Fresher) -> Result<IndexExpr> {
    let array = convert_expression(expr.array, fresher)?;
    let index = convert_expression(expr.index, fresher)?;
    Ok(IndexExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        array,
        index,
    })
}

fn convert_pipe_expression(expr: ast::PipeExpr, fresher: &mut Fresher) -> Result<CallExpr> {
    let mut call = convert_call_expression(expr.call, fresher)?;
    let pipe = convert_expression(expr.argument, fresher)?;
    call.pipe = Some(pipe);
    Ok(call)
}

fn convert_binary_expression(expr: ast::BinaryExpr, fresher: &mut Fresher) -> Result<BinaryExpr> {
    let left = convert_expression(expr.left, fresher)?;
    let right = convert_expression(expr.right, fresher)?;
    Ok(BinaryExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        operator: expr.operator,
        left,
        right,
    })
}

fn convert_unary_expression(expr: ast::UnaryExpr, fresher: &mut Fresher) -> Result<UnaryExpr> {
    let argument = convert_expression(expr.argument, fresher)?;
    Ok(UnaryExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        operator: expr.operator,
        argument,
    })
}

fn convert_logical_expression(
    expr: ast::LogicalExpr,
    fresher: &mut Fresher,
) -> Result<LogicalExpr> {
    let left = convert_expression(expr.left, fresher)?;
    let right = convert_expression(expr.right, fresher)?;
    Ok(LogicalExpr {
        loc: expr.base.location,
        operator: expr.operator,
        left,
        right,
    })
}

fn convert_conditional_expression(
    expr: ast::ConditionalExpr,
    fresher: &mut Fresher,
) -> Result<ConditionalExpr> {
    let test = convert_expression(expr.test, fresher)?;
    let consequent = convert_expression(expr.consequent, fresher)?;
    let alternate = convert_expression(expr.alternate, fresher)?;
    Ok(ConditionalExpr {
        loc: expr.base.location,
        test,
        consequent,
        alternate,
    })
}

fn convert_object_expression(expr: ast::ObjectExpr, fresher: &mut Fresher) -> Result<ObjectExpr> {
    let properties = expr
        .properties
        .into_iter()
        .map(|p| convert_property(p, fresher))
        .collect::<Result<Vec<Property>>>()?;
    let with = match expr.with {
        Some(with) => Some(convert_identifier_expression(with.source, fresher)?),
        None => None,
    };
    Ok(ObjectExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        with,
        properties,
    })
}

fn convert_property(prop: ast::Property, fresher: &mut Fresher) -> Result<Property> {
    let key = match prop.key {
        ast::PropertyKey::Identifier(id) => convert_identifier(id, fresher)?,
        ast::PropertyKey::StringLit(lit) => Identifier {
            loc: lit.base.location.clone(),
            name: convert_string_literal(lit, fresher)?.value,
        },
    };
    let value = match prop.value {
        Some(expr) => convert_expression(expr, fresher)?,
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

fn convert_array_expression(expr: ast::ArrayExpr, fresher: &mut Fresher) -> Result<ArrayExpr> {
    let elements = expr
        .elements
        .into_iter()
        .map(|e| convert_expression(e.expression, fresher))
        .collect::<Result<Vec<Expression>>>()?;
    Ok(ArrayExpr {
        loc: expr.base.location,
        typ: MonoType::Var(fresher.fresh()),
        elements,
    })
}

fn convert_identifier(id: ast::Identifier, _fresher: &mut Fresher) -> Result<Identifier> {
    Ok(Identifier {
        loc: id.base.location,
        name: id.name,
    })
}

fn convert_identifier_expression(
    id: ast::Identifier,
    fresher: &mut Fresher,
) -> Result<IdentifierExpr> {
    Ok(IdentifierExpr {
        loc: id.base.location,
        typ: MonoType::Var(fresher.fresh()),
        name: id.name,
    })
}

fn convert_string_expression(expr: ast::StringExpr, fresher: &mut Fresher) -> Result<StringExpr> {
    let parts = expr
        .parts
        .into_iter()
        .map(|p| convert_string_expression_part(p, fresher))
        .collect::<Result<Vec<StringExprPart>>>()?;
    Ok(StringExpr {
        loc: expr.base.location,
        parts,
    })
}

fn convert_string_expression_part(
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
                expression: convert_expression(itp.expression, fresher)?,
            }))
        }
    }
}

fn convert_string_literal(lit: ast::StringLit, _: &mut Fresher) -> Result<StringLit> {
    Ok(StringLit {
        loc: lit.base.location,
        value: lit.value,
    })
}

fn convert_boolean_literal(lit: ast::BooleanLit, _: &mut Fresher) -> Result<BooleanLit> {
    Ok(BooleanLit {
        loc: lit.base.location,
        value: lit.value,
    })
}

fn convert_float_literal(lit: ast::FloatLit, _: &mut Fresher) -> Result<FloatLit> {
    Ok(FloatLit {
        loc: lit.base.location,
        value: lit.value,
    })
}

fn convert_integer_literal(lit: ast::IntegerLit, _: &mut Fresher) -> Result<IntegerLit> {
    Ok(IntegerLit {
        loc: lit.base.location,
        value: lit.value,
    })
}

fn convert_unsigned_integer_literal(lit: ast::UintLit, _: &mut Fresher) -> Result<UintLit> {
    Ok(UintLit {
        loc: lit.base.location,
        value: lit.value,
    })
}

fn convert_regexp_literal(lit: ast::RegexpLit, _: &mut Fresher) -> Result<RegexpLit> {
    Ok(RegexpLit {
        loc: lit.base.location,
        value: lit.value,
    })
}

fn convert_duration_literal(lit: ast::DurationLit, _: &mut Fresher) -> Result<DurationLit> {
    Ok(DurationLit {
        loc: lit.base.location,
        value: convert_duration(&lit.values)?,
    })
}

fn convert_date_time_literal(lit: ast::DateTimeLit, _: &mut Fresher) -> Result<DateTimeLit> {
    Ok(DateTimeLit {
        loc: lit.base.location,
        value: lit.value,
    })
}

// In these tests we test the results of semantic analysis on some ASTs.
// NOTE: we do not care about locations.
// We create a default base node and clone it in various AST nodes.
#[cfg(test)]
mod tests {
    use super::*;
    use crate::semantic::fresh;
    use crate::semantic::types::{MonoType, Tvar};
    use pretty_assertions::assert_eq;

    // type_info() is used for the expected semantic graph.
    // The id for the Tvar does not matter, because that is not compared.
    fn type_info() -> MonoType {
        MonoType::Var(Tvar(0))
    }

    fn test_convert(pkg: ast::Package) -> Result<Package> {
        convert_with(pkg, &mut fresh::Fresher::default())
    }

    #[test]
    fn test_convert_empty() {
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
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_package() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: Some(ast::PackageClause {
                    base: b.clone(),
                    name: ast::Identifier {
                        base: b.clone(),
                        name: "foo".to_string(),
                    },
                }),
                imports: Vec::new(),
                body: Vec::new(),
                eof: None,
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
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_imports() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
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
                eof: None,
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
                            value: "path/foo".to_string(),
                        },
                        alias: None,
                    },
                    ImportDeclaration {
                        loc: b.location.clone(),
                        path: StringLit {
                            loc: b.location.clone(),
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
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_var_assignment() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![
                    ast::Statement::Variable(Box::new(ast::VariableAssgn {
                        base: b.clone(),
                        id: ast::Identifier {
                            base: b.clone(),
                            name: "a".to_string(),
                        },
                        init: ast::Expression::Boolean(ast::BooleanLit {
                            base: b.clone(),
                            value: true,
                        }),
                    })),
                    ast::Statement::Expr(Box::new(ast::ExprStmt {
                        base: b.clone(),
                        expression: ast::Expression::Identifier(ast::Identifier {
                            base: b.clone(),
                            name: "a".to_string(),
                        }),
                    })),
                ],
                eof: None,
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
                    Statement::Variable(Box::new(VariableAssgn::new(
                        Identifier {
                            loc: b.location.clone(),
                            name: "a".to_string(),
                        },
                        Expression::Boolean(BooleanLit {
                            loc: b.location.clone(),
                            value: true,
                        }),
                        b.location.clone(),
                    ))),
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
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_object() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(Box::new(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Object(Box::new(ast::ObjectExpr {
                        base: b.clone(),
                        lbrace: None,
                        with: None,
                        properties: vec![ast::Property {
                            base: b.clone(),
                            key: ast::PropertyKey::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "a".to_string(),
                            }),
                            separator: None,
                            value: Some(ast::Expression::Integer(ast::IntegerLit {
                                base: b.clone(),
                                value: 10,
                            })),
                            comma: None,
                        }],
                        rbrace: None,
                    })),
                }))],
                eof: None,
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
                                value: 10,
                            }),
                        }],
                    })),
                })],
            }],
        };
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_object_with_string_key() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(Box::new(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Object(Box::new(ast::ObjectExpr {
                        base: b.clone(),
                        lbrace: None,
                        with: None,
                        properties: vec![ast::Property {
                            base: b.clone(),
                            key: ast::PropertyKey::StringLit(ast::StringLit {
                                base: b.clone(),
                                value: "a".to_string(),
                            }),
                            separator: None,
                            value: Some(ast::Expression::Integer(ast::IntegerLit {
                                base: b.clone(),
                                value: 10,
                            })),
                            comma: None,
                        }],
                        rbrace: None,
                    })),
                }))],
                eof: None,
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
                                value: 10,
                            }),
                        }],
                    })),
                })],
            }],
        };
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_object_with_mixed_keys() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(Box::new(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Object(Box::new(ast::ObjectExpr {
                        base: b.clone(),
                        lbrace: None,
                        with: None,
                        properties: vec![
                            ast::Property {
                                base: b.clone(),
                                key: ast::PropertyKey::StringLit(ast::StringLit {
                                    base: b.clone(),
                                    value: "a".to_string(),
                                }),
                                separator: None,
                                value: Some(ast::Expression::Integer(ast::IntegerLit {
                                    base: b.clone(),
                                    value: 10,
                                })),
                                comma: None,
                            },
                            ast::Property {
                                base: b.clone(),
                                key: ast::PropertyKey::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "b".to_string(),
                                }),
                                separator: None,
                                value: Some(ast::Expression::Integer(ast::IntegerLit {
                                    base: b.clone(),
                                    value: 11,
                                })),
                                comma: None,
                            },
                        ],
                        rbrace: None,
                    })),
                }))],
                eof: None,
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
                                    value: 11,
                                }),
                            },
                        ],
                    })),
                })],
            }],
        };
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_object_with_implicit_keys() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(Box::new(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Object(Box::new(ast::ObjectExpr {
                        base: b.clone(),
                        lbrace: None,
                        with: None,
                        properties: vec![
                            ast::Property {
                                base: b.clone(),
                                key: ast::PropertyKey::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "a".to_string(),
                                }),
                                separator: None,
                                value: None,
                                comma: None,
                            },
                            ast::Property {
                                base: b.clone(),
                                key: ast::PropertyKey::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "b".to_string(),
                                }),
                                separator: None,
                                value: None,
                                comma: None,
                            },
                        ],
                        rbrace: None,
                    })),
                }))],
                eof: None,
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
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_options_declaration() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Option(Box::new(ast::OptionStmt {
                    base: b.clone(),
                    assignment: ast::Assignment::Variable(Box::new(ast::VariableAssgn {
                        base: b.clone(),
                        id: ast::Identifier {
                            base: b.clone(),
                            name: "task".to_string(),
                        },
                        init: ast::Expression::Object(Box::new(ast::ObjectExpr {
                            base: b.clone(),
                            lbrace: None,
                            with: None,
                            properties: vec![
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "name".to_string(),
                                    }),
                                    separator: None,
                                    value: Some(ast::Expression::StringLit(ast::StringLit {
                                        base: b.clone(),
                                        value: "foo".to_string(),
                                    })),
                                    comma: None,
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "every".to_string(),
                                    }),
                                    separator: None,
                                    value: Some(ast::Expression::Duration(ast::DurationLit {
                                        base: b.clone(),
                                        values: vec![ast::Duration {
                                            magnitude: 1,
                                            unit: "h".to_string(),
                                        }],
                                    })),
                                    comma: None,
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "delay".to_string(),
                                    }),
                                    separator: None,
                                    value: Some(ast::Expression::Duration(ast::DurationLit {
                                        base: b.clone(),
                                        values: vec![ast::Duration {
                                            magnitude: 10,
                                            unit: "m".to_string(),
                                        }],
                                    })),
                                    comma: None,
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "cron".to_string(),
                                    }),
                                    separator: None,
                                    value: Some(ast::Expression::StringLit(ast::StringLit {
                                        base: b.clone(),
                                        value: "0 2 * * *".to_string(),
                                    })),
                                    comma: None,
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "retry".to_string(),
                                    }),
                                    separator: None,
                                    value: Some(ast::Expression::Integer(ast::IntegerLit {
                                        base: b.clone(),
                                        value: 5,
                                    })),
                                    comma: None,
                                },
                            ],
                            rbrace: None,
                        })),
                    })),
                }))],
                eof: None,
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![Statement::Option(Box::new(OptionStmt {
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
                                        value: Duration {
                                            months: 5,
                                            nanoseconds: 5000,
                                            negative: false,
                                        },
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
                                        value: Duration {
                                            months: 1,
                                            nanoseconds: 50,
                                            negative: true,
                                        },
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
                                        value: 5,
                                    }),
                                },
                            ],
                        })),
                        b.location.clone(),
                    )),
                }))],
            }],
        };
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_qualified_option_statement() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Option(Box::new(ast::OptionStmt {
                    base: b.clone(),
                    assignment: ast::Assignment::Member(Box::new(ast::MemberAssgn {
                        base: b.clone(),
                        member: ast::MemberExpr {
                            base: b.clone(),
                            object: ast::Expression::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "alert".to_string(),
                            }),
                            lbrack: None,
                            property: ast::PropertyKey::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "state".to_string(),
                            }),
                            rbrack: None,
                        },
                        init: ast::Expression::StringLit(ast::StringLit {
                            base: b.clone(),
                            value: "Warning".to_string(),
                        }),
                    })),
                }))],
                eof: None,
            }],
        };
        let want = Package {
            loc: b.location.clone(),
            package: "main".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: None,
                imports: Vec::new(),
                body: vec![Statement::Option(Box::new(OptionStmt {
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
                            value: "Warning".to_string(),
                        }),
                    }),
                }))],
            }],
        };
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_function() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![
                    ast::Statement::Variable(Box::new(ast::VariableAssgn {
                        base: b.clone(),
                        id: ast::Identifier {
                            base: b.clone(),
                            name: "f".to_string(),
                        },
                        init: ast::Expression::Function(Box::new(ast::FunctionExpr {
                            base: b.clone(),
                            lparen: None,
                            params: vec![
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "a".to_string(),
                                    }),
                                    separator: None,
                                    value: None,
                                    comma: None,
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "b".to_string(),
                                    }),
                                    separator: None,
                                    value: None,
                                    comma: None,
                                },
                            ],
                            rparen: None,
                            arrow: None,
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
                    })),
                    ast::Statement::Expr(Box::new(ast::ExprStmt {
                        base: b.clone(),
                        expression: ast::Expression::Call(Box::new(ast::CallExpr {
                            base: b.clone(),
                            callee: ast::Expression::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "f".to_string(),
                            }),
                            lparen: None,
                            arguments: vec![ast::Expression::Object(Box::new(ast::ObjectExpr {
                                base: b.clone(),
                                lbrace: None,
                                with: None,
                                properties: vec![
                                    ast::Property {
                                        base: b.clone(),
                                        key: ast::PropertyKey::Identifier(ast::Identifier {
                                            base: b.clone(),
                                            name: "a".to_string(),
                                        }),
                                        separator: None,
                                        value: Some(ast::Expression::Integer(ast::IntegerLit {
                                            base: b.clone(),
                                            value: 2,
                                        })),
                                        comma: None,
                                    },
                                    ast::Property {
                                        base: b.clone(),
                                        key: ast::PropertyKey::Identifier(ast::Identifier {
                                            base: b.clone(),
                                            name: "b".to_string(),
                                        }),
                                        separator: None,
                                        value: Some(ast::Expression::Integer(ast::IntegerLit {
                                            base: b.clone(),
                                            value: 3,
                                        })),
                                        comma: None,
                                    },
                                ],
                                rbrace: None,
                            }))],
                            rparen: None,
                        })),
                    })),
                ],
                eof: None,
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
                    Statement::Variable(Box::new(VariableAssgn::new(
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
                            body: Block::Return(ReturnStmt {
                                loc: b.location.clone(),
                                argument: Expression::Binary(Box::new(BinaryExpr {
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
                            }),
                        })),
                        b.location.clone(),
                    ))),
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
                                        value: 3,
                                    }),
                                },
                            ],
                        })),
                    }),
                ],
            }],
        };
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_function_with_defaults() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![
                    ast::Statement::Variable(Box::new(ast::VariableAssgn {
                        base: b.clone(),
                        id: ast::Identifier {
                            base: b.clone(),
                            name: "f".to_string(),
                        },
                        init: ast::Expression::Function(Box::new(ast::FunctionExpr {
                            base: b.clone(),
                            lparen: None,
                            params: vec![
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "a".to_string(),
                                    }),
                                    separator: None,
                                    value: Some(ast::Expression::Integer(ast::IntegerLit {
                                        base: b.clone(),
                                        value: 0,
                                    })),
                                    comma: None,
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "b".to_string(),
                                    }),
                                    separator: None,
                                    value: Some(ast::Expression::Integer(ast::IntegerLit {
                                        base: b.clone(),
                                        value: 0,
                                    })),
                                    comma: None,
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "c".to_string(),
                                    }),
                                    separator: None,
                                    value: None,
                                    comma: None,
                                },
                            ],
                            rparen: None,
                            arrow: None,
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
                    })),
                    ast::Statement::Expr(Box::new(ast::ExprStmt {
                        base: b.clone(),
                        expression: ast::Expression::Call(Box::new(ast::CallExpr {
                            base: b.clone(),
                            callee: ast::Expression::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "f".to_string(),
                            }),
                            lparen: None,
                            arguments: vec![ast::Expression::Object(Box::new(ast::ObjectExpr {
                                base: b.clone(),
                                lbrace: None,
                                with: None,
                                properties: vec![ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "c".to_string(),
                                    }),
                                    separator: None,
                                    value: Some(ast::Expression::Integer(ast::IntegerLit {
                                        base: b.clone(),
                                        value: 42,
                                    })),
                                    comma: None,
                                }],
                                rbrace: None,
                            }))],
                            rparen: None,
                        })),
                    })),
                ],
                eof: None,
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
                    Statement::Variable(Box::new(VariableAssgn::new(
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
                            body: Block::Return(ReturnStmt {
                                loc: b.location.clone(),
                                argument: Expression::Binary(Box::new(BinaryExpr {
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
                                })),
                            }),
                        })),
                        b.location.clone(),
                    ))),
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
                                    value: 42,
                                }),
                            }],
                        })),
                    }),
                ],
            }],
        };
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_function_multiple_pipes() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Variable(Box::new(ast::VariableAssgn {
                    base: b.clone(),
                    id: ast::Identifier {
                        base: b.clone(),
                        name: "f".to_string(),
                    },
                    init: ast::Expression::Function(Box::new(ast::FunctionExpr {
                        base: b.clone(),
                        lparen: None,
                        params: vec![
                            ast::Property {
                                base: b.clone(),
                                key: ast::PropertyKey::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "a".to_string(),
                                }),
                                separator: None,
                                value: None,
                                comma: None,
                            },
                            ast::Property {
                                base: b.clone(),
                                key: ast::PropertyKey::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "piped1".to_string(),
                                }),
                                separator: None,
                                value: Some(ast::Expression::PipeLit(ast::PipeLit {
                                    base: b.clone(),
                                })),
                                comma: None,
                            },
                            ast::Property {
                                base: b.clone(),
                                key: ast::PropertyKey::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "piped2".to_string(),
                                }),
                                separator: None,
                                value: Some(ast::Expression::PipeLit(ast::PipeLit {
                                    base: b.clone(),
                                })),
                                comma: None,
                            },
                        ],
                        rparen: None,
                        arrow: None,
                        body: ast::FunctionBody::Expr(ast::Expression::Identifier(
                            ast::Identifier {
                                base: b.clone(),
                                name: "a".to_string(),
                            },
                        )),
                    })),
                }))],
                eof: None,
            }],
        };
        let got = test_convert(pkg).err().unwrap().to_string();
        assert_eq!("only a single argument may be piped".to_string(), got);
    }

    #[test]
    fn test_convert_call_multiple_object_arguments() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(Box::new(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Call(Box::new(ast::CallExpr {
                        base: b.clone(),
                        callee: ast::Expression::Identifier(ast::Identifier {
                            base: b.clone(),
                            name: "f".to_string(),
                        }),
                        lparen: None,
                        arguments: vec![
                            ast::Expression::Object(Box::new(ast::ObjectExpr {
                                base: b.clone(),
                                lbrace: None,
                                with: None,
                                properties: vec![ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "a".to_string(),
                                    }),
                                    separator: None,
                                    value: Some(ast::Expression::Integer(ast::IntegerLit {
                                        base: b.clone(),
                                        value: 0,
                                    })),
                                    comma: None,
                                }],
                                rbrace: None,
                            })),
                            ast::Expression::Object(Box::new(ast::ObjectExpr {
                                base: b.clone(),
                                lbrace: None,
                                with: None,
                                properties: vec![ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "b".to_string(),
                                    }),
                                    separator: None,
                                    value: Some(ast::Expression::Integer(ast::IntegerLit {
                                        base: b.clone(),
                                        value: 1,
                                    })),
                                    comma: None,
                                }],
                                rbrace: None,
                            })),
                        ],
                        rparen: None,
                    })),
                }))],
                eof: None,
            }],
        };
        let got = test_convert(pkg).err().unwrap().to_string();
        assert_eq!(
            "arguments are more than one object expression".to_string(),
            got
        );
    }

    #[test]
    fn test_convert_pipe_expression() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![
                    ast::Statement::Variable(Box::new(ast::VariableAssgn {
                        base: b.clone(),
                        id: ast::Identifier {
                            base: b.clone(),
                            name: "f".to_string(),
                        },
                        init: ast::Expression::Function(Box::new(ast::FunctionExpr {
                            base: b.clone(),
                            lparen: None,
                            params: vec![
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "piped".to_string(),
                                    }),
                                    separator: None,
                                    value: Some(ast::Expression::PipeLit(ast::PipeLit {
                                        base: b.clone(),
                                    })),
                                    comma: None,
                                },
                                ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "a".to_string(),
                                    }),
                                    separator: None,
                                    value: None,
                                    comma: None,
                                },
                            ],
                            rparen: None,
                            arrow: None,
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
                    })),
                    ast::Statement::Expr(Box::new(ast::ExprStmt {
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
                                lparen: None,
                                arguments: vec![ast::Expression::Object(Box::new(
                                    ast::ObjectExpr {
                                        base: b.clone(),
                                        lbrace: None,
                                        with: None,
                                        properties: vec![ast::Property {
                                            base: b.clone(),
                                            key: ast::PropertyKey::Identifier(ast::Identifier {
                                                base: b.clone(),
                                                name: "a".to_string(),
                                            }),
                                            separator: None,
                                            value: Some(ast::Expression::Integer(
                                                ast::IntegerLit {
                                                    base: b.clone(),
                                                    value: 2,
                                                },
                                            )),
                                            comma: None,
                                        }],
                                        rbrace: None,
                                    },
                                ))],
                                rparen: None,
                            },
                        })),
                    })),
                ],
                eof: None,
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
                    Statement::Variable(Box::new(VariableAssgn::new(
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
                            body: Block::Return(ReturnStmt {
                                loc: b.location.clone(),
                                argument: Expression::Binary(Box::new(BinaryExpr {
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
                                })),
                            }),
                        })),
                        b.location.clone(),
                    ))),
                    Statement::Expr(ExprStmt {
                        loc: b.location.clone(),
                        expression: Expression::Call(Box::new(CallExpr {
                            loc: b.location.clone(),
                            typ: type_info(),
                            pipe: Some(Expression::Integer(IntegerLit {
                                loc: b.location.clone(),
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
                                    value: 2,
                                }),
                            }],
                        })),
                    }),
                ],
            }],
        };
        let got = test_convert(pkg).unwrap();
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
            body: Block::Return(ReturnStmt {
                loc: b.location.clone(),
                argument: Expression::Binary(Box::new(BinaryExpr {
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
            }),
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
            body: Block::Return(ReturnStmt {
                loc: b.location.clone(),
                argument: Expression::Binary(Box::new(BinaryExpr {
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
            }),
        };
        assert_eq!(defaults, f.defaults());
        assert_eq!(Some(&piped), f.pipe());
    }

    #[test]
    fn test_convert_index_expression() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(Box::new(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Index(Box::new(ast::IndexExpr {
                        base: b.clone(),
                        array: ast::Expression::Identifier(ast::Identifier {
                            base: b.clone(),
                            name: "a".to_string(),
                        }),
                        lbrack: None,
                        index: ast::Expression::Integer(ast::IntegerLit {
                            base: b.clone(),
                            value: 3,
                        }),
                        rbrack: None,
                    })),
                }))],
                eof: None,
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
                            value: 3,
                        }),
                    })),
                })],
            }],
        };
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_nested_index_expression() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(Box::new(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Index(Box::new(ast::IndexExpr {
                        base: b.clone(),
                        array: ast::Expression::Index(Box::new(ast::IndexExpr {
                            base: b.clone(),
                            array: ast::Expression::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "a".to_string(),
                            }),
                            lbrack: None,
                            index: ast::Expression::Integer(ast::IntegerLit {
                                base: b.clone(),
                                value: 3,
                            }),
                            rbrack: None,
                        })),
                        lbrack: None,
                        index: ast::Expression::Integer(ast::IntegerLit {
                            base: b.clone(),
                            value: 5,
                        }),
                        rbrack: None,
                    })),
                }))],
                eof: None,
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
                                value: 3,
                            }),
                        })),
                        index: Expression::Integer(IntegerLit {
                            loc: b.location.clone(),
                            value: 5,
                        }),
                    })),
                })],
            }],
        };
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_access_idexed_object_returned_from_function_call() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(Box::new(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Index(Box::new(ast::IndexExpr {
                        base: b.clone(),
                        array: ast::Expression::Call(Box::new(ast::CallExpr {
                            base: b.clone(),
                            callee: ast::Expression::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "f".to_string(),
                            }),
                            lparen: None,
                            arguments: vec![],
                            rparen: None,
                        })),
                        lbrack: None,
                        index: ast::Expression::Integer(ast::IntegerLit {
                            base: b.clone(),
                            value: 3,
                        }),
                        rbrack: None,
                    })),
                }))],
                eof: None,
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
                            value: 3,
                        }),
                    })),
                })],
            }],
        };
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_nested_member_expression() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(Box::new(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Member(Box::new(ast::MemberExpr {
                        base: b.clone(),
                        object: ast::Expression::Member(Box::new(ast::MemberExpr {
                            base: b.clone(),
                            object: ast::Expression::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "a".to_string(),
                            }),
                            lbrack: None,
                            property: ast::PropertyKey::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "b".to_string(),
                            }),
                            rbrack: None,
                        })),
                        lbrack: None,
                        property: ast::PropertyKey::Identifier(ast::Identifier {
                            base: b.clone(),
                            name: "c".to_string(),
                        }),
                        rbrack: None,
                    })),
                }))],
                eof: None,
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
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_member_with_call_expression() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(Box::new(ast::ExprStmt {
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
                                lbrack: None,
                                property: ast::PropertyKey::Identifier(ast::Identifier {
                                    base: b.clone(),
                                    name: "b".to_string(),
                                }),
                                rbrack: None,
                            })),
                            lparen: None,
                            arguments: vec![],
                            rparen: None,
                        })),
                        lbrack: None,
                        property: ast::PropertyKey::Identifier(ast::Identifier {
                            base: b.clone(),
                            name: "c".to_string(),
                        }),
                        rbrack: None,
                    })),
                }))],
                eof: None,
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
        let got = test_convert(pkg).unwrap();
        assert_eq!(want, got);
    }
    #[test]
    fn test_convert_bad_stmt() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Bad(Box::new(ast::BadStmt {
                    base: b.clone(),
                    text: "bad statement".to_string(),
                }))],
                eof: None,
            }],
        };
        let want: Result<Package> =
            Err("BadStatement is not supported in semantic analysis".to_string());
        let got = test_convert(pkg);
        assert_eq!(want, got);
    }
    #[test]
    fn test_convert_bad_expr() {
        let b = ast::BaseNode::default();
        let pkg = ast::Package {
            base: b.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![ast::File {
                base: b.clone(),
                name: "foo.flux".to_string(),
                metadata: String::new(),
                package: None,
                imports: Vec::new(),
                body: vec![ast::Statement::Expr(Box::new(ast::ExprStmt {
                    base: b.clone(),
                    expression: ast::Expression::Bad(Box::new(ast::BadExpr {
                        base: b.clone(),
                        text: "bad expression".to_string(),
                        expression: None,
                    })),
                }))],
                eof: None,
            }],
        };
        let want: Result<Package> =
            Err("BadExpression is not supported in semantic analysis".to_string());
        let got = test_convert(pkg);
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_monotype_int() {
        let b = ast::BaseNode::default();
        let monotype = ast::MonoType::Basic(ast::NamedType {
            base: b.clone(),
            name: ast::Identifier {
                base: b.clone(),
                name: "int".to_string(),
            },
        });
        let mut m = HashMap::<String, types::Tvar>::new();
        let got = convert_monotype(monotype, &mut m, &mut fresh::Fresher::default()).unwrap();
        let want = MonoType::Int;
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_monotype_record() {
        let b = ast::BaseNode::default();
        let monotype = ast::MonoType::Record(ast::RecordType {
            base: b.clone(),
            tvar: Some(ast::Identifier {
                base: b.clone(),
                name: "A".to_string(),
            }),
            properties: vec![ast::PropertyType {
                base: b.clone(),
                name: ast::Identifier {
                    base: b.clone(),
                    name: "B".to_string(),
                },
                monotype: ast::MonoType::Basic(ast::NamedType {
                    base: b.clone(),
                    name: ast::Identifier {
                        base: b.clone(),
                        name: "int".to_string(),
                    },
                }),
            }],
        });
        let mut m = HashMap::<String, types::Tvar>::new();
        let got = convert_monotype(monotype, &mut m, &mut fresh::Fresher::default()).unwrap();
        let want = MonoType::Row(Box::new(types::Row::Extension {
            head: types::Property {
                k: "B".to_string(),
                v: MonoType::Int,
            },
            tail: MonoType::Var(Tvar(0)),
        }));
        assert_eq!(want, got);
    }
    #[test]

    fn test_convert_monotype_function() {
        let b = ast::BaseNode::default();
        let monotype_ex = ast::MonoType::Function(Box::new(ast::FunctionType {
            base: b.clone(),
            parameters: vec![ast::ParameterType::Optional {
                base: b.clone(),
                name: ast::Identifier {
                    base: b.clone(),
                    name: "A".to_string(),
                },
                monotype: ast::MonoType::Basic(ast::NamedType {
                    base: b.clone(),
                    name: ast::Identifier {
                        base: b.clone(),
                        name: "int".to_string(),
                    },
                }),
            }],
            monotype: ast::MonoType::Basic(ast::NamedType {
                base: b.clone(),
                name: ast::Identifier {
                    base: b.clone(),
                    name: "int".to_string(),
                },
            }),
        }));
        let mut m = HashMap::<String, types::Tvar>::new();
        let got = convert_monotype(monotype_ex, &mut m, &mut fresh::Fresher::default()).unwrap();
        let mut opt = MonoTypeMap::new();
        opt.insert(String::from("A"), MonoType::Int);
        let want = MonoType::Fun(Box::new(types::Function {
            req: MonoTypeMap::new(),
            opt: opt,
            pipe: None,
            retn: MonoType::Int,
        }));
        assert_eq!(want, got);
    }
}
