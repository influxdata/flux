package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,6535598080,active,mem,host.local
,,0,2018-05-22T19:53:36Z,6390587392,active,mem,host.local
,,0,2018-05-22T19:53:46Z,6445174784,active,mem,host.local
,,0,2018-05-22T19:53:56Z,6387265536,active,mem,host.local
,,0,2018-05-22T19:54:06Z,6375489536,active,mem,host.local
,,0,2018-05-22T19:54:16Z,6427201536,active,mem,host.local
,,1,2018-05-22T19:53:26Z,6347390976,available,mem,host.local
,,1,2018-05-22T19:53:36Z,6405451776,available,mem,host.local
,,1,2018-05-22T19:53:46Z,6461759488,available,mem,host.local
,,1,2018-05-22T19:53:56Z,6400196608,available,mem,host.local
,,1,2018-05-22T19:54:06Z,6394032128,available,mem,host.local
,,1,2018-05-22T19:54:16Z,6448041984,available,mem,host.local

#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,2,2018-05-22T19:53:26Z,36.946678161621094,available_percent,mem,host.local
,,2,2018-05-22T19:53:36Z,37.28463649749756,available_percent,mem,host.local
,,2,2018-05-22T19:53:46Z,37.61239051818848,available_percent,mem,host.local
,,2,2018-05-22T19:53:56Z,37.25404739379883,available_percent,mem,host.local
,,2,2018-05-22T19:54:06Z,37.21816539764404,available_percent,mem,host.local
,,2,2018-05-22T19:54:16Z,37.53254413604736,available_percent,mem,host.local

