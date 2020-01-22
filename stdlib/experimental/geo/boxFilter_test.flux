package geo_test

import "experimental/geo"
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#group,false,false,false,false,true,true,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,string,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_time,_value,_field,_g1,_g2,_g3,_g4,_g5,_measurement,_pt,id
,,0,2019-11-10T11:08:34Z,dr5ruud5tqgq,geohash,d,dr,dr5,dr5r,dr5ru,bike,end,vehicleB
,,0,2019-11-10T21:17:47Z,dr5ruuefkfxf,geohash,d,dr,dr5,dr5r,dr5ru,bike,end,vehicleB
,,1,2019-11-10T11:07:12Z,dr5ruudb3t3t,geohash,d,dr,dr5,dr5r,dr5ru,bike,start,vehicleB
,,1,2019-11-10T21:16:00Z,dr5ruueq7t2w,geohash,d,dr,dr5,dr5r,dr5ru,bike,start,vehicleB
,,2,2019-11-10T11:07:35Z,dr5ruudc4j6z,geohash,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,2,2019-11-10T11:07:38Z,dr5ruudc2cx4,geohash,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,2,2019-11-10T11:07:43Z,dr5ruud9vchz,geohash,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,2,2019-11-10T11:07:48Z,dr5ruudd4pse,geohash,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,2,2019-11-10T11:07:52Z,dr5ruud6xmg0,geohash,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,2,2019-11-10T11:08:01Z,dr5ruud774zf,geohash,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,2,2019-11-10T11:08:16Z,dr5ruud71q89,geohash,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,2,2019-11-10T21:16:06Z,dr5ruueqjqex,geohash,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,2,2019-11-10T21:16:18Z,dr5ruuetchtj,geohash,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,2,2019-11-10T21:16:31Z,dr5ruuettz5d,geohash,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,2,2019-11-10T21:16:48Z,dr5ruuev1s43,geohash,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,2,2019-11-10T21:17:08Z,dr5ruueu9z2u,geohash,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,2,2019-11-10T21:17:23Z,dr5ruuesputr,geohash,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,2,2019-11-10T21:17:36Z,dr5ruuefsffw,geohash,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,2,2019-11-10T21:17:46Z,dr5ruuefm16y,geohash,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,3,2019-11-20T10:18:00Z,dr5ze3r6djkt,geohash,d,dr,dr5,dr5z,dr5ze,bike,end,vehicleA
,,4,2019-11-20T10:17:17Z,dr5ze3q2xypb,geohash,d,dr,dr5,dr5z,dr5ze,bike,start,vehicleA
,,5,2019-11-20T10:17:18Z,dr5ze3q88qg0,geohash,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,5,2019-11-20T10:17:24Z,dr5ze3q8fxvt,geohash,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,5,2019-11-20T10:17:26Z,dr5ze3q9h4gh,geohash,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,5,2019-11-20T10:17:32Z,dr5ze3q9ryr8,geohash,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,5,2019-11-20T10:17:35Z,dr5ze3qc6rd4,geohash,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,5,2019-11-20T10:17:42Z,dr5ze3qcqery,geohash,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,5,2019-11-20T10:17:47Z,dr5ze3r176sr,geohash,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,5,2019-11-20T10:17:54Z,dr5ze3r3b6gn,geohash,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA

