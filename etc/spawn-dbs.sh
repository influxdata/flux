#!/bin/bash

set -e

PREFIX=flux-integ-tests

PG_NAME="${PREFIX}-postgres"
PG_TAG="postgres:14"
MYSQL_NAME="${PREFIX}-mysql"
MYSQL_TAG="mysql:8"

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

# Cleanup previous runs.
docker rm -f "${PG_NAME}" "${MYSQL_NAME}"

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
  ${MYSQL_TAG} \
  --general-log=1 --general-log-file=/tmp/query.log

docker run --rm --detach \
  --name "${PG_NAME}" \
  --publish 5432:5432 \
  -e POSTGRES_HOST_AUTH_METHOD=trust \
  ${PG_TAG} \
  postgres -c log_statement=all

until docker exec "${MYSQL_NAME}" env MYSQL_PWD=flux mysql --database=flux --host=127.0.0.1 --user=flux --execute '\q'; do
  >&2 echo "MySQL: Waiting"
  sleep 1
done
echo "MySQL: Ready"

until docker exec "${PG_NAME}" psql -U postgres -c '\q'; do
  >&2 echo "Postgres: Waiting"
  sleep 1
done
echo "Postgres: Ready"

docker exec "${PG_NAME}" psql -U postgres -c "${PG_SEED}"
# XXX: query logs don't seem to show up in stdout even when this is set...
# docker exec "${MYSQL_NAME}" mysql --host=127.0.0.1 --password=flux --user=root --execute "SET GLOBAL general_log = 'ON';"
docker exec "${MYSQL_NAME}" env MYSQL_PWD=flux mysql --database=flux --host=127.0.0.1 --user=flux --execute "${MYSQL_SEED}"
