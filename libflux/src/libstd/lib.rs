use flatbuffers;
use flux::ctypes::*;
use flux::flux_buffer_t;
use flux::semantic::analyze::analyze_file;
use flux::semantic::env::Environment;
use flux::semantic::flatbuffers::semantic_generated::fbsemantic as fb;
use flux::semantic::fresh::Fresher;
use flux::semantic::nodes::{infer_pkg_types, inject_pkg_types};
use std::ffi::*;
use std::os::raw::c_char;

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

#[no_mangle]
pub unsafe extern "C" fn flux_semantic_analyze(
    src_ptr: *const c_char,
    flux_buf: *mut flux_buffer_t,
) -> *mut flux_error_t {
    let buf = CStr::from_ptr(src_ptr).to_bytes(); // Unsafe
    let s = String::from_utf8(buf.to_vec()).unwrap();
    match analyze(s.as_str()) {
        Ok(vec) => {
            let flux_buf = &mut *flux_buf;
            flux_buf.data = vec.as_ptr();
            flux_buf.len = vec.len();
            std::mem::forget(vec);
            std::ptr::null_mut()
        }
        Err(err) => {
            let errh = flux::ErrorHandle { err: Box::new(err) };
            return Box::into_raw(Box::new(errh)) as *mut flux_error_t;
        }
    }
}

fn analyze(src: &str) -> Result<Vec<u8>, flux::Error> {
    let mut f = fresher();

    let ast_file = flux::parser::parse_string("", src);
    let sem_file = analyze_file(ast_file, &mut f).unwrap();
    let mut sem_pkg = flux::semantic::nodes::Package {
        loc: flux::ast::SourceLocation {
            ..flux::ast::SourceLocation::default()
        },
        package: String::from(flux::DEFAULT_PACKAGE_NAME),
        files: vec![sem_file],
    };

    let prelude = Environment::new(prelude().unwrap());
    let imports = imports().unwrap();
    let (_, sub) = infer_pkg_types(&mut sem_pkg, prelude, &mut f, &imports, &None)?;
    sem_pkg = inject_pkg_types(sem_pkg, &sub);

    let (mut vec, offset) = flux::semantic::flatbuffers::serialize(&mut sem_pkg)?;
    Ok(vec.split_off(offset))
}

#[cfg(test)]
mod tests {
    use flux::semantic;
    use flux::semantic::analyze::analyze_file;
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

        let mut file = analyze_file(ast, &mut f).unwrap();
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
