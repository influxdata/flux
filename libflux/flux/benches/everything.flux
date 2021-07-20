builtin from : (token: string, project: string, instance: string, table: string) => [T] where T: Record

// from will construct a table from the input rows.
//
// This function takes the `rows` parameter. The rows
// parameter is an array of records that will be constructed.
// All of the records must have the same keys and the same types
// for the values.
builtin from : (rows: [A]) => [A] where A: Record

rate = (tables=<-, every, groupColumns=[], unit=1s) => tables
    |> derivative(nonNegative: true, unit: unit)
    |> aggregateWindow(
        every: every,
        fn: (tables=<-, column) => tables
            |> mean(column: column)
            |> group(columns: groupColumns)
            |> experimental.group(columns: ["_start", "_stop"], mode: "extend")
            |> sum(),
    )

// Provides functions for geographic location filtering and grouping based on S2 cells.
// Units
option units = {
    distance: "km",
}

//
// Builtin GIS functions
//
// Returns boolean whether the region contains specified geometry.
builtin stContains : (region: A, geometry: B, units: {distance: string}) => bool where A: Record, B: Record

// Returns distance from given region to specified geometry.
builtin stDistance : (region: A, geometry: B, units: {distance: string}) => float where A: Record, B: Record

// Returns length of a curve.
builtin stLength : (geometry: A, units: {distance: string}) => float where A: Record

//
// Flux GIS ST functions
//
ST_Contains = (region, geometry, units=units) => stContains(region: region, geometry: geometry, units: units)
ST_Distance = (region, geometry, units=units) => stDistance(region: region, geometry: geometry, units: units)
ST_DWithin = (region, geometry, distance, units=units) => stDistance(region: region, geometry: geometry, units: units) <= distance
ST_Intersects = (region, geometry, units=units) => stDistance(region: region, geometry: geometry, units: units) <= 0.0
ST_Length = (geometry, units=units) => stLength(geometry: geometry, units: units)

// Non-standard
ST_LineString = (tables=<-) => tables
    |> reduce(
        fn: (r, accumulator) => ({
            __linestring: accumulator.__linestring + (if accumulator.__count > 0 then ", " else "") + string(v: r.lon) + " " + string(v: r.lat),
            __count: accumulator.__count + 1,
        }),
        identity: {
            __linestring: "",
            __count: 0,
        },
    )
    |> drop(columns: ["__count"])
    |> rename(columns: {__linestring: "st_linestring"})

//
// None of the following builtin functions are intended to be used by end users.
//
// Calculates grid (set of cell ID tokens) for given region and according to options.
builtin getGrid : (
    region: T,
    ?minSize: int,
    ?maxSize: int,
    ?level: int,
    ?maxLevel: int,
    units: {distance: string},
) => {level: int, set: [string]} where
    T: Record

// Returns level of specified cell ID token.
builtin getLevel : (token: string) => int

// Returns cell ID token for given cell or lat/lon point at specified level.
builtin s2CellIDToken : (?token: string, ?point: {lat: float, lon: float}, level: int) => string

// Returns lat/lon coordinates of given cell ID token.
builtin s2CellLatLon : (token: string) => {lat: float, lon: float}

//
// Flux functions
//
// Gets level of cell ID tag `s2cellID` from the first record from the first table in the stream.
_detectLevel = (tables=<-) => {
    _r0 = tables
        |> tableFind(fn: (key) => exists key.s2_cell_id)
        |> getRecord(idx: 0)
    _level = if exists _r0 then
        getLevel(token: _r0.s2_cell_id)
else
        666

    return _level
}

//
// Convenience functions
//
// Pivots values to row-wise sets.
toRows = (tables=<-) => tables
    |> v1.fieldsAsCols()

// Shapes data to meet the requirements of the geo package.
// Renames fields containing latitude and longitude values to lat and lon.
// Pivots values to row-wise sets.
// Generates an s2_cell_id tag for each reach using lat and lon values.
// Adds the s2_cell_id column to the group key.
shapeData = (tables=<-, latField, lonField, level) => tables
    |> map(
        fn: (r) => ({r with
            _field: if r._field == latField then
                "lat"
else if r._field == lonField then
                "lon"
else
                r._field,
        }),
    )
    |> toRows()
    |> map(
        fn: (r) => ({r with
            s2_cell_id: s2CellIDToken(point: {lat: r.lat, lon: r.lon}, level: level),
        }),
    )
    |> experimental.group(
        columns: ["s2_cell_id"],
        mode: "extend",
    )

//
// Filtering functions
//
// Filters records by a box, a circle or a polygon area using S2 cell ID tag.
// It is a coarse filter, as the grid always overlays the region, the result will likely contain records
// with lat/lon outside the specified region.
gridFilter = (tables=<-, region, minSize=24, maxSize=-1, level=-1, s2cellIDLevel=-1, units=units) => {
    _s2cellIDLevel = if s2cellIDLevel == -1 then
        tables
            |> _detectLevel()
else
        s2cellIDLevel
    _grid = getGrid(
        region: region,
        minSize: minSize,
        maxSize: maxSize,
        level: level,
        maxLevel: _s2cellIDLevel,
        units: units,
    )

    return tables
        |> filter(
            fn: (r) => if _grid.level == _s2cellIDLevel then
                contains(value: r.s2_cell_id, set: _grid.set)
else
                contains(value: s2CellIDToken(token: r.s2_cell_id, level: _grid.level), set: _grid.set),
        )
}

// Filters records by specified region.
// It is an exact filter and must be used after `toRows()` because it requires `lat` and `lon` columns in input row sets.
strictFilter = (tables=<-, region) => tables
    |> filter(fn: (r) => ST_Contains(region: region, geometry: {lat: r.lat, lon: r.lon}))

// Two-phase filtering by specified region.
// Checks to see if data is already pivoted and contains a lat column.
// Returns pivoted data.
filterRows = (tables=<-, region, minSize=24, maxSize=-1, level=-1, s2cellIDLevel=-1, strict=true) => {
    _columns = tables
        |> columns(column: "_value")
        |> tableFind(fn: (key) => true)
        |> getColumn(column: "_value")
    _rows = if contains(value: "lat", set: _columns) then
        tables
            |> gridFilter(
                region: region,
                minSize: minSize,
                maxSize: maxSize,
                level: level,
                s2cellIDLevel: s2cellIDLevel,
            )
else
        tables
            |> gridFilter(
                region: region,
                minSize: minSize,
                maxSize: maxSize,
                level: level,
                s2cellIDLevel: s2cellIDLevel,
            )
            |> toRows()
    _result = if strict then
        _rows
            |> strictFilter(region)
else
        _rows

    return _result
}

//
// Grouping functions
//
// intended to be used row-wise sets (i.e after `toRows()`)
// Groups data by area of size specified by level. Result is grouped by `newColumn`.
// Grouping levels: https://s2geometry.io/resources/s2cell_statistics.html
groupByArea = (tables=<-, newColumn, level, s2cellIDLevel=-1) => {
    _s2cellIDLevel = if s2cellIDLevel == -1 then
        tables
            |> _detectLevel()
else
        s2cellIDLevel
    _prepared = if level == _s2cellIDLevel then
        tables
            |> duplicate(column: "s2_cell_id", as: newColumn)
else
        tables
            |> map(
                fn: (r) => ({r with
                    _s2_cell_id_xxx: s2CellIDToken(point: {lat: r.lat, lon: r.lon}, level: level),
                }),
            )
            |> rename(columns: {_s2_cell_id_xxx: newColumn})

    return _prepared
        |> group(columns: [newColumn])
}

// Groups rows into tracks.
asTracks = (tables=<-, groupBy=["id", "tid"], orderBy=["_time"]) => tables
    |> group(columns: groupBy)
    |> sort(columns: orderBy)

// Parse will consume json data as bytes and return a value.
// Lists, objects, strings, booleans and float values can be produced.
// All numeric values are represented using the float type.
builtin parse : (data: bytes) => A

// Get submits an HTTP get request to the specified URL with headers
// Returns HTTP status code and body as a byte array
builtin get : (url: string, ?headers: A, ?timeout: duration) => {statusCode: int, body: bytes, headers: B} where A: Record, B: Record

from = (url) => c.from(csv: string(v: http.get(url: url).body))

// scrape enables scraping of a prometheus metrics endpoint and converts 
// that input into flux tables. Each metric is put into an individual flux 
// table, including each histogram and summary value.  
builtin scrape : (url: string) => [A] where A: Record

// histogramQuantile enables the user to calculate quantiles on a set of given values
// This function assumes that the given histogram data is being scraped or read from a 
// Prometheus source. 
histogramQuantile = (tables=<-, quantile) => tables
    |> filter(fn: (r) => r._measurement == "prometheus")
    |> group(mode: "except", columns: ["le", "_value", "_time"])
    |> map(fn: (r) => ({r with le: float(v: r.le)}))
    |> universe.histogramQuantile(quantile: quantile)

builtin addDuration : (d: duration, to: time) => time
builtin subDuration : (d: duration, from: time) => time

// An experimental version of group that has mode: "extend"
builtin group : (<-tables: [A], mode: string, columns: [string]) => [A] where A: Record

// objectKeys produces a list of the keys existing on the object
builtin objectKeys : (o: A) => [string] where A: Record

// set adds the values from the object onto each row of a table
builtin set : (<-tables: [A], o: B) => [C] where A: Record, B: Record, C: Record

// An experimental version of "to" that:
// - Expects pivoted data
// - Any column in the group key is made a tag in storage
// - All other columns are fields
// - An error will be thrown for incompatible data types
builtin to : (
    <-tables: [A],
    ?bucket: string,
    ?bucketID: string,
    ?org: string,
    ?orgID: string,
    ?host: string,
    ?token: string,
) => [A] where
    A: Record

// An experimental version of join.
builtin join : (left: [A], right: [B], fn: (left: A, right: B) => C) => [C] where A: Record, B: Record, C: Record
builtin chain : (first: [A], second: [B]) => [B] where A: Record, B: Record

// Aligns all tables to a common start time by using the same _time value for
// the first record in each table and incrementing all subsequent _time values
// using time elapsed between input records.
// By default, it aligns to tables to 1970-01-01T00:00:00Z UTC.
alignTime = (tables=<-, alignTo=time(v: 0)) => tables
    |> stateDuration(
        fn: (r) => true,
        column: "timeDiff",
        unit: 1ns,
    )
    |> map(fn: (r) => ({r with _time: time(v: int(v: alignTo) + r.timeDiff)}))
    |> drop(columns: ["timeDiff"])

