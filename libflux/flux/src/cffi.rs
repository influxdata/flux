use std::{
    any::Any,
    ffi::*,
    mem,
    os::raw::c_char,
    panic::{catch_unwind, resume_unwind},
};

use anyhow::anyhow;
use fluxcore::{
    ast,
    errors::SalvageResult,
    formatter, merge_packages,
    parser::Parser,
    semantic::{
        self,
        env::Environment,
        flatbuffers::{
            semantic_generated::fbsemantic as fb,
            types::{build_env, build_type},
        },
        import::Importer,
        import::Packages,
        nodes::{Package, Symbol},
        types::{BoundTvar, MonoType},
        Analyzer, AnalyzerConfig, Feature, PackageExports,
    },
};

use crate::semantic::flatbuffers::semantic_generated::fbsemantic::MonoTypeHolderArgs;

use super::{new_semantic_analyzer, prelude, Error, Result, IMPORTS};

/// An error handle designed to allow passing `Error` instances to library
/// consumers across language boundaries.
pub struct ErrorHandle {
    /// A heap-allocated `Error` message
    message: CString,

    /// The actual error
    err: Error,
}

impl From<Error> for Box<ErrorHandle> {
    fn from(err: Error) -> Self {
        Box::new(ErrorHandle {
            message: CString::new(format!("{}", err)).unwrap(),
            err,
        })
    }
}

impl From<Box<dyn Any + Send>> for Box<ErrorHandle> {
    fn from(err: Box<dyn Any + Send>) -> Self {
        // `panic!` will make `err` a `&str` or `String` so we try to extract those
        // If there is something else we resume the unwinding and let the caller deal with it
        let msg = err
            .downcast::<&str>()
            .map(|s| s.to_string())
            .or_else(|err| err.downcast::<String>().map(|s| *s))
            .unwrap_or_else(|err| resume_unwind(err));
        let err = Error::Other(anyhow!("{}", msg));
        Box::new(ErrorHandle {
            message: CString::new(msg).unwrap(),
            err,
        })
    }
}

static SEMANTIC_PACKAGES: &[u8] = include_bytes!(concat!(env!("OUT_DIR"), "/packages.data"));

/// Returns the flatbuffer of the semantic packages for the standard library
#[no_mangle]
pub extern "C" fn flux_semantic_packages(out: &mut flux_buffer_t) {
    out.data = SEMANTIC_PACKAGES.as_ptr();
    out.len = SEMANTIC_PACKAGES.len();
}

/// Frees a previously allocated error.
///
/// ## Memory layout
///
/// We use the memory layout pattern described in the [`std::boxed`] module,
/// wherein a pointer where ownership is being transferred is modeled as a [`Box`], and if it could be
/// null, then it's wrapped in an [`Option`].
///
/// [`std::boxed`]: https://doc.rust-lang.org/std/boxed/index.html#memory-layout
#[no_mangle]
pub extern "C" fn flux_free_error(_err: Option<Box<ErrorHandle>>) {}

/// Frees a pointer to characters.
///
/// # Safety
///
/// This function is unsafe because improper use may lead to
/// memory problems. For example, a double-free may occur if the
/// function is called twice on the same raw pointer.
#[no_mangle]
pub unsafe extern "C" fn flux_free_bytes(cstr: *mut c_char) {
    Box::from_raw(cstr);
}

/// A buffer of flux source.
#[repr(C)]
pub struct flux_buffer_t {
    /// A pointer to a byte array.
    pub data: *const u8,
    /// The length of the byte array.
    pub len: usize,
}

/// flux_parse parses a string containing Flux source code into an AST.
///
/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer passed
/// in as a parameter. For example, if that pointer is NULL, undefined behavior
/// could occur.
#[no_mangle]
pub unsafe extern "C" fn flux_parse(
    cfname: *const c_char,
    csrc: *const c_char,
) -> Box<ast::Package> {
    let fname = String::from_utf8(CStr::from_ptr(cfname).to_bytes().to_vec()).unwrap();
    let src = String::from_utf8(CStr::from_ptr(csrc).to_bytes().to_vec()).unwrap();
    let pkg = parse(fname, &src);
    Box::new(pkg)
}

