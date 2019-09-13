// THIS PACKAGE IS NOT MEANT FOR EXTERNAL USE.
package promql

import "math" 
import "universe"

// changes() implements functionality equivalent to PromQL's changes() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#changes
builtin changes

// promqlDayOfMonth() implements functionality equivalent to PromQL's day_of_month() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#day_of_month
builtin promqlDayOfMonth

// promqlDayOfWeek() implements functionality equivalent to PromQL's day_of_week() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#day_of_week
builtin promqlDayOfWeek

// promqlDaysInMonth() implements functionality equivalent to PromQL's days_in_month() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#days_in_month
builtin promqlDaysInMonth

// emptyTable() returns an empty table, which is used as a helper function to implement
// PromQL's time() and vector() functions:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#time
// https://prometheus.io/docs/prometheus/latest/querying/functions/#vector
builtin emptyTable

// extrapolatedRate() is a helper function that calculates extrapolated rates over
// counters and is used to implement PromQL's rate(), delta(), and increase() functions.
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#rate
// https://prometheus.io/docs/prometheus/latest/querying/functions/#increase
// https://prometheus.io/docs/prometheus/latest/querying/functions/#delta
builtin extrapolatedRate

// holtWinters() implements functionality equivalent to PromQL's holt_winters()
// function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#holt_winters
builtin holtWinters

// promqlHour() implements functionality equivalent to PromQL's hour() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#hour
builtin promqlHour

// instantRate() is a helper function that calculates instant rates over
// counters and is used to implement PromQL's irate() and idelta() functions.
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#irate
// https://prometheus.io/docs/prometheus/latest/querying/functions/#idelta
builtin instantRate

// labelReplace implements functionality equivalent to PromQL's label_replace() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#label_replace
builtin labelReplace

// linearRegression implements linear regression functionality required to implement
// PromQL's deriv() and predict_linear() functions:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#deriv
// https://prometheus.io/docs/prometheus/latest/querying/functions/#predict_linear
builtin linearRegression

// promqlMinute() implements functionality equivalent to PromQL's minute() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#minute
builtin promqlMinute

// promqlMonth() implements functionality equivalent to PromQL's month() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#month
builtin promqlMonth

// promHistogramQuantile() implements functionality equivalent to PromQL's
// histogram_quantile() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#histogram_quantile
builtin promHistogramQuantile

// resets() implements functionality equivalent to PromQL's resets() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#resets
builtin resets

// timestamp() implements functionality equivalent to PromQL's timestamp() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#timestamp
builtin timestamp

// promqlYear() implements functionality equivalent to PromQL's year() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#year
builtin promqlYear

// quantile() accounts checks for quantile values that are out of range, above 1.0 or 
// below 0.0, by either returning positive infinity or negative infinity in the `_value` 
// column respectively. q must be a float 

quantile = (q, tables=<-, method="exact_mean") => 
    // value is in normal range. We can use the normal quantile function
    if q <= 1 and q >= 0 then 
    (tables
        |> universe.quantile(q: q, method: method))
    else if q < 0 then
    (tables
        |> reduce(identity: {_value: math.mInf(sign: -1)}, fn: (r, accumulator) => accumulator))
    else 
    (tables
        |> reduce(identity: {_value: math.mInf(sign: 1)}, fn: (r, accumulator) => accumulator))