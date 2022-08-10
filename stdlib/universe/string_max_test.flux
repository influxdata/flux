package universe_test


import "testing"
import "csv"

inData =
    "
#datatype,string,long,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:54:16Z,13F2,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:56Z,2COTDe,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:54:06Z,cLnSkNMI,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,a,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:46Z,b,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:36Z,k9ngm,used_percent,disk,disk1,apfs,host.local,/
"

testcase string_max {
    testing.shouldError(
        fn: () =>
            csv.from(csv: inData)
                |> max()
                |> tableFind(fn: (key) => true),
        want:
            "error calling function \"tableFind\" @stdlib/universe/string_max_test.flux|26:20-26:48: runtime error @stdlib/universe/string_max_test.flux|25:20-25:25: max: invalid use of function: *universe.MaxSelector has no implementation for type string",
    )
}
