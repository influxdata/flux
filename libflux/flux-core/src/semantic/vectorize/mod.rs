use crate::semantic::Error;

use crate::semantic::nodes;
use crate::semantic::infer;

use crate::semantic::types;
use crate::semantic::types::MonoType;

use crate::semantic::walk;
use crate::semantic::walk::NodeMut;

use crate::semantic::fresh::Fresher;
use crate::semantic::nodes::VariableAssgn;
use crate::semantic::env::Environment;
use crate::semantic::infer::Constraints;

// Execute the a function for each proprty of a record type.
fn for_each_property_mut<FN, T>(mt: &mut MonoType, mut f: FN)
  where FN: FnMut(&mut types::Property) -> T {
    let mut mt = mt;
    loop {
        if let MonoType::Record(rt) = mt {
            if let types::Record::Extension {head, tail} = &mut **rt {
                f(head);
                mt = tail
            }  else {
                return
            }
        } else {
            return
        }
    }
}

// fn vectorize_type(mt: MonoType) -> MonoType {
//     return MonoType::Vector(Box::new(types::Vector(mt)))
// }

struct FresheningVisitor<'a> {
    fresher: &'a mut Fresher
}

impl<'a> FresheningVisitor<'a> {
    fn new(fresher: &'a mut Fresher) -> Self {
        FresheningVisitor {
            fresher
        }
    }
}

impl<'a> walk::VisitorMut for FresheningVisitor<'a> {
     fn visit(&mut self, node: &mut NodeMut) -> bool {
         match node {
             NodeMut::IdentifierExpr(ref mut n) => n.typ = MonoType::Var(self.fresher.fresh()),
             NodeMut::ArrayExpr(ref mut n) => n.typ = MonoType::Var(self.fresher.fresh()),
             NodeMut::DictExpr(ref mut n) => n.typ = MonoType::Var(self.fresher.fresh()),
             NodeMut::FunctionExpr(ref mut n) => n.typ = MonoType::Var(self.fresher.fresh()),
             NodeMut::ObjectExpr(ref mut n) => n.typ = MonoType::Var(self.fresher.fresh()),
             NodeMut::MemberExpr(ref mut n) => n.typ = MonoType::Var(self.fresher.fresh()),
             NodeMut::IndexExpr(ref mut n) => n.typ = MonoType::Var(self.fresher.fresh()),
             NodeMut::BinaryExpr(ref mut n) => n.typ = MonoType::Var(self.fresher.fresh()),
             NodeMut::UnaryExpr(ref mut n) => n.typ = MonoType::Var(self.fresher.fresh()),
             NodeMut::CallExpr(ref mut n) => n.typ = MonoType::Var(self.fresher.fresh()),
             _ => (),
         };
         true
     }

    fn done(&mut self, _node: &mut NodeMut<'_>) {
    }
}

struct ExpandingVisitor<'a> {
    fresher: &'a mut Fresher
}

impl<'a> ExpandingVisitor<'a> {
    fn new(fresher: &'a mut Fresher) -> Self {
        ExpandingVisitor{fresher}
    }
}

fn is_scalar_literal(e: &nodes::Expression) -> bool {
    match e {
        nodes::Expression::Integer(_)
        | nodes::Expression::Float(_)
        | nodes::Expression::StringLit(_)
        | nodes::Expression::Duration(_)
        | nodes::Expression::Uint(_)
        | nodes::Expression::Boolean(_)
        | nodes::Expression::DateTime(_)
        | nodes::Expression::Regexp(_) => true,
        _ => false
    }
}

impl<'a> walk::VisitorMut for ExpandingVisitor<'a> {
    fn visit(&mut self, _node: &mut NodeMut) -> bool {
        true
    }

    fn done(&mut self, node: &mut NodeMut) {
        match node {
            // literals might appear in lots of places, but this is just a prototype
            NodeMut::BinaryExpr(be) => {
                if is_scalar_literal(&be.left) {
                    be.left = nodes::Expression::Expand(Box::new(nodes::ExpandExpr {
                        typ: MonoType::Var(self.fresher.fresh()),
                        argument: be.left.clone(),
                    }))
                }
                if is_scalar_literal(&be.right) {
                    be.right = nodes::Expression::Expand(Box::new(nodes::ExpandExpr {
                        typ: MonoType::Var(self.fresher.fresh()),
                        argument: be.right.clone(),
                    }))
                }
            }
            _ => ()
        }
    }
}

