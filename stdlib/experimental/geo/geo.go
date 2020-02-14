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
	"math"
)

var boxT = semantic.NewObjectPolyType(map[string]semantic.PolyType{
	"minLat": semantic.Float,
	"minLon": semantic.Float,
	"maxLat": semantic.Float,
	"maxLon": semantic.Float,
}, nil, semantic.LabelSet{"minLat", "minLon", "maxLat", "maxLon"})

var circleT = semantic.NewObjectPolyType(map[string]semantic.PolyType{
	"lat":    semantic.Float,
	"lon":    semantic.Float,
	"radius": semantic.Float,
}, nil, semantic.LabelSet{"lat", "lon", "radius"})

var polygonT = semantic.NewObjectPolyType(map[string]semantic.PolyType{
	"points": semantic.NewArrayPolyType(pointT.Nature()),
}, nil, semantic.LabelSet{"points"})

var pointT = semantic.NewObjectType(map[string]semantic.Type{
	"lat": semantic.Float,
	"lon": semantic.Float,
})

func generateGetGridFunc() values.Function {
	return values.NewFunction(
		"getGrid",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				"box":      boxT,
				"circle":   circleT,
				"polygon":  polygonT,
				"level":    semantic.Int,
				"maxLevel": semantic.Int,
				"minSize":  semantic.Int,
				"maxSize":  semantic.Int,
			},
			Return: semantic.NewObjectPolyType(map[string]semantic.PolyType{"level": semantic.Int, "set": semantic.NewArrayPolyType(semantic.String)}, semantic.LabelSet{"level", "set"}, nil),
		}),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			a := interpreter.NewArguments(args)
			box, boxOk, err := a.GetObject("box")
			if err != nil {
				return nil, err
			}
			circle, circleOk, err := a.GetObject("circle")
			if err != nil {
				return nil, err
			}
			polygon, polygonOk, err := unwrapPolygonArg(a.GetObject("polygon"))
			if err != nil {
				return nil, err
			}

			if err := regionArgumentCheck(box, circle, polygon, boxOk, circleOk, polygonOk); err != nil {
				return nil, err
			}
			if isArrayArgumentOk(polygon, polygonOk) && polygon.Len() < 3 {
				return nil, fmt.Errorf("code %d: polygon must have at least 3 points", codes.Invalid)
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
			} else if maxLevel > MaxLevel {
				return nil, fmt.Errorf("code %d: invalid maxLevel (%d, must be < %d)", codes.Invalid, maxLevel, MaxLevel)
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

			if isObjectArgumentOk(box, boxOk) {
				minLat, minLatOk := box.Get("minLat")
				minLon, minLonOk := box.Get("minLon")
				maxLat, maxLatOk := box.Get("maxLat")
				maxLon, maxLonOk := box.Get("maxLon")

				if !minLatOk || !minLonOk || !maxLatOk || !maxLonOk {
					return nil, fmt.Errorf("code %d: invalid box specification - must have minLat, minLon, maxLat, maxLon fields", codes.Invalid)
				}

				region = getRectRegion(minLat.Float(), minLon.Float(), maxLat.Float(), maxLon.Float())
			} else if isObjectArgumentOk(circle, circleOk) {
				lat, latOk := circle.Get("lat")
				lon, lonOk := circle.Get("lon")
				radius, radiusOk := circle.Get("radius")

				if !latOk || !lonOk || !radiusOk {
					return nil, fmt.Errorf("code %d: invalid circle specification - must have lat, lon and radius fields", codes.Invalid)
				}

				region = getCapRegion(lat.Float(), lon.Float(), radius.Float())
			} else if isArrayArgumentOk(polygon, polygonOk) {
				points := make([]s2.Point, polygon.Len())
				for i := 0; i < polygon.Len(); i++ {
					point := polygon.Get(i).Object()
					lat, latOk := point.Get("lat")
					lon, lonOk := point.Get("lon")

					if !latOk || !lonOk {
						return nil, fmt.Errorf("code %d: invalid polygon point specification - must have lat, lon fields", codes.Invalid)
					}
					points[i] = s2.PointFromLatLng(s2.LatLngFromDegrees(lat.Float(), lon.Float()))
				}

				region = getLoopRegion(points)
			}

			grid, err := getGrid(region, int(level), int(maxLevel), int(minSize), int(maxSize))
			if err != nil {
				return nil, err
			}

			levelVal := values.NewInt(-1)
			setVal := values.NewArray(semantic.String)
			if grid != nil {
				levelVal = values.NewInt(int64(grid.getLevel()))
				for _, hash := range grid.getSet() {
					setVal.Append(values.NewString(hash))
				}
			}

			return values.NewObjectWithValues(map[string]values.Value{
				"level": levelVal,
				"set":   setVal,
			}), nil
		}, false,
	)
}

