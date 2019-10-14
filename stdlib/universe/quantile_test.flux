package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,Reiva,LgT6,BnR,2019-01-09T19:44:58Z,k5Uym
,,0,Reiva,LgT6,BnR,2019-01-09T19:45:08Z,csheb
,,0,Reiva,LgT6,BnR,2019-01-09T19:45:18Z,xUPF
,,0,Reiva,LgT6,BnR,2019-01-09T19:45:28Z,fJTWEh
,,0,Reiva,LgT6,BnR,2019-01-09T19:45:38Z,oF7g
,,0,Reiva,LgT6,BnR,2019-01-09T19:45:48Z,NvfS
,,1,Reiva,LgT6,qCnJDC,2019-01-09T19:44:58Z,eWoKiN
,,1,Reiva,LgT6,qCnJDC,2019-01-09T19:45:08Z,oE4S
,,1,Reiva,LgT6,qCnJDC,2019-01-09T19:45:18Z,mRC
,,1,Reiva,LgT6,qCnJDC,2019-01-09T19:45:28Z,SNwh8
,,1,Reiva,LgT6,qCnJDC,2019-01-09T19:45:38Z,pwH
,,1,Reiva,LgT6,qCnJDC,2019-01-09T19:45:48Z,jmJqsA

#datatype,string,long,string,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,2,Reiva,OAOJWe7,BnR,2019-01-09T19:44:58Z,7.940387008821781
,,2,Reiva,OAOJWe7,BnR,2019-01-09T19:45:08Z,49.460104214779086
,,2,Reiva,OAOJWe7,BnR,2019-01-09T19:45:18Z,-36.564150808873954
,,2,Reiva,OAOJWe7,BnR,2019-01-09T19:45:28Z,34.319039251798635
,,2,Reiva,OAOJWe7,BnR,2019-01-09T19:45:38Z,79.27019811403116
,,2,Reiva,OAOJWe7,BnR,2019-01-09T19:45:48Z,41.91029522104053
,,3,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:44:58Z,-61.68790887989735
,,3,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:08Z,-6.3173755351186465
,,3,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:18Z,-26.049728557657513
,,3,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:28Z,114.285955884979
,,3,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:38Z,16.140262630578995
,,3,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:48Z,29.50336437998469

#datatype,string,long,string,string,string,dateTime:RFC3339,long
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,4,Reiva,dGpnr,BnR,2019-01-09T19:44:58Z,17
,,4,Reiva,dGpnr,BnR,2019-01-09T19:45:08Z,-44
,,4,Reiva,dGpnr,BnR,2019-01-09T19:45:18Z,-99
,,4,Reiva,dGpnr,BnR,2019-01-09T19:45:28Z,-85
,,4,Reiva,dGpnr,BnR,2019-01-09T19:45:38Z,18
,,4,Reiva,dGpnr,BnR,2019-01-09T19:45:48Z,99
,,5,Reiva,dGpnr,qCnJDC,2019-01-09T19:44:58Z,-44
,,5,Reiva,dGpnr,qCnJDC,2019-01-09T19:45:08Z,-25
,,5,Reiva,dGpnr,qCnJDC,2019-01-09T19:45:18Z,46
,,5,Reiva,dGpnr,qCnJDC,2019-01-09T19:45:28Z,-2
,,5,Reiva,dGpnr,qCnJDC,2019-01-09T19:45:38Z,-14
,,5,Reiva,dGpnr,qCnJDC,2019-01-09T19:45:48Z,-53

