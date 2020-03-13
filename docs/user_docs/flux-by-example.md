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
