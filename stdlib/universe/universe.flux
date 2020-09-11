package universe

import "system"
import "date"
import "math"
import "strings"
import "regexp"

// now is a function option whose default behaviour is to return the current system time
option now = system.time

// Booleans
builtin true : bool
builtin false : bool

// Transformation functions
builtin chandeMomentumOscillator : (<-tables: [A],  n: int, ?columns: [string]) => [B] where A: Record, B: Record
builtin columns : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin count : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin covariance : (<-tables: [A], ?pearsonr: bool, ?valueDst: string, columns: [string]) => [B] where A: Record, B: Record
builtin cumulativeSum : (<-tables: [A], ?columns: [string]) => [B] where A: Record, B: Record
builtin derivative : (<-tables: [A], ?unit: duration, ?nonNegative: bool, ?columns: [string], ?timeColumn: string) => [B] where A: Record, B: Record
builtin die : (msg: string) => A
builtin difference : (<-tables: [T], ?nonNegative: bool, ?columns: [string], ?keepFirst: bool) => [R] where T: Record, R: Record
builtin distinct : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin drop : (<-tables: [A], ?fn: (column: string) => bool, ?columns: [string]) => [B] where A: Record, B: Record
builtin duplicate : (<-tables: [A], column: string, as: string) => [B] where A: Record, B: Record
builtin elapsed : (<-tables: [A], ?unit: duration, ?timeColumn: string, ?columnName: string) => [B] where A: Record, B: Record
builtin exponentialMovingAverage : (<-tables: [{ B with _value: A}], n: int) => [{ B with _value: A }] where A: Numeric
builtin fill : (<-tables: [A], ?column: string, ?value: B, ?usePrevious: bool) => [C] where A: Record, C: Record
builtin filter : (<-tables: [A], fn: (r: A) => bool, ?onEmpty: string) => [A] where A: Record
builtin first : ( <-tables: [A], ?column: string) => [A] where A: Record
builtin group : (<-tables: [A], ?mode: string, ?columns: [string]) => [A] where A: Record
builtin histogram : (<-tables: [A], ?column: string, ?upperBoundColumn: string, ?countColumn: string, bins: [float], ?normalize: bool) => [B] where A: Record, B: Record
builtin histogramQuantile : (<-tables: [A], ?quantile: float, ?countColumn: string, ?upperBoundColumn: string, ?valueColumn: string, ?minValue: float) => [B] where A: Record, B: Record
builtin holtWinters : (<-tables: [A], n: int, interval: duration, ?withFit: bool, ?column: string, ?timeColumn: string, ?seasonality: int) => [B] where A: Record, B: Record
builtin hourSelection : (<-tables: [A], start: int, stop: int, ?timeColumn: string) => [A] where A: Record
builtin integral : (<-tables: [A], ?unit: duration, ?timeColumn: string, ?column: string, ?interpolate: string) => [B] where A: Record, B: Record
builtin join : (<-tables: A, ?method: string, ?on: [string]) => [B] where A: Record, B: Record
builtin kaufmansAMA : (<-tables: [A], n: int, ?column: string) => [B] where A: Record, B: Record
builtin keep : (<-tables: [A], ?columns: [string], ?fn: (column: string) => bool) => [B] where A: Record, B: Record
builtin keyValues : (<-tables: [A], ?keyColumns: [string]) => [{C with _key: string , _value: B}] where A: Record, C: Record
builtin keys : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin last : (<-tables: [A], ?column: string) => [A] where A: Record
builtin limit : (<-tables: [A], n: int, ?offset: int) => [A]
builtin map : (<-tables: [A], fn: (r: A) => B, ?mergeKey: bool) => [B]
builtin max : (<-tables: [A], ?column: string) => [A] where A: Record
builtin mean : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin min : (<-tables: [A], ?column: string) => [A] where A: Record
builtin mode : (<-tables: [A], ?column: string) => [{C with _value: B}] where A: Record, C: Record
builtin movingAverage : (<-tables: [{B with _value: A}], n: int) => [{B with _value: float}] where A: Numeric
builtin quantile : (<-tables: [A], ?column: string, q: float, ?compression: float, ?method: string) => [A] where A: Record
builtin pivot : (<-tables: [A], rowKey: [string], columnKey: [string], valueColumn: string) => [B] where A: Record, B: Record
builtin range : (<-tables: [A], start: B, ?stop: C, ?timeColumn: string, ?startColumn: string, ?stopColumn: string) => [D] where A: Record, D: Record
builtin reduce : (<-tables: [A], fn: (r: A, accumulator: B) => B, identity: B) => [C] where A: Record, B: Record, C: Record
builtin relativeStrengthIndex : (<-tables: [A], n: int, ?columns: [string]) => [B] where A: Record, B: Record
builtin rename : (<-tables: [A], ?fn: (column: string) => string, ?columns: B) => [C] where A: Record, B: Record, C: Record
builtin sample : (<-tables: [A], n: int, ?pos: int, ?column: string) => [A] where A: Record
builtin set : (<-tables: [A], key: string, value: string) => [A] where A: Record
builtin tail : (<-tables: [A], n: int, ?offset: int) => [A]
builtin timeShift : (<-tables: [A], duration: duration, ?columns: [string]) => [A]
builtin skew : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin spread : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin sort : (<-tables: [A], ?columns: [string], ?desc: bool) => [A] where A: Record
builtin stateTracking : (<-tables: [A], fn: (r: A) => bool, ?countColumn: string, ?durationColumn: string, ?durationUnit: duration, ?timeColumn: string) => [B] where A: Record, B: Record
builtin stddev : (<-tables: [A], ?column: string, ?mode: string) => [B] where A: Record, B: Record
builtin sum : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin tripleExponentialDerivative : (<-tables: [{B with _value: A}], n: int) => [{B with _value: float}] where A: Numeric, B: Record
builtin union : (tables: [[A]]) => [A] where A: Record
builtin unique : (<-tables: [A], ?column: string) => [A] where A: Record
builtin window : (<-tables: [A], ?every: duration, ?period: duration, ?offset: duration, ?timeColumn: string, ?startColumn: string, ?stopColumn: string, ?createEmpty: bool) => [B] where A: Record, B: Record
builtin yield : (<-tables: [A], ?name: string) => [A] where A: Record