/// Parse the contents of a string.
pub fn parse(fname: String, src: &str) -> ast::Package {
    let mut p = Parser::new(src);
    p.parse_file(fname).into()
}

/// Format the Flux AST.
#[no_mangle]
pub extern "C" fn flux_ast_format(
    ast_pkg: &ast::Package,
    out: &mut flux_buffer_t,
) -> Option<Box<ErrorHandle>> {
    let mut out_str = String::new();
    for file in &ast_pkg.files {
        let s = match formatter::convert_to_string(file) {
            Ok(v) => v,
            Err(e) => return Some(Error::from(e).into()),
        };
        out_str.push_str(&s);
    }

    let len = out_str.len();
    let cstr = match CString::new(out_str) {
        Ok(bytes) => bytes,
        Err(e) => return Some(Error::from(anyhow::Error::from(e)).into()),
    };
    out.data = cstr.into_raw() as *mut u8;
    out.len = len;
    None
}

/// flux_ast_get_error returns the first error in the given AST.
///
/// # Safety
///
/// This funtion is unsafe because it dereferences a raw pointer.
#[no_mangle]
pub unsafe extern "C" fn flux_ast_get_error(
    ast_pkg: *const ast::Package,
) -> Option<Box<ErrorHandle>> {
    catch_unwind(|| {
        let ast_pkg = ast::walk::Node::Package(&*ast_pkg);
        match ast::check::check(ast_pkg) {
            Err(e) => Some(Error::from(anyhow::Error::from(e)).into()),
            Ok(_) => None,
        }
    })
    .unwrap_or_else(|err| Some(err.into()))
}

/// Frees an AST package.
///
/// ## Memory layout
///
/// We use the memory layout pattern described in the [`std::boxed`] module,
/// wherein a pointer where ownership is being transferred is modeled as a [`Box`], and if it could be
/// null, then it's wrapped in an [`Option`].
///
/// [`std::boxed`]: https://doc.rust-lang.org/std/boxed/index.html#memory-layout
#[no_mangle]
pub extern "C" fn flux_free_ast_pkg(_: Option<Box<ast::Package>>) {}

/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer passed
/// in as a parameter. For example, if that pointer is NULL, undefined behavior
/// could occur.
#[no_mangle]
pub unsafe extern "C" fn flux_parse_json(
    cstr: *mut c_char,
    out_pkg: *mut Option<Box<ast::Package>>,
) -> Option<Box<ErrorHandle>> {
    catch_unwind(|| {
        let buf = CStr::from_ptr(cstr).to_bytes(); // Unsafe
        let res: Result<ast::Package, serde_json::error::Error> = serde_json::from_slice(buf);
        match res {
            Ok(pkg) => {
                *out_pkg = Some(Box::new(pkg));
                None
            }
            Err(err) => Some(Error::from(anyhow::Error::from(err)).into()),
        }
    })
    .unwrap_or_else(|err| Some(err.into()))
}

/// # Safety
///
/// This function is unsafe because it dereferences raw pointers passed
/// in as parameters. For example, if that pointer is NULL, undefined behavior
/// could occur.
#[no_mangle]
pub unsafe extern "C" fn flux_ast_marshal_json(
    ast_pkg: *const ast::Package,
    buf: *mut flux_buffer_t,
) -> Option<Box<ErrorHandle>> {
    catch_unwind(|| {
        let ast_pkg = &*ast_pkg;
        let data = match serde_json::to_vec(ast_pkg) {
            Ok(v) => v,
            Err(err) => {
                return Some(Error::from(anyhow::Error::from(err)).into());
            }
        };

        (*buf).len = data.len();
        (*buf).data = Box::into_raw(data.into_boxed_slice()) as *mut u8;
        None
    })
    .unwrap_or_else(|err| Some(err.into()))
}

/// Frees a semantic package.
#[no_mangle]
pub extern "C" fn flux_free_semantic_pkg(_: Option<Box<semantic::nodes::Package>>) {}

