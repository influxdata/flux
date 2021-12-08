#!/bin/bash

# This script will try to spin up a docker container for a series of database
# engines which select flux tests will run against.
#
# Typically you will not invoke this script yourself.
# Instead use: `make test-flux-integration` from the repo root since the Make
# target knows specifically which tests to run.
#
# This script expects to find `docker` in your `PATH` and to be run as a user
# with the privs to create/destroy containers.
# Additionally, the script expects to find `sqlite3` (the cli sqlite client) in
# your `PATH`.
#
# As a diagnostic consideration, the docker containers are left running after
# the tests run to allow you to inspect the records and/or logs.
# These containers are destroyed and recreated with each invocation of this script.
#
# To shutdown all the containers (after you're done running
# integration tests), you should be able to do something like:
# ```
# docker ps --format '{{.Names}}' | grep flux-integ-tests | xargs docker rm -f
# ```


set -e

PREFIX=flux-integ-tests

PG_NAME="${PREFIX}-postgres"
PG_TAG="postgres:14"
MYSQL_NAME="${PREFIX}-mysql"
MYSQL_TAG="mysql:8"
MARIADB_NAME="${PREFIX}-mariadb"
MARIADB_TAG="mariadb:10"
MS_NAME="${PREFIX}-mssql"
MS_TAG="mcr.microsoft.com/mssql/server:2019-latest"
VERTICA_NAME="${PREFIX}-vertica"
VERTICA_TAG="vertica/vertica-ce:11.0.0-0"
SQLITE_DB_PATH="/tmp/${PREFIX}-sqlite.db"
# XXX: The SAP HANA docker image requires you to be logged in to pull (but it's
# free). We'll need some shared creds if we want to run this in CI.
# The image is also LARGE. 1.2G+.
# HDB_NAME="${PREFIX-hdb}"
# HDB_TAG="store/saplabs/hanaexpress:2.00.054.00.20210603.1"

# FIXME(onelson): get hdb into the test flow.
#  the hdb container doesn't seem to run without jumping through many hoops
#  including:
#  - custom kernel tuning.
#  - elevating the container security context.
#  Combine these hassles with the fact the documented procedures for
#  configuration don't seem to work as advertised _and_ the docker hub login
#  requirement, I'm punting on this one for now. When we get to refactoring the
#  SQL package, we'll have to look at the risk to hdb support carefully to
#  decide if we can afford to skip this or not.

# Seed Data
# ---------
#
# Each db engine will be seeded with an equivalent schema and sample data to help
# exercise each driver as it is exposed to Flux.
#
# 4 columns: id (auto inc pk), name (varchar), age (int), and seeded (bool or
# equivalent).
# The `seeded` column is used to separate initial data from new rows written
# during testing.

PG_SEED="
CREATE TABLE pets (
  id SERIAL PRIMARY KEY,
  name VARCHAR(20),
  age INT,
  seeded BOOL NOT NULL DEFAULT false
);
INSERT INTO pets (name, age, seeded)
VALUES
  ('Stanley', 15, true),
  ('Lucy', 14, true)
;
"

MYSQL_SEED="
CREATE TABLE pets (
  id SERIAL,
  name VARCHAR(20),
  age INT,
  seeded TINYINT(1) NOT NULL DEFAULT false,
  PRIMARY KEY (id)
);
INSERT INTO pets (name, age, seeded)
VALUES
  ('Stanley', 15, true),
  ('Lucy', 14, true)
;
"

MSSQL_SEED="
CREATE TABLE pets (
  id INT IDENTITY(1, 1) PRIMARY KEY,
  name VARCHAR(20),
  age INT,
  seeded BIT NOT NULL DEFAULT 0
);
INSERT INTO pets (name, age, seeded)
VALUES
  ('Stanley', 15, 1),
  ('Lucy', 14, 1)
;
"

VERTICA_SEED="
CREATE TABLE pets (
  id IDENTITY(1, 1) PRIMARY KEY,
  name VARCHAR(20),
  age INT,
  seeded BOOLEAN NOT NULL DEFAULT false
);
-- Vertica doesn't seem to support inserting more than one record at a time?
INSERT INTO pets (name, age, seeded) VALUES ('Stanley', 15, true);
INSERT INTO pets (name, age, seeded) VALUES ('Lucy', 14, true);
"

