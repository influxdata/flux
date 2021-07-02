// Package SQL provides tools for working with data in SQL
// databases such as:
// - Amazon Athena
// - Google BigQuery
// - Microsoft SQL Server
// - MySQL
// - PostgreSQL
// - Snowflake
// - SQLite
package sql


// from is a function that retrieves data from a SQL data source.
//
// ## Parameters
// - `driverName` is the driver used to connect to the SQL database.
//
//   The following drivers are available:
//    - awsathena
//    - bigquery
//    - mysql
//    - postgres
//    - snowflake
//    - sqlite3 - Does not work with InfluxDB OSS or InfluxDB Cloud
//    - sqlserver, mssql
//
// - `dataSourceName` is the data source name (DNS) or connection string used
//   to connect to the SQL database.
//
//   The string's form and structure depend on the driver used.
//
// - `query` is the query to run against the SQL database.
//
// ## Driver dataSourceName examples
//
// ```
// # Amazon Athena Driver DSN
// s3://myorgqueryresults/?accessID=AKIAJLO3F...&region=us-west-1&secretAccessKey=NnQ7MUMp9PYZsmD47c%2BSsXGOFsd%2F...
// s3://myorgqueryresults/?accessID=AKIAJLO3F...&db=dbname&missingAsDefault=false&missingAsEmptyString=false&region=us-west-1&secretAccessKey=NnQ7MUMp9PYZsmD47c%2BSsXGOFsd%2F...&WGRemoteCreation=false
//
// # MySQL Driver DSN
// username:password@tcp(localhost:3306)/dbname?param=value
//
// # Postgres Driver DSN
// postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full
//
// # Snowflake Driver DSNs
// username[:password]@accountname/dbname/schemaname?param1=value1&paramN=valueN
// username[:password]@accountname/dbname?param1=value1&paramN=valueN
// username[:password]@hostname:port/dbname/schemaname?account=<your_account>&param1=value1&paramN=valueN
//
// # SQLite Driver DSN
// file:/path/to/test.db?cache=shared&mode=ro
//
// # Microsoft SQL Server Driver DSNs
// sqlserver://username:password@localhost:1234?database=examplebdb
// server=localhost;user id=username;database=examplebdb;
// server=localhost;user id=username;database=examplebdb;azure auth=ENV
// server=localhost;user id=username;database=examplebdbr;azure tenant id=77e7d537;azure client id=58879ce8;azure client secret=0123456789
//
// # Google BigQuery DSNs
// bigquery://projectid/?param1=value&param2=value
// bigquery://projectid/location?param1=value&param2=value
// ```
//
// ## Query a MySQL database
//
// ```
// import "sql"
// import "influxdata/influxdb/secrets"
//
// username = secrets.get(key: "MYSQL_USER")
// password = secrets.get(key: "MYSQL_PASS")
//
// sql.from(
//  driverName: "mysql",
//  dataSourceName: "${username}:${password}@tcp(localhost:3306)/db",
//  query:"SELECT * FROM example_table"
// )
// ```
//
// ## Query a Postgres database
//
// ```
// import "sql"
// import "influxdata/influxdb/secrets"
//
// username = secrets.get(key: "POSTGRES_USER")
// password = secrets.get(key: "POSTGRES_PASS")
//
// sql.from(
//   driverName: "postgres",
//   dataSourceName: "postgresql://${username}:${password}@localhost",
//   query:"SELECT * FROM example_table"
// )
// ```
//
// ## Query a Snowflake database
//
// ```
// import "sql"
// import "influxdata/influxdb/secrets"
//
// username = secrets.get(key: "SNOWFLAKE_USER")
// password = secrets.get(key: "SNOWFLAKE_PASS")
// account = secrets.get(key: "SNOWFLAKE_ACCT")
//
// sql.from(
//   driverName: "snowflake",
//   dataSourceName: "${username}:${password}@${account}/db/exampleschema?warehouse=wh",
//   query: "SELECT * FROM example_table"
// )
// ```
//
// ## Query a SQLite database
//
// ```
// import "sql"
//
// sql.from(
//   driverName: "sqlite3",
//   dataSourceName: "file:/path/to/test.db?cache=shared&mode=ro",
//   query: "SELECT * FROM example_table"
// )
// ```
// InfluxDB OSS and InfluxDB Cloud do not have direct access to the local filesystem
// and cannot query SQLite data sources. Use the Flux REPL to query a SQLite data
// source on your local filesystem.
//
// ## Query an Amazon Athena database
//
// ```
// import "sql"
// import "influxdata/influxdb/secrets"
//
// region = us-west-1
// accessID = secrets.get(key: "ATHENA_ACCESS_ID")
// secretKey = secrets.get(key: "ATHENA_SECRET_KEY")
//
// sql.from(
//  driverName: "awsathena",
//  dataSourceName: "s3://myorgqueryresults/?accessID=${accessID}&region=${region}&secretAccessKey=${secretKey}",
//  query:"SELECT * FROM example_table"
// )
// ```
// # Athena connection strings
// To query an Amazon Athena database, use the following querry parameters in your Athena
// S3 connection string (DNS):
// * Required
// - *region - AWS region
// - *accessID - AWS IAM access ID
// - *SecretAccessKey - AWS IAM secret key
// - db - database name
// - WGRemoteCreation - controls workgroup and tag creation
// - missingAsDefault - replace missing data with default values
// - missingAsEmptyString - replace missing data with empty strings
//
// ## Query a SQL Server database
//
// ```
// import "sql"
// import "influxdata/influxdb/secrets"
//
// username = secrets.get(key: "SQLSERVER_USER")
// password = secrets.get(key: "SQLSERVER_PASS")
//
// sql.from(
//   driverName: "sqlserver",
//   dataSourceName: "sqlserver://${username}:${password}@localhost:1234?database=examplebdb",
//   query: "GO SELECT * FROM Example.Table"
// )
// ```
// # SQL Server ADO authentication
// Use one of the following methods to provide SQL Server authentication
// credentials as ActiveX Data Objects (ADO) connection string parameters:
//
// # Retrieve authentication credentials from environment variables
// ```
// azure auth=ENV
// ```
//
// # Retrieve authentication credentials from a file
// ```
// azure auth=C:\secure\azure.auth
// ```
// InfluxDB OSS and InfluxDB Cloud user interfaces do not provide access to the underlying
// filesystem and do not support reading credentials from a file. To retrieve SQL Server
// credentials from a file, execute the query in the Flux REPL on your local machine.
//
// # Specify authentication credentials in the connection string
// ```
// # Example of providing tenant ID, client ID, and client secret token
// azure tenant id=77...;azure client id=58...;azure client secret=0cf123..
// # Example of providing tenant ID, client ID, certificate path and certificate password
// azure tenant id=77...;azure client id=58...;azure certificate path=C:\secure\...;azure certificate password=xY...
// # Example of providing tenant ID, client ID, and Azure username and password
// azure tenant id=77...;azure client id=58...;azure username=some@myorg;azure password=a1...
// ```
//
// # Use a managed identity in an Azure VM
// ```
// azure auth=MSI
// ```
//
// ## Query a BigQuery database
//
// ```
// import "sql"
// import "influxdata/influxdb/secrets"
// projectID = secrets.get(key: "BIGQUERY_PROJECT_ID")
// apiKey = secrets.get(key: "BIGQUERY_APIKEY")
// sql.from(
//  driverName: "bigquery",
//  dataSourceName: "bigquery://${projectID}/?apiKey=${apiKey}",
//  query:"SELECT * FROM exampleTable"
// )
// ```
// # Common BigQuery URL parameters
// The Flux BigQuery Implementation uses the Google Cloud Go SDK. Provide your
// authentication credentials using one of the following methods:
//
// - The `GOOGLE_APPLICATION_CREDENTIALS` environment variable that identifies the
//   location of yur credential JSON file.
//
// - Provide your BigQuery API key using the apiKey URL parameters in your BigQuery DSN.
//
// # Example apiKey URL parameter
// ```
// bigquery://projectid/?apiKey=AIzaSyB6XK8IO5AzKZXoioQOVNTFYzbDBjY5hy4
// ```
//
// - Provide your base-64 encoded service account, refresh token, or JSON credentials
//   using the credentials URL parameter in your BigQuery DSN.
//
// # Example credential URL parameter
// ```
// bigquery://projectid/?credentials=eyJ0eXBlIjoiYXV0...
// ```
builtin from : (driverName: string, dataSourceName: string, query: string) => [A]

