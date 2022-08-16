//! This module defines methods for serializing and deserializing MonoTypes
//! and PolyTypes using the flatbuffer encoding.

use crate::{
    map::HashMap,
    semantic::{flatbuffers::semantic_generated::fbsemantic as fb, PackageExports},
};

use crate::semantic::{
    bootstrap::Module,
    flatbuffers::serialize_pkg_into,
    import::Packages,
    nodes::Symbol,
    types::{
        self, BoundTvar, BoundTvarKinds, BuiltinType, Collection, CollectionType, Dictionary,
        Function, Kind, MonoType, MonoTypeMap, PolyType, PolyTypeMap, Property, Record,
        RecordLabel, Tvar, TvarKinds,
    },
};

#[derive(Default)]
struct DeserializeFlatBuffer {
    symbols: HashMap<*const u8, Symbol>,
}

impl DeserializeFlatBuffer {
    fn deserialize_packages(&mut self, fb_packages: fb::Packages<'_>) -> Option<Packages> {
        let fb_packages = fb_packages.packages()?;
        let mut packages = Packages::new();
        for package in fb_packages.iter() {
            let (id, package) = self.deserialize_package_entry(package)?;
            packages.insert(id, package);
        }
        Some(packages)
    }

    fn deserialize_package_entry(
        &mut self,
        a: fb::PackageExports<'_>,
    ) -> Option<(String, PackageExports)> {
        let id: String = a.id()?.into();
        let exports: Option<PackageExports> = self.deserialize_package_exports(a.package()?);
        Some((id, exports?))
    }

    fn deserialize_package_exports<'a>(
        &mut self,
        env: fb::TypeEnvironment<'a>,
    ) -> Option<PackageExports> {
        let env = env.assignments()?;
        let mut types = Vec::new();
        for value in env.iter() {
            let assignment: Option<(&'a str, PolyType)> = value.into();
            let (id, ty) = assignment?;
            types.push((
                self.symbols
                    .entry(id.as_ptr())
                    .or_insert_with(|| Symbol::from(id))
                    .clone(),
                ty,
            ));
        }
        PackageExports::try_from(types).ok()
    }
}

impl From<fb::Packages<'_>> for Option<Packages> {
    fn from(fb_packages: fb::Packages<'_>) -> Option<Packages> {
        DeserializeFlatBuffer::default().deserialize_packages(fb_packages)
    }
}

impl From<fb::TypeEnvironment<'_>> for Option<PackageExports> {
    fn from(env: fb::TypeEnvironment) -> Option<PackageExports> {
        DeserializeFlatBuffer::default().deserialize_package_exports(env)
    }
}

impl<'a> From<fb::TypeAssignment<'a>> for Option<(&'a str, PolyType)> {
    fn from(a: fb::TypeAssignment<'a>) -> Self {
        let ty: Option<PolyType> = a.ty()?.into();
        Some((a.id()?, ty?))
    }
}

/// Decodes a PolyType from a flatbuffer
impl From<fb::PolyType<'_>> for Option<PolyType> {
    fn from(t: fb::PolyType) -> Option<PolyType> {
        let v = t.vars()?;
        let mut vars = Vec::new();
        for value in v.iter() {
            vars.push(value.into());
        }
        let c = t.cons()?;
        let mut cons = BoundTvarKinds::new();
        for value in c.iter() {
            let constraint: Option<(BoundTvar, Kind)> = value.into();
            let (tv, kind) = constraint?;
            cons.entry(tv).or_insert_with(Vec::new).push(kind);
        }
        Some(PolyType {
            vars,
            cons,
            expr: from_table(t.expr()?, t.expr_type())?,
        })
    }
}

impl From<fb::Constraint<'_>> for Option<(BoundTvar, Kind)> {
    fn from(c: fb::Constraint) -> Self {
        Some((c.tvar()?.into(), c.kind().into()))
    }
}

