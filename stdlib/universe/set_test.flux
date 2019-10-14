package universe_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,string,dateTime:RFC3339,boolean
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,thmWJ,GK1Ji,SbkiNS9,2018-12-19T22:13:30Z,false
,,0,thmWJ,GK1Ji,SbkiNS9,2018-12-19T22:13:40Z,true
,,0,thmWJ,GK1Ji,SbkiNS9,2018-12-19T22:13:50Z,false
,,0,thmWJ,GK1Ji,SbkiNS9,2018-12-19T22:14:00Z,false
,,0,thmWJ,GK1Ji,SbkiNS9,2018-12-19T22:14:10Z,true
,,0,thmWJ,GK1Ji,SbkiNS9,2018-12-19T22:14:20Z,true
,,1,thmWJ,GK1Ji,gpmhNEw,2018-12-19T22:13:30Z,false
,,1,thmWJ,GK1Ji,gpmhNEw,2018-12-19T22:13:40Z,true
,,1,thmWJ,GK1Ji,gpmhNEw,2018-12-19T22:13:50Z,false
,,1,thmWJ,GK1Ji,gpmhNEw,2018-12-19T22:14:00Z,true
,,1,thmWJ,GK1Ji,gpmhNEw,2018-12-19T22:14:10Z,true
,,1,thmWJ,GK1Ji,gpmhNEw,2018-12-19T22:14:20Z,true

#datatype,string,long,string,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,2,thmWJ,c9wjx7r,SbkiNS9,2018-12-19T22:13:30Z,-61.68790887989735
,,2,thmWJ,c9wjx7r,SbkiNS9,2018-12-19T22:13:40Z,-6.3173755351186465
,,2,thmWJ,c9wjx7r,SbkiNS9,2018-12-19T22:13:50Z,-26.049728557657513
,,2,thmWJ,c9wjx7r,SbkiNS9,2018-12-19T22:14:00Z,114.285955884979
,,2,thmWJ,c9wjx7r,SbkiNS9,2018-12-19T22:14:10Z,16.140262630578995
,,2,thmWJ,c9wjx7r,SbkiNS9,2018-12-19T22:14:20Z,29.50336437998469
,,3,thmWJ,c9wjx7r,gpmhNEw,2018-12-19T22:13:30Z,7.940387008821781
,,3,thmWJ,c9wjx7r,gpmhNEw,2018-12-19T22:13:40Z,49.460104214779086
,,3,thmWJ,c9wjx7r,gpmhNEw,2018-12-19T22:13:50Z,-36.564150808873954
,,3,thmWJ,c9wjx7r,gpmhNEw,2018-12-19T22:14:00Z,34.319039251798635
,,3,thmWJ,c9wjx7r,gpmhNEw,2018-12-19T22:14:10Z,79.27019811403116
,,3,thmWJ,c9wjx7r,gpmhNEw,2018-12-19T22:14:20Z,41.91029522104053

#datatype,string,long,string,string,string,dateTime:RFC3339,long
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,4,thmWJ,iUcIq,SbkiNS9,2018-12-19T22:13:30Z,-44
,,4,thmWJ,iUcIq,SbkiNS9,2018-12-19T22:13:40Z,-25
,,4,thmWJ,iUcIq,SbkiNS9,2018-12-19T22:13:50Z,46
,,4,thmWJ,iUcIq,SbkiNS9,2018-12-19T22:14:00Z,-2
,,4,thmWJ,iUcIq,SbkiNS9,2018-12-19T22:14:10Z,-14
,,4,thmWJ,iUcIq,SbkiNS9,2018-12-19T22:14:20Z,-53
,,5,thmWJ,iUcIq,gpmhNEw,2018-12-19T22:13:30Z,17
,,5,thmWJ,iUcIq,gpmhNEw,2018-12-19T22:13:40Z,-44
,,5,thmWJ,iUcIq,gpmhNEw,2018-12-19T22:13:50Z,-99
,,5,thmWJ,iUcIq,gpmhNEw,2018-12-19T22:14:00Z,-85
,,5,thmWJ,iUcIq,gpmhNEw,2018-12-19T22:14:10Z,18
,,5,thmWJ,iUcIq,gpmhNEw,2018-12-19T22:14:20Z,99

