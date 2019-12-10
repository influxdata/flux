extern crate flatbuffers;
extern crate walkdir;

use super::ast_generated::fbast;
use crate::ast;
use chrono::FixedOffset;

#[test]
fn test_flatbuffers_ast() {
    use super::ast_generated::fbast::*;
    let mut b = flatbuffers::FlatBufferBuilder::new_with_capacity(256);

    // Generate a flatbuffer representation of `40+60`.

    let lit1 = IntegerLiteral::create(
        &mut b,
        &IntegerLiteralArgs {
            value: 40,
            ..IntegerLiteralArgs::default()
        },
    );
    let lit2 = IntegerLiteral::create(
        &mut b,
        &IntegerLiteralArgs {
            value: 60,
            ..IntegerLiteralArgs::default()
        },
    );
    let add = BinaryExpression::create(
        &mut b,
        &BinaryExpressionArgs {
            operator: Operator::AdditionOperator,
            left_type: Expression::IntegerLiteral,
            left: Some(lit1.as_union_value()),
            right_type: Expression::IntegerLiteral,
            right: Some(lit2.as_union_value()),
            ..BinaryExpressionArgs::default()
        },
    );

    let stmt = ExpressionStatement::create(
        &mut b,
        &ExpressionStatementArgs {
            expression_type: Expression::BinaryExpression,
            expression: Some(add.as_union_value()),
            ..ExpressionStatementArgs::default()
        },
    );

    let wrapped_stmt = WrappedStatement::create(
        &mut b,
        &WrappedStatementArgs {
            statement_type: Statement::ExpressionStatement,
            statement: Some(stmt.as_union_value()),
        },
    );

    let stmts = b.create_vector(&[wrapped_stmt]);

    let file = File::create(
        &mut b,
        &FileArgs {
            body: Some(stmts),
            ..FileArgs::default()
        },
    );

    let files = b.create_vector(&[file]);

    let pkg = Package::create(
        &mut b,
        &PackageArgs {
            files: Some(files),
            ..PackageArgs::default()
        },
    );

    b.finish(pkg, None);
    let bytes = b.finished_data();
    assert_ne!(bytes.len(), 0);
}

#[test]
fn test_flatbuffers_serialize() {
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
    let pkg = ast::Package {
        base: ast::BaseNode {
            ..ast::BaseNode::default()
        },
        path: String::from("./"),
        package: String::from("test"),
        files: f,
    };
    let (vec, offset) = match super::serialize(&pkg) {
        Ok((v, o)) => (v, o),
        Err(e) => {
            assert!(false, e);
            return;
        }
    };
    let fb = &vec.as_slice()[offset..];
    match compare_ast_fb(&pkg, fb) {
        Err(e) => assert!(false, e),
        _ => (),
    }
}

#[test]
fn test_serialize_all_flux_files() {
    use walkdir::WalkDir;
    for entry in WalkDir::new("../stdlib").into_iter().filter_map(|e| e.ok()) {
        let f_name = entry.file_name().to_string_lossy();
        if f_name.ends_with(".flux") {
            let flux_script = match std::fs::read_to_string(entry.path()) {
                Ok(s) => s,
                Err(e) => {
                    assert!(false, format!("{}", e));
                    return;
                }
            };
            let path = if let Some(s) = entry.path().to_str() {
                s
            } else {
                ""
            };
            match serialize_and_compare(path, flux_script.as_str()) {
                Ok(()) => (),
                Err(e) => assert!(false, e),
            };
        }
    }
}

fn serialize_and_compare(path: &str, flux_script: &str) -> Result<(), String> {
    use std::time::Instant;
    println!("{}", path);
    let now = Instant::now();
    let ast_file = crate::parser::parse_string("test", flux_script);
    println!(
        "  parsing took {}s",
        now.elapsed().as_micros() as f64 / 1_000_000.0
    );
    let pkg = ast::Package {
        base: ast::BaseNode {
            ..ast::BaseNode::default()
        },
        path: String::from("./"),
        package: String::from("test"),
        files: vec![ast_file],
    };
    let now = Instant::now();
    let (vec, offset) = super::serialize(&pkg)?;
    println!(
        "  serializing took {}s",
        now.elapsed().as_nanos() as f64 / 1_000_000_000.0
    );
    let fb = &vec.as_slice()[offset..];
    compare_ast_fb(&pkg, fb)
}

fn compare_ast_fb(ast_pkg: &ast::Package, fb: &[u8]) -> Result<(), String> {
    let fb_pkg = fbast::get_root_as_package(fb);
    compare_pkg_fb(ast_pkg, &fb_pkg)?;
    Ok(())
}

