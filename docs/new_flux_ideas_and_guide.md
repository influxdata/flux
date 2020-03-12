# Getting Started Querying with Flux and InfluxDB
**NOTE BEFORE YOU GET STARTED:** Forget much of what you know about Flux. This is a new version with some very 
important differences. As you work through this guide, please keep notes on your thoughts for each section 
as you encounter it. This ends in a test and you are expected to create a secret Gist with your thoughts and 
the answers to the questions. You can close with a summary of your feelings after going through the whole thing 
but please note your thoughts as you go along. It's very likely that you'll think something is crazy or off 
initially, but maybe think it's ok when you look back later to write your summary. We're interested in 
seeing this progression of your learning and thought process.

You're allowed to ask questions in the private channel set up for this usability test in Slack. #flux-usability-yourname

Welcome to the getting started guide for Flux. Flux is a functional scripting and query language that looks 
very similar to Javascript in some parts, but with shortcut syntax for common queries. Its goal is to make 
querying and working with data, and time series data in particular, easy and fast. While InfluxDB is the first 
class data source for Flux, it also provides adaptors to other databases, third party APIs, files, and formats that 
data can be represented in.

This guide doesn't present the entirety of the API or its functionality, but a subset to convey the essential 
concepts of the language. Go here for a [complete Flux API reference]().

## Data Model
The data model of InfluxDB consists of Buckets, which hold data. Time series data is written in 
via the HTTP API and is represented as InfluxDB Line Protocol. It separates time series into measurements, tags, 
fields, and a time. Here are some examples of the InfluxDB line protocol:

```
cpu,host=serverA,region=west usage_system=23.2,usage_user=55.1 1566000000
events,app=foo description="something happened",state=1i 1566000000
```

In the first line you have a measurement named `cpu`, two tag key/value pairs of `host=serverA` and `region=west`, 
two fields of `usage_system=23.2` and `usage_user=55.1` and a timestamp at the end of `1566000000`. Measurement 
names and tag values are always strings. Fields can be either float64, int64, bool, or string. The second line in
the snippet has data for the `events` measurement and shows a string and integer field value.

The combination of a measurement, tags, and a specific field represent an individual time series, which are an
ordered list of time, value pairs. For example, `cpu,host=serverA,region=west usage_system` would be one time 
series from line 1 in the above example. You can generally think of tags as the dimensions on which you can
query and slice and dice your time series data. However, you can also do that with fields, provided that you
pay attention to performance and schema design considerations.

With the data in InfluxDB it's important to keep this context of a time series in mind. There are different types 
of numbers that can be recorded in time series and many operations only make sense when performed on an individual 
series and then later combined in some other way. Here are some examples of data that is fed in a time series:

### Metrics
Metrics are regular time series that are values captured at specific intervals of time (like every 10 seconds or 1 
minute or 1 hour or 1 day). There are different kinds of metrics that are commonly captured

* counter - A value that counts the occurrence of an event. Counters alway increase unless they have 
            been wrapped or reset. A reset occurs when what ever is doing the counting is restarted in 
            which case it starts counting again from 0. A wrapped counter is one that has reached the 
            largest number allowed by the byte representation of the value, in which case it starts over 
            again at 0. These wraps and resets need to be accounted for when performing operations on counters. 
            An example of a counter is the number of bytes a network interface has received. 
* gauge - A value that measures a specific value, which can increase or decrease. An example of a gauge is 
          taking a temperature reading.

### Events
Events are single data points in time. This could be an individual request to an API and how long that took, a user 
logging into a service, a car parking, a light turning on, a container starting or stopping, or anything you can think 
of. Events are what are called irregular time series given that they can occur at any time. It is possible to create 
regular time series from an underlying event stream (like counting how many occurred every five minutes for the last hour).

### Working with data in Flux
When working with data in Flux, it translates the line protocol/time series structure into to a table. You 
can think of a measurement name as a table with the tag and field keys representing columns. Flux also 
includes the reserved column, `time`, in every table pulled from InfluxDB. The `time` column is the time at the end 
of the line protocol when data is written in.

Here's an example of line protocol and its translation to a Flux table:

```
h2o_temperature,location=santa_monica,state=CA surface_degrees=65.2,bottom_degrees=50.4 1568756160
h2o_temperature,location=santa_monica,state=CA surface_degrees=63.6,bottom_degrees=49.2 1600756160
h2o_temperature,location=coyote_creek,state=CA surface_degrees=55.1,bottom_degrees=51.3 1568756160
h2o_temperature,location=coyote_creek,state=CA surface_degrees=50.2,bottom_degrees=50.9 1600756160
h2o_temperature,location=puget_sound,state=WA surface_degrees=55.8,bottom_degrees=40.2 1568756160
h2o_temperature,location=puget_sound,state=WA surface_degrees=54.7,bottom_degrees=40.1 1600756160
```

| location | state | surface_degrees | bottom_degrees | time |
| --- | --- | --- | --- | --- |
| coyote_creek | CA | 55.1 | 51.3 | 1568756160 |
| coyote_creek | CA | 50.2 | 50.9 | 1600756160 |
| puget_sound | WA | 55.8 | 40.2  | 1568756160 |
| puget_sound | WA | 54.7 | 40.1  | 1600756160 |
| santa_monica | CA | 65.2 | 50.4 | 1568756160 |
| santa_monica | CA | 63.6 | 49.2 | 1600756160 |

