use std::env;
use std::io;

// module for all flux WASM functions
pub use crate::ast::*;
use crate::docs;
pub use crate::formatter::convert_to_string;
pub use crate::{ast, find_var_type};
pub use fluxcore::parser::Parser;
use fluxcore::semantic::bootstrap;
pub use fluxcore::semantic::types::{MonoType, Tvar};
pub use wasm_bindgen::prelude::*;

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
pub fn get_json_documentation(flux_path: &str) -> JsValue {
    let d = docs();
    let mut doc = JsValue::UNDEFINED;

    for i in &d {
        // look for the given identifier
        if flux_path == i.path {
            // return that doc package
            let param = serde_json::to_string(i).unwrap();
            doc = JsValue::from_serde(&param).unwrap();
            break;
        }
    }

    doc
}

/// Gets json docs for all Stdlib
#[wasm_bindgen]
pub fn get_all_json() -> JsValue {
    let d = docs();
    let param = serde_json::to_string(&d).unwrap();
    let doc = JsValue::from_serde(&param).unwrap();
    doc
}

#[cfg(test)]
mod tests {
    use super::*;
    use wasm_bindgen_test::*;

    #[wasm_bindgen_test]
    pub fn json_docs_test() {
        let csv_doc = r#"{"path":"csv","name":"csv","headline":"Package csv provides tools for working with data in annotated CSV format.","description":null,"members":{"from":{"kind":"Function","name":"from","headline":"from is a function that retrieves data from a comma separated value (CSV) data source. ","description":"A stream of tables are returned, each unique series contained within its own table. Each record in the table represents a single point in the series. ## Query anotated CSV data from file\n```\nimport \"csv\"\n\ncsv.from(file: \"path/to/data-file.csv\")\n```\n\n## Query raw data from CSV file\n```\nimport \"csv\"\n\ncsv.from(\n  file: \"/path/to/data-file.csv\",\n  mode: \"raw\"\n)\n```\n\n## Query an annotated CSV string\n```\nimport \"csv\"\n\ncsvData = \"\n#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double\n#group,false,false,false,false,false,false,false,false\n#default,,,,,,,,\n,result,table,_start,_stop,_time,region,host,_value\n,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43\n,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25\n,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62\n\"\n\ncsv.from(csv: csvData)\n\n```\n\n## Query a raw CSV string\n```\nimport \"csv\"\n\ncsvData = \"\n_start,_stop,_time,region,host,_value\n2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43\n2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25\n2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62\n\"\n\ncsv.from(\n  csv: csvData,\n  mode: \"raw\"\n)\n```\n\n","parameters":[{"name":"csv","headline":" is CSV data.","description":"Supports anonotated CSV or raw CSV. Use mode to specify the parsing mode.","required":false},{"name":"file","headline":" is the file path of the CSV file to query.","description":"The path can be absolute or relative. If relative, it is relative to the working directory of the  fluxd  process. The CSV file must exist in the same file system running the  fluxd  process.","required":false},{"name":"mode","headline":" is the CSV parsing mode. Default is annotations.","description":"Available annotation modes: annotations: Use CSV notations to determine column data types. raw: Parse all columns as strings and use the first row as the header row and all subsequent rows as data.","required":false}],"flux_type":"(?csv:string, ?file:string, ?mode:string) => [t8500]","link":"https://docs.influxdata.com/influxdb/cloud/reference/flux/stdlib/csv/from"}},"link":"https://docs.influxdata.com/influxdb/cloud/reference/flux/stdlib/csv"}"#;
        let want = JsValue::from_serde(csv_doc).unwrap();
        let got = get_json_documentation("csv");
        assert_eq!(want, got);
    }
}
