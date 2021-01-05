// Profiler exposes an API to profile queries.
// Profile results are returned as an extra result in the response named according to the profiles which are enabled.
package profiler


// EnabledProfilers sets a list of profilers that should be enabled during execution.
//
// Available profilers are:
//   * query - Profiles time spent in the various phases of query execution.
//   * operator - Profiles time spent in each operator of the query.
//
// Example:
//
//    import "profiler"
//    option profiler.enabledProfilers = ["query", "operator"]
//
option enabledProfilers = [""]
