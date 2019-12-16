use flux::semantic::env::Environment;
use flux::semantic::flatbuffers::semantic_generated::fbsemantic as fb;
use flux::semantic::fresh::Fresher;

use flatbuffers;

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
