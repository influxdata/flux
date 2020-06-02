package anomalydetection_test

import "testing"
import "contrib/anaisdg/anomalydetection" 

inData= "
#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,0,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:00Z,18.6,usage_user,cpu,cpu4,Anais.attlocal.net
,,0,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:10Z,25.474525474525475,usage_user,cpu,cpu4,Anais.attlocal.net
,,0,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:20Z,21.878121878121878,usage_user,cpu,cpu4,Anais.attlocal.net
,,0,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:30Z,24.524524524524523,usage_user,cpu,cpu4,Anais.attlocal.net
,,0,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:40Z,21.12112112112112,usage_user,cpu,cpu4,Anais.attlocal.net
,,1,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:00Z,3.1,usage_user,cpu,cpu3,Anais.attlocal.net
,,1,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:10Z,4.795204795204795,usage_user,cpu,cpu3,Anais.attlocal.net
,,1,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:20Z,6.806806806806807,usage_user,cpu,cpu3,Anais.attlocal.net
,,1,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:30Z,4.095904095904096,usage_user,cpu,cpu3,Anais.attlocal.net
,,1,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:40Z,3,usage_user,cpu,cpu3,Anais.attlocal.net
,,2,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:00Z,21.8,usage_user,cpu,cpu0,Anais.attlocal.net
,,2,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:10Z,29.87012987012987,usage_user,cpu,cpu0,Anais.attlocal.net
,,2,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:20Z,25.725725725725727,usage_user,cpu,cpu0,Anais.attlocal.net
,,2,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:30Z,28,usage_user,cpu,cpu0,Anais.attlocal.net
,,2,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:40Z,25.8,usage_user,cpu,cpu0,Anais.attlocal.net
,,3,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:00Z,2.9,usage_user,cpu,cpu7,Anais.attlocal.net
,,3,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:10Z,4.6,usage_user,cpu,cpu7,Anais.attlocal.net
,,3,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:20Z,6.793206793206793,usage_user,cpu,cpu7,Anais.attlocal.net
,,3,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:30Z,4.004004004004004,usage_user,cpu,cpu7,Anais.attlocal.net
,,3,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:40Z,2.9,usage_user,cpu,cpu7,Anais.attlocal.net
,,4,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:00Z,3.4,usage_user,cpu,cpu1,Anais.attlocal.net
,,4,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:10Z,4.7,usage_user,cpu,cpu1,Anais.attlocal.net
,,4,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:20Z,7.3,usage_user,cpu,cpu1,Anais.attlocal.net
,,4,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:30Z,4.104104104104104,usage_user,cpu,cpu1,Anais.attlocal.net
,,4,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:40Z,3.196803196803197,usage_user,cpu,cpu1,Anais.attlocal.net
,,5,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:00Z,16.6,usage_user,cpu,cpu6,Anais.attlocal.net
,,5,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:10Z,24.6,usage_user,cpu,cpu6,Anais.attlocal.net
,,5,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:20Z,20.4,usage_user,cpu,cpu6,Anais.attlocal.net
,,5,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:30Z,21.7,usage_user,cpu,cpu6,Anais.attlocal.net
,,5,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:40Z,18.81881881881882,usage_user,cpu,cpu6,Anais.attlocal.net
,,6,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:00Z,19.660678642714572,usage_user,cpu,cpu2,Anais.attlocal.net
,,6,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:10Z,26.726726726726728,usage_user,cpu,cpu2,Anais.attlocal.net
,,6,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:20Z,23.3,usage_user,cpu,cpu2,Anais.attlocal.net
,,6,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:30Z,26,usage_user,cpu,cpu2,Anais.attlocal.net
,,6,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:40Z,22.9,usage_user,cpu,cpu2,Anais.attlocal.net
,,7,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:00Z,11.169415292353824,usage_user,cpu,cpu-total,Anais.attlocal.net
,,7,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:10Z,15.67108222944264,usage_user,cpu,cpu-total,Anais.attlocal.net
,,7,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:20Z,14.889361170146268,usage_user,cpu,cpu-total,Anais.attlocal.net
,,7,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:30Z,14.5625,usage_user,cpu,cpu-total,Anais.attlocal.net
,,7,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:40Z,12.567212704764287,usage_user,cpu,cpu-total,Anais.attlocal.net
,,8,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:00Z,3.2967032967032965,usage_user,cpu,cpu5,Anais.attlocal.net
,,8,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:10Z,4.6,usage_user,cpu,cpu5,Anais.attlocal.net
,,8,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:20Z,6.9,usage_user,cpu,cpu5,Anais.attlocal.net
,,8,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:30Z,3.996003996003996,usage_user,cpu,cpu5,Anais.attlocal.net
,,8,2020-04-29T21:04:59.916661Z,2020-05-29T21:04:59.916661Z,2020-05-06T16:13:40Z,2.9029029029029028,usage_user,cpu,cpu5,Anais.attlocal.net
"

outData = "
#group,false,false,false,true,false,false
#datatype,string,long,double,dateTime:RFC3339,double,string
#default,_result,,,,,
,result,table,MAD,_time,_value,level
,,0,11.67208280475147,2020-05-06T16:13:00Z,0.6744907594765952,normal
,,1,16.265726513371657,2020-05-06T16:13:10Z,0.6744907594765952,normal
,,2,11.845026870858856,2020-05-06T16:13:20Z,0.6744907594765952,normal
,,3,15.517775087412586,2020-05-06T16:13:30Z,0.6744907594765952,normal
,,4,14.184349556083532,2020-05-06T16:13:40Z,0.6744907594765952,normal
"

t_mad = (table=<-) =>
	table
		|> range(start: 2020-04-29T21:04:00Z, stop: 2020-05-29T21:05:00Z)
		|> anomalyDetection.mad(threshold: 3.0)


test _mad = () =>
({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_mad})
