package usage_test

import "testing"

// This dataset has been generated with this query:
// from(bucket: "system_usage")
//     |> range(start: 2019-08-01T12:00:00Z, stop: 2019-08-01T14:00:00Z)
//     |> filter(fn: (r) =>
//         (r.org_id == "03d01b74c8e09000" or r.org_id == "03c19003200d7000" or r.org_id == "0395bd7401aa3000")
//         and r._measurement == "queryd_billing"
//     )
inData = "
#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,0,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:01.034938691Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000
,,0,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:41:00.525870701Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000
,,0,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.950355604Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000
,,0,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:26:00.503203471Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,1,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:46:00.486323917Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,1,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:54:00.481370575Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,1,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:13:00.607707752Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,1,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:17:00.488540074Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,1,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:23:00.483726433Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,1,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:26:00.499075289Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,1,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:29:00.477223164Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,1,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:31:00.474147376Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,1,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:46:00.525979289Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,1,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:48:00.601566851Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,1,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:49:00.491419111Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,1,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.497453495Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,1,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:58:00.488056387Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,2,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:01.880930931Z,840,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03d01b74c8e09000
,,2,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.977436912Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03d01b74c8e09000
,,2,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:01.179598985Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,3,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.651898933Z,46961,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000
,,3,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:03:00.477061908Z,27780,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000
,,3,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.490292928Z,29545,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000
,,3,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.547483082Z,23813,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,4,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:01:00.570149909Z,59993468,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,4,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.867434006Z,397361,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,4,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:01.272127486Z,632614,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,4,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.970514197Z,431528,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,4,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:01.078136693Z,567648,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,4,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.572721595Z,1108515,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,4,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.973699773Z,366409,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,4,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:01.870137074Z,1218519,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,4,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:46:00.520840844Z,59995218,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,5,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.988611166Z,514081,total_duration_us,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,5,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.976033727Z,407542,total_duration_us,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,5,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.875974229Z,376080,total_duration_us,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,5,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:56:00.540440239Z,59993983,total_duration_us,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,5,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:06:00.488297265Z,59995256,total_duration_us,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,6,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.588746873Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,6,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.616685301Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,6,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.701174358Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,6,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.685121539Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,6,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.529396168Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,6,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.549283022Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,6,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.511155262Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,7,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.533096623Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,7,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.657260033Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,7,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.523745277Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,7,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.521292944Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,7,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.508784116Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,7,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.519831768Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,7,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.684879Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,7,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.720897719Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,8,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.954866323Z,464,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,9,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.690868149Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,03c19003200d7000
,,9,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.704345198Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,10,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.95995624Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000
,,10,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.962184778Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000
,,10,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.958940279Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000
,,10,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.870540202Z,464,read_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,11,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:02.095923069Z,60,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,11,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:01.450962024Z,90,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,11,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:26:00.554024549Z,7451,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,11,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:01.096838578Z,90,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,11,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:01.694674526Z,102,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,11,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:01:00.545725631Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,11,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:01.200032709Z,90,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,11,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:01.198296884Z,90,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,11,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:01.697935952Z,105,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,11,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:01.097195574Z,90,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,11,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:01.29974034Z,90,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,11,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:01.098950778Z,90,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,11,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:01.418781835Z,87,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,12,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:11:00.478414062Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,12,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:18:00.488037379Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,12,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:24:00.493400923Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,12,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.557061177Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,12,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:49:00.483622115Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,12,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:53:00.48923721Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,12,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:59:00.486840652Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,12,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.531132434Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,12,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:21:00.491145961Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,12,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:27:00.487253317Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,12,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:37:00.491349011Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,12,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.621251682Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,13,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:46:00.486323917Z,23060,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,13,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:54:00.481370575Z,26931,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,13,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:13:00.607707752Z,154617,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,13,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:17:00.488540074Z,21512,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,13,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:23:00.483726433Z,23151,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,13,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:26:00.499075289Z,35627,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,13,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:29:00.477223164Z,23999,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,13,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:31:00.474147376Z,23965,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,13,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:46:00.525979289Z,61358,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,13,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:48:00.601566851Z,142701,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,13,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:49:00.491419111Z,25007,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,13,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.497453495Z,25336,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,13,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:58:00.488056387Z,24574,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,14,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.501845561Z,40373,total_duration_us,queryd_billing,queryd-a-567d7b8464-nkr8z,03c19003200d7000
,,14,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.513513395Z,49685,total_duration_us,queryd_billing,queryd-a-567d7b8464-nkr8z,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,15,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.533096623Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,15,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.657260033Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,15,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.523745277Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,15,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.521292944Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,15,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.508784116Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,15,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.519831768Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,15,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.684879Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,15,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.720897719Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,16,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:01:00.516347576Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,0395bd7401aa3000
,,16,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:04:00.483414339Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,0395bd7401aa3000
,,16,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:16:00.491232711Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,17,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.296146195Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,17,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.523493998Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,17,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.524743262Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,17,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.53008227Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,17,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:52:00.479092929Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,18,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:01.292879825Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,18,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.398377049Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,18,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.893352315Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,18,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.894114015Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,18,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:31:00.496776599Z,62400,read_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,19,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.498459129Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,03c19003200d7000
,,19,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.702457381Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,20,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:01.283537001Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,20,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.887420221Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,20,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:01.685343496Z,632,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,20,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.888061441Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,20,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:56:00.478723104Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,21,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:03.91066756Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,21,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:21:00.513655817Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,21,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:01.114372345Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,21,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:31:00.46674743Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,21,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.813728039Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,21,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:01.012124567Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,21,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:51:00.511063137Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,21,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:01.713330923Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,21,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:36:00.479939528Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,21,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:01.310517133Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,22,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.647640246Z,144073,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,22,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.673152688Z,141689,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,22,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.503074424Z,41794,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,22,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.883388065Z,270053,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,22,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:01.766565584Z,1239294,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,22,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.784667596Z,160419,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,22,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.485214725Z,31292,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,22,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.496265331Z,41546,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,23,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:02:00.479534485Z,0,read_values,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,23,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:04:00.470594397Z,0,read_values,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,23,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:12:00.476923814Z,0,read_values,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,23,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.695016629Z,0,read_values,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,23,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.632624131Z,0,read_values,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,23,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:32:00.482130432Z,0,read_values,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,24,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.576894936Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03c19003200d7000
,,24,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.513387882Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03c19003200d7000
,,24,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.672684244Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,25,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:02.095923069Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,25,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:01.450962024Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,25,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:26:00.554024549Z,59608,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,25,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:01.096838578Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,25,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:01.694674526Z,816,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,25,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:01:00.545725631Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,25,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:01.200032709Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,25,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:01.198296884Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,25,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:01.697935952Z,840,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,25,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:01.097195574Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,25,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:01.29974034Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,25,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:01.098950778Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,25,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:01.418781835Z,696,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,26,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:01.292879825Z,750084,total_duration_us,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,26,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.398377049Z,853524,total_duration_us,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,26,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.893352315Z,431990,total_duration_us,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,26,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.894114015Z,430767,total_duration_us,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,26,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:31:00.496776599Z,59994185,total_duration_us,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,27,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:01:00.484147634Z,25560,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,27,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:07:00.480924824Z,23716,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,27,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:27:00.490781811Z,29321,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,27,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:29:00.473468521Z,24092,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,27,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.607478294Z,75188,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,27,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.620991461Z,22280,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,27,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:51:00.479647924Z,21761,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,27,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:56:00.588645189Z,140323,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,27,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:03:00.487133158Z,20973,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,27,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:06:00.484321671Z,24501,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,27,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:33:00.482739897Z,21389,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,27,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.516558315Z,27926,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,27,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.517181373Z,28245,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,27,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:44:00.479659339Z,25773,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,28,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:08:00.627574839Z,171018,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,28,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:14:00.49067527Z,20779,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,28,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:17:00.487130291Z,21888,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,28,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:38:00.479543081Z,21535,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,28,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:18:00.47614158Z,21897,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,29,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.562048755Z,23764,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,29,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:26:00.556708235Z,100060,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,29,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:34:00.487483035Z,21705,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,29,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:41:00.67594936Z,208953,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,29,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:44:00.499102169Z,37811,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,29,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.56992199Z,38063,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,29,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:57:00.488454706Z,23394,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,29,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:58:00.504724054Z,33464,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,29,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:02:00.482659525Z,22009,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,29,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:08:00.478316256Z,23595,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,29,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:12:00.492881833Z,26502,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,29,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:14:00.491639042Z,24259,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,29,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:22:00.482878191Z,25873,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,29,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:24:00.4878516Z,22082,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,30,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.988611166Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,30,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.976033727Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,30,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.875974229Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,30,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:56:00.540440239Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,30,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:06:00.488297265Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,31,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.968529532Z,90,read_values,queryd_billing,queryd-a-567d7b8464-wm622,03d01b74c8e09000
,,31,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:01.171849185Z,90,read_values,queryd_billing,queryd-a-567d7b8464-wm622,03d01b74c8e09000
,,31,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:01.484631063Z,90,read_values,queryd_billing,queryd-a-567d7b8464-wm622,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,32,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.29244898Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,32,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.511146754Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,32,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.647563651Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,32,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.51837934Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,32,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.513211046Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,33,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:31:00.478958223Z,26108,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,33,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.501944968Z,36181,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,33,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:36:00.488893386Z,25383,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,33,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:42:00.489699624Z,28196,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,33,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:47:00.483799335Z,23553,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,33,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.585722242Z,48795,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,33,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.558041208Z,31334,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,33,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.656737151Z,39660,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,34,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:01.039264398Z,60,read_values,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000
,,34,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:06:00.470512067Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000
,,34,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:01.05279611Z,60,read_values,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000
,,34,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:36:00.456164223Z,7459,read_values,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,35,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.636022941Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000
,,35,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.517726347Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000
,,35,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.504779428Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000
,,35,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.689278921Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,36,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.533173951Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,36,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.533579345Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,36,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:36:00.475702635Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,36,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:39:00.490094464Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,36,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:41:00.49089932Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,36,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:42:00.483353951Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,36,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:47:00.506875644Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,37,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.533096623Z,39606,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,37,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.657260033Z,51450,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,37,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.523745277Z,34774,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,37,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.521292944Z,41630,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,37,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.508784116Z,52142,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,37,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.519831768Z,46471,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,37,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.684879Z,32051,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,37,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.720897719Z,30012,total_duration_us,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,38,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:01.039264398Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000
,,38,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:06:00.470512067Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000
,,38,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:01.05279611Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000
,,38,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:36:00.456164223Z,59672,read_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,39,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:46:00.486323917Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,39,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:54:00.481370575Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,39,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:13:00.607707752Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,39,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:17:00.488540074Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,39,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:23:00.483726433Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,39,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:26:00.499075289Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,39,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:29:00.477223164Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,39,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:31:00.474147376Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,39,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:46:00.525979289Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,39,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:48:00.601566851Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,39,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:49:00.491419111Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,39,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.497453495Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,39,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:58:00.488056387Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,40,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.498459129Z,0,read_values,queryd_billing,queryd-a-567d7b8464-wxzxp,03c19003200d7000
,,40,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.702457381Z,0,read_values,queryd_billing,queryd-a-567d7b8464-wxzxp,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,41,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:01.086269488Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,41,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.888541465Z,464,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,41,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:02.882870556Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,41,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.988118366Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,41,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:03.78792612Z,928,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,41,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.98470962Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,41,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.988438498Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,41,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:11:00.496709641Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,41,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:16:00.524449747Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,41,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.887275324Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,42,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.542910288Z,0,read_values,queryd_billing,queryd-a-567d7b8464-8hmpp,03c19003200d7000
,,42,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.001227019Z,0,read_values,queryd_billing,queryd-a-567d7b8464-8hmpp,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,43,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:01.217638223Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-zpgpf,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,44,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:02:00.479534485Z,24931,total_duration_us,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,44,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:04:00.470594397Z,22903,total_duration_us,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,44,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:12:00.476923814Z,27532,total_duration_us,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,44,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.695016629Z,46163,total_duration_us,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,44,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.632624131Z,85421,total_duration_us,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,44,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:32:00.482130432Z,30867,total_duration_us,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,45,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.588746873Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,45,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.616685301Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,45,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.701174358Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,45,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.685121539Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,45,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.529396168Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,45,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.549283022Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,45,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.511155262Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,46,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.296146195Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,46,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.523493998Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,46,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.524743262Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,46,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.53008227Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,46,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:52:00.479092929Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,47,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:08:00.627574839Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,47,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:14:00.49067527Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,47,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:17:00.487130291Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,47,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:38:00.479543081Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,47,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:18:00.47614158Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,48,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:01.283537001Z,60,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,48,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.887420221Z,60,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,48,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:01.685343496Z,79,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,48,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.888061441Z,60,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,48,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:56:00.478723104Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,49,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:03.91066756Z,3408055,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,49,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:21:00.513655817Z,59995305,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,49,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:01.114372345Z,618328,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,49,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:31:00.46674743Z,59993448,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,49,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.813728039Z,361976,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,49,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:01.012124567Z,505638,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,49,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:51:00.511063137Z,59992818,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,49,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:01.713330923Z,1251029,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,49,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:36:00.479939528Z,59994222,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,49,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:01.310517133Z,774601,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,50,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:11:00.478414062Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,50,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:18:00.488037379Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,50,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:24:00.493400923Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,50,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.557061177Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,50,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:49:00.483622115Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,50,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:53:00.48923721Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,50,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:59:00.486840652Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,50,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.531132434Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,50,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:21:00.491145961Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,50,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:27:00.487253317Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,50,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:37:00.491349011Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,50,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.621251682Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.697901808Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.500304872Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:09:00.475715338Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:13:00.4883411Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:16:00.471327771Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:19:00.472909982Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:21:00.473764932Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:22:00.483040753Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:23:00.718566244Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:28:00.491388187Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.475034596Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:43:00.477120009Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:48:00.537010408Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:52:00.470429139Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.530864823Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.525931238Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:34:00.481362068Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,51,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:57:00.492651355Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,52,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.616067917Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,52,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.553305038Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,52,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.596901203Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,52,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.598127073Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,52,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.622914101Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,52,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.512773568Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,52,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.515196338Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,53,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.501845561Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,03c19003200d7000
,,53,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.513513395Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,54,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.690868149Z,92455,total_duration_us,queryd_billing,queryd-a-567d7b8464-mr4s8,03c19003200d7000
,,54,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.704345198Z,90191,total_duration_us,queryd_billing,queryd-a-567d7b8464-mr4s8,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,55,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:01.039264398Z,488222,total_duration_us,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000
,,55,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:06:00.470512067Z,59995594,total_duration_us,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000
,,55,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:01.05279611Z,476219,total_duration_us,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000
,,55,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:36:00.456164223Z,59993535,total_duration_us,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,56,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.542910288Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,03c19003200d7000
,,56,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.001227019Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,57,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:01.682352598Z,1031699,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,03c19003200d7000
,,57,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.650243169Z,53191,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,58,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:01:00.484147634Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,58,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:07:00.480924824Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,58,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:27:00.490781811Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,58,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:29:00.473468521Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,58,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.607478294Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,58,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.620991461Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,58,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:51:00.479647924Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,58,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:56:00.588645189Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,58,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:03:00.487133158Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,58,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:06:00.484321671Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,58,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:33:00.482739897Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,58,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.516558315Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,58,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.517181373Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,58,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:44:00.479659339Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,59,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:01.880930931Z,105,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,03d01b74c8e09000
,,59,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.977436912Z,60,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,03d01b74c8e09000
,,59,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:01.179598985Z,90,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,60,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:11:00.492570159Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000
,,60,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:16:00.650605498Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000
,,60,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:01.070125773Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000
,,60,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:21:00.496795386Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,61,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:46:00.590968998Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,61,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.930211883Z,59,read_values,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,61,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:41:00.546438039Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,61,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:01.034314226Z,60,read_values,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,61,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.82933305Z,60,read_values,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,62,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:31:00.478958223Z,0,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,62,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.501944968Z,0,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,62,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:36:00.488893386Z,0,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,62,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:42:00.489699624Z,0,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,62,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:47:00.483799335Z,0,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,62,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.585722242Z,0,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,62,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.558041208Z,0,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,62,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.656737151Z,0,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,63,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.542910288Z,45321,total_duration_us,queryd_billing,queryd-a-567d7b8464-8hmpp,03c19003200d7000
,,63,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.001227019Z,365683,total_duration_us,queryd_billing,queryd-a-567d7b8464-8hmpp,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,64,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:11:00.492570159Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000
,,64,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:16:00.650605498Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000
,,64,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:01.070125773Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000
,,64,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:21:00.496795386Z,59264,read_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,65,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:01.167845429Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2sbq5,03d01b74c8e09000
,,65,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:01.264460977Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2sbq5,03d01b74c8e09000
,,65,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.885957683Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2sbq5,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,66,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:46:00.486323917Z,0,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,66,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:54:00.481370575Z,0,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,66,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:13:00.607707752Z,0,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,66,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:17:00.488540074Z,0,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,66,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:23:00.483726433Z,0,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,66,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:26:00.499075289Z,0,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,66,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:29:00.477223164Z,0,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,66,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:31:00.474147376Z,0,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,66,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:46:00.525979289Z,0,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,66,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:48:00.601566851Z,0,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,66,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:49:00.491419111Z,0,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,66,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.497453495Z,0,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000
,,66,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:58:00.488056387Z,0,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,67,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:01.140329663Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,67,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.937088519Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,67,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.837769465Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,67,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.938965657Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,67,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:03.735264596Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,67,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.959621221Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,67,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:51:00.579710518Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,68,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.988611166Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,68,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.976033727Z,472,read_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,68,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.875974229Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,68,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:56:00.540440239Z,61920,read_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,68,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:06:00.488297265Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,69,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:01.034938691Z,394383,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000
,,69,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:41:00.525870701Z,59994575,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000
,,69,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.950355604Z,427451,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000
,,69,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:26:00.503203471Z,59995176,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,70,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.29244898Z,781129,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,70,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.511146754Z,31950,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,70,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.647563651Z,45831,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,70,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.51837934Z,31406,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,70,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.513211046Z,39264,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,71,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.501845561Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,03c19003200d7000
,,71,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.513513395Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,72,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:01.65498559Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-zpgpf,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,73,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:01.140329663Z,704,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,73,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.937088519Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,73,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.837769465Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,73,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.938965657Z,464,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,73,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:03.735264596Z,1040,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,73,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.959621221Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,73,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:51:00.579710518Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,74,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.522923171Z,21715,total_duration_us,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000
,,74,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:07:00.49308561Z,27149,total_duration_us,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000
,,74,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:19:00.484455371Z,21030,total_duration_us,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000
,,74,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:28:00.488167633Z,21735,total_duration_us,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,75,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:01.65498559Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-zpgpf,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,76,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:01.034938691Z,60,read_values,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000
,,76,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:41:00.525870701Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000
,,76,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.950355604Z,60,read_values,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000
,,76,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:26:00.503203471Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,77,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:01.086269488Z,584819,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,77,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.888541465Z,419385,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,77,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:02.882870556Z,2329207,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,77,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.988118366Z,468365,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,77,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:03.78792612Z,3250067,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,77,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.98470962Z,501423,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,77,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.988438498Z,458628,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,77,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:11:00.496709641Z,59994974,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,77,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:16:00.524449747Z,59995037,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,77,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.887275324Z,427337,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,78,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:01.167845429Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-2sbq5,03d01b74c8e09000
,,78,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:01.264460977Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-2sbq5,03d01b74c8e09000
,,78,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.885957683Z,472,read_bytes,queryd_billing,queryd-a-567d7b8464-2sbq5,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,79,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:06:00.492582743Z,0,read_values,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,79,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.497626928Z,0,read_values,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,79,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.497621429Z,0,read_values,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,79,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:33:00.484687388Z,0,read_values,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,79,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:37:00.478695875Z,0,read_values,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,79,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:39:00.481188183Z,0,read_values,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,79,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.532824377Z,0,read_values,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,79,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.303708334Z,0,read_values,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,80,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.562048755Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,80,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:26:00.556708235Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,80,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:34:00.487483035Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,80,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:41:00.67594936Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,80,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:44:00.499102169Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,80,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.56992199Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,80,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:57:00.488454706Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,80,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:58:00.504724054Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,80,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:02:00.482659525Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,80,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:08:00.478316256Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,80,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:12:00.492881833Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,80,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:14:00.491639042Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,80,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:22:00.482878191Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,80,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:24:00.4878516Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,81,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:01.167845429Z,603201,total_duration_us,queryd_billing,queryd-a-567d7b8464-2sbq5,03d01b74c8e09000
,,81,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:01.264460977Z,645041,total_duration_us,queryd_billing,queryd-a-567d7b8464-2sbq5,03d01b74c8e09000
,,81,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.885957683Z,421589,total_duration_us,queryd_billing,queryd-a-567d7b8464-2sbq5,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,82,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:01.167845429Z,90,read_values,queryd_billing,queryd-a-567d7b8464-2sbq5,03d01b74c8e09000
,,82,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:01.264460977Z,90,read_values,queryd_billing,queryd-a-567d7b8464-2sbq5,03d01b74c8e09000
,,82,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.885957683Z,59,read_values,queryd_billing,queryd-a-567d7b8464-2sbq5,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,83,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:11:00.478414062Z,21881,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,83,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:18:00.488037379Z,22319,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,83,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:24:00.493400923Z,22810,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,83,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.557061177Z,28062,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,83,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:49:00.483622115Z,22174,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,83,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:53:00.48923721Z,23815,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,83,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:59:00.486840652Z,23198,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,83,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.531132434Z,23960,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,83,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:21:00.491145961Z,26512,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,83,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:27:00.487253317Z,22819,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,83,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:37:00.491349011Z,22097,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,83,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.621251682Z,23616,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,84,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.562048755Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,84,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:26:00.556708235Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,84,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:34:00.487483035Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,84,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:41:00.67594936Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,84,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:44:00.499102169Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,84,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.56992199Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,84,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:57:00.488454706Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,84,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:58:00.504724054Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,84,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:02:00.482659525Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,84,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:08:00.478316256Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,84,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:12:00.492881833Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,84,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:14:00.491639042Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,84,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:22:00.482878191Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,84,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:24:00.4878516Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,85,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.636022941Z,144816,total_duration_us,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000
,,85,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.517726347Z,42722,total_duration_us,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000
,,85,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.504779428Z,40384,total_duration_us,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000
,,85,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.689278921Z,46353,total_duration_us,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,86,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.647838288Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,86,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.608466293Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,86,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.528214136Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,86,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.543848534Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,86,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.521564682Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,86,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.525124227Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,86,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.613172125Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,87,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.647640246Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,87,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.673152688Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,87,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.503074424Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,87,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.883388065Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,87,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:01.766565584Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,87,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.784667596Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,87,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.485214725Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,87,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.496265331Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.697901808Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.500304872Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:09:00.475715338Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:13:00.4883411Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:16:00.471327771Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:19:00.472909982Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:21:00.473764932Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:22:00.483040753Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:23:00.718566244Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:28:00.491388187Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.475034596Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:43:00.477120009Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:48:00.537010408Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:52:00.470429139Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.530864823Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.525931238Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:34:00.481362068Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,88,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:57:00.492651355Z,0,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,89,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:56:00.54287719Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-zpgpf,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,90,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.651898933Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000
,,90,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:03:00.477061908Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000
,,90,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.490292928Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000
,,90,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.547483082Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,91,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:01:00.516347576Z,35858,total_duration_us,queryd_billing,queryd-a-567d7b8464-nkr8z,0395bd7401aa3000
,,91,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:04:00.483414339Z,27456,total_duration_us,queryd_billing,queryd-a-567d7b8464-nkr8z,0395bd7401aa3000
,,91,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:16:00.491232711Z,24531,total_duration_us,queryd_billing,queryd-a-567d7b8464-nkr8z,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,92,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.954866323Z,58,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,93,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.651898933Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000
,,93,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:03:00.477061908Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000
,,93,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.490292928Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000
,,93,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.547483082Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,94,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:01.039264398Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000
,,94,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:06:00.470512067Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000
,,94,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:01.05279611Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000
,,94,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:36:00.456164223Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,95,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:01.034938691Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000
,,95,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:41:00.525870701Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000
,,95,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.950355604Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000
,,95,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:26:00.503203471Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,96,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:01.292879825Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,96,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.398377049Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,96,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.893352315Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,96,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.894114015Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,96,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:31:00.496776599Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,97,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.498459129Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,03c19003200d7000
,,97,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.702457381Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,98,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.647838288Z,129095,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,98,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.608466293Z,83090,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,98,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.528214136Z,41782,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,98,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.543848534Z,37021,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,98,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.521564682Z,30355,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,98,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.525124227Z,33811,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,98,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.613172125Z,68858,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,99,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:01:00.570149909Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,99,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.867434006Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,99,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:01.272127486Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,99,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.970514197Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,99,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:01.078136693Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,99,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.572721595Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,99,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.973699773Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,99,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:01.870137074Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,99,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:46:00.520840844Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,100,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.95995624Z,60,read_values,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000
,,100,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.962184778Z,60,read_values,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000
,,100,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.958940279Z,60,read_values,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000
,,100,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.870540202Z,58,read_values,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,101,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:02.095923069Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,101,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:01.450962024Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,101,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:26:00.554024549Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,101,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:01.096838578Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,101,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:01.694674526Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,101,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:01:00.545725631Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,101,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:01.200032709Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,101,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:01.198296884Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,101,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:01.697935952Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,101,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:01.097195574Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,101,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:01.29974034Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,101,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:01.098950778Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,101,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:01.418781835Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,102,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.635201896Z,44435,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,102,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.527170667Z,33115,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,102,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.495654052Z,30931,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,102,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.242752012Z,754406,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,102,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.532006543Z,34172,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,102,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.515414938Z,44005,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,102,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.74769122Z,38870,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,102,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.60047454Z,43362,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,102,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.563200167Z,43153,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,103,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.522092597Z,43622,total_duration_us,queryd_billing,queryd-a-567d7b8464-rlmq5,03c19003200d7000
,,103,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.593976825Z,65016,total_duration_us,queryd_billing,queryd-a-567d7b8464-rlmq5,03c19003200d7000
,,103,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.630789006Z,55658,total_duration_us,queryd_billing,queryd-a-567d7b8464-rlmq5,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,104,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:02:00.479534485Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,104,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:04:00.470594397Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,104,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:12:00.476923814Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,104,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.695016629Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,104,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.632624131Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,104,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:32:00.482130432Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,105,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.296146195Z,736208,total_duration_us,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,105,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.523493998Z,23799,total_duration_us,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,105,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.524743262Z,24000,total_duration_us,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,105,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.53008227Z,22275,total_duration_us,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,105,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:52:00.479092929Z,22159,total_duration_us,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,106,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.29244898Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,106,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.511146754Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,106,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.647563651Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,106,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.51837934Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,106,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.513211046Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,107,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.533173951Z,31523,total_duration_us,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,107,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.533579345Z,26772,total_duration_us,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,107,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:36:00.475702635Z,21018,total_duration_us,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,107,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:39:00.490094464Z,23167,total_duration_us,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,107,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:41:00.49089932Z,23753,total_duration_us,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,107,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:42:00.483353951Z,26153,total_duration_us,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,107,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:47:00.506875644Z,37408,total_duration_us,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,108,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.522092597Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,03c19003200d7000
,,108,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.593976825Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,03c19003200d7000
,,108,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.630789006Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,109,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.954866323Z,459087,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,110,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:02:00.479534485Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,110,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:04:00.470594397Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,110,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:12:00.476923814Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,110,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.695016629Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,110,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.632624131Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000
,,110,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:32:00.482130432Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.697901808Z,122019,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.500304872Z,22685,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:09:00.475715338Z,24385,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:13:00.4883411Z,25841,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:16:00.471327771Z,24433,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:19:00.472909982Z,26378,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:21:00.473764932Z,24648,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:22:00.483040753Z,24904,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:23:00.718566244Z,270897,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:28:00.491388187Z,36758,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.475034596Z,25031,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:43:00.477120009Z,24231,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:48:00.537010408Z,74870,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:52:00.470429139Z,24455,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.530864823Z,30052,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.525931238Z,25591,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:34:00.481362068Z,26440,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,111,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:57:00.492651355Z,29075,total_duration_us,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,112,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.690868149Z,0,read_values,queryd_billing,queryd-a-567d7b8464-mr4s8,03c19003200d7000
,,112,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.704345198Z,0,read_values,queryd_billing,queryd-a-567d7b8464-mr4s8,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,113,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.954866323Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,114,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:01:00.484147634Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,114,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:07:00.480924824Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,114,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:27:00.490781811Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,114,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:29:00.473468521Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,114,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.607478294Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,114,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.620991461Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,114,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:51:00.479647924Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,114,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:56:00.588645189Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,114,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:03:00.487133158Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,114,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:06:00.484321671Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,114,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:33:00.482739897Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,114,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.516558315Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,114,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.517181373Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,114,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:44:00.479659339Z,0,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,115,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.29244898Z,0,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,115,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.511146754Z,0,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,115,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.647563651Z,0,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,115,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.51837934Z,0,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000
,,115,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.513211046Z,0,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,116,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:03.91066756Z,1040,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,116,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:21:00.513655817Z,62296,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,116,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:01.114372345Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,116,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:31:00.46674743Z,58568,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,116,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.813728039Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,116,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:01.012124567Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,116,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:51:00.511063137Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,116,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:01.713330923Z,608,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,116,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:36:00.479939528Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,116,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:01.310517133Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,117,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.968529532Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-wm622,03d01b74c8e09000
,,117,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:01.171849185Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-wm622,03d01b74c8e09000
,,117,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:01.484631063Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-wm622,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,118,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:01:00.516347576Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,0395bd7401aa3000
,,118,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:04:00.483414339Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,0395bd7401aa3000
,,118,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:16:00.491232711Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-nkr8z,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,119,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:01.65498559Z,1063097,total_duration_us,queryd_billing,queryd-a-567d7b8464-zpgpf,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,120,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:09:00.553102941Z,101695,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,120,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:11:00.482733139Z,22767,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,120,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.54992025Z,25121,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,120,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:32:00.476587805Z,26238,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,120,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:38:00.48083445Z,21545,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,120,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:43:00.475526043Z,24713,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,120,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:51:00.491071984Z,22770,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,120,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:53:00.478711805Z,23025,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,120,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:54:00.481033973Z,27596,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,120,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:59:00.491954975Z,28856,total_duration_us,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,121,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:06:00.492582743Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,121,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.497626928Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,121,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.497621429Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,121,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:33:00.484687388Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,121,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:37:00.478695875Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,121,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:39:00.481188183Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,121,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.532824377Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,121,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.303708334Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,122,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:56:00.54287719Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-zpgpf,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,123,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.635201896Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,123,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.527170667Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,123,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.495654052Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,123,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.242752012Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,123,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.532006543Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,123,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.515414938Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,123,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.74769122Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,123,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.60047454Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,123,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.563200167Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,124,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:31:00.478958223Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,124,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.501944968Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,124,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:36:00.488893386Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,124,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:42:00.489699624Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,124,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:47:00.483799335Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,124,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.585722242Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,124,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.558041208Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,124,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.656737151Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,125,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.651898933Z,0,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000
,,125,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:03:00.477061908Z,0,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000
,,125,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.490292928Z,0,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000
,,125,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.547483082Z,0,read_values,queryd_billing,queryd-a-567d7b8464-ggvwp,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,126,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.533173951Z,0,read_values,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,126,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.533579345Z,0,read_values,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,126,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:36:00.475702635Z,0,read_values,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,126,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:39:00.490094464Z,0,read_values,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,126,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:41:00.49089932Z,0,read_values,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,126,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:42:00.483353951Z,0,read_values,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,126,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:47:00.506875644Z,0,read_values,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,127,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:01.217638223Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-zpgpf,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,128,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:06:00.492582743Z,31249,total_duration_us,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,128,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.497626928Z,32192,total_duration_us,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,128,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.497621429Z,22854,total_duration_us,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,128,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:33:00.484687388Z,22891,total_duration_us,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,128,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:37:00.478695875Z,28481,total_duration_us,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,128,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:39:00.481188183Z,29004,total_duration_us,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,128,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.532824377Z,22289,total_duration_us,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,128,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.303708334Z,753881,total_duration_us,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,129,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.616067917Z,24993,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,129,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.553305038Z,71653,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,129,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.596901203Z,31766,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,129,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.598127073Z,48384,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,129,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.622914101Z,32684,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,129,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.512773568Z,38812,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,129,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.515196338Z,42132,total_duration_us,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,130,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:46:00.590968998Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,130,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.930211883Z,472,read_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,130,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:41:00.546438039Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,130,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:01.034314226Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,130,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.82933305Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,131,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:06:00.492582743Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,131,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.497626928Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,131,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.497621429Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,131,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:33:00.484687388Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,131,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:37:00.478695875Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,131,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:39:00.481188183Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,131,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.532824377Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000
,,131,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.303708334Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,132,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.588746873Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,132,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.616685301Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,132,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.701174358Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,132,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.685121539Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,132,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.529396168Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,132,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.549283022Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,132,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.511155262Z,0,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,133,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:11:00.492570159Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000
,,133,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:16:00.650605498Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000
,,133,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:01.070125773Z,60,read_values,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000
,,133,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:21:00.496795386Z,7408,read_values,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,134,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:01.682352598Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03c19003200d7000
,,134,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.650243169Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,135,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.636022941Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000
,,135,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.517726347Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000
,,135,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.504779428Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000
,,135,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.689278921Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,136,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.988611166Z,60,read_values,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,136,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.976033727Z,59,read_values,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,136,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.875974229Z,60,read_values,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,136,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:56:00.540440239Z,7740,read_values,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000
,,136,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:06:00.488297265Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-mr4s8,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,137,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:03.91066756Z,130,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,137,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:21:00.513655817Z,7787,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,137,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:01.114372345Z,60,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,137,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:31:00.46674743Z,7321,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,137,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.813728039Z,60,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,137,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:01.012124567Z,60,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,137,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:51:00.511063137Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,137,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:01.713330923Z,76,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,137,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:36:00.479939528Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000
,,137,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:01.310517133Z,60,read_values,queryd_billing,queryd-a-567d7b8464-7m48t,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,138,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:02.095923069Z,1552369,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,138,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:01.450962024Z,948628,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,138,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:26:00.554024549Z,59994227,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,138,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:01.096838578Z,587964,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,138,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:01.694674526Z,1199539,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,138,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:01:00.545725631Z,59993505,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,138,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:01.200032709Z,677237,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,138,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:01.198296884Z,666745,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,138,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:01.697935952Z,1189003,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,138,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:01.097195574Z,579423,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,138,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:01.29974034Z,773744,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,138,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:01.098950778Z,613731,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000
,,138,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:01.418781835Z,927293,total_duration_us,queryd_billing,queryd-a-567d7b8464-ggvwp,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,139,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:46:00.590968998Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,139,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.930211883Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,139,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:41:00.546438039Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,139,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:01.034314226Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,139,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.82933305Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,140,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:46:00.590968998Z,59994614,total_duration_us,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,140,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.930211883Z,435426,total_duration_us,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,140,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:41:00.546438039Z,59995714,total_duration_us,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,140,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:01.034314226Z,465452,total_duration_us,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000
,,140,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.82933305Z,361287,total_duration_us,queryd_billing,queryd-a-567d7b8464-2t9vn,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,141,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:01.292879825Z,90,read_values,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,141,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.398377049Z,60,read_values,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,141,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.893352315Z,60,read_values,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,141,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.894114015Z,60,read_values,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000
,,141,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:31:00.496776599Z,7800,read_values,queryd_billing,queryd-a-567d7b8464-wxzxp,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,142,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:11:00.492570159Z,59994307,total_duration_us,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000
,,142,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:16:00.650605498Z,59994795,total_duration_us,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000
,,142,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:01.070125773Z,502740,total_duration_us,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000
,,142,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:21:00.496795386Z,59995024,total_duration_us,queryd_billing,queryd-a-567d7b8464-nkr8z,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,143,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:08:00.627574839Z,0,read_values,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,143,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:14:00.49067527Z,0,read_values,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,143,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:17:00.487130291Z,0,read_values,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,143,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:38:00.479543081Z,0,read_values,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,143,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:18:00.47614158Z,0,read_values,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,144,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.588746873Z,35188,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,144,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.616685301Z,36350,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,144,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.701174358Z,154422,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,144,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.685121539Z,26011,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,144,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.529396168Z,45003,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,144,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.549283022Z,42857,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000
,,144,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.511155262Z,26785,total_duration_us,queryd_billing,queryd-a-567d7b8464-l9rhl,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,145,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:01.140329663Z,583338,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,145,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.937088519Z,383755,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,145,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.837769465Z,371740,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,145,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.938965657Z,482270,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,145,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:03.735264596Z,3246376,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,145,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.959621221Z,440089,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,145,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:51:00.579710518Z,59994096,total_duration_us,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,146,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.576894936Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03c19003200d7000
,,146,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.513387882Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03c19003200d7000
,,146,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.672684244Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,147,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.647640246Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,147,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.673152688Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,147,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.503074424Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,147,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.883388065Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,147,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:01.766565584Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,147,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.784667596Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,147,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.485214725Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,147,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.496265331Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,148,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.522092597Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,03c19003200d7000
,,148,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.593976825Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,03c19003200d7000
,,148,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.630789006Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,149,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.690868149Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,03c19003200d7000
,,149,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.704345198Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-mr4s8,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,150,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.498459129Z,28301,total_duration_us,queryd_billing,queryd-a-567d7b8464-wxzxp,03c19003200d7000
,,150,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.702457381Z,51019,total_duration_us,queryd_billing,queryd-a-567d7b8464-wxzxp,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,151,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:01.086269488Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,151,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.888541465Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,151,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:02.882870556Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,151,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.988118366Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,151,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:03.78792612Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,151,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.98470962Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,151,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.988438498Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,151,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:11:00.496709641Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,151,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:16:00.524449747Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,151,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.887275324Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,152,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:31:00.478958223Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,152,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.501944968Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,152,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:36:00.488893386Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,152,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:42:00.489699624Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,152,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:47:00.483799335Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,152,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.585722242Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,152,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.558041208Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000
,,152,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.656737151Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,153,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.522923171Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000
,,153,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:07:00.49308561Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000
,,153,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:19:00.484455371Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000
,,153,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:28:00.488167633Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,154,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:01.140329663Z,88,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,154,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.937088519Z,60,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,154,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.837769465Z,60,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,154,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.938965657Z,58,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,154,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:03.735264596Z,130,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,154,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.959621221Z,60,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000
,,154,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:51:00.579710518Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,155,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:01.682352598Z,0,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,03c19003200d7000
,,155,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.650243169Z,0,read_values,queryd_billing,queryd-a-567d7b8464-xb8ww,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,156,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.95995624Z,388105,total_duration_us,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000
,,156,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.962184778Z,481771,total_duration_us,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000
,,156,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.958940279Z,425120,total_duration_us,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000
,,156,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.870540202Z,402779,total_duration_us,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,157,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.968529532Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-wm622,03d01b74c8e09000
,,157,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:01.171849185Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-wm622,03d01b74c8e09000
,,157,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:01.484631063Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-wm622,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,158,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.533096623Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,158,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.657260033Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,158,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.523745277Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,158,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.521292944Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,158,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.508784116Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,158,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.519831768Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,158,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.684879Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000
,,158,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.720897719Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,159,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.522923171Z,0,read_values,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000
,,159,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:07:00.49308561Z,0,read_values,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000
,,159,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:19:00.484455371Z,0,read_values,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000
,,159,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:28:00.488167633Z,0,read_values,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,160,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.616067917Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,160,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.553305038Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,160,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.596901203Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,160,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.598127073Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,160,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.622914101Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,160,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.512773568Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,160,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.515196338Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,161,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:01:00.570149909Z,60000,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,161,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.867434006Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,161,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:01.272127486Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,161,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.970514197Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,161,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:01.078136693Z,720,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,161,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.572721595Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,161,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.973699773Z,480,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,161,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:01.870137074Z,640,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,161,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:46:00.520840844Z,59936,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,162,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.647640246Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,162,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.673152688Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,162,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.503074424Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,162,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:00.883388065Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,162,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:01.766565584Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,162,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.784667596Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,162,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.485214725Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000
,,162,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.496265331Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,163,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:56:00.54287719Z,78286,total_duration_us,queryd_billing,queryd-a-567d7b8464-zpgpf,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,164,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:09:00.553102941Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,164,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:11:00.482733139Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,164,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.54992025Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,164,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:32:00.476587805Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,164,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:38:00.48083445Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,164,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:43:00.475526043Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,164,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:51:00.491071984Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,164,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:53:00.478711805Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,164,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:54:00.481033973Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,164,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:59:00.491954975Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,165,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.95995624Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000
,,165,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.962184778Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000
,,165,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.958940279Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000
,,165,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.870540202Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,166,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.562048755Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,166,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:26:00.556708235Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,166,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:34:00.487483035Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,166,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:41:00.67594936Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,166,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:44:00.499102169Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,166,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.56992199Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,166,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:57:00.488454706Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,166,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:58:00.504724054Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,166,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:02:00.482659525Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,166,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:08:00.478316256Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,166,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:12:00.492881833Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,166,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:14:00.491639042Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,166,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:22:00.482878191Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000
,,166,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:24:00.4878516Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-l9rhl,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,167,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:01.217638223Z,724895,total_duration_us,queryd_billing,queryd-a-567d7b8464-zpgpf,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,168,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.576894936Z,30245,total_duration_us,queryd_billing,queryd-a-567d7b8464-2t9vn,03c19003200d7000
,,168,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.513387882Z,27447,total_duration_us,queryd_billing,queryd-a-567d7b8464-2t9vn,03c19003200d7000
,,168,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.672684244Z,31100,total_duration_us,queryd_billing,queryd-a-567d7b8464-2t9vn,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,169,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:01:00.516347576Z,0,read_values,queryd_billing,queryd-a-567d7b8464-nkr8z,0395bd7401aa3000
,,169,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:04:00.483414339Z,0,read_values,queryd_billing,queryd-a-567d7b8464-nkr8z,0395bd7401aa3000
,,169,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:16:00.491232711Z,0,read_values,queryd_billing,queryd-a-567d7b8464-nkr8z,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,170,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:01.283537001Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,170,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.887420221Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,170,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:01.685343496Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,170,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.888061441Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,170,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:56:00.478723104Z,0,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,171,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.522923171Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000
,,171,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:07:00.49308561Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000
,,171,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:19:00.484455371Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000
,,171,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:28:00.488167633Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-wxzxp,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,172,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:01.682352598Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03c19003200d7000
,,172,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.650243169Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-xb8ww,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,173,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.647838288Z,0,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,173,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.608466293Z,0,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,173,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.528214136Z,0,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,173,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.543848534Z,0,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,173,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.521564682Z,0,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,173,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.525124227Z,0,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,173,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.613172125Z,0,read_values,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,174,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.616067917Z,0,read_values,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,174,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.553305038Z,0,read_values,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,174,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.596901203Z,0,read_values,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,174,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:00.598127073Z,0,read_values,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,174,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.622914101Z,0,read_values,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,174,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.512773568Z,0,read_values,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000
,,174,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.515196338Z,0,read_values,queryd_billing,queryd-a-567d7b8464-db72l,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,175,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.533173951Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,175,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.533579345Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,175,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:36:00.475702635Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,175,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:39:00.490094464Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,175,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:41:00.49089932Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,175,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:42:00.483353951Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000
,,175,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:47:00.506875644Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-rlmq5,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,176,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.522092597Z,0,read_values,queryd_billing,queryd-a-567d7b8464-rlmq5,03c19003200d7000
,,176,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.593976825Z,0,read_values,queryd_billing,queryd-a-567d7b8464-rlmq5,03c19003200d7000
,,176,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:00.630789006Z,0,read_values,queryd_billing,queryd-a-567d7b8464-rlmq5,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,177,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:56:00.54287719Z,0,read_values,queryd_billing,queryd-a-567d7b8464-zpgpf,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,178,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.635201896Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,178,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.527170667Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,178,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.495654052Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,178,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.242752012Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,178,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.532006543Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,178,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.515414938Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,178,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.74769122Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,178,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.60047454Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,178,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.563200167Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,179,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:01.086269488Z,90,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,179,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.888541465Z,58,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,179,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:35:02.882870556Z,90,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,179,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.988118366Z,60,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,179,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:03.78792612Z,116,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,179,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.98470962Z,60,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,179,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.988438498Z,60,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,179,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:11:00.496709641Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,179,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:16:00.524449747Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000
,,179,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.887275324Z,60,read_values,queryd_billing,queryd-a-567d7b8464-l9rhl,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,180,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.501845561Z,0,read_values,queryd_billing,queryd-a-567d7b8464-nkr8z,03c19003200d7000
,,180,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.513513395Z,0,read_values,queryd_billing,queryd-a-567d7b8464-nkr8z,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,181,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:01:00.484147634Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,181,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:07:00.480924824Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,181,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:27:00.490781811Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,181,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:29:00.473468521Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,181,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.607478294Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,181,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.620991461Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,181,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:51:00.479647924Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,181,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:56:00.588645189Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,181,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:03:00.487133158Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,181,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:06:00.484321671Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,181,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:33:00.482739897Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,181,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.516558315Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,181,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.517181373Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000
,,181,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:44:00.479659339Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-twpfq,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,182,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:11:00.478414062Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,182,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:18:00.488037379Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,182,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:24:00.493400923Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,182,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.557061177Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,182,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:49:00.483622115Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,182,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:53:00.48923721Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,182,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:59:00.486840652Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,182,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.531132434Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,182,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:21:00.491145961Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,182,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:27:00.487253317Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,182,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:37:00.491349011Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000
,,182,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.621251682Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,183,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.576894936Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2t9vn,03c19003200d7000
,,183,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.513387882Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2t9vn,03c19003200d7000
,,183,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.672684244Z,0,read_values,queryd_billing,queryd-a-567d7b8464-2t9vn,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,184,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:09:00.553102941Z,0,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,184,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:11:00.482733139Z,0,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,184,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.54992025Z,0,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,184,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:32:00.476587805Z,0,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,184,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:38:00.48083445Z,0,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,184,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:43:00.475526043Z,0,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,184,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:51:00.491071984Z,0,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,184,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:53:00.478711805Z,0,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,184,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:54:00.481033973Z,0,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,184,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:59:00.491954975Z,0,read_values,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,185,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:01.880930931Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03d01b74c8e09000
,,185,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.977436912Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03d01b74c8e09000
,,185,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:01.179598985Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,186,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:10:01.283537001Z,794129,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,186,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:00.887420221Z,379452,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,186,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:01.685343496Z,1217499,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,186,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.888061441Z,387678,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000
,,186,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:56:00.478723104Z,59994378,total_duration_us,queryd_billing,queryd-a-567d7b8464-2dw7s,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,187,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:01.880930931Z,1333857,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,03d01b74c8e09000
,,187,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:00.977436912Z,505862,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,03d01b74c8e09000
,,187,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:40:01.179598985Z,617529,total_duration_us,queryd_billing,queryd-a-567d7b8464-q82tv,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,188,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.968529532Z,1466985,total_duration_us,queryd_billing,queryd-a-567d7b8464-wm622,03d01b74c8e09000
,,188,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:01.171849185Z,645481,total_duration_us,queryd_billing,queryd-a-567d7b8464-wm622,03d01b74c8e09000
,,188,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:01.484631063Z,970039,total_duration_us,queryd_billing,queryd-a-567d7b8464-wm622,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,189,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:01.65498559Z,0,read_values,queryd_billing,queryd-a-567d7b8464-zpgpf,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,190,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:09:00.553102941Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,190,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:11:00.482733139Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,190,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.54992025Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,190,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:32:00.476587805Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,190,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:38:00.48083445Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,190,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:43:00.475526043Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,190,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:51:00.491071984Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,190,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:53:00.478711805Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,190,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:54:00.481033973Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000
,,190,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:59:00.491954975Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-svsmx,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,191,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:55:01.217638223Z,90,read_values,queryd_billing,queryd-a-567d7b8464-zpgpf,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,192,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.635201896Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,192,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.527170667Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,192,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.495654052Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,192,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.242752012Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,192,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:15:00.532006543Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,192,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.515414938Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,192,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.74769122Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,192,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.60047454Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000
,,192,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.563200167Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2dw7s,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,193,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:08:00.627574839Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,193,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:14:00.49067527Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,193,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:17:00.487130291Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,193,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:38:00.479543081Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000
,,193,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:18:00.47614158Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-db72l,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:00:00.697901808Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.500304872Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:09:00.475715338Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:13:00.4883411Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:16:00.471327771Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:19:00.472909982Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:21:00.473764932Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:22:00.483040753Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:23:00.718566244Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:28:00.491388187Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:30:00.475034596Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:43:00.477120009Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:48:00.537010408Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:52:00.470429139Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:20:00.530864823Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.525931238Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:34:00.481362068Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000
,,194,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:57:00.492651355Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-7m48t,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,195,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:01:00.570149909Z,7500,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,195,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:05:00.867434006Z,60,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,195,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:15:01.272127486Z,90,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,195,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.970514197Z,60,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,195,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:01.078136693Z,90,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,195,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.572721595Z,60,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,195,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:05:00.973699773Z,60,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,195,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:01.870137074Z,80,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000
,,195,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:46:00.520840844Z,7492,read_values,queryd_billing,queryd-a-567d7b8464-twpfq,03d01b74c8e09000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,196,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.647838288Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,196,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:45:00.608466293Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,196,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:50:00.528214136Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,196,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:55:00.543848534Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,196,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.521564682Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,196,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.525124227Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000
,,196,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:50:00.613172125Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-q82tv,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,197,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:20:00.542910288Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,03c19003200d7000
,,197,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.001227019Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-8hmpp,03c19003200d7000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,198,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:00:01.296146195Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,198,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:10:00.523493998Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,198,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:30:00.524743262Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,198,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:40:00.53008227Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000
,,198,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:52:00.479092929Z,0,read_bytes,queryd_billing,queryd-a-567d7b8464-2t9vn,0395bd7401aa3000

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,hostname,org_id
,,199,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T12:25:00.636022941Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000
,,199,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:25:00.517726347Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000
,,199,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:35:00.504779428Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000
,,199,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,2019-08-01T13:45:00.689278921Z,2,response_bytes,queryd_billing,queryd-a-567d7b8464-2sbq5,03c19003200d7000
"

outData = "
#group,false,false,true,true,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,long,dateTime:RFC3339
#default,duration_us,,,,,
,result,table,_start,_stop,duration_us,_time
,,0,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,750144902,2019-08-01T13:00:00Z
,,0,2019-08-01T12:00:00Z,2019-08-01T14:00:00Z,745212595,2019-08-01T14:00:00Z
"

_f = (table=<-) => table
    |> range(start: 2019-08-01T12:00:00Z, stop: 2019-08-01T14:00:00Z)
    |> filter(fn: (r) =>
        r.org_id == "03d01b74c8e09000"
        and r._measurement == "queryd_billing"
        and r._field == "total_duration_us"
    )
    |> group()
    |> aggregateWindow(every: 1h, fn: sum)
    |> fill(column: "_value", value: 0)
    |> rename(columns: {_value: "duration_us"})
    |> yield(name: "duration_us")

test query_duration = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: _f})
