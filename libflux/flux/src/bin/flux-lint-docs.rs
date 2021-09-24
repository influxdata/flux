use std::env;
use std::fs;
use std::path::Path;

use flux::semantic::doc::parse_doc_comments;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<String> = env::args().collect();
    if args.len() != 2 {
        panic!("must pass arg")
    }
    let fpath = Path::new(&args[1]);
    let fname = fpath
        .file_name()
        .expect("Must provide a path to a Flux file");
    let contents = fs::read_to_string(fpath)?;
    let ast_pkg = flux::parse(fname.to_str().unwrap().to_owned(), &contents);
    let types = flux::analyze_to_map(ast_pkg.clone())?;
    let (_doc, mut diags) = parse_doc_comments(&ast_pkg, fpath.to_str().unwrap(), &types)?;
    if !diags.is_empty() {
        let limit = 10;
        let rest = diags.len() as i64 - limit as i64;
        println!("Found {} diagnostic errors", diags.len());
        diags.truncate(limit);
        for d in diags {
            println!("{}", d);
        }
        if rest > 0 {
            println!("Hidding the remaining {} diagnostics", rest);
        }
    }
    Ok(())
}