func generateGetParentFunc() values.Function {
	return values.NewFunction(
		"getParent",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				"token": semantic.String,
				"point": pointT.Nature(),
				"level": semantic.Int,
			},
			Required: semantic.LabelSet{"level"},
			Return:   semantic.String,
		}),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			a := interpreter.NewArguments(args)

			token, tokenOk, err := a.GetString("token")
			if err != nil {
				return nil, err
			}
			point, pointOk, err := a.GetObject("point")
			if err != nil {
				return nil, err
			}

			if !tokenOk && !pointOk {
				return nil, fmt.Errorf("code %d: either token or point parameter must be specified", codes.Invalid)
			}

			// TODO (alespour@bonitoo.io) would not be needed if we knew how to specify default null object/array value in flux
			if tokenOk && pointOk {
				if (len(token) == 0 && point.Len() == 0) || (len(token) > 0 && point.Len() > 0) {
					return nil, fmt.Errorf("code %d: either token or point parameter must be specified and must not be empty", codes.Invalid)
				}
			}

			level, err := a.GetRequiredInt("level")
			if err != nil {
				return nil, err
			}
			if level < 1 || level > MaxLevel {
				return nil, fmt.Errorf("code %d: level value must be [1, 30]", codes.Invalid)
			}

			var parentToken string
			if tokenOk && len(token) > 0 {
				parentToken, err = getParentFromToken(token, int(level))
			} else {
				lat, latOk := point.Get("lat")
				lon, lonOk := point.Get("lon")
				if !latOk || !lonOk {
					return nil, fmt.Errorf("code %d: invalid point specification - must have lat, lon fields", codes.Invalid)
				}
				parentToken, err = getParentFromLatLon(lat.Float(), lon.Float(), int(level))
			}

			return values.NewString(parentToken), err
		}, false,
	)
}

func generateGetLevelFunc() values.Function {
	return values.NewFunction(
		"getLevel",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				"token": semantic.String,
			},
			Required: semantic.LabelSet{"token"},
			Return:   semantic.Int,
		}),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			a := interpreter.NewArguments(args)

			token, err := a.GetRequiredString("token")
			if err != nil {
				return nil, err
			}
			level, err := getLevel(token)

			return values.NewInt(int64(level)), err
		}, false,
	)
}

