//! Bootstrap provides an API for compiling the Flux standard library.
//!
//! This package does not assume a location of the source code but does assume which packages are
//! part of the prelude.

use std::{env::consts, fs, io, io::Write, path::Path};

use anyhow::{anyhow, bail, Result};
use libflate::gzip::Encoder;
use walkdir::WalkDir;

use crate::{
    ast,
    map::{HashMap, HashSet},
    parser,
    semantic::{
        env::Environment,
        flatbuffers::types::{build_module, finish_serialize},
        fs::{FileSystemImporter, StdFS},
        import::{Importer, Packages},
        nodes::{self, Package, Symbol},
        sub::{Substitutable, Substituter},
        types::{
            MonoType, PolyType, PolyTypeHashMap, Record, RecordLabel, SemanticMap, Tvar, TvarKinds,
        },
        Analyzer, PackageExports,
    },
};

// List of packages to include into the Flux prelude
const PRELUDE: [&str; 4] = [
    "internal/boolean",
    "internal/location",
    "universe",
    "influxdata/influxdb",
];

/// A mapping of package import paths to the corresponding AST package.
pub type ASTPackageMap = SemanticMap<String, ast::Package>;
/// A mapping of package import paths to the corresponding semantic graph package.
pub type SemanticPackageMap = SemanticMap<String, Package>;

