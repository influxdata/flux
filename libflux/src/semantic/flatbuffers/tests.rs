extern crate flatbuffers; 

use super::fbsemantic::fbsemantic::*; 

#[test]
fn test_flatbuffers_semantic() {
    let mut builder = flatbuffers::FlatBufferBuilder::new_with_capacity(256);

    // Testing out a unary expression using a float
    let floatval = FloatLiteral::create(
        &mut builder,
        &FloatLiteralArgs {
            value: 3.5,
            ..FloatLiteralArgs::default()
        },
    );

    let increment = UnaryExpression::create(
        &mut builder,
        &UnaryExpressionArgs {
            operator: Operator::SubtractionOperator,
            argument: Expression::FloatLiteral,
            ..UnaryExpressionArgs::default()
        },
    );

    let statement = ExpressionStatement::create(
        &mut builder,
        &ExpressionStatementArgs {
            expression_type: Expression::UnaryExpression,
            expression: Some(add.as_union_value()),
            ..ExpressionStatementArgs::default()
        },
    );

    let wrappedStatement = WrappedStatement::create(
        &mut builder,
        &WrappedStatementArgs {
            statement_type: Statement::ExpressionStatement,
            statement: Some(statement.as_union_value()),
        },
    );

    let statements = b.create_vector(&[wrappedStatement]);

    let file = File::create(
        &mut builder,
        &FileArgs {
            body: Some(statements),
            ..FileArgs::default()
        },
    );

    let files = b.create_vector(&[file]);

    let pkg = Package::create(
        &mut builder,
        &PackageArgs {
            files: Some(files),
            ..PackageArgs::default()
        },
    );

    builder.finish(pkg, None);
    let bytes = builder.finished_data();
    assert_ne!(bytes.len(), 0);
}
