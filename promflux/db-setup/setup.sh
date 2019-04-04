#!/bin/bash

# This script assumes you have the following in your path:
#
# - go
# - influx
# - influxd (a version built against latest Flux changes in https://github.com/influxdata/flux/pull/999)
# - prometheus

set -e
set -x

# Change to directory of this script.
cd "${0%/*}"

rm -rf ./influx-data

influxd \
  --bolt-path=influx-data/influxd.bolt \
  --engine-path=influx-data/engine \
  --reporting-disabled &
INFLUX_PID=$!

trap shutdown_influxdb INT
trap shutdown_influxdb EXIT

function shutdown_influxdb() {
  kill $INFLUX_PID # Kill the backgrounded InfluxDB when we interrupt the script.
}

# Sleep a bit to make sure InfluxDB is operational before setting it up.
sleep 8

influx setup -b prometheus -f -o prometheus -p prometheus -u prometheus -r 9999

# Dump the Prometheus test TSDB into the InfluxDB line protocol, but ignore NaN values (not supported in InfluxDB yet).
go run prometheus/prom_tsdb_to_influxdb.go --tsdb.path=prometheus/data | grep -v f64=NaN > influx-data.txt

export INFLUX_TOKEN=$(influx auth find | grep active | awk '{ print $2 }')

# Store the test data in InfluxDB.
curl --globoff "http://localhost:9999/api/v2/write?bucket=prometheus&org=prometheus" -XPOST -H "Authorization: Token $INFLUX_TOKEN" -H "content-type: text/plain; charset=utf-8" --data-binary @influx-data.txt

prometheus --config.file=prometheus/prometheus-noscrape.yml --storage.tsdb.path=prometheus/data
