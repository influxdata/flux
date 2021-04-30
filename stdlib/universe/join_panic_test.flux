package universe_test


import "testing"

passData = "
#group,false,false,true,true,false,false,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,

,result,table,_start,_stop,_time,_value,_field,_measurement,connection_security_policy,destination_app,destination_canonical_revision,destination_canonical_service,destination_principal,destination_service,destination_service_name,destination_service_namespace,destination_version,destination_workload,destination_workload_namespace,env,host,hostname,nodename,reporter,request_protocol,response_code,response_flags,source_app,source_canonical_revision,source_canonical_service,source_principal,source_version,source_workload,source_workload_namespace,url
,,4500,2020-12-21T17:49:44.773856591Z,2020-12-21T17:50:44.773856591Z,2020-12-21T17:50:20Z,20131,counter,istio_requests_total,mutual_tls,gateway,v4,gateway,spiffe://prod101-us-east-1.aws.influxdata.io/ns/twodotoh/sa/gateway,gateway-internal-meta.twodotoh.svc.prod101-us-east-1.aws.influxdata.io,gateway-internal-meta,twodotoh,v4,gateway-internal-meta,twodotoh,prod101-us-east-1,gateway-internal-meta-546b695987-6fwfg,gateway-internal-meta-546b695987-6fwfg,ip-10-143-10-201.ec2.internal,destination,http,200,-,queryd,v1,queryd,spiffe://prod101-us-east-1.aws.influxdata.io/ns/twodotoh/sa/queryd-service-account,v1,queryd-v1,twodotoh,http://127.0.0.1:15090/stats/prometheus
,,4510,2020-12-21T17:49:44.773856591Z,2020-12-21T17:50:44.773856591Z,2020-12-21T17:50:20Z,75794,counter,istio_requests_total,none,gateway,v4,gateway,unknown,gateway-external-write.twodotoh.svc.prod01.us-west-2.local,gateway-external-write,twodotoh,v4,gateway-external-write,twodotoh,prod01-us-west-2,gateway-external-write-9c6585b49-2qdpg,gateway-external-write-9c6585b49-2qdpg,ip-10-130-16-200.us-west-2.compute.internal,destination,http,429,-,unknown,latest,unknown,unknown,unknown,unknown,unknown,http://127.0.0.1:15090/stats/prometheus
,,4510,2020-12-21T17:49:44.773856591Z,2020-12-21T17:50:44.773856591Z,2020-12-21T17:50:30Z,75810,counter,istio_requests_total,none,gateway,v4,gateway,unknown,gateway-external-write.twodotoh.svc.prod01.us-west-2.local,gateway-external-write,twodotoh,v4,gateway-external-write,twodotoh,prod01-us-west-2,gateway-external-write-9c6585b49-2qdpg,gateway-external-write-9c6585b49-2qdpg,ip-10-130-16-200.us-west-2.compute.internal,destination,http,429,-,unknown,latest,unknown,unknown,unknown,unknown,unknown,http://127.0.0.1:15090/stats/prometheus
,,4535,2020-12-21T17:49:44.773856591Z,2020-12-21T17:50:44.773856591Z,2020-12-21T17:50:00Z,5000,counter,istio_requests_total,none,gateway,v4,gateway,unknown,gateway-external-query.twodotoh.svc.prod01.us-west-2.local,gateway-external-query,twodotoh,v4,gateway-external-query,twodotoh,prod01-us-west-2,gateway-external-query-b499584c-65xvz,gateway-external-query-b499584c-65xvz,ip-10-130-16-33.us-west-2.compute.internal,destination,http,400,-,unknown,latest,unknown,unknown,unknown,unknown,unknown,http://127.0.0.1:15090/stats/prometheus
,,4535,2020-12-21T17:49:44.773856591Z,2020-12-21T17:50:44.773856591Z,2020-12-21T17:50:10Z,5002,counter,istio_requests_total,none,gateway,v4,gateway,unknown,gateway-external-query.twodotoh.svc.prod01.us-west-2.local,gateway-external-query,twodotoh,v4,gateway-external-query,twodotoh,prod01-us-west-2,gateway-external-query-b499584c-65xvz,gateway-external-query-b499584c-65xvz,ip-10-130-16-33.us-west-2.compute.internal,destination,http,400,-,unknown,latest,unknown,unknown,unknown,unknown,unknown,http://127.0.0.1:15090/stats/prometheus
"
failData = "
#group,false,false,true,true,false,false,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,

