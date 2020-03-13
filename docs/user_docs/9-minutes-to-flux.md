# Set up the Data Explorer
Open InfluxDB 2.0 Cloud, and use the left hand navigation to open the Data Explorer.

![data explorer left nav](images/image1.png?raw=true)

Click the Script Editor button to open the Flux editor.

![script editor button](images/image2.png?raw=true)

Use the switch to switch to the raw data view.
![raw data view](images/image3.png?raw=true)

Paste the following code into the script editor, and click submit.

```flux
import "experimental/csv"
csv.from(url: "https://influx-testdata.s3.amazonaws.com/noaa.csv")
```

![copy query](images/image4.png?raw=true)

Data is now loaded into the data explorer, and we can start transforming it.

![ready](images/image5.png?raw=true)


# The Flux Data Model
Stream of Tables
The data that we are working with is sensor data from various weather stations taken over a period of a few weeks. There are multiple measurements taken from multiple locations.

In order to understand the data, you must first understand the Flux data model. In Flux, you work with *streams* of tables. A stream of tables is a collection of tables where each table represents a logical grouping of data. What this means in practice, is that when you run a query, you get back a set of tables to work with, not just one giant table. 

If you scroll through the data, notice that there is a table column, and that the data is organized into separate tables.

Use the following query to take a look at just average temperatures for the last few hours of data:

```flux
import "experimental/csv"
csv.from(url: "https://influx-testdata.s3.amazonaws.com/noaa.csv")
 
|> range(start: 2019-09-17T16:00:00Z)
|> filter(fn: (r) => r._measurement == "average_temperature")
```

Scrolling through the data, you can see that the results are in a set of tables, one table each, the difference being that each table is from a different location.
Notice the table column and the location column.
Group Keys
When Flux returns a query, it creates separate tables. Each table has a set of columns where the value for those columns are the same for each row. For example, notice below that each table either has santa_monica or coyote_creek as the location.

|#group   |false  |false|true        |true               |false                         |false                         |true   |true               |true        |
|---------|-------|-----|------------|-------------------|------------------------------|------------------------------|-------|-------------------|------------|
|#datatype|string |long |dateTime:RFC3339|dateTime:RFC3339   |dateTime:RFC3339              |double                        |string |string             |string      |
|#default |_result|     |            |                   |                              |                              |       |                   |            |
|         |result |table|_start      |_stop              |_time                         |_value                        |_field |_measurement       |location    |
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:00:00Z          |89                            |degrees|average_temperature|coyote_creek|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:06:00Z          |86                            |degrees|average_temperature|coyote_creek|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:12:00Z          |70                            |degrees|average_temperature|coyote_creek|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:18:00Z          |70                            |degrees|average_temperature|coyote_creek|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:24:00Z          |89                            |degrees|average_temperature|coyote_creek|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:00:00Z          |90                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:06:00Z          |73                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:12:00Z          |82                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:18:00Z          |71                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:24:00Z          |82                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:30:00Z          |82                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:36:00Z          |88                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:42:00Z          |72                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:48:00Z          |79                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:54:00Z          |80                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:00:00Z          |71                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:06:00Z          |71                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:12:00Z          |81                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:18:00Z          |89                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:24:00Z          |71                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:30:00Z          |90                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:36:00Z          |89                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:42:00Z          |70                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:48:00Z          |78                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:54:00Z          |90                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:00:00Z          |84                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:06:00Z          |83                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:12:00Z          |88                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:18:00Z          |82                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:24:00Z          |83                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:30:00Z          |86                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:36:00Z          |86                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:42:00Z          |81                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:48:00Z          |82                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:54:00Z          |84                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:00:00Z          |74                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:06:00Z          |80                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:12:00Z          |77                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:18:00Z          |73                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:24:00Z          |86                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:30:00Z          |89                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:36:00Z          |83                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:42:00Z          |71                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:48:00Z          |83                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:54:00Z          |85                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:00:00Z          |74                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:06:00Z          |83                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:12:00Z          |79                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:18:00Z          |90                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:24:00Z          |79                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:30:00Z          |74                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:36:00Z          |75                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:42:00Z          |71                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:48:00Z          |79                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:54:00Z          |71                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:00:00Z          |79                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:06:00Z          |86                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:12:00Z          |73                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:18:00Z          |70                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:24:00Z          |79                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:30:00Z          |90                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:36:00Z          |83                            |degrees|average_temperature|santa_monica|
|         |       |1    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:42:00Z          |85                            |degrees|average_temperature|santa_monica|