impl From<fb::Kind> for Kind {
    fn from(kind: fb::Kind) -> Kind {
        match kind {
            fb::Kind::Addable => Kind::Addable,
            fb::Kind::Subtractable => Kind::Subtractable,
            fb::Kind::Divisible => Kind::Divisible,
            fb::Kind::Numeric => Kind::Numeric,
            fb::Kind::Comparable => Kind::Comparable,
            fb::Kind::Equatable => Kind::Equatable,
            fb::Kind::Label => Kind::Label,
            fb::Kind::Nullable => Kind::Nullable,
            fb::Kind::Record => Kind::Record,
            fb::Kind::Negatable => Kind::Negatable,
            fb::Kind::Timeable => Kind::Timeable,
            fb::Kind::Stringable => Kind::Stringable,
            fb::Kind::Basic => Kind::Basic,
            _ => unreachable!("Unknown fb::Kind"),
        }
    }
}

impl From<Kind> for fb::Kind {
    fn from(kind: Kind) -> fb::Kind {
        match kind {
            Kind::Addable => fb::Kind::Addable,
            Kind::Subtractable => fb::Kind::Subtractable,
            Kind::Divisible => fb::Kind::Divisible,
            Kind::Numeric => fb::Kind::Numeric,
            Kind::Comparable => fb::Kind::Comparable,
            Kind::Equatable => fb::Kind::Equatable,
            Kind::Label => fb::Kind::Label,
            Kind::Nullable => fb::Kind::Nullable,
            Kind::Record => fb::Kind::Record,
            Kind::Negatable => fb::Kind::Negatable,
            Kind::Timeable => fb::Kind::Timeable,
            Kind::Stringable => fb::Kind::Stringable,
            Kind::Basic => fb::Kind::Basic,
        }
    }
}

fn record_label_from_table(table: flatbuffers::Table, t: fb::RecordLabel) -> Option<RecordLabel> {
    match t {
        fb::RecordLabel::Var => {
            let var = fb::Var::init_from_table(table);
            Some(RecordLabel::BoundVariable(BoundTvar::from(var)))
        }
        fb::RecordLabel::Concrete => {
            let concrete = fb::Concrete::init_from_table(table);
            let id = concrete.id()?;
            Some(RecordLabel::from(id))
        }
        fb::RecordLabel::NONE => None,
        _ => unreachable!("Unknown type from table"),
    }
}

fn from_table(table: flatbuffers::Table, t: fb::MonoType) -> Option<MonoType> {
    match t {
        fb::MonoType::Basic => {
            let basic = fb::Basic::init_from_table(table);
            Some(basic.into())
        }
        fb::MonoType::Var => {
            let var = fb::Var::init_from_table(table);
            Some(MonoType::BoundVar(BoundTvar::from(var)))
        }
        fb::MonoType::Collection => {
            let opt: Option<Collection> = fb::Collection::init_from_table(table).into();
            Some(MonoType::from(opt?))
        }
        fb::MonoType::Fun => {
            let opt: Option<Function> = fb::Fun::init_from_table(table).into();
            Some(MonoType::from(opt?))
        }
        fb::MonoType::Record => fb::Record::init_from_table(table).into(),
        fb::MonoType::Dict => {
            let opt: Option<Dictionary> = fb::Dict::init_from_table(table).into();
            Some(MonoType::from(opt?))
        }
        fb::MonoType::NONE => None,
        _ => unreachable!("Unknown type from table"),
    }
}

impl From<fb::Basic<'_>> for MonoType {
    fn from(t: fb::Basic) -> MonoType {
        MonoType::from(match t.t() {
            fb::Type::Bool => BuiltinType::Bool,
            fb::Type::Int => BuiltinType::Int,
            fb::Type::Uint => BuiltinType::Uint,
            fb::Type::Float => BuiltinType::Float,
            fb::Type::String => BuiltinType::String,
            fb::Type::Duration => BuiltinType::Duration,
            fb::Type::Time => BuiltinType::Time,
            fb::Type::Regexp => BuiltinType::Regexp,
            fb::Type::Bytes => BuiltinType::Bytes,
            _ => unreachable!("Unknown fb::Type"),
        })
    }
}