,result,table,_start,_stop,_time,_value,_field,_measurement,connection_security_policy,destination_app,destination_canonical_revision,destination_canonical_service,destination_principal,destination_service,destination_service_name,destination_service_namespace,destination_version,destination_workload,destination_workload_namespace,env,host,hostname,nodename,reporter,request_protocol,response_code,response_flags,source_app,source_canonical_revision,source_canonical_service,source_principal,source_version,source_workload,source_workload_namespace,url
,,4500,2020-12-21T17:49:44.773856591Z,2020-12-21T17:50:44.773856591Z,2020-12-21T17:50:20Z,20131,counter,istio_requests_total,mutual_tls,gateway,v4,gateway,spiffe://prod101-us-east-1.aws.influxdata.io/ns/twodotoh/sa/gateway,gateway-internal-meta.twodotoh.svc.prod101-us-east-1.aws.influxdata.io,gateway-internal-meta,twodotoh,v4,gateway-internal-meta,twodotoh,prod101-us-east-1,gateway-internal-meta-546b695987-6fwfg,gateway-internal-meta-546b695987-6fwfg,ip-10-143-10-201.ec2.internal,destination,http,200,-,queryd,v1,queryd,spiffe://prod101-us-east-1.aws.influxdata.io/ns/twodotoh/sa/queryd-service-account,v1,queryd-v1,twodotoh,http://127.0.0.1:15090/stats/prometheus
,,4500,2020-12-21T17:49:44.773856591Z,2020-12-21T17:50:44.773856591Z,2020-12-21T17:50:30Z,20133,counter,istio_requests_total,mutual_tls,gateway,v4,gateway,spiffe://prod101-us-east-1.aws.influxdata.io/ns/twodotoh/sa/gateway,gateway-internal-meta.twodotoh.svc.prod101-us-east-1.aws.influxdata.io,gateway-internal-meta,twodotoh,v4,gateway-internal-meta,twodotoh,prod101-us-east-1,gateway-internal-meta-546b695987-6fwfg,gateway-internal-meta-546b695987-6fwfg,ip-10-143-10-201.ec2.internal,destination,http,200,-,queryd,v1,queryd,spiffe://prod101-us-east-1.aws.influxdata.io/ns/twodotoh/sa/queryd-service-account,v1,queryd-v1,twodotoh,http://127.0.0.1:15090/stats/prometheus
,,4510,2020-12-21T17:49:44.773856591Z,2020-12-21T17:50:44.773856591Z,2020-12-21T17:50:20Z,75794,counter,istio_requests_total,none,gateway,v4,gateway,unknown,gateway-external-write.twodotoh.svc.prod01.us-west-2.local,gateway-external-write,twodotoh,v4,gateway-external-write,twodotoh,prod01-us-west-2,gateway-external-write-9c6585b49-2qdpg,gateway-external-write-9c6585b49-2qdpg,ip-10-130-16-200.us-west-2.compute.internal,destination,http,429,-,unknown,latest,unknown,unknown,unknown,unknown,unknown,http://127.0.0.1:15090/stats/prometheus
,,4510,2020-12-21T17:49:44.773856591Z,2020-12-21T17:50:44.773856591Z,2020-12-21T17:50:30Z,75810,counter,istio_requests_total,none,gateway,v4,gateway,unknown,gateway-external-write.twodotoh.svc.prod01.us-west-2.local,gateway-external-write,twodotoh,v4,gateway-external-write,twodotoh,prod01-us-west-2,gateway-external-write-9c6585b49-2qdpg,gateway-external-write-9c6585b49-2qdpg,ip-10-130-16-200.us-west-2.compute.internal,destination,http,429,-,unknown,latest,unknown,unknown,unknown,unknown,unknown,http://127.0.0.1:15090/stats/prometheus
,,4535,2020-12-21T17:49:44.773856591Z,2020-12-21T17:50:44.773856591Z,2020-12-21T17:49:50Z,5000,counter,istio_requests_total,none,gateway,v4,gateway,unknown,gateway-external-query.twodotoh.svc.prod01.us-west-2.local,gateway-external-query,twodotoh,v4,gateway-external-query,twodotoh,prod01-us-west-2,gateway-external-query-b499584c-65xvz,gateway-external-query-b499584c-65xvz,ip-10-130-16-33.us-west-2.compute.internal,destination,http,400,-,unknown,latest,unknown,unknown,unknown,unknown,unknown,http://127.0.0.1:15090/stats/prometheus
,,4535,2020-12-21T17:49:44.773856591Z,2020-12-21T17:50:44.773856591Z,2020-12-21T17:50:00Z,5000,counter,istio_requests_total,none,gateway,v4,gateway,unknown,gateway-external-query.twodotoh.svc.prod01.us-west-2.local,gateway-external-query,twodotoh,v4,gateway-external-query,twodotoh,prod01-us-west-2,gateway-external-query-b499584c-65xvz,gateway-external-query-b499584c-65xvz,ip-10-130-16-33.us-west-2.compute.internal,destination,http,400,-,unknown,latest,unknown,unknown,unknown,unknown,unknown,http://127.0.0.1:15090/stats/prometheus
,,4535,2020-12-21T17:49:44.773856591Z,2020-12-21T17:50:44.773856591Z,2020-12-21T17:50:10Z,5002,counter,istio_requests_total,none,gateway,v4,gateway,unknown,gateway-external-query.twodotoh.svc.prod01.us-west-2.local,gateway-external-query,twodotoh,v4,gateway-external-query,twodotoh,prod01-us-west-2,gateway-external-query-b499584c-65xvz,gateway-external-query-b499584c-65xvz,ip-10-130-16-33.us-west-2.compute.internal,destination,http,400,-,unknown,latest,unknown,unknown,unknown,unknown,unknown,http://127.0.0.1:15090/stats/prometheus
"
outData = "
#group,false,false,false,false,false,false,false,false
#datatype,string,long,double,double,double,string,string,string
#default,_result,,,,,,,

