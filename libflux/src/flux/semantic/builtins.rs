use maplit::hashmap;
use std::collections::hash_map;
use std::collections::HashMap;
use std::iter::Iterator;
use std::vec::Vec;

pub struct Builtins<'a> {
    pkgs: HashMap<&'a str, Node<'a>>,
}

impl<'a> Builtins<'a> {
    pub fn iter(&'a self) -> NodeIterator {
        NodeIterator::new(self)
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
                // https://github.com/influxdata/flux/issues/2249
                "from" => Node::Builtin("forall [t0] where t0: Row (?csv: string, ?file: string) -> [t0]"),
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
                     "from" => Node::Builtin("forall [t0] where t0: Row (token: string, project: string, instance: string, table: string) -> [t0]"),
                 }),
                 "http" => Node::Package(maplit::hashmap! {
                     "get" => Node::Builtin(r#"
                         forall [t0, t1] where t0: Row, t1: Row (
                             url: string,
                             ?headers: t0, 
                             ?timeout: duration
                         ) -> {statusCode: int | body: bytes | headers: t1}
                     "#),
                 }),
                 "mqtt" => Node::Package(maplit::hashmap! {
                     "to" => Node::Builtin(r#"
                         forall [t0, t1] where t0: Row, t1: Row (
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
                     "scrape" => Node::Builtin("forall [t0] where t0: Row (url: string) -> [t0]"),
                 }),
                 "addDuration" => Node::Builtin("forall [] (d: duration, to: time) -> time"),
                 "subDuration" => Node::Builtin("forall [] (d: duration, from: time) -> time"),
                 "group" => Node::Builtin("forall [t0] where t0: Row (<-tables: [t0], mode: string, columns: [string]) -> [t0]"),
                 "objectKeys" => Node::Builtin("forall [t0] where t0: Row (o: t0) -> [string]"),
                 "set" => Node::Builtin("forall [t0, t1, t2] where t0: Row, t1: Row, t2: Row (<-tables: [t0], o: t1) -> [t2]"),
                 // must specify exactly one of bucket, bucketID
                 // must specify exactly one of org, orgID
                 // if host is specified, token must be too.
                 // https://github.com/influxdata/flux/issues/1660
                 "to" => Node::Builtin("forall [t0] where t0: Row (<-tables: [t0], ?bucket: string, ?bucketID: string, ?org: string, ?orgID: string, ?host: string, ?token: string) -> [t0]"),
            }),
            "generate" => Node::Package(maplit::hashmap! {
                "from" => Node::Builtin("forall [] (start: time, stop: time, count: int, fn: (n: int) -> int) -> [{ _start: time | _stop: time | _time: time | _value:int }]"),
            }),
            "http" => Node::Package(maplit::hashmap! {
                "post" => Node::Builtin("forall [t0] where t0: Row (url: string, ?headers: t0, ?data: bytes) -> int"),
                "basicAuth" => Node::Builtin("forall [] (u: string, p: string) -> string"),
            }),
            "influxdata" => Node::Package(maplit::hashmap! {
                "influxdb" => Node::Package(maplit::hashmap! {
                    "secrets" => Node::Package(maplit::hashmap! {
                         "get" => Node::Builtin("forall [] (key: string) -> string"),
                    }),
                    "v1" => Node::Package(maplit::hashmap! {
                        // exactly one of json and file must be specified
                        // https://github.com/influxdata/flux/issues/2250
                        "json" => Node::Builtin("forall [t0] where t0: Row (?json: string, ?file: string) -> [t0]"),
                        "databases" => Node::Builtin(r#"
                            forall [] () -> [{
                                organizationID: string |
                                databaseName: string |
                                retentionPolicy: string |
                                retentionPeriod: int |
                                default: bool |
                                bucketID: string
                            }]
                        "#),
                    }),
                    // This is a one-or-the-other parameters function
                    // https://github.com/influxdata/flux/issues/1659
                    "from" => Node::Builtin("forall [t0, t1] (?bucket: string, ?bucketID: string) -> [{_measurement: string | _field: string | _time: time | _value: t0 | t1}]"),
                    // exactly one of (bucket, bucketID) must be specified
                    // exactly one of (org, orgID) must be specified
                    // https://github.com/influxdata/flux/issues/1660
                    "to" => Node::Builtin(r#"
                        forall [t0, t1] where t0: Row, t1: Row (
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
                    "buckets" => Node::Builtin(r#"
                        forall [] () -> [{
                            name: string |
                            id: string |
                            organizationID: string |
                            retentionPolicy: string |
                            retentionPeriod: int
                        }]
                    "#),
                }),

            }),
            "internal" => Node::Package(maplit::hashmap! {
                "gen" => Node::Package(maplit::hashmap! {
                    "tables" => Node::Builtin("forall [t0] (n: int, tags: [{name: string | cardinality: int}]) -> [{_time: time | _value: float | t0}]"),
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
                    "promHistogramQuantile" => Node::Builtin("forall [t0, t1] where t0: Row, t1: Row (<-tables: [t0], ?quantile: float, ?countColumn: string, ?upperBoundColumn: string, ?valueColumn: string) -> [t1]"),
                    "resets" => Node::Builtin("forall [t0, t1] (<-tables: [{_value: float | t0}]) -> [{_value: float | t1}]"),
                    "timestamp" => Node::Builtin("forall [t0] (<-tables: [{_value: float | t0}]) -> [{_value: float | t0}]"),
                    "promqlYear" => Node::Builtin("forall [] (timestamp: float) -> float"),
                    "join" => Node::Builtin("forall [t0, t1, t2] where t0: Row, t1: Row, t2: Row (left: [t0], right: [t1], fn: (left: t0, right: t1) -> t2) -> [t2]"),
                }),
            }),
            "json" => Node::Package(maplit::hashmap! {
                "encode" => Node::Builtin("forall [t0] (v: t0) -> bytes"),
            }),
            "kafka" => Node::Package(maplit::hashmap! {
                "to" => Node::Builtin(r#"
                    forall [t0] where t0: Row (
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
            "runtime" => Node::Package(maplit::hashmap! {
                "version" => Node::Builtin("forall [] () -> string"),
            }),
            "slack" => Node::Package(maplit::hashmap! {
                "validateColorString" => Node::Builtin("forall [] (color: string) -> string"),
            }),
            "socket" => Node::Package(maplit::hashmap! {
                "from" => Node::Builtin("forall [t0] (url: string, ?decoder: string) -> [t0]"),
            }),
            "sql" => Node::Package(maplit::hashmap! {
                "from" => Node::Builtin("forall [t0] (driverName: string, dataSourceName: string, query: string) -> [t0]"),
                "to" => Node::Builtin("forall [t0] (<-tables: [t0], driverName: string, dataSourceName: string, table: string, ?batchSize: int) -> [t0]"),
            }),
            "strings" => Node::Package(maplit::hashmap! {
                "title" => Node::Builtin("forall [] (v: string) -> string"),
                "toUpper" => Node::Builtin("forall [] (v: string) -> string"),
                "toLower" => Node::Builtin("forall [] (v: string) -> string"),
                "trim" => Node::Builtin("forall [] (v: string, cutset: string) -> string"),
                "trimPrefix" => Node::Builtin("forall [] (v: string, prefix: string) -> string"),
                "trimSpace" => Node::Builtin("forall [] (v: string) -> string"),
                "trimSuffix" => Node::Builtin("forall [] (v: string, suffix: string) -> string"),
                "trimRight" => Node::Builtin("forall [] (v: string, cutset: string) -> string"),
                "trimLeft" => Node::Builtin("forall [] (v: string, cutset: string) -> string"),
                "toTitle" => Node::Builtin("forall [] (v: string) -> string"),
                "hasPrefix" => Node::Builtin("forall [] (v: string, prefix: string) -> bool"),
                "hasSuffix" => Node::Builtin("forall [] (v: string, suffix: string) -> bool"),
                "containsStr" => Node::Builtin("forall [] (v: string, substr: string) -> bool"),
                "containsAny" => Node::Builtin("forall [] (v: string, chars: string) -> bool"),
                "equalFold" => Node::Builtin("forall [] (v: string, t: string) -> bool"),
                "compare" => Node::Builtin("forall [] (v: string, t: string) -> int"),
                "countStr" => Node::Builtin("forall [] (v: string, substr: string) -> int"),
                "index" => Node::Builtin("forall [] (v: string, substr: string) -> int"),
                "indexAny" => Node::Builtin("forall [] (v: string, chars: string) -> int"),
                "lastIndex" => Node::Builtin("forall [] (v: string, substr: string) -> int"),
                "lastIndexAny" => Node::Builtin("forall [] (v: string, chars: string) -> int"),
                "isDigit" => Node::Builtin("forall [] (v: string) -> bool"),
                "isLetter" => Node::Builtin("forall [] (v: string) -> bool"),
                "isLower" => Node::Builtin("forall [] (v: string) -> bool"),
                "isUpper" => Node::Builtin("forall [] (v: string) -> bool"),
                "repeat" => Node::Builtin("forall [] (v: string, count: int) -> string"),
                "replace" => Node::Builtin("forall [] (v: string, old: string, new: string, n: int) -> string"),
                "replaceAll" => Node::Builtin("forall [] (v: string, old: string, new: string) -> string"),
                "split" => Node::Builtin("forall [] (v: string, t: string) -> string"),
                "splitAfter" => Node::Builtin("forall [] (v: string, t: string) -> string"),
                "splitN" => Node::Builtin("forall [] (v: string, t: string, n: int) -> string"),
                "splitAfterN" => Node::Builtin("forall [] (v: string, t: string, i: int) -> string"),
                "joinStr" => Node::Builtin("forall [] (a: [string], v: string) -> {}"),
                "strlen" => Node::Builtin("forall [] (v: string) -> int"),
                "substring" => Node::Builtin("forall [] (v: string, start: int, end: int) -> string"),
            }),
            "system" => Node::Package(maplit::hashmap! {
                "time" => Node::Builtin("forall [] () -> time"),
            }),
            "testing" => Node::Package(maplit::hashmap! {
                "assertEquals" => Node::Builtin("forall [t0] (name: string, <-got: [t0], want: [t0]) -> [t0]"),
                "assertEmpty" => Node::Builtin("forall [t0] (<-tables: [t0]) -> [t0]"),
                "diff" => Node::Builtin("forall [t0] (<-got: [t0], want: [t0], ?verbose: bool) -> [{_diff: string | t0}]"),
            }),
            "universe" => Node::Package(maplit::hashmap! {
                "bool" => Node::Builtin("forall [t0] (v: t0) -> bool"),
                "bytes" => Node::Builtin("forall [t0] (v: t0) -> bytes"),
                "chandeMomentumOscillator" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        n: int,
                        ?columns: [string]
                    ) -> [t1]
                "#),
                "columns" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        column: string
                    ) -> [t1]
                "#),
                "contains" => Node::Builtin(r#"
                    forall [t0] where t0: Nullable (
                        value: t0,
                        set: [t0]
                    ) -> bool
                "#),
                "count" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: [string]
                    ) -> [t1]
                "#),
                "covariance" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?pearsonr: bool,
                        ?valueDst: string,
                        columns: [string]
                    ) -> [t1]
                "#),
                "cumulativeSum" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?columns: [string]
                    ) -> [t1]
                "#),
                "derivative" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?unit: duration,
                        ?nonNegative: bool,
                        ?columns: [string],
                        ?timeColumn: string
                    ) -> [t1]
                "#),
                "difference" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?nonNegative: bool,
                        ?columns: [string],
                        ?keepFirst: bool
                    ) -> [t1]
                "#),
                "distinct" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t1]
                "#),
                "drop" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?fn: (column: string) -> bool,
                        ?columns: [string]
                    ) -> [t1]
                "#),
                "duplicate" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        column: string,
                        as: string
                    ) -> [t1]
                "#),
                "duration" => Node::Builtin("forall [t0] (v: t0) -> duration"),
                "elapsed" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?unit: duration,
                        ?timeColumn: string,
                        ?columnName: string
                    ) -> [t1]
                "#),
                "exponentialMovingAverage" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Numeric (
                        <-tables: [{ _value: t0 | t1 }],
                        n: int
                    ) -> [{ _value: t0 | t1}]
                "#),
                "false" => Node::Builtin("forall [] bool"),
                "fill" => Node::Builtin(r#"
                    forall [t0, t1, t2] where t0: Row, t2: Row (
                        <-tables: [t0],
                        ?column: string,
                        value: [t1],
                        usePrevious: bool
                    ) -> [t2]
                "#),
                "filter" => Node::Builtin(r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        fn: (r: t0) -> bool,
                        ?onEmpty: string
                    ) -> [t0]
                "#),
                "first" => Node::Builtin(r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t0]
                "#),
                "float" => Node::Builtin("forall [t0] (v: t0) -> float"),
                "getColumn" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row (
                        <-table: [t0],
                        column: string
                    ) -> [t1]
                "#),
                "getRecord" => Node::Builtin(r#"
                    forall [t0] where t0: Row (
                        <-table: [t0],
                        idx: int
                    ) -> t0
                "#),
                "group" => Node::Builtin(r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        ?mode: string,
                        ?columns: [string]
                    ) -> [t0]
                "#),
                "histogram" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string,
                        ?upperBoundColumn: string,
                        ?countColumn: string,
                        bins: [float],
                        normalize: bool
                    ) -> [t1]
                "#),
                "histogramQuantile" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?quantile: float,
                        ?countColumn: string,
                        ?upperBoundColumn: string,
                        ?valueColumn: string,
                        ?minValue: float
                    ) -> [t1]
                "#),
                "holtWinters" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?withFit: bool,
                        ?column: string,
                        ?timeColumn: string,
                        n: int,
                        seasonality: int,
                        interval: duration
                    ) -> [t1]
                "#),
                "hourSelection" => Node::Builtin(r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        start: int,
                        stop: int,
                        ?timeColumn: string
                    ) -> [t0]
                "#),
                "inf" => Node::Builtin("forall [] duration"),
                "int" => Node::Builtin("forall [t0] (v: t0) -> int"),
                "integral" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?unit: duration,
                        ?timeColumn: string,
                        ?column: string
                    ) -> [t1]
                "#),
                "join" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: t0,
                        ?method: string,
                        ?on: [string]
                    ) -> [t1]
                "#),
                // This function would almost have input/output types that match, but:
                // input column may start as int, uint or float, and always ends up as float.
                // https://github.com/influxdata/flux/issues/2252
                "kaufmansAMA" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        n: int,
                        ?column: string
                    ) -> [t1]
                "#),
                // either column list or predicate must be provided
                // https://github.com/influxdata/flux/issues/2248
                "keep" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?columns: [string],
                        ?fn: (column: string) -> bool
                    ) -> [t1]
                "#),
                "keyValues" => Node::Builtin(r#"
                    forall [t0, t1, t2] where t0: Row, t2: Row (
                        <-tables: [t0],
                        ?keyColumns: [string]
                    ) -> [{_key: string | _value: t1 | t2}]
                "#),
                "keys" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t1]
                "#),
                "last" => Node::Builtin("forall [t0] where t0: Row (<-tables: [t0], ?column: string) -> [t0]"),
                "length" => Node::Builtin("forall [t0] (arr: [t0]) -> int"),
                "limit"  => Node::Builtin("forall [t0] (<-tables: [t0], n: int, ?offset: int) -> [t0]"),
                "linearBins" => Node::Builtin(r#"
                    forall [] (
                        start: float,
                        width: float,
                        count: int,
                        ?infinity: bool
                    ) -> [float]
                "#),
                "logarithmicBins" => Node::Builtin(r#"
                    forall [] (
                        start: float,
                        factor: float,
                        count: int,
                        ?infinity: bool
                    ) -> [float]
                "#),
                // Note: mergeKey parameter could be removed from map once the transpiler is updated:
                // https://github.com/influxdata/flux/issues/816
                "map" => Node::Builtin("forall [t0, t1] (<-tables: [t0], fn: (r: t0) -> t1, ?mergeKey: bool) -> [t1]"),
                "max" => Node::Builtin("forall [t0] where t0: Row (<-tables: [t0], ?column: string) -> [t0]"),
                "mean" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t1]
                "#),
                "min" => Node::Builtin("forall [t0] where t0: Row (<-tables: [t0], ?column: string) -> [t0]"),
                "mode" => Node::Builtin(r#"
                    forall [t0, t1, t2] where t0: Row, t2: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [{_value: t1 | t2}]
                "#),
                "movingAverage" => Node::Builtin("forall [t0, t1] where t0: Numeric (<-tables: [{_value: t0 | t1}], n: int) -> [{_value: float | t1}]"),
                "pivot" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        rowKey: [string],
                        columnKey: [string],
                        valueColumn: string
                    ) -> [t1]
                "#),
                "quantile" => Node::Builtin(r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        ?column: string,
                        q: float,
                        ?compression: float,
                        ?method: string
                    ) -> [t0]
                "#),
                // start and stop should be able to constrained to time or duration with a kind constraint:
                //   https://github.com/influxdata/flux/issues/2243
                // Also, we should remove the column arguments so we can reuse t0 in the return type:
                //   https://github.com/influxdata/flux/issues/2253
                "range" => Node::Builtin(r#"
                    forall [t0, t1, t2, t3] where t0: Row, t3: Row (
                        <-tables: [t0],
                        start: t1,
                        ?stop: t2,
                        ?timeColumn: string,
                        ?startColumn: string,
                        ?stopColumn: string
                    ) -> [t3]
                "#),
                // This function could be updated to get better type inference:
                //   https://github.com/influxdata/flux/issues/2254
                "reduce" => Node::Builtin(r#"
                    forall [t0, t1, t2] where t0: Row, t1: Row, t2: Row (
                        <-tables: [t0],
                        fn: (r: t0, accumulator: t1) -> t1,
                        identity: t1
                    ) -> [t2]
                "#),
                "relativeStrengthIndex" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        n: int,
                        ?columns: [string]
                    ) -> [t1]
                "#),
                // Either fn or columns should be specified
                // https://github.com/influxdata/flux/issues/2251
                "rename" => Node::Builtin(r#"
                    forall [t0, t1, t2] where t0: Row, t1: Row, t2: Row (
                        <-tables: [t0],
                        ?fn: (column: string) -> string,
                        ?columns: t1
                    ) -> [t2]
                "#),
                "sample" => Node::Builtin(r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        n: int,
                        ?pos: int,
                        ?column: string
                    ) -> [t0]
                "#),
                "set" => Node::Builtin(r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        key: string,
                        value: string
                    ) -> [t0]
                "#),
                // This is an aggregate function, and may clobber value columns
                "skew" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t1]
                "#),
                "sleep" => Node::Builtin(r#"
                    forall [t0] (
                        <-v: t0,
                        "duration": duration
                    ) -> t0
                "#),
                "sort" => Node::Builtin(r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        ?columns: [string],
                        ?desc: bool
                    ) -> [t0]
                "#),
                "spread" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t1]
                "#),
                "stateTracking" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        fn: (r: t0) -> bool,
                        ?countColumn: string,
                        ?durationColumn: string,
                        ?durationUnit: duration,
                        ?timeColumn: string
                    ) -> [t1]
                "#),
                "stddev" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string,
                        mode: string
                    ) -> [t1]
                "#),
                "string" => Node::Builtin("forall [t0] (v: t0) -> string"),
                "sum" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t1]
                "#),
                "tableFind" => Node::Builtin(r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        fn: (key: t1) -> bool
                    ) -> [t0]
                "#),
                "tail" => Node::Builtin(r#"
                    forall [t0] (
                        <-tables: [t0],
                        n: int,
                        ?offset: int
                    ) -> [t0]
                "#),
                "time" => Node::Builtin("forall [t0] (v: t0) -> time"),
                "timeShift" => Node::Builtin(r#"
                    forall [t0] (
                        <-tables: [t0],
                        "duration": duration,
                        ?columns: [string]
                    ) -> [t0]
                "#),
                "tripleExponentialDerivative" => Node::Builtin(r#"
                    forall [t0] where t0: Numeric, t1: Row (
                        <-tables: [{_value: t0 | t1}],
                        n: int
                    ) -> [{_value: float | t1}]
                "#),
                "true" => Node::Builtin("forall [] bool"),
                "uint" => Node::Builtin("forall [t0] (v: t0) -> uint"),
                "union" => Node::Builtin(r#"
                    forall [t0] where t0: Row (
                        tables: [[t0]]
                    ) -> [t0]
                "#),
                "unique" => Node::Builtin(r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t0]
                "#),
                // This would produce an output the same as the input,
                // except that startColumn and stopColumn will be added if they don't
                // already exist.
                // https://github.com/influxdata/flux/issues/2255
                "window" => Node::Builtin(r#"
                    forall [t0] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?every: duration,
                        ?period: duration,
                        ?offset: duration,
                        ?timeColumn: string,
                        ?startColumn: string,
                        ?stopColumn: string,
                        ?createEmpty: bool
                    ) -> [t1]
                "#),
                "yield" => Node::Builtin(r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        ?name: string
                    ) -> [t0]
                "#),
            }),
        },
    }
}

pub struct NodeIterator<'a> {
    path_elems: Vec<&'a str>,
    iter_stack: Vec<hash_map::Iter<'a, &'a str, Node<'a>>>,
}

impl<'a> NodeIterator<'a> {
    pub fn new(builtins: &'a Builtins) -> NodeIterator<'a> {
        NodeIterator {
            path_elems: Vec::new(),
            iter_stack: vec![builtins.pkgs.iter()],
        }
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
                        self.iter_stack.push(m.iter());
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
