package universe_test


import "array"
import "csv"
import "testing"
import "testing/expect"
import "internal/debug"

option now = () => 2030-01-01T00:00:00Z

testcase basic {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,100,load1,system,host.local
,,0,2018-05-22T19:53:36Z,101,load1,system,host.local
,,0,2018-05-22T19:53:46Z,102,load1,system,host.local
"
    outData =
        "
#datatype,string,long,double
#group,false,false,false
#default,_result,,
,result,table,newValue
,,0,100.0
,,0,101.0
,,0,102.0
"

    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z)
            |> map(fn: (r) => ({newValue: float(v: r._value)}))
    want = csv.from(csv: outData)

    testing.diff(want: want, got: got) |> yield()
}

testcase nulls_map {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,string,long,string
#group,false,false,false,true,false,true
#default,_result,,,,,
,result,table,_time,_field,_value,_measurement
,,0,2018-05-22T19:53:26Z,a,1,aa
,,0,2018-05-22T19:53:36Z,a,1,aa
,,0,2018-05-22T19:53:46Z,a,1,aa
,,1,2018-05-22T19:53:36Z,b,1,aa
,,1,2018-05-22T19:53:46Z,b,1,aa
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,long
#group,false,false,false,false
#default,0,,,
,result,table,_time,_value
,,0,2018-05-22T19:53:26Z,
,,0,2018-05-22T19:53:36Z,1
,,0,2018-05-22T19:53:46Z,1
"

    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z)
            |> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
            |> map(fn: (r) => ({_time: r._time, _value: r.a / r.b}))
    want = csv.from(csv: outData)

    testing.diff(want: want, got: got) |> yield()
}

testcase missing_column {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,string,long,string
#group,false,false,false,true,false,true
#default,_result,,,,,
,result,table,_time,_field,_value,_measurement
,,0,2018-05-22T19:53:26Z,a,1,aa
,,0,2018-05-22T19:53:36Z,a,1,aa
,,0,2018-05-22T19:53:46Z,a,1,aa
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,long
#group,false,false,false,false
#default,0,,,
,result,table,_time,_value
,,0,2018-05-22T19:53:26Z,1
,,0,2018-05-22T19:53:36Z,1
,,0,2018-05-22T19:53:46Z,1
"

    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z)
            |> map(fn: (r) => ({_time: r._time, _value: r._value, key: r.key}))
    want = csv.from(csv: outData)

    testing.diff(want: want, got: got) |> yield()
}

testcase local_var {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,a ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:36Z,k9n  gm ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:46Z,b  ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:56Z,2COTDe   ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:54:06Z,cLnSkNMI ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:54:16Z,13F2,used_percent,disk,disk1,apfs,host.local,/
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,const,used_percent,disk,disk1,apfs,host.local,/
"

    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z)
            |> map(
                fn: (r) => {
                    myVal = "const"

                    return {r with _value: myVal}
                },
            )
    want = csv.from(csv: outData)

    testing.diff(want: want, got: got) |> yield()
}

testcase shadow_var {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,a ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:36Z,k9n  gm ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:46Z,b  ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:56Z,2COTDe   ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:54:06Z,cLnSkNMI ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:54:16Z,13F2,used_percent,disk,disk1,apfs,host.local,/
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,const,used_percent,disk,disk1,apfs,host.local,/
"

    myVal = "wrong"
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z)
            |> map(
                fn: (r) => {
                    myVal = "const"

                    return {r with _value: myVal}
                },
            )
    want = csv.from(csv: outData)

    testing.diff(want: want, got: got) |> yield()
}

testcase polymorphism {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,long,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-22T19:53:26Z,49,load1,system
,,0,2018-05-22T19:53:36Z,50,load1,system
,,0,2018-05-22T19:53:46Z,51,load1,system
"
    outData =
        "
#datatype,string,long,string
#group,false,false,false
#default,_result,,
,result,table,out
,,0,Y
,,0,N
,,0,N
"
    f = (r) => r._value < 50
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:16Z)
            |> map(fn: (r) => ({out: if f(r: r) then "Y" else "N"}))
    want = csv.from(csv: outData)

    testing.diff(want: want, got: got) |> yield()
}

testcase extension_with {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,100,load1,system,host.local
,,0,2018-05-22T19:53:36Z,101,load1,system,host.local
,,0,2018-05-22T19:53:46Z,102,load1,system,host.local
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,100.0,load1,system,host.local
,,0,2018-05-22T19:53:36Z,101.0,load1,system,host.local
,,0,2018-05-22T19:53:46Z,102.0,load1,system,host.local
"

    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T00:00:00Z)
            |> drop(columns: ["_start", "_stop"])
            |> map(fn: (r) => ({r with _value: float(v: r._value)}))
    want = csv.from(csv: outData)

    testing.diff(want: want, got: got) |> yield()
}

