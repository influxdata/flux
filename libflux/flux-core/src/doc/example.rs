//! Parse documentation examples for their code and execute them collecting their inputs and
//! outputs.

use std::{collections::HashMap, ops::Range};

use anyhow::{bail, Context, Result};
use csv::StringRecord;
use pad::PadStr;
use pulldown_cmark::{CodeBlockKind, Event, Parser, Tag};

use crate::{
    doc::{Doc, Example, PackageDoc, Table},
    formatter::format,
};

/// Executes Flux code producing input and output tables.
pub trait Executor {
    /// Execute the provided Flux code and return the CSV Flux results.
    fn execute(&self, code: &str) -> Result<String>;
}

/// A list of test results
pub type TestResults = Vec<Result<String>>;

/// Evaluates all examples in the documentation perfoming an inplace update of their inputs/outpts
/// and content.
pub fn evaluate_package_examples(docs: &mut PackageDoc, executor: &impl Executor) -> TestResults {
    let path = &docs.path;
    let mut results = TestResults::new();
    for example in docs.examples.iter_mut() {
        results.push(
            evaluate_example(example, executor)
                .with_context(|| format!("executing example for package {}", path))
                .map(|()| path.to_string()),
        );
    }
    for (name, doc) in docs.members.iter_mut() {
        results.extend(
            evaluate_doc_examples(doc, executor)
                .into_iter()
                .map(|result| {
                    result
                        .with_context(|| format!("executing example for {}.{}", path, name))
                        .map(|title| format!("{}.{}.{}", path, name, title))
                }),
        );
    }

    results
}

fn evaluate_doc_examples(doc: &mut Doc, executor: &impl Executor) -> TestResults {
    match doc {
        Doc::Package(pkg) => evaluate_package_examples(pkg, executor),
        Doc::Value(v) => v
            .examples
            .iter_mut()
            .map(|example| evaluate_example(example, executor).map(|()| example.title.clone()))
            .collect(),
        Doc::Function(f) => f
            .examples
            .iter_mut()
            .map(|example| {
                evaluate_example(example, executor)
                    .with_context(|| format!("executing `{}`", example.title))
                    .map(|()| example.title.clone())
            })
            .collect(),
    }
}

fn evaluate_example(example: &mut Example, executor: &impl Executor) -> Result<()> {
    let blocks = preprocess_code_blocks(&example.content)?;
    if blocks.len() > 1 {
        bail!(
            "examples must contain at most one Flux code block, found {}",
            blocks.len()
        )
    }
    for b in blocks {
        example.content.replace_range(b.range, &b.display);
        match b.mode {
            BlockMode::Run => {
                let results = executor.execute(&b.exec)?;
                let (input, output) = parse_results(&results)?;
                example.input = input;
                example.output = output;
            }
            BlockMode::NoRun => {}
        }
    }
    Ok(())
}

enum BlockMode {
    Run,
    NoRun,
}

struct CodeBlock {
    range: Range<usize>,
    display: String,
    exec: String,
    mode: BlockMode,
}

fn preprocess_code_blocks(content: &str) -> Result<Vec<CodeBlock>> {
    let mut blocks: Vec<CodeBlock> = Vec::with_capacity(1);
    let parser = Parser::new(content).into_offset_iter();
    for item in parser {
        if let (Event::Start(Tag::CodeBlock(kind)), range) = item {
            if let Some(mode) = block_mode(kind) {
                let (display, exec) = preprocess(&content[range.clone()])?;
                blocks.push(CodeBlock {
                    range,
                    display,
                    exec,
                    mode,
                });
            }
        }
    }
    Ok(blocks)
}

fn block_mode(kind: CodeBlockKind) -> Option<BlockMode> {
    match kind {
        CodeBlockKind::Indented => Some(BlockMode::Run),
        CodeBlockKind::Fenced(lang) => match lang.as_ref() {
            "" => Some(BlockMode::Run),
            "flux" => Some(BlockMode::Run),
            "no_run" => Some(BlockMode::NoRun),
            _ => None,
        },
    }
}