impl From<fb::Var<'_>> for Tvar {
    fn from(t: fb::Var) -> Self {
        Self(t.i())
    }
}

impl From<fb::Var<'_>> for BoundTvar {
    fn from(t: fb::Var) -> Self {
        Self(t.i())
    }
}

impl From<fb::Collection<'_>> for Option<Collection> {
    fn from(t: fb::Collection) -> Self {
        Some(Collection {
            collection: match t.collection() {
                fb::CollectionType::Array => CollectionType::Array,
                fb::CollectionType::Vector => CollectionType::Vector,
                fb::CollectionType::Stream => CollectionType::Stream,
                _ => return None,
            },
            arg: from_table(t.arg()?, t.arg_type())?,
        })
    }
}

impl From<fb::Dict<'_>> for Option<Dictionary> {
    fn from(t: fb::Dict) -> Option<Dictionary> {
        Some(Dictionary {
            key: from_table(t.k()?, t.k_type())?,
            val: from_table(t.v()?, t.v_type())?,
        })
    }
}

impl From<fb::Record<'_>> for Option<MonoType> {
    fn from(t: fb::Record) -> Option<MonoType> {
        let mut r = match t.extends() {
            None => MonoType::from(Record::Empty),
            Some(tv) => MonoType::BoundVar(tv.into()),
        };
        let p = t.props()?;
        for value in p.iter().rev() {
            let prop: Option<Property> = value.into();
            r = MonoType::from(Record::Extension {
                head: prop?,
                tail: r,
            });
        }
        Some(r)
    }
}

impl From<fb::Prop<'_>> for Option<Property> {
    fn from(t: fb::Prop) -> Option<Property> {
        Some(Property {
            k: record_label_from_table(t.k()?, t.k_type())?,
            v: from_table(t.v()?, t.v_type())?,
        })
    }
}

impl From<fb::Fun<'_>> for Option<Function> {
    fn from(t: fb::Fun) -> Option<Function> {
        let args = t.args()?;
        let mut req = MonoTypeMap::new();
        let mut opt = MonoTypeMap::new();
        let mut pipe = None;
        for value in args.iter() {
            match value.into() {
                None => {
                    return None;
                }
                Some((k, v, true, _)) => {
                    pipe = Some(Property { k, v });
                }
                Some((name, t, _, true)) => {
                    opt.insert(name, t.into());
                }
                Some((name, t, false, false)) => {
                    req.insert(name, t);
                }
            };
        }
        Some(Function {
            req,
            opt,
            pipe,
            retn: from_table(t.retn()?, t.retn_type())?,
        })
    }
}

impl From<fb::Argument<'_>> for Option<(String, MonoType, bool, bool)> {
    fn from(t: fb::Argument) -> Option<(String, MonoType, bool, bool)> {
        Some((
            t.name()?.to_owned(),
            from_table(t.t()?, t.t_type())?,
            t.pipe(),
            t.optional(),
        ))
    }
}

pub fn finish_serialize<'a, 'b, S>(
    builder: &'a mut flatbuffers::FlatBufferBuilder<'b>,
    offset: flatbuffers::WIPOffset<S>,
) -> &'a [u8] {
    builder.finish(offset, None);
    builder.finished_data()
}

pub fn serialize<'a, 'b, T, S, F>(
    builder: &'a mut flatbuffers::FlatBufferBuilder<'b>,
    t: T,
    f: F,
) -> &'a [u8]
where
    F: Fn(&mut flatbuffers::FlatBufferBuilder<'b>, T) -> flatbuffers::WIPOffset<S>,
{
    let offset = f(builder, t);
    builder.finish(offset, None);
    builder.finished_data()
}

pub fn deserialize<'a, T: 'a, S>(buf: &'a [u8]) -> S
where
    T: flatbuffers::Follow<'a> + flatbuffers::Verifiable,
    S: std::convert::From<T::Inner>,
{
    flatbuffers::root::<T>(buf).unwrap().into()
}

fn build_vec<T, S, F, B>(v: Vec<T>, b: &mut B, f: F) -> Vec<S>
where
    F: Fn(&mut B, T) -> S,
{
    let mut mapped = Vec::new();
    for t in v {
        mapped.push(f(b, t));
    }
    mapped
}

pub fn build_packages<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    env: Packages,
) -> flatbuffers::WIPOffset<fb::Packages<'a>> {
    let packages = build_vec(env.into_iter().collect(), builder, build_package);
    let packages = builder.create_vector(packages.as_slice());
    fb::Packages::create(
        builder,
        &fb::PackagesArgs {
            packages: Some(packages),
        },
    )
}

