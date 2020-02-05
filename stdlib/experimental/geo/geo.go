package geo

import (
	"context"
	"fmt"
	"github.com/golang/geo/r1"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
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
				"circle": semantic.NewObjectPolyType(map[string]semantic.PolyType{
					"lat": semantic.Float,
					"lon": semantic.Float,
					"radius": semantic.Float,
				}, semantic.LabelSet{"lat", "lon", "radius"}, nil),
				"level":    semantic.Int,
				"maxLevel": semantic.Int,
				"minSize":  semantic.Int,
				"maxSize":  semantic.Int,
			},
			Return:   semantic.NewObjectPolyType(map[string]semantic.PolyType{"level": semantic.Int, "set": semantic.NewArrayPolyType(semantic.String)}, semantic.LabelSet{"level", "set"}, nil), // { level: int, array: []string }
		}),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			a := interpreter.NewArguments(args)
			box, boxOk := a.Get("box")
			circle, circleOk := a.Get("circle")

			if !boxOk && !circleOk {
				return nil, fmt.Errorf("code %d: either box or circle parameter must be specified", codes.Invalid)
			}

			// TODO (alespour@bonitoo.io) how to specify default object value nil at Flux?
			if boxOk && circleOk {
				if (box.Object().Len() == 0 && circle.Object().Len() == 0) || (box.Object().Len() > 0 && circle.Object().Len() > 0) {
					return nil, fmt.Errorf("code %d: either box or circle parameter must be specified and must not be empty", codes.Invalid)
				}
			}

			level, levelOk, err := a.GetInt("level")
			if err != nil {
				return nil, err
			}
			if !levelOk {
				level = -1
			}

			maxLevel, maxLevelOk, err := a.GetInt("maxLevel")
			if err != nil {
				return nil, err
			}
			if !maxLevelOk {
				maxLevel = -1
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

			if minSize > 0 && maxSize > 0 && minSize > maxSize {
				return nil, fmt.Errorf("code %d: minSize > maxSize (%d > %d)", codes.Invalid, minSize, maxSize)
			}

			var region s2.Region

			if boxOk && box.Object().Len() > 0 {
				minLat, minLatOk := box.Object().Get("minLat")
				minLon, minLonOk := box.Object().Get("minLon")
				maxLat, maxLatOk := box.Object().Get("maxLat")
				maxLon, maxLonOk := box.Object().Get("maxLon")

				if !minLatOk || !minLonOk || !maxLatOk || !maxLonOk {
					return nil, fmt.Errorf("code %d: invalid box specification - must have minLat, minLon, maxLat, maxLon fields", codes.Invalid)
				}

				region = getRectRegion(minLat.Float(), minLon.Float(), maxLat.Float(), maxLon.Float())
			} else if circleOk && circle.Object().Len() > 0 {
				lat, latOk := circle.Object().Get("lat")
				lon, lonOk := circle.Object().Get("lon")
				radius, radiusOk := circle.Object().Get("radius")

				if !latOk || !lonOk || !radiusOk {
					return nil, fmt.Errorf("code %d: invalid circle specification - must have lat, lon, radius fields", codes.Invalid)
				}

				region = getCapRegion(lat.Float(), lon.Float(), radius.Float())
			}

			grid, err := getGrid(region, int(level), int(maxLevel), int(minSize), int(maxSize))
			if err != nil {
				return nil, err
			}

			levelVal := values.NewInt(-1)
			setVal := values.NewArray(semantic.String)
			if grid != nil {
				levelVal = values.NewInt(int64(grid.getPrecision()))
				for _, hash := range grid.getSet() {
					setVal.Append(values.NewString(hash))
				}
			}

			return values.NewObjectWithValues(map[string]values.Value{
				"level": levelVal,
				"set":	 setVal,
			}), nil
		}, false,
	)
}

func generateGetParentFunc() values.Function {
	return values.NewFunction(
		"getParent",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				"token":    semantic.String,
				"level": semantic.Int,
			},
			Required: semantic.LabelSet{"token", "level"},
			Return:   semantic.String,
		}),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			a := interpreter.NewArguments(args)
			token, err := a.GetRequiredString("token")
			if err != nil {
				return nil, err
			}

			level, err := a.GetRequiredInt("level")
			if err != nil {
				return nil, err
			}
			if level < 1 || level > MaxLevel {
				return nil, fmt.Errorf("code %d: level value must be [1, 30]", codes.Invalid)
			}

			parentToken := getParent(token, int(level))

			return values.NewString(parentToken), nil
		}, false,
	)
}

