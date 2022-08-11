package rdp

import (
	"math"
	"sort"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

type Rdp struct {
	timeColumn       string
	valColumn        string
	epsilon          float64
	retentionPercent float64
	vs               *array.Float
	ts               *array.Float
	alloc            memory.Allocator
}

func New(timeColumn string, valColumn string, threshold float64, retentionPercent float64, alloc memory.Allocator) *Rdp {
	return &Rdp{
		timeColumn:       timeColumn,
		valColumn:        valColumn,
		epsilon:          threshold,
		retentionPercent: retentionPercent,
		alloc:            alloc,
	}
}

type Point struct {
	xCoordinate float64
	yCoordinate float64
}

type ConnectLine struct {
	Starting Point
	Ending   Point
}

// Finds the perpendicular distance from a given 2D point to a line

func (l ConnectLine) perpendicularDistanceFromPointToLine(pt Point) float64 {
	a, b, c := l.Coefficients()
	return math.Abs(a*pt.xCoordinate+b*pt.yCoordinate+c) / math.Sqrt(a*a+b*b)
}

func (l ConnectLine) Coefficients() (a, b, c float64) {
	a = l.Starting.yCoordinate - l.Ending.yCoordinate
	b = l.Ending.xCoordinate - l.Starting.xCoordinate
	c = l.Starting.xCoordinate*l.Ending.yCoordinate - l.Ending.xCoordinate*l.Starting.yCoordinate
	return a, b, c
}

// Finds the farthest 2D point from a given line

func FindFarthestPoint(line ConnectLine, ipdata []Point) (farthestPointIndex int, farthestPointDistanceFromLine float64) {
	for i := 0; i < len(ipdata); i++ {
		distance := line.perpendicularDistanceFromPointToLine(ipdata[i])
		if distance > farthestPointDistanceFromLine {
			farthestPointDistanceFromLine = distance
			farthestPointIndex = i
		}
	}
	return farthestPointIndex, farthestPointDistanceFromLine
}

// Uses the times array and values array to build a single array containing 2D points

func (r *Rdp) ToPoints() []Point {
	points := make([]Point, 0, r.vs.Len())
	for i := 0; i < r.vs.Len(); i++ {
		points = append(points, Point{xCoordinate: r.ts.Value(i), yCoordinate: r.vs.Value(i)})
	}

	return points
}

// Recursive implementation of the RDP Algorithm given the epsilon value

func DownSampleIpdata(ipdata []Point, epsilon float64) []Point {
	if len(ipdata) <= 2 {
		return ipdata
	}
	line := ConnectLine{Starting: ipdata[0], Ending: ipdata[len(ipdata)-1]}

	farthestPointIndex, farthestPointDistanceFromLine := FindFarthestPoint(line, ipdata)

	if farthestPointDistanceFromLine >= epsilon {
		leftSplit := DownSampleIpdata(ipdata[:farthestPointIndex+1], epsilon)
		rightSplit := DownSampleIpdata(ipdata[farthestPointIndex:], epsilon)
		return append(leftSplit[:len(leftSplit)-1], rightSplit...)
	}

	return []Point{ipdata[0], ipdata[len(ipdata)-1]}

}

//Iterative implementation of the RDP algorithm for the given retention rate

func DownSampleRetention(ipdata []Point, startIndex int64, endIndex int64, weights []float64) []float64 {
	var stack [][]int64
	globalStartIndex := startIndex
	globalEndIndex := endIndex
	stack = append(stack, []int64{startIndex, endIndex})
	for len(stack) > 0 {
		n := len(stack) - 1
		lineOfInterest := stack[n]
		stack = stack[:n]
		startIndex := lineOfInterest[0]
		endIndex := lineOfInterest[1]
		maxDistance := 0.0
		index := startIndex
		line := ConnectLine{Starting: ipdata[startIndex], Ending: ipdata[endIndex]}
		for i := index + 1; i < endIndex; i++ {

			distance := line.perpendicularDistanceFromPointToLine(ipdata[i])
			if distance > maxDistance {
				index = i
				maxDistance = distance
			}

		}
		if maxDistance > 0.0 {
			weights[index] = maxDistance
			stack = append(stack, []int64{startIndex, index})
			stack = append(stack, []int64{index, endIndex})
		}
	}
	weights[globalStartIndex] = math.Inf(1)
	weights[globalEndIndex] = math.Inf(1)
	return weights
}

//Learning the value of epsilon (maxTolerance) automatically if both the epsilon value and the retention rate have not been given

func DownSampleIpdataAuto(ipdata []Point) []Point {
	if len(ipdata) <= 2 {
		return ipdata
	}
	line := ConnectLine{Starting: ipdata[0], Ending: ipdata[len(ipdata)-1]}

	farthestPointIndex, farthestPointDistanceFromLine := FindFarthestPoint(line, ipdata)
	x1 := ipdata[0].xCoordinate
	x2 := ipdata[len(ipdata)-1].xCoordinate
	y1 := ipdata[0].yCoordinate
	y2 := ipdata[len(ipdata)-1].yCoordinate
	lengthOfLine := math.Sqrt(math.Pow((x2-x1), 2) + math.Pow((y2-y1), 2))
	slope := 0.0
	if (x2 - x1) != 0 {
		slope = (y2 - y1) / (x2 - x1)
	}
	phi := math.Atan(slope) //Phi is the angle between the digitized line and the continuous 2D line made with the X-axis
	tmax := (1 / lengthOfLine) * (math.Abs(math.Sin(phi)) + math.Abs(math.Cos(phi)))
	dophimaxone := math.Atan((1 / lengthOfLine) * (math.Abs(math.Sin(phi) + math.Cos(phi))) * (1 - tmax + math.Pow(tmax, 2)))
	dophimaxtwo := math.Atan((1 / lengthOfLine) * (math.Abs(math.Sin(phi) - math.Cos(phi))) * (1 - tmax + math.Pow(tmax, 2)))
	dophimax := math.Max(dophimaxone, dophimaxtwo) //Maximum deviation
	window_threshold := lengthOfLine * dophimax
	if farthestPointDistanceFromLine >= window_threshold {
		leftSplit := DownSampleIpdataAuto(ipdata[:farthestPointIndex+1])
		rightSplit := DownSampleIpdataAuto(ipdata[farthestPointIndex:])
		return append(leftSplit[:len(leftSplit)-1], rightSplit...)
	}

	return []Point{ipdata[0], ipdata[len(ipdata)-1]}

}

// Given the array of 2D points split them into separate x and y arrays

func (r *Rdp) SplitCoordinates(sampledPoints []Point) (*array.Int, *array.Float) {
	newts := array.NewIntBuilder(r.alloc)
	newvs := array.NewFloatBuilder(r.alloc)
	for i := range sampledPoints {
		newts.Append(int64(sampledPoints[i].xCoordinate))
		newvs.Append(sampledPoints[i].yCoordinate)
	}
	return newts.NewIntArray(), newvs.NewFloatArray()
}

// Main implementation

func (r *Rdp) Do(vs *array.Float, ts *array.Float) (*array.Int, *array.Float, error) {
	r.vs = vs
	r.ts = ts
	points := r.ToPoints()
	var sampledPoints []Point
	// if both epsilon and retention rate is given together

	if r.epsilon != 0.0 && r.retentionPercent != 0.0 {
		return nil, nil, errors.New(codes.Invalid, "Please provide either epsilon or retention rate and not both")
	}
	// If epsilon has been given

	if r.epsilon != 0.0 {
		sampledPoints = DownSampleIpdata(points, r.epsilon)
		Newts, Newvs := r.SplitCoordinates(sampledPoints)
		return Newts, Newvs, nil
	} else if r.retentionPercent != 0.0 { // if the retention percent is given
		numOfPoints := int64(vs.Len())
		retention := r.retentionPercent
		weights := make([]float64, numOfPoints)
		pointsToKeep := math.Floor((retention / 100) * float64(numOfPoints))
		if pointsToKeep >= 1.0 {
			weightsUpdated := DownSampleRetention(points, 0, numOfPoints-1, weights)
			weightsDescending := make([]float64, len(weightsUpdated))
			copy(weightsDescending, weightsUpdated)
			sort.Sort(sort.Reverse(sort.Float64Slice(weightsDescending)))
			maxTolerance := weightsDescending[int64(pointsToKeep)-1]
			for i := range points {
				if weightsUpdated[i] >= maxTolerance {
					sampledPoints = append(sampledPoints, points[i])
				}
			}
			Newts, Newvs := r.SplitCoordinates(sampledPoints)
			return Newts, Newvs, nil
		} else {
			return nil, nil, errors.New(codes.Invalid, "The number of points to retain is less than 1. Please consider increasing the retention rate")
		}
	} else { // if both epsilon and retention rate is not given then we try to learn the maximum tolerance by ourselves.
		sampledPoints := DownSampleIpdataAuto(points)
		Newts, Newvs := r.SplitCoordinates(sampledPoints)
		return Newts, Newvs, nil
	}
}
