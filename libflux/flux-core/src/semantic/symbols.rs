//! Visitors for

use crate::{
    ast,
    errors::located,
    map::HashSet,
    semantic::{
        nodes::{Package, Statement},
        types::MonoType,
        walk::{walk, Node, Visitor},
        Symbol, Warning, WarningKind,
    },
};

pub struct DefinitionVisitor<F> {
    consumer: F,
}

impl<F> DefinitionVisitor<F> {
    pub fn new(consumer: F) -> Self {
        DefinitionVisitor { consumer }
    }
}

impl<'a, F> Visitor<'a> for DefinitionVisitor<F>
where
    F: FnMut(&'a Symbol, &'a ast::SourceLocation, Node<'a>),
{
    fn visit(&mut self, node: Node<'a>) -> bool {
        match node {
            Node::VariableAssgn(va) => (self.consumer)(&va.id.name, &va.id.loc, node),
            Node::BuiltinStmt(builtin) => (self.consumer)(&builtin.id.name, &builtin.id.loc, node),
            Node::ImportDeclaration(import) => {
                (self.consumer)(&import.import_symbol, &import.loc, node)
            }
            Node::FunctionExpr(func) => {
                for param in &func.params {
                    (self.consumer)(&param.key.name, &param.key.loc, node);
                }
            }
            _ => (),
        }
        true
    }
}

pub struct UseVisitor<F> {
    consumer: F,
}

impl<F> UseVisitor<F> {
    pub fn new(consumer: F) -> Self {
        UseVisitor { consumer }
    }
}

impl<'a, F> Visitor<'a> for UseVisitor<F>
where
    F: FnMut(&Symbol, &ast::SourceLocation),
{
    fn visit(&mut self, node: Node<'a>) -> bool {
        match node {
            Node::IdentifierExpr(id) => (self.consumer)(&id.name, &id.loc),
            Node::File(file) => {
                // All top-level bindings are "used" through being exported
                for stmt in &file.body {
                    match stmt {
                        Statement::Variable(va) => (self.consumer)(&va.id.name, &va.loc),
                        Statement::Builtin(builtin) => {
                            (self.consumer)(&builtin.id.name, &builtin.loc)
                        }
                        _ => (),
                    }
                }
            }
            _ => (),
        }
        true
    }
}

pub fn unused_symbols(node: &Package) -> Vec<Warning> {
    let mut uses = HashSet::new();

    walk(
        &mut UseVisitor::new(|symbol: &Symbol, _: &ast::SourceLocation| {
            uses.insert(symbol.clone());
        }),
        node.into(),
    );

    let mut warnings = Vec::new();

    walk(
        &mut DefinitionVisitor::new(|id, loc: &ast::SourceLocation, node| {
            // Function parameters are part of the function type so users may need to
            // specify a parameter for the program to type check even if they do not use it.
            // So we do not emit warnings for unused function parameters.
            if !matches!(node, Node::FunctionExpr(_)) && !uses.contains(id) {
                warnings.push(located(
                    loc.clone(),
                    WarningKind::UnusedSymbol(id.to_string()),
                ));
            }
        }),
        node.into(),
    );

    warnings
}

/// Given a Flux source and a variable name, find out the type of that variable in the Flux source code.
pub fn find_var_type(pkg: &Package, var_name: &str) -> Option<MonoType> {
    // `var_name` refers to an identifier without a definition so we gather all the symbols that
    // have definitions to filter those out
    let mut definitions = HashSet::new();

    walk(
        &mut DefinitionVisitor::new(|symbol: &Symbol, _, _| {
            definitions.insert(symbol.clone());
        }),
        pkg.into(),
    );

    let mut typ = None;
    walk(
        &mut |node| {
            if let Node::IdentifierExpr(id) = node {
                if id.name == var_name && !definitions.contains(&id.name) {
                    typ = Some(id.typ.clone());
                }
            }
        },
        pkg.into(),
    );

    typ
}