func generateContainsLatLonFunc() values.Function {
	return values.NewFunction(
		"containsLatLon",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				"box": semantic.NewObjectPolyType(map[string]semantic.PolyType{
					"minLat": semantic.Float,
					"minLon": semantic.Float,
					"maxLat": semantic.Float,
					"maxLon": semantic.Float,
				}, semantic.LabelSet{"minLat", "minLon", "maxLat", "maxLon"}, nil),
				"circle": semantic.NewObjectPolyType(map[string]semantic.PolyType{
					"lat": semantic.Float,
					"lon": semantic.Float,
					"radius": semantic.Float,
				}, semantic.LabelSet{"lat", "lon", "radius"}, nil),
				"lat":    semantic.Float,
				"lon": semantic.Float,
			},
			Required: semantic.LabelSet{"lat", "lon"},
			Return:   semantic.Bool,
		}),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			a := interpreter.NewArguments(args)
			box, boxOk := a.Get("box")
			circle, circleOk := a.Get("circle")

			if !boxOk && !circleOk {
				return nil, fmt.Errorf("code %d: either box or circle parameter must be specified", codes.Invalid)
			}

			// TODO (alespour@bonitoo.io) how to specify default object value nil at Flux?
			if boxOk && circleOk {
				if (box.Object().Len() == 0 && circle.Object().Len() == 0) || (box.Object().Len() > 0 && circle.Object().Len() > 0) {
					return nil, fmt.Errorf("code %d: either box or circle parameter must be specified and must not be empty", codes.Invalid)
				}
			}

			lat, err := a.GetRequiredFloat("lat")
			if err != nil {
				return nil, err
			}
			lon, err := a.GetRequiredFloat("lon")
			if err != nil {
				return nil, err
			}

			var region s2.Region

			if boxOk && box.Object().Len() > 0 {
				minLat, minLatOk := box.Object().Get("minLat")
				minLon, minLonOk := box.Object().Get("minLon")
				maxLat, maxLatOk := box.Object().Get("maxLat")
				maxLon, maxLonOk := box.Object().Get("maxLon")

				if !minLatOk || !minLonOk || !maxLatOk || !maxLonOk {
					return nil, fmt.Errorf("code %d: invalid box specification - must have minLat, minLon, maxLat, maxLon fields", codes.Invalid)
				}

				region = getRectRegion(minLat.Float(), minLon.Float(), maxLat.Float(), maxLon.Float())
			} else if circleOk && circle.Object().Len() > 0 {
				lat, latOk := circle.Object().Get("lat")
				lon, lonOk := circle.Object().Get("lon")
				radius, radiusOk := circle.Object().Get("radius")

				if !latOk || !lonOk || !radiusOk {
					return nil, fmt.Errorf("code %d: invalid circle specification - must have lat, lon, radius fields", codes.Invalid)
				}

				region = getCapRegion(lat.Float(), lon.Float(), radius.Float())
			}

			point := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lon))
			retVal := region.ContainsPoint(point)

			return values.NewBool(retVal), nil
		}, false,
	)
}

func init() {
	flux.RegisterPackageValue("experimental/geo", "getGrid", generateGetGridFunc())
	flux.RegisterPackageValue("experimental/geo", "getParent", generateGetParentFunc())
	flux.RegisterPackageValue("experimental/geo", "containsLatLon", generateContainsLatLonFunc())
}

//
// Implementation
//

const MaxLevel = 30
const AbsoluteMaxSize = 4096

type grid struct {
	set []string
	precision int
}

func (g *grid) getSet() []string {
	return g.set
}

func (g *grid) getSize() int {
	return len(g.set)
}

func (g *grid) getPrecision() int {
	if len(g.set) > 0 {
		return g.precision
	}
	return -1
}

func getSpecGrid(region s2.Region, precision int) grid {
	var result grid

	rc := &s2.RegionCoverer{MaxLevel: int(precision), MinLevel: int(precision), MaxCells: AbsoluteMaxSize}
	covering := rc.Covering(region)
	size := len(covering)
	if size > 0 {
		result.set = make([]string, size)
		for i, cellId := range covering {
			result.set[i] = cellId.ToToken()
		}
		result.precision = precision
	}

	return result
}

func getRectRegion(minLat, minLon, maxLat, maxLon float64) s2.Region {
	min := s2.LatLngFromDegrees(minLat, minLon)
	max := s2.LatLngFromDegrees(maxLat, maxLon)
	return s2.Rect{
		Lat:r1.Interval{Lo: min.Lat.Radians(), Hi: max.Lat.Radians()},
		Lng:s1.Interval{Lo: min.Lng.Radians(), Hi: max.Lng.Radians()},
	}
}

// The Earth's mean radius in kilometers (according to NASA).
const earthRadiusKm = 6371.01

func getCapRegion(lat ,lon, radius float64) s2.Region {
	center := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lon))
	return s2.CapFromCenterChordAngle(center,s1.ChordAngleFromAngle(s1.Angle(radius / earthRadiusKm)))
}

func getGrid(region s2.Region, reqLevel, maxLevel, minSize, maxSize int) (*grid, error) {
	var result *grid

	if reqLevel > -1 {
		g := getSpecGrid(region, reqLevel)
		result = &g
	} else if minSize > 0 || maxSize > 0 {
		if maxLevel <= 0 {
			maxLevel = MaxLevel
		}
		if minSize <= 0 {
			minSize = 0
		}
		if maxSize <= 0 {
			maxSize = AbsoluteMaxSize
		}
		n := 0
		for i := 1; i <= maxLevel; i++ {
			g := getSpecGrid(region, i)
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

func getParent(token string, level int) string {
	return s2.CellIDFromToken(token).Parent(level).ToToken()
}