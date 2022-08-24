// Package monitor provides tools for monitoring and alerting with InfluxDB.
//
// ## Metadata
// introduced: 0.39.0
// tag: monitor, alerts
//
package monitor


import "experimental"
import "influxdata/influxdb/v1"
import "influxdata/influxdb"

// bucket is the default bucket to store InfluxDB monitoring data in.
bucket = "_monitoring"

// write persists check statuses to an InfluxDB bucket.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
option write = (tables=<-) => tables |> experimental.to(bucket: bucket)

// log persists notification events to an InfluxDB bucket.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
option log = (tables=<-) => tables |> experimental.to(bucket: bucket)

// logs retrieves notification events stored in the `notifications` measurement
// in the `_monitoring` bucket.
//
// ## Parameters
// - start: Earliest time to include in results.
//
//      Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//      For example, `-1h`, `2019-08-28T22:00:00Z`, or `1567029600`.
//      Durations are relative to `now()`.
//
// - stop: Latest time to include in results. Default is `now()`.
//
//      Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//      For example, `-1h`, `2019-08-28T22:00:00Z`, or `1567029600`.
//      Durations are relative to `now()`.
//
// - fn: Predicate function that evaluates true or false.
//
//      Records or rows (`r`) that evaluate to `true` are included in output tables.
//      Records that evaluate to _null_ or `false` are not included in output tables.
//
// ## Examples
//
// ### Query notification events from the last hour
// ```no_run
// import "influxdata/influxdb/monitor"
//
// monitor.logs(
//     start: -2h,
//     fn: (r) => true,
// )
// ```
//
// ## Metadata
// tags: inputs
//
logs = (start, stop=now(), fn) =>
    influxdb.from(bucket: bucket)
        |> range(start: start, stop: stop)
        |> filter(fn: (r) => r._measurement == "notifications")
        |> filter(fn: fn)
        |> v1.fieldsAsCols()

// from retrieves check statuses stored in the `statuses` measurement in the
// `_monitoring` bucket.
//
// ## Parameters
// - start: Earliest time to include in results.
//
//      Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//      For example, `-1h`, `2019-08-28T22:00:00Z`, or `1567029600`.
//      Durations are relative to `now()`.
//
// - stop: Latest time to include in results. Default is `now()`.
//
//      Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//      For example, `-1h`, `2019-08-28T22:00:00Z`, or `1567029600`.
//      Durations are relative to `now()`
//
// - fn: Predicate function that evaluates true or false.
//
//      Records or rows (`r`) that evaluate to `true` are included in output tables.
//      Records that evaluate to _null_ or `false` are not included in output tables.
//
// ## Examples
//
// ### View critical check statuses from the last hour
// ```no_run
// import "influxdata/influxdb/monitor"
//
// monitor.from(
//     start: -1h,
//     fn: (r) => r._level == "crit",
// )
// ```
//
from = (start, stop=now(), fn=(r) => true) =>
    influxdb.from(bucket: bucket)
        |> range(start: start, stop: stop)
        |> filter(fn: (r) => r._measurement == "statuses")
        |> filter(fn: fn)
        |> v1.fieldsAsCols()

// levelOK is the string representation of the "ok" level.
levelOK = "ok"

// levelInfo is the string representation of the "info" level.
levelInfo = "info"

// levelWarn is the string representation of the "warn" level.
levelWarn = "warn"

// levelCrit is the string representation of the "crit" level.
levelCrit = "crit"

// levelUnknown is the string representation of the an unknown level.
levelUnknown = "unknown"

_stateChanges = (fromLevel="any", toLevel="any", tables=<-) => {
    toLevelFilter =
        if toLevel == "any" then
            (r) => r._level != fromLevel and exists r._level
        else
            (r) => r._level == toLevel
    fromLevelFilter =
        if fromLevel == "any" then
            (r) => r._level != toLevel and exists r._level
        else
            (r) => r._level == fromLevel

    return
        tables
            |> map(
                fn: (r) =>
                    ({r with level_value:
                            if toLevelFilter(r: r) then
                                1
                            else if fromLevelFilter(r: r) then
                                0
                            else
                                -10,
                    }),
            )
            |> duplicate(column: "_level", as: "____temp_level____")
            |> drop(columns: ["_level"])
            |> rename(columns: {"____temp_level____": "_level"})
            |> sort(columns: ["_source_timestamp", "_time"], desc: false)
            |> difference(columns: ["level_value"])
            |> filter(fn: (r) => r.level_value == 1)
            |> drop(columns: ["level_value"])
            |> experimental.group(mode: "extend", columns: ["_level"])
}

