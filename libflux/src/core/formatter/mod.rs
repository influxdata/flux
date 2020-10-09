#![allow(missing_docs)]
use crate::ast::{self, walk::Node, File};
use crate::parser::parse_string;
use crate::Error;

use chrono::SecondsFormat;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub fn format_from_js_file(js_file: JsValue) -> String {
    if let Ok(file) = js_file.into_serde::<File>() {
        if let Ok(converted) = convert_to_string(&file) {
            return converted;
        }
    }
    "".to_string()
}

pub fn convert_to_string(file: &File) -> Result<String, String> {
    let mut formatter = Formatter::default();
    formatter.format_file(file, true);
    formatter.output()
}

pub fn format(contents: String) -> Result<String, String> {
    let file = parse_string("", contents.as_str());

    convert_to_string(&file)
}

pub struct Formatter {
    builder: String,
    indentation: u32,
    clear: bool,
    temp_indent: bool,
    err: Option<Error>,
}

impl Default for Formatter {
    fn default() -> Self {
        Formatter {
            builder: String::new(),
            indentation: 0,
            clear: true,
            temp_indent: false,
            err: None,
        }
    }
}

impl Formatter {
    // returns the final string and error msg
    pub fn output(&self) -> Result<String, String> {
        if let Some(err) = &self.err {
            return Err(err.msg.clone());
        }

        Ok(self.builder.clone())
    }

    // Do not use to send newline. This is not (yet) setting clear
    // appropriately.
    fn write_string(&mut self, s: &str) {
        self.clear = false;
        (&mut self.builder).push_str(s);
    }

    fn write_rune(&mut self, c: char) {
        if c == '\n' {
            self.clear = true;
            if self.temp_indent {
                self.temp_indent = false;
                self.unindent();
            }
        } else if c != '\t' && c != ' ' {
            self.clear = false;
        }
        (&mut self.builder).push(c);
    }

