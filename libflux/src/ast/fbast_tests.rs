extern crate flatbuffers;

use super::fbast::fbast::*;

#[test]
fn test_flatbuffers_ast() {
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

    let wrappedStmt = WrappedStatement::create(
        &mut b,
        &WrappedStatementArgs {
            statement_type: Statement::ExpressionStatement,
            statement: Some(stmt.as_union_value()),
        },
    );

    let stmts = b.create_vector(&[wrappedStmt]);

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
