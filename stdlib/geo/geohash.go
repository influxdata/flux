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

			fmt.Printf("box: %v, precision: %d, maxPrecision: %d, minSize: %d, maxSize: %d\n", box, precision, maxPrecision, minSize, maxSize)

			minLat, minLatOk := box.Object().Get("minLat")
			minLon, minLonOk := box.Object().Get("minLon")
			maxLat, maxLatOk := box.Object().Get("maxLat")
			maxLon, maxLonOk := box.Object().Get("maxLon")

			if !minLatOk || !minLonOk || !maxLatOk || !maxLonOk {
				return nil, fmt.Errorf("code %d: invalid box specification - it must have minLat, minLon, maxLat, maxLon items", codes.Invalid)
			}

			grid, err := getGrid(&rectangle{
				minLat: minLat.Float(),
				minLon: minLon.Float(),
				maxLat: maxLat.Float(),
				maxLon: maxLon.Float(),
			}, int(precision), int(maxPrecision), int(minSize), int(maxSize))
			if err != nil {
				return nil, err
			}

			set := values.NewArray(semantic.String)
			for _, hash := range grid.asSet() {
				set.Append(values.NewString(hash))
			}

			return values.NewObjectWithValues(map[string]values.Value{
				"precision": values.NewInt(int64(grid.getPrecision())),
				"set": set,
			}), nil
		}, false,
	)
}

func init() {
	flux.RegisterPackageValue("geo", "getGrid", generateGetGridFunc())
}

//
// Implementation
//

const MaxSize = 100
const MaxPrecision = 9

type rectangle struct {
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

func getSpecGrid(rect *rectangle, precision int) grid {
	var result grid

	gridLineX0 := getSpecGridLine(rect.maxLat, rect.minLon, geohash.South, rect.minLat, rect.minLon, precision)
	fmt.Printf("X0 line: %v\n", gridLineX0)
	gridHeight := len(gridLineX0)
	result.set = make([][]string, gridHeight)
	for i, hash := range gridLineX0 {
		lat, _ := geohash.Decode(hash)
		result.set[i] = getSpecGridLine(lat, rect.minLon, geohash.East, lat, rect.maxLon, precision)
	}

	fmt.Println("+ ---")
	for _, line := range result.set {
		printGridLine(geohash.East, line)
	}
	fmt.Println("+ ---")
	fmt.Printf("%v\n", result.asSet())

	return result
}

func printGridLine(orientation geohash.Direction, gridLine []string) {
	switch orientation {
	case geohash.East:
		for _, hash := range gridLine {
			fmt.Printf("| %s ", hash)
		}
		fmt.Println("|")
		break
	case geohash.South:
		for _, hash := range gridLine {
			fmt.Printf("| %s |\n", hash)
		}
		break
	}
}

func getGrid(rect *rectangle, reqPrecision, maxPrecision, minSize, maxSize int) (*grid, error) {
	var result grid

	if reqPrecision > -1 {
		if maxPrecision > -1 || minSize > -1 || maxSize > -1  {
			return nil, fmt.Errorf("code %d: reqPrecision is mutually exlusive with other parameters", codes.Invalid)
		}
		result = getSpecGrid(rect, reqPrecision)
		fmt.Printf("grid with specific precision %d\n", reqPrecision)
	} else if minSize > 0  || maxSize > 0 {
		if maxPrecision < 0 {
			maxPrecision = MaxPrecision
		}
		if minSize < 0 {
			minSize = 0
		}
		if maxSize < 0 {
			maxSize = MaxSize
		}
		if minSize > maxSize {
			return nil, fmt.Errorf("code %d: minSize > maxSize (%d > %d)", codes.Invalid, minSize, maxSize)
		}
		n := 0
		fmt.Printf("for until i <= %d and n < %d\n", maxPrecision, minSize)
		for i := 1; i <= maxPrecision && n < minSize; i++ {
			g := getSpecGrid(rect, i)
			n = g.getSize()
			if n > maxSize {
				fmt.Printf("break %d > %d\n", n, maxSize)
				break
			}
			result = g
		}
		n = result.getSize()
		fmt.Printf("grid length = %d, min size = %d\n", n, minSize)
		if n < minSize {
			return nil, nil
		}
	} else {
		return nil, fmt.Errorf("code %d: either minSize or maxSize must be specified", codes.Invalid)
	}

	return &result, nil
}
