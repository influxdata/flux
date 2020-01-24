package influxql_test

import "testing"
import "internal/influxql"

option now = () => 2019-10-24T19:07:30Z

inData = "
#datatype,string,long,string,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,cpu,value,Duzw4c,2019-10-24T19:06:30Z,-61.68790887989735
,,0,cpu,value,Duzw4c,2019-10-24T19:06:40Z,-6.3173755351186465
,,0,cpu,value,Duzw4c,2019-10-24T19:06:50Z,-26.049728557657513
,,0,cpu,value,Duzw4c,2019-10-24T19:07:00Z,114.285955884979
,,0,cpu,value,Duzw4c,2019-10-24T19:07:10Z,16.140262630578995
,,0,cpu,value,Duzw4c,2019-10-24T19:07:20Z,29.50336437998469
,,1,cpu,value,EmU470,2019-10-24T19:06:30Z,49.48552101042658
,,1,cpu,value,EmU470,2019-10-24T19:06:40Z,-30.761174263888247
,,1,cpu,value,EmU470,2019-10-24T19:06:50Z,-71.75234610661141
,,1,cpu,value,EmU470,2019-10-24T19:07:00Z,-107.57183413713223
,,1,cpu,value,EmU470,2019-10-24T19:07:10Z,6.867518678667539
,,1,cpu,value,EmU470,2019-10-24T19:07:20Z,22.14113135132833
,,2,cpu,value,LbQrlPU,2019-10-24T19:06:30Z,126.51192216762033
,,2,cpu,value,LbQrlPU,2019-10-24T19:06:40Z,18.551103465915904
,,2,cpu,value,LbQrlPU,2019-10-24T19:06:50Z,-63.51213592477466
,,2,cpu,value,LbQrlPU,2019-10-24T19:07:00Z,-108.9405569292533
,,2,cpu,value,LbQrlPU,2019-10-24T19:07:10Z,-21.031390631066174
,,2,cpu,value,LbQrlPU,2019-10-24T19:07:20Z,-42.87508368305759
,,3,cpu,value,PHtSS,2019-10-24T19:06:30Z,-38.81066049368835
,,3,cpu,value,PHtSS,2019-10-24T19:06:40Z,33.5971884838324
,,3,cpu,value,PHtSS,2019-10-24T19:06:50Z,71.013974839444
,,3,cpu,value,PHtSS,2019-10-24T19:07:00Z,-18.438923037368127
,,3,cpu,value,PHtSS,2019-10-24T19:07:10Z,-16.52283812374848
,,3,cpu,value,PHtSS,2019-10-24T19:07:20Z,-3.201704561305113
,,4,cpu,value,b3C6Do,2019-10-24T19:06:30Z,0.5003098992419464
,,4,cpu,value,b3C6Do,2019-10-24T19:06:40Z,-3.590624775919784
,,4,cpu,value,b3C6Do,2019-10-24T19:06:50Z,-22.91564241031196
,,4,cpu,value,b3C6Do,2019-10-24T19:07:00Z,-108.34620004260354
,,4,cpu,value,b3C6Do,2019-10-24T19:07:10Z,-23.82287894830945
,,4,cpu,value,b3C6Do,2019-10-24T19:07:20Z,-27.865458085515993
,,5,cpu,value,n69gsUs,2019-10-24T19:06:30Z,35.13751307810906
,,5,cpu,value,n69gsUs,2019-10-24T19:06:40Z,17.306944807744788
,,5,cpu,value,n69gsUs,2019-10-24T19:06:50Z,-53.477270117651024
,,5,cpu,value,n69gsUs,2019-10-24T19:07:00Z,-41.66259545948701
,,5,cpu,value,n69gsUs,2019-10-24T19:07:10Z,16.514469280633087
,,5,cpu,value,n69gsUs,2019-10-24T19:07:20Z,87.28708658889737
,,6,cpu,value,pMA,2019-10-24T19:06:30Z,-41.92963812596751
,,6,cpu,value,pMA,2019-10-24T19:06:40Z,5.071048992474987
,,6,cpu,value,pMA,2019-10-24T19:06:50Z,-70.15463641868325
,,6,cpu,value,pMA,2019-10-24T19:07:00Z,-116.32149276007675
,,6,cpu,value,pMA,2019-10-24T19:07:10Z,48.71415819443108
,,6,cpu,value,pMA,2019-10-24T19:07:20Z,12.330309153464318
,,7,cpu,value,sDjZtMO,2019-10-24T19:06:30Z,64.94204237587171
,,7,cpu,value,sDjZtMO,2019-10-24T19:06:40Z,26.367919652993084
,,7,cpu,value,sDjZtMO,2019-10-24T19:06:50Z,36.622096290225656
,,7,cpu,value,sDjZtMO,2019-10-24T19:07:00Z,-53.65899105443762
,,7,cpu,value,sDjZtMO,2019-10-24T19:07:10Z,35.00604512199924
,,7,cpu,value,sDjZtMO,2019-10-24T19:07:20Z,21.57653593480266
,,8,cpu,value,unBUUi,2019-10-24T19:06:30Z,25.505805524300808
,,8,cpu,value,unBUUi,2019-10-24T19:06:40Z,-23.606840424793397
,,8,cpu,value,unBUUi,2019-10-24T19:06:50Z,12.57700220295746
,,8,cpu,value,unBUUi,2019-10-24T19:07:00Z,-4.096873576678456
,,8,cpu,value,unBUUi,2019-10-24T19:07:10Z,8.334535945914878
,,8,cpu,value,unBUUi,2019-10-24T19:07:20Z,18.202255926527776
,,9,cpu,value,wT37nhV,2019-10-24T19:06:30Z,-52.5832648801972
,,9,cpu,value,wT37nhV,2019-10-24T19:06:40Z,-24.81736985563042
,,9,cpu,value,wT37nhV,2019-10-24T19:06:50Z,-33.876177313156965
,,9,cpu,value,wT37nhV,2019-10-24T19:07:00Z,-94.85213297484836
,,9,cpu,value,wT37nhV,2019-10-24T19:07:10Z,-98.08189555424542
,,9,cpu,value,wT37nhV,2019-10-24T19:07:20Z,-36.77222399347868
"

outData = "
#datatype,string,long,string,dateTime:RFC3339,long
#group,false,false,true,false,false
#default,_result,,,,
,result,table,_measurement,time,count
,,0,cpu,1970-01-01T00:00:00Z,60
"

// SELECT count("value") FROM cpu
t_aggregate_count = (tables=<-) => tables
	|> range(start: influxql.minTime, stop: influxql.maxTime)
	|> filter(fn: (r) => r._measurement == "cpu")
	|> filter(fn: (r) => r._field == "value")
	|> group(columns: ["_measurement", "_field"], mode: "by")
	|> count()
	|> map(fn: (r) => ({r with time: influxql.epoch, count: r._value}))
	|> drop(columns: ["_time", "_start", "_stop", "_field", "_value"])

test _aggregate_count = () => ({
	input: testing.loadStorage(csv: inData),
	want: testing.loadMem(csv: outData),
	fn: t_aggregate_count,
})
