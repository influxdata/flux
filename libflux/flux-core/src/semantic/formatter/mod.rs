//! Semantic graph formatter.
#![cfg_attr(feature = "strict", allow(warnings))]

use anyhow::{anyhow, Error, Result};
use chrono::SecondsFormat;

use crate::{
    ast, semantic,
    semantic::{
        types::{CollectionType, MonoType, PolyType, Tvar, TvarKinds},
        walk,
    },
};

#[cfg(test)]
mod tests;

/// Format a Package.
pub fn convert_to_string(pkg: &semantic::nodes::Package) -> Result<String, Error> {
    let mut formatter = Formatter::default();
    formatter.format_package(pkg);
    formatter.output()
}

/// Format a semantic graph
pub fn format(pkg: &semantic::nodes::Package) -> Result<String, Error> {
    convert_to_string(pkg)
}

/// Format a `Node`
pub fn format_node<'a>(node: impl Into<walk::Node<'a>>) -> Result<String, Error> {
    let mut formatter = Formatter::default();
    formatter.format_node(&node.into());
    formatter.output()
}

/// Format a `MonoType`
pub fn format_monotype(typ: &MonoType) -> String {
    let arena = Arena::default();
    let mut formatter = DocFormatter { arena: &arena };
    let doc = formatter.format_monotype(typ);
    doc.pretty(120).to_string()
}

/// Struct to hold data related to formatting such as formatted code,
/// options, and errors.
/// Provides methods for formatting files and strings of source code.
#[derive(Default)]
pub struct Formatter {
    builder: String,
    indentation: u32,
    err: Option<Error>,
}

// INDENT_BYTES is 4 spaces as a constant byte slice
const INDENT_BYTES: &str = "    ";

impl Formatter {
    /// Returns the final formatted string and error message.
    pub fn output(self) -> Result<String, Error> {
        if let Some(err) = self.err {
            return Err(err);
        }

        Ok(self.builder)
    }

    fn write_string(&mut self, s: &str) {
        (&mut self.builder).push_str(s);
    }

    fn write_rune(&mut self, c: char) {
        (&mut self.builder).push(c);
    }

    fn write_indent(&mut self) {
        for _ in 0..self.indentation {
            (&mut self.builder).push_str(INDENT_BYTES);
        }
    }
    fn indent(&mut self) {
        self.indentation += 1;
    }

    fn unindent(&mut self) {
        self.indentation -= 1;
    }

    fn set_indent(&mut self, i: u32) {
        self.indentation = i;
    }

    /// Format a file.
    pub fn format_file(&mut self, n: &semantic::nodes::File, include_pkg: bool) {
        let sep = '\n';
        if let Some(pkg) = &n.package {
            if include_pkg && !pkg.name.name.is_empty() {
                self.write_indent();
                self.format_node(&walk::Node::PackageClause(pkg));
                if !n.imports.is_empty() || !n.body.is_empty() {
                    self.write_rune(sep);
                    self.write_rune(sep)
                }
            }
        }
        for (i, value) in n.imports.iter().enumerate() {
            if i != 0 {
                self.write_rune(sep)
            }
            self.write_indent();
            self.format_import_declaration(value)
        }
        if !n.imports.is_empty() && !n.body.is_empty() {
            self.write_rune(sep);
            self.write_rune(sep);
        }
        // format the file statements
        self.format_statement_list(&n.body);
    }

