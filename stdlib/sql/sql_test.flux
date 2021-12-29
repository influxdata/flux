package sql_test


import "array"
import "internal/debug"
import "sql"
import "testing"

hdbDsn = "hdb://SYSTEM:fluX!234@localhost:39041"
mssqlDsn = "sqlserver://sa:fluX!234@localhost:1433?database=master"
mysqlDsn = "flux:flux@tcp(127.0.0.1:3306)/flux"
mariaDbDsn = "flux:flux@tcp(127.0.0.1:3307)/flux"
pgDsn = "postgresql://postgres@127.0.0.1:5432/postgres?sslmode=disable"
verticaDsn = "vertica://dbadmin@localhost:5433/flux"
sqliteDsn = "file:///tmp/flux-integ-tests-sqlite.db"

// Some db engines will UPPERCASE table/column names when the identifiers are unquoted.
// At time of writing the DDL used to create the tables has unquoted column names, so
// for engines such as Snowflake and SAP HANA (hdb) we need to UPPERCASE these identifiers
// here so the quoted version matches the table definition.
stanley = {name: "Stanley", age: 15, "fav food": "chicken"}
STANLEY = {NAME: "Stanley", AGE: 15, "FAV FOOD": "chicken"}

lucy = {name: "Lucy", age: 14}
LUCY = {NAME: "Lucy", AGE: 14}

sophie = {name: "Sophie", age: 15, "fav food": "salmon"}
SOPHIE = {NAME: "Sophie", AGE: 15, "FAV FOOD": "salmon"}

nonseed_want = array.from(rows: [sophie])
NONSEED_WANT = array.from(rows: [SOPHIE])

// SQL Injection attempt simulation.
// Try to write the row (each crafted for a particular dialect) to a new table.
// Flux will try to create automatically and in the process, drop the seeded
// table, or not if the injection is mitigated.
// If the injection is successful, the "read_from_seed" tests should fail.
evil = array.from(rows: [{"x\" INT); drop table \"pet info\"; --": 1}])
EVIL = array.from(rows: [{"x\" INT); drop table \"PET INFO\"; --": 1}])
myevil = array.from(rows: [{"x` INT); drop table `pet info`; --": 1}])

testcase integration_hdb_read_from_seed {
    got =
        sql.from(
            driverName: "hdb",
            dataSourceName: hdbDsn,
            // n.b. we must explicitly UPPER CASE the table name here to match the DDL.
            query: "SELECT name, age, \"FAV FOOD\" FROM \"PET INFO\" WHERE seeded = true",
        )

    testing.diff(
        got: got,
        want:
            union(tables: [array.from(rows: [STANLEY]) |> debug.opaque(), array.from(rows: [LUCY]) |> debug.opaque()])
                |> sort(columns: ["AGE"], desc: true),
    )
        |> yield()
}

testcase integration_hdb_read_from_nonseed {
    got =
        sql.from(
            driverName: "hdb",
            dataSourceName: hdbDsn,
            // n.b. we must explicitly UPPER CASE the table name here to match the DDL.
            query: "SELECT name, age, \"FAV FOOD\" FROM \"PET INFO\" WHERE seeded = false",
        )

    testing.diff(got: got, want: NONSEED_WANT)
        |> yield()
}

testcase integration_hdb_injection {
    EVIL
        |> sql.to(driverName: "hdb", dataSourceName: hdbDsn, table: "injection attempt", batchSize: 1)
        |> filter(fn: (r) => false)
        |> yield()
}

testcase integration_hdb_write_to {
    NONSEED_WANT
        // n.b. our handling of identifiers for HDB mean the table name will
        // automatically be upper cased here (matching the UPPER CASEd name in the DDL).
        |> sql.to(driverName: "hdb", dataSourceName: hdbDsn, table: "pet info", batchSize: 1)
        // The array.from() will be returned and will cause the test to fail.
        // Filtering will mean the test can pass. Hopefully `sql.to()` will
        // error if there's a problem.
        |> filter(fn: (r) => false)
        // Without the yield, the flux script can "finish", closing the db
        // connection before the insert commits!
        |> yield()
}

