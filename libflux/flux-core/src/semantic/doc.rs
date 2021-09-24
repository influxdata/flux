//! Generate documentation from source code comments.

use pulldown_cmark::{Event, OffsetIter, Parser as MarkdownParser, Tag};
use std::collections::BTreeMap;
use std::ops::Range;

use crate::ast;
use crate::semantic::types::{Function, MonoType, PolyType, PolyTypeMap};
use derive_more::Display;

/// Diagnostic represents an issue with the documentation comments.
/// Something about the formatting or content of the comments does not meet expectations.
#[derive(PartialEq, Debug, Display)]
#[display(fmt = "error {}: {}", loc, msg)]
pub struct Diagnostic {
    msg: String,
    loc: ast::SourceLocation,
}

/// Diagnostics is a set of diagnostics
pub type Diagnostics = Vec<Diagnostic>;

/// Doc is an enum that can take the form of the various types of flux documentation structures through polymorphism.
#[derive(PartialEq, Debug, Serialize, Deserialize)]
#[serde(tag = "kind")]
pub enum Doc {
    /// Package represents documentation for an entire Flux package.
    Package(Box<PackageDoc>),
    /// Value represents documentation for a value exposed from a package.
    Value(Box<ValueDoc>),
    /// Option represents documentation for a option value exposed from a package.
    Opt(Box<ValueDoc>),
    /// Function represents documentation for a function value exposed from a package.
    Function(Box<FunctionDoc>),
}

/// PackageDoc represents the documentation for a package and its sub packages
#[derive(PartialEq, Debug, Serialize, Deserialize)]
pub struct PackageDoc {
    /// the relative path to the package
    pub path: String,
    /// the name of the comments package
    pub name: String,
    /// the headline of the package
    pub headline: String,
    /// the description of the package
    pub description: Option<String>,
    /// the members are the values and functions of a package
    pub members: BTreeMap<String, Doc>,
}

/// ValueDoc represents the documentation for a single value within a package.
/// Values include options, builtins, or any variable assignment within the top level scope of a
/// package.
#[derive(PartialEq, Debug, Serialize, Deserialize)]
pub struct ValueDoc {
    /// the name of the value
    pub name: String,
    /// the headline of the value
    pub headline: String,
    /// the description of the value
    pub description: Option<String>,
    /// the type of the value
    pub flux_type: String,
}

/// FunctionDoc represents the documentation for a single Function within a package.
#[derive(PartialEq, Debug, Serialize, Deserialize)]
pub struct FunctionDoc {
    /// the name of the function
    pub name: String,
    /// the headline of the function
    pub headline: String,
    /// the description of the function
    pub description: String,
    /// the parameters of the function
    pub parameters: Vec<ParameterDoc>,
    /// the type of the function
    pub flux_type: String,
}

/// ParameterDoc represents the documentation for a single parameter within a function.
#[derive(PartialEq, Debug, Serialize, Deserialize)]
pub struct ParameterDoc {
    /// the name of the parameter
    pub name: String,
    /// the headline of the parameter
    pub headline: String,
    /// the description of the parameter
    pub description: Option<String>,
    /// a boolean indicating if the parameter is required
    pub required: bool,
}

/// Error when generating documentation.
#[derive(Debug, Display)]
#[display(fmt = "error: {}", msg)]
pub struct Error {
    /// The error message
    pub msg: String,
}

impl std::error::Error for Error {}
// Private errors that can occur when parsing markdown.
enum ParseError {
    EmptyComment,
}

/// Parse the package documentation for all values within the package.
/// The list of diagnostics reports problems found with formatting or otherwise of the comments.
/// An empty list of diagnostics implies that doc comments are all property formatted.
pub fn parse_doc_comments(
    pkg: &ast::Package,
    pkgpath: &str,
    types: &PolyTypeMap,
) -> Result<(PackageDoc, Diagnostics), Error> {
    // TODO(nathanielc): Support package with more than one file.
    parse_file_doc_comments(&pkg.files[0], pkgpath, types)
}

