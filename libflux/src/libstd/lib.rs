use flatbuffers;
use flux::ast;
use flux::ctypes::*;
use flux::flux_buffer_t;
use flux::semantic::env::Environment;
use flux::semantic::flatbuffers::semantic_generated::fbsemantic as fb;
use flux::semantic::fresh::Fresher;
use flux::semantic::nodes::{infer_pkg_types, inject_pkg_types};
use flux::semantic::flatbuffers::types::build_env; 

pub fn prelude() -> Option<Environment> {
    let buf = include_bytes!(concat!(env!("OUT_DIR"), "/prelude.data"));
    flatbuffers::get_root::<fb::TypeEnvironment>(buf).into()
}

pub fn imports() -> Option<Environment> {
    let buf = include_bytes!(concat!(env!("OUT_DIR"), "/stdlib.data")); 
    flatbuffers::get_root::<fb::TypeEnvironment>(buf).into()
}

pub fn fresher() -> Fresher {
    let buf = include_bytes!(concat!(env!("OUT_DIR"), "/fresher.data"));
    flatbuffers::get_root::<fb::Fresher>(buf).into()
}

/// # Safety
///
/// Ths function is unsafe because it dereferences a raw pointer.
#[no_mangle]
pub unsafe extern "C" fn flux_analyze(
    ast_pkg: *mut flux_ast_pkg_t,
    out_sem_pkg: *mut *const flux_semantic_pkg_t,
) -> *mut flux_error_t {
    let ast_pkg = *Box::from_raw(ast_pkg as *mut ast::Package);
    match analyze(ast_pkg) {
        Ok(sem_pkg) => {
            let sem_pkg = Box::into_raw(Box::new(sem_pkg)) as *const flux_semantic_pkg_t;
            *out_sem_pkg = sem_pkg;
            std::ptr::null_mut()
        }
        Err(err) => {
            let errh = flux::ErrorHandle { err: Box::new(err) };
            Box::into_raw(Box::new(errh)) as *mut flux_error_t
        }
    }
}

/// analyze consumes the given AST package and returns a semantic package
/// that has been type-inferred.  This function is aware of the standard library
/// and prelude.
pub fn analyze(ast_pkg: ast::Package) -> Result<flux::semantic::nodes::Package, flux::Error> {
    let mut f = fresher();
    let mut sem_pkg = flux::semantic::convert::convert_with(ast_pkg, &mut f)?;

    let prelude = match prelude() {
        Some(prelude) => Environment::new(prelude),
        None => return Err(flux::Error::from("missing prelude")),
    };
    let imports = match imports() {
        Some(imports) => imports,
        None => return Err(flux::Error::from("missing stdlib imports")),
    };
    let (_, sub) = infer_pkg_types(&mut sem_pkg, prelude, &mut f, &imports, &None)?;
    sem_pkg = inject_pkg_types(sem_pkg, &sub);
    Ok(sem_pkg)
}

#[no_mangle]
pub unsafe extern "C" fn flux_get_env_stdlib(buf: *mut flux_buffer_t) {
    let env = imports().unwrap(); 
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
    use flux::semantic;
    use flux::semantic::convert::convert_file;
    use flux::semantic::env::Environment;
    use flux::semantic::nodes::infer_file;

    #[test]
    fn deserialize_and_infer() {
        let prelude = Environment::new(super::prelude().unwrap());
        let imports = super::imports().unwrap();

        let src = r#"
            x = from(bucket: "b")
                |> filter(fn: (r) => r.region == "west")
                |> map(fn: (r) => ({r with _value: r._value + r._value}))
        "#;

        let ast = flux::parser::parse_string("main.flux", src);
        let mut f = super::fresher();

        let mut file = convert_file(ast, &mut f).unwrap();
        let (got, _) = infer_file(&mut file, prelude, &mut f, &imports, &None).unwrap();

        let want = semantic::parser::parse(
            r#"forall [t0, t1] where t0: Addable [{
                _value: t0
                    | _value: t0
                    | _time: time
                    | _measurement: string
                    | _field: string
                    | region: string
                    | t1
                    }]
            "#,
        )
        .unwrap();

        assert_eq!(want, got.lookup("x").expect("'x' not found").clone());
    }
}
