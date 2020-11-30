package dict_test

import "testing"
import "dict"

option now = () => (2030-01-01T00:00:00Z)

codes = dict.fromList(pairs: [
  {key: "internal", value: 0},
  {key: "invalid", value: 1},
  {key: "unimplemented", value: 2},
])

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,string
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,error_type,_value
,,0,2018-05-22T19:53:26Z,requests,error,internal,some internal error
,,0,2018-05-22T19:53:36Z,requests,error,internal,another internal error
,,1,2018-05-22T19:53:46Z,requests,error,invalid,unknown parameter
,,1,2018-05-22T19:53:56Z,requests,error,invalid,cannot use duration as value
,,2,2018-05-22T19:54:06Z,requests,error,unimplemented,implement me
,,2,2018-05-22T19:54:16Z,requests,error,unimplemented,not implemented
,,3,2018-05-22T19:53:26Z,requests,error,unknown,unknown error
,,3,2018-05-22T19:53:36Z,requests,error,unknown,network error
"

outData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,string,long
#group,false,false,false,true,true,true,false,false
#default,_result,,,,,,,
,result,table,_time,_measurement,_field,error_type,_value,error_code
,,0,2018-05-22T19:53:26Z,requests,error,internal,some internal error,0
,,0,2018-05-22T19:53:36Z,requests,error,internal,another internal error,0
,,1,2018-05-22T19:53:46Z,requests,error,invalid,unknown parameter,1
,,1,2018-05-22T19:53:56Z,requests,error,invalid,cannot use duration as value,1
,,2,2018-05-22T19:54:06Z,requests,error,unimplemented,implement me,2
,,2,2018-05-22T19:54:16Z,requests,error,unimplemented,not implemented,2
,,3,2018-05-22T19:53:26Z,requests,error,unknown,unknown error,-1
,,3,2018-05-22T19:53:36Z,requests,error,unknown,network error,-1
"

t_dict = (table=<-) =>
  table
  |> range(start: 2018-05-22T19:53:26Z)
  |> drop(columns: ["_start", "_stop"])
  |> map(fn: (r) => {
    error_code = dict.get(dict: codes, key: r.error_type, default: -1)
    return {r with error_code: error_code}
  })

test _dict = () => ({
  input: testing.loadStorage(csv: inData),
  want: testing.loadMem(csv: outData),
  fn: t_dict,
})