#group,false,false,false,false,true,true,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_time,_value,_field,_g1,_g2,_g3,_g4,_g5,_measurement,_pt,id
,,6,2019-11-10T11:08:34Z,40.762662,lat,d,dr,dr5,dr5r,dr5ru,bike,end,vehicleB
,,6,2019-11-10T21:17:47Z,40.762424,lat,d,dr,dr5,dr5r,dr5ru,bike,end,vehicleB
,,7,2019-11-10T11:07:12Z,40.762096,lat,d,dr,dr5,dr5r,dr5ru,bike,start,vehicleB
,,7,2019-11-10T21:16:00Z,40.763126,lat,d,dr,dr5,dr5r,dr5ru,bike,start,vehicleB
,,8,2019-11-10T11:07:35Z,40.762225,lat,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,8,2019-11-10T11:07:38Z,40.762247,lat,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,8,2019-11-10T11:07:43Z,40.762331,lat,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,8,2019-11-10T11:07:48Z,40.762408,lat,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,8,2019-11-10T11:07:52Z,40.762484,lat,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,8,2019-11-10T11:08:01Z,40.762597,lat,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,8,2019-11-10T11:08:16Z,40.762574,lat,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,8,2019-11-10T21:16:06Z,40.76309,lat,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,8,2019-11-10T21:16:18Z,40.763036,lat,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,8,2019-11-10T21:16:31Z,40.763006,lat,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,8,2019-11-10T21:16:48Z,40.762904,lat,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,8,2019-11-10T21:17:08Z,40.762836,lat,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,8,2019-11-10T21:17:23Z,40.762736,lat,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,8,2019-11-10T21:17:36Z,40.762469,lat,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,8,2019-11-10T21:17:46Z,40.762418,lat,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,9,2019-11-20T10:18:00Z,40.700684,lat,d,dr,dr5,dr5z,dr5ze,bike,end,vehicleA
,,10,2019-11-20T10:17:17Z,40.700344,lat,d,dr,dr5,dr5z,dr5ze,bike,start,vehicleA
,,11,2019-11-20T10:17:18Z,40.700348,lat,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,11,2019-11-20T10:17:24Z,40.700397,lat,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,11,2019-11-20T10:17:26Z,40.700413,lat,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,11,2019-11-20T10:17:32Z,40.700474,lat,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,11,2019-11-20T10:17:35Z,40.700481,lat,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,11,2019-11-20T10:17:42Z,40.700459,lat,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,11,2019-11-20T10:17:47Z,40.700455,lat,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,11,2019-11-20T10:17:54Z,40.700542,lat,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,12,2019-11-10T11:08:34Z,-73.967971,lon,d,dr,dr5,dr5r,dr5ru,bike,end,vehicleB
,,12,2019-11-10T21:17:47Z,-73.965583,lon,d,dr,dr5,dr5r,dr5ru,bike,end,vehicleB
,,13,2019-11-10T11:07:12Z,-73.967104,lon,d,dr,dr5,dr5r,dr5ru,bike,start,vehicleB
,,13,2019-11-10T21:16:00Z,-73.966333,lon,d,dr,dr5,dr5r,dr5ru,bike,start,vehicleB
,,14,2019-11-10T11:07:35Z,-73.967081,lon,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,14,2019-11-10T11:07:38Z,-73.967129,lon,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,14,2019-11-10T11:07:43Z,-73.967261,lon,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,14,2019-11-10T11:07:48Z,-73.967422,lon,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,14,2019-11-10T11:07:52Z,-73.967542,lon,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,14,2019-11-10T11:08:01Z,-73.967718,lon,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,14,2019-11-10T11:08:16Z,-73.967803,lon,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,14,2019-11-10T21:16:06Z,-73.966254,lon,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,14,2019-11-10T21:16:18Z,-73.966091,lon,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,14,2019-11-10T21:16:31Z,-73.965889,lon,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,14,2019-11-10T21:16:48Z,-73.96573,lon,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,14,2019-11-10T21:17:08Z,-73.965721,lon,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,14,2019-11-10T21:17:23Z,-73.965801,lon,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,14,2019-11-10T21:17:36Z,-73.96559,lon,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,14,2019-11-10T21:17:46Z,-73.965579,lon,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,15,2019-11-20T10:18:00Z,-73.323692,lon,d,dr,dr5,dr5z,dr5ze,bike,end,vehicleA
,,16,2019-11-20T10:17:17Z,-73.324814,lon,d,dr,dr5,dr5z,dr5ze,bike,start,vehicleA
,,17,2019-11-20T10:17:18Z,-73.324799,lon,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,17,2019-11-20T10:17:24Z,-73.324699,lon,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,17,2019-11-20T10:17:26Z,-73.324638,lon,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,17,2019-11-20T10:17:32Z,-73.324471,lon,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,17,2019-11-20T10:17:35Z,-73.324371,lon,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,17,2019-11-20T10:17:42Z,-73.324181,lon,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,17,2019-11-20T10:17:47Z,-73.323982,lon,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,17,2019-11-20T10:17:54Z,-73.323769,lon,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA

#group,false,false,false,false,true,true,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_time,_value,_field,_g1,_g2,_g3,_g4,_g5,_measurement,_pt,id
,,18,2019-11-10T11:08:34Z,1573384032000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,end,vehicleB
,,18,2019-11-10T21:17:47Z,1573420560000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,end,vehicleB
,,19,2019-11-10T11:07:12Z,1573384032000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,start,vehicleB
,,19,2019-11-10T21:16:00Z,1573420560000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,start,vehicleB
,,20,2019-11-10T11:07:35Z,1573384032000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,20,2019-11-10T11:07:38Z,1573384032000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,20,2019-11-10T11:07:43Z,1573384032000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,20,2019-11-10T11:07:48Z,1573384032000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,20,2019-11-10T11:07:52Z,1573384032000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,20,2019-11-10T11:08:01Z,1573384032000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,20,2019-11-10T11:08:16Z,1573384032000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,20,2019-11-10T21:16:06Z,1573420560000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,20,2019-11-10T21:16:18Z,1573420560000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,20,2019-11-10T21:16:31Z,1573420560000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,20,2019-11-10T21:16:48Z,1573420560000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,20,2019-11-10T21:17:08Z,1573420560000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,20,2019-11-10T21:17:23Z,1573420560000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,20,2019-11-10T21:17:36Z,1573420560000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,20,2019-11-10T21:17:46Z,1573420560000000000,tid,d,dr,dr5,dr5r,dr5ru,bike,via,vehicleB
,,21,2019-11-20T10:18:00Z,1574245037000000000,tid,d,dr,dr5,dr5z,dr5ze,bike,end,vehicleA
,,22,2019-11-20T10:17:17Z,1574245037000000000,tid,d,dr,dr5,dr5z,dr5ze,bike,start,vehicleA
,,23,2019-11-20T10:17:18Z,1574245037000000000,tid,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,23,2019-11-20T10:17:24Z,1574245037000000000,tid,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,23,2019-11-20T10:17:26Z,1574245037000000000,tid,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,23,2019-11-20T10:17:32Z,1574245037000000000,tid,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,23,2019-11-20T10:17:35Z,1574245037000000000,tid,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,23,2019-11-20T10:17:42Z,1574245037000000000,tid,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,23,2019-11-20T10:17:47Z,1574245037000000000,tid,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
,,23,2019-11-20T10:17:54Z,1574245037000000000,tid,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA
"

outData = "
#group,false,false,false,true,true,true,true,true,true,true,true,false,false,false,false
#datatype,string,long,dateTime:RFC3339,string,string,string,string,string,string,string,string,string,double,double,long
#default,_result,,,,,,,,,,,,,,
,result,table,_time,_g1,_g2,_g3,_g4,_g5,_measurement,_pt,id,geohash,lat,lon,tid
,,0,2019-11-20T10:18:00Z,d,dr,dr5,dr5z,dr5ze,bike,end,vehicleA,dr5ze3r6djkt,40.700684,-73.323692,1574245037000000000
,,1,2019-11-20T10:17:17Z,d,dr,dr5,dr5z,dr5ze,bike,start,vehicleA,dr5ze3q2xypb,40.700344,-73.324814,1574245037000000000
,,2,2019-11-20T10:17:18Z,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA,dr5ze3q88qg0,40.700348,-73.324799,1574245037000000000
,,2,2019-11-20T10:17:24Z,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA,dr5ze3q8fxvt,40.700397,-73.324699,1574245037000000000
,,2,2019-11-20T10:17:26Z,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA,dr5ze3q9h4gh,40.700413,-73.324638,1574245037000000000
,,2,2019-11-20T10:17:32Z,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA,dr5ze3q9ryr8,40.700474,-73.324471,1574245037000000000
,,2,2019-11-20T10:17:35Z,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA,dr5ze3qc6rd4,40.700481,-73.324371,1574245037000000000
,,2,2019-11-20T10:17:42Z,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA,dr5ze3qcqery,40.700459,-73.324181,1574245037000000000
,,2,2019-11-20T10:17:47Z,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA,dr5ze3r176sr,40.700455,-73.323982,1574245037000000000
,,2,2019-11-20T10:17:54Z,d,dr,dr5,dr5z,dr5ze,bike,via,vehicleA,dr5ze3r3b6gn,40.700542,-73.323769,1574245037000000000
"

t_boxFilter = (table=<-) =>
	table
		|> range(start: 2019-11-01T00:00:00Z)
		|> filter(fn: (r) => r._measurement == "bike")
		|> geo.toRows()
		|> geo.boxFilter(box: {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875})
		|> drop(columns: ["_start","_stop"])

test _boxFilter = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_boxFilter})
