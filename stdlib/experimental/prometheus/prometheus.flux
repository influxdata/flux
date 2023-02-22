// Package prometheus provides tools for working with
// [Prometheus-formatted metrics](https://prometheus.io/docs/instrumenting/exposition_formats/).
//
// ## Metadata
// introduced: 0.50.0
// tags: prometheus
//
package prometheus


import "universe"
import "experimental"

// scrape scrapes Prometheus metrics from an HTTP-accessible endpoint and returns
// them as a stream of tables.
//
// ## Parameters
//
// - url: URL to scrape Prometheus metrics from.
//
// ## Examples
//
// ### Scrape InfluxDB OSS internal metrics
// ```no_run
//  import "experimental/prometheus"
//
//  prometheus.scrape(url: "http://localhost:8086/metrics")
// ```
//
// ## Metadata
// tags: inputs,prometheus
//
builtin scrape : (url: string) => stream[A] where A: Record

// histogramQuantile calculates a quantile on a set of Prometheus histogram values.
//
// This function supports [Prometheus metric parsing formats](https://docs.influxdata.com/influxdb/latest/reference/prometheus-metrics/)
// used by `prometheus.scrape()`, the Telegraf `promtheus` input plugin, and
// InfluxDB scrapers available in InfluxDB OSS.
//
// ## Parameters
//
// - quantile: Quantile to compute. Must be a float value between 0.0 and 1.0.
// - metricVersion: [Prometheus metric parsing format](https://docs.influxdata.com/influxdb/latest/reference/prometheus-metrics/)
//   used to parse queried Prometheus data.
//   Available versions are `1` and `2`.
//   Default is `2`.
// - tables: Input data. Default is piped-forward data (`<-`).
// - onNonmonotonic: Describes behavior when counts are not monotonically increasing
//   when sorted by upper bound. Default is `error`.
//
//   **Supported values**:
//   - **error**: Produce an error.
//   - **force**: Force bin counts to be monotonic by adding to each bin such that it
//     is equal to the next smaller bin.
//   - **drop**: When a nonmonotonic table is encountered, produce no output.
//
// ## Examples
//
// ### Compute the 0.99 quantile of a Prometheus histogram
// ```no_run
// import "experimental/prometheus"
//
// prometheus.scrape(url: "http://localhost:8086/metrics")
//     |> filter(fn: (r) => r._measurement == "prometheus")
//     |> filter(fn: (r) => r._field == "qc_all_duration_seconds")
//     |> prometheus.histogramQuantile(quantile: 0.99)
// ```
//
// ### Compute the 0.99 quantile of a Prometheus histogram parsed with metric version 1
// ```no_run
// import "experimental/prometheus"
//
// from(bucket: "example-bucket")
//     |> range(start: -1h)
//     |> filter(fn: (r) => r._measurement == "qc_all_duration_seconds")
//     |> prometheus.histogramQuantile(quantile: 0.99, metricVersion: 1)
// ```
//
// ## Metadata
// tags: transformations,aggregates,prometheus
//
histogramQuantile = (tables=<-, quantile, metricVersion=2, onNonmonotonic="error") => {
    _version2 = (onNonmonotonic) =>
        tables
            |> group(mode: "except", columns: ["le", "_value"])
            |> map(fn: (r) => ({r with le: float(v: r.le)}))
            |> universe.histogramQuantile(quantile: quantile, onNonmonotonic: onNonmonotonic)
            |> group(mode: "except", columns: ["le", "_value", "_time"])
            |> set(key: "quantile", value: string(v: quantile))
            |> experimental.group(columns: ["quantile"], mode: "extend")

    _version1 = (onNonmonotonic) =>
        tables
            |> filter(fn: (r) => r._field != "sum" and r._field != "count")
            |> map(fn: (r) => ({r with le: float(v: r._field)}))
            |> group(mode: "except", columns: ["_field", "le", "_value"])
            |> universe.histogramQuantile(quantile: quantile, onNonmonotonic: onNonmonotonic)
            |> group(mode: "except", columns: ["le", "_value", "_time"])
            |> set(key: "quantile", value: string(v: quantile))
            |> experimental.group(columns: ["quantile"], mode: "extend")

    output =
        if metricVersion == 2 then
            _version2(onNonmonotonic)
        else if metricVersion == 1 then
            _version1(onNonmonotonic)
        else
            universe.die(msg: "Invalid metricVersion. Available versions are 1 and 2.")

    return output
}
