package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/user"
	"path"
	"reflect"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/prometheus/prometheus/promql"
)

func main() {
	influxURL := flag.String("influx-url", "http://localhost:9999/", "InfluxDB server URL.")
	influxBucket := flag.String("influx-bucket", "prometheus", "InfluxDB bucket name.")
	influxToken := flag.String("influx-token", "", "InfluxDB authentication token.")
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

	promqlNode, err := promql.ParseExpr(*promqlExpr)
	if err != nil {
		log.Fatalln("Error parsing PromQL expression:", err)
	}
	fluxNode, err := transpile(*influxBucket, promqlNode, startTime, endTime, *queryRes)
	if err != nil {
		log.Fatalln("Error transpiling PromQL expression to Flux:", err)
	}

	promMatrix, err := queryPrometheus(*promURL, *promqlExpr, startTime, endTime, *queryRes)
	if err != nil {
		log.Fatalln("Error querying Prometheus:", err)
	}
	influxResult, err := queryInfluxDB(*influxURL, *influxBucket, *influxToken, ast.Format(fluxNode))
	if err != nil {
		log.Fatalln("Error querying InfluxDB:", err)
	}
	influxMatrix, err := influxResultToPromMatrix(influxResult)
	if err != nil {
		log.Fatalln("Error processing InfluxDB results:", err)
	}

	fmt.Println("======== PROMETHEUS RESULTS:")
	fmt.Println(promMatrix)

	fmt.Println("======== INFLUXDB RESULTS:")
	fmt.Println(influxMatrix)

	fmt.Println("======== RESULTS EQUAL:", reflect.DeepEqual(influxMatrix, promMatrix))
}