/// flux_semantic_marshal_fb populates the supplied buffer with a FlatBuffers serialization
/// of the given AST.
///
/// # Safety
///
/// This function is unsafe because it takes a dereferences a raw pointer passed
/// in as a parameter. For example, if that pointer is NULL, undefined behavior
/// could occur.
#[no_mangle]
pub unsafe extern "C" fn flux_semantic_marshal_fb(
    sem_pkg: *const semantic::nodes::Package,
    buf: *mut flux_buffer_t,
) -> Option<Box<ErrorHandle>> {
    catch_unwind(|| {
        let sem_pkg = &*sem_pkg;
        let (mut vec, offset) = match semantic::flatbuffers::serialize_pkg(sem_pkg) {
            Ok(vec_offset) => vec_offset,
            Err(err) => {
                return Some(Error::from(err).into());
            }
        };

        // Note, split_off() does a copy: https://github.com/influxdata/flux/issues/2194
        let data = vec.split_off(offset);
        (*buf).len = data.len();
        (*buf).data = Box::into_raw(data.into_boxed_slice()) as *mut u8;
        None
    })
    .unwrap_or_else(|err| Some(err.into()))
}

/// flux_error_str returns the error message associated with the given error.
///
/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer passed as a
/// parameter
#[no_mangle]
pub unsafe extern "C" fn flux_error_str(errh: &ErrorHandle) -> *const c_char {
    errh.message.as_ptr()
}

/// flux_error_print prints the error message associated with the given error to stdout.
///
/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer passed as a
/// parameter
#[no_mangle]
pub unsafe extern "C" fn flux_error_print(errh: &ErrorHandle) {
    match &errh.err {
        Error::Semantic(err) => err.print(),
        err => println!("{}", err),
    }
}

/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer passed as a
/// parameter
///
/// flux_merge_ast_pkg_files merges the files of a given input ast::Package into the file
/// vector of an output ast::Package.
#[no_mangle]
pub unsafe extern "C" fn flux_merge_ast_pkgs(
    out_pkg: *mut ast::Package,
    in_pkg: *mut ast::Package,
) -> Option<Box<ErrorHandle>> {
    catch_unwind(|| {
        // Do not change ownership here so that Go maintains ownership of packages
        let out_pkg = &mut *out_pkg;
        let in_pkg = &mut *in_pkg;

        match merge_packages(out_pkg, in_pkg) {
            Ok(_) => None,
            Err(e) => Some(Error::from(e).into()),
        }
    })
    .unwrap_or_else(|err| Some(err.into()))
}

/// flux_analyze is a C-compatible wrapper around the analyze() function below
///
/// Note that Box<T> is used to indicate we are receiving/returning a C pointer and also
/// transferring ownership.
///
/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer.
#[no_mangle]
#[allow(clippy::boxed_local)]
pub unsafe extern "C" fn flux_analyze(
    ast_pkg: Box<ast::Package>,
    options: *const c_char,
    out_sem_pkg: *mut Option<Box<semantic::nodes::Package>>,
) -> Option<Box<ErrorHandle>> {
    catch_unwind(|| {
        let options = match Options::from_c_str(options) {
            Ok(x) => x,
            Err(err) => return Some(err.into()),
        };
        match analyze(&ast_pkg, options) {
            Ok(sem_pkg) => {
                *out_sem_pkg = Some(Box::new(sem_pkg));
                None
            }
            Err(salvage) => {
                *out_sem_pkg = salvage.value.map(Box::new);
                Some(salvage.error.into())
            }
        }
    })
    .unwrap_or_else(|err| Some(err.into()))
}

