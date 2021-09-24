//! Generate documentation from source code comments.

use pulldown_cmark::CodeBlockKind;
use pulldown_cmark::{Event, Parser};
use std::collections::BTreeMap;

use crate::ast;
use crate::semantic::types::{MonoType, PolyTypeMapMap};

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
    /// the docs site link for a package
    pub link: String,
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
    /// the docs site link for a Value
    pub link: String,
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
    /// the docs site link for a function
    pub link: String,
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

/// Generates the docs by parsing the sources and checking type inference.
pub fn generate_docs(
    types: &PolyTypeMapMap,
    file: &ast::File,
    pkgpath: &str,
) -> Result<PackageDoc, Box<dyn std::error::Error>> {
    // construct the package documentation
    // use type inference to determine types of all values
    let mut all_comment = String::new();
    let members = generate_values(file, types, pkgpath)?;
    if Some(&file.package) != None {
        all_comment = comments_to_string(&file.package.as_ref().unwrap().base.comments);
    }
    let (headline, description) = separate_description(&all_comment);

    //TODO check if package name exists and if it doesn't throw an error message
    Ok(PackageDoc {
        path: pkgpath.to_string(),
        name: file.package.clone().unwrap().name.name,
        headline,
        description,
        members,
        link: "https://docs.influxdata.com/influxdb/cloud/reference/flux/stdlib/".to_owned()
            + &pkgpath.to_string(),
    })
}

// Separates headline from description
fn separate_description(all_comment: &str) -> (String, Option<String>) {
    let mut headline: String = "".to_string();
    let mut reached_end: bool = false;
    let mut description_text: String = "".to_string();
    let parser = Parser::new(all_comment);
    for event in parser {
        match event {
            Event::Text(t) => {
                if !reached_end {
                    headline.push_str(&t.to_string());
                } else {
                    description_text.push_str(&t.to_string());
                    if description_text.ends_with('.') {
                        description_text.push(' ');
                    }
                }
                format!("{}", t)
            }
            Event::Start(tag) => {
                format!("start: {:?}", tag)
            }
            Event::End(tag) => {
                reached_end = true;
                format!("end: {:?}", tag)
            }
            _ => "Unsupported markdown in documentation comment".to_string(),
        };
    }
    if !description_text.is_empty() {
        (headline, Option::from(description_text))
    } else {
        (headline, Option::None)
    }
}

// Separates function document parameters and returns a newly generated FuncDoc struct
fn separate_func_docs(all_doc: &str, name: &str) -> FunctionDoc {
    let mut funcdocs = FunctionDoc {
        name: name.to_string(),
        headline: String::new(),
        description: String::new(),
        parameters: Vec::new(),
        flux_type: String::new(),
        link: String::new(),
    };
    let mut tmp = &mut funcdocs.headline;
    let mut param_flag = false;

    let parser = Parser::new(all_doc);
    let events: Vec<pulldown_cmark::Event> = parser.collect();
    for (_, event) in events.windows(2).enumerate() {
        match &event[0] {
            Event::Start(pulldown_cmark::Tag::Heading(2)) => match &event[1] {
                Event::Text(t) => {
                    if "Parameters".eq(&t.to_string()) {
                        param_flag = true;
                    } else {
                        tmp.push_str("## ");
                    }
                }
                _ => {
                    param_flag = false;
                }
            },
            Event::Start(pulldown_cmark::Tag::Item) => {
                if param_flag {
                    funcdocs.parameters.push(ParameterDoc {
                        name: String::new(),
                        headline: String::new(),
                        description: None,
                        required: false,
                    });
                } else {
                    tmp.push_str(" - ");
                }
            }
            Event::Start(pulldown_cmark::Tag::CodeBlock(CodeBlockKind::Fenced(_))) => {
                tmp.push_str("\n```\n");
            }
            Event::Code(c) => {
                if param_flag {
                    let len = funcdocs.parameters.len() - 1;
                    if funcdocs.parameters[len].name.is_empty() {
                        funcdocs.parameters[len].name = c.to_string();
                    } else {
                        if funcdocs.parameters[len].headline.is_empty() {
                            funcdocs.parameters[len].headline.push_str(&c.to_string());
                        }
                        if funcdocs.parameters[len].description != None {
                            let doc = &funcdocs.parameters[len].description;
                            let x = doc.as_ref().map(|d| format!("{} {}", d, c.to_string()));
                            funcdocs.parameters[len].description = x;
                        } else {
                            funcdocs.parameters[len].description = Some(c.to_string());
                        }
                    }
                } else {
                    tmp.push_str(&c.to_string());
                }
            }
            Event::Text(t) => {
                if param_flag && !(funcdocs.parameters.is_empty()) {
                    let len = funcdocs.parameters.len() - 1;
                    if funcdocs.parameters[len].headline.is_empty() {
                        funcdocs.parameters[len].headline = t.to_string();
                        continue;
                    }
                    if funcdocs.parameters[len].description != None {
                        let doc = &funcdocs.parameters[len].description;
                        let x = doc.as_ref().map(|d| format!("{} {}", d, t.to_string()));
                        funcdocs.parameters[len].description = x;
                    } else {
                        funcdocs.parameters[len].description = Option::from(t.to_string());
                    }
                } else if !("Parameters".eq(&t.to_string())) {
                    tmp.push_str(&t.to_string());
                    if tmp.ends_with('.') {
                        tmp.push_str(&" ".to_string());
                    }
                    if let Event::End(pulldown_cmark::Tag::CodeBlock(CodeBlockKind::Fenced(_))) =
                        &event[1]
                    {
                        tmp.push_str("```\n\n");
                    }
                }
            }
            Event::End(pulldown_cmark::Tag::List(None)) => {
                if param_flag {
                    param_flag = false;
                }
            }
            Event::End(_) => {
                tmp = &mut funcdocs.description;
            }
            _ => {
                // unused event tag found. Can be safely ignored.
            }
        }
    }
    funcdocs
}

