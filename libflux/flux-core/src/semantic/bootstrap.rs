//! Bootstrap provides an API for compiling the Flux standard library.
//!
//! This package does not assume a location of the source code but does assume which packages are
//! part of the prelude.

use std::{collections::HashSet, env::consts, fs, io, io::Write, path::Path};

use anyhow::{anyhow, bail, Result};
use libflate::gzip::Encoder;
use walkdir::WalkDir;

use crate::{
    ast, parser,
    semantic::{
        convert::convert_package,
        env::Environment,
        flatbuffers::types::{build_module, finish_serialize},
        fs::{FileSystemImporter, StdFS},
        import::Importer,
        infer,
        infer::Constraints,
        nodes,
        nodes::{infer_package, inject_pkg_types, Package},
        sub::{Substitutable, Substitution},
        types::{MonoType, PolyType, PolyTypeMap, Property, Record, SemanticMap, Tvar, TvarKinds},
        ExternalEnvironment,
    },
};

// List of packages to include into the Flux prelude
const PRELUDE: [&str; 3] = ["internal/boolean", "universe", "influxdata/influxdb"];

/// A mapping of package import paths to the corresponding AST package.
pub type ASTPackageMap = SemanticMap<String, ast::Package>;
/// A mapping of package import paths to the corresponding semantic graph package.
pub type SemanticPackageMap = SemanticMap<String, Package>;

/// Infers the Flux standard library given the path to the source code.
/// The prelude and the imports are returned.
#[allow(clippy::type_complexity)]
pub fn infer_stdlib_dir(path: &Path) -> Result<(PolyTypeMap, PolyTypeMap, SemanticPackageMap)> {
    let mut sub = Substitution::default();

    let ast_packages = parse_dir(path)?;

    let (prelude, importer) = infer_pre(&mut sub, &ast_packages)?;
    let (imports, sem_pkg_map) = infer_std(&mut sub, &ast_packages, prelude.clone(), importer)?;

    Ok((prelude, imports, sem_pkg_map))
}

fn infer_pre(
    sub: &mut Substitution,
    ast_packages: &ASTPackageMap,
) -> Result<(PolyTypeMap, PolyTypeMap)> {
    let mut prelude_map = PolyTypeMap::new();
    let mut imports = PolyTypeMap::new();
    for name in PRELUDE {
        // Infer each package in the prelude allowing the earlier packages to be used by later
        // packages within the prelude list.
        let (types, importer, _sem_pkg) =
            infer_pkg(name, sub, ast_packages, prelude_map.clone(), imports)?;
        for (k, v) in types {
            prelude_map.insert(k, v);
        }
        imports = importer;
    }
    Ok((prelude_map, imports))
}

#[allow(clippy::type_complexity)]
fn infer_std(
    sub: &mut Substitution,
    ast_packages: &ASTPackageMap,
    prelude: PolyTypeMap,
    mut imports: PolyTypeMap,
) -> Result<(PolyTypeMap, SemanticPackageMap)> {
    let mut sem_pkg_map = SemanticPackageMap::new();
    for (path, _) in ast_packages.iter() {
        let (types, mut importer, sem_pkg) =
            infer_pkg(path, sub, ast_packages, prelude.clone(), imports.clone())?;
        sem_pkg_map.insert(path.to_string(), sem_pkg);
        if !imports.contains_key(path) {
            importer.insert(path.to_string(), build_polytype(types, sub)?);
            imports = importer;
        }
    }
    Ok((imports, sem_pkg_map))
}

