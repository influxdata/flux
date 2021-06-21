// #![cfg_attr(feature = "strict", deny(warnings))]
//
// //use std::collections::HashMap;
// use fluxcore::semantic::bootstrap::DocPackage;
// use flux::docs::{self};
// use std::fs::{self, File};
// use std::io::Write;
// use std::path::{Path, PathBuf};
// use structopt::StructOpt;
// use tera::{Context, Tera};
//
// #[macro_use]
// extern crate lazy_static;
//
// //use flux::ast;
// //use flux::docs;
// //use flux::parser::Parser;
// //use flux::semantic;
// //use flux::semantic::types::{PolyType, TvarKinds};
//
// #[derive(Debug, StructOpt)]
// struct Args {
//     // The root path of packages for which to generate documentation
//     #[structopt(parse(from_os_str), long)]
//     pkg: PathBuf,
//     // The name of the file to write the documentation JSON data
//     #[structopt(parse(from_os_str), long, default_value = "")]
//     json: PathBuf,
//     // The name of the directory into which to write the documentation html files
//     #[structopt(parse(from_os_str), long, default_value = "")]
//     html: PathBuf,
// }
//
//fn main() -> Result<(), Box<dyn std::error::Error>> {
//     let args = Args::from_args();
//     let pkg = docs::walk_pkg(&args.pkg, &args.pkg)?;
//
//     if args.json != Path::new("") {
//         let f = File::create(args.json)?;
//         serde_json::to_writer(f, &pkg)?;
//     }
//
//     if args.html != Path::new("") {
//         write_home(&args.html)?;
//         return write_html(&args.html, &pkg);
//     }
//     Ok(())
//}
fn main() {
    print!("hi");
}
//
// lazy_static! {
//     static ref TEMPLATES: Tera = {
//         let mut tera = match Tera::new("flux/templates/*.html") {
//             Ok(t) => t,
//             Err(e) => {
//                 println!("Parsing error(s): {}", e);
//                 ::std::process::exit(1);
//             }
//         };
//         tera.autoescape_on(vec!["html"]);
//         tera
//     };
// }
//
// // Write out a tree of html like this
// // pkgRoot
// //      index.html -- contains pkgRoot description and index
// //      valuea.html -- contains value A description
// //      subpkgA
// //          index.html -- Contains subpkgA index
// //          valueb.html -- contains value B description
// fn write_html(dir: &Path, pkg: &DocPackage) -> Result<(), Box<dyn std::error::Error>> {
//     let pkgdir = dir.join(&pkg.name);
//     fs::create_dir(&pkgdir)?;
//     let mut f = File::create(pkgdir.join("index.html"))?;
//     let data = TEMPLATES.render("package.html", &Context::from_serialize(&pkg)?)?;
//     f.write_all(data.as_bytes())?;
//     for v in &pkg.values {
//         let mut vf = File::create(pkgdir.join(format!("{}.html", v.name)))?;
//         let data = TEMPLATES.render("value.html", &Context::from_serialize(&v)?)?;
//         vf.write_all(data.as_bytes())?;
//     }
//     for p in &pkg.packages {
//         write_html(&pkgdir, &p)?;
//     }
//     Ok(())
// }
//
// // Render home.html template
// fn write_home(dir: &Path) -> Result<(), Box<dyn std::error::Error>> {
//     let ctx = Context::new();
//     let data = TEMPLATES.render("home.html", &ctx)?;
//     let mut f = File::create(dir.join("index.html"))?;
//     f.write_all(data.as_bytes())?;
//     Ok(())
// }
