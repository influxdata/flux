use crate::ast::get_err_type_expression;
use crate::parser;
use crate::semantic::convert::convert_polytype;
use crate::semantic::fresh::Fresher;
use crate::semantic::import::Importer;
use crate::semantic::types::{PolyTypeMap, SemanticMap, SemanticMapIter};
type BuiltinsMapValue<'a> = SemanticMap<&'a str, &'a str>;
type BuiltinsMap<'a> = SemanticMap<&'a str, SemanticMap<&'a str, &'a str>>;

pub struct Builtins<'a> {
    pkgs: BuiltinsMap<'a>,
}

impl<'a> Builtins<'a> {
    pub fn iter(&'a self) -> SemanticMapIter<&'a str, BuiltinsMapValue<'a>> {
        self.pkgs.iter()
    }

    pub fn importer_for(&'a self, pkgpath: &str, f: &mut Fresher) -> impl Importer {
        let mut h = PolyTypeMap::new();
        if let Some(values) = self.pkgs.get(pkgpath) {
            for (name, expr) in values {
                // let pty = parser::Parse(expr).unwrap().fresh(f, &mut TvarMap::new());
                let mut p = parser::Parser::new(expr);

                let typ_expr = p.parse_type_expression();
                let err = get_err_type_expression(typ_expr.clone());

                if err != "" {
                    let msg = format!("TypeExpression parsing failed for {}. {:?}", name, err);
                    panic!(msg)
                }
                let pty = convert_polytype(typ_expr, f);

                if let Ok(p) = pty {
                    h.insert((*name).to_string(), p);
                }
            }
        }
        h
    }
}

