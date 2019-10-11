package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,iZquGj,Au1iY,2018-12-18T20:52:33Z,7
,,0,iZquGj,Au1iY,2018-12-18T20:52:43Z,38
,,0,iZquGj,Au1iY,2018-12-18T20:52:53Z,79
,,0,iZquGj,Au1iY,2018-12-18T20:53:03Z,51
,,0,iZquGj,Au1iY,2018-12-18T20:53:13Z,94
,,0,iZquGj,Au1iY,2018-12-18T20:53:23Z,85

#datatype,string,long,string,string,dateTime:RFC3339,string
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,1,iZquGj,HSbYC,2018-12-18T20:52:33Z,A9JcEV
,,1,iZquGj,HSbYC,2018-12-18T20:52:43Z,iNI7Bqy
,,1,iZquGj,HSbYC,2018-12-18T20:52:53Z,TFIS
,,1,iZquGj,HSbYC,2018-12-18T20:53:03Z,q6h9yU
,,1,iZquGj,HSbYC,2018-12-18T20:53:13Z,X8Ks
,,1,iZquGj,HSbYC,2018-12-18T20:53:23Z,aOMgU

#datatype,string,long,string,string,dateTime:RFC3339,boolean
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,2,iZquGj,J1u,2018-12-18T20:52:33Z,false
,,2,iZquGj,J1u,2018-12-18T20:52:43Z,false
,,2,iZquGj,J1u,2018-12-18T20:52:53Z,true
,,2,iZquGj,J1u,2018-12-18T20:53:03Z,true
,,2,iZquGj,J1u,2018-12-18T20:53:13Z,true
,,2,iZquGj,J1u,2018-12-18T20:53:23Z,false

#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,3,iZquGj,ei77f8T,2018-12-18T20:52:33Z,-61.68790887989735
,,3,iZquGj,ei77f8T,2018-12-18T20:52:43Z,-6.3173755351186465
,,3,iZquGj,ei77f8T,2018-12-18T20:52:53Z,-26.049728557657513
,,3,iZquGj,ei77f8T,2018-12-18T20:53:03Z,114.285955884979
,,3,iZquGj,ei77f8T,2018-12-18T20:53:13Z,16.140262630578995
,,3,iZquGj,ei77f8T,2018-12-18T20:53:23Z,29.50336437998469

#datatype,string,long,string,string,dateTime:RFC3339,long
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,4,iZquGj,ucyoZ,2018-12-18T20:52:33Z,-66
,,4,iZquGj,ucyoZ,2018-12-18T20:52:43Z,59
,,4,iZquGj,ucyoZ,2018-12-18T20:52:53Z,64
,,4,iZquGj,ucyoZ,2018-12-18T20:53:03Z,84
,,4,iZquGj,ucyoZ,2018-12-18T20:53:13Z,68
,,4,iZquGj,ucyoZ,2018-12-18T20:53:23Z,49
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,long
#group,false,false,true,true,true,true,false
#default,_result,,,,,,
,result,table,_start,_stop,_measurement,_field,_value
,,0,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,iZquGj,Au1iY,6
,,1,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,iZquGj,HSbYC,6
,,2,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,iZquGj,J1u,6
,,3,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,iZquGj,ei77f8T,6
,,4,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,iZquGj,ucyoZ,6
"

t_count = (table=<-) =>
	(table
		|> range(start: 2018-12-01T00:00:00Z)
		|> count())

test _count = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_count})

