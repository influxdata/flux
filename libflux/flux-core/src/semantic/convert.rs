//! Various conversions from AST nodes to their associated
//! types in the semantic graph.

use std::{collections::BTreeMap, fmt, sync::Arc};

use codespan_reporting::diagnostic;
use serde::{Serialize, Serializer};
use thiserror::Error;

use crate::{
    ast,
    errors::{located, AsDiagnostic, Errors, Located},
    semantic::{
        env::Environment,
        nodes::*,
        sub::Substitution,
        types::{self, BuiltinType, MonoType, MonoTypeMap, SemanticMap},
    },
};

/// Error that categorizes errors when converting from AST to semantic graph.
pub type Error = Located<ErrorKind>;

/// Error that categorizes errors when converting from AST to semantic graph.
#[derive(Error, Debug, PartialEq)]
#[allow(missing_docs)]
pub enum ErrorKind {
    #[error("TestCase is not supported in semantic analysis")]
    TestCase,
    #[error("invalid named type {0}")]
    InvalidNamedType(String),
    #[error("function types can have at most one pipe parameter")]
    AtMostOnePipe,
    #[error("invalid constraint {0}")]
    InvalidConstraint(String),
    #[error("a pipe literal may only be used as a default value for an argument in a function definition")]
    InvalidPipeLit,
    #[error("function parameters must be identifiers")]
    FunctionParameterIdents,
    #[error("missing return statement in block")]
    MissingReturn,
    #[error("invalid {0} statement in function block")]
    InvalidFunctionStatement(&'static str),
    #[error("function parameters is not a record expression")]
    ParametersNotRecord,
    #[error("function parameters are more than one record expression")]
    ExtraParameterRecord,
    #[error("invalid duration, {0}")]
    InvalidDuration(String),
}

impl AsDiagnostic for ErrorKind {
    fn as_diagnostic(&self, _source: &dyn crate::semantic::Source) -> diagnostic::Diagnostic<()> {
        diagnostic::Diagnostic::error().with_message(self.to_string())
    }
}

/// Result encapsulates any error during the conversion process.
pub type Result<T, E = Error> = std::result::Result<T, E>;

/// convert_package converts an [AST package] node to its semantic representation.
///
/// Note: most external callers of this function will want to use the analyze()
/// function in the flux crate instead, which is aware of everything in the Flux stdlib and prelude.
///
/// The function explicitly moves the `ast::Package` because it adds information to it.
/// We follow here the principle that every compilation step should be isolated and should add meaning
/// to the previous one. In other terms, once one converts an AST he should not use it anymore.
/// If one wants to do so, he should explicitly pkg.clone() and incur consciously in the memory
/// overhead involved.
///
/// [AST package]: ast::Package
pub fn convert_package(
    pkg: ast::Package,
    env: &Environment,
    sub: &mut Substitution,
) -> Result<Package, Errors<Error>> {
    let mut converter = Converter::with_env(sub, env);
    let r = converter.convert_package(pkg);
    converter.finish(r)
}

/// Converts a [type expression] in the AST into a [`PolyType`].
///
/// [type expression]: ast::TypeExpression
/// [`PolyType`]: types::PolyType
pub fn convert_polytype(
    type_expression: ast::TypeExpression,
    sub: &mut Substitution,
) -> Result<types::PolyType, Errors<Error>> {
    let mut converter = Converter::new(sub);
    let r = converter.convert_polytype(type_expression);
    converter.finish(r)
}

#[cfg(test)]
pub(crate) fn convert_monotype(
    ty: ast::MonoType,
    tvars: &mut BTreeMap<String, types::Tvar>,
    sub: &mut Substitution,
) -> Result<MonoType, Errors<Error>> {
    let mut converter = Converter::new(sub);
    let r = converter.convert_monotype(ty, tvars);
    converter.finish(r)
}

#[allow(missing_docs)]
#[derive(Debug, PartialOrd, Ord, PartialEq, Eq, Hash, Clone)]
pub struct Symbol {
    name: Arc<str>,
}

impl Serialize for Symbol {
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        self.name.serialize(serializer)
    }
}

impl std::ops::Deref for Symbol {
    type Target = str;
    fn deref(&self) -> &Self::Target {
        self.as_str()
    }
}

impl PartialEq<str> for Symbol {
    fn eq(&self, other: &str) -> bool {
        &self[..] == other
    }
}

impl PartialEq<&str> for Symbol {
    fn eq(&self, other: &&str) -> bool {
        &self[..] == *other
    }
}

impl fmt::Display for Symbol {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(&self[..])
    }
}

impl From<&str> for Symbol {
    fn from(name: &str) -> Self {
        Self {
            name: Arc::from(name),
        }
    }
}

impl From<String> for Symbol {
    fn from(name: String) -> Self {
        Self {
            name: Arc::from(name),
        }
    }
}

impl Symbol {
    /// Casts self into a `&str`
    pub fn as_str(&self) -> &str {
        self.name.split_once('@').map_or(&self.name, |x| x.0)
    }

    /// Returns just the name of the symbol
    pub fn name(&self) -> &str {
        self.as_str()
    }

    /// Returns the package that his symbol was defined in (if it exists)
    pub fn package(&self) -> Option<&str> {
        self.name.split_once('@').map(|x| x.1)
    }

    /// Returns the full name, package qualified name of `Symbol`
    pub fn full_name(&self) -> &str {
        &self.name
    }

    /// Attaches a package identifier to `self`
    pub fn with_package(self, package: &str) -> Self {
        Symbol::from(format!("{}@{}", self.as_str(), package))
    }
}

#[derive(Debug, Default)]
struct Symbols<'a> {
    parent: Option<Box<Symbols<'a>>>,
    env: Option<&'a Environment<'a>>,
    symbols: BTreeMap<String, Symbol>,
}

impl<'a> Symbols<'a> {
    fn with_env(env: &'a Environment) -> Self {
        Symbols {
            parent: None,
            env: Some(env),
            symbols: BTreeMap::default(),
        }
    }

    fn new_symbol(&mut self, name: String) -> Symbol {
        Symbol::from(name)
    }

    fn insert(&mut self, package: Option<&str>, name: String) -> Symbol {
        let symbol = self.new_symbol(match package {
            Some(package) => format!("{}@{}", name, package),
            None => name.clone(),
        });
        self.symbols.insert(name, symbol.clone());
        symbol
    }

    fn lookup(&mut self, name: &str) -> Symbol {
        self.lookup_option(name)
            .unwrap_or_else(|| self.new_symbol(name.into()))
    }