/// Parse the package documentation for all values within the package.
/// The list of diagnostics reports problems found with formatting or otherwise of the comments.
/// An empty list of diagnostics implies that doc comments are all property formatted.
pub fn parse_file_doc_comments(
    file: &ast::File,
    pkgpath: &str,
    types: &PolyTypeMap,
) -> Result<(PackageDoc, Diagnostics), Error> {
    let mut diagnostics: Diagnostics = Vec::new();
    let (name, headline, description) = match &file.package {
        Some(pkg_clause) => {
            let comment = comments_to_string(&pkg_clause.base.comments);
            let (headline, description) = match parse_headline_desc(&comment) {
                Ok(hd) => (hd.headline, hd.description),
                Err(ParseError::EmptyComment) => {
                    diagnostics.push(Diagnostic {
                        msg: format!(
                            "package {} must contain a non empty package comment",
                            pkgpath
                        ),
                        loc: pkg_clause.base.location.clone(),
                    });
                    ("".to_string(), None)
                }
            };
            (pkg_clause.name.name.clone(), headline, description)
        }
        None => {
            diagnostics.push(Diagnostic {
                msg: format!("package {} must contain a package clause", pkgpath),
                loc: file.base.location.clone(),
            });
            ("".to_string(), "".to_string(), None)
        }
    };
    let (members, mut diags) = parse_package_values(file, types)?;
    diagnostics.append(&mut diags);
    Ok((
        PackageDoc {
            path: pkgpath.to_string(),
            name,
            headline,
            description,
            members,
        },
        diagnostics,
    ))
}

// Generates docs for the values in a given source file.
fn parse_package_values(
    f: &ast::File,
    pkgtypes: &PolyTypeMap,
) -> Result<(BTreeMap<String, Doc>, Diagnostics), Error> {
    let mut members: BTreeMap<String, Doc> = BTreeMap::new();
    let mut diagnostics: Diagnostics = Vec::new();
    for stmt in &f.body {
        if let Some((name, comment, loc)) = match stmt {
            ast::Statement::Variable(s) => {
                let comment = comments_to_string(&s.id.base.comments);
                let name = s.id.name.clone();
                Some((name, comment, &s.base.location))
            }
            ast::Statement::Builtin(s) => {
                let comment = comments_to_string(&s.base.comments);
                let name = s.id.name.clone();
                Some((name, comment, &s.base.location))
            }
            ast::Statement::Option(s) => {
                match &s.assignment {
                    ast::Assignment::Variable(v) => {
                        let comment = comments_to_string(&s.base.comments);
                        let name = v.id.name.clone();
                        Some((name, comment, &s.base.location))
                    }
                    // Member assignments are not exported values from a package
                    // and do not need documentation.
                    _ => None,
                }
            }
            // Other statements do not assign any value and therefore are not exported from a
            // package.
            _ => None,
        } {
            let typ = &pkgtypes[name.as_str()];
            let (doc, mut diags) = parse_value(&name, &comment, typ, loc)?;
            diagnostics.append(&mut diags);
            members.insert(name.clone(), doc);
        }
    }
    Ok((members, diagnostics))
}

struct HeadlineDesc {
    headline: String,
    description: Option<String>,
}

// parses a headline and description from a comment.
// The headline is defined as the first paragraph in the comment.
fn parse_headline_desc(comment: &str) -> Result<HeadlineDesc, ParseError> {
    if comment.trim().is_empty() {
        return Err(ParseError::EmptyComment);
    }
    let mut parser = MarkdownParser::new(comment).into_offset_iter();
    let range = headline_range(&mut parser)?;
    let headline = &comment[range.clone()];
    // the rest of the comment is the description
    let description = &comment[range.end..comment.len()];
    Ok(HeadlineDesc {
        headline: headline.to_string(),
        description: match description.len() {
            0 => None,
            _ => Some(description.to_string()),
        },
    })
}

