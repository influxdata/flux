package main
import "testing"

option now = () =>
	(2030-01-01T00:00:00Z)

inData = "
#datatype,string,string
#group,true,true
#default,,
,error,reference
,failed to execute query: failed to initialize execute state: EOF,
"
outData = "err: error calling function "

filter
": name "
null
" does not exist in scope
"

t_null_as_value = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> filter(fn: (r) =>
			(r._value == null)))

test null_as_value = {input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_null_as_value}