pub fn builtins() -> Builtins<'static> {
    Builtins {
        pkgs: semantic_map! {
            "csv" => semantic_map! {
                // This is a "provide exactly one argument" function
                // https://github.com/influxdata/flux/issues/2249
                "from" => "(?csv: string, ?file: string) => [A] where A: Row",
            },
            "date" => semantic_map! {
                 "second" => "(t: T) => int where T: Timeable",
                 "minute" => "(t: T) => int where T: Timeable",
                 "hour" => "(t: T) => int where T: Timeable",
                 "weekDay" => "(t: T) => int where T: Timeable",
                 "monthDay" => "(t: T) => int where T: Timeable",
                 "yearDay" => "(t: T) => int where T: Timeable",
                 "month" => "(t: T) => int where T: Timeable",
                 "year" => "(t: T) => int where T: Timeable",
                 "week" => "(t: T) => int where T: Timeable",
                 "quarter" => "(t: T) => int where T: Timeable",
                 "millisecond" => "(t: T) => int where T: Timeable",
                 "microsecond" => "(t: T) => int where T: Timeable",
                 "nanosecond" => "(t: T) => int where T: Timeable",
                 "truncate" => "(t: T, unit: duration) => time where T : Timeable",
            },
            "experimental/array" => semantic_map! {
                "from" => "(rows: [A]) => [A] where A: Row ",
            },
            "experimental/bigtable" => semantic_map! {
                     "from" => "(token: string, project: string, instance: string, table: string) => [T] where T: Row",
            },
            "experimental/geo" => semantic_map! {
                     "getGrid" => "(region: T, ?minSize: int, ?maxSize: int, ?level: int, ?maxLevel: int, units: {distance: string}) => {level: int , set: [string]} where T: Row",
                     "getLevel" => "(token: string) => int",
                     "s2CellIDToken" => "(?token: string, ?point: {lat: float , lon: float}, level: int) => string",
                     "s2CellLatLon" => "(token: string) => {lat: float , lon: float}",
                     "stContains" => "(region: A, geometry: B, units: {distance: string}) => bool where A: Row, B: Row",
                     "stDistance" => "(region: A, geometry: B, units: {distance: string}) => float where A: Row, B: Row",
                     "stLength" => "(geometry: A, units: {distance: string}) => float where A: Row",
            },
            "experimental/json" => semantic_map! {
                "parse" => "(data: bytes) => A",
            },
            // parse(data: 12)
            // A parse(int data)
            "experimental/http" => semantic_map! {
                "get" => r#"(
                        url: string,
                        ?headers: A,
                        ?timeout: duration
                    ) => {statusCode: int , body: bytes , headers: B} where A: Row, B: Row "#,
            },
            "experimental/mqtt" => semantic_map! {
                "to" => r#"(
                        <-tables: [A],
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
                    ) => [B] where A: Row, B: Row "#,
            },
            "experimental/prometheus" => semantic_map! {
                "scrape" => "(url: string) => [A] where A: Row",
            },
            "experimental" => semantic_map! {
                 "addDuration" => "(d: duration, to: time) => time",
                 "chain" => "(first: [A], second: [B]) => [B] where A: Row, B: Row",
                 "subDuration" => "(d: duration, from: time) => time",
                 "group" => "(<-tables: [A], mode: string, columns: [string]) => [A] where A: Row",
                 "objectKeys" => "(o: A) => [string] where A: Row",
                 "set" => "(<-tables: [A], o: B) => [C] where A: Row, B: Row, C: Row",
                 // must specify exactly one of bucket, bucketID
                 // must specify exactly one of org, orgID
                 // if host is specified, token must be too.
                 // https://github.com/influxdata/flux/issues/1660
                 "to" => "(<-tables: [A], ?bucket: string, ?bucketID: string, ?org: string, ?orgID: string, ?host: string, ?token: string) => [A] where A: Row",
                 "join" => "(left: [A], right: [B], fn: (left: A, right: B) => C) => [C] where A: Row, B: Row, C: Row ",
                 "table" => "(rows: [A]) => [A] where A: Row ",
            },
            "generate" => semantic_map! {
                "from" => "(start: A, stop: A, count: int, fn: (n: int) => int) => [{ _start: time , _stop: time , _time: time , _value:int }] where A: Timeable",
            },
            "http" => semantic_map! {
                "post" => "(url: string, ?headers: A, ?data: bytes) => int where A: Row",
                "basicAuth" => "(u: string, p: string) => string",
                "pathEscape" => "(inputString: string) => string",
            },
            "influxdata/influxdb/secrets" => semantic_map! {
                "get" => "(key: string) => string",
            },
            "influxdata/influxdb/v1" => semantic_map! {
                // exactly one of json and file must be specified
                // https://github.com/influxdata/flux/issues/2250
                "json" => "(?json: string, ?file: string) => [A] where A: Row",
                "databases" => r#"
                    (
                        ?org: string,
                        ?orgID: string,
                        ?host: string,
                        ?token: string
                    ) => [{
                        organizationID: string ,
                        databaseName: string ,
                        retentionPolicy: string ,
                        retentionPeriod: int ,
                        default: bool ,
                        bucketID: string
                    }]
                "#,
            },
            "influxdata/influxdb" => semantic_map! {
                // This is a one-or-the-other parameters function
                // https://github.com/influxdata/flux/issues/1659
                "from" => r#"(
                    ?bucket: string,
                    ?bucketID: string,
                    ?org: string,
                    ?orgID: string,
                    ?host: string,
                    ?token: string
                ) => [{B with _measurement: string , _field: string , _time: time , _value: A}] "#,
                // exactly one of (bucket, bucketID) must be specified
                // exactly one of (org, orgID) must be specified
                // https://github.com/influxdata/flux/issues/1660
                "to" => r#"(
                        <-tables: [A],
                        ?bucket: string,
                        ?bucketID: string,
                        ?org: string,
                        ?orgID: string,
                        ?host: string,
                        ?token: string,
                        ?timeColumn: string,
                        ?measurementColumn: string,
                        ?tagColumns: [string],
                        ?fieldFn: (r: A) => B
                    ) => [A] where A: Row, B: Row "#,
                "buckets" => r#"
                    (
                        ?org: string,
                        ?orgID: string,
                        ?host: string,
                        ?token: string
                    ) => [{
                        name: string ,
                        id: string ,
                        organizationID: string ,
                        retentionPolicy: string ,
                        retentionPeriod: int
                    }]
                "#,
            },
            "internal/gen" => semantic_map! {
                "tables" => "(n: int, ?nulls: float, ?tags: [{name: string , cardinality: int}]) => [{A with _time: time , _value: float}]",
            },
            "internal/debug" => semantic_map! {
                "pass" => "(<-tables: [A]) => [A] where A: Row",
            },
            "internal/promql" => semantic_map! {
                "changes" => "(<-tables: [{A with _value: float}]) => [{B with _value: float}]",
                "promqlDayOfMonth" => "(timestamp: float) => float",
                "promqlDayOfWeek" => "(timestamp: float) => float",
                "promqlDaysInMonth" => "(timestamp: float) => float",
                "emptyTable" => "() => [{_start: time , _stop: time , _time: time , _value: float}]",
                "extrapolatedRate" => "(<-tables: [{A with _start: time , _stop: time , _time: time , _value: float}], ?isCounter: bool, ?isRate: bool) => [{B with _value: float}]",
                "holtWinters" => "(<-tables: [{A with _time: time , _value: float}], ?smoothingFactor: float, ?trendFactor: float) => [{B with _value: float}]",
                "promqlHour" => "(timestamp: float) => float",
                "instantRate" => "(<-tables: [{A with _time: time , _value: float}], ?isRate: bool) => [{B with _value: float}]",
                "labelReplace" => "(<-tables: [{A with _value: float}], source: string, destination: string, regex: string, replacement: string) => [{B with _value: float}]",
                "linearRegression" => "(<-tables: [{A with _time: time , _stop: time , _value: float}], ?predict: bool, ?fromNow: float) => [{B with _value: float}]",
                "promqlMinute" => "(timestamp: float) => float",
                "promqlMonth" => "(timestamp: float) => float",
                "promHistogramQuantile" => "(<-tables: [A], ?quantile: float, ?countColumn: string, ?upperBoundColumn: string, ?valueColumn: string) => [B] where A: Row, B: Row",
                "resets" => "(<-tables: [{A with _value: float}]) => [{B with _value: float}]",
                "timestamp" => "(<-tables: [{A with _value: float}]) => [{A with _value: float}]",
                "promqlYear" => "(timestamp: float) => float",
            },
            "internal/testutil" => semantic_map! {
                "fail" => "() => bool",
                "yield" => r#"(<-v: A) => A "#,
                "makeRecord" => "(o: A) => B where A: Row, B: Row",
            },
            "json" => semantic_map! {
                "encode" => "(v: A) => bytes",
            },
            "kafka" => semantic_map! {
                "to" => r#"(
                        <-tables: [A],
                        brokers: [string],
                        topic: string,
                        ?balancer: string,
                        ?name: string,
                        ?nameColumn: string,
                        ?timeColumn: string,
                        ?tagColumns: [string],
                        ?valueColumns: [string]
                    ) => [A] where A: Row "#,
            },
            "math" => semantic_map! {
                "pi" => "float",
                "e" => "float",
                "phi" => "float",
                "sqrt2" => "float",
                "sqrte" => "float",
                "sqrtpi" => "float",
                "sqrtphi" => "float",
                "log2e" => "float",
                "ln2" => "float",
                "ln10" => "float",
                "log10e" => "float",

                "maxfloat" => "float",
                "smallestNonzeroFloat" => "float",
                "maxint" => "int",
                "minint" => "int",
                "maxuint" => "uint",

                "abs" => "(x: float) => float",
                "acos" => "(x: float) => float",
                "acosh" => "(x: float) => float",
                "asin" => "(x: float) => float",
                "asinh" => "(x: float) => float",
                "atan" => "(x: float) => float",
                "atan2" => "(x: float, y: float) => float",
                "atanh" => "(x: float) => float",
                "cbrt" => "(x: float) => float",
                "ceil" => "(x: float) => float",
                "copysign" => "(x: float, y: float) => float",
                "cos" => "(x: float) => float",
                "cosh" => "(x: float) => float",
                "dim" => "(x: float, y: float) => float",
                "erf" => "(x: float) => float",
                "erfc" => "(x: float) => float",
                "erfcinv" => "(x: float) => float",
                "erfinv" => "(x: float) => float",
                "exp" => "(x: float) => float",
                "exp2" => "(x: float) => float",
                "expm1" => "(x: float) => float",
                "floor" => "(x: float) => float",
                "gamma" => "(x: float) => float",
                "hypot" => "(x: float, y: float) => float",
                "j0" => "(x: float) => float",
                "j1" => "(x: float) => float",
                "log" => "(x: float) => float",
                "log10" => "(x: float) => float",
                "log1p" => "(x: float) => float",
                "log2" => "(x: float) => float",
                "logb" => "(x: float) => float",
                "mMax" => "(x: float, y: float) => float",
                "mMin" => "(x: float, y: float) => float",
                "mod" => "(x: float, y: float) => float",
                "nextafter" => "(x: float, y: float) => float",
                "pow" => "(x: float, y: float) => float",
                "remainder" => "(x: float, y: float) => float",
                "round" => "(x: float) => float",
                "roundtoeven" => "(x: float) => float",
                "sin" => "(x: float) => float",
                "sinh" => "(x: float) => float",
                "sqrt" => "(x: float) => float",
                "tan" => "(x: float) => float",
                "tanh" => "(x: float) => float",
                "trunc" => "(x: float) => float",
                "y0" => "(x: float) => float",
                "y1" => "(x: float) => float",

                "float64bits" => "(f: float) => uint",
                "float64frombits" => "(b: uint) => float",
                "ilogb" => "(x: float) => int",
                "frexp" => "(f: float) => {frac: float , exp: int}",
                "lgamma" => "(x: float) => {lgamma: float , sign: int}",
                "modf" => r#"(f: float) => {int: float , frac: float} "#,
                "sincos" => "(x: float) => {sin: float , cos: float}",
                "isInf" => "(f: float, sign: int) => bool",
                "isNaN" => "(f: float) => bool",
                "signbit" => "(x: float) => bool",
                "NaN" => "() => float",
                "mInf" => "(sign: int) => float",
                "jn" => "(n: int, x: float) => float",
                "yn" => "(n: int, x: float) => float",
                "ldexp" => "(frac: float, exp: int) => float",
                "pow10" => "(n: int) => float",
            },
            "pagerduty" => semantic_map! {
                "dedupKey" => "(<-tables: [A]) => [{A with _pagerdutyDedupKey: string }]",
            },
            "regexp" => semantic_map! {
                "compile" => "(v: string) => regexp",
                "quoteMeta" => "(v: string) => string",
                "findString" => "(r: regexp, v: string) => string",
                "findStringIndex" => "(r: regexp, v: string) => [int]",
                "matchRegexpString" => "(r: regexp, v: string) => bool",
                "replaceAllString" => "(r: regexp, v: string, t: string) => string",
                "splitRegexp" => "(r: regexp, v: string, i: int) => [string]",
                "getString" => "(r: regexp) => string",
            },
            "runtime" => semantic_map! {
                "version" => "() => string",
            },
            "slack" => semantic_map! {
                "validateColorString" => "(color: string) => string",
            },
            "socket" => semantic_map! {
                "from" => "(url: string, ?decoder: string) => [A]",
            },
            "sql" => semantic_map! {
                "from" => "(driverName: string, dataSourceName: string, query: string) => [A]",
                "to" => "(<-tables: [A], driverName: string, dataSourceName: string, table: string, ?batchSize: int) => [A]",
            },
            "strings" => semantic_map! {
                "title" => "(v: string) => string",
                "toUpper" => "(v: string) => string",
                "toLower" => "(v: string) => string",
                "trim" => "(v: string, cutset: string) => string",
                "trimPrefix" => "(v: string, prefix: string) => string",
                "trimSpace" => "(v: string) => string",
                "trimSuffix" => "(v: string, suffix: string) => string",
                "trimRight" => "(v: string, cutset: string) => string",
                "trimLeft" => "(v: string, cutset: string) => string",
                "toTitle" => "(v: string) => string",
                "hasPrefix" => "(v: string, prefix: string) => bool",
                "hasSuffix" => "(v: string, suffix: string) => bool",
                "containsStr" => "(v: string, substr: string) => bool",
                "containsAny" => "(v: string, chars: string) => bool",
                "equalFold" => "(v: string, t: string) => bool",
                "compare" => "(v: string, t: string) => int",
                "countStr" => "(v: string, substr: string) => int",
                "index" => "(v: string, substr: string) => int",
                "indexAny" => "(v: string, chars: string) => int",
                "lastIndex" => "(v: string, substr: string) => int",
                "lastIndexAny" => "(v: string, chars: string) => int",
                "isDigit" => "(v: string) => bool",
                "isLetter" => "(v: string) => bool",
                "isLower" => "(v: string) => bool",
                "isUpper" => "(v: string) => bool",
                "repeat" => "(v: string, i: int) => string",
                "replace" => "(v: string, t: string, u: string, i: int) => string",
                "replaceAll" => "(v: string, t: string, u: string) => string",
                "split" => "(v: string, t: string) => [string]",
                "splitAfter" => "(v: string, t: string) => [string]",
                "splitN" => "(v: string, t: string, n: int) => [string]",
                "splitAfterN" => "(v: string, t: string, i: int) => [string]",
                "joinStr" => "(arr: [string], v: string) => string",
                "strlen" => "(v: string) => int",
                "substring" => "(v: string, start: int, end: int) => string",
            },
            "system" => semantic_map! {
                "time" => "() => time",
            },
            "testing" => semantic_map! {
                "assertEquals" => "(name: string, <-got: [A], want: [A]) => [A]",
                "assertEmpty" => "(<-tables: [A]) => [A]",
                "diff" => "(<-got: [A], want: [A], ?verbose: bool, ?epsilon: float) => [{A with _diff: string}]",
            },
            "universe" => semantic_map! {
                "bool" => "(v: A) => bool",
                "bytes" => "(v: A) => bytes",
                "chandeMomentumOscillator" => r#"(
                        <-tables: [A],
                        n: int,
                        ?columns: [string]
                    ) => [B] where A: Row, B: Row "#,
                "columns" => r#"(
                        <-tables: [A],
                        ?column: string
                    ) => [B] where A: Row, B: Row "#,
                "contains" => r#"(
                        value: A,
                        set: [A]
                    ) => bool where A: Nullable "#,
                "count" => r#"(
                        <-tables: [A],
                        ?column: string
                    ) => [B] where A: Row, B: Row "#,
                "covariance" => r#"(
                        <-tables: [A],
                        ?pearsonr: bool,
                        ?valueDst: string,
                        columns: [string]
                    ) => [B] where A: Row, B: Row "#,
                "cumulativeSum" => r#"(
                        <-tables: [A],
                        ?columns: [string]
                    ) => [B] where A: Row, B: Row "#,
                "derivative" => r#"(
                        <-tables: [A],
                        ?unit: duration,
                        ?nonNegative: bool,
                        ?columns: [string],
                        ?timeColumn: string
                    ) => [B] where A: Row, B: Row "#,
                "difference" => r#"
                   (
                        <-tables: [T],
                        ?nonNegative: bool,
                        ?columns: [string],
                        ?keepFirst: bool
                    ) => [R] where T: Row, R: Row
                "#,
                "distinct" => r#"(
                        <-tables: [A],
                        ?column: string
                    ) => [B] where A: Row, B: Row "#,
                "drop" => r#"(
                        <-tables: [A],
                        ?fn: (column: string) => bool,
                        ?columns: [string]
                    ) => [B] where A: Row, B: Row "#,
                "duplicate" => r#"(
                        <-tables: [A],
                        column: string,
                        as: string
                    ) => [B] where A: Row, B: Row "#,
                "duration" => "(v: A) => duration",
                "elapsed" => r#"(
                        <-tables: [A],
                        ?unit: duration,
                        ?timeColumn: string,
                        ?columnName: string
                    ) => [B] where A: Row, B: Row "#,
                "exponentialMovingAverage" => r#"(
                        <-tables: [{ B with _value: A}],
                        n: int
                    ) => [{ B with _value: A }] where A: Numeric "#,
                "false" => "bool",
                "fill" => r#"(
                        <-tables: [A],
                        ?column: string,
                        ?value: B,
                        ?usePrevious: bool
                    ) => [C] where A: Row, C: Row "#,
                "filter" => r#"(
                        <-tables: [A],
                        fn: (r: A) => bool,
                        ?onEmpty: string
                    ) => [A] where A: Row "#,
                "first" => r#"(
                        <-tables: [A],
                        ?column: string
                    ) => [A] where A: Row "#,
                "float" => "(v: A) => float",
                "getColumn" => r#"(
                        <-table: [A],
                        column: string
                    ) => [B] where A: Row "#,
                "getRecord" => r#"(
                        <-table: [A],
                        idx: int
                    ) => A where A: Row "#,
                "findColumn" => r#"(
                        <-tables: [A],
                        fn: (key: B) => bool,
                        column: string
                    ) => [C] where A: Row, B: Row "#,
                "findRecord" => r#"(
                        <-tables: [A],
                        fn: (key: B) => bool,
                        idx: int
                    ) => A where A: Row, B: Row "#,
                "group" => r#"(
                        <-tables: [A],
                        ?mode: string,
                        ?columns: [string]
                    ) => [A] where A: Row "#,
                "histogram" => r#"(
                        <-tables: [A],
                        ?column: string,
                        ?upperBoundColumn: string,
                        ?countColumn: string,
                        bins: [float],
                        ?normalize: bool
                    ) => [B] where A: Row, B: Row "#,
                "histogramQuantile" => r#"(
                        <-tables: [A],
                        ?quantile: float,
                        ?countColumn: string,
                        ?upperBoundColumn: string,
                        ?valueColumn: string,
                        ?minValue: float
                    ) => [B] where A: Row, B: Row "#,
                "holtWinters" => r#"(
                        <-tables: [A],
                        n: int,
                        interval: duration,
                        ?withFit: bool,
                        ?column: string,
                        ?timeColumn: string,
                        ?seasonality: int
                    ) => [B] where A: Row, B: Row "#,
                "hourSelection" => r#"(
                        <-tables: [A],
                        start: int,
                        stop: int,
                        ?timeColumn: string
                    ) => [A] where A: Row "#,
                "inf" => "duration",
                "int" => "(v: A) => int",
                "integral" => r#"(
                        <-tables: [A],
                        ?unit: duration,
                        ?timeColumn: string,
                        ?column: string
                    ) => [B] where A: Row, B: Row "#,
                "join" => r#"(
                        <-tables: A,
                        ?method: string,
                        ?on: [string]
                    ) => [B] where A: Row, B: Row "#,
                // This function would almost have input/output types that match, but:
                // input column may start as int, uint or float, and always ends up as float.
                // https://github.com/influxdata/flux/issues/2252
                "kaufmansAMA" => r#"(
                        <-tables: [A],
                        n: int,
                        ?column: string
                    ) => [B] where A: Row, B: Row "#,
                // either column list or predicate must be provided
                // https://github.com/influxdata/flux/issues/2248
                "keep" => r#"(
                        <-tables: [A],
                        ?columns: [string],
                        ?fn: (column: string) => bool
                    ) => [B] where A: Row, B: Row "#,
                "keyValues" => r#"(
                        <-tables: [A],
                        ?keyColumns: [string]
                    ) => [{C with _key: string , _value: B}] where A: Row, C: Row "#,
                "keys" => r#"(
                        <-tables: [A],
                        ?column: string
                    ) => [B] where A: Row, B: Row "#,
                "last" => "(<-tables: [A], ?column: string) => [A] where A: Row",
                "length" => "(arr: [A]) => int",
                "limit"  => "(<-tables: [A], n: int, ?offset: int) => [A]",
                "linearBins" => r#"(
                        start: float,
                        width: float,
                        count: int,
                        ?infinity: bool
                    ) => [float] "#,
                "logarithmicBins" => r#"(
                        start: float,
                        factor: float,
                        count: int,
                        ?infinity: bool
                    ) => [float] "#,
                // Note: mergeKey parameter could be removed from map once the transpiler is updated:
                // https://github.com/influxdata/flux/issues/816
                "map" => "(<-tables: [A], fn: (r: A) => B, ?mergeKey: bool) => [B]",
                "max" => "(<-tables: [A], ?column: string) => [A] where A: Row",
                "mean" => r#"(
                        <-tables: [A],
                        ?column: string
                    ) => [B] where A: Row, B: Row "#,
                "min" => "(<-tables: [A], ?column: string) => [A] where A: Row",
                "mode" => r#"(
                        <-tables: [A],
                        ?column: string
                    ) => [{C with _value: B}] where A: Row, C: Row "#,
                "movingAverage" => "(<-tables: [{B with _value: A}], n: int) => [{B with _value: float}] where A: Numeric",
                "pivot" => r#"(
                        <-tables: [A],
                        rowKey: [string],
                        columnKey: [string],
                        valueColumn: string
                    ) => [B] where A: Row, B: Row "#,
                "quantile" => r#"(
                        <-tables: [A],
                        ?column: string,
                        q: float,
                        ?compression: float,
                        ?method: string
                    ) => [A] where A: Row "#,
                // start and stop should be able to constrained to time or duration with a kind constraint:
                //   https://github.com/influxdata/flux/issues/2243
                // Also, we should remove the column arguments so we can reuse A in the return type:
                //   https://github.com/influxdata/flux/issues/2253
                "range" => r#"(
                        <-tables: [A],
                        start: B,
                        ?stop: C,
                        ?timeColumn: string,
                        ?startColumn: string,
                        ?stopColumn: string
                    ) => [D] where A: Row, D: Row "#,
                // This function could be updated to get better type inference:
                //   https://github.com/influxdata/flux/issues/2254
                "reduce" => r#"(
                        <-tables: [A],
                        fn: (r: A, accumulator: B) => B,
                        identity: B
                    ) => [C] where A: Row, B: Row, C: Row "#,
                "relativeStrengthIndex" => r#"(
                        <-tables: [A],
                        n: int,
                        ?columns: [string]
                    ) => [B] where A: Row, B: Row "#,
                // Either fn or columns should be specified
                // https://github.com/influxdata/flux/issues/2251
                "rename" => r#"(
                        <-tables: [A],
                        ?fn: (column: string) => string,
                        ?columns: B
                    ) => [C] where A: Row, B: Row, C: Row "#,
                "sample" => r#"(
                        <-tables: [A],
                        n: int,
                        ?pos: int,
                        ?column: string
                    ) => [A] where A: Row "#,
                "set" => r#"(
                        <-tables: [A],
                        key: string,
                        value: string
                    ) => [A] where A: Row "#,
                // This is an aggregate function, and may clobber value columns
                "skew" => r#"(
                        <-tables: [A],
                        ?column: string
                    ) => [B] where A: Row, B: Row "#,
                "sleep" => r#"
                    (
                        <-v: A,
                        duration: duration
                    ) => A
                "#,
                "sort" => r#"(
                        <-tables: [A],
                        ?columns: [string],
                        ?desc: bool
                    ) => [A] where A: Row "#,
                "spread" => r#"(
                        <-tables: [A],
                        ?column: string
                    ) => [B] where A: Row, B: Row "#,
                "stateTracking" => r#"(
                        <-tables: [A],
                        fn: (r: A) => bool,
                        ?countColumn: string,
                        ?durationColumn: string,
                        ?durationUnit: duration,
                        ?timeColumn: string
                    ) => [B] where A: Row, B: Row "#,
                "stddev" => r#"(
                        <-tables: [A],
                        ?column: string,
                        ?mode: string
                    ) => [B] where A: Row, B: Row "#,
                "string" => "(v: A) => string",
                "sum" => r#"(
                        <-tables: [A],
                        ?column: string
                    ) => [B] where A: Row, B: Row "#,
                "tableFind" => r#"(
                        <-tables: [A],
                        fn: (key: B) => bool
                    ) => [A] where A: Row, B: Row "#,
                "tail" => r#"(
                        <-tables: [A],
                        n: int,
                        ?offset: int
                    ) => [A] "#,
                "time" => "(v: A) => time",
                "timeShift" => r#"
                    (
                        <-tables: [A],
                        duration: duration,
                        ?columns: [string]
                    ) => [A]
                "#,
                "tripleExponentialDerivative" => r#"(
                        <-tables: [{B with _value: A}],
                        n: int
                    ) => [{B with _value: float}] where A: Numeric, B: Row "#,
                "true" => "bool",
                "uint" => "(v: A) => uint",
                "union" => r#"(
                        tables: [[A]]
                    ) => [A] where A: Row "#,
                "unique" => r#"(
                        <-tables: [A],
                        ?column: string
                    ) => [A] where A: Row "#,
                // This would produce an output the same as the input,
                // except that startColumn and stopColumn will be added if they don't
                // already exist.
                // https://github.com/influxdata/flux/issues/2255
                "window" => r#"(
                        <-tables: [A],
                        ?every: duration,
                        ?period: duration,
                        ?offset: duration,
                        ?timeColumn: string,
                        ?startColumn: string,
                        ?stopColumn: string,
                        ?createEmpty: bool
                    ) => [B] where A: Row, B: Row "#,
                "yield" => r#"(
                        <-tables: [A],
                        ?name: string
                    ) => [A] where A: Row "#,
            },
            "contrib/jsternberg/math" => semantic_map! {
                "minIndex" => r#"(
                        values: [A]
                    ) => int where A: Numeric "#,
                "maxIndex" => r#"(
                        values: [A]
                    ) => int where A: Numeric "#,
                "sum" => r#"(
                        values: [A]
                    ) => A where A: Numeric "#,
            },
            "contrib/jsternberg/aggregate" => semantic_map! {
                "table" => r#"(
                        <-tables: [A],
                        columns: C
                    ) => [B] where A: Row, B: Row, C: Row "#,
                "null" => r#"A"#,
                "none" => r#"A"#,
            },
            "contrib/jsternberg/influxdb" => semantic_map! {
                "_mask" => r#"(
                        <-tables: [A],
                        columns: [string]
                    ) => [B] where A: Row, B: Row "#,
            },
            "contrib/jsternberg/rows" => semantic_map! {
                "map" => r#"(
                        <-tables: [A],
                        fn: (r: A) => B
                    ) => [B] where A: Row, B: Row "#,
            },
        },
    }
}

#[cfg(test)]
mod test {
    use crate::ast::get_err_type_expression;
    use crate::parser;
    use crate::semantic::builtins::builtins;
    use crate::semantic::convert::convert_polytype;
    use crate::semantic::fresh::Fresher;
    #[test]
    fn parse_builtin_types() {
        for (path, values) in builtins().iter() {
            for (name, expr) in values {
                let mut p = parser::Parser::new(expr);

                let typ_expr = p.parse_type_expression();
                let err = get_err_type_expression(typ_expr.clone());

                if err != "" {
                    let msg = format!(
                        "TypeExpression parsing failed for {}.{}. {:?}",
                        path, name, err
                    );
                    panic!(msg)
                }
                let expr = convert_polytype(typ_expr, &mut Fresher::default());

                match expr {
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