// find the range for the headline which is defined as the first paragraph.
fn headline_range(parser: &mut OffsetIter) -> Result<Range<usize>, ParseError> {
    let mut range = Range::<usize> { start: 0, end: 0 };
    loop {
        match parser.next() {
            Some((Event::Start(Tag::Paragraph), r)) => {
                range.start = r.start;
            }
            Some((Event::End(Tag::Paragraph), r)) => {
                range.end = r.end;
                return Ok(range);
            }
            Some(_) => {} //do nothing but catch the event
            None => {
                return Err(ParseError::EmptyComment);
            }
        }
    }
}

fn parse_value(
    name: &str,
    comment: &str,
    typ: &PolyType,
    loc: &ast::SourceLocation,
) -> Result<(Doc, Diagnostics), Error> {
    match &typ.expr {
        MonoType::Fun(f) => {
            let (doc, diags) = parse_function_doc(name, comment, typ, f, loc)?;
            Ok((Doc::Function(Box::new(doc)), diags))
        }
        _ => {
            let (doc, diags) = parse_value_doc(name, comment, typ, loc)?;
            Ok((Doc::Value(Box::new(doc)), diags))
        }
    }
}

fn parse_function_doc(
    name: &str,
    comment: &str,
    typ: &PolyType,
    fun_typ: &Function,
    loc: &ast::SourceLocation,
) -> Result<(FunctionDoc, Diagnostics), Error> {
    let mut diagnostics: Diagnostics = Vec::new();
    let (headline, description) = match parse_headline_desc(comment) {
        Ok(hd) => (hd.headline, hd.description),
        Err(ParseError::EmptyComment) => {
            diagnostics.push(Diagnostic {
                msg: format!("function \"{}\" must contain a non empty comment", name),
                loc: loc.clone(),
            });
            ("".to_string(), None)
        }
    };
    let mut parameters: Vec<ParameterDoc> = Vec::new();
    let description = match description {
        Some(description) => {
            let mut parser = MarkdownParser::new(&description).into_offset_iter();
            let mut parameter_range: Range<usize> = Range::default();
            loop {
                match parser.next() {
                    Some((Event::Start(Tag::Heading(2)), range)) => {
                        if let Some((Event::Text(_), _)) = parser.next() {
                            parameter_range.start = range.start;
                            let (params, range, mut diags) = parse_function_parameter_list(
                                &mut parser,
                                fun_typ,
                                &description,
                                loc,
                            )?;
                            parameter_range.end = range.end;
                            parameters = params;
                            diagnostics.append(&mut diags);
                            // Validate all parameters were documented
                            let params_on_type: Vec<&String> =
                                fun_typ.req.keys().chain(fun_typ.opt.keys()).collect();
                            for name in &params_on_type {
                                if !contains_parameter(&parameters, name.as_str()) {
                                    diagnostics.push(Diagnostic {
                                        msg: format!(
                                            "missing documentation for parameter \"{}\"",
                                            name
                                        ),
                                        loc: loc.clone(),
                                    });
                                }
                            }
                            // Validate extra parameters are not documented
                            for param in &parameters {
                                if !params_on_type.iter().any(|&name| name == &param.name) {
                                    diagnostics.push(Diagnostic {
                                        msg: format!("found extra parameter \"{}\"", name),
                                        loc: loc.clone(),
                                    });
                                }
                            }
                            break;
                        }
                    }
                    // else do nothing
                    None => break,
                    _ => {}
                }
            }
            if !parameter_range.is_empty() {
                // Return the description with the parameter list removed
                description
                    .chars()
                    .take(parameter_range.start)
                    .chain(description.chars().skip(parameter_range.end))
                    .collect()
            } else {
                // Its possible the parameter list was not found or was invalid.
                // In such cases a diagnostic would have been reported, so just return the
                // description unmodified.
                description
            }
        }
        None => {
            diagnostics.push(Diagnostic {
                msg: format!(
                    "function \"{}\" comment must contain both a headline and a description",
                    name
                ),
                loc: loc.clone(),
            });
            "".to_string()
        }
    };
    Ok((
        FunctionDoc {
            name: name.to_string(),
            headline,
            description,
            parameters,
            flux_type: format!("{}", &typ.normal()),
        },
        diagnostics,
    ))
}

