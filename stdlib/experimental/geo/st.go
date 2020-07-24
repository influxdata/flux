package geo

import (
	"context"
	"fmt"
	"math"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

func generateSTContainsFunc() values.Function {
	stContainsSignature := runtime.MustLookupBuiltinType("experimental/geo", "stContains")
	return values.NewFunction(
		"stContains",
		stContainsSignature,
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

			geom1Arg, err := a.GetRequiredObject("region")
			if err != nil {
				return nil, err
			}

			geom2Arg, err := a.GetRequiredObject("geometry")
			if err != nil {
				return nil, err
			}

			geom1, err := parseGeometryArgument("region", geom1Arg, units)
			if err != nil {
				return nil, err
			}

			geom2, err := parseGeometryArgument("geometry", geom2Arg, units)
			if err != nil {
				return nil, err
			}

			var region s2.Region
			switch v := geom1.(type) {
			case box:
				region = getS2RectRegion(v)
			case circle:
				region = getS2CapRegion(v)
			case polygon:
				region = getS2LoopRegion(v)
			default:
				return nil, fmt.Errorf("code %d: unsupported region type: %T", codes.Invalid, geom1)
			}

			var retVal bool
			switch v := geom2.(type) {
			case point:
				retVal = region.ContainsPoint(getS2Point(v))
			case polyline:
				for _, ll := range v.latlngs {
					retVal = region.ContainsPoint(s2.PointFromLatLng(ll))
					if !retVal {
						break
					}
				}
			default:
				return nil, fmt.Errorf("code %d: unsupported geometry type: %T", codes.Invalid, geom2)
			}

			return values.NewBool(retVal), nil
		}, false,
	)
}

func generateSTDistanceFunc() values.Function {
	stDistanceSignature := runtime.MustLookupBuiltinType("experimental/geo", "stDistance")
	return values.NewFunction(
		"stDistance",
		stDistanceSignature,
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

			geom1Arg, err := a.GetRequiredObject("region")
			if err != nil {
				return nil, err
			}

			geom2Arg, err := a.GetRequiredObject("geometry")
			if err != nil {
				return nil, err
			}

			geom1, err := parseGeometryArgument("region", geom1Arg, units)
			if err != nil {
				return nil, err
			}

			geom2, err := parseGeometryArgument("geometry", geom2Arg, units)
			if err != nil {
				return nil, err
			}

			var distance s1.Angle

			switch v := geom2.(type) {
			case point:
				to := getS2Point(v)
				switch v := geom1.(type) {
				case point: // point-point distance
					distance = getS2Point(v).Distance(to)
				case box: // rect-point distance
					distance = getS2RectRegion(v).DistanceToLatLng(s2.LatLngFromPoint(to))
				case circle: // circle-point distance
					distance = getS2Point(v.point).Distance(to) - s1.Angle(v.radius)
					if distance < 0.0 {
						distance = 0.0
					}
				case polygon: // polygon-point distance
					index := shapeToIndex(getS2LoopRegion(v))
					distance = minDistanceToPoint(index, to)
				}
			case polyline: // linestring represents path (track) in GIS
				toIndex := shapeToIndex(s2.PolylineFromLatLngs(v.latlngs))
				switch v := geom1.(type) {
				case point: // point-polyline distance
					distance = minDistanceToPoint(toIndex, getS2Point(v))
				case box: // box-polyline distance
					index := shapeToIndex(getS2LoopRegion(boxToPolygon(v))) // represent box as polygon
					distance = minDistanceToShapeIndex(index, toIndex)
				case circle: // circle-polyline distance
					distance = minDistanceToPoint(toIndex, getS2Point(v.point)) - s1.Angle(v.radius)
					if distance < 0.0 {
						distance = 0.0
					}
				case polygon: // polygon-polyline distance
					index := shapeToIndex(getS2LoopRegion(v))
					distance = minDistanceToShapeIndex(index, toIndex)
				}
			default:
				return nil, fmt.Errorf("code %d: unsupported geometry type: %T", codes.Invalid, geom2)
			}

			return values.NewFloat(units.distanceToUser(distance.Radians())), nil
		}, false,
	)
}

func generateSTLengthFunc() values.Function {
	stLengthSignature := runtime.MustLookupBuiltinType("experimental/geo", "stLength")
	return values.NewFunction(
		"stLength",
		stLengthSignature,
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

			geomArg, err := a.GetRequiredObject("geometry")
			if err != nil {
				return nil, err
			}

			geom, err := parseGeometryArgument("geometry", geomArg, units)
			if err != nil {
				return nil, err
			}

			var length s1.Angle
			switch v := geom.(type) {
			case point:
				length = 0.0
			case polyline:
				for i := 0; i < len(v.latlngs)-1; i++ {
					length += v.latlngs[i].Distance(v.latlngs[i+1])
				}
			default:
				return nil, fmt.Errorf("code %d: unsupported geometry type: %T", codes.Invalid, geom)
			}

			return values.NewFloat(units.distanceToUser(length.Radians())), nil
		}, false,
	)
}

//
// helper functions
//

// Returns index containing specified shape
func shapeToIndex(shape s2.Shape) *s2.ShapeIndex {
	index := s2.NewShapeIndex()
	index.Add(shape)
	return index
}

// Convert box to polygon
func boxToPolygon(b box) polygon {
	points := make([]s2.Point, 4)
	points[0] = s2.PointFromLatLng(s2.LatLngFromDegrees(b.minLat, b.minLon))
	points[1] = s2.PointFromLatLng(s2.LatLngFromDegrees(b.minLat, b.maxLon))
	points[2] = s2.PointFromLatLng(s2.LatLngFromDegrees(b.maxLat, b.maxLon))
	points[3] = s2.PointFromLatLng(s2.LatLngFromDegrees(b.maxLat, b.minLon))
	return polygon{
		points: points,
	}
}

// Returns distance as angle
func minDistanceToPoint(index *s2.ShapeIndex, point s2.Point) s1.Angle {
	var distance s1.Angle
	query := s2.NewClosestEdgeQuery(index, s2.NewClosestEdgeQueryOptions().MaxResults(1))
	target := s2.NewMinDistanceToPointTarget(point)
	results := query.FindEdges(target)
	if results != nil || len(results) == 1 {
		distance = results[0].Distance().Angle()
	} else {
		distance = s1.Angle(math.NaN())
	}
	return distance
}

// Returns distance as angle
func minDistanceToShapeIndex(index *s2.ShapeIndex, index2 *s2.ShapeIndex) s1.Angle {
	var distance s1.Angle
	query := s2.NewClosestEdgeQuery(index, s2.NewClosestEdgeQueryOptions().MaxResults(1))
	target := s2.NewMinDistanceToShapeIndexTarget(index2)
	results := query.FindEdges(target)
	if results != nil || len(results) == 1 {
		distance = results[0].Distance().Angle()
	} else {
		distance = s1.Angle(math.NaN())
	}
	return distance
}
