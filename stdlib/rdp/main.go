package main

import (
	"fmt"

	rdp "github.com/influxdata/flux/stdlib/rdp/lib"
	"gonum.org/v1/plot/plotter"
)

const Threshold = 1

func ToPoints(xys plotter.XYs) []rdp.Point {
	points := make([]rdp.Point, 0, len(xys))
	for i := range xys {
		points = append(points, rdp.Point{X_coordinate: xys[i].X, Y_coordinate: xys[i].Y})
	}

	return points
}

func ToXYs(points []rdp.Point) plotter.XYs {
	fmt.Println("The number of downsampled points is:", len(points))
	xys := make(plotter.XYs, len(points))
	for i := range points {
		xys[i].X = points[i].X_coordinate
		xys[i].Y = points[i].Y_coordinate
	}

	return xys
}

func main() {
	xys := RandomXYs(200, 0.5)
	simXYs := ToXYs(rdp.DownSampleIpdata(ToPoints(xys), Threshold))
	mask := rdp.DownSampleIpdata_iter(ToPoints(xys), 0, 199, Threshold)
	fmt.Print(mask)
	fmt.Println("Saving plot...")
	if err := SavePlot(xys, simXYs); err != nil {
		panic(err)
	}
}
