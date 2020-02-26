package geo_test

import "experimental/geo"
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#group,false,false,false,false,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string
#default,_result,,,,,,,,
,result,table,_time,_value,_ci,_field,_measurement,_pt,id
,,0,2019-11-10T11:08:34Z,40.762662,89c258c,lat,bikes,end,vehicleB
,,0,2019-11-10T21:17:47Z,40.762424,89c258c,lat,bikes,end,vehicleB
,,1,2019-11-10T11:07:12Z,40.762096,89c258c,lat,bikes,start,vehicleB
,,1,2019-11-10T21:16:00Z,40.763126,89c258c,lat,bikes,start,vehicleB
,,2,2019-11-10T11:07:35Z,40.762225,89c258c,lat,bikes,via,vehicleB
,,2,2019-11-10T11:07:38Z,40.762247,89c258c,lat,bikes,via,vehicleB
,,2,2019-11-10T11:07:43Z,40.762331,89c258c,lat,bikes,via,vehicleB
,,2,2019-11-10T11:07:48Z,40.762408,89c258c,lat,bikes,via,vehicleB
,,2,2019-11-10T11:07:52Z,40.762484,89c258c,lat,bikes,via,vehicleB
,,2,2019-11-10T11:08:01Z,40.762597,89c258c,lat,bikes,via,vehicleB
,,2,2019-11-10T11:08:16Z,40.762574,89c258c,lat,bikes,via,vehicleB
,,2,2019-11-10T21:16:06Z,40.76309,89c258c,lat,bikes,via,vehicleB
,,2,2019-11-10T21:16:18Z,40.763036,89c258c,lat,bikes,via,vehicleB
,,2,2019-11-10T21:16:31Z,40.763006,89c258c,lat,bikes,via,vehicleB
,,2,2019-11-10T21:16:48Z,40.762904,89c258c,lat,bikes,via,vehicleB
,,2,2019-11-10T21:17:08Z,40.762836,89c258c,lat,bikes,via,vehicleB
,,2,2019-11-10T21:17:23Z,40.762736,89c258c,lat,bikes,via,vehicleB
,,2,2019-11-10T21:17:36Z,40.762469,89c258c,lat,bikes,via,vehicleB
,,2,2019-11-10T21:17:46Z,40.762418,89c258c,lat,bikes,via,vehicleB
,,3,2019-11-10T11:08:34Z,-73.967971,89c258c,lon,bikes,end,vehicleB
,,3,2019-11-10T21:17:47Z,-73.965583,89c258c,lon,bikes,end,vehicleB
,,4,2019-11-10T11:07:12Z,-73.967104,89c258c,lon,bikes,start,vehicleB
,,4,2019-11-10T21:16:00Z,-73.966333,89c258c,lon,bikes,start,vehicleB
,,5,2019-11-10T11:07:35Z,-73.967081,89c258c,lon,bikes,via,vehicleB
,,5,2019-11-10T11:07:38Z,-73.967129,89c258c,lon,bikes,via,vehicleB
,,5,2019-11-10T11:07:43Z,-73.967261,89c258c,lon,bikes,via,vehicleB
,,5,2019-11-10T11:07:48Z,-73.967422,89c258c,lon,bikes,via,vehicleB
,,5,2019-11-10T11:07:52Z,-73.967542,89c258c,lon,bikes,via,vehicleB
,,5,2019-11-10T11:08:01Z,-73.967718,89c258c,lon,bikes,via,vehicleB
,,5,2019-11-10T11:08:16Z,-73.967803,89c258c,lon,bikes,via,vehicleB
,,5,2019-11-10T21:16:06Z,-73.966254,89c258c,lon,bikes,via,vehicleB
,,5,2019-11-10T21:16:18Z,-73.966091,89c258c,lon,bikes,via,vehicleB
,,5,2019-11-10T21:16:31Z,-73.965889,89c258c,lon,bikes,via,vehicleB
,,5,2019-11-10T21:16:48Z,-73.96573,89c258c,lon,bikes,via,vehicleB
,,5,2019-11-10T21:17:08Z,-73.965721,89c258c,lon,bikes,via,vehicleB
,,5,2019-11-10T21:17:23Z,-73.965801,89c258c,lon,bikes,via,vehicleB
,,5,2019-11-10T21:17:36Z,-73.96559,89c258c,lon,bikes,via,vehicleB
,,5,2019-11-10T21:17:46Z,-73.965579,89c258c,lon,bikes,via,vehicleB