    fn format_node(&mut self, n: &walk::Node) {
        // save current indentation
        let curr_ind = self.indentation;
        match n {
            walk::Node::File(m) => self.format_file(m, true),
            walk::Node::Block(m) => self.format_block(m),
            walk::Node::ExprStmt(m) => self.format_expression_statement(m),
            walk::Node::PackageClause(m) => self.format_package_clause(m),
            walk::Node::ImportDeclaration(m) => self.format_import_declaration(m),
            walk::Node::ReturnStmt(m) => self.format_return_statement(m),
            walk::Node::OptionStmt(m) => self.format_option_statement(m),
            walk::Node::TestStmt(m) => self.format_test_statement(m),
            walk::Node::TestCaseStmt(m) => self.format_testcase_statement(m),
            walk::Node::VariableAssgn(m) => self.format_variable_assignment(m),
            walk::Node::IndexExpr(m) => self.format_index_expression(m),
            walk::Node::MemberAssgn(m) => self.format_member_assignment(m),
            walk::Node::CallExpr(m) => self.format_call_expression(m),
            walk::Node::ConditionalExpr(m) => self.format_conditional_expression(m),
            walk::Node::StringExpr(m) => self.format_string_expression(m),
            walk::Node::ArrayExpr(m) => self.format_array_expression(m),
            walk::Node::DictExpr(m) => self.format_dict_expression(m),
            walk::Node::MemberExpr(m) => self.format_member_expression(m),
            walk::Node::UnaryExpr(m) => self.format_unary_expression(m),
            walk::Node::BinaryExpr(m) => self.format_binary_expression(m),
            walk::Node::LogicalExpr(m) => self.format_logical_expression(m),
            walk::Node::FunctionExpr(m) => self.format_function_expression(m),
            walk::Node::IdentifierExpr(m) => self.format_identifier_expression(m),
            walk::Node::Property(m) => self.format_property(m),
            walk::Node::TextPart(m) => self.format_text_part(m),
            walk::Node::InterpolatedPart(m) => self.format_interpolated_part(m),
            walk::Node::StringLit(m) => self.format_string_literal(m),
            walk::Node::BooleanLit(m) => self.format_boolean_literal(m),
            walk::Node::FloatLit(m) => self.format_float_literal(m),
            walk::Node::IntegerLit(m) => self.format_integer_literal(m),
            walk::Node::UintLit(m) => self.format_unsigned_integer_literal(m),
            walk::Node::RegexpLit(m) => self.format_regexp_literal(m),
            walk::Node::DurationLit(m) => self.format_duration_literal(m),
            walk::Node::DateTimeLit(m) => self.format_date_time_literal(m),
            walk::Node::Identifier(m) => self.format_identifier(m),
            walk::Node::ObjectExpr(m) => self.format_record_expression_braces(m, true),
            walk::Node::Package(m) => self.format_package(m),
            walk::Node::BuiltinStmt(m) => self.format_builtin(m),
            walk::Node::Expr(m) => self.format_node(&walk::Node::reduce_expr(m)),
            _ => self.err = Some(anyhow!(format!("bad expression: {:?}", n))),
        }
        self.set_indent(curr_ind)
    }

    fn format_package(&mut self, n: &semantic::nodes::Package) {
        let pkg_name = &n.package;
        self.format_package_clause(&semantic::nodes::PackageClause {
            name: semantic::nodes::Identifier {
                name: semantic::nodes::Symbol::from(pkg_name.as_str()),
                loc: ast::SourceLocation::default(),
            },
            loc: ast::SourceLocation::default(),
        });
        for (i, file) in n.files.iter().enumerate() {
            if i != 0 {
                self.write_rune('\n');
                self.write_rune('\n');
            }
            self.format_file(file, false)
        }
    }

    fn format_monotype(&mut self, n: &MonoType) {
        match n {
            MonoType::Var(tv) => self.format_tvar(tv),
            MonoType::BoundVar(tv) => self.format_bound_tvar(tv),
            MonoType::Collection(app) => match app.collection {
                CollectionType::Array => self.format_array_type(&app.arg),
                // Collection::Vector => self.format_vector_type(vec),
                _ => self.err = Some(anyhow!("bad expression")),
            },
            MonoType::Dict(dict) => self.format_dict_type(dict),
            MonoType::Record(rec) => self.format_record_type(rec),
            MonoType::Fun(fun) => self.format_function_type(fun),
            _ => self.err = Some(anyhow!("bad expression")),
        }
    }

    fn format_builtin(&mut self, n: &semantic::nodes::BuiltinStmt) {
        self.write_string("builtin ");
        self.format_identifier(&n.id);
        self.write_string(": ");
        self.format_type_expression(&n.typ_expr);
    }

    fn format_type_expression(&mut self, n: &PolyType) {
        self.format_monotype(&n.expr);
        if !n.vars.is_empty() {
            let multiline = n.vars.len() > 4;
            self.write_string(" where");
            if multiline {
                self.write_rune('\n');
                self.indent();
                self.write_indent();
            } else {
                self.write_rune(' ');
            }
            let sep = match multiline {
                true => ",\n",
                false => ", ",
            };
            for (i, c) in (&n.vars).iter().enumerate() {
                if i != 0 {
                    self.write_string(sep);
                    if multiline {
                        self.write_indent();
                    }
                }
                self.write_string(&format!("{}", c));
            }
            if multiline {
                self.unindent();
            }
        }
    }

    fn format_bound_tvar(&mut self, n: &semantic::types::BoundTvar) {
        self.write_string(&format!("{}", &n));
    }

    fn format_tvar(&mut self, n: &semantic::types::Tvar) {
        self.write_string(&format!("{}", &n));
    }

    fn format_property_type(&mut self, n: &semantic::types::Property) {
        self.write_string(&n.k.to_string());
        self.write_string(": ");
        self.format_monotype(&n.v);
    }

    fn format_dict_type(&mut self, n: &semantic::types::Dictionary) {
        self.write_rune('[');
        self.format_monotype(&n.key);
        self.write_rune(':');
        self.format_monotype(&n.val);
        self.write_rune(']');
    }

