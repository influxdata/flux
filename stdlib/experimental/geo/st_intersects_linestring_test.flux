package geo_test


import "experimental/geo"
import "influxdata/influxdb/v1"
import "testing"

option now = () => 2030-01-01T00:00:00Z

inData = "
#group,false,false,false,false,true,true,true,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,,
,result,table,_time,_value,_field,_measurement,_pt,area,id,s2_cell_id,seq_idx,status,stop_id,trip_id
,,0,2020-04-08T15:44:58Z,40.820317,lat,mta,start,LLIR,GO506_20_6431,89c288c54,1,STOPPED_AT,171,GO506_20_6431
,,1,2020-04-08T16:19:27Z,40.745249,lat,mta,via,LLIR,GO506_20_6431,89c2592bc,13,IN_TRANSIT_TO,237,GO506_20_6431
,,2,2020-04-08T16:16:50Z,40.751085,lat,mta,via,LLIR,GO506_20_6431,89c25f18c,13,IN_TRANSIT_TO,237,GO506_20_6431
,,3,2020-04-08T16:15:59Z,40.748141,lat,mta,via,LLIR,GO506_20_6431,89c25f1c4,12,STOPPED_AT,214,GO506_20_6431
,,4,2020-04-08T16:16:16Z,40.748909,lat,mta,via,LLIR,GO506_20_6431,89c25f1c4,13,IN_TRANSIT_TO,237,GO506_20_6431
,,5,2020-04-08T16:14:32Z,40.743989,lat,mta,via,LLIR,GO506_20_6431,89c25f1d4,12,IN_TRANSIT_TO,214,GO506_20_6431
,,6,2020-04-08T16:14:49Z,40.745851,lat,mta,via,LLIR,GO506_20_6431,89c25f1d4,12,STOPPED_AT,214,GO506_20_6431
,,7,2020-04-08T16:18:18Z,40.748495,lat,mta,via,LLIR,GO506_20_6431,89c25f29c,13,IN_TRANSIT_TO,237,GO506_20_6431
,,8,2020-04-08T16:11:03Z,40.749543,lat,mta,via,LLIR,GO506_20_6431,89c25fdb4,12,IN_TRANSIT_TO,214,GO506_20_6431
,,9,2020-04-08T16:10:46Z,40.750688,lat,mta,via,LLIR,GO506_20_6431,89c25fdbc,12,IN_TRANSIT_TO,214,GO506_20_6431
,,10,2020-04-08T16:09:36Z,40.757106,lat,mta,via,LLIR,GO506_20_6431,89c2600c4,11,STOPPED_AT,56,GO506_20_6431
,,11,2020-04-08T16:09:53Z,40.7565,lat,mta,via,LLIR,GO506_20_6431,89c2600c4,12,IN_TRANSIT_TO,214,GO506_20_6431
,,12,2020-04-08T16:09:01Z,40.757895,lat,mta,via,LLIR,GO506_20_6431,89c2600dc,11,STOPPED_AT,56,GO506_20_6431
,,13,2020-04-08T16:08:19Z,40.759873,lat,mta,via,LLIR,GO506_20_6431,89c26010c,11,IN_TRANSIT_TO,56,GO506_20_6431
,,14,2020-04-08T16:08:38Z,40.759128,lat,mta,via,LLIR,GO506_20_6431,89c26011c,11,STOPPED_AT,56,GO506_20_6431
,,15,2020-04-08T16:07:43Z,40.762143,lat,mta,via,LLIR,GO506_20_6431,89c26022c,11,IN_TRANSIT_TO,56,GO506_20_6431
,,16,2020-04-08T16:06:49Z,40.762709,lat,mta,via,LLIR,GO506_20_6431,89c260234,10,STOPPED_AT,130,GO506_20_6431
,,17,2020-04-08T16:06:31Z,40.762441,lat,mta,via,LLIR,GO506_20_6431,89c26024c,10,IN_TRANSIT_TO,130,GO506_20_6431
,,18,2020-04-08T16:06:11Z,40.762373,lat,mta,via,LLIR,GO506_20_6431,89c260254,10,IN_TRANSIT_TO,130,GO506_20_6431
,,19,2020-04-08T16:03:54Z,40.76117,lat,mta,via,LLIR,GO506_20_6431,89c2602cc,9,IN_TRANSIT_TO,11,GO506_20_6431
,,20,2020-04-08T16:04:28Z,40.761653,lat,mta,via,LLIR,GO506_20_6431,89c2602ec,9,STOPPED_AT,11,GO506_20_6431
,,21,2020-04-08T16:05:46Z,40.761908,lat,mta,via,LLIR,GO506_20_6431,89c2602f4,9,STOPPED_AT,11,GO506_20_6431
,,22,2020-04-08T16:10:29Z,40.753279,lat,mta,via,LLIR,GO506_20_6431,89c260754,12,IN_TRANSIT_TO,214,GO506_20_6431
,,23,2020-04-08T16:03:36Z,40.761241,lat,mta,via,LLIR,GO506_20_6431,89c261d34,8,STOPPED_AT,2,GO506_20_6431
,,24,2020-04-08T16:02:33Z,40.761443,lat,mta,via,LLIR,GO506_20_6431,89c261d3c,8,STOPPED_AT,2,GO506_20_6431
,,25,2020-04-08T16:01:55Z,40.761878,lat,mta,via,LLIR,GO506_20_6431,89c261d74,8,IN_TRANSIT_TO,2,GO506_20_6431
,,26,2020-04-08T15:59:24Z,40.763356,lat,mta,via,LLIR,GO506_20_6431,89c261e04,7,STOPPED_AT,25,GO506_20_6431
,,27,2020-04-08T15:45:33Z,40.815413,lat,mta,via,LLIR,GO506_20_6431,89c288dcc,4,STOPPED_AT,72,GO506_20_6431
,,28,2020-04-08T15:45:51Z,40.814096,lat,mta,via,LLIR,GO506_20_6431,89c288ddc,4,STOPPED_AT,72,GO506_20_6431
,,29,2020-04-08T15:46:12Z,40.810699,lat,mta,via,LLIR,GO506_20_6431,89c288e0c,4,STOPPED_AT,72,GO506_20_6431
,,30,2020-04-08T15:47:45Z,40.808517,lat,mta,via,LLIR,GO506_20_6431,89c288e14,4,STOPPED_AT,72,GO506_20_6431
,,31,2020-04-08T15:48:04Z,40.806882,lat,mta,via,LLIR,GO506_20_6431,89c288e3c,4,STOPPED_AT,72,GO506_20_6431
,,32,2020-04-08T15:49:01Z,40.800397,lat,mta,via,LLIR,GO506_20_6431,89c288fcc,4,STOPPED_AT,72,GO506_20_6431
,,33,2020-04-08T15:49:22Z,40.796724,lat,mta,via,LLIR,GO506_20_6431,89c288fe4,4,STOPPED_AT,72,GO506_20_6431
,,34,2020-04-08T15:50:44Z,40.794998,lat,mta,via,LLIR,GO506_20_6431,89c289004,4,STOPPED_AT,72,GO506_20_6431
,,35,2020-04-08T15:53:56Z,40.786038,lat,mta,via,LLIR,GO506_20_6431,89c289974,5,IN_TRANSIT_TO,120,GO506_20_6431
,,36,2020-04-08T15:52:38Z,40.787216,lat,mta,via,LLIR,GO506_20_6431,89c28997c,4,STOPPED_AT,72,GO506_20_6431
,,37,2020-04-08T15:52:16Z,40.788413,lat,mta,via,LLIR,GO506_20_6431,89c289a2c,4,STOPPED_AT,72,GO506_20_6431
,,38,2020-04-08T15:51:55Z,40.789185,lat,mta,via,LLIR,GO506_20_6431,89c289a3c,4,STOPPED_AT,72,GO506_20_6431
,,39,2020-04-08T15:51:20Z,40.791541,lat,mta,via,LLIR,GO506_20_6431,89c289a64,4,STOPPED_AT,72,GO506_20_6431
,,40,2020-04-08T15:51:37Z,40.791009,lat,mta,via,LLIR,GO506_20_6431,89c289a6c,4,STOPPED_AT,72,GO506_20_6431
,,41,2020-04-08T15:51:03Z,40.793905,lat,mta,via,LLIR,GO506_20_6431,89c289aa4,4,STOPPED_AT,72,GO506_20_6431
,,42,2020-04-08T15:54:57Z,40.775044,lat,mta,via,LLIR,GO506_20_6431,89c289f3c,5,STOPPED_AT,120,GO506_20_6431
,,43,2020-04-08T15:55:54Z,40.773334,lat,mta,via,LLIR,GO506_20_6431,89c289f6c,5,STOPPED_AT,120,GO506_20_6431
,,44,2020-04-08T15:56:13Z,40.772144,lat,mta,via,LLIR,GO506_20_6431,89c289f74,6,IN_TRANSIT_TO,42,GO506_20_6431
,,45,2020-04-08T15:56:33Z,40.770953,lat,mta,via,LLIR,GO506_20_6431,89c289f7c,6,IN_TRANSIT_TO,42,GO506_20_6431
,,46,2020-04-08T15:58:10Z,40.76594,lat,mta,via,LLIR,GO506_20_6431,89c28a014,7,IN_TRANSIT_TO,25,GO506_20_6431
,,47,2020-04-08T15:56:53Z,40.768069,lat,mta,via,LLIR,GO506_20_6431,89c28a024,6,STOPPED_AT,42,GO506_20_6431
,,48,2020-04-08T15:57:52Z,40.76661,lat,mta,via,LLIR,GO506_20_6431,89c28a03c,6,STOPPED_AT,42,GO506_20_6431
,,49,2020-04-08T15:44:58Z,-73.68691,lon,mta,start,LLIR,GO506_20_6431,89c288c54,1,STOPPED_AT,171,GO506_20_6431
,,50,2020-04-08T16:19:27Z,-73.940563,lon,mta,via,LLIR,GO506_20_6431,89c2592bc,13,IN_TRANSIT_TO,237,GO506_20_6431
,,51,2020-04-08T16:16:50Z,-73.912119,lon,mta,via,LLIR,GO506_20_6431,89c25f18c,13,IN_TRANSIT_TO,237,GO506_20_6431
,,52,2020-04-08T16:15:59Z,-73.905367,lon,mta,via,LLIR,GO506_20_6431,89c25f1c4,12,STOPPED_AT,214,GO506_20_6431
,,53,2020-04-08T16:16:16Z,-73.906398,lon,mta,via,LLIR,GO506_20_6431,89c25f1c4,13,IN_TRANSIT_TO,237,GO506_20_6431
,,54,2020-04-08T16:14:32Z,-73.900815,lon,mta,via,LLIR,GO506_20_6431,89c25f1d4,12,IN_TRANSIT_TO,214,GO506_20_6431
,,55,2020-04-08T16:14:49Z,-73.902975,lon,mta,via,LLIR,GO506_20_6431,89c25f1d4,12,STOPPED_AT,214,GO506_20_6431
,,56,2020-04-08T16:18:18Z,-73.927597,lon,mta,via,LLIR,GO506_20_6431,89c25f29c,13,IN_TRANSIT_TO,237,GO506_20_6431
,,57,2020-04-08T16:11:03Z,-73.852242,lon,mta,via,LLIR,GO506_20_6431,89c25fdb4,12,IN_TRANSIT_TO,214,GO506_20_6431
,,58,2020-04-08T16:10:46Z,-73.848962,lon,mta,via,LLIR,GO506_20_6431,89c25fdbc,12,IN_TRANSIT_TO,214,GO506_20_6431
,,59,2020-04-08T16:09:36Z,-73.833095,lon,mta,via,LLIR,GO506_20_6431,89c2600c4,11,STOPPED_AT,56,GO506_20_6431
,,60,2020-04-08T16:09:53Z,-73.834298,lon,mta,via,LLIR,GO506_20_6431,89c2600c4,12,IN_TRANSIT_TO,214,GO506_20_6431
,,61,2020-04-08T16:09:01Z,-73.831347,lon,mta,via,LLIR,GO506_20_6431,89c2600dc,11,STOPPED_AT,56,GO506_20_6431
,,62,2020-04-08T16:08:19Z,-73.826087,lon,mta,via,LLIR,GO506_20_6431,89c26010c,11,IN_TRANSIT_TO,56,GO506_20_6431
,,63,2020-04-08T16:08:38Z,-73.828799,lon,mta,via,LLIR,GO506_20_6431,89c26011c,11,STOPPED_AT,56,GO506_20_6431
,,64,2020-04-08T16:07:43Z,-73.817956,lon,mta,via,LLIR,GO506_20_6431,89c26022c,11,IN_TRANSIT_TO,56,GO506_20_6431
,,65,2020-04-08T16:06:49Z,-73.814539,lon,mta,via,LLIR,GO506_20_6431,89c260234,10,STOPPED_AT,130,GO506_20_6431
,,66,2020-04-08T16:06:31Z,-73.810154,lon,mta,via,LLIR,GO506_20_6431,89c26024c,10,IN_TRANSIT_TO,130,GO506_20_6431
,,67,2020-04-08T16:06:11Z,-73.809437,lon,mta,via,LLIR,GO506_20_6431,89c260254,10,IN_TRANSIT_TO,130,GO506_20_6431
,,68,2020-04-08T16:03:54Z,-73.795812,lon,mta,via,LLIR,GO506_20_6431,89c2602cc,9,IN_TRANSIT_TO,11,GO506_20_6431
,,69,2020-04-08T16:04:28Z,-73.801766,lon,mta,via,LLIR,GO506_20_6431,89c2602ec,9,STOPPED_AT,11,GO506_20_6431
,,70,2020-04-08T16:05:46Z,-73.804421,lon,mta,via,LLIR,GO506_20_6431,89c2602f4,9,STOPPED_AT,11,GO506_20_6431
,,71,2020-04-08T16:10:29Z,-73.841791,lon,mta,via,LLIR,GO506_20_6431,89c260754,12,IN_TRANSIT_TO,214,GO506_20_6431
,,72,2020-04-08T16:03:36Z,-73.792932,lon,mta,via,LLIR,GO506_20_6431,89c261d34,8,STOPPED_AT,2,GO506_20_6431
,,73,2020-04-08T16:02:33Z,-73.789959,lon,mta,via,LLIR,GO506_20_6431,89c261d3c,8,STOPPED_AT,2,GO506_20_6431
,,74,2020-04-08T16:01:55Z,-73.785035,lon,mta,via,LLIR,GO506_20_6431,89c261d74,8,IN_TRANSIT_TO,2,GO506_20_6431
,,75,2020-04-08T15:59:24Z,-73.769272,lon,mta,via,LLIR,GO506_20_6431,89c261e04,7,STOPPED_AT,25,GO506_20_6431
,,76,2020-04-08T15:45:33Z,-73.690054,lon,mta,via,LLIR,GO506_20_6431,89c288dcc,4,STOPPED_AT,72,GO506_20_6431
,,77,2020-04-08T15:45:51Z,-73.693214,lon,mta,via,LLIR,GO506_20_6431,89c288ddc,4,STOPPED_AT,72,GO506_20_6431
,,78,2020-04-08T15:46:12Z,-73.695214,lon,mta,via,LLIR,GO506_20_6431,89c288e0c,4,STOPPED_AT,72,GO506_20_6431
,,79,2020-04-08T15:47:45Z,-73.695594,lon,mta,via,LLIR,GO506_20_6431,89c288e14,4,STOPPED_AT,72,GO506_20_6431
,,80,2020-04-08T15:48:04Z,-73.695858,lon,mta,via,LLIR,GO506_20_6431,89c288e3c,4,STOPPED_AT,72,GO506_20_6431
,,81,2020-04-08T15:49:01Z,-73.695023,lon,mta,via,LLIR,GO506_20_6431,89c288fcc,4,STOPPED_AT,72,GO506_20_6431
,,82,2020-04-08T15:49:22Z,-73.699899,lon,mta,via,LLIR,GO506_20_6431,89c288fe4,4,STOPPED_AT,72,GO506_20_6431
,,83,2020-04-08T15:50:44Z,-73.702982,lon,mta,via,LLIR,GO506_20_6431,89c289004,4,STOPPED_AT,72,GO506_20_6431
,,84,2020-04-08T15:53:56Z,-73.72924,lon,mta,via,LLIR,GO506_20_6431,89c289974,5,IN_TRANSIT_TO,120,GO506_20_6431
,,85,2020-04-08T15:52:38Z,-73.7261,lon,mta,via,LLIR,GO506_20_6431,89c28997c,4,STOPPED_AT,72,GO506_20_6431
,,86,2020-04-08T15:52:16Z,-73.721973,lon,mta,via,LLIR,GO506_20_6431,89c289a2c,4,STOPPED_AT,72,GO506_20_6431
,,87,2020-04-08T15:51:55Z,-73.720059,lon,mta,via,LLIR,GO506_20_6431,89c289a3c,4,STOPPED_AT,72,GO506_20_6431
,,88,2020-04-08T15:51:20Z,-73.713571,lon,mta,via,LLIR,GO506_20_6431,89c289a64,4,STOPPED_AT,72,GO506_20_6431
,,89,2020-04-08T15:51:37Z,-73.715622,lon,mta,via,LLIR,GO506_20_6431,89c289a6c,4,STOPPED_AT,72,GO506_20_6431
,,90,2020-04-08T15:51:03Z,-73.705485,lon,mta,via,LLIR,GO506_20_6431,89c289aa4,4,STOPPED_AT,72,GO506_20_6431
,,91,2020-04-08T15:54:57Z,-73.740647,lon,mta,via,LLIR,GO506_20_6431,89c289f3c,5,STOPPED_AT,120,GO506_20_6431
,,92,2020-04-08T15:55:54Z,-73.742812,lon,mta,via,LLIR,GO506_20_6431,89c289f6c,5,STOPPED_AT,120,GO506_20_6431
,,93,2020-04-08T15:56:13Z,-73.74431,lon,mta,via,LLIR,GO506_20_6431,89c289f74,6,IN_TRANSIT_TO,42,GO506_20_6431
,,94,2020-04-08T15:56:33Z,-73.745805,lon,mta,via,LLIR,GO506_20_6431,89c289f7c,6,IN_TRANSIT_TO,42,GO506_20_6431
,,95,2020-04-08T15:58:10Z,-73.752465,lon,mta,via,LLIR,GO506_20_6431,89c28a014,7,IN_TRANSIT_TO,25,GO506_20_6431
,,96,2020-04-08T15:56:53Z,-73.749413,lon,mta,via,LLIR,GO506_20_6431,89c28a024,6,STOPPED_AT,42,GO506_20_6431
,,97,2020-04-08T15:57:52Z,-73.751322,lon,mta,via,LLIR,GO506_20_6431,89c28a03c,6,STOPPED_AT,42,GO506_20_6431

