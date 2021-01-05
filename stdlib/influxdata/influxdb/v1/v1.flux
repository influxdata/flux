// V1 provides an API for working with an InfluxDB 1 instance
// >NOTE: Must functions in this package are now deprecated see influxdata/influxdb/schema.
package v1


import "influxdata/influxdb/schema"

// Json parses an InfluxDB 1.x json result into a table stream.
builtin json : (?json: string, ?file: string) => [A] where A: Record

// Databases returns the list of available databases, it has no parameters.
builtin databases : (
    ?org: string,
    ?orgID: string,
    ?host: string,
    ?token: string,
) => [{
    organizationID: string,
    databaseName: string,
    retentionPolicy: string,
    retentionPeriod: int,
    default: bool,
    bucketID: string,
}]

// Deprecated: See influxdata/influxdata/schema.fieldsAsCols
fieldsAsCols = schema.fieldsAsCols

// Deprecated: See influxdata/influxdata/schema.tagValues
tagValues = schema.tagValues

// Deprecated: See influxdata/influxdata/schema.measurementTagValues
measurementTagValues = schema.measurementTagValues

// Deprecated: See influxdata/influxdata/schema.tagKeys
tagKeys = schema.tagKeys

// Deprecated: See influxdata/influxdata/schema.measurementTagKeys
measurementTagKeys = schema.measurementTagKeys

// Deprecated: See influxdata/influxdata/schema.fieldKeys
fieldKeys = schema.fieldKeys

// Deprecated: See influxdata/influxdata/schema.measurementFieldKeys
measurementFieldKeys = schema.measurementFieldKeys

// Deprecated: See influxdata/influxdata/schema.measurements
measurements = schema.measurements
// Maintain backwards compatibility by mapping the functions into the schema package.
