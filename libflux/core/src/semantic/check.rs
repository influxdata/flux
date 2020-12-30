use crate::ast;
use crate::semantic::nodes;
use crate::semantic::nodes::Assignment;
use crate::semantic::nodes::Expression;
use crate::semantic::nodes::Statement;
use crate::semantic::walk;
use crate::semantic::walk::Node;

use std::collections::HashMap;
use std::rc::Rc;

// OptionMap maps the name of a Flux option (including an optional package qualifier)
// to its corresponding option statement.
type OptionMap<'a> = HashMap<(Option<&'a str>, &'a str), &'a nodes::OptionStmt>;
type VariableAssignMap<'a> = HashMap<&'a str, Option<&'a nodes::VariableAssgn>>;

/// This function checks a semantic graph, looking for errors.
///
/// This pass is typically run before type inference, so type-related errors occur
/// at a later stage.
///
/// These are the kind of errors that `check()` will find:
/// - Option reassignment: options may only be assigned once within a package
/// - Variable reassignment: variables may only be assigned once within a scope.
///     A variable of the same name may be declared in a different scope.
/// - Dependent options: options declared within the same package must not depend on one another.
///
/// If any of these errors are found, `check()` will return the first one it finds, and `Ok(())` otherwise.
pub fn check(pkg: &nodes::Package) -> Result<(), Error> {
    let opts = check_option_stmts(pkg)?;
    check_vars(pkg, &opts)?;
    check_option_dependencies(&opts)
}

/// This is the error type for errors returned by the `check()` function.
#[derive(Debug)]
pub enum Error {
    /// An assignment after the `option` keyword is not correctly formed.
    InvalidOption(ast::SourceLocation),
    /// An option has been assigned at least two places in the package source.
    OptionReassign(ast::SourceLocation, String),
    /// A variable has been assigned more than once in the same scope.
    VarReassign(ast::SourceLocation, String),
    /// A variable name conflicts with an option name.
    VarReassignOption(ast::SourceLocation, String),
    /// An option depends on another option declared in the same package.
    DependentOptions(ast::SourceLocation, String, String),
}