#group,false,false,false,false,true,true,true,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,,
,result,table,_time,_value,_field,_measurement,_pt,area,id,s2_cell_id,seq_idx,status,stop_id,trip_id
,,98,2020-04-08T15:44:58Z,1586304000,tid,mta,start,LLIR,GO506_20_6431,89c288c54,1,STOPPED_AT,171,GO506_20_6431
,,99,2020-04-08T16:19:27Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c2592bc,13,IN_TRANSIT_TO,237,GO506_20_6431
,,100,2020-04-08T16:16:50Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c25f18c,13,IN_TRANSIT_TO,237,GO506_20_6431
,,101,2020-04-08T16:15:59Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c25f1c4,12,STOPPED_AT,214,GO506_20_6431
,,102,2020-04-08T16:16:16Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c25f1c4,13,IN_TRANSIT_TO,237,GO506_20_6431
,,103,2020-04-08T16:14:32Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c25f1d4,12,IN_TRANSIT_TO,214,GO506_20_6431
,,104,2020-04-08T16:14:49Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c25f1d4,12,STOPPED_AT,214,GO506_20_6431
,,105,2020-04-08T16:18:18Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c25f29c,13,IN_TRANSIT_TO,237,GO506_20_6431
,,106,2020-04-08T16:11:03Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c25fdb4,12,IN_TRANSIT_TO,214,GO506_20_6431
,,107,2020-04-08T16:10:46Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c25fdbc,12,IN_TRANSIT_TO,214,GO506_20_6431
,,108,2020-04-08T16:09:36Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c2600c4,11,STOPPED_AT,56,GO506_20_6431
,,109,2020-04-08T16:09:53Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c2600c4,12,IN_TRANSIT_TO,214,GO506_20_6431
,,110,2020-04-08T16:09:01Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c2600dc,11,STOPPED_AT,56,GO506_20_6431
,,111,2020-04-08T16:08:19Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c26010c,11,IN_TRANSIT_TO,56,GO506_20_6431
,,112,2020-04-08T16:08:38Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c26011c,11,STOPPED_AT,56,GO506_20_6431
,,113,2020-04-08T16:07:43Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c26022c,11,IN_TRANSIT_TO,56,GO506_20_6431
,,114,2020-04-08T16:06:49Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c260234,10,STOPPED_AT,130,GO506_20_6431
,,115,2020-04-08T16:06:31Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c26024c,10,IN_TRANSIT_TO,130,GO506_20_6431
,,116,2020-04-08T16:06:11Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c260254,10,IN_TRANSIT_TO,130,GO506_20_6431
,,117,2020-04-08T16:03:54Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c2602cc,9,IN_TRANSIT_TO,11,GO506_20_6431
,,118,2020-04-08T16:04:28Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c2602ec,9,STOPPED_AT,11,GO506_20_6431
,,119,2020-04-08T16:05:46Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c2602f4,9,STOPPED_AT,11,GO506_20_6431
,,120,2020-04-08T16:10:29Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c260754,12,IN_TRANSIT_TO,214,GO506_20_6431
,,121,2020-04-08T16:03:36Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c261d34,8,STOPPED_AT,2,GO506_20_6431
,,122,2020-04-08T16:02:33Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c261d3c,8,STOPPED_AT,2,GO506_20_6431
,,123,2020-04-08T16:01:55Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c261d74,8,IN_TRANSIT_TO,2,GO506_20_6431
,,124,2020-04-08T15:59:24Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c261e04,7,STOPPED_AT,25,GO506_20_6431
,,125,2020-04-08T15:45:33Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c288dcc,4,STOPPED_AT,72,GO506_20_6431
,,126,2020-04-08T15:45:51Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c288ddc,4,STOPPED_AT,72,GO506_20_6431
,,127,2020-04-08T15:46:12Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c288e0c,4,STOPPED_AT,72,GO506_20_6431
,,128,2020-04-08T15:47:45Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c288e14,4,STOPPED_AT,72,GO506_20_6431
,,129,2020-04-08T15:48:04Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c288e3c,4,STOPPED_AT,72,GO506_20_6431
,,130,2020-04-08T15:49:01Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c288fcc,4,STOPPED_AT,72,GO506_20_6431
,,131,2020-04-08T15:49:22Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c288fe4,4,STOPPED_AT,72,GO506_20_6431
,,132,2020-04-08T15:50:44Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c289004,4,STOPPED_AT,72,GO506_20_6431
,,133,2020-04-08T15:53:56Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c289974,5,IN_TRANSIT_TO,120,GO506_20_6431
,,134,2020-04-08T15:52:38Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c28997c,4,STOPPED_AT,72,GO506_20_6431
,,135,2020-04-08T15:52:16Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c289a2c,4,STOPPED_AT,72,GO506_20_6431
,,136,2020-04-08T15:51:55Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c289a3c,4,STOPPED_AT,72,GO506_20_6431
,,137,2020-04-08T15:51:20Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c289a64,4,STOPPED_AT,72,GO506_20_6431
,,138,2020-04-08T15:51:37Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c289a6c,4,STOPPED_AT,72,GO506_20_6431
,,139,2020-04-08T15:51:03Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c289aa4,4,STOPPED_AT,72,GO506_20_6431
,,140,2020-04-08T15:54:57Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c289f3c,5,STOPPED_AT,120,GO506_20_6431
,,141,2020-04-08T15:55:54Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c289f6c,5,STOPPED_AT,120,GO506_20_6431
,,142,2020-04-08T15:56:13Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c289f74,6,IN_TRANSIT_TO,42,GO506_20_6431
,,143,2020-04-08T15:56:33Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c289f7c,6,IN_TRANSIT_TO,42,GO506_20_6431
,,144,2020-04-08T15:58:10Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c28a014,7,IN_TRANSIT_TO,25,GO506_20_6431
,,145,2020-04-08T15:56:53Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c28a024,6,STOPPED_AT,42,GO506_20_6431
,,146,2020-04-08T15:57:52Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c28a03c,6,STOPPED_AT,42,GO506_20_6431
"
outData = "
#group,false,false,false,true,false,true
#datatype,string,long,boolean,string,string,string
#default,_result,,,,,
,result,table,_st_intersects,id,st_linestring,trip_id
,,0,true,GO506_20_6431,\"-73.68691 40.820317, -73.690054 40.815413, -73.693214 40.814096, -73.695214 40.810699, -73.695594 40.808517, -73.695858 40.806882, -73.695023 40.800397, -73.699899 40.796724, -73.702982 40.794998, -73.705485 40.793905, -73.713571 40.791541, -73.715622 40.791009, -73.720059 40.789185, -73.721973 40.788413, -73.7261 40.787216, -73.72924 40.786038, -73.740647 40.775044, -73.742812 40.773334, -73.74431 40.772144, -73.745805 40.770953, -73.749413 40.768069, -73.751322 40.76661, -73.752465 40.76594, -73.769272 40.763356, -73.785035 40.761878, -73.789959 40.761443, -73.792932 40.761241, -73.795812 40.76117, -73.801766 40.761653, -73.804421 40.761908, -73.809437 40.762373, -73.810154 40.762441, -73.814539 40.762709, -73.817956 40.762143, -73.826087 40.759873, -73.828799 40.759128, -73.831347 40.757895, -73.833095 40.757106, -73.834298 40.7565, -73.841791 40.753279, -73.848962 40.750688, -73.852242 40.749543, -73.900815 40.743989, -73.902975 40.745851, -73.905367 40.748141, -73.906398 40.748909, -73.912119 40.751085, -73.927597 40.748495, -73.940563 40.745249\",GO506_20_6431
"

// polygon in Brooklyn
bt = {points: [{lat: 40.671659, lon: -73.936631}, {lat: 40.706543, lon: -73.749177}, {lat: 40.791333, lon: -73.880327}]}
t_stIntersectsLinestring = (table=<-) => table
    |> range(start: 2020-04-01T00:00:00Z)
    |> v1.fieldsAsCols()
    // optional but it helps to see train crossing defined region
    |> geo.asTracks(groupBy: ["id", "trip_id"])
    |> geo.ST_LineString()
    |> map(fn: (r) => ({r with _st_intersects: geo.ST_Intersects(region: bt, geometry: {linestring: r.st_linestring})}))
    |> drop(columns: ["_start", "_stop"])

test _stIntersectsLinestring = () => ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_stIntersectsLinestring})
