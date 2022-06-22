package main

import (
	"math"
	"math/rand"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func SavePlot(orig, simp plotter.XYs) error {
	p := plot.New()
	p.Title.Text = "Visualize Path"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	err := plotutil.AddLinePoints(p, "Original Path", orig, "Simplified Path", simp)
	if err != nil {
		return err
	}

	return p.Save(14*vg.Inch, 7*vg.Inch, "path.png")
}

func RandomXYs(n int, scale float64) plotter.XYs {
	xys := make(plotter.XYs, n)
	increment := float64(2*math.Pi) / float64(n)

	for i := range xys {
		xys[i].X = float64(i) * increment
		xys[i].Y = math.Sin(xys[i].X) + scale*rand.Float64()
	}

	return xys
}