fn build_package<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    (id, package): (String, PackageExports),
) -> flatbuffers::WIPOffset<fb::PackageExports<'a>> {
    let id = builder.create_string(&id);
    let package = build_env(builder, package);
    fb::PackageExports::create(
        builder,
        &fb::PackageExportsArgs {
            id: Some(id),
            package: Some(package),
        },
    )
}

pub fn build_env<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    env: PackageExports,
) -> flatbuffers::WIPOffset<fb::TypeEnvironment<'a>> {
    let assignments = build_vec(
        env.into_bindings().collect(),
        builder,
        build_type_assignment,
    );
    let assignments = builder.create_vector(assignments.as_slice());
    fb::TypeEnvironment::create(
        builder,
        &fb::TypeEnvironmentArgs {
            assignments: Some(assignments),
        },
    )
}

pub fn build_module<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    module: Module,
) -> flatbuffers::WIPOffset<fb::Module<'a>> {
    let polytype = module.polytype.map(|pt| build_polytype(builder, pt));
    let code = module
        .code
        .map(|pkg| serialize_pkg_into(&pkg, builder).expect("serialize package"));
    fb::Module::create(builder, &fb::ModuleArgs { polytype, code })
}

fn build_type_assignment<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    assignment: (Symbol, PolyType),
) -> flatbuffers::WIPOffset<fb::TypeAssignment<'a>> {
    let id = builder.create_string(assignment.0.full_name());
    let ty = build_polytype(builder, assignment.1);
    fb::TypeAssignment::create(
        builder,
        &fb::TypeAssignmentArgs {
            id: Some(id),
            ty: Some(ty),
        },
    )
}

/// Encodes a polytype as a flatbuffer
pub fn build_polytype<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    t: PolyType,
) -> flatbuffers::WIPOffset<fb::PolyType<'a>> {
    let vars: Vec<_> = t
        .vars
        .iter()
        .map(|v| build_var(builder, types::Tvar(v.0)))
        .collect();
    let vars = builder.create_vector(vars.as_slice());

    let mut cons = Vec::new();
    for (tv, kinds) in t.cons {
        for k in kinds {
            cons.push((tv, k));
        }
    }
    let cons: Vec<_> = cons
        .into_iter()
        .map(|(t, k)| build_constraint(builder, (Tvar(t.0), k)))
        .collect();
    let cons = builder.create_vector(cons.as_slice());

    let (buf_offset, expr) = build_type(builder, &t.expr);
    fb::PolyType::create(
        builder,
        &fb::PolyTypeArgs {
            vars: Some(vars),
            cons: Some(cons),
            expr_type: expr,
            expr: Some(buf_offset),
        },
    )
}

fn build_constraint<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    constraint: (Tvar, Kind),
) -> flatbuffers::WIPOffset<fb::Constraint<'a>> {
    let tvar = build_var(builder, constraint.0);
    fb::Constraint::create(
        builder,
        &fb::ConstraintArgs {
            tvar: Some(tvar),
            kind: constraint.1.into(),
        },
    )
}