    fn lookup_option(&mut self, name: &str) -> Option<Symbol> {
        self.symbols
            .get(name)
            .or_else(|| self.env.and_then(|env| env.lookup_symbol(name)))
            .cloned()
            .or_else(|| {
                if let Some(parent) = &mut self.parent {
                    parent.lookup_option(name)
                } else {
                    None
                }
            })
    }

    fn enter_scope(&mut self) {
        let parent = std::mem::take(self);
        self.parent = Some(Box::new(parent));
    }

    fn exit_scope(&mut self) {
        match self.parent.take() {
            Some(env) => *self = *env,
            None => panic!("cannot pop final stack frame from symbols"),
        }
    }
}

struct Converter<'a> {
    sub: &'a mut Substitution,
    symbols: Symbols<'a>,
    errors: Errors<Error>,
}

impl<'a> Converter<'a> {
    fn new(sub: &'a mut Substitution) -> Self {
        Converter {
            sub,
            symbols: Symbols::default(),
            errors: Errors::new(),
        }
    }

    fn with_env(sub: &'a mut Substitution, env: &'a Environment) -> Self {
        Converter {
            sub,
            symbols: Symbols::with_env(env),
            errors: Errors::new(),
        }
    }

    fn finish<R>(mut self, result: Result<R>) -> Result<R, Errors<Error>> {
        let r = match result {
            Ok(r) => r,
            Err(err) => {
                self.errors.push(err);
                return Err(self.errors);
            }
        };
        if self.errors.has_errors() {
            Err(self.errors)
        } else {
            Ok(r)
        }
    }

    fn convert_package(&mut self, pkg: ast::Package) -> Result<Package> {
        let package = pkg.package;

        self.symbols.enter_scope();

        let files = pkg
            .files
            .into_iter()
            .map(|file| self.convert_file(&package, file))
            .collect::<Result<Vec<File>>>()?;

        self.symbols.exit_scope();

        Ok(Package {
            loc: pkg.base.location,
            package,
            files,
        })
    }

    fn convert_file(&mut self, package_name: &str, file: ast::File) -> Result<File> {
        let package = self.convert_package_clause(file.package)?;
        let imports = file
            .imports
            .into_iter()
            .map(|i| self.convert_import_declaration(i))
            .collect::<Result<Vec<ImportDeclaration>>>()?;
        let body = file
            .body
            .into_iter()
            .map(|s| self.convert_statement(package_name, s))
            .collect::<Result<Vec<Statement>>>()?;

        Ok(File {
            loc: file.base.location,
            package,
            imports,
            body,
        })
    }

    fn convert_package_clause(
        &mut self,
        pkg: Option<ast::PackageClause>,
    ) -> Result<Option<PackageClause>> {
        if pkg.is_none() {
            return Ok(None);
        }
        let pkg = pkg.unwrap();
        let name = self.convert_identifier(pkg.name)?;
        Ok(Some(PackageClause {
            loc: pkg.base.location,
            name,
        }))
    }

    fn convert_import_declaration(
        &mut self,
        imp: ast::ImportDeclaration,
    ) -> Result<ImportDeclaration> {
        let path = &imp.path.value;
        let (import_symbol, alias) = match imp.alias {
            None => {
                let name = path.rsplit_once('/').map_or(&path[..], |t| t.1).to_owned();
                (self.symbols.insert(None, name), None)
            }
            Some(id) => {
                let id = self.define_identifier(None, id)?;
                (id.name.clone(), Some(id))
            }
        };
        let path = self.convert_string_literal(imp.path)?;

        Ok(ImportDeclaration {
            loc: imp.base.location,
            alias,
            path,
            import_symbol,
        })
    }

    fn convert_statement(&mut self, package: &str, stmt: ast::Statement) -> Result<Statement> {
        match stmt {
            ast::Statement::Option(s) => Ok(Statement::Option(Box::new(
                self.convert_option_statement(*s)?,
            ))),
            ast::Statement::Builtin(s) => Ok(Statement::Builtin(
                self.convert_builtin_statement(package, *s)?,
            )),
            ast::Statement::Test(s) => {
                Ok(Statement::Test(Box::new(self.convert_test_statement(*s)?)))
            }
            ast::Statement::TestCase(s) => {
                self.errors
                    .push(located(s.base.location.clone(), ErrorKind::TestCase));
                Ok(Statement::Error(s.base.location.clone()))
            }
            ast::Statement::Expr(s) => Ok(Statement::Expr(self.convert_expression_statement(*s)?)),
            ast::Statement::Return(s) => Ok(Statement::Return(self.convert_return_statement(*s)?)),
            // TODO(affo): we should fix this to include MemberAssignement.
            //  The error lies in AST: the Statement enum does not include that.
            //  This is not a problem when parsing, because we parse it only in the option assignment case,
            //  and we return an OptionStmt, which is a Statement.
            ast::Statement::Variable(s) => Ok(Statement::Variable(Box::new(
                self.convert_variable_assignment(Some(package), *s)?,
            ))),
            ast::Statement::Bad(s) => Ok(Statement::Error(s.base.location.clone())),
        }
    }

    fn convert_assignment(&mut self, assign: ast::Assignment) -> Result<Assignment> {
        match assign {
            ast::Assignment::Variable(a) => Ok(Assignment::Variable(
                self.convert_variable_assignment(None, *a)?,
            )),
            ast::Assignment::Member(a) => {
                Ok(Assignment::Member(self.convert_member_assignment(*a)?))
            }
        }
    }

    fn convert_option_statement(&mut self, stmt: ast::OptionStmt) -> Result<OptionStmt> {
        Ok(OptionStmt {
            loc: stmt.base.location,
            assignment: self.convert_assignment(stmt.assignment)?,
        })
    }

    fn convert_builtin_statement(
        &mut self,
        package: &str,
        stmt: ast::BuiltinStmt,
    ) -> Result<BuiltinStmt> {
        Ok(BuiltinStmt {
            loc: stmt.base.location,
            id: self.define_identifier(Some(package), stmt.id)?,
            typ_expr: self.convert_polytype(stmt.ty)?,
        })
    }

    fn convert_builtintype(&mut self, basic: ast::NamedType) -> Result<BuiltinType> {
        Ok(match basic.name.name.as_str() {
            "bool" => BuiltinType::Bool,
            "int" => BuiltinType::Int,
            "uint" => BuiltinType::Uint,
            "float" => BuiltinType::Float,
            "string" => BuiltinType::String,
            "duration" => BuiltinType::Duration,
            "time" => BuiltinType::Time,
            "regexp" => BuiltinType::Regexp,
            "bytes" => BuiltinType::Bytes,
            _ => {
                return Err(located(
                    basic.base.location,
                    ErrorKind::InvalidNamedType(basic.name.name.to_string()),
                ))
            }
        })
    }

