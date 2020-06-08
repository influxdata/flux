package statsmodels_test

import "testing"
import "contrib/anaisdg/statsmodels" 


inData = "
#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,shelter,type
,,0,2020-05-21T21:30:48.901Z,2020-05-21T21:50:48.9Z,2020-05-21T21:43:45Z,7,young,cats,A,calico
,,0,2020-05-21T21:30:48.901Z,2020-05-21T21:50:48.9Z,2020-05-21T21:45:08Z,5,young,cats,A,calico
,,0,2020-05-21T21:30:48.901Z,2020-05-21T21:50:48.9Z,2020-05-21T21:46:25Z,4,young,cats,A,calico
,,0,2020-05-21T21:30:48.901Z,2020-05-21T21:50:48.9Z,2020-05-21T21:48:38Z,3,young,cats,A,calico
"
outData = "
#group,false,false,false,true,true,true,true,false,false,true,false,false,false,false,false,true,false,false,false
#datatype,string,long,double,string,string,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,double,double,double,double,double,string,double,double,double
#default,_result,,,,,,,,,,,,,,,,,,
,result,table,N,_field,_measurement,_start,_stop,_time,errors,shelter,slope,sx,sxx,sxy,sy,type,x,y,y_hat
,,0,4,young,cats,2020-05-21T21:30:48.901Z,2020-05-21T21:50:48.9Z,2020-05-21T21:43:45Z,0.0899999999999999,A,-1.3,10,30,41,19,calico,1,7,6.7
,,0,4,young,cats,2020-05-21T21:30:48.901Z,2020-05-21T21:50:48.9Z,2020-05-21T21:45:08Z,0.16000000000000028,A,-1.3,10,30,41,19,calico,2,5,5.4
,,0,4,young,cats,2020-05-21T21:30:48.901Z,2020-05-21T21:50:48.9Z,2020-05-21T21:46:25Z,0.009999999999999929,A,-1.3,10,30,41,19,calico,3,4,4.1
,,0,4,young,cats,2020-05-21T21:30:48.901Z,2020-05-21T21:50:48.9Z,2020-05-21T21:48:38Z,0.04000000000000007,A,-1.3,10,30,41,19,calico,4,3,2.8
"


t_linearRegression = (table=<-) =>
	table
		|> range(start: 2020-05-21T21:30:48.901Z, stop: 2020-05-21T21:50:48.9Z)
		|> statsmodels.linearRegression()

test _linearRegression = () =>
({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_linearRegression})