pub fn build_type(
    builder: &mut flatbuffers::FlatBufferBuilder,
    t: &MonoType,
) -> (
    flatbuffers::WIPOffset<flatbuffers::UnionWIPOffset>,
    fb::MonoType,
) {
    match t {
        MonoType::Error => unreachable!(),
        MonoType::Builtin(typ) => build_basic_type(builder, typ),
        MonoType::Label(_) => build_basic_type(builder, &BuiltinType::String),
        MonoType::Var(tvr) => {
            let offset = build_var(builder, *tvr);
            (offset.as_union_value(), fb::MonoType::Var)
        }
        MonoType::BoundVar(tvr) => {
            let offset = build_var(builder, types::Tvar(tvr.0));
            (offset.as_union_value(), fb::MonoType::Var)
        }
        MonoType::Collection(app) => {
            let offset = build_app(builder, app.collection, &app.arg);
            (offset.as_union_value(), fb::MonoType::Collection)
        }
        MonoType::Dict(dict) => {
            let offset = build_dict(builder, dict);
            (offset.as_union_value(), fb::MonoType::Dict)
        }
        MonoType::Record(record) => {
            let offset = build_record(builder, record);
            (offset.as_union_value(), fb::MonoType::Record)
        }
        MonoType::Fun(fun) => {
            let offset = build_fun(builder, fun);
            (offset.as_union_value(), fb::MonoType::Fun)
        }
    }
}

fn build_basic_type(
    builder: &mut flatbuffers::FlatBufferBuilder,
    t: &BuiltinType,
) -> (
    flatbuffers::WIPOffset<flatbuffers::UnionWIPOffset>,
    fb::MonoType,
) {
    let t = match t {
        BuiltinType::Bool => fb::Type::Bool,
        BuiltinType::Int => fb::Type::Int,
        BuiltinType::Uint => fb::Type::Uint,
        BuiltinType::Float => fb::Type::Float,
        BuiltinType::String => fb::Type::String,
        BuiltinType::Duration => fb::Type::Duration,
        BuiltinType::Time => fb::Type::Time,
        BuiltinType::Regexp => fb::Type::Regexp,
        BuiltinType::Bytes => fb::Type::Bytes,
    };
    let a = fb::BasicArgs { t };
    let v = fb::Basic::create(builder, &a);
    (v.as_union_value(), fb::MonoType::Basic)
}

fn build_var<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    var: Tvar,
) -> flatbuffers::WIPOffset<fb::Var<'a>> {
    fb::Var::create(builder, &fb::VarArgs { i: var.0 })
}

fn build_app<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    collection: CollectionType,
    typ: &MonoType,
) -> flatbuffers::WIPOffset<fb::Collection<'a>> {
    let (off, typ) = build_type(builder, typ);
    fb::Collection::create(
        builder,
        &fb::CollectionArgs {
            collection: match collection {
                CollectionType::Array => fb::CollectionType::Array,
                CollectionType::Vector => fb::CollectionType::Vector,
                CollectionType::Stream => fb::CollectionType::Stream,
            },
            arg_type: typ,
            arg: Some(off),
        },
    )
}

fn build_dict<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    mut dict: &Dictionary,
) -> flatbuffers::WIPOffset<fb::Dict<'a>> {
    let (k_offset, k_type) = build_type(builder, &dict.key);
    let (v_offset, v_type) = build_type(builder, &dict.val);
    let (k, v) = (Some(k_offset), Some(v_offset));
    fb::Dict::create(
        builder,
        &fb::DictArgs {
            k_type,
            k,
            v_type,
            v,
        },
    )
}

fn build_record<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    mut record: &Record,
) -> flatbuffers::WIPOffset<fb::Record<'a>> {
    let mut props = Vec::new();

    let mut fields = record.fields();
    for field in &mut fields {
        props.push(field);
    }
    let extends = fields.tail().and_then(|typ| match typ {
        MonoType::Var(t) => Some(*t),
        MonoType::BoundVar(t) => Some(types::Tvar(t.0)),
        _ => None,
    });

    let props = build_vec(props, builder, build_prop);
    let props = builder.create_vector(props.as_slice());
    let extends = extends.map(|typevar| build_var(builder, typevar));
    fb::Record::create(
        builder,
        &fb::RecordArgs {
            props: Some(props),
            extends,
        },
    )
}

