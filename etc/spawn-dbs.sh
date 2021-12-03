#!/bin/bash

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


PG_SEED="
CREATE TABLE pets (
 id SERIAL PRIMARY KEY,
 name VARCHAR(20),
 age INT,
 -- When seeded is true, this indicates the rows were a part of the initial
 -- data load (prior to tests making changes).
 seeded BOOL NOT NULL DEFAULT false
);
INSERT INTO pets (name, age, seeded)
VALUES
 ('Stanley', 15, true),
 ('Lucy', 14, true)
;"

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
;"

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
;"

VERTICA_SEED="
CREATE TABLE pets (
 id IDENTITY(1, 1) PRIMARY KEY,
 name VARCHAR(20),
 age INT,
 seeded BOOLEAN NOT NULL DEFAULT false
);
-- Vertica doesn't seem to support inserting more than one record at a time?
INSERT INTO pets (name, age, seeded)
VALUES ('Stanley', 15, true);
INSERT INTO pets (name, age, seeded)
VALUES ('Lucy', 14, true);
"

# Cleanup previous runs.
docker rm -f "${PG_NAME}" "${MYSQL_NAME}" "${MARIADB_NAME}" "${MS_NAME}" "${VERTICA_NAME}"
# XXX: if you want to shutdown all the containers (after you're done running
# integration tests), you should be able to run something like:
# ```
# docker ps --format '{{.Names}}' | grep flux-integ- | xargs docker rm -f
# ```

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

until docker exec "${VERTICA_NAME}" /opt/vertica/bin/vsql -l;  do
  >&2 echo "Vertica: Waiting"
  sleep 1
done
>&2 echo "Vertica: Ready"

until docker exec "${MS_NAME}" /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P 'fluX!234' -Q "EXIT"; do
  >&2 echo "MSSQL: Waiting"
  sleep 1
done
>&2 echo "MSSQL: Ready"

until docker exec "${MYSQL_NAME}" env MYSQL_PWD=flux mysql --database=flux --host=127.0.0.1 --user=flux --execute '\q'; do
  >&2 echo "MySQL: Waiting"
  sleep 1
done
>&2 echo "MySQL: Ready"

until docker exec "${MARIADB_NAME}" env MYSQL_PWD=flux mysql --database=flux --host=127.0.0.1 --user=flux --execute '\q'; do
  >&2 echo "MariaDB: Waiting"
  sleep 1
done
>&2 echo "MariaDB: Ready"

until docker exec "${PG_NAME}" psql -U postgres -c '\q'; do
  >&2 echo "Postgres: Waiting"
  sleep 1
done
>&2 echo "Postgres: Ready"

docker exec "${VERTICA_NAME}" /opt/vertica/bin/vsql -d flux -v AUTOCOMMIT=on -c "${VERTICA_SEED}"
docker exec "${PG_NAME}" psql -U postgres -c "${PG_SEED}"
docker exec "${MYSQL_NAME}" env MYSQL_PWD=flux mysql --database=flux --host=127.0.0.1 --user=flux --execute "${MYSQL_SEED}"
docker exec "${MARIADB_NAME}" env MYSQL_PWD=flux mysql --database=flux --host=127.0.0.1 --user=flux --execute "${MYSQL_SEED}"
docker exec "${MS_NAME}" /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P 'fluX!234' -Q "${MSSQL_SEED}";
