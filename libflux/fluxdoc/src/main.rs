#![cfg_attr(feature = "strict", deny(warnings))]

use std::collections::HashMap;
use std::fs::{self, File};
use std::io::Write;
use std::path::{Path, PathBuf};
use structopt::StructOpt;
use tera::{Context, Tera};

#[macro_use]
extern crate lazy_static;

#[macro_use]
extern crate serde_derive;

mod doc;
use crate::doc::*;

use flux::ast;
use flux::parser::Parser;
use flux::semantic;
use flux::semantic::types::{PolyType, TvarKinds};

#[derive(Debug, StructOpt)]
struct Args {
    // The root path of packages for which to generate documentation
    #[structopt(parse(from_os_str), long)]
    pkg: PathBuf,
    // The name of the file to write the documentation JSON data
    #[structopt(parse(from_os_str), long, default_value = "")]
    json: PathBuf,
    // The name of the directory into which to write the documentation html files
    #[structopt(parse(from_os_str), long, default_value = "")]
    html: PathBuf,
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::from_args();
    let pkg = walk_pkg(&args.pkg, &args.pkg)?;

    if args.json != Path::new("") {
        let f = File::create(args.json)?;
        serde_json::to_writer(f, &pkg)?;
    }

    if args.html != Path::new("") {
        write_home(&args.html)?;
        return write_html(&args.html, &pkg);
    }
    Ok(())
}

// Walks the directory and generates docs for the package found at topdir and any sub packages.
fn walk_pkg(topdir: &Path, dir: &Path) -> Result<DocPackage, Box<dyn std::error::Error>> {
    let mut packages = Vec::<DocPackage>::new();
    let mut src = Vec::<PathBuf>::new();
    for entry in fs::read_dir(dir)? {
        let entry = entry?;
        let path = entry.path();
        if path.is_dir() {
            let pkg = walk_pkg(topdir, &path)?;
            packages.push(pkg);
            continue;
        }
        match path.extension() {
            Some(ext) => {
                if ext != "flux" {
                    continue;
                }
            }
            None => {
                continue;
            }
        }
        src.push(path.clone());
    }
    let pkgpath = dir.strip_prefix(topdir.parent().unwrap())?;
    generate_docs(&pkgpath, src, packages)
}

// Generates the docs by parsing the sources and checking type inference.
fn generate_docs(
    pkgpath: &Path,
    srcs: Vec<PathBuf>,
    mut packages: Vec<DocPackage>,
) -> Result<DocPackage, Box<dyn std::error::Error>> {
    // determine path vector
    let mut path: Vec<String> = Vec::new();
    let mut curr = pkgpath;
    loop {
        if curr == Path::new("") {
            break;
        }
        path.push(curr.file_name().unwrap().to_str().unwrap().to_string());
        if let Some(parent) = curr.parent() {
            curr = parent
        }
    }
    path.reverse();

    // parse each src in the package
    let mut pkg: Option<ast::Package> = None;
    for src in srcs {
        let source = fs::read_to_string(&src)?;
        let file_name = src.file_name().unwrap();
        let mut p = Parser::new(source.as_str());
        let mut fpkg: ast::Package = p.parse_file(file_name.to_str().unwrap().to_owned()).into();
        // skip test packages
        if !fpkg.package.ends_with("_test") {
            match pkg {
                None => pkg = Some(fpkg),
                Some(ref mut pkg) => {
                    if let Some(err) = flux::merge_packages(&mut fpkg, pkg) {
                        return Err(Box::new(err));
                    }
                }
            }
        }
    }
    packages.sort_by_key(|p| p.name.clone());

    // construct the package documentation
    if let Some(pkg) = pkg {
        // use type inference to determine types of all values
        let sem_pkg = flux::analyze(pkg.clone())?;
        let types = pkg_types(&sem_pkg);
        let mut values: Vec<DocValue> = Vec::new();
        values.sort_by_key(|v| v.name.clone());
        let mut doc = String::new();
        for f in &pkg.files {
            let vs = generate_values(&f, &types, &path)?;
            values.extend(vs);
            if let Some(comment) = &f.package {
                doc = comments_to_string(&comment.base.comments);
            }
        }
        Ok(DocPackage {
            path,
            name: pkg.package,
            doc,
            values,
            packages,
        })
    } else {
        let name = path.last().unwrap().clone();
        Ok(DocPackage {
            path,
            name,
            doc: "".to_string(),
            values: vec![],
            packages,
        })
    }
}

