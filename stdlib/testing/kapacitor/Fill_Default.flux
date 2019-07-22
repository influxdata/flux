package testdata_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m1,f1,server01,2018-12-19T22:13:30Z,
,,0,m1,f1,server01,2018-12-19T22:13:40Z,banana
,,0,m1,f1,server01,2018-12-19T22:13:50Z,apple
,,0,m1,f1,server01,2018-12-19T22:14:00Z,peach
,,0,m1,f1,server01,2018-12-19T22:14:10Z,
,,0,m1,f1,server01,2018-12-19T22:14:20Z,raspberry
,,1,m1,f1,server02,2018-12-19T22:13:30Z,watermelon
,,1,m1,f1,server02,2018-12-19T22:13:40Z,grape
,,1,m1,f1,server02,2018-12-19T22:13:50Z,
,,1,m1,f1,server02,2018-12-19T22:14:00Z,mango
,,1,m1,f1,server02,2018-12-19T22:14:10Z,orange
,,1,m1,f1,server02,2018-12-19T22:14:20Z,
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:30Z,tomato
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:40Z,banana
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:50Z,apple
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:00Z,peach
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:10Z,tomato
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:20Z,raspberry
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:30Z,watermelon
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:40Z,grape
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:50Z,tomato
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:00Z,mango
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:10Z,orange
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:20Z,tomato
"

t_fill_string = (table=<-) =>
	(table
		|> range(start: 2018-12-15T00:00:00Z)
		|> fill(value: "tomato"))

test _fill = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_fill_string})

// Equivalent TICKscript query:
// stream
//   |default()
//     .field('_value', tomato)