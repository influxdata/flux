package sql_test


import "array"
import "sql"
import "testing"

hdbDsn = "hdb://SYSTEM:fluX!234@localhost:39041"
mssqlDsn = "sqlserver://sa:fluX!234@localhost:1433?database=master"
mysqlDsn = "flux:flux@tcp(127.0.0.1:3306)/flux"
mariaDbDsn = "flux:flux@tcp(127.0.0.1:3307)/flux"
pgDsn = "postgresql://postgres@127.0.0.1:5432/postgres?sslmode=disable"
verticaDsn = "vertica://dbadmin@localhost:5433/flux"
sqliteDsn = "file:///tmp/flux-integ-tests-sqlite.db"

stanley = {name: "Stanley", age: 15}
lucy = {name: "Lucy", age: 14}
sophie = {name: "Sophie", age: 15}

// Some db engines will UPPERCASE table/column names when the identifiers are unquoted
STANLEY = {NAME: "Stanley", AGE: 15}
LUCY = {NAME: "Lucy", AGE: 14}
SOPHIE = {NAME: "Sophie", AGE: 15}

testcase integration_hdb_read_from_seed {
    want = array.from(rows: [STANLEY, LUCY])
    got = sql.from(driverName: "hdb", dataSourceName: hdbDsn, query: "SELECT name, age FROM pets WHERE seeded = true")

    testing.diff(got: got, want: want)
        |> yield()
}

testcase integration_hdb_read_from_nonseed {
    want = array.from(rows: [SOPHIE])
    got = sql.from(driverName: "hdb", dataSourceName: hdbDsn, query: "SELECT name, age FROM pets WHERE seeded = false")

    testing.diff(got: got, want: want)
        |> yield()
}

testcase integration_hdb_write_to {
    array.from(rows: [sophie])
        |> sql.to(
            driverName: "hdb",
            dataSourceName: hdbDsn,
            // n.b. if we don't UPPERCASE the table name here, the automatic table
            // create (if not exists) will incorrectly think it needs to create the
            // table. This will result in an error since the table already exists.
            table: "PETS",
            batchSize: 1,
        )
        // The array.from() will be returned and will cause the test to fail.
        // Filtering will mean the test can pass. Hopefully `sql.to()` will
        // error if there's a problem.
        |> filter(fn: (r) => false)
        // Without the yield, the flux script can "finish", closing the db
        // connection before the insert commits!
        |> yield()
}

testcase integration_pg_read_from_seed {
    want = array.from(rows: [stanley, lucy])
    got =
        sql.from(
            driverName: "postgres",
            dataSourceName: pgDsn,
            query: "SELECT name, age FROM \"pet info\" WHERE seeded = true",
        )

    testing.diff(got: got, want: want)
        |> yield()
}

testcase integration_pg_read_from_nonseed {
    want = array.from(rows: [sophie])
    got =
        sql.from(
            driverName: "postgres",
            dataSourceName: pgDsn,
            query: "SELECT name, age FROM \"pet info\" WHERE seeded = false",
        )

    testing.diff(got: got, want: want)
        |> yield()
}

testcase integration_pg_write_to {
    array.from(rows: [sophie])
        |> sql.to(driverName: "postgres", dataSourceName: pgDsn, table: "pet info", batchSize: 1)
        // The array.from() will be returned and will cause the test to fail.
        // Filtering will mean the test can pass. Hopefully `sql.to()` will
        // error if there's a problem.
        |> filter(fn: (r) => false)
        // Without the yield, the flux script can "finish", closing the db
        // connection before the insert commits!
        |> yield()
}

testcase integration_mysql_read_from_seed {
    want = array.from(rows: [stanley, lucy])
    got =
        sql.from(
            driverName: "mysql",
            dataSourceName: mysqlDsn,
            query: "SELECT name, age FROM `pet info` WHERE seeded = true",
        )

    testing.diff(got: got, want: want)
        |> yield()
}

testcase integration_mysql_read_from_nonseed {
    want = array.from(rows: [sophie])
    got =
        sql.from(
            driverName: "mysql",
            dataSourceName: mysqlDsn,
            query: "SELECT name, age FROM `pet info` WHERE seeded = false",
        )

    testing.diff(got: got, want: want)
        |> yield()
}

testcase integration_mysql_write_to {
    array.from(rows: [sophie])
        |> sql.to(driverName: "mysql", dataSourceName: mysqlDsn, table: "pet info", batchSize: 1)
        // The array.from() will be returned and will cause the test to fail.
        // Filtering will mean the test can pass. Hopefully `sql.to()` will
        // error if there's a problem.
        |> filter(fn: (r) => false)
        // Without the yield, the flux script can "finish", closing the db
        // connection before the insert commits!
        |> yield()
}

testcase integration_mariadb_read_from_seed {
    want = array.from(rows: [stanley, lucy])
    got =
        sql.from(
            driverName: "mysql",
            dataSourceName: mariaDbDsn,
            query: "SELECT name, age FROM `pet info` WHERE seeded = true",
        )

    testing.diff(got: got, want: want)
        |> yield()
}

