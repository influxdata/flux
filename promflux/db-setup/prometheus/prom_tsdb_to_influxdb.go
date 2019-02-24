package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/tsdb"
	"github.com/prometheus/tsdb/labels"
	"github.com/prometheus/tsdb/wal"
)

func main() {
	path := flag.String("tsdb.path", "./data", "The path to the Prometheus TSDB directory to dump.")
	flag.Parse()

	db, err := tsdb.Open(*path, nil, nil, &tsdb.Options{
		WALSegmentSize:    wal.DefaultSegmentSize,
		RetentionDuration: 99999 * 24 * 60 * 60 * 1000, // 99999 days in milliseconds
		BlockRanges:       tsdb.ExponentialBlockRanges(int64(2*time.Hour)/1e6, 3, 5),
	})
	if err != nil {
		log.Fatal(err)
	}
	dumpSamples(db)
}

func dumpSamples(db *tsdb.DB) {
	q, err := db.Querier(math.MinInt64, math.MaxInt64)
	if err != nil {
		log.Fatal(err)
	}

	ss, err := q.Select(labels.NewMustRegexpMatcher("__name__", ".*"))
	if err != nil {
		log.Fatal(err)
	}

	for ss.Next() {
		series := ss.At()
		labels := series.Labels()
		tags := make([]string, 0, len(labels))
		measurement := ""
		for _, l := range labels {
			if l.Name == "__name__" {
				measurement = l.Value
				continue
			}
			tags = append(tags, escapeInfluxDBChars(l.Name)+"="+escapeInfluxDBChars(l.Value))
		}
		if measurement == "" {
			log.Fatalf("no metric name found in series %v", labels)
		}
		it := series.Iterator()
		for it.Next() {
			ts, val := it.At()
			fmt.Printf("%s,%s f64=%s %d\n", measurement, strings.Join(tags, ","), strconv.FormatFloat(val, 'f', -1, 64), ts*1e6)
		}
		if it.Err() != nil {
			log.Fatal(ss.Err())
		}
	}

	if ss.Err() != nil {
		log.Fatal(ss.Err())
	}
}

func escapeInfluxDBChars(str string) string {
	specialChars := []string{`,`, `=`, ` `, `\`}
	for _, c := range specialChars {
		str = strings.Replace(str, c, `\`+c, -1)
	}
	return str
}
