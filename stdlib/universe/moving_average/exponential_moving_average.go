package moving_average

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/values"
)

type ExponentialMovingAverage struct {
	inTimePeriod  int
	i             []int
	count         []float64
	value         []float64
	periodReached []bool
	lastVal       []interface{}

	Multiplier float64
	ema        [][]interface{}
}

func New(inTimePeriod int, numCols int) *ExponentialMovingAverage {
	return &ExponentialMovingAverage{
		inTimePeriod:  inTimePeriod,
		i:             make([]int, numCols),
		count:         make([]float64, numCols),
		value:         make([]float64, numCols),
		periodReached: make([]bool, numCols),
		lastVal:       make([]interface{}, numCols),
		Multiplier:    2 / float64(inTimePeriod+1),
		ema:           make([][]interface{}, numCols),
	}
}

func (r *ExponentialMovingAverage) PassThrough(vs *ArrayContainer, b execute.TableBuilder, bj int) error {
	j := 0

	for ; r.i[bj] < r.inTimePeriod && j < vs.Len(); r.i[bj]++ {
		if vs.IsNull(j) {
			r.lastVal[bj] = nil
		} else {
			r.lastVal[bj] = vs.OrigValue(j)
		}
		j++
	}

	if r.i[bj] == r.inTimePeriod && !r.periodReached[bj] {
		if vs.IsNull(j - 1) {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		} else {
			if err := b.AppendValue(bj, values.New(vs.OrigValue(j-1))); err != nil {
				return err
			}
		}
		r.periodReached[bj] = true
	}

	for ; r.i[bj] >= r.inTimePeriod && j < vs.Len(); r.i[bj]++ {
		if vs.IsNull(j) {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		} else {
			if err := b.AppendValue(bj, values.New(vs.OrigValue(j))); err != nil {
				return err
			}
		}
		j++
	}
	return nil
}

func (r *ExponentialMovingAverage) DoNumeric(vs *ArrayContainer, b execute.TableBuilder, bj int, doExponentialMovingAverage bool, appendToTable bool) error {
	if !doExponentialMovingAverage {
		return r.PassThrough(vs, b, bj)
	}

	var appendVal func(v float64) error
	var appendNil func() error
	if appendToTable {
		appendVal = func(v float64) error {
			if err := b.AppendFloat(bj, v); err != nil {
				return err
			}
			return nil
		}
		appendNil = func() error {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
			return nil
		}
	} else {
		appendVal = func(v float64) error {
			r.ema[bj] = append(r.ema[bj], v)
			return nil
		}
		appendNil = func() error {
			r.ema[bj] = append(r.ema[bj], nil)
			return nil
		}
	}

	j := 0

	// Build the first value of the EMA
	for ; r.i[bj] < r.inTimePeriod && j < vs.Len(); r.i[bj]++ {
		if !vs.IsNull(j) {
			r.value[bj] += vs.Value(j).Float()
			r.count[bj]++
			r.lastVal[bj] = vs.OrigValue(j)
		} else {
			r.lastVal[bj] = nil
		}
		j++
	}

	// Append the first value of the EMA
	if r.i[bj] == r.inTimePeriod && !r.periodReached[bj] {
		if r.count[bj] != 0 {
			r.value[bj] = r.value[bj] / r.count[bj]
			if err := appendVal(r.value[bj]); err != nil {
				return err
			}
		} else {
			if err := appendNil(); err != nil {
				return err
			}
		}
		r.periodReached[bj] = true
	}

	l := vs.Len()
	for ; j < l; j++ {
		if vs.IsNull(j) {
			if r.count[bj] == 0 {
				if err := appendNil(); err != nil {
					return err
				}
			} else {
				if err := appendVal(r.value[bj]); err != nil {
					return err
				}
			}
		} else {
			cValue := vs.Value(j).Float()
			var ema float64
			if r.count[bj] == 0 {
				ema = cValue
				r.count[bj]++
			} else {
				ema = (cValue * r.Multiplier) + (r.value[bj] * (1.0 - r.Multiplier))
			}
			if err := appendVal(ema); err != nil {
				return err
			}
			r.value[bj] = ema
		}
		r.i[bj]++
	}
	return nil
}

func (r *ExponentialMovingAverage) PassThroughTime(vs *array.Int64, b execute.TableBuilder, bj int) error {
	j := 0

	for ; r.i[bj] < r.inTimePeriod && j < vs.Len(); r.i[bj]++ {
		if vs.IsNull(j) {
			r.lastVal[bj] = nil
		} else {
			r.lastVal[bj] = execute.Time(vs.Value(j))
		}
		j++
	}

	if r.i[bj] == r.inTimePeriod && !r.periodReached[bj] {
		if vs.IsNull(j - 1) {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		} else {
			if err := b.AppendTime(bj, execute.Time(vs.Value(j-1))); err != nil {
				return err
			}
		}
		r.periodReached[bj] = true
	}

	for ; r.i[bj] >= r.inTimePeriod && j < vs.Len(); r.i[bj]++ {
		if vs.IsNull(j) {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		} else {
			if err := b.AppendTime(bj, execute.Time(vs.Value(j))); err != nil {
				return err
			}
		}
		j++
	}
	return nil
}

func (r *ExponentialMovingAverage) GetEMA(bj int) []interface{} {
	return r.ema[bj]
}

// If we don't have enough values for a proper EMA, just append the last value (which is the average of the values so far)
func (r *ExponentialMovingAverage) Finish(cols []flux.ColMeta, builder execute.TableBuilder, doExponentialMovingAverage []bool) error {
	for j := range cols {
		if !r.periodReached[j] {
			if !doExponentialMovingAverage[j] {
				if r.lastVal[j] == nil {
					if err := builder.AppendNil(j); err != nil {
						return err
					}
				} else {
					if err := builder.AppendValue(j, values.New(r.lastVal[j])); err != nil {
						return err
					}
				}
			} else {
				if r.count[j] != 0 {
					average := r.value[j] / r.count[j]
					if err := builder.AppendFloat(j, average); err != nil {
						return err
					}
				} else {
					if err := builder.AppendNil(j); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
