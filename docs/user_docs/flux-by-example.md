# Find Unique Values for a Column
Drop all the data except the column you are interested in, and use unique().
```flux
import "experimental/csv"
csv.from(url: "https://influx-testdata.s3.amazonaws.com/noaa.csv")
|> keep(columns: ["location"])
|> unique(column: "location")
```

|#group   |false  |false|true        |
|---------|-------|-----|------------|
|#datatype|string |long |string      |
|#default |_result|     |            |
|         |result |table|location    |
|         |       |0    |coyote_creek|
|         |       |1    |santa_monica|

# Count the Number of Unique Values for a Column
Drop all the data except the column you are interested in, use unique(), remove the grouping, and then count().

```flux
import "experimental/csv"
csv.from(url: "https://influx-testdata.s3.amazonaws.com/noaa.csv")
|> keep(columns: ["location"])
|> unique(column: "location")
|> group(columns: [])
|> count(column: "location")
```

|#group   |false  |false|false       |
|---------|-------|-----|------------|
|#datatype|string |long |long        |
|#default |_result|     |            |
|         |result |table|location    |
|         |       |0    |2           |

# Calculate a New Column Based On Values in a Row
Use the with keyboard in a map function to create the new column.

```
import "experimental/csv"
csv.from(url: "https://influx-testdata.s3.amazonaws.com/noaa.csv")
|> filter(fn: (r) => r._measurement == "average_temperature")
|> map(fn: (r) => ({r with celsius: ((r._value - 32.0) * 5.0 / 9.0)} ))
```
|#group   |false  |false|true        |true               |true                          |true                          |false               |false |false             |true        |
|---------|-------|-----|------------|-------------------|------------------------------|------------------------------|--------------------|------|------------------|------------|
|#datatype|string |long |string      |string             |dateTime:RFC3339              |dateTime:RFC3339              |dateTime:RFC3339    |double|double            |string      |
|#default |_result|     |            |                   |                              |                              |                    |      |                  |            |
|         |result |table|_field      |_measurement       |_start                        |_stop                         |_time               |_value|celsius           |location    |
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:00:00Z|82    |27.77777777777778 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:06:00Z|73    |22.77777777777778 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:12:00Z|86    |30                |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:18:00Z|89    |31.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:24:00Z|77    |25                |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:30:00Z|70    |21.11111111111111 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:36:00Z|84    |28.88888888888889 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:42:00Z|76    |24.444444444444443|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:48:00Z|85    |29.444444444444443|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:54:00Z|80    |26.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|
| ... |

# Recalculate the _value Column in Place
Use “with _value” in a map function.

```flux
import "experimental/csv"
csv.from(url: "https://influx-testdata.s3.amazonaws.com/noaa.csv")
|> filter(fn: (r) => r._measurement == "average_temperature")
|> map(fn: (r) => ({r with _value: ((r._value - 32.0) * 5.0 / 9.0)} ))
```

|#group   |false  |false|true        |true               |true                          |true                          |false               |false |true              |
|---------|-------|-----|------------|-------------------|------------------------------|------------------------------|--------------------|------|------------------|
|#datatype|string |long |string      |string             |dateTime:RFC3339              |dateTime:RFC3339              |dateTime:RFC3339    |double|string            |
|#default |_result|     |            |                   |                              |                              |                    |      |                  |
|         |result |table|_field      |_measurement       |_start                        |_stop                         |_time               |_value|location          |
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:00:00Z|27.77777777777778|coyote_creek      |
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:06:00Z|22.77777777777778|coyote_creek      |
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:12:00Z|30    |coyote_creek      |
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:18:00Z|31.666666666666668|coyote_creek      |
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:24:00Z|25    |coyote_creek      |
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:30:00Z|21.11111111111111|coyote_creek      |
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:36:00Z|28.88888888888889|coyote_creek      |
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:42:00Z|24.444444444444443|coyote_creek      |
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:48:00Z|29.444444444444443|coyote_creek      |
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:54:00Z|26.666666666666668|coyote_creek      |
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T01:00:00Z|21.11111111111111|coyote_creek      |
| ... |

