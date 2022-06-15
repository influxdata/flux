//! Source code formatter.

use anyhow::{anyhow, Error, Result};
use chrono::SecondsFormat;
use pretty::{docs, DocAllocator};

use crate::{
    ast::{self, walk::Node, File, Statement},
    parser::parse_string,
};

/// Format a [`File`].
pub fn convert_to_string(file: &File) -> Result<String> {
    format_to_string(file, true)
}

/// Format a string of Flux code.
///
/// # Example
///
/// ```rust
/// # use fluxcore::formatter::format;
/// let source = "(r) => r.user ==              \"user1\"";
/// let formatted = format(source).unwrap();
/// assert_eq!(formatted, "(r) => r.user == \"user1\"");
/// ```
pub fn format(contents: &str) -> Result<String> {
    let file = parse_string("".to_string(), contents);
    let node = ast::walk::Node::File(&file);
    ast::check::check(node)?;

    convert_to_string(&file)
}

const MULTILINE: usize = 4;

type Arena<'doc> = pretty::Arena<'doc>;
type Doc<'doc> = pretty::DocBuilder<'doc, Arena<'doc>, ()>;

fn format_item_list<'doc>(
    arena: &'doc Arena<'doc>,
    (open, close): (&'doc str, &'doc str),
    trailing_comments: Doc<'doc>,
    items: impl ExactSizeIterator<Item = Doc<'doc>>,
) -> (Doc<'doc>, Doc<'doc>, Doc<'doc>) {
    let multiline = items.len() > MULTILINE;
    let line = if multiline {
        arena.hardline()
    } else {
        arena.line()
    };
    let line_ = if multiline {
        arena.hardline()
    } else {
        arena.line_()
    };

    (
        arena.text(open),
        docs![arena, line_.clone(), comma_list_with(arena, items, line),],
        docs![
            arena,
            if let ::pretty::Doc::Nil = &*trailing_comments {
                line_
            } else {
                arena.nil()
            },
            trailing_comments,
            arena.text(close)
        ],
    )
}

pub(crate) fn comma_list_with<'doc, I>(
    arena: &'doc Arena<'doc>,
    docs: impl IntoIterator<Item = Doc<'doc>, IntoIter = I>,
    line: Doc<'doc>,
) -> Doc<'doc>
where
    I: Iterator<Item = Doc<'doc>>,
{
    let mut docs = docs.into_iter().peekable();
    let trailing_comma = if docs.peek().is_none() {
        arena.nil()
    } else {
        arena.text(",").flat_alt(arena.nil())
    };
    arena
        .intersperse(
            docs.map(move |doc| doc.group()),
            arena.text(",").append(line.clone()),
        )
        .append(trailing_comma)
}

fn comma_list_without_trailing_comma<'doc>(
    arena: &'doc Arena<'doc>,
    docs: impl IntoIterator<Item = Doc<'doc>>,
    line: Doc<'doc>,
) -> Doc<'doc> {
    arena.intersperse(docs, arena.text(",").append(line))
}

/// Constructs a document which attempts to "hang" its prefix  on the same line to reduce how much
/// the `body` needs to be indented. To do this it tries to fit these (example) layouts in order,
/// selecting the first that fits
///
/// ```flux
/// foo = () => { x: 1 }
/// ```
/// gets turned into
///
/// Prefixes: `[foo =, () =>, {]`
/// Body: `x: 1`
/// Suffixes: `[}]`
///
/// Layouts:
///
/// ```flux
/// foo = () => {
///     x: 1,
/// }
/// ```
///
/// ```flux
/// foo = () =>
///     {
///         x: 1,
///     }
/// ```
///
/// ```flux
/// foo =
///     () =>
///         {
///             x: 1,
///         }
/// ```
///
/// (The record in these layouts are laid out on multiple lines, however if they fit they will of
/// course still be on a single line)
fn format_hang_doc<'doc>(
    arena: &'doc Arena<'doc>,
    surrounding: &[Affixes<'doc>],
    body: Doc<'doc>,
) -> Doc<'doc> {
    let fail_on_multi_line = arena.fail().flat_alt(arena.nil());

    (1..surrounding.len() + 1)
        .rev()
        .map(|split| {
            let (before, after) = surrounding.split_at(split);
            let last = before.len() == 1;
            docs![
                arena,
                docs![
                    arena,
                    arena.concat(before.iter().map(|affixes| affixes.prefix.clone())),
                    if last {
                        arena.nil()
                    } else {
                        fail_on_multi_line.clone()
                    }
                ]
                .group(),
                docs![
                    arena,
                    after.iter().rev().cloned().fold(
                        docs![
                            arena,
                            body.clone(),
                            // If there is no prefix then we must not allow the body to laid out on multiple
                            // lines without nesting
                            if !last
                                && before
                                    .iter()
                                    .all(|affixes| matches!(&*affixes.prefix.1, ::pretty::Doc::Nil))
                            {
                                fail_on_multi_line.clone()
                            } else {
                                arena.nil()
                            },
                        ]
                        .nest(INDENT)
                        .append(arena.concat(after.iter().map(|affixes| affixes.suffix.clone()))),
                        |acc, affixes| {
                            let mut doc = affixes.prefix.append(acc);
                            if affixes.nest {
                                doc = doc.nest(INDENT);
                            }
                            doc.group()
                        },
                    ),
                    arena.concat(before.iter().map(|affixes| affixes.suffix.clone())),
                ]
                .group(),
            ]
        })
        .fold(None::<Doc<'doc>>, |acc, doc| {
            Some(match acc {
                None => doc,
                Some(acc) => acc.union(doc),
            })
        })
        .unwrap_or(body)
}

#[derive(Clone)]
struct Affixes<'doc> {
    prefix: Doc<'doc>,
    suffix: Doc<'doc>,
    nest: bool,
}

impl Affixes<'_> {
    fn nest(mut self) -> Self {
        self.nest = true;
        self
    }
}

fn affixes<'doc>(prefix: Doc<'doc>, suffix: Doc<'doc>) -> Affixes<'doc> {
    Affixes {
        prefix,
        suffix,
        nest: false,
    }
}

struct HangDoc<'doc> {
    affixes: Vec<Affixes<'doc>>,
    body: Doc<'doc>,
}

impl<'doc> HangDoc<'doc> {
    fn add_prefix(&mut self, doc: Doc<'doc>) {
        if let Some(affixes) = self.affixes.last_mut() {
            affixes.prefix = doc.append(affixes.prefix.clone());
        } else {
            self.body = doc.append(self.body.clone());
        }
    }

    fn format(mut self) -> Doc<'doc> {
        self.affixes.reverse();
        format_hang_doc(self.body.0, &self.affixes, self.body)
    }
}

