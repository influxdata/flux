//! Generate documentation from source code comments.

use lazy_static::lazy_static;
use pulldown_cmark::{Event, OffsetIter, Parser as MarkdownParser, Tag};
use regex::Regex;
use std::collections::BTreeMap;
use std::ops::Range;

use crate::{
    ast,
    semantic::{
        env::Environment,
        types::{Function, MonoType, PolyType},
    },
};

use anyhow::{bail, Result};
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
    /// any Metadata associated with the package
    pub metadata: Option<Metadata>,
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
    /// indicates if this value is a Flux option
    pub is_option: bool,
    /// the location in the source code of the value
    pub source_location: ast::SourceLocation,
    /// any Metadata associated with the value
    pub metadata: Option<Metadata>,
}

/// FunctionDoc represents the documentation for a single Function within a package.
#[derive(PartialEq, Debug, Serialize, Deserialize)]
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
    /// any Metadata associated with the function
    pub metadata: Option<Metadata>,
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

/// Parse the package documentation for all values within the package.
/// The list of diagnostics reports problems found with formatting or otherwise of the comments.
/// An empty list of diagnostics implies that doc comments are all property formatted.
pub fn parse_package_doc_comments(
    pkg: &ast::Package,
    pkgpath: &str,
    types: &Environment,
) -> Result<(PackageDoc, Diagnostics)> {
    // TODO(nathanielc): Support package with more than one file.
    parse_file_doc_comments(&pkg.files[0], pkgpath, types)
}

const PACKAGE_LIT: &str = "Package";

