package integration_test

import "array"
import "sql"
import "testing"


mysqlDsn = "flux:flux@tcp(127.0.0.1:3306)/flux"
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

        |> filter(fn: (r) => false)
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
        |> filter(fn: (r) => false)
        |> yield()
}