builtin to : (
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
    ?valueColumns: [string],
) => [B] where
    A: Record,
    B: Record

fromRange = (bucket, start, stop=now()) => from(bucket: bucket)
    |> range(start: start, stop: stop)
filterMeasurement = (table=<-, measurement) => table |> filter(fn: (r) => r._measurement == measurement)
filterFields = (table=<-, fields=[]) => if length(arr: fields) == 0 then
    table
else
    table |> filter(fn: (r) => contains(value: r._field, set: fields))
inBucket = (bucket, measurement, start, stop=now(), fields=[], predicate=(r) => true) => fromRange(bucket: bucket, start: start, stop: stop)
    |> filterMeasurement(measurement)
    |> filter(fn: predicate)
    |> filterFields(fields)

// Transformation functions
builtin title : (v: string) => string
builtin toUpper : (v: string) => string
builtin toLower : (v: string) => string
builtin trim : (v: string, cutset: string) => string
builtin trimPrefix : (v: string, prefix: string) => string
builtin trimSpace : (v: string) => string
builtin trimSuffix : (v: string, suffix: string) => string
builtin trimRight : (v: string, cutset: string) => string
builtin trimLeft : (v: string, cutset: string) => string
builtin toTitle : (v: string) => string
builtin hasPrefix : (v: string, prefix: string) => bool
builtin hasSuffix : (v: string, suffix: string) => bool
builtin containsStr : (v: string, substr: string) => bool
builtin containsAny : (v: string, chars: string) => bool
builtin equalFold : (v: string, t: string) => bool
builtin compare : (v: string, t: string) => int
builtin countStr : (v: string, substr: string) => int
builtin index : (v: string, substr: string) => int
builtin indexAny : (v: string, chars: string) => int
builtin lastIndex : (v: string, substr: string) => int
builtin lastIndexAny : (v: string, chars: string) => int
builtin isDigit : (v: string) => bool
builtin isLetter : (v: string) => bool
builtin isLower : (v: string) => bool
builtin isUpper : (v: string) => bool
builtin repeat : (v: string, i: int) => string
builtin replace : (v: string, t: string, u: string, i: int) => string
builtin replaceAll : (v: string, t: string, u: string) => string
builtin split : (v: string, t: string) => [string]
builtin splitAfter : (v: string, t: string) => [string]
builtin splitN : (v: string, t: string, n: int) => [string]
builtin splitAfterN : (v: string, t: string, i: int) => [string]
builtin joinStr : (arr: [string], v: string) => string
builtin strlen : (v: string) => int
builtin substring : (v: string, start: int, end: int) => string

// Json parses an InfluxDB 1.x json result into a table stream.
builtin json : (?json: string, ?file: string) => [A] where A: Record

// Databases returns the list of available databases, it has no parameters.
builtin databases : (
    ?org: string,
    ?orgID: string,
    ?host: string,
    ?token: string,
) => [{
    organizationID: string,
    databaseName: string,
    retentionPolicy: string,
    retentionPeriod: int,
    default: bool,
    bucketID: string,
}]

// Maintain backwards compatibility by mapping the functions into the schema package.
fieldsAsCols = schema.fieldsAsCols
tagValues = schema.tagValues
measurementTagValues = schema.measurementTagValues
tagKeys = schema.tagKeys
measurementTagKeys = schema.measurementTagKeys
fieldKeys = schema.fieldKeys
measurementFieldKeys = schema.measurementFieldKeys
measurements = schema.measurements

// Package tasks is an experimental package.
// The API for this package is not stable and should not
// be counted on for production code.
// _zeroTime is a sentinel value for the zero time.
// This is used to mark that the lastSuccessTime has not been set.
builtin _zeroTime : time

// lastSuccessTime is the last time this task had run successfully.
option lastSuccessTime = _zeroTime

// _lastSuccess will return the time set on the option lastSuccessTime
// or it will return the orTime.
builtin _lastSuccess : (orTime: T, lastSuccessTime: time) => time where T: Timeable

// lastSuccess will return the last successful time a task ran
// within an influxdb task. If the task has not successfully run,
// the orTime will be returned.
lastSuccess = (orTime) => _lastSuccess(orTime, lastSuccessTime)
bucket = "_monitoring"

// Write persists the check statuses
option write = (tables=<-) => tables |> experimental.to(bucket: bucket)

// Log records notification events
option log = (tables=<-) => tables |> experimental.to(bucket: bucket)

// From retrieves the check statuses that have been stored.
from = (start, stop=now(), fn=(r) => true) => influxdb.from(bucket: bucket)
    |> range(start: start, stop: stop)
    |> filter(fn: (r) => r._measurement == "statuses")
    |> filter(fn: fn)
    |> v1.fieldsAsCols()

