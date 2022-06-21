use crate::{
    ast,
    parser::parse_string,
    semantic::{convert, nodes},
};

pub fn compile(source: &str) -> nodes::Package {
    let file = parse_string("".to_string(), source);
    ast::check::check(ast::walk::Node::File(&file)).expect("errors on parsing");
    let ast_pkg = ast::Package {
        base: file.base.clone(),
        path: "".to_string(),
        package: "main".to_string(),
        files: vec![file],
    };
    convert::convert_package(&ast_pkg, &Default::default()).unwrap()
}