fn compare_pkg_fb(ast_pkg: &ast::Package, fb_pkg: &fbast::Package) -> Result<(), String> {
    compare_base(&ast_pkg.base, &fb_pkg.base_node())?;
    compare_strings("package path", &ast_pkg.path, &fb_pkg.path())?;
    compare_strings("package name", &ast_pkg.package, &fb_pkg.package())?;

    let fb_files = &fb_pkg.files();
    let fb_files = unwrap_or_fail("package files", fb_files)?;
    compare_vec_len(&ast_pkg.files, &fb_files)?;
    let mut i: usize = 0;
    loop {
        if i >= ast_pkg.files.len() {
            return Ok(());
        }

        let ast_file = &ast_pkg.files[i];
        let fb_file = &fb_files.get(i);
        compare_files(ast_file, fb_file)?;
        i = i + 1;
    }
}

fn compare_files(ast_file: &ast::File, fb_file: &fbast::File) -> Result<(), String> {
    compare_base(&ast_file.base, &fb_file.base_node())?;
    compare_strings("file name", &ast_file.name, &fb_file.name())?;
    compare_strings("metadata", &ast_file.metadata, &fb_file.metadata())?;
    compare_package_clause(&ast_file.package, &fb_file.package())?;
    compare_imports(&ast_file.imports, &fb_file.imports())?;
    compare_stmt_vectors(&ast_file.body, &fb_file.body())?;
    Ok(())
}

fn compare_stmt_vectors(
    ast_stmts: &Vec<ast::Statement>,
    fb_stmts: &Option<flatbuffers::Vector<flatbuffers::ForwardsUOffset<fbast::WrappedStatement>>>,
) -> Result<(), String> {
    let fb_stmts = unwrap_or_fail("statement list", fb_stmts)?;
    compare_vec_len(ast_stmts, fb_stmts)?;
    let mut i: usize = 0;
    loop {
        if i >= ast_stmts.len() {
            break Ok(());
        }
        let fb_stmt_ty = fb_stmts.get(i).statement_type();
        let fb_stmt = &fb_stmts.get(i).statement();
        compare_stmts(&ast_stmts[i], fb_stmt_ty, fb_stmt)?;
        i = i + 1;
    }
}

fn compare_stmts(
    ast_stmt: &ast::Statement,
    fb_stmt_ty: fbast::Statement,
    fb_stmt: &Option<flatbuffers::Table>,
) -> Result<(), String> {
    let fb_tbl = unwrap_or_fail("statement", fb_stmt)?;
    match (ast_stmt, fb_stmt_ty) {
        (ast::Statement::Variable(ast_stmt), fbast::Statement::VariableAssignment) => {
            let fb_stmt = fbast::VariableAssignment::init_from_table(*fb_tbl);
            compare_var_assign(&ast_stmt, &Some(fb_stmt))
        }
        (ast::Statement::Expr(ast_stmt), fbast::Statement::ExpressionStatement) => {
            let fb_stmt = fbast::ExpressionStatement::init_from_table(*fb_tbl);
            compare_base(&ast_stmt.base, &fb_stmt.base_node())?;
            compare_exprs(
                &ast_stmt.expression,
                fb_stmt.expression_type(),
                &fb_stmt.expression(),
            )?;
            Ok(())
        }
        (ast::Statement::Option(ast_stmt), fbast::Statement::OptionStatement) => {
            let fb_stmt = fbast::OptionStatement::init_from_table(*fb_tbl);
            compare_base(&ast_stmt.base, &fb_stmt.base_node())?;
            compare_assignments(
                &ast_stmt.assignment,
                fb_stmt.assignment_type(),
                &fb_stmt.assignment(),
            )
        }
        (ast::Statement::Return(ast_stmt), fbast::Statement::ReturnStatement) => {
            let fb_stmt = fbast::ReturnStatement::init_from_table(*fb_tbl);
            compare_base(&ast_stmt.base, &fb_stmt.base_node())?;
            compare_exprs(
                &ast_stmt.argument,
                fb_stmt.argument_type(),
                &fb_stmt.argument(),
            )
        }
        (ast::Statement::Bad(ast_stmt), fbast::Statement::BadStatement) => {
            let fb_stmt = fbast::BadStatement::init_from_table(*fb_tbl);
            compare_base(&ast_stmt.base, &fb_stmt.base_node())?;
            compare_strings("bad stmt", &ast_stmt.text, &fb_stmt.text())
        }
        (ast::Statement::Test(ast_stmt), fbast::Statement::TestStatement) => {
            let fb_stmt = fbast::TestStatement::init_from_table(*fb_tbl);
            compare_base(&ast_stmt.base, &fb_stmt.base_node())?;
            match fb_stmt.assignment_type() == fbast::Assignment::VariableAssignment {
                false => Err(String::from("expected var assignment in test stmt")),
                true => {
                    let fb_var_assign = &fb_stmt.assignment_as_variable_assignment();
                    compare_var_assign(&ast_stmt.assignment, fb_var_assign)
                }
            }
        }
        (ast::Statement::Builtin(ast_stmt), fbast::Statement::BuiltinStatement) => {
            let fb_stmt = fbast::BuiltinStatement::init_from_table(*fb_tbl);
            compare_base(&ast_stmt.base, &fb_stmt.base_node())?;
            compare_ids(&ast_stmt.id, &fb_stmt.id())
        }
        (ast_stmt, fb_ty) => {
            let ast_stmt_ty = ast::walk::Node::from_stmt(ast_stmt);
            let fb_ty = fbast::enum_name_statement(fb_ty);
            Err(String::from(format!(
                "wrong statement type; ast = {}, fb = {}",
                ast_stmt_ty, fb_ty
            )))
        }
    }
}

