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
	retentionpercent float64
	vs               *array.Float
	ts               *array.Float
	alloc            memory.Allocator
}

func New(timeColumn string, column string, threshold float64, retentionpercent float64, alloc memory.Allocator) *Rdp {
	return &Rdp{
		timeColumn:       timeColumn,
		column:           column,
		epsilon:          threshold,
		retentionpercent: retentionpercent,
		alloc:            alloc,
	}
}

type Point struct {
	X_coordinate float64
	Y_coordinate float64
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

	return p.Save(14*vg.Inch, 7*vg.Inch, "2paths.png")
}

// This functions saves a png containing only the original path

func SaveOrigPlot(orig plotter.XYs) error {
	p_orig := plot.New()
	p_orig.Title.Text = "Visualize Path"
	p_orig.X.Label.Text = "X"
	p_orig.Y.Label.Text = "Y"
	err := plotutil.AddLinePoints(p_orig, "Original Path", orig)
	if err != nil {
		return err
	}

	return p_orig.Save(14*vg.Inch, 7*vg.Inch, "orig_path.png")
}

// This functions saves a png containing only the Simplified path

func SaveSimpPlot(simp plotter.XYs) error {
	p_simp := plot.New()
	p_simp.Title.Text = "Visualize Path"
	p_simp.X.Label.Text = "X"
	p_simp.Y.Label.Text = "Y"
	err := plotutil.AddLinePoints(p_simp, "Simplified Path", simp)
	if err != nil {
		return err
	}

	return p_simp.Save(14*vg.Inch, 7*vg.Inch, "simplified_path.png")
}

func ToXYs(points []Point) plotter.XYs {
	xys := make(plotter.XYs, len(points))
	for i := range points {
		xys[i].X = points[i].X_coordinate
		xys[i].Y = points[i].Y_coordinate
	}

	return xys
}

// Find the perpendicular distance from a given point to a line

func (l ConnectLine) perpendicularDistanceFromPointToLine(pt Point) float64 {
	a, b, c := l.Coefficients()
	return math.Abs(a*pt.X_coordinate+b*pt.Y_coordinate+c) / math.Sqrt(a*a+b*b)
}

func (l ConnectLine) Coefficients() (a, b, c float64) {
	a = l.Starting.Y_coordinate - l.Ending.Y_coordinate
	b = l.Ending.X_coordinate - l.Starting.X_coordinate
	c = l.Starting.X_coordinate*l.Ending.Y_coordinate - l.Ending.X_coordinate*l.Starting.Y_coordinate
	return a, b, c
}

// Find the farthest point from the given line

func FindFarthestPoint(line ConnectLine, ipdata []Point) (farthest_point_index int, farthest_point_distance_from_line float64) {
	for i := 0; i < len(ipdata); i++ {
		distance := line.perpendicularDistanceFromPointToLine(ipdata[i])
		if distance > farthest_point_distance_from_line {
			farthest_point_distance_from_line = distance
			farthest_point_index = i
		}
	}
	return farthest_point_index, farthest_point_distance_from_line
}

func (r *Rdp) ToPoints() []Point {
	points := make([]Point, 0, r.vs.Len())
	for i := 0; i < r.vs.Len(); i++ {
		points = append(points, Point{X_coordinate: r.ts.Value(i), Y_coordinate: r.vs.Value(i)})
	}

	return points
}

// Recursive implementation of the RDP Algorithm given the epsilon value

func DownSampleIpdata(ipdata []Point, epsilon float64) []Point {
	if len(ipdata) <= 2 {
		return ipdata
	}
	line := ConnectLine{Starting: ipdata[0], Ending: ipdata[len(ipdata)-1]}

	farthest_point_index, farthest_point_distance_from_line := FindFarthestPoint(line, ipdata)

	if farthest_point_distance_from_line >= epsilon {
		left := DownSampleIpdata(ipdata[:farthest_point_index+1], epsilon)
		right := DownSampleIpdata(ipdata[farthest_point_index:], epsilon)
		return append(left[:len(left)-1], right...)
	}

	return []Point{ipdata[0], ipdata[len(ipdata)-1]}

}

//Iterative implementation of the RDP algorithm for the given retention rate

func DownSample_iter(ipdata []Point, start_index int64, last_index int64, weights []float64) []float64 {
	var stack [][]int64
	global_start_index := start_index
	global_end_index := last_index
	stack = append(stack, []int64{start_index, last_index})
	for len(stack) > 0 {
		n := len(stack) - 1
		curr := stack[n]
		stack = stack[:n]
		start_index := curr[0]
		last_index := curr[1]
		dmax := 0.0
		index := start_index
		line := ConnectLine{Starting: ipdata[start_index], Ending: ipdata[last_index]}
		for i := index + 1; i < last_index; i++ {

			distance := line.perpendicularDistanceFromPointToLine(ipdata[i])
			if distance > dmax {
				index = i
				dmax = distance
			}

		}
		if dmax > 0.0 {
			weights[index] = dmax
			stack = append(stack, []int64{start_index, index})
			stack = append(stack, []int64{index, last_index})
		}
	}
	weights[global_start_index] = math.Inf(1)
	weights[global_end_index] = math.Inf(1)
	return weights
}