fn wrap_fn_expr(f: nodes::FunctionExpr, name: &str) -> nodes::Package {
    let id = nodes::Identifier {
        loc: Default::default(),
        name: name.to_string(),
    };
    let var_asgn = VariableAssgn::new(
        id,
        nodes::Expression::Function(Box::new(f)),
        Default::default(),
    );
    nodes::Package {
        loc: Default::default(),
        package: "".to_string(),
        files: vec![
            nodes::File {
                loc: Default::default(),
                package: None,
                imports: vec![],
                body: vec![
                    nodes::Statement::Variable(Box::new(var_asgn))
                ]
            }
        ]
    }
}

fn unwrap_fn_expr(sem_pkg: nodes::Package) -> nodes::FunctionExpr {
    let mut sem_pkg = sem_pkg;
    let stmt = sem_pkg.files[0].body.remove(0);
    let fe = match stmt {
      nodes::Statement::Variable(va) => {
          match va.init {
              nodes::Expression::Function(fe) => *fe,
              _ => panic!("expected function expression")
          }
      },
        _ => panic!("expected assignment")

    };
    fe
}

fn vectorize_fn_type(fresher: &mut Fresher, fn_type: &MonoType, vector_arg: &str) -> (MonoType, Constraints) {
    let mut fn_type = fn_type.clone();
    let mut cons = infer::Constraints::empty();

    match &mut fn_type {
        MonoType::Fun(f) => {
            let arg_type = f.req.get_mut(vector_arg).unwrap();
            for_each_property_mut(arg_type, |prop| {
                let new_tvar = MonoType::Var(fresher.fresh());
                cons.add(infer::Constraint::Equal {
                    exp: new_tvar.clone(),
                    //act: MonoType::Vector(Box::new(types::Vector(prop.v.clone()))),
                    act: MonoType::Vector(Box::new(types::Vector(MonoType::Var(fresher.fresh())))),
                    loc: Default::default()
                });
                prop.v = new_tvar;
            });
            for_each_property_mut(&mut f.retn, |prop| {
                let new_tvar = MonoType::Var(fresher.fresh());
                cons.add(infer::Constraint::Equal {
                    exp: new_tvar.clone(),
                    //act: MonoType::Vector(Box::new(types::Vector(prop.v.clone()))),
                    act: MonoType::Vector(Box::new(types::Vector(MonoType::Var(fresher.fresh())))),
                    loc: Default::default()
                });
                prop.v = new_tvar;
            })
        },
        _ => panic!("expected a function type!")
    };
    (fn_type, cons)
}

pub fn get_init_type(pkg: &nodes::Package) -> MonoType {
    match &pkg.files[0].body[0] {
        nodes::Statement::Variable(v) => {
            v.init.type_of()
        }
        bad => panic!("expected variable assignment, got {:?}", bad)
    }
}

pub fn vectorize(env: Environment, fresher: &mut Fresher, f: nodes::FunctionExpr, vector_arg: &str) -> (nodes::FunctionExpr, Result<bool, Error>) {
    let (vectorized_fn_type, mut init_cons) = vectorize_fn_type(fresher, &f.typ, vector_arg);
    let vectorized_fn_name = "__vectorized_fn";

    println!("vectorized_fn_type:\n{:#?}", vectorized_fn_type);
    // println!("new constraints: {:#?}", init_cons);


    let mut sem_pkg = wrap_fn_expr(f, vectorized_fn_name);
    //println!("input function: \n{:#?}", sem_pkg);

    let mut visitor = FresheningVisitor::new(fresher);
    walk::walk_mut(&mut visitor, &mut NodeMut::Package(&mut sem_pkg));

    let mut visitor = ExpandingVisitor::new(fresher);
    walk::walk_mut(&mut visitor, &mut NodeMut::Package(&mut sem_pkg));

    //println!("expanded sem_pkg: \n{:#?}", sem_pkg);

    init_cons = init_cons + infer::Constraints::from(infer::Constraint::Equal {
        exp: get_init_type(&sem_pkg),
        act: vectorized_fn_type,
        loc: Default::default()
    });

    let (_env, subst) = nodes::infer_pkg_types_with_constraints(&mut sem_pkg, env, init_cons, fresher, &None).unwrap();
    let sem_pkg = nodes::inject_pkg_types(sem_pkg, &subst);

    //println!("new sem_pkg: \n{:#?}", sem_pkg);

    (unwrap_fn_expr(sem_pkg), Ok(true))
}