fn compare_assignments(
    ast_asgn: &ast::Assignment,
    fb_asgn_ty: fbast::Assignment,
    fb_asgn: &Option<flatbuffers::Table>,
) -> Result<(), String> {
    let fb_tbl = unwrap_or_fail("assign", fb_asgn)?;
    match (ast_asgn, fb_asgn_ty) {
        (ast::Assignment::Variable(ast_va), fbast::Assignment::VariableAssignment) => {
            let fb_va = fbast::VariableAssignment::init_from_table(*fb_tbl);
            compare_var_assign(ast_va, &Some(fb_va))
        }
        (ast::Assignment::Member(ast_ma), fbast::Assignment::MemberAssignment) => {
            let fb_ma = fbast::MemberAssignment::init_from_table(*fb_tbl);
            compare_base(&ast_ma.base, &fb_ma.base_node())?;
            compare_member_expr(&ast_ma.member, &fb_ma.member())?;
            compare_exprs(&ast_ma.init, fb_ma.init__type(), &fb_ma.init_())
        }
        _ => Err(String::from("assignment mismatch")),
    }
}

fn compare_var_assign(
    ast_va: &ast::VariableAssgn,
    fb_va: &Option<fbast::VariableAssignment>,
) -> Result<(), String> {
    let fb_va = unwrap_or_fail("var assign", fb_va)?;
    compare_base(&ast_va.base, &fb_va.base_node())?;
    compare_ids(&ast_va.id, &fb_va.id())?;
    compare_exprs(&ast_va.init, fb_va.init__type(), &fb_va.init_())
}