    fn convert_monotype(
        &mut self,
        ty: ast::MonoType,
        tvars: &mut BTreeMap<String, types::Tvar>,
    ) -> Result<MonoType> {
        match ty {
            ast::MonoType::Tvar(tv) => {
                let tvar = tvars
                    .entry(tv.name.name)
                    .or_insert_with(|| self.sub.fresh());
                Ok(MonoType::Var(*tvar))
            }

            ast::MonoType::Basic(basic) => Ok(MonoType::from(self.convert_builtintype(basic)?)),
            ast::MonoType::Array(arr) => Ok(MonoType::from(types::Array(
                self.convert_monotype(arr.element, tvars)?,
            ))),
            ast::MonoType::Dict(dict) => {
                let key = self.convert_monotype(dict.key, tvars)?;
                let val = self.convert_monotype(dict.val, tvars)?;
                Ok(MonoType::from(types::Dictionary { key, val }))
            }
            ast::MonoType::Function(func) => {
                let mut req = MonoTypeMap::new();
                let mut opt = MonoTypeMap::new();
                let mut _pipe = None;
                let mut dirty = false;
                for param in func.parameters {
                    match param {
                        ast::ParameterType::Required { name, monotype, .. } => {
                            req.insert(name.name, self.convert_monotype(monotype, tvars)?);
                        }
                        ast::ParameterType::Optional { name, monotype, .. } => {
                            opt.insert(name.name, self.convert_monotype(monotype, tvars)?);
                        }
                        ast::ParameterType::Pipe {
                            name,
                            monotype,
                            base,
                        } => {
                            if !dirty {
                                _pipe = Some(types::Property {
                                    k: match name {
                                        Some(n) => n.name,
                                        None => String::from("<-"),
                                    },
                                    v: self.convert_monotype(monotype, tvars)?,
                                });
                                dirty = true;
                            } else {
                                self.errors
                                    .push(located(base.location, ErrorKind::AtMostOnePipe));
                            }
                        }
                    }
                }
                Ok(MonoType::from(types::Function {
                    req,
                    opt,
                    pipe: _pipe,
                    retn: self.convert_monotype(func.monotype, tvars)?,
                }))
            }
            ast::MonoType::Record(rec) => {
                let mut r = match rec.tvar {
                    None => MonoType::from(types::Record::Empty),
                    Some(id) => {
                        let tv = ast::MonoType::Tvar(ast::TvarType {
                            base: id.clone().base,
                            name: id,
                        });
                        self.convert_monotype(tv, tvars)?
                    }
                };
                for prop in rec.properties {
                    let property = types::Property {
                        k: types::Label::from(self.symbols.lookup(&prop.name.name)),
                        v: self.convert_monotype(prop.monotype, tvars)?,
                    };
                    r = MonoType::from(types::Record::Extension {
                        head: property,
                        tail: r,
                    })
                }
                Ok(r)
            }
        }
    }

    // [`PolyType`]: types::PolyType
    fn convert_polytype(
        &mut self,
        type_expression: ast::TypeExpression,
    ) -> Result<types::PolyType> {
        let mut tvars = BTreeMap::<String, types::Tvar>::new();
        let expr = self.convert_monotype(type_expression.monotype, &mut tvars)?;
        let mut vars = Vec::<types::Tvar>::new();
        let mut cons = SemanticMap::<types::Tvar, Vec<types::Kind>>::new();

        for (name, tvar) in tvars {
            vars.push(tvar);
            let mut kinds = Vec::<types::Kind>::new();
            for con in &type_expression.constraints {
                if con.tvar.name == name {
                    for k in &con.kinds {
                        match k.name.as_str() {
                            "Addable" => kinds.push(types::Kind::Addable),
                            "Subtractable" => kinds.push(types::Kind::Subtractable),
                            "Divisible" => kinds.push(types::Kind::Divisible),
                            "Numeric" => kinds.push(types::Kind::Numeric),
                            "Comparable" => kinds.push(types::Kind::Comparable),
                            "Equatable" => kinds.push(types::Kind::Equatable),
                            "Nullable" => kinds.push(types::Kind::Nullable),
                            "Negatable" => kinds.push(types::Kind::Negatable),
                            "Timeable" => kinds.push(types::Kind::Timeable),
                            "Record" => kinds.push(types::Kind::Record),
                            "Basic" => kinds.push(types::Kind::Basic),
                            "Stringable" => kinds.push(types::Kind::Stringable),
                            _ => {
                                self.errors.push(located(
                                    k.base.location.clone(),
                                    ErrorKind::InvalidConstraint(k.name.clone()),
                                ));
                            }
                        }
                    }
                    cons.insert(tvar, kinds.clone());
                }
            }
        }
        Ok(types::PolyType { vars, cons, expr })
    }

    fn convert_test_statement(&mut self, stmt: ast::TestStmt) -> Result<TestStmt> {
        Ok(TestStmt {
            loc: stmt.base.location,
            assignment: self.convert_variable_assignment(None, stmt.assignment)?,
        })
    }

    fn convert_expression_statement(&mut self, stmt: ast::ExprStmt) -> Result<ExprStmt> {
        Ok(ExprStmt {
            loc: stmt.base.location,
            expression: self.convert_expression(stmt.expression)?,
        })
    }

    fn convert_return_statement(&mut self, stmt: ast::ReturnStmt) -> Result<ReturnStmt> {
        Ok(ReturnStmt {
            loc: stmt.base.location,
            argument: self.convert_expression(stmt.argument)?,
        })
    }

    fn convert_variable_assignment(
        &mut self,
        package: Option<&str>,
        stmt: ast::VariableAssgn,
    ) -> Result<VariableAssgn> {
        let expr = self.convert_expression(stmt.init)?;
        Ok(VariableAssgn::new(
            self.define_identifier(package, stmt.id)?,
            expr,
            stmt.base.location,
        ))
    }

    fn convert_member_assignment(&mut self, stmt: ast::MemberAssgn) -> Result<MemberAssgn> {
        let init = self.convert_expression(stmt.init)?;
        Ok(MemberAssgn {
            loc: stmt.base.location,
            member: self.convert_member_expression(stmt.member)?,
            init,
        })
    }