SQLITE_SEED="
CREATE TABLE pets (
  id INT PRIMARY KEY,
  name VARCHAR(20),
  age INT,
  seeded BOOLEAN NOT NULL DEFAULT false
);
INSERT INTO pets (name, age, seeded)
VALUES
  ('Stanley', 15, true),
  ('Lucy', 14, true);
"

# Cleanup previous runs (just in case).
echo "Cleaning up prior db data..."
rm -f "$SQLITE_DB_PATH"
docker rm -f "${PG_NAME}" "${MYSQL_NAME}" "${MARIADB_NAME}" "${MS_NAME}" "${VERTICA_NAME}"

# mysql is sort of annoying when it comes to logging so to look at the query log,
# you'll probably want to either use `docker cp` to get a copy of `/tmp/query.log`
# out of the container, or `docker exec ${MYSQL_NAME} cat /tmp/query.log` and
# redirect the output to a host-local file.
docker run --rm --detach \
  --name "${MYSQL_NAME}" \
  --publish 3306:3306 \
  -e MYSQL_USER=flux \
  -e MYSQL_ROOT_PASSWORD=flux \
  -e MYSQL_PASSWORD=flux \
  -e MYSQL_DATABASE=flux \
  "${MYSQL_TAG}" \
  --general-log=1 --general-log-file=/tmp/query.log

docker run --rm --detach \
  --name "${MARIADB_NAME}" \
  --publish 3307:3306 \
  -e MARIADB_USER=flux \
  -e MARIADB_ROOT_PASSWORD=flux \
  -e MARIADB_PASSWORD=flux \
  -e MARIADB_DATABASE=flux \
  "${MARIADB_TAG}" \
  --general-log=1 --general-log-file=/tmp/query.log

docker run --rm --detach \
  --name "${PG_NAME}" \
  --publish 5432:5432 \
  -e POSTGRES_HOST_AUTH_METHOD=trust \
  "${PG_TAG}" \
  postgres -c log_statement=all

# To look at the query log for MSSQL, try something like the following:
# ```
# docker exec -it flux-integ-tests-mssql /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P 'fluX!234' -Q 'SELECT TOP(100) t.TEXT FROM sys.dm_exec_query_stats s CROSS APPLY sys.dm_exec_sql_text(s.sql_handle) t ORDER BY s.last_execution_time'
# ```
docker run --rm --detach \
  --name "${MS_NAME}" \
  --publish 1433:1433 \
  -e ACCEPT_EULA=Y \
  -e 'SA_PASSWORD=fluX!234' \
  -e MSSQL_PID=Developer \
  "${MS_TAG}"

docker run --rm --detach \
  --name "${VERTICA_NAME}" \
  --publish 5433:5433 \
  -e VERTICA_DB_NAME=flux \
  "${VERTICA_TAG}"

function wait_for () {
  name="${1}"
  cmd="${2}"
  until eval "${cmd}";  do
    >&2 echo "${name}: Waiting"
    sleep 1
  done
  >&2 echo "${name}: Ready"
}

wait_for "MariaDB" "docker exec ${MARIADB_NAME} env MYSQL_PWD=flux mysql --database=flux --host=127.0.0.1 --user=flux --execute '\q'"
docker exec "${MARIADB_NAME}" env MYSQL_PWD=flux mysql --database=flux --host=127.0.0.1 --user=flux --execute "${MYSQL_SEED}"

wait_for "MSSQL" "docker exec ${MS_NAME} /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P 'fluX!234' -Q 'EXIT'"
docker exec "${MS_NAME}" /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P 'fluX!234' -Q "${MSSQL_SEED}";

wait_for "MySQL" "docker exec ${MYSQL_NAME} env MYSQL_PWD=flux mysql --database=flux --host=127.0.0.1 --user=flux --execute '\q'"
docker exec "${MYSQL_NAME}" env MYSQL_PWD=flux mysql --database=flux --host=127.0.0.1 --user=flux --execute "${MYSQL_SEED}"

wait_for "Postgres" "docker exec ${PG_NAME} psql -U postgres -c '\q'"
docker exec "${PG_NAME}" psql -U postgres -c "${PG_SEED}"

wait_for "Vertica" "docker exec ${VERTICA_NAME} /opt/vertica/bin/vsql -l"
docker exec "${VERTICA_NAME}" /opt/vertica/bin/vsql -d flux -v AUTOCOMMIT=on -c "${VERTICA_SEED}"

sqlite3 "${SQLITE_DB_PATH}" "${SQLITE_SEED}"
