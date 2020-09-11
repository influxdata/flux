package sql

builtin from : (driverName: string, dataSourceName: string, query: string) => [A]
builtin to : (<-tables: [A], driverName: string, dataSourceName: string, table: string, ?batchSize: int) => [A]