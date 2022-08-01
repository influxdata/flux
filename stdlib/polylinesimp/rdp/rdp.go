package rdp

import (
	"math"
	"sort"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux/array"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type Rdp struct {
	timeColumn       string
	column           string
	epsilon          float64
	retentionPercent float64
	vs               *array.Float
	ts               *array.Float
	alloc            memory.Allocator
}

func New(timeColumn string, column string, threshold float64, retentionPercent float64, alloc memory.Allocator) *Rdp {
	return &Rdp{
		timeColumn:       timeColumn,
		column:           column,
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

//This function saves a png containing line plots for both the original and the simplified path

func SavePlot(orig, simp plotter.XYs) error {
	p := plot.New()
	p.Title.Text = "Visualize Path"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"
	err := plotutil.AddLinePoints(p, "Original Path", orig, "Simplified Path", simp)
	if err != nil {
		return err
	}
	length := 14 * vg.Inch
	width := 7 * vg.Inch
	return p.Save(length, width, "bothpaths.png")
}

// This functions saves a png containing only the original path

func SaveOrigPlot(orig plotter.XYs) error {
	pOrig := plot.New()
	pOrig.Title.Text = "Visualize Path"
	pOrig.X.Label.Text = "X"
	pOrig.Y.Label.Text = "Y"
	err := plotutil.AddLinePoints(pOrig, "Original Path", orig)
	if err != nil {
		return err
	}
	length := 14 * vg.Inch
	width := 7 * vg.Inch
	return pOrig.Save(length, width, "origpath.png")
}

// This functions saves a png containing only the Simplified path

func SaveSimpPlot(simp plotter.XYs) error {
	pSimp := plot.New()
	pSimp.Title.Text = "Visualize Path"
	pSimp.X.Label.Text = "X"
	pSimp.Y.Label.Text = "Y"
	err := plotutil.AddLinePoints(pSimp, "Simplified Path", simp)
	if err != nil {
		return err
	}
	length := 14 * vg.Inch
	width := 7 * vg.Inch
	return pSimp.Save(length, width, "simplifiedpath.png")
}

func ToXYs(points []Point) plotter.XYs {
	xys := make(plotter.XYs, len(points))
	for i := range points {
		xys[i].X = points[i].xCoordinate
		xys[i].Y = points[i].yCoordinate
	}

	return xys
}

// Find the perpendicular distance from a given point to a line

func (l ConnectLine) perpendicularDistanceFromPointToLine(pt Point) float64 {
	a, b, c := l.Coefficients()
	return math.Abs(a*pt.xCoordinate+b*pt.yCoordinate+c) / math.Sqrt(a*a+b*b)
}

func (l ConnectLine) Coefficients() (a, b, c float64) {
	a = l.Starting.yCoordinate - l.Ending.yCoordinate
	b = l.Ending.xCoordinate - l.Starting.yCoordinate
	c = l.Starting.xCoordinate*l.Ending.yCoordinate - l.Ending.xCoordinate*l.Starting.yCoordinate
	return a, b, c
}

// Find the farthest point from the given line

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
		left := DownSampleIpdata(ipdata[:farthestPointIndex+1], epsilon)
		right := DownSampleIpdata(ipdata[farthestPointIndex:], epsilon)
		return append(left[:len(left)-1], right...)
	}

	return []Point{ipdata[0], ipdata[len(ipdata)-1]}

}

//Iterative implementation of the RDP algorithm for the given retention rate

func DownSampleIter(ipdata []Point, startIndex int64, lastIndex int64, weights []float64) []float64 {
	var stack [][]int64
	globalStartIndex := startIndex
	globalEndIndex := lastIndex
	stack = append(stack, []int64{startIndex, lastIndex})
	for len(stack) > 0 {
		n := len(stack) - 1
		curr := stack[n]
		stack = stack[:n]
		startIndex := curr[0]
		last_index := curr[1]
		dmax := 0.0
		index := startIndex
		line := ConnectLine{Starting: ipdata[startIndex], Ending: ipdata[lastIndex]}
		for i := index + 1; i < last_index; i++ {

			distance := line.perpendicularDistanceFromPointToLine(ipdata[i])
			if distance > dmax {
				index = i
				dmax = distance
			}

		}
		if dmax > 0.0 {
			weights[index] = dmax
			stack = append(stack, []int64{startIndex, index})
			stack = append(stack, []int64{index, lastIndex})
		}
	}
	weights[globalStartIndex] = math.Inf(1)
	weights[globalEndIndex] = math.Inf(1)
	return weights
}

