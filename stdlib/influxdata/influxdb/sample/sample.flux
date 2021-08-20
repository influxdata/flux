// Package sample provides functions for downloading and ouputting InfluxDB sample datasets.
package sample


import "array"
import "dict"
import "experimental/csv"

sets = [
    "airSensor": {
        url: "https://influx-testdata.s3.amazonaws.com/air-sensor-data-annotated.csv",
        desc: "Simulated office building air sensor data with temperature, humidity, and carbon monoxide metrics. Data is updated approximately every 15m.",
        size: "~600 KB",
        type: "live",
    },
    "birdMigration": {
        url: "https://influx-testdata.s3.amazonaws.com/bird-migration.csv",
        desc: "2019 African bird migration data from the \"Movebank: Animal Tracking\" dataset. Contains geotemporal data between 2019-01-01 and 2019-12-31.",
        size: "~1.2 MB",
        type: "static",
    },
    "machineProduction": {
        url: "https://influx-testdata.s3.amazonaws.com/machine-production.csv",
        desc: "States and metrics reported from four automated grinding wheel stations on a production line. Contains data from 2021-08-01.",
        size: "~11.9 MB",
        type: "static",
    },
    "noaa": {
        url: "https://influx-testdata.s3.amazonaws.com/noaa-ndbc-latest-observations-annotated.csv",
        desc: "Latest observations from the NOAA National Data Buoy Center (NDBC). Contains only the most recent observations (no historical data). Data is updated approximately every 15m.",
        size: "~1.3 MB",
        type: "live",
    },
    "noaaWater": {
        url: "https://influx-testdata.s3.amazonaws.com/noaa.csv",
        desc: "Water level observations from two stations reported by the NOAA Center for Operational Oceanographic Products and Services. Contains data between 2019-08-17 and 2019-09-17.",
        size: "~10.3 MB",
        type: "static",
    },
    "usgs": {
        url: "https://influx-testdata.s3.amazonaws.com/usgs-earthquake-all-week-annotated.csv",
        desc: "USGS earthquake data from the last week. Contains geotemporal data collected from USGS seismic sensors around the world. Data is updated approximately every 15m.",
        size: "~6 MB",
        type: "live",
    },
]

_setInfo = (set) => {
    _setDict = dict.get(dict: sets, key: set, default: {url: "", desc: "", size: "", type: ""})

    return {name: set, description: _setDict.desc, url: _setDict.url, size: _setDict.size, type: _setDict.type}
}

// data downloads a specified InfluxDB sample dataset.
//
// ## Parameters
//
// - `set` is the sample data set to download and output. Valid datasets:
//    - **airSensor**: Simulated temperature, humidity, and CO data from an office building.
//    - **birdMigration**: 2019 African bird migration data from [Movebank: Animal Tracking](https://www.kaggle.com/pulkit8595/movebank-animal-tracking).
//    - **noaa**: Latest observations from the [NOAA National Data Buoy Center (NDBC)](https://www.ndbc.noaa.gov/).
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

    return csv.from(url: url)
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
        _setInfo(set: "noaa"),
        _setInfo(set: "usgs"),
    ],
)

// alignToNow shifts time values in input data to align the chronological last point to _now_.
// When writing static historical sample datasets to **InfluxDB Cloud**, use alignToNow
// to avoid losing sample data with timestamps outside of the retention period
// associated with your InfluxDB Cloud account.
//
// ## Align sample data to now
//
// ```
// import "influxdata/influxdb/sample"
//
// sample.data(set: "birdMigration")
//    |> sample.alignToNow()
// ```

alignToNow = (tables=<-) => {
    _lastTime = (tables |> keep(columns: ["_time"]) |> last(column: "_time") |> findRecord(fn: (key) => true, idx: 0))._time
    _offset = int(v: now()) - int(v: _lastTime)
    _offsetDuration = duration(v: _offset)

    return tables |> timeShift(duration: _offsetDuration)
}