/// Recursively parse all flux files within a directory.
pub fn parse_dir(dir: &Path) -> io::Result<ASTPackageMap> {
    let mut files = Vec::new();
    let entries = WalkDir::new(dir)
        .into_iter()
        .filter_map(|r| r.ok())
        .filter(|r| r.path().is_file());

    let is_windows = consts::OS == "windows";

    for entry in entries {
        if let Some(path) = entry.path().to_str() {
            if path.ends_with(".flux") && !path.ends_with("_test.flux") {
                let mut normalized_path = path.to_string();
                if is_windows {
                    // When building on Windows, the paths generated by WalkDir will
                    // use `\` instead of `/` as their separator. It's easier to normalize
                    // the separators to always be `/` here than it is to change the
                    // rest of this buildscript & the flux runtime initialization logic
                    // to work with either separator.
                    normalized_path = normalized_path.replace("\\", "/");
                }
                files.push(parser::parse_string(
                    normalized_path
                        .rsplitn(2, "/stdlib/")
                        .collect::<Vec<&str>>()[0]
                        .to_owned(),
                    &fs::read_to_string(entry.path())?,
                ));
            }
        }
    }
    Ok(ast_map(files))
}

// Associates an import path with each file
fn ast_map(files: Vec<ast::File>) -> ASTPackageMap {
    files
        .into_iter()
        .fold(ASTPackageMap::new(), |mut acc, file| {
            let path = file.name.rsplitn(2, '/').collect::<Vec<&str>>()[1].to_string();
            acc.insert(
                path.clone(),
                ast::Package {
                    base: ast::BaseNode {
                        ..ast::BaseNode::default()
                    },
                    path,
                    package: String::from(file.get_package()),
                    files: vec![file],
                },
            );
            acc
        })
}

fn imports(pkg: &ast::Package) -> Vec<&str> {
    let mut dependencies = Vec::new();
    for file in &pkg.files {
        for import in &file.imports {
            dependencies.push(&import.path.value[..]);
        }
    }
    dependencies
}