/// flux_find_var_type() is a C-compatible wrapper around the find_var_type() function below.
/// Note that Box<T> is used to indicate we are receiving/returning a C pointer and also
/// transferring ownership.
///
/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer.
#[no_mangle]
#[allow(clippy::boxed_local)]
pub unsafe extern "C" fn flux_find_var_type(
    sem_pkg: *const semantic::nodes::Package,
    var_name: *const c_char,
    out_type: *mut flux_buffer_t,
) -> Option<Box<ErrorHandle>> {
    catch_unwind(|| {
        let buf = CStr::from_ptr(var_name).to_bytes(); // Unsafe
        let name = std::str::from_utf8(buf).unwrap();
        find_var_type(&*sem_pkg, &name).map_or_else(
            |e| Some(Box::from(e)),
            |t| {
                let mut builder = flatbuffers::FlatBufferBuilder::new();
                let (fb_mono_type, typ_type) = build_type(&mut builder, &t);
                let fb_mono_type_holder = fb::MonoTypeHolder::create(
                    &mut builder,
                    &MonoTypeHolderArgs {
                        typ_type,
                        typ: Some(fb_mono_type),
                    },
                );
                builder.finish(fb_mono_type_holder, None);
                let (mut vec, offset) = builder.collapse();
                // Note, split_off() does a copy: https://github.com/influxdata/flux/issues/2194
                let data = vec.split_off(offset);
                let out_type = &mut *out_type; // Unsafe
                out_type.len = data.len();
                out_type.data = Box::into_raw(data.into_boxed_slice()) as *mut u8;
                None
            },
        )
    })
    .unwrap_or_else(|err| Some(err.into()))
}

fn new_stateful_analyzer(options: Options) -> Result<StatefulAnalyzer> {
    let env = match prelude() {
        Some(prelude) => prelude,
        None => return Err(anyhow!("missing prelude").into()),
    };
    let imports = match &*IMPORTS {
        Some(imports) => imports,
        None => return Err(anyhow!("missing stdlib imports").into()),
    };
    Ok(StatefulAnalyzer {
        env,
        imports,
        options,
    })
}

/// StatefulAnalyzer updates its environment with the contents of any previously analyzed package.
/// This enables uses cases where analysis is performed iteratively, for example in a REPL.
pub struct StatefulAnalyzer {
    env: PackageExports,
    imports: &'static Packages,
    options: Options,
}

impl StatefulAnalyzer {
    fn analyze(&mut self, ast_pkg: &ast::Package) -> Result<fluxcore::semantic::nodes::Package> {
        let Options { features } = self.options.clone();
        let mut analyzer = Analyzer::new(
            Environment::from(&self.env),
            self.imports,
            AnalyzerConfig { features },
        );
        let (mut env, sem_pkg) = match analyzer.analyze_ast(ast_pkg) {
            Ok(r) => r,
            Err(e) => {
                // In the face of an error we need to get the imports
                // back from the analyzer.
                let (_env, imports) = analyzer.drop();
                self.imports = imports;
                return Err(e.error.into());
            }
        };
        // Restore the imports.
        // We restore the env below.
        let (_, imports) = analyzer.drop();
        self.imports = imports;

        // Re-export any imported names into the env.
        // Normally we do not do this but we need to remember
        // any previous import statements since
        // each line of source is analyzed independently.
        for file in &sem_pkg.files {
            for dec in &file.imports {
                let path = &dec.path.value;

                // A failure should have already happened if any of these
                // imports would have failed.
                if let Some(typ) = self.imports.import(path) {
                    env.add(dec.import_symbol.clone(), typ);
                }
            }
        }
        self.env.copy_bindings_from(&env);
        Ok(sem_pkg)
    }
}

/// Create a new semantic analyzer.
///
/// # Safety
///
/// Ths function is unsafe because it dereferences a raw pointer.
#[no_mangle]
pub unsafe extern "C" fn flux_new_stateful_analyzer(
    options: *const c_char,
) -> Box<Result<StatefulAnalyzer>> {
    let options = match Options::from_c_str(options) {
        Ok(x) => x,
        Err(err) => return Box::new(Err(err)),
    };
    Box::new(new_stateful_analyzer(options))
}

/// Free a previously allocated semantic analyzer
#[no_mangle]
pub extern "C" fn flux_free_stateful_analyzer(_: Option<Box<Result<StatefulAnalyzer>>>) {}

