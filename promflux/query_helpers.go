package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/execute"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func queryPrometheus(url string, expr string, start time.Time, end time.Time, resolution time.Duration) (model.Matrix, error) {
	c, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating Prometheus API client: %s", err)
	}

	promAPI := v1.NewAPI(c)
	v, err := promAPI.QueryRange(context.Background(), expr, v1.Range{
		Start: start,
		End:   end,
		Step:  resolution,
	})
	if err != nil {
		return nil, fmt.Errorf("error querying Prometheus: %s", err)
	}

	return v.(model.Matrix), nil
}

func queryInfluxDB(url string, org string, token string, bucket string, query string) (flux.ResultIterator, error) {
	jsonObj := map[string]interface{}{
		"dialect": map[string]interface{}{
			"annotations": []string{"group", "datatype", "default"},
		},
		"query": query,
		"type":  "flux",
	}
	jsonBody, err := json.Marshal(jsonObj)
	if err != nil {
		return nil, fmt.Errorf("error marshaling query JSON: %s", err)
	}

	body := bytes.NewBuffer(jsonBody)
	req, err := http.NewRequest("POST", url+"api/v2/query?org="+org, body)
	if err != nil {
		return nil, fmt.Errorf("error creating InfluxDB request: %s", err)
	}
	req.Header.Add("Authorization", "Token "+token)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error querying InfluxDB: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading InfluxDB response body: %s", err)
		}

		return nil, fmt.Errorf("bad HTTP status code from InfluxDB: %s - Body: %s", resp.Status, b)
	}
	decoder := csv.NewMultiResultDecoder(csv.ResultDecoderConfig{})
	return decoder.Decode(resp.Body)
}

func influxResultToPromMatrix(resultIt flux.ResultIterator) (model.Matrix, error) {
	fpToSS := map[model.Fingerprint]*model.SampleStream{}
	for resultIt.More() {
		r := resultIt.Next()
		tableIt := r.Tables()
		tableIt.Do(func(tbl flux.Table) error {
			tbl.Do(func(cr flux.ColReader) error {
				for i := 0; i < cr.Len(); i++ {
					met := model.Metric{}
					var val model.SampleValue
					var ts model.Time

					for j, col := range cr.Cols() {
						switch col.Label {
						case "_measurement":
							met[model.MetricNameLabel] = model.LabelValue(cr.Strings(j).Value(i))
						case "_time":
							ts = model.TimeFromUnixNano(execute.ValueForRow(cr, i, j).Time().Time().UnixNano())
						case "_value":
							val = model.SampleValue(execute.ValueForRow(cr, i, j).Float())
						case "_start", "_stop", "_field":
							// Ignore.
						default:
							met[model.LabelName(col.Label)] = model.LabelValue(cr.Strings(j).Value(i))
						}
					}

					sp := model.SamplePair{
						Timestamp: ts,
						Value:     val,
					}
					fp := met.Fingerprint()
					if ss, ok := fpToSS[fp]; !ok {
						fpToSS[fp] = &model.SampleStream{
							Metric: met,
							Values: []model.SamplePair{sp},
						}
					} else {
						ss.Values = append(ss.Values, sp)
					}
				}
				return nil
			})
			return nil
		})
	}
	if err := resultIt.Err(); err != nil {
		return nil, fmt.Errorf("error processing InfluxDB results: %s", err)
	}

	matrix := make(model.Matrix, 0, len(fpToSS))
	for _, ss := range fpToSS {
		// TODO: Also sort sample stream by time? Or are these always sorted coming from InfluxDB?
		matrix = append(matrix, ss)
	}
	sort.Sort(matrix)
	return matrix, nil
}