#[cfg(test)]
mod test {
    use crate::semantic;
    use crate::semantic::vectorize::vectorize;
    use crate::semantic::nodes::{FunctionExpr, Statement, Expression};
    use crate::semantic::env::Environment;
    use crate::semantic::fresh::Fresher;
    use crate::semantic::types;
    use crate::semantic::types::{MonoType, PolyType};
    use crate::semantic_map;

    fn compile(fresher: &mut Fresher, env: Environment, source: &str) -> (Environment, FunctionExpr) {
        if let Result::Ok((env, mut pkg)) = semantic::convert_source_with_env(fresher, env, source) {
            let stmt = pkg.files[0].body.remove(0);
            match stmt {
                Statement::Expr(e) => match e.expression {
                    Expression::Function(fe) => (env, *fe),
                    _ => panic!("expected function expression")
                }
                _ => panic!("expected expression statement")
            }
        } else {
            panic!("could not parse and infer source");
        }
    }

    fn vectorize_flux_ez(src: &str) {
        let mut fresher = Fresher::from(8888);
        vectorize_flux(Environment::empty(false), &mut fresher, src)
    }

    fn vectorize_flux(env: Environment, fresher: &mut Fresher, src: &str) {
        let (env, fn_expr) = compile(fresher, env, src);
        match vectorize(env, fresher, fn_expr, "r") {
            (fe, Ok(b)) => {
                assert_eq!(true, b);
                println!("{:#?}", Expression::Function(Box::new(fe)))
            },
            (_, Err(e)) => panic!("got error vectorizing: {:#?}", e)
        }
    }

    #[test]
    fn test_vectorize_identity() {
        vectorize_flux_ez("(r) => ({a: r.a, b: r.b})");
    }

    #[test]
    fn test_vectorize_addition() {
        vectorize_flux_ez("(r) => ({a: r.a, b: r.b, c: r.a + r.b})");
    }

    #[test]
    fn test_vectorize_addition_with_constant() {
        vectorize_flux_ez("(r) => ({a: r.a + 1})");
    }

    #[test]
    fn test_vectorize_id_fn() {
        let mut fresher = Fresher::from(7777);
        let tv = fresher.fresh();
        let fnty = PolyType {
            vars: vec![tv.clone()],
            cons: Default::default(),
            expr: types::MonoType::Fun(Box::new(types::Function {
                req: semantic_map!["v".to_string() => MonoType::Var(tv.clone())],
                opt: Default::default(),
                pipe: None,
                retn: MonoType::Var(tv.clone())
            }))
        };
        let env = Environment::from(semantic_map! {
            "id_fn".to_string() => fnty
        });
        vectorize_flux(env.clone(), &mut fresher, "(r) => ({a: id_fn(v: r.a)})");
        vectorize_flux(env, &mut fresher, "(r) => ({a: id_fn(v: r.a + 1)})");
    }

    #[test]
    fn test_vectorize_int_fn() {
        let mut fresher = Fresher::from(7777);
        //let tv = fresher.fresh();
        let fnty = PolyType {
            vars: vec![],
            cons: Default::default(),
            expr: types::MonoType::Fun(Box::new(types::Function {
                req: semantic_map!["v".to_string() => MonoType::Int],
                opt: Default::default(),
                pipe: None,
                retn: MonoType::Int
            }))
        };
        let env = Environment::from(semantic_map! {
            "id_fn".to_string() => fnty
        });
        //vectorize_flux(env.clone(), &mut fresher, "(r) => ({a: id_fn(v: r.a)})");
        vectorize_flux(env, &mut fresher, "(r) => ({a: id_fn(v: r.a + 1)})");
    }
}