/// Infers the Flux standard library given the path to the source code.
/// The prelude and the imports are returned.
#[allow(clippy::type_complexity)]
pub fn infer_stdlib_dir(path: &Path) -> Result<(PackageExports, Packages, SemanticPackageMap)> {
    let ast_packages = parse_dir(path)?;

    let mut infer_state = InferState::default();
    let prelude = infer_state.infer_pre(&ast_packages)?;
    infer_state.infer_std(&ast_packages, &prelude)?;

    Ok((prelude, infer_state.imports, infer_state.sem_pkg_map))
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
                    normalized_path = normalized_path.replace('\\', "/");
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

fn imports(pkg: &ast::Package) -> impl Iterator<Item = &str> {
    pkg.files
        .iter()
        .flat_map(|file| file.imports.iter().map(|import| &import.path.value[..]))
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

#[derive(Default)]
struct InferState {
    // types available for import
    imports: Packages,
    sem_pkg_map: SemanticPackageMap,
}

impl InferState {
    fn infer_pre(&mut self, ast_packages: &ASTPackageMap) -> Result<PackageExports> {
        let mut prelude_map = HashMap::new();
        for name in PRELUDE {
            // Infer each package in the prelude allowing the earlier packages to be used by later
            // packages within the prelude list.
            let (types, _sem_pkg) = self.infer_pkg(
                name,
                ast_packages,
                &PackageExports::try_from(prelude_map.clone())?,
            )?;
            for (k, v) in types.into_bindings() {
                prelude_map.insert(k, v);
            }
        }
        Ok(PackageExports::try_from(prelude_map)?)
    }

    #[allow(clippy::type_complexity)]
    fn infer_std(&mut self, ast_packages: &ASTPackageMap, prelude: &PackageExports) -> Result<()> {
        for (path, _) in ast_packages.iter() {
            // No need to infer the package again if it has already been inferred through a
            // dependency
            if !self.sem_pkg_map.contains_key(path) {
                let (types, sem_pkg) = self.infer_pkg(path, ast_packages, prelude)?;

                self.sem_pkg_map.insert(path.to_string(), sem_pkg);
                if !self.imports.contains_key(path) {
                    self.imports.insert(path.to_string(), types);
                }
            }
        }
        Ok(())
    }

    // Infer the types in a package(file), returning a hash map containing
    // the inferred types along with a possibly updated map of package imports.
    //
    #[allow(clippy::type_complexity)]
    fn infer_pkg(
        &mut self,
        name: &str,                   // name of package to infer
        ast_packages: &ASTPackageMap, // ast_packages available for inference
        prelude: &PackageExports,     // prelude types
    ) -> Result<(
        PackageExports, // inferred types
        Package,        // semantic graph
    )> {
        // Determine the order in which we must infer dependencies
        let (deps, _, _) = dependencies(
            name,
            ast_packages,
            Vec::new(),
            HashSet::new(),
            HashSet::new(),
        )?;

        // Infer all dependencies
        for pkg in deps {
            if self.imports.import(pkg).is_none() {
                let (env, sem_pkg) =
                    self.infer_pkg_with_resolved_deps(pkg, ast_packages, prelude)?;

                self.sem_pkg_map.insert(pkg.to_string(), sem_pkg);
                self.imports.insert(pkg.to_string(), env);
            }
        }

        self.infer_pkg_with_resolved_deps(name, ast_packages, prelude)
    }

    fn infer_pkg_with_resolved_deps(
        &mut self,
        name: &str,                   // name of package to infer
        ast_packages: &ASTPackageMap, // ast_packages available for inference
        prelude: &PackageExports,     // prelude types
    ) -> Result<(
        PackageExports, // inferred types
        Package,        // semantic graph
    )> {
        let file = ast_packages
            .get(name)
            .ok_or_else(|| anyhow!(r#"package import "{}" not found"#, name))?;

        let env = Environment::new(prelude.into());
        let mut analyzer = Analyzer::new_with_defaults(env, &mut self.imports);
        let (exports, sem_pkg) = analyzer
            .analyze_ast(file)
            .map_err(|err| err.error.pretty_error())?;

        Ok((exports, sem_pkg))
    }
}

fn stdlib_importer(path: &Path) -> FileSystemImporter<StdFS> {
    let fs = StdFS::new(path);
    FileSystemImporter::new(fs)
}

fn prelude_from_importer<I>(importer: &mut I) -> Result<PackageExports>
where
    I: Importer,
{
    let mut env = PolyTypeHashMap::new();
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
    let exports = PackageExports::try_from(env)?;
    Ok(exports)
}

// Collects any `MonoType::BoundVar`s in the type
struct CollectBoundVars(Vec<Tvar>);

impl Substituter for CollectBoundVars {
    fn try_apply(&mut self, _var: Tvar) -> Option<MonoType> {
        None
    }

    fn try_apply_bound(&mut self, var: Tvar) -> Option<MonoType> {
        let vars = &mut self.0;
        if let Err(i) = vars.binary_search(&var) {
            vars.insert(i, var);
        }
        None
    }
}

fn add_record_to_map(
    env: &mut PolyTypeHashMap<Symbol>,
    r: &Record,
    free_vars: &[Tvar],
    cons: &TvarKinds,
) -> Result<()> {
    for field in r.fields() {
        let new_vars = {
            let mut new_vars = CollectBoundVars(Vec::new());
            field.v.visit(&mut new_vars);
            new_vars.0
        };

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
            match &field.k {
                RecordLabel::Concrete(s) => s.clone().into(),
                RecordLabel::BoundVariable(_) | RecordLabel::Variable(_) => {
                    bail!("Record contains variable labels")
                }
                RecordLabel::Error => {
                    bail!("Record contains type error")
                }
            },
            PolyType {
                vars: new_vars,
                cons: new_cons,
                expr: field.v.clone(),
            },
        );
    }
    Ok(())
}

/// Stdlib returns the prelude and importer for the Flux standard library given a path to a
/// compiled directory structure.
pub fn stdlib(dir: &Path) -> Result<(PackageExports, FileSystemImporter<StdFS>)> {
    let mut stdlib_importer = stdlib_importer(dir);
    let prelude = prelude_from_importer(&mut stdlib_importer)?;
    Ok((prelude, stdlib_importer))
}

/// Compiles the stdlib found at the srcdir into the outdir.
pub fn compile_stdlib(srcdir: &Path, outdir: &Path) -> Result<()> {
    let (_, imports, mut sem_pkgs) = infer_stdlib_dir(srcdir)?;
    // Write each file as compiled module
    for (path, exports) in &imports {
        if let Some(code) = sem_pkgs.remove(path) {
            let module = Module {
                polytype: Some(exports.typ()),
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
        ast,
        parser::{self, parse_string},
        semantic::{convert::convert_polytype, sub::Substitution},
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
            package c
            import "b"

            z = b.y
        "#;
        let ast_packages: ASTPackageMap = semantic_map! {
            String::from("a") => parse_string("a.flux".to_string(), a).into(),
            String::from("b") => parse_string("b.flux".to_string(), b).into(),
            String::from("c") => parse_string("c.flux".to_string(), c).into(),
        };
        let mut infer_state = InferState::default();
        let (types, _) = infer_state.infer_pkg("c", &ast_packages, &PackageExports::new())?;

        let want = PackageExports::try_from(vec![(types.lookup_symbol("z").unwrap().clone(), {
            let mut p = parser::Parser::new("int");
            let typ_expr = p.parse_type_expression();
            if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
                panic!("TypeExpression parsing failed for int. {:?}", err);
            }
            convert_polytype(&typ_expr, &mut Substitution::default())?
        })])
        .unwrap();
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
                if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
                    panic!(
                        "TypeExpression parsing failed for int. {:?}", err
                    );
                }
                convert_polytype(&typ_expr, &mut Substitution::default())?
            },
            String::from("b") => {
                let mut p = parser::Parser::new("{x: int , y: int}");
                let typ_expr = p.parse_type_expression();
                if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
                    panic!(
                        "TypeExpression parsing failed for int. {:?}", err
                    );
                }
                convert_polytype(&typ_expr, &mut Substitution::default())?
            },
        };
        if want
            != infer_state
                .imports
                .into_iter()
                .map(|(k, v)| (k, v.typ()))
                .collect::<SemanticMap<_, _>>()
        {
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
