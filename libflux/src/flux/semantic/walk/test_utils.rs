use crate::ast;
use crate::parser::parse_string;
use crate::semantic::convert;
use crate::semantic::nodes;

pub fn compile(source: &str) -> nodes::Package {
    let file = parse_string("", source);
    let errs = ast::check::check(ast::walk::Node::File(&file));
    if errs.len() > 0 {
        panic!(format!("got errors on parsing: {:?}", errs));
    }
    let ast_pkg = ast::Package {
        base: file.base.clone(),
        path: "".to_string(),
        package: "main".to_string(),
        files: vec![file],
    };
    convert::convert(ast_pkg).unwrap()
}
