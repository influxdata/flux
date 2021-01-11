# Tickscript package

The `tickscript` package can be used to convert TICKscripts to InfluxDB 2.x tasks.

Most TICKscript functions have the same or similar counterparts in Flux.
This package only provides a set of convenience functions for easier conversion and creation of custom `checks` executed as `tasks` and triggering `alerts` in InfluxDB 2.x.

To learn more about monitoring and alerting in InfluxDB 2.x and Flux, please see
 - [Monitor data and send alerts](https://docs.influxdata.com/influxdb/v2.0/monitor-alert/)
 - [Flux InfluxDB Monitor package](https://docs.influxdata.com/influxdb/v2.0/reference/flux/stdlib/monitor/)
 - [Process Data with InfluxDB tasks](https://docs.influxdata.com/influxdb/v2.0/process-data/)

## Available functions

- `select`
- `selectWindow`
- `groupBy`
- `compute`
- `join`
- `alert`
- `topic`

## Conversion guidelines

* Both `batch` and `stream` translates to `from(bucket: ...)` in Flux.

* Variable conversions  
  `var realm = 'qa'` becomes `realm = "qa"` in Flux  
  `var warnLevel = lambda: "device_count" > 2000` is a function `warnLevel = (r) => r["device_count"] > 2000` in Flux

* `every(duration)` and `offset(duration)` property maps to `every` and `offset` fields in task's option record.
  For better control or aligned scheduling, use `cron` option instead.

* `period(duration)` property maps to `range(start: -duration)` in Flux pipeline.

* `groupBy(columns)` maps to `group(columns)`.  
  Columns must include internal `_measurement` column.
  For convenience, function `groupBy` is provided in this package.  
  Flux results are grouped by all tags by default. To ungroup, call `group()` without argument.

* `groupBy(time(t), columns)` maps to `group(columns)` and `aggregateWindow(..., every: t, ...)`  
  The package provides convenience functions `groupBy(columns)` and `selectWindow(..., every(t), ...)` to achieve the same.

```js
    query('''
        SELECT sum("counter") AS "total_sum" WHERE ...
        ''')
        .groupBy(time(10s), 'host')
        .period(1m)
        .fill(0)
```
can be rewritten to Flux as
```js
    from(bucket: ...)
       |> range(start: -1m)
       |> filter(fn: (r) => ... )
       |> schema.fieldsAsCols()
       |> ts.grouBy(columns: ["host"])
       |> ts.selectWindow(column: "counter", fn: sum, as: "total_sum", every: 10s, defaultValue: 0)
```

* TICKscript `alert` provides property methods to send alerts to event handlers or a topic.
  In Flux, use `tickscript.topic()` pipeline function.

* `stateChangesOnly` is a filter available in InfluxDB notification rule.

* TICKscript pipeline with multiple alerts translates to multiple Flux pipelines, ie.

```js
var data = batch
    | query(...)
data
    | alert()
        .topic('A')
    | alert()
        .topic('B')
```

becomes

```js
data = from(bucket: ...)
    |> range(start: -duration)
    ...
data
    |> alert()
    |> topic('A')
data
    |> alert()
    |> topic('B')
```

## Functions

### tickscript.select

`tickscript.select()` is a convenience function for selecting a column with optional aggregation.
Intended to work like `"SELECT x AS y"` or `SELECT f(x) AS y` query (without time grouping).

Parameters:
- `column` - Existing column. Default value is `_value`.
- `fn` - Optional aggregation function. Default is none.
- `as` - Desired column name.

Examples:
```js
from(bucket: "test")
    ...
    |> ts.select(column: "message_rate", as: "MsgRate") // query('''SELECT "message_rate" AS "MsgRate"''')
```
```js
from(bucket: "test")
    ...
    |> ts.select(column: "counter", fn: mean, as: "count") // query('''SELECT mean("counter") AS "count"''')
```

### tickscript.selectWindow

`tickscript.selectWindow()` is a convenience function for selecting a column with time grouping and aggregation.
Intended to work like `"SELECT f(x) AS y"` query with `.groupBy(time(t), ...)`.

Parameters:
- `column` - Existing column. Default value is `_value`.
- `fn` - Aggregation function.
- `as` - Desired column name.
- `every` - Duration of windows.
- `defaultValue` - Value to fill windows with null aggregate value.

Examples:
```js
from(bucket: "test")
    ...
    |> ts.selectWindow(column: "counter", fn: mean, as: "rate", every: 1m, defaultValue: 0.0) // query('''"SELECT mean("counter") AS "rate"''').groupBy(time(1m))
```

### tickscript.groupBy

`tickscript.groupBy()` is a convenience function for result grouping.

Parameters:
- `columns` - Set of columns to group by. The implementation adds `_measurement` column required by underlying `monitor` package. 

_See "Examples" paragraph._

### tickscript.compute

`tickscript.compute()` is a convenience function for computing an aggregation on the data.
Intended to be used like TICKscript `|f(x).as(y)` where `f` is an aggregation function.

Parameters:
- `column` - Existing column. Default value is `_value`.
- `fn` - Aggregation function.
- `as` - Desired column name.

Examples:
```js
from(bucket: "test")
    ...
    |> ts.compute(column: "message_rate", fn: median, as: "median_message_rate") // query|median('message_rate').as('median_message_rate)
```

### tickscript.join

`tickscript.join()` is a convenience function for joining two streams.
It ensures the result has `_measurement` column and it is in the group key.

Parameters:
- `tables` - Record with two streams. See Flux `join` documentation for details.
- `on` - Optional list of columns to join on. Default is `["_time"]`.
- `measurement` - Measurement name.

Examples:
```js
requests = from(bucket: "test")
    ...
    |> ts.select(column: "counter", fn: sum, as: "total_sum")

failures = from(bucket: "test")
    ...
    |> ts.select(column: "counter", fn: sum, as: "failure_sum")

ts.join(tables: { requests: requests, failures: failures }, measurement: "xte")
    |> map(fn: (r) => ({ r with error_percent: float(v: failures.failure_sum) / float(v: requests.total_sum) }))
```

### tickscript.alert

`tickscript.alert()` checks input data and create alerts.
It requires pivoted data (just call `schema.fieldsAsCols()` before `tickscript.alert()`).

Parameters:
- `check` - Check description. It is a record required by the underlying `monitor` package with the following required fields:  
   `_check_id` - check unique identifier. Checks created via InfluxDB UI have GUID generated value, but the value can be arbitrary (but unique).  
   `_check_name` - check name  
   `_type` - check type (string value, must be `"custom"`)  
   `tags` - record with additional tags
- `id` - Function that constructs alert ID. Default is `(r) => "${r._check_id}"`.
- `message` - Function that constructs alert message. Default is `Check: ${r._check_name} is: ${r._level}"`.
- `details` - Function that constructs detailed alert message. Default is `(r) => ""`.
- `crit` - Predicate function that determines `crit` status. Default is `(r) => false`.
- `warn` - Predicate function that determines `warn` status. Default is `(r) => false`.
- `info` - Predicate function that determines `info` status. Default is `(r) => false`.
- `ok` - Predicate function that determines `info` status. Default is `(r) => true`.

_See Examples._

### tickscript.topic

`tickscript.topic()` sends alerts to a topic.

Parameters:
- `name` - Topic name.

_See Examples._

## Examples

### Using topic

##### Original TICKscript
```js
duration = 5m
every = 1m
db = 'gw'
tier = 'qa'
metric_type = 'kafka_message_in_rate'
h_threshold = 5000

batch
    |query('SELECT mean(' + metric_type + ') AS "KafkaMsgRate" FROM ' +  db  + ' WHERE realm = \'' + tier + '\' AND "host" =~ /^kafka.+.m02/')
    .period(duration)
    .every(frequency)
    .groupBy('host','realm')
   |alert()
        .id('Realm: {{index .Tags "realm"}} - Hostname: {{index .Tags "host"}} / Metric: ' + metric_type + ' threshold alert' )
        .message('{{.ID }}: {{ .Level }} - {{ index .Fields "KafkaMsgRate" | printf "%0.2f"}}')
        .crit(lambda: "KafkaMsgRate" > h_threshold)
        .stateChangesOnly()
        .topic('TESTING')
```

##### InfluxDB alert task

```js
import ts "contrib/bonitoo-io/tickscript"
import "influxdata/influxdb/schema"

// required task option
option task = {
  name: "Kafka Message Rate",
  every: 1m
}

// custom check info
check = {
  _check_id: "${task.name}-check",
  _check_name: "${task.name} Check",
  _type: "custom",
  tags: {}
} 

// converted TICKscript

duration = 5m
db = "gw"
tier = "qa"
metric_type = "kafka_message_in_rate"
h_threshold = 5000

from(bucket: db)
    |> range(start: -duration)
    |> filter(fn: (r) => r.measurement == db)
    |> filter(fn: (r) => r.realm == tier and r.host =~ /^kafka.+.m02/)
    |> filter(fn: (r) => r._field == metric_type)
    |> schema.fieldsAsCols()
    |> ts.groupBy(columns: ["host", "realm"])
    |> ts.select(column: metric_type, fn: mean, as: "KafkaMsgRate")
    |> ts.alert(
        check: check,
        id: (r) => "Realm: ${r.realm} - Hostname: ${r.host} / Metric: ${metric_type} threshold alert",
        message: (r) => "${r.id}: ${r._level} - ${string(v:r.KafkaMsgRate)}",
        crit: (r) => r.KafkaMsgRate > h_threshold
    )
    |> ts.topic(name: "TESTING")
```

##### Topic handler

Use notification rule as topic handler with additional filter for `_topic` tag value `"TESTING"`.
