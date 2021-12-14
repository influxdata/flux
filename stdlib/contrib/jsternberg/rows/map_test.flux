package rows_test


import "testing"
import "contrib/jsternberg/rows"

option now = () => 2020-08-02T17:24:00Z

inData =
    "
#datatype,string,long,string,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,false,false
#default,,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m0,f0,a-0,2020-08-02T17:22:00Z,-43.09452210525144
,,0,m0,f0,a-0,2020-08-02T17:22:10Z,30.353812994348537
,,0,m0,f0,a-0,2020-08-02T17:22:20Z,-19.17028701626966
,,0,m0,f0,a-0,2020-08-02T17:22:30Z,-31.713408760790323
,,0,m0,f0,a-0,2020-08-02T17:22:40Z,-16.22173130975937
,,0,m0,f0,a-0,2020-08-02T17:22:50Z,14.631305556841284
,,0,m0,f0,a-0,2020-08-02T17:23:00Z,85.5542463240766
,,0,m0,f0,a-0,2020-08-02T17:23:10Z,-77.18220390886191
,,0,m0,f0,a-0,2020-08-02T17:23:20Z,50.062559688977814
,,0,m0,f0,a-0,2020-08-02T17:23:30Z,22.17256401464515
,,0,m0,f0,a-0,2020-08-02T17:23:40Z,-112.47430195827386
,,0,m0,f0,a-0,2020-08-02T17:23:50Z,16.85801752656638
,,1,m0,f0,a-1,2020-08-02T17:22:00Z,-28.65256634110021
,,1,m0,f0,a-1,2020-08-02T17:22:10Z,-11.021368187315897
,,1,m0,f0,a-1,2020-08-02T17:22:20Z,18.04898637542153
,,1,m0,f0,a-1,2020-08-02T17:22:30Z,24.555312299824035
,,1,m0,f0,a-1,2020-08-02T17:22:40Z,5.543823619638458
,,1,m0,f0,a-1,2020-08-02T17:22:50Z,-64.34272303286494
,,1,m0,f0,a-1,2020-08-02T17:23:00Z,-54.40142609111467
,,1,m0,f0,a-1,2020-08-02T17:23:10Z,-6.68919215397088
,,1,m0,f0,a-1,2020-08-02T17:23:20Z,-36.36364746675186
,,1,m0,f0,a-1,2020-08-02T17:23:30Z,-31.041492590916768
,,1,m0,f0,a-1,2020-08-02T17:23:40Z,-8.461569912796826
,,1,m0,f0,a-1,2020-08-02T17:23:50Z,9.025669280720571
,,2,m0,f0,a-2,2020-08-02T17:22:00Z,-8.640246126337203
,,2,m0,f0,a-2,2020-08-02T17:22:10Z,-43.365488430173706
,,2,m0,f0,a-2,2020-08-02T17:22:20Z,-25.198611516637676
,,2,m0,f0,a-2,2020-08-02T17:22:30Z,16.593516600485213
,,2,m0,f0,a-2,2020-08-02T17:22:40Z,-76.42451523676915
,,2,m0,f0,a-2,2020-08-02T17:22:50Z,-67.78699694188528
,,2,m0,f0,a-2,2020-08-02T17:23:00Z,14.77477027658923
,,2,m0,f0,a-2,2020-08-02T17:23:10Z,28.521034402304263
,,2,m0,f0,a-2,2020-08-02T17:23:20Z,-53.47644712761566
,,2,m0,f0,a-2,2020-08-02T17:23:30Z,83.38193426782863
,,2,m0,f0,a-2,2020-08-02T17:23:40Z,-92.49751968643372
,,2,m0,f0,a-2,2020-08-02T17:23:50Z,2.187536871928522
"
outData =
    "
#datatype,string,long,string,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,false,false
#default,,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m0,f0,a-0,2020-08-02T17:22:00Z,43.09452210525144
,,0,m0,f0,a-0,2020-08-02T17:22:10Z,-30.353812994348537
,,0,m0,f0,a-0,2020-08-02T17:22:20Z,19.17028701626966
,,0,m0,f0,a-0,2020-08-02T17:22:30Z,31.713408760790323
,,0,m0,f0,a-0,2020-08-02T17:22:40Z,16.22173130975937
,,0,m0,f0,a-0,2020-08-02T17:22:50Z,-14.631305556841284
,,0,m0,f0,a-0,2020-08-02T17:23:00Z,-85.5542463240766
,,0,m0,f0,a-0,2020-08-02T17:23:10Z,77.18220390886191
,,0,m0,f0,a-0,2020-08-02T17:23:20Z,-50.062559688977814
,,0,m0,f0,a-0,2020-08-02T17:23:30Z,-22.17256401464515
,,0,m0,f0,a-0,2020-08-02T17:23:40Z,112.47430195827386
,,0,m0,f0,a-0,2020-08-02T17:23:50Z,-16.85801752656638
,,1,m0,f0,a-1,2020-08-02T17:22:00Z,28.65256634110021
,,1,m0,f0,a-1,2020-08-02T17:22:10Z,11.021368187315897
,,1,m0,f0,a-1,2020-08-02T17:22:20Z,-18.04898637542153
,,1,m0,f0,a-1,2020-08-02T17:22:30Z,-24.555312299824035
,,1,m0,f0,a-1,2020-08-02T17:22:40Z,-5.543823619638458
,,1,m0,f0,a-1,2020-08-02T17:22:50Z,64.34272303286494
,,1,m0,f0,a-1,2020-08-02T17:23:00Z,54.40142609111467
,,1,m0,f0,a-1,2020-08-02T17:23:10Z,6.68919215397088
,,1,m0,f0,a-1,2020-08-02T17:23:20Z,36.36364746675186
,,1,m0,f0,a-1,2020-08-02T17:23:30Z,31.041492590916768
,,1,m0,f0,a-1,2020-08-02T17:23:40Z,8.461569912796826
,,1,m0,f0,a-1,2020-08-02T17:23:50Z,-9.025669280720571
,,2,m0,f0,a-2,2020-08-02T17:22:00Z,8.640246126337203
,,2,m0,f0,a-2,2020-08-02T17:22:10Z,43.365488430173706
,,2,m0,f0,a-2,2020-08-02T17:22:20Z,25.198611516637676
,,2,m0,f0,a-2,2020-08-02T17:22:30Z,-16.593516600485213
,,2,m0,f0,a-2,2020-08-02T17:22:40Z,76.42451523676915
,,2,m0,f0,a-2,2020-08-02T17:22:50Z,67.78699694188528
,,2,m0,f0,a-2,2020-08-02T17:23:00Z,-14.77477027658923
,,2,m0,f0,a-2,2020-08-02T17:23:10Z,-28.521034402304263
,,2,m0,f0,a-2,2020-08-02T17:23:20Z,53.47644712761566
,,2,m0,f0,a-2,2020-08-02T17:23:30Z,-83.38193426782863
,,2,m0,f0,a-2,2020-08-02T17:23:40Z,92.49751968643372
,,2,m0,f0,a-2,2020-08-02T17:23:50Z,-2.187536871928522
"
t_map = (table=<-) =>
    table
        |> range(start: -2m)
        |> drop(columns: ["_start", "_stop"])
        |> rows.map(fn: (r) => ({_time: r._time, _value: -r._value}))

test _map = () => ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_map})