impl std::fmt::Display for Error {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        match self {
            Error::InvalidOption(sl) => {
                // This seems to be impossible to hit due structure of semantic graphs.
                f.write_fmt(format_args!("{}: invalid option", sl))
            }
            Error::OptionReassign(sl, name) => {
                f.write_fmt(format_args!(r#"{}: option "{}" reassigned"#, sl, name))
            }
            Error::VarReassign(sl, name) => {
                f.write_fmt(format_args!(r#"{}: variable "{}" reassigned"#, sl, name))
            }
            Error::VarReassignOption(sl, name) => f.write_fmt(format_args!(
                r#"{}: variable "{}" conflicts with option of same name"#,
                sl, name
            )),
            Error::DependentOptions(sl, depender, dependee) => f.write_fmt(format_args!(
                r#"{}: option "{}" depends on option "{}", which is defined in the same package"#,
                sl, depender, dependee
            )),
        }
    }
}

impl std::error::Error for Error {}

/// `check_option_stmts` checks that options are not reassigned within a package.
/// Note that options can only appear at file scope since the structure of the semantic
/// graph only allows expression statements, assignments and return statements inside function bodies.
/// As a convenience to later checks, it returns a map of all the option statements in the package.
fn check_option_stmts(pkg: &nodes::Package) -> Result<OptionMap, Error> {
    let mut opt_stmts = vec![];
    for f in &pkg.files {
        for st in &f.body {
            if let Statement::Option(o) = st {
                opt_stmts.push(o.as_ref())
            }
        }
    }

    let mut opts = OptionMap::new();
    for o in opt_stmts {
        let name = get_option_name(o)?;
        if opts.contains_key(&name) {
            return Err(Error::OptionReassign(o.loc.clone(), format_option(name)));
        }
        opts.insert(name, o);
    }
    Ok(opts)
}

fn format_option(opt: (Option<&str>, &str)) -> String {
    match opt {
        (None, id) => String::from(id),
        (Some(pkg), id) => format!("{}.{}", pkg, id),
    }
}

fn get_option_name(o: &nodes::OptionStmt) -> Result<(Option<&str>, &str), Error> {
    match &o.assignment {
        Assignment::Variable(va) => Ok((None, va.id.name.as_str())),
        Assignment::Member(nodes::MemberAssgn {
            member:
                nodes::MemberExpr {
                    object: Expression::Identifier(nodes::IdentifierExpr { name: id, .. }),
                    property,
                    ..
                },
            ..
        }) => Ok((Some(id), property)),
        _ => Err(Error::InvalidOption(o.loc.clone())),
    }
}

/// `check_vars()` returns an error if:
/// - Variables are reassigned within the same block
/// - A variable name clashes with an option name
fn check_vars<'a>(pkg: &'a nodes::Package, opts: &'a OptionMap) -> Result<(), Error> {
    let mut v = VarVisitor {
        opts,
        vars_stack: vec![VariableAssignMap::new()],
        in_option: false,
        err: None,
    };
    walk::walk(&mut v, Rc::new(walk::Node::Package(pkg)));
    match v.err {
        Some(e) => Err(e),
        None => Ok(()),
    }
}

struct VarVisitor<'a> {
    /// a map of all the options in the package
    opts: &'a OptionMap<'a>,
    /// a stack of maps showing which variables are currently in scope
    /// (the last item in the Vec is the most nested scope)
    vars_stack: Vec<VariableAssignMap<'a>>,
    in_option: bool,
    err: Option<Error>,
}

impl<'a> walk::Visitor<'a> for VarVisitor<'a> {
    fn visit(&mut self, node: Rc<Node<'a>>) -> bool {
        if self.err.is_some() {
            return false;
        }
        match *node {
            walk::Node::OptionStmt(_) => {
                self.in_option = true;
            }
            walk::Node::MemberAssgn(_) => {
                // These can only be inside option statements
                self.in_option = false;
            }
            walk::Node::FunctionExpr(_) => self.vars_stack.push(VariableAssignMap::new()),
            walk::Node::FunctionParameter(fp) => {
                let name = fp.key.name.as_str();
                self.vars_stack.last_mut().unwrap().insert(name, None);
            }
            walk::Node::VariableAssgn(va) => {
                if self.in_option {
                    self.in_option = false;
                    return true;
                }
                let name = va.id.name.as_str();
                // if we are at file scope (only one map in vars_stack), a variable assignment could collide with an option.
                if self.vars_stack.len() == 1 && self.opts.contains_key(&(None, name)) {
                    self.err = Some(Error::VarReassignOption(va.loc.clone(), String::from(name)))
                }
                // if most nested (current) scope, already has a variable of this name, return error.
                if self.vars_stack.last().unwrap().contains_key(name) {
                    self.err = Some(Error::VarReassign(va.loc.clone(), String::from(name)));
                    return false;
                }
                self.vars_stack.last_mut().unwrap().insert(name, Some(va));
            }
            _ => (),
        }
        true
    }

    fn done(&mut self, node: Rc<Node<'a>>) {
        if let walk::Node::FunctionExpr(_) = *node {
            self.vars_stack.pop();
        }
    }
}

/// `check_option_dependencies()` checks that no options declared in a package depend on other
/// options also declared in the same package.
fn check_option_dependencies(opts: &OptionMap) -> Result<(), Error> {
    let mut v = OptionDepVisitor {
        opts,
        vars_stack: vec![VariableAssignMap::new()],
        bad_id: None,
    };
    for &o in opts.values() {
        // An option statement like
        //   option foo.bar = "baz"
        // is referring to an option in package "foo", so is allowed.
        match o.assignment {
            Assignment::Member(_) => continue,
            Assignment::Variable(_) => (),
        }
        v.vars_stack[0].clear();
        walk::walk(&mut v, Rc::new(walk::Node::OptionStmt(o)));
        if let Some(id) = v.bad_id {
            let opt_name = get_option_name(o)?;
            return Err(Error::DependentOptions(
                id.loc.clone(),
                format_option(opt_name),
                id.name.clone(),
            ));
        }
    }
    Ok(())
}

struct OptionDepVisitor<'a> {
    opts: &'a OptionMap<'a>,
    vars_stack: Vec<VariableAssignMap<'a>>,
    bad_id: Option<&'a nodes::IdentifierExpr>,
}