fn build_prop<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    prop: &Property,
) -> flatbuffers::WIPOffset<fb::Prop<'a>> {
    let (off, v_type) = build_type(builder, &prop.v);
    let (k, k_type) = match &prop.k {
        RecordLabel::Variable(var) => {
            let concrete = build_var(builder, *var);
            (concrete.as_union_value(), fb::RecordLabel::Var)
        }
        RecordLabel::BoundVariable(var) => {
            let concrete = build_var(builder, types::Tvar(var.0));
            (concrete.as_union_value(), fb::RecordLabel::Var)
        }
        RecordLabel::Concrete(name) => {
            let id = builder.create_string(name);
            let concrete = fb::Concrete::create(builder, &fb::ConcreteArgs { id: Some(id) });
            (concrete.as_union_value(), fb::RecordLabel::Concrete)
        }
        RecordLabel::Error => unreachable!(),
    };
    fb::Prop::create(
        builder,
        &fb::PropArgs {
            k_type,
            k: Some(k),
            v_type,
            v: Some(off),
        },
    )
}

fn build_fun<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    mut fun: &Function,
) -> flatbuffers::WIPOffset<fb::Fun<'a>> {
    let mut args: Vec<(&str, _, _, _)> = Vec::new();
    if let Some(pipe) = &fun.pipe {
        args.push((&pipe.k, &pipe.v, true, false))
    };
    for (k, v) in &fun.req {
        args.push((k, v, false, false));
    }
    for (k, v) in &fun.opt {
        args.push((k, &v.typ, false, true));
    }
    let args = build_vec(args, builder, build_arg);
    let args = builder.create_vector(args.as_slice());

    let (ret, typ) = build_type(builder, &fun.retn);
    fb::Fun::create(
        builder,
        &fb::FunArgs {
            args: Some(args),
            retn_type: typ,
            retn: Some(ret),
        },
    )
}