Note that the measurement name doesn't appear anywhere. That's because the table itself represents data from the 
`h2o_temperature` measurement. Later we'll see how to work with multiple measurements (i.e. tables). Note that the 
underlying time series would be represented by the combination of `location`, `state`, and the specific field. 
Results are ordered by default by the columns of the tags and their values and then by time. So we can see that 
`coyote_creek`, `CA` rows are together in time ascending order.

We'll show later how to specify to the graphing engine what represents an individual time series that you want 
visualized.

## Basic Queries
This section will explore some basic queries in Flux to highlight the syntax and a few of the basic concepts of
the language. For these examples, we'll assume a dataset that has a schema like the one seen above.

### Introduction to basics and getting the last value
In this example, we'll get the last value of `surface_degrees` for every series in the `h2o_temperature` 
measurement in the `my_data` bucket. Flux has a shorthand syntax for the common operation of selecting data from 
a bucket with given criteria for a given time range. Let's look at that first for this example:

```
@my_data.h2o_temperature{surface_degrees, time}
```

The `@` symbol at the beginning indicates that we're using the shorthand syntax and that the next set of characters
will identify the bucket name. After that we have a `.` separator, followed by the measurement name of 
`h2o_temperature`. Immediately following that we have the **Column Filter Predicate** syntax (the curly braces
`{` and `}` and what's inside of them) to indicate which columns we want in the result. The identifiers can
also be wrapped in double quotes, which you'll need to use if they contain spaces:

```
@"my_data"."h2o_temperature"{"surface_degrees", "time"}
```

Given the previous data set, this query would produce a result like this: 

| surface_degrees | time |
| --- | --- |
| 50.2 | 1600756160 |
| 54.7 | 1600756160 |
| 63.6 | 1600756160 |

Since we didn't specify a time range in the query, we get the most recent data point for each series. How far 
back in time the database looks for a recent value is dependent on the setup of the backend InfluxDB system you 
are working with (this will typically be less than a few hours). We'll show how to specify explicitly how far 
back to look for the most recent value in a moment.

In this result set we have three rows, but we can't see their series information (i.e. the tags) because we 
didn't select them. Selecting them is fairly easy:

```
@my_data.h2o_temperature{location, state, surface_degrees, time}
```

Would return the last value of surface_degrees in each series:

| location | state | surface_degrees | time |
| --- | --- | --- | --- |
| coyote_creek | CA | 50.2 | 1600756160 |
| puget_sound | WA | 54.7 | 1600756160 |
| santa_monica | CA | 63.6 | 1600756160 |


In Flux, a tag key and a field key are both called columns. This means that tag keys and field keys must
be unique within a measurement. Here's an example that explicitly selects all the columns we know about:

```
@my_data.h2o_temperature{location, state, surface_degrees, bottom_degrees, time}
``` 

We can also select all columns by leaving the column list empty:

```
@my_data.h2o_temperature{}
``` 

Or you can leave out the curly braces:

```
@my_data.h2o_temperature
```

If your **Column Filter Predicate** selects only a tag, you'll get the same number of records as you have series for 
that tag since the last value for each series is what gets defined if you're not filtering by time range or other 
criteria.

```
@my_data.h2o_temperature{state}
```

Returns these results:

| state |
| --- |
| CA |
| CA |
| WA |

#### Filtering on Criteria
You can limit the results returned in the query by specifying filtering criteria in the *Column Filter Predicate* block. 
Here's an example that filters based on the location:

```
@my_data.h2o_temperature{location, surface_degrees, time, state == "CA"}
```

| location | state | surface_degrees | time |
| --- | --- | --- | --- |
| coyote_creek | CA | 50.2 | 1600756160 |
| santa_monica | CA | 63.6 | 1600756160 |

Note that `state` is returned in the result because it was listed in the filter criteria. You can have more complex 
logic to match criteria like:

```
@my_data.h2o_temperature{location, surface_degrees, time, location == "coyote_creek" OR location == "puget_sound"}
```

See that while the column names are separated by commas, the filter criteria is separated by the different matching 
clauses.


#### Matching Critera
There are many operators you can use to filter down data. Here is the list of operators:
* `==` (equal to)
* `!=` (not equal)
* `=~` (regex match)
* `!~` (not a regex match)
* `>` (greater than)
* `>=` (greater than or equal to)
* `<` (less than)
* `<=` (less than or equal to)
* `in`

The regex and in operators are probably the trickiest ones to work with, so here are a few examples:

```
// this is a comment because it has // before it

// status is a tag and this will match any stats between 200 and 500
@my_data.http_requests{status =~ /[2-5][0-9][0-9]/}

// will match states in the array
@my_data.h2o_temperature{state in ["CA", "WA"]}
```

### Aggregate, Select, Transform and Group Data
Flux has many built in functions to work with time series data. **Aggregate** functions refer to functions that summarize 
some period of time. They take N rows and output a single result, examples include count, mean, and percentile, 
among many others. **Select** functions select a set of rows from an input set based on some criteria. Min, max, top, and 
bottom are examples of selectors. Some selectors also double as aggregates like min and max.

This is an important distinction. An aggregate produces new data from some source data. That new data is frequently a 
summary statistic that describes the source data in some meaningful way. A selector specifies criteria for selecting 
rows from some source data. So you're not transforming the data when using a selector. For example, `count` is not a 
function that could be used as a selector.

#### Sample Dataset
Here's an example dataset that we'll use for the rest of the examples in this section. The raw number timestamps 
have been replaced with human readable times and the values have been set to simple whole numbers to make things 
easier. The example times only go down to minute precision to make things easier, but you can summarize down to 
the nanosecond.

```
h2o_temperature,location=santa_monica,state=CA surface_degrees=65,bottom_degrees=50 2020-02-22T15:01
h2o_temperature,location=santa_monica,state=CA surface_degrees=64,bottom_degrees=49 2020-02-22T15:31
h2o_temperature,location=santa_monica,state=CA surface_degrees=63,bottom_degrees=49 2020-02-22T16:01
h2o_temperature,location=santa_monica,state=CA surface_degrees=62,bottom_degrees=49 2020-02-22T16:31
h2o_temperature,location=santa_monica,state=CA surface_degrees=61,bottom_degrees=48 2020-02-22T17:01
h2o_temperature,location=santa_monica,state=CA surface_degrees=60,bottom_degrees=48 2020-02-22T17:31
h2o_temperature,location=santa_monica,state=CA surface_degrees=60,bottom_degrees=48 2020-02-22T17:46
h2o_temperature,location=coyote_creek,state=CA surface_degrees=55,bottom_degrees=51 2020-02-22T16:05
h2o_temperature,location=coyote_creek,state=CA surface_degrees=53,bottom_degrees=50 2020-02-22T17:05
h2o_temperature,location=puget_sound,state=WA surface_degrees=55,bottom_degrees=40 2020-02-22T17:10
h2o_temperature,location=puget_sound,state=WA surface_degrees=54,bottom_degrees=40 2020-02-22T17:40
```

In the data set we can see there are three distinct keys and that the time range of data is from 
`2020-02-22T15:01` to `2020-02-22T17:46` (about a two and a half hour span). There are more data 
points for the `santa_monica`, `CA` location than for the others. The differing number of values 
will be instructive later when we shape and summarize this data.

#### Summarizing and Grouping with Aggregates
Let's get into an example for summarizing some of our water data. For the sake of these queries, assume that 
we're executing them at `2020-02-22T18:00`, that way we can use relative times in the query. Say we want to get the 
last 3 hours of data, with the min, max, and mean of `bottom_degrees` for that time period. With that 
definition, we'd expect to get 1 row per series in the dataset (assuming that we have min, max, and mean as 
columns in the resulting table. Here's the query:

```
@my_data.h2o_temperature{bottom_degrees, time > -3h}
    |> aggregate({min(bottom_degrees), max(bottom_degrees), mean(bottom_degrees)}) 
```

Will return the following dataset:

| min_bottom_degrees | max_bottom_degrees | mean_bottom_degrees |
| --- | --- | --- |
| 48 | 51 | 47.45 |

Even though the sample data set has three series, we get back only a single row that gives the summary statistics we 
requested. The call to the aggregate function passed in an **Aggregate Summary Block** which looks similar in 
construction to the **Predicate Filter Block** that we used on the first line of the query. The returned column names 
are given automatically based on the construction `<function name>_<field name>`. As with the other blocks, field 
names can be wrapped in double quotes to handle cases like spaces or other special characters in a field name.

Our results don't have the descriptive tag metadata because we didn't ask for it. Selecting `location` and `state` 
which are the two tags that we have for data in the `h20_measurements` measurement will ensure it comes through in 
the aggregate and we can then group by those columns:

```
@my_data.h2o_temperature{location, state, bottom_degrees, time > -3h}
    |> aggregate({min(bottom_degrees), max(bottom_degrees), mean(bottom_degrees)}, by: ["location", "state"]) 
```

| location | state | min_bottom_degrees | max_bottom_degrees | mean_bottom_degrees |
| --- | --- | --- | --- | --- |
| coyote_creek | CA | 48 | 50 | 48.714 |
| puget_sound | CA | 40 | 40 | 40 |
| santa_monica | WA | 50 | 51 | 50.5 |

Although we could have left out the `state` grouping from our example because in this case the `location` is 
enough to unique identify a series.

Say we wanted to get a summary per state: 

```
@my_data.h2o_temperature{location, state, bottom_degrees, _time > -3h}
    |> aggregate({min(bottom_degrees), max(bottom_degrees), mean(bottom_degrees)}, by: ["state"]) 
```

| state | min_bottom_degrees | max_bottom_degrees | mean_bottom_degrees |
| --- | --- | --- | --- |
| CA | 48 | 51 | 49.11 |
| WA | 40 | 40 | 40 |

See that the results don't have the `location` column even though we selected it on the first line. That's because 
it wasn't included in the `by` set of columns in the call to `aggregate`.

The aggregate function can window the data by time intervals. For example, if we wanted to summarize each hour for 
the last three hours:

```
@my_data.h2o_temperature{location, bottom_degrees, time > -3h}
    |> aggregate({min(bottom_degrees), max(bottom_degrees), mean(bottom_degrees)}, by: ["location"], window: 1h) 
```

Because the `aggregate` function is breaking the records out into one hour intervals, the `time` column is kept, 
but the values are updated to the end of the window. For example, data at 2020-02-22T16:05 would get summarized in 
a record with the `time` value (window) of `2020-02-22T17:00`. 

| location | min_bottom_degrees | max_bottom_degrees | mean_bottom_degrees | time |
| --- | --- | --- | --- | --- |
| santa_monica | 48 | 50 | 48.714 | 2020-02-22T16:00 |
| santa_monica | 48 | 50 | 48.714 | 2020-02-22T17:00 |
| santa_monica | 48 | 50 | 48.714 | 2020-02-22T18:00 |
| coyote_creek | 51 | 51 | 50 | 2020-02-22T17:00 |
| coyote_creek | 50 | 50 | 50 | 2020-02-22T18:00 |
| puget_sound | 40 | 40 | 40 | 2020-02-22T18:00 |

Remember there were three series, so for 1h summaries for the last 3 hours we'd expect to have three rows per group 
for nine total rows. But looking at the results, we only see the full three summaries for the `santa_monica` location. 
This is because the other two series don't have data in every window of time.

You can create rows for the missing summaries by calling `interpolate` after the call to aggregate. It can be used to 
fill in default values or values based on a summary statistic of the rows like `first`, `min`, `max`, `mean`, etc. You 
can find the [details for interpolate in the function reference]().

#### Select
The `select` function works similar to `aggregate` in that it operates on windows of time and groupings of rows.
However, The difference is that select returns specific rows, rather than a summary data point. Here are a few 
examples:

```
@my_data.h2o_temperature{location, state, bottom_degrees, surface_degrees, time > -3h}
    |> select(fn: min("bottom_degrees"), by: [])
```

| location | state | bottom_degrees | surface_degrees | time
| --- | --- | --- | --- | --- |
| puget_sound | WA | 40 | 55 | 2020-02-22T17:10 |
| puget_sound | WA | 40 | 54 | 2020-02-22T17:40 |

Notice that all the columns are pulled through, the selected column name remains the same and note that 
we've pulled in two records. We pull in two records because both of those have the same `min` `bottom_degrees` 
value of `40`, which is the min for all records in that time. Notice also that `time` is the same value as those 
individual records, which you don't see in calls to `aggregate`.

You can also select in groups based on time or by other columns, just like you can with `aggregate`. Here's 
an example using time grouping and inserting the time window for associated row selections:

```
@my_data.h2o_temperature{location, state, bottom_degrees, _time > -3h}
    |> selector(fn: max(bottom_degrees), by: ["state"], window: 1h, windowColumn: "window")

// if you include window_column, the result will put the window timestamp for each selected row in that column
```

Produces this result:

| location | state | bottom_degrees | time | window |
| --- | --- | --- | --- | --- |
| santa_monica | CA | 50 | 2020-02-22T15:01 | 2020-02-22T16:00 |
| coyote_creek | CA | 51 | 2020-02-22T16:05 | 2020-02-22T17:00 |
| santa_monica| CA | 48 | 2020-02-22T17:01 | 2020-02-22T18:00 |
| santa_monica| CA | 48 | 2020-02-22T17:31 | 2020-02-22T18:00 |
| santa_monica| CA | 48 | 2020-02-22T17:46 | 2020-02-22T18:00 |
| puget_sound | WA | 40 | 2020-02-22T17:10 | 2020-02-22T18:00 |
| puget_sound | WA | 40 | 2020-02-22T17:40 | 2020-02-22T18:00 |

What this does is for each 1 hour period, it groups the rows by state and then selects the rows with the max value. 
Because we included the `windowColumn` argument, that is included in the result with the ending timestamp of the 
window that each record fell into.

Since so many of the records have tied `max(bottom_degrees)` values, we see that many of the original 
rows were selected in the result. We can also see the `window` that is inserted to specify 
which block of time the selector was working across.

The following functions can be used in the **Selector Predicate Block**: `first`, `last`, `min`, `max`, `top`, `bottom`

#### Calling Aggregates without the Aggregate function
The `aggregate` function provides a shorthand for computing multiple aggregates at a time. However, the 
functions passed in can also be called on tables. For example:

```
// get the most recently written record
@my_data.h2o_temperature |> last()

// get the most recently written record for each location
@my_data.h2o_temperature |> last(by: ["location"])

// get the average for each hour for each location
@my_data.h2o_temperature{time > -3h} |> mean(by: ["location"], window: 1h)
``` 

Selectors can only be called with the `select` function.

#### Count, Distinct, and CountDistinct
The `count` function is an aggregate while the `distinct` function is a selector. The `countDistinct` function is 
a helper function that will combine operations for you.
 
```
@my_data.h2o_temperature{state, time > -3h}
    |> select(fn: distinct(state), window: 1h, windowColumn: "time")
```

Would return the following result

| state | time
| --- | --- |
| CA | 2020-02-22T16:00
| CA | 2020-02-22T17:00
| CA | 2020-02-22T18:00
| WA | 2020-02-22T18:00

We can see that CA reported in all three hours while both CA and WA reported in only the last hour. Also see that 
distinct discards columns that aren't in the `by` argument or the `distinct` column itself. So we use the `time` 
name for where we put the window markers for each distinct state. You could then count the number of states reporting 
per hour:

```
@my_data.h2o_temperature{state, time > -3h}
    |> select(fn: distinct(state), window: 1h)
    |> timeShift(-1m)
    |> count("state", window: 1h)
```

Would produce:

| state | time
| --- | --- |
| 1 | 2020-02-22T16:00
| 1 | 2020-02-22T17:00
| 2 | 2020-02-22T18:00

See that we had to use the `timeShift` function to move the timestamps of the windows back so that the resulting 
times would be accurate for `aggregate`.

All of this is encapsulated by the `countDistinct` function. For example, if we wanted to count the number of 
distinct locations that reported in each sate for each hour in the last three hours.

```
@my_data.h2o_temperature{location, state, bottom_degrees, time > -3h}
    |> countDistinct(location, by: ["state"], window: 1h)
```

### Math, basic transformations, and using previous results in new queries
Flux makes it easy to do basic math and transformations across tables of data. First, let's look at 
renaming columns. All of the different predicate blocks recognize the `as` keyword to rename. For 
example:

```
@my_data.h2o_temperature{location, state, bottom_degrees as degrees, top_degrees, time}

// or renaming columns in an aggregate:
@my_data.h2o_temperature{bottom_degrees, time > -1h}
    |> aggregate({min(bottom_degrees) as min, max(bottom_degrees) as max})
```

Here are some examples of doing math and introducing the concept of variables:

```
temps = @my_data.h2o_temperature{bottom_degrees, surface_degrees, time, location == "coyote_creek"}

// do some math and return a new table with an addition column. Note that as is required on the end. Resulting
// table will have all the columns of the previous on and this new one
new_table = temps.surface_degrees - temps.bottom_degrees as bottom_surface_difference

// or you can pipe forward it to another function
temps.surface_degrees - temps.bottom_degrees as bottom_surface_difference
    |> sort(by: ["bottom_surface_differrence"])

// convert from fahrenheit to celsius and replace the value previously held by bottom_degrees
(temps.bottom_degrees - 32) * 5 / 9 as bottom_degrees
```

The above example shows assigning a result set to a variable so that it can be used multiple times later. You 
can also use this construction to do implicit joins across tables. Here's a made up example:

```
foo_data = @my_data.foo{value as foo_val, and time > -1h}
bar_data = @my_data.bar{value as bar_val, and time > -1h}

foo_data.foo_val - bar_data.bar_val as difference_to_bar
```

The above call will do a join between the two tables and rows based on all string columns matching and the `time`. 
The result data set will include all columns from both tables along with the new column 
`differece_to_bar`. If a column of the same name exists in both tables, the value from the left hand side 
table will be used (`foo_data` in this example). So this result would have the columns: 

| foo_val | bar_val | difference_to_bar | time
| --- | --- | --- | --- | 

## Advanced Concepts
We'll fill this in later. Topics will include joins, unions, map, and reduce. We shouldn't need any of these 
more explicit operations to answer the questions in this test. The implicit joins via math should be enough 
to make this work. However, there is one trick you'll likely need...

### Using Data From a Previous Result in a new Query
You can query the database for results and use either a record, an individual value, or an entire column in 
subsequent queries. This is useful for pulling back data from top results, etc. Here's an example:

```
// get the top 10 locations in terms of surface degrees
top_temps = @my_data.h2o_temperature{surface_degrees, location}
    |> selector({top(surface_degrees, 10)})

// now get the last hour of data for those 10 locations
@my_data.h2o_temperature{location in top_temps.location and time > -1h}
```

## Function Reference for Test
In addition to the getting started guide above, you'll need some combination of the following functions 
to complete the answers for the test. This is by no means a complete API reference, but only a limited set of 
functions required to complete this task. You may or may not use all of these functions, depending on how 
you choose to solve each problem.

Note that the function definitions all contain named arguments, which are also positional. The first argument 
is always `table`, since the first argument is always what receives the pipe forward. For example if we had this:

```
// define functon foo
foo = (table, columnName) => {
    // some function definition here
}

// we can call it like
some_table |> foo("my_column")

// or like this
some_table |> foo(columnName: "my_column")

// or like this
foo(some_table, "my_column")

// or like this
foo(some_column, columnName: "my_column")

// or like this
foo(table: some_column, columnName: "my_column")
``` 

### Aggregate
Aggregate will take aggregate functions like `quantile`, `min`, `max`, `mean`, `count` and others. It 
groups data by the given columns and windows of time.

```
aggregate = (table, fn, by = [], window = 0) => {}
```

The default `window` argument of zero means that the aggregate will be computed for all time seen in the table. 
That is, a single row will be returned per group.

### Select
Select will take selector functions like `top`, `bottom`, `min`, and `max` and select rows matching 
the given function by the given columns and time.

```
select = (table, fn, by = [], window = 0) => {}
```

### Max
Max works as either an aggregate or a selector. If used in an aggregate it will return a single value that is 
the max for the given grouping and time window. If used in a selector, it will return the rows that have the 
max value for the grouping and window.

```
max = (table, column, by = [], window = 0) => {}
``` 

### Min
Min works as either an aggregate or a selector. If used in an aggregate it will return a single value that is 
the min for the given grouping and time window. If used in a selector, it will return the rows that have the 
min value for the grouping and window.

```
min = (table, column, by = [], window = 0) => {}
``` 

### Mean
Mean is an aggregate function (so it can only be used as an argument to `aggregate`).

```
mean = (table, column, by = [], window = 0) => {}
```

### Count
Count is an aggregate function (so it can only be used as an argument to `aggregate`).

```
count = (table, column, by = [], window = 0) => {}
```

### Top
The top function is passed into `select` to produce the top n rows per grouping per window of time.

```
// notice that top takes a view, which is what the selector function sends to it. Top can only be 
// used as an argument to selector. N is the number of results you'd like
top = (table, column, n) => {}
```

### Sort
Sort will sort the results in the table.

```
sort = (table, by, desc = true) => {}
```

### Rate
Rate will compute the rate on a units per interval based on the passed in parameters. It is called on its own and not
as an argument to `select` or `aggregate`. It can be used to compute the rate for individual series, or for the 
sum across a number of series, all depending on what is passed to `by`.

```
// function signature
rate = (table, columns, period = 1s, by = [], window = 0) = {}
```

We can see that only the `table` and `columns` arguments are required with the other arguments being 
optional. Columns is a map with the key being the new column name and the value being the column you're 
computing a rate on. So you can compute the rate for multiple columns on a single call to the rate function.

Period determines what the rate period is (like per second, minute or hour). The `by` argument 
specifies what dimensions you'd like to compute a `rate` on. If this is passed in, but no `window` is passed 
the result will be a single row per grouping, which represents the average rate of that entire grouping 
over the whole time period.

If you wanted to get rate in time windows, you'd pass in `window` so for example, you could see the rate in per 
minute units for every hour for the last 24 hours (assuming you queried 24 hours of data to begin with).

The rate function wraps up all of the math operations required to calculate a rate across multiple 
dimensions so you won't have to worry about coming up with a nonsensical result.

```
// computing the rate on individual series. This will calculate the rate on a point to point basis from 
// column named value and put the result in units_per_second. value will not be in the result set as it
// no longer applies
data |> rate({value as units_per_second}, by: ["column1"])

// you can compute the rate on a per minute 
data |> rate({value as units_per_minute}, by: ["column1", "column2"], period: 1m)

// you can compute on some other dimension in per second units with one point per 10m
data |> rate({value as units_per_secod}, by: ["region"], window: 10m)
```

Rate wraps a number of calls into one thing depending on which arguments you pass in. With any call to rate, it starts with computing the rate between individual value,time pairs in each series. If you havn't passed in a `time` or `by` arugment, you're done. If you passed in a `time` arugment, rate will then call `aggregate` and compute the `mean` rate across those windows of time. If you passed in a `by` argument, it will then call `aggregate` again, but this time computing the `sum` of the previous means. This progression is very specific and often trips people up, so we've wrapped it in this function so you don't have to remember the correct order to get a useful result.

### Increase
Increase works on columns that are counters. Like the rate function, it is called by itself (not as an argument to 
aggregate or selector). Its arguments `by` and `window` operate in the same way that rate does. It will calculate 
the total increase accounting for resets in the counter (note this increase is not like the old Flux one). For 
example, this could be used to determine the total number of bytes send over a network card in a given window of time.

```
increase = (table, columns, by = [], window = 0) = {}
```

### Filter
Filter will filter a result set to the requested columns and records matching the predicate criteria.

```
// function signature
filter = (table, fn) => {}

// calling it
data |> filter(fn: (r) => r.foo == "bar") // will return all columns in data, but only rows where the value of foo is bar

data |> filter({fn: (r) => r.foo == "bar" and r.time > -3h and r.time < -1h) // will return only rows in specified time range
```

### Drop
Drop will drop the specified columns from the table, including system columns if requested:
```
// function signature
drop = (table, columns) => {}

// calling it
data |> drop(["foo", "_key"])
```

### Keep
Keep will keep only the specified columns in the table, while keeping any system columns (starting with `_`):

```
// function signature
keep = (table, columns) => {}

// calling it
data |> keep(["foo", "_time"])
```

### Rename
Rename will rename a column from one thing to another. Note that the `_key` values will not be altered even 
if that column appears in the key. Here's the function signature and how it can be called:

```
// function signature
rename = (table, from, to) => {}

// calling it
data |> rename(from: "foo", to: "bar")
```

### Graphing
In order for the visualization front end to know which rows group into a series, you can call the `graph` function 
that will include in the result set information about what constitutes a series that should be graphed. This function 
must be called if you will be visualizing with any kind of graph with an x and y axis. Nothing need be done for 
tables and single stat panes. In the future, Flux will have functions instructing the front end to visualize and 
structure the data in any of the ways it supports, like single stat panes, tables, bar charts, line graphs, heatmaps, 
histograms and many more.

Here's an example of the basic graph call:

```
@my_data.cpu{host, usage_system, usage_user, time, cpu == "cpu-total" AND time > -1h}
    |> graph(yCols: ["usage_system", "usage_user"], xCol: "time", by: ["host"])
```

## The Test
For the following questions, you will need to write the query that you think will give the desired result and 
write a small sample of what you expect the text result from the API to look like. Your answer should be a 
secret Gist as a single Markdown file. Your query should be wrapped in code blocks and your result expectation 
should be a markdown table (like used in this guide earlier).

These questions intentionally leave out metadata queries as those are likely to be single functions once we 
refine things. These questions are based on something like Telegraf data from `cpu`, `mem`, `disk`, and `net`. 
There’s also another measurement for `http_requests_total` from some service.

### Example Data
The cpu measurement looks like this:

```
cpu,host=serverA,region=east usage_system=0.5,usage_user=0.6 1568756160
cpu,host=serverB,region=east usage_system=0.5,usage_user=0.6 1568756160
cpu,host=serverC,region=west usage_system=0.5,usage_user=0.6 1568756160
cpu,host=serverD,region=west usage_system=0.5,usage_user=0.6 1568756160
```

where the fields `usage_system` and `usage_user` are gauges.

The `mem` measurement looks like this:

```
mem,host=serverA,region=east used_percent=50 1568756160
mem,host=serverB,region=east used_percent=50 1568756160
mem,host=serverC,region=west used_percent=50 1568756160
mem,host=serverD,region=west used_percent=50 1568756160
```

where the field `used_percent` is a gauge between 0 and 100.

The `disk` measurement looks like this:

```
disk,host=serverA,region=east,path=/ used_precent=50 1568756160
disk,host=serverA,region=east,path=/mnt/media used_precent=50 1568756160
disk,host=serverC,region=west,path=/ used_precent=50 1568756160
disk,host=serverC,region=west,path=/mnt/media used_precent=50 1568756160
```

where the field `used_percent` is a gauge between 0 and 100.

The `net` measurement looks like this:

```
net,host=serverA,region=east,interface=eth0 bytes_in=1000,bytes_out=18548 1568756160
net,host=serverA,region=east,interface=eth1 bytes_in=1000,bytes_out=18548 1568756160
net,host=serverC,region=west,interface=eth0 bytes_in=1000,bytes_out=18548 1568756160
net,host=serverC,region=west,interface=eth1 bytes_in=1000,bytes_out=18548 1568756160
```

where the fields `bytes_in` and `bytes_out` are counters.

The `http_request_total` measurement looks like this:

```
http_request_total,host=serverA,region=east,service=log,status=200 value=8978 1568756160
http_request_total,host=serverA,region=east,service=auth,status=200 value=8978 1568756160
http_request_total,host=serverA,region=east,service=auth,status=503 value=8978 1568756160
http_request_total,host=serverC,region=east,service=log,status=200 value=8978 1568756160
http_request_total,host=serverC,region=east,service=auth,status=200 value=8978 1568756160
http_request_total,host=serverC,region=west,service=auth,status=503 value=8978 1568756160
```

where the field `value` is a counter.

### Questions
For the following questions, answer each letter (e.g. `1.i`, `1.ii`, etc). Make each a `##` heading in your Gist. 
Please also include your thoughts about the language and each question as you go through. These can be raw 
stream of consciousness thoughts, we want any random thoughts you have as you go through the test.
 
1. Get the last value and host seen for the `cpu` measurement and the `usage_system` field
    1. You want the host, value, time (this is for a single stat pane, a single host)
    1. You want the last value, host, time for each host (this is for a tabular view)
    1. You want the last value, host, time for each host, but only in the uswest region (this is for a tabular view)
    
1. Get `usage_system` and `usage_user` from the `cpu` measurement for the last hour
    1. Only where usage_user is greater than 50.0 and host is serverA (we expect two lines in a graph)
    1. For each host (we expect two lines on the graph per host)
    1. For the host, serverA, add usage_system and usage_user and put that result in a new column titled 
       usage_system_and_user, while keeping usage_system and usage_user. (we expect three lines in the graph)
    
1. Calculate the min, max, mean for `used_percent` for the `mem` measurement
    1. For the last hour for a host named serverA (we expect a tabular view)
    1. For each region, returning the value of each and the region name (we expect a tabular view)
    1. For a host named serverA in 5m buckets for the last 4h (we expect three lines in a graph)
    1. For each region in 5m buckets for the last 4h, returning the value, region name, and time (we expect three 
       lines in the graph for each region)
    
1. Get the rate in bytes per second from the `bytes_in` and `bytes_out` fields for the `net` measurement. These fields 
   are counters that always increase except when they have a reset.
    1. For a host named serverA for the last 5 minutes (we expect two numbers to be displayed in single stat panes).
    1. For each host (we expect two stats per host in a tabular view)
    1. For each service (many hosts in a single service) (we expect two stats per service in a tabular view)
    1. Do i but in MB per minute for the last 4 hours with one point every 5m per line (we expect graphs with the 
       same number of lines as stats as before)
    1. Do ii but in MB per minute for the last 4 hours with one point every 5m per line (we expect graphs with the 
       same number of lines as stats as before)
    1. Do iii but in MB per minute for the last 4 hours with one point every 5m per line (we expect graphs with the 
       same number of lines as stats as before)
    
1. Calculate the error percentage per minute for the last hour from the `http_requests_total` measurement and the 
   field `value`, where there is a tag named status that has a status code ranging from 200-500. Error 
   percentage is defined as (rate of requests where status is 500) / (total rate of requests)
    1. For requests to host named serverA (we expect a single stat pane)
    1. For requests to each service (we expect a tabular view)
    1. Same as i, but in 5m windows for the last 4 hours (we expect a single line on a graph)
    1. Same as ii, but in 5m windows for the last 4 hours (we expect a line per service on a graph)
    
1. Show 10 hosts using the most disk space (`disk` measurement and `used_percent` field)
    1. For the last hour (we expect a tabular view)
    1. Use those 10 hosts and get their disk space used over the last hour (we expect 10 lines on the graph 
       representing the 10 hosts)

## Supplemental Information for Alternative Ways to Represent Shortcut Syntax (this isn't updated and is just cruft)
As mentioned before, Flux is a functional language. That means the results of one operation can be forwarded
to a function that can transform those results, reshape them, modify or insert new values, or join two separate
tables (results) together. Let's look at what our shorthand syntax looks like in the full functional 
representation:

```
@my_data.h2o_temperature{location, state, surface_degrees}
// This is a comment because it follows // so it is ignored by the query engine
// The above query line becomes:
from(bucket: "my_data", measurement: "h2o_temperature", columns: ["location", "state", "surface_degrees"])
```

We've introduced comments, which follow anywhere in a line after `//`. This example only produces one function so 
it's not very instructive, but let's look at the `from` function for a moment. We can see from the example that 
Flux has named arguments. Named arguments can either be required or have default values. In the case of `from`, 
only `bucket` and `measurement` are required, which means it always returns data from a single measurement (which 
is why you can think of a measurement as a table). We'll look at 
[how to join data from multiple measurements (or tables)]() in a later section.

Flux also supports positional arguments. The call to `from` above can also be written in the following two ways:

```
from("my_data", "h2o_temperature", columns: ["location", "state", "surface_degrees"])
// is equivalent to

from(bucket: "my_data", measurement: "h2o_temperature", columns: ["location", "state", "surface_degrees"])
// Because the definition of the from function has bucket, measurement, and columns as the first three arguements.
// There are others, which we'll introduce later
``` 

To show calling another function, let's look at getting the last value, but only for series in the `state` of 
`CA`:

```
@my_data.h2o_temperature{location, surface_degrees, state == "CA"}
// will map to this Flux functional code
from(bucket: "my_data", measurement: "h2o_temperature", columns: ["location", "state", "surface_degrees"])
    |> filter({state == "CA"})
```

We've introduced a number of new concepts here. From the top line, we can see that the last part of the **Column 
Filter Predicate** is a matcher of `state == "CA"`. Matchers like this are always the last argument in the filter
predicate. They can be any boolean expression and can be nested like `(location == "platte_river" or location == 
"cherry_creek") and state == "CO"`. Notice also that in our shortcut syntax on the top line, we left `state` out 
of the list of columns that we want in the result, but that columns is still brought in because the matcher 
`state == "CA"` tells Flux to include it. That's why we see `state` included in the list of columns in our call
to `from` (it has to be included for the call to `filter` to work). 

In the matcher part of the filter predicate, the column identifier is always on the left hand side. This can
either be wrapped in double quotes or not. We'll cover all of the [available matching operators]() in a later 
section.

The next thing to notice is the pipe forward operator `|>`, which says, take the value returned from the function 
call to `from` and send that as the first argument into the `filter` function. The functional representation could 
be a one-liner (everything on a single line), but it's idiomatic to split function calls up on separate lines. 

Another thing to note on the functional translation is how we're passing the filter predicate to the `filter` 
function. It's passed as a positional argument `{state == "CA"}` The following would be equivalent ways to get 
the same thing:

```
from(bucket: "my_data", measurement: "h2o_temperature", columns: ["location", "state", "surface_degrees"])
    |> filter(fn: {state == "CA"})

// is the same as
from(bucket: "my_data", measurement: "h2o_temperature")
    |> filter(fn: {"location", "surface_degrees", state == "CA"})
// Because columns is an optional argument to from and if left out, all columns will be included. We can
// also see that the filter predicate block includes the columns to filter down to

// is the same as
from(bucket: "my_data", measurement: "h2o_temperature", columns: ["location", "state", "surface_degrees"])
    |> filter(fn: (r) => r.state == "CA")
// in this case, we're not passing a filter predicate block, but instead an anonymous function, which
// won't change the columns in the result, rather it will only specify which rows to include
```

The last example shows Flux' anonymous function syntax. Don't worry if it looks strange to you, we'll come back 
to it later.

Finally, let's look at how we can get the last value for each series in a measurement where the series must have
`state == "CA"`,  but only if that value appears in the last five minutes:

```
@my_data.h2o_temperature{location, state, surface_degrees, state == "CA" and _time > -5m}
  |> select(fn: last, by: ["_key"]) 
```

First, we're selecting the data we want to work with, which is the `surface_degrees` field from the 
`h2o_temperature` measurement for everything in the `state` of `CA` and we want all data from the last 
five minutes. From there we're passing the result to a new function `select` which takes arguments `fn` and 
`by`. The `fn` argument takes a function that can be applied to select rows in a grouping (in this case `last`). 
The `by` argument takes an array of column names to group rows by. In this case we group by `_key` because we 
want to get the last value for each series. We'll dig into the other grouping functions [aggregate]() and 
[transform]() and how to group by different columns and dimensions later in this guide.

Select can also be called with positional arugments like this: `select(last, ["_key"])`. Also, since the 
`_key` arugment for `select` has a default value of `[_key]`, it can be called with just the function like 
this: `select(last)`.

We expect the result set to look like this:

| location | state| surface_degrees | _key | _time |
| --- | --- | --- | --- | --- |
| coyote_creek | CA | 50.2 | location=coyote_creek,state=CA | 1600756160 |
| santa_monica | CA | 63.6 | location=santa_monica,state=CA | 1600756160 |

The above query can also be represented in the full Flux functional notation like this:

```
from(bucket: "my_data", measurement: "h2o_temperature", columns: ["location", "state", "surface_degrees"])
    |> range(start: -5m)
    |> filter(fn: (r) => r.state == "CA")
    |> select(fn: last, by: ["_key"])
```

The `range` function operates explicitly on the `_time` column to limit results.
Generally it's preferred to use the shortcut syntax when possible, but showing what the full translation to Flux 
functions has hopefully served as a gentle introduction to Flux functional concepts.
