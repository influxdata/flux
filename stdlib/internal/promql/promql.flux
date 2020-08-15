// THIS PACKAGE IS NOT MEANT FOR EXTERNAL USE.
package promql

import "math" 
import "universe"
import "experimental"

// changes() implements functionality equivalent to PromQL's changes() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#changes
builtin changes : (<-tables: [{A with _value: float}]) => [{B with _value: float}]

// promqlDayOfMonth() implements functionality equivalent to PromQL's day_of_month() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#day_of_month
builtin promqlDayOfMonth : (timestamp: float) => float

// promqlDayOfWeek() implements functionality equivalent to PromQL's day_of_week() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#day_of_week
builtin promqlDayOfWeek : (timestamp: float) => float

// promqlDaysInMonth() implements functionality equivalent to PromQL's days_in_month() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#days_in_month
builtin promqlDaysInMonth : (timestamp: float) => float

// emptyTable() returns an empty table, which is used as a helper function to implement
// PromQL's time() and vector() functions:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#time
// https://prometheus.io/docs/prometheus/latest/querying/functions/#vector
builtin emptyTable : () => [{_start: time , _stop: time , _time: time , _value: float}]

// extrapolatedRate() is a helper function that calculates extrapolated rates over
// counters and is used to implement PromQL's rate(), delta(), and increase() functions.
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#rate
// https://prometheus.io/docs/prometheus/latest/querying/functions/#increase
// https://prometheus.io/docs/prometheus/latest/querying/functions/#delta
builtin extrapolatedRate : (<-tables: [{A with _start: time , _stop: time , _time: time , _value: float}], ?isCounter: bool, ?isRate: bool) => [{B with _value: float}]

// holtWinters() implements functionality equivalent to PromQL's holt_winters()
// function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#holt_winters
builtin holtWinters : (<-tables: [{A with _time: time , _value: float}], ?smoothingFactor: float, ?trendFactor: float) => [{B with _value: float}]

// promqlHour() implements functionality equivalent to PromQL's hour() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#hour
builtin promqlHour : (timestamp: float) => float

// instantRate() is a helper function that calculates instant rates over
// counters and is used to implement PromQL's irate() and idelta() functions.
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#irate
// https://prometheus.io/docs/prometheus/latest/querying/functions/#idelta
builtin instantRate : (<-tables: [{A with _time: time , _value: float}], ?isRate: bool) => [{B with _value: float}]

// labelReplace implements functionality equivalent to PromQL's label_replace() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#label_replace
builtin labelReplace : (<-tables: [{A with _value: float}], source: string, destination: string, regex: string, replacement: string) => [{B with _value: float}]

// linearRegression implements linear regression functionality required to implement
// PromQL's deriv() and predict_linear() functions:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#deriv
// https://prometheus.io/docs/prometheus/latest/querying/functions/#predict_linear
builtin linearRegression : (<-tables: [{A with _time: time , _stop: time , _value: float}], ?predict: bool, ?fromNow: float) => [{B with _value: float}]

// promqlMinute() implements functionality equivalent to PromQL's minute() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#minute
builtin promqlMinute : (timestamp: float) => float

// promqlMonth() implements functionality equivalent to PromQL's month() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#month
builtin promqlMonth : (timestamp: float) => float

// promHistogramQuantile() implements functionality equivalent to PromQL's
// histogram_quantile() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#histogram_quantile
builtin promHistogramQuantile : (<-tables: [A], ?quantile: float, ?countColumn: string, ?upperBoundColumn: string, ?valueColumn: string) => [B] where A: Record, B: Record

// resets() implements functionality equivalent to PromQL's resets() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#resets
builtin resets : (<-tables: [{A with _value: float}]) => [{B with _value: float}]

// timestamp() implements functionality equivalent to PromQL's timestamp() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#timestamp
builtin timestamp : (<-tables: [{A with _value: float}]) => [{A with _value: float}]

// promqlYear() implements functionality equivalent to PromQL's year() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#year
builtin promqlYear : (timestamp: float) => float

// quantile() accounts checks for quantile values that are out of range, above 1.0 or 
// below 0.0, by either returning positive infinity or negative infinity in the `_value` 
// column respectively. q must be a float 

quantile = (q, tables=<-, method="exact_mean") => 
    // value is in normal range. We can use the normal quantile function
    if q <= 1.0 and q >= 0.0 then 
    (tables
        |> universe.quantile(q: q, method: method))
    else if q < 0.0 then
    (tables
        |> reduce(identity: {_value: math.mInf(sign: -1)}, fn: (r, accumulator) => accumulator))
    else 
    (tables
        |> reduce(identity: {_value: math.mInf(sign: 1)}, fn: (r, accumulator) => accumulator))

join = experimental.join