,result,table,_value_errors,_value_total,availability,env,response_code_errors,response_code_total
,,0,2,2,0,prod01-us-west-2,400,400
,,0,2,16,87.5,prod01-us-west-2,400,429
,,0,16,2,-700,prod01-us-west-2,429,400
,,0,16,16,0,prod01-us-west-2,429,429
"
t_join_panic = (table=<-) => {
    api_requests = table |> difference()
    errors = api_requests
        |> filter(fn: (r) => r.response_code == "400" or r.response_code == "401" or r.response_code == "404" or r.response_code == "429" or r.response_code == "500" or r.response_code == "503")
        |> group(columns: ["env", "response_code"])
        |> sum()
        |> filter(fn: (r) => r._value > 0)
    total = api_requests
        //|> group(columns: ["env"])
        |> group(columns: ["env", "response_code"])
        |> sum()
        |> filter(fn: (r) => r._value > 0)

    return join(tables: {errors: errors, total: total}, on: ["env"])
        |> map(fn: (r) => ({r with availability: (1.0 - float(v: r._value_errors) / float(v: r._value_total)) * 100.0}))
        |> sort(columns: ["availability"], desc: true)
        |> group()
}

test _join_panic = () => 
    // to trigger the panic, switch the testing.loadStorage() csv from `passData` to `failData`
    ({input: testing.loadStorage(csv: passData), want: testing.loadMem(csv: outData), fn: t_join_panic})