fn compare_exprs(
    ast_expr: &ast::Expression,
    fb_expr_ty: fbast::Expression,
    fb_tbl: &Option<flatbuffers::Table>,
) -> Result<(), String> {
    let fb_tbl = unwrap_or_fail("expr", fb_tbl)?;
    match (ast_expr, fb_expr_ty) {
        (ast::Expression::Integer(ast_int), fbast::Expression::IntegerLiteral) => {
            let fb_int = fbast::IntegerLiteral::init_from_table(*fb_tbl);
            compare_base(&ast_expr.base(), &fb_int.base_node())?;
            match ast_int.value == fb_int.value() {
                true => Ok(()),
                false => Err(String::from(format!(
                    "int lit mismatch; ast = {}, fb = {}",
                    ast_int.value,
                    fb_int.value()
                ))),
            }
        }
        (ast::Expression::Float(ast_float), fbast::Expression::FloatLiteral) => {
            let fb_float = fbast::FloatLiteral::init_from_table(*fb_tbl);
            compare_base(&ast_float.base, &fb_float.base_node())?;
            match ast_float.value == fb_float.value() {
                true => Ok(()),
                false => Err(String::from(format!(
                    "float lit mismatch; ast = {}, fb = {}",
                    ast_float.value,
                    fb_float.value()
                ))),
            }
        }
        (ast::Expression::StringLit(ast_string), fbast::Expression::StringLiteral) => {
            let fb_string = fbast::StringLiteral::init_from_table(*fb_tbl);
            compare_base(&ast_string.base, &fb_string.base_node())?;
            let fb_value = fb_string.value();
            let fb_value = unwrap_or_fail("string lit string", &fb_value)?;
            match &ast_string.value.as_str() == fb_value {
                true => Ok(()),
                false => Err(String::from(format!(
                    "string lit mismatch; ast = {}, fb = {}",
                    ast_string.value, fb_value,
                ))),
            }
        }
        (ast::Expression::Duration(ast_dur), fbast::Expression::DurationLiteral) => {
            let fb_dur_lit = fbast::DurationLiteral::init_from_table(*fb_tbl);
            compare_base(&ast_dur.base, &fb_dur_lit.base_node())?;
            let fb_values = fb_dur_lit.values();
            let fb_values = unwrap_or_fail("dur lit values", &fb_values)?;
            compare_vec_len(&ast_dur.values, fb_values)?;
            let mut i: usize = 0;
            loop {
                if i >= ast_dur.values.len() {
                    break Ok(());
                }
                let ast_d = &ast_dur.values[i];
                let fb_d = fb_values.get(i);
                if ast_d.magnitude != fb_d.magnitude() {
                    return Err(String::from("invalid duration magnitude"));
                }
                if ast_d.unit != fbast::enum_name_time_unit(fb_d.unit()) {
                    return Err(String::from("invalid duration time unit"));
                }
                i = i + 1;
            }
        }
        (ast::Expression::DateTime(ast_dtl), fbast::Expression::DateTimeLiteral) => {
            let fb_dtl = fbast::DateTimeLiteral::init_from_table(*fb_tbl);
            let dtl = chrono::DateTime::<FixedOffset>::from_utc(
                chrono::NaiveDateTime::from_timestamp(fb_dtl.secs(), fb_dtl.nsecs()),
                FixedOffset::east(fb_dtl.offset()),
            );
            compare_base(&ast_dtl.base, &fb_dtl.base_node())?;
            if ast_dtl.value != dtl {
                return Err(String::from("invalid DateTimeLiteral value"));
            }
            Ok(())
        }
        (ast::Expression::Regexp(ast_rel), fbast::Expression::RegexpLiteral) => {
            let fb_rel = fbast::RegexpLiteral::init_from_table(*fb_tbl);
            compare_base(&ast_rel.base, &fb_rel.base_node())?;
            compare_strings("regexp lit value", &ast_rel.value, &fb_rel.value())?;
            Ok(())
        }
        (ast::Expression::PipeLit(ast_pl), fbast::Expression::PipeLiteral) => {
            let fb_pl = fbast::PipeLiteral::init_from_table(*fb_tbl);
            compare_base(&ast_pl.base, &fb_pl.base_node())?;
            Ok(())
        }
        (ast::Expression::Identifier(ast_id), fbast::Expression::Identifier) => {
            let fb_id = fbast::Identifier::init_from_table(*fb_tbl);
            compare_ids(ast_id, &Some(fb_id))?;
            Ok(())
        }
        (ast::Expression::Array(ast_ae), fbast::Expression::ArrayExpression) => {
            let fb_ae = fbast::ArrayExpression::init_from_table(*fb_tbl);
            compare_base(&ast_ae.base, &fb_ae.base_node())?;
            let fb_elems = fb_ae.elements();
            let fb_elems = unwrap_or_fail("array elems", &fb_elems)?;
            compare_vec_len(&ast_ae.elements, fb_elems)?;

            let mut i: usize = 0;
            loop {
                if i >= ast_ae.elements.len() {
                    break Ok(());
                }
                let fb_we = &fb_elems.get(i);
                let fb_e = &fb_we.expr();
                compare_exprs(&ast_ae.elements[i], fb_we.expr_type(), fb_e)?;
                i = i + 1
            }
        }
        (ast::Expression::Function(ast_fe), fbast::Expression::FunctionExpression) => {
            let fb_fe = fbast::FunctionExpression::init_from_table(*fb_tbl);
            compare_base(&ast_fe.base, &fb_fe.base_node())?;
            compare_property_list(&ast_fe.params, &fb_fe.params())?;
            match (&ast_fe.body, fb_fe.body_type()) {
                (
                    ast::FunctionBody::Expr(ast_expr),
                    fbast::ExpressionOrBlock::WrappedExpression,
                ) => {
                    let fb_we = fb_fe.body_as_wrapped_expression();
                    let fb_we = unwrap_or_fail("function body wrapped expr", &fb_we)?;
                    compare_exprs(ast_expr, fb_we.expr_type(), &fb_we.expr())
                }
                (ast::FunctionBody::Block(ast_bl), fbast::ExpressionOrBlock::Block) => {
                    let fb_bl = fb_fe.body_as_block();
                    let fb_bl = unwrap_or_fail("function body block", &fb_bl)?;
                    compare_base(&ast_bl.base, &fb_bl.base_node())?;
                    compare_stmt_vectors(&ast_bl.body, &fb_bl.body())
                }
                _ => Err(String::from("function body mismatch")),
            }
        }
        (ast::Expression::Logical(ast_le), fbast::Expression::LogicalExpression) => {
            let fb_le = fbast::LogicalExpression::init_from_table(*fb_tbl);
            compare_base(&ast_le.base, &fb_le.base_node())?;
            compare_exprs(&ast_le.left, fb_le.left_type(), &fb_le.left())?;
            compare_exprs(&ast_le.right, fb_le.right_type(), &fb_le.right())?;
            match ast_logical_operator(&fb_le.operator()) == ast_le.operator {
                true => Ok(()),
                false => Err(String::from("logical operator mismatch")),
            }
        }
        (ast::Expression::Object(ast_oe), fbast::Expression::ObjectExpression) => {
            let fb_oe = fbast::ObjectExpression::init_from_table(*fb_tbl);
            compare_base(&ast_oe.base, &fb_oe.base_node())?;
            compare_property_list(&ast_oe.properties, &fb_oe.properties())?;
            compare_opt_ids(&ast_oe.with, &fb_oe.with())
        }
        (ast::Expression::Member(ast_me), fbast::Expression::MemberExpression) => {
            let fb_me = fbast::MemberExpression::init_from_table(*fb_tbl);
            compare_member_expr(&ast_me, &Some(fb_me))
        }
        (ast::Expression::Index(ast_ie), fbast::Expression::IndexExpression) => {
            let fb_ie = fbast::IndexExpression::init_from_table(*fb_tbl);
            compare_base(&ast_ie.base, &fb_ie.base_node())?;
            compare_exprs(&ast_ie.array, fb_ie.array_type(), &fb_ie.array())?;
            compare_exprs(&ast_ie.index, fb_ie.index_type(), &fb_ie.index())
        }
        (ast::Expression::Binary(ast_be), fbast::Expression::BinaryExpression) => {
            let fb_be = fbast::BinaryExpression::init_from_table(*fb_tbl);
            compare_base(&ast_be.base, &fb_be.base_node())?;
            compare_exprs(&ast_be.left, fb_be.left_type(), &fb_be.left())?;
            compare_exprs(&ast_be.right, fb_be.right_type(), &fb_be.right())?;
            match ast_operator(fb_be.operator()) == ast_be.operator {
                true => Ok(()),
                false => Err(String::from("binary operator mismatch")),
            }
        }
        (ast::Expression::Unary(ast_ue), fbast::Expression::UnaryExpression) => {
            let fb_ue = fbast::UnaryExpression::init_from_table(*fb_tbl);
            compare_base(&ast_ue.base, &fb_ue.base_node())?;
            compare_exprs(&ast_ue.argument, fb_ue.argument_type(), &fb_ue.argument())?;
            match ast_operator(fb_ue.operator()) == ast_ue.operator {
                true => Ok(()),
                false => Err(String::from("unary operator mismatch")),
            }
        }
        (ast::Expression::PipeExpr(ast_pe), fbast::Expression::PipeExpression) => {
            let fb_pe = fbast::PipeExpression::init_from_table(*fb_tbl);
            compare_base(&ast_pe.base, &fb_pe.base_node())?;
            compare_exprs(&ast_pe.argument, fb_pe.argument_type(), &fb_pe.argument())?;
            compare_call_exprs(&ast_pe.call, &fb_pe.call())
        }
        (ast::Expression::Call(ast_ce), fbast::Expression::CallExpression) => {
            let fb_ce = fbast::CallExpression::init_from_table(*fb_tbl);
            compare_call_exprs(&ast_ce, &Some(fb_ce))
        }
        (ast::Expression::Conditional(ast_ce), fbast::Expression::ConditionalExpression) => {
            let fb_ce = fbast::ConditionalExpression::init_from_table(*fb_tbl);
            compare_base(&ast_ce.base, &fb_ce.base_node())?;
            compare_exprs(&ast_ce.test, fb_ce.test_type(), &fb_ce.test())?;
            compare_exprs(
                &ast_ce.consequent,
                fb_ce.consequent_type(),
                &fb_ce.consequent(),
            )?;
            compare_exprs(
                &ast_ce.alternate,
                fb_ce.alternate_type(),
                &fb_ce.alternate(),
            )
        }
        (ast::Expression::StringExpr(ast_se), fbast::Expression::StringExpression) => {
            let fb_se = fbast::StringExpression::init_from_table(*fb_tbl);
            compare_base(&ast_se.base, &fb_se.base_node())?;
            compare_string_expr_part_list(&ast_se.parts, &fb_se.parts())
        }
        (ast::Expression::Paren(ast_pe), fbast::Expression::ParenExpression) => {
            let fb_pe = fbast::ParenExpression::init_from_table(*fb_tbl);
            compare_base(&ast_pe.base, &fb_pe.base_node())?;
            compare_exprs(
                &ast_pe.expression,
                fb_pe.expression_type(),
                &fb_pe.expression(),
            )
        }
        (ast::Expression::Bad(ast_be), fbast::Expression::BadExpression) => {
            let fb_be = fbast::BadExpression::init_from_table(*fb_tbl);
            compare_base(&ast_be.base, &fb_be.base_node())?;
            compare_strings("bad expr text", &ast_be.text, &fb_be.text())?;
            compare_opt_exprs(
                &ast_be.expression,
                fb_be.expression_type(),
                &fb_be.expression(),
            )
        }
        (ast_expr, fb_expr_ty) => {
            let ast_expr_ty = ast::walk::Node::from_expr(ast_expr);
            let fb_ty = fbast::enum_name_expression(fb_expr_ty);
            Err(String::from(format!(
                "wrong expr type; ast = {}, fb = {}",
                ast_expr_ty, fb_ty
            )))
        }
    }
}