    fn format_array_type(&mut self, n: &MonoType) {
        self.write_rune('[');
        self.format_monotype(n);
        self.write_rune(']');
    }

    fn format_kinds(&mut self, n: &[semantic::nodes::Identifier]) {
        self.format_identifier(&n[0]);
        for k in &n[1..] {
            self.write_string(" + ");
            self.format_identifier(k);
        }
    }

    fn format_record_type(&mut self, n: &semantic::types::Record) {
        self.write_string((format!("{}", n)).as_str());
    }

    fn format_function_type(&mut self, n: &semantic::types::Function) {
        self.write_string((format!("{}", n)).as_str());
    }

    fn format_string_expression(&mut self, n: &semantic::nodes::StringExpr) {
        self.write_rune('"');
        for p in &n.parts {
            self.format_string_expression_part(p)
        }
        self.write_rune('"');
    }

    fn format_string_expression_part(&mut self, n: &semantic::nodes::StringExprPart) {
        match n {
            semantic::nodes::StringExprPart::Text(p) => self.format_text_part(p),
            semantic::nodes::StringExprPart::Interpolated(p) => self.format_interpolated_part(p),
        }
    }

    fn format_property(&mut self, n: &semantic::nodes::Property) {
        self.format_identifier(&n.key);
        self.write_string(": ");
        self.format_node(&walk::Node::from_expr(&n.value));
    }

    fn format_text_part(&mut self, n: &semantic::nodes::TextPart) {
        let escaped_string = escape_string(&n.value);
        self.write_string(&escaped_string);
    }

    fn format_interpolated_part(&mut self, n: &semantic::nodes::InterpolatedPart) {
        self.write_string("${");
        self.format_node(&walk::Node::from_expr(&n.expression));
        self.write_rune('}')
    }

    fn format_array_expression(&mut self, n: &semantic::nodes::ArrayExpr) {
        let multiline = n.elements.len() > 4 || n.loc.is_multiline();
        self.write_rune('[');
        if multiline {
            self.write_rune('\n');
            self.indent();
            self.write_indent();
        }
        let sep = match multiline {
            true => ",\n",
            false => ", ",
        };
        for (i, item) in (&n.elements).iter().enumerate() {
            if i != 0 {
                self.write_string(sep);
                if multiline {
                    self.write_indent()
                }
            }
            self.format_node(&walk::Node::from_expr(item));
        }
        if multiline {
            self.write_string(sep);
            self.unindent();
            self.write_indent();
        }
        self.write_rune(']');
        self.write_string(&format!(":{}", &n.typ));
    }

    fn format_dict_expression(&mut self, n: &semantic::nodes::DictExpr) {
        let multiline = n.elements.len() > 4 || n.loc.is_multiline();
        self.write_rune('[');
        if multiline {
            self.write_rune('\n');
            self.indent();
            self.write_indent();
        }
        let sep = match multiline {
            true => ",\n",
            false => ", ",
        };
        if !n.elements.is_empty() {
            for (i, item) in (&n.elements).iter().enumerate() {
                if i != 0 {
                    self.write_string(sep);
                    if multiline {
                        self.write_indent()
                    }
                }
                self.format_node(&walk::Node::from_expr(&item.0));
                self.write_rune(':');
                self.write_rune(' ');
                self.format_node(&walk::Node::from_expr(&item.1));
            }
        } else {
            self.write_rune(':');
        }
        if multiline {
            self.write_string(sep);
            self.unindent();
            self.write_indent();
        }
        self.write_rune(']');
        self.write_string(&format!(":{}", &n.typ));
    }

    fn format_index_expression(&mut self, n: &semantic::nodes::IndexExpr) {
        self.format_child_with_parens(walk::Node::IndexExpr(n), walk::Node::from_expr(&n.array));
        self.write_rune('[');
        self.format_node(&walk::Node::from_expr(&n.index));
        self.write_rune(']');
        self.write_string(&format!(":{}", &n.typ));
    }

    fn format_identifier_expression(&mut self, n: &semantic::nodes::IdentifierExpr) {
        self.write_string(&n.name);
        self.write_string(&format!(":{}", &n.typ));
    }

    fn format_statement_list(&mut self, n: &[semantic::nodes::Statement]) {
        let sep = '\n';
        for (i, stmt) in n.iter().enumerate() {
            if i != 0 {
                self.write_rune(sep);
            }
            self.write_indent();
            self.format_node(&walk::Node::from_stmt(stmt));
        }
    }

    fn format_return_statement(&mut self, n: &semantic::nodes::ReturnStmt) {
        self.write_string("return ");
        self.format_node(&walk::Node::from_expr(&n.argument));
    }