// Generates docs for the values in a given source file.
fn generate_values(
    f: &ast::File,
    types: &PolyTypeMapMap,
    pkgpath: &str,
) -> Result<BTreeMap<String, Doc>, Box<dyn std::error::Error>> {
    let mut members: BTreeMap<String, Doc> = BTreeMap::new();
    for stmt in &f.body {
        match stmt {
            ast::Statement::Variable(s) => {
                let doc = comments_to_string(&s.id.base.comments);
                let name = s.id.name.clone();
                let mut funcdoc = separate_func_docs(&doc, &name);
                let pkgtype = &types[pkgpath];
                let typ = &pkgtype[name.as_str()];
                match &typ.expr {
                    MonoType::Fun(_f) => {
                        funcdoc.flux_type = format!("{}", &typ);
                        funcdoc.link =
                            "https://docs.influxdata.com/influxdb/cloud/reference/flux/stdlib/"
                                .to_owned()
                                + &pkgpath.to_string()
                                + "/"
                                + &name.to_string();
                        members.insert(name.clone(), Doc::Function(Box::new(funcdoc)));
                    }
                    _ => {
                        let variable = ValueDoc {
                            name: name.clone(),
                            headline: funcdoc.headline,
                            description: Option::from(funcdoc.description),
                            flux_type: format!("{}", typ.normal()),
                            link:
                                "https://docs.influxdata.com/influxdb/cloud/reference/flux/stdlib/"
                                    .to_owned()
                                    + &pkgpath.to_string()
                                    + "/"
                                    + &name.to_string(),
                        };
                        members.insert(name.clone(), Doc::Value(Box::new(variable)));
                    }
                }
            }
            ast::Statement::Builtin(s) => {
                let doc = comments_to_string(&s.base.comments);
                let name = s.id.name.clone();
                let mut funcdoc = separate_func_docs(&doc, &name);
                let pkgtype = &types[pkgpath];
                let typ = &pkgtype[name.as_str()];
                match &typ.expr {
                    MonoType::Fun(_f) => {
                        funcdoc.flux_type = format!("{}", typ.normal());
                        funcdoc.link =
                            "https://docs.influxdata.com/influxdb/cloud/reference/flux/stdlib/"
                                .to_owned()
                                + &pkgpath.to_string()
                                + "/"
                                + &name.to_string();
                        members.insert(name.clone(), Doc::Function(Box::new(funcdoc)));
                    }
                    _ => {
                        let builtin = ValueDoc {
                            name: name.clone(),
                            headline: funcdoc.headline,
                            description: Option::from(funcdoc.description),
                            flux_type: format!("{}", typ),
                            link:
                                "https://docs.influxdata.com/influxdb/cloud/reference/flux/stdlib/"
                                    .to_owned()
                                    + &pkgpath.to_string()
                                    + "/"
                                    + &name.to_string(),
                        };
                        members.insert(name.clone(), Doc::Value(Box::new(builtin)));
                    }
                }
            }
            ast::Statement::Option(s) => {
                if let ast::Assignment::Variable(v) = &s.assignment {
                    let doc = comments_to_string(&s.base.comments);
                    let name = v.id.name.clone();
                    let mut funcdoc = separate_func_docs(&doc, &name);
                    let pkgtype = &types[pkgpath];
                    let typ = &pkgtype[name.as_str()];
                    match &typ.expr {
                        MonoType::Fun(_f) => {
                            funcdoc.flux_type = format!("{}", typ.normal());
                            funcdoc.link =
                                "https://docs.influxdata.com/influxdb/cloud/reference/flux/stdlib/"
                                    .to_owned()
                                    + &pkgpath.to_string()
                                    + "/"
                                    + &name.to_string();
                            members.insert(name.clone(), Doc::Function(Box::new(funcdoc)));
                        }
                        _ => {
                            let option = ValueDoc {
                                    name: name.clone(),
                                    headline: funcdoc.headline,
                                    description: Option::from(funcdoc.description),
                                    flux_type: format!("{}", typ),
                                    link: "https://docs.influxdata.com/influxdb/cloud/reference/flux/stdlib/".to_owned() + &pkgpath.to_string() + "/" + &name.to_string(),
                                };
                            members.insert(name.clone(), Doc::Value(Box::new(option)));
                        }
                    }
                }
            }
            _ => {}
        }
    }
    Ok(members)
}

fn comments_to_string(comments: &[ast::Comment]) -> String {
    let mut s = String::new();
    if !comments.is_empty() {
        for c in comments {
            s.push_str(c.text.as_str().strip_prefix("//").unwrap());
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
        link: "https://docs.influxdata.com/influxdb/cloud/reference/flux/stdlib/".to_string(),
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
                link: "https://docs.influxdata.com/influxdb/cloud/reference/flux/stdlib/"
                    .to_owned()
                    + path,
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
