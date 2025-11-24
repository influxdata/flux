//! Generate documentation from source code comments.

pub mod example;

use std::{
    collections::BTreeMap,
    iter::{Iterator, Peekable},
    mem,
    ops::Range,
    sync::LazyLock,
};

use anyhow::{bail, Result};
use derive_more::Display;
use pulldown_cmark::{Event, HeadingLevel, OffsetIter, Parser as MarkdownParser, Tag};
use regex::Regex;

use crate::{
    ast,
    semantic::{
        types::{Function, MonoType, PolyType},
        PackageExports,
    },
};

/// Diagnostic represents an issue with the documentation comments.
/// Something about the formatting or content of the comments does not meet expectations.
#[derive(Eq, PartialEq, Debug, Display)]
#[display(fmt = "error {}: {}", loc, msg)]
pub struct Diagnostic {
    msg: String,
    loc: ast::SourceLocation,
}

/// Diagnostics is a set of diagnostics
pub type Diagnostics = Vec<Diagnostic>;

/// Metadata is arbitrary key value data associated with documentation.
pub type Metadata = BTreeMap<String, String>;

/// Doc is an enum that can take the form of the various types of flux documentation structures through polymorphism.
#[derive(PartialEq, Debug, Serialize, Deserialize)]
#[serde(tag = "kind")]
pub enum Doc {
    /// Package represents documentation for an entire Flux package.
    Package(Box<PackageDoc>),
    /// Value represents documentation for a value exposed from a package.
    Value(Box<ValueDoc>),
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
    /// list of any examples of the value
    pub examples: Vec<Example>,
    /// any Metadata associated with the package
    pub metadata: Option<Metadata>,
}

/// ValueDoc represents the documentation for a single value within a package.
/// Values include options, builtins, or any variable assignment within the top level scope of a
/// package.
#[derive(Eq, PartialEq, Debug, Serialize, Deserialize)]
pub struct ValueDoc {
    /// the name of the value
    pub name: String,
    /// the headline of the value
    pub headline: String,
    /// the description of the value
    pub description: Option<String>,
    /// the type of the value
    pub flux_type: String,
    /// indicates if this value is a Flux option
    pub is_option: bool,
    /// the location in the source code of the value
    pub source_location: ast::SourceLocation,
    /// list of any examples of the value
    pub examples: Vec<Example>,
    /// any Metadata associated with the value
    pub metadata: Option<Metadata>,
}

/// FunctionDoc represents the documentation for a single Function within a package.
#[derive(Eq, PartialEq, Debug, Serialize, Deserialize)]
pub struct FunctionDoc {
    /// the name of the function
    pub name: String,
    /// the headline of the function
    pub headline: String,
    /// the description of the function
    pub description: Option<String>,
    /// the parameters of the function
    pub parameters: Vec<ParameterDoc>,
    /// the type of the function
    pub flux_type: String,
    /// indicates if this function is a Flux option
    pub is_option: bool,
    /// the location in the source code of the function
    pub source_location: ast::SourceLocation,
    /// list of any examples of the function
    pub examples: Vec<Example>,
    /// any Metadata associated with the function
    pub metadata: Option<Metadata>,
}

/// ParameterDoc represents the documentation for a single parameter within a function.
#[derive(Eq, PartialEq, Debug, Serialize, Deserialize)]
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

/// Example represents an extracted example with optional input and outputs.
#[derive(Eq, PartialEq, Debug, Serialize, Deserialize)]
pub struct Example {
    /// Title is the heading given to the example.
    pub title: String,
    /// Content is the source code and other markdown content of the example.
    pub content: String,
    /// If something represents the input to the example.
    pub input: Option<Vec<Table>>,
    /// If something represents the output to the example.
    pub output: Option<Vec<Table>>,
}

/// Rendered markdown of table data.
pub type Table = String;

/// Parse the package documentation for all values within the package.
/// The list of diagnostics reports problems found with formatting or otherwise of the comments.
/// An empty list of diagnostics implies that doc comments are all property formatted.
pub fn parse_package_doc_comments(
    pkg: &ast::Package,
    pkgpath: &str,
    types: &PackageExports,
) -> Result<(PackageDoc, Diagnostics)> {
    // TODO(nathanielc): Support package with more than one file.
    parse_file_doc_comments(&pkg.files[0], pkgpath, types)
}

fn parse_file_doc_comments(
    file: &ast::File,
    pkgpath: &str,
    types: &PackageExports,
) -> Result<(PackageDoc, Diagnostics)> {
    let mut diagnostics: Diagnostics = Vec::new();
    let mut pkg = match &file.package {
        Some(pkg_clause) => {
            let comment = comments_to_string(&pkg_clause.base.comments);
            let pr = parse_comment(
                comment.as_str(),
                false,
                &pkg_clause.base.location,
                &mut diagnostics,
            )?;
            if pr.headline.is_empty() {
                diagnostics.push(Diagnostic {
                    msg: format!(
                        "package {} must contain a non empty package comment",
                        pkgpath
                    ),
                    loc: pkg_clause.base.location.clone(),
                });
            }
            let name = pkg_clause.name.name.clone();
            let words = two_words(pr.headline.as_str());
            let start = format!("Package {}", name);
            if start != words {
                diagnostics.push(Diagnostic {
                    msg: format!(
                        "package headline must start with \"{}\" found \"{}\"",
                        start, words
                    ),
                    loc: pkg_clause.base.location.clone(),
                })
            }
            PackageDoc {
                path: pkgpath.to_string(),
                name,
                headline: pr.headline,
                description: pr.description,
                members: BTreeMap::new(),
                examples: pr.examples,
                metadata: pr.metadata,
            }
        }
        None => {
            diagnostics.push(Diagnostic {
                msg: format!("package {} must contain a package clause", pkgpath),
                loc: file.base.location.clone(),
            });
            // Create a skeleton package doc since we know basically nothing
            PackageDoc {
                path: pkgpath.to_string(),
                name: "".to_string(),
                headline: "".to_string(),
                description: None,
                members: BTreeMap::new(),
                examples: vec![],
                metadata: None,
            }
        }
    };

    let members = parse_package_values(file, types, &mut diagnostics)?;
    pkg.members = members;
    Ok((pkg, diagnostics))
}

// Union of values that can be parsed from a comment
struct ParseResult {
    headline: String,
    description: Option<String>,
    parameters: Vec<HeadlineDescription>,
    examples: Vec<Example>,
    metadata: Option<Metadata>,
}

struct HeadlineDescription {
    headline: String,
    description: Option<String>,
}
fn parse_comment(
    comment: &str,
    expect_parameters: bool,
    loc: &ast::SourceLocation,
    diagnostics: &mut Diagnostics,
) -> Result<ParseResult> {
    let mut parser = Parser::new(comment);
    let tokens_vec = match parser.parse() {
        Ok(t) => t,
        Err(e) => {
            diagnostics.push(Diagnostic {
                msg: format!("parse error {}", e),
                loc: loc.clone(),
            });
            // We didn't get any tokens so return a completely empty parse result.
            // This should only happen if the parser failed to understand the markdown.
            return Ok(ParseResult {
                headline: "".to_string(),
                description: None,
                parameters: Vec::new(),
                examples: Vec::new(),
                metadata: None,
            });
        }
    };

    let mut tokens = tokens_vec.iter().peekable();
    let headline = headline_from_tokens(&mut tokens);
    let description = description_from_tokens(&mut tokens);
    let parameters = if expect_parameters {
        parameters_from_tokens(&mut tokens)
    } else {
        if let Some(Token::Parameters) = tokens.peek() {
            diagnostics.push(Diagnostic {
                msg: "extra Parameters heading".to_string(),
                loc: loc.clone(),
            });
        }
        Vec::new()
    };
    let more_description = description_from_tokens(&mut tokens);
    let description = match (description, more_description) {
        (Some(d), Some(m)) => Some(format!("{}\n\n{}", d, m)),
        (d, None) => d,
        (None, d) => d,
    };
    let description = if let Some(d) = description {
        if d.is_empty() {
            None
        } else {
            Some(d)
        }
    } else {
        description
    };
    let examples = examples_from_tokens(&mut tokens);
    let metadata = metadata_from_tokens(&mut tokens, loc, diagnostics);
    Ok(ParseResult {
        headline,
        description,
        parameters,
        examples,
        metadata,
    })
}