    fn format_option_statement(&mut self, n: &semantic::nodes::OptionStmt) {
        self.write_string("option ");
        self.format_assignment(&n.assignment);
    }

    fn format_test_statement(&mut self, n: &semantic::nodes::TestStmt) {
        self.write_string("test ");
        self.format_node(&walk::Node::VariableAssgn(&n.assignment));
    }

    fn format_testcase_statement(&mut self, n: &semantic::nodes::TestCaseStmt) {
        self.write_string("testcase ");
        self.format_node(&walk::Node::Identifier(&n.id));
        self.write_rune(' ');
        // format the testcase statements
        self.format_statement_list(&n.body);
    }

    fn format_assignment(&mut self, n: &semantic::nodes::Assignment) {
        match &n {
            semantic::nodes::Assignment::Variable(m) => {
                self.format_node(&walk::Node::VariableAssgn(m));
            }
            semantic::nodes::Assignment::Member(m) => {
                self.format_node(&walk::Node::MemberAssgn(m));
            }
        }
    }

    // format_child_with_parens applies the generic rule for parenthesis (not for binary expressions).
    fn format_child_with_parens(&mut self, parent: walk::Node, child: walk::Node) {
        self.format_left_child_with_parens(&parent, &child)
    }

    // format_right_child_with_parens applies the generic rule for parenthesis to the right child of a binary expression.
    fn format_right_child_with_parens(&mut self, parent: &walk::Node, child: &walk::Node) {
        let pvp = get_precedence(parent);
        let pvc = get_precedence(child);
        if needs_parenthesis(pvp, pvc, true) {
            self.format_node_with_parens(child);
        } else {
            self.format_node(child);
        }
    }

    // format_left_child_with_parens applies the generic rule for parenthesis to the left child of a binary expression.
    fn format_left_child_with_parens(&mut self, parent: &walk::Node, child: &walk::Node) {
        let pvp = get_precedence(parent);
        let pvc = get_precedence(child);
        if needs_parenthesis(pvp, pvc, false) {
            self.format_node_with_parens(child);
        } else {
            self.format_node(child);
        }
    }

    #[allow(clippy::branches_sharing_code)]
    fn format_node_with_parens(&mut self, node: &walk::Node) {
        self.write_rune('(');
        self.format_node(node);
        self.write_rune(')')
    }

    fn format_member_expression(&mut self, n: &semantic::nodes::MemberExpr) {
        self.format_child_with_parens(walk::Node::MemberExpr(n), walk::Node::from_expr(&n.object));
        self.write_rune('.');
        self.write_string(&n.property);
        self.write_string(&format!(":{}", &n.typ));
    }

    fn format_record_expression_as_function_argument(&mut self, n: &semantic::nodes::ObjectExpr) {
        let i = self.indentation;
        self.format_record_expression_braces(n, false);
        self.set_indent(i);
    }

    fn format_record_expression_braces(&mut self, n: &semantic::nodes::ObjectExpr, braces: bool) {
        let multiline = n.properties.len() > 4 || n.loc.is_multiline();
        if braces {
            self.write_rune('{');
        }
        if let Some(with) = &n.with {
            self.format_identifier_expression(with);
            self.write_string(" with");
            if !multiline {
                self.write_rune(' ');
            }
        }
        if multiline {
            self.write_rune('\n');
            self.indent();
            self.write_indent();
        }
        let sep = match multiline {
            true => ",\n",
            false => ", ",
        };
        for (i, property) in (&n.properties).iter().enumerate() {
            if i != 0 {
                self.write_string(sep);
                if multiline {
                    self.write_indent()
                }
            }
            self.format_node(&walk::Node::Property(property));
        }
        if multiline {
            self.write_string(sep);
            self.unindent();
            self.write_indent();
        }
        if braces {
            self.write_rune('}');
        }
        self.write_string(&format!(":{}", &n.typ));
    }

    fn format_function_expression(&mut self, n: &semantic::nodes::FunctionExpr) {
        let multiline = n.params.len() > 4 && n.loc.is_multiline();
        self.write_rune('(');
        let sep;
        if multiline && n.params.len() > 1 {
            self.indent();
            sep = ",\n";
            self.write_string("\n");
            self.indent();
            self.write_indent();
        } else {
            sep = ", ";
        }
        for (i, function_parameter) in (&n.params).iter().enumerate() {
            if i != 0 {
                self.write_string(sep);
                if multiline {
                    self.write_indent();
                }
            }
            // treat properties differently than in general case
            self.format_function_argument(function_parameter);
        }
        if multiline {
            self.unindent();
            self.unindent();
            self.write_string(sep);
        }
        self.write_string(") ");
        self.write_string("=>");
        self.write_rune(' ');
        self.format_block(&n.body);
        self.write_string(&format!(":{}", &n.typ));
    }