testcase extern_var {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,a ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:36Z,k9n  gm ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:46Z,b  ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:56Z,2COTDe   ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:54:06Z,cLnSkNMI ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:54:16Z,13F2,used_percent,disk,disk1,apfs,host.local,/
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,const,used_percent,disk,disk1,apfs,host.local,/
"
    myVal = "const"

    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z)
            |> map(
                fn: (r) => {
                    return {r with _value: myVal}
                },
            )
    want = csv.from(csv: outData)

    testing.diff(want: want, got: got) |> yield()
}

testcase extern_dynamic_var {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,a ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:36Z,k9n  gm ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:46Z,b  ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:56Z,2COTDe   ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:54:06Z,cLnSkNMI ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:54:16Z,13F2,used_percent,disk,disk1,apfs,host.local,/
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,const,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,const,used_percent,disk,disk1,apfs,host.local,/
"
    myVal = () => "const"

    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z)
            |> map(
                fn: (r) => {
                    return {r with _value: myVal()}
                },
            )
    want = csv.from(csv: outData)

    testing.diff(want: want, got: got) |> yield()
}

testcase with_obj {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,100,load1,system,host.local
,,0,2018-05-22T19:53:36Z,101,load1,system,host.local
,,0,2018-05-22T19:53:46Z,102,load1,system,host.local
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,long,long,long,long,long,string,dateTime:RFC3339,long
#group,false,false,true,true,true,true,true,false,false,false,false,false,false,false,false,false
#default,got,,,,,,,,,,,,,,,
,result,table,_start,_stop,_field,_measurement,host,_time,_value,array,boolAdd,floatAdd,intAdd,string,time,uintAdd
,,0,1947-11-13T00:00:00Z,2030-01-01T00:00:00Z,load1,system,host.local,2018-05-22T19:53:26Z,100,101,101,101,99,1,2018-05-22T19:53:26Z,101
,,0,1947-11-13T00:00:00Z,2030-01-01T00:00:00Z,load1,system,host.local,2018-05-22T19:53:36Z,101,102,102,102,100,1,2018-05-22T19:53:26Z,102
,,0,1947-11-13T00:00:00Z,2030-01-01T00:00:00Z,load1,system,host.local,2018-05-22T19:53:46Z,102,103,103,103,101,1,2018-05-22T19:53:26Z,103
"
    obj = {
        b: true,
        i: -1,
        d: 1.0,
        u: 1,
        s: "1",
        t: 2018-05-22T19:53:26Z,
        r: -30000d,
    }
    arr = [1, 2, 3, 4]

    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: obj.r)
            |> map(
                fn: (r) =>
                    ({r with
                        boolAdd: int(v: obj.b) + r._value,
                        intAdd: obj.i + r._value,
                        floatAdd: int(v: obj.d) + r._value,
                        uintAdd: int(v: obj.u) + r._value,
                        string: obj.s,
                        time: obj.t,
                        array: arr[0] + r._value,
                    }),
            )
    want = csv.from(csv: outData)

    testing.diff(want: want, got: got) |> yield()
}

testcase field_type_change {
    inData =
        "
#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,m,f,2018-01-01T00:00:00Z,2
"
    outData =
        "
#datatype,string,long,string,string,dateTime:RFC3339,string
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,m,f,2018-01-01T00:00:00Z,hello
"

    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-01-01T00:00:00Z)
            |> drop(columns: ["_start", "_stop"])
            // establish _value as a double column in output
            |> map(fn: (r) => ({r with _value: 2.0}))
            // convert to a string
            |> map(fn: (r) => ({r with _value: "hello"}))
            // previously this would produce an error
            |> filter(fn: (r) => r._value == "hello")
    want = csv.from(csv: outData)

    testing.diff(want: want, got: got) |> yield()
}