// stream/table index functions
builtin tableFind : (<-tables: [A], fn: (key: B) => bool) => [A] where A: Record, B: Record
builtin getColumn : (<-table: [A], column: string) => [B] where A: Record
builtin getRecord : (<-table: [A], idx: int) => A where A: Record
builtin findColumn : (<-tables: [A], fn: (key: B) => bool, column: string) => [C] where A: Record, B: Record
builtin findRecord : (<-tables: [A], fn: (key: B) => bool, idx: int) => A where A: Record, B: Record

// type conversion functions
builtin bool : (v: A) => bool
builtin bytes : (v: A) => bool
builtin duration : (v: A) => duration
builtin float : (v: A) => float
builtin int : (v: A) => int
builtin string : (v: A) => string
builtin time : (v: A) => time
builtin uint : (v: A) => uint

// contains function
builtin contains : (value: A, set: [A]) => bool where A: Nullable

// other builtins
builtin inf : duration
builtin length : (arr: [A]) => int
builtin linearBins : (start: float, width: float, count: int, ?infinity: bool) => [float]
builtin logarithmicBins : (start: float, factor: float, count: int, ?infinity: bool) => [float]

// sleep is the identity function with the side effect of delaying execution by a specified duration
builtin sleep : (<-v: A, duration: duration) => A
// die returns a fatal error from within a flux script
builtin die : (msg: string) => A

// Time weighted average where values at the beginning and end of the range are linearly interpolated.
timeWeightedAvg = (tables=<-, unit) => tables
    |> integral(unit: unit, interpolate: "linear")
    |> map(fn: (r) => ({ r with _value: (r._value * float(v: uint(v: unit))) / float(v: int(v: r._stop) - int(v: r._start)) }))

// covariance function with automatic join
cov = (x,y,on,pearsonr=false) =>
    join(
        tables:{x:x, y:y},
        on:on,
    )
    |> covariance(pearsonr:pearsonr, columns:["_value_x","_value_y"])

pearsonr = (x,y,on) => cov(x:x, y:y, on:on, pearsonr:true)

// AggregateWindow applies an aggregate function to fixed windows of time.
// The procedure is to window the data, perform an aggregate operation,
// and then undo the windowing to produce an output table for every input table.
aggregateWindow = (every, fn, column="_value", timeSrc="_stop",timeDst="_time", createEmpty=true, tables=<-) =>
    tables
        |> window(every:every, createEmpty: createEmpty)
        |> fn(column:column)
        |> duplicate(column:timeSrc,as:timeDst)
        |> window(every:inf, timeColumn:timeDst)

