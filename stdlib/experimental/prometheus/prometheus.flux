package prometheus
import "universe"

// scrape enables scraping of a prometheus metrics endpoint and converts 
// that input into flux tables. Each metric is put into an individual flux 
// table, including each histogram and summary value.  
builtin scrape

// histogramQuantile enables the user to calculate quantiles on a set of given values
// This function assumes that the given histogram data is being scraped or read from a 
// Prometheus source. 
histogramQuantile = (tables=<-, field, quantile) => 
    tables
        |> filter(fn: (r) => r._measurement == "prometheus" and r._field == field)
        |> group(mode: "except", columns: ["le", "_value"]) 
        |> map(fn:(r) => ({r with le: float(v:r.le)})) 
        |> universe.histogramQuantile(quantile: quantile)