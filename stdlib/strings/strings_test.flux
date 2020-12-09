package strings_test

import "testing"
import "strings"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,字,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:36Z,k9n  gm,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:46Z,b  ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:56Z,2COTDe,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:54:06Z,cLnSkNMI  ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:54:16Z,13F2  ,used_percent,disk,disk1,apfs,host.local,/
"

testcase string_trim {
    outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,字,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,k9n  gm,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,b,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,2COTDe,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,cLnSkNMI,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,13F2,used_percent,disk,disk1,apfs,host.local,/
"
    result = testing.loadStorage(csv: inData)
		|> range(start: 2018-05-22T19:53:26Z)
		|> map(fn: (r) =>
        			({r with _value: strings.trimSpace(v: r._value)}))
    testing.diff(got: result, want: testing.loadMem(csv: outData))
}

testcase string_toUpper {
    outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,字,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,K9N  GM,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,B  ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,2COTDE,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,CLNSKNMI  ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,13F2  ,used_percent,disk,disk1,apfs,host.local,/
"
    result = testing.loadStorage(csv: inData)
		|> range(start: 2018-05-22T19:53:26Z)
		|> map(fn: (r) =>
        			({r with _value: strings.toUpper(v: r._value)}))
    testing.diff(got: result, want: testing.loadMem(csv: outData))
}

testcase string_toLower {
    outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,字,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,k9n  gm,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,b  ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,2cotde,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,clnsknmi  ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,13f2  ,used_percent,disk,disk1,apfs,host.local,/
"
    result = testing.loadStorage(csv: inData)
		|> range(start: 2018-05-22T19:53:26Z)
		|> map(fn: (r) =>
        			({r with _value: strings.toLower(v: r._value)}))
    testing.diff(got: result, want: testing.loadMem(csv: outData))
}

testcase string_title {
    outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,字,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,K9n  Gm,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,B  ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,2COTDe,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,CLnSkNMI  ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,13F2  ,used_percent,disk,disk1,apfs,host.local,/
"
    result = testing.loadStorage(csv: inData)
		|> range(start: 2018-05-22T19:53:26Z)
		|> map(fn: (r) =>
        			({r with _value: strings.title(v: r._value)}))
    testing.diff(got: result, want: testing.loadMem(csv: outData))
}

testcase string_subset {
    outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,字,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,k,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,b,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,2,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,c,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,1,used_percent,disk,disk1,apfs,host.local,/
"
    result = testing.loadStorage(csv: inData)
		|> range(start: 2018-05-22T19:53:26Z)
		|> map(fn: (r) =>
        			({r with _value: strings.substring(v: r._value, start: 0, end: 1)}))
    testing.diff(got: result, want: testing.loadMem(csv: outData))
}

testcase string_subset {
    outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,字,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,k,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,b,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,2,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,c,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,1,used_percent,disk,disk1,apfs,host.local,/
"
    result = testing.loadStorage(csv: inData)
		|> range(start: 2018-05-22T19:53:26Z)
		|> map(fn: (r) =>
        			({r with _value: strings.substring(v: r._value, start: 0, end: 1)}))
    testing.diff(got: result, want: testing.loadMem(csv: outData))
}

testcase string_replaceAll {
    outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,字,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,k9ngm,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,b,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,2COTDe,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,cLnSkNMI,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,13F2,used_percent,disk,disk1,apfs,host.local,/
"
    result = testing.loadStorage(csv: inData)
		|> range(start: 2018-05-22T19:53:26Z)
		|> map(fn: (r) => ({r with _value: strings.replaceAll(v: r._value, t: " ", u: "")}))
    testing.diff(got: result, want: testing.loadMem(csv: outData))
}

testcase string_replace {
    outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,字,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,k9n  gm,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,b  ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,2COTDe,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,c NMI  ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,13F2  ,used_percent,disk,disk1,apfs,host.local,/
"
    result = testing.loadStorage(csv: inData)
		|> range(start: 2018-05-22T19:53:26Z)
		|> map(fn: (r) =>
        			({r with _value: strings.replace(v: r._value, t: "LnSk", u: " ", i: 1)}))
    testing.diff(got: result, want: testing.loadMem(csv: outData))
}

testcase string_substring {
    outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true,false
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path,sub
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,字,used_percent,disk,disk1,apfs,host.local,/,字
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,k9n  gm,used_percent,disk,disk1,apfs,host.local,/,m
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,b  ,used_percent,disk,disk1,apfs,host.local,/, 
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,2COTDe,used_percent,disk,disk1,apfs,host.local,/,e
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,cLnSkNMI  ,used_percent,disk,disk1,apfs,host.local,/, 
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,13F2  ,used_percent,disk,disk1,apfs,host.local,/, 
"
    result = testing.loadStorage(csv: inData)
		|> range(start: 2018-05-22T19:53:26Z)
		|> map(fn: (r) =>
        			({r with sub: strings.substring(v: r._value, start: strings.strlen(v: r._value)-1, end: strings.strlen(v: r._value))}))
    testing.diff(got: result, want: testing.loadMem(csv: outData))
}

testcase string_length {
    outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string,long
#group,false,false,true,true,false,false,true,true,true,true,true,true,false
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path,len
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,字,used_percent,disk,disk1,apfs,host.local,/,1
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,k9n  gm,used_percent,disk,disk1,apfs,host.local,/,7
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,b  ,used_percent,disk,disk1,apfs,host.local,/,3
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,2COTDe,used_percent,disk,disk1,apfs,host.local,/,6
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,cLnSkNMI  ,used_percent,disk,disk1,apfs,host.local,/,10
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,13F2  ,used_percent,disk,disk1,apfs,host.local,/,6
"
    result = testing.loadStorage(csv: inData)
		|> range(start: 2018-05-22T19:53:26Z)
		|> map(fn: (r) =>
        			({r with len: strings.strlen(v: r._value)}))
    testing.diff(got: result, want: testing.loadMem(csv: outData))
}
