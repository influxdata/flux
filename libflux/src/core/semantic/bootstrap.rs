use std::collections::HashSet;
use std::fs;
use std::io;
use std::path::{Path, PathBuf};

use crate::ast;
use crate::parser;
use crate::semantic::builtins::builtins;
use crate::semantic::convert::convert_file;
use crate::semantic::convert::convert_polytype;
use crate::semantic::env::Environment;
use crate::semantic::fresh::Fresher;
use crate::semantic::import::Importer;
use crate::semantic::infer;
use crate::semantic::infer::Constraints;
use crate::semantic::nodes;
use crate::semantic::nodes::infer_file;
use crate::semantic::sub::Substitutable;
use crate::semantic::types;
use crate::semantic::types::{
    MaxTvar, MonoType, PolyType, PolyTypeMap, PolyTypeMapMap, Property, Record, SemanticMap, Tvar,
    TvarKinds,
};

use walkdir::WalkDir;

const PRELUDE: [&str; 2] = ["universe", "influxdata/influxdb"];

type AstFileMap = SemanticMap<String, ast::File>;

#[derive(Debug, PartialEq)]
pub struct Error {
    pub msg: String,
}

impl From<io::Error> for Error {
    fn from(err: io::Error) -> Error {
        Error {
            msg: format!("{:?}", err),
        }
    }
}

impl From<nodes::Error> for Error {
    fn from(err: nodes::Error) -> Error {
        Error {
            msg: err.to_string(),
        }
    }
}

impl From<types::Error> for Error {
    fn from(err: types::Error) -> Error {
        Error {
            msg: err.to_string(),
        }
    }
}

impl From<infer::Error> for Error {
    fn from(err: infer::Error) -> Error {
        Error {
            msg: err.to_string(),
        }
    }
}

impl From<String> for Error {
    fn from(msg: String) -> Error {
        Error { msg }
    }
}

impl From<&str> for Error {
    fn from(msg: &str) -> Error {
        Error {
            msg: msg.to_string(),
        }
    }
}

#[allow(clippy::type_complexity)]
// Infer the types of the standard library returning two importers, one for the prelude
// and one for the standard library, as well as a type variable fresher.
pub fn infer_stdlib() -> Result<(PolyTypeMap, PolyTypeMap, Fresher, Vec<String>), Error> {
    let (builtins, mut f) = builtin_types()?;

    let dir = "../../../stdlib";
    let files = file_map(parse_flux_files(dir)?);
    let rerun_if_changed = compute_file_dependencies(dir);

    let (prelude, importer) = infer_pre(&mut f, &files, &builtins)?;
    let importer = infer_std(&mut f, &files, &builtins, prelude.clone(), importer)?;

    Ok((prelude, importer, f, rerun_if_changed))
}

fn compute_file_dependencies(root: &str) -> Vec<String> {
    // Iterate through each ast file and canonicalize the
    // file path to an absolute path.
    // Canonicalize the root path to the absolute directory.
    let rootpath = std::env::current_dir()
        .unwrap()
        .join(root)
        .canonicalize()
        .unwrap();
    WalkDir::new(rootpath)
        .into_iter()
        .filter_map(|r| r.ok())
        .filter(|r| r.path().is_dir() || (r.path().is_file() && r.path().ends_with(".flux")))
        .map(|r| path_to_string(r.path()))
        .collect()
}

fn path_to_string(path: &Path) -> String {
    path.to_str().expect("valid path").to_string()
}

#[allow(clippy::type_complexity)]
fn builtin_types() -> Result<(PolyTypeMapMap, Fresher), Error> {
    let mut tv = Tvar(0);
    let mut ty = PolyTypeMapMap::new();
    for (path, values) in builtins().iter() {
        for (name, expr) in values {
            let mut p = parser::Parser::new(expr);
            let expr = convert_polytype(p.parse_type_expression(), &mut Fresher::default())?;

            let tvar = expr.max_tvar();
            if tvar > tv {
                tv = tvar;
            }

            ty.entry((*path).to_string())
                .or_insert_with(PolyTypeMap::new)
                .insert((*name).to_string(), expr);
        }
    }
    Ok((ty, Fresher::from(tv.0 + 1)))
}