    fn write_indent(&mut self) {
        for _ in 0..self.indentation {
            self.write_rune('\t')
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
        self.temp_indent = false;
    }

    fn format_comments(&mut self, mut comment: &ast::CommentList) {
        while let Some(boxed) = comment {
            if !self.clear {
                self.write_rune('\n');
                self.temp_indent = true;
                self.indent();
                self.write_indent();
            }
            self.write_string((*boxed).lit.as_str());
            self.clear = true;
            self.write_indent();
            comment = &(*boxed).next;
        }
    }

    fn write_comment(&mut self, comment: &str) {
        self.write_string("// ");
        self.write_string(comment);
        self.write_rune('\n')
    }

    pub fn format_file(&mut self, n: &File, include_pkg: bool) {
        let sep = '\n';
        if let Some(pkg) = &n.package {
            if include_pkg && !pkg.name.name.is_empty() {
                self.write_indent();
                self.format_node(&Node::PackageClause(pkg));
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

        let mut prev: i8 = -1;
        for (i, value) in n.body.iter().enumerate() {
            let cur = n.body.get(i).unwrap().typ();
            if i != 0 {
                self.write_rune(sep);
                // separate different statements with double newline
                if cur != prev {
                    self.write_rune(sep)
                }
            }
            self.write_indent();
            self.format_node(&Node::from_stmt(value));
            prev = cur;
        }

        if n.eof.is_some() {
            self.write_rune(sep);
            self.set_indent(0);
            self.clear = true;
            self.format_comments(&n.eof);
        }
    }

    fn format_node(&mut self, n: &Node) {
        // save current indentation
        let curr_ind = self.indentation;
        match n {
            Node::File(m) => self.format_file(m, true),
            Node::Block(m) => self.format_block(m),
            Node::ExprStmt(m) => self.format_expression_statement(m),
            Node::PackageClause(m) => self.format_package_clause(m),
            Node::ImportDeclaration(m) => self.format_import_declaration(m),
            Node::ReturnStmt(m) => self.format_return_statement(m),
            Node::OptionStmt(m) => self.format_option_statement(m),
            Node::TestStmt(m) => self.format_test_statement(m),
            Node::VariableAssgn(m) => self.format_variable_assignment(m),
            Node::IndexExpr(m) => self.format_index_expression(m),
            Node::MemberAssgn(m) => self.format_member_assignment(m),
            Node::CallExpr(m) => self.format_call_expression(m),
            Node::PipeExpr(m) => self.format_pipe_expression(m),
            Node::ConditionalExpr(m) => self.format_conditional_expression(m),
            Node::StringExpr(m) => self.format_string_expression(m),
            Node::ArrayExpr(m) => self.format_array_expression(m),
            Node::MemberExpr(m) => self.format_member_expression(m),
            Node::UnaryExpr(m) => self.format_unary_expression(m),
            Node::BinaryExpr(m) => self.format_binary_expression(m),
            Node::LogicalExpr(m) => self.format_logical_expression(m),
            Node::ParenExpr(m) => self.format_paren_expression(m),
            Node::FunctionExpr(m) => self.format_function_expression(m),
            Node::Property(m) => self.format_property(m),
            Node::TextPart(m) => self.format_text_part(m),
            Node::InterpolatedPart(m) => self.format_interpolated_part(m),
            Node::StringLit(m) => self.format_string_literal(m),
            Node::BooleanLit(m) => self.format_boolean_literal(m),
            Node::FloatLit(m) => self.format_float_literal(m),
            Node::IntegerLit(m) => self.format_integer_literal(m),
            Node::UintLit(m) => self.format_unsigned_integer_literal(m),
            Node::RegexpLit(m) => self.format_regexp_literal(m),
            Node::DurationLit(m) => self.format_duration_literal(m),
            Node::DateTimeLit(m) => self.format_date_time_literal(m),
            Node::PipeLit(m) => self.format_pipe_literal(m),
            Node::Identifier(m) => self.format_identifier(m),
            Node::ObjectExpr(m) => self.format_object_expression_braces(m, true),
            Node::Package(m) => self.format_package(m),
            Node::BadStmt(_) => self.err = Some(Error::from("bad statement")),
            Node::BadExpr(_) => self.err = Some(Error::from("bad expression")),
            Node::BuiltinStmt(m) => self.format_builtin(m),
        }
        self.set_indent(curr_ind)
    }

    fn format_package(&mut self, n: &ast::Package) {
        let pkg_name = &n.package;
        self.format_package_clause(&ast::PackageClause {
            name: ast::Identifier {
                name: String::from(pkg_name),
                base: ast::BaseNode::default(),
            },
            base: ast::BaseNode::default(),
        });
        for (i, file) in n.files.iter().enumerate() {
            if i != 0 {
                self.write_rune('\n');
                self.write_rune('\n');
            }
            if !file.name.is_empty() {
                self.write_comment(&file.name);
            }
            self.format_file(file, false)
        }
    }

    fn format_builtin(&mut self, n: &ast::BuiltinStmt) {
        self.format_comments(&n.base.comments);
        self.write_string("builtin ");
        self.format_identifier(&n.id);
        self.format_comments(&n.colon);
        if n.colon == None {
            self.write_string(" ");
        }
        self.write_string(": ");
        self.format_type_expression(&n.ty);
    }

    fn format_type_expression(&mut self, n: &ast::TypeExpression) {
        self.format_monotype(&n.monotype);
        if !n.constraints.is_empty() {
            self.write_string(" where ");
            self.format_constraints(&n.constraints);
        }
    }

    fn format_monotype(&mut self, n: &ast::MonoType) {
        match n {
            ast::MonoType::Tvar(tv) => self.format_tvar(tv),
            ast::MonoType::Basic(nt) => self.format_basic(&nt),
            ast::MonoType::Array(arr) => self.format_array(&*arr),
            ast::MonoType::Record(rec) => self.format_record(&rec),
            ast::MonoType::Function(fun) => self.format_function(&*fun),
        }
    }
    fn format_function(&mut self, n: &ast::FunctionType) {
        self.write_string("(");
        if !n.parameters.is_empty() {
            self.format_parameters(&n.parameters)
        }
        self.write_string(")");
        self.write_string(" => ");
        self.format_monotype(&n.monotype)
    }
    fn format_parameters(&mut self, n: &[ast::ParameterType]) {
        self.format_parameter(&n[0]);
        for p in &n[1..] {
            self.write_string(", ");
            self.format_parameter(p);
        }
    }
    fn format_parameter(&mut self, n: &ast::ParameterType) {
        match &n {
            ast::ParameterType::Required {
                base: _,
                name,
                monotype,
            } => {
                self.format_identifier(&name);
                self.write_string(": ");
                self.format_monotype(&monotype);
            }
            ast::ParameterType::Optional {
                base: _,
                name,
                monotype,
            } => {
                self.write_string("?");
                self.format_identifier(&name);
                self.write_string(": ");
                self.format_monotype(&monotype);
            }
            ast::ParameterType::Pipe {
                base: _,
                name,
                monotype,
            } => {
                self.write_string("<-");
                match name {
                    Some(n) => self.format_identifier(n),
                    None => {}
                }
                self.write_string(": ");
                self.format_monotype(&monotype);
            }
        }
    }
    fn format_record(&mut self, n: &ast::RecordType) {
        self.write_string("{");
        match &n.tvar {
            Some(tv) => {
                self.format_identifier(tv);
                self.write_string(" with ");
                self.format_properties(&n.properties);
            }
            None => {
                if !n.properties.is_empty() {
                    self.format_properties(&n.properties);
                }
            }
        }
        self.write_string("}");
    }
    fn format_properties(&mut self, n: &[ast::PropertyType]) {
        self.format_property_type(&n[0]);
        for p in &n[1..] {
            self.write_string(", ");
            self.format_property_type(&p);
        }
    }
    fn format_property_type(&mut self, n: &ast::PropertyType) {
        self.format_identifier(&n.name);
        self.write_string(": ");
        self.format_monotype(&n.monotype);
    }
    fn format_array(&mut self, n: &ast::ArrayType) {
        self.write_string("[");
        self.format_monotype(&n.element);
        self.write_string("]");
    }
    fn format_basic(&mut self, n: &ast::NamedType) {
        self.format_identifier(&n.name);
    }
    fn format_constraints(&mut self, n: &[ast::TypeConstraint]) {
        self.format_constraint(&n[0]);
        for c in &n[1..] {
            self.write_string(", ");
            self.format_constraint(c);
        }
    }
    fn format_constraint(&mut self, n: &ast::TypeConstraint) {
        self.format_identifier(&n.tvar);
        self.write_string(": ");
        self.format_kinds(&n.kinds);
    }
    fn format_kinds(&mut self, n: &[ast::Identifier]) {
        self.format_identifier(&n[0]);
        for k in &n[1..] {
            self.write_string(" + ");
            self.format_identifier(&k);
        }
    }
    fn format_tvar(&mut self, n: &ast::TvarType) {
        self.format_identifier(&n.name);
    }

    fn format_property(&mut self, n: &ast::Property) {
        self.format_property_key(&n.key);
        if let Some(v) = &n.value {
            self.format_comments(&n.separator);
            self.write_string(": ");
            self.format_node(&Node::from_expr(&v));
        }
    }

    fn format_function_expression(&mut self, n: &ast::FunctionExpr) {
        self.format_comments(&n.lparen);
        self.write_rune('(');
        let sep = ", ";
        for (i, _) in n.params.iter().enumerate() {
            if i != 0 {
                self.write_string(sep)
            }
            // treat properties differently than in general case
            let property = n.params.get(i).unwrap();
            self.format_function_argument(property);
            self.format_comments(&property.comma);
        }
        self.format_comments(&n.rparen);
        self.write_string(") ");
        self.format_comments(&n.arrow);
        self.write_string("=>");
        // must wrap body with parenthesis in order to discriminate between:
        //  - returning an object: (x) => ({foo: x})
        //  - and block statements:
        //		(x) => {
        //			return x + 1
        //		}
        match &n.body {
            ast::FunctionBody::Expr(b) => {
                self.write_rune('\n');
                self.indent();
                self.write_indent();
                self.write_rune('(');
                self.format_node(&Node::from_expr(&b));
                self.write_rune(')')
            }
            ast::FunctionBody::Block(b) => {
                self.write_rune(' ');
                self.format_block(&b);
            }
        }
    }

    fn format_function_argument(&mut self, n: &ast::Property) {
        if let Some(v) = &n.value {
            self.format_property_key(&n.key);
            self.format_comments(&n.separator);
            self.write_rune('=');
            self.format_node(&Node::from_expr(&v));
        } else {
            self.format_property_key(&n.key)
        }
    }

    fn format_property_key(&mut self, n: &ast::PropertyKey) {
        match n {
            ast::PropertyKey::StringLit(m) => self.format_string_literal(&m),
            ast::PropertyKey::Identifier(m) => self.format_identifier(&m),
        }
    }

    fn format_paren_expression(&mut self, n: &ast::ParenExpr) {
        // This could mix up ordering, but since the parens are programatically
        // added back, we would need to pass any closing comments down,
        // seriously complicating the function interface. For now, permit reordering.
        self.format_comments(&n.lparen);
        self.format_node(&Node::from_expr(&n.expression));
        self.format_comments(&n.rparen);
    }

    fn format_string_expression(&mut self, n: &ast::StringExpr) {
        self.format_comments(&n.base.comments);
        self.write_rune('"');
        for p in &n.parts {
            self.format_string_expression_part(p)
        }
        self.write_rune('"');
    }

    fn format_string_expression_part(&mut self, n: &ast::StringExprPart) {
        match n {
            ast::StringExprPart::Text(p) => self.format_text_part(&p),
            ast::StringExprPart::Interpolated(p) => self.format_interpolated_part(&p),
        }
    }

    fn format_text_part(&mut self, n: &ast::TextPart) {
        self.write_string(&n.value)
    }

    fn format_interpolated_part(&mut self, n: &ast::InterpolatedPart) {
        self.write_string("${");
        self.format_node(&Node::from_expr(&n.expression));
        self.write_string("}")
    }

    fn format_array_expression(&mut self, n: &ast::ArrayExpr) {
        self.format_comments(&n.lbrack);
        self.write_rune('[');
        let sep = ", ";
        for (i, _) in n.elements.iter().enumerate() {
            if i != 0 {
                self.write_string(sep)
            }
            let item = n.elements.get(i).unwrap();
            self.format_node(&Node::from_expr(&item.expression));
            self.format_comments(&item.comma);
        }
        self.format_comments(&n.rbrack);
        self.write_rune(']')
    }

    fn format_index_expression(&mut self, n: &ast::IndexExpr) {
        self.format_child_with_parens(Node::IndexExpr(n), Node::from_expr(&n.array));
        self.format_comments(&n.lbrack);
        self.write_rune('[');
        self.format_node(&Node::from_expr(&n.index));
        self.format_comments(&n.rbrack);
        self.write_rune(']');
    }

    fn format_block(&mut self, n: &ast::Block) {
        self.format_comments(&n.lbrace);
        self.write_rune('{');
        let sep = '\n';
        if !n.body.is_empty() {
            self.indent()
        }

        let mut prev: i8 = -1;
        for (i, smt) in n.body.iter().enumerate() {
            let cur = smt.typ();
            self.write_rune(sep);

            if i != 0 {
                // separate different statements with double newline
                if cur != prev {
                    self.write_rune(sep);
                }
            }
            self.write_indent();
            self.format_node(&Node::from_stmt(smt));
            prev = cur;
        }
        if !n.body.is_empty() {
            self.write_rune(sep);
            self.unindent();
            self.write_indent()
        }
        self.format_comments(&n.rbrace);
        self.write_rune('}')
    }

    fn format_return_statement(&mut self, n: &ast::ReturnStmt) {
        self.format_comments(&n.base.comments);
        self.write_string("return ");
        self.format_node(&Node::from_expr(&n.argument));
    }

    fn format_option_statement(&mut self, n: &ast::OptionStmt) {
        self.format_comments(&n.base.comments);
        self.write_string("option ");
        self.format_assignment(&n.assignment);
    }

    fn format_test_statement(&mut self, n: &ast::TestStmt) {
        self.format_comments(&n.base.comments);
        self.write_string("test ");
        self.format_node(&Node::VariableAssgn(&n.assignment));
    }

    fn format_assignment(&mut self, n: &ast::Assignment) {
        match &n {
            ast::Assignment::Variable(m) => {
                self.format_node(&Node::VariableAssgn(&m));
            }
            ast::Assignment::Member(m) => {
                self.format_node(&Node::MemberAssgn(&m));
            }
        }
    }

    // format_child_with_parens applies the generic rule for parenthesis (not for binary expressions).
    fn format_child_with_parens(&mut self, parent: Node, child: Node) {
        self.format_left_child_with_parens(&parent, &child)
    }

    // format_right_child_with_parens applies the generic rule for parenthesis to the right child of a binary expression.
    fn format_right_child_with_parens(&mut self, parent: &Node, child: &Node) {
        let (pvp, pvc) = get_precedences(parent, child);
        if needs_parenthesis(pvp, pvc, true) {
            self.format_node_with_parens(child);
        } else {
            self.format_node(child);
        }
    }

    // format_left_child_with_parens applies the generic rule for parenthesis to the left child of a binary expression.
    fn format_left_child_with_parens(&mut self, parent: &Node, child: &Node) {
        let (pvp, pvc) = get_precedences(parent, child);
        if needs_parenthesis(pvp, pvc, false) {
            self.format_node_with_parens(child);
        } else {
            self.format_node(child);
        }
    }

    fn format_node_with_parens(&mut self, node: &Node) {
        self.write_rune('(');
        self.format_node(node);
        self.write_rune(')')
    }

    fn format_member_expression(&mut self, n: &ast::MemberExpr) {
        self.format_child_with_parens(Node::MemberExpr(n), Node::from_expr(&n.object));

        match &n.property {
            ast::PropertyKey::Identifier(m) => {
                self.format_comments(&n.lbrack);
                self.write_rune('.');
                self.format_node(&Node::Identifier(&m));
            }
            ast::PropertyKey::StringLit(m) => {
                self.format_comments(&n.lbrack);
                self.write_rune('[');
                self.format_node(&Node::StringLit(&m));
                self.format_comments(&n.rbrack);
                self.write_rune(']');
            }
        }
    }

    fn format_pipe_expression(&mut self, n: &ast::PipeExpr) {
        self.format_node(&Node::from_expr(&n.argument));
        self.write_rune('\n');
        self.indent();
        self.write_indent();
        self.format_comments(&n.base.comments);
        self.write_string("|> ");
        self.format_node(&Node::CallExpr(&n.call));
    }

    fn format_call_expression(&mut self, n: &ast::CallExpr) {
        self.format_child_with_parens(Node::CallExpr(n), Node::from_expr(&n.callee));
        self.format_comments(&n.lparen);
        self.write_rune('(');
        let sep = ", ";
        for (i, c) in n.arguments.iter().enumerate() {
            if i != 0 {
                self.write_string(sep);
            }
            match c {
                ast::Expression::Object(s) => self.format_object_expression_as_function_argument(s),
                _ => self.format_node(&Node::from_expr(c)),
            }
        }
        self.format_comments(&n.rparen);
        self.write_rune(')');
    }

    fn format_object_expression_as_function_argument(&mut self, n: &ast::ObjectExpr) {
        // not called from formatNode, need to save indentation
        let i = self.indentation;
        self.format_object_expression_braces(n, false);
        self.set_indent(i);
    }

    fn format_object_expression_braces(&mut self, n: &ast::ObjectExpr, braces: bool) {
        let multiline = n.properties.len() > 3;
        self.format_comments(&n.lbrace);
        if braces {
            self.write_rune('{');
        }
        if let Some(with) = &n.with {
            self.format_identifier(&with.source);
            self.format_comments(&with.with);
            self.write_string(" with ");
        }
        if multiline {
            self.write_rune('\n');
            self.indent();
            self.write_indent();
        }
        let sep: &str;
        if multiline {
            sep = ",\n"
        } else {
            sep = ", "
        }
        for (i, _) in n.properties.iter().enumerate() {
            if i != 0 {
                self.write_string(sep);
                if multiline {
                    self.write_indent()
                }
            }
            let property = n.properties.get(i).unwrap();
            self.format_node(&Node::Property(property));
            self.format_comments(&property.comma);
        }
        if multiline {
            self.write_string(sep);
            self.unindent();
            self.write_indent();
        }
        self.format_comments(&n.rbrace);
        if braces {
            self.write_rune('}');
        }
    }

    fn format_identifier(&mut self, n: &ast::Identifier) {
        self.format_comments(&n.base.comments);
        self.write_string(&n.name);
    }

    fn format_variable_assignment(&mut self, n: &ast::VariableAssgn) {
        self.format_node(&Node::Identifier(&n.id));
        self.format_comments(&n.base.comments);
        self.write_string(" = ");
        self.format_node(&Node::from_expr(&n.init));
    }

    fn format_conditional_expression(&mut self, n: &ast::ConditionalExpr) {
        self.format_comments(&n.tk_if);
        self.write_string("if ");
        self.format_node(&Node::from_expr(&n.test));
        self.format_comments(&n.tk_then);
        self.write_string(" then ");
        self.format_node(&Node::from_expr(&n.consequent));
        self.format_comments(&n.tk_else);
        self.write_string(" else ");
        self.format_node(&Node::from_expr(&n.alternate));
    }

    fn format_member_assignment(&mut self, n: &ast::MemberAssgn) {
        self.format_node(&Node::MemberExpr(&n.member));
        self.format_comments(&n.base.comments);
        self.write_string(" = ");
        self.format_node(&Node::from_expr(&n.init));
    }

    fn format_unary_expression(&mut self, n: &ast::UnaryExpr) {
        self.format_comments(&n.base.comments);
        self.write_string(&n.operator.to_string());
        match n.operator {
            ast::Operator::SubtractionOperator => {}
            ast::Operator::AdditionOperator => {}
            _ => {
                self.write_rune(' ');
            }
        }
        self.format_child_with_parens(Node::UnaryExpr(&n), Node::from_expr(&n.argument));
    }

    fn format_binary_expression(&mut self, n: &ast::BinaryExpr) {
        self.format_binary(
            &n.base.comments,
            &n.operator.to_string(),
            Node::BinaryExpr(&n),
            Node::from_expr(&n.left),
            Node::from_expr(&n.right),
        );
    }

    fn format_logical_expression(&mut self, n: &ast::LogicalExpr) {
        self.format_binary(
            &n.base.comments,
            &n.operator.to_string(),
            Node::LogicalExpr(&n),
            Node::from_expr(&n.left),
            Node::from_expr(&n.right),
        );
    }

    fn format_binary(
        &mut self,
        comments: &ast::CommentList,
        op: &str,
        parent: Node,
        left: Node,
        right: Node,
    ) {
        self.format_left_child_with_parens(&parent, &left);
        self.write_rune(' ');
        self.format_comments(comments);
        self.write_string(op);
        self.write_rune(' ');
        self.format_right_child_with_parens(&parent, &right);
    }

    fn format_import_declaration(&mut self, n: &ast::ImportDeclaration) {
        self.format_comments(&n.base.comments);
        self.write_string("import ");
        if let Some(alias) = &n.alias {
            if !alias.name.is_empty() {
                self.format_node(&Node::Identifier(&alias));
                self.write_rune(' ')
            }
        }
        self.format_node(&Node::StringLit(&n.path))
    }

    fn format_expression_statement(&mut self, n: &ast::ExprStmt) {
        self.format_node(&Node::from_expr(&n.expression))
    }

    fn format_package_clause(&mut self, n: &ast::PackageClause) {
        self.format_comments(&n.base.comments);
        self.write_string("package ");
        self.format_node(&Node::Identifier(&n.name));
        self.write_rune('\n');
    }

    fn format_string_literal(&mut self, n: &ast::StringLit) {
        self.format_comments(&n.base.comments);
        if let Some(src) = &n.base.location.source {
            if !src.is_empty() {
                // Preserve the exact literal if we have it
                self.write_string(&src);
                return;
            }
        }

        // Write out escaped string value
        self.write_rune('"');
        let mut escaped: String;
        if n.value.contains("\"\\") {
            escaped = n.value.clone()
        } else {
            escaped = String::with_capacity(n.value.len() * 2);
            for r in n.value.chars() {
                if r == '"' || r == '\\' {
                    escaped.push('\\')
                }
                escaped.push(r)
            }
        }

        self.write_string(&escaped);
        self.write_rune('"');
    }

    // TODO(adriandt): this code appears dead. Boolean literal is no longer a node type?
    fn format_boolean_literal(&mut self, n: &ast::BooleanLit) {
        let s: &str;
        if n.value {
            s = "true"
        } else {
            s = "false"
        }
        self.write_string(s)
    }

    fn format_date_time_literal(&mut self, n: &ast::DateTimeLit) {
        // rust rfc3339NANO only support nano3, nano6, nano9 precisions
        // for frac nano6 timestamp in go like "2018-05-22T19:53:23.09012Z",
        // rust will append a zero at the end, like "2018-05-22T19:53:23.090120Z"
        // the following implementation will match go's rfc3339nano
        let mut f: String;
        let v = &n.value;
        let nano_sec = v.timestamp_subsec_nanos();
        if nano_sec > 0 {
            f = v.format("%FT%T").to_string();
            let mut frac_nano: String = v.format("%f").to_string();
            frac_nano.insert_str(0, ".");
            let mut r = frac_nano.chars().last().unwrap();
            while r == '0' {
                frac_nano.pop();
                r = frac_nano.chars().last().unwrap();
            }
            f.push_str(&frac_nano);

            if v.timezone().local_minus_utc() == 0 {
                f.push_str("Z")
            } else {
                f.push_str(&v.format("%:z").to_string());
            }
        } else {
            f = v.to_rfc3339_opts(SecondsFormat::Secs, true)
        }
        self.format_comments(&n.base.comments);
        self.write_string(&f);
    }

    fn format_duration_literal(&mut self, n: &ast::DurationLit) {
        self.format_comments(&n.base.comments);
        for d in &n.values {
            self.write_string(&format!("{}", d.magnitude));
            self.write_string(&d.unit)
        }
    }

    fn format_float_literal(&mut self, n: &ast::FloatLit) {
        self.format_comments(&n.base.comments);
        let mut s = format!("{}", n.value);
        if !s.contains('.') {
            s.push_str(".0");
        }
        self.write_string(&s)
    }

    fn format_integer_literal(&mut self, n: &ast::IntegerLit) {
        self.format_comments(&n.base.comments);
        self.write_string(&format!("{}", n.value));
    }

    fn format_unsigned_integer_literal(&mut self, n: &ast::UintLit) {
        self.format_comments(&n.base.comments);
        self.write_string(&format!("{0:10}", n.value))
    }

    fn format_pipe_literal(&mut self, n: &ast::PipeLit) {
        self.format_comments(&n.base.comments);
        self.write_string("<-")
    }

    fn format_regexp_literal(&mut self, n: &ast::RegexpLit) {
        self.format_comments(&n.base.comments);
        self.write_rune('/');
        self.write_string(&n.value.replace("/", "\\/"));
        self.write_rune('/')
    }
}

fn get_precedences(parent: &Node, child: &Node) -> (u32, u32) {
    let pvp: u32;
    let pvc: u32;
    match parent {
        Node::BinaryExpr(p) => pvp = Operator::new(&p.operator).get_precedence(),
        Node::LogicalExpr(p) => pvp = Operator::new_logical(&p.operator).get_precedence(),
        Node::UnaryExpr(p) => pvp = Operator::new(&p.operator).get_precedence(),
        Node::CallExpr(_) => pvp = 1,
        Node::MemberExpr(_) => pvp = 1,
        Node::IndexExpr(_) => pvp = 1,
        Node::ParenExpr(p) => return get_precedences(&(Node::from_expr(&p.expression)), child),
        _ => pvp = 0,
    }

    match child {
        Node::BinaryExpr(p) => pvc = Operator::new(&p.operator).get_precedence(),
        Node::LogicalExpr(p) => pvc = Operator::new_logical(&p.operator).get_precedence(),
        Node::UnaryExpr(p) => pvc = Operator::new(&p.operator).get_precedence(),
        Node::CallExpr(_) => pvc = 1,
        Node::MemberExpr(_) => pvc = 1,
        Node::IndexExpr(_) => pvc = 1,
        Node::ParenExpr(p) => return get_precedences(parent, &(Node::from_expr(&p.expression))),
        _ => pvc = 0,
    }

    (pvp, pvc)
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
                ast::Operator::PowerOperator => 2,
                ast::Operator::MultiplicationOperator => 4,
                ast::Operator::DivisionOperator => 4,
                ast::Operator::ModuloOperator => 3,
                ast::Operator::AdditionOperator => 4,
                ast::Operator::SubtractionOperator => 4,
                ast::Operator::LessThanEqualOperator => 5,
                ast::Operator::LessThanOperator => 5,
                ast::Operator::GreaterThanEqualOperator => 5,
                ast::Operator::GreaterThanOperator => 5,
                ast::Operator::StartsWithOperator => 5,
                ast::Operator::InOperator => 5,
                ast::Operator::NotEmptyOperator => 5,
                ast::Operator::EmptyOperator => 5,
                ast::Operator::EqualOperator => 5,
                ast::Operator::NotEqualOperator => 5,
                ast::Operator::RegexpMatchOperator => 5,
                ast::Operator::NotRegexpMatchOperator => 5,
                ast::Operator::NotOperator => 6,
                ast::Operator::ExistsOperator => 6,
                ast::Operator::InvalidOperator => 0,
            };
        }
        match self.l_op.unwrap() {
            ast::LogicalOperator::AndOperator => 7,
            ast::LogicalOperator::OrOperator => 8,
        }
    }
}