    fn convert_expression(&mut self, expr: ast::Expression) -> Result<Expression> {
        match expr {
            ast::Expression::Function(expr) => Ok(Expression::Function(Box::new(
                self.convert_function_expression(*expr)?,
            ))),
            ast::Expression::Call(expr) => Ok(Expression::Call(Box::new(
                self.convert_call_expression(*expr)?,
            ))),
            ast::Expression::Member(expr) => Ok(Expression::Member(Box::new(
                self.convert_member_expression(*expr)?,
            ))),
            ast::Expression::Index(expr) => Ok(Expression::Index(Box::new(
                self.convert_index_expression(*expr)?,
            ))),
            ast::Expression::PipeExpr(expr) => Ok(Expression::Call(Box::new(
                self.convert_pipe_expression(*expr)?,
            ))),
            ast::Expression::Binary(expr) => Ok(Expression::Binary(Box::new(
                self.convert_binary_expression(*expr)?,
            ))),
            ast::Expression::Unary(expr) => Ok(Expression::Unary(Box::new(
                self.convert_unary_expression(*expr)?,
            ))),
            ast::Expression::Logical(expr) => Ok(Expression::Logical(Box::new(
                self.convert_logical_expression(*expr)?,
            ))),
            ast::Expression::Conditional(expr) => Ok(Expression::Conditional(Box::new(
                self.convert_conditional_expression(*expr)?,
            ))),
            ast::Expression::Object(expr) => Ok(Expression::Object(Box::new(
                self.convert_object_expression(*expr)?,
            ))),
            ast::Expression::Array(expr) => Ok(Expression::Array(Box::new(
                self.convert_array_expression(*expr)?,
            ))),
            ast::Expression::Dict(expr) => Ok(Expression::Dict(Box::new(
                self.convert_dict_expression(*expr)?,
            ))),
            ast::Expression::Identifier(expr) => Ok(Expression::Identifier(
                self.convert_identifier_expression(expr)?,
            )),
            ast::Expression::StringExpr(expr) => Ok(Expression::StringExpr(Box::new(
                self.convert_string_expression(*expr)?,
            ))),
            ast::Expression::Paren(expr) => self.convert_expression(expr.expression),
            ast::Expression::StringLit(lit) => {
                Ok(Expression::StringLit(self.convert_string_literal(lit)?))
            }
            ast::Expression::Boolean(lit) => {
                Ok(Expression::Boolean(self.convert_boolean_literal(lit)?))
            }
            ast::Expression::Float(lit) => Ok(Expression::Float(self.convert_float_literal(lit)?)),
            ast::Expression::Integer(lit) => {
                Ok(Expression::Integer(self.convert_integer_literal(lit)?))
            }
            ast::Expression::Uint(lit) => Ok(Expression::Uint(
                self.convert_unsigned_integer_literal(lit)?,
            )),
            ast::Expression::Regexp(lit) => {
                Ok(Expression::Regexp(self.convert_regexp_literal(lit)?))
            }
            ast::Expression::Duration(lit) => {
                Ok(Expression::Duration(self.convert_duration_literal(lit)?))
            }
            ast::Expression::DateTime(lit) => {
                Ok(Expression::DateTime(self.convert_date_time_literal(lit)?))
            }
            ast::Expression::PipeLit(lit) => {
                self.errors.push(located(
                    lit.base.location.clone(),
                    ErrorKind::InvalidPipeLit,
                ));

                Ok(Expression::Error(lit.base.location))
            }
            ast::Expression::Bad(bad) => Ok(Expression::Error(bad.base.location.clone())),
        }
    }

    fn convert_function_expression(&mut self, expr: ast::FunctionExpr) -> Result<FunctionExpr> {
        self.symbols.enter_scope();

        let params = self.convert_function_params(expr.params)?;
        let body = self.convert_function_body(expr.body)?;

        self.symbols.exit_scope();

        Ok(FunctionExpr {
            loc: expr.base.location,
            typ: MonoType::Var(self.sub.fresh()),
            params,
            body,
            vectorized: None,
        })
    }