#[allow(clippy::type_complexity)]
fn infer_pre<I: Importer>(
    f: &mut Fresher,
    files: &AstFileMap,
    builtin: &SemanticMap<String, I>,
) -> Result<(PolyTypeMap, PolyTypeMap), Error> {
    let mut prelude = PolyTypeMap::new();
    let mut imports = PolyTypeMap::new();
    for name in &PRELUDE {
        let (types, importer) = infer_pkg(name, f, files, builtin, PolyTypeMap::new(), imports)?;
        for (k, v) in types {
            prelude.insert(k, v);
        }
        imports = importer;
    }
    Ok((prelude, imports))
}

#[allow(clippy::type_complexity)]
fn infer_std<I: Importer>(
    f: &mut Fresher,
    files: &AstFileMap,
    builtin: &SemanticMap<String, I>,
    prelude: PolyTypeMap,
    mut imports: PolyTypeMap,
) -> Result<PolyTypeMap, Error> {
    for (path, _) in files.iter() {
        if imports.contains_key(path) {
            continue;
        }
        let (types, mut importer) = infer_pkg(path, f, files, builtin, prelude.clone(), imports)?;
        importer.insert(path.to_string(), build_polytype(types, f)?);
        imports = importer;
    }
    Ok(imports)
}

// Recursively parse all flux files within a directory.
fn parse_flux_files(path: &str) -> io::Result<Vec<ast::File>> {
    let mut files = Vec::new();
    let entries = WalkDir::new(PathBuf::from(path))
        .into_iter()
        .filter_map(|r| r.ok())
        .filter(|r| r.path().is_file());

    for entry in entries {
        if let Some(path) = entry.path().to_str() {
            if path.ends_with(".flux") && !path.ends_with("_test.flux") {
                files.push(parser::parse_string(
                    path.rsplitn(2, "/stdlib/").collect::<Vec<&str>>()[0],
                    &fs::read_to_string(path)?,
                ));
            }
        }
    }
    Ok(files)
}

// Associates an import path with each file
fn file_map(files: Vec<ast::File>) -> AstFileMap {
    files.into_iter().fold(AstFileMap::new(), |mut acc, file| {
        let name = file.name.rsplitn(2, '/').collect::<Vec<&str>>()[1].to_string();
        acc.insert(name, file);
        acc
    })
}

fn imports(file: &ast::File) -> Vec<&str> {
    let mut dependencies = Vec::new();
    for import in &file.imports {
        dependencies.push(&import.path.value[..]);
    }
    dependencies
}