fn compare_call_exprs(
    ast_ce: &ast::CallExpr,
    fb_ce: &Option<fbast::CallExpression>,
) -> Result<(), String> {
    let fb_ce = unwrap_or_fail("call expr", fb_ce)?;
    compare_base(&ast_ce.base, &fb_ce.base_node())?;
    compare_exprs(&ast_ce.callee, fb_ce.callee_type(), &fb_ce.callee())?;
    let ast_args = &ast_ce.arguments;
    let fb_args = fb_ce.arguments();
    match (ast_args.len(), fb_args) {
        (0, None) => Ok(()),
        (1, Some(fb_arg)) => compare_exprs(
            &ast_args[0],
            fbast::Expression::ObjectExpression,
            &Some(fb_arg._tab),
        ),
        (0, Some(_)) => Err(String::from("found call arg where not expected")),
        (1, None) => Err(String::from("missing call arg")),
        _ => Err(String::from("strange ast with more than one arg")),
    }
}

fn compare_member_expr(
    ast_me: &ast::MemberExpr,
    fb_me: &Option<fbast::MemberExpression>,
) -> Result<(), String> {
    let fb_me = unwrap_or_fail("member expression", fb_me)?;
    compare_base(&ast_me.base, &fb_me.base_node())?;
    compare_exprs(&ast_me.object, fb_me.object_type(), &fb_me.object())?;
    compare_property_key(&ast_me.property, fb_me.property_type(), &fb_me.property())
}