// Produces a map of identifiers to their type for a package.
fn pkg_types(pkg: &semantic::nodes::Package) -> HashMap<String, PolyType> {
    let mut types: HashMap<String, PolyType> = HashMap::new();
    for f in &pkg.files {
        for s in &f.body {
            match s {
                semantic::nodes::Statement::Variable(s) => {
                    let typ = s.init.type_of();
                    types.insert(
                        s.id.name.clone(),
                        PolyType {
                            vars: vec![],
                            cons: TvarKinds::new(),
                            expr: typ,
                        },
                    );
                }
                semantic::nodes::Statement::Builtin(s) => {
                    types.insert(s.id.name.clone(), s.typ_expr.clone());
                }
                semantic::nodes::Statement::Option(s) => {
                    if let semantic::nodes::Assignment::Variable(v) = &s.assignment {
                        let typ = v.init.type_of();
                        types.insert(
                            v.id.name.clone(),
                            PolyType {
                                vars: vec![],
                                cons: TvarKinds::new(),
                                expr: typ,
                            },
                        );
                    }
                }
                _ => {}
            }
        }
    }
    types
}

// Generates docs for the values in a given source file.
fn generate_values(
    f: &ast::File,
    types: &HashMap<String, PolyType>,
    pkgpath: &[String],
) -> Result<Vec<DocValue>, Box<dyn std::error::Error>> {
    let mut values: Vec<DocValue> = Vec::new();
    for stmt in &f.body {
        match stmt {
            ast::Statement::Variable(s) => {
                let doc = comments_to_string(&s.id.base.comments);
                let name = s.id.name.clone();
                let typ = format!("{}", types[&name].normal());
                values.push(DocValue {
                    pkgpath: pkgpath.to_vec(),
                    name,
                    doc,
                    typ,
                });
            }
            ast::Statement::Builtin(s) => {
                let doc = comments_to_string(&s.base.comments);
                let name = s.id.name.clone();
                let typ = format!("{}", types[&name].normal());
                values.push(DocValue {
                    pkgpath: pkgpath.to_vec(),
                    name,
                    doc,
                    typ,
                });
            }
            ast::Statement::Option(s) => {
                if let ast::Assignment::Variable(v) = &s.assignment {
                    let doc = comments_to_string(&s.base.comments);
                    let name = v.id.name.clone();
                    let typ = format!("{}", types[&name].normal());
                    values.push(DocValue {
                        pkgpath: pkgpath.to_vec(),
                        name,
                        doc,
                        typ,
                    });
                }
            }
            _ => {}
        }
    }
    Ok(values)
}

fn comments_to_string(comments: &[ast::Comment]) -> String {
    let mut s = String::new();
    if !comments.is_empty() {
        for c in comments {
            s.push_str(c.text.as_str().strip_prefix("//").unwrap());
        }
    }
    comrak::markdown_to_html(s.as_str(), &comrak::ComrakOptions::default())
}

lazy_static! {
    pub static ref TEMPLATES: Tera = {
        let mut tera = match Tera::new("fluxdoc/templates/*.html") {
            Ok(t) => t,
            Err(e) => {
                println!("Parsing error(s): {}", e);
                ::std::process::exit(1);
            }
        };
        tera.autoescape_on(vec!["html"]);
        tera
    };
}

// Write out a tree of html like this
// pkgRoot
//      index.html -- contains pkgRoot description and index
//      valuea.html -- contains value A description
//      subpkgA
//          index.html -- Contains subpkgA index
//          valueb.html -- contains value B description
fn write_html(dir: &Path, pkg: &DocPackage) -> Result<(), Box<dyn std::error::Error>> {
    let pkgdir = dir.join(&pkg.name);
    fs::create_dir(&pkgdir)?;
    let mut f = File::create(pkgdir.join("index.html"))?;
    let data = TEMPLATES.render("package.html", &Context::from_serialize(&pkg)?)?;
    f.write_all(data.as_bytes())?;
    for v in &pkg.values {
        let mut vf = File::create(pkgdir.join(format!("{}.html", v.name)))?;
        let data = TEMPLATES.render("value.html", &Context::from_serialize(&v)?)?;
        vf.write_all(data.as_bytes())?;
    }
    for p in &pkg.packages {
        write_html(&pkgdir, &p)?;
    }
    Ok(())
}
// Render home.html template
fn write_home(dir: &Path) -> Result<(), Box<dyn std::error::Error>> {
    let ctx = Context::new();
    let data = TEMPLATES.render("home.html", &ctx)?;
    let mut f = File::create(dir.join("index.html"))?;
    f.write_all(data.as_bytes())?;
    Ok(())
}