// to is a function that writes data to an SQL database.
//
// ## Parameters
// - `driverName` is the driver used to connect to the SQL database.
//
//   The following drivers are available:
//    - bigquery
//    - mysql
//    - postgres
//    - snowflake
//    - sqlite3 - Does not work with InfluxDB OSS or InfluxDB Cloud
//    - sqlserver, mssql
//
// sql.to does not support Amazon Athena.
//
// - `dataSourceName` is the data source name (DNS) or connection string used
//   to connect to the SQL database.
//
// - `table` is the destination table.
//
// - `batchSize` is the number of parameters or columns that can be queued within
//   each call to Exec. Defaults to 10000.
//
//   If writing to SQLite database, set the batchSize to 999 or less.
//
// ## Driver dataSourceName examples
//
// ```
// # Postgres Driver DSN
// postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full
// # MySQL Driver DSN
// username:password@tcp(localhost:3306)/dbname?param=value
//
// # Snowflake Driver DSNs
// username[:password]@accountname/dbname/schemaname?param1=value1&paramN=valueN
// username[:password]@accountname/dbname?param1=value1&paramN=valueN
// username[:password]@hostname:port/dbname/schemaname?account=<your_account>&param1=value1&paramN=valueN
//
// # SQLite Driver DSN
// file:/path/to/test.db?cache=shared&mode=rw
//
// # Microsoft SQL Server Driver DSNs
// sqlserver://username:password@localhost:1234?database=examplebdb
// server=localhost;user id=username;database=examplebdb;
// server=localhost;user id=username;database=examplebdb;azure auth=ENV
// server=localhost;user id=username;database=examplebdbr;azure tenant id=77e7d537;azure client id=58879ce8;azure client secret=0123456789
//
// # Google BigQuery DSNs
// bigquery://projectid/?param1=value&param2=value
// bigquery://projectid/location?param1=value&param2=value
// ```
//
// ## Write data to a MySQL database
//
// ```
// import "sql"
// import "influxdata/influxdb/secrets"
//
// username = secrets.get(key: "MYSQL_USER")
// password = secrets.get(key: "MYSQL_PASS")
//
// sql.to(
//   driverName: "mysql",
//   dataSourceName: "${username}:${password}@tcp(localhost:3306)/db",
//   table: "example_table"
// )
// ```
//
// ## Write data to a Postgres database
//
// ```
// import "sql"
// import "influxdata/influxdb/secrets"
//
// username = secrets.get(key: "POSTGRES_USER")
// password = secrets.get(key: "POSTGRES_PASS")
//
// sql.to(
//   driverName: "postgres",
//   dataSourceName: "postgresql://${username}:${password}@localhost",
//   table: "example_table"
// )
// ```
//
// ## Write data to a snowflake database
//
// ```
// import "sql"
// import "influxdata/influxdb/secrets"
//
// username = secrets.get(key: "SNOWFLAKE_USER")
// password = secrets.get(key: "SNOWFLAKE_PASS")
// account = secrets.get(key: "SNOWFLAKE_ACCT")
//
// sql.to(
//   driverName: "snowflake",
//   dataSourceName: "${username}:${password}@${account}/db/exampleschema?warehouse=wh",
//   table: "example_table"
// )
// ```
//
// ## Write data to an SQLite database
//
// ```
// import "sql"
//
// sql.to(
//   driverName: "sqlite3",
//   dataSourceName: "file:/path/to/test.db?cache=shared&mode=rw",
//   table: "example_table"
// )
// ```
// InfluxDB OSS and InfluxDB Cloud do not have direct access to the local
// filesystem and cannot write to SQLite data sources. Use the Flux REPL
// to write to an SQLite data source on your local filesystem.
//
// ## Write data to a SQL Server database
//
// ```
// import "sql"
// import "influxdata/influxdb/secrets"
//
// username = secrets.get(key: "SQLSERVER_USER")
// password = secrets.get(key: "SQLSERVER_PASS")
//
// sql.to(
//   driverName: "sqlserver",
//   dataSourceName: "sqlserver://${username}:${password}@localhost:1234?database=examplebdb",
//   table: "Example.Table"
// )
// ```
//
// # SQL Server ADO authentication
// Use one of the following methods to provide SQL Server authentication credentials as
// ActiveX Data Objects (ADO) connection string parameters:
//
// # Retrieve authentication credentials from environment variables
// ```
// azure auth=ENV
// ```
//
// # Retrieve authentication credentials from a file
// ```
// azure auth=C:\secure\azure.auth
// ```
// InfluxDB OSS and InfluxDB Cloud user interfaces do not provide access to the underlying file
// system and do not support reading credentials from a file. To retrieve SQL Server credentials
// from a file, execute the query in the Flux REPL on your local machine.
//
// # Specify authentication credentials in the connection string
// ```
// # Example of providing tenant ID, client ID, and client secret token
// azure tenant id=77...;azure client id=58...;azure client secret=0cf123..
//
// # Example of providing tenant ID, client ID, certificate path and certificate password
// azure tenant id=77...;azure client id=58...;azure certificate path=C:\secure\...;azure certificate password=xY...
//
// # Example of providing tenant ID, client ID, and Azure username and password
// azure tenant id=77...;azure client id=58...;azure username=some@myorg;azure password=a1...
// ```
//
// # Use a managed identity in an Azure VM
// ```
// azure auth=MSI
// ```
//
// ## Write to a BigQuery database
//
// ```
// import "sql"
// import "influxdata/influxdb/secrets"
// projectID = secrets.get(key: "BIGQUERY_PROJECT_ID")
// apiKey = secrets.get(key: "BIGQUERY_APIKEY")
// sql.to(
//  driverName: "bigquery",
//  dataSourceName: "bigquery://${projectID}/?apiKey=${apiKey}",
//  table:"exampleTable"
// )
// ```
//
// # Common BigQuery URL parameters
// - dataset - BigQuery dataset ID. When set, you can use unqualified table
//   names in queries.
//
// # BigQuery authentication parameters
// - The `GOOGLE_APPLICATION_CREDENTIALS` environment variable that identifies
//   the location of your credential JSON file.
//
// - Provide your BigQuery API key using the apiKey URL parameter in your
//   BigQuery DSN.
//
// # Example apiKey URL parameter
// ```
// bigquery://projectid/?apiKey=AIzaSyB6XK8IO5AzKZXoioQOVNTFYzbDBjY5hy4
// ```
//
// - Provide your base-64 encoded service account, refresh token, or JSON credentials
//   URL parameter in your BigQuery DSN.
//
// # Example credentials URL parameter
// ```
// bigquery://projectid/?credentials=eyJ0eXBlIjoiYXV0...
// ```
builtin to : (
    <-tables: [A],
    driverName: string,
    dataSourceName: string,
    table: string,
    ?batchSize: int,
) => [A]