// Determines the dependencies of a package. That is, all packages
// that must be evaluated before the package in question. Each
// dependency is added to the `deps` vector in evaluation order.
#[allow(clippy::type_complexity)]
fn dependencies<'a>(
    name: &'a str,
    pkgs: &'a ASTPackageMap,
    mut deps: Vec<&'a str>,
    mut seen: HashSet<&'a str>,
    mut done: HashSet<&'a str>,
) -> Result<(Vec<&'a str>, HashSet<&'a str>, HashSet<&'a str>)> {
    if seen.contains(name) && !done.contains(name) {
        Err(anyhow!(r#"package "{}" depends on itself"#, name))
    } else {
        seen.insert(name);
        match pkgs.get(name) {
            None => Err(anyhow!(r#"package "{}" not found"#, name)),
            Some(pkg) => {
                for name in imports(pkg) {
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

/// Constructs a polytype, or more specifically a generic record type, from a hash map.
pub fn build_polytype(from: PolyTypeMap, sub: &mut Substitution) -> Result<PolyType> {
    let (r, cons) = build_record(from, sub);
    infer::solve(&cons, sub)?;
    let typ = MonoType::record(r).apply(sub);
    Ok(infer::generalize(
        &Environment::empty(false),
        sub.cons(),
        typ,
    ))
}

fn build_record(from: PolyTypeMap, sub: &mut Substitution) -> (Record, Constraints) {
    let mut r = Record::Empty;
    let mut cons = Constraints::empty();

    for (name, poly) in from {
        let (ty, constraints) = infer::instantiate(
            poly.clone(),
            sub,
            ast::SourceLocation {
                file: None,
                start: ast::Position::default(),
                end: ast::Position::default(),
                source: None,
            },
        );
        r = Record::Extension {
            head: Property { k: name, v: ty },
            tail: MonoType::record(r),
        };
        cons += constraints;
    }
    (r, cons)
}

// Infer the types in a package(file), returning a hash map containing
// the inferred types along with a possibly updated map of package imports.
//
#[allow(clippy::type_complexity)]
fn infer_pkg(
    name: &str,                   // name of package to infer
    sub: &mut Substitution,       // type variable substitution
    ast_packages: &ASTPackageMap, // ast_packages available for inference
    prelude: PolyTypeMap,         // prelude types
    imports: PolyTypeMap,         // types available for import
) -> Result<(
    PolyTypeMap, // inferred types
    PolyTypeMap, // types available for import (possibly updated)
    Package,     // semantic graph
)> {
    // Determine the order in which we must infer dependencies
    let (deps, _, _) = dependencies(
        name,
        ast_packages,
        Vec::new(),
        HashSet::new(),
        HashSet::new(),
    )?;
    let mut imports = imports;

    // Infer all dependencies
    for pkg in deps {
        if imports.import(pkg).is_none() {
            let file = ast_packages.get(pkg);
            if file.is_none() {
                bail!(r#"package import "{}" not found"#, pkg);
            }
            let file = file.unwrap().to_owned();

            let env = Environment::new(prelude.clone().into());
            let env = infer_package(
                &mut convert_package(file, &env, sub)?,
                env,
                sub,
                &mut imports,
            )?;

            imports.insert(pkg.to_string(), build_polytype(env.string_values(), sub)?);
        }
    }

    let file = ast_packages.get(name);
    if file.is_none() {
        bail!(r#"package "{}" not found"#, name);
    }
    let file = file.unwrap().to_owned();

    let env = Environment::new(prelude.into());
    let mut sem_pkg = convert_package(file, &env, sub)?;
    let env = infer_package(&mut sem_pkg, env, sub, &mut imports)?;
    sem_pkg = inject_pkg_types(sem_pkg, sub);

    Ok((env.string_values(), imports, sem_pkg))
}

fn stdlib_importer(path: &Path) -> FileSystemImporter<StdFS> {
    let fs = StdFS::new(path);
    FileSystemImporter::new(fs)
}

fn prelude_from_importer<I>(importer: &mut I) -> Result<ExternalEnvironment>
where
    I: Importer,
{
    let mut env = PolyTypeMap::new();
    for pkg in PRELUDE {
        if let Some(pkg_type) = importer.import(pkg) {
            if let MonoType::Record(typ) = pkg_type.expr {
                add_record_to_map(&mut env, typ.as_ref(), &pkg_type.vars, &pkg_type.cons)?;
            } else {
                bail!("package type is not a record");
            }
        } else {
            bail!("prelude package {} not found", pkg);
        }
    }
    Ok(env.into())
}

fn add_record_to_map(
    env: &mut PolyTypeMap,
    r: &Record,
    free_vars: &[Tvar],
    cons: &TvarKinds,
) -> Result<()> {
    match r {
        Record::Empty => Ok(()),
        Record::Extension { head, tail } => {
            let new_vars = head.v.free_vars();
            let mut new_cons = TvarKinds::new();
            for var in &new_vars {
                if !free_vars.iter().any(|v| v == var) {
                    bail!("monotype contains free var not in poly type free vars");
                }
                if let Some(con) = cons.get(var) {
                    new_cons.insert(*var, con.clone());
                }
            }
            env.insert(
                head.k.clone(),
                PolyType {
                    vars: new_vars,
                    cons: new_cons,
                    expr: head.v.clone(),
                },
            );
            match tail {
                MonoType::Record(r) => add_record_to_map(env, r, free_vars, cons),
                _ => Ok(()),
            }
        }
    }
}

/// Stdlib returns the prelude and importer for the Flux standard library given a path to a
/// compiled directory structure.
pub fn stdlib(dir: &Path) -> Result<(ExternalEnvironment, FileSystemImporter<StdFS>)> {
    let mut stdlib_importer = stdlib_importer(dir);
    let prelude = prelude_from_importer(&mut stdlib_importer)?;
    Ok((prelude, stdlib_importer))
}

/// Compiles the stdlib found at the srcdir into the outdir.
pub fn compile_stdlib(srcdir: &Path, outdir: &Path) -> Result<()> {
    let (_, imports, mut sem_pkgs) = infer_stdlib_dir(srcdir)?;
    // Write each file as compiled module
    for (path, pt) in &imports {
        if let Some(code) = sem_pkgs.remove(path) {
            let module = Module {
                polytype: Some(pt.clone()),
                code: Some(code),
            };
            let mut builder = flatbuffers::FlatBufferBuilder::new();
            let offset = build_module(&mut builder, module);
            let buf = finish_serialize(&mut builder, offset);

            // Write module contents to file
            let mut fpath = outdir.join(path);
            fpath.set_extension("fc");
            fs::create_dir_all(fpath.parent().unwrap())?;
            let file = fs::File::create(&fpath)?;
            let mut encoder = Encoder::new(file)?;
            encoder.write_all(buf)?;
            encoder.finish().into_result()?;
        } else {
            bail!("package {} missing code", &path);
        }
    }
    Ok(())
}

/// Module represenets the result of compiling Flux source code.
///
/// The polytype represents the type of the entire package as a record type.
/// The record properties represent the exported values from the package.
///
/// The package is the actual code of the package that can be used to execute the package.
///
/// This struct is experimental we anticipate it will change as we build more systems around
/// the concepts of modules.
pub struct Module {
    /// The polytype
    pub polytype: Option<PolyType>,
    /// The code
    pub code: Option<nodes::Package>,
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{
        ast::get_err_type_expression, parser, parser::parse_string,
        semantic::convert::convert_polytype,
    };

    #[test]
    fn infer_program() -> Result<()> {
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
        let ast_packages: ASTPackageMap = semantic_map! {
            String::from("a") => parse_string("a.flux".to_string(), a).into(),
            String::from("b") => parse_string("b.flux".to_string(), b).into(),
            String::from("c") => parse_string("c.flux".to_string(), c).into(),
        };
        let (types, imports, _) = infer_pkg(
            "c",
            &mut Substitution::default(),
            &ast_packages,
            PolyTypeMap::new(),
            PolyTypeMap::new(),
        )?;

        let want = semantic_map! {
            String::from("z") => {
                    let mut p = parser::Parser::new("int");
                    let typ_expr = p.parse_type_expression();
                    let err = get_err_type_expression(typ_expr.clone());
                    if err != "" {
                        panic!(
                            "TypeExpression parsing failed for int. {:?}", err
                        );
                    }
                    convert_polytype(typ_expr, &mut Substitution::default())?
            },
        };
        if want != types {
            bail!(
                "unexpected inference result:\n\nwant: {:?}\n\ngot: {:?}",
                want,
                types,
            );
        }

        let want = semantic_map! {
            String::from("a") => {
            let mut p = parser::Parser::new("{f: (x: A) => A}");
                    let typ_expr = p.parse_type_expression();
                    let err = get_err_type_expression(typ_expr.clone());
                    if err != "" {
                        panic!(
                            "TypeExpression parsing failed for int. {:?}", err
                        );
                    }
                    convert_polytype(typ_expr, &mut Substitution::default())?
            },
            String::from("b") => {
            let mut p = parser::Parser::new("{x: int , y: int}");
                    let typ_expr = p.parse_type_expression();
                    let err = get_err_type_expression(typ_expr.clone());
                    if err != "" {
                        panic!(
                            "TypeExpression parsing failed for int. {:?}", err
                        );
                    }
                    convert_polytype(typ_expr, &mut Substitution::default())?
            },
        };
        if want != imports {
            bail!(
                "unexpected type importer:\n\nwant: {:?}\n\ngot: {:?}",
                want,
                types,
            );
        }

        Ok(())
    }

    #[test]
    fn cyclic_dependency() {
        let a = r#"
            import "b"
        "#;
        let b = r#"
            import "a"
        "#;
        let ast_packages: ASTPackageMap = semantic_map! {
            String::from("a") => parse_string("a.flux".to_string(), a).into(),
            String::from("b") => parse_string("b.flux".to_string(), b).into(),
        };
        let got_err = dependencies(
            "b",
            &ast_packages,
            Vec::new(),
            HashSet::new(),
            HashSet::new(),
        )
        .expect_err("expected cyclic dependency error");

        assert_eq!(
            r#"package "b" depends on itself"#.to_string(),
            got_err.to_string(),
        );
    }
}