func generateContainsLatLonFunc() values.Function {
	return values.NewFunction(
		"containsLatLon",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				"box":     boxT,
				"circle":  circleT,
				"polygon": polygonT,
				"lat":     semantic.Float,
				"lon":     semantic.Float,
			},
			Required: semantic.LabelSet{"lat", "lon"},
			Return:   semantic.Bool,
		}),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			a := interpreter.NewArguments(args)
			box, boxOk, err := a.GetObject("box")
			if err != nil {
				return nil, err
			}
			circle, circleOk, err := a.GetObject("circle")
			if err != nil {
				return nil, err
			}
			polygon, polygonOk, err := unwrapPolygonArg(a.GetObject("polygon"))
			if err != nil {
				return nil, err
			}

			if err := regionArgumentCheck(box, circle, polygon, boxOk, circleOk, polygonOk); err != nil {
				return nil, err
			}
			if isArrayArgumentOk(polygon, polygonOk) && polygon.Len() < 3 {
				return nil, fmt.Errorf("code %d: polygon must have at least 3 points", codes.Invalid)
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

			if isObjectArgumentOk(box, boxOk) {
				minLat, minLatOk := box.Get("minLat")
				minLon, minLonOk := box.Get("minLon")
				maxLat, maxLatOk := box.Get("maxLat")
				maxLon, maxLonOk := box.Get("maxLon")

				if !minLatOk || !minLonOk || !maxLatOk || !maxLonOk {
					return nil, fmt.Errorf("code %d: invalid box specification - must have minLat, minLon, maxLat, maxLon fields", codes.Invalid)
				}

				region = getRectRegion(minLat.Float(), minLon.Float(), maxLat.Float(), maxLon.Float())
			} else if isObjectArgumentOk(circle, circleOk) {
				centerLat, centerLatOk := circle.Get("lat")
				centerLon, centerLonOk := circle.Get("lon")
				radius, radiusOk := circle.Get("radius")

				if !centerLatOk || !centerLonOk || !radiusOk {
					return nil, fmt.Errorf("code %d: invalid circle specification - must have lat, lon, radius fields", codes.Invalid)
				}

				region = getCapRegion(centerLat.Float(), centerLon.Float(), radius.Float())
			} else if isArrayArgumentOk(polygon, polygonOk) {
				points := make([]s2.Point, polygon.Len())
				for i := 0; i < polygon.Len(); i++ {
					point := polygon.Get(i).Object()
					pointLat, pointLatOk := point.Get("lat")
					pointLon, pointLonOk := point.Get("lon")

					if !pointLatOk || !pointLonOk {
						return nil, fmt.Errorf("code %d: invalid polygon point specification - must have lat, lon fields", codes.Invalid)
					}

					points[i] = s2.PointFromLatLng(s2.LatLngFromDegrees(pointLat.Float(), pointLon.Float()))
				}

				region = getLoopRegion(points)
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
	flux.RegisterPackageValue("experimental/geo", "getLevel", generateGetLevelFunc())
	flux.RegisterPackageValue("experimental/geo", "containsLatLon", generateContainsLatLonFunc())
}

//
// Flux helpers
//

func unwrapPolygonArg(pObject values.Object, pObjectOk bool, err error) (values.Array, bool, error) {
	if err != nil {
		return nil, false, err
	}
	if pObjectOk {
		points, pointsOk := pObject.Get("points")
		if pointsOk {
			return points.Array(), true, nil
		}
	}
	return nil, false, nil
}

func regionArgumentCheck(box, circle values.Object, polygon values.Array, boxOk, circleOk, polygonOk bool) error {
	oks := 0
	nzs := 0

	if boxOk {
		oks++
		if box.Len() > 0 {
			nzs++
		}
	}
	if circleOk {
		oks++
		if circle.Len() > 0 {
			nzs++
		}
	}
	if polygonOk {
		oks++
		if polygon.Len() > 0 {
			nzs++
		}
	}

	if oks == 0 || nzs == 0 {
		return fmt.Errorf("code %d: either box or circle or polygon parameter must be specified", codes.Invalid)
	}

	// TODO (alespour@bonitoo.io) would not be needed if we knew how to specify default null object/array value in flux
	if oks > 1 && nzs > 1 {
		return fmt.Errorf("code %d: either box or circle or polygon parameter must be specified and must not be empty", codes.Invalid)
	}

	return nil
}

func isObjectArgumentOk(v values.Object, vOk bool) bool {
	return vOk && v.Len() > 0
}

func isArrayArgumentOk(v values.Array, vOk bool) bool {
	return vOk && v.Len() > 0
}

//
// S2 geo implementation
//

const MaxLevel = 30 // https://s2geometry.io/resources/s2cell_statistics.html
const AbsoluteMaxSize = 100

type grid struct {
	set   []string
	level int
}

func (g *grid) getSet() []string {
	return g.set
}

func (g *grid) getSize() int {
	return len(g.set)
}

func (g *grid) getLevel() int {
	if len(g.set) > 0 {
		return g.level
	}
	return -1
}

func getSpecGrid(region s2.Region, level int) grid {
	var result grid

	rc := &s2.RegionCoverer{MaxLevel: int(level), MinLevel: int(level), MaxCells: AbsoluteMaxSize}
	covering := rc.Covering(region)
	size := len(covering)
	if size > 0 {
		result.set = make([]string, size)
		for i, cellId := range covering {
			result.set[i] = cellId.ToToken()
		}
		result.level = level
	}

	return result
}

func getRectRegion(minLat, minLon, maxLat, maxLon float64) s2.Region {
	min := s2.LatLngFromDegrees(minLat, minLon)
	max := s2.LatLngFromDegrees(maxLat, maxLon)
	return s2.Rect{
		Lat: r1.Interval{Lo: min.Lat.Radians(), Hi: max.Lat.Radians()},
		Lng: s1.Interval{Lo: min.Lng.Radians(), Hi: max.Lng.Radians()},
	}
}

// The Earth's mean radius in kilometers (according to NASA).
const earthRadiusKm = 6371.01

func getCapRegion(lat, lon, radius float64) s2.Region {
	center := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lon))
	return s2.CapFromCenterAngle(center, s1.Angle(radius/earthRadiusKm))
}

func getLoopRegion(points []s2.Point) s2.Region {
	loop := s2.LoopFromPoints(points)
	if loop.Area() >= 2*math.Pi { // points are not CCW but CW
		loop.Invert()
	}
	return loop
}

func getGrid(region s2.Region, reqLevel, maxLevel, minSize, maxSize int) (*grid, error) {
	var result *grid

	if reqLevel > -1 {
		g := getSpecGrid(region, reqLevel)
		result = &g
	} else if minSize > 0 || maxSize > 0 {
		maxLevelFallback := maxLevel
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
			if n < minSize && maxLevelFallback <= 0 {
				result = nil
			}
		}
	} else {
		return nil, fmt.Errorf("code %d: either minSize or maxSize must be specified", codes.Invalid)
	}

	return result, nil
}

func getParentFromToken(token string, level int) (string, error) {
	cellID := s2.CellIDFromToken(token)
	if cellID.IsValid() && level <= cellID.Level() {
		return cellID.Parent(level).ToToken(), nil
	}
	return "", fmt.Errorf("code %d: invalid token specified or requested level greater then current level", codes.Invalid)
}

func getParentFromLatLon(lat, lon float64, level int) (string, error) {
	cellID := s2.CellIDFromLatLng(s2.LatLngFromDegrees(lat, lon))
	if cellID.IsValid() {
		return cellID.Parent(level).ToToken(), nil
	}
	return "", fmt.Errorf("code %d: invalid coordinates", codes.Invalid)
}

func getLevel(token string) (int, error) {
	cellID := s2.CellIDFromToken(token)
	if cellID.IsValid() {
		return cellID.Level(), nil
	}
	return -1, fmt.Errorf("code %d: invalid token specified", codes.Invalid)
}
