package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/user"
	"path"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/ast"
	"github.com/prometheus/prometheus/promql"
)

func main() {
	influxURL := flag.String("influx-url", "http://localhost:9999/", "InfluxDB server URL.")
	influxBucket := flag.String("influx-bucket", "prometheus", "InfluxDB bucket name.")
	influxToken := flag.String("influx-token", "", "InfluxDB authentication token.")
	influxOrg := flag.String("influx-org", "prometheus", "The InfluxDB organization name.")
	promURL := flag.String("prometheus-url", "http://localhost:9090/", "Prometheus server URL.")
	promqlExpr := flag.String("query-expr", "up", "PromQL expression to query.")
	queryStart := flag.Int64("query-start", 0, "Query start timestamp in milliseconds.")
	queryEnd := flag.Int64("query-end", 0, "Query end timestamp in milliseconds.")
	queryRes := flag.Duration("query-resolution", 10*time.Second, "Query resolution in seconds.")

	flag.Parse()

	if *influxToken == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatalln("Error getting current user:", err)
		}
		tokenPath := path.Join(usr.HomeDir, ".influxdbv2/credentials")
		token, err := ioutil.ReadFile(tokenPath)
		if err != nil {
			log.Fatalf("Error reading auth token from %q (-influx-token was not set): %s", tokenPath, err)
		}
		*influxToken = string(token)
	}
	if *queryStart == 0 || *queryEnd == 0 {
		log.Fatalf("Must specify both -query-start and -query-end.")
	}

	startTime := time.Unix(0, *queryStart*1e6).UTC()
	endTime := time.Unix(0, *queryEnd*1e6).UTC()

	// Transpile PromQL into Flux.
	promqlNode, err := promql.ParseExpr(*promqlExpr)
	if err != nil {
		log.Fatalln("Error parsing PromQL expression:", err)
	}
	t := transpiler{
		bucket:     *influxBucket,
		start:      startTime,
		end:        endTime,
		resolution: *queryRes,
	}
	fluxFile, err := t.transpile(promqlNode)
	if err != nil {
		log.Fatalln("Error transpiling PromQL expression to Flux:", err)
	}

	// Query both Prometheus and InfluxDB, expect same result.
	promMatrix, err := queryPrometheus(*promURL, *promqlExpr, startTime, endTime, *queryRes)
	if err != nil {
		log.Fatalln("Error querying Prometheus:", err)
	}
	fmt.Printf("Running Flux query:\n============================================\n%s\n============================================\n\n", ast.Format(fluxFile))
	influxResult, err := queryInfluxDB(*influxURL, *influxOrg, *influxToken, *influxBucket, ast.Format(fluxFile))
	if err != nil {
		log.Fatalln("Error querying InfluxDB:", err)
	}
	// Make InfluxDB result comparable with the Prometheus result.
	influxMatrix, err := influxResultToPromMatrix(influxResult)
	if err != nil {
		log.Fatalln("Error processing InfluxDB results:", err)
	}

	if diff := cmp.Diff(promMatrix, influxMatrix); diff != "" {
		fmt.Println("FAILED! Prometheus and InfluxDB results differ:\n\n", diff)
		fmt.Println("Full results:")
		fmt.Println("=== InfluxDB results:\n", influxMatrix)
		fmt.Println("=== Prometheus results:\n", promMatrix)
	} else {
		fmt.Println("SUCCESS! Results equal.")
	}
}