// Determines the dependencies of a package. That is, all packages
// that must be evaluated before the package in question. Each
// dependency is added to the `deps` vector in evaluation order.
#[allow(clippy::type_complexity)]
fn dependencies<'a>(
    name: &'a str,
    pkgs: &'a AstFileMap,
    mut deps: Vec<&'a str>,
    mut seen: HashSet<&'a str>,
    mut done: HashSet<&'a str>,
) -> Result<(Vec<&'a str>, HashSet<&'a str>, HashSet<&'a str>), Error> {
    if seen.contains(name) && !done.contains(name) {
        Err(Error {
            msg: format!(r#"package "{}" depends on itself"#, name),
        })
    } else {
        seen.insert(name);
        match pkgs.get(name) {
            None => Err(Error {
                msg: format!(r#"package "{}" not found"#, name),
            }),
            Some(file) => {
                for name in imports(file) {
                    let (x, y, z) = dependencies(name, pkgs, deps, seen, done)?;
                    deps = x;
                    seen = y;
                    done = z;
                    if !deps.contains(&name) {
                        deps.push(name);
                    }
                }
                done.insert(name);
                Ok((deps, seen, done))
            }
        }
    }
}

// Constructs a polytype, or more specifically a generic row type, from a hash map
pub fn build_polytype(from: PolyTypeMap, f: &mut Fresher) -> Result<PolyType, Error> {
    let (r, cons) = build_row(from, f);
    let mut kinds = TvarKinds::new();
    let sub = infer::solve(&cons, &mut kinds, f)?;
    Ok(infer::generalize(
        &Environment::empty(false),
        &kinds,
        MonoType::Record(Box::new(r)).apply(&sub),
    ))
}

fn build_row(from: PolyTypeMap, f: &mut Fresher) -> (Record, Constraints) {
    let mut r = Record::Empty;
    let mut cons = Constraints::empty();

    for (name, poly) in from {
        let (ty, constraints) = infer::instantiate(
            poly.clone(),
            f,
            ast::SourceLocation {
                file: None,
                start: ast::Position::default(),
                end: ast::Position::default(),
                source: None,
            },
        );
        r = Record::Extension {
            head: Property { k: name, v: ty },
            tail: MonoType::Record(Box::new(r)),
        };
        cons = cons + constraints;
    }
    (r, cons)
}

// Infer the types in a package(file), returning a hash map containing
// the inferred types along with a possibly updated map of package imports.
//
#[allow(clippy::type_complexity)]
fn infer_pkg<I: Importer>(
    name: &str,                       // name of package to infer
    f: &mut Fresher,                  // type variable fresher
    files: &AstFileMap,               // files available for inference
    builtin: &SemanticMap<String, I>, // builtin types
    prelude: PolyTypeMap,             // prelude types
    imports: PolyTypeMap,             // types available for import
) -> Result<
    (
        PolyTypeMap, // inferred types
        PolyTypeMap, // types available for import (possibly updated)
    ),
    Error,
> {
    // Determine the order in which we must infer dependencies
    let (deps, _, _) = dependencies(name, files, Vec::new(), HashSet::new(), HashSet::new())?;

    let mut imports = imports;

    // Infer all dependencies
    for pkg in deps {
        if imports.import(pkg).is_none() {
            let file = files.get(pkg);
            if file.is_none() {
                return Err(Error {
                    msg: format!(r#"package "{}" not found"#, pkg),
                });
            }
            let file = file.unwrap().to_owned();

            let env = if let Some(builtins) = builtin.get(pkg) {
                infer_file(
                    &mut convert_file(file, f)?,
                    Environment::new(prelude.clone().into()),
                    f,
                    &imports,
                    builtins,
                )?
                .0
            } else {
                infer_file(
                    &mut convert_file(file, f)?,
                    Environment::new(prelude.clone().into()),
                    f,
                    &imports,
                    &None,
                )?
                .0
            };

            imports.insert(pkg.to_string(), build_polytype(env.values, f)?);
        }
    }

    let file = files.get(name);
    if file.is_none() {
        return Err(Error {
            msg: format!("package '{}' not found", name),
        });
    }
    let file = file.unwrap().to_owned();

    let env = if let Some(builtins) = builtin.get(name) {
        infer_file(
            &mut convert_file(file, f)?,
            Environment::new(prelude.into()),
            f,
            &imports,
            builtins,
        )?
        .0
    } else {
        infer_file(
            &mut convert_file(file, f)?,
            Environment::new(prelude.into()),
            f,
            &imports,
            &None,
        )?
        .0
    };

    Ok((env.values, imports))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::ast::get_err_type_expression;
    use crate::parser;
    use crate::parser::parse_string;
    use crate::semantic::env::Environment;

    #[test]
    fn infer_program() -> Result<(), Error> {
        let a = r#"
            f = (x) => x
        "#;
        let b = r#"
            import "a"

            builtin x : int

            y = a.f(x: x)
        "#;
        let c = r#"
            import "b"

            z = b.y
        "#;
        let files = semantic_map! {
            String::from("a") => parse_string("a.flux", a),
            String::from("b") => parse_string("b.flux", b),
            String::from("c") => parse_string("c.flux", c),
        };
        let builtins = semantic_map! {
            String::from("b") => Environment::from(semantic_map! {
                String::from("x") => {
                // parse("forall [] int")?
                    let mut p = parser::Parser::new("int");
                    let typ_expr = p.parse_type_expression();
                    let err = get_err_type_expression(typ_expr.clone());
                    if err != "" {
                        let msg = format!(
                            "TypeExpression parsing failed for int. {:?}", err
                        );
                        panic!(msg)
                    }
                    convert_polytype(typ_expr, &mut Fresher::default())?
                },
            }),
        };
        let (types, imports) = infer_pkg(
            "c",
            &mut Fresher::from(1),
            &files,
            &builtins,
            PolyTypeMap::new(),
            PolyTypeMap::new(),
        )?;

        let want = semantic_map! {
            String::from("z") => {
                    let mut p = parser::Parser::new("int");
                    let typ_expr = p.parse_type_expression();
                    let err = get_err_type_expression(typ_expr.clone());
                    if err != "" {
                        let msg = format!(
                            "TypeExpression parsing failed for int. {:?}", err
                        );
                        panic!(msg)
                    }
                    convert_polytype(typ_expr, &mut Fresher::default())?
            },
        };
        if want != types {
            return Err(Error {
                msg: format!(
                    "unexpected inference result:\n\nwant: {:?}\n\ngot: {:?}",
                    want, types
                ),
            });
        }

        let want = semantic_map! {
            String::from("a") => {
            let mut p = parser::Parser::new("{f: (x: A) => A}");
                    let typ_expr = p.parse_type_expression();
                    let err = get_err_type_expression(typ_expr.clone());
                    if err != "" {
                        let msg = format!(
                            "TypeExpression parsing failed for int. {:?}", err
                        );
                        panic!(msg)
                    }
                    convert_polytype(typ_expr, &mut Fresher::default())?
            },
            String::from("b") => {
            let mut p = parser::Parser::new("{x: int , y: int}");
                    let typ_expr = p.parse_type_expression();
                    let err = get_err_type_expression(typ_expr.clone());
                    if err != "" {
                        let msg = format!(
                            "TypeExpression parsing failed for int. {:?}", err
                        );
                        panic!(msg)
                    }
                    convert_polytype(typ_expr, &mut Fresher::default())?
            },
        };
        if want != imports {
            return Err(Error {
                msg: format!(
                    "unexpected type importer:\n\nwant: {:?}\n\ngot: {:?}",
                    want, types
                ),
            });
        }

        Ok(())
    }

    #[test]
    fn prelude_dependencies() {
        let files = file_map(parse_flux_files("../../../stdlib").unwrap());

        let r = PRELUDE.iter().try_fold(
            (Vec::new(), HashSet::new(), HashSet::new()),
            |(deps, seen, done), name| dependencies(name, &files, deps, seen, done),
        );

        let names = r.unwrap().0;

        assert_eq!(vec!["system", "date", "math", "strings", "regexp"], names,);
    }

    #[test]
    fn cyclic_dependency() {
        let a = r#"
            import "b"
        "#;
        let b = r#"
            import "a"
        "#;
        let files = semantic_map! {
            String::from("a") => parse_string("a.flux", a),
            String::from("b") => parse_string("b.flux", b),
        };
        let got_err = dependencies("b", &files, Vec::new(), HashSet::new(), HashSet::new())
            .expect_err("expected cyclic dependency error");

        assert_eq!(
            Error {
                msg: r#"package "b" depends on itself"#.to_string()
            },
            got_err
        );
    }
}
