# Flux Standard Library Design

The Flux standard library will define the set of functions, constants, and other
values that are available out of the box with Flux, providing the functionality that is
most frequently used. Everything in the Flux standard library is part of a package.  
The top-level package is called "universe."

The Flux compiler provides the concept of a "prelude," which is the set of packages whose members
may be used in a Flux program without namespace qualification.  For example,
in the vanilla Flux CLI, the prelude includes the `universe` and `influxdata/influxdb`
packages. 

## The `Universe` Package

Fundamental transformations:
- `bottom`
- `distinct`
- `filter`
- `group`
- `join`
- `keyValues`
- `keys`
- `limit`
- `map`
- `pivot`
- `range`
- `set`
- `shift`
- `sort`
- `stateCount`
- `stateDuration`
- `top`
- `union`
- `window`
- `yield`


Schema changing operations:
- `drop`
- `duplicate`
- `keep`
- `rename`

Aggregate operations:
- `aggregateWindow`
- `covariance`
- `cov`
- `pearsonr`
- `count`
- `integral`
- `mean`
- `median`
- `percentile`
- `skew`
- `spread`
- `stddev`
- `sum`

Selector operations:
- `first`
- `last`
- `max`
- `min`
- `percentile`
- `median`
- `sample`

Highest/Lowest
- `highestMax`
- `highestAverage`
- `highestCurrent`
- `lowestMax`
- `lowestAverage`
- `lowestCurrent`

Binning transformations:
- `histogram`
- `histogramQuantile`
- `linearBins`
- `logarithmicBins`

Other mathematical operations:
- `cumulativeSum`
- `derivative`
- `difference`

## Package `testing`
- `assertEquals`
- `loadStorage`
- `loadMem`
- `test`

Others?

## I/O Packages

### Package `csv`
- `from`

### Package `http`
- `to`

### Package `influxdata/influxdb`
- `buckets`
- `from`
- `to`

### Package `influxdata/influxdb/v1`

- `json`
- `databases`
- `fieldsAsCols`
- `tagValues`
- `measurementTagValues`
- `tagKeys`
- `measurementTagKeys`
- `measurements`

### Package `kafka`
- `from`
- `to`

### Package `sql`
- `from`
- `to`

## Package `time`

Constants representing months.  From the SPEC:
[IMPL#154](https://github.com/influxdata/flux/issues/154)
```
January   = 1
February  = 2
March     = 3
April     = 4
May       = 5
June      = 6
July      = 7
August    = 8
September = 9
October   = 10
November  = 11
December  = 12
```

Constants representing days of the week. From the SPEC:
```
Sunday    = 0
Monday    = 1
Tuesday   = 2
Wednesday = 3
Thursday  = 4
Friday    = 5
Saturday  = 6
```

Time and date functions.  Each of these accept a `time` value and return an integer.
From the SPEC:
[IMPL#155](https://github.com/influxdata/flux/issues/155)
- `second` - integer Second returns the second of the minute for the provided time in the range [0-59].
- `minute` - integer Minute returns the minute of the hour for the provided time in the range [0-59].
- `hour` - integer Hour returns the hour of the day for the provided time in the range [0-59].
- `weekDay` - integer WeekDay returns the day of the week for the provided time in the range [0-6].
- `monthDay` - integer MonthDay returns the day of the month for the provided time in the range [1-31].
- `yearDay` - integer YearDay returns the day of the year for the provided time in the range [1-366].
- `month` - integer Month returns the month of the year for the provided time in the range [1-12].

System time function:
- `systemTime` produces a `time` value that is the current system time.

Functions that deal with time zones:
- `loadLocation` - accepts a string like `"America/Denver"` and produces a `location`
[IMPL#157](https://github.com/influxdata/flux/issues/157)
- `fixedZone`
[IMPL#156](https://github.com/influxdata/flux/issues/156)

The intervals function:
- `intervals` [IMPL#407](https://github.com/influxdata/flux/issues/407)

Builtin intervals (from the SPEC):
```
// 1 second intervals
seconds = intervals(every:1s)
// 1 minute intervals
minutes = intervals(every:1m)
// 1 hour intervals
hours = intervals(every:1h)
// 1 day intervals
days = intervals(every:1d)
// 1 day intervals excluding Sundays and Saturdays
weekdays = intervals(every:1d, filter: (interval) => weekday(time:interval.start) not in [Sunday, Saturday])
// 1 day intervals including only Sundays and Saturdays
weekdends = intervals(every:1d, filter: (interval) => weekday(time:interval.start) in [Sunday, Saturday])
// 1 week intervals
weeks = intervals(every:1w)
// 1 month interval
months = intervals(every:1mo)
// 3 month intervals
quarters = intervals(every:3mo)
// 1 year intervals
years = intervals(every:1y)
```

## Package `math`
[IMPL#332](https://github.com/influxdata/flux/issues/332)

## Package `strings`
[IMPL#332](https://github.com/influxdata/flux/issues/332)

## Package `experimental`

This is the place for packages that we may be using internally, but are not yet ready
for production use.
