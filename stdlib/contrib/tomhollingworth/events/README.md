# Events Package

Use this Flux Package calculate the time between a record and the next record. The function `events.duration` peeks at the next record and calculates the duration an between records and associates it with the start of the event. For the final record it can be compared against a stop column or a timestamp. This function differs to existing `elapsed` which removes the first entry and `stateDuration` which totalized on a function.

See also
- [elapsed](https://docs.influxdata.com/influxdb/v2.0/reference/flux/stdlib/built-in/transformations/elapsed/)
- [stateDuration](https://docs.influxdata.com/influxdb/v2.0/reference/flux/stdlib/built-in/transformations/stateduration/)

## events.duration

`duration` calculates the duration of the event.

| Name        | Type     | Description                                                                  |
| ----------- | -------- | ---------------------------------------------------------------------------- |
| unit        | duration | Units of state duration 'ns', 'us', 'µs', 'ms', 's', 'm', 'h'                |
| columnName  | string   | The name of the result column. Default `duration`                            |
| timeColumn  | string   | The name of the time column, default `_time`                                 |
| stopColumn  | string   | The name of the stop column, default `_stop`                                 |
| stop        | time     | Optional. If provided, it will be used instead of the stop column  |

Basic Example:

```flux
import "contrib/tomhollingworth/events"

from(bucket: "example-bucket")
    |> range(start: -24h)
    |> events.duration()
```

### Last Record Duration

The last record needs a time to compare to. The following strategy is implemented:
- Use `stop` if provided.
- If no `stop` time is provided, then use the value from the `stopColumn` column on the last record.
- If no `stopColumn` is provided then use `_stop` by default.

### Comparison to other functions

Consider the following dataset of a door opening and closing: 

```flux
import "csv"

inData = "
#datatype,string,long,dateTime:RFC3339,string,string
#group,false,true,false,false,false
#default,,,,,
,result,table,_time,_value,_field
,,0,2020-01-01T08:00:00Z,Closed,value
,,0,2020-01-01T08:15:00Z,Open,value
,,0,2020-01-01T08:15:08Z,Closed,value
,,0,2020-01-01T08:21:00Z,Open,value
,,0,2020-01-01T08:21:07Z,Closed,value
,,0,2020-01-01T08:24:00Z,Open,value
,,0,2020-01-01T08:24:12Z,Closed,value
"

csv.from(csv: inData)
  |> range(start: 2020-01-01T08:00:00Z, stop: 2020-01-01T08:30:00Z)
```

`|> elapsed()` yields the following. The first record is dropped and durations are associated with subsequent records. Totalizing on filter on value and summing the elasped column would have the duration swapped between open and closed.

```diff
-,  result, table,               _start,                _stop, _time, _value, _field
+,  result, table,               _start,                _stop, _time, _value, _field,       elasped
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:00:00Z, Closed,  value
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:10:00Z, Closed,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:10:00Z, Closed,  value,          1600
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:15:00Z,   Open,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:15:00Z,   Open,  value,          1300
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:15:08Z, Closed,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:15:08Z, Closed,  value,             8
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:21:00Z,   Open,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:21:00Z,   Open,  value,          1352
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:21:07Z, Closed,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:21:07Z, Closed,  value,             7
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:24:00Z,   Open,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:24:00Z,   Open,  value,           173
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:24:12Z, Closed,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:24:12Z, Closed,  value,            12
```

`|> stateDuration(fn: (r) => true)` yields the following. The duration is continuously totalized. For a particular state we could also include that in our stateDuration function, however time is counted twice on subsequent events. Totalizing this would have that time counted twice.

```diff
-,  result, table,               _start,                _stop,                _time, _value, _field
+,  result, table,               _start,                _stop,                _time, _value, _field, stateDuration
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:00:00Z, Closed,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:00:00Z, Closed,  value,             0
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:10:00Z, Closed,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:10:00Z, Closed,  value,           600
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:15:00Z,   Open,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:15:00Z,   Open,  value,           900
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:15:08Z, Closed,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:15:08Z, Closed,  value,           908
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:21:00Z,   Open,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:21:00Z,   Open,  value,          1260
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:21:07Z, Closed,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:21:07Z, Closed,  value,          1267
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:24:00Z,   Open,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:24:00Z,   Open,  value,          1440
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:24:12Z, Closed,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:24:12Z, Closed,  value,          1452
```

`|> events.duration()` yields the following. 
```diff
-,  result, table,               _start,                _stop,                _time, _value, _field
+,  result, table,               _start,                _stop,                _time, _value, _field,      duration
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:00:00Z, Closed,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:00:00Z, Closed,  value,           600
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:10:00Z, Closed,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:10:00Z, Closed,  value,           300
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:15:00Z,   Open,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:15:00Z,   Open,  value,             8
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:15:08Z, Closed,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:15:08Z, Closed,  value,           352
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:21:00Z,   Open,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:21:00Z,   Open,  value,             7
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:21:07Z, Closed,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:21:07Z, Closed,  value,           173
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:24:00Z,   Open,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:24:00Z,   Open,  value,            12
-,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:24:12Z, Closed,  value
+,        ,     0, 2020-01-01T08:00:00Z, 2020-01-01T08:30:00Z, 2020-01-01T08:24:12Z, Closed,  value,           348
```

## Contact

- Author: Tom Hollingworth
- Email: tom.hollingworth@spruiktec.com
- Github: [@tomhollingworth](https://github.com/tomhollingworth)
- Influx Slack: [@tomhollingworth](https://influxdata.com/slack)
