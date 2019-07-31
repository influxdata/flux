package holt_winters

import (
	"math"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/internal/mutable"
)

const (
	defaultMaxIterations = 1000
	// reflection coefficient
	defaultAlpha = 1.0
	// contraction coefficient
	defaultBeta = 0.5
	// expansion coefficient
	defaultGamma = 2.0
)

// Optimizer is an implementation of the Nelder-Mead optimization method.
// Based on work by Michael F. Hutt: http://www.mikehutt.com/neldermead.html
type Optimizer struct {
	// Maximum number of iterations.
	MaxIterations int
	// Reflection coefficient.
	Alpha,
	// Contraction coefficient.
	Beta,
	// Expansion coefficient.
	Gamma float64

	alloc memory.Allocator
}

// NewOptimizer returns a new instance of Optimizer with all values set to the defaults.
func NewOptimizer(alloc memory.Allocator) *Optimizer {
	return &Optimizer{
		MaxIterations: defaultMaxIterations,
		Alpha:         defaultAlpha,
		Beta:          defaultBeta,
		Gamma:         defaultGamma,
		alloc:         alloc,
	}
}

// Optimize applies the Nelder-Mead simplex method with the Optimizer's settings.
// Optimize returns a new Float64Array, it is responsibility of the caller to Release it.
func (o *Optimizer) Optimize(
	objfunc func(*mutable.Float64Array) float64,
	start *mutable.Float64Array,
	epsilon,
	scale float64,
) (float64, *mutable.Float64Array) {
	n := start.Len()

	//holds vertices of simplex
	v := make([]*mutable.Float64Array, n+1)
	for i := range v {
		v[i] = mutable.NewFloat64Array(o.alloc)
		v[i].Resize(n)
	}
	defer func() {
		for i := range v {
			v[i].Release()
		}
	}()

	//value of function at each vertex
	f := mutable.NewFloat64Array(o.alloc)
	f.Resize(n + 1)
	defer f.Release()

	//reflection - coordinates
	vr := mutable.NewFloat64Array(o.alloc)
	vr.Resize(n)
	defer vr.Release()

	//expansion - coordinates
	ve := mutable.NewFloat64Array(o.alloc)
	ve.Resize(n)
	defer ve.Release()

	//contraction - coordinates
	vc := mutable.NewFloat64Array(o.alloc)
	vc.Resize(n)
	defer vc.Release()

	//centroid - coordinates
	vm := mutable.NewFloat64Array(o.alloc)
	vm.Resize(n)
	defer vm.Release()

	// create the initial simplex
	// assume one of the vertices is 0,0

	pn := scale * (math.Sqrt(float64(n+1)) - 1 + float64(n)) / (float64(n) * math.Sqrt(2))
	qn := scale * (math.Sqrt(float64(n+1)) - 1) / (float64(n) * math.Sqrt(2))

	for i := 0; i < n; i++ {
		v[0].Set(i, start.Value(i))
	}

	for i := 1; i <= n; i++ {
		for j := 0; j < n; j++ {
			if i-1 == j {
				v[i].Set(j, pn+start.Value(j))
			} else {
				v[i].Set(j, qn+start.Value(j))
			}
		}
	}

	// find the initial function values
	for j := 0; j <= n; j++ {
		f.Set(j, objfunc(v[j]))
	}

	// begin the main loop of the minimization
	for itr := 1; itr <= o.MaxIterations; itr++ {

		// find the indexes of the largest and smallest values
		vg := 0
		vs := 0
		for i := 0; i <= n; i++ {
			if f.Value(i) > f.Value(vg) {
				vg = i
			}
			if f.Value(i) < f.Value(vs) {
				vs = i
			}
		}
		// find the index of the second largest value
		vh := vs
		for i := 0; i <= n; i++ {
			if f.Value(i) > f.Value(vh) && f.Value(i) < f.Value(vg) {
				vh = i
			}
		}

		// calculate the centroid
		for i := 0; i <= n-1; i++ {
			cent := 0.0
			for m := 0; m <= n; m++ {
				if m != vg {
					cent += v[m].Value(i)
				}
			}
			vm.Set(i, cent/float64(n))
		}

		// reflect vg to new vertex vr
		for i := 0; i <= n-1; i++ {
			vr.Set(i, vm.Value(i)+o.Alpha*(vm.Value(i)-v[vg].Value(i)))
		}

		// value of function at reflection point
		fr := objfunc(vr)

		if fr < f.Value(vh) && fr >= f.Value(vs) {
			for i := 0; i <= n-1; i++ {
				v[vg].Set(i, vr.Value(i))
			}
			f.Set(vg, fr)
		}

		// investigate a step further in this direction
		if fr < f.Value(vs) {
			for i := 0; i <= n-1; i++ {
				ve.Set(i, vm.Value(i)+o.Gamma*(vr.Value(i)-vm.Value(i)))
			}

			// value of function at expansion point
			fe := objfunc(ve)

			// by making fe < fr as opposed to fe < f[vs],
			// Rosenbrocks function takes 63 iterations as opposed
			// to 64 when using double variables.

			if fe < fr {
				for i := 0; i <= n-1; i++ {
					v[vg].Set(i, ve.Value(i))
				}
				f.Set(vg, fe)
			} else {
				for i := 0; i <= n-1; i++ {
					v[vg].Set(i, vr.Value(i))
				}
				f.Set(vg, fr)
			}
		}

		// check to see if a contraction is necessary
		if fr >= f.Value(vh) {
			if fr < f.Value(vg) && fr >= f.Value(vh) {
				// perform outside contraction
				for i := 0; i <= n-1; i++ {
					vc.Set(i, vm.Value(i)+o.Beta*(vr.Value(i)-vm.Value(i)))
				}
			} else {
				// perform inside contraction
				for i := 0; i <= n-1; i++ {
					vc.Set(i, vm.Value(i)-o.Beta*(vm.Value(i)-v[vg].Value(i)))
				}
			}

			// value of function at contraction point
			fc := objfunc(vc)

			if fc < f.Value(vg) {
				for i := 0; i <= n-1; i++ {
					v[vg].Set(i, vc.Value(i))
				}
				f.Set(vg, fc)
			} else {
				// at this point the contraction is not successful,
				// we must halve the distance from vs to all the
				// vertices of the simplex and then continue.

				for row := 0; row <= n; row++ {
					if row != vs {
						for i := 0; i <= n-1; i++ {
							v[row].Set(i, v[vs].Value(i)+(v[row].Value(i)-v[vs].Value(i))/2.0)
						}
					}
				}
				f.Set(vg, objfunc(v[vg]))
				f.Set(vh, objfunc(v[vh]))
			}
		}

		// test for convergence
		fsum := 0.0
		for i := 0; i <= n; i++ {
			fsum += f.Value(i)
		}
		favg := fsum / float64(n+1)
		s := 0.0
		for i := 0; i <= n; i++ {
			s += math.Pow(f.Value(i)-favg, 2.0) / float64(n)
		}
		s = math.Sqrt(s)
		if s < epsilon {
			break
		}
	}

	// find the index of the smallest value
	vs := 0
	for i := 0; i <= n; i++ {
		if f.Value(i) < f.Value(vs) {
			vs = i
		}
	}

	parameters := mutable.NewFloat64Array(o.alloc)
	parameters.Resize(n)
	for i := 0; i < n; i++ {
		parameters.Set(i, v[vs].Value(i))
	}

	min := objfunc(v[vs])

	return min, parameters
}
