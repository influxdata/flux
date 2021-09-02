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

struct FresheningVisitor {
    fresher: Fresher
}

impl FresheningVisitor {
    fn new(fresher: Fresher) -> Self {
        FresheningVisitor {
            fresher
        }
    }
    fn finish(self: Self) -> Fresher {
        self.fresher
    }
}

impl walk::VisitorMut for FresheningVisitor {
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

fn vectorize_fn_type(mut fresher: Fresher, fn_type: &MonoType, vector_arg: &str) -> (MonoType, Constraints, Fresher) {
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
            })
        },
        _ => panic!("expected a function type!")
    };
    (fn_type, cons, fresher)
}

pub fn get_init_type(pkg: &nodes::Package) -> MonoType {
    match &pkg.files[0].body[0] {
        nodes::Statement::Variable(v) => {
            v.init.type_of()
        }
        bad => panic!("expected variable assignment, got {:?}", bad)
    }
}

pub fn vectorize(f: nodes::FunctionExpr, vector_arg: &str) -> (nodes::FunctionExpr, Result<bool, Error>) {
    let fresher = Fresher::from(8888);
    let (vectorized_fn_type, mut init_cons, fresher) = vectorize_fn_type(fresher, &f.typ, vector_arg);
    let vectorized_fn_name = "__vectorized_fn";

    // println!("vectorized_fn_type:\n{:#?}", vectorized_fn_type);
    // println!("new constraints: {:#?}", init_cons);


    let mut sem_pkg = wrap_fn_expr(f, vectorized_fn_name);
    //println!("input function: \n{:#?}", sem_pkg);

    let mut visitor = FresheningVisitor::new(fresher);
    walk::walk_mut(&mut visitor, &mut NodeMut::Package(&mut sem_pkg));
    let mut fresher = visitor.finish();

    //println!("freshly-typed sem_pkg: \n{:#?}", sem_pkg);

    let env = Environment::empty(true);
    init_cons = init_cons + infer::Constraints::from(infer::Constraint::Equal {
        exp: get_init_type(&sem_pkg),
        act: vectorized_fn_type,
        loc: Default::default()
    });

    let (_env, subst) = nodes::infer_pkg_types_with_constraints(&mut sem_pkg, env, init_cons, &mut fresher, &None).unwrap();
    let sem_pkg = nodes::inject_pkg_types(sem_pkg, &subst);

    //println!("new sem_pkg: \n{:#?}", sem_pkg);

    (unwrap_fn_expr(sem_pkg), Ok(true))
}


#[cfg(test)]
mod test {
    use crate::semantic::vectorize::vectorize;
    use crate::semantic::nodes::{FunctionExpr, Statement, Expression};

    fn compile(source: &str) -> FunctionExpr {
        if let Result::Ok(mut pkg) = crate::semantic::convert_source(source) {
            let stmt = pkg.files[0].body.remove(0);
            match stmt {
                Statement::Expr(e) => match e.expression {
                    Expression::Function(fe) => *fe,
                    _ => panic!("expected function expression")
                }
                _ => panic!("expected expression statement")
            }
        } else {
            panic!("could not parse and infer source");
        }
    }


    #[test]
    fn test_vectorize_identity() {
        let f = compile("(r) => ({a: r.a, b: r.b})");
        match vectorize(f, "r") {
            (fe, Ok(b)) => {
                assert_eq!(true, b);
                println!("{:?}", Expression::Function(Box::new(fe)))
            },
            (_, Err(e)) => panic!("got error vectorizing: {:?}", e)
        }
    }

    #[test]
    fn test_vectorize_addition() {
        let f = compile("(r) => ({a: r.a, b: r.b, c: r.a + r.b})");
        match vectorize(f, "r") {
            (fe, Ok(b)) => {
                assert_eq!(true, b);
                println!("{:?}", Expression::Function(Box::new(fe)))
            },
            (_, Err(e)) => panic!("got error vectorizing: {:?}", e)
        }
    }

    #[test]
    fn test_vectorize_addition_with_constant() {
        let f = compile("(r) => ({a: r.a + 1})");
        match vectorize(f, "r") {
            (fe, Ok(b)) => {
                assert_eq!(true, b);
                println!("{:?}", Expression::Function(Box::new(fe)))
            },
            (_, Err(e)) => panic!("got error vectorizing: {:?}", e)
        }
    }

}