// Performs a basic transformation on any Flux code blocks to make them executable:
//
// Transformations:
//    '# ' - hide the line in the display
//    '< ' - mark the line as the input
//    '> ' - mark the line as the output
//    '#<' - mark the line as the input and hide
//    '#>' - mark the line as the output and hide
fn preprocess(code: &str) -> Result<(String, String)> {
    const OUTPUT_YIELD: &str = " |> yield(name:\"output\")";
    const INPUT_YIELD: &str = " |> yield(name:\"input\")";
    let mut display = String::new();
    let mut exec = String::new();
    for line in code.lines().filter(|l| !l.starts_with("```")) {
        if line == "#" {
            // skip lines that do not have any content
            continue;
        }
        if line.len() >= 2 {
            match &line[..2] {
                "# " => {
                    exec.push_str(&line[2..]);
                    exec.push('\n');
                    continue;
                }
                "#<" => {
                    exec.push_str(&line[2..]);
                    exec.push_str(INPUT_YIELD);
                    exec.push('\n');
                    continue;
                }
                "#>" => {
                    exec.push_str(&line[2..]);
                    exec.push_str(OUTPUT_YIELD);
                    exec.push('\n');
                    continue;
                }
                "< " => {
                    display.push_str(&line[2..]);
                    display.push('\n');
                    exec.push_str(&line[2..]);
                    exec.push_str(INPUT_YIELD);
                    exec.push('\n');
                    continue;
                }
                "> " => {
                    display.push_str(&line[2..]);
                    display.push('\n');
                    exec.push_str(&line[2..]);
                    exec.push_str(OUTPUT_YIELD);
                    exec.push('\n');
                    continue;
                }
                _ => {}
            }
        }
        display.push_str(line);
        display.push('\n');
        exec.push_str(line);
        exec.push('\n');
    }
    display = format(display.as_str())?;
    exec = format(exec.as_str())?;
    // Add the code fences back in as they were removed above.
    // TODO(nathanielc): Allow for specifying the code tag.
    display = format!("```flux\n{}\n```", display);
    Ok((display, exec))
}

struct StringTable {
    group: Vec<bool>,
    header: Vec<String>,
    rows: Vec<Vec<String>>,
    ready: bool,
}

impl StringTable {
    fn reset(&mut self) {
        self.rows = Vec::new();
        self.ready = false;
    }
    fn to_markdown(&self) -> Table {
        let mut t = Table::new();
        // Find width of each column
        let mut widths: Vec<usize> = Vec::with_capacity(self.header.len());
        for h in &self.header {
            widths.push(h.len() + 1);
        }
        for row in &self.rows {
            for (i, v) in row.iter().enumerate() {
                if widths[i] < v.len() {
                    widths[i] = v.len();
                }
            }
        }
        // Format into string buffer
        for (i, h) in self.header.iter().enumerate() {
            t.push_str("| ");
            let mut width = widths[i];
            if self.group[i] {
                t.push('*');
                width -= 1;
            }
            t.push_str(&h.pad_to_width(width));
            t.push(' ');
        }
        t.push_str("|\n");
        for (i, _) in self.header.iter().enumerate() {
            t.push_str("| ");
            for _i in 0..widths[i] {
                t.push('-');
            }
            t.push(' ');
        }
        t.push_str("|\n");
        for row in &self.rows {
            for (i, v) in row.iter().enumerate() {
                t.push_str("| ");
                t.push_str(&v.pad_to_width(widths[i]));
                t.push(' ');
            }
            t.push_str("|\n");
        }
        t
    }
}
const INPUT_NAME: &str = "input";
const OUTPUT_NAME: &str = "output";
const ANNOTATION_IDX: usize = 0;
const RESULT_IDX: usize = 1;
const TABLE_IDX: usize = 2;
const DATA_START_IDX: usize = 3;
const ERROR_LABEL: &str = "error";
const DATATYPE_LABEL: &str = "#datatype";
const DEFAULT_LABEL: &str = "#default";
const GROUP_LABEL: &str = "#group";

// Parses Flux CSV for input and output results into markdown formatted tables.
#[allow(clippy::type_complexity)]
fn parse_results(data: &str) -> Result<(Option<Vec<Table>>, Option<Vec<Table>>)> {
    let mut input: Option<Vec<Table>> = None;
    let mut output: Option<Vec<Table>> = None;
    let mut results = parse_all_results(data)?;
    if let Some(i) = results.remove(INPUT_NAME) {
        input = Some(i)
    }
    if let Some(o) = results.remove(OUTPUT_NAME) {
        output = Some(o)
    }
    Ok((input, output))
}