// Increase returns the total non-negative difference between values in a table.
// A main usage case is tracking changes in counter values which may wrap over time when they hit
// a threshold or are reset. In the case of a wrap/reset,
// we can assume that the absolute delta between two points will be at least their non-negative difference.
increase = (tables=<-, columns=["_value"]) =>
    tables
        |> difference(nonNegative: true, columns:columns)
        |> cumulativeSum(columns: columns)

// median returns the 50th percentile.
median = (method="estimate_tdigest", compression=0.0, column="_value", tables=<-) =>
    tables
        |> quantile(q:0.5, method: method, compression: compression, column: column)

// stateCount computes the number of consecutive records in a given state.
// The state is defined via the function fn. For each consecutive point for
// which the expression evaluates as true, the state count will be incremented
// When a point evaluates as false, the state count is reset.
//
// The state count will be added as an additional column to each record. If the
// expression evaluates as false, the value will be -1. If the expression
// generates an error during evaluation, the point is discarded, and does not
// affect the state count.
stateCount = (fn, column="stateCount", tables=<-) =>
    tables
        |> stateTracking(countColumn:column, fn:fn)

// stateDuration computes the duration of a given state.
// The state is defined via the function fn. For each consecutive point for
// which the expression evaluates as true, the state duration will be
// incremented by the duration between points. When a point evaluates as false,
// the state duration is reset.
//
// The state duration will be added as an additional column to each record. If the
// expression evaluates as false, the value will be -1. If the expression
// generates an error during evaluation, the point is discarded, and does not
// affect the state duration.
//
// Note that as the first point in the given state has no previous point, its
// state duration will be 0.
//
// The duration is represented as an integer in the units specified.
stateDuration = (fn, column="stateDuration", timeColumn="_time", unit=1s, tables=<-) =>
    tables
        |> stateTracking(durationColumn:column, timeColumn:timeColumn, fn:fn, durationUnit:unit)

// _sortLimit is a helper function, which sorts and limits a table.
_sortLimit = (n, desc, columns=["_value"], tables=<-) =>
    tables
        |> sort(columns:columns, desc:desc)
        |> limit(n:n)

// top sorts a table by columns and keeps only the top n records.
top = (n, columns=["_value"], tables=<-) =>
    tables
        |> _sortLimit(n:n, columns:columns, desc:true)

// top sorts a table by columns and keeps only the bottom n records.
bottom = (n, columns=["_value"], tables=<-) =>
    tables
        |> _sortLimit(n:n, columns:columns, desc:false)

// _highestOrLowest is a helper function, which reduces all groups into a single group by specific tags and a reducer function,
// then it selects the highest or lowest records based on the column and the _sortLimit function.
// The default reducer assumes no reducing needs to be performed.
_highestOrLowest = (n, _sortLimit, reducer, column="_value", groupColumns=[], tables=<-) =>
    tables
        |> group(columns:groupColumns)
        |> reducer()
        |> group(columns:[])
        |> _sortLimit(n:n, columns:[column])

// highestMax returns the top N records from all groups using the maximum of each group.
highestMax = (n, column="_value", groupColumns=[], tables=<-) =>
    tables
        |> _highestOrLowest(
                n:n,
                column:column,
                groupColumns:groupColumns,
                // TODO(nathanielc): Once max/min support selecting based on multiple columns change this to pass all columns.
                reducer: (tables=<-) => tables |> max(column:column),
                _sortLimit: top,
            )

// highestAverage returns the top N records from all groups using the average of each group.
highestAverage = (n, column="_value", groupColumns=[], tables=<-) =>
    tables
        |> _highestOrLowest(
                n:n,
                column:column,
                groupColumns:groupColumns,
                reducer: (tables=<-) => tables |> mean(column:column),
                _sortLimit: top,
            )

// highestCurrent returns the top N records from all groups using the last value of each group.
highestCurrent = (n, column="_value", groupColumns=[], tables=<-) =>
    tables
        |> _highestOrLowest(
                n:n,
                column:column,
                groupColumns:groupColumns,
                reducer: (tables=<-) => tables |> last(column:column),
                _sortLimit: top,
            )