/// # Safety
///
/// Ths function is unsafe because it dereferences a raw pointer.
#[no_mangle]
#[allow(clippy::boxed_local)]
pub unsafe extern "C" fn flux_analyze_with(
    analyzer: *mut Result<StatefulAnalyzer>,
    csrc: *const c_char,
    ast_pkg: Box<ast::Package>,
    out_sem_pkg: *mut Option<Box<semantic::nodes::Package>>,
) -> Option<Box<ErrorHandle>> {
    catch_unwind(|| {
        let ast_pkg = &ast_pkg;
        let analyzer = &mut *analyzer;
        let analyzer = match analyzer {
            Ok(a) => a,
            Err(_) => {
                match mem::replace(
                    analyzer,
                    Err(Error::from(anyhow!("The error has already been return!"))),
                ) {
                    Err(err) => {
                        return Some(err.into());
                    }
                    Ok(_) => unreachable!(),
                }
            }
        };

        let src = if csrc.is_null() {
            None
        } else {
            Some(std::str::from_utf8(CStr::from_ptr(csrc).to_bytes()).unwrap())
        };

        let sem_pkg = Box::new(match analyzer.analyze(ast_pkg) {
            Ok(sem_pkg) => sem_pkg,
            Err(mut err) => {
                if let Some(src) = src {
                    if let Error::Semantic(err) = &mut err {
                        err.source = Some(src.into());
                    }
                }
                return Some(err.into());
            }
        });

        *out_sem_pkg = Some(sem_pkg);
        None
    })
    .unwrap_or_else(|err| Some(err.into()))
}

/// Compilation options. Deserialized from json when called via the C API
#[derive(Clone, Default, Debug)]
#[cfg_attr(feature = "serde", derive(serde::Deserialize))]
pub struct Options {
    /// Features used in the flux compiler
    #[serde(default)]
    pub features: Vec<Feature>,
}

impl Options {
    unsafe fn from_c_str(options: *const c_char) -> Result<Self> {
        let options = CStr::from_ptr(options).to_bytes();
        if options.is_empty() {
            return Ok(Self::default());
        }

        #[cfg(not(feature = "serde"))]
        {
            return Err(Error::Other(anyhow!(
                "`serde` feature is not enabled, unable to parse compilation options."
            )));
        }

        #[cfg(feature = "serde")]
        match serde_json::from_slice(options) {
            Ok(x) => Ok(x),
            Err(err) => Err(Error::InvalidOptions(err.to_string())),
        }
    }
}

/// analyze consumes the given AST package and returns a semantic package
/// that has been type-inferred.  This function is aware of the standard library
/// and prelude.
pub fn analyze(ast_pkg: &ast::Package, options: Options) -> SalvageResult<Package, Error> {
    let Options { features } = options;
    let mut analyzer = new_semantic_analyzer(AnalyzerConfig { features })?;
    let (_, sem_pkg) = analyzer
        .analyze_ast(ast_pkg)
        .map_err(|salvage| salvage.err_into().map(|(_, sem_pkg)| sem_pkg))?;
    Ok(sem_pkg)
}

/// Given a Flux source and a variable name, find out the type of that variable in the Flux source code.
/// A type variable will be automatically generated and injected into the type environment that
/// will be used in semantic analysis. The Flux source code itself should not contain any definition
/// for that variable.
/// This version of find_var_type is aware of the prelude and builtins.
fn find_var_type(pkg: &Package, var_name: &str) -> Result<MonoType> {
    Ok(semantic::find_var_type(pkg, var_name).unwrap_or_else(|| MonoType::BoundVar(BoundTvar(0))))
}

/// # Safety
///
/// This function is unsafe because it dereferences a raw pointer.
#[no_mangle]
pub unsafe extern "C" fn flux_get_env_stdlib(buf: *mut flux_buffer_t) {
    let imports = IMPORTS.as_ref().unwrap();
    let env = PackageExports::try_from(
        imports
            .iter()
            .map(|(k, v)| (Symbol::from(k.as_str()), v.typ()))
            .collect::<Vec<_>>(),
    )
    .unwrap();
    let mut builder = flatbuffers::FlatBufferBuilder::new();
    let fb_type_env = build_env(&mut builder, env);

    builder.finish(fb_type_env, None);
    let (mut vec, offset) = builder.collapse();

    // Note, split_off() does a copy: https://github.com/influxdata/flux/issues/2194
    let data = vec.split_off(offset);
    let buf = &mut *buf; // Unsafe
    buf.len = data.len();
    buf.data = Box::into_raw(data.into_boxed_slice()) as *mut u8;
}

#[cfg(test)]
mod tests {
    use fluxcore::{
        ast,
        semantic::{
            convert::convert_polytype,
            fresh::Fresher,
            types::{BoundTvar, Label, MonoType, Property, Ptr, Record, Tvar, TvarMap},
            walk,
        },
    };

