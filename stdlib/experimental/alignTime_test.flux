package experimental_test


import "testing"
import "experimental"
import "csv"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,country
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-02-22T00:00:00Z,76369,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-02-23T00:00:00Z,77016,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-02-24T00:00:00Z,77234,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-02-25T00:00:00Z,77749,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-02-26T00:00:00Z,78159,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-02-27T00:00:00Z,78598,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-02-28T00:00:00Z,78927,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-02-29T00:00:00Z,79355,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-01T00:00:00Z,79929,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-02T00:00:00Z,80134,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-03T00:00:00Z,80261,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-04T00:00:00Z,80380,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-05T00:00:00Z,80497,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-06T00:00:00Z,80667,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-07T00:00:00Z,80768,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-08T00:00:00Z,80814,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-09T00:00:00Z,80859,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-10T00:00:00Z,80879,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-11T00:00:00Z,80908,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-12T00:00:00Z,80932,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-13T00:00:00Z,80954,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-14T00:00:00Z,80973,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-15T00:00:00Z,80995,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-16T00:00:00Z,81020,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-17T00:00:00Z,81130,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-18T00:00:00Z,81163,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-19T00:00:00Z,81238,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-20T00:00:00Z,81337,total_cases,covid-19,China
,,0,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-21T00:00:00Z,81416,total_cases,covid-19,China

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,country
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-02-24T00:00:00Z,132,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-02-25T00:00:00Z,229,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-02-26T00:00:00Z,322,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-02-27T00:00:00Z,400,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-02-28T00:00:00Z,650,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-02-29T00:00:00Z,888,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-01T00:00:00Z,1128,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-02T00:00:00Z,1689,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-03T00:00:00Z,1835,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-04T00:00:00Z,2502,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-05T00:00:00Z,3089,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-06T00:00:00Z,3858,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-07T00:00:00Z,4636,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-08T00:00:00Z,5883,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-09T00:00:00Z,7375,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-10T00:00:00Z,9172,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-11T00:00:00Z,10149,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-12T00:00:00Z,12462,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-13T00:00:00Z,15113,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-14T00:00:00Z,17660,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-15T00:00:00Z,17750,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-16T00:00:00Z,23980,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-17T00:00:00Z,27980,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-18T00:00:00Z,31506,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-19T00:00:00Z,35713,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-20T00:00:00Z,41035,total_cases,covid-19,Italy
,,1,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-21T00:00:00Z,47021,total_cases,covid-19,Italy

#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,country
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-03T00:00:00Z,103,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-04T00:00:00Z,125,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-05T00:00:00Z,159,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-06T00:00:00Z,233,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-07T00:00:00Z,338,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-08T00:00:00Z,433,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-09T00:00:00Z,554,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-10T00:00:00Z,754,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-11T00:00:00Z,1025,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-12T00:00:00Z,1312,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-13T00:00:00Z,1663,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-14T00:00:00Z,2174,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-15T00:00:00Z,2951,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-16T00:00:00Z,3774,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-17T00:00:00Z,4661,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-18T00:00:00Z,6427,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-19T00:00:00Z,9415,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-20T00:00:00Z,14250,total_cases,covid-19,United States
,,2,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-03-21T00:00:00Z,19624,total_cases,covid-19,United States
"
outData =
    "
#group,false,false,true,true,true,true,false,false,true
#datatype,string,long,string,string,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string
#default,_result,,,,,,,,
,result,table,_field,_measurement,_start,_stop,_time,_value,country
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-01T00:00:00Z,76369,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-02T00:00:00Z,77016,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-03T00:00:00Z,77234,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-04T00:00:00Z,77749,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-05T00:00:00Z,78159,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-06T00:00:00Z,78598,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-07T00:00:00Z,78927,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-08T00:00:00Z,79355,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-09T00:00:00Z,79929,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-10T00:00:00Z,80134,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-11T00:00:00Z,80261,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-12T00:00:00Z,80380,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-13T00:00:00Z,80497,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-14T00:00:00Z,80667,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-15T00:00:00Z,80768,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-16T00:00:00Z,80814,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-17T00:00:00Z,80859,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-18T00:00:00Z,80879,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-19T00:00:00Z,80908,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-20T00:00:00Z,80932,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-21T00:00:00Z,80954,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-22T00:00:00Z,80973,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-23T00:00:00Z,80995,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-24T00:00:00Z,81020,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-25T00:00:00Z,81130,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-26T00:00:00Z,81163,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-27T00:00:00Z,81238,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-28T00:00:00Z,81337,China
,,0,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-29T00:00:00Z,81416,China
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-01T00:00:00Z,132,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-02T00:00:00Z,229,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-03T00:00:00Z,322,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-04T00:00:00Z,400,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-05T00:00:00Z,650,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-06T00:00:00Z,888,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-07T00:00:00Z,1128,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-08T00:00:00Z,1689,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-09T00:00:00Z,1835,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-10T00:00:00Z,2502,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-11T00:00:00Z,3089,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-12T00:00:00Z,3858,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-13T00:00:00Z,4636,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-14T00:00:00Z,5883,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-15T00:00:00Z,7375,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-16T00:00:00Z,9172,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-17T00:00:00Z,10149,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-18T00:00:00Z,12462,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-19T00:00:00Z,15113,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-20T00:00:00Z,17660,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-21T00:00:00Z,17750,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-22T00:00:00Z,23980,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-23T00:00:00Z,27980,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-24T00:00:00Z,31506,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-25T00:00:00Z,35713,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-26T00:00:00Z,41035,Italy
,,1,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-27T00:00:00Z,47021,Italy
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-01T00:00:00Z,103,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-02T00:00:00Z,125,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-03T00:00:00Z,159,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-04T00:00:00Z,233,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-05T00:00:00Z,338,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-06T00:00:00Z,433,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-07T00:00:00Z,554,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-08T00:00:00Z,754,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-09T00:00:00Z,1025,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-10T00:00:00Z,1312,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-11T00:00:00Z,1663,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-12T00:00:00Z,2174,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-13T00:00:00Z,2951,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-14T00:00:00Z,3774,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-15T00:00:00Z,4661,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-16T00:00:00Z,6427,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-17T00:00:00Z,9415,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-18T00:00:00Z,14250,United States
,,2,total_cases,covid-19,2020-02-22T00:00:00Z,2020-03-22T00:00:00Z,2020-01-19T00:00:00Z,19624,United States
"

testcase align_time {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2020-02-22T00:00:00Z, stop: 2020-03-22T00:00:00Z)
            |> experimental.alignTime(alignTo: 2020-01-01T00:00:00Z)
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
