# PromQL -> Flux Transpilation Concerns

## Data Model

### Scalars

Prometheus (sub)expressions may result in scalar values. Scalar values are raw float64 values that do not have the concept of labels attached to them. PromQL differentiates them type-wise from single-series vectors that have an empty label set, so it's not sufficient to just have an empty set of columns on the Flux side to indicate a scalar result.

Compare (see the result item name, `{}` vs. `scalar`):

* [Vector-result](http://demo.robustperception.io:9090/graph?g0.range_input=1h&g0.expr=count(node_cpu_seconds_total)&g0.tab=1)
* [Scalar result](http://demo.robustperception.io:9090/graph?g0.range_input=1h&g0.expr=scalar(count(node_cpu_seconds_total))&g0.tab=1)

I'm not sure how to handle this yet, since final outputs of Flux always have to be tables. Maybe a separate special field / column to indicate the output type?

### Special float values

Prometheus supports storing `NaN`, `+Inf`, and `-Inf` sample values. InfluxDB issue for this: [https://github.com/influxdata/influxdb/issues/10490](https://github.com/influxdata/influxdb/issues/4089)

To better support storing of Prometheus data, it would be important to support not only `NaN` in general, but to be able to store exact bit representations of different `NaN` values. See this Wikipedia article about the different `NaN` types: https://en.wikipedia.org/wiki/NaN#Signaling_NaN

Prometheus uses a **normal** (non-signalling) `NaN` for e.g.:

- A quantile value for which there were no observations yet.
- The outcome of certain floating-point operations.

In addition, Prometheus uses a special **signalling** `NaN` to indicate that a series has gone stale. E.g. a series was present in a target in the last scrape, but now it's not. Or when a target disappears in an orderly way, we write out stale markers for all of its series. This is then used by Prometheus to immediately not return stale-marked series anymore for time-step evaluations after the staleness timestamp.

Here is the definition of the two `NaN`s used in Prometheus: https://github.com/prometheus/prometheus/blob/master/pkg/value/value.go#L20-L29

Prometheus's aggregator etc. functions also have specific behaviors when special float values are involved. We'd have to go through those separately once we have general support for those values.

### Special column names

Flux has special column names (`_start`, `_stop`, `_time`, `_measurement` and suffixed ones from joins). These are perfectly valid (though unusual label names) in PromQL.

We either have to escape those names somehow when they come from PromQL (and when they get stored coming from Prometheus) or live with those (uncommon) queries breaking.

## Dynamically calculated / time-step dependent scalar function arguments

Scalar function parameters in PromQL can be arbitrary expressions that may involve or depend on the current evaluation step timestamp.

Consider the following query in PromQL:

```
quantile(0.5, rate(node_cpu_seconds_total[5m]))
```

This calculates the 50th percentile CPU usage among the set of input series.

Instead of hard-coding `0.5` you could do something crazy like the following (not that it would be sensical):

```
quantile(scalar(avg(up)), rate(node_cpu_seconds_total[5m]))
```

At every resolution evaluation step of the range query, this would compute the average value of the `up` metric at that timestamp and use it as the quantile argument for `quantile()`.

Note that the scalar expression does not necessarily have to rely on persisted data, but it could be dynamically generated out of thin air (but still perhaps depend on the current timestamp) like:

```
quantile(time(), node_cpu_seconds_total)
```

How can we achieve the same with Flux?

## Arithmetic operators

PromQL has the `%` and `^` arithmetic operators, which are still missing in Flux.

## Vector matching / joining

When doing binary arithmetic between vectors of time series, PromQL has a lot of subtle vector-matching and label propagation behaviors:

### Implicit joins

PromQL automatically joins two sets of time series (instant vectors) on identical label sets on the LHS and RHS. E.g. the following calculates the average response time for every label combination that is on the involved series without having to specify `on` labels:

```
  rate(prometheus_http_request_duration_seconds_sum[5m])
/
  rate(prometheus_http_request_duration_seconds_count[5m])
```

Flux's `join()` does not support an implicit `on` on all columns yet.

### TODO

## Function behaviors

### PromQL-specific functions

Almost all of the functions at https://prometheus.io/docs/prometheus/latest/querying/functions/ have very specific behaviors (e.g. exact extrapolation behavior in `rate()` and related functions) that cannot be emulated with current Flux features. We probably will need to port them as is into a new `promql` Flux package.

### `percentile()` edge case behavior

Prometheus's `quantile()` can be *almost* emulated using Flux's `percentile()`, except that PromQL returns `-Inf` / `+Inf` if one chooses quantile values below `0` or above `1`, whereas Flux doesn't.

We could either ignore this (unlikely to come up in real use cases), change the behavior in Flux, or additionally add a PromQL-compatible version into Flux.

### `stddev` behavior

PromQL's `stddev` calculates the population standard deviation, Flux's `stddev` calculates the sample standard deviation, see https://www.khanacademy.org/math/statistics-probability/summarizing-quantitative-data/variance-standard-deviation-sample/a/population-and-sample-standard-deviation-review.

Probably just add a modifier to Flux's `stddev` that allows choosing the mode?

### No `stdvar` yet

Flux has `stddev()` (like PromQL), but no `stdvar()` yet. We could probably work around it via `stddev(...) * stddev(...)` (no `^2` yet), but might be nice to have a native function.
