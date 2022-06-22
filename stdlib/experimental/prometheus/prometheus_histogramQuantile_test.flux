package prometheus_test


import "csv"
import "experimental/prometheus"
import "testing"

// Test prometheus.histogramQuantile with Prometheus histograms written using
// the Prometheus metric parser format version 2
testcase prometheus_histogramQuantile_v2 {
    inData =
        "#group,false,false,true,true,false,false,true,true
#datatype,string,long,string,string,dateTime:RFC3339,double,string,string
#default,_result,,,,,,,
,result,table,_field,_measurement,_time,_value,le,org
,,0,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,10,+Inf,0001
,,0,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,11,+Inf,0001
,,0,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,11,+Inf,0001
,,0,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,11,+Inf,0001
,,0,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,11,+Inf,0001
,,0,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,11,+Inf,0001
,,1,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,1402,+Inf,0002
,,1,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,1405,+Inf,0002
,,1,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,1407,+Inf,0002
,,1,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,1409,+Inf,0002
,,1,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,1411,+Inf,0002
,,1,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,1413,+Inf,0002
,,2,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,0,0.001,0001
,,2,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,0,0.001,0001
,,2,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,0,0.001,0001
,,2,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,0,0.001,0001
,,2,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,0,0.001,0001
,,2,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,0,0.001,0001
,,3,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,1,0.001,0002
,,3,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,1,0.001,0002
,,3,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,1,0.001,0002
,,3,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,1,0.001,0002
,,3,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,1,0.001,0002
,,3,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,1,0.001,0002
,,4,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,0,0.005,0001
,,4,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,0,0.005,0001
,,4,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,0,0.005,0001
,,4,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,0,0.005,0001
,,4,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,0,0.005,0001
,,4,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,0,0.005,0001
,,5,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,1,0.005,0002
,,5,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,1,0.005,0002
,,5,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,1,0.005,0002
,,5,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,1,0.005,0002
,,5,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,1,0.005,0002
,,5,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,1,0.005,0002
,,6,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,0,0.025,0001
,,6,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,0,0.025,0001
,,6,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,0,0.025,0001
,,6,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,0,0.025,0001
,,6,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,0,0.025,0001
,,6,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,0,0.025,0001
,,7,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,84,0.025,0002
,,7,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,84,0.025,0002
,,7,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,84,0.025,0002
,,7,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,84,0.025,0002
,,7,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,84,0.025,0002
,,7,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,84,0.025,0002
,,8,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,0,0.125,0001
,,8,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,0,0.125,0001
,,8,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,0,0.125,0001
,,8,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,0,0.125,0001
,,8,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,0,0.125,0001
,,8,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,0,0.125,0001
,,9,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,980,0.125,0002
,,9,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,981,0.125,0002
,,9,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,981,0.125,0002
,,9,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,981,0.125,0002
,,9,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,981,0.125,0002
,,9,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,981,0.125,0002
,,10,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,0,0.625,0001
,,10,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,0,0.625,0001
,,10,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,0,0.625,0001
,,10,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,0,0.625,0001
,,10,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,0,0.625,0001
,,10,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,0,0.625,0001
,,11,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,1370,0.625,0002
,,11,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,1373,0.625,0002
,,11,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,1375,0.625,0002
,,11,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,1377,0.625,0002
,,11,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,1379,0.625,0002
,,11,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,1381,0.625,0002
,,12,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,10,15.625,0001
,,12,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,11,15.625,0001
,,12,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,11,15.625,0001
,,12,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,11,15.625,0001
,,12,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,11,15.625,0001
,,12,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,11,15.625,0001
,,13,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,1402,15.625,0002
,,13,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,1405,15.625,0002
,,13,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,1407,15.625,0002
,,13,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,1409,15.625,0002
,,13,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,1411,15.625,0002
,,13,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,1413,15.625,0002
,,14,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,0,3.125,0001
,,14,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,0,3.125,0001
,,14,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,0,3.125,0001
,,14,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,0,3.125,0001
,,14,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,0,3.125,0001
,,14,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,0,3.125,0001
,,15,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00.412729Z,1398,3.125,0002
,,15,qc_all_duration_seconds,prometheus,2021-10-08T00:00:10.362635Z,1401,3.125,0002
,,15,qc_all_duration_seconds,prometheus,2021-10-08T00:00:20.374072Z,1403,3.125,0002
,,15,qc_all_duration_seconds,prometheus,2021-10-08T00:00:30.330013Z,1405,3.125,0002
,,15,qc_all_duration_seconds,prometheus,2021-10-08T00:00:40.377Z,1407,3.125,0002
,,15,qc_all_duration_seconds,prometheus,2021-10-08T00:00:50.373893Z,1409,3.125,0002
"
    outData =
        "#group,false,false,true,true,true,true,false,true,false,true
#datatype,string,long,string,string,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,double,string
#default,_result,,,,,,,,,
,result,table,_field,_measurement,_start,_stop,_time,org,_value,quantile
,,0,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:00.412729Z,0001,15.5,0.99
,,0,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:10.362635Z,0001,15.500000000000002,0.99
,,0,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:20.374072Z,0001,15.500000000000002,0.99
,,0,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:30.330013Z,0001,15.500000000000002,0.99
,,0,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:40.377Z,0001,15.500000000000002,0.99
,,0,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:50.373893Z,0001,15.500000000000002,0.99
,,1,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:00.412729Z,0002,2.2303571428571445,0.99
,,1,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:10.362635Z,0002,2.2276785714285756,0.99
,,1,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:20.374072Z,0002,2.225892857142863,0.99
,,1,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:30.330013Z,0002,2.22410714285715,0.99
,,1,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:40.377Z,0002,2.2223214285714374,0.99
,,1,qc_all_duration_seconds,prometheus,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:50.373893Z,0002,2.2205357142857047,0.99
"
    want = csv.from(csv: outData)
    got =
        csv.from(csv: inData)
            |> range(start: 2021-10-08T00:00:00Z, stop: 2021-10-08T00:01:00Z)
            |> prometheus.histogramQuantile(quantile: 0.99, metricVersion: 2)

    testing.diff(got: got, want: want)
}

