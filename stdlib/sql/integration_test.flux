package integration_test

import "array"
import "sql"
import "testing"

mssqlDsn = "sqlserver://sa:fluX!234@localhost:1433?database=master"
mysqlDsn = "flux:flux@tcp(127.0.0.1:3306)/flux"
mariaDbDsn = "flux:flux@tcp(127.0.0.1:3307)/flux"
pgDsn = "postgresql://postgres@127.0.0.1:5432/postgres?sslmode=disable"

stanley = { name: "Stanley", age: 15 }
lucy = { name: "Lucy", age: 14 }
sophie = { name: "Sophie", age: 15 }

testcase integration_pg_read_from_seed {
    want = array.from(rows: [stanley, lucy])
    got = sql.from(
        driverName: "postgres",
        dataSourceName: pgDsn,
        query: "SELECT name, age FROM pets WHERE seeded = true"
    )
    testing.diff(
        got: got,
        want: want
    ) |> yield()
}

testcase integration_pg_read_from_nonseed {
    want = array.from(rows: [sophie])
    got = sql.from(
        driverName: "postgres",
        dataSourceName: pgDsn,
        query: "SELECT name, age FROM pets WHERE seeded = false"
    )
    testing.diff(
        got: got,
        want: want
    ) |> yield()
}

testcase integration_pg_write_to {
    array.from(rows: [sophie])
        |> sql.to(
           driverName: "postgres",
           dataSourceName: pgDsn,
           table: "pets",
           batchSize: 1)
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
    got = sql.from(
        driverName: "mysql",
        dataSourceName: mysqlDsn,
        query: "SELECT name, age FROM pets WHERE seeded = true"
    )
    testing.diff(
        got: got,
        want: want
    ) |> yield()
}

testcase integration_mysql_read_from_nonseed {
    want = array.from(rows: [sophie])
    got = sql.from(
        driverName: "mysql",
        dataSourceName: mysqlDsn,
        query: "SELECT name, age FROM pets WHERE seeded = false"
    )
    testing.diff(
        got: got,
        want: want
    ) |> yield()
}

testcase integration_mysql_write_to {
    array.from(rows: [sophie])
        |> sql.to(
           driverName: "mysql",
           dataSourceName: mysqlDsn,
           table: "pets",
           batchSize: 1)
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
    got = sql.from(
        driverName: "mysql",
        dataSourceName: mariaDbDsn,
        query: "SELECT name, age FROM pets WHERE seeded = true"
    )
    testing.diff(
        got: got,
        want: want
    ) |> yield()
}

testcase integration_mariadb_read_from_nonseed {
    want = array.from(rows: [sophie])
    got = sql.from(
        driverName: "mysql",
        dataSourceName: mariaDbDsn,
        query: "SELECT name, age FROM pets WHERE seeded = false"
    )
    testing.diff(
        got: got,
        want: want
    ) |> yield()
}

testcase integration_mariadb_write_to {
    array.from(rows: [sophie])
        |> sql.to(
           driverName: "mysql",
           dataSourceName: mariaDbDsn,
           table: "pets",
           batchSize: 1)
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
    got = sql.from(
        driverName: "sqlserver",
        dataSourceName: mssqlDsn,
        query: "SELECT name, age FROM pets WHERE seeded = 1"
    )
    testing.diff(
        got: got,
        want: want
    ) |> yield()
}

testcase integration_mssql_read_from_nonseed {
    want = array.from(rows: [sophie])
    got = sql.from(
        driverName: "sqlserver",
        dataSourceName: mssqlDsn,
        query: "SELECT name, age FROM pets WHERE seeded = 0"
    )
    testing.diff(
        got: got,
        want: want
    ) |> yield()
}

// n.b. selecting "mssql" as the driver name changes the behavior of the
// driver re: parameter binding, causing our `to()` implementation to break at
// runtime.
testcase integration_mssql_write_to {
    array.from(rows: [sophie])
        |> sql.to(
           driverName: "sqlserver",
           dataSourceName: mssqlDsn,
           table: "pets",
           batchSize: 1)
        // The array.from() will be returned and will cause the test to fail.
        // Filtering will mean the test can pass. Hopefully `sql.to()` will
        // error if there's a problem.
        |> filter(fn: (r) => false)
        // Without the yield, the flux script can "finish", closing the db
        // connection before the insert commits!
        |> yield()
}