// notify sends a notification to an endpoint and logs it in the `notifications`
// measurement in the `_monitoring` bucket.
//
// ## Parameters
// - endpoint: A function that constructs and sends the notification to an endpoint.
// - data: Notification data to append to the output.
//
//     This data specifies which notification rule and notification endpoint to
//     associate with the sent notification.
//     The data record must contain the following properties:
//
//     - \_notification\_rule\_id
//     - \_notification\_rule\_name
//     - \_notification\_endpoint\_id
//     - \_notification\_endpoint\_name
//
//     The InfluxDB monitoring and alerting system uses `monitor.notify()` to store
//     information about sent notifications and automatically assigns these values.
//     If writing a custom notification task, we recommend using **unique arbitrary**
//     values for data record properties.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Send critical status notifications to Slack
// ```no_run
// import "influxdata/influxdb/monitor"
// import "influxdata/influxdb/secrets"
// import "slack"
//
// token = secrets.get(key: "SLACK_TOKEN")
//
// endpoint = slack.endpoint(token: token)(
//     mapFn: (r) => ({
//         channel: "Alerts",
//         text: r._message,
//         color: "danger",
//     }),
// )
//
// notification_data = {
//     _notification_rule_id: "0000000000000001",
//     _notification_rule_name: "example-rule-name",
//     _notification_endpoint_id: "0000000000000002",
//     _notification_endpoint_name: "example-endpoint-name",
// }
//
// monitor.from(
//     range: -5m,
//     fn: (r) => r._level == "crit",
// )
//     |> range(start: -5m)
//     |> monitor.notify(
//         endpoint: endpoint,
//         data: notification_data,
//     )
// ```
//
notify = (tables=<-, endpoint, data) =>
    tables
        |> experimental.set(o: data)
        |> experimental.group(mode: "extend", columns: experimental.objectKeys(o: data))
        |> map(
            fn: (r) =>
                ({r with _measurement: "notifications",
                    _status_timestamp: int(v: r._time),
                    _time: now(),
                }),
        )
        |> endpoint()
        |> experimental.group(mode: "extend", columns: ["_sent"])
        |> log()

// stateChangesOnly takes a stream of tables that contains a _level column
// and returns a stream of tables grouped by `_level` where each record
// represents a state change.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return records representing state changes
// ```
// import "array"
// import "influxdata/influxdb/monitor"
//
// data = array.from(
//     rows: [
//         {_time: 2021-01-01T00:00:00Z, _level: "ok"},
//         {_time: 2021-01-01T00:01:00Z, _level: "ok"},
//         {_time: 2021-01-01T00:02:00Z, _level: "warn"},
//         {_time: 2021-01-01T00:03:00Z, _level: "crit"},
//     ],
// )
//
// < data
// >     |> monitor.stateChangesOnly()
// ```
//
// ## Metadata
// tags: transformations
// introduced: 0.65.0
//
stateChangesOnly = (tables=<-) => {
    return
        tables
            |> map(
                fn: (r) =>
                    ({r with level_value:
                            if r._level == levelCrit then
                                4
                            else if r._level == levelWarn then
                                3
                            else if r._level == levelInfo then
                                2
                            else if r._level == levelOK then
                                1
                            else
                                0,
                    }),
            )
            |> duplicate(column: "_level", as: "____temp_level____")
            |> drop(columns: ["_level"])
            |> rename(columns: {"____temp_level____": "_level"})
            |> sort(columns: ["_source_timestamp", "_time"], desc: false)
            |> difference(columns: ["level_value"])
            |> filter(fn: (r) => r.level_value != 0)
            |> drop(columns: ["level_value"])
            |> experimental.group(mode: "extend", columns: ["_level"])
}

// stateChanges detects state changes in a stream of data with a `_level` column
// and outputs records that change from `fromLevel` to `toLevel`.
//
// ## Parameters
// - fromLevel: Level to detect a change from. Default is `"any"`.
// - toLevel: Level to detect a change to. Default is `"any"`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Detect when the state changes to critical
// ```
// import "array"
// import "influxdata/influxdb/monitor"
//
// data = array.from(
//     rows: [
//         {_time: 2021-01-01T00:00:00Z, _level: "ok"},
//         {_time: 2021-01-01T00:01:00Z, _level: "ok"},
//         {_time: 2021-01-01T00:02:00Z, _level: "warn"},
//         {_time: 2021-01-01T00:03:00Z, _level: "crit"},
//     ],
// )
//
// < data
// >     |> monitor.stateChanges(toLevel: "crit")
// ```
//
// ## Metadata
// introduced: 0.42.0
// tags: transformations
//
stateChanges = (fromLevel="any", toLevel="any", tables=<-) => {
    return
        if fromLevel == "any" and toLevel == "any" then
            tables |> stateChangesOnly()
        else
            tables |> _stateChanges(fromLevel: fromLevel, toLevel: toLevel)
}