// lowestMin returns the bottom N records from all groups using the minimum of each group.
lowestMin = (n, column="_value", groupColumns=[], tables=<-) =>
    tables
        |> _highestOrLowest(
                n:n,
                column:column,
                groupColumns:groupColumns,
                // TODO(nathanielc): Once max/min support selecting based on multiple columns change this to pass all columns.
                reducer: (tables=<-) => tables |> min(column:column),
                _sortLimit: bottom,
            )

// lowestAverage returns the bottom N records from all groups using the average of each group.
lowestAverage = (n, column="_value", groupColumns=[], tables=<-) =>
    tables
        |> _highestOrLowest(
                n:n,
                column:column,
                groupColumns:groupColumns,
                reducer: (tables=<-) => tables |> mean(column:column),
                _sortLimit: bottom,
            )

// lowestCurrent returns the bottom N records from all groups using the last value of each group.
lowestCurrent = (n, column="_value", groupColumns=[], tables=<-) =>
    tables
        |> _highestOrLowest(
                n:n,
                column:column,
                groupColumns:groupColumns,
                reducer: (tables=<-) => tables |> last(column:column),
                _sortLimit: bottom,
            )

// timedMovingAverage constructs a simple moving average over windows of 'period' duration
// eg: A 5 year moving average would be called as such:
//    movingAverage(1y, 5y)
timedMovingAverage = (every, period, column="_value", tables=<-) =>
    tables
        |> window(every: every, period: period)
        |> mean(column:column)
        |> duplicate(column: "_stop", as: "_time")
        |> window(every: inf)

// Double Exponential Moving Average computes the double exponential moving averages of the `_value` column.
// eg: A 5 point double exponential moving average would be called as such:
// from(bucket: "telegraf/autogen"):
//    |> range(start: -7d)
//    |> doubleEMA(n: 5)
doubleEMA = (n, tables=<-) =>
    tables
          |> exponentialMovingAverage(n:n)
          |> duplicate(column:"_value", as:"__ema")
          |> exponentialMovingAverage(n:n)
          |> map(fn: (r) => ({r with _value: 2.0*r.__ema - r._value}))
          |> drop(columns: ["__ema"])


// Triple Exponential Moving Average computes the triple exponential moving averages of the `_value` column.
// eg: A 5 point triple exponential moving average would be called as such:
// from(bucket: "telegraf/autogen"):
//    |> range(start: -7d)
//    |> tripleEMA(n: 5)
tripleEMA = (n, tables=<-) =>
	tables
		|> exponentialMovingAverage(n:n)
		|> duplicate(column:"_value", as:"__ema1")
		|> exponentialMovingAverage(n:n)
		|> duplicate(column:"_value", as:"__ema2")
		|> exponentialMovingAverage(n:n)
		|> map(fn: (r) => ({r with _value: 3.0*r.__ema1 - 3.0*r.__ema2 + r._value}))
		|> drop(columns: ["__ema1", "__ema2"])

// truncateTimeColumn takes in a time column t and a Duration unit and truncates each value of t to the given unit via map
// Change from _time to timeColumn once Flux Issue 1122 is resolved
truncateTimeColumn = (timeColumn="_time", unit, tables=<-) =>
    tables
        |> map(fn:(r) => ({r with _time: date.truncate(t: r._time, unit: unit)}))

// kaufmansER computes Kaufman's Efficiency Ratios of the `_value` column
kaufmansER = (n, tables=<-) =>
    tables
        |> chandeMomentumOscillator(n: n)
        |> map(fn:(r) => ({r with _value: (math.abs(x: r._value)/100.0)}))

toString   = (tables=<-) => tables |> map(fn:(r) => ({r with _value: string(v:r._value)}))
toInt      = (tables=<-) => tables |> map(fn:(r) => ({r with _value: int(v:r._value)}))
toUInt     = (tables=<-) => tables |> map(fn:(r) => ({r with _value: uint(v:r._value)}))
toFloat    = (tables=<-) => tables |> map(fn:(r) => ({r with _value: float(v:r._value)}))
toBool     = (tables=<-) => tables |> map(fn:(r) => ({r with _value: bool(v:r._value)}))
toTime     = (tables=<-) => tables |> map(fn:(r) => ({r with _value: time(v:r._value)}))
