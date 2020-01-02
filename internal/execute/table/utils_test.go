package table_test

import "github.com/influxdata/flux"

type TableIterator []flux.Table

func (t TableIterator) Do(f func(flux.Table) error) error {
	for _, tbl := range t {
		if err := f(tbl); err != nil {
			return err
		}
	}
	return nil
}
