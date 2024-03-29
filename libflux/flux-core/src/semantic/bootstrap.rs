//! Bootstrap provides an API for compiling the Flux standard library.
//!
//! This package does not assume a location of the source code but does assume which packages are
//! part of the prelude.

use std::{env::consts, io, path::Path, sync::Arc};

use anyhow::{bail, Result};
use walkdir::WalkDir;

use crate::{
    ast,
    db::{DatabaseBuilder, Flux},
    semantic::{
        fs::{FileSystemImporter, StdFS},
        import::{Importer, Packages},
        nodes::{self, Package, Symbol},
        sub::{Substitutable, Substituter},
        types::{
            BoundTvar, BoundTvarKinds, MonoType, PolyType, PolyTypeHashMap, Record, RecordLabel,
            SemanticMap, Tvar,
        },
        AnalyzerConfig, PackageExports,
    },
};

// List of packages to include into the Flux prelude
pub(crate) const PRELUDE: [&str; 4] = [
    "internal/boolean",
    "internal/location",
    "universe",
    "influxdata/influxdb",
];

/// A mapping of package import paths to the corresponding AST package.
pub type ASTPackageMap = SemanticMap<String, ast::Package>;
/// A mapping of package import paths to the corresponding semantic graph package.
pub type SemanticPackageMap = SemanticMap<String, Arc<Package>>;

/// Infers the Flux standard library given the path to the source code.
/// The prelude and the imports are returned.
#[allow(clippy::type_complexity)]
pub fn infer_stdlib_dir(
    path: impl AsRef<Path>,
    config: AnalyzerConfig,
) -> Result<(PackageExports, Packages, SemanticPackageMap)> {
    infer_stdlib_dir_(path.as_ref(), config)
}

#[allow(clippy::type_complexity)]
fn infer_stdlib_dir_(
    path: &Path,
    config: AnalyzerConfig,
) -> Result<(PackageExports, Packages, SemanticPackageMap)> {
    let package_list = parse_dir(path)?;

    let mut db = DatabaseBuilder::default()
        .filesystem_roots(vec![path.into()])
        .build();

    db.set_analyzer_config(config);

    let mut imports = Packages::default();
    let mut sem_pkg_map = SemanticPackageMap::default();
    for name in &package_list {
        let (exports, pkg) = db.semantic_package(name.clone())?;
        imports.insert(name.clone(), exports.clone());
        sem_pkg_map.insert(name.clone(), pkg.clone());
    }

    let prelude = db.prelude()?;
    Ok((PackageExports::clone(&prelude), imports, sem_pkg_map))
}

/// Recursively parse all flux files within a directory.
pub fn parse_dir(dir: &Path) -> io::Result<Vec<String>> {
    let mut package_names = Vec::new();
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

                let file_name = normalized_path
                    .rsplitn(2, "/stdlib/")
                    .collect::<Vec<&str>>()[0]
                    .to_owned();
                let path = file_name.rsplitn(2, '/').collect::<Vec<&str>>()[1].to_string();
                package_names.push(path);
            }
        }
    }

    Ok(package_names)
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
        if let Ok(pkg_type) = importer.import(pkg) {
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
struct CollectBoundVars(Vec<BoundTvar>);

impl Substituter for CollectBoundVars {
    fn try_apply(&mut self, _var: Tvar) -> Option<MonoType> {
        None
    }

    fn try_apply_bound(&mut self, var: BoundTvar) -> Option<MonoType> {
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
    free_vars: &[BoundTvar],
    cons: &BoundTvarKinds,
) -> Result<()> {
    for field in r.fields() {
        let new_vars = {
            let mut new_vars = CollectBoundVars(Vec::new());
            field.v.visit(&mut new_vars);
            new_vars.0
        };

        let mut new_cons = BoundTvarKinds::new();
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
    pub code: Option<Arc<nodes::Package>>,
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{
        ast, parser,
        semantic::{self, convert::convert_polytype},
    };

    use crate::db::{Database, Flux, FluxBase};

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

        let mut db = Database::default();
        db.set_use_prelude(false);

        for (k, v) in [("a/a.flux", a), ("b/b.flux", b), ("c/c.flux", c)] {
            db.set_source(k.into(), v.into());
        }
        let (types, _) = db.semantic_package("c".into())?;

        let want = PackageExports::try_from(vec![(types.lookup_symbol("z").unwrap().clone(), {
            let mut p = parser::Parser::new("int");
            let typ_expr = p.parse_type_expression();
            if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
                panic!("TypeExpression parsing failed for int. {:?}", err);
            }
            convert_polytype(&typ_expr, &Default::default())?
        })])
        .unwrap();
        if want != *types {
            bail!(
                "unexpected inference result:\n\nwant: {:?}\n\ngot: {:?}",
                want,
                types,
            );
        }

        let a = {
            let mut p = parser::Parser::new("{f: (x: A) => A}");
            let typ_expr = p.parse_type_expression();
            if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
                panic!("TypeExpression parsing failed for int. {:?}", err);
            }
            convert_polytype(&typ_expr, &Default::default())?
        };
        assert_eq!(db.import("a"), Ok(a));

        let b = {
            let mut p = parser::Parser::new("{x: int , y: int}");
            let typ_expr = p.parse_type_expression();
            if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
                panic!("TypeExpression parsing failed for int. {:?}", err);
            }
            convert_polytype(&typ_expr, &Default::default())?
        };
        assert_eq!(db.import("b"), Ok(b));

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

        let mut db = Database::default();

        db.set_use_prelude(false);

        for (k, v) in [("a/a.flux", a), ("b/b.flux", b)] {
            db.set_source(k.into(), v.into());
        }

        let got_err = db
            .semantic_package("b".into())
            .expect_err("expected cyclic dependency error");

        assert_eq!(
            r#"error @0:0-0:0: package "b" depends on itself: b -> a -> b"#.to_string(),
            got_err.to_string(),
        );
    }

    #[test]
    fn bootstrap() {
        infer_stdlib_dir("../../stdlib", AnalyzerConfig::default())
            .unwrap_or_else(|err| panic!("{}", err));
    }

    #[test]
    fn cross_module_error() {
        let a = r#"
            x = 1 + ""
        "#;
        let b = r#"
            import "a"

            y = a.x
        "#;

        let mut db = Database::default();

        db.set_use_prelude(false);
        db.set_analyzer_config(AnalyzerConfig {
            features: vec![semantic::Feature::PrettyError],
        });

        for (k, v) in [("a/a.flux", a), ("b/b.flux", b)] {
            db.set_source(k.into(), v.into());
        }

        let got_err = db
            .semantic_package("b".into())
            .expect_err("expected error error");
        let mut errors = db.package_errors();
        errors.push(got_err.error);

        expect_test::expect![[r#"
            error: expected int but found string
              ┌─ a/a.flux:1:1
              │
            1 │ x = 1 + ""
              │ ^



            error: invalid import path a
              ┌─ b/b.flux:3:12
              │
            3 │             y = a.x
              │            ^^^^^^^^^

        "#]]
        .assert_eq(&errors.to_string());
    }
}