fn build_arg<'a>(
    builder: &mut flatbuffers::FlatBufferBuilder<'a>,
    arg: (&str, &MonoType, bool, bool),
) -> flatbuffers::WIPOffset<fb::Argument<'a>> {
    let name = builder.create_string(arg.0);
    let (buf_offset, typ) = build_type(builder, arg.1);
    fb::Argument::create(
        builder,
        &fb::ArgumentArgs {
            name: Some(name),
            t_type: typ,
            t: Some(buf_offset),
            pipe: arg.2,
            optional: arg.3,
        },
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    use std::convert::TryInto;

    use crate::{
        ast, parser,
        semantic::{convert::convert_polytype, types::SemanticMap},
    };

    #[rustfmt::skip]
    use crate::semantic::flatbuffers::semantic_generated::fbsemantic::{
        Expression,
        ExpressionStatement,
        ExpressionStatementArgs,
        File,
        FileArgs,
        FloatLiteral,
        FloatLiteralArgs,
        Operator,
        Package,
        PackageArgs,
        Statement,
        UnaryExpression,
        UnaryExpressionArgs,
        WrappedStatement,
        WrappedStatementArgs,
    };

    fn test_serde(expr: &'static str) {
        // let want = parser::parse(expr).unwrap();
        let mut p = parser::Parser::new(expr);

        let typ_expr = p.parse_type_expression();
        if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
            panic!("TypeExpression parsing failed for {}. {:?}", expr, err);
        }
        let want = convert_polytype(&typ_expr, &Default::default()).unwrap();

        let mut builder = flatbuffers::FlatBufferBuilder::new();
        let buf = serialize(&mut builder, want.clone(), build_polytype);
        let got = deserialize::<fb::PolyType, Option<PolyType>>(buf);
        assert_eq!(want, got.unwrap())
    }

    #[test]
    fn serde_type_environment() {
        let mut p = parser::Parser::new("bool");
        let typ_expr = p.parse_type_expression();
        if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
            panic!("TypeExpression parsing failed for bool. {:?}", err);
        }
        let a = convert_polytype(&typ_expr, &Default::default()).unwrap();

        let mut p = parser::Parser::new("time");
        let typ_expr = p.parse_type_expression();
        if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
            panic!("TypeExpression parsing failed for time. {:?}", err);
        }
        let b = convert_polytype(&typ_expr, &Default::default()).unwrap();

        let want: PackageExports = vec![
            (Symbol::from("a"), a.clone()),
            (Symbol::from("b"), b.clone()),
        ]
        .try_into()
        .unwrap();

        let mut builder = flatbuffers::FlatBufferBuilder::new();
        let buf = serialize(&mut builder, want, build_env);
        let got = deserialize::<fb::TypeEnvironment, Option<PackageExports>>(buf);
        let mut deserializer = DeserializeFlatBuffer::default();
        let got = deserializer
            .deserialize_package_exports(flatbuffers::root::<fb::TypeEnvironment>(buf).unwrap());

        let want: PackageExports = vec![
            (
                deserializer
                    .symbols
                    .values()
                    .find(|s| *s == "a")
                    .unwrap()
                    .clone(),
                a,
            ),
            (
                deserializer
                    .symbols
                    .values()
                    .find(|s| *s == "b")
                    .unwrap()
                    .clone(),
                b,
            ),
        ]
        .try_into()
        .unwrap();

        assert_eq!(want, got.unwrap());
    }
    #[test]
    fn serde_basic_types() {
        test_serde("bool");
        test_serde("int");
        test_serde("uint");
        test_serde("float");
        test_serde("string");
        test_serde("duration");
        test_serde("time");
        test_serde("regexp");
        test_serde("bytes");
    }
    #[test]
    fn serde_array_type() {
        test_serde("[A]");
    }
    #[test]
    fn serde_vector_type() {
        let want = PolyType {
            vars: vec![],
            cons: BoundTvarKinds::new(),
            expr: MonoType::vector(MonoType::INT),
        };

        let mut builder = flatbuffers::FlatBufferBuilder::new();
        let buf = serialize(&mut builder, want.clone(), build_polytype);
        let got = deserialize::<fb::PolyType, Option<PolyType>>(buf);
        assert_eq!(want, got.unwrap())
    }
    #[test]
    fn serde_function_types() {
        test_serde("(<-tables: [A], ?flag: bool, fn: (r: A) => bool) => [A]");
        test_serde("(a: A, b: B) => bool where A: Addable, B: Divisible");
    }
    #[test]
    fn serde_record_types() {
        test_serde("{A with a: int , b: float , c: {d: string , d: string , d: time , d: {}}}");
    }
    #[test]
    fn test_flatbuffers_semantic() {
        let mut builder = flatbuffers::FlatBufferBuilder::with_capacity(256);

        // Testing out a unary expression using a float
        let floatval = FloatLiteral::create(
            &mut builder,
            &FloatLiteralArgs {
                value: 3.5,
                ..FloatLiteralArgs::default()
            },
        );

        let increment = UnaryExpression::create(
            &mut builder,
            &UnaryExpressionArgs {
                operator: Operator::SubtractionOperator,
                argument: Some(floatval.as_union_value()),
                ..UnaryExpressionArgs::default()
            },
        );

        let statement = ExpressionStatement::create(
            &mut builder,
            &ExpressionStatementArgs {
                expression_type: Expression::UnaryExpression,
                expression: Some(increment.as_union_value()),
                ..ExpressionStatementArgs::default()
            },
        );

        let wrappedStatement = WrappedStatement::create(
            &mut builder,
            &WrappedStatementArgs {
                statement_type: Statement::ExpressionStatement,
                statement: Some(statement.as_union_value()),
            },
        );

        let statements = builder.create_vector(&[wrappedStatement]);

        let file = File::create(
            &mut builder,
            &FileArgs {
                body: Some(statements),
                ..FileArgs::default()
            },
        );

        let files = builder.create_vector(&[file]);

        let pkg = Package::create(
            &mut builder,
            &PackageArgs {
                files: Some(files),
                ..PackageArgs::default()
            },
        );

        builder.finish(pkg, None);
        let bytes = builder.finished_data();
        assert_ne!(bytes.len(), 0);
    }
}
