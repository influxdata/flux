//! Various conversions from AST nodes to their associated
//! types in the semantic graph.

use std::{
    collections::{BTreeMap, HashMap},
    fmt,
    sync::Arc,
};

use codespan_reporting::diagnostic;
use serde::{Serialize, Serializer};
use thiserror::Error;

use crate::{
    ast,
    errors::{located, AsDiagnostic, Errors, Located},
    semantic::{
        env::Environment,
        nodes::*,
        types::{self, BuiltinType, MonoType, MonoTypeMap, SemanticMap},
        AnalyzerConfig, Feature,
    },
};

/// Error that categorizes errors when converting from AST to semantic graph.
pub type Error = Located<ErrorKind>;

/// Error that categorizes errors when converting from AST to semantic graph.
#[derive(Error, Debug, PartialEq)]
#[allow(missing_docs)]
pub enum ErrorKind {
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
    pkg: &ast::Package,
    env: &Environment,
    config: &AnalyzerConfig,
) -> Result<Package, Errors<Error>> {
    let mut converter = Converter::with_env(env, config);
    let r = converter.convert_package(pkg);
    converter.finish()?;
    Ok(r)
}

/// Converts a [type expression] in the AST into a [`PolyType`].
///
/// [type expression]: ast::TypeExpression
/// [`PolyType`]: types::PolyType
pub fn convert_polytype(
    type_expression: &ast::TypeExpression,
    config: &AnalyzerConfig,
) -> Result<types::PolyType, Errors<Error>> {
    let mut converter = Converter::new(config);
    let r = converter.convert_polytype(type_expression);
    converter.finish()?;
    Ok(r)
}

#[cfg(test)]
pub(crate) fn convert_monotype(
    ty: &ast::MonoType,
    tvars: &mut BTreeMap<String, types::BoundTvar>,
    config: &AnalyzerConfig,
) -> Result<MonoType, Errors<Error>> {
    let mut converter = Converter::new(config);
    let r = converter.convert_monotype(&ty, tvars);
    converter.finish()?;
    Ok(r)
}

#[allow(missing_docs)]
#[derive(Clone)]
pub struct Symbol {
    name: Arc<str>,
}

impl fmt::Debug for Symbol {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{:?}#{}", self.name.as_ptr(), &self.name)
    }
}

impl PartialEq for Symbol {
    fn eq(&self, other: &Self) -> bool {
        Arc::ptr_eq(&self.name, &other.name)
    }
}

impl Eq for Symbol {}

impl std::hash::Hash for Symbol {
    fn hash<H: std::hash::Hasher>(&self, hasher: &mut H) {
        self.name.as_ptr().hash(hasher)
    }
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

impl PartialEq<String> for Symbol {
    fn eq(&self, other: &String) -> bool {
        &self[..] == other
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
struct SymbolStack {
    symbols: Vec<BTreeMap<String, Symbol>>,
}

impl SymbolStack {
    fn get(&mut self, name: &str) -> Option<&Symbol> {
        self.symbols
            .iter()
            .rev()
            .find_map(|symbols| symbols.get(name))
    }

    fn insert(&mut self, package: Option<&str>, name: &str) -> Symbol {
        let symbol = Symbol::from(match package {
            Some(package) => format!("{}@{}", name, package),
            None => name.to_owned(),
        });
        self.symbols
            .last_mut()
            .unwrap()
            .insert(name.to_owned(), symbol.clone());
        symbol
    }

    fn enter_scope(&mut self) {
        self.symbols.push(Default::default());
    }

    fn exit_scope(&mut self) {
        match self.symbols.pop() {
            Some(_) => (),
            None => panic!("cannot pop final stack frame from symbols"),
        }
    }
}

#[allow(missing_docs)]
#[derive(Debug)]
pub struct SymbolInfo {
    pub comments: Vec<String>,
}

#[allow(missing_docs)]
pub type PackageInfo = HashMap<Symbol, SymbolInfo>;

fn get_attribute<'a>(comments: impl IntoIterator<Item = &'a str>, attr: &str) -> Option<&'a str> {
    comments.into_iter().find_map(|comment| {
        // Remove the comment and any preceding whitespace
        let comment = comment.trim_start_matches("//").trim_start();
        if let Some(content) = comment.strip_prefix('@') {
            let mut iter = content.splitn(2, char::is_whitespace);
            let name = iter.next().unwrap();
            if name == attr {
                Some(iter.next().unwrap_or("").trim())
            } else {
                None
            }
        } else {
            None
        }
    })
}

#[derive(Debug, Default)]
struct Symbols<'a> {
    symbols: SymbolStack,
    package_info: PackageInfo,
    local_labels: BTreeMap<String, Symbol>,
    env: Option<&'a Environment<'a>>,
}

impl<'a> Symbols<'a> {
    fn new(env: Option<&'a Environment>) -> Self {
        Symbols {
            env,
            package_info: Default::default(),
            local_labels: Default::default(),
            symbols: SymbolStack::default(),
        }
    }

    fn insert(&mut self, package: Option<&str>, name: &str, comments: &[ast::Comment]) -> Symbol {
        let symbol = self.symbols.insert(package, name);

        let symbol_info = SymbolInfo {
            comments: comments.iter().map(|c| c.text.clone()).collect(),
        };

        self.package_info.insert(symbol.clone(), symbol_info);

        if package.is_none() && !self.local_labels.contains_key(&symbol[..]) {
            self.local_labels.insert(symbol.to_string(), symbol.clone());
        }

        symbol
    }

    /// Property keys don't rely on `Symbol` equality so we can use a single `Symbol` for all
    /// properties of the same name in a single package
    fn lookup_property_key(&mut self, name: &str) -> Symbol {
        if let Some(symbol) = self.local_labels.get(name).cloned() {
            symbol
        } else {
            let symbol = Symbol::from(name);
            self.local_labels.insert(name.to_string(), symbol.clone());
            symbol
        }
    }

    fn lookup(&mut self, name: &str) -> Symbol {
        self.lookup_option(name).unwrap_or_else(|| {
            // Use the same symbol for every unbound variable
            self.local_labels
                .entry(name.into())
                .or_insert_with(|| Symbol::from(name))
                .clone()
        })
    }

    fn lookup_option(&mut self, name: &str) -> Option<Symbol> {
        self.symbols
            .get(name)
            .or_else(|| self.env.and_then(|env| env.lookup_symbol(name)))
            .cloned()
    }

    fn enter_scope(&mut self) {
        self.symbols.enter_scope()
    }

