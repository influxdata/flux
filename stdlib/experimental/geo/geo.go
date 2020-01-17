package geo

import (
	"context"
	"fmt"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/mmcloughlin/geohash"
)

var Functions = map[string]values.Function {
	"getGrid": generateGetGridFunc(),
}

func generateGetGridFunc() values.Function {
	return values.NewFunction(
		"getGrid",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				"box": semantic.NewObjectPolyType(map[string]semantic.PolyType{
					"minLat": semantic.Float,
					"minLon": semantic.Float,
					"maxLat": semantic.Float,
					"maxLon": semantic.Float,
				}, semantic.LabelSet{"minLat", "minLon", "maxLat", "maxLon"}, nil),
				"precision": semantic.Int,
				"maxPrecision": semantic.Int,
				"minSize": semantic.Int,
				"maxSize": semantic.Int,
			},
			Required: semantic.LabelSet{"box"},
			Return: semantic.NewObjectPolyType(map[string]semantic.PolyType{"precision": semantic.Int, "set": semantic.NewArrayPolyType(semantic.String)}, semantic.LabelSet{"precision", "set"}, nil), // { level: int, array: []string }
		}),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			a := interpreter.NewArguments(args)
			box, boxOk := a.Get("box")
			if !boxOk {
				return nil, fmt.Errorf("code %d: box parameter not specified", codes.Invalid)
			}

			precision, precisionOk, err := a.GetInt("precision")
			if err != nil {
				return nil, err
			}
			if !precisionOk {
				precision = -1
			}

			maxPrecision, maxPrecisionOk, err := a.GetInt("maxPrecision")
			if err != nil {
				return nil, err
			}
			if !maxPrecisionOk {
				maxPrecision = -1
			}

			minSize, minSizeOk, err := a.GetInt("minSize")
			if err != nil {
				return nil, err
			}
			if !minSizeOk {
				minSize = -1
			}

			maxSize, maxSizeOk, err := a.GetInt("maxSize")
			if err != nil {
				return nil, err
			}
			if !maxSizeOk {
				maxSize = -1
			}

			minLat, minLatOk := box.Object().Get("minLat")
			minLon, minLonOk := box.Object().Get("minLon")
			maxLat, maxLatOk := box.Object().Get("maxLat")
			maxLon, maxLonOk := box.Object().Get("maxLon")

			if !minLatOk || !minLonOk || !maxLatOk || !maxLonOk {
				return nil, fmt.Errorf("code %d: invalid box specification - must have minLat, minLon, maxLat, maxLon fields", codes.Invalid)
			}

			if minSize > 0 && maxSize > 0 && minSize > maxSize {
				return nil, fmt.Errorf("code %d: minSize > maxSize (%d > %d)", codes.Invalid, minSize, maxSize)
			}

			grid, err := getGrid(&latLonBox{
				minLat: minLat.Float(),
				minLon: minLon.Float(),
				maxLat: maxLat.Float(),
				maxLon: maxLon.Float(),
			}, int(precision), int(maxPrecision), int(minSize), int(maxSize))
			if err != nil {
				return nil, err
			}

			precisionVal := values.NewInt(-1)
			setVal := values.NewArray(semantic.String)
			if grid != nil {
				precisionVal = values.NewInt(int64(grid.getPrecision()))
				for _, hash := range grid.asSet() {
					setVal.Append(values.NewString(hash))
				}
			}

			return values.NewObjectWithValues(map[string]values.Value{
				"precision": precisionVal,
				"set": setVal,
			}), nil
		}, false,
	)
}

func init() {
	flux.RegisterPackageValue("experimental/geo", "getGrid", generateGetGridFunc())
}

//
// Implementation
//

const MaxPrecision = 12
const AbsoluteMaxSize = 4096

type latLonBox struct {
	minLat float64
	maxLat float64
	minLon float64
	maxLon float64
}

type grid struct {
	set [][]string
}

func (g *grid) asSet() []string {
	var result []string
	for _, line := range g.set {
		for _, hash := range line {
			result = append(result, hash)
		}
	}
	return result
}

func (g *grid) getSize() int {
	if len(g.set) > 0 {
		return len(g.set) * len(g.set[0])
	}
	return 0
}

func (g *grid) getPrecision() int {
	if len(g.set) > 0 {
		return len(g.set[0][0])
	}
	return -1
}

func getSpecGridLine(latSrc, lonSrc float64, direction geohash.Direction, latDist, lonDist float64, precision int) []string {
	var result []string

	startPoint := geohash.EncodeWithPrecision(latSrc, lonSrc, uint(precision))
	startPointBox := geohash.BoundingBox(startPoint)
	distHit := startPointBox.Contains(latDist, lonDist)
	result = append(result, startPoint)

	for !distHit { // TODO add check(s) to avoid infinite loop
		neighbor := geohash.Neighbor(startPoint, direction)
		nBox := geohash.BoundingBox(neighbor)
		distHit = nBox.Contains(latDist, lonDist)
		startPoint = neighbor
		result = append(result, neighbor)
	}

	return result
}

func getSpecGrid(rect *latLonBox, precision int) grid {
	var result grid

	gridLineX0 := getSpecGridLine(rect.maxLat, rect.minLon, geohash.South, rect.minLat, rect.minLon, precision)
	gridHeight := len(gridLineX0)
	result.set = make([][]string, gridHeight)
	for i, hash := range gridLineX0 {
		lat, _ := geohash.Decode(hash)
		result.set[i] = getSpecGridLine(lat, rect.minLon, geohash.East, lat, rect.maxLon, precision)
	}

	return result
}

func getGrid(rect *latLonBox, reqPrecision, maxPrecision, minSize, maxSize int) (*grid, error) {
	var result *grid

	if reqPrecision > -1 {
		if maxPrecision > -1 || minSize > -1 || maxSize > -1  {
			return nil, fmt.Errorf("code %d: precision is mutually exclusive with other parameters", codes.Invalid)
		}
		g := getSpecGrid(rect, reqPrecision)
		result = &g
	} else if minSize > 0 || maxSize > 0 {
		if maxPrecision <= 0 {
			maxPrecision = MaxPrecision
		}
		if minSize <= 0 {
			minSize = 0
		}
		if maxSize <= 0 {
			maxSize = AbsoluteMaxSize
		}
		n := 0
		for i := 1; i <= maxPrecision; i++ {
			g := getSpecGrid(rect, i)
			n = g.getSize()
			if n > maxSize {
				break
			}
			result = &g
			if minSize > 0 && n >= minSize {
				break
			}
		}
		if result != nil {
			n = result.getSize()
			if n < minSize {
				result = nil
			}
		}
	} else {
		return nil, fmt.Errorf("code %d: either minSize or maxSize must be specified", codes.Invalid)
	}

	return result, nil
}