//Learning the value of epsilon (maxTolerance) if both the epsilon value and the retention rate have not been given

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
	s := math.Sqrt(math.Pow((x2-x1), 2) + math.Pow((y2-y1), 2))
	m := 0.0
	if (x2 - x1) != 0 {
		m = (y2 - y1) / (x2 - x1)
	}
	phi := math.Atan(m)
	tmax := (1 / s) * (math.Abs(math.Sin(phi)) + math.Abs(math.Cos(phi)))
	dophimaxone := math.Atan((1 / s) * (math.Abs(math.Sin(phi) + math.Cos(phi))) * (1 - tmax + math.Pow(tmax, 2)))
	dophimaxtwo := math.Atan((1 / s) * (math.Abs(math.Sin(phi) - math.Cos(phi))) * (1 - tmax + math.Pow(tmax, 2)))
	dophimax := math.Max(dophimaxone, dophimaxtwo)
	window_threshold := s * dophimax
	if farthestPointDistanceFromLine >= window_threshold {
		left := DownSampleIpdataAuto(ipdata[:farthestPointIndex+1])
		right := DownSampleIpdataAuto(ipdata[farthestPointIndex:])
		return append(left[:len(left)-1], right...)
	}

	return []Point{ipdata[0], ipdata[len(ipdata)-1]}

}

// Split x and y into separate arrays given the 2D points

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

func (r *Rdp) Do(vs *array.Float, ts *array.Float) (*array.Int, *array.Float) {
	r.vs = vs
	r.ts = ts
	points := r.ToPoints()
	var sampledPoints []Point
	// if both epsilon and retention rate is given together

	if r.epsilon != 0.0 && r.retentionPercent != 0.0 {
		panic("Please provide either epsilon or retention rate and not both")
	}
	// If epsilon has been given

	if r.epsilon != 0.0 {
		sampledPoints = DownSampleIpdata(points, r.epsilon)
		if err := SavePlot(ToXYs(points), ToXYs(sampledPoints)); err != nil {
			panic(err)
		}
		if err := SaveOrigPlot(ToXYs(points)); err != nil {
			panic(err)
		}
		if err := SaveSimpPlot(ToXYs(sampledPoints)); err != nil {
			panic(err)
		}
		Newts, Newvs := r.SplitCoordinates(sampledPoints)
		return Newts, Newvs
	} else if r.retentionPercent != 0.0 { // if the retention percent is given
		numOfPoints := int64(vs.Len())
		retention := r.retentionPercent
		weights := make([]float64, numOfPoints)
		pointsToKeep := math.Floor((retention / 100) * float64(numOfPoints))
		if pointsToKeep > 0 {
			weights_updated := DownSampleIter(points, 0, numOfPoints-1, weights)
			weights_descending := make([]float64, len(weights_updated))
			copy(weights_descending, weights_updated)
			sort.Sort(sort.Reverse(sort.Float64Slice(weights_descending)))
			maxTolerance := weights_descending[int64(pointsToKeep)-1]
			for i := range points {
				if weights_updated[i] >= maxTolerance {
					sampledPoints = append(sampledPoints, points[i])
				}
			}
			if err := SavePlot(ToXYs(points), ToXYs(sampledPoints)); err != nil {
				panic(err)
			}
			if err := SaveOrigPlot(ToXYs(points)); err != nil {
				panic(err)
			}
			if err := SaveSimpPlot(ToXYs(sampledPoints)); err != nil {
				panic(err)
			}
			Newts, Newvs := r.SplitCoordinates(sampledPoints)
			return Newts, Newvs
		} else {
			panic("The number of points to retain is less than 1. Please increase the retention rate")
		}
	} else { // if both epsilon and retention rate is not given then we try to learn the maximum tolerance by ourselves.
		sampledPoints := DownSampleIpdataAuto(points)
		if err := SavePlot(ToXYs(points), ToXYs(sampledPoints)); err != nil {
			panic(err)
		}
		if err := SaveOrigPlot(ToXYs(points)); err != nil {
			panic(err)
		}
		if err := SaveSimpPlot(ToXYs(sampledPoints)); err != nil {
			panic(err)
		}
		Newts, Newvs := r.SplitCoordinates(sampledPoints)
		return Newts, Newvs
	}
}