#group,false,false,false,false,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string,string
#default,_result,,,,,,,,
,result,table,_time,_value,_ci,_field,_measurement,_pt,id
,,6,2019-11-10T11:08:34Z,1573384032,89c258c,tid,bikes,end,vehicleB
,,6,2019-11-10T21:17:47Z,1573420560,89c258c,tid,bikes,end,vehicleB
,,7,2019-11-10T11:07:12Z,1573384032,89c258c,tid,bikes,start,vehicleB
,,7,2019-11-10T21:16:00Z,1573420560,89c258c,tid,bikes,start,vehicleB
,,8,2019-11-10T11:07:35Z,1573384032,89c258c,tid,bikes,via,vehicleB
,,8,2019-11-10T11:07:38Z,1573384032,89c258c,tid,bikes,via,vehicleB
,,8,2019-11-10T11:07:43Z,1573384032,89c258c,tid,bikes,via,vehicleB
,,8,2019-11-10T11:07:48Z,1573384032,89c258c,tid,bikes,via,vehicleB
,,8,2019-11-10T11:07:52Z,1573384032,89c258c,tid,bikes,via,vehicleB
,,8,2019-11-10T11:08:01Z,1573384032,89c258c,tid,bikes,via,vehicleB
,,8,2019-11-10T11:08:16Z,1573384032,89c258c,tid,bikes,via,vehicleB
,,8,2019-11-10T21:16:06Z,1573420560,89c258c,tid,bikes,via,vehicleB
,,8,2019-11-10T21:16:18Z,1573420560,89c258c,tid,bikes,via,vehicleB
,,8,2019-11-10T21:16:31Z,1573420560,89c258c,tid,bikes,via,vehicleB
,,8,2019-11-10T21:16:48Z,1573420560,89c258c,tid,bikes,via,vehicleB
,,8,2019-11-10T21:17:08Z,1573420560,89c258c,tid,bikes,via,vehicleB
,,8,2019-11-10T21:17:23Z,1573420560,89c258c,tid,bikes,via,vehicleB
,,8,2019-11-10T21:17:36Z,1573420560,89c258c,tid,bikes,via,vehicleB
,,8,2019-11-10T21:17:46Z,1573420560,89c258c,tid,bikes,via,vehicleB

#group,false,false,false,false,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string
#default,_result,,,,,,,,
,result,table,_time,_value,_ci,_field,_measurement,_pt,id
,,9,2019-11-20T10:17:17Z,40.700344,89e82cc,lat,bikes,start,vehicleA
,,10,2019-11-20T10:17:18Z,40.700348,89e82cc,lat,bikes,via,vehicleA
,,10,2019-11-20T10:17:24Z,40.700397,89e82cc,lat,bikes,via,vehicleA
,,10,2019-11-20T10:17:26Z,40.700413,89e82cc,lat,bikes,via,vehicleA
,,10,2019-11-20T10:17:32Z,40.700474,89e82cc,lat,bikes,via,vehicleA
,,10,2019-11-20T10:17:35Z,40.700481,89e82cc,lat,bikes,via,vehicleA
,,10,2019-11-20T10:17:42Z,40.700459,89e82cc,lat,bikes,via,vehicleA
,,10,2019-11-20T10:17:47Z,40.700455,89e82cc,lat,bikes,via,vehicleA
,,10,2019-11-20T10:17:54Z,40.700542,89e82cc,lat,bikes,via,vehicleA
,,11,2019-11-20T10:17:17Z,-73.324814,89e82cc,lon,bikes,start,vehicleA
,,12,2019-11-20T10:17:18Z,-73.324799,89e82cc,lon,bikes,via,vehicleA
,,12,2019-11-20T10:17:24Z,-73.324699,89e82cc,lon,bikes,via,vehicleA
,,12,2019-11-20T10:17:26Z,-73.324638,89e82cc,lon,bikes,via,vehicleA
,,12,2019-11-20T10:17:32Z,-73.324471,89e82cc,lon,bikes,via,vehicleA
,,12,2019-11-20T10:17:35Z,-73.324371,89e82cc,lon,bikes,via,vehicleA
,,12,2019-11-20T10:17:42Z,-73.324181,89e82cc,lon,bikes,via,vehicleA
,,12,2019-11-20T10:17:47Z,-73.323982,89e82cc,lon,bikes,via,vehicleA
,,12,2019-11-20T10:17:54Z,-73.323769,89e82cc,lon,bikes,via,vehicleA

#group,false,false,false,false,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string,string
#default,_result,,,,,,,,
,result,table,_time,_value,_ci,_field,_measurement,_pt,id
,,13,2019-11-20T10:17:17Z,1574245037,89e82cc,tid,bikes,start,vehicleA
,,14,2019-11-20T10:17:18Z,1574245037,89e82cc,tid,bikes,via,vehicleA
,,14,2019-11-20T10:17:24Z,1574245037,89e82cc,tid,bikes,via,vehicleA
,,14,2019-11-20T10:17:26Z,1574245037,89e82cc,tid,bikes,via,vehicleA
,,14,2019-11-20T10:17:32Z,1574245037,89e82cc,tid,bikes,via,vehicleA
,,14,2019-11-20T10:17:35Z,1574245037,89e82cc,tid,bikes,via,vehicleA
,,14,2019-11-20T10:17:42Z,1574245037,89e82cc,tid,bikes,via,vehicleA
,,14,2019-11-20T10:17:47Z,1574245037,89e82cc,tid,bikes,via,vehicleA
,,14,2019-11-20T10:17:54Z,1574245037,89e82cc,tid,bikes,via,vehicleA

