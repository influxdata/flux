package transformations

import (
	"github.com/influxdata/flux"
)

func init() {
	flux.RegisterBuiltIn("helpers", helpersBuiltIn)
}

var helpersBuiltIn = `
// AggregateWindow applies an aggregate function to fixed windows of time.
aggregateWindow = (every, period=0s, column, fn=(column,table=<-) => table, table=<-) =>
	table
		|> window(every:every,period:period)
		|> fn(column:column)
		|> window(every:inf)
		|> duplicate(column:"_stop",as:"_time")
`
