package rdp

import "math"

type Point struct {
	X_coordinate float64
	Y_coordinate float64
}

type ConnectLine struct {
	Starting Point
	Ending   Point
}

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