#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,3,2018-05-22T19:53:26Z,0,blocked,processes,host.local
,,3,2018-05-22T19:53:36Z,0,blocked,processes,host.local
,,3,2018-05-22T19:53:46Z,0,blocked,processes,host.local
,,3,2018-05-22T19:53:56Z,2,blocked,processes,host.local
,,3,2018-05-22T19:54:06Z,0,blocked,processes,host.local
,,3,2018-05-22T19:54:16Z,0,blocked,processes,host.local
,,4,2018-05-22T19:53:26Z,0,buffered,mem,host.local
,,4,2018-05-22T19:53:36Z,0,buffered,mem,host.local
,,4,2018-05-22T19:53:46Z,0,buffered,mem,host.local
,,4,2018-05-22T19:53:56Z,0,buffered,mem,host.local
,,4,2018-05-22T19:54:06Z,0,buffered,mem,host.local
,,4,2018-05-22T19:54:16Z,0,buffered,mem,host.local
,,5,2018-05-22T19:53:26Z,0,cached,mem,host.local
,,5,2018-05-22T19:53:36Z,0,cached,mem,host.local
,,5,2018-05-22T19:53:46Z,0,cached,mem,host.local
,,5,2018-05-22T19:53:56Z,0,cached,mem,host.local
,,5,2018-05-22T19:54:06Z,0,cached,mem,host.local
,,5,2018-05-22T19:54:16Z,0,cached,mem,host.local
,,6,2018-05-22T19:53:26Z,54624256,free,mem,host.local
,,6,2018-05-22T19:53:36Z,17481728,free,mem,host.local
,,6,2018-05-22T19:53:46Z,17805312,free,mem,host.local
,,6,2018-05-22T19:53:56Z,16089088,free,mem,host.local
,,6,2018-05-22T19:54:06Z,20774912,free,mem,host.local
,,6,2018-05-22T19:54:16Z,20930560,free,mem,host.local
,,7,2018-05-22T19:53:26Z,1461714944,free,swap,host.local
,,7,2018-05-22T19:53:36Z,1494745088,free,swap,host.local
,,7,2018-05-22T19:53:46Z,1494745088,free,swap,host.local
,,7,2018-05-22T19:53:56Z,1494745088,free,swap,host.local
,,7,2018-05-22T19:54:06Z,1494745088,free,swap,host.local
,,7,2018-05-22T19:54:16Z,1491075072,free,swap,host.local
,,8,2018-05-22T19:53:26Z,0,idle,processes,host.local
,,8,2018-05-22T19:53:36Z,0,idle,processes,host.local
,,8,2018-05-22T19:53:46Z,0,idle,processes,host.local
,,8,2018-05-22T19:53:56Z,0,idle,processes,host.local
,,8,2018-05-22T19:54:06Z,0,idle,processes,host.local
,,8,2018-05-22T19:54:16Z,0,idle,processes,host.local
,,9,2018-05-22T19:53:26Z,0,in,swap,host.local
,,9,2018-05-22T19:53:36Z,0,in,swap,host.local
,,9,2018-05-22T19:53:46Z,0,in,swap,host.local
,,9,2018-05-22T19:53:56Z,0,in,swap,host.local
,,9,2018-05-22T19:54:06Z,0,in,swap,host.local
,,9,2018-05-22T19:54:16Z,0,in,swap,host.local
,,10,2018-05-22T19:53:26Z,6292766720,inactive,mem,host.local
,,10,2018-05-22T19:53:36Z,6387970048,inactive,mem,host.local
,,10,2018-05-22T19:53:46Z,6443954176,inactive,mem,host.local
,,10,2018-05-22T19:53:56Z,6384107520,inactive,mem,host.local
,,10,2018-05-22T19:54:06Z,6373257216,inactive,mem,host.local
,,10,2018-05-22T19:54:16Z,6427111424,inactive,mem,host.local
,,11,2018-05-22T19:53:26Z,0,out,swap,host.local
,,11,2018-05-22T19:53:36Z,0,out,swap,host.local
,,11,2018-05-22T19:53:46Z,0,out,swap,host.local
,,11,2018-05-22T19:53:56Z,0,out,swap,host.local
,,11,2018-05-22T19:54:06Z,0,out,swap,host.local
,,11,2018-05-22T19:54:16Z,0,out,swap,host.local
,,12,2018-05-22T19:53:26Z,3,running,processes,host.local
,,12,2018-05-22T19:53:36Z,1,running,processes,host.local
,,12,2018-05-22T19:53:46Z,1,running,processes,host.local
,,12,2018-05-22T19:53:56Z,2,running,processes,host.local
,,12,2018-05-22T19:54:06Z,2,running,processes,host.local
,,12,2018-05-22T19:54:16Z,3,running,processes,host.local
,,13,2018-05-22T19:53:26Z,414,sleeping,processes,host.local
,,13,2018-05-22T19:53:36Z,415,sleeping,processes,host.local
,,13,2018-05-22T19:53:46Z,415,sleeping,processes,host.local
,,13,2018-05-22T19:53:56Z,414,sleeping,processes,host.local
,,13,2018-05-22T19:54:06Z,415,sleeping,processes,host.local
,,13,2018-05-22T19:54:16Z,414,sleeping,processes,host.local
,,14,2018-05-22T19:53:26Z,0,stopped,processes,host.local
,,14,2018-05-22T19:53:36Z,0,stopped,processes,host.local
,,14,2018-05-22T19:53:46Z,0,stopped,processes,host.local
,,14,2018-05-22T19:53:56Z,0,stopped,processes,host.local
,,14,2018-05-22T19:54:06Z,0,stopped,processes,host.local
,,14,2018-05-22T19:54:16Z,0,stopped,processes,host.local
,,15,2018-05-22T19:53:26Z,17179869184,total,mem,host.local
,,15,2018-05-22T19:53:36Z,17179869184,total,mem,host.local
,,15,2018-05-22T19:53:46Z,17179869184,total,mem,host.local
,,15,2018-05-22T19:53:56Z,17179869184,total,mem,host.local
,,15,2018-05-22T19:54:06Z,17179869184,total,mem,host.local
,,15,2018-05-22T19:54:16Z,17179869184,total,mem,host.local
,,16,2018-05-22T19:53:26Z,417,total,processes,host.local
,,16,2018-05-22T19:53:36Z,416,total,processes,host.local
,,16,2018-05-22T19:53:46Z,416,total,processes,host.local
,,16,2018-05-22T19:53:56Z,418,total,processes,host.local
,,16,2018-05-22T19:54:06Z,418,total,processes,host.local
,,16,2018-05-22T19:54:16Z,417,total,processes,host.local
,,17,2018-05-22T19:53:26Z,8589934592,total,swap,host.local
,,17,2018-05-22T19:53:36Z,8589934592,total,swap,host.local
,,17,2018-05-22T19:53:46Z,8589934592,total,swap,host.local
,,17,2018-05-22T19:53:56Z,8589934592,total,swap,host.local
,,17,2018-05-22T19:54:06Z,8589934592,total,swap,host.local
,,17,2018-05-22T19:54:16Z,8589934592,total,swap,host.local
,,18,2018-05-22T19:53:26Z,0,unknown,processes,host.local
,,18,2018-05-22T19:53:36Z,0,unknown,processes,host.local
,,18,2018-05-22T19:53:46Z,0,unknown,processes,host.local
,,18,2018-05-22T19:53:56Z,0,unknown,processes,host.local
,,18,2018-05-22T19:54:06Z,1,unknown,processes,host.local
,,18,2018-05-22T19:54:16Z,0,unknown,processes,host.local
,,19,2018-05-22T19:53:26Z,10832478208,used,mem,host.local
,,19,2018-05-22T19:53:36Z,10774417408,used,mem,host.local
,,19,2018-05-22T19:53:46Z,10718109696,used,mem,host.local
,,19,2018-05-22T19:53:56Z,10779672576,used,mem,host.local
,,19,2018-05-22T19:54:06Z,10785837056,used,mem,host.local
,,19,2018-05-22T19:54:16Z,10731827200,used,mem,host.local
,,20,2018-05-22T19:53:26Z,7128219648,used,swap,host.local
,,20,2018-05-22T19:53:36Z,7095189504,used,swap,host.local
,,20,2018-05-22T19:53:46Z,7095189504,used,swap,host.local
,,20,2018-05-22T19:53:56Z,7095189504,used,swap,host.local
,,20,2018-05-22T19:54:06Z,7095189504,used,swap,host.local
,,20,2018-05-22T19:54:16Z,7098859520,used,swap,host.local

