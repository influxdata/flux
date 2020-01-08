extern crate flatbuffers;
extern crate walkdir;

use super::semantic_generated::fbsemantic;
use crate::ast;
use crate::semantic;
use crate::semantic::convert;
use chrono::FixedOffset;

#[test]
fn test_serialize() {
    let f = vec![
        crate::parser::parse_string(
            "test1",
            r#"
package testpkg
import other "my_other_pkg"
import "yet_another_pkg"
option now = () => (2030-01-01T00:00:00Z)
option foo.bar = "baz"
builtin foo

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
    let mut pkg = match convert::convert(pkg) {
        Ok(pkg) => pkg,
        Err(e) => {
            assert!(false, e);
            return;
        }
    };
    let (vec, offset) = match super::serialize(&mut pkg) {
        Ok((v, o)) => (v, o),
        Err(e) => {
            assert!(false, e);
            return;
        }
    };
    let fb = &vec.as_slice()[offset..];
    match compare_semantic_fb(&pkg, fb) {
        Err(e) => assert!(false, e),
        _ => (),
    }
}

fn compare_semantic_fb(semantic_pkg: &semantic::nodes::Package, fb: &[u8]) -> Result<(), String> {
    let fb_pkg = fbsemantic::get_root_as_package(fb);
    compare_pkg_fb(semantic_pkg, &fb_pkg)?;
    Ok(())
}

fn compare_pkg_fb(
    semantic_pkg: &semantic::nodes::Package,
    fb_pkg: &fbsemantic::Package,
) -> Result<(), String> {
    compare_loc(&semantic_pkg.loc, &fb_pkg.loc())?;
    compare_strings("package name", &semantic_pkg.package, &fb_pkg.package())?;

    let fb_files = &fb_pkg.files();
    let fb_files = unwrap_or_fail("package files", fb_files)?;
    compare_vec_len(&semantic_pkg.files, &fb_files)?;
    let mut i: usize = 0;
    loop {
        if i >= semantic_pkg.files.len() {
            return Ok(());
        }

        let semantic_file = &semantic_pkg.files[i];
        let fb_file = &fb_files.get(i);
        compare_files(semantic_file, fb_file)?;
        i = i + 1;
    }
}

fn compare_files(
    semantic_file: &semantic::nodes::File,
    fb_file: &fbsemantic::File,
) -> Result<(), String> {
    compare_loc(&semantic_file.loc, &fb_file.loc())?;
    let sem_file_ref = &semantic_file.package.as_ref();
    if let Some(package) = sem_file_ref {
        let semantic_file_name = &sem_file_ref.unwrap().name.name;
        let fb_file_name = &fb_file.package().unwrap().name().unwrap().name();
        compare_strings("file name", &semantic_file_name, fb_file_name)?;
        compare_package_clause(&semantic_file.package, &fb_file.package())?;
    }
    compare_imports(&semantic_file.imports, &fb_file.imports())?;
    compare_stmt_vectors(&semantic_file.body, &fb_file.body())?;
    Ok(())
}

fn compare_stmt_vectors(
    semantic_stmts: &Vec<semantic::nodes::Statement>,
    fb_stmts: &Option<
        flatbuffers::Vector<flatbuffers::ForwardsUOffset<fbsemantic::WrappedStatement>>,
    >,
) -> Result<(), String> {
    let fb_stmts = unwrap_or_fail("statement list", fb_stmts)?;
    compare_vec_len(semantic_stmts, fb_stmts)?;
    let mut i: usize = 0;
    loop {
        if i >= semantic_stmts.len() {
            break Ok(());
        }
        let fb_stmt_ty = fb_stmts.get(i).statement_type();
        let fb_stmt = &fb_stmts.get(i).statement();
        compare_stmts(&semantic_stmts[i], fb_stmt_ty, fb_stmt)?;
        i += 1;
    }
}

fn compare_stmts(
    semantic_stmt: &semantic::nodes::Statement,
    fb_stmt_ty: fbsemantic::Statement,
    fb_stmt: &Option<flatbuffers::Table>,
) -> Result<(), String> {
    let fb_tbl = unwrap_or_fail("statement", fb_stmt)?;
    match (semantic_stmt, fb_stmt_ty) {
        (
            semantic::nodes::Statement::Variable(semantic_stmt),
            fbsemantic::Statement::NativeVariableAssignment,
        ) => {
            let fb_stmt = fbsemantic::NativeVariableAssignment::init_from_table(*fb_tbl);
            compare_var_assign(&semantic_stmt, &Some(fb_stmt))
        }
        (
            semantic::nodes::Statement::Expr(semantic_stmt),
            fbsemantic::Statement::ExpressionStatement,
        ) => {
            let fb_stmt = fbsemantic::ExpressionStatement::init_from_table(*fb_tbl);
            compare_loc(&semantic_stmt.loc, &fb_stmt.loc())?;
            compare_exprs(
                &semantic_stmt.expression,
                fb_stmt.expression_type(),
                &fb_stmt.expression(),
            )?;
            Ok(())
        }
        (
            semantic::nodes::Statement::Option(semantic_stmt),
            fbsemantic::Statement::OptionStatement,
        ) => {
            let fb_stmt = fbsemantic::OptionStatement::init_from_table(*fb_tbl);
            compare_loc(&semantic_stmt.loc, &fb_stmt.loc())?;
            compare_assignments(
                &semantic_stmt.assignment,
                fb_stmt.assignment_type(),
                &fb_stmt.assignment(),
            )
        }
        (
            semantic::nodes::Statement::Return(semantic_stmt),
            fbsemantic::Statement::ReturnStatement,
        ) => {
            let fb_stmt = fbsemantic::ReturnStatement::init_from_table(*fb_tbl);
            compare_loc(&semantic_stmt.loc, &fb_stmt.loc())?;
            compare_exprs(
                &semantic_stmt.argument,
                fb_stmt.argument_type(),
                &fb_stmt.argument(),
            )
        }
        (semantic::nodes::Statement::Test(semantic_stmt), fbsemantic::Statement::TestStatement) => {
            let fb_stmt = fbsemantic::TestStatement::init_from_table(*fb_tbl);
            compare_loc(&semantic_stmt.loc, &fb_stmt.loc())
        }
        (
            semantic::nodes::Statement::Builtin(semantic_stmt),
            fbsemantic::Statement::BuiltinStatement,
        ) => {
            let fb_stmt = fbsemantic::BuiltinStatement::init_from_table(*fb_tbl);
            compare_loc(&semantic_stmt.loc, &fb_stmt.loc())?;
            compare_ids(&semantic_stmt.id, &fb_stmt.id())
        }
        (semantic_stmt, fb_ty) => {
            let semantic_stmt_ty = semantic::walk::Node::from_stmt(semantic_stmt);
            let fb_ty = fbsemantic::enum_name_statement(fb_ty);
            Err(String::from(format!(
                "wrong statement type; semantic = {}, fb = {}",
                semantic_stmt_ty, fb_ty
            )))
        }
    }
}

fn translate_block_to_stmt(sem_block: &semantic::nodes::Block) -> semantic::nodes::Statement {
    match sem_block {
        semantic::nodes::Block::Variable(va, _) => semantic::nodes::Statement::Variable(va.clone()),
        semantic::nodes::Block::Expr(expr, _) => semantic::nodes::Statement::Expr(expr.clone()),
        semantic::nodes::Block::Return(rtn) => semantic::nodes::Statement::Return(rtn.clone()),
    }
}

fn compare_ids(
    semantic_id: &semantic::nodes::Identifier,
    fb_id: &Option<fbsemantic::Identifier>,
) -> Result<(), String> {
    let fb_id = unwrap_or_fail("id", fb_id)?;
    compare_loc(&semantic_id.loc, &fb_id.loc())?;
    compare_strings("id", &semantic_id.name, &fb_id.name())?;
    Ok(())
}

fn compare_id_exprs(
    semantic_id: &semantic::nodes::IdentifierExpr,
    fb_id: &Option<fbsemantic::IdentifierExpression>,
) -> Result<(), String> {
    let fb_id = unwrap_or_fail("id", fb_id)?;
    compare_loc(&semantic_id.loc, &fb_id.loc())?;
    compare_strings("id", &semantic_id.name, &fb_id.name())?;
    Ok(())
}

fn compare_opt_ids(
    semantic_id: &Option<semantic::nodes::Identifier>,
    fb_id: &Option<fbsemantic::Identifier>,
) -> Result<(), String> {
    match (semantic_id, fb_id) {
        (None, None) => Ok(()),
        (Some(_), None) => Err(String::from(
            "compare opt ids, semantic had one, fb did not",
        )),
        (None, Some(_)) => Err(String::from("compare opt ids, semantic had none, fb did")),
        (Some(semantic_id), fb_id) => compare_ids(semantic_id, fb_id),
    }
}

fn compare_opt_expr_ids(
    semantic_id: &Option<semantic::nodes::IdentifierExpr>,
    fb_id: &Option<fbsemantic::IdentifierExpression>,
) -> Result<(), String> {
    match (semantic_id, fb_id) {
        (None, None) => Ok(()),
        (Some(_), None) => Err(String::from(
            "compare opt ids, semantic had one, fb did not",
        )),
        (None, Some(_)) => Err(String::from("compare opt ids, semantic had none, fb did")),
        (Some(semantic_id), fb_id) => compare_id_exprs(semantic_id, fb_id),
    }
}

fn compare_assignments(
    semantic_asgn: &semantic::nodes::Assignment,
    fb_asgn_ty: fbsemantic::Assignment,
    fb_asgn: &Option<flatbuffers::Table>,
) -> Result<(), String> {
    let fb_tbl = unwrap_or_fail("assign", fb_asgn)?;
    match (semantic_asgn, fb_asgn_ty) {
        (
            semantic::nodes::Assignment::Variable(semantic_va),
            fbsemantic::Assignment::NativeVariableAssignment,
        ) => {
            let fb_va = fbsemantic::NativeVariableAssignment::init_from_table(*fb_tbl);
            compare_var_assign(semantic_va, &Some(fb_va))
        }
        (
            semantic::nodes::Assignment::Member(semantic_ma),
            fbsemantic::Assignment::MemberAssignment,
        ) => {
            let fb_ma = fbsemantic::MemberAssignment::init_from_table(*fb_tbl);
            compare_loc(&semantic_ma.loc, &fb_ma.loc())?;
            compare_member_expr(&semantic_ma.member, &fb_ma.member())?;
            compare_exprs(&semantic_ma.init, fb_ma.init__type(), &fb_ma.init_())
        }
        _ => Err(String::from("assignment mismatch")),
    }
}

fn compare_member_expr(
    semantic_me: &semantic::nodes::MemberExpr,
    fb_me: &Option<fbsemantic::MemberExpression>,
) -> Result<(), String> {
    let fb_me = unwrap_or_fail("member expression", fb_me)?;
    compare_loc(&semantic_me.loc, &fb_me.loc())?;
    compare_exprs(&semantic_me.object, fb_me.object_type(), &fb_me.object())
}

fn compare_var_assign(
    semantic_va: &semantic::nodes::VariableAssgn,
    fb_va: &Option<fbsemantic::NativeVariableAssignment>,
) -> Result<(), String> {
    let fb_va = unwrap_or_fail("var assign", fb_va)?;
    compare_loc(&semantic_va.loc, &fb_va.loc())?;
    compare_ids(&semantic_va.id, &fb_va.identifier())?;
    compare_exprs(&semantic_va.init, fb_va.init__type(), &fb_va.init_())
}

fn compare_exprs(
    semantic_expr: &semantic::nodes::Expression,
    fb_expr_ty: fbsemantic::Expression,
    fb_tbl: &Option<flatbuffers::Table>,
) -> Result<(), String> {
    let fb_tbl = unwrap_or_fail("expr", fb_tbl)?;
    match (semantic_expr, fb_expr_ty) {
        (
            semantic::nodes::Expression::Integer(semantic_int),
            fbsemantic::Expression::IntegerLiteral,
        ) => {
            let fb_int = fbsemantic::IntegerLiteral::init_from_table(*fb_tbl);
            compare_loc(&semantic_expr.loc(), &fb_int.loc())?;
            match semantic_int.value == fb_int.value() {
                true => Ok(()),
                false => Err(String::from(format!(
                    "int lit mismatch; semantic = {}, fb = {}",
                    semantic_int.value,
                    fb_int.value()
                ))),
            }
        }
        (
            semantic::nodes::Expression::Float(semantic_float),
            fbsemantic::Expression::FloatLiteral,
        ) => {
            let fb_float = fbsemantic::FloatLiteral::init_from_table(*fb_tbl);
            compare_loc(&semantic_float.loc, &fb_float.loc())?;
            match semantic_float.value == fb_float.value() {
                true => Ok(()),
                false => Err(String::from(format!(
                    "float lit mismatch; semantic = {}, fb = {}",
                    semantic_float.value,
                    fb_float.value()
                ))),
            }
        }
        (
            semantic::nodes::Expression::StringLit(semantic_string),
            fbsemantic::Expression::StringLiteral,
        ) => {
            let fb_string = fbsemantic::StringLiteral::init_from_table(*fb_tbl);
            compare_loc(&semantic_string.loc, &fb_string.loc())?;
            let fb_value = fb_string.value();
            let fb_value = unwrap_or_fail("string lit string", &fb_value)?;
            match &semantic_string.value.as_str() == fb_value {
                true => Ok(()),
                false => Err(String::from(format!(
                    "string lit mismatch; semantic = {}, fb = {}",
                    semantic_string.value, fb_value,
                ))),
            }
        }
        (
            semantic::nodes::Expression::Duration(semantic_dur),
            fbsemantic::Expression::DurationLiteral,
        ) => {
            let fb_dur_lit = fbsemantic::DurationLiteral::init_from_table(*fb_tbl);
            compare_loc(&semantic_dur.loc, &fb_dur_lit.loc())?;
            let fb_val = fb_dur_lit.value();
            let fb_val_unwrap = unwrap_or_fail("dur lit values", &fb_val)?;
            let fb_d = fb_val_unwrap.get(0);
            compare_duration(&semantic_dur.value, &fb_d)?;
            Ok(())
        }
        (
            semantic::nodes::Expression::DateTime(semantic_dtl),
            fbsemantic::Expression::DateTimeLiteral,
        ) => {
            let fb_dtl = fbsemantic::DateTimeLiteral::init_from_table(*fb_tbl);
            let fb_dtl_val = fb_dtl.value().unwrap();
            let dtl = chrono::DateTime::<FixedOffset>::from_utc(
                chrono::NaiveDateTime::from_timestamp(fb_dtl_val.secs(), fb_dtl_val.nsecs()),
                FixedOffset::east(fb_dtl_val.offset()),
            );
            compare_loc(&semantic_dtl.loc, &fb_dtl.loc())?;
            if semantic_dtl.value != dtl {
                return Err(String::from("invalid DateTimeLiteral value"));
            }
            Ok(())
        }
        (
            semantic::nodes::Expression::Regexp(semantic_rel),
            fbsemantic::Expression::RegexpLiteral,
        ) => {
            let fb_rel = fbsemantic::RegexpLiteral::init_from_table(*fb_tbl);
            compare_loc(&semantic_rel.loc, &fb_rel.loc())?;
            compare_strings("regexp lit value", &semantic_rel.value, &fb_rel.value())?;
            Ok(())
        }
        (
            semantic::nodes::Expression::Identifier(semantic_id),
            fbsemantic::Expression::IdentifierExpression,
        ) => {
            let fb_id = fbsemantic::IdentifierExpression::init_from_table(*fb_tbl);
            compare_id_exprs(semantic_id, &Some(fb_id))?;
            Ok(())
        }
        (
            semantic::nodes::Expression::Array(semantic_ae),
            fbsemantic::Expression::ArrayExpression,
        ) => {
            let fb_ae = fbsemantic::ArrayExpression::init_from_table(*fb_tbl);
            compare_loc(&semantic_ae.loc, &fb_ae.loc())?;
            let fb_elems = fb_ae.elements();
            let fb_elems = unwrap_or_fail("array elems", &fb_elems)?;
            compare_vec_len(&semantic_ae.elements, fb_elems)?;

            let mut i: usize = 0;
            loop {
                if i >= semantic_ae.elements.len() {
                    break Ok(());
                }
                let fb_we = &fb_elems.get(i);
                let fb_e = &fb_we.expression();
                compare_exprs(&semantic_ae.elements[i], fb_we.expression_type(), fb_e)?;
                i = i + 1
            }
        }
        (
            semantic::nodes::Expression::Function(semantic_fe),
            fbsemantic::Expression::FunctionExpression,
        ) => {
            let fb_fe = fbsemantic::FunctionExpression::init_from_table(*fb_tbl);
            compare_loc(&semantic_fe.loc, &fb_fe.loc())?;
            compare_params(&semantic_fe.params, &fb_fe.params())?;

            // compare function bodies
            compare_loc(&semantic_fe.body.loc(), &fb_fe.body().unwrap().loc());
            let mut block_len: usize = 0;
            let mut current_sem = &semantic_fe.body;
            let fb_list = fb_fe.body().unwrap().body().unwrap();
            loop {
                compare_stmts(
                    &translate_block_to_stmt(current_sem),
                    fb_list.get(block_len).statement_type(),
                    &fb_list.get(block_len).statement(),
                )?;

                match current_sem {
                    semantic::nodes::Block::Expr(_, next)
                    | semantic::nodes::Block::Variable(_, next) => {
                        current_sem = next.as_ref();
                    }
                    semantic::nodes::Block::Return(_) => {
                        break;
                    }
                }
                block_len += 1;
            }
            Ok(())
        }
        (
            semantic::nodes::Expression::Logical(semantic_le),
            fbsemantic::Expression::LogicalExpression,
        ) => {
            let fb_le = fbsemantic::LogicalExpression::init_from_table(*fb_tbl);
            compare_loc(&semantic_le.loc, &fb_le.loc())?;
            compare_exprs(&semantic_le.left, fb_le.left_type(), &fb_le.left())?;
            compare_exprs(&semantic_le.right, fb_le.right_type(), &fb_le.right())?;
            match semantic_logical_operator(&fb_le.operator()) == semantic_le.operator {
                true => Ok(()),
                false => Err(String::from("logical operator mismatch")),
            }
        }
        (
            semantic::nodes::Expression::Object(semantic_oe),
            fbsemantic::Expression::ObjectExpression,
        ) => {
            let fb_oe = fbsemantic::ObjectExpression::init_from_table(*fb_tbl);
            compare_loc(&semantic_oe.loc, &fb_oe.loc())?;
            compare_property_list(&semantic_oe.properties, &fb_oe.properties())?;
            compare_opt_expr_ids(&semantic_oe.with, &fb_oe.with())
        }
        (
            semantic::nodes::Expression::Member(semantic_me),
            fbsemantic::Expression::MemberExpression,
        ) => {
            let fb_me = fbsemantic::MemberExpression::init_from_table(*fb_tbl);
            compare_member_expr(&semantic_me, &Some(fb_me))
        }
        (
            semantic::nodes::Expression::Index(semantic_ie),
            fbsemantic::Expression::IndexExpression,
        ) => {
            let fb_ie = fbsemantic::IndexExpression::init_from_table(*fb_tbl);
            compare_loc(&semantic_ie.loc, &fb_ie.loc())?;
            compare_exprs(&semantic_ie.array, fb_ie.array_type(), &fb_ie.array())?;
            compare_exprs(&semantic_ie.index, fb_ie.index_type(), &fb_ie.index())
        }
        (
            semantic::nodes::Expression::Binary(semantic_be),
            fbsemantic::Expression::BinaryExpression,
        ) => {
            let fb_be = fbsemantic::BinaryExpression::init_from_table(*fb_tbl);
            compare_loc(&semantic_be.loc, &fb_be.loc())?;
            compare_exprs(&semantic_be.left, fb_be.left_type(), &fb_be.left())?;
            compare_exprs(&semantic_be.right, fb_be.right_type(), &fb_be.right())?;
            match semantic_operator(fb_be.operator()) == semantic_be.operator {
                true => Ok(()),
                false => Err(String::from("binary operator mismatch")),
            }
        }
        (
            semantic::nodes::Expression::Unary(semantic_ue),
            fbsemantic::Expression::UnaryExpression,
        ) => {
            let fb_ue = fbsemantic::UnaryExpression::init_from_table(*fb_tbl);
            compare_loc(&semantic_ue.loc, &fb_ue.loc())?;
            compare_exprs(
                &semantic_ue.argument,
                fb_ue.argument_type(),
                &fb_ue.argument(),
            )?;
            match semantic_operator(fb_ue.operator()) == semantic_ue.operator {
                true => Ok(()),
                false => Err(String::from("unary operator mismatch")),
            }
        }
        (
            semantic::nodes::Expression::Call(semantic_ce),
            fbsemantic::Expression::CallExpression,
        ) => {
            let fb_ce = fbsemantic::CallExpression::init_from_table(*fb_tbl);
            compare_call_exprs(&semantic_ce, &Some(fb_ce))
        }
        (
            semantic::nodes::Expression::Conditional(semantic_ce),
            fbsemantic::Expression::ConditionalExpression,
        ) => {
            let fb_ce = fbsemantic::ConditionalExpression::init_from_table(*fb_tbl);
            compare_loc(&semantic_ce.loc, &fb_ce.loc())?;
            compare_exprs(&semantic_ce.test, fb_ce.test_type(), &fb_ce.test())?;
            compare_exprs(
                &semantic_ce.consequent,
                fb_ce.consequent_type(),
                &fb_ce.consequent(),
            )?;
            compare_exprs(
                &semantic_ce.alternate,
                fb_ce.alternate_type(),
                &fb_ce.alternate(),
            )
        }
        (
            semantic::nodes::Expression::StringExpr(semantic_se),
            fbsemantic::Expression::StringExpression,
        ) => {
            let fb_se = fbsemantic::StringExpression::init_from_table(*fb_tbl);
            compare_loc(&semantic_se.loc, &fb_se.loc())?;
            compare_string_expr_part_list(&semantic_se.parts, &fb_se.parts())
        }
        (semantic_expr, fb_expr_ty) => {
            let semantic_expr_ty = semantic::walk::Node::from_expr(semantic_expr);
            let fb_ty = fbsemantic::enum_name_expression(fb_expr_ty);
            Err(String::from(format!(
                "wrong expr type; semantic = {} {}, fb = {}",
                semantic_expr_ty,
                semantic_expr_ty.loc(),
                fb_ty,
            )))
        }
    }
}

fn compare_duration(
    semantic_dur: &semantic::nodes::Duration,
    fb_dur: &fbsemantic::Duration,
) -> Result<(), String> {
    if semantic_dur.months != fb_dur.months() {
        return Err(String::from(format!(
            "duration months do not match; semantic = {}, fb = {}",
            semantic_dur.months,
            fb_dur.months()
        )));
    }

    if semantic_dur.nanoseconds != fb_dur.nanoseconds() {
        return Err(String::from(format!(
            "duration nanoseconds do not match; semantic = {}, fb = {}",
            semantic_dur.nanoseconds,
            fb_dur.nanoseconds()
        )));
    }

    if semantic_dur.negative != fb_dur.negative() {
        return Err(String::from(format!(
            "durations are not the same sign; semantic is negative = {}, fb is negative = {}",
            semantic_dur.negative,
            fb_dur.negative()
        )));
    }
    Ok(())
}

fn compare_property_list(
    semantic_pl: &Vec<semantic::nodes::Property>,
    fb_pl: &Option<flatbuffers::Vector<flatbuffers::ForwardsUOffset<fbsemantic::Property>>>,
) -> Result<(), String> {
    let fb_pl = unwrap_or_fail("property list", fb_pl)?;
    compare_vec_len(semantic_pl, fb_pl)?;
    let mut i: usize = 0;
    loop {
        if i >= semantic_pl.len() {
            return Ok(());
        }

        compare_property(&semantic_pl[i], &fb_pl.get(i))?;
        i = i + 1;
    }
}

fn compare_params(
    semantic_params: &Vec<semantic::nodes::FunctionParameter>,
    fb_params: &Option<
        flatbuffers::Vector<flatbuffers::ForwardsUOffset<fbsemantic::FunctionParameter>>,
    >,
) -> Result<(), String> {
    let fb_params = unwrap_or_fail("params", fb_params)?;
    compare_vec_len(semantic_params, fb_params)?;
    let mut i: usize = 0;
    loop {
        if i >= semantic_params.len() {
            return Ok(());
        }

        compare_param(&semantic_params[i], &fb_params.get(i))?;
        i = i + 1;
    }
}

fn compare_param(
    semantic_param: &semantic::nodes::FunctionParameter,
    fb_param: &fbsemantic::FunctionParameter,
) -> Result<(), String> {
    compare_loc(&semantic_param.loc, &fb_param.loc())?;
    if semantic_param.is_pipe != fb_param.is_pipe() {
        return Err(format!(
            "mismatch: semantic: {}, fb: {}",
            semantic_param.is_pipe,
            fb_param.is_pipe()
        ));
    }
    compare_ids(&semantic_param.key, &fb_param.key());
    if let Some(def) = &semantic_param.default {
        compare_exprs(&def, fb_param.default_type(), &fb_param.default());
    }
    Ok(())
}

fn compare_property(
    semantic_prop: &semantic::nodes::Property,
    fb_prop: &fbsemantic::Property,
) -> Result<(), String> {
    compare_loc(&semantic_prop.loc, &fb_prop.loc())?;
    compare_ids(&semantic_prop.key, &fb_prop.key());
    let expr_prop = &semantic_prop.value;
    compare_exprs(&expr_prop, fb_prop.value_type(), &fb_prop.value())
}

fn compare_package_clause(
    semantic_pkg_clause: &Option<semantic::nodes::PackageClause>,
    fb_pkg_clause: &Option<fbsemantic::PackageClause>,
) -> Result<(), String> {
    let (semantic_pkg_clause, fb_pkg_clause) = match (semantic_pkg_clause, fb_pkg_clause) {
        (None, None) => return Ok(()),
        (None, Some(_)) => return Err(String::from("found package clause where not expected")),
        (Some(_), None) => return Err(String::from("missing package clause")),
        (Some(ac), Some(fc)) => (ac, fc),
    };
    compare_loc(&semantic_pkg_clause.loc, &fb_pkg_clause.loc())?;
    compare_ids(&semantic_pkg_clause.name, &fb_pkg_clause.name())?;
    Ok(())
}

fn compare_imports(
    semantic_imports: &Vec<semantic::nodes::ImportDeclaration>,
    fb_imports: &Option<
        flatbuffers::Vector<flatbuffers::ForwardsUOffset<fbsemantic::ImportDeclaration>>,
    >,
) -> Result<(), String> {
    let fb_imports = unwrap_or_fail("imports", fb_imports)?;
    compare_vec_len(semantic_imports, fb_imports)?;
    let mut i: usize = 0;
    loop {
        if i >= semantic_imports.len() {
            break Ok(());
        }

        compare_import_decls(&semantic_imports[i], &fb_imports.get(i))?;
        i = i + 1;
    }
}

fn compare_import_decls(
    semantic_id: &semantic::nodes::ImportDeclaration,
    fb_id: &fbsemantic::ImportDeclaration,
) -> Result<(), String> {
    compare_opt_ids(&semantic_id.alias, &fb_id.alias())?;
    compare_string_lits(&semantic_id.path, &fb_id.path())?;
    Ok(())
}

fn compare_loc(
    semantic_loc: &ast::SourceLocation,
    fb_loc: &Option<fbsemantic::SourceLocation>,
) -> Result<(), String> {
    let fb_loc = unwrap_or_fail("source location", fb_loc)?;
    compare_opt_strings("source location file", &semantic_loc.file, &fb_loc.file())?;
    compare_pos(&semantic_loc.start, &fb_loc.start())?;
    compare_pos(&semantic_loc.end, &fb_loc.end())?;
    compare_opt_strings(
        "source location source",
        &semantic_loc.source,
        &fb_loc.source(),
    )?;
    Ok(())
}

fn compare_vec_len<T, U>(
    semantic_vec: &Vec<T>,
    fb_vec: &flatbuffers::Vector<U>,
) -> Result<(), String> {
    match semantic_vec.len() == fb_vec.len() {
        true => Ok(()),
        false => Err(String::from(format!(
            "vectors have different lengths: semantic = {}, fb = {}",
            semantic_vec.len(),
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

fn compare_strings(msg: &str, semantic_str: &String, fb_str: &Option<&str>) -> Result<(), String> {
    let fb_str = unwrap_or_fail("string", fb_str)?;
    if semantic_str.as_str() != *fb_str {
        return Err(format!(
            "{} mismatch: semantic: {}, fb: {}",
            msg, semantic_str, fb_str
        ));
    };
    Ok(())
}

fn compare_opt_strings(
    msg: &str,
    semantic_str: &Option<String>,
    fb_str: &Option<&str>,
) -> Result<(), String> {
    match (semantic_str, fb_str) {
        (None, None) => return Ok(()),
        (None, Some(s)) => Err(String::from(format!(
            "comparing opt string for {}: semantic had none, fb had {}",
            msg, s,
        ))),
        (Some(s), None) => Err(String::from(format!(
            "comparing opt string for {}: semantic had {}, fb had none",
            msg, s,
        ))),
        (Some(semantic_str), fb_str) => compare_strings(msg, semantic_str, fb_str),
    }
}

fn compare_pos(
    semantic_pos: &ast::Position,
    fb_pos: &Option<&fbsemantic::Position>,
) -> Result<(), String> {
    let fb_pos = unwrap_or_fail("position", fb_pos)?;
    if semantic_pos.line != fb_pos.line() as u32 {
        return Err(String::from(format!(
            "semantic line position is {}, fb is {}",
            semantic_pos.line,
            fb_pos.line()
        )));
    }
    if semantic_pos.column != fb_pos.column() as u32 {
        return Err(String::from(format!(
            "semantic column position is {}, fb is {}",
            semantic_pos.column,
            fb_pos.column()
        )));
    }
    Ok(())
}

fn compare_string_lits(
    semantic_lit: &semantic::nodes::StringLit,
    fb_lit: &Option<fbsemantic::StringLiteral>,
) -> Result<(), String> {
    let fb_lit = unwrap_or_fail("string literal", fb_lit)?;
    compare_loc(&semantic_lit.loc, &fb_lit.loc())?;
    compare_strings("string literal value", &semantic_lit.value, &fb_lit.value())?;
    Ok(())
}

fn compare_opt_exprs(
    semantic_expr: &Option<semantic::nodes::Expression>,
    fb_expr_ty: fbsemantic::Expression,
    fb_expr: &Option<flatbuffers::Table>,
) -> Result<(), String> {
    match (semantic_expr, fb_expr_ty) {
        (None, fbsemantic::Expression::NONE) => Ok(()),
        (None, _) => Err(String::from("expected no expr but got one")),
        (Some(_), fbsemantic::Expression::NONE) => {
            Err(String::from("expected an expr but got none"))
        }
        (Some(semantic_expr), _) => compare_exprs(semantic_expr, fb_expr_ty, fb_expr),
    }
}

fn compare_call_exprs(
    semantic_ce: &semantic::nodes::CallExpr,
    fb_ce: &Option<fbsemantic::CallExpression>,
) -> Result<(), String> {
    let fb_ce = unwrap_or_fail("call expr", fb_ce)?;
    compare_loc(&semantic_ce.loc, &fb_ce.loc())?;
    compare_exprs(&semantic_ce.callee, fb_ce.callee_type(), &fb_ce.callee())?;
    let fb_args = fb_ce.arguments().unwrap();
    let mut index = 0;
    loop {
        if index >= semantic_ce.arguments.len() {
            break;
        }
        compare_property(&semantic_ce.arguments[index], &fb_args.get(index));
        index += 1;
    }
    Ok(())
}

fn compare_string_expr_part_list(
    semantic_parts: &Vec<semantic::nodes::StringExprPart>,
    fb_parts: &Option<
        flatbuffers::Vector<flatbuffers::ForwardsUOffset<fbsemantic::StringExpressionPart>>,
    >,
) -> Result<(), String> {
    let fb_parts = unwrap_or_fail("string expr parts", fb_parts)?;
    compare_vec_len(semantic_parts, fb_parts)?;
    let mut i: usize = 0;
    loop {
        if i >= semantic_parts.len() {
            break Ok(());
        }

        compare_string_expr_part(&semantic_parts[i], &fb_parts.get(i))?;
        i = i + 1
    }
}

fn compare_string_expr_part(
    semantic_part: &semantic::nodes::StringExprPart,
    fb_part: &fbsemantic::StringExpressionPart,
) -> Result<(), String> {
    match (
        semantic_part,
        fb_part.text_value(),
        fb_part.interpolated_expression_type(),
        fb_part.interpolated_expression(),
    ) {
        (
            semantic::nodes::StringExprPart::Text(semantic_text),
            Some(fb_text),
            fbsemantic::Expression::NONE,
            None,
        ) => {
            compare_loc(&semantic_text.loc, &fb_part.loc())?;
            match semantic_text.value.as_str() == fb_text {
                true => Ok(()),
                false => Err(String::from(
                    "mismatch in value of text part of string expr",
                )),
            }
        }
        (semantic::nodes::StringExprPart::Interpolated(semantic_ip), None, fb_expr_ty, fb_expr) => {
            compare_loc(&semantic_ip.loc, &fb_part.loc())?;
            compare_exprs(&semantic_ip.expression, fb_expr_ty, &fb_expr)
        }
        _ => Err(String::from(
            "mismatch in string expr part text/interpolated",
        )),
    }
}

fn semantic_operator(fb_op: fbsemantic::Operator) -> ast::Operator {
    match fb_op {
        fbsemantic::Operator::MultiplicationOperator => ast::Operator::MultiplicationOperator,
        fbsemantic::Operator::DivisionOperator => ast::Operator::DivisionOperator,
        fbsemantic::Operator::ModuloOperator => ast::Operator::ModuloOperator,
        fbsemantic::Operator::PowerOperator => ast::Operator::PowerOperator,
        fbsemantic::Operator::AdditionOperator => ast::Operator::AdditionOperator,
        fbsemantic::Operator::SubtractionOperator => ast::Operator::SubtractionOperator,
        fbsemantic::Operator::LessThanEqualOperator => ast::Operator::LessThanEqualOperator,
        fbsemantic::Operator::LessThanOperator => ast::Operator::LessThanOperator,
        fbsemantic::Operator::GreaterThanEqualOperator => ast::Operator::GreaterThanEqualOperator,
        fbsemantic::Operator::GreaterThanOperator => ast::Operator::GreaterThanOperator,
        fbsemantic::Operator::StartsWithOperator => ast::Operator::StartsWithOperator,
        fbsemantic::Operator::InOperator => ast::Operator::InOperator,
        fbsemantic::Operator::NotOperator => ast::Operator::NotOperator,
        fbsemantic::Operator::ExistsOperator => ast::Operator::ExistsOperator,
        fbsemantic::Operator::NotEmptyOperator => ast::Operator::NotEmptyOperator,
        fbsemantic::Operator::EmptyOperator => ast::Operator::EmptyOperator,
        fbsemantic::Operator::EqualOperator => ast::Operator::EqualOperator,
        fbsemantic::Operator::NotEqualOperator => ast::Operator::NotEqualOperator,
        fbsemantic::Operator::RegexpMatchOperator => ast::Operator::RegexpMatchOperator,
        fbsemantic::Operator::NotRegexpMatchOperator => ast::Operator::NotRegexpMatchOperator,
        fbsemantic::Operator::InvalidOperator => ast::Operator::InvalidOperator,
    }
}

fn semantic_logical_operator(lo: &fbsemantic::LogicalOperator) -> ast::LogicalOperator {
    match lo {
        fbsemantic::LogicalOperator::AndOperator => ast::LogicalOperator::AndOperator,
        fbsemantic::LogicalOperator::OrOperator => ast::LogicalOperator::OrOperator,
    }
}