    use super::*;

    use crate::parser;

    pub struct MonoTypeNormalizer {
        tv_map: TvarMap,
        f: Fresher,
    }

    impl MonoTypeNormalizer {
        pub fn new() -> Self {
            Self {
                tv_map: TvarMap::new(),
                f: Fresher::default(),
            }
        }

        pub fn normalize(&mut self, t: &mut MonoType) {
            match t {
                MonoType::BoundVar(tv) => {
                    // This is to avoid using self directly inside a closure,
                    // otherwise it will be captured by that closure and the compiler
                    // will complain that closure requires unique access to `self`
                    let f = &mut self.f;
                    let v = self.tv_map.entry(Tvar(tv.0)).or_insert_with(|| f.fresh());
                    *t = MonoType::BoundVar(BoundTvar(v.0));
                }
                MonoType::Var(tv) => {
                    // This is to avoid using self directly inside a closure,
                    // otherwise it will be captured by that closure and the compiler
                    // will complain that closure requires unique access to `self`
                    let f = &mut self.f;
                    let v = self.tv_map.entry(*tv).or_insert_with(|| f.fresh());
                    *t = MonoType::BoundVar(BoundTvar(v.0));
                }
                MonoType::Collection(app) => {
                    self.normalize(&mut Ptr::make_mut(app).arg);
                }
                MonoType::Record(r) => {
                    if let Record::Extension { head, tail } = Ptr::make_mut(r) {
                        self.normalize(&mut head.v);
                        self.normalize(tail);
                    }
                }
                MonoType::Fun(f) => {
                    let f = Ptr::make_mut(f);
                    for (_, mut v) in f.req.iter_mut() {
                        self.normalize(&mut v);
                    }
                    for (_, v) in f.opt.iter_mut() {
                        self.normalize(&mut v.typ);
                        if let Some(default) = &mut v.default {
                            self.normalize(default);
                        }
                    }
                    if let Some(p) = &mut f.pipe {
                        self.normalize(&mut p.v);
                    }
                    self.normalize(&mut f.retn);
                }
                _ => {}
            }
        }
    }