fn get_cell(record: &StringRecord, i: usize, defaults: &[String]) -> Result<String> {
    match record.get(i) {
        Some(c) => {
            if c.is_empty() {
                Ok(defaults[i].to_owned())
            } else {
                Ok(c.to_owned())
            }
        }
        None => bail!("could not read cell at index {}", i),
    }
}
fn get_row(record: &StringRecord, defaults: &[String]) -> Result<Vec<String>> {
    Ok(record
        .iter()
        .enumerate()
        .skip(DATA_START_IDX)
        .map(|(i, c)| {
            if c.is_empty() {
                defaults[i].to_owned()
            } else {
                c.to_owned()
            }
        })
        .collect())
}

fn parse_all_results(data: &str) -> Result<HashMap<String, Vec<Table>>> {
    let mut reader = csv::ReaderBuilder::new()
        .has_headers(false)
        .flexible(true)
        .from_reader(data.as_bytes());
    let records = reader.records();

    let mut results: HashMap<String, Vec<Table>> = HashMap::new();
    let mut result_name = String::new();

    // Current set of tables in the result
    let mut tables: Vec<Table> = Vec::new();
    // Current table
    let mut table = StringTable {
        header: Vec::new(),
        group: Vec::new(),
        rows: Vec::new(),
        ready: false,
    };
    // Current defaults specified by the annotation
    let mut defaults: Vec<String> = Vec::new();

    // Index of current rows starting with 0 at the header row.
    // Any annotation rows reset this counter.
    let mut i: usize = 0;

    // Indicates if we found an error encoded into the csv table.
    let mut is_error = false;

    // Current ID of the table
    let mut table_id = "".to_string();
    for record in records {
        let record = record?;
        let annotation = match record.get(ANNOTATION_IDX) {
            Some(a) => a,
            None => bail!("could not read annotation column"),
        };
        match annotation {
            ERROR_LABEL => {
                is_error = true;
            }
            DATATYPE_LABEL => {
                if table.ready {
                    tables.push(table.to_markdown());
                    table.reset();
                }
                i = 0;
            }
            DEFAULT_LABEL => {
                if table.ready {
                    tables.push(table.to_markdown());
                    table.reset();
                }
                i = 0;
                defaults = record.iter().map(|c| c.to_string()).collect();
            }
            GROUP_LABEL => {
                if table.ready {
                    tables.push(table.to_markdown());
                    table.reset();
                }
                i = 0;
                table.group = record
                    .iter()
                    .skip(DATA_START_IDX)
                    .map(|v| v == "true")
                    .collect();
            }
            _ => {
                if i == 0 {
                    if is_error {
                        // We have an error and not a header so return the error
                        match record.get(ANNOTATION_IDX) {
                            Some(err) => bail!("flux error: {}", err),
                            None => bail!("flux error: unknown"),
                        }
                    }
                    // Header row
                    table.header = record
                        .iter()
                        .skip(DATA_START_IDX)
                        .map(|h| h.to_string())
                        .collect();
                    table.ready = true;
                } else {
                    if i == 1 {
                        // First data row: Are we in a new result?
                        let name = get_cell(&record, RESULT_IDX, &defaults)?;
                        if result_name.is_empty() {
                            result_name = name.to_owned();
                        }
                        if name != result_name {
                            results.insert(result_name.clone(), tables);
                            tables = Vec::new();
                            table_id = String::new();
                            result_name = name;
                        }
                    }
                    let id = get_cell(&record, TABLE_IDX, &defaults)?;
                    if table_id.is_empty() {
                        table_id = id.to_string();
                    }
                    if table_id != id {
                        table_id = id.to_string();
                        tables.push(table.to_markdown());
                        table.reset();
                        // table remains ready since its headers didn't change
                        table.ready = true;
                    }
                    table.rows.push(get_row(&record, &defaults)?);
                };
                i += 1;
            }
        }
    }
    if table.ready {
        tables.push(table.to_markdown());
        table.reset();
    }
    results.insert(result_name, tables);
    Ok(results)
}

#[cfg(test)]
mod tests {
    use std::collections::BTreeMap;

    use anyhow::Result;
    use expect_test::{expect, Expect};