fn headline_from_tokens<'a: 'b, 'b, I>(tokens: &mut Peekable<I>) -> String
where
    I: Iterator<Item = &'b Token<'a>>,
{
    if let Some(Token::Headline(h)) = tokens.peek() {
        tokens.next();
        h.to_string()
    } else {
        String::new()
    }
}
fn description_from_tokens<'a: 'b, 'b, I>(tokens: &mut Peekable<I>) -> Option<String>
where
    I: Iterator<Item = &'b Token<'a>>,
{
    if let Some(Token::Description(h)) = tokens.peek() {
        tokens.next();
        Some(h.to_string())
    } else {
        None
    }
}
fn parameters_from_tokens<'a: 'b, 'b, I>(tokens: &mut Peekable<I>) -> Vec<HeadlineDescription>
where
    I: Iterator<Item = &'b Token<'a>>,
{
    let mut parameters = Vec::with_capacity(tokens.size_hint().0);
    if let Some(Token::Parameters) = tokens.peek() {
        tokens.next();
        loop {
            if let Some(Token::Parameter) = tokens.peek() {
                tokens.next();
                let headline = param_headline_from_tokens(tokens);
                let description = param_description_from_tokens(tokens);
                parameters.push(HeadlineDescription {
                    headline: headline.to_string(),
                    description: description.to_owned(),
                });
            } else {
                return parameters;
            }
        }
    };
    parameters
}
fn param_headline_from_tokens<'a: 'b, 'b, I>(tokens: &mut Peekable<I>) -> String
where
    I: Iterator<Item = &'b Token<'a>>,
{
    if let Some(Token::ParamHeadline(h)) = tokens.peek() {
        tokens.next();
        h.to_string()
    } else {
        String::new()
    }
}
fn param_description_from_tokens<'a: 'b, 'b, I>(tokens: &mut Peekable<I>) -> Option<String>
where
    I: Iterator<Item = &'b Token<'a>>,
{
    if let Some(Token::ParamDescription(h)) = tokens.peek() {
        tokens.next();
        Some(h.to_string())
    } else {
        None
    }
}
fn examples_from_tokens<'a: 'b, 'b, I>(tokens: &mut Peekable<I>) -> Vec<Example>
where
    I: Iterator<Item = &'b Token<'a>>,
{
    let mut examples = Vec::with_capacity(tokens.size_hint().0);
    if let Some(Token::Examples) = tokens.peek() {
        tokens.next();
        loop {
            if let Some(Token::ExampleTitle(title)) = tokens.peek() {
                tokens.next();
                if let Some(Token::ExampleContent(content)) = tokens.peek() {
                    tokens.next();
                    examples.push(Example {
                        title: title.to_string(),
                        content: content.to_string(),
                        input: None,
                        output: None,
                    });
                } else {
                    return examples;
                }
            } else {
                return examples;
            }
        }
    };
    examples
}

fn metadata_from_tokens<'a: 'b, 'b, I>(
    tokens: &mut Peekable<I>,
    loc: &ast::SourceLocation,
    diagnostics: &mut Diagnostics,
) -> Option<Metadata>
where
    I: Iterator<Item = &'b Token<'a>>,
{
    static KEY_VALUE_PATTERN: LazyLock<Regex> =
        LazyLock::new(|| Regex::new("^(\\w[\\w_]+): (.+)$").unwrap());

    if let Some(Token::Metadata) = tokens.peek() {
        tokens.next();
        let mut metadata = Metadata::new();
        while let Some(Token::MetadataLine(line)) = tokens.peek() {
            tokens.next();
            for cap in KEY_VALUE_PATTERN.captures_iter(line) {
                let key = &cap[1];
                let value = &cap[2];
                if metadata.contains_key(key) {
                    diagnostics.push(Diagnostic {
                        msg: format!("found duplicate metadata key \"{}\"", key),
                        loc: loc.clone(),
                    });
                };
                metadata.insert(key.to_string(), value.to_string());
            }
        }
        if !metadata.is_empty() {
            Some(metadata)
        } else {
            None
        }
    } else {
        None
    }
}