#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,21,2018-05-22T19:53:26Z,63.053321838378906,used_percent,mem,host.local
,,21,2018-05-22T19:53:36Z,62.71536350250244,used_percent,mem,host.local
,,21,2018-05-22T19:53:46Z,62.38760948181152,used_percent,mem,host.local
,,21,2018-05-22T19:53:56Z,62.74595260620117,used_percent,mem,host.local
,,21,2018-05-22T19:54:06Z,62.78183460235596,used_percent,mem,host.local
,,21,2018-05-22T19:54:16Z,62.46745586395264,used_percent,mem,host.local

#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,22,2018-05-22T19:53:26Z,0,zombies,processes,host.local
,,22,2018-05-22T19:53:36Z,0,zombies,processes,host.local
,,22,2018-05-22T19:53:46Z,0,zombies,processes,host.local
,,22,2018-05-22T19:53:56Z,0,zombies,processes,host.local
,,22,2018-05-22T19:54:06Z,0,zombies,processes,host.local
,,22,2018-05-22T19:54:16Z,0,zombies,processes,host.local
"

outData = "
#datatype,string,long,string,string,string,string,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,long,string
#group,false,false,true,true,true,true,true,true,false,false,false,true
#default,_result,,,,,,,,,,,
,result,table,_field_mem,_field_proc,_measurement_mem,_measurement_proc,_start,_stop,_time,_value_mem,_value_proc,host
,,0,used,total,mem,processes,2018-05-22T19:53:00Z,2018-05-22T19:55:00Z,2018-05-22T19:53:26Z,10832478208,417,host.local
,,0,used,total,mem,processes,2018-05-22T19:53:00Z,2018-05-22T19:55:00Z,2018-05-22T19:53:36Z,10774417408,416,host.local
,,0,used,total,mem,processes,2018-05-22T19:53:00Z,2018-05-22T19:55:00Z,2018-05-22T19:53:46Z,10718109696,416,host.local
,,0,used,total,mem,processes,2018-05-22T19:53:00Z,2018-05-22T19:55:00Z,2018-05-22T19:53:56Z,10779672576,418,host.local
,,0,used,total,mem,processes,2018-05-22T19:53:00Z,2018-05-22T19:55:00Z,2018-05-22T19:54:06Z,10785837056,418,host.local
,,0,used,total,mem,processes,2018-05-22T19:53:00Z,2018-05-22T19:55:00Z,2018-05-22T19:54:16Z,10731827200,417,host.local
"

t_join = (table=<-) => {
    mem = table
        |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:55:00Z)
        |> filter(fn: (r) => r._measurement == "mem" and r._field == "used")

    proc = table
        |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:55:00Z)
        |> filter(fn: (r) => r._measurement == "processes" and r._field == "total")

    return join(tables: {mem: mem, proc: proc}, on: ["_time", "_stop", "_start", "host"])
}

test _join = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_join})