testcase integration_pg_read_from_seed {
    got =
        sql.from(
            driverName: "postgres",
            dataSourceName: pgDsn,
            query: "SELECT name, age, \"fav food\" FROM \"pet info\" WHERE seeded = true",
        )

    testing.diff(
        got: got,
        want:
            union(tables: [array.from(rows: [stanley]) |> debug.opaque(), array.from(rows: [lucy]) |> debug.opaque()])
                |> sort(columns: ["age"], desc: true),
    )
        |> yield()
}

testcase integration_pg_read_from_nonseed {
    got =
        sql.from(
            driverName: "postgres",
            dataSourceName: pgDsn,
            query: "SELECT name, age, \"fav food\" FROM \"pet info\" WHERE seeded = false",
        )

    testing.diff(got: got, want: nonseed_want)
        |> yield()
}

testcase integration_pg_injection {
    evil
        |> sql.to(driverName: "postgres", dataSourceName: pgDsn, table: "injection attempt", batchSize: 1)
        |> filter(fn: (r) => false)
        |> yield()
}

testcase integration_pg_write_to {
    nonseed_want
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
    got =
        sql.from(
            driverName: "mysql",
            dataSourceName: mysqlDsn,
            query: "SELECT name, age, `fav food` FROM `pet info` WHERE seeded = true",
        )

    testing.diff(
        got: got,
        want:
            union(tables: [array.from(rows: [stanley]) |> debug.opaque(), array.from(rows: [lucy]) |> debug.opaque()])
                |> sort(columns: ["age"], desc: true),
    )
        |> yield()
}

testcase integration_mysql_read_from_nonseed {
    got =
        sql.from(
            driverName: "mysql",
            dataSourceName: mysqlDsn,
            query: "SELECT name, age, `fav food` FROM `pet info` WHERE seeded = false",
        )

    testing.diff(got: got, want: nonseed_want)
        |> yield()
}

testcase integration_mysql_injection {
    myevil
        |> sql.to(driverName: "mysql", dataSourceName: mysqlDsn, table: "injection attempt", batchSize: 1)
        |> filter(fn: (r) => false)
        |> yield()
}

testcase integration_mysql_write_to {
    nonseed_want
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
    got =
        sql.from(
            driverName: "mysql",
            dataSourceName: mariaDbDsn,
            query: "SELECT name, age, `fav food` FROM `pet info` WHERE seeded = true",
        )

    testing.diff(
        got: got,
        want:
            union(tables: [array.from(rows: [stanley]) |> debug.opaque(), array.from(rows: [lucy]) |> debug.opaque()])
                |> sort(columns: ["age"], desc: true),
    )
        |> yield()
}

testcase integration_mariadb_read_from_nonseed {
    got =
        sql.from(
            driverName: "mysql",
            dataSourceName: mariaDbDsn,
            query: "SELECT name, age, `fav food` FROM `pet info` WHERE seeded = false",
        )

    testing.diff(got: got, want: nonseed_want)
        |> yield()
}

testcase integration_mariadb_injection {
    myevil
        |> sql.to(driverName: "mysql", dataSourceName: mariaDbDsn, table: "injection attempt", batchSize: 1)
        |> filter(fn: (r) => false)
        |> yield()
}

testcase integration_mariadb_write_to {
    nonseed_want
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
    got =
        sql.from(
            driverName: "sqlserver",
            dataSourceName: mssqlDsn,
            query: "SELECT name, age, \"fav food\" FROM \"pet info\" WHERE seeded = 1",
        )

    testing.diff(
        got: got,
        want:
            union(tables: [array.from(rows: [stanley]) |> debug.opaque(), array.from(rows: [lucy]) |> debug.opaque()])
                |> sort(columns: ["age"], desc: true),
    )
        |> yield()
}

