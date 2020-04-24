package prometheus
import "universe" 

// scrape enables scraping of a prometheus metrics endpoint and converts 
// that input into flux tables. Each metric is put into an individual flux 
// table, including each histogram and summary value.  
builtin scrape

// scrape Opt1 is a prototype of the option 1 data model that is described in
// this issue: https://github.com/influxdata/idpe/issues/6696
builtin scrapeOpt1

// histogramQuantile enables the user to calculate quantiles on a set of given values
// This function assumes that the given histogram data is being scraped or read from a 
// Prometheus source. 
histogramQuantile = (tables=<-, quantile) => 
    tables
        |> filter(fn: (r) => r._measurement == "prometheus")
        |> group(mode: "except", columns: ["le", "_value", "_time"]) 
        |> map(fn:(r) => ({r with le: float(v:r.le)})) 
        |> universe.histogramQuantile(quantile: quantile)