fn contains_parameter(params: &[ParameterDoc], name: &str) -> bool {
    params.iter().any(|pd| pd.name == name)
}

fn parse_function_parameter_list(
    parser: &mut OffsetIter,
    typ: &Function,
    content: &str,
    loc: &ast::SourceLocation,
) -> Result<(Vec<ParameterDoc>, Range<usize>, Diagnostics), Error> {
    let mut parameters: Vec<ParameterDoc> = Vec::new();
    let mut diagnostics: Diagnostics = Vec::new();
    loop {
        match parser.next() {
            Some((Event::Start(Tag::List(_)), range)) => println!("found list start {:?}", range),
            Some((Event::Start(Tag::Item), r)) => {
                println!("item {:?}", r);
                let (doc, mut diags) = parse_function_parameter(parser, typ, content, loc)?;
                diagnostics.append(&mut diags);
                parameters.push(doc);
            }
            Some((Event::End(Tag::List(_)), range)) => return Ok((parameters, range, diagnostics)),
            None => return Ok((parameters, Range::default(), diagnostics)),
            _ => {}
        }
    }
}

fn parse_function_parameter(
    parser: &mut OffsetIter,
    typ: &Function,
    content: &str,
    loc: &ast::SourceLocation,
) -> Result<(ParameterDoc, Diagnostics), Error> {
    let mut diagnostics: Diagnostics = Vec::new();
    let mut headline_range: Range<usize> = Range::default();
    let mut name = String::new();
    // Find Code event for the parameter name. For example "`rows`".
    loop {
        match parser.next() {
            Some((Event::Start(Tag::Paragraph), _)) => {}
            Some((Event::Code(c), range)) => {
                headline_range.start = range.start;
                name = c.to_string();
                break;
            }
            Some(_) => {
                diagnostics.push(Diagnostic {
                    msg:
                        "parameter list entry does not begin with the parameter name in backticks "
                            .to_string(),
                    loc: loc.clone(),
                });
                break;
            }
            None => {
                diagnostics.push(Diagnostic {
                    msg: "parameter list entry ends without content".to_string(),
                    loc: loc.clone(),
                });
                break;
            }
        }
    }
    // Find Text event for the parameter headline. For example "is the array of records ... ".
    match parser.next() {
        Some((Event::Text(_), range)) => {
            headline_range.end = range.end;
        }
        Some(_) => {
            diagnostics.push(Diagnostic {
                msg: "parameter list entry does not contain a headline".to_string(),
                loc: loc.clone(),
            });
        }
        None => {
            diagnostics.push(Diagnostic {
                msg: "parameter list entry ends unexpectedly".to_string(),
                loc: loc.clone(),
            });
        }
    }
    let headline = content[headline_range.clone()].to_string();
    // The rest of the list item is the description.
    let mut desc_range = Range::<usize> {
        start: headline_range.end,
        end: 0,
    };
    let mut depth = 0;
    loop {
        match parser.next() {
            Some((Event::Start(Tag::List(_)), _)) => {
                depth += 1;
            }
            Some((Event::End(Tag::List(_)), _)) => {
                depth -= 1;
            }
            Some((Event::End(Tag::Item), range)) => {
                if depth == 0 {
                    desc_range.end = range.end;
                    break;
                }
            }
            // Consume all other events into the description.
            // It is valid for the description to contain arbitrary markdown.
            Some(_) => {}
            None => {
                diagnostics.push(Diagnostic {
                    msg: "parameter list entry ends unexpectedly".to_string(),
                    loc: loc.clone(),
                });
                break;
            }
        }
    }
    let description = if desc_range.is_empty() {
        None
    } else {
        let d = content[desc_range].to_string();
        if d.trim().is_empty() {
            None
        } else {
            Some(d)
        }
    };
    let required = typ.req.contains_key(&name);
    Ok((
        ParameterDoc {
            name,
            headline,
            description,
            required,
        },
        diagnostics,
    ))
}