fn format_to_string(file: &File, include_pkg: bool) -> Result<String> {
    let arena = Arena::new();
    let mut formatter = Formatter {
        arena: &arena,
        err: None,
    };
    let doc = formatter.format_file(file, include_pkg).group().1;
    if let Some(err) = formatter.err {
        return Err(err);
    }
    let formatted = doc.pretty(120).to_string();
    // Remove indentation from whitespace only lines
    Ok(formatted
        .split('\n')
        .map(|s| s.trim_end())
        .collect::<Vec<_>>()
        .join("\n"))
}

struct Formatter<'doc> {
    arena: &'doc Arena<'doc>,
    err: Option<Error>,
}

#[allow(dead_code, unused_variables)]
impl<'doc> Formatter<'doc> {
    fn base_multiline(&self, base: &ast::BaseNode) -> Doc<'doc> {
        self.multiline(base.is_multiline())
    }

    fn multiline(&self, multiline: bool) -> Doc<'doc> {
        if multiline {
            self.arena.hardline()
        } else {
            self.arena.line()
        }
    }

    fn base_multiline_(&self, base: &ast::BaseNode) -> Doc<'doc> {
        self.multiline_(base.is_multiline())
    }

    fn multiline_(&self, multiline: bool) -> Doc<'doc> {
        if multiline {
            self.arena.hardline()
        } else {
            self.arena.line_()
        }
    }

    fn format_file(&mut self, file: &'doc File, include_pkg: bool) -> Doc<'doc> {
        let arena = self.arena;
        let mut doc = arena.nil();
        if let Some(pkg) = &file.package {
            if include_pkg && !pkg.name.name.is_empty() {
                doc = docs![
                    arena,
                    doc,
                    self.format_comments(&pkg.base.comments),
                    "package ",
                    self.format_identifier(&pkg.name),
                    arena.hardline(),
                    if !file.imports.is_empty() || !file.body.is_empty() {
                        arena.hardline().append(arena.hardline())
                    } else {
                        arena.nil()
                    }
                ];
            }
        }

        doc = docs![
            arena,
            doc,
            arena.intersperse(
                file.imports
                    .iter()
                    .map(|import| self.format_import_declaration(import)),
                arena.hardline()
            ),
        ];
        if !file.imports.is_empty() && !file.body.is_empty() {
            doc = docs![arena, doc, arena.hardline(), arena.hardline(),];
        }

        // format the file statements
        doc = doc.append(self.format_statement_list(&file.body));

        if !file.eof.is_empty() {
            doc = doc.append(self.format_comments(&file.eof));
        }

        doc
    }

    fn format_append_comments(&mut self, comments: &'doc [ast::Comment]) -> Doc<'doc> {
        let arena = self.arena;
        let mut doc = arena.nil();
        if !comments.is_empty() {
            doc = arena.line();
        }
        docs![arena, doc, self.format_comments(comments)].nest(INDENT)
    }

    fn format_comments(&mut self, comments: &'doc [ast::Comment]) -> Doc<'doc> {
        let arena = self.arena;
        arena.concat(comments.iter().map(|c| {
            arena.intersperse(
                c.text.split('\n').map(|part| arena.text(part)),
                arena.hardline(),
            )
        }))
    }

    fn format_type_expression(&mut self, n: &'doc ast::TypeExpression) -> Doc<'doc> {
        let arena = self.arena;
        docs![
            arena,
            self.format_monotype(&n.monotype),
            if !n.constraints.is_empty() {
                let line = self.multiline(n.constraints.len() > MULTILINE);

                docs![
                    arena,
                    line.clone(),
                    "where",
                    line.clone(),
                    comma_list_without_trailing_comma(
                        arena,
                        n.constraints.iter().map(|c| docs![
                            arena,
                            self.format_identifier(&c.tvar),
                            ": ",
                            self.format_kinds(&c.kinds),
                        ]
                        .group()),
                        line,
                    ),
                ]
            } else {
                arena.nil()
            }
        ]
    }

    fn format_kinds(&mut self, n: &'doc [ast::Identifier]) -> Doc<'doc> {
        let arena = self.arena;
        arena.intersperse(
            n.iter().map(|k| self.format_identifier(k)),
            arena.text(" + "),
        )
    }

    fn format_monotype(&mut self, n: &'doc ast::MonoType) -> Doc<'doc> {
        let arena = self.arena;
        match n {
            ast::MonoType::Tvar(tv) => self.format_identifier(&tv.name),
            ast::MonoType::Basic(nt) => self.format_identifier(&nt.name),
            ast::MonoType::Array(arr) => {
                docs![arena, "[", self.format_monotype(&arr.element), "]",]
            }
            ast::MonoType::Stream(stream) => {
                docs![arena, "stream[", self.format_monotype(&stream.element), "]",]
            }
            ast::MonoType::Vector(vector) => {
                docs![arena, "vector[", self.format_monotype(&vector.element), "]",]
            }
            ast::MonoType::Dict(dict) => {
                docs![
                    arena,
                    "[",
                    self.format_monotype(&dict.key),
                    ":",
                    self.format_monotype(&dict.val),
                    "]",
                ]
            }
            ast::MonoType::Record(n) => {
                let multiline = n.properties.len() > MULTILINE;
                let line = self.multiline(multiline);
                let line_ = self.multiline_(multiline);

                docs![
                    arena,
                    self.format_comments(&n.base.comments),
                    "{",
                    docs![
                        arena,
                        line_.clone(),
                        if let Some(tv) = &n.tvar {
                            docs![
                                arena,
                                self.format_identifier(tv),
                                arena.line(),
                                "with",
                                arena.line(),
                            ]
                        } else {
                            arena.nil()
                        },
                        comma_list_with(
                            arena,
                            n.properties.iter().map(|p| {
                                docs![
                                    arena,
                                    self.format_property_key(&p.name),
                                    ": ",
                                    self.format_monotype(&p.monotype),
                                ]
                                .group()
                            }),
                            line,
                        ),
                    ]
                    .nest(INDENT),
                    line_,
                    "}",
                ]
            }
            ast::MonoType::Function(n) => {
                let multiline = n.parameters.len() > MULTILINE;
                let line = self.multiline(multiline);
                let line_ = self.multiline_(multiline);

                docs![
                    arena,
                    self.format_comments(&n.base.comments),
                    "(",
                    docs![
                        arena,
                        line_.clone(),
                        comma_list_with(
                            arena,
                            n.parameters.iter().map(|p| self.format_parameter_type(p)),
                            line,
                        ),
                    ]
                    .nest(INDENT),
                    line_.clone(),
                    ")",
                    " => ",
                    self.format_monotype(&n.monotype),
                ]
            }
            ast::MonoType::Label(label) => self.format_string_literal(label),
        }
        .group()
    }

    fn format_parameter_type(&mut self, n: &'doc ast::ParameterType) -> Doc<'doc> {
        let arena = self.arena;
        match &n {
            ast::ParameterType::Required {
                base: _,
                name,
                monotype,
            } => {
                docs![
                    arena,
                    self.format_identifier(name),
                    ": ",
                    self.format_monotype(monotype),
                ]
            }
            ast::ParameterType::Optional {
                base: _,
                name,
                monotype,
                default,
            } => {
                docs![
                    arena,
                    "?",
                    self.format_identifier(name),
                    ": ",
                    self.format_monotype(monotype),
                    match default {
                        Some(default) => docs![arena, " = ", self.format_string_literal(default)],
                        None => arena.nil(),
                    }
                ]
            }
            ast::ParameterType::Pipe {
                base: _,
                name,
                monotype,
            } => {
                docs![
                    arena,
                    "<-",
                    match name {
                        Some(n) => self.format_identifier(n),
                        None => arena.nil(),
                    },
                    ": ",
                    self.format_monotype(monotype),
                ]
            }
        }
    }

    fn format_import_declaration(&mut self, n: &'doc ast::ImportDeclaration) -> Doc<'doc> {
        let arena = self.arena;
        docs![
            arena,
            self.format_comments(&n.base.comments),
            "import ",
            if let Some(alias) = &n.alias {
                if !alias.name.is_empty() {
                    docs![arena, self.format_identifier(alias), " "]
                } else {
                    arena.nil()
                }
            } else {
                arena.nil()
            },
            self.format_string_literal(&n.path)
        ]
    }

    fn format_block(&mut self, n: &'doc ast::Block) -> HangDoc<'doc> {
        let arena = self.arena;
        HangDoc {
            affixes: vec![affixes(
                docs![arena, self.format_comments(&n.lbrace), "{"],
                docs![arena, arena.hardline(), "}"],
            )],
            body: docs![
                arena,
                arena.hardline(),
                // format the block statements
                self.format_statement_list(&n.body),
                self.format_comments(&n.rbrace),
            ],
        }
    }

    fn format_statement_list(&mut self, s: &'doc [Statement]) -> Doc<'doc> {
        let arena = self.arena;

        let mut prev: i8 = -1;
        let mut previous_location: i32 = -1;
        arena.intersperse(
            s.iter().enumerate().map(|(i, stmt)| {
                let mut extra_line = arena.nil();

                let cur = stmt.typ();
                if i != 0 {
                    let current_location: i32 = stmt.base().location.start.line as i32;
                    //compare the line position of adjacent lines to preserve formatted double new lines
                    let line_gap = current_location - previous_location;
                    // separate different statements with double newline or statements with comments
                    if line_gap > 1 || cur != prev || starts_with_comment(Node::from_stmt(stmt)) {
                        extra_line = arena.hardline();
                    }
                }
                previous_location = stmt.base().location.end.line as i32;
                prev = cur;

                extra_line.append(self.format_statement(stmt))
            }),
            arena.hardline(),
        )
    }

    fn format_assignment(&mut self, n: &'doc ast::Assignment) -> Doc<'doc> {
        let arena = self.arena;
        match n {
            ast::Assignment::Variable(n) => {
                let mut hang_doc = self.hang_expression(&n.init);
                hang_doc.add_prefix(arena.line());
                hang_doc.affixes.push(
                    affixes(
                        docs![
                            arena,
                            self.format_identifier(&n.id),
                            self.format_append_comments(&n.base.comments),
                            " =",
                        ],
                        arena.nil(),
                    )
                    .nest(),
                );
                hang_doc.format()
            }
            ast::Assignment::Member(n) => {
                let mut hang_doc = self.hang_expression(&n.init);
                hang_doc.add_prefix(arena.line());
                hang_doc.affixes.push(
                    affixes(
                        docs![
                            arena,
                            self.format_member_expression(&n.member),
                            self.format_append_comments(&n.base.comments),
                            " =",
                        ],
                        arena.nil(),
                    )
                    .nest(),
                );
                hang_doc.format()
            }
        }
    }

    fn format_statement(&mut self, s: &'doc Statement) -> Doc<'doc> {
        let arena = self.arena;
        match s {
            Statement::Expr(s) => self.format_expression(&s.expression),
            Statement::Variable(s) => self.format_variable_assignment(s),
            Statement::Option(s) => {
                docs![
                    arena,
                    self.format_comments(&s.base.comments),
                    "option ",
                    self.format_assignment(&s.assignment),
                ]
            }
            Statement::Return(s) => {
                let prefix = docs![arena, self.format_comments(&s.base.comments), "return"];
                let mut hang_doc = self.hang_expression(&s.argument);
                hang_doc.add_prefix(arena.line());
                hang_doc.affixes.push(affixes(prefix, arena.nil()).nest());
                hang_doc.format()
            }
            Statement::Bad(s) => {
                self.err = Some(anyhow!("bad statement"));
                arena.nil()
            }
            Statement::Test(n) => {
                docs![
                    arena,
                    self.format_comments(&n.base.comments),
                    "test ",
                    self.format_variable_assignment(&n.assignment),
                ]
            }
            Statement::TestCase(n) => {
                let comment = self.format_comments(&n.base.comments);
                let prefix = docs![
                    arena,
                    "testcase",
                    arena.line(),
                    self.format_identifier(&n.id),
                    if let Some(extends) = &n.extends {
                        docs![
                            arena,
                            arena.line(),
                            "extends",
                            arena.line(),
                            self.format_string_literal(extends),
                        ]
                    } else {
                        arena.nil()
                    },
                    arena.line(),
                ];

                let mut hang_doc = self.format_block(&n.block);
                hang_doc.affixes.push(affixes(prefix, arena.nil()).nest());
                docs![
                    arena,
                    // Do not put the leading comment into the hang_doc so that
                    // the comment size doesn't affect the hang layout.
                    comment,
                    hang_doc.format(),
                ]
            }
            Statement::Builtin(n) => docs![
                arena,
                self.format_comments(&n.base.comments),
                docs![
                    arena,
                    docs![
                        arena,
                        "builtin",
                        arena.line(),
                        self.format_identifier(&n.id)
                    ]
                    .group(),
                    if n.colon.is_empty() {
                        arena.text(" ")
                    } else {
                        arena.line()
                    },
                    self.format_comments(&n.colon),
                    ": ",
                    self.format_type_expression(&n.ty),
                ]
                .nest(INDENT)
                .group()
            ],
        }
        .group()
    }

    fn format_record_expression_as_function_argument(
        &mut self,
        n: &'doc ast::ObjectExpr,
    ) -> (Doc<'doc>, Doc<'doc>, Doc<'doc>) {
        self.format_record_expression_braces(n, false)
    }

    fn format_record_expression_braces(
        &mut self,
        n: &'doc ast::ObjectExpr,
        braces: bool,
    ) -> (Doc<'doc>, Doc<'doc>, Doc<'doc>) {
        let arena = self.arena;
        let multiline = n.properties.len() > MULTILINE;
        let line = self.multiline(multiline);
        let line_ = self.multiline_(multiline);

        let first = docs![
            arena,
            self.format_comments(&n.lbrace),
            if braces { arena.text("{") } else { arena.nil() },
        ];
        let doc = docs![
            arena,
            if let Some(with) = &n.with {
                docs![
                    arena,
                    self.format_identifier(&with.source),
                    self.format_comments(&with.with),
                    if with.with.is_empty() {
                        arena.text(" ")
                    } else {
                        arena.nil()
                    },
                    "with",
                    line.clone(),
                ]
                .group()
            } else {
                line_.clone()
            },
            comma_list_with(
                arena,
                n.properties.iter().map(|property| {
                    docs![
                        arena,
                        self.format_property(property),
                        self.format_append_comments(&property.comma),
                    ]
                }),
                line,
            ),
            self.format_append_comments(&n.rbrace),
        ];
        (
            first,
            doc,
            docs![
                arena,
                line_,
                if braces { arena.text("}") } else { arena.nil() },
            ],
        )
    }

    // format_child_with_parens applies the generic rule for parenthesis (not for binary expressions).
    fn format_child_with_parens(
        &mut self,
        parent: Node<'doc>,
        child: ChildNode<'doc>,
    ) -> Doc<'doc> {
        self.format_left_child_with_parens(parent, child)
    }

    // format_right_child_with_parens applies the generic rule for parenthesis to the right child of a binary expression.
    fn format_right_child_with_parens(
        &mut self,
        parent: Node<'doc>,
        child: ChildNode<'doc>,
    ) -> Doc<'doc> {
        let (pvp, pvc) = get_precedences(&parent, &child.as_node());
        if needs_parenthesis(pvp, pvc, true) {
            self.format_node_with_parens(child)
        } else {
            self.format_childnode(child)
        }
    }

    // format_left_child_with_parens applies the generic rule for parenthesis to the left child of a binary expression.
    fn format_left_child_with_parens(
        &mut self,
        parent: Node<'doc>,
        child: ChildNode<'doc>,
    ) -> Doc<'doc> {
        let (pvp, pvc) = get_precedences(&parent, &child.as_node());
        if needs_parenthesis(pvp, pvc, false) {
            self.format_node_with_parens(child)
        } else {
            self.format_childnode(child)
        }
    }

    // XXX: rockstar (17 Jun 2021) - This clippy lint erroneously flags this
    // function with lint. It's allowed here, for now.
    // See https://github.com/rust-lang/rust-clippy/issues/7369
    #[allow(clippy::branches_sharing_code)]
    fn format_node_with_parens(&mut self, node: ChildNode<'doc>) -> Doc<'doc> {
        let arena = self.arena;
        if has_parens(&node.as_node()) {
            // If the AST already has parens here do not double add them
            self.format_childnode(node)
        } else {
            docs![arena, "(", self.format_childnode(node), ")"]
        }
    }

    fn format_childnode(&mut self, node: ChildNode<'doc>) -> Doc<'doc> {
        match node {
            ChildNode::Call(c) => self.format_call_expression(c),
            ChildNode::Expr(e) => self.format_expression(e),
        }
    }

    fn format_function_expression(&mut self, n: &'doc ast::FunctionExpr) -> HangDoc<'doc> {
        let arena = self.arena;

        let multiline = n.params.len() > MULTILINE;
        let line = self.multiline(multiline);
        let line_ = self.multiline_(multiline);

        let lparen_comments = self.format_comments(&n.lparen);

        let args = docs![
            arena,
            "(",
            docs![
                arena,
                line_.clone(),
                comma_list_with(
                    arena,
                    n.params.iter().map(|property| {
                        docs![
                            arena,
                            // treat properties differently than in general case
                            self.format_function_argument(property),
                            self.format_comments(&property.comma),
                        ]
                    }),
                    line
                ),
            ]
            .nest(INDENT),
            line_,
            self.format_comments(&n.rparen),
            ")",
            if n.arrow.is_empty() {
                arena.softline()
            } else {
                arena.nil()
            },
            self.format_append_comments(&n.arrow),
            "=>",
        ]
        .group();

        // must wrap body with parenthesis in order to discriminate between:
        //  - returning a record: (x) => ({foo: x})
        //  - and block statements:
        //		(x) => {
        //			return x + 1
        //		}
        match &n.body {
            ast::FunctionBody::Expr(b) => {
                // Remove any parentheses around the body, we will re add them if needed.
                let b = strip_parens(b);
                HangDoc {
                    affixes: vec![
                        affixes(args, arena.nil()),
                        affixes(lparen_comments, arena.nil()).nest(),
                    ],
                    body: docs![
                        arena,
                        arena.line(),
                        match b {
                            ast::Expression::Object(_) => {
                                // Add parens because we have an object literal for the body
                                docs![arena, "(", self.format_expression(b), ")",].group()
                            }
                            _ => {
                                // Do not add parens for everything else
                                self.format_expression(b)
                            }
                        },
                    ],
                }
            }
            ast::FunctionBody::Block(b) => {
                let mut hang_doc = self.format_block(b);
                hang_doc.add_prefix(arena.line());
                hang_doc.affixes.push(affixes(args, arena.nil()));
                hang_doc
                    .affixes
                    .push(affixes(lparen_comments, arena.nil()).nest());
                hang_doc
            }
        }
    }

    fn format_property(&mut self, n: &'doc ast::Property) -> Doc<'doc> {
        let arena = self.arena;
        if let Some(v) = &n.value {
            let prefix = docs![
                arena,
                self.format_property_key(&n.key),
                self.format_append_comments(&n.separator),
                ":",
            ];
            let mut hang_doc = self.hang_expression(v);
            hang_doc.add_prefix(arena.line());
            hang_doc.affixes.push(affixes(prefix, arena.nil()).nest());
            hang_doc.format()
        } else {
            self.format_property_key(&n.key)
        }
    }

    fn format_function_argument(&mut self, n: &'doc ast::Property) -> Doc<'doc> {
        let arena = self.arena;
        if let Some(v) = &n.value {
            let prefix = docs![
                arena,
                self.format_property_key(&n.key),
                self.format_comments(&n.separator),
                "=",
            ];

            let mut hang_doc = self.hang_expression(v);
            hang_doc.affixes.push(affixes(prefix, arena.nil()).nest());
            hang_doc.format()
        } else {
            self.format_property_key(&n.key)
        }
    }

    fn format_property_key(&mut self, n: &'doc ast::PropertyKey) -> Doc<'doc> {
        match n {
            ast::PropertyKey::StringLit(m) => self.format_string_literal(m),
            ast::PropertyKey::Identifier(m) => self.format_identifier(m),
        }
    }

    fn format_string_literal(&mut self, n: &'doc ast::StringLit) -> Doc<'doc> {
        let hang_doc = self.hang_string_literal(n);
        hang_doc.format()
    }

    fn hang_string_literal(&mut self, n: &'doc ast::StringLit) -> HangDoc<'doc> {
        let arena = self.arena;

        let doc = self.format_comments(&n.base.comments);

        if let Some(src) = &n.base.location.source {
            if !src.is_empty() {
                // Preserve the exact literal if we have it
                return HangDoc {
                    affixes: Vec::new(),
                    body: docs![arena, doc, src],
                };
            }
        }

        let escaped_string = escape_string(&n.value);
        HangDoc {
            affixes: vec![affixes(docs![arena, doc, arena.text("\"")], arena.text("\"")).nest()],
            body: docs![
                arena,
                // Write out escaped string value
                escaped_string,
            ],
        }
    }

    fn format_identifier(&mut self, id: &'doc ast::Identifier) -> Doc<'doc> {
        let (x, y) = self.format_split_identifier(id);
        x.append(y)
    }

    fn format_split_identifier(&mut self, id: &'doc ast::Identifier) -> (Doc<'doc>, Doc<'doc>) {
        (
            self.format_comments(&id.base.comments),
            docs![self.arena, &id.name,],
        )
    }

    fn format_variable_assignment(&mut self, n: &'doc ast::VariableAssgn) -> Doc<'doc> {
        let arena = self.arena;
        let (comment, id) = self.format_split_identifier(&n.id);
        let prefix = docs![
            arena,
            id,
            self.format_append_comments(&n.base.comments),
            " =",
        ];
        let mut hang_doc = self.hang_expression(&n.init);
        hang_doc.add_prefix(arena.line());
        hang_doc.affixes.push(affixes(prefix, arena.nil()).nest());
        docs![arena, comment, hang_doc.format()]
    }

    fn format_date_time_literal(&mut self, n: &'doc ast::DateTimeLit) -> Doc<'doc> {
        // rust rfc3339NANO only support nano3, nano6, nano9 precisions
        // for frac nano6 timestamp in go like "2018-05-22T19:53:23.09012Z",
        // rust will append a zero at the end, like "2018-05-22T19:53:23.090120Z"
        // the following implementation will match go's rfc3339nano
        let mut f: String;
        let v = &n.value;
        let nano_sec = v.timestamp_subsec_nanos();
        if nano_sec > 0 {
            f = v.format("%FT%T").to_string();
            let mut frac_nano: String = v.format("%f").to_string();
            frac_nano.insert(0, '.');
            let mut r = frac_nano.chars().last().unwrap();
            while r == '0' {
                frac_nano.pop();
                r = frac_nano.chars().last().unwrap();
            }
            f.push_str(&frac_nano);

            if v.timezone().local_minus_utc() == 0 {
                f.push('Z')
            } else {
                f.push_str(&v.format("%:z").to_string());
            }
        } else {
            f = v.to_rfc3339_opts(SecondsFormat::Secs, true)
        }

        let arena = self.arena;
        docs![arena, self.format_comments(&n.base.comments), f]
    }

    fn format_member_expression(&mut self, n: &'doc ast::MemberExpr) -> Doc<'doc> {
        let arena = self.arena;
        docs![
            arena,
            self.format_child_with_parens(Node::MemberExpr(n), ChildNode::Expr(&n.object)),
            match &n.property {
                ast::PropertyKey::Identifier(m) => {
                    docs![
                        arena,
                        self.format_append_comments(&n.lbrack),
                        ".",
                        self.format_identifier(m),
                    ]
                }
                ast::PropertyKey::StringLit(m) => {
                    docs![
                        arena,
                        self.format_comments(&n.lbrack),
                        "[",
                        self.format_string_literal(m),
                        self.format_append_comments(&n.rbrack),
                        "]",
                    ]
                }
            }
        ]
    }

    fn hang_expression(&mut self, expr: &'doc ast::Expression) -> HangDoc<'doc> {
        let arena = self.arena;
        match expr {
            ast::Expression::Array(n) => {
                let (prefix, body, suffix) = format_item_list(
                    arena,
                    ("[", "]"),
                    self.format_comments(&n.rbrack),
                    n.elements.iter().map(|item| {
                        docs![
                            arena,
                            self.format_expression(&item.expression),
                            self.format_comments(&item.comma),
                        ]
                    }),
                );
                HangDoc {
                    affixes: vec![
                        affixes(prefix, suffix),
                        affixes(self.format_comments(&n.lbrack), arena.nil()).nest(),
                    ],
                    body,
                }
            }

            ast::Expression::Object(expr) => {
                let (prefix, body, suffix) = self.format_record_expression_braces(expr, true);
                HangDoc {
                    affixes: vec![affixes(prefix, suffix).nest()],
                    body,
                }
            }

            ast::Expression::StringExpr(n) => HangDoc {
                affixes: vec![affixes(
                    docs![arena, self.format_comments(&n.base.comments), "\""],
                    arena.text("\""),
                )
                .nest()],
                body: docs![
                    arena,
                    arena.concat(n.parts.iter().map(|n| {
                        match n {
                            ast::StringExprPart::Text(p) => self.format_text_part(p),
                            ast::StringExprPart::Interpolated(p) => {
                                self.format_interpolated_part(p)
                            }
                        }
                    })),
                ],
            },

            ast::Expression::Function(expr) => self.format_function_expression(expr),

            ast::Expression::StringLit(expr) => self.hang_string_literal(expr),

            ast::Expression::Index(n) => HangDoc {
                affixes: vec![affixes(
                    docs![
                        arena,
                        self.format_child_with_parens(
                            Node::IndexExpr(n),
                            ChildNode::Expr(&n.array)
                        ),
                        self.format_comments(&n.lbrack),
                        "[",
                    ],
                    docs![arena, self.format_comments(&n.rbrack), "]",],
                )
                .nest()],
                body: self.format_expression(&n.index),
            },

            _ => HangDoc {
                affixes: Vec::new(),
                body: self.format_expression(expr),
            },
        }
    }

    fn format_expression(&mut self, expr: &'doc ast::Expression) -> Doc<'doc> {
        let arena = self.arena;
        let parent = expr;
        match expr {
            ast::Expression::Array(_)
            | ast::Expression::Object(_)
            | ast::Expression::StringExpr(_)
            | ast::Expression::Function(_)
            | ast::Expression::StringLit(_)
            | ast::Expression::Index(_) => {
                let hang_doc = self.hang_expression(expr);
                hang_doc.format()
            }
            ast::Expression::Identifier(expr) => self.format_identifier(expr),
            ast::Expression::Dict(n) => {
                let line = self.base_multiline(&n.base);
                docs![
                    arena,
                    self.format_comments(&n.lbrack),
                    "[",
                    docs![
                        arena,
                        arena.line_(),
                        if n.elements.is_empty() {
                            arena.text(":")
                        } else {
                            comma_list_with(
                                arena,
                                n.elements.iter().map(|item| {
                                    docs![
                                        arena,
                                        self.format_expression(&item.key),
                                        ":",
                                        " ",
                                        self.format_expression(&item.val),
                                        self.format_comments(&item.comma),
                                    ]
                                }),
                                line,
                            )
                        },
                        self.format_comments(&n.rbrack),
                    ]
                    .nest(INDENT),
                    arena.line_(),
                    "]",
                ]
            }
            ast::Expression::Logical(expr) => self.format_binary_expression(
                parent,
                &expr.left,
                expr.operator.as_str(),
                &expr.right,
            ),
            ast::Expression::Member(n) => self.format_member_expression(n),
            ast::Expression::Binary(expr) => self.format_binary_expression(
                parent,
                &expr.left,
                expr.operator.as_str(),
                &expr.right,
            ),
            ast::Expression::Unary(n) => {
                docs![
                    arena,
                    self.format_comments(&n.base.comments),
                    n.operator.to_string(),
                    match n.operator {
                        ast::Operator::SubtractionOperator => arena.nil(),
                        ast::Operator::AdditionOperator => arena.nil(),
                        _ => {
                            arena.text(" ")
                        }
                    },
                    self.format_child_with_parens(Node::UnaryExpr(n), ChildNode::Expr(&n.argument)),
                ]
            }
            ast::Expression::PipeExpr(n) => self.format_pipe_expression(n),
            ast::Expression::Call(expr) => self.format_call_expression(expr),
            ast::Expression::Conditional(n) => {
                let line = self.base_multiline(&n.base);
                let mut alternate = &n.alternate;
                let mut doc = docs![
                    arena,
                    docs![
                        arena,
                        "if ",
                        self.format_expression(&n.test).nest(INDENT),
                        self.format_comments(&n.tk_then),
                        if n.tk_then.is_empty() {
                            arena.line()
                        } else {
                            arena.nil()
                        },
                        "then",
                    ]
                    .group(),
                    docs![arena, line.clone(), self.format_expression(&n.consequent)].nest(INDENT),
                    line.clone(),
                    self.format_comments(&n.tk_else),
                ];
                loop {
                    match alternate {
                        ast::Expression::Conditional(n) => {
                            doc = docs![
                                arena,
                                doc,
                                self.format_comments(&n.tk_if),
                                "else if ",
                                self.format_expression(&n.test).nest(INDENT),
                                self.format_comments(&n.tk_then),
                                " then",
                                docs![arena, line.clone(), self.format_expression(&n.consequent)]
                                    .nest(INDENT),
                                line.clone(),
                                self.format_comments(&n.tk_else),
                            ];
                            alternate = &n.alternate;
                        }
                        _ => {
                            doc = docs![
                                arena,
                                doc,
                                "else",
                                docs![arena, line, self.format_expression(alternate)].nest(INDENT),
                            ];
                            break;
                        }
                    }
                }
                docs![arena, self.format_comments(&n.tk_if), doc.group()]
            }
            ast::Expression::Integer(expr) => {
                docs![
                    arena,
                    self.format_comments(&expr.base.comments),
                    format!("{}", expr.value),
                ]
            }
            ast::Expression::Float(expr) => {
                docs![arena, self.format_comments(&expr.base.comments), {
                    let mut s = format!("{}", expr.value);
                    if !s.contains('.') {
                        s.push_str(".0");
                    }
                    s
                }]
            }
            ast::Expression::Duration(n) => {
                docs![
                    arena,
                    self.format_comments(&n.base.comments),
                    arena.concat(
                        n.values
                            .iter()
                            .map(|d| { docs![arena, format!("{}", d.magnitude), &d.unit,] })
                    )
                ]
            }
            ast::Expression::Uint(n) => {
                docs![
                    arena,
                    self.format_comments(&n.base.comments),
                    format!("{0:10}", n.value),
                ]
            }
            ast::Expression::Boolean(expr) => {
                let s = if expr.value { "true" } else { "false" };
                arena.text(s)
            }
            ast::Expression::DateTime(expr) => self.format_date_time_literal(expr),
            ast::Expression::Regexp(expr) => self.format_regexp_literal(expr),
            ast::Expression::PipeLit(expr) => {
                docs![arena, self.format_comments(&expr.base.comments), "<-"]
            }
            ast::Expression::Bad(expr) => {
                self.err = Some(anyhow!("bad expression"));
                arena.nil()
            }
            ast::Expression::Paren(n) => {
                if has_parens(&Node::ParenExpr(n)) {
                    docs![
                        arena,
                        // The paren node has comments so we should format them
                        self.format_comments(&n.lparen),
                        "(",
                        self.format_expression(&n.expression),
                        self.format_append_comments(&n.rparen),
                        ")",
                    ]
                } else {
                    // The paren node does not have comments so we can skip adding the parens
                    self.format_expression(&n.expression)
                }
            }
        }
        .group()
    }

    fn format_text_part(&mut self, n: &'doc ast::TextPart) -> Doc<'doc> {
        let arena = self.arena;

        arena.intersperse(
            n.value.split('\n').map(|s| {
                let escaped_string = escape_string(s);
                arena.text(escaped_string)
            }),
            arena.nesting(move |indentation| {
                arena.hardline().nest(-(indentation as isize)).into_doc()
            }),
        )
    }

    fn format_interpolated_part(&mut self, n: &'doc ast::InterpolatedPart) -> Doc<'doc> {
        let arena = self.arena;
        docs![arena, "${", self.format_expression(&n.expression), "}",]
    }

    fn layout_binary_expressions(
        &mut self,
        mut arguments: impl Iterator<Item = Doc<'doc>>,
        mut operators: impl Iterator<Item = Doc<'doc>>,
        line: Doc<'doc>,
    ) -> Doc<'doc> {
        let arena = self.arena;
        let first = match arguments.next() {
            Some(doc) => doc,
            None => return arena.nil(),
        };
        let mut doc = line.clone();
        let mut arguments = arguments.peekable();
        loop {
            match (operators.next(), arguments.next()) {
                (Some(operator), Some(arg)) => {
                    doc += docs![arena, operator, arg].group();
                    if arguments.peek().is_some() {
                        doc += line.clone();
                    }
                }
                _ => return docs![arena, first, doc.nest(INDENT)].group(),
            }
        }
    }

    fn format_pipe_expression(&mut self, mut pipe: &'doc ast::PipeExpr) -> Doc<'doc> {
        let arena = self.arena;

        let mut arguments = Vec::new();
        let mut operators = Vec::new();
        let line = self.base_multiline(&pipe.base);
        loop {
            arguments.push(
                self.format_right_child_with_parens(
                    Node::PipeExpr(pipe),
                    ChildNode::Call(&pipe.call),
                )
                .group(),
            );
            operators.push(docs![
                arena,
                self.format_comments(&pipe.base.comments),
                arena.text("|> "),
            ]);
            match &pipe.argument {
                ast::Expression::PipeExpr(expr) => {
                    pipe = expr;
                }
                _ => {
                    arguments.push(
                        self.format_left_child_with_parens(
                            Node::PipeExpr(pipe),
                            ChildNode::Expr(&pipe.argument),
                        )
                        .group(),
                    );
                    break;
                }
            }
        }
        self.layout_binary_expressions(
            arguments.into_iter().rev(),
            operators.into_iter().rev(),
            line,
        )
    }

    fn format_call_expression(&mut self, n: &'doc ast::CallExpr) -> Doc<'doc> {
        let arena = self.arena;
        let line = self.base_multiline(&n.base);
        let line_ = self.base_multiline_(&n.base);
        docs![
            arena,
            self.format_child_with_parens(Node::CallExpr(n), ChildNode::Expr(&n.callee)),
            self.format_append_comments(&n.lparen),
            match n.arguments.first() {
                Some(ast::Expression::Object(o)) if n.arguments.len() == 1 => {
                    let (prefix, body, suffix) =
                        self.format_record_expression_as_function_argument(o);

                    docs![
                        arena,
                        "(",
                        format_hang_doc(arena, &[affixes(prefix, suffix).nest()], body),
                        self.format_comments(&n.rparen),
                        ")"
                    ]
                }
                _ => {
                    let (prefix, body, suffix) = format_item_list(
                        arena,
                        ("(", ")"),
                        self.format_comments(&n.rparen),
                        n.arguments.iter().map(|c| self.format_expression(c)),
                    );
                    format_hang_doc(arena, &[affixes(prefix, suffix).nest()], body)
                }
            },
        ]
        .group()
    }

    fn format_binary_expression(
        &mut self,
        mut parent: &'doc ast::Expression,
        lhs: &'doc ast::Expression,
        mut operator: &'doc str,
        mut rhs: &'doc ast::Expression,
    ) -> Doc<'doc> {
        let arena = self.arena;
        let l = self.format_left_child_with_parens(Node::from_expr(parent), ChildNode::Expr(lhs));
        let mut doc = arena.nil();
        loop {
            match rhs {
                ast::Expression::Binary(expr) => {
                    doc = docs![
                        arena,
                        doc,
                        docs![
                            arena,
                            arena.line(),
                            self.format_comments(&parent.base().comments),
                            operator,
                            arena.line(),
                        ]
                        .group(),
                        self.format_left_child_with_parens(
                            Node::BinaryExpr(expr),
                            ChildNode::Expr(&expr.left)
                        ),
                    ];

                    parent = rhs;
                    operator = expr.operator.as_str();
                    rhs = &expr.right;
                }
                _ => {
                    doc = docs![
                        arena,
                        doc,
                        docs![
                            arena,
                            arena.line(),
                            self.format_comments(&parent.base().comments),
                            operator,
                            arena.line(),
                        ]
                        .group(),
                        self.format_right_child_with_parens(
                            Node::from_expr(parent),
                            ChildNode::Expr(rhs)
                        ),
                    ];
                    break;
                }
            }
        }
        docs![arena, l, doc.nest(INDENT)].group()
    }

    fn format_regexp_literal(&mut self, n: &'doc ast::RegexpLit) -> Doc<'doc> {
        let arena = self.arena;
        docs![
            arena,
            self.format_comments(&n.base.comments),
            "/",
            n.value.replace('/', "\\/"),
            "/",
        ]
    }
}

fn escape_string(s: &str) -> String {
    if !(s.contains('\"') || s.contains('\\')) {
        return s.to_string();
    }
    let mut escaped = String::with_capacity(s.len() * 2);
    for r in s.chars() {
        if r == '"' || r == '\\' {
            escaped.push('\\')
        }
        escaped.push(r)
    }
    escaped
}

enum ChildNode<'doc> {
    Call(&'doc ast::CallExpr),
    Expr(&'doc ast::Expression),
}

impl<'doc> ChildNode<'doc> {
    fn as_node(&self) -> Node<'doc> {
        match *self {
            Self::Call(c) => Node::CallExpr(c),
            Self::Expr(c) => Node::from_expr(c),
        }
    }
}

// INDENT_BYTES is 4 spaces as a constant byte slice
const INDENT_BYTES: &str = "    ";
const INDENT: isize = INDENT_BYTES.len() as isize;

fn get_precedences(parent: &Node, child: &Node) -> (u32, u32) {
    let pvp: u32 = match parent {
        Node::BinaryExpr(p) => Operator::new(&p.operator).get_precedence(),
        Node::LogicalExpr(p) => Operator::new_logical(&p.operator).get_precedence(),
        Node::UnaryExpr(p) => Operator::new(&p.operator).get_precedence(),
        Node::FunctionExpr(_) => 3,
        Node::PipeExpr(_) => 2,
        Node::CallExpr(_) => 1,
        Node::MemberExpr(_) => 1,
        Node::IndexExpr(_) => 1,
        Node::ParenExpr(p) => return get_precedences(&(Node::from_expr(&p.expression)), child),
        Node::ConditionalExpr(_) => 11,
        _ => 0,
    };

    let pvc: u32 = match child {
        Node::BinaryExpr(p) => Operator::new(&p.operator).get_precedence(),
        Node::LogicalExpr(p) => Operator::new_logical(&p.operator).get_precedence(),
        Node::UnaryExpr(p) => Operator::new(&p.operator).get_precedence(),
        Node::FunctionExpr(_) => 3,
        Node::PipeExpr(_) => 2,
        Node::CallExpr(_) => 1,
        Node::MemberExpr(_) => 1,
        Node::IndexExpr(_) => 1,
        Node::ParenExpr(p) => return get_precedences(parent, &(Node::from_expr(&p.expression))),
        Node::ConditionalExpr(_) => 11,
        _ => 0,
    };

    (pvp, pvc)
}

struct Operator<'a> {
    op: Option<&'a ast::Operator>,
    l_op: Option<&'a ast::LogicalOperator>,
    is_logical: bool,
}

impl<'a> Operator<'a> {
    fn new(op: &ast::Operator) -> Operator {
        Operator {
            op: Some(op),
            l_op: None,
            is_logical: false,
        }
    }

    fn new_logical(op: &ast::LogicalOperator) -> Operator {
        Operator {
            op: None,
            l_op: Some(op),
            is_logical: true,
        }
    }

    fn get_precedence(&self) -> u32 {
        if !self.is_logical {
            return match self.op.unwrap() {
                ast::Operator::PowerOperator => 4,
                ast::Operator::MultiplicationOperator => 5,
                ast::Operator::DivisionOperator => 5,
                ast::Operator::ModuloOperator => 5,
                ast::Operator::AdditionOperator => 6,
                ast::Operator::SubtractionOperator => 6,
                ast::Operator::LessThanEqualOperator => 7,
                ast::Operator::LessThanOperator => 7,
                ast::Operator::GreaterThanEqualOperator => 7,
                ast::Operator::GreaterThanOperator => 7,
                ast::Operator::StartsWithOperator => 7,
                ast::Operator::InOperator => 7,
                ast::Operator::NotEmptyOperator => 7,
                ast::Operator::EmptyOperator => 7,
                ast::Operator::EqualOperator => 7,
                ast::Operator::NotEqualOperator => 7,
                ast::Operator::RegexpMatchOperator => 7,
                ast::Operator::NotRegexpMatchOperator => 7,
                ast::Operator::NotOperator => 8,
                ast::Operator::ExistsOperator => 8,
                ast::Operator::InvalidOperator => 0,
            };
        }
        match self.l_op.unwrap() {
            ast::LogicalOperator::AndOperator => 9,
            ast::LogicalOperator::OrOperator => 10,
        }
    }
}

// About parenthesis:
// We need parenthesis if a child node has lower precedence (bigger value) than its parent node.
// The same stands for the left child of a binary expression; while, for the right child, we need parenthesis if its
// precedence is lower or equal then its parent's.
//
// To explain parenthesis logic, we must to understand how the parser generates the AST.
// (A) - The parser always puts lower precedence operators at the root of the AST.
// (B) - When there are multiple operators with the same precedence, the right-most expression is at root.
// (C) - When there are parenthesis, instead, the parser recursively generates a AST for the expression contained
// in the parenthesis, and makes it the right child.
// So, when formatting:
//  - if we encounter a child with lower precedence on the left, this means it requires parenthesis, because, for sure,
//    the parser detected parenthesis to break (A);
//  - if we encounter a child with higher or equal precedence on the left, it doesn't need parenthesis, because
//    that was the natural parsing order of elements (see (B));
//  - if we encounter a child with lower or equal precedence on the right, it requires parenthesis, otherwise, it
//    would have been at root (see (C)).
fn needs_parenthesis(pvp: u32, pvc: u32, is_right: bool) -> bool {
    // If one of the precedence values is invalid, then we shouldn't apply any parenthesis.
    let par = pvc != 0 && pvp != 0;
    par && ((!is_right && pvc > pvp) || (is_right && pvc >= pvp))
}

// has_parens reports whether the node will be formatted with parens.
//
// Only format parens if they have associated comments.
// Otherwise we skip formatting them because anytime they are needed they are explicitly
// added back in.
fn has_parens(n: &Node) -> bool {
    if let Node::ParenExpr(p) = &n {
        return !p.lparen.is_empty() || !p.rparen.is_empty();
    }
    false
}

// strip_parens returns the expression removing any wrapping paren expressions
// that do not have comments attached
fn strip_parens(n: &ast::Expression) -> &ast::Expression {
    if let ast::Expression::Paren(p) = n {
        if p.lparen.is_empty() && p.rparen.is_empty() {
            return strip_parens(&p.expression);
        }
    }
    n
}

// starts_with_comment reports if the node has a comment that it would format before anything else as part
// of the node.
fn starts_with_comment(n: Node) -> bool {
    match n {
        Node::Package(n) => !n.base.comments.is_empty(),
        Node::File(n) => {
            if let Some(pkg) = &n.package {
                return starts_with_comment(Node::PackageClause(pkg));
            }
            if let Some(imp) = &n.imports.first() {
                return starts_with_comment(Node::ImportDeclaration(imp));
            }
            if let Some(stmt) = &n.body.first() {
                return starts_with_comment(Node::from_stmt(stmt));
            }
            !n.eof.is_empty()
        }
        Node::PackageClause(n) => !n.base.comments.is_empty(),
        Node::ImportDeclaration(n) => !n.base.comments.is_empty(),
        Node::Identifier(n) => !n.base.comments.is_empty(),
        Node::ArrayExpr(n) => !n.lbrack.is_empty(),
        Node::DictExpr(n) => !n.lbrack.is_empty(),
        Node::FunctionExpr(n) => !n.lparen.is_empty(),
        Node::LogicalExpr(n) => starts_with_comment(Node::from_expr(&n.left)),
        Node::ObjectExpr(n) => !n.lbrace.is_empty(),
        Node::MemberExpr(n) => starts_with_comment(Node::from_expr(&n.object)),
        Node::IndexExpr(n) => starts_with_comment(Node::from_expr(&n.array)),
        Node::BinaryExpr(n) => starts_with_comment(Node::from_expr(&n.left)),
        Node::UnaryExpr(n) => !n.base.comments.is_empty(),
        Node::PipeExpr(n) => starts_with_comment(Node::from_expr(&n.argument)),
        Node::CallExpr(n) => starts_with_comment(Node::from_expr(&n.callee)),
        Node::ConditionalExpr(n) => !n.tk_if.is_empty(),
        Node::StringExpr(n) => !n.base.comments.is_empty(),
        Node::ParenExpr(n) => !n.lparen.is_empty(),
        Node::IntegerLit(n) => !n.base.comments.is_empty(),
        Node::FloatLit(n) => !n.base.comments.is_empty(),
        Node::StringLit(n) => !n.base.comments.is_empty(),
        Node::DurationLit(n) => !n.base.comments.is_empty(),
        Node::UintLit(n) => !n.base.comments.is_empty(),
        Node::BooleanLit(n) => !n.base.comments.is_empty(),
        Node::DateTimeLit(n) => !n.base.comments.is_empty(),
        Node::RegexpLit(n) => !n.base.comments.is_empty(),
        Node::PipeLit(n) => !n.base.comments.is_empty(),
        Node::BadExpr(_) => false,
        Node::ExprStmt(n) => starts_with_comment(Node::from_expr(&n.expression)),
        Node::OptionStmt(n) => !n.base.comments.is_empty(),
        Node::ReturnStmt(n) => !n.base.comments.is_empty(),
        Node::BadStmt(_) => false,
        Node::TestStmt(n) => !n.base.comments.is_empty(),
        Node::TestCaseStmt(n) => !n.base.comments.is_empty(),
        Node::BuiltinStmt(n) => !n.base.comments.is_empty(),
        Node::Block(n) => !n.lbrace.is_empty(),
        Node::Property(_) => false,
        Node::TextPart(_) => false,
        Node::InterpolatedPart(_) => false,
        Node::VariableAssgn(n) => starts_with_comment(Node::Identifier(&n.id)),
        Node::MemberAssgn(n) => starts_with_comment(Node::MemberExpr(&n.member)),
        Node::TypeExpression(n) => !n.base.comments.is_empty(),
        Node::MonoType(n) => !n.base().comments.is_empty(),
        Node::ParameterType(n) => !n.base().comments.is_empty(),
        Node::PropertyType(n) => !n.base.comments.is_empty(),
        Node::TypeConstraint(n) => !n.base.comments.is_empty(),
    }
}

#[cfg(test)]
pub mod tests;