    fn format_function_argument(&mut self, n: &semantic::nodes::FunctionParameter) {
        self.format_identifier(&n.key);
        if let Some(v) = &n.default {
            self.write_rune('=');
            self.format_node(&walk::Node::from_expr(v));
        }
    }

    fn format_block(&mut self, n: &semantic::nodes::Block) {
        self.write_rune('{');
        let sep = '\n';
        self.indent();
        self.write_rune(sep);
        let mut current = n;
        let mut multiline = false;

        loop {
            match current {
                semantic::nodes::Block::Variable(assign, next) => {
                    self.write_indent();
                    self.format_variable_assignment(assign.as_ref());
                    multiline = true;
                    current = next.as_ref();
                }
                semantic::nodes::Block::Expr(expr_stmt, next) => {
                    self.write_indent();
                    self.format_expression_statement(expr_stmt);
                    multiline = true;
                    current = next.as_ref();
                }
                semantic::nodes::Block::Return(ret) => {
                    if multiline {
                        self.write_rune(sep);
                    }
                    self.write_indent();
                    self.format_return_statement(ret);
                    break;
                }
            }
        }
        self.write_rune(sep);
        self.unindent();
        self.write_indent();
        self.write_rune('}');
    }

    fn format_identifier(&mut self, n: &semantic::nodes::Identifier) {
        self.write_string(&n.name);
    }

    fn format_variable_assignment(&mut self, n: &semantic::nodes::VariableAssgn) {
        self.format_node(&walk::Node::Identifier(&n.id));
        self.write_string(" = ");
        self.format_node(&walk::Node::from_expr(&n.init));
    }

    fn format_call_expression(&mut self, n: &semantic::nodes::CallExpr) {
        self.format_child_with_parens(walk::Node::CallExpr(n), walk::Node::from_expr(&n.callee));
        self.write_rune('(');
        let sep = ", ";
        for (i, c) in n.arguments.iter().enumerate() {
            if i != 0 {
                self.write_string(sep);
            }
            self.format_property(c);
        }
        self.write_rune(')');
        self.write_string(&format!(":{}", &n.typ));
    }

    fn format_conditional_expression(&mut self, n: &semantic::nodes::ConditionalExpr) {
        let multiline = n.loc.is_multiline();
        let nested = matches!(&n.alternate, semantic::nodes::Expression::Conditional(_));
        self.write_rune('(');
        self.write_string("if ");
        self.format_node(&walk::Node::from_expr(&n.test));
        self.write_string(" then");
        if multiline {
            self.write_rune('\n');
            self.indent();
            self.write_indent();
        } else {
            self.write_rune(' ');
        }
        self.format_node(&walk::Node::from_expr(&n.consequent));
        if multiline {
            self.write_rune('\n');
            self.unindent();
            self.write_indent();
        } else {
            self.write_rune(' ');
        }
        self.write_string("else");
        if multiline && !nested {
            self.write_rune('\n');
            self.indent();
            self.write_indent();
        } else {
            self.write_rune(' ');
        }
        self.format_node(&walk::Node::from_expr(&n.alternate));
        if multiline && !nested {
            self.unindent();
        }
        self.write_rune(')');
        self.write_string(&format!(":{}", &n.typ));
    }

    fn format_member_assignment(&mut self, n: &semantic::nodes::MemberAssgn) {
        self.format_node(&walk::Node::MemberExpr(&n.member));
        self.write_string(" = ");
        self.format_node(&walk::Node::from_expr(&n.init));
    }

    fn format_unary_expression(&mut self, n: &semantic::nodes::UnaryExpr) {
        self.write_string(&n.operator.to_string());
        match n.operator {
            ast::Operator::SubtractionOperator => {}
            ast::Operator::AdditionOperator => {}
            _ => {
                self.write_rune(' ');
            }
        }
        self.format_child_with_parens(walk::Node::UnaryExpr(n), walk::Node::from_expr(&n.argument));
        self.write_string(&format!(":{}", &n.typ));
    }

    fn format_binary_expression(&mut self, n: &semantic::nodes::BinaryExpr) {
        self.format_binary(
            &n.operator.to_string(),
            walk::Node::BinaryExpr(n),
            walk::Node::from_expr(&n.left),
            walk::Node::from_expr(&n.right),
            &n.typ,
        );
    }

    fn format_logical_expression(&mut self, n: &semantic::nodes::LogicalExpr) {
        self.format_binary(
            &n.operator.to_string(),
            walk::Node::LogicalExpr(n),
            walk::Node::from_expr(&n.left),
            walk::Node::from_expr(&n.right),
            &MonoType::BOOL,
        );
    }