The collection of columns that defines how a table is grouped is called the group key. When the group key changes in a stream, you can see that Flux inserts a new annotation with a group named “#group,” and labels each column that is part of the group key as true. If the group key does not change, you can see the new table by the third column labeled “table.” The table id changes. 

#Operating on Tables
In Flux, the tables are a *stream*, and the stream could theoretically be infinite. Therefore, it is important to keep in mind that you can’t operate on all the tables at once, but you can operate on each table in turn.

## Aggregate Functions Return a New Table For Each Table in the Stream
Aggregate functions are functions that collapse a table into fewer rows, often a single row. For example, the mean() function will return a new table with a single row, that is the mean for each table in the stream. 

```
import "experimental/csv"
csv.from(url: "https://influx-testdata.s3.amazonaws.com/noaa.csv")
 
|> range(start: 2019-09-02T12:44:00Z)
|> filter(fn: (r) => r._measurement == "average_temperature")
|> mean()
```

Now notice that there are still the same number of tables as before, but each table has a single row, where the _value column is the mean.

|#group   |false  |false|true        |true               |true                          |true                          |true   |false              |
|---------|-------|-----|------------|-------------------|------------------------------|------------------------------|-------|-------------------|
|#datatype|string |long |dateTime:RFC3339|dateTime:RFC3339   |string                        |string                        |string |double             |
|#default |_result|     |            |                   |                              |                              |       |                   |
|         |result |table|_start      |_stop              |_field                        |_measurement                  |location|_value             |
|         |       |0    |2019-09-02T12:44:00Z|2020-03-05T22:10:01.711964667Z|degrees                       |average_temperature           |coyote_creek|79.91006600660066  |
|         |       |1    |2019-09-02T12:44:00Z|2020-03-05T22:10:01.711964667Z|degrees                       |average_temperature           |santa_monica|80.08708627238198  |


## Changing Group Keys
Some functions may change the tables in the stream. For example, keep() and drop() go through each table and keep or drop columns. If you drop columns that are part of the group key, then the stream will be adjusted. 

For example, if we drop the location column:

```flux
import "experimental/csv"
csv.from(url: "https://influx-testdata.s3.amazonaws.com/noaa.csv")
 
|> range(start: 2019-09-17T16:00:00Z)
|> filter(fn: (r) => r._measurement == "average_temperature")
|> drop(columns: ["location"])
```

Notice that the tables column is different, because Flux is no longer grouping on that columns:

|#group   |false  |false|true        |true               |false                         |false                         |true   |true               |
|---------|-------|-----|------------|-------------------|------------------------------|------------------------------|-------|-------------------|
|#datatype|string |long |dateTime:RFC3339|dateTime:RFC3339   |dateTime:RFC3339              |double                        |string |string             |
|#default |_result|     |            |                   |                              |                              |       |                   |
|         |result |table|_start      |_stop              |_time                         |_value                        |_field |_measurement       |
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:00:00Z          |89                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:06:00Z          |86                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:12:00Z          |70                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:18:00Z          |70                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:24:00Z          |89                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:00:00Z          |90                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:06:00Z          |73                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:12:00Z          |82                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:18:00Z          |71                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:24:00Z          |82                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:30:00Z          |82                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:36:00Z          |88                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:42:00Z          |72                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:48:00Z          |79                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T16:54:00Z          |80                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:00:00Z          |71                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:06:00Z          |71                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:12:00Z          |81                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:18:00Z          |89                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:24:00Z          |71                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:30:00Z          |90                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:36:00Z          |89                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:42:00Z          |70                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:48:00Z          |78                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T17:54:00Z          |90                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:00:00Z          |84                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:06:00Z          |83                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:12:00Z          |88                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:18:00Z          |82                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:24:00Z          |83                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:30:00Z          |86                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:36:00Z          |86                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:42:00Z          |81                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:48:00Z          |82                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T18:54:00Z          |84                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:00:00Z          |74                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:06:00Z          |80                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:12:00Z          |77                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:18:00Z          |73                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:24:00Z          |86                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:30:00Z          |89                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:36:00Z          |83                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:42:00Z          |71                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:48:00Z          |83                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T19:54:00Z          |85                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:00:00Z          |74                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:06:00Z          |83                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:12:00Z          |79                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:18:00Z          |90                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:24:00Z          |79                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:30:00Z          |74                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:36:00Z          |75                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:42:00Z          |71                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:48:00Z          |79                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T20:54:00Z          |71                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:00:00Z          |79                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:06:00Z          |86                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:12:00Z          |73                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:18:00Z          |70                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:24:00Z          |79                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:30:00Z          |90                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:36:00Z          |83                            |degrees|average_temperature|
|         |       |0    |2019-09-17T16:00:00Z|2020-03-05T22:10:01.711964667Z|2019-09-17T21:42:00Z          |85                            |degrees|average_temperature|

# Operating on Rows
In addition to being able to operate on tables, with Flux you can also operate on each row. For example.
Execute a Function on Each Row
Using map() you can execute a function on each row. For example, we can add a new column with the temperatures converted to celsius:

```flux
import "experimental/csv"
csv.from(url: "https://influx-testdata.s3.amazonaws.com/noaa.csv")
 
|> filter(fn: (r) => r._measurement == "average_temperature")
|> map(fn: (r) => ({r with celsius: ((r._value - 32.0) * 5.0 / 9.0)} ))
```