testcase integration_mssql_read_from_nonseed {
    got =
        sql.from(
            driverName: "sqlserver",
            dataSourceName: mssqlDsn,
            query: "SELECT name, age, \"fav food\" FROM \"pet info\" WHERE seeded = 0",
        )

    testing.diff(got: got, want: nonseed_want)
        |> yield()
}

testcase integration_mssql_injection {
    evil
        |> sql.to(driverName: "sqlserver", dataSourceName: mssqlDsn, table: "injection attempt", batchSize: 1)
        |> filter(fn: (r) => false)
        |> yield()
}

testcase integration_mssql_write_to {
    nonseed_want
        // n.b. selecting "mssql" as the driver name changes the behavior of the
        // driver re: parameter binding, causing our `sql.to()` implementation to break
        // at runtime. As such, we only technically support "sqlserver" though you can
        // skate by with "mssql" if you only ever use `sql.from()` (which doesn't
        // attempt to bind parameters!)
        // <https://github.com/denisenkom/go-mssqldb#deprecated>
        |> sql.to(driverName: "sqlserver", dataSourceName: mssqlDsn, table: "pet info", batchSize: 1)
        // The array.from() will be returned and will cause the test to fail.
        // Filtering will mean the test can pass. Hopefully `sql.to()` will
        // error if there's a problem.
        |> filter(fn: (r) => false)
        |> yield()
}

testcase integration_vertica_read_from_seed {
    got =
        sql.from(
            driverName: "vertica",
            dataSourceName: verticaDsn,
            query: "SELECT name, age, \"fav food\" FROM \"pet info\" where seeded = true",
        )

    testing.diff(
        got: got,
        want:
            union(tables: [array.from(rows: [stanley]) |> debug.opaque(), array.from(rows: [lucy]) |> debug.opaque()])
                |> sort(columns: ["age"], desc: true),
    )
        |> yield()
}

testcase integration_vertica_read_from_nonseed {
    got =
        sql.from(
            driverName: "vertica",
            dataSourceName: verticaDsn,
            query: "SELECT name, age, \"fav food\" FROM \"pet info\" where seeded = false",
        )

    testing.diff(got: got, want: nonseed_want)
        |> yield()
}

testcase integration_vertica_injection {
    evil
        |> sql.to(driverName: "vertica", dataSourceName: verticaDsn, table: "injection attempt", batchSize: 1)
        |> filter(fn: (r) => false)
        |> yield()
}

testcase integration_vertica_write_to {
    nonseed_want
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
    got =
        sql.from(
            driverName: "sqlite3",
            dataSourceName: sqliteDsn,
            query: "SELECT name, age, \"fav food\" FROM \"pet info\" where seeded = true",
        )

    testing.diff(
        got: got,
        want:
            union(tables: [array.from(rows: [stanley]) |> debug.opaque(), array.from(rows: [lucy]) |> debug.opaque()])
                |> sort(columns: ["age"], desc: true),
    )
        |> yield()
}

testcase integration_sqlite_read_from_nonseed {
    got =
        sql.from(
            driverName: "sqlite3",
            dataSourceName: sqliteDsn,
            query: "SELECT name, age, \"fav food\" FROM \"pet info\" where seeded = false",
        )

    testing.diff(got: got, want: nonseed_want)
        |> yield()
}

testcase integration_sqlite_injection {
    evil
        |> sql.to(driverName: "sqlite3", dataSourceName: sqliteDsn, table: "injection attempt", batchSize: 1)
        |> filter(fn: (r) => false)
        |> yield()
}

testcase integration_sqlite_write_to {
    nonseed_want
        |> sql.to(driverName: "sqlite3", dataSourceName: sqliteDsn, table: "pet info", batchSize: 1)
        // The array.from() will be returned and will cause the test to fail.
        // Filtering will mean the test can pass. Hopefully `sql.to()` will
        // error if there's a problem.
        |> filter(fn: (r) => false)
        // Without the yield, the flux script can "finish", closing the db
        // connection before the insert commits!
        |> yield()
}
