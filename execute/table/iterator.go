package table

import "github.com/mvn-trinhnguyen2-dn/flux"

type Iterator []flux.Table

func (t Iterator) Do(f func(flux.Table) error) error {
	for _, tbl := range t {
		if err := f(tbl); err != nil {
			return err
		}
	}
	return nil
}