|#group   |false  |false|true        |true               |true                          |true                          |false  |false              |false             |true        |
|---------|-------|-----|------------|-------------------|------------------------------|------------------------------|-------|-------------------|------------------|------------|
|#datatype|string |long |string      |string             |dateTime:RFC3339              |dateTime:RFC3339              |dateTime:RFC3339|double             |double            |string      |
|#default |_result|     |            |                   |                              |                              |       |                   |                  |            |
|         |result |table|_field      |_measurement       |_start                        |_stop                         |_time  |_value             |celsius           |location    |
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:00:00Z|82                 |27.77777777777778 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:06:00Z|73                 |22.77777777777778 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:12:00Z|86                 |30                |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:18:00Z|89                 |31.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:24:00Z|77                 |25                |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:30:00Z|70                 |21.11111111111111 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:36:00Z|84                 |28.88888888888889 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:42:00Z|76                 |24.444444444444443|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:48:00Z|85                 |29.444444444444443|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T00:54:00Z|80                 |26.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T01:00:00Z|70                 |21.11111111111111 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T01:06:00Z|77                 |25                |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T01:12:00Z|90                 |32.22222222222222 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T01:18:00Z|90                 |32.22222222222222 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T01:24:00Z|85                 |29.444444444444443|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T01:30:00Z|84                 |28.88888888888889 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T01:36:00Z|81                 |27.22222222222222 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T01:42:00Z|76                 |24.444444444444443|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T01:48:00Z|74                 |23.333333333333332|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T01:54:00Z|80                 |26.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T02:00:00Z|71                 |21.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T02:06:00Z|88                 |31.11111111111111 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T02:12:00Z|72                 |22.22222222222222 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T02:18:00Z|88                 |31.11111111111111 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T02:24:00Z|89                 |31.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T02:30:00Z|87                 |30.555555555555557|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T02:36:00Z|70                 |21.11111111111111 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T02:42:00Z|85                 |29.444444444444443|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T02:48:00Z|72                 |22.22222222222222 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T02:54:00Z|83                 |28.333333333333332|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T03:00:00Z|71                 |21.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T03:06:00Z|76                 |24.444444444444443|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T03:12:00Z|85                 |29.444444444444443|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T03:18:00Z|83                 |28.333333333333332|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T03:24:00Z|77                 |25                |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T03:30:00Z|78                 |25.555555555555557|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T03:36:00Z|80                 |26.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T03:42:00Z|90                 |32.22222222222222 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T03:48:00Z|77                 |25                |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T03:54:00Z|87                 |30.555555555555557|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T04:00:00Z|74                 |23.333333333333332|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T04:06:00Z|85                 |29.444444444444443|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T04:12:00Z|88                 |31.11111111111111 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T04:18:00Z|75                 |23.88888888888889 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T04:24:00Z|80                 |26.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T04:30:00Z|83                 |28.333333333333332|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T04:36:00Z|89                 |31.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T04:42:00Z|74                 |23.333333333333332|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T04:48:00Z|76                 |24.444444444444443|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T04:54:00Z|85                 |29.444444444444443|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T05:00:00Z|77                 |25                |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T05:06:00Z|85                 |29.444444444444443|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T05:12:00Z|90                 |32.22222222222222 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T05:18:00Z|72                 |22.22222222222222 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T05:24:00Z|71                 |21.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T05:30:00Z|87                 |30.555555555555557|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T05:36:00Z|89                 |31.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T05:42:00Z|76                 |24.444444444444443|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T05:48:00Z|82                 |27.77777777777778 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T05:54:00Z|79                 |26.11111111111111 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T06:00:00Z|72                 |22.22222222222222 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T06:06:00Z|84                 |28.88888888888889 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T06:12:00Z|87                 |30.555555555555557|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T06:18:00Z|73                 |22.77777777777778 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T06:24:00Z|77                 |25                |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T06:30:00Z|82                 |27.77777777777778 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T06:36:00Z|86                 |30                |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T06:42:00Z|83                 |28.333333333333332|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T06:48:00Z|72                 |22.22222222222222 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T06:54:00Z|81                 |27.22222222222222 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T07:00:00Z|78                 |25.555555555555557|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T07:06:00Z|79                 |26.11111111111111 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T07:12:00Z|82                 |27.77777777777778 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T07:18:00Z|87                 |30.555555555555557|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T07:24:00Z|84                 |28.88888888888889 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T07:30:00Z|77                 |25                |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T07:36:00Z|75                 |23.88888888888889 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T07:42:00Z|83                 |28.333333333333332|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T07:48:00Z|71                 |21.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T07:54:00Z|82                 |27.77777777777778 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T08:00:00Z|78                 |25.555555555555557|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T08:06:00Z|78                 |25.555555555555557|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T08:12:00Z|74                 |23.333333333333332|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T08:18:00Z|77                 |25                |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T08:24:00Z|71                 |21.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T08:30:00Z|72                 |22.22222222222222 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T08:36:00Z|72                 |22.22222222222222 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T08:42:00Z|82                 |27.77777777777778 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T08:48:00Z|88                 |31.11111111111111 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T08:54:00Z|70                 |21.11111111111111 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T09:00:00Z|73                 |22.77777777777778 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T09:06:00Z|79                 |26.11111111111111 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T09:12:00Z|81                 |27.22222222222222 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T09:18:00Z|87                 |30.555555555555557|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T09:24:00Z|80                 |26.666666666666668|coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T09:30:00Z|75                 |23.88888888888889 |coyote_creek|
|         |       |0    |degrees     |average_temperature|1920-03-05T22:10:01.711964667Z|2020-03-05T22:10:01.711964667Z|2019-08-17T09:36:00Z|85                 |29.444444444444443|coyote_creek|
| ...       |

# Windowing
It is often useful to create a stream of tables grouped by a specific time period, for example, by day:
```flux
import "experimental/csv"
csv.from(url: "https://influx-testdata.s3.amazonaws.com/noaa.csv")
 
|> range(start: 2019-01-01T00:00:00Z)
|> filter(fn: (r) => r._measurement == "average_temperature")
|> window(every: 1d)
```