    use super::{evaluate_package_examples, parse_results, preprocess, Executor};
    use crate::doc::{Example, PackageDoc};

    struct MockExecutor<'a> {
        code: Expect,
        results: &'a str,
    }
    impl<'a> Executor for MockExecutor<'a> {
        fn execute(&self, code: &str) -> Result<String> {
            self.code.assert_eq(code);
            Ok(self.results.to_string())
        }
    }

    #[test]
    fn test_simple() {
        let mut doc = PackageDoc {
            path: "".to_string(),
            name: "".to_string(),
            headline: "".to_string(),
            description: None,
            members: BTreeMap::new(),
            examples: vec![Example {
                title: "".to_string(),
                content: r#"
Example on using array.from

```
# import "array"
< array.from(rows: [{_value: "a"}, {_value: "b"}])
>   |> map(fn: (r) => ({r with _value: "b"}))
```
"#
                .to_string(),
                input: None,
                output: None,
            }],
            metadata: None,
        };
        let executor = MockExecutor {
            code: expect![[r#"
                import "array"

                array.from(rows: [{_value: "a"}, {_value: "b"}])
                    |> yield(name: "input")
                    |> map(fn: (r) => ({r with _value: "b"}))
                    |> yield(name: "output")
            "#]],
            results: r#"#datatype,string,long,string
#group,false,false,false
#default,input,,
,result,table,_value
,,0,a
,,0,b

#datatype,string,long,string
#group,false,false,false
#default,output,,
,result,table,_value
,,0,b
,,0,b
"#,
        };

        for result in evaluate_package_examples(&mut doc, &executor) {
            result.unwrap();
        }

        let example = doc.examples.first().unwrap();
        let input = example.input.as_ref().unwrap().join("\n");
        let output = example.output.as_ref().unwrap().join("\n");

        let want_content = expect![[r#"

            Example on using array.from

            ```flux
            array.from(rows: [{_value: "a"}, {_value: "b"}])
                |> map(fn: (r) => ({r with _value: "b"}))

            ```
        "#]];
        want_content.assert_eq(&example.content);

        let want_input = expect![[r#"
            | _value  |
            | ------- |
            | a       |
            | b       |
        "#]];
        want_input.assert_eq(input.as_str());

        let want_output = expect![[r#"
            | _value  |
            | ------- |
            | b       |
            | b       |
        "#]];
        want_output.assert_eq(output.as_str());
    }
    #[test]
    fn test_preprocess() {
        let (display, exec) = preprocess(
            r#"
# import "array"
#
#
< array.from(rows:[{_value:"a"}])
>   |> map(fn: (r) => ({r with _value: "b"}))
"#,
        )
        .unwrap();

        let want_display = expect![[r#"
            ```flux
            array.from(rows: [{_value: "a"}])
                |> map(fn: (r) => ({r with _value: "b"}))

            ```"#]];
        want_display.assert_eq(display.as_str());

        let want_exec = expect![[r#"
            import "array"

            array.from(rows: [{_value: "a"}])
                |> yield(name: "input")
                |> map(fn: (r) => ({r with _value: "b"}))
                |> yield(name: "output")
        "#]];
        want_exec.assert_eq(exec.as_str());
    }
    #[test]
    fn test_preprocess_large() {
        let (display, exec) = preprocess(r#"import "array"

array.from(
    rows: [
        {_measurement: "m0", _field: "f0", t0: "tagvalue", _time: 2018-12-19T22:13:30Z, _value: false},
        {_measurement: "m0", _field: "f0", t0: "tagvalue", _time: 2018-12-19T22:13:40Z, _value: true},
        {_measurement: "m0", _field: "f0", t0: "tagvalue", _time: 2018-12-19T22:13:50Z, _value: false},
        {_measurement: "m0", _field: "f0", t0: "tagvalue", _time: 2018-12-19T22:14:00Z, _value: false},
        {_measurement: "m0", _field: "f0", t0: "tagvalue", _time: 2018-12-19T22:14:10Z, _value: true},
        {_measurement: "m0", _field: "f0", t0: "tagvalue", _time: 2018-12-19T22:14:20Z, _value: true},
    ],
< )
>   |> map(fn: (r) => ({r with _value: "b"}))"#).unwrap();

        let want_display = expect![[r#"
            ```flux
            import "array"

            array.from(
                rows: [
                    {
                        _measurement: "m0",
                        _field: "f0",
                        t0: "tagvalue",
                        _time: 2018-12-19T22:13:30Z,
                        _value: false,
                    },
                    {
                        _measurement: "m0",
                        _field: "f0",
                        t0: "tagvalue",
                        _time: 2018-12-19T22:13:40Z,
                        _value: true,
                    },
                    {
                        _measurement: "m0",
                        _field: "f0",
                        t0: "tagvalue",
                        _time: 2018-12-19T22:13:50Z,
                        _value: false,
                    },
                    {
                        _measurement: "m0",
                        _field: "f0",
                        t0: "tagvalue",
                        _time: 2018-12-19T22:14:00Z,
                        _value: false,
                    },
                    {
                        _measurement: "m0",
                        _field: "f0",
                        t0: "tagvalue",
                        _time: 2018-12-19T22:14:10Z,
                        _value: true,
                    },
                    {
                        _measurement: "m0",
                        _field: "f0",
                        t0: "tagvalue",
                        _time: 2018-12-19T22:14:20Z,
                        _value: true,
                    },
                ],
            )
                |> map(fn: (r) => ({r with _value: "b"}))

            ```"#]];
        want_display.assert_eq(display.as_str());

        let want_exec = expect![[r#"
            import "array"

            array.from(
                rows: [
                    {
                        _measurement: "m0",
                        _field: "f0",
                        t0: "tagvalue",
                        _time: 2018-12-19T22:13:30Z,
                        _value: false,
                    },
                    {
                        _measurement: "m0",
                        _field: "f0",
                        t0: "tagvalue",
                        _time: 2018-12-19T22:13:40Z,
                        _value: true,
                    },
                    {
                        _measurement: "m0",
                        _field: "f0",
                        t0: "tagvalue",
                        _time: 2018-12-19T22:13:50Z,
                        _value: false,
                    },
                    {
                        _measurement: "m0",
                        _field: "f0",
                        t0: "tagvalue",
                        _time: 2018-12-19T22:14:00Z,
                        _value: false,
                    },
                    {
                        _measurement: "m0",
                        _field: "f0",
                        t0: "tagvalue",
                        _time: 2018-12-19T22:14:10Z,
                        _value: true,
                    },
                    {
                        _measurement: "m0",
                        _field: "f0",
                        t0: "tagvalue",
                        _time: 2018-12-19T22:14:20Z,
                        _value: true,
                    },
                ],
            )
                |> yield(name: "input")
                |> map(fn: (r) => ({r with _value: "b"}))
                |> yield(name: "output")
        "#]];
        want_exec.assert_eq(exec.as_str());
    }
    #[test]
    fn test_parse_results() {
        let data = r#"#datatype,string,long,string,string
#group,false,false,true,false
#default,input,,,
,result,table,tag,_value
,,0,a,1
,,0,a,2
,,1,b,3
,,1,b,4

#datatype,string,long,string,string,string
#group,false,false,true,true,false
#default,output,,,,
,result,table,tag,othertag,_value
,,0,a,x,11
,,0,a,x,12
,,1,b,x,13
,,1,b,x,14
"#;

        let (input, output) = parse_results(data).unwrap();

        let input = input.unwrap().join("\n");
        let output = output.unwrap().join("\n");

        let want_input = expect![[r#"
            | *tag | _value  |
            | ---- | ------- |
            | a    | 1       |
            | a    | 2       |

            | *tag | _value  |
            | ---- | ------- |
            | b    | 3       |
            | b    | 4       |
        "#]];
        want_input.assert_eq(input.as_str());

        let want_output = expect![[r#"
            | *tag | *othertag | _value  |
            | ---- | --------- | ------- |
            | a    | x         | 11      |
            | a    | x         | 12      |

            | *tag | *othertag | _value  |
            | ---- | --------- | ------- |
            | b    | x         | 13      |
            | b    | x         | 14      |
        "#]];
        want_output.assert_eq(output.as_str());
    }
    #[test]
    fn test_parse_results_error_table() {
        let data = r#"error,reference
encoded error message,5
"#;

        let err = parse_results(data).unwrap_err();
        assert_eq!(
            "flux error: encoded error message",
            err.to_string().as_str()
        );
    }
}