# Calculate  a Weekly Mean and Add it to a New Bucket 
Useful for comparing values to a historical mean. Use window() to group by day, then mean(), then create a _time column, send to the new bucket.

```flux
import "experimental/csv"
csv.from(url: "https://influx-testdata.s3.amazonaws.com/noaa.csv")
|> filter(fn: (r) => r._measurement == "average_temperature")
|> range(start: 2019-09-01T11:24:00Z)
|> window(every: 1w)
|> mean()
|> rename(columns: {_stop: "_time"})
|> to(bucket: "weekly_means")
```

|#group   |false  |false|true        |true               |true                          |true                          |true                |false |
|---------|-------|-----|------------|-------------------|------------------------------|------------------------------|--------------------|------|
|#datatype|string |long |dateTime:RFC3339|dateTime:RFC3339   |string                        |string                        |string              |double|
|#default |to6    |     |            |                   |                              |                              |                    |      |
|         |result |table|_start      |_time              |_field                        |_measurement                  |location            |_value|
|         |       |0    |2019-09-01T11:24:00Z|2019-09-05T00:00:00Z|degrees                       |average_temperature           |coyote_creek        |80.31005917159763|
|         |       |1    |2019-09-01T11:24:00Z|2019-09-05T00:00:00Z|degrees                       |average_temperature           |santa_monica        |80.19952494061758|
|         |       |2    |2019-09-05T00:00:00Z|2019-09-12T00:00:00Z|degrees                       |average_temperature           |coyote_creek        |79.8422619047619|
|         |       |3    |2019-09-05T00:00:00Z|2019-09-12T00:00:00Z|degrees                       |average_temperature           |santa_monica        |80.01964285714286|
|         |       |4    |2019-09-12T00:00:00Z|2019-09-19T00:00:00Z|degrees                       |average_temperature           |coyote_creek        |79.82710622710623|
|         |       |5    |2019-09-12T00:00:00Z|2019-09-19T00:00:00Z|degrees                       |average_temperature           |santa_monica        |80.20451339915374|

# Compare the Last Measurement to a Mean Stored in Another Bucket
Useful for writing to a bucket and using as a threshold check. Get the last value in the means bucket, compare it to the last value in your main bucket, use join() to combine the results, and use map() to calculate the differences.

```flux
means = from(bucket: "weekly_means")
|> range(start: 2019-09-01T00:00:00Z)
|> last()
|> keep(columns: ["_value", "location"])
 
latest = from(bucket: "noaa")
|> range(start: 2019-09-01T00:00:00Z)
|> filter(fn: (r) => r._measurement == "average_temperature")
|> last()
|> keep(columns: ["_value", "location"])
 
join(tables: {mean: means, reading: latest}, on: ["location"])
|> map(fn: (r) => ({r with deviation: r._value_reading - r._value_mean}))
```

|#group   |false  |false|false       |false              |false                         |true                          |
|---------|-------|-----|------------|-------------------|------------------------------|------------------------------|
|#datatype|string |long |double      |double             |double                        |string                        |
|#default |_result|     |            |                   |                              |                              |
|         |result |table|_value_mean |_value_reading     |deviation                     |location                      |
|         |       |0    |79.82710622710623|89                 |9.172893772893772             |coyote_creek                  |
|         |       |1    |80.20451339915374|85                 |4.79548660084626              |santa_monica                  |

# Convert Results to JSON and Post to an Endpoint
Use json.encode() and http.post(). The follow will make a separate http call for each record.

```
import "http"
import "json"
import "experimental/csv"
csv.from(url: "https://influx-testdata.s3.amazonaws.com/noaa.csv")
|> filter(fn: (r) => r._measurement == "average_temperature")
|> mean()
|> map(fn: (r) => ({ r with jsonStr: string(v: json.encode(v: {"location":r.location,"mean":r._value}))}))
|> map(fn: (r) => ({r with status_code: http.post(url: "http://somehost.com/", headers: {x:"a", y:"b"}, data: bytes(v: r.jsonStr))}))
```
