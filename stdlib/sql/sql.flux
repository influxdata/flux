// Package sql provides tools for working with data in SQL databases.
//
// ## SQL data source names
// The `sql` packages uses Go drivers to connect to SQL database.
// The data source name (DSN) (also known as connection string) is determined by
// the driver used.
//
// ```sh
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
//
// # SAP HANA driver DSN
// hdb://<user>:<password>@<host>:<port>?<connection-property>=<value>&<connection-property>=<value>&...
// hdb://<user>:<password>@<host>:<port>?DATABASENAME=<tenant-db-name>
// hdb://?KEY=<keyname>
//
// # Vertica driver DSN
// vertica://<user>:<password>@<host>:<port>/<database>?<queryArgs>
// ```
//
// ### AWS Athena connection strings
// To query an Amazon Athena database, use the following query parameters in
// your Athena S3 connection string (DNS):
//
// - **region**: _(Required)_ AWS region.
// - **accessID**: _(Required)_ AWS IAM access ID.
// - **SecretAccessKey**: _(Required)_ AWS IAM secret key.
// - **db**: Database name.
// - **WGRemoteCreation**: Controls workgroup and tag creation.
// - **missingAsDefault**: Replace missing data with default values.
// - **missingAsEmptyString**: Replace missing data with empty strings.
//
// ### Common BigQuery URL parameters
// The Flux BigQuery implementation uses the Google Cloud Go SDK. Provide your
// authentication credentials using one of the following methods:
//
// - The `GOOGLE_APPLICATION_CREDENTIALS` environment variable that identifies the
//   location of yur credential JSON file.
// - Provide your BigQuery credentials using the `credentials` URL parameters in your BigQuery DSN.
//
// #### BigQuery credential URL parameter
// Provide your base-64 encoded service account, refresh token, or JSON credentials
// using the credentials URL parameter in your BigQuery DSN.
//
// ##### BigQuery credential URL parameter
// ```txt
// bigquery://projectid/?credentials=eyJ0eXBlIjoiYXV0...
// ```
//
// ### SQL Server ADO authentication
// Use one of the following methods to provide SQL Server authentication
// credentials as ActiveX Data Objects (ADO) connection string parameters:
//
// #### Retrieve authentication credentials from environment variables
// ```txt
// azure auth=ENV
// ```
//
// #### Retrieve authentication credentials from a file
// ```txt
// azure auth=C:\secure\azure.auth
// ```
//
// **Note**: InfluxDB OSS and InfluxDB Cloud user interfaces do not provide access
// to the underlying filesystem and do not support reading credentials from a file.
// To retrieve SQL Server credentials from a file, execute the query in the Flux
// REPL on your local machine.
//
// #### Specify authentication credentials in the connection string
// ```sh
// # Example of providing tenant ID, client ID, and client secret token
// azure tenant id=77...;azure client id=58...;azure client secret=0cf123..
// # Example of providing tenant ID, client ID, certificate path and certificate password
// azure tenant id=77...;azure client id=58...;azure certificate path=C:\secure\...;azure certificate password=xY...
// # Example of providing tenant ID, client ID, and Azure username and password
// azure tenant id=77...;azure client id=58...;azure username=some@myorg;azure password=a1...
// ```
//
// #### Use a managed identity in an Azure VM
// ```txt
// azure auth=MSI
// ```
//
// ## Metadata
// introduced: 0.34.0
//
package sql


// from retrieves data from a SQL data source.
//
// ## Parameters
// - driverName: Driver to use to connect to the SQL database.
//
//   **Supported drivers**:
//   - awsathena
//   - bigquery
//   - hdb
//   - mysql
//   - postgres
//   - snowflake
//   - sqlite3 _(Does not work with InfluxDB OSS or InfluxDB Cloud)_
//   - sqlserver
//   - vertica, vertigo
//
// - dataSourceName: Data source name (DNS) or connection string used to connect
//   to the SQL database.
// - query: Query to run against the SQL database.
//
// ## Examples
// For examples and more information about each supported SQL database, see
// [Query SQL databases](https://docs.influxdata.com/flux/v0.x/query-data/sql/).
//
// ## Metadata
// tags: inputs,sql
//
builtin from : (driverName: string, dataSourceName: string, query: string) => stream[A]

// to writes data to an SQL database.
//
// ## Parameters
// - driverName: Driver used to connect to the SQL database.
//
//   **Supported drivers**:
//   - bigquery
//   - hdb
//   - mysql
//   - postgres
//   - snowflake
//   - sqlite3 _(Does not work with InfluxDB OSS or InfluxDB Cloud)_
//   - sqlserver
//   - vertica, vertigo
//
// - dataSourceName: Data source name (DNS) or connection string used
//   to connect to the SQL database.
// - table: Destination table.
// - batchSize: Number of parameters or columns that can be queued within each
//   call to `Exec`. Default is `10000`.
//
//   If writing to SQLite database, set the batchSize to `999` or less.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// For examples and more information about each supported SQL database, see
// [Write to SQL databases](https://docs.influxdata.com/flux/v0.x/write-data/sql/).
//
// ## Metadata
// tags: outputs,sql
//
builtin to : (
        <-tables: stream[A],
        driverName: string,
        dataSourceName: string,
        table: string,
        ?batchSize: int,
    ) => stream[A]