testcase vectorize_addition_operator {
    inData =
        "
#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,m,a,2018-01-01T00:00:00Z,1
,,0,m,a,2018-01-02T00:00:00Z,2
,,0,m,a,2018-01-03T00:00:00Z,3

#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,m,b,2018-01-01T00:00:00Z,3
,,0,m,b,2018-01-02T00:00:00Z,4
,,0,m,b,2018-01-03T00:00:00Z,5


#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,n,a,2018-01-04T00:00:00Z,10
,,0,n,a,2018-01-05T00:00:00Z,20
,,0,n,a,2018-01-06T00:00:00Z,30

#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,n,b,2018-01-04T00:00:00Z,30
,,0,n,b,2018-01-05T00:00:00Z,40
,,0,n,b,2018-01-06T00:00:00Z,50
"
    outData =
        "
#datatype,string,long,string,dateTime:RFC3339,double,double,double
#group,false,false,true,false,false,false,false
#default,_result,,,,,,
,result,table,_measurement,_time,a,b,x
,,0,m,2018-01-01T00:00:00Z,1,3,4
,,0,m,2018-01-02T00:00:00Z,2,4,6
,,0,m,2018-01-03T00:00:00Z,3,5,8

#datatype,string,long,string,dateTime:RFC3339,double,double,double
#group,false,false,true,false,false,false,false
#default,_result,,,,,,
,result,table,_measurement,_time,a,b,x
,,0,n,2018-01-04T00:00:00Z,10,30,40
,,0,n,2018-01-05T00:00:00Z,20,40,60
,,0,n,2018-01-06T00:00:00Z,30,50,80
"

    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-01-01T00:00:00Z, stop: 2018-01-07T00:00:00Z)
            |> drop(columns: ["_start", "_stop"])
            |> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
            |> map(fn: (r) => ({r with x: r.a + r.b}))
    want = csv.from(csv: outData)

    testing.diff(want: want, got: got) |> yield()
}

testcase vectorize_and_operator {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want = array.from(rows: [{a: true, b: false, c: false}])

    got =
        array.from(rows: [{a: true, b: false}])
            |> map(fn: (r) => ({r with c: r.a and r.b}))

    testing.diff(want: want, got: got) |> yield()
}

testcase vectorize_or_operator {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want = array.from(rows: [{a: true, b: false, c: true}])

    got =
        array.from(rows: [{a: true, b: false}])
            |> map(fn: (r) => ({r with c: r.a or r.b}))

    testing.diff(want: want, got: got) |> yield()
}

testcase vectorize_nested_logicals {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want = array.from(rows: [{a: true, b: false, c: true}])

    got =
        array.from(rows: [{a: true, b: false}])
            |> map(fn: (r) => ({r with c: r.a and r.b or r.a}))

    testing.diff(want: want, got: got)
}

testcase logical_typed_null_vectorized {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want =
        array.from(
            rows: [
                {
                    nullOrFalse: debug.null(type: "bool"),
                    falseOrNull: debug.null(type: "bool"),
                    nullOrTrue: true,
                    trueOrNull: true,
                    nullAndFalse: false,
                    falseAndNull: false,
                    nullAndTrue: debug.null(type: "bool"),
                    trueAndNull: debug.null(type: "bool"),
                    nullAndNull: debug.null(type: "bool"),
                    nullOrNull: debug.null(type: "bool"),
                },
            ],
        )

    got =
        array.from(rows: [{_true: true, _false: false, _null: debug.null(type: "bool")}])
            |> map(
                fn: (r) =>
                    ({r with
                        nullOrFalse: r._null or r._false,
                        falseOrNull: r._false or r._null,
                        nullAndFalse: r._null and r._false,
                        falseAndNull: r._false and r._null,
                        nullOrTrue: r._null or r._true,
                        trueOrNull: r._true or r._null,
                        nullAndTrue: r._null and r._true,
                        trueAndNull: r._true and r._null,
                        nullAndNull: r._null and r._null,
                        nullOrNull: r._null or r._null,
                    }),
            )
            |> drop(columns: ["_null", "_true", "_false"])

    testing.diff(want: want, got: got)
}

testcase logical_untyped_null_vectorized {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want =
        array.from(
            // N.b. the "untyped" nulls are dropped silently by map() when it produces its output.
            rows: [{nullOrTrue: true, trueOrNull: true, nullAndFalse: false, falseAndNull: false}],
        )
            |> debug.opaque()

    got =
        array.from(rows: [{_true: true, _false: false}])
            // N.b. The "untyped" null here is introduced by hiding the type
            // information from flux via `debug.opaque()` then accessing the
            // undefined field `_null` inside our map func.
            |> debug.opaque()
            |> map(
                fn: (r) =>
                    ({r with
                        nullOrFalse: r._null or r._false,
                        falseOrNull: r._false or r._null,
                        nullAndFalse: r._null and r._false,
                        falseAndNull: r._false and r._null,
                        nullOrTrue: r._null or r._true,
                        trueOrNull: r._true or r._null,
                        nullAndTrue: r._null and r._true,
                        trueAndNull: r._true and r._null,
                        nullAndNull: r._null and r._null,
                        nullOrNull: r._null or r._null,
                    }),
            )
            |> drop(columns: ["_true", "_false"])

    testing.diff(want: want, got: got)
}

