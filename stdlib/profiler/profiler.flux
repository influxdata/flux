// Package profiler provides performance profiling tools for Flux queries and operations.
//
// Profile results are returned as an extra result in the response named according to the profiles which are enabled.
//
// ## Metadata
// introduced: 0.82.0
// tags: optimize
//
package profiler


// enabledProfilers is a list of profilers to enable during execution.
//
// ## Available profilers
// - [query](#query)
// - [operator](#operator)
//
// ### query
// Provides statistics about the execution of an entire Flux script.
// When enabled, results include a table with the following columns:
//
// - **TotalDuration**: total query duration in nanoseconds.
// - **CompileDuration**: number of nanoseconds spent compiling the query.
// - **QueueDuration**: number of nanoseconds spent queueing.
// - **RequeueDuration**: number fo nanoseconds spent requeueing.
// - **PlanDuration**: number of nanoseconds spent planning the query.
// - **ExecuteDuration**: number of nanoseconds spent executing the query.
// - **Concurrency**: number of goroutines allocated to process the query.
// - **MaxAllocated**: maximum number of bytes the query allocated.
// - **TotalAllocated**: total number of bytes the query allocated (includes memory that was freed and then used again).
// - **RuntimeErrors**: error messages returned during query execution.
// - **flux/query-plan**: Flux query plan.
// - **influxdb/scanned-values**: value scanned by InfluxDB.
// - **influxdb/scanned-bytes**: number of bytes scanned by InfluxDB.
//
// ### operator
// The `operator` profiler output statistics about each operation in a query.
// [Operations executed in the storage tier](https://docs.influxdata.com/influxdb/cloud/query-data/optimize-queries/#start-queries-with-pushdown-functions)
// return as a single operation.
// When the `operator` profile is enabled, results include a table with a row
// for each operation and the following columns:
//
// - **Type:** operation type
// - **Label:** operation name
// - **Count:** total number of times the operation executed
// - **MinDuration:** minimum duration of the operation in nanoseconds
// - **MaxDuration:** maximum duration of the operation in nanoseconds
// - **DurationSum:** total duration of all operation executions in nanoseconds
// - **MeanDuration:** average duration of all operation executions in nanoseconds
//
// ## Examples
//
// ### Enable profilers in a query
// ```no_run
// import "profiler"
//
// option profiler.enabledProfilers = ["query", "operator"]
// ```
//
option enabledProfilers = [""]
