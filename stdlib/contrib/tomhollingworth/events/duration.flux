package events

// duration will calculate the duration between records
// for each record. The duration calculated is between
// the current record and the next. The last record will
// compare against either the stopColum (default: _stop)
// or a stop timestamp value.
//
// `timeColumn` - Optional string. Default '_time'. The value used to calculate duration
// `columnName` - Optional string. Default 'duration'. The name of the result column
// `stopColumn` - Optional string. Default '_stop'. The name of the column to compare the last record on
// `stop` - Optional Time. Use a fixed time to compare the last record against instead of stop column.
builtin duration : (<-tables: [A], ?unit: duration, ?timeColumn: string, ?columnName: string, ?stopColumn: string, ?stop: time) => [B] where A: Record, B: Record
