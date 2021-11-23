#!/bin/bash

set -e

PREFIX=flux-integ-tests

PG_NAME="${PREFIX}-postgres"
PG_TAG="postgres"

SEED="
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

# Cleanup in case of failed previous runs.
docker rm -f "${PG_NAME}"

docker run --rm --detach \
  --name "${PG_NAME}" \
  --publish 5432:5432 \
  -e POSTGRES_HOST_AUTH_METHOD=trust \
  ${PG_TAG} \
  postgres -c log_statement=all

until docker exec "${PREFIX}-postgres" psql -U postgres -c '\q'; do
  >&2 echo "Postgres: Waiting"
  sleep 1
done

echo "Postgres: Ready"

docker exec "${PG_NAME}" psql -U postgres -c "${SEED}"