testcase integration_mariadb_read_from_nonseed {
    want = array.from(rows: [sophie])
    got =
        sql.from(
            driverName: "mysql",
            dataSourceName: mariaDbDsn,
            query: "SELECT name, age FROM `pet info` WHERE seeded = false",
        )

    testing.diff(got: got, want: want)
        |> yield()
}

testcase integration_mariadb_write_to {
    array.from(rows: [sophie])
        |> sql.to(driverName: "mysql", dataSourceName: mariaDbDsn, table: "pet info", batchSize: 1)
        // The array.from() will be returned and will cause the test to fail.
        // Filtering will mean the test can pass. Hopefully `sql.to()` will
        // error if there's a problem.
        |> filter(fn: (r) => false)
        // Without the yield, the flux script can "finish", closing the db
        // connection before the insert commits!
        |> yield()
}

testcase integration_mssql_read_from_seed {
    want = array.from(rows: [stanley, lucy])
    got =
        sql.from(
            driverName: "sqlserver",
            dataSourceName: mssqlDsn,
            query: "SELECT name, age FROM \"pet info\" WHERE seeded = 1",
        )

    testing.diff(got: got, want: want)
        |> yield()
}

testcase integration_mssql_read_from_nonseed {
    want = array.from(rows: [sophie])
    got =
        sql.from(
            driverName: "sqlserver",
            dataSourceName: mssqlDsn,
            query: "SELECT name, age FROM \"pet info\" WHERE seeded = 0",
        )

    testing.diff(got: got, want: want)
        |> yield()
}

// n.b. selecting "mssql" as the driver name changes the behavior of the
// driver re: parameter binding, causing our `sql.to()` implementation to break
// at runtime. As such, we only technically support "sqlserver" though you can
// skate by with "mssql" if you only ever use `sql.from()` (which doesn't
// attempt to bind parameters!)
// <https://github.com/denisenkom/go-mssqldb#deprecated>
testcase integration_mssql_write_to
{
        array.from(rows: [sophie])
            |> sql.to(driverName: "sqlserver", dataSourceName: mssqlDsn, table: "pet info", batchSize: 1)
            // The array.from() will be returned and will cause the test to fail.
            // Filtering will mean the test can pass. Hopefully `sql.to()` will
            // error if there's a problem.
            |> filter(fn: (r) => false)
            // Without the yield, the flux script can "finish", closing the db
            // connection before the insert commits!
            |> yield()
    }

testcase integration_vertica_read_from_seed {
    want = array.from(rows: [stanley, lucy])
    got =
        sql.from(
            driverName: "vertica",
            dataSourceName: verticaDsn,
            query: "SELECT name, age FROM \"pet info\" where seeded = true",
        )

    testing.diff(got: got, want: want)
        |> yield()
}

testcase integration_vertica_read_from_nonseed {
    want = array.from(rows: [sophie])
    got =
        sql.from(
            driverName: "vertica",
            dataSourceName: verticaDsn,
            query: "SELECT name, age FROM \"pet info\" where seeded = false",
        )

    testing.diff(got: got, want: want)
        |> yield()
}

testcase integration_vertica_write_to {
    array.from(rows: [sophie])
        |> sql.to(driverName: "vertica", dataSourceName: verticaDsn, table: "pet info", batchSize: 1)
        // The array.from() will be returned and will cause the test to fail.
        // Filtering will mean the test can pass. Hopefully `sql.to()` will
        // error if there's a problem.
        |> filter(fn: (r) => false)
        // Without the yield, the flux script can "finish", closing the db
        // connection before the insert commits!
        |> yield()
}

testcase integration_sqlite_read_from_seed {
    want = array.from(rows: [stanley, lucy])
    got =
        sql.from(
            driverName: "sqlite3",
            dataSourceName: sqliteDsn,
            query: "SELECT name, age FROM \"pet info\" where seeded = true",
        )

    testing.diff(got: got, want: want)
        |> yield()
}

testcase integration_sqlite_read_from_nonseed {
    want = array.from(rows: [sophie])
    got =
        sql.from(
            driverName: "sqlite3",
            dataSourceName: sqliteDsn,
            query: "SELECT name, age FROM \"pet info\" where seeded = false",
        )

    testing.diff(got: got, want: want)
        |> yield()
}

testcase integration_sqlite_write_to {
    array.from(rows: [sophie])
        |> sql.to(driverName: "sqlite3", dataSourceName: sqliteDsn, table: "pet info", batchSize: 1)
        // The array.from() will be returned and will cause the test to fail.
        // Filtering will mean the test can pass. Hopefully `sql.to()` will
        // error if there's a problem.
        |> filter(fn: (r) => false)
        // Without the yield, the flux script can "finish", closing the db
        // connection before the insert commits!
        |> yield()
}