    fn exit_scope(&mut self) {
        self.symbols.exit_scope()
    }
}

pub(crate) struct Converter<'a> {
    symbols: Symbols<'a>,
    errors: Errors<Error>,
    config: &'a AnalyzerConfig,
}

impl<'a> Converter<'a> {
    fn new(config: &'a AnalyzerConfig) -> Self {
        Converter {
            symbols: Symbols::new(None),
            errors: Errors::new(),
            config,
        }
    }

    pub(crate) fn with_env(env: &'a Environment, config: &'a AnalyzerConfig) -> Self {
        Converter {
            symbols: Symbols::new(Some(env)),
            errors: Errors::new(),
            config,
        }
    }

    pub(crate) fn finish(self) -> Result<(), Errors<Error>> {
        if self.errors.has_errors() {
            Err(self.errors)
        } else {
            Ok(())
        }
    }

    pub(crate) fn take_package_info(&mut self) -> PackageInfo {
        std::mem::take(&mut self.symbols.package_info)
    }

    pub(crate) fn convert_package(&mut self, pkg: &ast::Package) -> Package {
        let package = pkg.package.clone();

        self.symbols.enter_scope();

        let files = pkg
            .files
            .iter()
            .map(|file| self.convert_file(&package, file))
            .collect::<Vec<File>>();

        self.symbols.exit_scope();

        Package {
            loc: pkg.base.location.clone(),
            package,
            files,
        }
    }

    fn convert_file(&mut self, package_name: &str, file: &ast::File) -> File {
        let package = self.convert_package_clause(file.package.as_ref());
        let imports = file
            .imports
            .iter()
            .map(|i| self.convert_import_declaration(i))
            .collect::<Vec<ImportDeclaration>>();
        let body = self.convert_statements(package_name, &file.body);

        File {
            loc: file.base.location.clone(),
            package,
            imports,
            body,
        }
    }

    fn convert_package_clause(
        &mut self,
        pkg: Option<&ast::PackageClause>,
    ) -> Option<PackageClause> {
        let pkg = pkg?;
        let name = self.convert_identifier(&pkg.name);
        Some(PackageClause {
            loc: pkg.base.location.clone(),
            name,
        })
    }

    fn convert_import_declaration(&mut self, imp: &ast::ImportDeclaration) -> ImportDeclaration {
        let path = &imp.path.value;
        let (import_symbol, alias) = match &imp.alias {
            None => {
                let name = path.rsplit_once('/').map_or(&path[..], |t| t.1);
                (self.symbols.insert(None, name, &[]), None)
            }
            Some(id) => {
                let id = self.define_identifier(None, id, &imp.base.comments);
                (id.name.clone(), Some(id))
            }
        };
        let path = self.convert_string_literal(&imp.path);

        ImportDeclaration {
            loc: imp.base.location.clone(),
            alias,
            path,
            import_symbol,
        }
    }

    fn convert_statements(&mut self, package: &str, stmts: &[ast::Statement]) -> Vec<Statement> {
        stmts
            .iter()
            .filter_map(|s| self.convert_statement(package, s))
            .collect::<Vec<_>>()
    }

    fn convert_statement(&mut self, package: &str, stmt: &ast::Statement) -> Option<Statement> {
        Some(match stmt {
            ast::Statement::Option(s) => {
                Statement::Option(Box::new(self.convert_option_statement(s)))
            }
            ast::Statement::Builtin(s) => {
                Statement::Builtin(self.convert_builtin_statement(package, s)?)
            }
            ast::Statement::TestCase(s) => {
                Statement::TestCase(Box::new(self.convert_testcase(package, s)))
            }
            ast::Statement::Expr(s) => Statement::Expr(self.convert_expression_statement(s)),
            ast::Statement::Return(s) => Statement::Return(self.convert_return_statement(s)),
            // TODO(affo): we should fix this to include MemberAssignement.
            //  The error lies in AST: the Statement enum does not include that.
            //  This is not a problem when parsing, because we parse it only in the option assignment case,
            //  and we return an OptionStmt, which is a Statement.
            ast::Statement::Variable(s) => {
                Statement::Variable(Box::new(self.convert_variable_assignment(Some(package), s)))
            }
            ast::Statement::Bad(s) => Statement::Error(BadStmt {
                loc: s.base.location.clone(),
            }),
        })
    }

    fn convert_assignment(&mut self, assign: &ast::Assignment) -> Assignment {
        match assign {
            ast::Assignment::Variable(a) => {
                Assignment::Variable(self.convert_variable_assignment(None, a))
            }
            ast::Assignment::Member(a) => Assignment::Member(self.convert_member_assignment(a)),
        }
    }

    fn convert_option_statement(&mut self, stmt: &ast::OptionStmt) -> OptionStmt {
        OptionStmt {
            loc: stmt.base.location.clone(),
            assignment: self.convert_assignment(&stmt.assignment),
        }
    }

    fn convert_builtin_statement(
        &mut self,
        package: &str,
        stmt: &ast::BuiltinStmt,
    ) -> Option<BuiltinStmt> {
        // Only include builtin statements that have the `feature` attribute if a matching
        // feature is detected
        let opt_attr = get_attribute(
            stmt.base.comments.iter().map(|c| c.text.as_str()),
            "feature",
        )
        .and_then(|attr| attr.parse::<Feature>().ok());
        if let Some(attr) = opt_attr {
            if self.config.features.iter().all(|feature| *feature != attr) {
                return None;
            }
        }

        Some(BuiltinStmt {
            loc: stmt.base.location.clone(),
            id: self.define_identifier(Some(package), &stmt.id, &stmt.base.comments),
            typ_expr: self.convert_polytype(&stmt.ty),
        })
    }
    fn convert_testcase(&mut self, package: &str, stmt: &ast::TestCaseStmt) -> TestCaseStmt {
        TestCaseStmt {
            loc: stmt.base.location.clone(),
            id: self.convert_identifier(&stmt.id),
            extends: stmt
                .extends
                .as_ref()
                .map(|e| self.convert_string_literal(e)),
            body: self.convert_statements(package, &stmt.block.body),
        }
    }