// About parenthesis:
// We need parenthesis if a child node has lower precedence (bigger value) than its parent node.
// The same stands for the left child of a binary expression; while, for the right child, we need parenthesis if its
// precedence is lower or equal then its parent's.
//
// To explain parenthesis logic, we must to understand how the parser generates the AST.
// (A) - The parser always puts lower precedence operators at the root of the AST.
// (B) - When there are multiple operators with the same precedence, the right-most expression is at root.
// (C) - When there are parenthesis, instead, the parser recursively generates a AST for the expression contained
// in the parenthesis, and makes it the right child.
// So, when formatting:
//  - if we encounter a child with lower precedence on the left, this means it requires parenthesis, because, for sure,
//    the parser detected parenthesis to break (A);
//  - if we encounter a child with higher or equal precedence on the left, it doesn't need parenthesis, because
//    that was the natural parsing order of elements (see (B));
//  - if we encounter a child with lower or equal precedence on the right, it requires parenthesis, otherwise, it
//    would have been at root (see (C)).
fn needs_parenthesis(pvp: u32, pvc: u32, is_right: bool) -> bool {
    // If one of the precedence values is invalid, then we shouldn't apply any parenthesis.
    let par = pvc != 0 && pvp != 0;
    par && ((!is_right && pvc > pvp) || (is_right && pvc >= pvp))
}

#[cfg(test)]
pub mod tests;
