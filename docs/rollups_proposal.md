# Rollups

Users want to be able to downsample their data and then query across each downsampled bucket transparently.
Something similar to Graphite's or RRD metric databases.
There are two key aspects to how those systems work:

* Data is downsampled automatically via a predefined configuration.
* Data is queried across the entire range of data without the user needing to know how downsampling was configured.

This creates a nice user experience where queries over the data range return quickly because as the query covers a larger time period less dense data is processed.
Instead of querying data proportional to the time range the data processed is relatively constant.
This should then result in constant time queries regardless of the time range queried.


## How should this use case be addressed in the InfluxDB ecosystem?

This proposal is that we create a _rollups_ feature.
There has been lots of talk over the years about _intelligent rollups_, however we can deliver on a simpler form of rollups that will more than satisfy the user's needs.
A _rollup_ is a predefined downsampling configuration, where the user provides the platform with the information it needs to rollup data and to query data back out of the rollup.
This simple definition eliminates some of the complexities in the _intelligent rollups_ proposals from the past that attempted to in one form or another determine how data should be downsampled automatically.
This  _rollup_ is equivalent to how Graphite and RRD provided these same features.


### How does one use a rollup?

We should create a Flux package named `rollups` that can query data across different downsampled retentions transparently for the user.
At a high level the Flux API would be similar to this:

```flux
import "influxdata/influxdb/rollups"


// Query data across the entire rollup
rollups.from(name: "mydata", start: -1y, stop: now())
    |> filter()
```

This query returns the rolledup data across the entire time range.
It reads similar to using a normal `from` call but we are being explicit that we want data from a rollup.

### How do ones know which rollups exist?

In order to use a _rollup_ as user must configure it via the API.
At a minimum we would need information about which buckets cover which time ranges.
At a maximum we would need information about the source raw bucket, which aggregate methods to apply and over which retentions and resolutions.

We should explore what users are looking for if we want to build to full system of defining the aggregates for the rollup etc or if we want them to create those manually and then just inform Flux where data lives for which time periods.


## How would we build this?

Here is the high level design:

* Store the data for each retention period into a bucket.
* Define tasks that downsample data from one bucket to the next using a naming convention.
* Flux logic that can examine the range of a query and return data across each of the buckets based on the naming convention.

Given this basic design the parts we are missing are two-fold:

* A system to store/retrieve configuration for rollups
* A Flux package to query rollups

### Rollup configuration

We need to decide if we are going to generate the downsample tasks for the user or if we expect them to build them manually.

If users create their own downsample tasks then all we need to know is which buckets contain data for which time ranges.
We _might_ even be able to time range information from the retention policies on the buckets.
This approach keeps the implementation simple and flexible while pushing complexity onto the user.
Specifically users will need to ensure certain assumptions hold about the data within each bucket, like having the same shape etc.

If we create the downsample tasks for the users we will be restricted in which downsample methods they can use but would mean less of a burden on users to setup.
Users would need to specify the following:

* The bucket containing the raw data
* An aggregation method to apply i.e. mean, sum count.
* Any number of _archives_ which need a retention period and data resolution.

An example archive definition looks like (using Graphite's syntax) `10m:1h,1h:1d` which reads as keep data at 10 minute intervals for 1 hour, then keep data at 1 hour for 1 day.

With this information we could automatically generate the Flux tasks to do the downsampling between the buckets for the user.

### Flux package

The Flux package generally needs to know the names of the buckets for a given rollup and the time ranges associated with each bucket.
With that information Flux can determine how to query across each bucket as needed.

There are some details we need to consider in how the Flux package works such that we don't loose access to pushdowns and other optimizations.

For example a query that does more with the data beyond just query it will still want to take advantage of pushdowns.

```
import "influxdata/influxdb/rollups"


rollups.from(name: "mydata", start: -1y, stop: now())
    |> filter()
    |> sum()
```

We need to be careful how we _rewrite_ this query under the hood.
A naive approach would look like this:

```
union(tables:[
    from(bucket: "mydata_rollup_0") |> filter(),
    from(bucket: "mydata_rollup_1") |> filter(),
])
    |> sum()
```

However in current Flux that query would no longer be able to pushdown the `sum()` operation because of the `union` in between the `from` and `sum`.
A better way to rewrite the query would be like this.

```
union(tables:[
    from(bucket: "mydata_rollup_0") |> filter() |> sum(),
    from(bucket: "mydata_rollup_1") |> filter() |> sum(),
])
```

This way we do the union on the result of summing each bucket.
We will need to explore if this kind of rewrite inside the Flux planner is possible to do.


## What about UI support?

We will need UI support for defining rollup configurations.
We might want to create a mode for the query editor that knows about `rollups` and can generate `rollup.from()` calls instead of normal `from()` calls.
However that functionally would not be necessary in extracting value from `rollups` generally and as such can be built later if at all.

## What are the limitations of this approach?


Rollups generally are limited in the kinds of aggregations they can perform.
This is because the aggregation needs to statistically valid even after being applied to each level in the rollup hierarchy.
Graphite only allows mean, sum, min, max, and last. We might be able to provide a few more but would require very careful consideration.

The data in each bucket needs to have the same shape, otherwise trying to union the data back together will cause errors and or very confusing results.
This means that you cannot reduce the cardinality of the data within a single rollup. Each bucket within the rollup will have the same cardinality.

## Why not make buckets multi retention?

Another approach would be to push down the rollup logic into the bucket itself which would make querying the data exactly the same as query data in a normal bucket.
This approach has one large drawback: the implementation of rollups at the storage layer practically means a rewrite of TSM and/or IOx as neither storage system is designed as a round robin database.
A round robin database is fundamentally different from the architecture of both TSM and IOx. 

Additionally the usage of a Flux package for rollups means that querying rollups is the same level of effort as query a normal bucket.

## What next?


Given that this design requires building new API surface area for both the HTTP API and the Flux API it will need a more thorough design process than what has been presented here.
However knowing the direction we want to go let us make small steps in this direction.
We already have a set of tasks that downsample data for a client's use case.
What remains is a way to query the data across the various buckets given the time range.

There are a few ways forward:

1. Spike on a Flux rollups package that is hardcoded to work for a single client.
2. Use invokable scripts to encapsulate the complex queries that union the data and share the logic for picking which script to call.
3. Build Flux modules and use them to build a client specific rollup package.


The disadvantage to the first approach is that we have to hardcode some of client's rollup configuration directly into Flux's standard library.
This will mean its fragile and inflexible as a new Flux release will be required if they want to change something about their rollups.
However this approach allows us to prove out if this is a viable path forward.

The disadvantage of the second approach is that the logic that we really want to hide is the logic we can't hide, i.e. the logic to check the time range.
However if copy/paste is not a big deterrent it would mean we could deliver a working feature quickly.

Finally the third approach likely has the longest timeline but doesn't have the draw backs of the other two approaches.
It would allow us to build a `rollups` package that works for a client that they control and therefore can modify without needing to wait for a Flux release.
Additionally it will be able to hide the time range picking logic from the user and they would be able to do `rollups.from()`.

