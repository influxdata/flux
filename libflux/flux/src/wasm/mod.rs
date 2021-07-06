use std::env;
use std::io;

// module for all flux WASM functions
pub use crate::ast::*;
pub use crate::formatter::convert_to_string;
pub use crate::{ast, find_var_type};
pub use fluxcore::parser::Parser;
use fluxcore::semantic::bootstrap;
pub use fluxcore::semantic::types::{MonoType, Tvar};
pub use wasm_bindgen::prelude::*;
use crate::docs;
//use wasm_bindgen_test::*;

#[derive(Debug)]
struct Error {
    msg: String,
}

impl From<env::VarError> for Error {
    fn from(err: env::VarError) -> Error {
        Error {
            msg: err.to_string(),
        }
    }
}

impl From<io::Error> for Error {
    fn from(err: io::Error) -> Error {
        Error {
            msg: format!("{:?}", err),
        }
    }
}

impl From<bootstrap::Error> for Error {
    fn from(err: bootstrap::Error) -> Error {
        Error { msg: err.msg }
    }
}


/// (Generated by WASM.)
#[wasm_bindgen]
pub fn parse(s: &str) -> JsValue {
    let mut p = Parser::new(s);
    let file = p.parse_file(String::from(""));

    JsValue::from_serde(&file).unwrap()
}

/// Format a JS file.
#[wasm_bindgen]
pub fn format_from_js_file(js_file: JsValue) -> String {
    if let Ok(file) = js_file.into_serde::<File>() {
        if let Ok(converted) = convert_to_string(&file) {
            return converted;
        }
    }
    "".to_string()
}

/// wasm version of the flux_find_var_type() API. Instead of returning a flat buffer that contains
/// the MonoType, it returns a JsValue。
#[wasm_bindgen]
pub fn wasm_find_var_type(source: &str, file_name: &str, var_name: &str) -> JsValue {
    let mut p = Parser::new(source);
    let pkg: ast::Package = p.parse_file(file_name.to_string()).into();
    let ty = find_var_type(pkg, var_name.to_string()).unwrap_or(MonoType::Var(Tvar(0)));
    JsValue::from_serde(&ty).unwrap()
}

/// Gets json docs from a Flux identifier
#[wasm_bindgen]
pub fn get_json_documentation(flux_identifier: String) -> Result<String, Error> {
    let d = docs();
    for i in &d {
        // look for the given identifier
        if flux_identifier == i.name {
            // return that doc package
            Ok(String::from_utf8(serde_json::to_vec(&i.values).unwrap()))
        }
    }
    //return d[flux_identifier];
    return Err(Error {
        msg: format!("Identifier given not found: {}", flux_identifier),
    });
}

#[wasm_bindgen_test]
pub fn json_docs_test() {
   let want = r#"[{"pkgpath":"csv","name":"from","doc":"<p>from is a function that retrieves data from a comma separated value (CSV) data source.</p>\n<p>A stream of tables are returned, each unique series contained within its own table.\nEach record in the table represents a single point in the series.</p>\n<h2>Parameters</h2>\n<ul>\n<li>\n<p><code>csv</code> is CSV data.</p>\n<p>Supports anonotated CSV or raw CSV. Use mode to specify the parsing mode.</p>\n</li>\n<li>\n<p><code>file</code> if the file path of the CSV file to query.</p>\n<p>The path can be absolute or relative. If relative, it is relative to the working\ndirectory of the <code>fluxd</code> process. The CSV file must exist in the same file\nsystem running the <code>fluxd</code> process.</p>\n</li>\n<li>\n<p><code>mode</code> is the CSV parsing mode. Default is annotations.</p>\n<p>Available annotation modes:\nannotations: Use CSV notations to determine column data types.\nraw: Parse all columns as strings and use the first row as the header row\nand all subsequent rows as data.</p>\n</li>\n</ul>\n<h2>Query anotated CSV data from file</h2>\n<pre><code>import &quot;csv&quot;\n\ncsv.from(file: &quot;path/to/data-file.csv&quot;)\n</code></pre>\n<h2>Query raw data from CSV file</h2>\n<pre><code>import &quot;csv&quot;\n\ncsv.from(\n  file: &quot;/path/to/data-file.csv&quot;,\n  mode: &quot;raw&quot;\n)\n</code></pre>\n<h2>Query an annotated CSV string</h2>\n<pre><code>import &quot;csv&quot;\n\ncsvData = &quot;\n#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double\n#group,false,false,false,false,false,false,false,false\n#default,,,,,,,,\n,result,table,_start,_stop,_time,region,host,_value\n,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43\n,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25\n,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62\n&quot;\n\ncsv.from(csv: csvData)\n\n</code></pre>\n<h2>Query a raw CSV string</h2>\n<pre><code>import &quot;csv&quot;\n\ncsvData = &quot;\n_start,_stop,_time,region,host,_value\n2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43\n2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25\n2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62\n&quot;\n\ncsv.from(\n  csv: csvData,\n  mode: &quot;raw&quot;\n)\n</code></pre>\n","typ":"(?bucket:string, ?bucketID:string, ?host:string, ?org:string, ?orgID:string, ?token:string) => [{A with _value:B, _time:time, _measurement:string, _field:string}]"}]"#;
   let got = get_json_documentation("CSV");
   assert_eq!(want, got);


}