|#group   |false  |false|true        |true               |false                         |false                         |true   |true               |true              |
|---------|-------|-----|------------|-------------------|------------------------------|------------------------------|-------|-------------------|------------------|
|#datatype|string |long |dateTime:RFC3339|dateTime:RFC3339   |dateTime:RFC3339              |double                        |string |string             |string            |
|#default |_result|     |            |                   |                              |                              |       |                   |                  |
|         |result |table|_start      |_stop              |_time                         |_value                        |_field |_measurement       |location          |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T00:00:00Z          |82                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T00:06:00Z          |73                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T00:12:00Z          |86                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T00:18:00Z          |89                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T00:24:00Z          |77                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T00:30:00Z          |70                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T00:36:00Z          |84                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T00:42:00Z          |76                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T00:48:00Z          |85                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T00:54:00Z          |80                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T01:00:00Z          |70                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T01:06:00Z          |77                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T01:12:00Z          |90                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T01:18:00Z          |90                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T01:24:00Z          |85                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T01:30:00Z          |84                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T01:36:00Z          |81                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T01:42:00Z          |76                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T01:48:00Z          |74                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T01:54:00Z          |80                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T02:00:00Z          |71                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T02:06:00Z          |88                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T02:12:00Z          |72                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T02:18:00Z          |88                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T02:24:00Z          |89                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T02:30:00Z          |87                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T02:36:00Z          |70                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T02:42:00Z          |85                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T02:48:00Z          |72                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T02:54:00Z          |83                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T03:00:00Z          |71                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T03:06:00Z          |76                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T03:12:00Z          |85                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T03:18:00Z          |83                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T03:24:00Z          |77                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T03:30:00Z          |78                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T03:36:00Z          |80                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T03:42:00Z          |90                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T03:48:00Z          |77                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T03:54:00Z          |87                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T04:00:00Z          |74                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T04:06:00Z          |85                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T04:12:00Z          |88                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T04:18:00Z          |75                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T04:24:00Z          |80                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T04:30:00Z          |83                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T04:36:00Z          |89                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T04:42:00Z          |74                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T04:48:00Z          |76                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T04:54:00Z          |85                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T05:00:00Z          |77                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T05:06:00Z          |85                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T05:12:00Z          |90                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T05:18:00Z          |72                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T05:24:00Z          |71                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T05:30:00Z          |87                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T05:36:00Z          |89                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T05:42:00Z          |76                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T05:48:00Z          |82                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T05:54:00Z          |79                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T06:00:00Z          |72                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T06:06:00Z          |84                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T06:12:00Z          |87                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T06:18:00Z          |73                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T06:24:00Z          |77                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T06:30:00Z          |82                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T06:36:00Z          |86                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T06:42:00Z          |83                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T06:48:00Z          |72                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T06:54:00Z          |81                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T07:00:00Z          |78                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T07:06:00Z          |79                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T07:12:00Z          |82                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T07:18:00Z          |87                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T07:24:00Z          |84                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T07:30:00Z          |77                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T07:36:00Z          |75                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T07:42:00Z          |83                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T07:48:00Z          |71                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T07:54:00Z          |82                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T08:00:00Z          |78                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T08:06:00Z          |78                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T08:12:00Z          |74                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T08:18:00Z          |77                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T08:24:00Z          |71                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T08:30:00Z          |72                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T08:36:00Z          |72                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T08:42:00Z          |82                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T08:48:00Z          |88                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T08:54:00Z          |70                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T09:00:00Z          |73                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T09:06:00Z          |79                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T09:12:00Z          |81                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T09:18:00Z          |87                            |degrees|average_temperature|coyote_creek      |
|         |       |0    |2019-08-17T00:00:00Z|2019-08-18T00:00:00Z|2019-08-17T09:24:00Z          |80                            |degrees|average_temperature|coyote_creek      |
| ... |