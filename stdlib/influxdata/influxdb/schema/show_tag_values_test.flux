package schema_test


import "testing"

option now = () => 2030-01-01T00:00:00Z

input = "
#group,false,false,false,false,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,mode,path
,,0,2020-10-21T20:48:30Z,4881964326,inodes_free,disk,disk1s5,apfs,euterpe.local,ro,/
,,0,2020-10-21T20:48:40Z,4881964326,inodes_free,disk,disk1s5,apfs,euterpe.local,ro,/
,,0,2020-10-21T20:48:50Z,4881964326,inodes_free,disk,disk1s5,apfs,euterpe.local,ro,/
,,1,2020-10-21T20:48:30Z,4294963701,inodes_free,disk,disk2s1,hfs,euterpe.local,ro,/Volumes/IntelliJ IDEA CE
,,1,2020-10-21T20:48:40Z,4294963701,inodes_free,disk,disk2s1,hfs,euterpe.local,ro,/Volumes/IntelliJ IDEA CE
,,1,2020-10-21T20:48:50Z,4294963701,inodes_free,disk,disk2s1,hfs,euterpe.local,ro,/Volumes/IntelliJ IDEA CE
,,2,2020-10-21T20:48:30Z,488514,inodes_used,disk,disk1s5,apfs,euterpe.local,ro,/
,,2,2020-10-21T20:48:40Z,488514,inodes_used,disk,disk1s5,apfs,euterpe.local,ro,/
,,2,2020-10-21T20:48:50Z,488514,inodes_used,disk,disk1s5,apfs,euterpe.local,ro,/
,,3,2020-10-21T20:48:30Z,3578,inodes_used,disk,disk2s1,hfs,euterpe.local,ro,/Volumes/IntelliJ IDEA CE
,,3,2020-10-21T20:48:40Z,3578,inodes_used,disk,disk2s1,hfs,euterpe.local,ro,/Volumes/IntelliJ IDEA CE
,,3,2020-10-21T20:48:50Z,3578,inodes_used,disk,disk2s1,hfs,euterpe.local,ro,/Volumes/IntelliJ IDEA CE

#group,false,false,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,cpu,host
,,4,2020-10-21T20:48:30Z,69.30000000167638,usage_idle,cpu,cpu0,euterpe.local
,,4,2020-10-21T20:48:40Z,67.36736736724372,usage_idle,cpu,cpu0,euterpe.local
,,4,2020-10-21T20:48:50Z,69.23076923005354,usage_idle,cpu,cpu0,euterpe.local
,,5,2020-10-21T20:48:30Z,96.10000000102445,usage_idle,cpu,cpu1,euterpe.local
,,5,2020-10-21T20:48:40Z,95.70000000055181,usage_idle,cpu,cpu1,euterpe.local
,,5,2020-10-21T20:48:50Z,95.89999999860534,usage_idle,cpu,cpu1,euterpe.local

#group,false,false,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,cpu,host
,,6,2020-10-21T20:48:30Z,69.30000000167638,usage_idle,cpu,cpu0,mnemosyne.local
,,6,2020-10-21T20:48:40Z,67.36736736724372,usage_idle,cpu,cpu0,mnemosyne.local
,,6,2020-10-21T20:48:50Z,69.23076923005354,usage_idle,cpu,cpu0,mnemosyne.local
,,7,2020-10-21T20:48:30Z,96.10000000102445,usage_idle,cpu,cpu1,mnemosyne.local
,,7,2020-10-21T20:48:40Z,95.70000000055181,usage_idle,cpu,cpu1,mnemosyne.local
,,7,2020-10-21T20:48:50Z,95.89999999860534,usage_idle,cpu,cpu1,mnemosyne.local

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,string,string,dateTime:RFC3339,double,string,string,string
#default,_result,,,,,,,,
,result,table,_field,_measurement,_time,_value,cpu,host,region
,,8,usage_user,cpu,2020-10-21T20:48:30Z,19.30000000007567,cpu0,euterpe.local,north
,,8,usage_user,cpu,2020-10-21T20:48:40Z,20.020020020038682,cpu0,euterpe.local,north
,,8,usage_user,cpu,2020-10-21T20:48:50Z,18.581418581407107,cpu0,euterpe.local,north
,,9,usage_user,cpu,2020-10-21T20:48:30Z,2.3000000000138243,cpu1,euterpe.local,north
,,9,usage_user,cpu,2020-10-21T20:48:40Z,2.4000000000536965,cpu1,euterpe.local,north
,,9,usage_user,cpu,2020-10-21T20:48:50Z,2.0999999999423746,cpu1,euterpe.local,north
"
output = "
#datatype,string,long,string
#group,false,false,false
#default,0,,
,result,table,_value
,,0,euterpe.local
,,0,mnemosyne.local
"
show_tag_values_fn = (tables=<-) => tables
    |> range(start: 2018-01-01T00:00:00Z)
    |> filter(fn: (r) => true)
    |> keep(columns: ["host"])
    |> group()
    |> distinct(column: "host")
    |> sort()

test show_tag_values = () => ({input: testing.loadStorage(csv: input), want: testing.loadMem(csv: output), fn: show_tag_values_fn})
