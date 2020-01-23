use flux::ast;
use flux::parser::parse_string;
use flux::semantic::nodes;
use libstd;
use std::io::{self, Read};

fn main() {
    let mut src = String::new();
    match io::stdin().read_to_string(&mut src) {
        Ok(_) => {}
        Err(err) => {
            print!("{}\n", err);
            std::process::exit(1);
        }
    };
    match convert_source(&src) {
        Ok(_) => {}
        Err(err) => {
            print!("{}\n", err);
            std::process::exit(2);
        }
    };
}

pub fn convert_source(source: &str) -> Result<nodes::Package, flux::Error> {
    let file = parse_string("", source);
    let errs = ast::check::check(ast::walk::Node::File(&file));
    if !errs.is_empty() {
        return Err(flux::Error::from(format!("Parsing Error: {:?}", errs)));
    }
    let ast_pkg: ast::Package = file.into();
    libstd::analyze(ast_pkg)
}
