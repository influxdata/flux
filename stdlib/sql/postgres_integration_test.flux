package integration_pg_test

import "array"
import "experimental"
import "sql"
import "testing"

dsn = "postgresql://postgres@127.0.0.1:5432/postgres?sslmode=disable"

testcase integration_pg_read_from {
    want = array.from(rows: [
        {name: "Stanley", age: 15},
        {name: "Lucy", age: 14}]
    )
    got = sql.from(
        driverName: "postgres",
        dataSourceName: dsn,
        query: "SELECT name, age FROM pets where seeded = true"
    )
    testing.diff(
        got: got,
        want: want
    ) |> yield()
}

// FIXME: the `sql.to` seems to fire twice but shouldn't
testcase integration_pg_write_to {
    sophie = [{name: "Sophie", age: 15}]

    want = array.from(rows: sophie)

    got = experimental.chain(
        first: want |> sql.to(
            driverName: "postgres",
            dataSourceName: dsn,
            table: "pets",
            batchSize: 1
        ),
        second: sql.from(
            driverName: "postgres",
            dataSourceName: dsn,
            query: "SELECT name, age FROM pets where seeded = false"
        )
    )

    testing.diff(
        got: got,
        want: want
    ) |> yield()
}