#datatype,string,long,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,6,thmWJ,urO72,SbkiNS9,2018-12-19T22:13:30Z,xRbS
,,6,thmWJ,urO72,SbkiNS9,2018-12-19T22:13:40Z,PtTh
,,6,thmWJ,urO72,SbkiNS9,2018-12-19T22:13:50Z,ZjN2je
,,6,thmWJ,urO72,SbkiNS9,2018-12-19T22:14:00Z,YZNBh
,,6,thmWJ,urO72,SbkiNS9,2018-12-19T22:14:10Z,pu08
,,6,thmWJ,urO72,SbkiNS9,2018-12-19T22:14:20Z,ixlOdT
,,7,thmWJ,urO72,gpmhNEw,2018-12-19T22:13:30Z,YqV
,,7,thmWJ,urO72,gpmhNEw,2018-12-19T22:13:40Z,GjbWF
,,7,thmWJ,urO72,gpmhNEw,2018-12-19T22:13:50Z,GiX1Bb
,,7,thmWJ,urO72,gpmhNEw,2018-12-19T22:14:00Z,DQCZXZ
,,7,thmWJ,urO72,gpmhNEw,2018-12-19T22:14:10Z,atopRR2
,,7,thmWJ,urO72,gpmhNEw,2018-12-19T22:14:20Z,TNKKB