    fn format_binary(
        &mut self,
        op: &str,
        parent: walk::Node,
        left: walk::Node,
        right: walk::Node,
        typ: &MonoType,
    ) {
        self.format_left_child_with_parens(&parent, &left);
        self.write_rune(' ');
        self.write_string(&format!("{}:{}", op, typ));
        self.write_rune(' ');
        self.format_right_child_with_parens(&parent, &right);
    }

    fn format_import_declaration(&mut self, n: &semantic::nodes::ImportDeclaration) {
        self.write_string("import ");
        if let Some(alias) = &n.alias {
            if !alias.name.is_empty() {
                self.format_node(&walk::Node::Identifier(alias));
                self.write_rune(' ')
            }
        }
        self.format_node(&walk::Node::StringLit(&n.path))
    }

    fn format_expression_statement(&mut self, n: &semantic::nodes::ExprStmt) {
        self.format_node(&walk::Node::from_expr(&n.expression))
    }

    fn format_package_clause(&mut self, n: &semantic::nodes::PackageClause) {
        self.write_string("package ");
        self.format_node(&walk::Node::Identifier(&n.name));
        self.write_rune('\n');
    }

    fn format_string_literal(&mut self, n: &semantic::nodes::StringLit) {
        if let Some(src) = &n.loc.source {
            if !src.is_empty() {
                // Preserve the exact literal if we have it
                self.write_string(src);
                // self.write_string(&format!(":{}", MonoType::String.to_string()));
                return;
            }
        }
        // Write out escaped string value
        self.write_rune('"');
        let escaped_string = escape_string(&n.value);
        self.write_string(&escaped_string);
        self.write_rune('"');
        // self.write_string(&format!(":{}", MonoType::String.to_string()));
    }

    fn format_boolean_literal(&mut self, n: &semantic::nodes::BooleanLit) {
        let s: &str = if n.value { "true" } else { "false" };
        self.write_string(s);
        // self.write_string(&format!(":{}", MonoType::Bool.to_string()));
    }

    fn format_date_time_literal(&mut self, n: &semantic::nodes::DateTimeLit) {
        let mut f: String;
        let v = &n.value;
        let nano_sec = v.timestamp_subsec_nanos();
        if nano_sec > 0 {
            f = v.format("%FT%T").to_string();
            let mut frac_nano: String = v.format("%f").to_string();
            frac_nano.insert(0, '.');
            let mut r = frac_nano.chars().last().unwrap();
            while r == '0' {
                frac_nano.pop();
                r = frac_nano.chars().last().unwrap();
            }
            f.push_str(&frac_nano);

            if v.timezone().local_minus_utc() == 0 {
                f.push('Z')
            } else {
                f.push_str(&v.format("%:z").to_string());
            }
        } else {
            f = v.to_rfc3339_opts(SecondsFormat::Secs, true)
        }
        self.write_string(&f);
        // self.write_string(&format!(":{}", MonoType::Time.to_string()));
    }

    fn format_duration_literal(&mut self, n: &semantic::nodes::DurationLit) {
        // format negative sign
        if n.value.negative {
            self.write_string("-");
        }

        // format months
        let mut inp_months = n.value.months;
        if inp_months > 0 {
            let years = inp_months / 12;
            if years > 0 {
                self.write_string(&format!("{}y", years));
            }
            let months = inp_months % 12;
            if months > 0 {
                self.write_string(&format!("{}mo", months));
            }
        }

        // format nanoseconds
        let mut inp_nsecs = n.value.nanoseconds;
        if inp_nsecs > 0 {
            let nsecs = inp_nsecs % 1000;
            inp_nsecs /= 1000;
            let usecs = inp_nsecs % 1000;
            inp_nsecs /= 1000;
            let msecs = inp_nsecs % 1000;
            inp_nsecs /= 1000;
            let secs = inp_nsecs % 60;
            inp_nsecs /= 60;
            let mins = inp_nsecs % 60;
            inp_nsecs /= 60;
            let hours = inp_nsecs % 24;
            inp_nsecs /= 24;
            let days = inp_nsecs % 7;
            inp_nsecs /= 7;
            let weeks = inp_nsecs;

            if weeks > 0 {
                self.write_string(&format!("{}w", weeks));
            }
            if days > 0 {
                self.write_string(&format!("{}d", days));
            }
            if hours > 0 {
                self.write_string(&format!("{}h", hours));
            }
            if mins > 0 {
                self.write_string(&format!("{}m", mins));
            }
            if secs > 0 {
                self.write_string(&format!("{}s", secs));
            }
            if msecs > 0 {
                self.write_string(&format!("{}ms", msecs));
            }
            if usecs > 0 {
                self.write_string(&format!("{}us", usecs));
            }
            if nsecs > 0 {
                self.write_string(&format!("{}ns", nsecs));
            }
        }
        // self.write_string(&format!(":{}", MonoType::Duration.to_string()));
    }