fn compare_string_expr_part_list(
    ast_parts: &Vec<ast::StringExprPart>,
    fb_parts: &Option<
        flatbuffers::Vector<flatbuffers::ForwardsUOffset<fbast::StringExpressionPart>>,
    >,
) -> Result<(), String> {
    let fb_parts = unwrap_or_fail("string expr parts", fb_parts)?;
    compare_vec_len(ast_parts, fb_parts)?;
    let mut i: usize = 0;
    loop {
        if i >= ast_parts.len() {
            break Ok(());
        }

        compare_string_expr_part(&ast_parts[i], &fb_parts.get(i))?;
        i = i + 1
    }
}

fn compare_string_expr_part(
    ast_part: &ast::StringExprPart,
    fb_part: &fbast::StringExpressionPart,
) -> Result<(), String> {
    match (
        ast_part,
        fb_part.text_value(),
        fb_part.interpolated_expression_type(),
        fb_part.interpolated_expression(),
    ) {
        (ast::StringExprPart::Text(ast_text), Some(fb_text), fbast::Expression::NONE, None) => {
            compare_base(&ast_text.base, &fb_part.base_node())?;
            match ast_text.value.as_str() == fb_text {
                true => Ok(()),
                false => Err(String::from(
                    "mismatch in value of text part of string expr",
                )),
            }
        }
        (ast::StringExprPart::Interpolated(ast_ip), None, fb_expr_ty, fb_expr) => {
            compare_base(&ast_ip.base, &fb_part.base_node())?;
            compare_exprs(&ast_ip.expression, fb_expr_ty, &fb_expr)
        }
        _ => Err(String::from(
            "mismatch in string expr part text/interpolated",
        )),
    }
}

fn compare_property_list(
    ast_pl: &Vec<ast::Property>,
    fb_pl: &Option<flatbuffers::Vector<flatbuffers::ForwardsUOffset<fbast::Property>>>,
) -> Result<(), String> {
    let fb_pl = unwrap_or_fail("property list", fb_pl)?;
    compare_vec_len(ast_pl, fb_pl)?;
    let mut i: usize = 0;
    loop {
        if i >= ast_pl.len() {
            return Ok(());
        }

        compare_property(&ast_pl[i], &fb_pl.get(i))?;
        i = i + 1;
    }
}

fn compare_property(ast_prop: &ast::Property, fb_prop: &fbast::Property) -> Result<(), String> {
    compare_base(&ast_prop.base, &fb_prop.base_node())?;
    // compare keys
    compare_property_key(&ast_prop.key, fb_prop.key_type(), &fb_prop.key())?;
    match (&ast_prop.key, fb_prop.key_type()) {
        (ast::PropertyKey::Identifier(ast_id), fbast::PropertyKey::Identifier) => {
            let fb_id = &fb_prop.key_as_identifier();
            compare_ids(ast_id, fb_id)?;
        }
        (ast::PropertyKey::StringLit(ast_str), fbast::PropertyKey::StringLiteral) => {
            let fb_str = &fb_prop.key_as_string_literal();
            compare_string_lits(ast_str, fb_str)?;
        }
        _ => return Err(String::from("property key mismatch")),
    }
    compare_opt_exprs(&ast_prop.value, fb_prop.value_type(), &fb_prop.value())
}

fn compare_property_key(
    ast_key: &ast::PropertyKey,
    fb_key_ty: fbast::PropertyKey,
    fb_key: &Option<flatbuffers::Table>,
) -> Result<(), String> {
    let fb_key = unwrap_or_fail("property key", &fb_key)?;
    match (&ast_key, fb_key_ty) {
        (ast::PropertyKey::Identifier(ast_id), fbast::PropertyKey::Identifier) => {
            let fb_id = &fbast::Identifier::init_from_table(*fb_key);
            compare_ids(ast_id, &Some(*fb_id))
        }
        (ast::PropertyKey::StringLit(ast_str), fbast::PropertyKey::StringLiteral) => {
            let fb_str = &fbast::StringLiteral::init_from_table(*fb_key);
            compare_string_lits(ast_str, &Some(*fb_str))
        }
        _ => Err(String::from("property key mismatch")),
    }
}