fn parse_value_doc(
    name: &str,
    comment: &str,
    typ: &PolyType,
    loc: &ast::SourceLocation,
) -> Result<(ValueDoc, Diagnostics), Error> {
    let mut diagnostics: Diagnostics = Vec::new();
    let (headline, description) = match parse_headline_desc(comment) {
        Ok(hd) => (hd.headline, hd.description),
        Err(ParseError::EmptyComment) => {
            diagnostics.push(Diagnostic {
                msg: format!("value {} must contain a non empty comment", name),
                loc: loc.clone(),
            });
            ("".to_string(), None)
        }
    };
    Ok((
        ValueDoc {
            name: name.to_string(),
            headline,
            description,
            flux_type: format!("{}", &typ.normal()),
        },
        diagnostics,
    ))
}

fn comments_to_string(comments: &[ast::Comment]) -> String {
    let mut s = String::new();
    if !comments.is_empty() {
        for c in comments {
            let text = c.text.as_str();
            if let Some(t) = text.strip_prefix("// ") {
                // Strip the leading space if it is present.
                s.push_str(t);
            } else if let Some(t) = text.strip_prefix("//") {
                // An empty comment line will not have the extra space.
                s.push_str(t);
            } else {
                panic!("found invalid comment, all comments must start with //")
            }
        }
    }
    s
}

/// Restructures the Vector of PackageDocs into a hierarchical format where subpackages are in the member section
/// of their parent packages. Ex: monitor.flux docs are in the members section of influxdb docs which are in the members of InfluxData docs.
pub fn nest_docs(original_docs: Vec<PackageDoc>) -> PackageDoc {
    let mut nested_docs = PackageDoc {
        path: "stdlib".to_string(),
        name: "stdlib".to_string(),
        headline: String::new(),
        description: None,
        members: std::collections::BTreeMap::new(),
    };
    for current_pkg in original_docs {
        let parent = find_parent(current_pkg.path.clone(), &mut nested_docs);
        parent.members.insert(
            current_pkg.name.clone(),
            Doc::Package(Box::new(current_pkg)),
        );
    }
    nested_docs
}

/// Find the package directly above the input package and returns it so that
/// we can insert documentation into its members section.
/// Creates an empty parent package if one did not exist.
fn find_parent(path: String, nested_docs: &mut PackageDoc) -> &mut PackageDoc {
    let mut parents: Vec<&str> = path.split('/').collect();
    let mut parent = nested_docs;
    while parents.len() > 1 {
        let pkg = parents.remove(0);
        let path = parent.path.clone();
        let current = parent.members.entry(pkg.to_string()).or_insert_with(|| {
            let path = path + "/" + pkg;
            let path = path.trim_start_matches("stdlib/");
            Doc::Package(Box::new(PackageDoc {
                path: path.to_string(),
                name: pkg.to_string(),
                headline: String::new(),
                description: None,
                members: std::collections::BTreeMap::new(),
            }))
        });
        match current {
            Doc::Package(current) => parent = current,
            _ => panic!(
                "package has a member with the same name as child package: {}",
                pkg,
            ),
        }
    }
    parent
}

#[cfg(test)]
mod test {
    use super::{
        parse_doc_comments, Diagnostic, Diagnostics, Doc, Error, FunctionDoc, PackageDoc,
        ParameterDoc, ValueDoc,
    };

    use crate::ast;
    use crate::ast::tests::Locator;
    use crate::parser::parse_string;
    use crate::semantic::convert::convert_with;
    use crate::semantic::env::Environment;
    use crate::semantic::fresh::Fresher;
    use crate::semantic::nodes;
    use crate::semantic::types::PolyTypeMap;

    use std::collections::BTreeMap;

    macro_rules! map {
        ($( $key: expr => $val: expr ),*$(,)?) => {{
             let mut map = BTreeMap::default();
             $( map.insert($key.to_string(), $val); )*
             map
        }}
    }