#group,false,false,false,false,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string
#default,_result,,,,,,,,
,result,table,_time,_value,_ci,_field,_measurement,_pt,id
,,15,2019-11-20T10:18:00Z,40.700684,89e82d4,lat,bikes,end,vehicleA
,,16,2019-11-20T10:18:00Z,-73.323692,89e82d4,lon,bikes,end,vehicleA

#group,false,false,false,false,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string,string
#default,_result,,,,,,,,
,result,table,_time,_value,_ci,_field,_measurement,_pt,id
,,17,2019-11-20T10:18:00Z,1574245037,89e82d4,tid,bikes,end,vehicleA
"

outData = "
#group,false,false,false,false,false,false,true,false,true,false
#datatype,string,long,dateTime:RFC3339,string,string,string,string,double,long,double
#default,_result,,,,,,,,,
,result,table,_time,_ci,_measurement,_pt,id,lat,tid,lon
,,0,2019-11-20T10:17:17Z,89e82cc,bikes,start,vehicleA,40.700344,1574245037,-73.324814
,,0,2019-11-20T10:17:18Z,89e82cc,bikes,via,vehicleA,40.700348,1574245037,-73.324799
,,0,2019-11-20T10:17:24Z,89e82cc,bikes,via,vehicleA,40.700397,1574245037,-73.324699
,,0,2019-11-20T10:17:26Z,89e82cc,bikes,via,vehicleA,40.700413,1574245037,-73.324638
,,0,2019-11-20T10:17:32Z,89e82cc,bikes,via,vehicleA,40.700474,1574245037,-73.324471
,,0,2019-11-20T10:17:35Z,89e82cc,bikes,via,vehicleA,40.700481,1574245037,-73.324371
,,0,2019-11-20T10:17:42Z,89e82cc,bikes,via,vehicleA,40.700459,1574245037,-73.324181
,,0,2019-11-20T10:17:47Z,89e82cc,bikes,via,vehicleA,40.700455,1574245037,-73.323982
,,0,2019-11-20T10:17:54Z,89e82cc,bikes,via,vehicleA,40.700542,1574245037,-73.323769
,,0,2019-11-20T10:18:00Z,89e82d4,bikes,end,vehicleA,40.700684,1574245037,-73.323692
,,1,2019-11-10T11:07:12Z,89c258c,bikes,start,vehicleB,40.762096,1573384032,-73.967104
,,1,2019-11-10T11:07:35Z,89c258c,bikes,via,vehicleB,40.762225,1573384032,-73.967081
,,1,2019-11-10T11:07:38Z,89c258c,bikes,via,vehicleB,40.762247,1573384032,-73.967129
,,1,2019-11-10T11:07:43Z,89c258c,bikes,via,vehicleB,40.762331,1573384032,-73.967261
,,1,2019-11-10T11:07:48Z,89c258c,bikes,via,vehicleB,40.762408,1573384032,-73.967422
,,1,2019-11-10T11:07:52Z,89c258c,bikes,via,vehicleB,40.762484,1573384032,-73.967542
,,1,2019-11-10T11:08:01Z,89c258c,bikes,via,vehicleB,40.762597,1573384032,-73.967718
,,1,2019-11-10T11:08:16Z,89c258c,bikes,via,vehicleB,40.762574,1573384032,-73.967803
,,1,2019-11-10T11:08:34Z,89c258c,bikes,end,vehicleB,40.762662,1573384032,-73.967971
,,2,2019-11-10T21:16:00Z,89c258c,bikes,start,vehicleB,40.763126,1573420560,-73.966333
,,2,2019-11-10T21:16:06Z,89c258c,bikes,via,vehicleB,40.76309,1573420560,-73.966254
,,2,2019-11-10T21:16:18Z,89c258c,bikes,via,vehicleB,40.763036,1573420560,-73.966091
,,2,2019-11-10T21:16:31Z,89c258c,bikes,via,vehicleB,40.763006,1573420560,-73.965889
,,2,2019-11-10T21:16:48Z,89c258c,bikes,via,vehicleB,40.762904,1573420560,-73.96573
,,2,2019-11-10T21:17:08Z,89c258c,bikes,via,vehicleB,40.762836,1573420560,-73.965721
,,2,2019-11-10T21:17:23Z,89c258c,bikes,via,vehicleB,40.762736,1573420560,-73.965801
,,2,2019-11-10T21:17:36Z,89c258c,bikes,via,vehicleB,40.762469,1573420560,-73.96559
,,2,2019-11-10T21:17:46Z,89c258c,bikes,via,vehicleB,40.762418,1573420560,-73.965579
,,2,2019-11-10T21:17:47Z,89c258c,bikes,end,vehicleB,40.762424,1573420560,-73.965583
"

t_asTracks = (table=<-) =>
  table
    |> range(start: 2019-11-01T00:00:00Z)
    |> geo.toRows(correlationKey: ["id", "_time"])
    |> geo.asTracks()
    |> drop(columns: ["_start", "_stop"])
test _asTracks = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_asTracks})
