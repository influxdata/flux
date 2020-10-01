package v1

import "influxdata/influxdb/schema"

// Json parses an InfluxDB 1.x json result into a table stream.
builtin json : (?json: string, ?file: string) => [A] where A: Record

// Databases returns the list of available databases, it has no parameters.
builtin databases : (?org: string, ?orgID: string, ?host: string, ?token: string) => [{organizationID: string , databaseName: string , retentionPolicy: string , retentionPeriod: int , default: bool , bucketID: string}]

// Maintain backwards compatibility by mapping the functions into the schema package.
fieldsAsCols = schema.fieldsAsCols
tagValues = schema.tagValues
measurementTagValues = schema.measurementTagValues
tagKeys = schema.tagKeys
measurementTagKeys = schema.measurementTagKeys
fieldKeys = schema.fieldKeys
measurementFieldKeys = schema.measurementFieldKeys
measurements = schema.measurements
