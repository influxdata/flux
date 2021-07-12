// Package sample provides functions for downloading and ouputting InfluxDB sample datasets.
package sample


import "array"
import "dict"
import "experimental/csv"

sets = [
    "airSensor": {
        url: "https://raw.githubusercontent.com/influxdata/influxdb2-sample-data/master/air-sensor-data/air-sensor-data-annotated.csv",
        desc: "Simulated office building air sensor data with temperature, humidity, and carbon monoxide metrics. Data is updated approximately every 5m.",
        size: "~600 KB",
    },
    "birdMigration": {
        url: "https://raw.githubusercontent.com/influxdata/influxdb2-sample-data/master/bird-migration-data/bird-migration.csv",
        desc: "2019 African bird migration data from the \"Movebank: Animal Tracking\" dataset. Contains geotemporal data between 2019-01-01 and 2019-12-31.",
        size: "~1.25 MB",
    },
    "usgs": {
        url: "https://raw.githubusercontent.com/influxdata/influxdb2-sample-data/master/usgs-earthquake-data/all_week-annotated.csv",
        desc: "USGS earthquake data from the last week. Contains geotemporal data collected from USGS seismic sensors around the world. Data is updated approximately every 5m.",
        size: "~5 MB",
    },
]

_setInfo = (set) => {
    _setDict = dict.get(dict: sets, key: set, default: {url: "", desc: "", size: ""})

    return {name: set, description: _setDict.desc, url: _setDict.url, size: _setDict.size}
}

// data downloads a specified InfluxDB sample dataset.
//
// ## Parameters
//
// - `set` is the sample data set to download and output. Valid datasets:
//    - **airSensor**: Simulated temperature, humidity, and CO data from and office building.
//    - **birdMigration**: 2019 African bird migration data from [Movebank: Animal Tracking](https://www.kaggle.com/pulkit8595/movebank-animal-tracking).
//    - **usgs**: USGS earthquake data from the last week.
//
// ## Load InfluxDB sample data
//
// ```
// import "influxdata/influxdb/sample"
//
// sample.data(set: "airSensor")
// ```
//
data = (set) => {
    setInfo = _setInfo(set: set)

    url = if setInfo.url == "" then
        die(msg: "Invalid sample data set. Use sample.list to view available datasets.")
    else
        setInfo.url

    return csv.from(url)
}

// list outputs information about available InfluxDB sample datasets.
//
// ## List available InfluxDB sample datasets
//
// ```
// import "influxdata/influxdb/sample"
//
// sample.list()
// ```
//
list = () => array.from(
    rows: [
        _setInfo(set: "airSensor"),
        _setInfo(set: "birdMigration"),
        _setInfo(set: "usgs"),
    ],
)