    fn convert_function_params(
        &mut self,
        mut props: Vec<ast::Property>,
    ) -> Result<Vec<FunctionParameter>> {
        // The defaults must be converted first so that the parameters are not in scope
        let mut piped = false;
        enum Default {
            Expr(Expression),
            Piped,
            None,
        }
        let defaults: Vec<_> = props
            .iter_mut()
            .map(|prop| {
                if let Some(expr) = prop.value.take() {
                    match expr {
                        ast::Expression::PipeLit(lit) => {
                            if piped {
                                return Err(located(lit.base.location, ErrorKind::AtMostOnePipe));
                            } else {
                                piped = true;
                            }
                            Ok(Default::Piped)
                        }
                        e => Ok(Default::Expr(self.convert_expression(e)?)),
                    }
                } else {
                    Ok(Default::None)
                }
            })
            .collect::<Result<_>>()?;

        // The iteration here is complex, cannot use iter().map()..., better to write it explicitly.
        let mut params: Vec<FunctionParameter> = Vec::new();
        for (prop, default) in props.into_iter().zip(defaults) {
            let id = match prop.key {
                ast::PropertyKey::Identifier(id) => id,
                _ => {
                    self.errors.push(located(
                        prop.base.location.clone(),
                        ErrorKind::FunctionParameterIdents,
                    ));
                    continue;
                }
            };
            let key = self.define_identifier(None, id)?;

            let (is_pipe, default) = match default {
                Default::Expr(expr) => (false, Some(expr)),
                Default::Piped => (true, None),
                Default::None => (false, None),
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

    fn convert_function_body(&mut self, body: ast::FunctionBody) -> Result<Block> {
        match body {
            ast::FunctionBody::Expr(expr) => {
                let argument = self.convert_expression(expr)?;
                Ok(Block::Return(ReturnStmt {
                    loc: argument.loc().clone(),
                    argument,
                }))
            }
            ast::FunctionBody::Block(block) => Ok(self.convert_block(block)?),
        }
    }

    fn convert_block(&mut self, block: ast::Block) -> Result<Block> {
        enum TempBlock {
            Variable(Box<VariableAssgn>),
            Expr(ExprStmt),
            Return(ReturnStmt),
        }

        impl TempBlock {
            fn loc(&self) -> &ast::SourceLocation {
                match self {
                    TempBlock::Variable(dec) => &dec.loc,
                    TempBlock::Expr(stmt) => &stmt.loc,
                    TempBlock::Return(s) => &s.loc,
                }
            }
        }

        let mut body = Vec::with_capacity(block.body.len());
        for s in block.body {
            match s {
                ast::Statement::Variable(dec) => body.push(TempBlock::Variable(Box::new(
                    self.convert_variable_assignment(None, *dec)?,
                ))),
                ast::Statement::Expr(stmt) => {
                    body.push(TempBlock::Expr(self.convert_expression_statement(*stmt)?))
                }
                ast::Statement::Return(stmt) => {
                    let argument = self.convert_expression(stmt.argument)?;
                    body.push(TempBlock::Return(ReturnStmt {
                        loc: stmt.base.location.clone(),
                        argument,
                    }));
                }
                _ => {
                    self.errors.push(located(
                        s.base().location.clone(),
                        ErrorKind::InvalidFunctionStatement(s.type_name()),
                    ));
                }
            }
        }

        let mut body = body.into_iter().rev();
        let block = match body.next() {
            Some(TempBlock::Return(stmt)) => Block::Return(stmt),
            Some(s) => {
                self.errors
                    .push(located(s.loc().clone(), ErrorKind::MissingReturn));
                Block::Return(ReturnStmt {
                    loc: s.loc().clone(),
                    argument: Expression::Error(s.loc().clone()),
                })
            }
            None => {
                self.errors.push(located(
                    block.base.location.clone(),
                    ErrorKind::MissingReturn,
                ));
                Block::Return(ReturnStmt {
                    loc: block.base.location.clone(),
                    argument: Expression::Error(block.base.location),
                })
            }
        };

        body.try_fold(block, |acc, s| match s {
            TempBlock::Variable(dec) => Ok(Block::Variable(dec, Box::new(acc))),
            TempBlock::Expr(stmt) => Ok(Block::Expr(stmt, Box::new(acc))),
            TempBlock::Return(s) => {
                self.errors.push(located(
                    s.loc,
                    ErrorKind::InvalidFunctionStatement("return"),
                ));
                Ok(acc)
            }
        })
    }

    fn convert_call_expression(&mut self, expr: ast::CallExpr) -> Result<CallExpr> {
        let callee = self.convert_expression(expr.callee)?;
        // TODO(affo): I'd prefer these checks to be in ast.Check().
        if expr.arguments.len() > 1 {
            return Err(located(expr.base.location, ErrorKind::ExtraParameterRecord));
        }
        let mut args = expr
            .arguments
            .into_iter()
            .map(|a| match a {
                ast::Expression::Object(obj) => self.convert_object_expression(*obj),
                _ => Err(located(
                    a.base().location.clone(),
                    ErrorKind::ParametersNotRecord,
                )),
            })
            .collect::<Result<Vec<ObjectExpr>>>()?;
        let arguments = match args.len() {
            0 => Ok(Vec::new()),
            1 => Ok(args.pop().expect("there must be 1 element").properties),
            _ => Err(located(
                expr.base.location.clone(),
                ErrorKind::ExtraParameterRecord,
            )),
        }?;
        Ok(CallExpr {
            loc: expr.base.location,
            typ: MonoType::Var(self.sub.fresh()),
            callee,
            arguments,
            pipe: None,
        })
    }

    fn convert_member_expression(&mut self, expr: ast::MemberExpr) -> Result<MemberExpr> {
        let object = self.convert_expression(expr.object)?;
        let property = match expr.property {
            ast::PropertyKey::Identifier(id) => id.name,
            ast::PropertyKey::StringLit(lit) => lit.value,
        };
        let property = self.symbols.lookup(&property);
        Ok(MemberExpr {
            loc: expr.base.location,
            typ: MonoType::Var(self.sub.fresh()),
            object,
            property,
        })
    }

    fn convert_index_expression(&mut self, expr: ast::IndexExpr) -> Result<IndexExpr> {
        let array = self.convert_expression(expr.array)?;
        let index = self.convert_expression(expr.index)?;
        Ok(IndexExpr {
            loc: expr.base.location,
            typ: MonoType::Var(self.sub.fresh()),
            array,
            index,
        })
    }

    fn convert_pipe_expression(&mut self, expr: ast::PipeExpr) -> Result<CallExpr> {
        let mut call = self.convert_call_expression(expr.call)?;
        let pipe = self.convert_expression(expr.argument)?;
        call.pipe = Some(pipe);
        Ok(call)
    }

    fn convert_binary_expression(&mut self, expr: ast::BinaryExpr) -> Result<BinaryExpr> {
        let left = self.convert_expression(expr.left)?;
        let right = self.convert_expression(expr.right)?;
        Ok(BinaryExpr {
            loc: expr.base.location,
            typ: MonoType::Var(self.sub.fresh()),
            operator: expr.operator,
            left,
            right,
        })
    }

    fn convert_unary_expression(&mut self, expr: ast::UnaryExpr) -> Result<UnaryExpr> {
        let argument = self.convert_expression(expr.argument)?;
        Ok(UnaryExpr {
            loc: expr.base.location,
            typ: MonoType::Var(self.sub.fresh()),
            operator: expr.operator,
            argument,
        })
    }

    fn convert_logical_expression(&mut self, expr: ast::LogicalExpr) -> Result<LogicalExpr> {
        let left = self.convert_expression(expr.left)?;
        let right = self.convert_expression(expr.right)?;
        Ok(LogicalExpr {
            loc: expr.base.location,
            operator: expr.operator,
            left,
            right,
        })
    }

    fn convert_conditional_expression(
        &mut self,
        expr: ast::ConditionalExpr,
    ) -> Result<ConditionalExpr> {
        let test = self.convert_expression(expr.test)?;
        let consequent = self.convert_expression(expr.consequent)?;
        let alternate = self.convert_expression(expr.alternate)?;
        Ok(ConditionalExpr {
            loc: expr.base.location,
            test,
            consequent,
            alternate,
        })
    }

    fn convert_object_expression(&mut self, expr: ast::ObjectExpr) -> Result<ObjectExpr> {
        let properties = expr
            .properties
            .into_iter()
            .map(|p| self.convert_property(p))
            .collect::<Result<Vec<Property>>>()?;
        let with = match expr.with {
            Some(with) => Some(self.convert_identifier_expression(with.source)?),
            None => None,
        };
        Ok(ObjectExpr {
            loc: expr.base.location,
            typ: MonoType::Var(self.sub.fresh()),
            with,
            properties,
        })
    }

    fn convert_property(&mut self, prop: ast::Property) -> Result<Property> {
        let key = match prop.key {
            ast::PropertyKey::Identifier(id) => self.convert_identifier(id)?,
            ast::PropertyKey::StringLit(lit) => {
                let loc = lit.base.location.clone();
                let name = self.convert_string_literal(lit)?.value;
                Identifier {
                    name: self.symbols.lookup(&name),
                    loc,
                }
            }
        };
        let value = match prop.value {
            Some(expr) => self.convert_expression(expr)?,
            None => Expression::Identifier(IdentifierExpr {
                loc: key.loc.clone(),
                typ: MonoType::Var(self.sub.fresh()),
                name: key.name.clone(),
            }),
        };
        Ok(Property {
            loc: prop.base.location,
            key,
            value,
        })
    }

    fn convert_array_expression(&mut self, expr: ast::ArrayExpr) -> Result<ArrayExpr> {
        let elements = expr
            .elements
            .into_iter()
            .map(|e| self.convert_expression(e.expression))
            .collect::<Result<Vec<Expression>>>()?;
        Ok(ArrayExpr {
            loc: expr.base.location,
            typ: MonoType::Var(self.sub.fresh()),
            elements,
        })
    }

    fn convert_dict_expression(&mut self, expr: ast::DictExpr) -> Result<DictExpr> {
        let mut elements = Vec::new();
        for item in expr.elements.into_iter() {
            elements.push((
                self.convert_expression(item.key)?,
                self.convert_expression(item.val)?,
            ));
        }
        Ok(DictExpr {
            loc: expr.base.location,
            typ: MonoType::Var(self.sub.fresh()),
            elements,
        })
    }

    fn define_identifier(
        &mut self,
        package: Option<&str>,
        id: ast::Identifier,
    ) -> Result<Identifier> {
        let name = self.symbols.insert(package, id.name);
        Ok(Identifier {
            loc: id.base.location,
            name,
        })
    }

    fn convert_identifier(&mut self, id: ast::Identifier) -> Result<Identifier> {
        Ok(Identifier {
            name: self.symbols.lookup(&id.name),
            loc: id.base.location,
        })
    }

    fn convert_identifier_expression(&mut self, id: ast::Identifier) -> Result<IdentifierExpr> {
        Ok(IdentifierExpr {
            typ: MonoType::Var(self.sub.fresh()),
            name: self.symbols.lookup(&id.name),
            loc: id.base.location,
        })
    }

    fn convert_string_expression(&mut self, expr: ast::StringExpr) -> Result<StringExpr> {
        let parts = expr
            .parts
            .into_iter()
            .map(|p| self.convert_string_expression_part(p))
            .collect::<Result<Vec<StringExprPart>>>()?;
        Ok(StringExpr {
            loc: expr.base.location,
            parts,
        })
    }

    fn convert_string_expression_part(
        &mut self,
        expr: ast::StringExprPart,
    ) -> Result<StringExprPart> {
        match expr {
            ast::StringExprPart::Text(txt) => Ok(StringExprPart::Text(TextPart {
                loc: txt.base.location,
                value: txt.value,
            })),
            ast::StringExprPart::Interpolated(itp) => {
                Ok(StringExprPart::Interpolated(InterpolatedPart {
                    loc: itp.base.location,
                    expression: self.convert_expression(itp.expression)?,
                }))
            }
        }
    }

    fn convert_string_literal(&mut self, lit: ast::StringLit) -> Result<StringLit> {
        Ok(StringLit {
            loc: lit.base.location,
            value: lit.value,
        })
    }

    fn convert_boolean_literal(&mut self, lit: ast::BooleanLit) -> Result<BooleanLit> {
        Ok(BooleanLit {
            loc: lit.base.location,
            value: lit.value,
        })
    }

    fn convert_float_literal(&mut self, lit: ast::FloatLit) -> Result<FloatLit> {
        Ok(FloatLit {
            loc: lit.base.location,
            value: lit.value,
        })
    }

    fn convert_integer_literal(&mut self, lit: ast::IntegerLit) -> Result<IntegerLit> {
        Ok(IntegerLit {
            loc: lit.base.location,
            value: lit.value,
        })
    }

    fn convert_unsigned_integer_literal(&mut self, lit: ast::UintLit) -> Result<UintLit> {
        Ok(UintLit {
            loc: lit.base.location,
            value: lit.value,
        })
    }

    fn convert_regexp_literal(&mut self, lit: ast::RegexpLit) -> Result<RegexpLit> {
        Ok(RegexpLit {
            loc: lit.base.location,
            value: lit.value,
        })
    }

    fn convert_duration_literal(&mut self, lit: ast::DurationLit) -> Result<DurationLit> {
        Ok(DurationLit {
            value: convert_duration(&lit.values).map_err(|e| {
                located(
                    lit.base.location.clone(),
                    ErrorKind::InvalidDuration(e.to_string()),
                )
            })?,
            loc: lit.base.location,
        })
    }

    fn convert_date_time_literal(&mut self, lit: ast::DateTimeLit) -> Result<DateTimeLit> {
        Ok(DateTimeLit {
            loc: lit.base.location,
            value: lit.value,
        })
    }
}

// In these tests we test the results of semantic analysis on some ASTs.
// NOTE: we do not care about locations.
// We create a default base node and clone it in various AST nodes.
#[cfg(test)]
mod tests {
    use expect_test::expect;
    use pretty_assertions::assert_eq;

    use super::*;
    use crate::{
        ast,
        parser::Parser,
        semantic::{
            sub,
            types::{MonoType, Tvar},
            walk::{walk_mut, NodeMut},
        },
    };

    fn parse_package(pkg: &str) -> ast::Package {
        let pkg = Parser::new(pkg).parse_single_package("path".to_string(), "foo.flux".to_string());
        ast::check::check(ast::walk::Node::Package(&pkg)).unwrap_or_else(|err| panic!("{}", err));
        pkg
    }

    // type_info() is used for the expected semantic graph.
    // The id for the Tvar does not matter, because that is not compared.
    fn type_info() -> MonoType {
        MonoType::Var(Tvar(0))
    }

    fn test_convert(pkg: ast::Package) -> Result<Package, Errors<Error>> {
        let mut sub = sub::Substitution::default();
        let mut converter = Converter::new(&mut sub);
        let r = converter.convert_package(pkg);
        let mut pkg = converter.finish(r)?;

        // We don't want to specifc the exact locations for each node in the tests
        walk_mut(
            &mut |n: &mut NodeMut| n.set_loc(ast::BaseNode::default().location),
            &mut NodeMut::Package(&mut pkg),
        );

        Ok(pkg)
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
        let pkg = parse_package("package foo");
        let want = Package {
            loc: b.location.clone(),
            package: "foo".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: Some(PackageClause {
                    loc: b.location.clone(),
                    name: Identifier {
                        loc: b.location.clone(),
                        name: Symbol::from("foo"),
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
        let pkg = parse_package(
            r#"package foo
            import "path/foo"
            import b "path/bar"
            "#,
        );
        let want = Package {
            loc: b.location.clone(),
            package: "foo".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: Some(PackageClause {
                    loc: b.location.clone(),
                    name: Identifier {
                        loc: b.location.clone(),
                        name: Symbol::from("foo"),
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
                        import_symbol: Symbol::from("foo"),
                    },
                    ImportDeclaration {
                        loc: b.location.clone(),
                        path: StringLit {
                            loc: b.location.clone(),
                            value: "path/bar".to_string(),
                        },
                        alias: Some(Identifier {
                            loc: b.location.clone(),
                            name: Symbol::from("b"),
                        }),
                        import_symbol: Symbol::from("b"),
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
                eof: vec![],
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
                            name: Symbol::from("a@main"),
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
                            name: Symbol::from("a@main"),
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
        let pkg = parse_package(r#"{ a: 10 }"#);
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
                                name: Symbol::from("a"),
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
        let pkg = parse_package(r#"{ "a": 10 }"#);
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
                                name: Symbol::from("a"),
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
        let pkg = parse_package(r#"{ "a": 10, b: 11 }"#);
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
                                    name: Symbol::from("a"),
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
                                    name: Symbol::from("b"),
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
        let pkg = parse_package("{ a, b }");
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
                                    name: Symbol::from("a"),
                                },
                                value: Expression::Identifier(IdentifierExpr {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    name: Symbol::from("a"),
                                }),
                            },
                            Property {
                                loc: b.location.clone(),
                                key: Identifier {
                                    loc: b.location.clone(),
                                    name: Symbol::from("b"),
                                },
                                value: Expression::Identifier(IdentifierExpr {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    name: Symbol::from("b"),
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
        let pkg = parse_package(
            r#"option task = { name: "foo", every: 1h, delay: 10m, cron: "0 2 * * *", retry: 5}"#,
        );
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
                            name: Symbol::from("task"),
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
                                        name: Symbol::from("name"),
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
                                        name: Symbol::from("every"),
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
                                        name: Symbol::from("delay"),
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
                                        name: Symbol::from("cron"),
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
                                        name: Symbol::from("retry"),
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
        let pkg = parse_package(r#"option alert.state = "Warning""#);
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
                                name: Symbol::from("alert"),
                            }),
                            property: Symbol::from("state"),
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
        let pkg = parse_package(
            "f = (a, b) => a + b
            f(a: 2, b: 3)",
        );
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
                            name: Symbol::from("f@main"),
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
                                        name: Symbol::from("a"),
                                    },
                                    default: None,
                                },
                                FunctionParameter {
                                    loc: b.location.clone(),
                                    is_pipe: false,
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: Symbol::from("b"),
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
                                        name: Symbol::from("a"),
                                    }),
                                    right: Expression::Identifier(IdentifierExpr {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        name: Symbol::from("b"),
                                    }),
                                })),
                            }),
                            vectorized: None,
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
                                name: Symbol::from("f@main"),
                            }),
                            arguments: vec![
                                Property {
                                    loc: b.location.clone(),
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: Symbol::from("a"),
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
                                        name: Symbol::from("b"),
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
        let pkg = parse_package(
            "f = (a=0, b=0, c) => a + b + c
            f(c: 42)",
        );
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
                            name: Symbol::from("f@main"),
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
                                        name: Symbol::from("a"),
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
                                        name: Symbol::from("b"),
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
                                        name: Symbol::from("c"),
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
                                            name: Symbol::from("a"),
                                        }),
                                        right: Expression::Identifier(IdentifierExpr {
                                            loc: b.location.clone(),
                                            typ: type_info(),
                                            name: Symbol::from("b"),
                                        }),
                                    })),
                                    right: Expression::Identifier(IdentifierExpr {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        name: Symbol::from("c"),
                                    }),
                                })),
                            }),
                            vectorized: None,
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
                                name: Symbol::from("f@main"),
                            }),
                            arguments: vec![Property {
                                loc: b.location.clone(),
                                key: Identifier {
                                    loc: b.location.clone(),
                                    name: Symbol::from("c"),
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
        let pkg = parse_package("f = (a, piped1=<-, piped2=<-) => a");
        let got = test_convert(pkg).err().unwrap().to_string();
        expect![["error foo.flux@1:27-1:29: function types can have at most one pipe parameter"]]
            .assert_eq(&got);
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
                        lparen: vec![],
                        arguments: vec![
                            ast::Expression::Object(Box::new(ast::ObjectExpr {
                                base: b.clone(),
                                lbrace: vec![],
                                with: None,
                                properties: vec![ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "a".to_string(),
                                    }),
                                    separator: vec![],
                                    value: Some(ast::Expression::Integer(ast::IntegerLit {
                                        base: b.clone(),
                                        value: 0,
                                    })),
                                    comma: vec![],
                                }],
                                rbrace: vec![],
                            })),
                            ast::Expression::Object(Box::new(ast::ObjectExpr {
                                base: b.clone(),
                                lbrace: vec![],
                                with: None,
                                properties: vec![ast::Property {
                                    base: b.clone(),
                                    key: ast::PropertyKey::Identifier(ast::Identifier {
                                        base: b.clone(),
                                        name: "b".to_string(),
                                    }),
                                    separator: vec![],
                                    value: Some(ast::Expression::Integer(ast::IntegerLit {
                                        base: b.clone(),
                                        value: 1,
                                    })),
                                    comma: vec![],
                                }],
                                rbrace: vec![],
                            })),
                        ],
                        rparen: vec![],
                    })),
                }))],
                eof: vec![],
            }],
        };
        let got = test_convert(pkg).err().unwrap().to_string();

        expect![["error @0:0-0:0: function parameters are more than one record expression"]]
            .assert_eq(&got);
    }

    #[test]
    fn test_convert_pipe_expression() {
        let b = ast::BaseNode::default();
        let pkg = parse_package(
            "f = (piped=<-, a) => a + piped
            3 |> f(a: 2)",
        );

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
                            name: Symbol::from("f@main"),
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
                                        name: Symbol::from("piped"),
                                    },
                                    default: None,
                                },
                                FunctionParameter {
                                    loc: b.location.clone(),
                                    is_pipe: false,
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: Symbol::from("a"),
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
                                        name: Symbol::from("a"),
                                    }),
                                    right: Expression::Identifier(IdentifierExpr {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        name: Symbol::from("piped"),
                                    }),
                                })),
                            }),
                            vectorized: None,
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
                                name: Symbol::from("f@main"),
                            }),
                            arguments: vec![Property {
                                loc: b.location.clone(),
                                key: Identifier {
                                    loc: b.location.clone(),
                                    name: Symbol::from("a"),
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
                        name: Symbol::from("a"),
                    },
                    default: None,
                },
                FunctionParameter {
                    loc: b.location.clone(),
                    is_pipe: false,
                    key: Identifier {
                        loc: b.location.clone(),
                        name: Symbol::from("b"),
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
                        name: Symbol::from("a"),
                    }),
                    right: Expression::Identifier(IdentifierExpr {
                        loc: b.location.clone(),
                        typ: type_info(),
                        name: Symbol::from("b"),
                    }),
                })),
            }),
            vectorized: None,
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
                name: Symbol::from("a"),
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
                name: Symbol::from("b"),
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
                name: Symbol::from("c"),
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
                name: Symbol::from("d"),
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
                        name: Symbol::from("a"),
                    }),
                    right: Expression::Identifier(IdentifierExpr {
                        loc: b.location.clone(),
                        typ: type_info(),
                        name: Symbol::from("b"),
                    }),
                })),
            }),
            vectorized: None,
        };
        assert_eq!(defaults, f.defaults());
        assert_eq!(Some(&piped), f.pipe());
    }

    #[test]
    fn test_convert_index_expression() {
        let b = ast::BaseNode::default();
        let pkg =
            Parser::new("a[3]").parse_single_package("path".to_string(), "foo.flux".to_string());
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
                            name: Symbol::from("a"),
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
        let pkg =
            Parser::new("a[3][5]").parse_single_package("path".to_string(), "foo.flux".to_string());
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
                                name: Symbol::from("a"),
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
    fn test_convert_access_indexed_object_returned_from_function_call() {
        let b = ast::BaseNode::default();
        let pkg =
            Parser::new("f()[3]").parse_single_package("path".to_string(), "foo.flux".to_string());
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
                                name: Symbol::from("f"),
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
                            lbrack: vec![],
                            property: ast::PropertyKey::Identifier(ast::Identifier {
                                base: b.clone(),
                                name: "b".to_string(),
                            }),
                            rbrack: vec![],
                        })),
                        lbrack: vec![],
                        property: ast::PropertyKey::Identifier(ast::Identifier {
                            base: b.clone(),
                            name: "c".to_string(),
                        }),
                        rbrack: vec![],
                    })),
                }))],
                eof: vec![],
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
                                name: Symbol::from("a"),
                            }),
                            property: Symbol::from("b"),
                        })),
                        property: Symbol::from("c"),
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
        let pkg =
            Parser::new("a.b().c").parse_single_package("path".to_string(), "foo.flux".to_string());

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
                                    name: Symbol::from("a"),
                                }),
                                property: Symbol::from("b"),
                            })),
                            arguments: Vec::new(),
                        })),
                        property: Symbol::from("c"),
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
                eof: vec![],
            }],
        };
        test_convert(pkg).unwrap();
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
                eof: vec![],
            }],
        };
        test_convert(pkg).unwrap();
    }

    #[test]
    fn test_convert_monotype_int() {
        let monotype = Parser::new("int").parse_monotype();
        let mut m = BTreeMap::<String, types::Tvar>::new();
        let got = convert_monotype(monotype, &mut m, &mut sub::Substitution::default()).unwrap();
        let want = MonoType::INT;
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_monotype_record() {
        let monotype = Parser::new("{ A with B: int }").parse_monotype();

        let mut m = BTreeMap::<String, types::Tvar>::new();
        let got = convert_monotype(monotype, &mut m, &mut sub::Substitution::default()).unwrap();
        let want = MonoType::from(types::Record::Extension {
            head: types::Property {
                k: types::Label::from("B"),
                v: MonoType::INT,
            },
            tail: MonoType::Var(Tvar(0)),
        });
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_monotype_function() {
        let monotype_ex = Parser::new("(?A: int) => int").parse_monotype();

        let mut m = BTreeMap::<String, types::Tvar>::new();
        let got = convert_monotype(monotype_ex, &mut m, &mut sub::Substitution::default()).unwrap();
        let mut opt = MonoTypeMap::new();
        opt.insert(String::from("A"), MonoType::INT);
        let want = MonoType::from(types::Function {
            req: MonoTypeMap::new(),
            opt,
            pipe: None,
            retn: MonoType::INT,
        });
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_polytype() {
        let type_exp =
            Parser::new("(A: T, B: S) => T where T: Addable, S: Divisible").parse_type_expression();
        let got = convert_polytype(type_exp, &mut sub::Substitution::default()).unwrap();
        let mut vars = Vec::<types::Tvar>::new();
        vars.push(types::Tvar(0));
        vars.push(types::Tvar(1));
        let mut cons = types::TvarKinds::new();
        let mut kind_vector_1 = Vec::<types::Kind>::new();
        kind_vector_1.push(types::Kind::Addable);
        cons.insert(types::Tvar(0), kind_vector_1);

        let mut kind_vector_2 = Vec::<types::Kind>::new();
        kind_vector_2.push(types::Kind::Divisible);
        cons.insert(types::Tvar(1), kind_vector_2);

        let mut req = MonoTypeMap::new();
        req.insert("A".to_string(), MonoType::Var(Tvar(0)));
        req.insert("B".to_string(), MonoType::Var(Tvar(1)));
        let expr = MonoType::from(types::Function {
            req,
            opt: MonoTypeMap::new(),
            pipe: None,
            retn: MonoType::Var(Tvar(0)),
        });
        let want = types::PolyType { vars, cons, expr };
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_polytype_2() {
        let type_exp = Parser::new("(A: T, B: S) => T where T: Addable").parse_type_expression();

        let got = convert_polytype(type_exp, &mut sub::Substitution::default()).unwrap();
        let mut vars = Vec::<types::Tvar>::new();
        vars.push(types::Tvar(0));
        vars.push(types::Tvar(1));
        let mut cons = types::TvarKinds::new();
        let mut kind_vector_1 = Vec::<types::Kind>::new();
        kind_vector_1.push(types::Kind::Addable);
        cons.insert(types::Tvar(0), kind_vector_1);

        let mut req = MonoTypeMap::new();
        req.insert("A".to_string(), MonoType::Var(Tvar(0)));
        req.insert("B".to_string(), MonoType::Var(Tvar(1)));
        let expr = MonoType::from(types::Function {
            req,
            opt: MonoTypeMap::new(),
            pipe: None,
            retn: MonoType::Var(Tvar(0)),
        });
        let want = types::PolyType { vars, cons, expr };
        assert_eq!(want, got);
    }
}