    #[test]
    fn monotype_normalizer() {
        let mut ty = MonoType::from(Record::new(
            [
                Property {
                    k: Label::from("a").into(),
                    v: MonoType::BoundVar(BoundTvar(4949)),
                },
                Property {
                    k: Label::from("b").into(),
                    v: MonoType::BoundVar(BoundTvar(4949)),
                },
                Property {
                    k: Label::from("e").into(),
                    v: MonoType::BoundVar(BoundTvar(4957)),
                },
                Property {
                    k: Label::from("f").into(),
                    v: MonoType::BoundVar(BoundTvar(4957)),
                },
                Property {
                    k: Label::from("g").into(),
                    v: MonoType::BoundVar(BoundTvar(4957)),
                },
            ],
            Some(MonoType::BoundVar(BoundTvar(4972))),
        ));
        assert_eq!(
            format!("{}", ty),
            r#"{
    t4972 with
    a: t4949,
    b: t4949,
    e: t4957,
    f: t4957,
    g: t4957,
}"#,
        );
        let mut v = MonoTypeNormalizer::new();
        v.normalize(&mut ty);
        assert_eq!(
            format!("{}", ty),
            r#"{
    C with
    a: A,
    b: A,
    e: B,
    f: B,
    g: B,
}"#
        );
    }

    fn find_var_type_from_source(source: &str, var_name: &str) -> Result<MonoType> {
        let mut analyzer = new_semantic_analyzer(AnalyzerConfig::default())?;
        let pkg = match analyzer.analyze_source("".into(), "".into(), source) {
            Ok((_, pkg)) => pkg,
            Err(err) => match err.value {
                Some((_, pkg)) => pkg,
                None => return Err(err.error.into()),
            },
        };

        find_var_type(&pkg, var_name)
    }

    #[test]
    fn find_var_ref() {
        let source = r#"
vint = v.int + 2
f = (v) => v.shadow
g = () => v.sweet
x = g()
vstr = v.str + "hello"
"#;
        let mut t =
            find_var_type_from_source(source, "v").expect("Should be able to get a MonoType.");
        let mut v = MonoTypeNormalizer::new();
        v.normalize(&mut t);
        assert_eq!(
            format!("{}", t),
            r#"{B with int: int, sweet: A, str: string}"#
        );

        expect_test::expect![[r#"
            {
              "Record": {
                "type": "Extension",
                "head": {
                  "k": {
                    "Concrete": "int"
                  },
                  "v": "Int"
                },
                "tail": {
                  "Record": {
                    "type": "Extension",
                    "head": {
                      "k": {
                        "Concrete": "sweet"
                      },
                      "v": {
                        "Var": 0
                      }
                    },
                    "tail": {
                      "Record": {
                        "type": "Extension",
                        "head": {
                          "k": {
                            "Concrete": "str"
                          },
                          "v": "String"
                        },
                        "tail": {
                          "Var": 1
                        }
                      }
                    }
                  }
                }
              }
            }"#]]
        .assert_eq(&serde_json::to_string_pretty(&t).unwrap());
    }

    #[test]
    fn find_var_ref_non_row_type() {
        let source = r#"
vint = v + 2
"#;
        let t = find_var_type_from_source(source, "v").expect("Should be able to get a MonoType.");
        assert_eq!(t, MonoType::INT);

        assert_eq!(serde_json::to_string_pretty(&t).unwrap(), "\"Int\"");
    }

    #[test]
    fn find_var_ref_obj_with() {
        let source = r#"
vint = v.int + 2
o = {v with x: 256}
p = o.ethan
"#;
        let mut t =
            find_var_type_from_source(source, "v").expect("Should be able to get a MonoType.");
        let mut v = MonoTypeNormalizer::new();
        v.normalize(&mut t);
        assert_eq!(format!("{}", t), r#"{B with int: int, ethan: A}"#);

        expect_test::expect![[r#"
            {
              "Record": {
                "type": "Extension",
                "head": {
                  "k": {
                    "Concrete": "int"
                  },
                  "v": "Int"
                },
                "tail": {
                  "Record": {
                    "type": "Extension",
                    "head": {
                      "k": {
                        "Concrete": "ethan"
                      },
                      "v": {
                        "Var": 0
                      }
                    },
                    "tail": {
                      "Var": 1
                    }
                  }
                }
              }
            }"#]]
        .assert_eq(&serde_json::to_string_pretty(&t).unwrap());
    }

    #[test]
    fn find_var_ref_query() {
        // Test the find_var_type() function with some calls to stdlib functions.
        let source = r#"
from(bucket: v.bucket)
|> range(start: v.timeRangeStart, stop: v.timeRangeStop)
|> filter(fn: (r) => r._measurement == v.measurement or r._measurement == "cpu")
|> filter(fn: (r) => r.host == "host.local")
|> aggregateWindow(every: 30s, fn: count)
"#;
        let mut ty =
            find_var_type_from_source(source, "v").expect("should be able to find var type");
        let mut v = MonoTypeNormalizer::new();
        v.normalize(&mut ty);
        assert_eq!(
            format!("{}", ty),
            "{D with measurement: A, timeRangeStart: B, timeRangeStop: C, bucket: string}"
        );
    }

    #[test]
    fn test_ast_get_error() {
        let ast = crate::parser::parse_string("test".to_string(), "x = 3 + / 10 - \"");
        let ast = Box::into_raw(Box::new(ast.into()));
        let errh = unsafe { flux_ast_get_error(ast) };

        expect_test::expect![[r#"
            error test@1:9-1:10: invalid expression: invalid token for primary expression: DIV

            error test@1:16-1:17: got unexpected token in string expression test@1:17-1:17: EOF"#]]
        .assert_eq(&errh.unwrap().message.into_string().unwrap());
    }

    #[test]
    fn deserialize_and_infer() {
        let mut analyzer = new_semantic_analyzer(AnalyzerConfig::default()).unwrap();

        let src = r#"
            x = from(bucket: "b")
                |> filter(fn: (r) => r.region == "west")
                |> map(fn: (r) => ({r with _value: r._value + r._value}))
        "#;

        let (got, _) = analyzer
            .analyze_source("".to_string(), "main.flux".to_string(), src)
            .unwrap();

        // TODO(algow): re-introduce equality constraints for binary comparison operators
        // https://github.com/influxdata/flux/issues/2466
        let code = "stream[{ C with
                _value: A
                    , _value: A
                    , _time: time
                    , _measurement: string
                    , _field: string
                    , region: B
                    }] where A: Addable, B: Equatable ";
        let mut p = parser::Parser::new(code);

        let typ_expr = p.parse_type_expression();
        if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
            panic!("TypeExpression parsing failed. {:?}", err);
        }
        let want = convert_polytype(&typ_expr, &Default::default()).unwrap();

        assert_eq!(want, got.lookup("x").expect("'x' not found").clone());
    }

    #[test]
    fn infer_union() {
        let mut analyzer = new_semantic_analyzer(AnalyzerConfig::default()).unwrap();

        let src = r#"
            a = from(bucket: "b")
                |> filter(fn: (r) => r.A_ == "A")
            b = from(bucket: "b")
                |> filter(fn: (r) => r.B_ == "B")
            c = union(tables: [a, b])
        "#;

        let (got, _) = analyzer
            .analyze_source("".to_string(), "main.flux".to_string(), src)
            .unwrap();

        // TODO(algow): re-introduce equality constraints for binary comparison operators
        // https://github.com/influxdata/flux/issues/2466
        let code = "stream[{ D with
                _value: A
                    , A_: B
                    , _time: time
                    , _measurement: string
                    , _field: string
                    }] where B: Equatable ";
        let mut p = parser::Parser::new(code);

        let typ_expr = p.parse_type_expression();
        if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
            panic!("TypeExpression parsing failed for {:?}", err);
        }
        let want_a = convert_polytype(&typ_expr, &Default::default()).unwrap();

        let code = "stream[{ D with
                _value: A
                    , B_: B
                    , _time: time
                    , _measurement: string
                    , _field: string
                    }] where  B: Equatable";

        let mut p = parser::Parser::new(code);

        let typ_expr = p.parse_type_expression();
        if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
            panic!("TypeExpression parsing failed for {:?}", err);
        }
        let want_b = convert_polytype(&typ_expr, &Default::default()).unwrap();

        let code = "stream[{ D with
                _value: A
                    , A_: B
                    , B_: C
                    , _time: time
                    , _measurement: string
                    , _field: string
                    }] where B: Equatable, C: Equatable ";
        let mut p = parser::Parser::new(code);

        let typ_expr = p.parse_type_expression();
        if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
            panic!("TypeExpression parsing failed for {:?}", err);
        }
        let want_c = convert_polytype(&typ_expr, &Default::default()).unwrap();

        assert_eq!(want_a, got.lookup("a").expect("'a' not found").clone());
        assert_eq!(want_b, got.lookup("b").expect("'b' not found").clone());
        assert_eq!(want_c, got.lookup("c").expect("'c' not found").clone());
    }

    #[test]
    fn analyze_error() {
        let ast: ast::Package = fluxcore::parser::parse_string("".to_string(), "x = ()").into();
        match analyze(&ast, Options::default()) {
            Ok(_) => panic!("expected an error, got none"),
            Err(e) => {
                expect_test::expect![[r#"
                    error @1:5-1:7: expected ARROW, got EOF

                    error @1:7-1:7: invalid expression: invalid token for primary expression: EOF"#]].assert_eq(&e.to_string());
            }
        }
    }

    #[test]
    fn prelude_symbols_retain_their_package() {
        let mut analyzer = new_semantic_analyzer(AnalyzerConfig::default()).unwrap();

        let src = r#"
            derivative
        "#;

        let (_, pkg) = analyzer
            .analyze_source("".to_string(), "main.flux".to_string(), src)
            .unwrap();

        let mut identifier = None;
        walk::walk(
            &mut |node| {
                if let walk::Node::IdentifierExpr(id) = node {
                    identifier = Some(id);
                }
            },
            walk::Node::Package(&pkg),
        );

        dbg!(&pkg);

        assert_eq!(identifier.unwrap().name.package(), Some("universe"));
    }
}
