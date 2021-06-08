// `profiler`
//
// Profiler exposes an API to profile queries. Profile results are returned as
// an extra result in the response named according to the profiles which are enabled.
package profiler


// `enabledProfilers` sets a list of profilers that should be enabled during execution.
//
// There are two profilers available: the query profiler and the operator profiler.
//
// - The `query` profiler measures time spent in various phases of query execution
// - The `operator` profiler measures time spent in each operator of the query
//
// # Enabling the profilers
//
// Add the following lines to your flux query to see profiler results in the output:
// 
// ```
// import "profiler"
// option profiler.enabledProfilers = ["query", "operator"]
// ```
option enabledProfilers = [""]
