package geo

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

type point struct {
	lat, lon float64
}

type box struct {
	minLat, minLon, maxLat, maxLon float64
}

type circle struct {
	point
	radius float64
}

type polygon struct {
	points []s2.Point
}

type polyline struct {
	latlngs []s2.LatLng
}

type units struct {
	distance    string
	earthRadius float64
}

func (u *units) distanceToRad(v float64) float64 {
	return v / u.earthRadius
}

func (u *units) distanceToUser(v float64) float64 {
	return v * u.earthRadius
}

// WGS-84 Earth's mean radius in kilometers
const earthRadiusKm = 6371.01

var earthRadiuses = map[string]float64{
	"m":    earthRadiusKm * 1000,
	"km":   earthRadiusKm,
	"mile": earthRadiusKm / 1.609344,
}

func generateGetGridFunc() values.Function {
	getGridSignature := runtime.MustLookupBuiltinType("experimental/geo", "getGrid")
	return values.NewFunction(
		"getGrid",
		getGridSignature,
		func(ctx context.Context, args values.Object) (values.Value, error) {
			a := interpreter.NewArguments(args)
			unitsArg, err := a.GetRequiredObject("units")
			if err != nil {
				return nil, err
			}
			units, err := parseUnitsArgument(unitsArg)
			if err != nil {
				return nil, err
			}

			regionArg, err := a.GetRequiredObject("region")
			if err != nil {
				return nil, err
			}

			geom, err := parseGeometryArgument("region", regionArg, units)
			if err != nil {
				return nil, err
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
				return nil, errors.Newf(codes.Invalid, "invalid maxLevel (%d, must be < %d)", maxLevel, MaxLevel)
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
				return nil, errors.Newf(codes.Invalid,"minSize > maxSize (%d > %d)", minSize, maxSize)
			}

			var region s2.Region
			switch v := geom.(type) {
			case box:
				region = getS2RectRegion(v)
			case circle:
				region = getS2CapRegion(v)
			case polygon:
				region = getS2LoopRegion(v)
			default:
				return nil, errors.Newf(codes.Invalid, "unsupported region type: %T", geom)
			}

			grid, err := getGrid(region, int(level), int(maxLevel), int(minSize), int(maxSize))
			if err != nil {
				return nil, err
			}

			levelVal := values.NewInt(-1)
			setVal := values.NewArray(semantic.NewArrayType(semantic.BasicString))
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

func generateGetLevelFunc() values.Function {
	getLevelSignature := runtime.MustLookupBuiltinType("experimental/geo", "getLevel")
	return values.NewFunction(
		"getLevel",
		getLevelSignature,
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

func generateS2CellIDTokenFunc() values.Function {
	s2CellIDTokenSignature := runtime.MustLookupBuiltinType("experimental/geo", "s2CellIDToken")
	return values.NewFunction(
		"s2CellIDToken",
		s2CellIDTokenSignature,
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
				return nil, errors.Newf(codes.Invalid,"either token or point parameter must be specified")
			}

			if tokenOk && pointOk {
				if (len(token) == 0 && point.Len() == 0) || (len(token) > 0 && point.Len() > 0) {
					return nil, errors.Newf(codes.Invalid,"either token or point parameter must be specified and must not be empty")
				}
			}

			level, err := a.GetRequiredInt("level")
			if err != nil {
				return nil, err
			}
			if level < 1 || level > MaxLevel {
				return nil, errors.Newf(codes.Invalid,"level value must be [1, 30]")
			}

			var parentToken string
			if tokenOk && len(token) > 0 {
				parentToken, err = getParentFromToken(token, int(level))
			} else {
				lat, latOk := point.Get("lat")
				lon, lonOk := point.Get("lon")
				if !latOk || !lonOk {
					return nil, errors.Newf(codes.Invalid,"invalid point specification - must have lat, lon fields")
				}
				parentToken, err = getParentFromLatLon(lat.Float(), lon.Float(), int(level))
			}

			return values.NewString(parentToken), err
		}, false,
	)
}

func generateS2CellLatLonFunc() values.Function {
	s2CellLatLonSignature := runtime.MustLookupBuiltinType("experimental/geo", "s2CellLatLon")
	return values.NewFunction(
		"s2CellLatLon",
		s2CellLatLonSignature,
		func(ctx context.Context, args values.Object) (values.Value, error) {
			a := interpreter.NewArguments(args)

			token, err := a.GetRequiredString("token")
			if err != nil {
				return nil, err
			}

			ll, err := getLatLng(token)
			if err != nil {
				return nil, err
			}

			return values.NewObjectWithValues(map[string]values.Value{
				"lat": values.NewFloat(ll.Lat.Degrees()),
				"lon": values.NewFloat(ll.Lng.Degrees()),
			}), nil
		}, false,
	)
}

func init() {
	runtime.RegisterPackageValue("experimental/geo", "getGrid", generateGetGridFunc())
	runtime.RegisterPackageValue("experimental/geo", "getLevel", generateGetLevelFunc())
	runtime.RegisterPackageValue("experimental/geo", "s2CellIDToken", generateS2CellIDTokenFunc())
	runtime.RegisterPackageValue("experimental/geo", "s2CellLatLon", generateS2CellLatLonFunc())
	runtime.RegisterPackageValue("experimental/geo", "stContains", generateSTContainsFunc())
	runtime.RegisterPackageValue("experimental/geo", "stDistance", generateSTDistanceFunc())
	runtime.RegisterPackageValue("experimental/geo", "stLength", generateSTLengthFunc())
}

//
// Flux helpers
//

func parseGeometryArgument(name string, arg values.Object, units *units) (geom interface{}, err error) {
	_, pointOk := arg.Get("lat")
	if pointOk && arg.Len() == 2 {
		lat, latOk := arg.Get("lat")
		lon, lonOk := arg.Get("lon")

		if !latOk || !lonOk {
			return nil, errors.Newf(codes.Invalid,"invalid point specification - must have lat, lon fields")
		}

		geom = point{
			lat: lat.Float(),
			lon: lon.Float(),
		}
	}

	_, boxOk := arg.Get("minLat")
	if boxOk && arg.Len() == 4 {
		minLat, minLatOk := arg.Get("minLat")
		minLon, minLonOk := arg.Get("minLon")
		maxLat, maxLatOk := arg.Get("maxLat")
		maxLon, maxLonOk := arg.Get("maxLon")

		if !minLatOk || !minLonOk || !maxLatOk || !maxLonOk {
			return nil, errors.Newf(codes.Invalid,"invalid box specification - must have minLat, minLon, maxLat, maxLon fields")
		}

		// fix user mistakes
		if minLat.Float() > maxLat.Float() {
			minLat, maxLat = maxLat, minLat
		}
		if minLon.Float() > maxLon.Float() {
			minLon, maxLon = maxLon, minLon
		}

		geom = box{
			minLat: minLat.Float(),
			minLon: minLon.Float(),
			maxLat: maxLat.Float(),
			maxLon: maxLon.Float(),
		}
	}

	_, circleOk := arg.Get("radius")
	if circleOk && arg.Len() == 3 {
		centerLat, centerLatOk := arg.Get("lat")
		centerLon, centerLonOk := arg.Get("lon")
		radius, radiusOk := arg.Get("radius")

		if !centerLatOk || !centerLonOk || !radiusOk {
			return nil, errors.Newf(codes.Invalid,"invalid circle specification - must have lat, lon, radius fields")
		}

		geom = circle{
			point: point{
				lat: centerLat.Float(),
				lon: centerLon.Float(),
			},
			radius: units.distanceToRad(radius.Float()),
		}
	}

	points, polygonOk := arg.Get("points")
	if polygonOk && arg.Len() == 1 {
		array := points.Array()
		if array.Len() < 3 {
			err = errors.Newf(codes.Invalid,"polygon must have at least 3 points")
		}

		s2points := make([]s2.Point, array.Len())
		for i := 0; i < array.Len(); i++ {
			p := array.Get(i).Object()
			lat, latOk := p.Get("lat")
			lon, lonOk := p.Get("lon")

			if !latOk || !lonOk {
				return nil, errors.Newf(codes.Invalid,"invalid polygon point specification - must have lat, lon fields")
			}

			s2points[i] = s2.PointFromLatLng(s2.LatLngFromDegrees(lat.Float(), lon.Float()))
		}

		geom = polygon{
			points: s2points,
		}
	}

	ls, lsOk := arg.Get("linestring")
	if lsOk && arg.Len() == 1 {
		if ls.IsNull() {
			return nil, errors.Newf(codes.Invalid,"empty linestring")
		}
		lsArray := strings.Split(ls.Str(), ",")
		lsLength := len(lsArray)
		latlngs := make([]s2.LatLng, lsLength)
		for i, pair := range lsArray {
			fields := strings.Fields(pair)
			if len(fields) == 0 {
				return nil, errors.Newf(codes.Invalid,"invalid linestring - empty part")
			}
			lon, lonErr := strconv.ParseFloat(fields[0], 64)
			lat, latErr := strconv.ParseFloat(fields[1], 64)
			if latErr != nil || lonErr != nil {
				return nil, errors.Newf(codes.Invalid,"invalid linestring - %v, %v", lonErr, latErr)
			}

			latlngs[i] = s2.LatLngFromDegrees(lat, lon)
		}
		geom = polyline{
			latlngs: latlngs,
		}
	}

	if geom == nil {
		err = errors.Newf(codes.Invalid,"unsupported geometry specified for '%s'", name)
	}

	return geom, err
}

// Return units object
func parseUnitsArgument(arg values.Object) (*units, error) {
	var u *units
	du, ok := arg.Get("distance")
	if ok && arg.Len() == 1 {
		if du.IsNull() {
			return nil, errors.Newf(codes.Invalid,"invalid units parameter: distance field is null")
		}
		r, ok := earthRadiuses[du.Str()]
		if ok {
			u = &units{
				distance:    du.Str(),
				earthRadius: r,
			}
		} else {
			return nil, errors.Newf(codes.Invalid,"invalid units parameter: unsupported distance unit '%s'", du.Str())
		}
	} else {
		return nil, errors.Newf(codes.Invalid,"invalid units parameter: missing field: distance")
	}
	return u, nil
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

func getS2Point(p point) s2.Point {
	return s2.PointFromLatLng(s2.LatLngFromDegrees(p.lat, p.lon))
}

func getS2RectRegion(b box) s2.Rect {
	min := s2.LatLngFromDegrees(b.minLat, b.minLon)
	max := s2.LatLngFromDegrees(b.maxLat, b.maxLon)
	return s2.Rect{
		Lat: r1.Interval{Lo: min.Lat.Radians(), Hi: max.Lat.Radians()},
		Lng: s1.Interval{Lo: min.Lng.Radians(), Hi: max.Lng.Radians()},
	}
}

func getS2CapRegion(c circle) s2.Cap {
	center := s2.PointFromLatLng(s2.LatLngFromDegrees(c.lat, c.lon))
	return s2.CapFromCenterAngle(center, s1.Angle(c.radius))
}

func getS2LoopRegion(p polygon) *s2.Loop {
	loop := s2.LoopFromPoints(p.points)
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
		return nil, errors.Newf(codes.Invalid, "either minSize or maxSize must be specified")
	}

	return result, nil
}

func getParentFromToken(token string, level int) (string, error) {
	cellID := s2.CellIDFromToken(token)
	if cellID.IsValid() && level <= cellID.Level() {
		return cellID.Parent(level).ToToken(), nil
	}
	return "", errors.Newf(codes.Invalid,"invalid token specified or requested level greater then current level")
}

func getParentFromLatLon(lat, lon float64, level int) (string, error) {
	cellID := s2.CellIDFromLatLng(s2.LatLngFromDegrees(lat, lon))
	if cellID.IsValid() {
		return cellID.Parent(level).ToToken(), nil
	}
	return "", errors.Newf(codes.Invalid,"invalid coordinates")
}

func getLevel(token string) (int, error) {
	cellID := s2.CellIDFromToken(token)
	if cellID.IsValid() {
		return cellID.Level(), nil
	}
	return -1, errors.Newf(codes.Invalid, "invalid token specified")
}

func getLatLng(token string) (*s2.LatLng, error) {
	cellID := s2.CellIDFromToken(token)
	if cellID.IsValid() {
		ll := cellID.LatLng()
		return &ll, nil
	}
	return nil, errors.Newf(codes.Invalid,"invalid token specified")
}