    fn format_float_literal(&mut self, n: &semantic::nodes::FloatLit) {
        let mut s = format!("{}", n.value);
        if !s.contains('.') {
            s.push_str(".0");
        }
        // s.push_str(&format!(":{}", MonoType::Float.to_string()));
        self.write_string(&s)
    }

    fn format_integer_literal(&mut self, n: &semantic::nodes::IntegerLit) {
        self.write_string(&format!("{}", n.value));
        // self.write_string(&format!(":{}", MonoType::Int.to_string()));
    }

    fn format_unsigned_integer_literal(&mut self, n: &semantic::nodes::UintLit) {
        self.write_string(&format!("{0:10}", n.value));
        // self.write_string(&format!(":{}", MonoType::Uint.to_string()));
    }

    fn format_regexp_literal(&mut self, n: &semantic::nodes::RegexpLit) {
        self.write_rune('/');
        self.write_string(&n.value.replace('/', "\\/"));
        self.write_rune('/');
        // self.write_string(&format!(":{}", MonoType::Regexp.to_string()));
    }
}

use pretty::{docs, DocAllocator};

use crate::formatter::comma_list_with;

type Arena<'doc> = pretty::Arena<'doc>;
type Doc<'doc> = pretty::DocBuilder<'doc, Arena<'doc>, ()>;

const MULTILINE: usize = 4;
const INDENT: isize = INDENT_BYTES.len() as isize;

struct DocFormatter<'doc> {
    arena: &'doc Arena<'doc>,
}

impl<'doc> DocFormatter<'doc> {
    fn base_multiline(&self, base: &ast::BaseNode) -> Doc<'doc> {
        self.multiline(base.is_multiline())
    }

    fn multiline(&self, multiline: bool) -> Doc<'doc> {
        if multiline {
            self.arena.hardline()
        } else {
            self.arena.line()
        }
    }

    fn base_multiline_(&self, base: &ast::BaseNode) -> Doc<'doc> {
        self.multiline_(base.is_multiline())
    }

    fn multiline_(&self, multiline: bool) -> Doc<'doc> {
        if multiline {
            self.arena.hardline()
        } else {
            self.arena.line_()
        }
    }

    fn format_monotype(&self, n: &'doc MonoType) -> Doc<'doc> {
        let arena = self.arena;
        match n {
            MonoType::Error => arena.text("<error>"),
            MonoType::Var(tv) => docs![arena, "#", tv.to_string()],
            MonoType::BoundVar(tv) => arena.text(tv.to_string()),
            MonoType::Builtin(nt) => arena.text(nt.to_string()),
            MonoType::Collection(col) => match col.collection {
                CollectionType::Array => {
                    docs![arena, "[", self.format_monotype(&col.arg), "]",]
                }
                CollectionType::Vector => {
                    docs![arena, "v[", self.format_monotype(&col.arg), "]",]
                }
                CollectionType::Stream => {
                    docs![arena, "stream[", self.format_monotype(&col.arg), "]",]
                }
            },
            MonoType::Dict(dict) => {
                docs![
                    arena,
                    "[",
                    self.format_monotype(&dict.key),
                    ":",
                    self.format_monotype(&dict.val),
                    "]",
                ]
            }
            MonoType::Record(n) => {
                let multiline = n.fields().count() > MULTILINE;
                let line = self.multiline(multiline);
                let line_ = self.multiline_(multiline);

                let mut fields = n.fields();

                let fields_doc = comma_list_with(
                    arena,
                    fields.by_ref().map(|p| {
                        docs![arena, p.k.to_string(), ": ", self.format_monotype(&p.v),].group()
                    }),
                    line,
                );

                docs![
                    arena,
                    "{",
                    docs![
                        arena,
                        line_.clone(),
                        if let Some(typ) = fields.tail() {
                            docs![
                                arena,
                                docs![arena, self.format_monotype(typ), arena.line(), "with",]
                                    .group(),
                                arena.line(),
                            ]
                        } else {
                            arena.nil()
                        },
                        fields_doc,
                    ]
                    .nest(INDENT),
                    line_,
                    "}",
                ]
            }
            MonoType::Fun(n) => {
                let multiline = n.parameters_len() > MULTILINE;
                let line = self.multiline(multiline);
                let line_ = self.multiline_(multiline);

                docs![
                    arena,
                    "(",
                    docs![
                        arena,
                        line_.clone(),
                        comma_list_with(
                            arena,
                            n.pipe
                                .iter()
                                .map(|p| {
                                    docs![
                                        arena,
                                        if p.k == "<-" {
                                            docs![arena, &p.k]
                                        } else {
                                            docs![arena, "<-", &p.k]
                                        },
                                        ": ",
                                        self.format_monotype(&p.v),
                                    ]
                                })
                                .chain(n.req.iter().map(|(k, v)| {
                                    docs![arena, k.as_str(), ": ", self.format_monotype(v),]
                                }))
                                .chain(n.opt.iter().map(|(name, argument)| {
                                    docs![
                                        arena,
                                        "?",
                                        name.as_str(),
                                        ": ",
                                        self.format_monotype(&argument.typ),
                                        match &argument.default {
                                            Some(default) =>
                                                docs![arena, " = ", self.format_monotype(default)],
                                            None => arena.nil(),
                                        }
                                    ]
                                })),
                            line,
                        ),
                    ]
                    .nest(INDENT),
                    line_.clone(),
                    ")",
                    " => ",
                    self.format_monotype(&n.retn),
                ]
            }
            MonoType::Label(label) => self.format_string_literal(label),
        }
        .group()
    }

    fn format_string_literal(&self, value: &str) -> Doc<'doc> {
        let arena = self.arena;
        docs![arena, "\"", escape_string(value), "\""]
    }
}