impl<'a> walk::Visitor<'a> for OptionDepVisitor<'a> {
    fn visit(&mut self, node: Rc<Node<'a>>) -> bool {
        if self.bad_id.is_some() {
            return false;
        }

        match *node {
            Node::FunctionExpr(_) => self.vars_stack.push(VariableAssignMap::new()),
            Node::FunctionParameter(fp) => {
                let name = fp.key.name.as_str();
                self.vars_stack.last_mut().unwrap().insert(name, None);
            }
            Node::VariableAssgn(va) => {
                let name = va.id.name.as_str();
                self.vars_stack.last_mut().unwrap().insert(name, Some(va));
            }
            Node::IdentifierExpr(ie) => {
                let name = ie.name.as_str();
                if self.opts.contains_key(&(None, name)) {
                    let found = self
                        .vars_stack
                        .iter()
                        .any(|opt_map| opt_map.contains_key(name));
                    if !found {
                        // This is an option that is not shadowed by an enclosing scope
                        // being used inside another option definition
                        self.bad_id = Some(ie);
                        return false;
                    }
                }
            }
            _ => (),
        }
        true
    }

    fn done(&mut self, node: Rc<Node<'a>>) {
        if let Node::FunctionExpr(_) = *node {
            self.vars_stack.pop();
        }
    }
}

#[cfg(test)]
mod tests {
    use crate::ast;
    use crate::parser;
    use crate::semantic::check;
    use crate::semantic::convert;
    use crate::semantic::fresh;
    use crate::semantic::nodes;

    fn merge_ast_files(files: Vec<ast::File>) -> ast::Package {
        let pkg = ast::Package {
            base: ast::BaseNode {
                ..ast::BaseNode::default()
            },
            path: String::from(""),
            package: String::from(&files[0].name),
            files: vec![],
        };
        files.into_iter().fold(pkg, |mut p, f| {
            p.files.push(f);
            p
        })
    }

    fn parse_and_convert(files: Vec<&str>) -> Result<nodes::Package, String> {
        let mut ast_files = vec![];
        let mut ctr = 0;
        for f in files {
            let file = parser::parse_string(format!("file_{}.flux", ctr).as_str(), f);
            ast_files.push(file);
            ctr = ctr + 1;
        }
        let ast_pkg = merge_ast_files(ast_files);
        convert::convert_with(ast_pkg, &mut fresh::Fresher::default())
    }

    fn check_success(files: Vec<&str>) {
        let pkg = match parse_and_convert(files) {
            Err(e) => panic!(e),
            Ok(pkg) => pkg,
        };
        if let Err(e) = check::check(&pkg) {
            panic!(format!("check failed: {}", e))
        }
    }