// Test prometheus.histogramQuantile with Prometheus histograms written using
// the Prometheus metric parser format version 1
testcase prometheus_histogramQuantile_v1 {
    inData =
        "#group,false,false,true,true,false,false,true
#datatype,string,long,string,string,dateTime:RFC3339,double,string
#default,_result,,,,,,
,result,table,_field,_measurement,_time,_value,org
,,0,+Inf,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,10,0001
,,0,+Inf,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,11,0001
,,0,+Inf,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,11,0001
,,0,+Inf,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,11,0001
,,0,+Inf,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,11,0001
,,0,+Inf,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,11,0001
,,1,+Inf,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,1405,0002
,,1,+Inf,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,1407,0002
,,1,+Inf,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,1409,0002
,,1,+Inf,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,1411,0002
,,1,+Inf,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,1413,0002
,,1,+Inf,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,1415,0002
,,2,0.001,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,0,0001
,,2,0.001,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,0,0001
,,2,0.001,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,0,0001
,,2,0.001,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,0,0001
,,2,0.001,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,0,0001
,,2,0.001,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,0,0001
,,3,0.001,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,1,0002
,,3,0.001,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,1,0002
,,3,0.001,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,1,0002
,,3,0.001,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,1,0002
,,3,0.001,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,1,0002
,,3,0.001,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,1,0002
,,4,0.005,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,0,0001
,,4,0.005,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,0,0001
,,4,0.005,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,0,0001
,,4,0.005,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,0,0001
,,4,0.005,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,0,0001
,,4,0.005,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,0,0001
,,5,0.005,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,1,0002
,,5,0.005,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,1,0002
,,5,0.005,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,1,0002
,,5,0.005,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,1,0002
,,5,0.005,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,1,0002
,,5,0.005,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,1,0002
,,6,0.025,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,0,0001
,,6,0.025,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,0,0001
,,6,0.025,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,0,0001
,,6,0.025,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,0,0001
,,6,0.025,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,0,0001
,,6,0.025,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,0,0001
,,7,0.025,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,84,0002
,,7,0.025,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,84,0002
,,7,0.025,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,84,0002
,,7,0.025,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,84,0002
,,7,0.025,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,84,0002
,,7,0.025,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,84,0002
,,8,0.125,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,0,0001
,,8,0.125,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,0,0001
,,8,0.125,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,0,0001
,,8,0.125,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,0,0001
,,8,0.125,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,0,0001
,,8,0.125,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,0,0001
,,9,0.125,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,981,0002
,,9,0.125,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,981,0002
,,9,0.125,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,981,0002
,,9,0.125,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,981,0002
,,9,0.125,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,981,0002
,,9,0.125,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,981,0002
,,10,0.625,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,0,0001
,,10,0.625,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,0,0001
,,10,0.625,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,0,0001
,,10,0.625,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,0,0001
,,10,0.625,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,0,0001
,,10,0.625,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,0,0001
,,11,0.625,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,1373,0002
,,11,0.625,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,1375,0002
,,11,0.625,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,1377,0002
,,11,0.625,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,1379,0002
,,11,0.625,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,1381,0002
,,11,0.625,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,1383,0002
,,12,15.625,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,10,0001
,,12,15.625,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,11,0001
,,12,15.625,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,11,0001
,,12,15.625,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,11,0001
,,12,15.625,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,11,0001
,,12,15.625,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,11,0001
,,13,15.625,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,1405,0002
,,13,15.625,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,1407,0002
,,13,15.625,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,1409,0002
,,13,15.625,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,1411,0002
,,13,15.625,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,1413,0002
,,13,15.625,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,1415,0002
,,14,3.125,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,0,0001
,,14,3.125,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,0,0001
,,14,3.125,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,0,0001
,,14,3.125,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,0,0001
,,14,3.125,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,0,0001
,,14,3.125,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,0,0001
,,15,3.125,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,1401,0002
,,15,3.125,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,1403,0002
,,15,3.125,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,1405,0002
,,15,3.125,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,1407,0002
,,15,3.125,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,1409,0002
,,15,3.125,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,1411,0002
,,16,count,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,10,0001
,,16,count,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,11,0001
,,16,count,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,11,0001
,,16,count,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,11,0001
,,16,count,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,11,0001
,,16,count,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,11,0001
,,17,count,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,1405,0002
,,17,count,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,1407,0002
,,17,count,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,1409,0002
,,17,count,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,1411,0002
,,17,count,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,1413,0002
,,17,count,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,1415,0002
,,18,sum,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,45.746700925,0001
,,18,sum,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,50.714616303999996,0001
,,18,sum,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,50.714616303999996,0001
,,18,sum,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,50.714616303999996,0001
,,18,sum,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,50.714616303999996,0001
,,18,sum,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,50.714616303999996,0001
,,19,sum,qc_all_duration_seconds,2021-10-08T00:00:01.866064Z,178.4667627259998,0002
,,19,sum,qc_all_duration_seconds,2021-10-08T00:00:11.869239Z,178.85329525599983,0002
,,19,sum,qc_all_duration_seconds,2021-10-08T00:00:21.869603Z,179.18850017099984,0002
,,19,sum,qc_all_duration_seconds,2021-10-08T00:00:31.870017Z,179.52159502899983,0002
,,19,sum,qc_all_duration_seconds,2021-10-08T00:00:41.871188Z,179.87429533899981,0002
,,19,sum,qc_all_duration_seconds,2021-10-08T00:00:51.871403Z,180.2172711799998,0002
"
    outData =
        "#group,false,false,true,true,true,false,true,false,true
#datatype,string,long,string,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,double,string
#default,_result,,,,,,,,
,result,table,_measurement,_start,_stop,_time,org,_value,quantile
,,0,qc_all_duration_seconds,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:01.866064Z,0001,15.5,0.99
,,0,qc_all_duration_seconds,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:11.869239Z,0001,15.500000000000002,0.99
,,0,qc_all_duration_seconds,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:21.869603Z,0001,15.500000000000002,0.99
,,0,qc_all_duration_seconds,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:31.870017Z,0001,15.500000000000002,0.99
,,0,qc_all_duration_seconds,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:41.871188Z,0001,15.500000000000002,0.99
,,0,qc_all_duration_seconds,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:51.871403Z,0001,15.500000000000002,0.99
,,1,qc_all_duration_seconds,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:01.866064Z,0002,2.2276785714285756,0.99
,,1,qc_all_duration_seconds,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:11.869239Z,0002,2.225892857142863,0.99
,,1,qc_all_duration_seconds,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:21.869603Z,0002,2.22410714285715,0.99
,,1,qc_all_duration_seconds,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:31.870017Z,0002,2.2223214285714374,0.99
,,1,qc_all_duration_seconds,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:41.871188Z,0002,2.2205357142857047,0.99
,,1,qc_all_duration_seconds,2021-10-08T00:00:00Z,2021-10-08T00:01:00Z,2021-10-08T00:00:51.871403Z,0002,2.218749999999992,0.99
"
    want = csv.from(csv: outData)
    got =
        csv.from(csv: inData)
            |> range(start: 2021-10-08T00:00:00Z, stop: 2021-10-08T00:01:00Z)
            |> prometheus.histogramQuantile(quantile: 0.99, metricVersion: 1)

    testing.diff(got: got, want: want)
}
