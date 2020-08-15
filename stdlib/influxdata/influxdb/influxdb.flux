package influxdb

builtin from : (?bucket: string, ?bucketID: string, ?org: string, ?orgID: string, ?host: string, ?token: string) => [{B with _measurement: string , _field: string , _time: time , _value: A}]
builtin to : (<-tables: [A], ?bucket: string, ?bucketID: string, ?org: string, ?orgID: string, ?host: string, ?token: string, ?timeColumn: string, ?measurementColumn: string, ?tagColumns: [string], ?fieldFn: (r: A) => B) => [A] where A: Record, B: Record
builtin buckets : (?org: string, ?orgID: string, ?host: string, ?token: string) => [{name: string , id: string , organizationID: string , retentionPolicy: string , retentionPeriod: int}]