testcase logical_typed_null_vectorized_const {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want =
        array.from(
            rows: [
                {
                    nullOrFalse: debug.null(type: "bool"),
                    falseOrNull: debug.null(type: "bool"),
                    nullOrTrue: true,
                    trueOrNull: true,
                    nullAndFalse: false,
                    falseAndNull: false,
                    nullAndTrue: debug.null(type: "bool"),
                    trueAndNull: debug.null(type: "bool"),
                    nullAndNull: debug.null(type: "bool"),
                    nullOrNull: debug.null(type: "bool"),
                },
            ],
        )

    got =
        array.from(rows: [{x: 0, _null: debug.null(type: "bool")}])
            |> map(
                fn: (r) =>
                    ({r with
                        nullOrFalse: r._null or false,
                        falseOrNull: false or r._null,
                        nullAndFalse: r._null and false,
                        falseAndNull: false and r._null,
                        nullOrTrue: r._null or true,
                        trueOrNull: true or r._null,
                        nullAndTrue: r._null and true,
                        trueAndNull: true and r._null,
                        nullAndNull: r._null and r._null,
                        nullOrNull: r._null or r._null,
                    }),
            )
            |> drop(columns: ["x", "_null"])

    testing.diff(want: want, got: got)
}

testcase logical_untyped_null_vectorized_const {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want =
        array.from(
            // N.b. the "untyped" nulls are dropped silently by map() when it produces its output.
            rows: [{nullOrTrue: true, trueOrNull: true, nullAndFalse: false, falseAndNull: false}],
        )
            |> debug.opaque()

    got =
        array.from(rows: [{x: 0}])
            // N.b. The "untyped" null here is introduced by hiding the type
            // information from flux via `debug.opaque()` then accessing the
            // undefined field `_null` inside our map func.
            |> debug.opaque()
            |> map(
                fn: (r) =>
                    ({r with
                        nullOrFalse: r._null or false,
                        falseOrNull: false or r._null,
                        nullAndFalse: r._null and false,
                        falseAndNull: false and r._null,
                        nullOrTrue: r._null or true,
                        trueOrNull: true or r._null,
                        nullAndTrue: r._null and true,
                        trueAndNull: true and r._null,
                        nullAndNull: r._null and r._null,
                        nullOrNull: r._null or r._null,
                    }),
            )
            |> drop(columns: ["x"])

    testing.diff(want: want, got: got)
}

testcase logical_typed_null {
    // FIXME: disabled while looking at perf regression
    option testing.tags = ["skip"]

    expect.planner(rules: ["vectorizeMapRule": 0])

    want =
        array.from(
            rows: [
                {
                    nullOrFalse: debug.null(type: "bool"),
                    falseOrNull: debug.null(type: "bool"),
                    nullOrTrue: true,
                    trueOrNull: true,
                    nullAndFalse: false,
                    falseAndNull: false,
                    nullAndTrue: debug.null(type: "bool"),
                    trueAndNull: debug.null(type: "bool"),
                    nullAndNull: debug.null(type: "bool"),
                    nullOrNull: debug.null(type: "bool"),
                },
            ],
        )

    got =
        array.from(rows: [{_true: true, _false: false, _null: debug.null(type: "bool")}])
            |> map(
                fn: (r) =>
                    ({r with
                        nullOrFalse: r._null or r._false,
                        falseOrNull: r._false or r._null,
                        nullAndFalse: r._null and r._false,
                        falseAndNull: r._false and r._null,
                        nullOrTrue: r._null or r._true,
                        trueOrNull: r._true or r._null,
                        nullAndTrue: r._null and r._true,
                        trueAndNull: r._true and r._null,
                        nullAndNull: r._null and r._null,
                        nullOrNull: r._null or r._null,
                        // forces vectorization to fail
                        _drop: (() => true)(),
                    }),
            )
            |> drop(columns: ["_drop", "_null", "_true", "_false"])

    testing.diff(want: want, got: got)
}