// levels describing the result of a check
levelOK = "ok"
levelInfo = "info"
levelWarn = "warn"
levelCrit = "crit"
levelUnknown = "unknown"
_stateChanges = (fromLevel="any", toLevel="any", tables=<-) => {
    toLevelFilter = if toLevel == "any" then
        (r) => r._level != fromLevel and exists r._level
else
        (r) => r._level == toLevel
    fromLevelFilter = if fromLevel == "any" then
        (r) => r._level != toLevel and exists r._level
else
        (r) => r._level == fromLevel

    return tables
        |> map(
            fn: (r) => ({r with
                level_value: if toLevelFilter(r: r) then
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
        |> sort(columns: ["_source_timestamp"], desc: false)
        |> difference(columns: ["level_value"])
        |> filter(fn: (r) => r.level_value == 1)
        |> drop(columns: ["level_value"])
        |> experimental.group(mode: "extend", columns: ["_level"])
}

// stateChangesOnly takes a stream of tables that contains a _level column and
// returns a stream of tables where each record in a table represents a state change
// of the _level column.
// Statuses are sorted by source timestamp, because default sort order of statuses may differ
// (`_time` column holds the time when check was executed) and that could result in detecting
// status changes at the wrong time or even false changes.
stateChangesOnly = (tables=<-) => {
    return tables
        |> map(
            fn: (r) => ({r with
                level_value: if r._level == levelCrit then
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
        |> sort(columns: ["_source_timestamp"], desc: false)
        |> difference(columns: ["level_value"])
        |> filter(fn: (r) => r.level_value != 0)
        |> drop(columns: ["level_value"])
        |> experimental.group(mode: "extend", columns: ["_level"])
}

// StateChanges takes a stream of tables, fromLevel, and toLevel and returns
// a stream of tables where status has gone from fromLevel to toLevel.
//
// StateChanges only operates on data with data where r._level exists.
stateChanges = (fromLevel="any", toLevel="any", tables=<-) => {
    return if fromLevel == "any" and toLevel == "any" then
        tables |> stateChangesOnly()
else
        tables |> _stateChanges(fromLevel: fromLevel, toLevel: toLevel)
}

// Notify will call the endpoint and log the results.
notify = (tables=<-, endpoint, data) => tables
    |> experimental.set(o: data)
    |> experimental.group(mode: "extend", columns: experimental.objectKeys(o: data))
    |> map(
        fn: (r) => ({r with
            _measurement: "notifications",
            _status_timestamp: int(v: r._time),
            _time: now(),
        }),
    )
    |> endpoint()
    |> experimental.group(mode: "extend", columns: ["_sent"])
    |> log()

// Logs retrieves notification events that have been logged.
logs = (start, stop=now(), fn) => influxdb.from(bucket: bucket)
    |> range(start: start, stop: stop)
    |> filter(fn: (r) => r._measurement == "notifications")
    |> filter(fn: fn)
    |> v1.fieldsAsCols()

// Deadman takes in a stream of tables and reports which tables
// were observed strictly before t and which were observed after.
//
deadman = (t, tables=<-) => tables
    |> max(column: "_time")
    |> map(fn: (r) => ({r with dead: r._time < t}))

// Check performs a check against its input using the given ok, info, warn and crit functions
// and writes the result to a system bucket.
check = (tables=<-, data, messageFn, crit=(r) => false, warn=(r) => false, info=(r) => false, ok=(r) => true) => tables
    |> experimental.set(o: data.tags)
    |> experimental.group(mode: "extend", columns: experimental.objectKeys(o: data.tags))
    |> map(
        fn: (r) => ({r with
            _measurement: "statuses",
            _source_measurement: r._measurement,
            _type: data._type,
            _check_id: data._check_id,
            _check_name: data._check_name,
            _level: if crit(r: r) then
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
    |> map(
        fn: (r) => ({r with
            _message: messageFn(r: r),
        }),
    )
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

builtin get : (key: string) => string

// fieldsAsCols is a special application of pivot that will automatically align fields within each measurement that have the same timestamp.
fieldsAsCols = (tables=<-) => tables
    |> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")

// TagValues returns the unique values for a given tag.
// The return value is always a single table with a single column "_value".
tagValues = (bucket, tag, predicate=(r) => true, start=-30d) => from(bucket: bucket)
    |> range(start: start)
    |> filter(fn: predicate)
    |> keep(columns: [tag])
    |> group()
    |> distinct(column: tag)

// MeasurementTagValues returns a single table with a single column "_value" that contains the
// The return value is always a single table with a single column "_value".
measurementTagValues = (bucket, measurement, tag) => tagValues(bucket: bucket, tag: tag, predicate: (r) => r._measurement == measurement)

// TagKeys returns the list of tag keys for all series that match the predicate.
// The return value is always a single table with a single column "_value".
tagKeys = (bucket, predicate=(r) => true, start=-30d) => from(bucket: bucket)
    |> range(start: start)
    |> filter(fn: predicate)
    |> keys()
    |> keep(columns: ["_value"])
    |> distinct()

// MeasurementTagKeys returns the list of tag keys for a specific measurement.
measurementTagKeys = (bucket, measurement) => tagKeys(bucket: bucket, predicate: (r) => r._measurement == measurement)

// FieldKeys is a special application of tagValues that returns field keys in a given bucket.
// The return value is always a single table with a single column, "_value".
fieldKeys = (bucket, predicate=(r) => true, start=-30d) => tagValues(bucket: bucket, tag: "_field", predicate: predicate, start: start)

// MeasurementFieldKeys returns field keys in a given measurement.
// The return value is always a single table with a single column, "_value".
measurementFieldKeys = (bucket, measurement, start=-30d) => fieldKeys(bucket: bucket, predicate: (r) => r._measurement == measurement, start: start)

// Measurements returns the list of measurements in a specific bucket.
measurements = (bucket) => tagValues(bucket: bucket, tag: "_measurement")

builtin from : (
    ?bucket: string,
    ?bucketID: string,
    ?org: string,
    ?orgID: string,
    ?host: string,
    ?token: string,
) => [{B with _measurement: string, _field: string, _time: time, _value: A}]
builtin to : (
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
    ?fieldFn: (r: A) => B,
) => [A] where
    A: Record,
    B: Record
builtin buckets : (
    ?org: string,
    ?orgID: string,
    ?host: string,
    ?token: string,
) => [{
    name: string,
    id: string,
    organizationID: string,
    retentionPolicy: string,
    retentionPeriod: int,
}]

// cardinality will return the cardinality of data for a given bucket.
// If a predicate is specified, then the cardinality only includes series
// that match the predicate.
builtin cardinality : (
    ?bucket: string,
    ?bucketID: string,
    ?org: string,
    ?orgID: string,
    ?host: string,
    ?token: string,
    start: A,
    ?stop: B,
    ?predicate: (r: {T with _measurement: string, _field: string, _value: S}) => bool,
) => [{_start: time, _stop: time, _value: int}] where
    A: Timeable,
    B: Timeable
builtin compile : (v: string) => regexp
builtin quoteMeta : (v: string) => string
builtin findString : (r: regexp, v: string) => string
builtin findStringIndex : (r: regexp, v: string) => [int]
builtin matchRegexpString : (r: regexp, v: string) => bool
builtin replaceAllString : (r: regexp, v: string, t: string) => string
builtin splitRegexp : (r: regexp, v: string, i: int) => [string]
builtin getString : (r: regexp) => string

option disableLogicalRules = [""]
option disablePhysicalRules = [""]

// `dedupKey` - adds a newline concatinated value of the sorted group key that is then sha256-hashed and hex-encoded to a column with the key `_pagerdutyDedupKey`.
builtin dedupKey : (<-tables: [A]) => [{A with _pagerdutyDedupKey: string}]

option defaultURL = "https://events.pagerduty.com/v2/enqueue"

// severity levels on status objects can be one of the following: ok,info,warn,crit,unknown
// but pagerduty only accepts critical, error, warning or info.
// severityFromLevel turns a level from the status object into a pagerduty severity
severityFromLevel = (level) => {
    lvl = strings.toLower(v: level)
    sev = if lvl == "warn" then
        "warning"
else if lvl == "crit" then
        "critical"
else if lvl == "info" then
        "info"
else if lvl == "ok" then
        "info"
else
        "error"

    return sev
}

// `actionFromLevel` converts a monitoring level to an action; "ok" becomes "resolve" everything else converts to "trigger".
actionFromLevel = (level) => if strings.toLower(v: level) == "ok" then "resolve" else "trigger"

// `sendEvent` sends an event to PagerDuty, the description of some of these parameters taken from the pagerduty documentation at https://v2.developer.pagerduty.com/docs/send-an-event-events-api-v2
// `pagerdutyURL` - sring - URL of the pagerduty endpoint.  Defaults to: `option defaultURL = "https://events.pagerduty.com/v2/enqueue"`
// `routingKey` - string - routingKey.
// `client` - string - name of the client sending the alert.
// `clientURL` - string - url of the client sending the alert.
// `dedupkey` - string - a per alert ID. It acts as deduplication key, that allows you to ack or change the severity of previous messages. Supports a maximum of 255 characters.
// `class` - string - The class/type of the event, for example ping failure or cpu load.
// `group` - string - Logical grouping of components of a service, for example app-stack.
// `severity` - string - The perceived severity of the status the event is describing with respect to the affected system. This can be critical, error, warning or info.
// `eventAction` - string - The type of event to send to PagerDuty (ex. trigger, resolve, acknowledge)
// `source` - string - The unique location of the affected system, preferably a hostname or FQDN.
// `summary` - string - A brief text summary of the event, used to generate the summaries/titles of any associated alerts. The maximum permitted length of this property is 1024 characters.
// `timestamp` - string - The time at which the emitting tool detected or generated the event, in RFC 3339 nano format.
sendEvent = (pagerdutyURL=defaultURL, routingKey, client, clientURL, dedupKey, class, group, severity, eventAction, source, summary, timestamp) => {
    payload = {
        summary: summary,
        timestamp: timestamp,
        source: source,
        severity: severity,
        group: group,
        class: class,
    }
    data = {
        payload: payload,
        routing_key: routingKey,
        dedup_key: dedupKey,
        event_action: eventAction,
        client: client,
        client_url: clientURL,
    }
    headers = {
        "Accept": "application/vnd.pagerduty+json;version=2",
        "Content-Type": "application/json",
    }
    enc = json.encode(v: data)

    return http.post(headers: headers, url: pagerdutyURL, data: enc)
}

// `endpoint` creates the endpoint for the PagerDuty external service.
// `url` - string - URL of the Pagerduty endpoint. Defaults to: "https://events.pagerduty.com/v2/enqueue".
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` parameter must be a function that returns an object with `routingKey`, `client`, `client_url`, `class`, `group`, `severity`, `eventAction`, `source`, `summary`, and `timestamp` as defined in the sendEvent function.
// Note that while sendEvent accepts a dedup key, endpoint gets the dedupkey from the groupkey of the input table instead of it being handled by the `mapFn`.
endpoint = (url=defaultURL) => (mapFn) => (tables=<-) => tables
    |> dedupKey()
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(
                    v: 2 == sendEvent(
                        pagerdutyURL: url,
                        routingKey: obj.routingKey,
                        client: obj.client,
                        clientURL: obj.clientURL,
                        dedupKey: r._pagerdutyDedupKey,
                        class: obj.class,
                        group: obj.group,
                        severity: obj.severity,
                        eventAction: obj.eventAction,
                        source: obj.source,
                        summary: obj.summary,
                        timestamp: obj.timestamp,
                    ) / 100,
                ),
            }
        },
    )

builtin second : (t: T) => int where T: Timeable
builtin minute : (t: T) => int where T: Timeable
builtin hour : (t: T) => int where T: Timeable
builtin weekDay : (t: T) => int where T: Timeable
builtin monthDay : (t: T) => int where T: Timeable
builtin yearDay : (t: T) => int where T: Timeable
builtin month : (t: T) => int where T: Timeable
builtin year : (t: T) => int where T: Timeable
builtin week : (t: T) => int where T: Timeable
builtin quarter : (t: T) => int where T: Timeable
builtin millisecond : (t: T) => int where T: Timeable
builtin microsecond : (t: T) => int where T: Timeable
builtin nanosecond : (t: T) => int where T: Timeable
builtin truncate : (t: T, unit: duration) => time where T: Timeable

Sunday = 0
Monday = 1
Tuesday = 2
Wednesday = 3
Thursday = 4
Friday = 5
Saturday = 6
January = 1
February = 2
March = 3
April = 4
May = 5
June = 6
July = 7
August = 8
September = 9
October = 10
November = 11
December = 12

builtin version : () => string
builtin tables : (n: int, ?nulls: float, ?tags: [{name: string, cardinality: int}]) => [{A with _time: time, _value: float}]

epoch = 1970-01-01T00:00:00Z
minTime = 1677-09-21T00:12:43.145224194Z
maxTime = 2262-04-11T23:47:16.854775806Z

builtin fail : () => bool
builtin yield : (<-v: A) => A
builtin makeRecord : (o: A) => B where A: Record, B: Record

// THIS PACKAGE IS NOT MEANT FOR EXTERNAL USE.
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
builtin emptyTable : () => [{_start: time, _stop: time, _time: time, _value: float}]

// extrapolatedRate() is a helper function that calculates extrapolated rates over
// counters and is used to implement PromQL's rate(), delta(), and increase() functions.
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#rate
// https://prometheus.io/docs/prometheus/latest/querying/functions/#increase
// https://prometheus.io/docs/prometheus/latest/querying/functions/#delta
builtin extrapolatedRate : (<-tables: [{A with _start: time, _stop: time, _time: time, _value: float}], ?isCounter: bool, ?isRate: bool) => [{B with _value: float}]

// holtWinters() implements functionality equivalent to PromQL's holt_winters()
// function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#holt_winters
builtin holtWinters : (<-tables: [{A with _time: time, _value: float}], ?smoothingFactor: float, ?trendFactor: float) => [{B with _value: float}]

// promqlHour() implements functionality equivalent to PromQL's hour() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#hour
builtin promqlHour : (timestamp: float) => float

// instantRate() is a helper function that calculates instant rates over
// counters and is used to implement PromQL's irate() and idelta() functions.
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#irate
// https://prometheus.io/docs/prometheus/latest/querying/functions/#idelta
builtin instantRate : (<-tables: [{A with _time: time, _value: float}], ?isRate: bool) => [{B with _value: float}]

// labelReplace implements functionality equivalent to PromQL's label_replace() function:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#label_replace
builtin labelReplace : (
    <-tables: [{A with _value: float}],
    source: string,
    destination: string,
    regex: string,
    replacement: string,
) => [{B with _value: float}]

// linearRegression implements linear regression functionality required to implement
// PromQL's deriv() and predict_linear() functions:
//
// https://prometheus.io/docs/prometheus/latest/querying/functions/#deriv
// https://prometheus.io/docs/prometheus/latest/querying/functions/#predict_linear
builtin linearRegression : (<-tables: [{A with _time: time, _stop: time, _value: float}], ?predict: bool, ?fromNow: float) => [{B with _value: float}]

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
builtin promHistogramQuantile : (
    <-tables: [A],
    ?quantile: float,
    ?countColumn: string,
    ?upperBoundColumn: string,
    ?valueColumn: string,
) => [B] where
    A: Record,
    B: Record

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
        tables
            |> universe.quantile(q: q, method: method)
else if q < 0.0 then
        tables
            |> reduce(identity: {_value: math.mInf(sign: -1)}, fn: (r, accumulator) => accumulator)
else
        tables
            |> reduce(identity: {_value: math.mInf(sign: 1)}, fn: (r, accumulator) => accumulator)
join = experimental.join

// pass will pass any incoming tables directly next to the following transformation.
// It is best used to interrupt any planner rules that rely on a specific ordering.
builtin pass : (<-tables: [A]) => [A] where A: Record

// respondersToJSON converts an array of responder strings to JSON array that can be embedded into an alert message
builtin respondersToJSON : (v: [string]) => string

// `sendAlert` sends a message that creates an alert in Opsgenie. See https://docs.opsgenie.com/docs/alert-api#create-alert for details.
// `url`         - string - Opsgenie API URL. Defaults to "https://api.opsgenie.com/v2/alerts". 
// `apiKey`      - string - API Authorization key. 
// `message`     - string - Alert message text, at most 130 characters. 
// `alias`       - string - Opsgenie alias, at most 250 characters that are used to de-deduplicate alerts. Defaults to message. 
// `description` - string - Description field of an alert, at most 15000 characters. Optional. 
// `priority`    - string - "P1", "P2", "P3", "P4" or "P5". Defaults to "P3". 
// `responders`  - array  - Array of strings to identify responder teams or teams, a 'user:' prefix is used for users, 'teams:' prefix for teams. 
// `tags`        - array  - Array of string tags. Optional. 
// `entity`      - string - Entity of the alert, used to specify domain of the alert. Optional. 
// `actions`     - array  - Array of strings that specifies actions that will be available for the alert. 
// `details`     - string - Additional details of an alert, it must be a JSON-encoded map of key-value string pairs. 
// `visibleTo`   - array  - Arrays of teams and users that the alert will become visible to without sending any notification. Optional. 
sendAlert = (url="https://api.opsgenie.com/v2/alerts", apiKey, message, alias="", description="", priority="P3", responders=[], tags=[], entity="", actions=[], visibleTo=[], details="{}") => {
    headers = {
        "Content-Type": "application/json; charset=utf-8",
        "Authorization": "GenieKey " + apiKey,
    }
    cutEncode = (v, max, defV="") => {
        v2 = if strings.strlen(v: v) != 0 then v else defV

        return if strings.strlen(v: v2) > max then
            string(v: json.encode(v: "${strings.substring(v: v2, start: 0, end: max)}"))
else
            string(v: json.encode(v: v2))
    }
    body = "{
\"message\": ${cutEncode(v: message, max: 130)},
\"alias\": ${cutEncode(v: alias, max: 512, defV: message)},
\"description\": ${cutEncode(v: description, max: 15000)},
\"responders\": ${respondersToJSON(v: responders)},
\"visibleTo\": ${respondersToJSON(v: visibleTo)},
\"actions\": ${string(v: json.encode(v: actions))},
\"tags\": ${string(v: json.encode(v: tags))},
\"details\": ${details},
\"entity\": ${cutEncode(v: entity, max: 512)},
\"priority\": ${cutEncode(v: priority, max: 2)}
}"

    return http.post(headers: headers, url: url, data: bytes(v: body))
}

// `endpoint` creates a factory function that creates a target function for pipeline `|>` to send alerts to opsgenie for each table row.
// `url`         - string - Opsgenie API URL. Defaults to "https://api.opsgenie.com/v2/alerts". 
// `apiKey`      - string - API Authorization key. 
// `entity`      - string - Entity of the alert, used to specify domain of the alert. Optional. 
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with all properties defined in the `sendAlert` function arguments (except url, apiKey and entity).
endpoint = (url="https://api.opsgenie.com/v2/alerts", apiKey, entity="") => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(
                    v: 2 == sendAlert(
                        url: url,
                        apiKey: apiKey,
                        entity: entity,
                        message: obj.message,
                        alias: obj.alias,
                        description: obj.description,
                        priority: obj.priority,
                        responders: obj.responders,
                        tags: obj.tags,
                        actions: obj.actions,
                        visibleTo: obj.visibleTo,
                        details: obj.details,
                    ) / 100,
                ),
            }
        },
    )

option defaultURL = "https://api.telegram.org/bot"
option defaultParseMode = "MarkdownV2"
option defaultDisableWebPagePreview = false
option defaultSilent = true

// `message` sends a single message to a Telegram channel using the API descibed in https://core.telegram.org/bots/api#sendmessage
// `url` - string - URL of the telegram bot endpoint. Defaults to: "https://api.telegram.org/bot"
// `token` - string - Required telegram bot token string, such as 123456789:AAxSFgij0ln9C7zUKnr4ScDi5QXTGF71S
// `channel` - string - Required id of the telegram channel.
// `text` - string - The text to display.
// `parseMode` - string - Parse mode of the message text per https://core.telegram.org/bots/api#formatting-options . Defaults to "MarkdownV2"
// `disableWebPagePreview` - bool - Disables preview of web links in the sent messages when "true". Defaults to "false"
// `silent` - bool - Messages are sent silently (https://telegram.org/blog/channels-2-0#silent-messages) when "true". Defaults to "true".
message = (url=defaultURL, token, channel, text, parseMode=defaultParseMode, disableWebPagePreview=defaultDisableWebPagePreview, silent=defaultSilent) => {
    data = {
        chat_id: channel,
        text: text,
        parse_mode: parseMode,
        disable_web_page_preview: disableWebPagePreview,
        disable_notification: silent,
    }
    headers = {
        "Content-Type": "application/json; charset=utf-8",
    }
    enc = json.encode(v: data)

    return http.post(headers: headers, url: url + token + "/sendMessage", data: enc)
}

// `endpoint` creates a factory function that creates a target function for pipeline `|>` to send messages to telegram for each table row.
// `url` - string - URL of the telegram bot endpoint. Defaults to: "https://api.telegram.org/bot"
// `token` - string - Required telegram bot token string, such as 123456789:AAxSFgij0ln9C7zUKnr4ScDi5QXTGF71S
// `parseMode` - string - Parse mode of the message text per https://core.telegram.org/bots/api#formatting-options . Defaults to "MarkdownV2"
// `disableWebPagePreview` - bool - Disables preview of web links in the sent messages when "true". Defaults to "false"
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `channel`, `text`, and `silent`, as defined in the `message` function arguments.
endpoint = (url=defaultURL, token, parseMode=defaultParseMode, disableWebPagePreview=defaultDisableWebPagePreview) => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(
                    v: 2 == message(
                        url: url,
                        token: token,
                        channel: obj.channel,
                        text: obj.text,
                        parseMode: parseMode,
                        disableWebPagePreview: disableWebPagePreview,
                        silent: obj.silent,
                    ) / 100,
                ),
            }
        },
    )

// `summaryCutoff` is used 
option summaryCutoff = 70

// `message` sends a single message to Microsoft Teams via incoming web hook.
// `url` - string - incoming web hook URL
// `title` - string - Message card title.
// `text` - string - Message card text.
// `summary` - string - Message card summary, it can be an empty string to generate summary from text.
message = (url, title, text, summary="") => {
    headers = {
        "Content-Type": "application/json; charset=utf-8",
    }

    // see https://docs.microsoft.com/en-us/outlook/actionable-messages/message-card-reference#card-fields
    // using string body, object cannot be used because '@' is an illegal character in the object property key
    summary2 = if summary == "" then
        text
else
        summary
    shortSummary = if strings.strlen(v: summary2) > summaryCutoff then
        "${strings.substring(v: summary2, start: 0, end: summaryCutoff)}..."
else
        summary2
    body = "{
\"@type\": \"MessageCard\",
\"@context\": \"http://schema.org/extensions\",
\"title\": ${string(v: json.encode(v: title))},
\"text\": ${string(v: json.encode(v: text))},
\"summary\": ${string(v: json.encode(v: shortSummary))}
}"

    return http.post(headers: headers, url: url, data: bytes(v: body))
}

// `endpoint` creates the endpoint for the Microsoft Teams external service.
// `url` - string - URL of the incoming web hook.
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `title`, `text`, and `summary`, as defined in the `message` function arguments.
endpoint = (url) => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(
                    v: 2 == message(
                        url: url,
                        title: obj.title,
                        text: obj.text,
                        summary: if exists obj.summary then obj.summary else "",
                    ) / 100,
                ),
            }
        },
    )

// toSensuName translates a string value to a Sensu name.
// Characters not being [a-zA-Z0-9_.\-] are replaced by underscore.
builtin toSensuName : (v: string) => string

// `event` sends a single event to Sensu as described in https://docs.sensu.io/sensu-go/latest/api/events/#create-a-new-event API. 
// `url` - string - base URL of [Sensu API](https://docs.sensu.io/sensu-go/latest/migrate/#architecture) without a trailing slash, for example "http://localhost:8080" .
// `apiKey` - string - Sensu [API Key](https://docs.sensu.io/sensu-go/latest/operations/control-access/).
// `checkName` - string - Check name, it can contain [a-zA-Z0-9_.\-] characters, other characters are replaced by underscore.
// `text` - string - The event text (named output in a Sensu Event).
// `handlers` - array<string> - Sensu handlers to execute, optional.
// `status` - int - The event status, 0 (default) indicates "OK", 1 indicates "WARNING", 2 indicates "CRITICAL", any other value indicates an “UNKNOWN” or custom status.
// `state` - string - The event state can be "failing", "passing" or "flapping". Defaults to "passing" for 0 status, "failing" otherwise. 
// `namespace` - string - The Sensu namespace. Defaults to "default".
// `entityName` - string - Source of the event, it can contain [a-zA-Z0-9_.\-] characters, other characters are replaced by underscore. Defaults to "influxdb".
event = (url, apiKey, checkName, text, handlers=[], status=0, state="", namespace="default", entityName="influxdb") => {
    data = {
        entity: {
            entity_class: "proxy",
            metadata: {
                name: toSensuName(v: entityName),
            },
        },
        check: {
            output: text,
            state: if state != "" then state else if status == 0 then "passing" else "failing",
            status: status,
            handlers: handlers,
            interval: 60,
            // required
            metadata: {
                name: toSensuName(v: checkName),
            },
        },
    }
    headers = {
        "Content-Type": "application/json; charset=utf-8",
        "Authorization": "Key " + apiKey,
    }
    enc = json.encode(v: data)

    return http.post(headers: headers, url: url + "/api/core/v2/namespaces/" + namespace + "/events", data: enc)
}

// `endpoint` creates a factory function that creates a target function for pipeline `|>` to send event to Sensu for each table row.
// `url` - string - base URL of [Sensu API](https://docs.sensu.io/sensu-go/latest/migrate/#architecture) without a trailing slash, for example "http://localhost:8080" .
// `apiKey` - string - Sensu [API Key](https://docs.sensu.io/sensu-go/latest/operations/control-access/).
// `handlers` - array<string> - Sensu handlers to execute.
// `namespace` - string - The Sensu namespace. Defaults to "default".
// `entityName` - string - Source of the event, it can contain [a-zA-Z0-9_.\-] characters, other characters are replaced by underscore. Defaults to "influxdb".
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `checkName`, `text`, and `status`, as defined in the `event` function arguments.
endpoint = (url, apiKey, handlers=[], namespace="default", entityName="influxdb") => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(
                    v: 2 == event(
                        url: url,
                        apiKey: apiKey,
                        checkName: obj.checkName,
                        text: obj.text,
                        handlers: handlers,
                        status: obj.status,
                        namespace: namespace,
                        entityName: entityName,
                    ) / 100,
                ),
            }
        },
    )

// table will aggregate columns and create tables with a single
// row containing the aggregated value.
//
// This function takes a single parameter of `columns`. The parameter
// is an object with the output column name as the key and the aggregate
// object as the value.
//
// The aggregate object is composed of at least the following required attributes:
//     column = string
//         The column name for the input.
//     init = (values) -> state
//         An initial function to compute the initial state of the
//         output. This can return either the final aggregate or a
//         temporary state object that can be used to compute the
//         final aggregate. The values parameter will always be a
//         non-empty array of values from the specified column.
//     reduce = (values, state) -> state
//         A function that takes in another buffer of values
//         and the current state of the aggregate and computes
//         the updated state.
//     compute = (state) -> value
//         A function that takes the state and computes the final
//         aggregate.
//     fill = value
//         The value passed to fill, if present, will determine what
//         the aggregate does when there are no values.
//         This can either be a value or one of the predefined
//         identifiers of null or none.
//         This value must be the same type as the value return from
//         compute.
//
// An example of usage is:
//     tables |> aggregate.table(columns: {
//         "min_bottom_degrees": aggregate.min(column: "bottom_degrees"),
//     ])
builtin table : (<-tables: [A], columns: C) => [B] where A: Record, B: Record, C: Record

// window will aggregate columns and create tables by
// organizing incoming points into windows.
//
// Each table will have two additional columns: start and stop.
// These are the start and stop times for each interval.
// It is not possible to use start or stop as destination column
// names with this function. The start and stop columns are not
// added to the group key.
//
// The same options as for table apply to window.
// In addition to those options, window requires one
// additional parameter.
//     every = duration
//         The duration between the start of each interval.
//
// Along with the above required option, there are a few additional
// optional parameters.
//     time = string
//         The column name for the time input.
//         This defaults to _time or time, whichever is earlier in
//         the list of columns.
//     period = duration
//         The length of the interval. This defaults to the
//         every duration.
builtin window : (
    <-tables: [A],
    ?time: string,
    every: duration,
    ?period: duration,
    columns: C,
) => [B] where
    A: Record,
    B: Record,
    C: Record

// null is a sentinel value for fill that will fill
// in a null value if there were no values for an interval.
builtin null : A

// none is a sentinel value for fill that will skip
// emitting a row if there are no values for an interval.
builtin none : A

// define will define an aggregate function.
define = (init, reduce, compute, fill=null) => (column="_value", fill=fill) => ({
    column: column,
    init: init,
    reduce: reduce,
    compute: compute,
    fill: fill,
})
_make_selector = (fn) => define(
    init: (values) => fn(values),
    reduce: (values, state) => {
        v = fn(values)

        return fn(values: [state, v])
    },
    compute: (state) => state,
)

// min constructs a min aggregate or selector for the column.
min = _make_selector(fn: math.min)

// max constructs a max aggregate or selector for the column.
max = _make_selector(fn: math.max)

// sum constructs a sum aggregate for the column.
sum = define(
    init: (values) => math.sum(values),
    reduce: (values, state) => {
        return state + math.sum(values)
    },
    compute: (state) => state,
)

// count constructs a count aggregate for the column.
count = define(
    init: (values) => length(arr: values),
    reduce: (values, state) => {
        return state + length(arr: values)
    },
    compute: (state) => state,
    fill: 0,
)

// mean constructs a mean aggregate for the column.
mean = define(
    init: (values) => ({
        sum: math.sum(values),
        count: length(arr: values),
    }),
    reduce: (values, state) => ({
        sum: state.sum + math.sum(values),
        count: state.count + length(arr: values),
    }),
    compute: (state) => float(v: state.sum) / float(v: state.count),
)

builtin minIndex : (values: [A]) => int where A: Numeric

min = (values) => {
    index = minIndex(values)

    return values[index]
}

builtin maxIndex : (values: [A]) => int where A: Numeric

max = (values) => {
    index = maxIndex(values)

    return values[index]
}

builtin sum : (values: [A]) => A where A: Numeric

// map will map each of the rows to a new value.
// The function will be invoked for each row and the
// return value will be used as the values in the output
// row.
//
// The record that is passed to the function will contain
// all of the keys and values in the record including group
// keys, but the group key cannot be changed. Attempts to
// change the group key will be ignored.
//
// The returned record does not need to contain values that are
// part of the group key.
builtin map : (<-tables: [A], fn: (r: A) => B) => [B] where A: Record, B: Record

// _mask will hide the given columns from downstream
// transformations. It will not perform any copies and
// it will not regroup. This should only be used when
// the user knows it can't cause a key conflict.
builtin _mask : (<-tables: [A], columns: [string]) => [B] where A: Record, B: Record

// from will retrieve data from a bucket between the start and stop time.
// This version of from is the equivalent of doing from |> range
// as a single call.
from = (bucket, start, stop=now(), org="", host="", token="") => {
    source = if org != "" and host != "" and token != "" then
        influxdb.from(bucket, org, host, token)
else if org != "" and token != "" then
        influxdb.from(bucket, org, token)
else if org != "" and host != "" then
        influxdb.from(bucket, org, host)
else if host != "" and token != "" then
        influxdb.from(bucket, host, token)
else if org != "" then
        influxdb.from(bucket, org)
else if host != "" then
        influxdb.from(bucket, host)
else if token != "" then
        influxdb.from(bucket, token)
else
        influxdb.from(bucket)

    return source |> range(start, stop)
}

// _from allows us to reference the from function from
// within the select call which has a function parameter
// with the same name.
_from = from

// select will select data from an influxdb instance within
// the range between `start` and `stop` from the bucket specified by
// the `from` parameter. It will select the specific measurement
// and it will only include fields that are included in the list of
// `fields`.
//
// In order to filter by tags, the `where` function can be used to further
// limit the amount of data selected.
select = (from, start, stop=now(), m, fields=[], org="", host="", token="", where=(r) => true) => {
    bucket = from
    tables = _from(
        bucket,
        start,
        stop,
        org,
        host,
        token,
    )
        |> filter(fn: (r) => r._measurement == m)
        |> filter(fn: where)
    nfields = length(arr: fields)
    fn = if nfields == 0 then
        (r) => true
else if nfields == 1 then
        (r) => r._field == fields[0]
else if nfields == 2 then
        (r) => r._field == fields[0] or r._field == fields[1]
else if nfields == 3 then
        (r) => r._field == fields[0] or r._field == fields[1] or r._field == fields[2]
else if nfields == 4 then
        (r) => r._field == fields[0] or r._field == fields[1] or r._field == fields[2] or r._field == fields[3]
else if nfields == 5 then
        (r) => r._field == fields[0] or r._field == fields[1] or r._field == fields[2] or r._field == fields[3] or r._field == fields[4]
else
        (r) => contains(value: r._field, set: fields)

    return tables
        |> filter(fn)
        |> v1.fieldsAsCols()
        |> _mask(columns: ["_measurement", "_start", "_stop"])
}

// duration will calculate the duration between records
// for each record. The duration calculated is between
// the current record and the next. The last record will
// compare against either the stopColum (default: _stop)
// or a stop timestamp value.
//
// `timeColumn` - Optional string. Default '_time'. The value used to calculate duration
// `columnName` - Optional string. Default 'duration'. The name of the result column
// `stopColumn` - Optional string. Default '_stop'. The name of the column to compare the last record on
// `stop` - Optional Time. Use a fixed time to compare the last record against instead of stop column.
builtin duration : (
    <-tables: [A],
    ?unit: duration,
    ?timeColumn: string,
    ?columnName: string,
    ?stopColumn: string,
    ?stop: time,
) => [B] where
    A: Record,
    B: Record

mad = (table=<-, threshold=3.0) => {
    // MEDiXi = med(x)
    data = table |> group(columns: ["_time"], mode: "by")
    med = data |> median(column: "_value")

    // diff = |Xi - MEDiXi| = math.abs(xi-med(xi))
    diff = join(tables: {data: data, med: med}, on: ["_time"], method: "inner")
        |> map(fn: (r) => ({r with _value: math.abs(x: r._value_data - r._value_med)}))
        |> drop(columns: ["_start", "_stop", "_value_med", "_value_data"])

    // The constant k is needed to make the estimator consistent for the parameter of interest.
    // In the case of the usual parameter at Gaussian distributions k = 1.4826
    k = 1.4826

    // MAD =  k * MEDi * |Xi - MEDiXi| 
    diff_med = diff
        |> median(column: "_value")
        |> map(fn: (r) => ({r with MAD: k * r._value}))
        |> filter(fn: (r) => r.MAD > 0.0)
    output = join(tables: {diff: diff, diff_med: diff_med}, on: ["_time"], method: "inner")
        |> map(fn: (r) => ({r with _value: r._value_diff / r._value_diff_med}))
        |> map(
            fn: (r) => ({r with
                level: if r._value >= threshold then
                    "anomaly"
else
                    "normal",
            }),
        )

    return output
}

// performs linear regression, calculates y_hat, and residuals squared (rse) 
linearRegression = (tables=<-) => {
    renameAndSum = tables
        |> rename(columns: {_value: "y"})
        |> map(fn: (r) => ({r with x: 1.0}))
        |> cumulativeSum(columns: ["x"])
    t = renameAndSum
        |> reduce(
            fn: (r, accumulator) => ({
                sx: r.x + accumulator.sx,
                sy: r.y + accumulator.sy,
                N: accumulator.N + 1.0,
                sxy: r.x * r.y + accumulator.sxy,
                sxx: r.x * r.x + accumulator.sxx,
            }),
            identity: {
                sxy: 0.0,
                sx: 0.0,
                sy: 0.0,
                sxx: 0.0,
                N: 0.0,
            },
        )
        |> tableFind(fn: (key) => true)
        |> getRecord(idx: 0)
    xbar = t.sx / t.N
    ybar = t.sy / t.N
    slope = (t.sxy - xbar * ybar * t.N) / (t.sxx - t.N * xbar * xbar)
    intercept = ybar - slope * xbar
    y_hat = (r) => ({r with
        y_hat: slope * r.x + intercept,
        slope: slope,
        sx: t.sx,
        sxy: t.sxy,
        sxx: t.sxx,
        N: t.N,
        sy: t.sy,
    })
    rse = (r) => ({r with errors: (r.y - r.y_hat) ^ 2.0})
    output = renameAndSum
        |> map(fn: y_hat)
        |> map(fn: rse)

    return output
}

option discordURL = "https://discordapp.com/api/webhooks/"

// `webhookToken` - string - the secure token of the webhook.
// `webhookID` - string - the ID of the webhook.
// `username` - string - username posting the message.
// `content` - string - the text to display in discord.
// `avatar_url` -  override the default avatar of the webhook.
send = (webhookToken, webhookID, username, content, avatar_url="") => {
    data = {
        username: username,
        content: content,
        avatar_url: avatar_url,
    }
    headers = {
        "Content-Type": "application/json",
    }
    encode = json.encode(v: data)

    return http.post(headers: headers, url: discordURL + webhookID + "/" + webhookToken, data: encode)
}

// `endpoint` creates a factory function that creates a target function for pipeline `|>` to send messages to discord for each table row.
// `webhookToken` - string - the secure token of the webhook.
// `webhookID` - string - the ID of the webhook.
// `username` - string - username posting the message.
// `avatar_url` -  override the default avatar of the webhook.
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `content`, as defined in the `send` function arguments.
endpoint = (webhookToken, webhookID, username, avatar_url="") => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(
                    v: 2 == send(
                        webhookToken: webhookToken,
                        webhookID: webhookID,
                        username: username,
                        avatar_url: avatar_url,
                        content: obj.content,
                    ) / 100,
                ),
            }
        },
    )

//Final working code as of August 11, 2020
//Currently supports single field classification and binary data sets 
//Please ensure Ruby is installed
naiveBayes = (tables=<-, myClass, myField, myMeasurement) => {
    training_data = tables
        |> range(start: 2020-01-02T00:00:00Z, stop: 2020-01-06T23:00:00Z)
        //data for 3 days 
        |> filter(fn: (r) => r["_measurement"] == myMeasurement and r["_field"] == myField)
        |> group()

    //|> yield(name: "trainingData")
    test_data = tables
        |> range(start: 2020-01-01T00:00:00Z, stop: 2020-01-01T23:00:00Z)
        //data for 1 day
        |> filter(fn: (r) => r["_measurement"] == myMeasurement and r["_field"] == myField)
        |> group()

    //|> yield(name: "test data")
    //data preparation 
    r = training_data
        |> group(columns: ["_field"])
        |> count()
        |> tableFind(fn: (key) => key._field == myField)
    r2 = getRecord(table: r, idx: 0)
    total_count = r2._value
    P_Class_k = training_data
        |> group(columns: [myClass, "_field"])
        |> count()
        |> map(fn: (r) => ({r with p_k: float(v: r._value) / float(v: total_count), tc: total_count}))
        |> group()

    //one table for each class, where r.p_k == P(Class_k)
    P_value_x = training_data
        |> group(columns: ["_value", "_field"])
        |> count(column: myClass)
        |> map(fn: (r) => ({r with p_x: float(v: r.airborne) / float(v: total_count), tc: total_count}))

    // one table for each value, where r.p_x == P(value_x)
    P_k_x = training_data
        |> group(columns: ["_field", "_value", myClass])
        |> reduce(
            fn: (r, accumulator) => ({sum: 1.0 + accumulator.sum}),
            identity: {sum: 0.0},
        )
        |> group()

    // one table for each value and Class pair, where r.p_k_x == P(value_x | Class_k)
    P_k_x_class = join(tables: {P_k_x: P_k_x, P_Class_k: P_Class_k}, on: [myClass], method: "inner")
        |> group(columns: [myClass, "_value_P_k_x"])
        |> limit(n: 1)
        |> map(fn: (r) => ({r with P_x_k: r.sum / float(v: r._value_P_Class_k)}))
        |> drop(columns: ["_field_P_Class_k", "_value_P_Class_k"])
        |> rename(columns: {_field_P_k_x: "_field", _value_P_k_x: "_value"})
    P_k_x_class_Drop = join(tables: {P_k_x: P_k_x, P_Class_k: P_Class_k}, on: [myClass], method: "inner")
        |> drop(columns: ["_field_P_Class_k", "_value_P_Class_k", "_field_P_k_x"])
        |> group(columns: [myClass, "_value_P_k_x"])
        |> limit(n: 1)
        |> map(fn: (r) => ({r with P_x_k: r.sum / float(v: r._value_P_Class_k)}))

    //added P(value_x) to table
    //calculated probabilities for training data 
    Probability_table = join(tables: {P_k_x_class: P_k_x_class, P_value_x: P_value_x}, on: ["_value", "_field"], method: "inner")
        |> map(fn: (r) => ({r with Probability: r.P_x_k * r.p_k / r.p_x}))

    //|> yield(name: "final")
    //predictions for test data computed 
    predictOverall = (tables=<-) => {
        r = tables
            |> keep(columns: ["_value", "Animal_name", "_field"])
        output = join(tables: {Probability_table: Probability_table, r: r}, on: ["_value"], method: "inner")

        return output
    }

    return test_data |> predictOverall()
}

// builtin constants
builtin pi : float
builtin e : float
builtin phi : float
builtin sqrt2 : float
builtin sqrte : float
builtin sqrtpi : float
builtin sqrtphi : float
builtin ln2 : float
builtin log2e : float
builtin ln10 : float
builtin log10e : float
builtin maxfloat : float
builtin smallestNonzeroFloat : float
builtin maxint : int
builtin minint : int
builtin maxuint : uint

// builtin functions
builtin abs : (x: float) => float
builtin acos : (x: float) => float
builtin acosh : (x: float) => float
builtin asin : (x: float) => float
builtin asinh : (x: float) => float
builtin atan : (x: float) => float
builtin atan2 : (x: float, y: float) => float
builtin atanh : (x: float) => float
builtin cbrt : (x: float) => float
builtin ceil : (x: float) => float
builtin copysign : (x: float, y: float) => float
builtin cos : (x: float) => float
builtin cosh : (x: float) => float
builtin dim : (x: float, y: float) => float
builtin erf : (x: float) => float
builtin erfc : (x: float) => float
builtin erfcinv : (x: float) => float
builtin erfinv : (x: float) => float
builtin exp : (x: float) => float
builtin exp2 : (x: float) => float
builtin expm1 : (x: float) => float
builtin float64bits : (f: float) => uint
builtin float64frombits : (b: uint) => float
builtin floor : (x: float) => float
builtin frexp : (f: float) => {frac: float, exp: int}
builtin gamma : (x: float) => float
builtin hypot : (x: float) => float
builtin ilogb : (x: float) => float
builtin mInf : (sign: int) => float
builtin isInf : (f: float, sign: int) => bool
builtin isNaN : (f: float) => bool
builtin j0 : (x: float) => float
builtin j1 : (x: float) => float
builtin jn : (n: int, x: float) => float
builtin ldexp : (frac: float, exp: int) => float
builtin lgamma : (x: float) => {lgamma: float, sign: int}
builtin log : (x: float) => float
builtin log10 : (x: float) => float
builtin log1p : (x: float) => float
builtin log2 : (x: float) => float
builtin logb : (x: float) => float
builtin mMax : (x: float, y: float) => float
builtin mMin : (x: float, y: float) => float
builtin mod : (x: float, y: float) => float
builtin modf : (f: float) => {int: float, frac: float}
builtin NaN : () => float
builtin nextafter : (x: float, y: float) => float
builtin pow : (x: float, y: float) => float
builtin pow10 : (n: int) => float
builtin remainder : (x: float, y: float) => float
builtin round : (x: float) => float
builtin roundtoeven : (x: float) => float
builtin signbit : (x: float) => bool
builtin sin : (x: float) => float
builtin sincos : (x: float) => {sin: float, cos: float}
builtin sinh : (x: float) => float
builtin sqrt : (x: float) => float
builtin tan : (x: float) => float
builtin tanh : (x: float) => float
builtin trunc : (x: float) => float
builtin y0 : (x: float) => float
builtin y1 : (x: float) => float
builtin yn : (n: int, x: float) => float
builtin assertEquals : (name: string, <-got: [A], want: [A]) => [A]
builtin assertEmpty : (<-tables: [A]) => [A]
builtin diff : (<-got: [A], want: [A], ?verbose: bool, ?epsilon: float) => [{A with _diff: string}]

option loadStorage = (csv) => c.from(csv: csv)
    |> range(start: 1800-01-01T00:00:00Z, stop: 2200-12-31T11:59:59Z)
    |> map(
        fn: (r) => ({r with
            _field: if exists r._field then r._field else die(msg: "test input table does not have _field column"),
            _measurement: if exists r._measurement then r._measurement else die(msg: "test input table does not have _measurement column"),
            _time: if exists r._time then r._time else die(msg: "test input table does not have _time column"),
        }),
    )
option loadMem = (csv) => c.from(csv: csv)

inspect = (case) => {
    tc = case()
    got = tc.input |> tc.fn()
    dif = got |> diff(want: tc.want)

    return {
        fn: tc.fn,
        input: tc.input,
        want: tc.want |> yield(name: "want"),
        got: got |> yield(name: "got"),
        diff: dif |> yield(name: "diff"),
    }
}
run = (case) => {
    return inspect(case: case).diff |> assertEmpty()
}
benchmark = (case) => {
    tc = case()

    return tc.input |> tc.fn()
}

builtin time : () => time

// encode converts a value into JSON bytes
// Time values are encoded using RFC3339.
// Duration values are encoded in number of milleseconds since the epoch.
// Regexp values are encoded as their string representation.
// Bytes values are encodes as base64-encoded strings.
// Function values cannot be encoded and will produce an error.
builtin encode : (v: A) => bytes

// Post submits an HTTP post request to the specified URL with headers and data.
// The HTTP status code is returned.
builtin post : (url: string, ?headers: A, ?data: bytes) => int where A: Record

// basicAuth will take a username/password combination and return the authorization
// header value.
builtin basicAuth : (u: string, p: string) => string

// PathEscape escapes the string so it can be safely placed inside a URL path segment
// replacing special characters (including /) with %XX sequences as needed.
builtin pathEscape : (inputString: string) => string

endpoint = (url) => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(v: 200 == post(url: url, headers: obj.headers, data: obj.data)),
            }
        },
    )
    |> experimental.group(mode: "extend", columns: ["_sent"])

builtin to : (
    <-tables: [A],
    brokers: [string],
    topic: string,
    ?balancer: string,
    ?name: string,
    ?nameColumn: string,
    ?timeColumn: string,
    ?tagColumns: [string],
    ?valueColumns: [string],
) => [A] where
    A: Record
builtin from : (?csv: string, ?file: string) => [A] where A: Record
builtin linear : (
    <-tables: [{T with
        _time: time,
        _value: float,
    }],
    every: duration,
) => [{T with
    _time: time,
    _value: float,
}]
builtin from : (start: A, stop: A, count: int, fn: (n: int) => int) => [{_start: time, _stop: time, _time: time, _value: int}] where A: Timeable

option defaultURL = "https://api.pushbullet.com/v2/pushes"

// `pushData` sends a push notification using PushBullet's APIs.
// `url` - string - URL of the PushBullet endpoint. Defaults to: "https://api.pushbullet.com/v2/pushes".
// `token` - string - the api token string.  Defaults to: "".
// `data` - object - The data to send to the endpoint. It will be encoded in JSON and sent to PushBullet's endpoint.
// For how to structure data, see https://docs.pushbullet.com/#create-push.
pushData = (url=defaultURL, token="", data) => {
    headers = {
        "Access-Token": token,
        "Content-Type": "application/json",
    }
    enc = json.encode(v: data)

    return http.post(headers: headers, url: url, data: enc)
}

// `pushNote` sends a push notification of type `note` using PushBullet's APIs.
// `url` - string - URL of the PushBullet endpoint. Defaults to: "https://api.pushbullet.com/v2/pushes".
// `token` - string - the api token string.  Defaults to: "".
// `title` - string - The title of the notification.
// `text` - string - The text to display in the notification.
pushNote = (url=defaultURL, token="", title, text) => {
    data = {
        type: "note",
        title: title,
        body: text,
    }

    return pushData(token: token, url: url, data: data)
}

// `genericEndpoint` does not work for now for a bug in type inference in the compiler.
// // `genericEndpoint` creates the endpoint for the PushBullet external service.
// // `url` - string - URL of the PushBullet endpoint. Defaults to: "https://api.pushbullet.com/v2/pushes".
// // `token` - string - token for the PushBullet endpoint.
// // The returned factory function accepts a `mapFn` parameter.
// // The `mapFn` must return an object that will be used as payload as defined in `pushData` function arguments.
// genericEndpoint = (url=defaultURL, token="") =>
//     (mapFn) =>
//         (tables=<-) => tables
//             |> map(fn: (r) => {
//                 obj = mapFn(r: r)
//                 return {r with _sent: string(v: 2 == pushData(
//                   url: url,
//                   token: token,
//                   data: obj,
//                 ) / 100)}
//             })
// `endpoint` creates the endpoint for the PushBullet external service.
// It will push notifications of type `note`.
// If you want to push something else, see `genericEndpoint`.
// `url` - string - URL of the PushBullet endpoint. Defaults to: "https://api.pushbullet.com/v2/pushes".
// `token` - string - token for the PushBullet endpoint.
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `title` and `text` fields as defined in the `pushNote` function arguments.
endpoint = (url=defaultURL, token="") => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(
                    v: 2 == pushNote(
                        url: url,
                        token: token,
                        title: obj.title,
                        text: obj.text,
                    ) / 100,
                ),
            }
        },
    )

builtin validateColorString : (color: string) => string

option defaultURL = "https://slack.com/api/chat.postMessage"

// `message` sends a single message to a Slack channel. It will work either with the chat.postMessage API or with a slack webhook.
// `url` - string - URL of the slack endpoint. Defaults to: "https://slack.com/api/chat.postMessage", if one uses the webhook api this must be acquired as part of the slack API setup. This URL will be secret. Don't worry about secrets for the initial implementation.
// `token` - string - the api token string.  Defaults to: "", and can be ignored if one uses the webhook api URL.
// `channel` - string - Name of channel in which to post the message. No default.
// `text` - string - The text to display.
// `color` - string - Color to give message: one of good, warning, and danger, or any hex rgb color value ex. #439FE0.
message = (url=defaultURL, token="", channel, text, color) => {
    attachments = [
        {color: validateColorString(color), text: string(v: text), mrkdwn_in: ["text"]},
    ]
    data = {
        channel: channel,
        attachments: attachments,
        as_user: false,
    }
    headers = {
        "Authorization": "Bearer " + token,
        "Content-Type": "application/json",
    }
    enc = json.encode(v: data)

    return http.post(headers: headers, url: url, data: enc)
}

// `endpoint` creates the endpoint for the Slack external service.
// `url` - string - URL of the slack endpoint. Defaults to: "https://slack.com/api/chat.postMessage", if one uses the webhook api this must be acquired as part of the slack API setup, and this URL will be secret.
// `token` - string - token for the slack endpoint.  This can be ignored if one uses the webhook url acquired as part of the slack API setup, but must be supplied if the chat.postMessage API is used.
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `channel`, `text`, and `color` fields as defined in the `message` function arguments.
endpoint = (url=defaultURL, token="") => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(
                    v: 2 == message(
                        url: url,
                        token: token,
                        channel: obj.channel,
                        text: obj.text,
                        color: obj.color,
                    ) / 100,
                ),
            }
        },
    )

option enabledProfilers = [""]

// now is a function option whose default behaviour is to return the current system time
option now = system.time

// Booleans
builtin true : bool
builtin false : bool

// Transformation functions
builtin chandeMomentumOscillator : (<-tables: [A], n: int, ?columns: [string]) => [B] where A: Record, B: Record
builtin columns : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin count : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin covariance : (<-tables: [A], ?pearsonr: bool, ?valueDst: string, columns: [string]) => [B] where A: Record, B: Record
builtin cumulativeSum : (<-tables: [A], ?columns: [string]) => [B] where A: Record, B: Record
builtin derivative : (
    <-tables: [A],
    ?unit: duration,
    ?nonNegative: bool,
    ?columns: [string],
    ?timeColumn: string,
) => [B] where
    A: Record,
    B: Record
builtin die : (msg: string) => A
builtin difference : (<-tables: [T], ?nonNegative: bool, ?columns: [string], ?keepFirst: bool) => [R] where T: Record, R: Record
builtin distinct : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin drop : (<-tables: [A], ?fn: (column: string) => bool, ?columns: [string]) => [B] where A: Record, B: Record
builtin duplicate : (<-tables: [A], column: string, as: string) => [B] where A: Record, B: Record
builtin elapsed : (<-tables: [A], ?unit: duration, ?timeColumn: string, ?columnName: string) => [B] where A: Record, B: Record
builtin exponentialMovingAverage : (<-tables: [{B with _value: A}], n: int) => [{B with _value: A}] where A: Numeric
builtin fill : (<-tables: [A], ?column: string, ?value: B, ?usePrevious: bool) => [C] where A: Record, C: Record
builtin filter : (<-tables: [A], fn: (r: A) => bool, ?onEmpty: string) => [A] where A: Record
builtin first : (<-tables: [A], ?column: string) => [A] where A: Record
builtin group : (<-tables: [A], ?mode: string, ?columns: [string]) => [A] where A: Record
builtin histogram : (
    <-tables: [A],
    ?column: string,
    ?upperBoundColumn: string,
    ?countColumn: string,
    bins: [float],
    ?normalize: bool,
) => [B] where
    A: Record,
    B: Record
builtin histogramQuantile : (
    <-tables: [A],
    ?quantile: float,
    ?countColumn: string,
    ?upperBoundColumn: string,
    ?valueColumn: string,
    ?minValue: float,
) => [B] where
    A: Record,
    B: Record
builtin holtWinters : (
    <-tables: [A],
    n: int,
    interval: duration,
    ?withFit: bool,
    ?column: string,
    ?timeColumn: string,
    ?seasonality: int,
) => [B] where
    A: Record,
    B: Record
builtin hourSelection : (<-tables: [A], start: int, stop: int, ?timeColumn: string) => [A] where A: Record
builtin integral : (
    <-tables: [A],
    ?unit: duration,
    ?timeColumn: string,
    ?column: string,
    ?interpolate: string,
) => [B] where
    A: Record,
    B: Record
builtin join : (<-tables: A, ?method: string, ?on: [string]) => [B] where A: Record, B: Record
builtin kaufmansAMA : (<-tables: [A], n: int, ?column: string) => [B] where A: Record, B: Record
builtin keep : (<-tables: [A], ?columns: [string], ?fn: (column: string) => bool) => [B] where A: Record, B: Record
builtin keyValues : (<-tables: [A], ?keyColumns: [string]) => [{C with _key: string, _value: B}] where A: Record, C: Record
builtin keys : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin last : (<-tables: [A], ?column: string) => [A] where A: Record
builtin limit : (<-tables: [A], n: int, ?offset: int) => [A]
builtin map : (<-tables: [A], fn: (r: A) => B, ?mergeKey: bool) => [B]
builtin max : (<-tables: [A], ?column: string) => [A] where A: Record
builtin mean : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin min : (<-tables: [A], ?column: string) => [A] where A: Record
builtin mode : (<-tables: [A], ?column: string) => [{C with _value: B}] where A: Record, C: Record
builtin movingAverage : (<-tables: [{B with _value: A}], n: int) => [{B with _value: float}] where A: Numeric
builtin quantile : (
    <-tables: [A],
    ?column: string,
    q: float,
    ?compression: float,
    ?method: string,
) => [A] where
    A: Record
builtin pivot : (<-tables: [A], rowKey: [string], columnKey: [string], valueColumn: string) => [B] where A: Record, B: Record
builtin range : (
    <-tables: [{A with _time: time}],
    start: B,
    ?stop: C,
) => [{A with
    _time: time,
    _start: time,
    _stop: time,
}]
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
builtin stateTracking : (
    <-tables: [A],
    fn: (r: A) => bool,
    ?countColumn: string,
    ?durationColumn: string,
    ?durationUnit: duration,
    ?timeColumn: string,
) => [B] where
    A: Record,
    B: Record
builtin stddev : (<-tables: [A], ?column: string, ?mode: string) => [B] where A: Record, B: Record
builtin sum : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin tripleExponentialDerivative : (<-tables: [{B with _value: A}], n: int) => [{B with _value: float}] where A: Numeric, B: Record
builtin union : (tables: [[A]]) => [A] where A: Record
builtin unique : (<-tables: [A], ?column: string) => [A] where A: Record
builtin window : (
    <-tables: [A],
    ?every: duration,
    ?period: duration,
    ?offset: duration,
    ?timeColumn: string,
    ?startColumn: string,
    ?stopColumn: string,
    ?createEmpty: bool,
) => [B] where
    A: Record,
    B: Record
builtin yield : (<-tables: [A], ?name: string) => [A] where A: Record

// stream/table index functions
builtin tableFind : (<-tables: [A], fn: (key: B) => bool) => [A] where A: Record, B: Record
builtin getColumn : (<-table: [A], column: string) => [B] where A: Record
builtin getRecord : (<-table: [A], idx: int) => A where A: Record
builtin findColumn : (<-tables: [A], fn: (key: B) => bool, column: string) => [C] where A: Record, B: Record
builtin findRecord : (<-tables: [A], fn: (key: B) => bool, idx: int) => A where A: Record, B: Record

// type conversion functions
builtin bool : (v: A) => bool
builtin bytes : (v: A) => bytes
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

// die returns a fatal error from within a flux script
builtin die : (msg: string) => A

// Time weighted average where values at the beginning and end of the range are linearly interpolated.
timeWeightedAvg = (tables=<-, unit) => tables
    |> integral(unit: unit, interpolate: "linear")
    |> map(fn: (r) => ({r with _value: r._value * float(v: uint(v: unit)) / float(v: int(v: r._stop) - int(v: r._start))}))

// covariance function with automatic join
cov = (x, y, on, pearsonr=false) => join(
    tables: {x: x, y: y},
    on: on,
)
    |> covariance(pearsonr: pearsonr, columns: ["_value_x", "_value_y"])
pearsonr = (x, y, on) => cov(x: x, y: y, on: on, pearsonr: true)

// AggregateWindow applies an aggregate function to fixed windows of time.
// The procedure is to window the data, perform an aggregate operation,
// and then undo the windowing to produce an output table for every input table.
aggregateWindow = (every, fn, column="_value", timeSrc="_stop", timeDst="_time", createEmpty=true, tables=<-) => tables
    |> window(every: every, createEmpty: createEmpty)
    |> fn(column: column)
    |> duplicate(column: timeSrc, as: timeDst)
    |> window(every: inf, timeColumn: timeDst)

// Increase returns the total non-negative difference between values in a table.
// A main usage case is tracking changes in counter values which may wrap over time when they hit
// a threshold or are reset. In the case of a wrap/reset,
// we can assume that the absolute delta between two points will be at least their non-negative difference.
increase = (tables=<-, columns=["_value"]) => tables
    |> difference(nonNegative: true, columns: columns)
    |> cumulativeSum(columns: columns)

// median returns the 50th percentile.
median = (method="estimate_tdigest", compression=0.0, column="_value", tables=<-) => tables
    |> quantile(q: 0.5, method: method, compression: compression, column: column)

// stateCount computes the number of consecutive records in a given state.
// The state is defined via the function fn. For each consecutive point for
// which the expression evaluates as true, the state count will be incremented
// When a point evaluates as false, the state count is reset.
//
// The state count will be added as an additional column to each record. If the
// expression evaluates as false, the value will be -1. If the expression
// generates an error during evaluation, the point is discarded, and does not
// affect the state count.
stateCount = (fn, column="stateCount", tables=<-) => tables
    |> stateTracking(countColumn: column, fn: fn)

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
stateDuration = (fn, column="stateDuration", timeColumn="_time", unit=1s, tables=<-) => tables
    |> stateTracking(durationColumn: column, timeColumn: timeColumn, fn: fn, durationUnit: unit)

// _sortLimit is a helper function, which sorts and limits a table.
_sortLimit = (n, desc, columns=["_value"], tables=<-) => tables
    |> sort(columns: columns, desc: desc)
    |> limit(n: n)

// top sorts a table by columns and keeps only the top n records.
top = (n, columns=["_value"], tables=<-) => tables
    |> _sortLimit(n: n, columns: columns, desc: true)

// top sorts a table by columns and keeps only the bottom n records.
bottom = (n, columns=["_value"], tables=<-) => tables
    |> _sortLimit(n: n, columns: columns, desc: false)

// _highestOrLowest is a helper function, which reduces all groups into a single group by specific tags and a reducer function,
// then it selects the highest or lowest records based on the column and the _sortLimit function.
// The default reducer assumes no reducing needs to be performed.
_highestOrLowest = (n, _sortLimit, reducer, column="_value", groupColumns=[], tables=<-) => tables
    |> group(columns: groupColumns)
    |> reducer()
    |> group(columns: [])
    |> _sortLimit(n: n, columns: [column])

// highestMax returns the top N records from all groups using the maximum of each group.
highestMax = (n, column="_value", groupColumns=[], tables=<-) => tables
    |> _highestOrLowest(
        n: n,
        column: column,
        groupColumns: groupColumns,
        // TODO(nathanielc): Once max/min support selecting based on multiple columns change this to pass all columns.
        reducer: (tables=<-) => tables |> max(column: column),
        _sortLimit: top,
    )

// highestAverage returns the top N records from all groups using the average of each group.
highestAverage = (n, column="_value", groupColumns=[], tables=<-) => tables
    |> _highestOrLowest(
        n: n,
        column: column,
        groupColumns: groupColumns,
        reducer: (tables=<-) => tables |> mean(column: column),
        _sortLimit: top,
    )

// highestCurrent returns the top N records from all groups using the last value of each group.
highestCurrent = (n, column="_value", groupColumns=[], tables=<-) => tables
    |> _highestOrLowest(
        n: n,
        column: column,
        groupColumns: groupColumns,
        reducer: (tables=<-) => tables |> last(column: column),
        _sortLimit: top,
    )

// lowestMin returns the bottom N records from all groups using the minimum of each group.
lowestMin = (n, column="_value", groupColumns=[], tables=<-) => tables
    |> _highestOrLowest(
        n: n,
        column: column,
        groupColumns: groupColumns,
        // TODO(nathanielc): Once max/min support selecting based on multiple columns change this to pass all columns.
        reducer: (tables=<-) => tables |> min(column: column),
        _sortLimit: bottom,
    )

// lowestAverage returns the bottom N records from all groups using the average of each group.
lowestAverage = (n, column="_value", groupColumns=[], tables=<-) => tables
    |> _highestOrLowest(
        n: n,
        column: column,
        groupColumns: groupColumns,
        reducer: (tables=<-) => tables |> mean(column: column),
        _sortLimit: bottom,
    )

// lowestCurrent returns the bottom N records from all groups using the last value of each group.
lowestCurrent = (n, column="_value", groupColumns=[], tables=<-) => tables
    |> _highestOrLowest(
        n: n,
        column: column,
        groupColumns: groupColumns,
        reducer: (tables=<-) => tables |> last(column: column),
        _sortLimit: bottom,
    )

// timedMovingAverage constructs a simple moving average over windows of 'period' duration
// eg: A 5 year moving average would be called as such:
//    movingAverage(1y, 5y)
timedMovingAverage = (every, period, column="_value", tables=<-) => tables
    |> window(every: every, period: period)
    |> mean(column: column)
    |> duplicate(column: "_stop", as: "_time")
    |> window(every: inf)

// Double Exponential Moving Average computes the double exponential moving averages of the `_value` column.
// eg: A 5 point double exponential moving average would be called as such:
// from(bucket: "telegraf/autogen"):
//    |> range(start: -7d)
//    |> doubleEMA(n: 5)
doubleEMA = (n, tables=<-) => tables
    |> exponentialMovingAverage(n: n)
    |> duplicate(column: "_value", as: "__ema")
    |> exponentialMovingAverage(n: n)
    |> map(fn: (r) => ({r with _value: 2.0 * r.__ema - r._value}))
    |> drop(columns: ["__ema"])

// Triple Exponential Moving Average computes the triple exponential moving averages of the `_value` column.
// eg: A 5 point triple exponential moving average would be called as such:
// from(bucket: "telegraf/autogen"):
//    |> range(start: -7d)
//    |> tripleEMA(n: 5)
tripleEMA = (n, tables=<-) => tables
    |> exponentialMovingAverage(n: n)
    |> duplicate(column: "_value", as: "__ema1")
    |> exponentialMovingAverage(n: n)
    |> duplicate(column: "_value", as: "__ema2")
    |> exponentialMovingAverage(n: n)
    |> map(fn: (r) => ({r with _value: 3.0 * r.__ema1 - 3.0 * r.__ema2 + r._value}))
    |> drop(columns: ["__ema1", "__ema2"])

// truncateTimeColumn takes in a time column t and a Duration unit and truncates each value of t to the given unit via map
// Change from _time to timeColumn once Flux Issue 1122 is resolved
truncateTimeColumn = (timeColumn="_time", unit, tables=<-) => tables
    |> map(fn: (r) => ({r with _time: date.truncate(t: r._time, unit: unit)}))

// kaufmansER computes Kaufman's Efficiency Ratios of the `_value` column
kaufmansER = (n, tables=<-) => tables
    |> chandeMomentumOscillator(n: n)
    |> map(fn: (r) => ({r with _value: math.abs(x: r._value) / 100.0}))
toString = (tables=<-) => tables |> map(fn: (r) => ({r with _value: string(v: r._value)}))
toInt = (tables=<-) => tables |> map(fn: (r) => ({r with _value: int(v: r._value)}))
toUInt = (tables=<-) => tables |> map(fn: (r) => ({r with _value: uint(v: r._value)}))
toFloat = (tables=<-) => tables |> map(fn: (r) => ({r with _value: float(v: r._value)}))
toBool = (tables=<-) => tables |> map(fn: (r) => ({r with _value: bool(v: r._value)}))
toTime = (tables=<-) => tables |> map(fn: (r) => ({r with _value: time(v: r._value)}))

builtin from : (url: string, ?decoder: string) => [A]
builtin from : (driverName: string, dataSourceName: string, query: string) => [A]
builtin to : (
    <-tables: [A],
    driverName: string,
    dataSourceName: string,
    table: string,
    ?batchSize: int,
) => [A]
