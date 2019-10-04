package prometheus

// scrape enables scraping of a prometheus metrics endpoint and converts 
// that input into flux tables. Each metric is put into an individual flux 
// table, including each histogram and summary value.  
builtin scrape