testcase logical_untyped_null {
    // FIXME: disabled while looking at perf regression
    option testing.tags = ["skip"]

    expect.planner(rules: ["vectorizeMapRule": 0])

    want =
        array.from(
            // N.b. the "untyped" nulls are dropped silently by map() when it produces its output.
            rows: [{nullOrTrue: true, trueOrNull: true, nullAndFalse: false, falseAndNull: false}],
        )
            |> debug.opaque()

    got =
        array.from(rows: [{_true: true, _false: false}])
            // N.b. The "untyped" null here is introduced by hiding the type
            // information from flux via `debug.opaque()` then accessing the
            // undefined field `_null` inside our map func.
            |> debug.opaque()
            |> map(
                fn: (r) =>
                    ({r with
                        nullOrFalse: r._null or r._false,
                        falseOrNull: r._false or r._null,
                        nullAndFalse: r._null and r._false,
                        falseAndNull: r._false and r._null,
                        nullOrTrue: r._null or r._true,
                        trueOrNull: r._true or r._null,
                        nullAndTrue: r._null and r._true,
                        trueAndNull: r._true and r._null,
                        nullAndNull: r._null and r._null,
                        nullOrNull: r._null or r._null,
                        // forces vectorization to fail
                        _drop: (() => true)(),
                    }),
            )
            |> drop(columns: ["_drop", "_true", "_false"])

    testing.diff(want: want, got: got)
}

// XXX(onelson): this test doesn't actually use `map`, but is so similar to the
// other logical expr tests! I'm sorry!
// FIXME: can't figure out how to write this test in a way that will work for untyped nulls.
//   In `map()`, the untyped nulls are quietly dropped from the output tables.
//   In this situation, the untyped nulls produce an error:
//   ```
//   cannot represent the type null as column data
//   ```
testcase logical_typed_null_interp {
    r = {_true: true, _false: false, _null: debug.null(type: "bool")}

    want =
        array.from(
            rows: [
                {
                    nullOrFalse: debug.null(type: "bool"),
                    falseOrNull: debug.null(type: "bool"),
                    nullOrTrue: true,
                    trueOrNull: true,
                    nullAndFalse: false,
                    falseAndNull: false,
                    nullAndTrue: debug.null(type: "bool"),
                    trueAndNull: debug.null(type: "bool"),
                    nullAndNull: debug.null(type: "bool"),
                    nullOrNull: debug.null(type: "bool"),
                },
            ],
        )

    got =
        array.from(
            rows: [
                {
                    nullOrFalse: r._null or r._false,
                    falseOrNull: r._false or r._null,
                    nullAndFalse: r._null and r._false,
                    falseAndNull: r._false and r._null,
                    nullOrTrue: r._null or r._true,
                    trueOrNull: r._true or r._null,
                    nullAndTrue: r._null and r._true,
                    trueAndNull: r._true and r._null,
                    nullAndNull: r._null and r._null,
                    nullOrNull: r._null or r._null,
                },
            ],
        )

    testing.diff(want: want, got: got)
}

testcase vectorize_with_operator_overwrite_attribute {
    expect.planner(rules: ["vectorizeMapRule": 1])

    got =
        array.from(rows: [{x: 1, y: "a"}])
            |> map(fn: (r) => ({r with x: r.x}))
            |> drop(columns: ["y"])

    want = array.from(rows: [{x: 1}])

    testing.diff(got, want)
}

testcase vectorize_div_by_zero_int {
    expect.planner(rules: ["vectorizeMapRule": 1])

    fn = () =>
        array.from(rows: [{a: 1, b: 0}])
            |> map(fn: (r) => ({r with x: r.a / r.b}))
            |> tableFind(fn: (key) => true)
            |> findRecord(fn: (key) => true, idx: 0)

    testing.shouldError(fn: fn, want: /cannot divide by zero/)
}

testcase vectorize_div_by_zero_uint {
    expect.planner(rules: ["vectorizeMapRule": 1])

    fn = () =>
        array.from(rows: [{a: uint(v: 1), b: uint(v: 0)}])
            |> map(fn: (r) => ({r with x: r.a / r.b}))
            |> tableFind(fn: (key) => true)
            |> findRecord(fn: (key) => true, idx: 0)

    testing.shouldError(fn: fn, want: /cannot divide by zero/)
}

testcase vectorize_div_by_zero_int_const {
    expect.planner(rules: ["vectorizeMapRule": 1])

    fn = () =>
        array.from(rows: [{a: 1}])
            |> map(fn: (r) => ({r with x: r.a / 0}))
            |> tableFind(fn: (key) => true)
            |> findRecord(fn: (key) => true, idx: 0)

    testing.shouldError(fn: fn, want: /cannot divide by zero/)
}

testcase vectorize_div_by_zero_int_const_const {
    expect.planner(rules: ["vectorizeMapRule": 1])

    fn = () =>
        array.from(rows: [{a: 1}])
            |> map(fn: (r) => ({r with x: 1 / 0}))
            |> tableFind(fn: (key) => true)
            |> findRecord(fn: (key) => true, idx: 0)

    testing.shouldError(fn: fn, want: /cannot divide by zero/)
}
