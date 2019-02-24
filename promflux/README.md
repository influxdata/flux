# PromQL -> Flux Transpiler Proof-of-Concept

This takes a PromQL query, transpiles it to Flux, and then runs it against both Prometheus and InfluxDB.

It compares the results and expects them to be the same. So far, only selecting a single metric name without any matchers is supported.

Start the test setup (brings up Prometheus & InfluxDB, both with identical test datasets):

**NOTE**: Read the script to see which binaries are expected to be in your path!

```bash
./db-setup/setup.sh
```

Run PromQL+Flux queries against it:

```bash
$ GO111MODULE=on go run . -influx-org=prom -query-expr="demo_cpu_usage_seconds_total" -query-start=1550781000000 -query-end=1550781900000 -query-resolution=10s
Running Flux query:
============================================
package main
//
option queryRangeStart = 2019-02-21T20:25:00Z
option queryRangeEnd = 2019-02-21T20:45:00Z
option queryResolution = 10000000000ns
option queryMetricName = "demo_cpu_usage_seconds_total"
option queryOffset = 0ns
option queryWindowCutoff = 2019-02-21T20:40:00Z

from(bucket: "prom")
	|> range(start: queryRangeStart, stop: queryRangeEnd)
	|> filter(fn: (r) =>
		(r._measurement == queryMetricName))
	|> window(every: queryResolution, period: 5m)
	|> filter(fn: (r) =>
		(r._start <= queryWindowCutoff))
	|> last()
	|> drop(columns: ["_time"])
	|> duplicate(column: "_stop", as: "_time")
	|> shift(shift: queryOffset)
============================================

SUCCESS! Results equal.
```

Or one with an offset:

```bash
GO111MODULE=on go run . -influx-org=prom -query-expr="demo_cpu_usage_seconds_total offset 3m" -query-start=1550781000000 -query-end=1550781900000 -query-resolution=10s
Running Flux query:
============================================
package main
//
option queryRangeStart = 2019-02-21T20:22:00Z
option queryRangeEnd = 2019-02-21T20:42:00Z
option queryResolution = 10000000000ns
option queryMetricName = "demo_cpu_usage_seconds_total"
option queryOffset = 180000000000ns
option queryWindowCutoff = 2019-02-21T20:37:00Z

from(bucket: "prom")
	|> range(start: queryRangeStart, stop: queryRangeEnd)
	|> filter(fn: (r) =>
		(r._measurement == queryMetricName))
	|> window(every: queryResolution, period: 5m)
	|> filter(fn: (r) =>
		(r._start <= queryWindowCutoff))
	|> last()
	|> drop(columns: ["_time"])
	|> duplicate(column: "_stop", as: "_time")
	|> shift(shift: queryOffset)
============================================

SUCCESS! Results equal.
```