fn compare_opt_exprs(
    ast_expr: &Option<ast::Expression>,
    fb_expr_ty: fbast::Expression,
    fb_expr: &Option<flatbuffers::Table>,
) -> Result<(), String> {
    match (ast_expr, fb_expr_ty) {
        (None, fbast::Expression::NONE) => Ok(()),
        (None, _) => Err(String::from("expected no expr but got one")),
        (Some(_), fbast::Expression::NONE) => Err(String::from("expected an expr but got none")),
        (Some(ast_expr), _) => compare_exprs(ast_expr, fb_expr_ty, fb_expr),
    }
}

fn compare_imports(
    ast_imports: &Vec<ast::ImportDeclaration>,
    fb_imports: &Option<
        flatbuffers::Vector<flatbuffers::ForwardsUOffset<fbast::ImportDeclaration>>,
    >,
) -> Result<(), String> {
    let fb_imports = unwrap_or_fail("imports", fb_imports)?;
    compare_vec_len(ast_imports, fb_imports)?;
    let mut i: usize = 0;
    loop {
        if i >= ast_imports.len() {
            break Ok(());
        }

        compare_import_decls(&ast_imports[i], &fb_imports.get(i))?;
        i = i + 1;
    }
}

fn compare_import_decls(
    ast_id: &ast::ImportDeclaration,
    fb_id: &fbast::ImportDeclaration,
) -> Result<(), String> {
    compare_opt_ids(&ast_id.alias, &fb_id.as_())?;
    compare_string_lits(&ast_id.path, &fb_id.path())?;
    Ok(())
}

fn compare_package_clause(
    ast_pkg_clause: &Option<ast::PackageClause>,
    fb_pkg_clause: &Option<fbast::PackageClause>,
) -> Result<(), String> {
    let (ast_pkg_clause, fb_pkg_clause) = match (ast_pkg_clause, fb_pkg_clause) {
        (None, None) => return Ok(()),
        (None, Some(_)) => return Err(String::from("found package clause where not expected")),
        (Some(_), None) => return Err(String::from("missing package clause")),
        (Some(ac), Some(fc)) => (ac, fc),
    };
    compare_base(&ast_pkg_clause.base, &fb_pkg_clause.base_node())?;
    compare_ids(&ast_pkg_clause.name, &fb_pkg_clause.name())?;
    Ok(())
}

fn compare_ids(ast_id: &ast::Identifier, fb_id: &Option<fbast::Identifier>) -> Result<(), String> {
    let fb_id = unwrap_or_fail("id", fb_id)?;
    compare_base(&ast_id.base, &fb_id.base_node())?;
    compare_strings("id", &ast_id.name, &fb_id.name())?;
    Ok(())
}

fn compare_opt_ids(
    ast_id: &Option<ast::Identifier>,
    fb_id: &Option<fbast::Identifier>,
) -> Result<(), String> {
    match (ast_id, fb_id) {
        (None, None) => Ok(()),
        (Some(_), None) => Err(String::from("compare opt ids, ast had one, fb did not")),
        (None, Some(_)) => Err(String::from("compare opt ids, ast had none, fb did")),
        (Some(ast_id), fb_id) => compare_ids(ast_id, fb_id),
    }
}

fn compare_vec_len<T, U>(ast_vec: &Vec<T>, fb_vec: &flatbuffers::Vector<U>) -> Result<(), String> {
    match ast_vec.len() == fb_vec.len() {
        true => Ok(()),
        false => Err(String::from(format!(
            "vectors have different lengths: ast = {}, fb = {}",
            ast_vec.len(),
            fb_vec.len(),
        ))),
    }
}

fn unwrap_or_fail<'a, T>(msg: &str, o: &'a Option<T>) -> Result<&'a T, String> {
    match o {
        None => Err(String::from(format!("missing {}", msg))),
        Some(t) => Ok(t),
    }
}

fn compare_strings(msg: &str, ast_str: &String, fb_str: &Option<&str>) -> Result<(), String> {
    let fb_str = unwrap_or_fail("string", fb_str)?;
    if ast_str.as_str() != *fb_str {
        return Err(format!(
            "{} mismatch: ast: {}, fb: {}",
            msg, ast_str, fb_str
        ));
    };
    Ok(())
}

fn compare_opt_strings(
    msg: &str,
    ast_str: &Option<String>,
    fb_str: &Option<&str>,
) -> Result<(), String> {
    match (ast_str, fb_str) {
        (None, None) => return Ok(()),
        (None, Some(s)) => Err(String::from(format!(
            "comparing opt string for {}: ast had none, fb had {}",
            msg, s,
        ))),
        (Some(s), None) => Err(String::from(format!(
            "comparing opt string for {}: ast had {}, fb had none",
            msg, s,
        ))),
        (Some(ast_str), fb_str) => compare_strings(msg, ast_str, fb_str),
    }
}