fn escape_string(s: &str) -> String {
    if !(s.contains('\"') || s.contains('\\')) {
        return s.to_string();
    }
    let mut escaped = String::with_capacity(s.len() * 2);
    for r in s.chars() {
        if r == '"' || r == '\\' {
            escaped.push('\\')
        }
        escaped.push(r)
    }
    escaped
}

fn get_precedence(node: &walk::Node) -> u32 {
    match node {
        walk::Node::BinaryExpr(p) => Operator::new(&p.operator).get_precedence(),
        walk::Node::LogicalExpr(p) => Operator::new_logical(&p.operator).get_precedence(),
        walk::Node::UnaryExpr(p) => Operator::new(&p.operator).get_precedence(),
        walk::Node::FunctionExpr(_) => 3,
        walk::Node::CallExpr(_) => 1,
        walk::Node::MemberExpr(_) => 1,
        walk::Node::IndexExpr(_) => 1,
        walk::Node::ConditionalExpr(_) => 11,
        walk::Node::Expr(e) => get_precedence(&walk::Node::reduce_expr(e)),
        _ => 0,
    }
}

struct Operator<'a> {
    op: Option<&'a ast::Operator>,
    l_op: Option<&'a ast::LogicalOperator>,
    is_logical: bool,
}

impl<'a> Operator<'a> {
    fn new(op: &ast::Operator) -> Operator {
        Operator {
            op: Some(op),
            l_op: None,
            is_logical: false,
        }
    }

    fn new_logical(op: &ast::LogicalOperator) -> Operator {
        Operator {
            op: None,
            l_op: Some(op),
            is_logical: true,
        }
    }

    fn get_precedence(&self) -> u32 {
        if !self.is_logical {
            return match self.op.unwrap() {
                ast::Operator::PowerOperator => 4,
                ast::Operator::MultiplicationOperator => 5,
                ast::Operator::DivisionOperator => 5,
                ast::Operator::ModuloOperator => 5,
                ast::Operator::AdditionOperator => 6,
                ast::Operator::SubtractionOperator => 6,
                ast::Operator::LessThanEqualOperator => 7,
                ast::Operator::LessThanOperator => 7,
                ast::Operator::GreaterThanEqualOperator => 7,
                ast::Operator::GreaterThanOperator => 7,
                ast::Operator::StartsWithOperator => 7,
                ast::Operator::InOperator => 7,
                ast::Operator::NotEmptyOperator => 7,
                ast::Operator::EmptyOperator => 7,
                ast::Operator::EqualOperator => 7,
                ast::Operator::NotEqualOperator => 7,
                ast::Operator::RegexpMatchOperator => 7,
                ast::Operator::NotRegexpMatchOperator => 7,
                ast::Operator::NotOperator => 8,
                ast::Operator::ExistsOperator => 8,
                ast::Operator::InvalidOperator => 0,
            };
        }
        match self.l_op.unwrap() {
            ast::LogicalOperator::AndOperator => 9,
            ast::LogicalOperator::OrOperator => 10,
        }
    }
}

fn needs_parenthesis(pvp: u32, pvc: u32, is_right: bool) -> bool {
    // If one of the precedence values is invalid, then we shouldn't apply any parenthesis.
    let par = pvc != 0 && pvp != 0;
    par && ((!is_right && pvc > pvp) || (is_right && pvc >= pvp))
}