#datatype,string,long,string,string,string,dateTime:RFC3339,boolean
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,6,Reiva,rREO,BnR,2019-01-09T19:44:58Z,false
,,6,Reiva,rREO,BnR,2019-01-09T19:45:08Z,true
,,6,Reiva,rREO,BnR,2019-01-09T19:45:18Z,false
,,6,Reiva,rREO,BnR,2019-01-09T19:45:28Z,true
,,6,Reiva,rREO,BnR,2019-01-09T19:45:38Z,false
,,6,Reiva,rREO,BnR,2019-01-09T19:45:48Z,true
,,7,Reiva,rREO,qCnJDC,2019-01-09T19:44:58Z,true
,,7,Reiva,rREO,qCnJDC,2019-01-09T19:45:08Z,true
,,7,Reiva,rREO,qCnJDC,2019-01-09T19:45:18Z,true
,,7,Reiva,rREO,qCnJDC,2019-01-09T19:45:28Z,false
,,7,Reiva,rREO,qCnJDC,2019-01-09T19:45:38Z,false
,,7,Reiva,rREO,qCnJDC,2019-01-09T19:45:48Z,false

#datatype,string,long,string,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,8,Reiva,rc2iOD1,BnR,2019-01-09T19:44:58Z,79
,,8,Reiva,rc2iOD1,BnR,2019-01-09T19:45:08Z,33
,,8,Reiva,rc2iOD1,BnR,2019-01-09T19:45:18Z,97
,,8,Reiva,rc2iOD1,BnR,2019-01-09T19:45:28Z,90
,,8,Reiva,rc2iOD1,BnR,2019-01-09T19:45:38Z,96
,,8,Reiva,rc2iOD1,BnR,2019-01-09T19:45:48Z,10
,,9,Reiva,rc2iOD1,qCnJDC,2019-01-09T19:44:58Z,84
,,9,Reiva,rc2iOD1,qCnJDC,2019-01-09T19:45:08Z,52
,,9,Reiva,rc2iOD1,qCnJDC,2019-01-09T19:45:18Z,23
,,9,Reiva,rc2iOD1,qCnJDC,2019-01-09T19:45:28Z,62
,,9,Reiva,rc2iOD1,qCnJDC,2019-01-09T19:45:38Z,22
,,9,Reiva,rc2iOD1,qCnJDC,2019-01-09T19:45:48Z,78
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2019-01-01T00:00:00Z,2030-01-01T00:00:00Z,Reiva,LgT6,BnR,2019-01-09T19:45:38Z,oF7g
,,1,2019-01-01T00:00:00Z,2030-01-01T00:00:00Z,Reiva,LgT6,qCnJDC,2019-01-09T19:45:08Z,oE4S
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,2,2019-01-01T00:00:00Z,2030-01-01T00:00:00Z,Reiva,OAOJWe7,BnR,2019-01-09T19:45:08Z,49.460104214779086
,,3,2019-01-01T00:00:00Z,2030-01-01T00:00:00Z,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:48Z,29.50336437998469
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,long
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,4,2019-01-01T00:00:00Z,2030-01-01T00:00:00Z,Reiva,dGpnr,BnR,2019-01-09T19:45:38Z,18
,,5,2019-01-01T00:00:00Z,2030-01-01T00:00:00Z,Reiva,dGpnr,qCnJDC,2019-01-09T19:45:28Z,-2
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,boolean
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,6,2019-01-01T00:00:00Z,2030-01-01T00:00:00Z,Reiva,rREO,BnR,2019-01-09T19:45:28Z,true
,,7,2019-01-01T00:00:00Z,2030-01-01T00:00:00Z,Reiva,rREO,qCnJDC,2019-01-09T19:45:08Z,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,8,2019-01-01T00:00:00Z,2030-01-01T00:00:00Z,Reiva,rc2iOD1,BnR,2019-01-09T19:45:38Z,96
,,9,2019-01-01T00:00:00Z,2030-01-01T00:00:00Z,Reiva,rc2iOD1,qCnJDC,2019-01-09T19:45:48Z,78
"

t_quantile = (table=<-) =>
	(table
		|> range(start: 2019-01-01T00:00:00Z)
		|> quantile(q: 0.75, method: "exact_selector"))

test _quantile = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_quantile})

