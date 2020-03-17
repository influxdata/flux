package aggregate_test

import "experimental/aggregate"
import "testing"

inData = "
#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,0,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_out,net,localhost,en5
,,0,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_out,net,localhost,en5

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,1,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,1,blocked,processes,localhost
,,1,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,1,blocked,processes,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,2,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_out,net,localhost,utun0
,,2,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_out,net,localhost,utun0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,3,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_out,net,localhost,utun1
,,3,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_out,net,localhost,utun1

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,4,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,2,running,processes,localhost
,,4,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,1,running,processes,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,5,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,stopped,processes,localhost
,,5,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,stopped,processes,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,6,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,466,total,processes,localhost
,,6,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,463,total,processes,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,7,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,zombies,processes,localhost
,,7,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,zombies,processes,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,8,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,22742,bytes_sent,net,localhost,utun0
,,8,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,22742,bytes_sent,net,localhost,utun0

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,9,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,unknown,processes,localhost
,,9,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,unknown,processes,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,10,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,bytes_recv,net,localhost,utun1
,,10,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,bytes_recv,net,localhost,utun1

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,11,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,319,packets_recv,net,localhost,en5
,,11,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,319,packets_recv,net,localhost,en5

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,12,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_in,net,localhost,utun0
,,12,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_in,net,localhost,utun0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,13,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,97772,bytes_recv,net,localhost,en5
,,13,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,97772,bytes_recv,net,localhost,en5

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,14,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,44381,bytes_sent,net,localhost,en5
,,14,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,44381,bytes_sent,net,localhost,en5

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,15,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,463,sleeping,processes,localhost
,,15,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,461,sleeping,processes,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,16,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,119,packets_sent,net,localhost,utun0
,,16,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,119,packets_sent,net,localhost,utun0

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,17,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,idle,processes,localhost
,,17,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,idle,processes,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,18,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_in,net,localhost,llw0
,,18,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_in,net,localhost,llw0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,19,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_out,net,localhost,en5
,,19,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_out,net,localhost,en5

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,20,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_out,net,localhost,awdl0
,,20,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_out,net,localhost,awdl0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,21,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_in,net,localhost,en5
,,21,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_in,net,localhost,en5

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,22,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,31632,bytes_sent,net,localhost,awdl0
,,22,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,31632,bytes_sent,net,localhost,awdl0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,23,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_in,net,localhost,en5
,,23,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_in,net,localhost,en5

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,24,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,packets_recv,net,localhost,llw0
,,24,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,packets_recv,net,localhost,llw0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,25,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,306,packets_sent,net,localhost,en5
,,25,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,306,packets_sent,net,localhost,en5

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,26,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_in,net,localhost,awdl0
,,26,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_in,net,localhost,awdl0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,27,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,119,packets_sent,net,localhost,utun1
,,27,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,119,packets_sent,net,localhost,utun1

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,28,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,packets_recv,net,localhost,en3
,,28,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,packets_recv,net,localhost,en3

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,29,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_out,net,localhost,utun1
,,29,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_out,net,localhost,utun1

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,30,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,bytes_recv,net,localhost,awdl0
,,30,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,bytes_recv,net,localhost,awdl0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,31,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,packets_recv,net,localhost,utun1
,,31,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,packets_recv,net,localhost,utun1

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,32,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,184,packets_sent,net,localhost,awdl0
,,32,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,184,packets_sent,net,localhost,awdl0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,33,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,22742,bytes_sent,net,localhost,utun1
,,33,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,22742,bytes_sent,net,localhost,utun1

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,34,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_out,net,localhost,p2p0
,,34,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_out,net,localhost,p2p0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,35,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_in,net,localhost,llw0
,,35,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_in,net,localhost,llw0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,36,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_out,net,localhost,en1
,,36,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_out,net,localhost,en1

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,37,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_in,net,localhost,utun1
,,37,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_in,net,localhost,utun1

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,38,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_in,net,localhost,p2p0
,,38,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_in,net,localhost,p2p0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,39,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_in,net,localhost,utun1
,,39,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_in,net,localhost,utun1

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,40,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,packets_recv,net,localhost,awdl0
,,40,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,packets_recv,net,localhost,awdl0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,41,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,packets_recv,net,localhost,utun0
,,41,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,packets_recv,net,localhost,utun0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,42,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_out,net,localhost,en4
,,42,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_out,net,localhost,en4

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,43,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,packets_sent,net,localhost,llw0
,,43,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,packets_sent,net,localhost,llw0

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,44,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,\"3 days, 19:21\",uptime_format,system,localhost
,,44,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,\"3 days, 19:21\",uptime_format,system,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,45,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,bytes_recv,net,localhost,utun0
,,45,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,bytes_recv,net,localhost,utun0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,46,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,bytes_sent,net,localhost,en4
,,46,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,bytes_sent,net,localhost,en4

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,47,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,bytes_recv,net,localhost,en2
,,47,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,bytes_recv,net,localhost,en2

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,48,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_out,net,localhost,p2p0
,,48,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_out,net,localhost,p2p0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,49,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_in,net,localhost,utun0
,,49,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_in,net,localhost,utun0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,50,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,packets_sent,net,localhost,en2
,,50,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,packets_sent,net,localhost,en2

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,51,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,bytes_sent,net,localhost,p2p0
,,51,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,bytes_sent,net,localhost,p2p0

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,52,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,239822708736,free,disk,disk1s4,apfs,localhost,rw,/private/var/vm
,,52,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,239822426112,free,disk,disk1s4,apfs,localhost,rw,/private/var/vm

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,53,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_out,net,localhost,llw0
,,53,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_out,net,localhost,llw0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,54,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,bytes_recv,net,localhost,en4
,,54,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,bytes_recv,net,localhost,en4

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,55,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,packets_recv,net,localhost,en1
,,55,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,packets_recv,net,localhost,en1

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,56,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,bytes_recv,net,localhost,p2p0
,,56,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,bytes_recv,net,localhost,p2p0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,57,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_out,net,localhost,utun0
,,57,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_out,net,localhost,utun0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,58,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_in,net,localhost,en2
,,58,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_in,net,localhost,en2

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,59,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_out,net,localhost,en2
,,59,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_out,net,localhost,en2

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,60,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,4882452836,inodes_free,disk,disk1s4,apfs,localhost,rw,/private/var/vm
,,60,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,4882452836,inodes_free,disk,disk1s4,apfs,localhost,rw,/private/var/vm

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,61,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_out,net,localhost,awdl0
,,61,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_out,net,localhost,awdl0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,62,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,packets_sent,net,localhost,en3
,,62,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,packets_sent,net,localhost,en3

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,63,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_in,net,localhost,en0
,,63,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_in,net,localhost,en0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,64,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_in,net,localhost,en3
,,64,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_in,net,localhost,en3

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,65,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_out,net,localhost,llw0
,,65,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_out,net,localhost,llw0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,66,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,bytes_sent,net,localhost,en1
,,66,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,bytes_sent,net,localhost,en1

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,67,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,bytes_recv,net,localhost,en1
,,67,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,bytes_recv,net,localhost,en1

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,68,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,499963170816,total,disk,disk1s4,apfs,localhost,rw,/private/var/vm
,,68,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,499963170816,total,disk,disk1s4,apfs,localhost,rw,/private/var/vm

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,69,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,packets_sent,net,localhost,p2p0
,,69,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,packets_sent,net,localhost,p2p0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,70,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_out,net,localhost,en3
,,70,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_out,net,localhost,en3

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,71,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,61413965824,write_bytes,diskio,localhost,disk0
,,71,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,61418156032,write_bytes,diskio,localhost,disk0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,72,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_out,net,localhost,en2
,,72,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_out,net,localhost,en2

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,73,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,bytes_sent,net,localhost,llw0
,,73,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,bytes_sent,net,localhost,llw0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,74,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_in,net,localhost,en0
,,74,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_in,net,localhost,en0

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,75,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,out,swap,localhost
,,75,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,out,swap,localhost

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,76,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,499963170816,total,disk,disk1s1,apfs,localhost,rw,/System/Volumes/Data
,,76,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,499963170816,total,disk,disk1s1,apfs,localhost,rw,/System/Volumes/Data

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,77,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_in,net,localhost,p2p0
,,77,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_in,net,localhost,p2p0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,78,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_out,net,localhost,en3
,,78,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_out,net,localhost,en3

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,79,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,merged_reads,diskio,localhost,disk0
,,79,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,merged_reads,diskio,localhost,disk0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,80,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_out,net,localhost,en1
,,80,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_out,net,localhost,en1

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,81,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,bytes_recv,net,localhost,llw0
,,81,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,bytes_recv,net,localhost,llw0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,82,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,6494393,packets_recv,net,localhost,en0
,,82,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,6496118,packets_recv,net,localhost,en0

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,83,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,239822708736,free,disk,disk1s5,apfs,localhost,ro,/
,,83,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,239822426112,free,disk,disk1s5,apfs,localhost,ro,/

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,84,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,4882452840,inodes_total,disk,disk1s1,apfs,localhost,rw,/System/Volumes/Data
,,84,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,4882452840,inodes_total,disk,disk1s1,apfs,localhost,rw,/System/Volumes/Data

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,85,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_in,net,localhost,en4
,,85,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_in,net,localhost,en4

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,86,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,bytes_sent,net,localhost,en3
,,86,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,bytes_sent,net,localhost,en3

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,87,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,3.18994140625,load5,system,localhost
,,87,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,3.2333984375,load5,system,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,88,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_out,net,localhost,en0
,,88,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_out,net,localhost,en0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,89,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_in,net,localhost,awdl0
,,89,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_in,net,localhost,awdl0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,90,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,merged_writes,diskio,localhost,disk0
,,90,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,merged_writes,diskio,localhost,disk0

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,91,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,shared,mem,localhost
,,91,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,shared,mem,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,92,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,43.00837516784668,available_percent,mem,localhost
,,92,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,43.38059425354004,available_percent,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,93,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_out,net,localhost,en4
,,93,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_out,net,localhost,en4

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,94,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,bytes_sent,net,localhost,en2
,,94,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,bytes_sent,net,localhost,en2

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,95,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,2.5400390625,load1,system,localhost
,,95,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,2.85693359375,load1,system,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,96,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,540017555,bytes_sent,net,localhost,en0
,,96,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,540066055,bytes_sent,net,localhost,en0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,97,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,packets_recv,net,localhost,p2p0
,,97,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,packets_recv,net,localhost,p2p0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,98,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,3164686,writes,diskio,localhost,disk0
,,98,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,3165333,writes,diskio,localhost,disk0

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,99,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,committed_as,mem,localhost
,,99,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,committed_as,mem,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,100,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,7388782592,available,mem,localhost
,,100,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,7452729344,available,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,101,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,bytes_recv,net,localhost,en3
,,101,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,bytes_recv,net,localhost,en3

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,102,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,packets_sent,net,localhost,en1
,,102,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,packets_sent,net,localhost,en1

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,103,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,4879320232,inodes_free,disk,disk1s1,apfs,localhost,rw,/System/Volumes/Data
,,103,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,4879320220,inodes_free,disk,disk1s1,apfs,localhost,rw,/System/Volumes/Data

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,104,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,2383655,packets_sent,net,localhost,en0
,,104,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,2383933,packets_sent,net,localhost,en0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,105,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,packets_sent,net,localhost,en4
,,105,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,packets_sent,net,localhost,en4

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,106,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,328894,uptime,system,localhost
,,106,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,328904,uptime,system,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,107,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,dirty,mem,localhost
,,107,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,dirty,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,108,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_irq,cpu,cpu6,localhost
,,108,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_irq,cpu,cpu6,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,109,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_in,net,localhost,en3
,,109,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_in,net,localhost,en3

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,110,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,err_in,net,localhost,en1
,,110,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,err_in,net,localhost,en1

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,111,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,3132608,inodes_used,disk,disk1s1,apfs,localhost,rw,/System/Volumes/Data
,,111,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,3132620,inodes_used,disk,disk1s1,apfs,localhost,rw,/System/Volumes/Data

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,112,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,3035975,reads,diskio,localhost,disk0
,,112,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,3036038,reads,diskio,localhost,disk0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,113,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,packets_recv,net,localhost,en4
,,113,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,packets_recv,net,localhost,en4

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,114,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,3.02734375,load15,system,localhost
,,114,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,3.0439453125,load15,system,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,115,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,swap_cached,mem,localhost
,,115,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,swap_cached,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,116,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_iowait,cpu,cpu5,localhost
,,116,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_iowait,cpu,cpu5,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,117,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_in,net,localhost,en1
,,117,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_in,net,localhost,en1

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,118,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,8022268513,bytes_recv,net,localhost,en0
,,118,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,8023573796,bytes_recv,net,localhost,en0

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,119,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,499963170816,total,disk,disk1s5,apfs,localhost,ro,/
,,119,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,499963170816,total,disk,disk1s5,apfs,localhost,ro,/

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,120,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,2004394,read_time,diskio,localhost,disk0
,,120,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,2004455,read_time,diskio,localhost,disk0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,121,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_in,net,localhost,en4
,,121,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_in,net,localhost,en4

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,122,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,4,n_users,system,localhost
,,122,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,4,n_users,system,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,123,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,huge_pages_total,mem,localhost
,,123,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,huge_pages_total,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,124,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,10.4,usage_system,cpu,cpu4,localhost
,,124,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,8.89110889110889,usage_system,cpu,cpu4,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,125,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,weighted_io_time,diskio,localhost,disk0
,,125,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,weighted_io_time,diskio,localhost,disk0

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,126,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,3052868,io_time,diskio,localhost,disk0
,,126,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,3053173,io_time,diskio,localhost,disk0

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,127,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,11009597440,used,disk,disk1s5,apfs,localhost,ro,/
,,127,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,11009597440,used,disk,disk1s5,apfs,localhost,ro,/

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,128,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,35.7421875,used_percent,swap,localhost
,,128,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,35.7421875,used_percent,swap,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,129,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,packets_recv,net,localhost,en2
,,129,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,packets_recv,net,localhost,en2

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,130,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,484367,inodes_used,disk,disk1s5,apfs,localhost,ro,/
,,130,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,484367,inodes_used,disk,disk1s5,apfs,localhost,ro,/

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,131,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,write_back,mem,localhost
,,131,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,write_back,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,132,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_iowait,cpu,cpu4,localhost
,,132,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_iowait,cpu,cpu4,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,133,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,iops_in_progress,diskio,localhost,disk0
,,133,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,iops_in_progress,diskio,localhost,disk0

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,134,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,2147483648,total,swap,localhost
,,134,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,2147483648,total,swap,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,135,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,page_tables,mem,localhost
,,135,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,page_tables,mem,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,136,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,8,n_cpus,system,localhost
,,136,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,8,n_cpus,system,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,137,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,drop_in,net,localhost,en2
,,137,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,drop_in,net,localhost,en2

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,138,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,4882452840,inodes_total,disk,disk1s5,apfs,localhost,ro,/
,,138,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,4882452840,inodes_total,disk,disk1s5,apfs,localhost,ro,/

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,139,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,low_total,mem,localhost
,,139,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,low_total,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,140,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest,cpu,cpu4,localhost
,,140,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest,cpu,cpu4,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,141,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,1379926016,free,swap,localhost
,,141,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,1379926016,free,swap,localhost

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,142,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,3222294528,used,disk,disk1s4,apfs,localhost,rw,/private/var/vm
,,142,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,3222294528,used,disk,disk1s4,apfs,localhost,rw,/private/var/vm

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,143,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,17179869184,total,mem,localhost
,,143,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,17179869184,total,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,144,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,4881968473,inodes_free,disk,disk1s5,apfs,localhost,ro,/
,,144,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,4881968473,inodes_free,disk,disk1s5,apfs,localhost,ro,/

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,145,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,89,drop_out,net,localhost,en0
,,145,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,89,drop_out,net,localhost,en0

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,146,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,swap_free,mem,localhost
,,146,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,swap_free,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,147,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,2.4,usage_system,cpu,cpu7,localhost
,,147,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,1.2987012987012987,usage_system,cpu,cpu7,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,148,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_nice,cpu,cpu2,localhost
,,148,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_nice,cpu,cpu2,localhost

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,149,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,4,inodes_used,disk,disk1s4,apfs,localhost,rw,/private/var/vm
,,149,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,4,inodes_used,disk,disk1s4,apfs,localhost,rw,/private/var/vm

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,150,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,1.325801594242151,used_percent,disk,disk1s4,apfs,localhost,rw,/private/var/vm
,,150,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,1.3258031359475162,used_percent,disk,disk1s4,apfs,localhost,rw,/private/var/vm

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,151,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,slab,mem,localhost
,,151,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,slab,mem,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,152,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,9791086592,used,mem,localhost
,,152,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,9727139840,used,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,153,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,1048474,write_time,diskio,localhost,disk0
,,153,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,1048718,write_time,diskio,localhost,disk0

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,154,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,swap_total,mem,localhost
,,154,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,swap_total,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,155,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,92.3,usage_idle,cpu,cpu7,localhost
,,155,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,96.7032967032967,usage_idle,cpu,cpu7,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,156,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,14.4,usage_system,cpu,cpu0,localhost
,,156,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,14.985014985014985,usage_system,cpu,cpu0,localhost

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,157,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,239822708736,free,disk,disk1s1,apfs,localhost,rw,/System/Volumes/Data
,,157,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,239822426112,free,disk,disk1s1,apfs,localhost,rw,/System/Volumes/Data

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,158,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,245139259392,used,disk,disk1s1,apfs,localhost,rw,/System/Volumes/Data
,,158,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,245139542016,used,disk,disk1s1,apfs,localhost,rw,/System/Volumes/Data

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,159,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,mapped,mem,localhost
,,159,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,mapped,mem,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,160,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,high_total,mem,localhost
,,160,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,high_total,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,161,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,69740752896,read_bytes,diskio,localhost,disk0
,,161,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,69751013376,read_bytes,diskio,localhost,disk0

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,162,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,vmalloc_used,mem,localhost
,,162,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,vmalloc_used,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,163,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_nice,cpu,cpu6,localhost
,,163,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_nice,cpu,cpu6,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,164,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,cached,mem,localhost
,,164,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,cached,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,165,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,50.548140988923564,used_percent,disk,disk1s1,apfs,localhost,rw,/System/Volumes/Data
,,165,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,50.54819926648316,used_percent,disk,disk1s1,apfs,localhost,rw,/System/Volumes/Data

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,166,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,17.025,usage_user,cpu,cpu-total,localhost
,,166,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,9.22038980509745,usage_user,cpu,cpu-total,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,167,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,sreclaimable,mem,localhost
,,167,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,sreclaimable,mem,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,168,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,in,swap,localhost
,,168,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,in,swap,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,169,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,56.99162483215332,used_percent,mem,localhost
,,169,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,56.61940574645996,used_percent,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,170,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest,cpu,cpu5,localhost
,,170,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest,cpu,cpu5,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,171,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,huge_pages_free,mem,localhost
,,171,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,huge_pages_free,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,172,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,4.389226255518682,used_percent,disk,disk1s5,apfs,localhost,ro,/
,,172,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,4.389231201062172,used_percent,disk,disk1s5,apfs,localhost,ro,/

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,173,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest,cpu,cpu7,localhost
,,173,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest,cpu,cpu7,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,174,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,buffered,mem,localhost
,,174,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,buffered,mem,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,175,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,767557632,used,swap,localhost
,,175,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,767557632,used,swap,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,176,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_irq,cpu,cpu-total,localhost
,,176,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_irq,cpu,cpu-total,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,177,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,5.3,usage_user,cpu,cpu5,localhost
,,177,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,2,usage_user,cpu,cpu5,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,178,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,high_free,mem,localhost
,,178,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,high_free,mem,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,179,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,5859033088,active,mem,localhost
,,179,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,5816074240,active,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,180,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_steal,cpu,cpu7,localhost
,,180,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_steal,cpu,cpu7,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,181,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,4620652544,inactive,mem,localhost
,,181,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,4612915200,inactive,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,182,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,4882452840,inodes_total,disk,disk1s4,apfs,localhost,rw,/private/var/vm
,,182,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,4882452840,inodes_total,disk,disk1s4,apfs,localhost,rw,/private/var/vm

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,183,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_nice,cpu,cpu-total,localhost
,,183,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_nice,cpu,cpu-total,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,184,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,2.797202797202797,usage_system,cpu,cpu3,localhost
,,184,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,1.7,usage_system,cpu,cpu3,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,185,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,write_back_tmp,mem,localhost
,,185,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,write_back_tmp,mem,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,186,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,huge_page_size,mem,localhost
,,186,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,huge_page_size,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,187,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,8.90890890890891,usage_system,cpu,cpu6,localhost
,,187,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,6.593406593406593,usage_system,cpu,cpu6,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,188,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_irq,cpu,cpu7,localhost
,,188,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_irq,cpu,cpu7,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,189,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,vmalloc_total,mem,localhost
,,189,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,vmalloc_total,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,190,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,25.925925925925927,usage_user,cpu,cpu6,localhost
,,190,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,13.486513486513486,usage_user,cpu,cpu6,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,191,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,59,usage_idle,cpu,cpu2,localhost
,,191,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,71.07107107107107,usage_idle,cpu,cpu2,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,192,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest,cpu,cpu-total,localhost
,,192,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest,cpu,cpu-total,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,193,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,commit_limit,mem,localhost
,,193,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,commit_limit,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,194,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest,cpu,cpu3,localhost
,,194,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest,cpu,cpu3,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,195,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_iowait,cpu,cpu7,localhost
,,195,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_iowait,cpu,cpu7,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,196,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,2768130048,free,mem,localhost
,,196,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,2839814144,free,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,197,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,65.16516516516516,usage_idle,cpu,cpu6,localhost
,,197,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,79.92007992007991,usage_idle,cpu,cpu6,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,198,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_iowait,cpu,cpu2,localhost
,,198,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_iowait,cpu,cpu2,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,199,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,7.1875,usage_system,cpu,cpu-total,localhost
,,199,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,6.184407796101949,usage_system,cpu,cpu-total,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,200,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,sunreclaim,mem,localhost
,,200,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,sunreclaim,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,201,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest_nice,cpu,cpu3,localhost
,,201,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest_nice,cpu,cpu3,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,202,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_softirq,cpu,cpu6,localhost
,,202,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_softirq,cpu,cpu6,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,203,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,2612187136,wired,mem,localhost
,,203,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,2582466560,wired,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,204,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,92,usage_idle,cpu,cpu5,localhost
,,204,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,96.5,usage_idle,cpu,cpu5,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,205,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_softirq,cpu,cpu0,localhost
,,205,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_softirq,cpu,cpu0,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,206,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_iowait,cpu,cpu-total,localhost
,,206,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_iowait,cpu,cpu-total,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,207,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_softirq,cpu,cpu-total,localhost
,,207,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_softirq,cpu,cpu-total,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,208,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_softirq,cpu,cpu3,localhost
,,208,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_softirq,cpu,cpu3,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,209,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest,cpu,cpu6,localhost
,,209,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest,cpu,cpu6,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,210,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,low_free,mem,localhost
,,210,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,low_free,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,211,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_nice,cpu,cpu4,localhost
,,211,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_nice,cpu,cpu4,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,212,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_steal,cpu,cpu0,localhost
,,212,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_steal,cpu,cpu0,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,213,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,5.3,usage_user,cpu,cpu7,localhost
,,213,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,1.998001998001998,usage_user,cpu,cpu7,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,214,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest_nice,cpu,cpu-total,localhost
,,214,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest_nice,cpu,cpu-total,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,215,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,28.4,usage_user,cpu,cpu2,localhost
,,215,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,16.916916916916918,usage_user,cpu,cpu2,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,216,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_nice,cpu,cpu5,localhost
,,216,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_nice,cpu,cpu5,localhost

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,217,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,vmalloc_chunk,mem,localhost
,,217,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,vmalloc_chunk,mem,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,218,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_irq,cpu,cpu4,localhost
,,218,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_irq,cpu,cpu4,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,219,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_nice,cpu,cpu7,localhost
,,219,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_nice,cpu,cpu7,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,220,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,75.7875,usage_idle,cpu,cpu-total,localhost
,,220,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,84.5952023988006,usage_idle,cpu,cpu-total,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,221,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_softirq,cpu,cpu1,localhost
,,221,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_softirq,cpu,cpu1,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,222,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,61.5,usage_idle,cpu,cpu4,localhost
,,222,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,75.82417582417582,usage_idle,cpu,cpu4,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,223,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_steal,cpu,cpu-total,localhost
,,223,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_steal,cpu,cpu-total,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,224,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_steal,cpu,cpu4,localhost
,,224,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_steal,cpu,cpu4,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,225,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_irq,cpu,cpu5,localhost
,,225,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_irq,cpu,cpu5,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,226,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest_nice,cpu,cpu7,localhost
,,226,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest_nice,cpu,cpu7,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,227,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,91.1,usage_idle,cpu,cpu1,localhost
,,227,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,95.2047952047952,usage_idle,cpu,cpu1,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,228,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,91.80819180819181,usage_idle,cpu,cpu3,localhost
,,228,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,96.3,usage_idle,cpu,cpu3,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,229,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_softirq,cpu,cpu7,localhost
,,229,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_softirq,cpu,cpu7,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,230,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,12.6,usage_system,cpu,cpu2,localhost
,,230,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,12.012012012012011,usage_system,cpu,cpu2,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,231,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_irq,cpu,cpu3,localhost
,,231,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_irq,cpu,cpu3,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,232,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_iowait,cpu,cpu6,localhost
,,232,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_iowait,cpu,cpu6,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,233,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest_nice,cpu,cpu0,localhost
,,233,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest_nice,cpu,cpu0,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,234,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_steal,cpu,cpu3,localhost
,,234,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_steal,cpu,cpu3,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,235,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest_nice,cpu,cpu6,localhost
,,235,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest_nice,cpu,cpu6,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,236,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_steal,cpu,cpu2,localhost
,,236,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_steal,cpu,cpu2,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,237,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_steal,cpu,cpu1,localhost
,,237,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_steal,cpu,cpu1,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,238,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_steal,cpu,cpu6,localhost
,,238,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_steal,cpu,cpu6,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,239,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_iowait,cpu,cpu3,localhost
,,239,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_iowait,cpu,cpu3,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,240,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest_nice,cpu,cpu5,localhost
,,240,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest_nice,cpu,cpu5,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,241,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_nice,cpu,cpu0,localhost
,,241,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_nice,cpu,cpu0,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,242,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,5.6,usage_user,cpu,cpu1,localhost
,,242,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,2.2977022977022976,usage_user,cpu,cpu1,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,243,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_softirq,cpu,cpu5,localhost
,,243,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_softirq,cpu,cpu5,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,244,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest_nice,cpu,cpu2,localhost
,,244,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest_nice,cpu,cpu2,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,245,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_steal,cpu,cpu5,localhost
,,245,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_steal,cpu,cpu5,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,246,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_irq,cpu,cpu0,localhost
,,246,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_irq,cpu,cpu0,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,247,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_nice,cpu,cpu1,localhost
,,247,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_nice,cpu,cpu1,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,248,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,2.7,usage_system,cpu,cpu5,localhost
,,248,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,1.5,usage_system,cpu,cpu5,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,249,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest,cpu,cpu2,localhost
,,249,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest,cpu,cpu2,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,250,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,28.1,usage_user,cpu,cpu4,localhost
,,250,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,15.284715284715285,usage_user,cpu,cpu4,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,251,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest,cpu,cpu0,localhost
,,251,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest,cpu,cpu0,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,252,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,53.4,usage_idle,cpu,cpu0,localhost
,,252,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,65.23476523476523,usage_idle,cpu,cpu0,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,253,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest_nice,cpu,cpu4,localhost
,,253,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest_nice,cpu,cpu4,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,254,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_iowait,cpu,cpu0,localhost
,,254,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_iowait,cpu,cpu0,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,255,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_softirq,cpu,cpu4,localhost
,,255,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_softirq,cpu,cpu4,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,256,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,5.394605394605395,usage_user,cpu,cpu3,localhost
,,256,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,2,usage_user,cpu,cpu3,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,257,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_nice,cpu,cpu3,localhost
,,257,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_nice,cpu,cpu3,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,258,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_irq,cpu,cpu2,localhost
,,258,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_irq,cpu,cpu2,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,259,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_iowait,cpu,cpu1,localhost
,,259,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_iowait,cpu,cpu1,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,260,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_softirq,cpu,cpu2,localhost
,,260,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_softirq,cpu,cpu2,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,261,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_irq,cpu,cpu1,localhost
,,261,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_irq,cpu,cpu1,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,262,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,3.3,usage_system,cpu,cpu1,localhost
,,262,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,2.4975024975024973,usage_system,cpu,cpu1,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,263,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest_nice,cpu,cpu1,localhost
,,263,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest_nice,cpu,cpu1,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,264,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,0,usage_guest,cpu,cpu1,localhost
,,264,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,0,usage_guest,cpu,cpu1,localhost

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,265,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:30Z,32.2,usage_user,cpu,cpu0,localhost
,,265,2020-03-17T20:10:20.967464Z,2020-03-17T20:10:40.967464Z,2020-03-17T20:10:40Z,19.78021978021978,usage_user,cpu,cpu0,localhost
"

outData ="
#group,false,false,false
#datatype,string,long,long
#default,_result,,
,result,table,cpu
,,0,9
"

t_distinctByTag = (table=<-) =>
    table
        |> aggregate.countDistinctByTag(measurement: "cpu", tag: "cpu")

test distinctByTag = () => ({
        input: testing.loadStorage(csv: inData),
        want: testing.loadMem(csv: outData),
        fn: t_distinctByTag
})