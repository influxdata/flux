package usage_test

import "testing"

// This dataset has been generated with this query:
// from(bucket: "system_usage")
//    |> range(start: 2019-12-03T10:00:00.000Z, stop: 2019-12-03T13:00:00.000Z)
//    |> filter(fn: (r) =>
//        (r.org_id == "03d01b74c8e09000" or r.org_id == "03c19003200d7000" or r.org_id == "0395bd7401aa3000")
//        and r._measurement == "queryd_billing"
//    )

inData = "
#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,0,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.832579331Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,0,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.727379572Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,0,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.812383085Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,1,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:01.16148791Z,180,read_values,queryd_billing,queryd-v2-5797867574-b2bc5,03d01b74c8e09000
,,1,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.15254861Z,180,read_values,queryd_billing,queryd-v2-5797867574-b2bc5,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,2,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.90791717Z,13593,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,3,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:01.463385504Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-5pjgb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,4,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:23:00.473713274Z,7829,total_duration_us,queryd_billing,queryd-v2-5797867574-tlhkl,0395bd7401aa3000
,,4,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.647808347Z,7960,total_duration_us,queryd_billing,queryd-v2-5797867574-tlhkl,0395bd7401aa3000
,,4,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:43:00.561679905Z,9878,total_duration_us,queryd_billing,queryd-v2-5797867574-tlhkl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,5,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:18:00.151904896Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,5,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:56:00.450344765Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,5,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:17:00.634378145Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,5,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:19:00.610282493Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,5,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:29:00.532131872Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,6,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.924157523Z,11211,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,6,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.579798761Z,12719,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,6,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.763690024Z,8583,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,6,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:18:00.533780271Z,7944,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,6,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.596326558Z,40729,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,6,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:42:00.44613233Z,7588,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,7,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.924157523Z,5,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,7,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.579798761Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,7,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.763690024Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,7,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:18:00.533780271Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,7,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.596326558Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,7,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:42:00.44613233Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,8,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.912019031Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-lj72r,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,9,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.755759982Z,18209,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000
,,9,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.774581825Z,69858,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,10,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.530529897Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-7s4z2,03c19003200d7000
,,10,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.642695941Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-7s4z2,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,11,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:12:00.513928572Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000
,,11,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:52:00.1555046Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,12,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:33:00.48797071Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,12,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:04:00.484123404Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,12,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.729548006Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,13,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.516727014Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-bd7fh,03c19003200d7000
,,13,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.776060648Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-bd7fh,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,14,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.912019031Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-lj72r,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,15,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:41:00.536966287Z,9006,total_duration_us,queryd_billing,queryd-v2-5797867574-7s4z2,0395bd7401aa3000
,,15,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:08:00.66861141Z,20765,total_duration_us,queryd_billing,queryd-v2-5797867574-7s4z2,0395bd7401aa3000
,,15,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:39:00.429663199Z,9001,total_duration_us,queryd_billing,queryd-v2-5797867574-7s4z2,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,16,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:01.463385504Z,108672,total_duration_us,queryd_billing,queryd-v2-5797867574-5pjgb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,17,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:21:00.679948866Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,17,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:46:00.57791668Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,17,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:47:00.665339757Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,17,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.885284853Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,17,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:37:00.490611137Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,18,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:04.170703823Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,19,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.789866842Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,19,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.635709149Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,19,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:01.216435523Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,20,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:01.463385504Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-5pjgb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,21,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.515673793Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000
,,21,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.590064004Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,22,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.730269127Z,8485,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,22,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:04:00.506173446Z,8185,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,22,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:08:00.618513396Z,8193,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,22,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:13:00.133001433Z,7794,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,22,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.921225459Z,8026,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,22,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:48:00.633450477Z,8344,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,23,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.644202789Z,180,read_values,queryd_billing,queryd-v2-5797867574-qqx49,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,24,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:01.003107451Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-qqx49,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,25,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.875119313Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,25,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.965625042Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,25,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.551411546Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,25,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.818106383Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,25,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.706945865Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,26,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.758278197Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,26,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.837604512Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,26,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.614967009Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,27,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:16:00.501326426Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-rj2ns,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,28,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:09:00.543511505Z,0,read_values,queryd_billing,queryd-v2-5797867574-kkzsw,0395bd7401aa3000
,,28,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:27:00.565066666Z,0,read_values,queryd_billing,queryd-v2-5797867574-kkzsw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,29,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:01:00.492154788Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,29,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:36:00.45783491Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,29,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:38:00.44515579Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,29,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:44:00.587165743Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,29,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.65000697Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,29,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:34:00.445880939Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,30,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.542805667Z,9430,total_duration_us,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,30,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.794869556Z,7236,total_duration_us,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,30,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:02:00.468421939Z,7817,total_duration_us,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,30,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:18:00.466768224Z,13313,total_duration_us,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,30,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.738772673Z,54614,total_duration_us,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,30,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:01:00.650919032Z,7479,total_duration_us,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,30,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:13:00.594089157Z,7084,total_duration_us,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,31,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.716500242Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,31,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.675300682Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,31,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.722782443Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,31,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:46:00.61084851Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,31,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:53:00.659149488Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,32,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.783829632Z,1440,read_bytes,queryd_billing,queryd-v2-5797867574-bd7fh,03d01b74c8e09000
,,32,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.757281711Z,1440,read_bytes,queryd_billing,queryd-v2-5797867574-bd7fh,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,33,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.821668012Z,158438,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,34,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:09:00.695676986Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000
,,34,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:16:00.423571342Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,35,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.774815315Z,0,read_values,queryd_billing,queryd-v2-5797867574-5pjgb,0395bd7401aa3000
,,35,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.710588962Z,0,read_values,queryd_billing,queryd-v2-5797867574-5pjgb,0395bd7401aa3000
,,35,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.518786657Z,0,read_values,queryd_billing,queryd-v2-5797867574-5pjgb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,36,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.924257631Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,37,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:17:00.44658031Z,8636,total_duration_us,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,37,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:22:00.620511505Z,8465,total_duration_us,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,37,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:49:00.504522138Z,8125,total_duration_us,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,37,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:03:00.458527039Z,9253,total_duration_us,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,37,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:32:00.562507962Z,11758,total_duration_us,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,38,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.53535901Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,39,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:01.040258084Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000
,,39,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.52727879Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000
,,39,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.472275138Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000
,,39,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.581634661Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,40,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:06:00.455340709Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,40,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:09:00.434414481Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,40,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.653985084Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,40,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.767478932Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,41,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.701333339Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,41,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.82139065Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,41,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.811423271Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,42,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.848026232Z,1003,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,42,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.861699773Z,1260,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,42,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.989693911Z,1260,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,42,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:04.171157376Z,139403,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,43,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.790663574Z,32668,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,43,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.830213309Z,19445,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,43,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.636824955Z,31779,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,43,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.163289319Z,16726,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,44,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:06:00.513213184Z,7948,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,44,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:22:00.448283291Z,8593,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,44,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:28:00.484967147Z,8764,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,44,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:56:00.684591295Z,7644,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,45,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.720881037Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-k254f,03c19003200d7000
,,45,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.759054637Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-k254f,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,46,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:06:00.455340709Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,46,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:09:00.434414481Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,46,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.653985084Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,46,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.767478932Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,47,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.863657529Z,0,read_values,queryd_billing,queryd-v2-5797867574-hh9fz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,48,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.894829972Z,90952,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,48,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.928682633Z,7,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,48,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.898959022Z,16,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,48,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.619773147Z,16,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,48,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.783903603Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,48,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.853962964Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,48,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.785243966Z,20,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,49,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:06:00.513213184Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,49,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:22:00.448283291Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,49,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:28:00.484967147Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,49,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:56:00.684591295Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,50,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.508804564Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000
,,50,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.720702585Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,51,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:02:00.478853385Z,0,read_values,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,51,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:07:00.556311114Z,0,read_values,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,51,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:19:00.239151116Z,0,read_values,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,51,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:51:00.592699963Z,0,read_values,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,51,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:54:00.433290693Z,0,read_values,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,51,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:01:00.637048958Z,0,read_values,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,51,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:56:00.503553023Z,0,read_values,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,51,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.693835864Z,0,read_values,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,52,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:01.012279784Z,26,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,53,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.515673793Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000
,,53,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.590064004Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,54,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.508804564Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000
,,54,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.720702585Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,55,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:33:00.48797071Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,55,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:04:00.484123404Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,55,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.729548006Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,56,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:34:00.547848426Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,56,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:51:00.496150236Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,56,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:21:00.620889507Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,56,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:26:00.675475921Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,56,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:43:00.567823817Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,57,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.776577884Z,254308,total_duration_us,queryd_billing,queryd-v2-5797867574-tlhkl,03d01b74c8e09000
,,57,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:01.170233871Z,446894,total_duration_us,queryd_billing,queryd-v2-5797867574-tlhkl,03d01b74c8e09000
,,57,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:01.25392002Z,105953,total_duration_us,queryd_billing,queryd-v2-5797867574-tlhkl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,58,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:02.629173563Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,58,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:04.58472994Z,8,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,58,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:01.025444871Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,59,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.90791717Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,60,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.929248638Z,49,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000
,,60,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:04.160876428Z,273464,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,61,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.848026232Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,61,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.861699773Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,61,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.989693911Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,61,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:04.171157376Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,62,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.821352355Z,1258,response_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,63,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.666298715Z,1440,read_bytes,queryd_billing,queryd-v2-5797867574-plnml,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,64,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:48:00.675272795Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,65,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:12:00.462956123Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-j8hm4,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,66,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.481079117Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-t7slt,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,67,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:04.41909794Z,3647335,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000
,,67,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:01.610831987Z,248722,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,68,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:02.661499487Z,10,read_values,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,68,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:01.057331112Z,270,read_values,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,68,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:02.601634416Z,10,read_values,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,69,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:34:00.547848426Z,8437,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,69,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:51:00.496150236Z,7529,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,69,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:21:00.620889507Z,7924,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,69,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:26:00.675475921Z,8272,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,69,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:43:00.567823817Z,10857,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,70,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.821668012Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,71,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.917582859Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,72,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.686930041Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,72,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:57:00.433078105Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,72,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:17:00.415381738Z,16,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,73,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.758278197Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,73,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.837604512Z,15,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,73,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.614967009Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,74,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:14:00.606703032Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,74,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.658228927Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,74,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.675835865Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,74,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:02:00.441622158Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,74,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:33:00.490131246Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,74,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:47:00.493465617Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,74,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:58:00.69307463Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,75,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.721543002Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000
,,75,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.578924493Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000
,,75,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.685639229Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000
,,75,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.610618315Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,76,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:02.629173563Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,76,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:04.58472994Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,76,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:01.025444871Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,77,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:11:00.505877871Z,0,read_values,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,77,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:27:00.574693886Z,0,read_values,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,77,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:59:00.572427992Z,0,read_values,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,77,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:07:00.569599945Z,0,read_values,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,77,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:22:00.588925323Z,0,read_values,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,77,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.50045533Z,0,read_values,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,77,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:53:00.457936128Z,0,read_values,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,78,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:09:00.695676986Z,16111,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000
,,78,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:16:00.423571342Z,7628,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,79,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.912019031Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,80,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.867420941Z,0,read_values,queryd_billing,queryd-v2-5797867574-s6t85,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,81,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.67710836Z,17155,total_duration_us,queryd_billing,queryd-v2-5797867574-kvktv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,82,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:29:00.521460892Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,82,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:34:00.474156777Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,82,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.914502559Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,82,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:39:00.487057271Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,83,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.585156227Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000
,,83,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.663976545Z,18,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,84,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:21:00.679948866Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,84,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:46:00.57791668Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,84,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:47:00.665339757Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,84,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.885284853Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,84,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:37:00.490611137Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,85,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.675124012Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000
,,85,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.92728318Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,86,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.685173538Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,86,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.693548351Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,86,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.623719778Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,86,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.755486749Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,87,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.767777421Z,0,read_values,queryd_billing,queryd-v2-5797867574-rj2ns,03c19003200d7000
,,87,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.779496475Z,0,read_values,queryd_billing,queryd-v2-5797867574-rj2ns,03c19003200d7000
,,87,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.727572623Z,0,read_values,queryd_billing,queryd-v2-5797867574-rj2ns,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,88,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.663457967Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000
,,88,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.674816715Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000
,,88,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.729511016Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000
,,88,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.501504684Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,89,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:53:00.603537064Z,19006,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,89,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.709640978Z,21772,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,89,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:41:00.659356314Z,7669,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,89,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:47:00.524120738Z,7573,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,89,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:07:00.552515712Z,10773,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,89,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:38:00.141966771Z,10107,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,89,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.625087256Z,9015,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,90,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.821668012Z,158399,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,91,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.755759982Z,21,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000
,,91,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.774581825Z,7,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,92,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.661988211Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,93,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:34:00.547848426Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,93,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:51:00.496150236Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,93,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:21:00.620889507Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,93,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:26:00.675475921Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,93,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:43:00.567823817Z,19,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,94,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.912019031Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,95,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:02:00.478853385Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,95,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:07:00.556311114Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,95,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:19:00.239151116Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,95,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:51:00.592699963Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,95,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:54:00.433290693Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,95,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:01:00.637048958Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,95,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:56:00.503553023Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,95,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.693835864Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,96,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:11:00.505877871Z,10036,total_duration_us,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,96,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:27:00.574693886Z,8074,total_duration_us,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,96,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:59:00.572427992Z,8607,total_duration_us,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,96,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:07:00.569599945Z,8242,total_duration_us,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,96,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:22:00.588925323Z,8405,total_duration_us,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,96,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.50045533Z,7694,total_duration_us,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,96,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:53:00.457936128Z,8061,total_duration_us,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,97,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:24:00.614708868Z,9929,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,97,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:27:00.607072565Z,10089,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,97,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:37:00.586640451Z,8979,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,97,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.627600735Z,9959,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,97,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:46:00.574052194Z,16319,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,98,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.716500242Z,0,read_values,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,98,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.675300682Z,0,read_values,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,98,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.722782443Z,0,read_values,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,98,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:46:00.61084851Z,0,read_values,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,98,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:53:00.659149488Z,0,read_values,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,99,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.90791717Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,100,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:48:00.675272795Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,101,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.952632098Z,0,read_values,queryd_billing,queryd-v2-5797867574-sn26z,0395bd7401aa3000
,,101,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.896619226Z,0,read_values,queryd_billing,queryd-v2-5797867574-sn26z,0395bd7401aa3000
,,101,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:23:00.191362562Z,0,read_values,queryd_billing,queryd-v2-5797867574-sn26z,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,102,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:52:00.663145742Z,9232,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,102,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:31:00.655419185Z,10276,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,102,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:44:00.514580555Z,11925,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,103,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:24:00.614708868Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,103,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:27:00.607072565Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,103,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:37:00.586640451Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,103,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.627600735Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,103,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:46:00.574052194Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,104,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:06:00.513213184Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,104,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:22:00.448283291Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,104,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:28:00.484967147Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,104,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:56:00.684591295Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,105,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:02:00.478853385Z,8894,total_duration_us,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,105,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:07:00.556311114Z,8191,total_duration_us,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,105,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:19:00.239151116Z,8461,total_duration_us,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,105,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:51:00.592699963Z,9682,total_duration_us,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,105,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:54:00.433290693Z,7592,total_duration_us,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,105,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:01:00.637048958Z,8079,total_duration_us,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,105,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:56:00.503553023Z,15843,total_duration_us,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,105,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.693835864Z,10247,total_duration_us,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,106,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.917582859Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,107,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.929248638Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000
,,107,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:04.160876428Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,108,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:17:00.44658031Z,0,read_values,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,108,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:22:00.620511505Z,0,read_values,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,108,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:49:00.504522138Z,0,read_values,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,108,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:03:00.458527039Z,0,read_values,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,108,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:32:00.562507962Z,0,read_values,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,109,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.924257631Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,110,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.832579331Z,22013,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,110,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.727379572Z,21594,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,110,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.812383085Z,40427,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,111,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:06:00.455332774Z,9475,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,111,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:42:00.541039897Z,8498,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,111,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:29:00.477956027Z,13400,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,111,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:33:00.515658208Z,9947,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,111,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:11:00.667516371Z,8766,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,112,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.821352355Z,223976,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,113,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.789866842Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,113,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.635709149Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,113,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:01.216435523Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,114,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.790663574Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,114,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.830213309Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,114,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.636824955Z,16,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,114,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.163289319Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,115,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.790663574Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,115,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.830213309Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,115,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.636824955Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,115,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.163289319Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,116,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.777690142Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,116,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:03.873228519Z,8,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,116,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975537336Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,116,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.976567453Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,117,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:04.437709166Z,273464,read_bytes,queryd_billing,queryd-v1-5f699b6b58-66kcw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,118,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.730269127Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,118,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:04:00.506173446Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,118,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:08:00.618513396Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,118,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:13:00.133001433Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,118,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.921225459Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,118,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:48:00.633450477Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,119,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.282532978Z,176308,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000
,,119,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.97843698Z,205332,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,120,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.758278197Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,120,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.837604512Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,120,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.614967009Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,121,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.952632098Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-sn26z,0395bd7401aa3000
,,121,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.896619226Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-sn26z,0395bd7401aa3000
,,121,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:23:00.191362562Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-sn26z,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,122,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.824636831Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000
,,122,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.625636236Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,123,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.894829972Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,123,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.928682633Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,123,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.898959022Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,123,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.619773147Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,123,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.783903603Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,123,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.853962964Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,123,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.785243966Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,124,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:01.003107451Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-qqx49,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,125,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:14:00.606703032Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,125,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.658228927Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,125,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.675835865Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,125,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:02:00.441622158Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,125,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:33:00.490131246Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,125,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:47:00.493465617Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,125,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:58:00.69307463Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,126,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.824636831Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000
,,126,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.625636236Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,127,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.760513754Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,127,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:24:00.625390315Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,127,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:57:00.617251549Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,128,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.53535901Z,12389,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,129,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.762544227Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-xbx7c,03c19003200d7000
,,129,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.830581156Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-xbx7c,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,130,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.749034925Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zmcl2,03c19003200d7000
,,130,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.721679848Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zmcl2,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,131,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:01.077941715Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,132,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.824636831Z,225267,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000
,,132,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.625636236Z,150858,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,133,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.90791717Z,16,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,134,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:49:00.426081648Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,135,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.813155264Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,135,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.697447893Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,135,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.780742525Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,136,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.492088454Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-t474n,03c19003200d7000
,,136,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.714536617Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-t474n,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,137,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.481079117Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,138,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.720881037Z,0,read_values,queryd_billing,queryd-v2-5797867574-k254f,03c19003200d7000
,,138,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.759054637Z,0,read_values,queryd_billing,queryd-v2-5797867574-k254f,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,139,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:06:00.455332774Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,139,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:42:00.541039897Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,139,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:29:00.477956027Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,139,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:33:00.515658208Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,139,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:11:00.667516371Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,140,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.755759982Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000
,,140,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.774581825Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,141,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.912019031Z,12057,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,142,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.512542619Z,16715,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,142,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.50851438Z,16402,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,142,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.741858095Z,22754,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,142,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.486155671Z,14949,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,143,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.730269127Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,143,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:04:00.506173446Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,143,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:08:00.618513396Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,143,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:13:00.133001433Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,143,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.921225459Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,143,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:48:00.633450477Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,144,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:34:00.547848426Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,144,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:51:00.496150236Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,144,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:21:00.620889507Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,144,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:26:00.675475921Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,144,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:43:00.567823817Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,145,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:06:00.455332774Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,145,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:42:00.541039897Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,145,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:29:00.477956027Z,23,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,145,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:33:00.515658208Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,145,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:11:00.667516371Z,16,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,146,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.675124012Z,12049,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000
,,146,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.92728318Z,19482,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,147,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:09:00.695676986Z,16145,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000
,,147,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:16:00.423571342Z,7661,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,148,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975813826Z,180,read_values,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,149,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.685173538Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,149,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.693548351Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,149,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.623719778Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,149,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.755486749Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,150,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.90791717Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,151,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.760513754Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,151,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:24:00.625390315Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,151,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:57:00.617251549Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,152,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.852953324Z,0,read_values,queryd_billing,queryd-v2-5797867574-5pjgb,03c19003200d7000
,,152,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.655575144Z,0,read_values,queryd_billing,queryd-v2-5797867574-5pjgb,03c19003200d7000
,,152,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.656976818Z,0,read_values,queryd_billing,queryd-v2-5797867574-5pjgb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,153,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:12:00.513928572Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000
,,153,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:52:00.1555046Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,154,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:29:00.521460892Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,154,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:34:00.474156777Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,154,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.914502559Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,154,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:39:00.487057271Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,155,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:52:00.663145742Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,155,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:31:00.655419185Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,155,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:44:00.514580555Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,156,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.471264316Z,0,read_values,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,156,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.532354811Z,0,read_values,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,156,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.58982965Z,0,read_values,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,156,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.736000374Z,0,read_values,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,156,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.62852928Z,0,read_values,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,157,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.875119313Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,157,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.965625042Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,157,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.551411546Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,157,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.818106383Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,157,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.706945865Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,158,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:53:00.603537064Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,158,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.709640978Z,16,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,158,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:41:00.659356314Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,158,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:47:00.524120738Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,158,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:07:00.552515712Z,17,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,158,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:38:00.141966771Z,24,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,158,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.625087256Z,20,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,159,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.675124012Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000
,,159,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.92728318Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,160,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:52:00.663145742Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,160,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:31:00.655419185Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,160,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:44:00.514580555Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,161,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975813826Z,327037,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,162,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:23:00.473713274Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-tlhkl,0395bd7401aa3000
,,162,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.647808347Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-tlhkl,0395bd7401aa3000
,,162,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:43:00.561679905Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-tlhkl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,163,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.686930041Z,10746,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,163,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:57:00.433078105Z,9014,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,163,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:17:00.415381738Z,11129,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,164,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.758278197Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,164,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.837604512Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,164,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.614967009Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,165,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.774815315Z,12644,total_duration_us,queryd_billing,queryd-v2-5797867574-5pjgb,0395bd7401aa3000
,,165,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.710588962Z,68065,total_duration_us,queryd_billing,queryd-v2-5797867574-5pjgb,0395bd7401aa3000
,,165,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.518786657Z,11253,total_duration_us,queryd_billing,queryd-v2-5797867574-5pjgb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,166,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.777690142Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,166,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:03.873228519Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,166,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975537336Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,166,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.976567453Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,167,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.508804564Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000
,,167,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.720702585Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,168,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.585156227Z,13689,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000
,,168,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.663976545Z,20217,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,169,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.749802396Z,1025,response_bytes,queryd_billing,queryd-v2-5797867574-rj2ns,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,170,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.917582859Z,1018,response_bytes,queryd_billing,queryd-v1-5f699b6b58-467pb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,171,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:04.437709166Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,172,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:01.463385504Z,0,read_values,queryd_billing,queryd-v2-5797867574-5pjgb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,173,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.6467672Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-mfspl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,174,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:04.41909794Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000
,,174,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:01.610831987Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,175,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.929248638Z,187789,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000
,,175,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:04.160876428Z,3484886,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,176,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.686930041Z,10712,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,176,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:57:00.433078105Z,8982,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,176,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:17:00.415381738Z,11089,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,177,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.686930041Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,177,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:57:00.433078105Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,177,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:17:00.415381738Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,178,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:06:00.455340709Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,178,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:09:00.434414481Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,178,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.653985084Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,178,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.767478932Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,179,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.819675585Z,1022,response_bytes,queryd_billing,queryd-v2-5797867574-xvkns,03d01b74c8e09000
,,179,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.79114313Z,1013,response_bytes,queryd_billing,queryd-v2-5797867574-xvkns,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,180,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:34:00.547848426Z,8466,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,180,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:51:00.496150236Z,7558,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,180,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:21:00.620889507Z,7964,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,180,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:26:00.675475921Z,8301,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,180,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:43:00.567823817Z,10899,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,181,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:12:00.513928572Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000
,,181,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:52:00.1555046Z,15,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,182,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.824636831Z,1518,response_bytes,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000
,,182,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.625636236Z,1019,response_bytes,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,183,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:02.661499487Z,169,read_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,183,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:01.057331112Z,2160,read_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,183,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:02.601634416Z,169,read_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,184,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:26:00.437242604Z,7606,total_duration_us,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,184,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:31:00.477712461Z,8670,total_duration_us,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,184,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:21:00.4986011Z,9084,total_duration_us,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,184,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:52:00.719041693Z,9913,total_duration_us,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,184,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:54:00.613488751Z,11342,total_duration_us,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,184,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:59:00.564883689Z,10291,total_duration_us,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,184,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:03:00.545102773Z,7880,total_duration_us,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,184,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.856694666Z,37640,total_duration_us,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,185,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.785640834Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,185,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.620654875Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,185,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:12:00.462658028Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,186,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.832579331Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,186,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.727379572Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,186,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.812383085Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,187,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.282532978Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000
,,187,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.97843698Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,188,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.585156227Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000
,,188,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.663976545Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,189,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.777690142Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,189,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:03.873228519Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,189,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975537336Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,189,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.976567453Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,190,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:04.41909794Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000
,,190,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:01.610831987Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,191,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:04.41909794Z,34183,read_values,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000
,,191,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:01.610831987Z,10,read_values,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,192,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.824636831Z,270,read_values,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000
,,192,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.625636236Z,180,read_values,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,193,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.783156493Z,180,read_values,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000
,,193,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.825524632Z,180,read_values,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,194,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.609126611Z,16,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,194,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.770656494Z,8,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,194,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.469463621Z,15,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,194,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.71182255Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,195,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:01:00.492154788Z,8306,total_duration_us,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,195,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:36:00.45783491Z,9454,total_duration_us,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,195,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:38:00.44515579Z,9339,total_duration_us,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,195,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:44:00.587165743Z,9127,total_duration_us,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,195,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.65000697Z,51075,total_duration_us,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,195,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:34:00.445880939Z,9621,total_duration_us,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,196,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.720881037Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-k254f,03c19003200d7000
,,196,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.759054637Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-k254f,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,197,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.790663574Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,197,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.830213309Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,197,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.636824955Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,197,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.163289319Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,198,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:37:00.423986232Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,198,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.759512962Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,198,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:28:00.377646402Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,198,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:44:00.420950673Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,198,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:32:00.605366491Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,198,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:49:00.463047225Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,199,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.976921236Z,270,read_values,queryd_billing,queryd-v2-5797867574-hh9fz,03d01b74c8e09000
,,199,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.833400375Z,180,read_values,queryd_billing,queryd-v2-5797867574-hh9fz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,200,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.857012416Z,0,read_values,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,200,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:28:00.450725793Z,0,read_values,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,200,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:32:00.590667734Z,0,read_values,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,200,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:39:00.577723384Z,0,read_values,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,200,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:11:00.598135316Z,0,read_values,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,201,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.783156493Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000
,,201,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.825524632Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,202,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.789866842Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,202,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.635709149Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,202,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:01.216435523Z,26,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,203,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.821352355Z,1440,read_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,204,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:31:00.627947836Z,31477,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-mfspl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,205,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:48:00.675272795Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,206,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:33:00.48797071Z,18234,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,206,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:04:00.484123404Z,11624,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,206,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.729548006Z,43454,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,207,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:49:00.426081648Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-lj72r,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,208,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.515673793Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000
,,208,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.590064004Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,209,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.789866842Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,209,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.635709149Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,209,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:01.216435523Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,210,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:36:00.654229914Z,8737,total_duration_us,queryd_billing,queryd-v2-5797867574-5ff7l,0395bd7401aa3000
,,210,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:59:00.617599152Z,11639,total_duration_us,queryd_billing,queryd-v2-5797867574-5ff7l,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,211,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:02:00.478853385Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,211,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:07:00.556311114Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,211,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:19:00.239151116Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,211,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:51:00.592699963Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,211,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:54:00.433290693Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,211,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:01:00.637048958Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,211,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:56:00.503553023Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000
,,211,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.693835864Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-c88sh,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,212,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:02.661499487Z,223749,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,212,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:01.057331112Z,200564,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,212,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:02.601634416Z,220607,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,213,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:49:00.426081648Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,214,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.832579331Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,214,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.727379572Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,214,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.812383085Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,215,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:14:00.606703032Z,8561,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,215,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.658228927Z,8868,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,215,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.675835865Z,20343,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,215,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:02:00.441622158Z,7786,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,215,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:33:00.490131246Z,11318,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,215,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:47:00.493465617Z,11698,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,215,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:58:00.69307463Z,20814,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,216,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.824636831Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000
,,216,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.625636236Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,217,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.755759982Z,8,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000
,,217,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.774581825Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,218,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:02.629173563Z,233833,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,218,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:04.58472994Z,3969073,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,218,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:01.025444871Z,223150,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,219,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:06:00.455340709Z,16,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,219,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:09:00.434414481Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,219,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.653985084Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,219,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.767478932Z,10,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,220,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.054579583Z,554585,total_duration_us,queryd_billing,queryd-v2-5797867574-xbx7c,03d01b74c8e09000
,,220,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:04.057304496Z,3377318,total_duration_us,queryd_billing,queryd-v2-5797867574-xbx7c,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,221,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.508804564Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000
,,221,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.720702585Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,222,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:21:00.679948866Z,8285,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,222,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:46:00.57791668Z,7455,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,222,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:47:00.665339757Z,7902,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,222,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.885284853Z,7397,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,222,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:37:00.490611137Z,7688,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,223,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:01.077941715Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,224,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.685173538Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,224,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.693548351Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,224,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.623719778Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,224,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.755486749Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,225,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.924157523Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,225,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.579798761Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,225,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.763690024Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,225,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:18:00.533780271Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,225,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.596326558Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,225,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:42:00.44613233Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,226,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.61520984Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-c88sh,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,227,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.512542619Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,227,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.50851438Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,227,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.741858095Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,227,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.486155671Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,228,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.530529897Z,19410,total_duration_us,queryd_billing,queryd-v2-5797867574-7s4z2,03c19003200d7000
,,228,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.642695941Z,30931,total_duration_us,queryd_billing,queryd-v2-5797867574-7s4z2,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,229,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.821668012Z,180,read_values,queryd_billing,queryd-v1-5f699b6b58-9lsws,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,230,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:12:00.513928572Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000
,,230,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:52:00.1555046Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,231,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.724536149Z,16383,total_duration_us,queryd_billing,queryd-v2-5797867574-xvkns,03c19003200d7000
,,231,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.585442572Z,14006,total_duration_us,queryd_billing,queryd-v2-5797867574-xvkns,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,232,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.282532978Z,1517,response_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000
,,232,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.97843698Z,1517,response_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,233,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:29:00.521460892Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,233,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:34:00.474156777Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,233,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.914502559Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,233,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:39:00.487057271Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,234,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:33:00.48797071Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,234,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:04:00.484123404Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,234,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.729548006Z,2,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,235,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.832140221Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,235,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:38:00.693295746Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,235,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:48:00.560832795Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,235,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.577141351Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,235,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.890962075Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,235,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.182960005Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,236,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:16:00.501326426Z,7968,total_duration_us,queryd_billing,queryd-v2-5797867574-rj2ns,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,237,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.777690142Z,15,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,237,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:03.873228519Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,237,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975537336Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,237,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.976567453Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,238,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:52:00.663145742Z,9201,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,238,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:31:00.655419185Z,10241,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,238,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:44:00.514580555Z,11890,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,239,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:14:00.606703032Z,15,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,239,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.658228927Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,239,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.675835865Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,239,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:02:00.441622158Z,26,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,239,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:33:00.490131246Z,17,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,239,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:47:00.493465617Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,239,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:58:00.69307463Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,240,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.813155264Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,240,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.697447893Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,240,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.780742525Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,241,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.925514969Z,146044,total_duration_us,queryd_billing,queryd-v2-5797867574-fmtgz,03d01b74c8e09000
,,241,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:01.84642159Z,227179,total_duration_us,queryd_billing,queryd-v2-5797867574-fmtgz,03d01b74c8e09000
,,241,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:01.561424858Z,105645,total_duration_us,queryd_billing,queryd-v2-5797867574-fmtgz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,242,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.783829632Z,1019,response_bytes,queryd_billing,queryd-v2-5797867574-bd7fh,03d01b74c8e09000
,,242,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.757281711Z,1260,response_bytes,queryd_billing,queryd-v2-5797867574-bd7fh,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,243,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.714848904Z,232444,total_duration_us,queryd_billing,queryd-v2-5797867574-c88sh,03d01b74c8e09000
,,243,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.67290226Z,223505,total_duration_us,queryd_billing,queryd-v2-5797867574-c88sh,03d01b74c8e09000
,,243,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:04.058154233Z,3556074,total_duration_us,queryd_billing,queryd-v2-5797867574-c88sh,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,244,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:48:00.675272795Z,7954,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,245,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:24:00.614708868Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,245,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:27:00.607072565Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,245,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:37:00.586640451Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,245,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.627600735Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,245,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:46:00.574052194Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,246,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.481079117Z,13423,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,247,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.52560325Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-plnml,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,248,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:01.040258084Z,17463,total_duration_us,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000
,,248,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.52727879Z,17030,total_duration_us,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000
,,248,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.472275138Z,13217,total_duration_us,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000
,,248,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.581634661Z,13690,total_duration_us,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,249,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:49:00.426081648Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-lj72r,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,250,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.767777421Z,12217,total_duration_us,queryd_billing,queryd-v2-5797867574-rj2ns,03c19003200d7000
,,250,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.779496475Z,62298,total_duration_us,queryd_billing,queryd-v2-5797867574-rj2ns,03c19003200d7000
,,250,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.727572623Z,43132,total_duration_us,queryd_billing,queryd-v2-5797867574-rj2ns,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,251,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.917582859Z,268825,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,252,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.790663574Z,32699,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,252,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.830213309Z,19476,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,252,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.636824955Z,31818,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,252,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.163289319Z,16759,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,253,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.701333339Z,36708,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,253,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.82139065Z,31284,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,253,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.811423271Z,12074,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,254,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.730269127Z,7,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,254,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:04:00.506173446Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,254,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:08:00.618513396Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,254,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:13:00.133001433Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,254,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.921225459Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,254,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:48:00.633450477Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,255,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.821668012Z,1440,read_bytes,queryd_billing,queryd-v1-5f699b6b58-9lsws,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,256,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.661988211Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,257,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.6467672Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-mfspl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,258,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.819675585Z,180,read_values,queryd_billing,queryd-v2-5797867574-xvkns,03d01b74c8e09000
,,258,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.79114313Z,180,read_values,queryd_billing,queryd-v2-5797867574-xvkns,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,259,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:04.423740313Z,273464,read_bytes,queryd_billing,queryd-v2-5797867574-kssph,03d01b74c8e09000
,,259,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.715319869Z,1440,read_bytes,queryd_billing,queryd-v2-5797867574-kssph,03d01b74c8e09000
,,259,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.898664906Z,169,read_bytes,queryd_billing,queryd-v2-5797867574-kssph,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,260,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.937640598Z,1260,response_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000
,,260,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.803909311Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,261,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.508804564Z,13774,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000
,,261,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.720702585Z,11740,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,262,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.755759982Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000
,,262,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.774581825Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,263,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:31:00.627947836Z,20358,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-mfspl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,264,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.924157523Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,264,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.579798761Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,264,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.763690024Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,264,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:18:00.533780271Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,264,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.596326558Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,264,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:42:00.44613233Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,265,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:08:00.566922506Z,8854,total_duration_us,queryd_billing,queryd-v2-5797867574-xbx7c,0395bd7401aa3000
,,265,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:19:00.411422463Z,9004,total_duration_us,queryd_billing,queryd-v2-5797867574-xbx7c,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,266,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.609126611Z,25467,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,266,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.770656494Z,17679,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,266,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.469463621Z,13179,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,266,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.71182255Z,46317,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,267,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:29:00.521460892Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,267,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:34:00.474156777Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,267,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.914502559Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,267,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:39:00.487057271Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,268,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:04.41909794Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000
,,268,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:01.610831987Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,269,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.53535901Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,270,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:31:00.627947836Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-mfspl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,271,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.832579331Z,18,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,271,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.727379572Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,271,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.812383085Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,272,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.609126611Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,272,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.770656494Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,272,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.469463621Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,272,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.71182255Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,273,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.675124012Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000
,,273,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.92728318Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,274,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.53535901Z,12418,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,275,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:29:00.521460892Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,275,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:34:00.474156777Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,275,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.914502559Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,275,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:39:00.487057271Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,276,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:01.077941715Z,270,read_values,queryd_billing,queryd-v1-5f699b6b58-t7slt,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,277,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.832140221Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,277,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:38:00.693295746Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,277,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:48:00.560832795Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,277,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.577141351Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,277,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.890962075Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,277,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.182960005Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,278,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.852953324Z,11990,total_duration_us,queryd_billing,queryd-v2-5797867574-5pjgb,03c19003200d7000
,,278,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.655575144Z,28837,total_duration_us,queryd_billing,queryd-v2-5797867574-5pjgb,03c19003200d7000
,,278,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.656976818Z,15104,total_duration_us,queryd_billing,queryd-v2-5797867574-5pjgb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,279,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:21:00.679948866Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,279,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:46:00.57791668Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,279,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:47:00.665339757Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,279,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.885284853Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,279,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:37:00.490611137Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,280,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:33:00.48797071Z,18278,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,280,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:04:00.484123404Z,11669,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,280,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.729548006Z,43494,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,281,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:14:00.606703032Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,281,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.658228927Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,281,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.675835865Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,281,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:02:00.441622158Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,281,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:33:00.490131246Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,281,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:47:00.493465617Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,281,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:58:00.69307463Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,282,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.867420941Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-s6t85,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,283,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:12:00.462956123Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,284,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.53535901Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,285,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.813155264Z,22,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,285,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.697447893Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,285,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.780742525Z,15,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,286,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:01.040258084Z,0,read_values,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000
,,286,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.52727879Z,0,read_values,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000
,,286,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.472275138Z,0,read_values,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000
,,286,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.581634661Z,0,read_values,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,287,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.70702057Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000
,,287,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:04:00.522827935Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,288,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:12:00.462956123Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,289,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.492227407Z,0,read_values,queryd_billing,queryd-v2-5797867574-zlf4s,03c19003200d7000
,,289,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.632944642Z,0,read_values,queryd_billing,queryd-v2-5797867574-zlf4s,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,290,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:23:00.482973775Z,8896,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,290,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:42:00.476029727Z,7981,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,290,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:24:00.494923251Z,12286,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,290,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.800071949Z,8280,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,291,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975813826Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,292,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:01.012279784Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,293,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.515673793Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000
,,293,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.590064004Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,294,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.67710836Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kvktv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,295,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.6467672Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-mfspl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,296,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.580021946Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000
,,296,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.59833827Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,297,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:14:00.572403792Z,7754,total_duration_us,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,297,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.687631407Z,53839,total_duration_us,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,297,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:14:00.48545546Z,10068,total_duration_us,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,297,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:41:00.534038417Z,7845,total_duration_us,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,297,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:51:00.491763198Z,18284,total_duration_us,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,298,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:23:00.482973775Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,298,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:42:00.476029727Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,298,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:24:00.494923251Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,298,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.800071949Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,299,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.609126611Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,299,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.770656494Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,299,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.469463621Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,299,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.71182255Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,300,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.755759982Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000
,,300,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.774581825Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,301,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.917582859Z,17,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,302,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:04.173174511Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,302,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:04.02017492Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,302,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.990040524Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,302,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:01.585834691Z,6,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,302,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.67997434Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,303,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:08:00.566922506Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-xbx7c,0395bd7401aa3000
,,303,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:19:00.411422463Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-xbx7c,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,304,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.758278197Z,25806,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,304,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.837604512Z,16171,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,304,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.614967009Z,11809,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,305,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.730269127Z,8525,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,305,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:04:00.506173446Z,8219,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,305,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:08:00.618513396Z,8224,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,305,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:13:00.133001433Z,7826,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,305,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.921225459Z,8062,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,305,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:48:00.633450477Z,8379,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,306,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:02.629173563Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,306,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:04.58472994Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,306,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:01.025444871Z,2,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,307,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.762544227Z,0,read_values,queryd_billing,queryd-v2-5797867574-xbx7c,03c19003200d7000
,,307,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.830581156Z,0,read_values,queryd_billing,queryd-v2-5797867574-xbx7c,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,308,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.912019031Z,12019,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,309,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.681462039Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,309,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.955855408Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,309,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.748825278Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,310,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.515673793Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000
,,310,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.590064004Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,311,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.819675585Z,153561,total_duration_us,queryd_billing,queryd-v2-5797867574-xvkns,03d01b74c8e09000
,,311,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.79114313Z,236098,total_duration_us,queryd_billing,queryd-v2-5797867574-xvkns,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,312,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:08:00.566922506Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-xbx7c,0395bd7401aa3000
,,312,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:19:00.411422463Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-xbx7c,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,313,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.917582859Z,1440,read_bytes,queryd_billing,queryd-v1-5f699b6b58-467pb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,314,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:12:00.462956123Z,7773,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,315,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:09:00.695676986Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000
,,315,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:16:00.423571342Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,316,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.609126611Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,316,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.770656494Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,316,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.469463621Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,316,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.71182255Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,317,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:01.256241012Z,2160,read_bytes,queryd_billing,queryd-v2-5797867574-cm6fz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,318,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.6467672Z,14610,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-mfspl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,319,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:02.661499487Z,223716,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,319,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:01.057331112Z,200530,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,319,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:02.601634416Z,220572,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,320,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:53:00.603537064Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,320,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.709640978Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,320,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:41:00.659356314Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,320,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:47:00.524120738Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,320,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:07:00.552515712Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,320,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:38:00.141966771Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,320,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.625087256Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,321,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:03:00.453652438Z,0,read_values,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000
,,321,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:58:00.467439389Z,0,read_values,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000
,,321,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:26:00.505504846Z,0,read_values,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000
,,321,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.863801527Z,0,read_values,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,322,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.836044529Z,11618,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000
,,322,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.425198714Z,12580,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,323,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.813155264Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,323,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.697447893Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,323,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.780742525Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,324,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.720881037Z,102242,total_duration_us,queryd_billing,queryd-v2-5797867574-k254f,03c19003200d7000
,,324,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.759054637Z,11429,total_duration_us,queryd_billing,queryd-v2-5797867574-k254f,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,325,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.530529897Z,0,read_values,queryd_billing,queryd-v2-5797867574-7s4z2,03c19003200d7000
,,325,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.642695941Z,0,read_values,queryd_billing,queryd-v2-5797867574-7s4z2,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,326,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:18:00.151904896Z,0,read_values,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,326,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:56:00.450344765Z,0,read_values,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,326,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:17:00.634378145Z,0,read_values,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,326,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:19:00.610282493Z,0,read_values,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,326,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:29:00.532131872Z,0,read_values,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,327,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.912019031Z,6,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,328,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.824636831Z,2160,read_bytes,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000
,,328,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.625636236Z,1440,read_bytes,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,329,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.53535901Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,330,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:57:00.444508534Z,0,read_values,queryd_billing,queryd-v2-5797867574-t474n,0395bd7401aa3000
,,330,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:54:00.757802699Z,0,read_values,queryd_billing,queryd-v2-5797867574-t474n,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,331,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.894829972Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,331,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.928682633Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,331,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.898959022Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,331,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.619773147Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,331,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.783903603Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,331,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.853962964Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,331,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.785243966Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,332,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.90791717Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,333,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:06:00.455332774Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,333,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:42:00.541039897Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,333,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:29:00.477956027Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,333,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:33:00.515658208Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,333,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:11:00.667516371Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,334,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.70702057Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000
,,334,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:04:00.522827935Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,335,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.666298715Z,514275,total_duration_us,queryd_billing,queryd-v2-5797867574-plnml,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,336,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:09:00.543511505Z,8089,total_duration_us,queryd_billing,queryd-v2-5797867574-kkzsw,0395bd7401aa3000
,,336,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:27:00.565066666Z,7697,total_duration_us,queryd_billing,queryd-v2-5797867574-kkzsw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,337,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:04.437709166Z,3871108,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,338,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:24:00.614708868Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,338,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:27:00.607072565Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,338,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:37:00.586640451Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,338,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.627600735Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,338,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:46:00.574052194Z,15,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,339,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.508804564Z,13743,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000
,,339,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.720702585Z,11709,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,340,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.863657529Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,341,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:01.077941715Z,213584,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,342,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.724536149Z,0,read_values,queryd_billing,queryd-v2-5797867574-xvkns,03c19003200d7000
,,342,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.585442572Z,0,read_values,queryd_billing,queryd-v2-5797867574-xvkns,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,343,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.783829632Z,180,read_values,queryd_billing,queryd-v2-5797867574-bd7fh,03d01b74c8e09000
,,343,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.757281711Z,180,read_values,queryd_billing,queryd-v2-5797867574-bd7fh,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,344,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:02.629173563Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,344,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:04.58472994Z,138631,response_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,344,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:01.025444871Z,1520,response_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,345,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:26:00.437242604Z,0,read_values,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,345,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:31:00.477712461Z,0,read_values,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,345,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:21:00.4986011Z,0,read_values,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,345,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:52:00.719041693Z,0,read_values,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,345,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:54:00.613488751Z,0,read_values,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,345,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:59:00.564883689Z,0,read_values,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,345,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:03:00.545102773Z,0,read_values,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,345,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.856694666Z,0,read_values,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,346,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:11:00.505877871Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,346,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:27:00.574693886Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,346,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:59:00.572427992Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,346,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:07:00.569599945Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,346,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:22:00.588925323Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,346,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.50045533Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,346,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:53:00.457936128Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,347,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.755759982Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000
,,347,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.774581825Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,348,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.925514969Z,270,read_values,queryd_billing,queryd-v2-5797867574-fmtgz,03d01b74c8e09000
,,348,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:01.84642159Z,10,read_values,queryd_billing,queryd-v2-5797867574-fmtgz,03d01b74c8e09000
,,348,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:01.561424858Z,0,read_values,queryd_billing,queryd-v2-5797867574-fmtgz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,349,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.952632098Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-sn26z,0395bd7401aa3000
,,349,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.896619226Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-sn26z,0395bd7401aa3000
,,349,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:23:00.191362562Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-sn26z,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,350,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:04.423740313Z,139402,response_bytes,queryd_billing,queryd-v2-5797867574-kssph,03d01b74c8e09000
,,350,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.715319869Z,1258,response_bytes,queryd_billing,queryd-v2-5797867574-kssph,03d01b74c8e09000
,,350,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.898664906Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kssph,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,351,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.609126611Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,351,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.770656494Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,351,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.469463621Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,351,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.71182255Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,352,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:24:00.614708868Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,352,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:27:00.607072565Z,5,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,352,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:37:00.586640451Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,352,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.627600735Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,352,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:46:00.574052194Z,8,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,353,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:06:00.513213184Z,7978,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,353,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:22:00.448283291Z,8629,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,353,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:28:00.484967147Z,8798,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,353,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:56:00.684591295Z,7678,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,354,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:31:00.627947836Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-mfspl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,355,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:52:00.663145742Z,8,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,355,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:31:00.655419185Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,355,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:44:00.514580555Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,356,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.515673793Z,12934,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000
,,356,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.590064004Z,35544,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,357,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:01.256241012Z,1519,response_bytes,queryd_billing,queryd-v2-5797867574-cm6fz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,358,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:48:00.675272795Z,7920,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,359,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.836044529Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000
,,359,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.425198714Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,360,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:01.256241012Z,160254,total_duration_us,queryd_billing,queryd-v2-5797867574-cm6fz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,361,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.832579331Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,361,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.727379572Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,361,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.812383085Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,362,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.783156493Z,159097,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000
,,362,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.825524632Z,235862,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,363,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.832140221Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,363,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:38:00.693295746Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,363,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:48:00.560832795Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,363,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.577141351Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,363,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.890962075Z,2,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,363,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.182960005Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,364,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:02.629173563Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,364,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:04.58472994Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,364,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:01.025444871Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,365,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:04.437709166Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,366,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:04.437709166Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,367,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.512542619Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,367,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.50851438Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,367,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.741858095Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,367,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.486155671Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,368,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.714848904Z,1260,response_bytes,queryd_billing,queryd-v2-5797867574-c88sh,03d01b74c8e09000
,,368,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.67290226Z,1260,response_bytes,queryd_billing,queryd-v2-5797867574-c88sh,03d01b74c8e09000
,,368,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:04.058154233Z,139260,response_bytes,queryd_billing,queryd-v2-5797867574-c88sh,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,369,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:01.256241012Z,270,read_values,queryd_billing,queryd-v2-5797867574-cm6fz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,370,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.813155264Z,11533,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,370,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.697447893Z,12936,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,370,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.780742525Z,11645,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,371,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:01.003107451Z,0,read_values,queryd_billing,queryd-v2-5797867574-qqx49,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,372,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.852953324Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-5pjgb,03c19003200d7000
,,372,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.655575144Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-5pjgb,03c19003200d7000
,,372,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.656976818Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-5pjgb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,373,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.685173538Z,17342,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,373,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.693548351Z,14727,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,373,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.623719778Z,39260,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,373,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.755486749Z,17634,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,374,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:57:00.444508534Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-t474n,0395bd7401aa3000
,,374,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:54:00.757802699Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-t474n,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,375,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.512542619Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,375,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.50851438Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,375,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.741858095Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,375,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.486155671Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,376,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:53:00.603537064Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,376,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.709640978Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,376,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:41:00.659356314Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,376,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:47:00.524120738Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,376,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:07:00.552515712Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,376,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:38:00.141966771Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,376,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.625087256Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,377,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.790663574Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,377,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.830213309Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,377,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.636824955Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,377,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.163289319Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,378,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.925514969Z,1519,response_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,03d01b74c8e09000
,,378,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:01.84642159Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,03d01b74c8e09000
,,378,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:01.561424858Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,379,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.758278197Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,379,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.837604512Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,379,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.614967009Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,380,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:04.173174511Z,3525139,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,380,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:04.02017492Z,3548275,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,380,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.990040524Z,330322,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,380,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:01.585834691Z,277366,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,380,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.67997434Z,197377,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,381,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.894829972Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,381,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.928682633Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,381,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.898959022Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,381,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.619773147Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,381,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.783903603Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,381,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.853962964Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,381,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.785243966Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,382,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:01.012279784Z,2160,read_bytes,queryd_billing,queryd-v1-5f699b6b58-2d862,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,383,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.924257631Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,384,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.168887758Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,385,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:04.170703823Z,3507369,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,386,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:58:00.427481584Z,0,read_values,queryd_billing,queryd-v2-5797867574-bd7fh,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,387,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.492227407Z,17565,total_duration_us,queryd_billing,queryd-v2-5797867574-zlf4s,03c19003200d7000
,,387,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.632944642Z,15596,total_duration_us,queryd_billing,queryd-v2-5797867574-zlf4s,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,388,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:09:00.543511505Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kkzsw,0395bd7401aa3000
,,388,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:27:00.565066666Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kkzsw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,389,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.848026232Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,389,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.861699773Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,389,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.989693911Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,389,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:04.171157376Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,390,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.821668012Z,1013,response_bytes,queryd_billing,queryd-v1-5f699b6b58-9lsws,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,391,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.749034925Z,20502,total_duration_us,queryd_billing,queryd-v2-5797867574-zmcl2,03c19003200d7000
,,391,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.721679848Z,12973,total_duration_us,queryd_billing,queryd-v2-5797867574-zmcl2,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,392,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975813826Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,393,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:23:00.482973775Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,393,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:42:00.476029727Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,393,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:24:00.494923251Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,393,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.800071949Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,394,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.821352355Z,180,read_values,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,395,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.976921236Z,2160,read_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,03d01b74c8e09000
,,395,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.833400375Z,1440,read_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,396,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:48:00.675272795Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,397,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:31:00.627947836Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-mfspl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,398,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:23:00.482973775Z,2,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,398,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:42:00.476029727Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,398,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:24:00.494923251Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,398,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.800071949Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,399,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.848026232Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,399,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.861699773Z,27,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,399,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.989693911Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,399,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:04.171157376Z,9,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,400,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.686930041Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,400,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:57:00.433078105Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,400,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:17:00.415381738Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,401,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:41:00.536966287Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-7s4z2,0395bd7401aa3000
,,401,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:08:00.66861141Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-7s4z2,0395bd7401aa3000
,,401,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:39:00.429663199Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-7s4z2,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,402,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.730269127Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,402,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:04:00.506173446Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,402,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:08:00.618513396Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,402,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:13:00.133001433Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,402,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.921225459Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,402,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:48:00.633450477Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,403,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:16:00.504902703Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000
,,403,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:43:00.504989877Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000
,,403,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.562309784Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000
,,403,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.70656956Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,404,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.471264316Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,404,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.532354811Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,404,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.58982965Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,404,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.736000374Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,404,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.62852928Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,405,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:49:00.426081648Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,406,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.492088454Z,17601,total_duration_us,queryd_billing,queryd-v2-5797867574-t474n,03c19003200d7000
,,406,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.714536617Z,52675,total_duration_us,queryd_billing,queryd-v2-5797867574-t474n,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,407,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.857012416Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,407,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:28:00.450725793Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,407,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:32:00.590667734Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,407,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:39:00.577723384Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,407,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:11:00.598135316Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,408,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:04.170703823Z,34183,read_values,queryd_billing,queryd-v1-5f699b6b58-wbczl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,409,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:06:00.455332774Z,9412,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,409,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:42:00.541039897Z,8465,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,409,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:29:00.477956027Z,13352,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,409,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:33:00.515658208Z,9918,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,409,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:11:00.667516371Z,8729,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,410,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:12:00.462956123Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,411,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:14:00.606703032Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,411,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.658228927Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,411,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.675835865Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,411,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:02:00.441622158Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,411,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:33:00.490131246Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,411,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:47:00.493465617Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,411,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:58:00.69307463Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,412,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.721543002Z,29964,total_duration_us,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000
,,412,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.578924493Z,18766,total_duration_us,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000
,,412,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.685639229Z,15690,total_duration_us,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000
,,412,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.610618315Z,11001,total_duration_us,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,413,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.912019031Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-lj72r,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,414,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:33:00.48797071Z,17,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,414,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:04:00.484123404Z,20,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,414,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.729548006Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,415,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.609126611Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,415,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.770656494Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,415,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.469463621Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,415,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.71182255Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,416,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.758278197Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,416,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.837604512Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,416,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.614967009Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,417,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.661988211Z,0,read_values,queryd_billing,queryd-v2-5797867574-b2bc5,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,418,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.168887758Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,419,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.925514969Z,2160,read_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,03d01b74c8e09000
,,419,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:01.84642159Z,169,read_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,03d01b74c8e09000
,,419,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:01.561424858Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,420,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.929248638Z,4,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000
,,420,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:04.160876428Z,34183,read_values,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,421,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.6467672Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-mfspl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,422,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.821668012Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,423,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:04.170703823Z,273464,read_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,424,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.282532978Z,176270,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000
,,424,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.97843698Z,205297,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,425,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:52:00.663145742Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,425,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:31:00.655419185Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,425,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:44:00.514580555Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,426,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:23:00.482973775Z,8918,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,426,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:42:00.476029727Z,8015,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,426,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:24:00.494923251Z,12328,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,426,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.800071949Z,8315,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,427,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:12:00.513928572Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000
,,427,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:52:00.1555046Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,428,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.929248638Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000
,,428,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:04.160876428Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,429,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.701333339Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,429,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.82139065Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,429,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.811423271Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,430,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.580021946Z,20144,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000
,,430,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.59833827Z,19217,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,431,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.515673793Z,12891,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000
,,431,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.590064004Z,35505,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,432,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.836044529Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000
,,432,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.425198714Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,433,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.976921236Z,166052,total_duration_us,queryd_billing,queryd-v2-5797867574-hh9fz,03d01b74c8e09000
,,433,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.833400375Z,176952,total_duration_us,queryd_billing,queryd-v2-5797867574-hh9fz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,434,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:41:00.536966287Z,0,read_values,queryd_billing,queryd-v2-5797867574-7s4z2,0395bd7401aa3000
,,434,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:08:00.66861141Z,0,read_values,queryd_billing,queryd-v2-5797867574-7s4z2,0395bd7401aa3000
,,434,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:39:00.429663199Z,0,read_values,queryd_billing,queryd-v2-5797867574-7s4z2,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,435,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:01.16148791Z,1440,read_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,03d01b74c8e09000
,,435,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.15254861Z,1440,read_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,436,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.663457967Z,19001,total_duration_us,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000
,,436,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.674816715Z,21092,total_duration_us,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000
,,436,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.729511016Z,18064,total_duration_us,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000
,,436,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.501504684Z,14932,total_duration_us,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,437,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.762544227Z,19546,total_duration_us,queryd_billing,queryd-v2-5797867574-xbx7c,03c19003200d7000
,,437,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.830581156Z,17062,total_duration_us,queryd_billing,queryd-v2-5797867574-xbx7c,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,438,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.929248638Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000
,,438,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:04.160876428Z,139413,response_bytes,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,439,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.924157523Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,439,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.579798761Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,439,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.763690024Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,439,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:18:00.533780271Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,439,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.596326558Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,439,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:42:00.44613233Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,440,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.976921236Z,1518,response_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,03d01b74c8e09000
,,440,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.833400375Z,1260,response_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,441,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:31:00.627947836Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-mfspl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,442,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:31:00.627947836Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-mfspl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,443,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.755759982Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000
,,443,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.774581825Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,444,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:04.173174511Z,273464,read_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,444,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:04.02017492Z,273464,read_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,444,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.990040524Z,1440,read_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,444,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:01.585834691Z,169,read_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,444,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.67997434Z,1440,read_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,445,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.937640598Z,180,read_values,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000
,,445,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.803909311Z,10,read_values,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,446,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:06:00.455340709Z,8513,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,446,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:09:00.434414481Z,7289,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,446,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.653985084Z,8158,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,446,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.767478932Z,7521,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,447,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.777690142Z,1015,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,447,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:03.873228519Z,139364,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,447,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975537336Z,1517,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,447,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.976567453Z,1519,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,448,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.875119313Z,7420,total_duration_us,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,448,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.965625042Z,25492,total_duration_us,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,448,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.551411546Z,8820,total_duration_us,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,448,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.818106383Z,7677,total_duration_us,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,448,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.706945865Z,7932,total_duration_us,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,449,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:53:00.603537064Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,449,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.709640978Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,449,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:41:00.659356314Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,449,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:47:00.524120738Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,449,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:07:00.552515712Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,449,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:38:00.141966771Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,449,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.625087256Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,450,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.749802396Z,180,read_values,queryd_billing,queryd-v2-5797867574-rj2ns,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,451,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.714848904Z,1440,read_bytes,queryd_billing,queryd-v2-5797867574-c88sh,03d01b74c8e09000
,,451,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.67290226Z,1440,read_bytes,queryd_billing,queryd-v2-5797867574-c88sh,03d01b74c8e09000
,,451,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:04.058154233Z,273464,read_bytes,queryd_billing,queryd-v2-5797867574-c88sh,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,452,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.848026232Z,350238,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,452,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.861699773Z,181140,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,452,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.989693911Z,271956,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,452,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:04.171157376Z,3494376,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,453,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.783156493Z,1440,read_bytes,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000
,,453,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.825524632Z,1440,read_bytes,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,454,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:04.173174511Z,3525104,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,454,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:04.02017492Z,3548237,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,454,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.990040524Z,199420,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,454,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:01.585834691Z,277320,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,454,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.67997434Z,197342,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,455,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:02.661499487Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,455,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:01.057331112Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,455,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:02.601634416Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,456,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:09:00.695676986Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000
,,456,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:16:00.423571342Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,457,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.61520984Z,0,read_values,queryd_billing,queryd-v2-5797867574-c88sh,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,458,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:12:00.462956123Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-j8hm4,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,459,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:34:00.547848426Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,459,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:51:00.496150236Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,459,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:21:00.620889507Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,459,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:26:00.675475921Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,459,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:43:00.567823817Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,460,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975813826Z,1440,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,461,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:14:00.606703032Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,461,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.658228927Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,461,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.675835865Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,461,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:02:00.441622158Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,461,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:33:00.490131246Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,461,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:47:00.493465617Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,461,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:58:00.69307463Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,462,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.90791717Z,13633,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,463,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.681462039Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,463,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.955855408Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,463,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.748825278Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,464,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.848026232Z,360,read_values,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,464,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.861699773Z,180,read_values,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,464,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.989693911Z,180,read_values,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,464,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:04.171157376Z,34183,read_values,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,465,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.53535901Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,466,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.609126611Z,25428,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,466,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.770656494Z,17651,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,466,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.469463621Z,13145,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,466,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.71182255Z,46282,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,467,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:52:00.663145742Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,467,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:31:00.655419185Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,467,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:44:00.514580555Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,468,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:36:00.654229914Z,0,read_values,queryd_billing,queryd-v2-5797867574-5ff7l,0395bd7401aa3000
,,468,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:59:00.617599152Z,0,read_values,queryd_billing,queryd-v2-5797867574-5ff7l,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,469,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.681462039Z,37697,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,469,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.955855408Z,11770,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,469,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.748825278Z,18159,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,470,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.054579583Z,2880,read_bytes,queryd_billing,queryd-v2-5797867574-xbx7c,03d01b74c8e09000
,,470,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:04.057304496Z,273392,read_bytes,queryd_billing,queryd-v2-5797867574-xbx7c,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,471,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.937640598Z,1440,read_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000
,,471,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.803909311Z,169,read_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,472,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:33:00.48797071Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,472,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:04:00.484123404Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,472,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.729548006Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,473,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:12:00.513928572Z,2,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000
,,473,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:52:00.1555046Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,474,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.790663574Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,474,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.830213309Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,474,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.636824955Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,474,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.163289319Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,475,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.580021946Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000
,,475,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.59833827Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,476,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:02.629173563Z,169,read_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,476,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:04.58472994Z,273464,read_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,476,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:01.025444871Z,2160,read_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,477,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975813826Z,1258,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,478,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:23:00.482973775Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,478,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:42:00.476029727Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,478,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:24:00.494923251Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,478,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.800071949Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,479,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.836044529Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000
,,479,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.425198714Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,480,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.785640834Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,480,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.620654875Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,480,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:12:00.462658028Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,481,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.832579331Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,481,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.727379572Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,481,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.812383085Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,482,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.924157523Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,482,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.579798761Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,482,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.763690024Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,482,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:18:00.533780271Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,482,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.596326558Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,482,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:42:00.44613233Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,483,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.675124012Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000
,,483,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.92728318Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,484,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:37:00.423986232Z,7282,total_duration_us,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,484,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.759512962Z,8036,total_duration_us,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,484,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:28:00.377646402Z,7145,total_duration_us,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,484,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:44:00.420950673Z,8281,total_duration_us,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,484,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:32:00.605366491Z,9103,total_duration_us,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,484,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:49:00.463047225Z,8117,total_duration_us,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,485,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.929248638Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000
,,485,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:04.160876428Z,17,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,486,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.867420941Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-s6t85,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,487,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.821352355Z,224021,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,488,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:23:00.473713274Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-tlhkl,0395bd7401aa3000
,,488,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.647808347Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-tlhkl,0395bd7401aa3000
,,488,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:43:00.561679905Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-tlhkl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,489,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:23:00.482973775Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,489,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:42:00.476029727Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,489,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:24:00.494923251Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,489,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.800071949Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,490,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:36:00.654229914Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-5ff7l,0395bd7401aa3000
,,490,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:59:00.617599152Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-5ff7l,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,491,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:06:00.455340709Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,491,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:09:00.434414481Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,491,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.653985084Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,491,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.767478932Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,492,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.777690142Z,159911,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,492,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:03.873228519Z,3399060,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,492,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975537336Z,163993,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,492,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.976567453Z,176156,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,493,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.821352355Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,494,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.863657529Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,495,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.52560325Z,12807,total_duration_us,queryd_billing,queryd-v2-5797867574-plnml,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,496,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.585156227Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000
,,496,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.663976545Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,497,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:02.661499487Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,497,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:01.057331112Z,1518,response_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,497,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:02.601634416Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,498,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.937640598Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000
,,498,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.803909311Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,499,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.828458662Z,180,read_values,queryd_billing,queryd-v2-5797867574-t474n,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,500,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:02.629173563Z,233871,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,500,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:04.58472994Z,3969110,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,500,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:01.025444871Z,223184,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,501,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.701333339Z,36676,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,501,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.82139065Z,31236,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,501,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.811423271Z,12037,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,502,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:02.661499487Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,502,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:01.057331112Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,502,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:02.601634416Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,503,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.609126611Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,503,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.770656494Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,503,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.469463621Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000
,,503,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.71182255Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,504,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.760513754Z,10050,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,504,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:24:00.625390315Z,9961,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,504,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:57:00.617251549Z,7932,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,505,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.875119313Z,0,read_values,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,505,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.965625042Z,0,read_values,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,505,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.551411546Z,0,read_values,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,505,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.818106383Z,0,read_values,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000
,,505,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.706945865Z,0,read_values,queryd_billing,queryd-v2-5797867574-s6t85,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,506,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.917582859Z,180,read_values,queryd_billing,queryd-v1-5f699b6b58-467pb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,507,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.054579583Z,360,read_values,queryd_billing,queryd-v2-5797867574-xbx7c,03d01b74c8e09000
,,507,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:04.057304496Z,34174,read_values,queryd_billing,queryd-v2-5797867574-xbx7c,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,508,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.481079117Z,13389,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,509,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.512542619Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,509,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.50851438Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,509,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.741858095Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,509,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.486155671Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,510,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.968941865Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kvktv,0395bd7401aa3000
,,510,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:13:00.694273957Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kvktv,0395bd7401aa3000
,,510,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:36:00.447798087Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kvktv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,511,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.580021946Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000
,,511,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.59833827Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,512,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.675124012Z,12019,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000
,,512,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.92728318Z,19450,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,513,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:06:00.455340709Z,8476,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,513,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:09:00.434414481Z,7259,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,513,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.653985084Z,8126,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,513,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.767478932Z,7493,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,514,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.776577884Z,10,read_values,queryd_billing,queryd-v2-5797867574-tlhkl,03d01b74c8e09000
,,514,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:01.170233871Z,360,read_values,queryd_billing,queryd-v2-5797867574-tlhkl,03d01b74c8e09000
,,514,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:01.25392002Z,0,read_values,queryd_billing,queryd-v2-5797867574-tlhkl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,515,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:33:00.48797071Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,515,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:04:00.484123404Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,515,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.729548006Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,516,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.789866842Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,516,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.635709149Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,516,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:01.216435523Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,517,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.749034925Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zmcl2,03c19003200d7000
,,517,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.721679848Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zmcl2,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,518,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.716500242Z,36893,total_duration_us,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,518,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.675300682Z,11224,total_duration_us,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,518,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.722782443Z,7814,total_duration_us,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,518,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:46:00.61084851Z,8050,total_duration_us,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,518,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:53:00.659149488Z,8040,total_duration_us,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,519,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.762544227Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-xbx7c,03c19003200d7000
,,519,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.830581156Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-xbx7c,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,520,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:48:00.675272795Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-66kcw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,521,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.054579583Z,999,response_bytes,queryd_billing,queryd-v2-5797867574-xbx7c,03d01b74c8e09000
,,521,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:04.057304496Z,138642,response_bytes,queryd_billing,queryd-v2-5797867574-xbx7c,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,522,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.857012416Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,522,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:28:00.450725793Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,522,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:32:00.590667734Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,522,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:39:00.577723384Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,522,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:11:00.598135316Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,523,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:04.437709166Z,139379,response_bytes,queryd_billing,queryd-v1-5f699b6b58-66kcw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,524,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.512542619Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,524,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.50851438Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,524,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.741858095Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,524,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.486155671Z,17,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,525,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:58:00.427481584Z,7357,total_duration_us,queryd_billing,queryd-v2-5797867574-bd7fh,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,526,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.666298715Z,180,read_values,queryd_billing,queryd-v2-5797867574-plnml,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,527,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.894829972Z,148258,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,527,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.928682633Z,31600,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,527,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.898959022Z,12194,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,527,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.619773147Z,19196,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,527,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.783903603Z,14151,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,527,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.853962964Z,27959,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,527,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.785243966Z,15842,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,528,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.912019031Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,529,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:29:00.521460892Z,19,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,529,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:34:00.474156777Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,529,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.914502559Z,15,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,529,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:39:00.487057271Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,530,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.828458662Z,1440,read_bytes,queryd_billing,queryd-v2-5797867574-t474n,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,531,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.542805667Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,531,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.794869556Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,531,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:02:00.468421939Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,531,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:18:00.466768224Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,531,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.738772673Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,531,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:01:00.650919032Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,531,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:13:00.594089157Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,532,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.61520984Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-c88sh,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,533,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.585156227Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000
,,533,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.663976545Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,534,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.70702057Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000
,,534,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:04:00.522827935Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,535,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.515673793Z,24,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000
,,535,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.590064004Z,17,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,536,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:04.173174511Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,536,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:04.02017492Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,536,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.990040524Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,536,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:01.585834691Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,536,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.67997434Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,537,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:23:00.482973775Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,537,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:42:00.476029727Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,537,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:24:00.494923251Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,537,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.800071949Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,538,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:02.661499487Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,538,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:01.057331112Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,538,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:02.601634416Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,539,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.644202789Z,156475,total_duration_us,queryd_billing,queryd-v2-5797867574-qqx49,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,540,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.516727014Z,13615,total_duration_us,queryd_billing,queryd-v2-5797867574-bd7fh,03c19003200d7000
,,540,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.776060648Z,16810,total_duration_us,queryd_billing,queryd-v2-5797867574-bd7fh,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,541,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:24:00.614708868Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,541,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:27:00.607072565Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,541,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:37:00.586640451Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,541,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.627600735Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,541,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:46:00.574052194Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,542,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.813155264Z,11573,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,542,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.697447893Z,12972,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,542,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.780742525Z,11680,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,543,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.760513754Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,543,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:24:00.625390315Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,543,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:57:00.617251549Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,544,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.821668012Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,545,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:04.437709166Z,26,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,546,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.730269127Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,546,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:04:00.506173446Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,546,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:08:00.618513396Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,546,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:13:00.133001433Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,546,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.921225459Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,546,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:48:00.633450477Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,547,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.508804564Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000
,,547,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.720702585Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,548,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:53:00.603537064Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,548,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.709640978Z,7,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,548,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:41:00.659356314Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,548,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:47:00.524120738Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,548,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:07:00.552515712Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,548,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:38:00.141966771Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,548,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.625087256Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,549,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:09:00.543511505Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kkzsw,0395bd7401aa3000
,,549,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:27:00.565066666Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kkzsw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,550,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.686930041Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,550,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:57:00.433078105Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,550,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:17:00.415381738Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,551,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:58:00.427481584Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-bd7fh,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,552,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.282532978Z,2160,read_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000
,,552,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.97843698Z,2160,read_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,553,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.471264316Z,20525,total_duration_us,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,553,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.532354811Z,12622,total_duration_us,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,553,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.58982965Z,12829,total_duration_us,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,553,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.736000374Z,41590,total_duration_us,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,553,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.62852928Z,18236,total_duration_us,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,554,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:01.040258084Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000
,,554,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.52727879Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000
,,554,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.472275138Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000
,,554,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.581634661Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kkzsw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,555,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:08:00.566922506Z,0,read_values,queryd_billing,queryd-v2-5797867574-xbx7c,0395bd7401aa3000
,,555,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:19:00.411422463Z,0,read_values,queryd_billing,queryd-v2-5797867574-xbx7c,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,556,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.90791717Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,557,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.675124012Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000
,,557,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.92728318Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,558,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.685173538Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,558,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.693548351Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,558,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.623719778Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,558,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.755486749Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,559,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:12:00.513928572Z,8261,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000
,,559,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:52:00.1555046Z,8251,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,560,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.937640598Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000
,,560,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.803909311Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,561,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:24:00.614708868Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,561,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:27:00.607072565Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,561,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:37:00.586640451Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,561,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.627600735Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,561,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:46:00.574052194Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,562,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.530529897Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-7s4z2,03c19003200d7000
,,562,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.642695941Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-7s4z2,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,563,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.644202789Z,1440,read_bytes,queryd_billing,queryd-v2-5797867574-qqx49,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,564,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.777690142Z,180,read_values,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,564,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:03.873228519Z,34183,read_values,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,564,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975537336Z,270,read_values,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,564,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.976567453Z,270,read_values,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,565,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:01.003107451Z,21324,total_duration_us,queryd_billing,queryd-v2-5797867574-qqx49,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,566,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:02.661499487Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,566,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:01.057331112Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000
,,566,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:02.601634416Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,567,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.821352355Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,568,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.767777421Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-rj2ns,03c19003200d7000
,,568,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.779496475Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-rj2ns,03c19003200d7000
,,568,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.727572623Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-rj2ns,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,569,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:06:00.455332774Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,569,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:42:00.541039897Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,569,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:29:00.477956027Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,569,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:33:00.515658208Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,569,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:11:00.667516371Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,570,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.917582859Z,268781,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,571,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.755759982Z,18258,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000
,,571,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.774581825Z,69890,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,572,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.789866842Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,572,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.635709149Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,572,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:01.216435523Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,573,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.580021946Z,20109,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000
,,573,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.59833827Z,19185,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,574,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.663457967Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000
,,574,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.674816715Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000
,,574,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.729511016Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000
,,574,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.501504684Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,575,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:01.077941715Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,576,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:36:00.654229914Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-5ff7l,0395bd7401aa3000
,,576,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:59:00.617599152Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-5ff7l,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,577,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.70702057Z,15,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000
,,577,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:04:00.522827935Z,16,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,578,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.282532978Z,270,read_values,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000
,,578,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.97843698Z,270,read_values,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,579,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.968941865Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kvktv,0395bd7401aa3000
,,579,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:13:00.694273957Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kvktv,0395bd7401aa3000
,,579,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:36:00.447798087Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kvktv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,580,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.53535901Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,581,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:57:00.444508534Z,8350,total_duration_us,queryd_billing,queryd-v2-5797867574-t474n,0395bd7401aa3000
,,581,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:54:00.757802699Z,7694,total_duration_us,queryd_billing,queryd-v2-5797867574-t474n,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,582,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:04.41909794Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000
,,582,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:01.610831987Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,583,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.686930041Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,583,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:57:00.433078105Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,583,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:17:00.415381738Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,584,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.924257631Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,585,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.492088454Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-t474n,03c19003200d7000
,,585,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.714536617Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-t474n,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,586,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.508804564Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000
,,586,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.720702585Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,587,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.894829972Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,587,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.928682633Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,587,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.898959022Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,587,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.619773147Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,587,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.783903603Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,587,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.853962964Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,587,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.785243966Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,588,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.580021946Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000
,,588,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.59833827Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,589,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.70702057Z,8358,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000
,,589,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:04:00.522827935Z,11397,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,590,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.53535901Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,591,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:34:00.547848426Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,591,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:51:00.496150236Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,591,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:21:00.620889507Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,591,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:26:00.675475921Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,591,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:43:00.567823817Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,592,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:01:00.492154788Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,592,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:36:00.45783491Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,592,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:38:00.44515579Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,592,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:44:00.587165743Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,592,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.65000697Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,592,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:34:00.445880939Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,593,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.749802396Z,1440,read_bytes,queryd_billing,queryd-v2-5797867574-rj2ns,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,594,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:37:00.423986232Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,594,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.759512962Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,594,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:28:00.377646402Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,594,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:44:00.420950673Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,594,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:32:00.605366491Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,594,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:49:00.463047225Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,595,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:06:00.513213184Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,595,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:22:00.448283291Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,595,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:28:00.484967147Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,595,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:56:00.684591295Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,596,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.832579331Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,596,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.727379572Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,596,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.812383085Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,597,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.836044529Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000
,,597,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.425198714Z,34,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,598,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.760513754Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,598,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:24:00.625390315Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,598,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:57:00.617251549Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,599,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.789866842Z,11149,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,599,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.635709149Z,11014,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,599,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:01.216435523Z,11706,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,600,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:29:00.521460892Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,600,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:34:00.474156777Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,600,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.914502559Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,600,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:39:00.487057271Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,601,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:18:00.151904896Z,7887,total_duration_us,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,601,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:56:00.450344765Z,8633,total_duration_us,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,601,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:17:00.634378145Z,7588,total_duration_us,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,601,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:19:00.610282493Z,8522,total_duration_us,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,601,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:29:00.532131872Z,9775,total_duration_us,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,602,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.681462039Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,602,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.955855408Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,602,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.748825278Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,603,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:52:00.663145742Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,603,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:31:00.655419185Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,603,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:44:00.514580555Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,604,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.481079117Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,605,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.585156227Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000
,,605,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.663976545Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,606,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.783156493Z,159063,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000
,,606,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.825524632Z,235830,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,607,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.685173538Z,17309,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,607,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.693548351Z,14696,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,607,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.623719778Z,39225,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,607,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.755486749Z,17599,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,608,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.785640834Z,7210,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,608,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.620654875Z,14400,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,608,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:12:00.462658028Z,8436,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,609,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.821352355Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,610,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.760513754Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,610,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:24:00.625390315Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,610,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:57:00.617251549Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,611,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:01.012279784Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,612,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.821352355Z,18,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,613,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.924257631Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,614,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975813826Z,327007,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,615,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:24:00.614708868Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,615,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:27:00.607072565Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,615,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:37:00.586640451Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,615,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.627600735Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,615,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:46:00.574052194Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,616,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:49:00.426081648Z,7870,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,617,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.824636831Z,225233,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000
,,617,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.625636236Z,150824,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,618,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.836044529Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000
,,618,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.425198714Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,619,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.70702057Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000
,,619,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:04:00.522827935Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,620,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.492227407Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,03c19003200d7000
,,620,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.632944642Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,621,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:01:00.492154788Z,0,read_values,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,621,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:36:00.45783491Z,0,read_values,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,621,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:38:00.44515579Z,0,read_values,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,621,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:44:00.587165743Z,0,read_values,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,621,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.65000697Z,0,read_values,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000
,,621,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:34:00.445880939Z,0,read_values,queryd_billing,queryd-v2-5797867574-b2bc5,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,622,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.824636831Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000
,,622,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.625636236Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,623,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.924257631Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,624,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:09:00.695676986Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000
,,624,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:16:00.423571342Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,625,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:16:00.501326426Z,0,read_values,queryd_billing,queryd-v2-5797867574-rj2ns,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,626,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.512542619Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,626,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.50851438Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,626,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.741858095Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,626,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.486155671Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,627,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.686930041Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,627,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:57:00.433078105Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,627,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:17:00.415381738Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,628,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.90791717Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,629,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.721543002Z,0,read_values,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000
,,629,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.578924493Z,0,read_values,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000
,,629,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.685639229Z,0,read_values,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000
,,629,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.610618315Z,0,read_values,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,630,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.848026232Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,630,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.861699773Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,630,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.989693911Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,630,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:04.171157376Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,631,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.819675585Z,1440,read_bytes,queryd_billing,queryd-v2-5797867574-xvkns,03d01b74c8e09000
,,631,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.79114313Z,1440,read_bytes,queryd_billing,queryd-v2-5797867574-xvkns,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,632,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:12:00.462956123Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-j8hm4,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,633,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.168887758Z,11306,total_duration_us,queryd_billing,queryd-v2-5797867574-fmtgz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,634,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.724536149Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-xvkns,03c19003200d7000
,,634,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.585442572Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-xvkns,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,635,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.790663574Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,635,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.830213309Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,635,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.636824955Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,635,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.163289319Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,636,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:58:00.427481584Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-bd7fh,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,637,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.6467672Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-mfspl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,638,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:18:00.151904896Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,638,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:56:00.450344765Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,638,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:17:00.634378145Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,638,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:19:00.610282493Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000
,,638,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:29:00.532131872Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zmcl2,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,639,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.52560325Z,0,read_values,queryd_billing,queryd-v2-5797867574-plnml,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,640,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.924157523Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,640,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.579798761Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,640,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.763690024Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,640,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:18:00.533780271Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,640,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.596326558Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,640,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:42:00.44613233Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,641,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.724536149Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-xvkns,03c19003200d7000
,,641,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.585442572Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-xvkns,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,642,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.282532978Z,17,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000
,,642,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.97843698Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,643,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.760513754Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,643,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:24:00.625390315Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,643,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:57:00.617251549Z,16,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,644,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.760513754Z,10020,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,644,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:24:00.625390315Z,9932,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,644,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:57:00.617251549Z,7893,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,645,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.675124012Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000
,,645,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.92728318Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,646,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.968941865Z,0,read_values,queryd_billing,queryd-v2-5797867574-kvktv,0395bd7401aa3000
,,646,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:13:00.694273957Z,0,read_values,queryd_billing,queryd-v2-5797867574-kvktv,0395bd7401aa3000
,,646,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:36:00.447798087Z,0,read_values,queryd_billing,queryd-v2-5797867574-kvktv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,647,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.937640598Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000
,,647,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.803909311Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,648,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:09:00.695676986Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000
,,648,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:16:00.423571342Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,649,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:21:00.679948866Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,649,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:46:00.57791668Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,649,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:47:00.665339757Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,649,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.885284853Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,649,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:37:00.490611137Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,650,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.777690142Z,159877,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,650,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:03.873228519Z,3399021,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,650,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975537336Z,163958,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,650,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.976567453Z,176120,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,651,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:09:00.695676986Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000
,,651,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:16:00.423571342Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,652,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.924157523Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,652,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.579798761Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,652,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.763690024Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,652,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:18:00.533780271Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,652,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.596326558Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,652,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:42:00.44613233Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,653,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:04.170703823Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,654,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.721543002Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000
,,654,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.578924493Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000
,,654,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.685639229Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000
,,654,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.610618315Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hzlgm,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,655,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.783156493Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000
,,655,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.825524632Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,656,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:01.077941715Z,1519,response_bytes,queryd_billing,queryd-v1-5f699b6b58-t7slt,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,657,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.585156227Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000
,,657,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.663976545Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,658,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:17:00.44658031Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,658,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:22:00.620511505Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,658,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:49:00.504522138Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,658,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:03:00.458527039Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,658,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:32:00.562507962Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,659,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.929248638Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000
,,659,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:04.160876428Z,8,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,660,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:04.41909794Z,3647368,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000
,,660,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:01.610831987Z,248754,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,661,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:12:00.513928572Z,8232,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000
,,661,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:52:00.1555046Z,8213,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,662,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.282532978Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000
,,662,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.97843698Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,663,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.776577884Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-tlhkl,03d01b74c8e09000
,,663,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:01.170233871Z,1003,response_bytes,queryd_billing,queryd-v2-5797867574-tlhkl,03d01b74c8e09000
,,663,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:01.25392002Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-tlhkl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,664,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.774815315Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-5pjgb,0395bd7401aa3000
,,664,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.710588962Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-5pjgb,0395bd7401aa3000
,,664,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.518786657Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-5pjgb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,665,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:12:00.462956123Z,7738,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,666,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.924257631Z,19116,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,667,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.685173538Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,667,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.693548351Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,667,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.623719778Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,667,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.755486749Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,668,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.783156493Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000
,,668,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.825524632Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,669,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:01.077941715Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,670,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.701333339Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,670,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.82139065Z,28,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,670,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.811423271Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,671,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.924257631Z,19160,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,672,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.937640598Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000
,,672,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.803909311Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,673,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.580021946Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000
,,673,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.59833827Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,674,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:41:00.536966287Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-7s4z2,0395bd7401aa3000
,,674,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:08:00.66861141Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-7s4z2,0395bd7401aa3000
,,674,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:39:00.429663199Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-7s4z2,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,675,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.681462039Z,37738,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,675,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.955855408Z,11803,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,675,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.748825278Z,18189,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,676,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.785640834Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,676,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.620654875Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,676,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:12:00.462658028Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,677,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.580021946Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000
,,677,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.59833827Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,678,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.701333339Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,678,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.82139065Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,678,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.811423271Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,679,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:16:00.504902703Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000
,,679,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:43:00.504989877Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000
,,679,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.562309784Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000
,,679,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.70656956Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,680,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.585156227Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000
,,680,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.663976545Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,681,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:48:00.675272795Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-66kcw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,682,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.542805667Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,682,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.794869556Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,682,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:02:00.468421939Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,682,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:18:00.466768224Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,682,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.738772673Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,682,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:01:00.650919032Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,682,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:13:00.594089157Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,683,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.790663574Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,683,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.830213309Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,683,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.636824955Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000
,,683,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.163289319Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,684,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:12:00.513928572Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000
,,684,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:52:00.1555046Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,685,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.785640834Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,685,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.620654875Z,29,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,685,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:12:00.462658028Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,686,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.512542619Z,16681,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,686,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.50851438Z,16368,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,686,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.741858095Z,22722,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,686,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.486155671Z,14911,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,687,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:06:00.513213184Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,687,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:22:00.448283291Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,687,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:28:00.484967147Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,687,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:56:00.684591295Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,688,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.785640834Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,688,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.620654875Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,688,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:12:00.462658028Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,689,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.828458662Z,1030,response_bytes,queryd_billing,queryd-v2-5797867574-t474n,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,690,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:34:00.547848426Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,690,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:51:00.496150236Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,690,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:21:00.620889507Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,690,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:26:00.675475921Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,690,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:43:00.567823817Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,691,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:16:00.504902703Z,11221,total_duration_us,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000
,,691,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:43:00.504989877Z,9868,total_duration_us,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000
,,691,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.562309784Z,7634,total_duration_us,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000
,,691,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.70656956Z,15392,total_duration_us,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,692,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.924257631Z,17,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,693,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:04.173174511Z,139413,response_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,693,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:04.02017492Z,139424,response_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,693,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.990040524Z,1021,response_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,693,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:01.585834691Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,693,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.67997434Z,1023,response_bytes,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,694,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:57:00.444508534Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-t474n,0395bd7401aa3000
,,694,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:54:00.757802699Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-t474n,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,695,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.481079117Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-t7slt,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,696,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.836044529Z,11650,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000
,,696,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.425198714Z,12637,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,697,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:06:00.455340709Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,697,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:09:00.434414481Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,697,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.653985084Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,697,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.767478932Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,698,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:21:00.679948866Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,698,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:46:00.57791668Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,698,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:47:00.665339757Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,698,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.885284853Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,698,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:37:00.490611137Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,699,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.767777421Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-rj2ns,03c19003200d7000
,,699,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.779496475Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-rj2ns,03c19003200d7000
,,699,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.727572623Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-rj2ns,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,700,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.542805667Z,0,read_values,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,700,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.794869556Z,0,read_values,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,700,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:02:00.468421939Z,0,read_values,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,700,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:18:00.466768224Z,0,read_values,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,700,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.738772673Z,0,read_values,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,700,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:01:00.650919032Z,0,read_values,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000
,,700,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:13:00.594089157Z,0,read_values,queryd_billing,queryd-v2-5797867574-k254f,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,701,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:17:00.44658031Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,701,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:22:00.620511505Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,701,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:49:00.504522138Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,701,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:03:00.458527039Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000
,,701,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:32:00.562507962Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-hh9fz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,702,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.512542619Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,702,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.50851438Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,702,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.741858095Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000
,,702,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.486155671Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-cd4cc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,703,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:31:00.627947836Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-mfspl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,704,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:14:00.572403792Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,704,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.687631407Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,704,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:14:00.48545546Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,704,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:41:00.534038417Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,704,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:51:00.491763198Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,705,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.661988211Z,24588,total_duration_us,queryd_billing,queryd-v2-5797867574-b2bc5,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,706,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975813826Z,2,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,707,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.832140221Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,707,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:38:00.693295746Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,707,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:48:00.560832795Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,707,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.577141351Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,707,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.890962075Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,707,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.182960005Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,708,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:04.437709166Z,34183,read_values,queryd_billing,queryd-v1-5f699b6b58-66kcw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,709,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:31:00.627947836Z,11088,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-mfspl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,710,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.952632098Z,7576,total_duration_us,queryd_billing,queryd-v2-5797867574-sn26z,0395bd7401aa3000
,,710,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.896619226Z,7755,total_duration_us,queryd_billing,queryd-v2-5797867574-sn26z,0395bd7401aa3000
,,710,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:23:00.191362562Z,8741,total_duration_us,queryd_billing,queryd-v2-5797867574-sn26z,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,711,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.663457967Z,0,read_values,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000
,,711,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.674816715Z,0,read_values,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000
,,711,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.729511016Z,0,read_values,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000
,,711,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.501504684Z,0,read_values,queryd_billing,queryd-v2-5797867574-5ff7l,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,712,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:01.16148791Z,714,response_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,03d01b74c8e09000
,,712,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.15254861Z,715,response_bytes,queryd_billing,queryd-v2-5797867574-b2bc5,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,713,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.481079117Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,714,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:06:00.455332774Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,714,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:42:00.541039897Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,714,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:29:00.477956027Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,714,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:33:00.515658208Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,714,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:11:00.667516371Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,715,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.580021946Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000
,,715,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.59833827Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,716,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:03:00.453652438Z,9401,total_duration_us,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000
,,716,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:58:00.467439389Z,9901,total_duration_us,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000
,,716,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:26:00.505504846Z,11429,total_duration_us,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000
,,716,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.863801527Z,10635,total_duration_us,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,717,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.516727014Z,0,read_values,queryd_billing,queryd-v2-5797867574-bd7fh,03c19003200d7000
,,717,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.776060648Z,0,read_values,queryd_billing,queryd-v2-5797867574-bd7fh,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,718,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.832140221Z,51283,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,718,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:38:00.693295746Z,7719,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,718,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:48:00.560832795Z,8350,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,718,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.577141351Z,8642,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,718,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.890962075Z,8223,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,718,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.182960005Z,9590,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,719,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.785640834Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,719,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.620654875Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,719,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:12:00.462658028Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,720,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:33:00.48797071Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,720,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:04:00.484123404Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000
,,720,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.729548006Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-wd7ww,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,721,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.168887758Z,0,read_values,queryd_billing,queryd-v2-5797867574-fmtgz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,722,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.894829972Z,57278,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,722,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.928682633Z,31571,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,722,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.898959022Z,12134,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,722,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.619773147Z,19154,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,722,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.783903603Z,14118,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,722,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.853962964Z,27925,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,722,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.785243966Z,15797,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,723,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.832140221Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,723,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:38:00.693295746Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,723,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:48:00.560832795Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,723,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.577141351Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,723,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.890962075Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,723,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.182960005Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,724,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.832140221Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,724,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:38:00.693295746Z,58633,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,724,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:48:00.560832795Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,724,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.577141351Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,724,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.890962075Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,724,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.182960005Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,725,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.730269127Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,725,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:04:00.506173446Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,725,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:08:00.618513396Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,725,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:13:00.133001433Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,725,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.921225459Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,725,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:48:00.633450477Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,726,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.681462039Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,726,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.955855408Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,726,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.748825278Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,727,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:01.012279784Z,270,read_values,queryd_billing,queryd-v1-5f699b6b58-2d862,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,728,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.481079117Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-t7slt,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,729,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:04.173174511Z,34183,read_values,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,729,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:04.02017492Z,34183,read_values,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,729,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.990040524Z,180,read_values,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,729,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:01.585834691Z,10,read_values,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,729,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.67997434Z,180,read_values,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,730,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:34:00.547848426Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,730,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:51:00.496150236Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,730,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:21:00.620889507Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,730,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:26:00.675475921Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000
,,730,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:43:00.567823817Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-zq4wb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,731,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:09:00.695676986Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000
,,731,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:16:00.423571342Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,732,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.774815315Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-5pjgb,0395bd7401aa3000
,,732,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.710588962Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-5pjgb,0395bd7401aa3000
,,732,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.518786657Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-5pjgb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,733,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:21:00.679948866Z,8255,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,733,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:46:00.57791668Z,7421,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,733,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:47:00.665339757Z,7872,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,733,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.885284853Z,7366,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,733,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:37:00.490611137Z,7656,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,734,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:03:00.453652438Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000
,,734,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:58:00.467439389Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000
,,734,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:26:00.505504846Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000
,,734,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.863801527Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,735,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.937640598Z,214514,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000
,,735,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.803909311Z,326743,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,736,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:23:00.473713274Z,0,read_values,queryd_billing,queryd-v2-5797867574-tlhkl,0395bd7401aa3000
,,736,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.647808347Z,0,read_values,queryd_billing,queryd-v2-5797867574-tlhkl,0395bd7401aa3000
,,736,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:43:00.561679905Z,0,read_values,queryd_billing,queryd-v2-5797867574-tlhkl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,737,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.70702057Z,8395,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000
,,737,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:04:00.522827935Z,11435,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,738,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.61520984Z,25932,total_duration_us,queryd_billing,queryd-v2-5797867574-c88sh,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,739,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:14:00.572403792Z,0,read_values,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,739,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.687631407Z,0,read_values,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,739,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:14:00.48545546Z,0,read_values,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,739,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:41:00.534038417Z,0,read_values,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,739,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:51:00.491763198Z,0,read_values,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,740,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:49:00.426081648Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-lj72r,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,741,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:52:00.663145742Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,741,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:31:00.655419185Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000
,,741,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:44:00.514580555Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,742,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:53:00.603537064Z,19038,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,742,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.709640978Z,21818,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,742,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:41:00.659356314Z,7700,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,742,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:47:00.524120738Z,7602,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,742,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:07:00.552515712Z,10813,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,742,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:38:00.141966771Z,10149,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,742,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.625087256Z,9057,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,743,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:12:00.462956123Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-j8hm4,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,744,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.785640834Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,744,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.620654875Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,744,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:12:00.462658028Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,745,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:06:00.455332774Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,745,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:42:00.541039897Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,745,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:29:00.477956027Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,745,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:33:00.515658208Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,745,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:11:00.667516371Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,746,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.492088454Z,0,read_values,queryd_billing,queryd-v2-5797867574-t474n,03c19003200d7000
,,746,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.714536617Z,0,read_values,queryd_billing,queryd-v2-5797867574-t474n,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,747,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:24:00.614708868Z,9957,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,747,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:27:00.607072565Z,10110,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,747,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:37:00.586640451Z,9011,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,747,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.627600735Z,9998,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000
,,747,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:46:00.574052194Z,16361,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,748,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:23:00.482973775Z,5,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,748,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:42:00.476029727Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,748,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:24:00.494923251Z,17,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000
,,748,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:25:00.800071949Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,749,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.917582859Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,750,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.730269127Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,750,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:04:00.506173446Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,750,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:08:00.618513396Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,750,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:13:00.133001433Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,750,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.921225459Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000
,,750,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:48:00.633450477Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,751,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.776577884Z,169,read_bytes,queryd_billing,queryd-v2-5797867574-tlhkl,03d01b74c8e09000
,,751,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:01.170233871Z,2880,read_bytes,queryd_billing,queryd-v2-5797867574-tlhkl,03d01b74c8e09000
,,751,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:01.25392002Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-tlhkl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,752,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.515673793Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000
,,752,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.590064004Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,753,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.67710836Z,0,read_values,queryd_billing,queryd-v2-5797867574-kvktv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,754,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:04.170703823Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,755,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:04.170703823Z,3507412,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,756,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.852953324Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-5pjgb,03c19003200d7000
,,756,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.655575144Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-5pjgb,03c19003200d7000
,,756,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.656976818Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-5pjgb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,757,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.681462039Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,757,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.955855408Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,757,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.748825278Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,758,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.685173538Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,758,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.693548351Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,758,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.623719778Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,758,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.755486749Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,759,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.821668012Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-9lsws,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,760,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.749802396Z,146032,total_duration_us,queryd_billing,queryd-v2-5797867574-rj2ns,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,761,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.937640598Z,214483,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000
,,761,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.803909311Z,326703,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,762,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.789866842Z,11179,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,762,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.635709149Z,11047,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,762,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:01.216435523Z,11752,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,763,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.848026232Z,350273,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,763,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.861699773Z,181192,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,763,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.989693911Z,271989,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,763,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:04.171157376Z,3494409,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,764,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:21:00.679948866Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,764,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:46:00.57791668Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,764,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:47:00.665339757Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,764,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.885284853Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,764,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:37:00.490611137Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,765,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.758278197Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,765,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.837604512Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,765,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.614967009Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,766,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:04.423740313Z,34183,read_values,queryd_billing,queryd-v2-5797867574-kssph,03d01b74c8e09000
,,766,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.715319869Z,180,read_values,queryd_billing,queryd-v2-5797867574-kssph,03d01b74c8e09000
,,766,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.898664906Z,10,read_values,queryd_billing,queryd-v2-5797867574-kssph,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,767,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.492227407Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,03c19003200d7000
,,767,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.632944642Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,768,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:14:00.572403792Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,768,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.687631407Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,768,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:14:00.48545546Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,768,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:41:00.534038417Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000
,,768,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:51:00.491763198Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-xvkns,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,769,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.863657529Z,20458,total_duration_us,queryd_billing,queryd-v2-5797867574-hh9fz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,770,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.6467672Z,14642,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-mfspl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,771,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.701333339Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,771,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.82139065Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,771,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.811423271Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,772,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.828458662Z,154525,total_duration_us,queryd_billing,queryd-v2-5797867574-t474n,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,773,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.783156493Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000
,,773,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.825524632Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,774,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:04.173174511Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,774,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:04.02017492Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,774,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.990040524Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,774,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:01.585834691Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,774,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.67997434Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,775,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.585156227Z,13657,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000
,,775,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.663976545Z,20175,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,776,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.832140221Z,51339,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,776,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:38:00.693295746Z,66379,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,776,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:48:00.560832795Z,8383,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,776,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.577141351Z,8671,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,776,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.890962075Z,8253,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,776,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.182960005Z,9621,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,777,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.924157523Z,11235,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,777,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.579798761Z,12751,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,777,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.763690024Z,8616,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,777,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:18:00.533780271Z,7975,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,777,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.596326558Z,40780,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000
,,777,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:42:00.44613233Z,7620,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-w4p96,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,778,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.813155264Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,778,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.697447893Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,778,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.780742525Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,779,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:03:00.453652438Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000
,,779,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:58:00.467439389Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000
,,779,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:26:00.505504846Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000
,,779,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.863801527Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-cm6fz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,780,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:04.170703823Z,138556,response_bytes,queryd_billing,queryd-v1-5f699b6b58-wbczl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,781,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.686930041Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,781,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:57:00.433078105Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000
,,781,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:17:00.415381738Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-cd4cc,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,782,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:26:00.437242604Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,782,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:31:00.477712461Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,782,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:21:00.4986011Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,782,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:52:00.719041693Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,782,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:54:00.613488751Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,782,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:59:00.564883689Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,782,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:03:00.545102773Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,782,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.856694666Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,783,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:01.012279784Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,784,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:16:00.504902703Z,0,read_values,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000
,,784,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:43:00.504989877Z,0,read_values,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000
,,784,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.562309784Z,0,read_values,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000
,,784,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.70656956Z,0,read_values,queryd_billing,queryd-v2-5797867574-fmtgz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,785,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.471264316Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,785,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.532354811Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,785,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.58982965Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,785,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.736000374Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000
,,785,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.62852928Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kssph,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,786,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:01.077941715Z,2160,read_bytes,queryd_billing,queryd-v1-5f699b6b58-t7slt,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,787,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.52560325Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-plnml,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,788,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:16:00.501326426Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-rj2ns,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,789,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.783156493Z,1260,response_bytes,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000
,,789,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.825524632Z,1258,response_bytes,queryd_billing,queryd-v1-5f699b6b58-wd7ww,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,790,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.675124012Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000
,,790,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.92728318Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,791,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.6467672Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-mfspl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,792,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.836044529Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000
,,792,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.425198714Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,793,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:26:00.437242604Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,793,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:31:00.477712461Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,793,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:21:00.4986011Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,793,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:52:00.719041693Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,793,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:54:00.613488751Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,793,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:59:00.564883689Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,793,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:03:00.545102773Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000
,,793,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.856694666Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-zlf4s,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,794,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.681462039Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,794,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.955855408Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,794,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.748825278Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,795,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.813155264Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,795,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.697447893Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,795,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.780742525Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,796,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.813155264Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,796,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:05:00.697447893Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000
,,796,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:55:00.780742525Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-qfpgc,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,797,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:01.077941715Z,213553,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,798,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.968941865Z,29809,total_duration_us,queryd_billing,queryd-v2-5797867574-kvktv,0395bd7401aa3000
,,798,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:13:00.694273957Z,8600,total_duration_us,queryd_billing,queryd-v2-5797867574-kvktv,0395bd7401aa3000
,,798,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:36:00.447798087Z,8046,total_duration_us,queryd_billing,queryd-v2-5797867574-kvktv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,799,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.894829972Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,799,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.928682633Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,799,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.898959022Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,799,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.619773147Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,799,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.783903603Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,799,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.853962964Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,799,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.785243966Z,5,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,800,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975813826Z,11,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zq4wb,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,801,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.644202789Z,1018,response_bytes,queryd_billing,queryd-v2-5797867574-qqx49,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,802,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.929248638Z,187820,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000
,,802,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:04.160876428Z,3484937,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,803,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.516727014Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-bd7fh,03c19003200d7000
,,803,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.776060648Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-bd7fh,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,804,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.716500242Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,804,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.675300682Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,804,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.722782443Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,804,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:46:00.61084851Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000
,,804,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:53:00.659149488Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-qqx49,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,805,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:06:00.455332774Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,805,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:42:00.541039897Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,805,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:29:00.477956027Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,805,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:33:00.515658208Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000
,,805,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:11:00.667516371Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-dc5cv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,806,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:06:00.513213184Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,806,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:22:00.448283291Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,806,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:28:00.484967147Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,806,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:56:00.684591295Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,807,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.783829632Z,170533,total_duration_us,queryd_billing,queryd-v2-5797867574-bd7fh,03d01b74c8e09000
,,807,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.757281711Z,144718,total_duration_us,queryd_billing,queryd-v2-5797867574-bd7fh,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,808,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:01.16148791Z,548348,total_duration_us,queryd_billing,queryd-v2-5797867574-b2bc5,03d01b74c8e09000
,,808,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.15254861Z,587370,total_duration_us,queryd_billing,queryd-v2-5797867574-b2bc5,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,809,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:01.282532978Z,2,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000
,,809,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.97843698Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-hmdwq,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,810,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:01.012279784Z,151251,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,811,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.857012416Z,12388,total_duration_us,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,811,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:28:00.450725793Z,8341,total_duration_us,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,811,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:32:00.590667734Z,64585,total_duration_us,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,811,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:39:00.577723384Z,7955,total_duration_us,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000
,,811,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:11:00.598135316Z,9037,total_duration_us,queryd_billing,queryd-v2-5797867574-plnml,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,812,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:06:00.513213184Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,812,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:22:00.448283291Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,812,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:28:00.484967147Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,812,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:56:00.684591295Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,813,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:02.629173563Z,10,read_values,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,813,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:04.58472994Z,34183,read_values,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000
,,813,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:01.025444871Z,270,read_values,queryd_billing,queryd-v1-5f699b6b58-4drxz,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,814,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:06:00.455340709Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,814,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:09:00.434414481Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,814,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.653985084Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000
,,814,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.767478932Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-4drxz,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,815,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.832140221Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,815,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:38:00.693295746Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,815,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:48:00.560832795Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,815,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.577141351Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,815,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.890962075Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000
,,815,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.182960005Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-6c7j6,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,816,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.760513754Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,816,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:24:00.625390315Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000
,,816,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:57:00.617251549Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-2d862,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,817,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:06:00.513213184Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,817,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:22:00.448283291Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,817,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:28:00.484967147Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000
,,817,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:56:00.684591295Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-vh94j,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,818,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.70702057Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000
,,818,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:04:00.522827935Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,819,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:11:00.505877871Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,819,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:27:00.574693886Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,819,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:59:00.572427992Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,819,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:07:00.569599945Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,819,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:22:00.588925323Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,819,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.50045533Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000
,,819,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:53:00.457936128Z,2,response_bytes,queryd_billing,queryd-v2-5797867574-kssph,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,820,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.685173538Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,820,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.693548351Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,820,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:00.623719778Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000
,,820,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.755486749Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-6c7j6,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,821,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.701333339Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,821,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.82139065Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,821,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.811423271Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,822,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:04.41909794Z,139553,response_bytes,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000
,,822,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:01.610831987Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,823,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.789866842Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,823,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.635709149Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000
,,823,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:01.216435523Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,824,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:14:00.606703032Z,8523,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,824,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.658228927Z,8836,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,824,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.675835865Z,20314,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,824,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:02:00.441622158Z,7739,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,824,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:33:00.490131246Z,11280,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,824,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:47:00.493465617Z,11663,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,824,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:58:00.69307463Z,20785,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,825,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:29:00.521460892Z,9330,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,825,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:34:00.474156777Z,7856,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,825,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.914502559Z,7875,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,825,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:39:00.487057271Z,7663,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,826,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:25:00.6467672Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-mfspl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,827,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:04.173174511Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,827,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:04.02017492Z,14,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,827,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:00.990040524Z,130841,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,827,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:20:01.585834691Z,15,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000
,,827,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.67997434Z,13,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-8rtxf,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,828,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.758278197Z,25832,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,828,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:05:00.837604512Z,16211,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000
,,828,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:35:00.614967009Z,11839,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,829,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:49:00.426081648Z,7901,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,830,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:15:00.867420941Z,13565,total_duration_us,queryd_billing,queryd-v2-5797867574-s6t85,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,831,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:48:00.675272795Z,2,response_bytes,queryd_billing,queryd-v1-5f699b6b58-66kcw,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,832,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:04.170703823Z,17,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,833,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.777690142Z,1440,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,833,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:03.873228519Z,273464,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,833,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.975537336Z,2160,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000
,,833,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.976567453Z,2160,read_bytes,queryd_billing,queryd-v1-5f699b6b58-zgpgs,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,834,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:04.423740313Z,3678954,total_duration_us,queryd_billing,queryd-v2-5797867574-kssph,03d01b74c8e09000
,,834,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:45:00.715319869Z,140181,total_duration_us,queryd_billing,queryd-v2-5797867574-kssph,03d01b74c8e09000
,,834,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.898664906Z,318624,total_duration_us,queryd_billing,queryd-v2-5797867574-kssph,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,835,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:49:00.426081648Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-lj72r,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,836,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:01.012279784Z,151196,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-2d862,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,837,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:00:04.41909794Z,273464,read_bytes,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000
,,837,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:01.610831987Z,169,read_bytes,queryd_billing,queryd-v1-5f699b6b58-lj72r,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,838,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.836044529Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000
,,838,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:25:00.425198714Z,0,read_bytes,queryd_billing,queryd-v1-5f699b6b58-d5wwj,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,839,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:10:00.67710836Z,0,read_bytes,queryd_billing,queryd-v2-5797867574-kvktv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,840,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.681462039Z,17,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,840,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:20:00.955855408Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000
,,840,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.748825278Z,12,queue_duration_us,queryd_billing,queryd-v1-5f699b6b58-4drxz,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,841,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:21:00.679948866Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,841,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:46:00.57791668Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,841,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:47:00.665339757Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,841,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.885284853Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000
,,841,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:37:00.490611137Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-d5wwj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,842,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.70702057Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000
,,842,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:04:00.522827935Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-l8pjj,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,843,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:20:00.785640834Z,7178,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,843,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.620654875Z,14345,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000
,,843,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:12:00.462658028Z,8406,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,844,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:35:00.481079117Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-t7slt,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,845,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.508804564Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000
,,845,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:40:00.720702585Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-467pb,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,846,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:29:00.521460892Z,9290,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,846,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:34:00.474156777Z,7824,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,846,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:50:00.914502559Z,7839,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000
,,846,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:39:00.487057271Z,7629,execute_duration_us,queryd_billing,queryd-v1-5f699b6b58-wbczl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,847,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:01.012279784Z,1518,response_bytes,queryd_billing,queryd-v1-5f699b6b58-2d862,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,848,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:37:00.423986232Z,0,read_values,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,848,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.759512962Z,0,read_values,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,848,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:28:00.377646402Z,0,read_values,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,848,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:44:00.420950673Z,0,read_values,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,848,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:32:00.605366491Z,0,read_values,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000
,,848,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:49:00.463047225Z,0,read_values,queryd_billing,queryd-v2-5797867574-hzlgm,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,849,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:14:00.606703032Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,849,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:10:00.658228927Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,849,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:30:00.675835865Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,849,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:02:00.441622158Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,849,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:33:00.490131246Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,849,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:47:00.493465617Z,4,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000
,,849,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:58:00.69307463Z,3,compile_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,850,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.714848904Z,180,read_values,queryd_billing,queryd-v2-5797867574-c88sh,03d01b74c8e09000
,,850,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:30:00.67290226Z,180,read_values,queryd_billing,queryd-v2-5797867574-c88sh,03d01b74c8e09000
,,850,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:04.058154233Z,34183,read_values,queryd_billing,queryd-v2-5797867574-c88sh,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,851,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:53:00.603537064Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,851,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.709640978Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,851,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:41:00.659356314Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,851,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:47:00.524120738Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,851,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:07:00.552515712Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,851,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:38:00.141966771Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,851,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.625087256Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,852,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.848026232Z,2880,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,852,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.861699773Z,1440,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,852,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:00.989693911Z,1440,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000
,,852,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:04.171157376Z,273464,read_bytes,queryd_billing,queryd-v1-5f699b6b58-dbnmw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,853,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:30:04.437709166Z,3871157,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-66kcw,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,854,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:55:00.749034925Z,0,read_values,queryd_billing,queryd-v2-5797867574-zmcl2,03c19003200d7000
,,854,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.721679848Z,0,read_values,queryd_billing,queryd-v2-5797867574-zmcl2,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,855,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:00:00.666298715Z,714,response_bytes,queryd_billing,queryd-v2-5797867574-plnml,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,856,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.894829972Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,856,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:00:00.928682633Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,856,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:10:00.898959022Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,856,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:35:00.619773147Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,856,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:55:00.783903603Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,856,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:45:00.853962964Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000
,,856,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:50:00.785243966Z,0,requeue_duration_us,queryd_billing,queryd-v1-5f699b6b58-ltbql,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,857,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:05:00.832579331Z,22053,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,857,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:40:00.727379572Z,21630,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000
,,857,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.812383085Z,40459,total_duration_us,queryd_billing,queryd-v1-5f699b6b58-dc5cv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,858,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:15:00.701333339Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,858,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:40:00.82139065Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000
,,858,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:45:00.811423271Z,0,read_values,queryd_billing,queryd-v1-5f699b6b58-vh94j,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,859,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T10:53:00.603537064Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,859,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:15:00.709640978Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,859,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:41:00.659356314Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,859,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T11:47:00.524120738Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,859,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:07:00.552515712Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,859,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:38:00.141966771Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000
,,859,2019-12-03T10:00:00Z,2019-12-03T13:00:00Z,2019-12-03T12:50:00.625087256Z,0,plan_duration_us,queryd_billing,queryd-v1-5f699b6b58-zgpgs,0395bd7401aa3000


"

outData = "
#group,false,false,true,true,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,long,dateTime:RFC3339
#default,duration_us,,,,,
,result,table,_start,_stop,duration_us,_time
,,0,2019-12-03T10:00:00Z,2019-12-03T12:00:00Z,12507024,2019-12-03T11:00:00Z
,,0,2019-12-03T10:00:00Z,2019-12-03T12:00:00Z,16069640,2019-12-03T12:00:00Z
"

_f = (table=<-) => table
    |> range(start: 2019-12-03T10:00:00Z, stop: 2019-12-03T12:00:00Z)
    |> filter(fn: (r) =>
        r.org_id == "03d01b74c8e09000"
        and r._measurement == "queryd_billing"
        and (r._field == "compile_duration_us" or r._field == "plan_duration_us" or r._field == "execute_duration_us")
    )
    |> group()
    |> aggregateWindow(every: 1h, fn: sum)
    |> fill(column: "_value", value: 0)
    |> rename(columns: {_value: "duration_us"})
    |> yield(name: "duration_us")

test query_duration = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: _f})