    fn parse_program(src: &str) -> ast::Package {
        let file = parse_string("", src);

        ast::Package {
            base: file.base.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![file],
        }
    }
    fn infer_types(pkg: ast::Package) -> Result<PolyTypeMap, Error> {
        let importer = PolyTypeMap::new();
        let mut f = Fresher::from(0);
        let types = match nodes::infer_pkg_types(
            &mut convert_with(pkg, &mut f).expect("analysis failed"),
            Environment::empty(true),
            &mut f,
            &importer,
        ) {
            Ok((env, _)) => env.values,
            Err(e) => {
                return Err(Error {
                    msg: format!("{}", e),
                })
            }
        };
        Ok(types)
    }
    fn assert_docs(src: &str, pkg: PackageDoc, diags: Diagnostics) {
        let ast_pkg = parse_program(src);
        let types = match infer_types(ast_pkg.clone()) {
            Ok(t) => t,
            Err(e) => panic!("error inferring types {}", e),
        };
        let (got_pkg, got_diags) = match parse_doc_comments(&ast_pkg, "path", &types) {
            Ok((p, d)) => (p, d),
            Err(e) => panic!("error parsing doc comments {}", e),
        };
        // assert the diagnostics first as they may contain clues as to why the rest of the docs do
        // not match.
        assert_eq!(
            diags, got_diags,
            "want:\n{:#?}\ngot:\n{:#?}\n",
            diags, got_diags
        );
        assert_eq!(pkg, got_pkg, "want:\n{:#?}\ngot:\n{:#?}\n", pkg, got_pkg);
    }
    #[test]
    fn test_package_doc() {
        let src = "
        // Package foo does a thing
        package foo
        ";
        assert_docs(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing\n".to_string(),
                description: None,
                members: BTreeMap::default(),
            },
            vec![],
        );
    }
    #[test]
    fn test_value_doc_no_desc() {
        let src = "
        // Package foo does a thing
        package foo

        // A is a constant
        a = 1
        ";
        assert_docs(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing\n".to_string(),
                description: None,
                members: map![
                    "a" => Doc::Value(Box::new(ValueDoc{
                        name: "a".to_string(),
                        headline: "A is a constant\n".to_string(),
                        description: None,
                        flux_type: "int".to_string(),
                    })),
                ],
            },
            vec![],
        );
    }
    #[test]
    fn test_value_doc_full() {
        let src = "
        // Package foo does a thing
        package foo

        // A is a constant.
        // The value is one.
        //
        // This is the start of the description.
        //
        // The description contains any remaining markdown content.
        a = 1
        ";
        assert_docs(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing\n".to_string(),
                description: None,
                members: map![
                    "a" => Doc::Value(Box::new(ValueDoc{
                        name: "a".to_string(),
                        headline: "A is a constant.\nThe value is one.\n".to_string(),
                        description: Some("\nThis is the start of the description.\n\nThe description contains any remaining markdown content.\n".to_string()),
                        flux_type: "int".to_string(),
                    })),
                ],
            },
            vec![],
        );
    }
    #[test]
    fn test_function_doc() {
        let src = "
        // Package foo does a thing
        package foo

        // F is a function.
        //
        // F is specifically the identity function, it returns any value it is passed as a
        // parameter.
        //
        // ## Parameters
        // - `x` is any value
        //
        // More description after the parameter list.
        f = (x) => x
        ";
        assert_docs(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing\n".to_string(),
                description: None,
                members: map![
                    "f" => Doc::Function(Box::new(FunctionDoc{
                        name: "f".to_string(),
                        headline: "F is a function.\n".to_string(),
                        description: "\nF is specifically the identity function, it returns any value it is passed as a\nparameter.\n\nMore description after the parameter list.\n".to_string(),
                        parameters: vec![ParameterDoc{
                            name: "x".to_string(),
                            headline: "`x` is any value".to_string(),
                            description: None,
                            required: true,
                        }],
                        flux_type: "(x:A) => A".to_string(),
                    })),
                ],
            },
            vec![],
        );
    }
    #[test]
    fn test_function_doc_parameter_desc() {
        let src = "
        // Package foo does a thing
        package foo

        // F is a function.
        //
        // F is specifically the identity function, it returns any value it is passed as a
        // parameter.
        //
        // ## Parameters
        // - `x` is any value.
        //
        //    Long description of x.
        //
        // - `y` is any value.
        //
        //    Y has a long description too.
        //
        // More description after the parameter list.
        f = (x,y) => x + y
        ";
        assert_docs(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing\n".to_string(),
                description: None,
                members: map![
                    "f" => Doc::Function(Box::new(FunctionDoc{
                        name: "f".to_string(),
                        headline: "F is a function.\n".to_string(),
                        description: "\nF is specifically the identity function, it returns any value it is passed as a\nparameter.\n\nMore description after the parameter list.\n".to_string(),
                        parameters: vec![ParameterDoc{
                            name: "x".to_string(),
                            headline: "`x` is any value.".to_string(),
                            description: Some("\n\n   Long description of x.\n\n".to_string()),
                            required: true,
                        },
                        ParameterDoc{
                            name: "y".to_string(),
                            headline: "`y` is any value.".to_string(),
                            description: Some("\n\n   Y has a long description too.\n\n".to_string()),
                            required: true,
                        }],
                        flux_type: "(x:A, y:A) => A where A: Addable".to_string(),
                    })),
                ],
            },
            vec![],
        );
    }
    #[test]
    fn test_function_doc_missing_description() {
        let src = "
        // Package foo does a thing
        package foo

        // F is a function.
        f = (x) => x
        ";
        let loc = Locator::new(&src[..]);
        assert_docs(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing\n".to_string(),
                description: None,
                members: map![
                    "f" => Doc::Function(Box::new(FunctionDoc{
                        name: "f".to_string(),
                        headline: "F is a function.\n".to_string(),
                        description: "".to_string(),
                        parameters: vec![],
                        flux_type: "(x:A) => A".to_string(),
                    })),
                ],
            },
            vec![Diagnostic {
                msg: "function \"f\" comment must contain both a headline and a description"
                    .to_string(),
                loc: loc.get(6, 9, 6, 21),
            }],
        );
    }
    #[test]
    fn test_function_doc_missing_parameter() {
        let src = "
        // Package foo does a thing
        package foo

        // Add is a function.
        //
        // ## Parameters
        // - `x` is any value
        add = (x,y) => x + y
        ";
        let loc = Locator::new(&src[..]);
        assert_docs(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing\n".to_string(),
                description: None,
                members: map![
                    "add" => Doc::Function(Box::new(FunctionDoc{
                        name: "add".to_string(),
                        headline: "Add is a function.\n".to_string(),
                        description: "\n".to_string(),
                        parameters: vec![ParameterDoc{
                            name: "x".to_string(),
                            headline: "`x` is any value".to_string(),
                            description: None,
                            required: true,
                        }],
                        flux_type: "(x:A, y:A) => A where A: Addable".to_string(),
                    })),
                ],
            },
            vec![Diagnostic {
                msg: "missing documentation for parameter \"y\"".to_string(),
                loc: loc.get(9, 9, 9, 29),
            }],
        );
    }
    #[test]
    fn test_function_doc_missing_optional_parameter() {
        let src = "
        // Package foo does a thing
        package foo

        // Add is a function.
        //
        // ## Parameters
        // - `x` is any value
        add = (x,y=1) => x + y
        ";
        let loc = Locator::new(&src[..]);
        assert_docs(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing\n".to_string(),
                description: None,
                members: map![
                    "add" => Doc::Function(Box::new(FunctionDoc{
                        name: "add".to_string(),
                        headline: "Add is a function.\n".to_string(),
                        description: "\n".to_string(),
                        parameters: vec![ParameterDoc{
                            name: "x".to_string(),
                            headline: "`x` is any value".to_string(),
                            description: None,
                            required: true,
                        }],
                        flux_type: "(x:int, ?y:int) => int".to_string(),
                    })),
                ],
            },
            vec![Diagnostic {
                msg: "missing documentation for parameter \"y\"".to_string(),
                loc: loc.get(9, 9, 9, 31),
            }],
        );
    }
}
