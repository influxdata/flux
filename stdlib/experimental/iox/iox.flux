// Package iox provides functions for querying data from IOx.
package iox


// from reads from the selected bucket and measurement in an iox storage node.
//
// This function creates a source that reads data from IOx. Output data is
// "pivoted" on the time column and includes columns for each returned
// tag and field per time value.
//
// ## Parameters
// - bucket: IOx bucket to read data from.
// - measurement: Measurement to read data from.
builtin from : (bucket: string, measurement: string) => [{A with _time: time}] where A: Record