    fn convert_builtintype(&mut self, basic: &ast::NamedType) -> Result<BuiltinType> {
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
                    basic.base.location.clone(),
                    ErrorKind::InvalidNamedType(basic.name.name.to_string()),
                ))
            }
        })
    }

    fn convert_tvar(
        &mut self,
        tv: &ast::Identifier,
        tvars: &mut BTreeMap<String, types::BoundTvar>,
    ) -> types::BoundTvar {
        // Variables `A` to `Z` may be typed in the signatures of builtins and `Tvar(0)` to `Tvar(25)`
        // are displayed as those letters so we consistently map these names to those variables
        let next_var = if tv.name.len() == 1 && tv.name.as_str() >= "A" && tv.name.as_str() <= "Z" {
            types::BoundTvar(tv.name.chars().next().unwrap() as u64 - 'A' as u64)
        } else {
            types::BoundTvar(tvars.len() as u64 + ('Z' as u64 - 'A' as u64))
        };
        *tvars.entry(tv.name.clone()).or_insert_with(|| next_var)
    }

    fn convert_monotype(
        &mut self,
        ty: &ast::MonoType,
        tvars: &mut BTreeMap<String, types::BoundTvar>,
    ) -> MonoType {
        match ty {
            ast::MonoType::Tvar(tv) => {
                let tvar = self.convert_tvar(&tv.name, tvars);
                MonoType::BoundVar(tvar)
            }

            ast::MonoType::Basic(basic) => match self.convert_builtintype(basic) {
                Ok(builtin) => MonoType::from(builtin),
                Err(err) => {
                    self.errors.push(err);
                    MonoType::Error
                }
            },
            ast::MonoType::Array(arr) => MonoType::arr(self.convert_monotype(&arr.element, tvars)),
            ast::MonoType::Stream(stream) => {
                MonoType::stream(self.convert_monotype(&stream.element, tvars))
            }
            ast::MonoType::Vector(vector) => {
                MonoType::vector(self.convert_monotype(&vector.element, tvars))
            }
            ast::MonoType::Dict(dict) => {
                let key = self.convert_monotype(&dict.key, tvars);
                let val = self.convert_monotype(&dict.val, tvars);
                MonoType::from(types::Dictionary { key, val })
            }
            ast::MonoType::Function(func) => {
                let mut req = MonoTypeMap::new();
                let mut opt = MonoTypeMap::new();
                let mut _pipe = None;
                let mut dirty = false;
                for param in &func.parameters {
                    match param {
                        ast::ParameterType::Required { name, monotype, .. } => {
                            req.insert(name.name.clone(), self.convert_monotype(monotype, tvars));
                        }
                        ast::ParameterType::Optional {
                            name,
                            monotype,
                            default,
                            ..
                        } => {
                            opt.insert(
                                name.name.clone(),
                                types::Argument {
                                    typ: self.convert_monotype(monotype, tvars),
                                    default: default.as_ref().map(|default| {
                                        MonoType::Label(types::Label::from(default.value.as_str()))
                                    }),
                                },
                            );
                        }
                        ast::ParameterType::Pipe {
                            name,
                            monotype,
                            base,
                        } => {
                            if !dirty {
                                _pipe = Some(types::Property {
                                    k: match name {
                                        Some(n) => n.name.clone(),
                                        None => String::from("<-"),
                                    },
                                    v: self.convert_monotype(monotype, tvars),
                                });
                                dirty = true;
                            } else {
                                self.errors
                                    .push(located(base.location.clone(), ErrorKind::AtMostOnePipe));
                            }
                        }
                    }
                }
                MonoType::from(types::Function {
                    req,
                    opt,
                    pipe: _pipe,
                    retn: self.convert_monotype(&func.monotype, tvars),
                })
            }
            ast::MonoType::Record(rec) => {
                let mut r = match &rec.tvar {
                    None => MonoType::from(types::Record::Empty),
                    Some(id) => {
                        let tv = ast::MonoType::Tvar(ast::TvarType {
                            base: id.clone().base,
                            name: id.clone(),
                        });
                        self.convert_monotype(&tv, tvars)
                    }
                };
                for prop in &rec.properties {
                    let property = types::Property {
                        k: match &prop.name {
                            ast::PropertyKey::Identifier(id) => {
                                if id.name.len() == 1 && id.name.starts_with(char::is_uppercase) {
                                    let tvar = self.convert_tvar(id, tvars);
                                    types::RecordLabel::BoundVariable(tvar)
                                } else {
                                    types::Label::from(self.symbols.lookup(&id.name)).into()
                                }
                            }
                            ast::PropertyKey::StringLit(lit) => {
                                types::Label::from(self.symbols.lookup(&lit.value)).into()
                            }
                        },
                        v: self.convert_monotype(&prop.monotype, tvars),
                    };
                    r = MonoType::from(types::Record::Extension {
                        head: property,
                        tail: r,
                    })
                }
                r
            }

            ast::MonoType::Label(string_lit) => {
                MonoType::Label(types::Label::from(string_lit.value.as_str()))
            }
        }
    }

    // [`PolyType`]: types::PolyType
    fn convert_polytype(&mut self, type_expression: &ast::TypeExpression) -> types::PolyType {
        let mut tvars = BTreeMap::<String, types::BoundTvar>::new();
        let expr = self.convert_monotype(&type_expression.monotype, &mut tvars);
        let mut vars = Vec::<types::BoundTvar>::new();
        let mut cons = SemanticMap::<types::BoundTvar, Vec<types::Kind>>::new();

        for (name, tvar) in tvars {
            vars.push(tvar);
            let mut kinds = Vec::<types::Kind>::new();
            for con in &type_expression.constraints {
                if con.tvar.name == name {
                    for k in &con.kinds {
                        match k.name.parse() {
                            Ok(kind) => kinds.push(kind),
                            Err(()) => {
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
        types::PolyType { vars, cons, expr }
    }

    fn convert_expression_statement(&mut self, stmt: &ast::ExprStmt) -> ExprStmt {
        ExprStmt {
            loc: stmt.base.location.clone(),
            expression: self.convert_expression(&stmt.expression),
        }
    }

    fn convert_return_statement(&mut self, stmt: &ast::ReturnStmt) -> ReturnStmt {
        ReturnStmt {
            loc: stmt.base.location.clone(),
            argument: self.convert_expression(&stmt.argument),
        }
    }

    fn convert_variable_assignment(
        &mut self,
        package: Option<&str>,
        stmt: &ast::VariableAssgn,
    ) -> VariableAssgn {
        let expr = self.convert_expression(&stmt.init);
        VariableAssgn::new(
            self.define_identifier(package, &stmt.id, &stmt.id.base.comments),
            expr,
            stmt.base.location.clone(),
        )
    }

    fn convert_member_assignment(&mut self, stmt: &ast::MemberAssgn) -> MemberAssgn {
        let init = self.convert_expression(&stmt.init);
        MemberAssgn {
            loc: stmt.base.location.clone(),
            member: self.convert_member_expression(&stmt.member),
            init,
        }
    }

    fn convert_expression(&mut self, expr: &ast::Expression) -> Expression {
        match expr {
            ast::Expression::Function(expr) => {
                Expression::Function(Box::new(self.convert_function_expression(expr)))
            }
            ast::Expression::Call(expr) => {
                Expression::Call(Box::new(self.convert_call_expression(expr)))
            }
            ast::Expression::Member(expr) => {
                Expression::Member(Box::new(self.convert_member_expression(expr)))
            }
            ast::Expression::Index(expr) => {
                Expression::Index(Box::new(self.convert_index_expression(expr)))
            }
            ast::Expression::PipeExpr(expr) => {
                Expression::Call(Box::new(self.convert_pipe_expression(expr)))
            }
            ast::Expression::Binary(expr) => {
                Expression::Binary(Box::new(self.convert_binary_expression(expr)))
            }
            ast::Expression::Unary(expr) => {
                Expression::Unary(Box::new(self.convert_unary_expression(expr)))
            }
            ast::Expression::Logical(expr) => {
                Expression::Logical(Box::new(self.convert_logical_expression(expr)))
            }
            ast::Expression::Conditional(expr) => {
                Expression::Conditional(Box::new(self.convert_conditional_expression(expr)))
            }
            ast::Expression::Object(expr) => {
                Expression::Object(Box::new(self.convert_object_expression(expr)))
            }
            ast::Expression::Array(expr) => {
                Expression::Array(Box::new(self.convert_array_expression(expr)))
            }
            ast::Expression::Dict(expr) => {
                Expression::Dict(Box::new(self.convert_dict_expression(expr)))
            }
            ast::Expression::Identifier(expr) => {
                Expression::Identifier(self.convert_identifier_expression(expr))
            }
            ast::Expression::StringExpr(expr) => {
                Expression::StringExpr(Box::new(self.convert_string_expression(expr)))
            }
            ast::Expression::Paren(expr) => self.convert_expression(&expr.expression),
            ast::Expression::StringLit(lit) => {
                Expression::StringLit(self.convert_string_literal(lit))
            }
            ast::Expression::Boolean(lit) => Expression::Boolean(self.convert_boolean_literal(lit)),
            ast::Expression::Float(lit) => Expression::Float(self.convert_float_literal(lit)),
            ast::Expression::Integer(lit) => Expression::Integer(self.convert_integer_literal(lit)),
            ast::Expression::Uint(lit) => {
                Expression::Uint(self.convert_unsigned_integer_literal(lit))
            }
            ast::Expression::Regexp(lit) => Expression::Regexp(self.convert_regexp_literal(lit)),
            ast::Expression::Duration(lit) => {
                let location = lit.base.location.clone();
                match self.convert_duration_literal(lit) {
                    Ok(d) => Expression::Duration(d),
                    Err(err) => {
                        self.errors.push(err);
                        Expression::Error(BadExpr { loc: location })
                    }
                }
            }
            ast::Expression::DateTime(lit) => {
                Expression::DateTime(self.convert_date_time_literal(lit))
            }
            ast::Expression::PipeLit(lit) => {
                self.errors.push(located(
                    lit.base.location.clone(),
                    ErrorKind::InvalidPipeLit,
                ));

                Expression::Error(BadExpr {
                    loc: lit.base.location.clone(),
                })
            }
            ast::Expression::Bad(bad) => Expression::Error(BadExpr {
                loc: bad.base.location.clone(),
            }),
        }
    }

    fn convert_function_expression(&mut self, expr: &ast::FunctionExpr) -> FunctionExpr {
        self.symbols.enter_scope();

        let params = self.convert_function_params(&expr.params);
        let body = self.convert_function_body(&expr.body);

        self.symbols.exit_scope();

        FunctionExpr {
            loc: expr.base.location.clone(),
            typ: MonoType::Error,
            params,
            body,
            vectorized: None,
        }
    }

    fn convert_function_params(&mut self, props: &[ast::Property]) -> Vec<FunctionParameter> {
        // The defaults must be converted first so that the parameters are not in scope
        let mut piped = false;
        enum Default {
            Expr(Expression),
            Piped,
            None,
        }
        let defaults: Vec<_> = props
            .iter()
            .map(|prop| {
                if let Some(expr) = &prop.value {
                    match expr {
                        ast::Expression::PipeLit(lit) => {
                            if piped {
                                self.errors.push(located(
                                    lit.base.location.clone(),
                                    ErrorKind::AtMostOnePipe,
                                ));
                            } else {
                                piped = true;
                            }
                            Default::Piped
                        }
                        e => Default::Expr(self.convert_expression(e)),
                    }
                } else {
                    Default::None
                }
            })
            .collect();

        // The iteration here is complex, cannot use iter().map()..., better to write it explicitly.
        let mut params: Vec<FunctionParameter> = Vec::new();
        for (prop, default) in props.iter().zip(defaults) {
            let id = match &prop.key {
                ast::PropertyKey::Identifier(id) => id,
                _ => {
                    self.errors.push(located(
                        prop.base.location.clone(),
                        ErrorKind::FunctionParameterIdents,
                    ));
                    continue;
                }
            };
            let key = self.define_identifier(None, id, &id.base.comments);

            let (is_pipe, default) = match default {
                Default::Expr(expr) => (false, Some(expr)),
                Default::Piped => (true, None),
                Default::None => (false, None),
            };

            params.push(FunctionParameter {
                loc: prop.base.location.clone(),
                is_pipe,
                key,
                default,
            });
        }
        params
    }

    fn convert_function_body(&mut self, body: &ast::FunctionBody) -> Block {
        match body {
            ast::FunctionBody::Expr(expr) => {
                let argument = self.convert_expression(expr);
                Block::Return(ReturnStmt {
                    loc: argument.loc().clone(),
                    argument,
                })
            }
            ast::FunctionBody::Block(block) => self.convert_block(block),
        }
    }

    fn convert_block(&mut self, block: &ast::Block) -> Block {
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
        for s in &block.body {
            match s {
                ast::Statement::Variable(dec) => body.push(TempBlock::Variable(Box::new(
                    self.convert_variable_assignment(None, dec),
                ))),
                ast::Statement::Expr(stmt) => {
                    body.push(TempBlock::Expr(self.convert_expression_statement(stmt)))
                }
                ast::Statement::Return(stmt) => {
                    let argument = self.convert_expression(&stmt.argument);
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
                    argument: Expression::Error(BadExpr {
                        loc: s.loc().clone(),
                    }),
                })
            }
            None => {
                self.errors.push(located(
                    block.base.location.clone(),
                    ErrorKind::MissingReturn,
                ));
                Block::Return(ReturnStmt {
                    loc: block.base.location.clone(),
                    argument: Expression::Error(BadExpr {
                        loc: block.base.location.clone(),
                    }),
                })
            }
        };

        body.fold(block, |acc, s| match s {
            TempBlock::Variable(dec) => Block::Variable(dec, Box::new(acc)),
            TempBlock::Expr(stmt) => Block::Expr(stmt, Box::new(acc)),
            TempBlock::Return(s) => {
                self.errors.push(located(
                    s.loc,
                    ErrorKind::InvalidFunctionStatement("return"),
                ));
                acc
            }
        })
    }

    fn convert_call_expression(&mut self, expr: &ast::CallExpr) -> CallExpr {
        let callee = self.convert_expression(&expr.callee);
        // TODO(affo): I'd prefer these checks to be in ast.Check().
        let mut args = expr
            .arguments
            .iter()
            .map(|a| match a {
                ast::Expression::Object(obj) => self.convert_object_expression(obj),
                _ => {
                    self.errors.push(located(
                        a.base().location.clone(),
                        ErrorKind::ParametersNotRecord,
                    ));

                    ObjectExpr {
                        loc: a.base().location.clone(),
                        typ: MonoType::Error,
                        with: None,
                        properties: Vec::new(),
                    }
                }
            })
            .collect::<Vec<ObjectExpr>>();
        let arguments = match args.len() {
            0 => Vec::new(),
            1 => args.pop().expect("there must be 1 element").properties,
            _ => {
                self.errors.push(located(
                    expr.base.location.clone(),
                    ErrorKind::ExtraParameterRecord,
                ));
                args.remove(0).properties
            }
        };
        CallExpr {
            loc: expr.base.location.clone(),
            typ: MonoType::Error,
            callee,
            arguments,
            pipe: None,
        }
    }

    fn convert_member_expression(&mut self, expr: &ast::MemberExpr) -> MemberExpr {
        let object = self.convert_expression(&expr.object);
        let property = match &expr.property {
            ast::PropertyKey::Identifier(id) => &id.name,
            ast::PropertyKey::StringLit(lit) => &lit.value,
        };
        let property = self.symbols.lookup_property_key(property);
        MemberExpr {
            loc: expr.base.location.clone(),
            typ: MonoType::Error,
            object,
            property,
        }
    }

    fn convert_index_expression(&mut self, expr: &ast::IndexExpr) -> IndexExpr {
        let array = self.convert_expression(&expr.array);
        let index = self.convert_expression(&expr.index);
        IndexExpr {
            loc: expr.base.location.clone(),
            typ: MonoType::Error,
            array,
            index,
        }
    }

    fn convert_pipe_expression(&mut self, expr: &ast::PipeExpr) -> CallExpr {
        let mut call = self.convert_call_expression(&expr.call);
        let pipe = self.convert_expression(&expr.argument);
        call.pipe = Some(pipe);
        call
    }

    fn convert_binary_expression(&mut self, expr: &ast::BinaryExpr) -> BinaryExpr {
        let left = self.convert_expression(&expr.left);
        let right = self.convert_expression(&expr.right);
        BinaryExpr {
            loc: expr.base.location.clone(),
            typ: MonoType::Error,
            operator: expr.operator.clone(),
            left,
            right,
        }
    }

    fn convert_unary_expression(&mut self, expr: &ast::UnaryExpr) -> UnaryExpr {
        let argument = self.convert_expression(&expr.argument);
        UnaryExpr {
            loc: expr.base.location.clone(),
            typ: MonoType::Error,
            operator: expr.operator.clone(),
            argument,
        }
    }

    fn convert_logical_expression(&mut self, expr: &ast::LogicalExpr) -> LogicalExpr {
        let left = self.convert_expression(&expr.left);
        let right = self.convert_expression(&expr.right);
        LogicalExpr {
            loc: expr.base.location.clone(),
            typ: MonoType::BOOL,
            operator: expr.operator.clone(),
            left,
            right,
        }
    }

    fn convert_conditional_expression(&mut self, expr: &ast::ConditionalExpr) -> ConditionalExpr {
        let test = self.convert_expression(&expr.test);
        let consequent = self.convert_expression(&expr.consequent);
        let alternate = self.convert_expression(&expr.alternate);
        ConditionalExpr {
            loc: expr.base.location.clone(),
            test,
            consequent,
            alternate,
            typ: MonoType::Error,
        }
    }

    fn convert_object_expression(&mut self, expr: &ast::ObjectExpr) -> ObjectExpr {
        let properties = expr
            .properties
            .iter()
            .map(|p| self.convert_property(p))
            .collect::<Vec<Property>>();
        let with = expr
            .with
            .as_ref()
            .map(|with| self.convert_identifier_expression(&with.source));
        ObjectExpr {
            loc: expr.base.location.clone(),
            typ: MonoType::Error,
            with,
            properties,
        }
    }

    fn convert_property(&mut self, prop: &ast::Property) -> Property {
        let key = match &prop.key {
            ast::PropertyKey::Identifier(id) => self.convert_property_key(id),
            ast::PropertyKey::StringLit(lit) => {
                let loc = lit.base.location.clone();
                let name = self.convert_string_literal(lit).value;
                Identifier {
                    name: self.symbols.lookup_property_key(&name),
                    loc,
                }
            }
        };
        let value = match &prop.value {
            Some(expr) => self.convert_expression(expr),
            None => Expression::Identifier(IdentifierExpr {
                loc: key.loc.clone(),
                typ: MonoType::Error,
                name: self
                    .symbols
                    .lookup_option(&key.name)
                    .unwrap_or_else(|| key.name.clone()),
            }),
        };
        Property {
            loc: prop.base.location.clone(),
            key,
            value,
        }
    }

    fn convert_array_expression(&mut self, expr: &ast::ArrayExpr) -> ArrayExpr {
        let elements = expr
            .elements
            .iter()
            .map(|e| self.convert_expression(&e.expression))
            .collect::<Vec<Expression>>();
        ArrayExpr {
            loc: expr.base.location.clone(),
            typ: MonoType::Error,
            elements,
        }
    }

    fn convert_dict_expression(&mut self, expr: &ast::DictExpr) -> DictExpr {
        let mut elements = Vec::new();
        for item in &expr.elements {
            elements.push((
                self.convert_expression(&item.key),
                self.convert_expression(&item.val),
            ));
        }
        DictExpr {
            loc: expr.base.location.clone(),
            typ: MonoType::Error,
            elements,
        }
    }

    fn define_identifier(
        &mut self,
        package: Option<&str>,
        id: &ast::Identifier,
        comments: &[ast::Comment],
    ) -> Identifier {
        let name = self.symbols.insert(package, &id.name, comments);
        Identifier {
            loc: id.base.location.clone(),
            name,
        }
    }

    fn convert_property_key(&mut self, id: &ast::Identifier) -> Identifier {
        Identifier {
            name: self.symbols.lookup_property_key(&id.name),
            loc: id.base.location.clone(),
        }
    }

    fn convert_identifier(&mut self, id: &ast::Identifier) -> Identifier {
        Identifier {
            name: self.symbols.lookup(&id.name),
            loc: id.base.location.clone(),
        }
    }

    fn convert_identifier_expression(&mut self, id: &ast::Identifier) -> IdentifierExpr {
        IdentifierExpr {
            typ: MonoType::Error,
            name: self.symbols.lookup(&id.name),
            loc: id.base.location.clone(),
        }
    }

    fn convert_string_expression(&mut self, expr: &ast::StringExpr) -> StringExpr {
        let parts = expr
            .parts
            .iter()
            .map(|p| self.convert_string_expression_part(p))
            .collect::<Vec<StringExprPart>>();
        StringExpr {
            loc: expr.base.location.clone(),
            parts,
        }
    }

    fn convert_string_expression_part(&mut self, expr: &ast::StringExprPart) -> StringExprPart {
        match expr {
            ast::StringExprPart::Text(txt) => StringExprPart::Text(TextPart {
                loc: txt.base.location.clone(),
                value: txt.value.clone(),
            }),
            ast::StringExprPart::Interpolated(itp) => {
                StringExprPart::Interpolated(InterpolatedPart {
                    loc: itp.base.location.clone(),
                    expression: self.convert_expression(&itp.expression),
                })
            }
        }
    }

    fn convert_string_literal(&mut self, lit: &ast::StringLit) -> StringLit {
        StringLit {
            loc: lit.base.location.clone(),
            value: lit.value.clone(),
            typ: None,
        }
    }

    fn convert_boolean_literal(&mut self, lit: &ast::BooleanLit) -> BooleanLit {
        BooleanLit {
            loc: lit.base.location.clone(),
            value: lit.value,
        }
    }

    fn convert_float_literal(&mut self, lit: &ast::FloatLit) -> FloatLit {
        FloatLit {
            loc: lit.base.location.clone(),
            value: lit.value,
        }
    }

    fn convert_integer_literal(&mut self, lit: &ast::IntegerLit) -> IntegerLit {
        IntegerLit {
            loc: lit.base.location.clone(),
            value: lit.value,
        }
    }

    fn convert_unsigned_integer_literal(&mut self, lit: &ast::UintLit) -> UintLit {
        UintLit {
            loc: lit.base.location.clone(),
            value: lit.value,
        }
    }

    fn convert_regexp_literal(&mut self, lit: &ast::RegexpLit) -> RegexpLit {
        RegexpLit {
            loc: lit.base.location.clone(),
            value: lit.value.clone(),
        }
    }

    fn convert_duration_literal(&mut self, lit: &ast::DurationLit) -> Result<DurationLit> {
        Ok(DurationLit {
            value: convert_duration(&lit.values).map_err(|e| {
                located(
                    lit.base.location.clone(),
                    ErrorKind::InvalidDuration(e.to_string()),
                )
            })?,
            loc: lit.base.location.clone(),
        })
    }

    fn convert_date_time_literal(&mut self, lit: &ast::DateTimeLit) -> DateTimeLit {
        DateTimeLit {
            loc: lit.base.location.clone(),
            value: lit.value,
        }
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
            types::{BoundTvar, MonoType, Tvar},
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
        let config = AnalyzerConfig::default();
        let mut converter = Converter::new(&config);
        let mut pkg = converter.convert_package(&pkg);
        converter.finish()?;

        // We don't want to specifc the exact locations for each node in the tests
        walk_mut(
            &mut |n: &mut NodeMut| n.set_loc(ast::BaseNode::default().location),
            NodeMut::Package(&mut pkg),
        );

        Ok(pkg)
    }

    fn collect_symbols(pkg: &Package) -> BTreeMap<String, Symbol> {
        use crate::semantic::walk;

        let mut map = BTreeMap::new();

        walk::walk(
            &mut |node| {
                let symbol = match node {
                    walk::Node::Identifier(id) => &id.name,
                    walk::Node::ImportDeclaration(import) => &import.import_symbol,
                    walk::Node::IdentifierExpr(id) => &id.name,
                    walk::Node::MemberExpr(member) => &member.property,
                    _ => return,
                };

                if !map.contains_key(symbol.full_name()) {
                    map.insert(symbol.full_name().to_string(), symbol.clone());
                }
            },
            walk::Node::Package(pkg),
        );

        map
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

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

        let want = Package {
            loc: b.location.clone(),
            package: "foo".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: Some(PackageClause {
                    loc: b.location.clone(),
                    name: Identifier {
                        loc: b.location.clone(),
                        name: symbols["foo"].clone(),
                    },
                }),
                imports: Vec::new(),
                body: Vec::new(),
            }],
        };
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_imports() {
        let b = ast::BaseNode::default();
        let pkg = parse_package(
            r#"package qux
            import "path/foo"
            import b "path/bar"
            "#,
        );

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

        let want = Package {
            loc: b.location.clone(),
            package: "qux".to_string(),
            files: vec![File {
                loc: b.location.clone(),
                package: Some(PackageClause {
                    loc: b.location.clone(),
                    name: Identifier {
                        loc: b.location.clone(),
                        name: symbols["qux"].clone(),
                    },
                }),
                imports: vec![
                    ImportDeclaration {
                        loc: b.location.clone(),
                        path: StringLit {
                            loc: b.location.clone(),
                            value: "path/foo".to_string(),
                            typ: None,
                        },
                        alias: None,
                        import_symbol: symbols["foo"].clone(),
                    },
                    ImportDeclaration {
                        loc: b.location.clone(),
                        path: StringLit {
                            loc: b.location.clone(),
                            value: "path/bar".to_string(),
                            typ: None,
                        },
                        alias: Some(Identifier {
                            loc: b.location.clone(),
                            name: symbols["b"].clone(),
                        }),
                        import_symbol: symbols["b"].clone(),
                    },
                ],
                body: Vec::new(),
            }],
        };
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

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

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
                            name: symbols["a@main"].clone(),
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
                            name: symbols["a@main"].clone(),
                        }),
                    }),
                ],
            }],
        };
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_object() {
        let b = ast::BaseNode::default();
        let pkg = parse_package(r#"{ a: 10 }"#);

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

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
                                name: symbols["a"].clone(),
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
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_object_with_string_key() {
        let b = ast::BaseNode::default();
        let pkg = parse_package(r#"{ "a": 10 }"#);

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

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
                                name: symbols["a"].clone(),
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
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_object_with_mixed_keys() {
        let b = ast::BaseNode::default();
        let pkg = parse_package(r#"{ "a": 10, b: 11 }"#);

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

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
                                    name: symbols["a"].clone(),
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
                                    name: symbols["b"].clone(),
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
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_object_with_implicit_keys() {
        let b = ast::BaseNode::default();
        let pkg = parse_package("{ a, b }");

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

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
                                    name: symbols["a"].clone(),
                                },
                                value: Expression::Identifier(IdentifierExpr {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    name: symbols["a"].clone(),
                                }),
                            },
                            Property {
                                loc: b.location.clone(),
                                key: Identifier {
                                    loc: b.location.clone(),
                                    name: symbols["b"].clone(),
                                },
                                value: Expression::Identifier(IdentifierExpr {
                                    loc: b.location.clone(),
                                    typ: type_info(),
                                    name: symbols["b"].clone(),
                                }),
                            },
                        ],
                    })),
                })],
            }],
        };
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_options_declaration() {
        let b = ast::BaseNode::default();
        let pkg = parse_package(
            r#"option task = { name: "foo", every: 1h, delay: 10m, cron: "0 2 * * *", retry: 5}"#,
        );

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

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
                            name: symbols["task"].clone(),
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
                                        name: symbols["name"].clone(),
                                    },
                                    value: Expression::StringLit(StringLit {
                                        loc: b.location.clone(),
                                        value: "foo".to_string(),
                                        typ: None,
                                    }),
                                },
                                Property {
                                    loc: b.location.clone(),
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: symbols["every"].clone(),
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
                                        name: symbols["delay"].clone(),
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
                                        name: symbols["cron"].clone(),
                                    },
                                    value: Expression::StringLit(StringLit {
                                        loc: b.location.clone(),
                                        value: "0 2 * * *".to_string(),
                                        typ: None,
                                    }),
                                },
                                Property {
                                    loc: b.location.clone(),
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: symbols["retry"].clone(),
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
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_qualified_option_statement() {
        let b = ast::BaseNode::default();
        let pkg = parse_package(r#"option alert.state = "Warning""#);

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

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
                                name: symbols["alert"].clone(),
                            }),
                            property: symbols["state"].clone(),
                        },
                        init: Expression::StringLit(StringLit {
                            loc: b.location.clone(),
                            value: "Warning".to_string(),
                            typ: None,
                        }),
                    }),
                }))],
            }],
        };
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_function() {
        let b = ast::BaseNode::default();
        let pkg = parse_package(
            "f = (a, b) => a + b
            f(a: 2, b: 3)",
        );

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

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
                            name: symbols["f@main"].clone(),
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
                                        name: symbols["a"].clone(),
                                    },
                                    default: None,
                                },
                                FunctionParameter {
                                    loc: b.location.clone(),
                                    is_pipe: false,
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: symbols["b"].clone(),
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
                                        name: symbols["a"].clone(),
                                    }),
                                    right: Expression::Identifier(IdentifierExpr {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        name: symbols["b"].clone(),
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
                                name: symbols["f@main"].clone(),
                            }),
                            arguments: vec![
                                Property {
                                    loc: b.location.clone(),
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: symbols["a"].clone(),
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
                                        name: symbols["b"].clone(),
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
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_function_with_defaults() {
        let b = ast::BaseNode::default();
        let pkg = parse_package(
            "f = (a=0, b=0, c) => a + b + c
            f(c: 42)",
        );

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

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
                            name: symbols["f@main"].clone(),
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
                                        name: symbols["a"].clone(),
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
                                        name: symbols["b"].clone(),
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
                                        name: symbols["c"].clone(),
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
                                            name: symbols["a"].clone(),
                                        }),
                                        right: Expression::Identifier(IdentifierExpr {
                                            loc: b.location.clone(),
                                            typ: type_info(),
                                            name: symbols["b"].clone(),
                                        }),
                                    })),
                                    right: Expression::Identifier(IdentifierExpr {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        name: symbols["c"].clone(),
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
                                name: symbols["f@main"].clone(),
                            }),
                            arguments: vec![Property {
                                loc: b.location.clone(),
                                key: Identifier {
                                    loc: b.location.clone(),
                                    name: symbols["c"].clone(),
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

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

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
                            name: symbols["f@main"].clone(),
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
                                        name: symbols["piped"].clone(),
                                    },
                                    default: None,
                                },
                                FunctionParameter {
                                    loc: b.location.clone(),
                                    is_pipe: false,
                                    key: Identifier {
                                        loc: b.location.clone(),
                                        name: symbols["a"].clone(),
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
                                        name: symbols["a"].clone(),
                                    }),
                                    right: Expression::Identifier(IdentifierExpr {
                                        loc: b.location.clone(),
                                        typ: type_info(),
                                        name: symbols["piped"].clone(),
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
                                name: symbols["f@main"].clone(),
                            }),
                            arguments: vec![Property {
                                loc: b.location.clone(),
                                key: Identifier {
                                    loc: b.location.clone(),
                                    name: symbols["a"].clone(),
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
        assert_eq!(want, got);
    }

    #[test]
    fn test_function_expression_simple() {
        let b = ast::BaseNode::default();
        let symbols = ["a", "b"]
            .into_iter()
            .map(|s| (s.to_string(), Symbol::from(s)))
            .collect::<BTreeMap<_, _>>();
        let f = FunctionExpr {
            loc: b.location.clone(),
            typ: type_info(),
            params: vec![
                FunctionParameter {
                    loc: b.location.clone(),
                    is_pipe: false,
                    key: Identifier {
                        loc: b.location.clone(),
                        name: symbols["a"].clone(),
                    },
                    default: None,
                },
                FunctionParameter {
                    loc: b.location.clone(),
                    is_pipe: false,
                    key: Identifier {
                        loc: b.location.clone(),
                        name: symbols["b"].clone(),
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
                        name: symbols["a"].clone(),
                    }),
                    right: Expression::Identifier(IdentifierExpr {
                        loc: b.location.clone(),
                        typ: type_info(),
                        name: symbols["b"].clone(),
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
        let symbols = ["a", "b", "c", "d"]
            .into_iter()
            .map(|s| (s.to_string(), Symbol::from(s)))
            .collect::<BTreeMap<_, _>>();
        let piped = FunctionParameter {
            loc: b.location.clone(),
            is_pipe: true,
            key: Identifier {
                loc: b.location.clone(),
                name: symbols["a"].clone(),
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
                name: symbols["b"].clone(),
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
                name: symbols["c"].clone(),
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
                name: symbols["d"].clone(),
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
                        name: symbols["a"].clone(),
                    }),
                    right: Expression::Identifier(IdentifierExpr {
                        loc: b.location.clone(),
                        typ: type_info(),
                        name: symbols["b"].clone(),
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

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

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
                            name: symbols["a"].clone(),
                        }),
                        index: Expression::Integer(IntegerLit {
                            loc: b.location.clone(),
                            value: 3,
                        }),
                    })),
                })],
            }],
        };
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_nested_index_expression() {
        let b = ast::BaseNode::default();
        let pkg =
            Parser::new("a[3][5]").parse_single_package("path".to_string(), "foo.flux".to_string());

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

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
                                name: symbols["a"].clone(),
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
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_access_indexed_object_returned_from_function_call() {
        let b = ast::BaseNode::default();
        let pkg =
            Parser::new("f()[3]").parse_single_package("path".to_string(), "foo.flux".to_string());

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

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
                                name: symbols["f"].clone(),
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

        let got = test_convert(pkg).unwrap();
        let symbols = collect_symbols(&got);

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
                                name: symbols["a"].clone(),
                            }),
                            property: symbols["b"].clone(),
                        })),
                        property: symbols["c"].clone(),
                    })),
                })],
            }],
        };
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_member_with_call_expression() {
        let b = ast::BaseNode::default();
        let pkg =
            Parser::new("a.b().c").parse_single_package("path".to_string(), "foo.flux".to_string());
        let got = test_convert(pkg).unwrap();

        let symbols = collect_symbols(&got);

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
                                    name: symbols["a"].clone(),
                                }),
                                property: symbols["b"].clone(),
                            })),
                            arguments: Vec::new(),
                        })),
                        property: symbols["c"].clone(),
                    })),
                })],
            }],
        };
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

    fn convert_monotype_test(
        ty: &ast::MonoType,
        tvars: &mut BTreeMap<String, types::BoundTvar>,
    ) -> Result<MonoType, Errors<Error>> {
        convert_monotype(ty, tvars, &Default::default())
    }

    #[test]
    fn test_convert_monotype_int() {
        let monotype = Parser::new("int").parse_monotype();
        let mut m = BTreeMap::new();
        let got = convert_monotype_test(&monotype, &mut m).unwrap();
        let want = MonoType::INT;
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_monotype_record() {
        let monotype = Parser::new("{ A with b: int }").parse_monotype();

        let mut m = BTreeMap::new();
        let got = convert_monotype_test(&monotype, &mut m).unwrap();
        let want = MonoType::from(types::Record::Extension {
            head: types::Property {
                k: types::RecordLabel::from("b"),
                v: MonoType::INT,
            },
            tail: MonoType::BoundVar(BoundTvar(0)),
        });
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_monotype_function() {
        let monotype_ex = Parser::new("(?A: int) => int").parse_monotype();

        let mut m = BTreeMap::new();
        let got = convert_monotype_test(&monotype_ex, &mut m).unwrap();
        let mut opt = MonoTypeMap::new();
        opt.insert(String::from("A"), MonoType::INT.into());
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
        let got = convert_polytype(&type_exp, &Default::default()).unwrap();

        let want = {
            let mut vars = Vec::<types::BoundTvar>::new();
            vars.push(types::BoundTvar(0));
            vars.push(types::BoundTvar(1));
            let mut cons = types::BoundTvarKinds::new();
            let mut kind_vector_1 = Vec::<types::Kind>::new();
            kind_vector_1.push(types::Kind::Addable);
            cons.insert(types::BoundTvar(0), kind_vector_1);

            let mut kind_vector_2 = Vec::<types::Kind>::new();
            kind_vector_2.push(types::Kind::Divisible);
            cons.insert(types::BoundTvar(1), kind_vector_2);

            let mut req = MonoTypeMap::new();
            req.insert("A".to_string(), MonoType::BoundVar(BoundTvar(0)));
            req.insert("B".to_string(), MonoType::BoundVar(BoundTvar(1)));
            let expr = MonoType::from(types::Function {
                req,
                opt: MonoTypeMap::new(),
                pipe: None,
                retn: MonoType::BoundVar(BoundTvar(0)),
            });
            types::PolyType { vars, cons, expr }
        };
        assert_eq!(want, got);
    }

    #[test]
    fn test_convert_polytype_2() {
        let type_exp = Parser::new("(A: T, B: S) => T where T: Addable").parse_type_expression();

        let got = convert_polytype(&type_exp, &Default::default()).unwrap();
        let mut vars = Vec::<types::BoundTvar>::new();
        vars.push(types::BoundTvar(0));
        vars.push(types::BoundTvar(1));
        let mut cons = types::BoundTvarKinds::new();
        let mut kind_vector_1 = Vec::<types::Kind>::new();
        kind_vector_1.push(types::Kind::Addable);
        cons.insert(types::BoundTvar(0), kind_vector_1);

        let mut req = MonoTypeMap::new();
        req.insert("A".to_string(), MonoType::BoundVar(BoundTvar(0)));
        req.insert("B".to_string(), MonoType::BoundVar(BoundTvar(1)));
        let expr = MonoType::from(types::Function {
            req,
            opt: MonoTypeMap::new(),
            pipe: None,
            retn: MonoType::BoundVar(BoundTvar(0)),
        });
        let want = types::PolyType { vars, cons, expr };
        assert_eq!(want, got);
    }

    #[test]
    fn test_get_attribute() {
        assert_eq!(
            get_attribute(["// @feature labelPolymorphism\n"], "feature"),
            Some("labelPolymorphism"),
        );
    }
}