#datatype,string,long,string,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,8,thmWJ,zmk1YWi,SbkiNS9,2018-12-19T22:13:30Z,84
,,8,thmWJ,zmk1YWi,SbkiNS9,2018-12-19T22:13:40Z,52
,,8,thmWJ,zmk1YWi,SbkiNS9,2018-12-19T22:13:50Z,23
,,8,thmWJ,zmk1YWi,SbkiNS9,2018-12-19T22:14:00Z,62
,,8,thmWJ,zmk1YWi,SbkiNS9,2018-12-19T22:14:10Z,22
,,8,thmWJ,zmk1YWi,SbkiNS9,2018-12-19T22:14:20Z,78
,,9,thmWJ,zmk1YWi,gpmhNEw,2018-12-19T22:13:30Z,79
,,9,thmWJ,zmk1YWi,gpmhNEw,2018-12-19T22:13:40Z,33
,,9,thmWJ,zmk1YWi,gpmhNEw,2018-12-19T22:13:50Z,97
,,9,thmWJ,zmk1YWi,gpmhNEw,2018-12-19T22:14:00Z,90
,,9,thmWJ,zmk1YWi,gpmhNEw,2018-12-19T22:14:10Z,96
,,9,thmWJ,zmk1YWi,gpmhNEw,2018-12-19T22:14:20Z,10
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,boolean
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,GK1Ji,server01,2018-12-19T22:13:30Z,false
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,GK1Ji,server01,2018-12-19T22:13:40Z,true
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,GK1Ji,server01,2018-12-19T22:13:50Z,false
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,GK1Ji,server01,2018-12-19T22:14:00Z,false
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,GK1Ji,server01,2018-12-19T22:14:10Z,true
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,GK1Ji,server01,2018-12-19T22:14:20Z,true
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,GK1Ji,server01,2018-12-19T22:13:30Z,false
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,GK1Ji,server01,2018-12-19T22:13:40Z,true
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,GK1Ji,server01,2018-12-19T22:13:50Z,false
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,GK1Ji,server01,2018-12-19T22:14:00Z,true
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,GK1Ji,server01,2018-12-19T22:14:10Z,true
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,GK1Ji,server01,2018-12-19T22:14:20Z,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,c9wjx7r,server01,2018-12-19T22:13:30Z,-61.68790887989735
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,c9wjx7r,server01,2018-12-19T22:13:40Z,-6.3173755351186465
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,c9wjx7r,server01,2018-12-19T22:13:50Z,-26.049728557657513
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,c9wjx7r,server01,2018-12-19T22:14:00Z,114.285955884979
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,c9wjx7r,server01,2018-12-19T22:14:10Z,16.140262630578995
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,c9wjx7r,server01,2018-12-19T22:14:20Z,29.50336437998469
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,c9wjx7r,server01,2018-12-19T22:13:30Z,7.940387008821781
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,c9wjx7r,server01,2018-12-19T22:13:40Z,49.460104214779086
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,c9wjx7r,server01,2018-12-19T22:13:50Z,-36.564150808873954
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,c9wjx7r,server01,2018-12-19T22:14:00Z,34.319039251798635
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,c9wjx7r,server01,2018-12-19T22:14:10Z,79.27019811403116
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,c9wjx7r,server01,2018-12-19T22:14:20Z,41.91029522104053
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,long
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,iUcIq,server01,2018-12-19T22:13:30Z,-44
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,iUcIq,server01,2018-12-19T22:13:40Z,-25
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,iUcIq,server01,2018-12-19T22:13:50Z,46
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,iUcIq,server01,2018-12-19T22:14:00Z,-2
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,iUcIq,server01,2018-12-19T22:14:10Z,-14
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,iUcIq,server01,2018-12-19T22:14:20Z,-53
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,iUcIq,server01,2018-12-19T22:13:30Z,17
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,iUcIq,server01,2018-12-19T22:13:40Z,-44
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,iUcIq,server01,2018-12-19T22:13:50Z,-99
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,iUcIq,server01,2018-12-19T22:14:00Z,-85
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,iUcIq,server01,2018-12-19T22:14:10Z,18
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,iUcIq,server01,2018-12-19T22:14:20Z,99
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,urO72,server01,2018-12-19T22:13:30Z,xRbS
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,urO72,server01,2018-12-19T22:13:40Z,PtTh
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,urO72,server01,2018-12-19T22:13:50Z,ZjN2je
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,urO72,server01,2018-12-19T22:14:00Z,YZNBh
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,urO72,server01,2018-12-19T22:14:10Z,pu08
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,urO72,server01,2018-12-19T22:14:20Z,ixlOdT
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,urO72,server01,2018-12-19T22:13:30Z,YqV
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,urO72,server01,2018-12-19T22:13:40Z,GjbWF
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,urO72,server01,2018-12-19T22:13:50Z,GiX1Bb
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,urO72,server01,2018-12-19T22:14:00Z,DQCZXZ
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,urO72,server01,2018-12-19T22:14:10Z,atopRR2
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,urO72,server01,2018-12-19T22:14:20Z,TNKKB
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,4,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,zmk1YWi,server01,2018-12-19T22:13:30Z,84
,,4,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,zmk1YWi,server01,2018-12-19T22:13:40Z,52
,,4,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,zmk1YWi,server01,2018-12-19T22:13:50Z,23
,,4,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,zmk1YWi,server01,2018-12-19T22:14:00Z,62
,,4,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,zmk1YWi,server01,2018-12-19T22:14:10Z,22
,,4,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,zmk1YWi,server01,2018-12-19T22:14:20Z,78
,,4,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,zmk1YWi,server01,2018-12-19T22:13:30Z,79
,,4,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,zmk1YWi,server01,2018-12-19T22:13:40Z,33
,,4,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,zmk1YWi,server01,2018-12-19T22:13:50Z,97
,,4,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,zmk1YWi,server01,2018-12-19T22:14:00Z,90
,,4,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,zmk1YWi,server01,2018-12-19T22:14:10Z,96
,,4,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,thmWJ,zmk1YWi,server01,2018-12-19T22:14:20Z,10
"

t_set = (table=<-) =>
	(table
		|> range(start: 2018-01-01T00:00:00Z)
		|> set(key: "t0", value: "server01"))

test _set = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_set})