//Learning the value of epsilon (maxTolerance) if both the epsilon value and the retention rate have not been given

func DownSampleIpdata_auto(ipdata []Point) []Point {
	if len(ipdata) <= 2 {
		return ipdata
	}
	line := ConnectLine{Starting: ipdata[0], Ending: ipdata[len(ipdata)-1]}

	farthest_point_index, farthest_point_distance_from_line := FindFarthestPoint(line, ipdata)
	x1 := ipdata[0].X_coordinate
	x2 := ipdata[len(ipdata)-1].X_coordinate
	y1 := ipdata[0].Y_coordinate
	y2 := ipdata[len(ipdata)-1].Y_coordinate
	s := math.Sqrt(math.Pow((x2-x1), 2) + math.Pow((y2-y1), 2))
	m := 0.0
	if (x2 - x1) != 0 {
		m = (y2 - y1) / (x2 - x1)
	}
	phi := math.Atan(m)
	tmax := (1 / s) * (math.Abs(math.Sin(phi)) + math.Abs(math.Cos(phi)))
	dophimax1 := math.Atan((1 / s) * (math.Abs(math.Sin(phi) + math.Cos(phi))) * (1 - tmax + math.Pow(tmax, 2)))
	dophimax2 := math.Atan((1 / s) * (math.Abs(math.Sin(phi) - math.Cos(phi))) * (1 - tmax + math.Pow(tmax, 2)))
	dophimax := math.Max(dophimax1, dophimax2)
	window_threshold := s * dophimax
	if farthest_point_distance_from_line >= window_threshold {
		left := DownSampleIpdata_auto(ipdata[:farthest_point_index+1])
		right := DownSampleIpdata_auto(ipdata[farthest_point_index:])
		return append(left[:len(left)-1], right...)
	}

	return []Point{ipdata[0], ipdata[len(ipdata)-1]}

}

// Split x and y into separate arrays given the 2D points

func (r *Rdp) SplitCoordinates(sampled_points []Point) (*array.Int, *array.Float) {
	newts := array.NewIntBuilder(r.alloc)
	newvs := array.NewFloatBuilder(r.alloc)
	for i := range sampled_points {
		newts.Append(int64(sampled_points[i].X_coordinate))
		newvs.Append(sampled_points[i].Y_coordinate)
	}
	return newts.NewIntArray(), newvs.NewFloatArray()
}

// Main implementation

func (r *Rdp) Do(vs *array.Float, ts *array.Float) (*array.Int, *array.Float) {
	r.vs = vs
	r.ts = ts
	points := r.ToPoints()
	var sampled_points []Point
	// if both epsilon and retention rate is given together

	if r.epsilon != 0.0 && r.retentionpercent != 0.0 {
		panic("Please provide either epsilon or retention rate and not both")
	}
	// If epsilon has been given

	if r.epsilon != 0.0 {
		sampled_points = DownSampleIpdata(points, r.epsilon)
		if err := SavePlot(ToXYs(points), ToXYs(sampled_points)); err != nil {
			panic(err)
		}
		if err := SaveOrigPlot(ToXYs(points)); err != nil {
			panic(err)
		}
		if err := SaveSimpPlot(ToXYs(sampled_points)); err != nil {
			panic(err)
		}
		Newts, Newvs := r.SplitCoordinates(sampled_points)
		return Newts, Newvs
	} else if r.retentionpercent != 0.0 { // if the retention percent is given
		num_of_points := int64(vs.Len())
		retention := r.retentionpercent
		weights := make([]float64, num_of_points)
		pointsToKeep := math.Floor((retention / 100) * float64(num_of_points))
		if pointsToKeep > 0 {
			weights_updated := DownSample_iter(points, 0, num_of_points-1, weights)
			weights_descending := make([]float64, len(weights_updated))
			copy(weights_descending, weights_updated)
			sort.Sort(sort.Reverse(sort.Float64Slice(weights_descending)))
			maxTolerance := weights_descending[int64(pointsToKeep)-1]
			for i := range points {
				if weights_updated[i] >= maxTolerance {
					sampled_points = append(sampled_points, points[i])
				}
			}
			if err := SavePlot(ToXYs(points), ToXYs(sampled_points)); err != nil {
				panic(err)
			}
			if err := SaveOrigPlot(ToXYs(points)); err != nil {
				panic(err)
			}
			if err := SaveSimpPlot(ToXYs(sampled_points)); err != nil {
				panic(err)
			}
			Newts, Newvs := r.SplitCoordinates(sampled_points)
			return Newts, Newvs
		} else {
			panic("The number of points to retain is less than 1. Please increase the retention rate")
		}
	} else { // if both epsilon and retention rate is not given then we try to learn the maximum tolerance by ourselves.
		sampled_points := DownSampleIpdata_auto(points)
		if err := SavePlot(ToXYs(points), ToXYs(sampled_points)); err != nil {
			panic(err)
		}
		if err := SaveOrigPlot(ToXYs(points)); err != nil {
			panic(err)
		}
		if err := SaveSimpPlot(ToXYs(sampled_points)); err != nil {
			panic(err)
		}
		Newts, Newvs := r.SplitCoordinates(sampled_points)
		return Newts, Newvs
	}
}