fn compare_string_lits(
    ast_lit: &ast::StringLit,
    fb_lit: &Option<fbast::StringLiteral>,
) -> Result<(), String> {
    let fb_lit = unwrap_or_fail("string literal", fb_lit)?;
    compare_base(&ast_lit.base, &fb_lit.base_node())?;
    compare_strings("string literal value", &ast_lit.value, &fb_lit.value())?;
    Ok(())
}

fn compare_base(ast_base: &ast::BaseNode, fb_base: &Option<fbast::BaseNode>) -> Result<(), String> {
    let fb_base = unwrap_or_fail("base node", fb_base)?;
    compare_loc(&ast_base.location, &fb_base.loc())?;
    compare_base_errs(&ast_base.errors, &fb_base.errors())?;
    Ok(())
}

fn compare_loc(
    ast_loc: &ast::SourceLocation,
    fb_loc: &Option<fbast::SourceLocation>,
) -> Result<(), String> {
    let fb_loc = unwrap_or_fail("source location", fb_loc)?;
    compare_opt_strings("source location file", &ast_loc.file, &fb_loc.file())?;
    compare_pos(&ast_loc.start, &fb_loc.start())?;
    compare_pos(&ast_loc.end, &fb_loc.end())?;
    compare_opt_strings("source location source", &ast_loc.source, &fb_loc.source())?;
    Ok(())
}

fn compare_pos(ast_pos: &ast::Position, fb_pos: &Option<&fbast::Position>) -> Result<(), String> {
    let fb_pos = unwrap_or_fail("position", fb_pos)?;
    if ast_pos.line != fb_pos.line() as u32 {
        return Err(String::from(format!(
            "ast line position is {}, fb is {}",
            ast_pos.line,
            fb_pos.line()
        )));
    }
    if ast_pos.column != fb_pos.column() as u32 {
        return Err(String::from(format!(
            "ast column position is {}, fb is {}",
            ast_pos.line,
            fb_pos.line()
        )));
    }
    Ok(())
}

fn compare_base_errs(
    ast_errs: &Vec<String>,
    fb_errs: &Option<flatbuffers::Vector<flatbuffers::ForwardsUOffset<&str>>>,
) -> Result<(), String> {
    let fb_errs = unwrap_or_fail("base errors", fb_errs)?;
    compare_vec_len(ast_errs, fb_errs)?;
    let mut i: usize = 0;
    loop {
        if i >= fb_errs.len() {
            break Ok(());
        }

        let fb_err = fb_errs.get(i);
        let ast_err = &ast_errs[i];
        compare_strings("base error", ast_err, &Some(fb_err))?;
        i = i + 1;
    }
}

fn ast_operator(fb_op: fbast::Operator) -> ast::Operator {
    match fb_op {
        fbast::Operator::MultiplicationOperator => ast::Operator::MultiplicationOperator,
        fbast::Operator::DivisionOperator => ast::Operator::DivisionOperator,
        fbast::Operator::ModuloOperator => ast::Operator::ModuloOperator,
        fbast::Operator::PowerOperator => ast::Operator::PowerOperator,
        fbast::Operator::AdditionOperator => ast::Operator::AdditionOperator,
        fbast::Operator::SubtractionOperator => ast::Operator::SubtractionOperator,
        fbast::Operator::LessThanEqualOperator => ast::Operator::LessThanEqualOperator,
        fbast::Operator::LessThanOperator => ast::Operator::LessThanOperator,
        fbast::Operator::GreaterThanEqualOperator => ast::Operator::GreaterThanEqualOperator,
        fbast::Operator::GreaterThanOperator => ast::Operator::GreaterThanOperator,
        fbast::Operator::StartsWithOperator => ast::Operator::StartsWithOperator,
        fbast::Operator::InOperator => ast::Operator::InOperator,
        fbast::Operator::NotOperator => ast::Operator::NotOperator,
        fbast::Operator::ExistsOperator => ast::Operator::ExistsOperator,
        fbast::Operator::NotEmptyOperator => ast::Operator::NotEmptyOperator,
        fbast::Operator::EmptyOperator => ast::Operator::EmptyOperator,
        fbast::Operator::EqualOperator => ast::Operator::EqualOperator,
        fbast::Operator::NotEqualOperator => ast::Operator::NotEqualOperator,
        fbast::Operator::RegexpMatchOperator => ast::Operator::RegexpMatchOperator,
        fbast::Operator::NotRegexpMatchOperator => ast::Operator::NotRegexpMatchOperator,
        fbast::Operator::InvalidOperator => ast::Operator::InvalidOperator,
    }
}

fn ast_logical_operator(lo: &fbast::LogicalOperator) -> ast::LogicalOperator {
    match lo {
        fbast::LogicalOperator::AndOperator => ast::LogicalOperator::AndOperator,
        fbast::LogicalOperator::OrOperator => ast::LogicalOperator::OrOperator,
    }
}
