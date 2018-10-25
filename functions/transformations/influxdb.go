package transformations

import (
	"github.com/influxdata/flux"
)

func init() {
	flux.RegisterBuiltIn("influxdb", influxdbBuiltIns)
}

var influxdbBuiltIns = `
_measurement = (m,table=<-) => table |> filter(fn:(r) => r._measurement == m)
_field = (f,table=<-) => table |> filter(fn:(r) => r._field == f)

_tagValues = (bucket, key, predicate=(r) => true) =>
	from(bucket:bucket)
	  |> range(start:-24h)
	  |> filter(fn: predicate)
	  |> group(by:[key])
	  |> distinct(column:key)
	  |> group(by:["_stop","_start"])

_measurementTagValues = (bucket, measurement) =>
	_tagValues(bucket:bucket, predicate:(r) => r._measurement == measurement)

_tagKeys = (bucket, predicate=(r) => true) =>
	from(bucket:bucket)
		|> range(start:-24h)
		|> filter(fn:predicate)
		|> keys()

_measurementTagKeys = (bucket, measurement) =>
	_tagKeys(bucket:bucket, predicate:(r) => r._measurement == measurement)

_measurements = (bucket) =>
	_tagValues(bucket:bucket, key:"_measurement")

// This object approximates a namespace for influxdb related helper functions
ifluxdb = {
	// I/O
	from: from,
	//to: to,

	// Filters
	field: _field,
	measurement: _measurement,

	// Meta
	measurements: _measurements,
	tagValues: _tagValues,
	measurementTagValues: _measurementTagValues,
	tagKeys: _tagKeys,
	measurementTagKeys: _measurementTagKeys,
}
`
