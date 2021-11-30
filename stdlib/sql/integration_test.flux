package integration_pg_test

import "array"
import "experimental"
import "sql"
import "testing"
import "profiler"
import "internal/debug"


mysqlDsn = "flux:flux@tcp(127.0.0.1:3306)/flux"
pgDsn = "postgresql://postgres@127.0.0.1:5432/postgres?sslmode=disable"

testcase integration_pg_read_from {
    want = array.from(rows: [
        {name: "Stanley", age: 15},
        {name: "Lucy", age: 14}]
    )
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

// FIXME: `sql.to` inside `experimental.chain()` seems to fire twice but shouldn't
testcase integration_pg_write_to {
    sophie = { name: "Sophie", age: 15 }
    got = experimental.chain(
        first: array.from(rows: [sophie])
                    |> sql.to(
                       driverName: "postgres",
                       dataSourceName: pgDsn,
                       table: "pets",
                       batchSize: 1
                    ),
        second: sql.from(
            driverName: "postgres",
            dataSourceName: pgDsn,
            query: "SELECT name, age FROM pets WHERE seeded = false LIMIT 1"
        )
    )
    want = array.from(rows: [sophie])
    testing.diff(want: want, got: got) |> yield()
}

testcase integration_mysql_read_from {
    want = array.from(rows: [
        {name: "Stanley", age: 15},
        {name: "Lucy", age: 14}]
    )
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

testcase integration_mysql_write_to {
    sophie = { name: "Sophie", age: 15 }

    got = experimental.chain(
        first: array.from(rows: [sophie])
                    |> sql.to(
                       driverName: "mysql",
                       dataSourceName: mysqlDsn,
                       table: "pets",
                       batchSize: 1
                    ),
        second: sql.from(
            driverName: "mysql",
            dataSourceName: mysqlDsn,
            query: "SELECT name, age FROM pets WHERE seeded = false LIMIT 1"
        )
    )
    want = array.from(rows: [sophie])
    testing.diff(want: want, got: got) |> yield()
}
