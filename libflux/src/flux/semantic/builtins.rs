use maplit::hashmap;
use std::collections::hash_map;
use std::collections::HashMap;
use std::iter::Iterator;
use std::vec::Vec;

pub struct Builtins<'a> {
    pkgs: HashMap<&'a str, Node<'a>>,
}

impl<'a> Builtins<'a> {
    fn iter(&'a self) -> NodeIterator {
        return NodeIterator::new(self);
    }
}

enum Node<'a> {
    Package(HashMap<&'a str, Node<'a>>),
    Builtin(&'a str),
}

pub fn builtins() -> Builtins<'static> {
    Builtins {
        pkgs: hashmap! {
            "csv" => Node::Package(maplit::hashmap! {
                // This is a "provide exactly one argument" function
                "from" => Node::Builtin("forall [t0] (?csv: string, ?file: string) -> [t0]"),
            }),
            "date" => Node::Package(maplit::hashmap! {
                 "second" => Node::Builtin("forall [] (t: time) -> int"),
                 "minute" => Node::Builtin("forall [] (t: time) -> int"),
                 "hour" => Node::Builtin("forall [] (t: time) -> int"),
                 "weekDay" => Node::Builtin("forall [] (t: time) -> int"),
                 "monthDay" => Node::Builtin("forall [] (t: time) -> int"),
                 "yearDay" => Node::Builtin("forall [] (t: time) -> int"),
                 "month" => Node::Builtin("forall [] (t: time) -> int"),
                 "year" => Node::Builtin("forall [] (t: time) -> int"),
                 "week" => Node::Builtin("forall [] (t: time) -> int"),
                 "quarter" => Node::Builtin("forall [] (t: time) -> int"),
                 "millisecond" => Node::Builtin("forall [] (t: time) -> int"),
                 "microsecond" => Node::Builtin("forall [] (t: time) -> int"),
                 "nanosecond" => Node::Builtin("forall [] (t: time) -> int"),
                 "truncate" => Node::Builtin("forall [] (t: time, unit: duration) -> time"),
            }),
            "experimental" => Node::Package(maplit::hashmap! {
                 "bigtable" => Node::Package(maplit::hashmap! {
                     "from" => Node::Builtin("forall [t0] (token: string, project: string, instance: string, table: string) -> [t0]"),
                 }),
                 "http" => Node::Package(maplit::hashmap! {
                     "get" => Node::Builtin("forall [t0, t1] (url: string, ?headers: t0, ?timeout: duration) -> {statusCode: int | body: bytes | headers: t1}"),
                 }),
                 "mqtt" => Node::Package(maplit::hashmap! {
                     "to" => Node::Builtin(r#"
                         forall [t0, t1] (
                             <-tables: [t0],
                             broker: string,
                             ?topic: string,
                             ?message: string,
                             ?qos: int,
                             ?clientid: string,
                             ?username: string,
                             ?password: string,
                             ?name: string,
                             ?timeout: duration,
                             ?timeColumn: string,
                             ?tagColumns: [string],
                             ?valueColumns: [string]
                         ) -> [t1]
                     "#),
                 }),
                 "prometheus" => Node::Package(maplit::hashmap! {
                     "scrape" => Node::Builtin("forall [t0] (url: string) -> [t0]"),
                 }),
                 "addDuration" => Node::Builtin("forall [] (d: duration, to: time) -> time"),
                 "subDuration" => Node::Builtin("forall [] (d: duration, from: time) -> time"),
                 "group" => Node::Builtin("forall [t0, t1] (<-tables: [t0], mode: string, columns: [string]) -> t1"),
                 "objectKeys" => Node::Builtin("forall [t0] (o: t0) -> [string]"),
                 "set" => Node::Builtin("forall [t0, t1, t2] (<-tables: t0, o: t1) -> t2"),
                 // must specify exactly one of bucket, bucketID
                 // must specify exactly one of org, orgID
                 // if host is specified, token must be too.
                 "to" => Node::Builtin("forall [t0] (?bucket: string, ?bucketID: string, ?org: string, ?orgID: string, ?host: string, ?token: string) -> t0"),
            }),
            "generate" => Node::Package(maplit::hashmap! {
                "from" => Node::Builtin("forall [] (start: time, stop: time, count: int, fn: (n: int) -> int) -> [{ _start: time | _stop: time | _time: time | _value:int }]"),
            }),
            "http" => Node::Package(maplit::hashmap! {
                "post" => Node::Builtin("forall [t0] (url: string, ?headers: t0, ?data: bytes) -> int"),
                "basicAuth" => Node::Builtin("forall [] (u: string, p: string) -> string"),
            }),
            "influxdata" => Node::Package(maplit::hashmap! {
                "influxdb" => Node::Package(maplit::hashmap! {
                    "secrets" => Node::Package(maplit::hashmap! {
                         "get" => Node::Builtin("forall [] (key: string) -> string"),
                    }),
                    "v1" => Node::Package(maplit::hashmap! {
                        // exactly one of json and file must be specified
                        "json" => Node::Builtin("forall [t0] (?json: string, ?file: string) -> [t0]"),
                        "databases" => Node::Builtin(r#"
                            forall [] () -> {
                                organizationID: string |
                                databaseName: string |
                                retentionPolicy: string |
                                retentionPeriod: int |
                                default: bool |
                                bucketID: string
                            }
                        "#),
                    }),
                    // This is a one-or-the-other parameters function
                    "from" => Node::Builtin("forall [t0, t1] (?bucket: string, ?bucketID: string) -> [{_measurement: string | _field: string | _time: time | _value: t0 | t1}]"),
                    // exactly one of (bucket, bucketID) must be specified
                    // exactly one of (org, orgID) must be specified
                    "to" => Node::Builtin(r#"
                        forall [t0, t1] (
                            <-tables: [t0],
                            ?bucket: string,
                            ?bucketID: string,
                            ?org: string,
                            ?orgID: string,
                            ?token: string,
                            ?timeColumn: string,
                            ?measurementColumn: string,
                            ?tagColumns: [string],
                            ?fieldFn: (r: t0) -> t1
                        ) -> [t1]
                    "#),
                    "buckets" => Node::Builtin("forall [] () -> {name: string | id: string | organizationID: string | retentionPolicy: string | retentionPeriod: int}"),
                }),

            }),
            "internal" => Node::Package(maplit::hashmap! {
                "gen" => Node::Package(maplit::hashmap! {
                    "tables" => Node::Builtin("forall [t0] (n: int, tags: [{name: string | cardinality: int}]) -> {_time: time | _value: float | t0}"),
                }),
                "promql" => Node::Package(maplit::hashmap! {
                    "changes" => Node::Builtin("forall [t0, t1] (<-tables: [{_value: float | t0}]) -> [{_value: float | t1}]"),
                    "promqlDayOfMonth" => Node::Builtin("forall [] (timestamp: float) -> float"),
                    "promqlDayOfWeek" => Node::Builtin("forall [] (timestamp: float) -> float"),
                    "promqlDaysInMonth" => Node::Builtin("forall [] (timestamp: float) -> float"),
                    "emptyTable" => Node::Builtin("forall [] () -> [{_start: time | _stop: time | _time: time | _value: float}]"),
                    "extrapolatedRate" => Node::Builtin("forall [t0, t1] (<-tables: [{_start: time | _stop: time | _time: time | _value: float | t0}], ?isCounter: bool, ?isRate: bool) -> [{_value: float | t1}]"),
                    "holtWinters" => Node::Builtin("forall [t0, t1] (<-tables: [{_time: time | _value: float | t0}], ?smoothingFactor: float, ?trendFactor: float) -> [{_value: float | t1}]"),
                    "promqlHour" => Node::Builtin("forall [] (timestamp: float) -> float"),
                    "instantRate" => Node::Builtin("forall [t0, t1] (<-tables: [{_time: time | _value: float | t0}], ?isRate: bool) -> [{_value: float | t1}]"),
                    "labelReplace" => Node::Builtin("forall [t0, t1] (<-tables: [{_value: float | t0}], source: string, destination: string, regex: string, replacement: string) -> [{_value: float | t1}]"),
                    "linearRegression" => Node::Builtin("forall [t0, t1] (<-tables: [{_time: time | _stop: time | _value: float | t0}], ?predict: bool, ?fromNow: float) -> [{_value: float | t1}]"),
                    "promqlMinute" => Node::Builtin("forall [] (timestamp: float) -> float"),
                    "promqlMonth" => Node::Builtin("forall [] (timestamp: float) -> float"),
                    "promHistogramQuantile" => Node::Builtin("forall [t0, t1] (<-tables: [t0], ?quantile: float, ?countColumn: string, ?upperBoundColumn: string, ?valueColumn: string) -> [t1]"),
                    "resets" => Node::Builtin("forall [t0, t1] (<-tables: [{_value: float | t0}]) -> [{_value: float | t1}]"),
                    "timestamp" => Node::Builtin("forall [t0] (<-tables: [{_value: float | t0}]) -> [{_value: float | t0}]"),
                    "promqlYear" => Node::Builtin("forall [] (timestamp: float) -> float"),
                    "join" => Node::Builtin("forall [t0, t1, t2] (left: [t0], right: [t1], fn: (left: t0, right: t1) -> t2) -> [t2]"),
                }),
            }),
            "json" => Node::Package(maplit::hashmap! {
                "encode" => Node::Builtin("forall [t0] (v: t0) -> bytes"),
            }),
            "kafka" => Node::Package(maplit::hashmap! {
                "to" => Node::Builtin(r#"
                    forall [t0] (
                        <-tables: [t0],
                        brokers: string,
                        topic: string,
                        ?balancer: string,
                        ?name: string,
                        ?nameColumn: string,
                        ?timeColumn: string,
                        ?tagColumns: [string],
                        ?valueColumns: [string]
                    ) -> [t0]"#),
            }),

            "math" => Node::Package(maplit::hashmap! {
                "pi" => Node::Builtin("forall [] float"),
                "e" => Node::Builtin("forall [] float"),
                "phi" => Node::Builtin("forall [] float"),
                "sqrt2" => Node::Builtin("forall [] float"),
                "sqrte" => Node::Builtin("forall [] float"),
                "sqrtpi" => Node::Builtin("forall [] float"),
                "sqrtphi" => Node::Builtin("forall [] float"),
                "log2e" => Node::Builtin("forall [] float"),
                "ln2" => Node::Builtin("forall [] float"),
                "ln10" => Node::Builtin("forall [] float"),
                "log10e" => Node::Builtin("forall [] float"),

                "maxfloat" => Node::Builtin("forall [] float"),
                "smallestNonzeroFloat" => Node::Builtin("forall [] float"),
                "maxint" => Node::Builtin("forall [] int"),
                "minint" => Node::Builtin("forall [] int"),
                "maxuint" => Node::Builtin("forall [] uint"),

                "abs" => Node::Builtin("forall [] (x: float) -> float"),
                "acos" => Node::Builtin("forall [] (x: float) -> float"),
                "acosh" => Node::Builtin("forall [] (x: float) -> float"),
                "asin" => Node::Builtin("forall [] (x: float) -> float"),
                "asinh" => Node::Builtin("forall [] (x: float) -> float"),
                "atan" => Node::Builtin("forall [] (x: float) -> float"),
                "atan2" => Node::Builtin("forall [] (x: float, y: float) -> float"),
                "atanh" => Node::Builtin("forall [] (x: float) -> float"),
                "cbrt" => Node::Builtin("forall [] (x: float) -> float"),
                "ceil" => Node::Builtin("forall [] (x: float) -> float"),
                "copysign" => Node::Builtin("forall [] (x: float, y: float) -> float"),
                "cos" => Node::Builtin("forall [] (x: float) -> float"),
                "cosh" => Node::Builtin("forall [] (x: float) -> float"),
                "dim" => Node::Builtin("forall [] (x: float, y: float) -> float"),
                "erf" => Node::Builtin("forall [] (x: float) -> float"),
                "erfc" => Node::Builtin("forall [] (x: float) -> float"),
                "erfcinv" => Node::Builtin("forall [] (x: float) -> float"),
                "erfinv" => Node::Builtin("forall [] (x: float) -> float"),
                "exp" => Node::Builtin("forall [] (x: float) -> float"),
                "exp2" => Node::Builtin("forall [] (x: float) -> float"),
                "expm1" => Node::Builtin("forall [] (x: float) -> float"),
                "floor" => Node::Builtin("forall [] (x: float) -> float"),
                "gamma" => Node::Builtin("forall [] (x: float) -> float"),
                "hypot" => Node::Builtin("forall [] (x: float, y: float) -> float"),
                "j0" => Node::Builtin("forall [] (x: float) -> float"),
                "j1" => Node::Builtin("forall [] (x: float) -> float"),
                "log" => Node::Builtin("forall [] (x: float) -> float"),
                "log10" => Node::Builtin("forall [] (x: float) -> float"),
                "log1p" => Node::Builtin("forall [] (x: float) -> float"),
                "log2" => Node::Builtin("forall [] (x: float) -> float"),
                "logb" => Node::Builtin("forall [] (x: float) -> float"),
                "mMax" => Node::Builtin("forall [] (x: float, y: float) -> float"),
                "mMin" => Node::Builtin("forall [] (x: float, y: float) -> float"),
                "mod" => Node::Builtin("forall [] (x: float, y: float) -> float"),
                "nextafter" => Node::Builtin("forall [] (x: float, y: float) -> float"),
                "pow" => Node::Builtin("forall [] (x: float, y: float) -> float"),
                "remainder" => Node::Builtin("forall [] (x: float, y: float) -> float"),
                "round" => Node::Builtin("forall [] (x: float) -> float"),
                "roundtoeven" => Node::Builtin("forall [] (x: float) -> float"),
                "sin" => Node::Builtin("forall [] (x: float) -> float"),
                "sinh" => Node::Builtin("forall [] (x: float) -> float"),
                "sqrt" => Node::Builtin("forall [] (x: float) -> float"),
                "tan" => Node::Builtin("forall [] (x: float) -> float"),
                "tanh" => Node::Builtin("forall [] (x: float) -> float"),
                "trunc" => Node::Builtin("forall [] (x: float) -> float"),
                "y0" => Node::Builtin("forall [] (x: float) -> float"),
                "y1" => Node::Builtin("forall [] (x: float) -> float"),

                "float64bits" => Node::Builtin("forall [] (f: float) -> uint"),
                "float64frombits" => Node::Builtin("forall [] (b: uint) -> float"),
                "ilogb" => Node::Builtin("forall [] (x: float) -> int"),
                "frexp" => Node::Builtin("forall [] (f: float) -> {frac: float | exp: int}"),
                "lgamma" => Node::Builtin("forall [] (x: float) -> {lgamma: float | sign: int}"),
                "modf" => Node::Builtin(r#"forall [] (f: float) -> {"int": float | frac: float}"#),
                "sincos" => Node::Builtin("forall [] (x: float) -> {sin: float | cos: float}"),
                "isInf" => Node::Builtin("forall [] (f: float, sign: int) -> bool"),
                "isNaN" => Node::Builtin("forall [] (f: float) -> bool"),
                "signbit" => Node::Builtin("forall [] (x: float) -> bool"),
                "NaN" => Node::Builtin("forall [] () -> float"),
                "mInf" => Node::Builtin("forall [] (sign: int) -> float"),
                "jn" => Node::Builtin("forall [] (n: int, x: float) -> float"),
                "yn" => Node::Builtin("forall [] (n: int, x: float) -> float"),
                "ldexp" => Node::Builtin("forall [] (frac: float, exp: int) -> float"),
                "pow10" => Node::Builtin("forall [] (n: int) -> float"),
            }),
            "pagerduty" => Node::Package(maplit::hashmap! {
                "dedupKey" => Node::Builtin("forall [t0] (<-tables: [t0]) -> [{_pagerdutyDedupKey: string | t0}]"),
            }),
            "regexp" => Node::Package(maplit::hashmap! {
                "compile" => Node::Builtin("forall [] (v: string) -> regexp"),
                "quoteMeta" => Node::Builtin("forall [] (v: string) -> string"),
                "findString" => Node::Builtin("forall [] (r: regexp, v: string) -> string"),
                "findStringIndex" => Node::Builtin("forall [] (r: regexp, v: string) -> [int]"),
                "matchRegexpString" => Node::Builtin("forall [] (r: regexp, v: string) -> bool"),
                "replaceAllString" => Node::Builtin("forall [] (r: regexp, v: string, t: string) -> string"),
                "splitRegexp" => Node::Builtin("forall [] (r: regexp, v: string, i: int) -> [string]"),
                "getString" => Node::Builtin("forall [] (r: regexp) -> string"),
            }),
        },
    }
}

struct NodeIterator<'a> {
    path_elems: Vec<&'a str>,
    iter_stack: Vec<hash_map::Iter<'a, &'a str, Node<'a>>>,
}

impl<'a> NodeIterator<'a> {
    fn new(builtins: &'a Builtins) -> NodeIterator<'a> {
        return NodeIterator {
            path_elems: Vec::new(),
            iter_stack: vec![builtins.pkgs.iter()],
        };
    }
}

impl<'a> Iterator for NodeIterator<'a> {
    type Item = (std::vec::Vec<&'a str>, &'a str);

    fn next(&mut self) -> Option<Self::Item> {
        let mut it = self.iter_stack.pop()?;

        match it.next() {
            None => {
                self.path_elems.pop();
                self.next()
            }
            Some((name, item)) => {
                self.path_elems.push(name);
                match item {
                    Node::Package(m) => {
                        self.iter_stack.push(it);
                        self.iter_stack.push(m.into_iter());
                        self.next()
                    }
                    Node::Builtin(ty) => {
                        let item = (self.path_elems.clone(), *ty);
                        self.path_elems.pop();
                        self.iter_stack.push(it);
                        Some(item)
                    }
                }
            }
        }
    }
}

#[cfg(test)]
mod test {
    use crate::semantic::builtins::builtins;
    use crate::semantic::parser as type_parser;

    #[test]
    fn parse_builtin_types() {
        for (path, ty) in builtins().iter() {
            match type_parser::parse(ty) {
                Ok(_) => {}
                Err(s) => {
                    let msg = format!("{} type failed to parse: {}", path.join("/"), s);
                    panic!(msg)
                }
            }
        }
    }
}