// Generates docs for the values in a given source file.
fn parse_package_values(
    f: &ast::File,
    pkgtypes: &PackageExports,
    diagnostics: &mut Diagnostics,
) -> Result<BTreeMap<String, Doc>> {
    let mut members: BTreeMap<String, Doc> = BTreeMap::new();
    for stmt in &f.body {
        if let Some((name, comment, loc, is_option)) = match stmt {
            ast::Statement::Variable(s) => {
                let comment = comments_to_string(&s.id.base.comments);
                let name = s.id.name.clone();
                Some((name, comment, &s.base.location, false))
            }
            ast::Statement::Builtin(s) => {
                let comment = comments_to_string(&s.base.comments);
                let name = s.id.name.clone();
                Some((name, comment, &s.base.location, false))
            }
            ast::Statement::Option(s) => {
                match &s.assignment {
                    ast::Assignment::Variable(v) => {
                        let comment = comments_to_string(&s.base.comments);
                        let name = v.id.name.clone();
                        Some((name, comment, &s.base.location, true))
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
            if let Some(typ) = &pkgtypes.lookup(name.as_str()) {
                if !name.starts_with('_')
                    // No need to check bindings with alternative signatures
                    // that are behind feature flags as there is an original binding as well
                    && !comment.contains("@feature")
                {
                    let doc = parse_any_value(&name, &comment, typ, loc, diagnostics, is_option)?;
                    members.insert(name.clone(), doc);
                }
            } else {
                bail!("type of value {} not found in environment", &name);
            }
        }
    }
    Ok(members)
}

// Returns the first word in the string where words are considered to be delimited by spaces.
fn first_word(s: &str) -> &str {
    n_words(1, s)
}
// Returns first two words in the string where words are considered to be delimited by spaces.
fn two_words(s: &str) -> &str {
    n_words(2, s)
}

// Returns the first n words in the string where words are considered to be delimited by spaces.
fn n_words(n: i32, s: &str) -> &str {
    let bytes = s.as_bytes();

    let mut count: i32 = 0;
    for (i, &item) in bytes.iter().enumerate() {
        if item == b' ' {
            count += 1;
            if count == n {
                return &s[..i];
            }
        }
    }

    s
}

fn check_headline(name: &str, headline: &str, loc: &ast::SourceLocation) -> Option<Diagnostic> {
    let word = first_word(headline);
    if word != name {
        Some(Diagnostic {
            msg: format!("headline must start with \"{}\" found \"{}\"", name, word),
            loc: loc.clone(),
        })
    } else {
        None
    }
}

fn parse_any_value(
    name: &str,
    comment: &str,
    typ: &PolyType,
    loc: &ast::SourceLocation,
    diagnostics: &mut Diagnostics,
    is_option: bool,
) -> Result<Doc> {
    match &typ.expr {
        MonoType::Fun(f) => {
            let doc = parse_function_doc(name, comment, typ, f, loc, diagnostics, is_option)?;
            Ok(Doc::Function(Box::new(doc)))
        }
        _ => {
            let doc = parse_value_doc(name, comment, typ, loc, diagnostics, is_option)?;
            Ok(Doc::Value(Box::new(doc)))
        }
    }
}

fn parse_function_doc(
    name: &str,
    comment: &str,
    typ: &PolyType,
    fun_typ: &Function,
    loc: &ast::SourceLocation,
    diagnostics: &mut Diagnostics,
    is_option: bool,
) -> Result<FunctionDoc> {
    let pr = parse_comment(comment, true, loc, diagnostics)?;
    if pr.headline.is_empty() {
        diagnostics.push(Diagnostic {
            msg: format!("function \"{}\" must contain a non empty comment", name),
            loc: loc.clone(),
        });
    } else if let Some(diagnostic) = check_headline(name, &pr.headline, loc) {
        diagnostics.push(diagnostic)
    }
    let mut parameters: Vec<ParameterDoc> = Vec::with_capacity(pr.parameters.len());
    for parameter in pr.parameters {
        let mut name = String::new();
        if let Some(n) = first_word(&parameter.headline).strip_suffix(':') {
            name = n.to_string();
        }
        if name.is_empty() {
            diagnostics.push(Diagnostic {
                msg: "parameter headline must start with \"{parameter_name}:\"".to_string(),
                loc: loc.clone(),
            });
        }
        let required = fun_typ.req.contains_key(&name);
        parameters.push(ParameterDoc {
            name,
            headline: parameter.headline,
            description: parameter.description,
            required,
        })
    }
    // Validate all parameters were documented
    let mut params_on_type: Vec<&String> = fun_typ.req.keys().chain(fun_typ.opt.keys()).collect();
    if let Some(pipe) = &fun_typ.pipe {
        // Add pipe parameter to set if it exists
        params_on_type.push(&pipe.k)
    }
    for name in &params_on_type {
        if !contains_parameter(&parameters, name.as_str()) {
            diagnostics.push(Diagnostic {
                msg: format!("missing documentation for parameter \"{}\"", name),
                loc: loc.clone(),
            });
        }
    }

    // Validate extra parameters are not documented
    for param in &parameters {
        if !param.name.is_empty() && !params_on_type.iter().any(|&name| name == &param.name) {
            diagnostics.push(Diagnostic {
                msg: format!("extra documentation for parameter \"{}\"", param.name,),
                loc: loc.clone(),
            });
        }
    }

    Ok(FunctionDoc {
        name: name.to_string(),
        headline: pr.headline,
        description: pr.description,
        parameters,
        flux_type: format!("{}", &typ.normal()),
        is_option,
        source_location: loc.clone(),
        examples: pr.examples,
        metadata: pr.metadata,
    })
}

fn contains_parameter(params: &[ParameterDoc], name: &str) -> bool {
    params.iter().any(|pd| pd.name == name)
}

fn parse_value_doc(
    name: &str,
    comment: &str,
    typ: &PolyType,
    loc: &ast::SourceLocation,
    diagnostics: &mut Diagnostics,
    is_option: bool,
) -> Result<ValueDoc> {
    let pr = parse_comment(comment, false, loc, diagnostics)?;
    if pr.headline.is_empty() {
        diagnostics.push(Diagnostic {
            msg: format!("value {} must contain a non empty comment", name),
            loc: loc.clone(),
        });
    } else if let Some(diagnostic) = check_headline(name, &pr.headline, loc) {
        diagnostics.push(diagnostic)
    }
    Ok(ValueDoc {
        name: name.to_string(),
        headline: pr.headline,
        description: pr.description,
        flux_type: format!("{}", &typ.normal()),
        is_option,
        source_location: loc.clone(),
        examples: pr.examples,
        metadata: pr.metadata,
    })
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

/// Shorten removes all long form descriptions from the docs structures leaving only the headlines
/// and other metadata.
pub fn shorten(doc: &mut PackageDoc) {
    doc.description = None;
    doc.examples = Vec::new();
    for (_, m) in doc.members.iter_mut() {
        shorten_doc(m);
    }
}

/// Removes the description from a Doc.
///
/// This function is recursive via the [`shorten`] function.
/// This design allows the implementation for the Doc::Package variant to share code with
/// [`shorten`] and keep the original data types as &mut instead of moving the data into these
/// functions.
fn shorten_doc(doc: &mut Doc) {
    match doc {
        Doc::Package(p) => shorten(p),
        Doc::Value(v) => {
            v.description = None;
            v.examples = Vec::new();
        }
        Doc::Function(f) => {
            f.description = None;
            f.examples = Vec::new();
            for p in f.parameters.iter_mut() {
                p.description = None
            }
        }
    }
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
        examples: Vec::new(),
        metadata: None,
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
                examples: Vec::new(),
                metadata: None,
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

const PARAMETER_HEADING: &str = "Parameters";
const EXAMPLES_HEADING: &str = "Examples";
const METADATA_HEADING: &str = "Metadata";

// Parser produces a series of tokens from documentation comments.
struct Parser<'a> {
    content: &'a str,
    iter: Peekable<OffsetIter<'a, 'a>>,
    tokens: Vec<Token<'a>>,
}

impl<'a> Parser<'a> {
    fn slice(&self, r: Range<usize>) -> &'a str {
        self.content[r].trim()
    }
}

#[derive(PartialEq, Debug)]
enum Token<'a> {
    Headline(&'a str),
    Description(&'a str),
    Parameters,
    Parameter,
    ParamHeadline(&'a str),
    ParamDescription(&'a str),
    Examples,
    ExampleTitle(&'a str),
    ExampleContent(&'a str),
    Metadata,
    MetadataLine(&'a str),
}

impl<'a> Parser<'a> {
    fn new(content: &'a str) -> Parser<'a> {
        Parser {
            content,
            iter: MarkdownParser::new(content).into_offset_iter().peekable(),
            // Most comments will have less than 100 tokens and 100 is a small enough value that
            // pre-allocating will not be a big concern.
            tokens: Vec::with_capacity(100),
        }
    }
    // In a single pass parse the content into its tokens.
    //
    // An error is returned only when an assumption about parsing markdown is violated (i.e. no
    // end event after a start event).
    //
    // Otherwise Tokens are produced with a best effort.
    fn parse(&mut self) -> Result<Vec<Token<'a>>> {
        self.parse_headline()?;
        Ok(mem::take(&mut self.tokens))
    }
    fn parse_headline(&mut self) -> Result<()> {
        let mut range = Range::<usize> { start: 0, end: 0 };
        // We will either have a paragraph or a single text node
        match self.iter.next() {
            Some((Event::Start(Tag::Paragraph), r)) => {
                range.start = r.start;
            }
            Some((Event::Text(_), r)) => {
                self.tokens.push(Token::Headline(self.slice(r)));
                return self.parse_description();
            }
            _ => {
                // We failed to parse a headline, move on to next possible tokens.
                return self.parse_description();
            }
        };
        // We have a paragraph so gather all events until the end of the paragraph.
        loop {
            match self.iter.next() {
                Some((Event::End(Tag::Paragraph), r)) => {
                    range.end = r.end;
                    self.tokens.push(Token::Headline(self.slice(range)));
                    return self.parse_description();
                }
                //do nothing but catch the event
                Some(_) => {}
                None => {
                    bail!("reached end of markdown without reaching end of paragraph")
                }
            }
        }
    }

    fn parse_description(&mut self) -> Result<()> {
        let mut range: Range<usize> = Range::default();
        if let Some((_, r)) = self.iter.peek() {
            range.start = r.start;
        } else {
            // We reached the end of the markdown content, stop lexing
            return Ok(());
        }
        // Peek and consume items until we see a delimiter heading
        loop {
            match self.iter.next() {
                Some((Event::Start(Tag::Heading(HeadingLevel::H2, _, _)), r)) => {
                    if let Some((Event::Text(t), _)) = self.iter.peek() {
                        // The description ends at the start of this heading
                        range.end = r.start;
                        match t.as_ref() {
                            PARAMETER_HEADING => {
                                self.tokens.push(Token::Description(self.slice(range)));
                                return self.parse_parameters();
                            }
                            EXAMPLES_HEADING => {
                                self.tokens.push(Token::Description(self.slice(range)));
                                return self.parse_examples();
                            }
                            METADATA_HEADING => {
                                self.tokens.push(Token::Description(self.slice(range)));
                                return self.parse_metadata();
                            }
                            // If we didn't find a delimiter heading then keep consuming items.
                            _ => {}
                        };
                    }
                }
                Some(_) => {}
                // We reached the end of the markdown content, stop lexing and return token
                None => {
                    range.end = self.content.len();
                    self.tokens.push(Token::Description(self.slice(range)));
                    return Ok(());
                }
            }
        }
    }

    fn parse_any_heading_or_description(&mut self) -> Result<()> {
        match self.iter.peek() {
            Some((Event::Start(Tag::Heading(HeadingLevel::H2, _, _)), _)) => {
                self.iter.next();
                self.parse_any_heading_text()
            }
            Some(_) => self.parse_description(),
            // We reached the end of the markdown content, stop lexing
            None => Ok(()),
        }
    }

    fn parse_any_heading_text(&mut self) -> Result<()> {
        if let Some((Event::Text(t), _)) = self.iter.peek() {
            match t.as_ref() {
                PARAMETER_HEADING => self.parse_parameters(),
                EXAMPLES_HEADING => self.parse_examples(),
                METADATA_HEADING => self.parse_metadata(),
                _ => {
                    // We didn't find any delimiting heading
                    // There is no where to go from here so simply end parsing.
                    Ok(())
                }
            }
        } else {
            bail!("expected heading text")
        }
    }

    fn parse_parameters(&mut self) -> Result<()> {
        // Discard the "Parameters" text item and heading end
        if self
            .iter
            .next_if(|e| matches!(e, (Event::Text(_), _)))
            .is_none()
        {
            bail!("missing parameters text")
        }
        if self
            .iter
            .next_if(|e| matches!(e, (Event::End(Tag::Heading(HeadingLevel::H2, _, _)), _)))
            .is_none()
        {
            bail!("missing end of heading")
        }
        match self.iter.next() {
            Some((Event::Start(Tag::List(_)), _)) => {
                self.tokens.push(Token::Parameters);
                // Note: parse_parameter is recursive calling itself until the end of the
                // parameter list is found.
                self.parse_parameter()
            }
            _ => {
                // We didn't find a list so we start over looking for the next heading.
                self.parse_any_heading_or_description()
            }
        }
    }

    fn parse_parameter(&mut self) -> Result<()> {
        match self.iter.next() {
            Some((Event::Start(Tag::Item), _)) => {
                self.tokens.push(Token::Parameter);
                self.parse_parameter_headline()
            }
            Some((Event::End(Tag::List(_)), _)) => {
                // We reached the end of the parameters list
                // Start lexing the next section.
                self.parse_any_heading_or_description()
            }
            _ => {
                // We didn't find another item, start over looking for the next heading.
                self.parse_any_heading_or_description()
            }
        }
    }
    fn parse_parameter_headline(&mut self) -> Result<()> {
        let mut range = Range::<usize> { start: 0, end: 0 };
        // We will either have a paragraph or content within the entire item.
        match self.iter.next() {
            Some((Event::Start(Tag::Paragraph), r)) => {
                range.start = r.start;
            }
            Some((_, start)) => {
                // We do not have an explicit paragraph so assume the entire item is the headline.
                loop {
                    match self.iter.next() {
                        Some((Event::End(Tag::Item), end)) => {
                            self.tokens.push(Token::ParamHeadline(self.slice(Range {
                                start: start.start,
                                end: end.end,
                            })));
                            // Parse the next parameter
                            return self.parse_parameter();
                        }
                        Some((Event::Start(Tag::Item), _)) => {
                            // We found a new list within the headline we should bail with a helpful message.
                            bail!("found a new list within a parameter headline. Use a new paragraph to separate the list from the headline.")
                        }
                        Some(_) => {}
                        None => bail!("reached end of markdown without reaching end of item"),
                    };
                }
            }
            None => bail!("reached end of markdown without reaching end of item"),
        };
        // We have a paragraph so gather all events until the end of the paragraph.
        loop {
            match self.iter.next() {
                Some((Event::End(Tag::Paragraph), r)) => {
                    range.end = r.end;
                    self.tokens.push(Token::ParamHeadline(self.slice(range)));
                    return self.parse_parameter_description();
                }
                //do nothing but catch the event
                Some(_) => {}
                None => {
                    bail!("reached end of markdown without reaching end of paragraph")
                }
            }
        }
    }

    fn parse_parameter_description(&mut self) -> Result<()> {
        let mut range: Range<usize> = Range::default();
        if let Some((_, r)) = self.iter.peek() {
            range.start = r.start;
        } else {
            bail!("reached the end of markdown without reaching end of item")
        }
        let mut depth = 0;
        // Peek and consume events until we see an end item
        loop {
            match self.iter.next() {
                Some((Event::Start(Tag::List(_)), _)) => {
                    depth += 1;
                }
                Some((Event::End(Tag::List(_)), _)) => {
                    depth -= 1;
                }
                Some((Event::End(Tag::Item), r)) => {
                    if depth == 0 {
                        range.end = r.end;
                        if range != r {
                            // If the outer range is the same as the Tag::Item range then we didn't
                            // find any new events, meaning we do not have a description.
                            self.tokens.push(Token::ParamDescription(self.slice(range)));
                        }
                        // Recurse back to parse_parameter to look for more parameters.
                        return self.parse_parameter();
                    }
                }
                Some(_) => {}
                None => bail!("reached the end of markdown without reaching end of item"),
            }
        }
    }

    fn parse_examples(&mut self) -> Result<()> {
        // Discard the "Examples" text item and heading end
        if self
            .iter
            .next_if(|e| matches!(e, (Event::Text(_), _)))
            .is_none()
        {
            bail!("missing parameters text")
        }
        if self
            .iter
            .next_if(|e| matches!(e, (Event::End(Tag::Heading(HeadingLevel::H2, _, _)), _)))
            .is_none()
        {
            bail!("missing end of heading")
        }
        self.tokens.push(Token::Examples);
        let mut range: Range<usize> = Range::default();
        let mut count = 0;
        loop {
            match self.iter.next() {
                Some((Event::Start(Tag::Heading(HeadingLevel::H2, _, _)), r)) => {
                    // Heading 2 means we are done with examples
                    // We found the begining of a new section, emit the content token.
                    range.end = r.start;
                    self.tokens.push(Token::ExampleContent(self.slice(range)));
                    return self.parse_any_heading_text();
                }
                Some((Event::End(Tag::Heading(HeadingLevel::H3, _, _)), r)) => {
                    range.end = r.start;
                    if count > 0 {
                        // We found another example emit the content token
                        self.tokens
                            .push(Token::ExampleContent(self.slice(range.clone())));
                    }
                    count += 1;
                    // The example content starts where the heading ends
                    range.start = r.end;
                    self.tokens.push(Token::ExampleTitle(self.slice(r)));
                }
                Some(_) => {}
                None => {
                    // We found the end of the markdown emit the final content token
                    range.end = self.content.len();
                    self.tokens.push(Token::ExampleContent(self.slice(range)));
                    return Ok(());
                }
            }
        }
    }
    fn parse_metadata(&mut self) -> Result<()> {
        // Discard the "Metadata" text item and heading end
        if self
            .iter
            .next_if(|e| matches!(e, (Event::Text(_), _)))
            .is_none()
        {
            bail!("missing parameters text")
        }
        if self
            .iter
            .next_if(|e| matches!(e, (Event::End(Tag::Heading(HeadingLevel::H2, _, _)), _)))
            .is_none()
        {
            bail!("missing end of heading")
        }
        self.tokens.push(Token::Metadata);
        let mut range: Range<usize> = Range::default();
        loop {
            match self.iter.next() {
                Some((Event::Start(Tag::Heading(HeadingLevel::H2, _, _)), r)) => {
                    // Heading 2 means we are done with metadata
                    // We found the begining of a new section, emit the line token.
                    range.end = r.start;
                    self.tokens.push(Token::MetadataLine(self.slice(r)));
                    return self.parse_any_heading_text();
                }
                Some((Event::Text(_), r)) => {
                    self.tokens.push(Token::MetadataLine(self.slice(r)));
                }
                Some(_) => {}
                None => {
                    return Ok(());
                }
            }
        }
    }
}

#[cfg(test)]
mod test {
    use expect_test::expect;

    use super::{parse_package_doc_comments, shorten, Diagnostics, PackageDoc, Parser, Token};
    use crate::{
        ast,
        parser::parse_string,
        semantic::{env::Environment, import::Packages, Analyzer},
    };

    fn parse_program(src: &str) -> ast::Package {
        let file = parse_string("".to_string(), src);

        ast::Package {
            base: file.base.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![file],
        }
    }
    fn assert_parser(src: &str, want: Vec<Token>) {
        let mut parser = Parser::new(src);
        let got = parser.parse().unwrap();
        assert_eq!(want, got, "\nwant:\n{:#?}\ngot:\n{:#?}\n", want, got);
    }
    fn assert_docs_full(src: &str) -> (PackageDoc, Diagnostics) {
        assert_docs(src, false)
    }
    fn assert_docs_short(src: &str) -> (PackageDoc, Diagnostics) {
        assert_docs(src, true)
    }
    fn assert_docs(src: &str, short: bool) -> (PackageDoc, Diagnostics) {
        let mut analyzer = Analyzer::new_with_defaults(Environment::empty(true), Packages::new());
        let ast_pkg = parse_program(src);
        let (types, _) = match analyzer.analyze_ast(&ast_pkg) {
            Ok(t) => t,
            Err(e) => panic!("error inferring types {}", e),
        };
        let (mut got_pkg, got_diags) = match parse_package_doc_comments(&ast_pkg, "path", &types) {
            Ok((p, d)) => (p, d),
            Err(e) => panic!("error parsing doc comments: {}", e),
        };
        if short {
            shorten(&mut got_pkg);
        }
        (got_pkg, got_diags)
    }
    #[test]
    fn test_package_doc() {
        let src = "
        // Package foo does a thing.
        package foo
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {},
                    examples: [],
                    metadata: None,
                },
                [],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_package_private_values() {
        let src = "
        // Package foo does a thing.
        package foo

        _thisIsPrivate = 1
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {},
                    examples: [],
                    metadata: None,
                },
                [],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_package_headline_invalid() {
        let src = "
        // foo does a thing.
        package foo
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "foo does a thing.",
                    description: None,
                    members: {},
                    examples: [],
                    metadata: None,
                },
                [
                    Diagnostic {
                        msg: "package headline must start with \"Package foo\" found \"foo does\"",
                        loc: SourceLocation {
                            start: "line: 3, column: 9",
                            end: "line: 3, column: 20",
                            source: "package foo",
                        },
                    },
                ],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_value_doc_no_desc() {
        let src = "
        // Package foo does a thing.
        package foo

        // a is a constant.
        a = 1
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "a": Value(
                            ValueDoc {
                                name: "a",
                                headline: "a is a constant.",
                                description: None,
                                flux_type: "int",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 6, column: 9",
                                    end: "line: 6, column: 14",
                                    source: "a = 1",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_value_doc_multiline_headline_no_desc() {
        let src = "
        // Package foo does a thing.
        package foo

        // a is a constant. This headline has `code`
        // and multiple lines.
        a = 1
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "a": Value(
                            ValueDoc {
                                name: "a",
                                headline: "a is a constant. This headline has `code`\nand multiple lines.",
                                description: None,
                                flux_type: "int",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 7, column: 9",
                                    end: "line: 7, column: 14",
                                    source: "a = 1",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [],
            )
        "#]].assert_debug_eq(&docs);
    }
    #[test]
    fn test_value_doc_code_headline_no_desc() {
        let src = "
        // Package foo does a thing.
        package foo

        // a is a constant. This headline has `code`.
        a = 1
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "a": Value(
                            ValueDoc {
                                name: "a",
                                headline: "a is a constant. This headline has `code`.",
                                description: None,
                                flux_type: "int",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 6, column: 9",
                                    end: "line: 6, column: 14",
                                    source: "a = 1",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_value_doc_headline_invalid() {
        let src = "
        // Package foo does a thing.
        package foo

        // A is a constant.
        a = 1
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "a": Value(
                            ValueDoc {
                                name: "a",
                                headline: "A is a constant.",
                                description: None,
                                flux_type: "int",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 6, column: 9",
                                    end: "line: 6, column: 14",
                                    source: "a = 1",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [
                    Diagnostic {
                        msg: "headline must start with \"a\" found \"A\"",
                        loc: SourceLocation {
                            start: "line: 6, column: 9",
                            end: "line: 6, column: 14",
                            source: "a = 1",
                        },
                    },
                ],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_value_doc_full() {
        let src = "
        // Package foo does a thing.
        package foo

        // a is a constant.
        // The value is one.
        //
        // This is the start of the description.
        //
        // The description contains any remaining markdown content.
        a = 1
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "a": Value(
                            ValueDoc {
                                name: "a",
                                headline: "a is a constant.\nThe value is one.",
                                description: Some(
                                    "This is the start of the description.\n\nThe description contains any remaining markdown content.",
                                ),
                                flux_type: "int",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 11, column: 9",
                                    end: "line: 11, column: 14",
                                    source: "a = 1",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [],
            )
        "#]].assert_debug_eq(&docs);
    }
    #[test]
    fn test_shorten() {
        let src = r#"
        // Package foo does a thing.
        //
        // This is a description.
        //
        // ## Examples
        //
        // ### Using foo
        //
        // ```
        // import "foo"
        //
        // foo.a
        // ```
        package foo

        // a is a constant.
        //
        // This is a description.
        //
        // ## Examples
        //
        // ### Using a
        //
        // ```
        // # import "foo"
        // foo.a
        // ```
        a = 1

        // f is a function.
        //
        // This is a description.
        //
        // ## Parameters
        //
        // - x: is a parameter.
        //
        //     This is a description of x.
        //
        // ## Examples
        //
        // ### Using f
        //
        // ```
        // # import "foo"
        // foo.f(x:1)
        // ```
        f = (x) => 1

        // o is an option.
        //
        // This is a description.
        //
        // ## Examples
        //
        // ### Using o
        //
        // ```
        // # import "foo"
        // option foo.o = 2
        // ```
        option o = 1
        "#;
        let docs = assert_docs_short(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "a": Value(
                            ValueDoc {
                                name: "a",
                                headline: "a is a constant.",
                                description: None,
                                flux_type: "int",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 29, column: 9",
                                    end: "line: 29, column: 14",
                                    source: "a = 1",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                        "f": Function(
                            FunctionDoc {
                                name: "f",
                                headline: "f is a function.",
                                description: None,
                                parameters: [
                                    ParameterDoc {
                                        name: "x",
                                        headline: "x: is a parameter.",
                                        description: None,
                                        required: true,
                                    },
                                ],
                                flux_type: "(x: A) => int",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 49, column: 9",
                                    end: "line: 49, column: 21",
                                    source: "f = (x) => 1",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                        "o": Value(
                            ValueDoc {
                                name: "o",
                                headline: "o is an option.",
                                description: None,
                                flux_type: "int",
                                is_option: true,
                                source_location: SourceLocation {
                                    start: "line: 63, column: 9",
                                    end: "line: 63, column: 21",
                                    source: "option o = 1",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_examples() {
        let src = r#"
        // Package foo does a thing.
        //
        // This is a description.
        //
        // ## Examples
        //
        // ### Using foo
        //
        // ```
        // import "foo"
        //
        // foo.a
        // ```
        package foo

        // a is a constant.
        //
        // This is a description.
        //
        // ## Examples
        //
        // ### Using a
        //
        // ```
        // # import "foo"
        // foo.a
        // ```
        a = 1

        // f is a function.
        //
        // This is a description.
        //
        // ## Parameters
        //
        // - x: is a parameter.
        //
        //     This is a description of x.
        //
        // ## Examples
        //
        // ### Using f
        //
        // ```
        // # import "foo"
        // foo.f(x:1)
        // ```
        f = (x) => 1

        // o is an option.
        //
        // This is a description.
        //
        // ## Examples
        //
        // ### Using o
        //
        // ```
        // # import "foo"
        // option foo.o = 2
        // ```
        option o = 1
        "#;
        let docs = assert_docs_full(src);
        expect![[r####"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: Some(
                        "This is a description.",
                    ),
                    members: {
                        "a": Value(
                            ValueDoc {
                                name: "a",
                                headline: "a is a constant.",
                                description: Some(
                                    "This is a description.",
                                ),
                                flux_type: "int",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 29, column: 9",
                                    end: "line: 29, column: 14",
                                    source: "a = 1",
                                },
                                examples: [
                                    Example {
                                        title: "### Using a",
                                        content: "```\n# import \"foo\"\nfoo.a\n```",
                                        input: None,
                                        output: None,
                                    },
                                ],
                                metadata: None,
                            },
                        ),
                        "f": Function(
                            FunctionDoc {
                                name: "f",
                                headline: "f is a function.",
                                description: Some(
                                    "This is a description.",
                                ),
                                parameters: [
                                    ParameterDoc {
                                        name: "x",
                                        headline: "x: is a parameter.",
                                        description: Some(
                                            "This is a description of x.",
                                        ),
                                        required: true,
                                    },
                                ],
                                flux_type: "(x: A) => int",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 49, column: 9",
                                    end: "line: 49, column: 21",
                                    source: "f = (x) => 1",
                                },
                                examples: [
                                    Example {
                                        title: "### Using f",
                                        content: "```\n# import \"foo\"\nfoo.f(x:1)\n```",
                                        input: None,
                                        output: None,
                                    },
                                ],
                                metadata: None,
                            },
                        ),
                        "o": Value(
                            ValueDoc {
                                name: "o",
                                headline: "o is an option.",
                                description: Some(
                                    "This is a description.",
                                ),
                                flux_type: "int",
                                is_option: true,
                                source_location: SourceLocation {
                                    start: "line: 63, column: 9",
                                    end: "line: 63, column: 21",
                                    source: "option o = 1",
                                },
                                examples: [
                                    Example {
                                        title: "### Using o",
                                        content: "```\n# import \"foo\"\noption foo.o = 2\n```",
                                        input: None,
                                        output: None,
                                    },
                                ],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [
                        Example {
                            title: "### Using foo",
                            content: "```\nimport \"foo\"\n\nfoo.a\n```",
                            input: None,
                            output: None,
                        },
                    ],
                    metadata: None,
                },
                [],
            )
        "####]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_metadata_all_docs() {
        let src = "
        // Package foo does a thing.
        //
        // This is a description.
        //
        // ## Metadata
        // k0: v0
        // k1: v1
        // k2: v2
        package foo

        // a is a constant.
        //
        // This is a description.
        //
        // ## Metadata
        // k3: v3
        // k4: v4
        // k5: v5
        a = 1

        // f is a function.
        //
        // This is a description.
        //
        // ## Parameters
        //
        // - x: is a parameter.
        //
        //     This is a description of x.
        //
        // ## Metadata
        // k6: v6
        // k7: v7
        // k8: v8
        f = (x) => 1

        // o is an option.
        //
        // This is a description.
        //
        // ## Metadata
        // k9: v9
        // k0: v0
        option o = 1
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: Some(
                        "This is a description.",
                    ),
                    members: {
                        "a": Value(
                            ValueDoc {
                                name: "a",
                                headline: "a is a constant.",
                                description: Some(
                                    "This is a description.",
                                ),
                                flux_type: "int",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 20, column: 9",
                                    end: "line: 20, column: 14",
                                    source: "a = 1",
                                },
                                examples: [],
                                metadata: Some(
                                    {
                                        "k3": "v3",
                                        "k4": "v4",
                                        "k5": "v5",
                                    },
                                ),
                            },
                        ),
                        "f": Function(
                            FunctionDoc {
                                name: "f",
                                headline: "f is a function.",
                                description: Some(
                                    "This is a description.",
                                ),
                                parameters: [
                                    ParameterDoc {
                                        name: "x",
                                        headline: "x: is a parameter.",
                                        description: Some(
                                            "This is a description of x.",
                                        ),
                                        required: true,
                                    },
                                ],
                                flux_type: "(x: A) => int",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 36, column: 9",
                                    end: "line: 36, column: 21",
                                    source: "f = (x) => 1",
                                },
                                examples: [],
                                metadata: Some(
                                    {
                                        "k6": "v6",
                                        "k7": "v7",
                                        "k8": "v8",
                                    },
                                ),
                            },
                        ),
                        "o": Value(
                            ValueDoc {
                                name: "o",
                                headline: "o is an option.",
                                description: Some(
                                    "This is a description.",
                                ),
                                flux_type: "int",
                                is_option: true,
                                source_location: SourceLocation {
                                    start: "line: 45, column: 9",
                                    end: "line: 45, column: 21",
                                    source: "option o = 1",
                                },
                                examples: [],
                                metadata: Some(
                                    {
                                        "k0": "v0",
                                        "k9": "v9",
                                    },
                                ),
                            },
                        ),
                    },
                    examples: [],
                    metadata: Some(
                        {
                            "k0": "v0",
                            "k1": "v1",
                            "k2": "v2",
                        },
                    ),
                },
                [],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_metadata_pkg() {
        let src = "
        // Package foo does a thing.
        //
        // This is a description.
        //
        // ## Metadata
        // key: valueA
        // key: valueB
        // key1: value with spaces
        // key_with_underscores: value
        package foo
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: Some(
                        "This is a description.",
                    ),
                    members: {},
                    examples: [],
                    metadata: Some(
                        {
                            "key": "valueB",
                            "key1": "value with spaces",
                            "key_with_underscores": "value",
                        },
                    ),
                },
                [
                    Diagnostic {
                        msg: "found duplicate metadata key \"key\"",
                        loc: SourceLocation {
                            start: "line: 11, column: 9",
                            end: "line: 11, column: 20",
                            source: "package foo",
                        },
                    },
                ],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_metadata_no_desc_pkg() {
        let src = "
        // Package foo does a thing.
        //
        // ## Metadata
        // key: valueA
        // key: valueB
        // key1: value with spaces
        // key_with_underscores: value
        package foo
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {},
                    examples: [],
                    metadata: Some(
                        {
                            "key": "valueB",
                            "key1": "value with spaces",
                            "key_with_underscores": "value",
                        },
                    ),
                },
                [
                    Diagnostic {
                        msg: "found duplicate metadata key \"key\"",
                        loc: SourceLocation {
                            start: "line: 9, column: 9",
                            end: "line: 9, column: 20",
                            source: "package foo",
                        },
                    },
                ],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_function_doc() {
        let src = "
        // Package foo does a thing.
        package foo

        // f is a function.
        //
        // More specifically f is the identity function, it returns any value it is passed as a
        // parameter.
        //
        // ## Parameters
        // - x: is any value.
        // - p: is any value piped to the function.
        //
        // More description after the parameter list.
        f = (x,p=<-) => p + x
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "f": Function(
                            FunctionDoc {
                                name: "f",
                                headline: "f is a function.",
                                description: Some(
                                    "More specifically f is the identity function, it returns any value it is passed as a\nparameter.\n\nMore description after the parameter list.",
                                ),
                                parameters: [
                                    ParameterDoc {
                                        name: "x",
                                        headline: "x: is any value.",
                                        description: None,
                                        required: true,
                                    },
                                    ParameterDoc {
                                        name: "p",
                                        headline: "p: is any value piped to the function.",
                                        description: None,
                                        required: false,
                                    },
                                ],
                                flux_type: "(<-p: A, x: A) => A where A: Addable",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 15, column: 9",
                                    end: "line: 15, column: 30",
                                    source: "f = (x,p=<-) => p + x",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [],
            )
        "#]].assert_debug_eq(&docs);
    }
    #[test]
    fn test_function_doc_multiline() {
        // It is possible in markdown for a list item to contain mutliple lines without
        // having an explicit paragraph tag, this test case validates that such soft paragraphs are
        // correctly captured into the headline.

        let src = "
        // Package foo does a thing.
        package foo

        // f is a function
        //
        // ## Parameters
        // - a: parameter with a multiline
        //     headline without a paragraph.
        // - b: parameter with `code` and a multiline
        //     headline without a paragraph.
        // - c: parameter with a multiline
        //     headline without a paragraph but with `code`.
        f = (a, b, c) => 1
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "f": Function(
                            FunctionDoc {
                                name: "f",
                                headline: "f is a function",
                                description: None,
                                parameters: [
                                    ParameterDoc {
                                        name: "a",
                                        headline: "a: parameter with a multiline\n    headline without a paragraph.",
                                        description: None,
                                        required: true,
                                    },
                                    ParameterDoc {
                                        name: "b",
                                        headline: "b: parameter with `code` and a multiline\n    headline without a paragraph.",
                                        description: None,
                                        required: true,
                                    },
                                    ParameterDoc {
                                        name: "c",
                                        headline: "c: parameter with a multiline\n    headline without a paragraph but with `code`.",
                                        description: None,
                                        required: true,
                                    },
                                ],
                                flux_type: "(a: A, b: B, c: C) => int",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 14, column: 9",
                                    end: "line: 14, column: 27",
                                    source: "f = (a, b, c) => 1",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [],
            )
        "#]].assert_debug_eq(&docs);
    }
    #[test]
    fn test_function_headline_invalid() {
        let src = "
        // Package foo does a thing.
        package foo

        // F is a function.
        f = () => 1
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "f": Function(
                            FunctionDoc {
                                name: "f",
                                headline: "F is a function.",
                                description: None,
                                parameters: [],
                                flux_type: "() => int",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 6, column: 9",
                                    end: "line: 6, column: 20",
                                    source: "f = () => 1",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [
                    Diagnostic {
                        msg: "headline must start with \"f\" found \"F\"",
                        loc: SourceLocation {
                            start: "line: 6, column: 9",
                            end: "line: 6, column: 20",
                            source: "f = () => 1",
                        },
                    },
                ],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_function_doc_parameter_desc() {
        let src = "
        // Package foo does a thing.
        package foo

        // f is a function.
        //
        // More specifically f is the identity function, it returns any value it is passed as a
        // parameter.
        //
        // ## Parameters
        // - x: is any value.
        //
        //    Long description of x.
        //
        // - y: is any value.
        //
        //    Y has a long description too.
        //
        // More description after the parameter list.
        f = (x,y) => x + y
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "f": Function(
                            FunctionDoc {
                                name: "f",
                                headline: "f is a function.",
                                description: Some(
                                    "More specifically f is the identity function, it returns any value it is passed as a\nparameter.\n\nMore description after the parameter list.",
                                ),
                                parameters: [
                                    ParameterDoc {
                                        name: "x",
                                        headline: "x: is any value.",
                                        description: Some(
                                            "Long description of x.",
                                        ),
                                        required: true,
                                    },
                                    ParameterDoc {
                                        name: "y",
                                        headline: "y: is any value.",
                                        description: Some(
                                            "Y has a long description too.",
                                        ),
                                        required: true,
                                    },
                                ],
                                flux_type: "(x: A, y: A) => A where A: Addable",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 20, column: 9",
                                    end: "line: 20, column: 27",
                                    source: "f = (x,y) => x + y",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [],
            )
        "#]].assert_debug_eq(&docs);
    }
    #[test]
    fn test_function_doc_parameter_name_invalid() {
        let src = "
        // Package foo does a thing.
        package foo

        // f is a function.
        //
        // More specifically f is the identity function, it returns any value it is passed as a
        // parameter.
        //
        // ## Parameters
        // - x is any value.
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
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "f": Function(
                            FunctionDoc {
                                name: "f",
                                headline: "f is a function.",
                                description: Some(
                                    "More specifically f is the identity function, it returns any value it is passed as a\nparameter.\n\nMore description after the parameter list.",
                                ),
                                parameters: [
                                    ParameterDoc {
                                        name: "",
                                        headline: "x is any value.",
                                        description: Some(
                                            "Long description of x.",
                                        ),
                                        required: false,
                                    },
                                    ParameterDoc {
                                        name: "",
                                        headline: "`y` is any value.",
                                        description: Some(
                                            "Y has a long description too.",
                                        ),
                                        required: false,
                                    },
                                ],
                                flux_type: "(x: A, y: A) => A where A: Addable",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 20, column: 9",
                                    end: "line: 20, column: 27",
                                    source: "f = (x,y) => x + y",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [
                    Diagnostic {
                        msg: "parameter headline must start with \"{parameter_name}:\"",
                        loc: SourceLocation {
                            start: "line: 20, column: 9",
                            end: "line: 20, column: 27",
                            source: "f = (x,y) => x + y",
                        },
                    },
                    Diagnostic {
                        msg: "parameter headline must start with \"{parameter_name}:\"",
                        loc: SourceLocation {
                            start: "line: 20, column: 9",
                            end: "line: 20, column: 27",
                            source: "f = (x,y) => x + y",
                        },
                    },
                    Diagnostic {
                        msg: "missing documentation for parameter \"x\"",
                        loc: SourceLocation {
                            start: "line: 20, column: 9",
                            end: "line: 20, column: 27",
                            source: "f = (x,y) => x + y",
                        },
                    },
                    Diagnostic {
                        msg: "missing documentation for parameter \"y\"",
                        loc: SourceLocation {
                            start: "line: 20, column: 9",
                            end: "line: 20, column: 27",
                            source: "f = (x,y) => x + y",
                        },
                    },
                ],
            )
        "#]].assert_debug_eq(&docs);
    }
    #[test]
    fn test_function_doc_missing_description() {
        let src = "
        // Package foo does a thing.
        package foo

        // f is a function.
        f = (x) => x
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "f": Function(
                            FunctionDoc {
                                name: "f",
                                headline: "f is a function.",
                                description: None,
                                parameters: [],
                                flux_type: "(x: A) => A",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 6, column: 9",
                                    end: "line: 6, column: 21",
                                    source: "f = (x) => x",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [
                    Diagnostic {
                        msg: "missing documentation for parameter \"x\"",
                        loc: SourceLocation {
                            start: "line: 6, column: 9",
                            end: "line: 6, column: 21",
                            source: "f = (x) => x",
                        },
                    },
                ],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_function_doc_missing_parameter() {
        let src = "
        // Package foo does a thing.
        package foo

        // add is a function.
        //
        // ## Parameters
        // - x: is any value.
        add = (x,y) => x + y
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "add": Function(
                            FunctionDoc {
                                name: "add",
                                headline: "add is a function.",
                                description: None,
                                parameters: [
                                    ParameterDoc {
                                        name: "x",
                                        headline: "x: is any value.",
                                        description: None,
                                        required: true,
                                    },
                                ],
                                flux_type: "(x: A, y: A) => A where A: Addable",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 9, column: 9",
                                    end: "line: 9, column: 29",
                                    source: "add = (x,y) => x + y",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [
                    Diagnostic {
                        msg: "missing documentation for parameter \"y\"",
                        loc: SourceLocation {
                            start: "line: 9, column: 9",
                            end: "line: 9, column: 29",
                            source: "add = (x,y) => x + y",
                        },
                    },
                ],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_function_doc_missing_pipe_parameter() {
        let src = "
        // Package foo does a thing.
        package foo

        // add is a function.
        //
        // ## Parameters
        // - x: is any value.
        add = (x,y=<-) => x + y
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "add": Function(
                            FunctionDoc {
                                name: "add",
                                headline: "add is a function.",
                                description: None,
                                parameters: [
                                    ParameterDoc {
                                        name: "x",
                                        headline: "x: is any value.",
                                        description: None,
                                        required: true,
                                    },
                                ],
                                flux_type: "(<-y: A, x: A) => A where A: Addable",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 9, column: 9",
                                    end: "line: 9, column: 32",
                                    source: "add = (x,y=<-) => x + y",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [
                    Diagnostic {
                        msg: "missing documentation for parameter \"y\"",
                        loc: SourceLocation {
                            start: "line: 9, column: 9",
                            end: "line: 9, column: 32",
                            source: "add = (x,y=<-) => x + y",
                        },
                    },
                ],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_function_doc_missing_optional_parameter() {
        let src = "
        // Package foo does a thing.
        package foo

        // add is a function.
        //
        // ## Parameters
        // - x: is any value.
        add = (x,y=1) => x + y
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "add": Function(
                            FunctionDoc {
                                name: "add",
                                headline: "add is a function.",
                                description: None,
                                parameters: [
                                    ParameterDoc {
                                        name: "x",
                                        headline: "x: is any value.",
                                        description: None,
                                        required: true,
                                    },
                                ],
                                flux_type: "(x: A, ?y: A) => A where A: Addable",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 9, column: 9",
                                    end: "line: 9, column: 31",
                                    source: "add = (x,y=1) => x + y",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [
                    Diagnostic {
                        msg: "missing documentation for parameter \"y\"",
                        loc: SourceLocation {
                            start: "line: 9, column: 9",
                            end: "line: 9, column: 31",
                            source: "add = (x,y=1) => x + y",
                        },
                    },
                ],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_function_doc_extra_parameter() {
        let src = "
        // Package foo does a thing.
        package foo

        // one is a function.
        //
        // ## Parameters
        // - x: is any value.
        one = () => 1
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "one": Function(
                            FunctionDoc {
                                name: "one",
                                headline: "one is a function.",
                                description: None,
                                parameters: [
                                    ParameterDoc {
                                        name: "x",
                                        headline: "x: is any value.",
                                        description: None,
                                        required: false,
                                    },
                                ],
                                flux_type: "() => int",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 9, column: 9",
                                    end: "line: 9, column: 22",
                                    source: "one = () => 1",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [
                    Diagnostic {
                        msg: "extra documentation for parameter \"x\"",
                        loc: SourceLocation {
                            start: "line: 9, column: 9",
                            end: "line: 9, column: 22",
                            source: "one = () => 1",
                        },
                    },
                ],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_function_no_parameters() {
        let src = "
        // Package foo does a thing.
        package foo

        // one returns the number one.
        one = () => 1
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "one": Function(
                            FunctionDoc {
                                name: "one",
                                headline: "one returns the number one.",
                                description: None,
                                parameters: [],
                                flux_type: "() => int",
                                is_option: false,
                                source_location: SourceLocation {
                                    start: "line: 6, column: 9",
                                    end: "line: 6, column: 22",
                                    source: "one = () => 1",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_value_option() {
        let src = "
        // Package foo does a thing.
        package foo

        // one is the number one.
        option one = 1
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "one": Value(
                            ValueDoc {
                                name: "one",
                                headline: "one is the number one.",
                                description: None,
                                flux_type: "int",
                                is_option: true,
                                source_location: SourceLocation {
                                    start: "line: 6, column: 9",
                                    end: "line: 6, column: 23",
                                    source: "option one = 1",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_function_option() {
        let src = "
        // Package foo does a thing.
        package foo

        // one returns the number one.
        option one = () => 1
        ";
        let docs = assert_docs_full(src);
        expect![[r#"
            (
                PackageDoc {
                    path: "path",
                    name: "foo",
                    headline: "Package foo does a thing.",
                    description: None,
                    members: {
                        "one": Function(
                            FunctionDoc {
                                name: "one",
                                headline: "one returns the number one.",
                                description: None,
                                parameters: [],
                                flux_type: "() => int",
                                is_option: true,
                                source_location: SourceLocation {
                                    start: "line: 6, column: 9",
                                    end: "line: 6, column: 29",
                                    source: "option one = () => 1",
                                },
                                examples: [],
                                metadata: None,
                            },
                        ),
                    },
                    examples: [],
                    metadata: None,
                },
                [],
            )
        "#]]
        .assert_debug_eq(&docs);
    }
    #[test]
    fn test_parser_headline() {
        let src = r#"

This is the first paragraph.
It has multiple sentences.
Each on their own line.
But it is still a single paragraph.


"#;
        assert_parser(src, vec![Token::Headline(src.trim())]);
    }
    #[test]
    fn test_parser_headline_and_description() {
        let src = r#"This is the headline.

This is the description.
"#;
        assert_parser(
            src,
            vec![
                Token::Headline("This is the headline."),
                Token::Description("This is the description."),
            ],
        );
    }
    #[test]
    fn test_parser_parameters() {
        let src = r#"
This is the headline.

This is the description.

## Parameters

- this is _parameter_ 1.

    Description of one.

- this is parameter 2.

More description of function.

"#;
        assert_parser(
            src,
            vec![
                Token::Headline("This is the headline."),
                Token::Description("This is the description."),
                Token::Parameters,
                Token::Parameter,
                Token::ParamHeadline("this is _parameter_ 1."),
                Token::ParamDescription("Description of one."),
                Token::Parameter,
                Token::ParamHeadline("this is parameter 2."),
                Token::Description("More description of function."),
            ],
        );
    }

    #[test]
    fn test_parser_examples() {
        let src = r#"
This is the headline.

This is the description.

## Examples

### Example 1

Subtraction:

```
3 - 2
```

### Example 2

Addition:

```
1 + 1
```


"#;
        assert_parser(
            src,
            vec![
                Token::Headline("This is the headline."),
                Token::Description("This is the description."),
                Token::Examples,
                Token::ExampleTitle("### Example 1"),
                Token::ExampleContent(
                    r#"Subtraction:

```
3 - 2
```"#,
                ),
                Token::ExampleTitle("### Example 2"),
                Token::ExampleContent(
                    r#"Addition:

```
1 + 1
```"#,
                ),
            ],
        );
    }
    #[test]
    fn test_parser_metadata() {
        let src = r#"
This is the headline.

This is the description.

## Metadata

k1: v1
k2: v2
k3: v3
"#;
        assert_parser(
            src,
            vec![
                Token::Headline("This is the headline."),
                Token::Description("This is the description."),
                Token::Metadata,
                Token::MetadataLine("k1: v1"),
                Token::MetadataLine("k2: v2"),
                Token::MetadataLine("k3: v3"),
            ],
        );
    }
    #[test]
    fn test_parser_all() {
        let src = r#"
This is the headline.

This is the description.

## Parameters

- this is _parameter_ 1.

    Description of one.

- this is parameter 2.

    Description of two.


## Examples

### Example 1

Subtraction:

```
3 - 2
```

### Example 2

Addition:

```
1 + 1
```

## Metadata

k1: v1
k2: v2
k3: v3
"#;
        assert_parser(
            src,
            vec![
                Token::Headline("This is the headline."),
                Token::Description("This is the description."),
                Token::Parameters,
                Token::Parameter,
                Token::ParamHeadline("this is _parameter_ 1."),
                Token::ParamDescription("Description of one."),
                Token::Parameter,
                Token::ParamHeadline("this is parameter 2."),
                Token::ParamDescription("Description of two."),
                Token::Examples,
                Token::ExampleTitle("### Example 1"),
                Token::ExampleContent(
                    r#"Subtraction:

```
3 - 2
```"#,
                ),
                Token::ExampleTitle("### Example 2"),
                Token::ExampleContent(
                    r#"Addition:

```
1 + 1
```"#,
                ),
                Token::Metadata,
                Token::MetadataLine("k1: v1"),
                Token::MetadataLine("k2: v2"),
                Token::MetadataLine("k3: v3"),
            ],
        );
    }
}
