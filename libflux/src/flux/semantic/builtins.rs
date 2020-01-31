use crate::semantic::fresh::{Fresh, Fresher};
use crate::semantic::import::Importer;
use crate::semantic::parser::parse;
use maplit::hashmap;
use std::collections::hash_map::Iter;
use std::collections::HashMap;

pub struct Builtins<'a> {
    pkgs: HashMap<&'a str, HashMap<&'a str, &'a str>>,
}

impl<'a> Builtins<'a> {
    pub fn iter(&'a self) -> Iter<&'a str, HashMap<&'a str, &'a str>> {
        self.pkgs.iter()
    }

    pub fn importer_for(&'a self, pkgpath: &str, f: &mut Fresher) -> impl Importer {
        let mut h = HashMap::new();
        if let Some(values) = self.pkgs.get(pkgpath) {
            for (name, expr) in values {
                let pty = parse(expr).unwrap().fresh(f, &mut HashMap::new());
                h.insert((*name).to_string(), pty);
            }
        }
        h
    }
}

pub fn builtins() -> Builtins<'static> {
    Builtins {
        pkgs: hashmap! {
            "csv" => maplit::hashmap! {
                // This is a "provide exactly one argument" function
                // https://github.com/influxdata/flux/issues/2249
                "from" => "forall [t0] where t0: Row (?csv: string, ?file: string) -> [t0]",
            },
            "date" => maplit::hashmap! {
                 "second" => "forall [] (t: time) -> int",
                 "minute" => "forall [] (t: time) -> int",
                 "hour" => "forall [] (t: time) -> int",
                 "weekDay" => "forall [] (t: time) -> int",
                 "monthDay" => "forall [] (t: time) -> int",
                 "yearDay" => "forall [] (t: time) -> int",
                 "month" => "forall [] (t: time) -> int",
                 "year" => "forall [] (t: time) -> int",
                 "week" => "forall [] (t: time) -> int",
                 "quarter" => "forall [] (t: time) -> int",
                 "millisecond" => "forall [] (t: time) -> int",
                 "microsecond" => "forall [] (t: time) -> int",
                 "nanosecond" => "forall [] (t: time) -> int",
                 "truncate" => "forall [] (t: time, unit: duration) -> time",
            },
            "experimental/bigtable" => maplit::hashmap! {
                     "from" => "forall [t0] where t0: Row (token: string, project: string, instance: string, table: string) -> [t0]",
            },
            "experimental/geo" => maplit::hashmap! {
                     "getGrid" => "forall [] (?box: {minLat: float | minLon: float | maxLat: float | maxLon: float}, ?minSize: int, ?maxSize: int, ?precision: int, ?maxPrecision: int) -> {precision: int | set: [string]}",
                     "containsTag" => "forall [t0] where t0: Row (row: t0, tagKey: string, set: [string]) -> bool",
            },
            "experimental/http" => maplit::hashmap! {
                "get" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        url: string,
                        ?headers: t0,
                        ?timeout: duration
                    ) -> {statusCode: int | body: bytes | headers: t1}
                "#,
            },
            "experimental/mqtt" => maplit::hashmap! {
                "to" => r#"
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
                "#,
            },
            "experimental/prometheus" => maplit::hashmap! {
                "scrape" => "forall [t0] where t0: Row (url: string) -> [t0]",
            },
            "experimental" => maplit::hashmap! {
                 "addDuration" => "forall [] (d: duration, to: time) -> time",
                 "subDuration" => "forall [] (d: duration, from: time) -> time",
                 "group" => "forall [t0] where t0: Row (<-tables: [t0], mode: string, columns: [string]) -> [t0]",
                 "objectKeys" => "forall [t0] where t0: Row (o: t0) -> [string]",
                 "set" => "forall [t0, t1, t2] where t0: Row, t1: Row, t2: Row (<-tables: [t0], o: t1) -> [t2]",
                 // must specify exactly one of bucket, bucketID
                 // must specify exactly one of org, orgID
                 // if host is specified, token must be too.
                 // https://github.com/influxdata/flux/issues/1660
                 "to" => "forall [t0] where t0: Row (<-tables: [t0], ?bucket: string, ?bucketID: string, ?org: string, ?orgID: string, ?host: string, ?token: string) -> [t0]",
            },
            "generate" => maplit::hashmap! {
                "from" => "forall [] (start: time, stop: time, count: int, fn: (n: int) -> int) -> [{ _start: time | _stop: time | _time: time | _value:int }]",
            },
            "http" => maplit::hashmap! {
                "post" => "forall [t0] where t0: Row (url: string, ?headers: t0, ?data: bytes) -> int",
                "basicAuth" => "forall [] (u: string, p: string) -> string",
            },
            "influxdata/influxdb/secrets" => maplit::hashmap! {
                "get" => "forall [] (key: string) -> string",
            },
            "influxdata/influxdb/v1" => maplit::hashmap! {
                // exactly one of json and file must be specified
                // https://github.com/influxdata/flux/issues/2250
                "json" => "forall [t0] where t0: Row (?json: string, ?file: string) -> [t0]",
                "databases" => r#"
                    forall [] () -> [{
                        organizationID: string |
                        databaseName: string |
                        retentionPolicy: string |
                        retentionPeriod: int |
                        default: bool |
                        bucketID: string
                    }]
                "#,
            },
            "influxdata/influxdb" => maplit::hashmap! {
                // This is a one-or-the-other parameters function
                // https://github.com/influxdata/flux/issues/1659
                "from" => "forall [t0, t1] (?bucket: string, ?bucketID: string) -> [{_measurement: string | _field: string | _time: time | _value: t0 | t1}]",
                // exactly one of (bucket, bucketID) must be specified
                // exactly one of (org, orgID) must be specified
                // https://github.com/influxdata/flux/issues/1660
                "to" => r#"
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
                    ) -> [t0]
                "#,
                "buckets" => r#"
                    forall [] () -> [{
                        name: string |
                        id: string |
                        organizationID: string |
                        retentionPolicy: string |
                        retentionPeriod: int
                    }]
                "#,
            },
            "internal/gen" => maplit::hashmap! {
                "tables" => "forall [t0] (n: int, tags: [{name: string | cardinality: int}]) -> [{_time: time | _value: float | t0}]",
            },
            "internal/promql" => maplit::hashmap! {
                "changes" => "forall [t0, t1] (<-tables: [{_value: float | t0}]) -> [{_value: float | t1}]",
                "promqlDayOfMonth" => "forall [] (timestamp: float) -> float",
                "promqlDayOfWeek" => "forall [] (timestamp: float) -> float",
                "promqlDaysInMonth" => "forall [] (timestamp: float) -> float",
                "emptyTable" => "forall [] () -> [{_start: time | _stop: time | _time: time | _value: float}]",
                "extrapolatedRate" => "forall [t0, t1] (<-tables: [{_start: time | _stop: time | _time: time | _value: float | t0}], ?isCounter: bool, ?isRate: bool) -> [{_value: float | t1}]",
                "holtWinters" => "forall [t0, t1] (<-tables: [{_time: time | _value: float | t0}], ?smoothingFactor: float, ?trendFactor: float) -> [{_value: float | t1}]",
                "promqlHour" => "forall [] (timestamp: float) -> float",
                "instantRate" => "forall [t0, t1] (<-tables: [{_time: time | _value: float | t0}], ?isRate: bool) -> [{_value: float | t1}]",
                "labelReplace" => "forall [t0, t1] (<-tables: [{_value: float | t0}], source: string, destination: string, regex: string, replacement: string) -> [{_value: float | t1}]",
                "linearRegression" => "forall [t0, t1] (<-tables: [{_time: time | _stop: time | _value: float | t0}], ?predict: bool, ?fromNow: float) -> [{_value: float | t1}]",
                "promqlMinute" => "forall [] (timestamp: float) -> float",
                "promqlMonth" => "forall [] (timestamp: float) -> float",
                "promHistogramQuantile" => "forall [t0, t1] where t0: Row, t1: Row (<-tables: [t0], ?quantile: float, ?countColumn: string, ?upperBoundColumn: string, ?valueColumn: string) -> [t1]",
                "resets" => "forall [t0, t1] (<-tables: [{_value: float | t0}]) -> [{_value: float | t1}]",
                "timestamp" => "forall [t0] (<-tables: [{_value: float | t0}]) -> [{_value: float | t0}]",
                "promqlYear" => "forall [] (timestamp: float) -> float",
                "join" => "forall [t0, t1, t2] where t0: Row, t1: Row, t2: Row (left: [t0], right: [t1], fn: (left: t0, right: t1) -> t2) -> [t2]",
            },
            "json" => maplit::hashmap! {
                "encode" => "forall [t0] (v: t0) -> bytes",
            },
            "kafka" => maplit::hashmap! {
                "to" => r#"
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
                    ) -> [t0]"#,
            },
            "math" => maplit::hashmap! {
                "pi" => "forall [] float",
                "e" => "forall [] float",
                "phi" => "forall [] float",
                "sqrt2" => "forall [] float",
                "sqrte" => "forall [] float",
                "sqrtpi" => "forall [] float",
                "sqrtphi" => "forall [] float",
                "log2e" => "forall [] float",
                "ln2" => "forall [] float",
                "ln10" => "forall [] float",
                "log10e" => "forall [] float",

                "maxfloat" => "forall [] float",
                "smallestNonzeroFloat" => "forall [] float",
                "maxint" => "forall [] int",
                "minint" => "forall [] int",
                "maxuint" => "forall [] uint",

                "abs" => "forall [] (x: float) -> float",
                "acos" => "forall [] (x: float) -> float",
                "acosh" => "forall [] (x: float) -> float",
                "asin" => "forall [] (x: float) -> float",
                "asinh" => "forall [] (x: float) -> float",
                "atan" => "forall [] (x: float) -> float",
                "atan2" => "forall [] (x: float, y: float) -> float",
                "atanh" => "forall [] (x: float) -> float",
                "cbrt" => "forall [] (x: float) -> float",
                "ceil" => "forall [] (x: float) -> float",
                "copysign" => "forall [] (x: float, y: float) -> float",
                "cos" => "forall [] (x: float) -> float",
                "cosh" => "forall [] (x: float) -> float",
                "dim" => "forall [] (x: float, y: float) -> float",
                "erf" => "forall [] (x: float) -> float",
                "erfc" => "forall [] (x: float) -> float",
                "erfcinv" => "forall [] (x: float) -> float",
                "erfinv" => "forall [] (x: float) -> float",
                "exp" => "forall [] (x: float) -> float",
                "exp2" => "forall [] (x: float) -> float",
                "expm1" => "forall [] (x: float) -> float",
                "floor" => "forall [] (x: float) -> float",
                "gamma" => "forall [] (x: float) -> float",
                "hypot" => "forall [] (x: float, y: float) -> float",
                "j0" => "forall [] (x: float) -> float",
                "j1" => "forall [] (x: float) -> float",
                "log" => "forall [] (x: float) -> float",
                "log10" => "forall [] (x: float) -> float",
                "log1p" => "forall [] (x: float) -> float",
                "log2" => "forall [] (x: float) -> float",
                "logb" => "forall [] (x: float) -> float",
                "mMax" => "forall [] (x: float, y: float) -> float",
                "mMin" => "forall [] (x: float, y: float) -> float",
                "mod" => "forall [] (x: float, y: float) -> float",
                "nextafter" => "forall [] (x: float, y: float) -> float",
                "pow" => "forall [] (x: float, y: float) -> float",
                "remainder" => "forall [] (x: float, y: float) -> float",
                "round" => "forall [] (x: float) -> float",
                "roundtoeven" => "forall [] (x: float) -> float",
                "sin" => "forall [] (x: float) -> float",
                "sinh" => "forall [] (x: float) -> float",
                "sqrt" => "forall [] (x: float) -> float",
                "tan" => "forall [] (x: float) -> float",
                "tanh" => "forall [] (x: float) -> float",
                "trunc" => "forall [] (x: float) -> float",
                "y0" => "forall [] (x: float) -> float",
                "y1" => "forall [] (x: float) -> float",

                "float64bits" => "forall [] (f: float) -> uint",
                "float64frombits" => "forall [] (b: uint) -> float",
                "ilogb" => "forall [] (x: float) -> int",
                "frexp" => "forall [] (f: float) -> {frac: float | exp: int}",
                "lgamma" => "forall [] (x: float) -> {lgamma: float | sign: int}",
                "modf" => r#"forall [] (f: float) -> {"int": float | frac: float}"#,
                "sincos" => "forall [] (x: float) -> {sin: float | cos: float}",
                "isInf" => "forall [] (f: float, sign: int) -> bool",
                "isNaN" => "forall [] (f: float) -> bool",
                "signbit" => "forall [] (x: float) -> bool",
                "NaN" => "forall [] () -> float",
                "mInf" => "forall [] (sign: int) -> float",
                "jn" => "forall [] (n: int, x: float) -> float",
                "yn" => "forall [] (n: int, x: float) -> float",
                "ldexp" => "forall [] (frac: float, exp: int) -> float",
                "pow10" => "forall [] (n: int) -> float",
            },
            "pagerduty" => maplit::hashmap! {
                "dedupKey" => "forall [t0] (<-tables: [t0]) -> [{_pagerdutyDedupKey: string | t0}]",
            },
            "regexp" => maplit::hashmap! {
                "compile" => "forall [] (v: string) -> regexp",
                "quoteMeta" => "forall [] (v: string) -> string",
                "findString" => "forall [] (r: regexp, v: string) -> string",
                "findStringIndex" => "forall [] (r: regexp, v: string) -> [int]",
                "matchRegexpString" => "forall [] (r: regexp, v: string) -> bool",
                "replaceAllString" => "forall [] (r: regexp, v: string, t: string) -> string",
                "splitRegexp" => "forall [] (r: regexp, v: string, i: int) -> [string]",
                "getString" => "forall [] (r: regexp) -> string",
            },
            "runtime" => maplit::hashmap! {
                "version" => "forall [] () -> string",
            },
            "slack" => maplit::hashmap! {
                "validateColorString" => "forall [] (color: string) -> string",
            },
            "socket" => maplit::hashmap! {
                "from" => "forall [t0] (url: string, ?decoder: string) -> [t0]",
            },
            "sql" => maplit::hashmap! {
                "from" => "forall [t0] (driverName: string, dataSourceName: string, query: string) -> [t0]",
                "to" => "forall [t0] (<-tables: [t0], driverName: string, dataSourceName: string, table: string, ?batchSize: int) -> [t0]",
            },
            "strings" => maplit::hashmap! {
                "title" => "forall [] (v: string) -> string",
                "toUpper" => "forall [] (v: string) -> string",
                "toLower" => "forall [] (v: string) -> string",
                "trim" => "forall [] (v: string, cutset: string) -> string",
                "trimPrefix" => "forall [] (v: string, prefix: string) -> string",
                "trimSpace" => "forall [] (v: string) -> string",
                "trimSuffix" => "forall [] (v: string, suffix: string) -> string",
                "trimRight" => "forall [] (v: string, cutset: string) -> string",
                "trimLeft" => "forall [] (v: string, cutset: string) -> string",
                "toTitle" => "forall [] (v: string) -> string",
                "hasPrefix" => "forall [] (v: string, prefix: string) -> bool",
                "hasSuffix" => "forall [] (v: string, suffix: string) -> bool",
                "containsStr" => "forall [] (v: string, substr: string) -> bool",
                "containsAny" => "forall [] (v: string, chars: string) -> bool",
                "equalFold" => "forall [] (v: string, t: string) -> bool",
                "compare" => "forall [] (v: string, t: string) -> int",
                "countStr" => "forall [] (v: string, substr: string) -> int",
                "index" => "forall [] (v: string, substr: string) -> int",
                "indexAny" => "forall [] (v: string, chars: string) -> int",
                "lastIndex" => "forall [] (v: string, substr: string) -> int",
                "lastIndexAny" => "forall [] (v: string, chars: string) -> int",
                "isDigit" => "forall [] (v: string) -> bool",
                "isLetter" => "forall [] (v: string) -> bool",
                "isLower" => "forall [] (v: string) -> bool",
                "isUpper" => "forall [] (v: string) -> bool",
                "repeat" => "forall [] (v: string, count: int) -> string",
                "replace" => "forall [] (v: string, old: string, new: string, n: int) -> string",
                "replaceAll" => "forall [] (v: string, old: string, new: string) -> string",
                "split" => "forall [] (v: string, t: string) -> string",
                "splitAfter" => "forall [] (v: string, t: string) -> string",
                "splitN" => "forall [] (v: string, t: string, n: int) -> string",
                "splitAfterN" => "forall [] (v: string, t: string, i: int) -> string",
                "joinStr" => "forall [] (a: [string], v: string) -> {}",
                "strlen" => "forall [] (v: string) -> int",
                "substring" => "forall [] (v: string, start: int, end: int) -> string",
            },
            "system" => maplit::hashmap! {
                "time" => "forall [] () -> time",
            },
            "testing" => maplit::hashmap! {
                "assertEquals" => "forall [t0] (name: string, <-got: [t0], want: [t0]) -> [t0]",
                "assertEmpty" => "forall [t0] (<-tables: [t0]) -> [t0]",
                "diff" => "forall [t0] (<-got: [t0], want: [t0], ?verbose: bool) -> [{_diff: string | t0}]",
            },
            "universe" => maplit::hashmap! {
                "bool" => "forall [t0] (v: t0) -> bool",
                "bytes" => "forall [t0] (v: t0) -> bytes",
                "chandeMomentumOscillator" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        n: int,
                        ?columns: [string]
                    ) -> [t1]
                "#,
                "columns" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        column: string
                    ) -> [t1]
                "#,
                "contains" => r#"
                    forall [t0] where t0: Nullable (
                        value: t0,
                        set: [t0]
                    ) -> bool
                "#,
                "count" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t1]
                "#,
                "covariance" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?pearsonr: bool,
                        ?valueDst: string,
                        columns: [string]
                    ) -> [t1]
                "#,
                "cumulativeSum" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?columns: [string]
                    ) -> [t1]
                "#,
                "derivative" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?unit: duration,
                        ?nonNegative: bool,
                        ?columns: [string],
                        ?timeColumn: string
                    ) -> [t1]
                "#,
                "difference" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?nonNegative: bool,
                        ?columns: [string],
                        ?keepFirst: bool
                    ) -> [t1]
                "#,
                "distinct" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t1]
                "#,
                "drop" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?fn: (column: string) -> bool,
                        ?columns: [string]
                    ) -> [t1]
                "#,
                "duplicate" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        column: string,
                        as: string
                    ) -> [t1]
                "#,
                "duration" => "forall [t0] (v: t0) -> duration",
                "elapsed" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?unit: duration,
                        ?timeColumn: string,
                        ?columnName: string
                    ) -> [t1]
                "#,
                "exponentialMovingAverage" => r#"
                    forall [t0, t1] where t0: Numeric (
                        <-tables: [{ _value: t0 | t1 }],
                        n: int
                    ) -> [{ _value: t0 | t1}]
                "#,
                "false" => "forall [] bool",
                "fill" => r#"
                    forall [t0, t1, t2] where t0: Row, t2: Row (
                        <-tables: [t0],
                        ?column: string,
                        value: t1,
                        usePrevious: bool
                    ) -> [t2]
                "#,
                "filter" => r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        fn: (r: t0) -> bool,
                        ?onEmpty: string
                    ) -> [t0]
                "#,
                "first" => r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t0]
                "#,
                "float" => "forall [t0] (v: t0) -> float",
                "getColumn" => r#"
                    forall [t0, t1] where t0: Row (
                        <-table: [t0],
                        column: string
                    ) -> [t1]
                "#,
                "getRecord" => r#"
                    forall [t0] where t0: Row (
                        <-table: [t0],
                        idx: int
                    ) -> t0
                "#,
                "group" => r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        ?mode: string,
                        ?columns: [string]
                    ) -> [t0]
                "#,
                "histogram" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string,
                        ?upperBoundColumn: string,
                        ?countColumn: string,
                        bins: [float],
                        normalize: bool
                    ) -> [t1]
                "#,
                "histogramQuantile" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?quantile: float,
                        ?countColumn: string,
                        ?upperBoundColumn: string,
                        ?valueColumn: string,
                        ?minValue: float
                    ) -> [t1]
                "#,
                "holtWinters" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?withFit: bool,
                        ?column: string,
                        ?timeColumn: string,
                        n: int,
                        seasonality: int,
                        interval: duration
                    ) -> [t1]
                "#,
                "hourSelection" => r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        start: int,
                        stop: int,
                        ?timeColumn: string
                    ) -> [t0]
                "#,
                "inf" => "forall [] duration",
                "int" => "forall [t0] (v: t0) -> int",
                "integral" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?unit: duration,
                        ?timeColumn: string,
                        ?column: string
                    ) -> [t1]
                "#,
                "join" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: t0,
                        ?method: string,
                        ?on: [string]
                    ) -> [t1]
                "#,
                // This function would almost have input/output types that match, but:
                // input column may start as int, uint or float, and always ends up as float.
                // https://github.com/influxdata/flux/issues/2252
                "kaufmansAMA" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        n: int,
                        ?column: string
                    ) -> [t1]
                "#,
                // either column list or predicate must be provided
                // https://github.com/influxdata/flux/issues/2248
                "keep" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?columns: [string],
                        ?fn: (column: string) -> bool
                    ) -> [t1]
                "#,
                "keyValues" => r#"
                    forall [t0, t1, t2] where t0: Row, t2: Row (
                        <-tables: [t0],
                        ?keyColumns: [string]
                    ) -> [{_key: string | _value: t1 | t2}]
                "#,
                "keys" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t1]
                "#,
                "last" => "forall [t0] where t0: Row (<-tables: [t0], ?column: string) -> [t0]",
                "length" => "forall [t0] (arr: [t0]) -> int",
                "limit"  => "forall [t0] (<-tables: [t0], n: int, ?offset: int) -> [t0]",
                "linearBins" => r#"
                    forall [] (
                        start: float,
                        width: float,
                        count: int,
                        ?infinity: bool
                    ) -> [float]
                "#,
                "logarithmicBins" => r#"
                    forall [] (
                        start: float,
                        factor: float,
                        count: int,
                        ?infinity: bool
                    ) -> [float]
                "#,
                // Note: mergeKey parameter could be removed from map once the transpiler is updated:
                // https://github.com/influxdata/flux/issues/816
                "map" => "forall [t0, t1] (<-tables: [t0], fn: (r: t0) -> t1, ?mergeKey: bool) -> [t1]",
                "max" => "forall [t0] where t0: Row (<-tables: [t0], ?column: string) -> [t0]",
                "mean" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t1]
                "#,
                "min" => "forall [t0] where t0: Row (<-tables: [t0], ?column: string) -> [t0]",
                "mode" => r#"
                    forall [t0, t1, t2] where t0: Row, t2: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [{_value: t1 | t2}]
                "#,
                "movingAverage" => "forall [t0, t1] where t0: Numeric (<-tables: [{_value: t0 | t1}], n: int) -> [{_value: float | t1}]",
                "pivot" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        rowKey: [string],
                        columnKey: [string],
                        valueColumn: string
                    ) -> [t1]
                "#,
                "quantile" => r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        ?column: string,
                        q: float,
                        ?compression: float,
                        ?method: string
                    ) -> [t0]
                "#,
                // start and stop should be able to constrained to time or duration with a kind constraint:
                //   https://github.com/influxdata/flux/issues/2243
                // Also, we should remove the column arguments so we can reuse t0 in the return type:
                //   https://github.com/influxdata/flux/issues/2253
                "range" => r#"
                    forall [t0, t1, t2, t3] where t0: Row, t3: Row (
                        <-tables: [t0],
                        start: t1,
                        ?stop: t2,
                        ?timeColumn: string,
                        ?startColumn: string,
                        ?stopColumn: string
                    ) -> [t3]
                "#,
                // This function could be updated to get better type inference:
                //   https://github.com/influxdata/flux/issues/2254
                "reduce" => r#"
                    forall [t0, t1, t2] where t0: Row, t1: Row, t2: Row (
                        <-tables: [t0],
                        fn: (r: t0, accumulator: t1) -> t1,
                        identity: t1
                    ) -> [t2]
                "#,
                "relativeStrengthIndex" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        n: int,
                        ?columns: [string]
                    ) -> [t1]
                "#,
                // Either fn or columns should be specified
                // https://github.com/influxdata/flux/issues/2251
                "rename" => r#"
                    forall [t0, t1, t2] where t0: Row, t1: Row, t2: Row (
                        <-tables: [t0],
                        ?fn: (column: string) -> string,
                        ?columns: t1
                    ) -> [t2]
                "#,
                "sample" => r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        n: int,
                        ?pos: int,
                        ?column: string
                    ) -> [t0]
                "#,
                "set" => r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        key: string,
                        value: string
                    ) -> [t0]
                "#,
                // This is an aggregate function, and may clobber value columns
                "skew" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t1]
                "#,
                "sleep" => r#"
                    forall [t0] (
                        <-v: t0,
                        "duration": duration
                    ) -> t0
                "#,
                "sort" => r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        ?columns: [string],
                        ?desc: bool
                    ) -> [t0]
                "#,
                "spread" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t1]
                "#,
                "stateTracking" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        fn: (r: t0) -> bool,
                        ?countColumn: string,
                        ?durationColumn: string,
                        ?durationUnit: duration,
                        ?timeColumn: string
                    ) -> [t1]
                "#,
                "stddev" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string,
                        mode: string
                    ) -> [t1]
                "#,
                "string" => "forall [t0] (v: t0) -> string",
                "sum" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t1]
                "#,
                "tableFind" => r#"
                    forall [t0, t1] where t0: Row, t1: Row (
                        <-tables: [t0],
                        fn: (key: t1) -> bool
                    ) -> [t0]
                "#,
                "tail" => r#"
                    forall [t0] (
                        <-tables: [t0],
                        n: int,
                        ?offset: int
                    ) -> [t0]
                "#,
                "time" => "forall [t0] (v: t0) -> time",
                "timeShift" => r#"
                    forall [t0] (
                        <-tables: [t0],
                        "duration": duration,
                        ?columns: [string]
                    ) -> [t0]
                "#,
                "tripleExponentialDerivative" => r#"
                    forall [t0] where t0: Numeric, t1: Row (
                        <-tables: [{_value: t0 | t1}],
                        n: int
                    ) -> [{_value: float | t1}]
                "#,
                "true" => "forall [] bool",
                "uint" => "forall [t0] (v: t0) -> uint",
                "union" => r#"
                    forall [t0] where t0: Row (
                        tables: [[t0]]
                    ) -> [t0]
                "#,
                "unique" => r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        ?column: string
                    ) -> [t0]
                "#,
                // This would produce an output the same as the input,
                // except that startColumn and stopColumn will be added if they don't
                // already exist.
                // https://github.com/influxdata/flux/issues/2255
                "window" => r#"
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
                "#,
                "yield" => r#"
                    forall [t0] where t0: Row (
                        <-tables: [t0],
                        ?name: string
                    ) -> [t0]
                "#,
            },
        },
    }
}

#[cfg(test)]
mod test {
    use crate::semantic::builtins::builtins;
    use crate::semantic::parser as type_parser;

    #[test]
    fn parse_builtin_types() {
        for (path, values) in builtins().iter() {
            for (name, expr) in values {
                match type_parser::parse(expr) {
                    Ok(_) => {}
                    Err(s) => {
                        let msg = format!("{}.{} type failed to parse: {}", path, name, s);
                        panic!(msg)
                    }
                }
            }
        }
    }
}