// deadman detects when a group stops reporting data.
// It takes a stream of tables and reports if groups have been observed since time `t`.
//
// `monitor.deadman()` retains the most recent row from each input table and adds a `dead` column.
// If a record appears after time `t`, `monitor.deadman()` sets `dead` to `false`.
// Otherwise, `dead` is set to `true`.
//
// ## Parameters
// - t: Time threshold for the deadman check.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Detect if a host hasn’t reported since a specific time
// ```
// import "array"
// import "influxdata/influxdb/monitor"
//
// data = array.from(
//     rows: [
//         {_time: 2021-01-01T00:00:00Z, host: "a", _value: 1.2},
//         {_time: 2021-01-01T00:01:00Z, host: "a", _value: 1.3},
//         {_time: 2021-01-01T00:02:00Z, host: "a", _value: 1.4},
//         {_time: 2021-01-01T00:03:00Z, host: "a", _value: 1.3},
//     ],
// )
//     |> group(columns: ["host"])
//
// < data
// >     |> monitor.deadman(t: 2021-01-01T00:05:00Z)
// ```
//
// ### Detect if a host hasn't reported since a relative time
//
// Use `date.add()` to return a time value relative to a specified time.
//
// ```no_run
// import "influxdata/influxdb/monitor"
// import "date"
//
// from(bucket: "example-bucket")
//     |> range(start: -10m)
//     |> filter(fn: (r) => r._measurement == "example-measurement")
//     |> monitor.deadman(t: date.add(d: -5m, to: now()))
// ```
//
// ## Metadata
// tags: transformations
//
deadman = (t, tables=<-) =>
    tables
        |> max(column: "_time")
        |> map(fn: (r) => ({r with dead: r._time < t}))

// check checks input data and assigns a level (`ok`, `info`, `warn`, or `crit`)
// to each row based on predicate functions.
//
// `monitor.check()` stores statuses in the `_level` column and writes results
// to the `statuses` measurement in the `_monitoring` bucket.
//
// ## Parameters
// - crit: Predicate function that determines `crit` status. Default is `(r) => false`.
// - warn: Predicate function that determines `warn` status. Default is `(r) => false`.
// - info: Predicate function that determines `info` status. Default is `(r) => false`.
// - ok: Predicate function that determines `ok` status. `Default is (r) => true`.
// - messageFn: Predicate function that constructs a message to append to each row.
//
//   The message is stored in the `_message` column.
//
// - data: Check data to append to output used to identify this check.
//
//     This data specifies which notification rule and notification endpoint to
//     associate with the sent notification.
//     The data record must contain the following properties:
//
// 	   - **\_check\_id**: check ID _(string)_
// 	   - **\_check\_name**: check name _(string)_
// 	   - **\_type**: check type (threshold, deadman, or custom) _(string)_
// 	   - **tags**: Custom tags to append to output rows _(record)_
//
//     The InfluxDB monitoring and alerting system uses `monitor.check()` to
//     check statuses and automatically assigns these values.
//     If writing a custom check task, we recommend using **unique arbitrary**
//     values for data record properties.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Monitor InfluxDB disk usage collected by Telegraf
// ```no_run
// import "influxdata/influxdb/monitor"
//
// from(bucket: "telegraf")
//     |> range(start: -1h)
//     |> filter(
//         fn: (r) => r._measurement == "disk" and r._field == "used_percent",
//     )
//     |> monitor.check(
//         crit: (r) => r._value > 90.0,
//         warn: (r) => r._value > 80.0,
//         info: (r) => r._value > 70.0,
//         ok: (r) => r._value <= 60.0,
//         messageFn: (r) => if r._level == "crit" then
//             "Critical alert!! Disk usage is at ${r._value}%!"
//         else if r._level == "warn" then
//             "Warning! Disk usage is at ${r._value}%."
//         else if r._level == "info" then
//             "Disk usage is at ${r._value}%."
//         else
//             "Things are looking good.",
//         data: {
//             _check_name: "Disk Utilization (Used Percentage)",
//             _check_id: "disk_used_percent",
//             _type: "threshold",
//             tags: {},
//         },
//     )
// ```
//
// ## Metadata
// tags: transformations
//
check = (
    tables=<-,
    data,
    messageFn,
    crit=(r) => false,
    warn=(r) => false,
    info=(r) => false,
    ok=(r) => true,
) =>
    tables
        |> experimental.set(o: data.tags)
        |> experimental.group(mode: "extend", columns: experimental.objectKeys(o: data.tags))
        |> map(
            fn: (r) =>
                ({r with
                    _measurement: "statuses",
                    _source_measurement: r._measurement,
                    _type: data._type,
                    _check_id: data._check_id,
                    _check_name: data._check_name,
                    _level:
                        if crit(r: r) then
                            levelCrit
                        else if warn(r: r) then
                            levelWarn
                        else if info(r: r) then
                            levelInfo
                        else if ok(r: r) then
                            levelOK
                        else
                            levelUnknown,
                    _source_timestamp: int(v: r._time),
                    _time: now(),
                }),
        )
        |> map(fn: (r) => ({r with _message: messageFn(r: r)}))
        |> experimental.group(
            mode: "extend",
            columns: [
                "_source_measurement",
                "_type",
                "_check_id",
                "_check_name",
                "_level",
            ],
        )
        |> write()