fn parse_file_doc_comments(
    file: &ast::File,
    pkgpath: &str,
    types: &Environment,
) -> Result<(PackageDoc, Diagnostics)> {
    let mut diagnostics: Diagnostics = Vec::new();
    let (name, headline, description) = match &file.package {
        Some(pkg_clause) => {
            let comment = comments_to_string(&pkg_clause.base.comments);
            let hd = parse_headline_desc(&comment)?;
            if hd.headline.is_empty() {
                diagnostics.push(Diagnostic {
                    msg: format!(
                        "package {} must contain a non empty package comment",
                        pkgpath
                    ),
                    loc: pkg_clause.base.location.clone(),
                });
            }
            let name = pkg_clause.name.name.clone();
            let words = two_words(hd.headline.as_str());
            let start = format!("{} {}", PACKAGE_LIT, name);
            if start != words {
                diagnostics.push(Diagnostic {
                    msg: format!(
                        "package headline must start with \"{}\" found \"{}\"",
                        start, words
                    ),
                    loc: pkg_clause.base.location.clone(),
                })
            }
            (name, hd.headline, hd.description)
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
    let (description, metadata, mut diags) = parse_metadata(description, &file.base.location);
    diagnostics.append(&mut diags);
    Ok((
        PackageDoc {
            path: pkgpath.to_string(),
            name,
            headline,
            description,
            members,
            metadata,
        },
        diagnostics,
    ))
}

// Generates docs for the values in a given source file.
fn parse_package_values(
    f: &ast::File,
    pkgtypes: &Environment,
) -> Result<(BTreeMap<String, Doc>, Diagnostics)> {
    let mut members: BTreeMap<String, Doc> = BTreeMap::new();
    let mut diagnostics: Diagnostics = Vec::new();
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
                let (doc, mut diags) = parse_value(&name, &comment, typ, loc, is_option)?;
                diagnostics.append(&mut diags);
                members.insert(name.clone(), doc);
            } else {
                bail!("type of value {} not found in environment", &name);
            }
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
fn parse_headline_desc(comment: &str) -> Result<HeadlineDesc> {
    if comment.is_empty() {
        return Ok(HeadlineDesc {
            headline: "".to_string(),
            description: None,
        });
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
fn headline_range(parser: &mut OffsetIter) -> Result<Range<usize>> {
    let mut range = Range::<usize> { start: 0, end: 0 };
    // We will either have a paragraph or a single text node
    match parser.next() {
        Some((Event::Start(Tag::Paragraph), r)) => {
            range.start = r.start;
        }
        Some((Event::Text(_), r)) => return Ok(r),
        Some((Event::Code(_), r)) => return Ok(r),
        Some(e) => bail!(
            "headline does not start with paragraph or text found {:?}",
            e
        ),
        None => bail!("headline does not start with paragraph or text found EOF"),
    };
    // We have a paragraph so gather all events until the end of the paragraph.
    loop {
        match parser.next() {
            Some((Event::End(Tag::Paragraph), r)) => {
                range.end = r.end;
                return Ok(range);
            }
            //do nothing but catch the event
            Some(_) => {}
            None => {
                bail!("reached end of markdown without reaching end of paragraph")
            }
        }
    }
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

// finds a the next heading 2 within the markdown that has the provided name.
fn find_heading_range(parser: &mut OffsetIter, heading: &str) -> Option<Range<usize>> {
    loop {
        match parser.next() {
            Some((Event::Start(Tag::Heading(2)), range)) => {
                if let Some((Event::Text(t), _)) = parser.next() {
                    if heading == &*t {
                        return Some(range);
                    }
                }
            }
            Some(_) => {}
            None => return None,
        }
    }
}

fn parse_value(
    name: &str,
    comment: &str,
    typ: &PolyType,
    loc: &ast::SourceLocation,
    is_option: bool,
) -> Result<(Doc, Diagnostics)> {
    match &typ.expr {
        MonoType::Fun(f) => {
            let (doc, diags) = parse_function_doc(name, comment, typ, f, loc, is_option)?;
            Ok((Doc::Function(Box::new(doc)), diags))
        }
        _ => {
            let (doc, diags) = parse_value_doc(name, comment, typ, loc, is_option)?;
            Ok((Doc::Value(Box::new(doc)), diags))
        }
    }
}

const PARAMETER_HEADING: &str = "Parameters";

fn parse_function_doc(
    name: &str,
    comment: &str,
    typ: &PolyType,
    fun_typ: &Function,
    loc: &ast::SourceLocation,
    is_option: bool,
) -> Result<(FunctionDoc, Diagnostics)> {
    let mut diagnostics: Diagnostics = Vec::new();
    let hd = parse_headline_desc(comment)?;
    if hd.headline.is_empty() {
        diagnostics.push(Diagnostic {
            msg: format!("function \"{}\" must contain a non empty comment", name),
            loc: loc.clone(),
        });
    }
    if let Some(diagnostic) = check_headline(name, &hd.headline, loc) {
        diagnostics.push(diagnostic)
    }

    let mut parameters: Vec<ParameterDoc> = Vec::new();
    let description = match hd.description {
        Some(description) => {
            let mut parser = MarkdownParser::new(&description).into_offset_iter();
            if let Some(heading_range) = find_heading_range(&mut parser, PARAMETER_HEADING) {
                let (params, range, mut diags) =
                    parse_function_parameter_list(&mut parser, fun_typ, &description, loc)?;
                let parameter_range = Range {
                    start: heading_range.start,
                    end: range.end,
                };
                diagnostics.append(&mut diags);
                parameters = params;
                diagnostics.append(&mut diags);
                // Validate all parameters were documented
                let params_on_type: Vec<&String> =
                    fun_typ.req.keys().chain(fun_typ.opt.keys()).collect();
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
                    if !param.name.is_empty()
                        && !params_on_type.iter().any(|&name| name == &param.name)
                    {
                        diagnostics.push(Diagnostic {
                            msg: format!("extra documentation for parameter \"{}\"", param.name,),
                            loc: loc.clone(),
                        });
                    }
                }
                // Return the description with the parameter list removed
                Some(
                    description
                        .chars()
                        .take(parameter_range.start)
                        .chain(description.chars().skip(parameter_range.end))
                        .collect(),
                )
            } else {
                // Its possible the parameter list was not found or was invalid.
                // In such cases a diagnostic would have been reported, so just return the
                // description unmodified.
                Some(description)
            }
        }
        None => {
            // A description is not necessary if there are no parameters.
            if !fun_typ.req.is_empty() || !fun_typ.opt.is_empty() || fun_typ.pipe.is_some() {
                diagnostics.push(Diagnostic {
                    msg: format!("function \"{}\" comment must contain a description", name),
                    loc: loc.clone(),
                });
            }
            None
        }
    };
    let (description, metadata, mut diags) = parse_metadata(description, loc);
    diagnostics.append(&mut diags);
    Ok((
        FunctionDoc {
            name: name.to_string(),
            headline: hd.headline,
            description,
            parameters,
            flux_type: format!("{}", &typ.normal()),
            is_option,
            source_location: loc.clone(),
            metadata,
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
) -> Result<(Vec<ParameterDoc>, Range<usize>, Diagnostics)> {
    let mut parameters: Vec<ParameterDoc> = Vec::new();
    let mut diagnostics: Diagnostics = Vec::new();
    loop {
        match parser.next() {
            Some((Event::Start(Tag::Item), _)) => {
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
) -> Result<(ParameterDoc, Diagnostics)> {
    let mut diagnostics: Diagnostics = Vec::new();
    // Parse headline
    let headline_range = headline_range(parser)?;
    let headline = content[headline_range.clone()].to_string();
    let mut name = String::new();
    let word = first_word(headline.as_str());
    if let Some(n) = word.strip_suffix(':') {
        name = n.to_string();
    } else {
        diagnostics.push(Diagnostic {
            msg: "parameter headline must start with \"<parameter name>:\"".to_string(),
            loc: loc.clone(),
        });
    }
    if headline.is_empty() {
        diagnostics.push(Diagnostic {
            msg: "parameter list entry does not contain a headline".to_string(),
            loc: loc.clone(),
        });
    }
    // The rest of the list item is the description.
    let desc_range = find_item_end(parser, headline_range.end)?;
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

fn find_item_end(parser: &mut OffsetIter, start: usize) -> Result<Range<usize>> {
    let mut desc_range = Range::<usize> { start, end: 0 };
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
                    return Ok(desc_range);
                }
            }
            // Consume all other events.
            Some(_) => {}
            None => {
                bail!("reached end of markdown without reaching end of item")
            }
        }
    }
}

fn parse_value_doc(
    name: &str,
    comment: &str,
    typ: &PolyType,
    loc: &ast::SourceLocation,
    is_option: bool,
) -> Result<(ValueDoc, Diagnostics)> {
    let mut diagnostics: Diagnostics = Vec::new();
    let hd = parse_headline_desc(comment)?;
    if hd.headline.is_empty() {
        diagnostics.push(Diagnostic {
            msg: format!("value {} must contain a non empty comment", name),
            loc: loc.clone(),
        });
    }
    if let Some(diagnostic) = check_headline(name, &hd.headline, loc) {
        diagnostics.push(diagnostic)
    }
    let (description, metadata, mut diags) = parse_metadata(hd.description, loc);
    diagnostics.append(&mut diags);
    Ok((
        ValueDoc {
            name: name.to_string(),
            headline: hd.headline,
            description,
            flux_type: format!("{}", &typ.normal()),
            is_option,
            source_location: loc.clone(),
            metadata,
        },
        diagnostics,
    ))
}

const METADATA_HEADING: &str = "Metadata";

// parses 'key: value' data from the end of a string returning the
// unused beginning of the string, the metadata and any diagnostics.
// Metadata begins after a ## Metadata heading is found.
fn parse_metadata(
    content: Option<String>,
    loc: &ast::SourceLocation,
) -> (Option<String>, Option<Metadata>, Diagnostics) {
    lazy_static! {
        static ref KEY_VALUE_PATTERN: Regex = Regex::new("^(\\w[\\w_]+): (.+)$").unwrap();
    }
    let mut diagnostics: Diagnostics = Vec::new();
    if let Some(content) = content {
        let mut parser = MarkdownParser::new(&content).into_offset_iter();
        let mut description_range: Range<usize> = Range {
            start: 0,
            end: content.len(),
        };
        // Find beginning of metadata
        if let Some(heading_range) = find_heading_range(&mut parser, METADATA_HEADING) {
            description_range.end = heading_range.start;
            let mut meta = Metadata::new();
            for line in content[heading_range.end..]
                .lines()
                .take_while(|l| l.is_empty() || KEY_VALUE_PATTERN.is_match(l))
            {
                for cap in KEY_VALUE_PATTERN.captures_iter(line) {
                    if meta.contains_key(&cap[1]) {
                        diagnostics.push(Diagnostic {
                            msg: format!("found duplicate metadata key \"{}\"", &cap[1]),
                            loc: loc.clone(),
                        });
                    };
                    meta.insert(cap[1].to_string(), cap[2].to_string());
                }
            }
            if !meta.is_empty() {
                (
                    Some(content[description_range].to_string()),
                    Some(meta),
                    diagnostics,
                )
            } else {
                (Some(content), None, diagnostics)
            }
        } else {
            (Some(content), None, diagnostics)
        }
    } else {
        (None, None, diagnostics)
    }
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
    for (_, m) in doc.members.iter_mut() {
        remove_desc(m);
    }
}

/// Removes the description from a Doc.
///
/// This function is recursive via the [`shorten`] function.
/// This design allows the implementation for the Doc::Package variant to share code with
/// [`shorten`] and keep the original data types as &mut instead of moving the data into these
/// functions.
fn remove_desc(doc: &mut Doc) {
    match doc {
        Doc::Package(p) => shorten(p),
        Doc::Value(v) => {
            v.description = None;
        }
        Doc::Function(f) => {
            f.description = None;
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

#[cfg(test)]
mod test {
    use super::{
        parse_package_doc_comments, shorten, Diagnostic, Diagnostics, Doc, FunctionDoc, PackageDoc,
        ParameterDoc, ValueDoc,
    };

    use crate::{
        ast,
        ast::tests::Locator,
        parser::parse_string,
        semantic::{env::Environment, types::PolyTypeMap, Analyzer},
    };

    use std::collections::BTreeMap;

    macro_rules! map {
        ($( $key: expr => $val: expr ),*$(,)?) => {{
             let mut map = BTreeMap::default();
             $( map.insert($key.to_string(), $val); )*
             map
        }}
    }

    fn parse_program(src: &str) -> ast::Package {
        let file = parse_string("".to_string(), src);

        ast::Package {
            base: file.base.clone(),
            path: "path".to_string(),
            package: "main".to_string(),
            files: vec![file],
        }
    }
    fn assert_docs_full(src: &str, pkg: PackageDoc, diags: Diagnostics) {
        assert_docs(src, pkg, diags, false)
    }
    fn assert_docs_short(src: &str, pkg: PackageDoc, diags: Diagnostics) {
        assert_docs(src, pkg, diags, true)
    }
    fn assert_docs(src: &str, pkg: PackageDoc, diags: Diagnostics, short: bool) {
        let mut analyzer =
            Analyzer::new_with_defaults(Environment::empty(true), PolyTypeMap::new());
        let ast_pkg = parse_program(src);
        let (types, _) = match analyzer.analyze_ast(ast_pkg.clone()) {
            Ok(t) => t,
            Err(e) => panic!("error inferring types {}", e),
        };
        let (mut got_pkg, got_diags) = match parse_package_doc_comments(&ast_pkg, "path", &types) {
            Ok((p, d)) => (p, d),
            Err(e) => panic!("error parsing doc comments {}", e),
        };
        if short {
            shorten(&mut got_pkg);
        }
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
        // Package foo does a thing.
        package foo
        ";
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: BTreeMap::default(),
                metadata: None,
            },
            vec![],
        );
    }
    #[test]
    fn test_package_headline_invalid() {
        let src = "
        // foo does a thing.
        package foo
        ";
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "foo does a thing.\n".to_string(),
                description: None,
                members: BTreeMap::default(),
                metadata: None,
            },
            vec![Diagnostic {
                msg: "package headline must start with \"Package foo\" found \"foo does\""
                    .to_string(),
                loc: loc.get(3, 9, 3, 20),
            }],
        );
    }
    #[test]
    fn test_value_doc_no_desc() {
        let src = "
        // Package foo does a thing.
        package foo

        // a is a constant.
        a = 1
        ";
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: map![
                    "a" => Doc::Value(Box::new(ValueDoc{
                        name: "a".to_string(),
                        headline: "a is a constant.\n".to_string(),
                        description: None,
                        flux_type: "int".to_string(),
                        is_option: false,
                        source_location: loc.get(6,9,6,14),
                        metadata: None,
                    })),
                ],
                metadata: None,
            },
            vec![],
        );
    }
    #[test]
    fn test_value_doc_headline_invalid() {
        let src = "
        // Package foo does a thing.
        package foo

        // A is a constant.
        a = 1
        ";
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: map![
                    "a" => Doc::Value(Box::new(ValueDoc{
                        name: "a".to_string(),
                        headline: "A is a constant.\n".to_string(),
                        description: None,
                        flux_type: "int".to_string(),
                        is_option: false,
                        source_location: loc.get(6,9,6,14),
                        metadata: None,
                    })),
                ],
                metadata: None,
            },
            vec![Diagnostic {
                msg: "headline must start with \"a\" found \"A\"".to_string(),
                loc: loc.get(6, 9, 6, 14),
            }],
        );
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
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: map![
                    "a" => Doc::Value(Box::new(ValueDoc{
                        name: "a".to_string(),
                        headline: "a is a constant.\nThe value is one.\n".to_string(),
                        description: Some("\nThis is the start of the description.\n\nThe description contains any remaining markdown content.\n".to_string()),
                        flux_type: "int".to_string(),
                        is_option: false,
                        source_location: loc.get(11,9,11,14),
                        metadata: None,
                    })),
                ],
                metadata: None,
            },
            vec![],
        );
    }
    #[test]
    fn test_shorten() {
        let src = "
        // Package foo does a thing.
        //
        // This is a description.
        package foo

        // a is a constant.
        //
        // This is a description.
        //
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
        f = (x) => 1

        // o is an option.
        //
        // This is a description.
        option o = 1
        ";
        let loc = Locator::new(&src[..]);
        assert_docs_short(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: map![
                    "a" => Doc::Value(Box::new(ValueDoc{
                        name: "a".to_string(),
                        headline: "a is a constant.\n".to_string(),
                        description: None,
                        flux_type: "int".to_string(),
                        is_option: false,
                        source_location: loc.get(11,9,11,14),
                        metadata: None,
                    })),
                    "f" => Doc::Function(Box::new(FunctionDoc{
                        name: "f".to_string(),
                        headline: "f is a function.\n".to_string(),
                        description: None,
                        parameters: vec![ParameterDoc{
                            name: "x".to_string(),
                            headline: "x: is a parameter.\n".to_string(),
                            description: None,
                            required: true,
                        } ],
                        flux_type: "(x:A) => int".to_string(),
                        is_option: false,
                        source_location: loc.get(23,9,23,21),
                        metadata: None,
                    })),
                    "o" => Doc::Value(Box::new(ValueDoc{
                        name: "o".to_string(),
                        headline: "o is an option.\n".to_string(),
                        description: None,
                        flux_type: "int".to_string(),
                        is_option: true,
                        source_location: loc.get(28,9,28,21),
                        metadata: None,
                    })),
                ],
                metadata: None,
            },
            vec![],
        );
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
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: Some("\nThis is a description.\n\n".to_string()),
                members: map![
                    "a" => Doc::Value(Box::new(ValueDoc{
                        name: "a".to_string(),
                        headline: "a is a constant.\n".to_string(),
                        description: Some("\nThis is a description.\n\n".to_string()),
                        flux_type: "int".to_string(),
                        is_option: false,
                        source_location: loc.get(20,9,20,14),
                        metadata: Some(map![
                            "k3" => "v3".to_string(),
                            "k4" => "v4".to_string(),
                            "k5" => "v5".to_string(),
                        ]),
                    })),
                    "f" => Doc::Function(Box::new(FunctionDoc{
                        name: "f".to_string(),
                        headline: "f is a function.\n".to_string(),
                        description: Some("\nThis is a description.\n\n".to_string()),
                        parameters: vec![ParameterDoc{
                            name: "x".to_string(),
                            headline: "x: is a parameter.\n".to_string(),
                            description: Some("\n    This is a description of x.\n\n".to_string()),
                            required: true,
                        } ],
                        flux_type: "(x:A) => int".to_string(),
                        is_option: false,
                        source_location: loc.get(36,9,36,21),
                        metadata: Some(map![
                            "k6" => "v6".to_string(),
                            "k7" => "v7".to_string(),
                            "k8" => "v8".to_string(),
                        ]),
                    })),
                    "o" => Doc::Value(Box::new(ValueDoc{
                        name: "o".to_string(),
                        headline: "o is an option.\n".to_string(),
                        description: Some("\nThis is a description.\n\n".to_string()),
                        flux_type: "int".to_string(),
                        is_option: true,
                        source_location: loc.get(45,9,45,21),
                        metadata: Some(map![
                            "k9" => "v9".to_string(),
                            "k0" => "v0".to_string(),
                        ]),
                    })),
                ],
                metadata: Some(map![
                    "k0" => "v0".to_string(),
                    "k1" => "v1".to_string(),
                    "k2" => "v2".to_string(),
                ]),
            },
            vec![],
        );
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
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: Some("\nThis is a description.\n\n".to_string()),
                members: BTreeMap::default(),
                metadata: Some(map![
                    "key" => "valueB".to_string(),
                    "key1" => "value with spaces".to_string(),
                    "key_with_underscores" => "value".to_string(),
                ]),
            },
            vec![Diagnostic {
                msg: "found duplicate metadata key \"key\"".to_string(),
                loc: loc.get(11, 9, 11, 20),
            }],
        );
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
        //
        // More description after the parameter list.
        f = (x) => x
        ";
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: map![
                    "f" => Doc::Function(Box::new(FunctionDoc{
                        name: "f".to_string(),
                        headline: "f is a function.\n".to_string(),
                        description: Some("\nMore specifically f is the identity function, it returns any value it is passed as a\nparameter.\n\nMore description after the parameter list.\n".to_string()),
                        parameters: vec![ParameterDoc{
                            name: "x".to_string(),
                            headline: "x: is any value.".to_string(),
                            description: None,
                            required: true,
                        }],
                        flux_type: "(x:A) => A".to_string(),
                        is_option: false,
                        source_location: loc.get(14,9,14,21),
                        metadata: None,
                    })),
                ],
                metadata: None,
            },
            vec![],
        );
    }
    #[test]
    fn test_function_headline_invalid() {
        let src = "
        // Package foo does a thing.
        package foo

        // F is a function.
        f = () => 1
        ";
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: map![
                    "f" => Doc::Function(Box::new(FunctionDoc{
                        name: "f".to_string(),
                        headline: "F is a function.\n".to_string(),
                        description: None,
                        parameters: vec![],
                        flux_type: "() => int".to_string(),
                        is_option: false,
                        source_location: loc.get(6,9,6,20),
                        metadata: None,
                    })),
                ],
                metadata: None,
            },
            vec![Diagnostic {
                msg: "headline must start with \"f\" found \"F\"".to_string(),
                loc: loc.get(6, 9, 6, 20),
            }],
        );
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
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: map![
                    "f" => Doc::Function(Box::new(FunctionDoc{
                        name: "f".to_string(),
                        headline: "f is a function.\n".to_string(),
                        description: Some("\nMore specifically f is the identity function, it returns any value it is passed as a\nparameter.\n\nMore description after the parameter list.\n".to_string()),
                        parameters: vec![ParameterDoc{
                            name: "x".to_string(),
                            headline: "x: is any value.\n".to_string(),
                            description: Some("\n   Long description of x.\n\n".to_string()),
                            required: true,
                        },
                        ParameterDoc{
                            name: "y".to_string(),
                            headline: "y: is any value.\n".to_string(),
                            description: Some("\n   Y has a long description too.\n\n".to_string()),
                            required: true,
                        }],
                        flux_type: "(x:A, y:A) => A where A: Addable".to_string(),
                        is_option: false,
                        source_location: loc.get(20,9,20,27),
                        metadata: None,
                    })),
                ],
                metadata: None,
            },
            vec![],
        );
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
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: map![
                    "f" => Doc::Function(Box::new(FunctionDoc{
                        name: "f".to_string(),
                        headline: "f is a function.\n".to_string(),
                        description: Some("\nMore specifically f is the identity function, it returns any value it is passed as a\nparameter.\n\nMore description after the parameter list.\n".to_string()),
                        parameters: vec![ParameterDoc{
                            name: "".to_string(),
                            headline: "x is any value.\n".to_string(),
                            description: Some("\n   Long description of x.\n\n".to_string()),
                            required: false,
                        },
                        ParameterDoc{
                            name: "".to_string(),
                            headline: "`y` is any value.\n".to_string(),
                            description: Some("\n   Y has a long description too.\n\n".to_string()),
                            required: false,
                        }],
                        flux_type: "(x:A, y:A) => A where A: Addable".to_string(),
                        is_option: false,
                        source_location: loc.get(20,9,20,27),
                        metadata: None,
                    })),
                ],
                metadata: None,
            },
            vec![
                Diagnostic {
                    msg: "parameter headline must start with \"<parameter name>:\"".to_string(),
                    loc: loc.get(20, 9, 20, 27),
                },
                Diagnostic {
                    msg: "parameter headline must start with \"<parameter name>:\"".to_string(),
                    loc: loc.get(20, 9, 20, 27),
                },
                Diagnostic {
                    msg: "missing documentation for parameter \"x\"".to_string(),
                    loc: loc.get(20, 9, 20, 27),
                },
                Diagnostic {
                    msg: "missing documentation for parameter \"y\"".to_string(),
                    loc: loc.get(20, 9, 20, 27),
                },
            ],
        );
    }
    #[test]
    fn test_function_doc_missing_description() {
        let src = "
        // Package foo does a thing.
        package foo

        // f is a function.
        f = (x) => x
        ";
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: map![
                    "f" => Doc::Function(Box::new(FunctionDoc{
                        name: "f".to_string(),
                        headline: "f is a function.\n".to_string(),
                        description: None,
                        parameters: vec![],
                        flux_type: "(x:A) => A".to_string(),
                        is_option: false,
                        source_location: loc.get(6, 9, 6, 21),
                        metadata: None,
                    })),
                ],
                metadata: None,
            },
            vec![Diagnostic {
                msg: "function \"f\" comment must contain a description".to_string(),
                loc: loc.get(6, 9, 6, 21),
            }],
        );
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
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: map![
                    "add" => Doc::Function(Box::new(FunctionDoc{
                        name: "add".to_string(),
                        headline: "add is a function.\n".to_string(),
                        description: Some("\n".to_string()),
                        parameters: vec![ParameterDoc{
                            name: "x".to_string(),
                            headline: "x: is any value.".to_string(),
                            description: None,
                            required: true,
                        }],
                        flux_type: "(x:A, y:A) => A where A: Addable".to_string(),
                        is_option: false,
                        source_location: loc.get(9, 9, 9, 29),
                        metadata: None,
                    })),
                ],
                metadata: None,
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
        // Package foo does a thing.
        package foo

        // add is a function.
        //
        // ## Parameters
        // - x: is any value.
        add = (x,y=1) => x + y
        ";
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: map![
                    "add" => Doc::Function(Box::new(FunctionDoc{
                        name: "add".to_string(),
                        headline: "add is a function.\n".to_string(),
                        description: Some("\n".to_string()),
                        parameters: vec![ParameterDoc{
                            name: "x".to_string(),
                            headline: "x: is any value.".to_string(),
                            description: None,
                            required: true,
                        }],
                        flux_type: "(x:int, ?y:int) => int".to_string(),
                        is_option: false,
                        source_location: loc.get(9, 9, 9, 31),
                        metadata: None,
                    })),
                ],
                metadata: None,
            },
            vec![Diagnostic {
                msg: "missing documentation for parameter \"y\"".to_string(),
                loc: loc.get(9, 9, 9, 31),
            }],
        );
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
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: map![
                    "one" => Doc::Function(Box::new(FunctionDoc{
                        name: "one".to_string(),
                        headline: "one is a function.\n".to_string(),
                        description: Some("\n".to_string()),
                        parameters: vec![ParameterDoc{
                            name: "x".to_string(),
                            headline: "x: is any value.".to_string(),
                            description: None,
                            required: false,
                        }],
                        flux_type: "() => int".to_string(),
                        is_option: false,
                        source_location: loc.get(9, 9, 9, 22),
                        metadata: None,
                    })),
                ],
                metadata: None,
            },
            vec![Diagnostic {
                msg: "extra documentation for parameter \"x\"".to_string(),
                loc: loc.get(9, 9, 9, 22),
            }],
        );
    }
    #[test]
    fn test_function_no_parameters() {
        let src = "
        // Package foo does a thing.
        package foo

        // one returns the number one.
        one = () => 1
        ";
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: map![
                    "one" => Doc::Function(Box::new(FunctionDoc{
                        name: "one".to_string(),
                        headline: "one returns the number one.\n".to_string(),
                        description: None,
                        parameters: vec![],
                        flux_type: "() => int".to_string(),
                        is_option: false,
                        source_location: loc.get(6, 9, 6, 22),
                        metadata: None,
                    })),
                ],
                metadata: None,
            },
            vec![],
        );
    }
    #[test]
    fn test_value_option() {
        let src = "
        // Package foo does a thing.
        package foo

        // one is the number one.
        option one = 1
        ";
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: map![
                    "one" => Doc::Value(Box::new(ValueDoc{
                        name: "one".to_string(),
                        headline: "one is the number one.\n".to_string(),
                        description: None,
                        flux_type: "int".to_string(),
                        is_option: true,
                        source_location: loc.get(6, 9, 6, 23),
                        metadata: None,
                    })),
                ],
                metadata: None,
            },
            vec![],
        );
    }
    #[test]
    fn test_function_option() {
        let src = "
        // Package foo does a thing.
        package foo

        // one returns the number one.
        option one = () => 1
        ";
        let loc = Locator::new(&src[..]);
        assert_docs_full(
            src,
            PackageDoc {
                path: "path".to_string(),
                name: "foo".to_string(),
                headline: "Package foo does a thing.\n".to_string(),
                description: None,
                members: map![
                    "one" => Doc::Function(Box::new(FunctionDoc{
                        name: "one".to_string(),
                        headline: "one returns the number one.\n".to_string(),
                        description: None,
                        parameters: vec![],
                        flux_type: "() => int".to_string(),
                        is_option: true,
                        source_location: loc.get(6, 9, 6, 29),
                        metadata: None,
                    })),
                ],
                metadata: None,
            },
            vec![],
        );
    }
}
