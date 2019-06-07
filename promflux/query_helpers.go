package promflux

import (
	"sort"

	"github.com/influxdata/flux"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"

	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/semantic"
)

func InfluxResultToPromMatrix(result flux.Result) promql.Matrix {
	hashToSeries := map[uint64]*promql.Series{}

	result.Tables().Do(func(tbl flux.Table) error {
		tbl.Do(func(cr flux.ColReader) error {
			for i := 0; i < cr.Len(); i++ {
				builder := labels.NewBuilder(nil)
				var val float64
				var ts int64

				for j, col := range cr.Cols() {
					switch col.Label {
					case "_time":
						ts = execute.ValueForRow(cr, i, j).Time().Time().UnixNano() / 1e6
					case "_value":
						v := execute.ValueForRow(cr, i, j)
						switch v.Type().Nature() {
						case semantic.Float:
							val = v.Float()
						case semantic.Int:
							// TODO: Should this be allowed to happen?
							val = float64(v.Int())
						case semantic.UInt:
							// TODO: Should this be allowed to happen?
							val = float64(v.UInt())
						default:
							panic("invalid value type")
						}
					case "_start", "_stop", "_measurement":
						// Ignore.
					default:
						ln := unescapeLabelName(col.Label)
						builder.Set(ln, cr.Strings(j).ValueString(i))
					}
				}

				lbls := builder.Labels()

				point := promql.Point{
					T: ts,
					V: val,
				}
				hash := lbls.Hash()
				if ser, ok := hashToSeries[hash]; !ok {
					hashToSeries[hash] = &promql.Series{
						Metric: lbls,
						Points: []promql.Point{point},
					}
				} else {
					ser.Points = append(ser.Points, point)
				}
			}
			return nil
		})
		return nil
	})

	matrix := make(promql.Matrix, 0, len(hashToSeries))
	for _, ser := range hashToSeries {
		// TODO: Also sort series by time? Or are these always sorted coming from InfluxDB?
		matrix = append(matrix, *ser)
	}
	sort.Sort(matrix)
	return matrix
}