    fn check_fail(files: Vec<&str>, want_msg: &str) {
        let pkg = match parse_and_convert(files) {
            Ok(pkg) => pkg,
            Err(got_msg) => {
                if !got_msg.contains(want_msg) {
                    panic!(format!(
                        r#"expected error "{}" but got "{}""#,
                        want_msg, got_msg
                    ));
                } else {
                    return ();
                }
            }
        };

        match check::check(&pkg) {
            Ok(()) => panic!(format!(r#"expected error "{}", got no error"#, want_msg)),
            Err(e) => {
                let got_msg = format!("{}", e);
                if !got_msg.contains(want_msg) {
                    panic!(format!(
                        r#"expected error "{}", got error "{}""#,
                        want_msg, got_msg
                    ))
                }
            }
        }
    }

    #[test]
    fn test_option_declarations() {
        // no error
        check_success(vec![
            r#"
                package foo
                option a = 0
                f = () => {
                    a = 0
                    return a + 1
                }
            "#,
        ]);
        // function block
        check_fail(
            vec![
                r#"
                package foo
                f = () => {
                    option bar = 0
                    return 0
                }
            "#,
            ],
            "invalid statement in function block",
        );
        // nested function block
        check_fail(
            vec![
                r#"
                package foo
                f = () => {
                    g = () => {
                        option bar = 0
                        return 0
                    }
                    return 0
                }
            "#,
            ],
            "invalid statement in function block",
        );
        // qualified option
        check_fail(
            vec![
                r#"
                package foo
                import "bar"
                f = () => {
                    option bar.baz = 0
                    return 0
                }
            "#,
            ],
            "invalid statement in function block",
        );
        // multiple files
        check_fail(
            vec![
                r#"
                package foo
                option a = 0
            "#,
                r#"
                package foo
                import "bar"

                x = bar.x
                option bar.x = 0
            "#,
                r#"
                package foo
                option b = 0

                f = () => {
                  a = 1
                  b = 1
                  c = 1
                  return a + b - c
                }
            "#,
                r#"
                package foo
                option c = 0
                g = () => {
                  option d = "d"
                  return 0
                }
            "#,
            ],
            "invalid statement in function block",
        )
    }

    #[test]
    fn test_option_reassignment() {
        // simple
        check_fail(
            vec![
                r#"
                    package foo
                    option a = 0
                    option a = 1
                "#,
            ],
            r#"file_0.flux@4:21-4:33: option "a" reassigned"#,
        );
        // multiple files
        check_fail(
            vec![
                r#"
                    package foo
                    option a = 0
                "#,
                r#"
                    package foo
                    b = 0
                "#,
                r#"
                    package foo
                    option c = 0
                "#,
                r#"
                    package foo
                    option c = 1
                "#,
            ],
            r#"file_3.flux@3:21-3:33: option "c" reassigned"#,
        );
        check_success(vec![
            r#"
                package foo
                option universe.now = () => 2020-01-01T00:00:00.000Z
            "#,
        ]);
        check_success(vec![
            r#"
                import "planner"
                option planner.disablePhysicalRules = ["fromRangeRule"]
                option planner.disableLogicalRules = ["removeCountRule"]
                from(bucket: "bkt") |> range(start: 0) |> filter(fn: (r) => r._value > 0) |> count()
           "#,
        ]);
    }

    #[test]
    fn test_var_reassignment() {
        // no error
        check_success(vec![
            r#"
                    package foo
                    a = 0
                    b = 1
                    c = 2
                "#,
        ]);
        // no error multiple files
        check_success(vec![
            r#"
                    package foo
                    a = 0
                "#,
            r#"
                    package foo
                    b = 0
                "#,
        ]);
        // redeclaration
        check_fail(
            vec![
                r#"
                    package foo
                    a = 0
                    a = 1
                "#,
            ],
            "file_0.flux@4:21-4:26: variable \"a\" reassigned",
        );
        // redec option
        check_fail(
            vec![
                r#"
                    package foo
                    option a = 0
                    a = 1
                "#,
            ],
            "file_0.flux@4:21-4:26: variable \"a\" conflicts with option of same name",
        );
        // shadow
        check_success(vec![
            r#"
                    package foo
                    a = 0
                    f = () => {
                        a = 2
                        return a
                    }
                "#,
        ]);
        // after shadow
        check_success(vec![
            r#"
                    package foo
                    f = () => {
                        a = 2
                        return a
                    }
                    a = 0
                "#,
        ]);
        // redeclaration inside function
        check_fail(
            vec![
                r#"
                    package foo
                    a = 0
                    f = () => {
                        a = 1
                        b = a
                        b = 1
                        return b
                    }
                "#,
            ],
            "file_0.flux@7:25-7:30: variable \"b\" reassigned",
        );
        // redeclaration inside option expression
        check_fail(
            vec![
                r#"
                    package foo
                    a = 0
                    option f = () => {
                        a = 1
                        b = a
                        b = 1
                        return b
                    }
                "#,
            ],
            "file_0.flux@7:25-7:30: variable \"b\" reassigned",
        );
        // reassign parameter
        check_fail(
            vec![
                r#"
                    package foo
                    f = (a) => {
                        a = 1
                        return a
                    }
                "#,
            ],
            "file_0.flux@4:25-4:30: variable \"a\" reassigned",
        );
        // no error option
        check_success(vec![
            r#"
                    package foo
                    option bar = () => {
                        bar = 0
                        return bar
                    }
                "#,
        ]);
        // redec after function
        check_fail(
            vec![
                r#"
                    package foo
                    x = 0
                    f = () => {
                        a = 0
                        b = 0
                        return a + b
                    }
                    x = 1
                "#,
            ],
            "file_0.flux@9:21-9:26: variable \"x\" reassigned",
        );
        // redeclaration multiple files
        check_fail(
            vec![
                r#"
                    package foo
                    a = 0
                    d = a
                "#,
                r#"
                    package foo
                    b = 0
                    f = () => {
                      a = 0
                      b = 0
                      c = 0
                      return a + b + c
                    }
                    c = 0
                "#,
                r#"
                    package foo
                    g = (a, b, c) => {
                      f = (a, b, c) => a + b + c
                      return f(a: a, b: b, c: c)
                    }
                    d = g(a: 0, b: 1, c: 2)
                "#,
            ],
            "file_2.flux@7:21-7:44: variable \"d\" reassigned",
        )
    }

    #[test]
    fn test_option_dependencies() {
        // no error
        check_success(vec![
            r#"
                    package foo
                    option bar = 0
                "#,
            r#"
                    package foo
                    option baz = 0
                "#,
        ]);
        // dependency
        check_fail(
            vec![
                r#"
                    package foo
                    option a = 0
                    option b = a
                "#,
            ],
            "file_0.flux@4:32-4:33: option \"b\" depends on option \"a\", which is defined in the same package"
        );
        // dependency across files
        check_fail(
            vec![
                r#"
                    package foo
                    option a = 0
                "#, r#"
                    package foo
                    option b = a
                "#,
            ],
            "file_1.flux@3:32-3:33: option \"b\" depends on option \"a\", which is defined in the same package"
        );
        // dependency on an export
        check_success(vec![
            r#"
                    package foo
                    import "bar"
                    option a = bar.x
                "#,
        ]);
        // option with same name as export
        check_success(vec![
            r#"
                    package foo
                    import "bar"
                    option a = bar.a.x
                "#,
        ]);
        // nested dependency
        check_fail(
            vec![
                r#"
                    package foo
                    option a = 0
                    option f = () => a
                "#,
            ],
            "file_0.flux@4:38-4:39: option \"f\" depends on option \"a\", which is defined in the same package"
        );
        // nested nested dependency
        check_fail(
            vec![
                r#"
                    package foo
                    option a = 0
                    option f = () => (() => a)()
                "#,
            ],
            "file_0.flux@4:45-4:46: option \"f\" depends on option \"a\", which is defined in the same package"
        );
        // shadow
        check_success(vec![
            r#"
                    package foo
                    import "bar"
                    option a = 0
                    option f = () => {
                        a = 1
                        return a + 1
                    }
                "#,
        ]);
        // param shadow
        check_success(vec![
            r#"
                    package foo
                    import "bar"
                    option a = 0
                    option f = (a) => a
                "#,
        ]);
        // nested shadow
        check_success(vec![
            r#"
                    package foo
                    import "bar"
                    option a = 0
                    option f = () => {
                        a = 0
                        return (() => a + 1)()
                    }
                "#,
        ]);
        // option that shadows import
        check_fail(
            vec![
                r#"
                    package foo
                    import "bar"
                    option bar = {x: 0}
                    option a = bar.x
                "#,
            ],
            "file_0.flux@5:32-5:35: option \"a\" depends on option \"bar\", which is defined in the same package"
        );
        // dependency after shadow
        check_fail(
            vec![
                r#"
                    package foo
                    option a = 0
                    option f = () => {
                        a = 1
                        return a + 1
                    }
                    option b = a
                "#,
            ],
            "file_0.flux@8:32-8:33: option \"b\" depends on option \"a\", which is defined in the same package"
        );
        // dependency with multiple files and shadows
        check_fail(
            vec![
                r#"
                    package foo
                    option a = 0
                "#, r#"
                    package foo
                    option f = () => {
                      a = 1
                      return a + 1
                    }
                "#, r#"
                    package foo
                    option g = (f) => {
                      a = 1
                      g = (g) => g |> f
                      h = (b) => g(g: b)
                      return h(b: a)
                    }
                "#, r#"
                    package foo
                    option b = a
                "#,
            ],
            "file_3.flux@3:32-3:33: option \"b\" depends on option \"a\", which is defined in the same package"
        );
        // assigning option from external package
        check_success(vec![
            r#"
                import "influxdata/influxdb/monitor"
                option monitor.log = (tables=<-) => tables
            "#,
        ]);
    }
}
