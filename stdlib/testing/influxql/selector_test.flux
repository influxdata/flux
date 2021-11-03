package influxql_test


import "testing"
import "internal/influxql"

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,string,double
#group,false,false,false,true,true,true,true,false
#default,0,,,,,,,
,result,table,_time,_measurement,t0,t1,_field,_value
,,0,1970-01-01T00:00:00Z,m,0,0,f,0.21546887461084024
,,0,1970-01-01T00:00:01Z,m,0,0,f,0.9576896132790585
,,0,1970-01-01T00:00:02Z,m,0,0,f,0.294953913000311
,,0,1970-01-01T00:00:03Z,m,0,0,f,0.4651741324883778
,,0,1970-01-01T00:00:04Z,m,0,0,f,0.9873388815871567
,,0,1970-01-01T00:00:05Z,m,0,0,f,0.3845474109517986
,,0,1970-01-01T00:00:06Z,m,0,0,f,0.2922442980858412
,,0,1970-01-01T00:00:07Z,m,0,0,f,0.03298588199059829
,,0,1970-01-01T00:00:08Z,m,0,0,f,0.969396406468683
,,0,1970-01-01T00:00:09Z,m,0,0,f,0.8126386582005671
,,0,1970-01-01T00:00:10Z,m,0,0,f,0.875468209815408
,,0,1970-01-01T00:00:11Z,m,0,0,f,0.43242435584494165
,,0,1970-01-01T00:00:12Z,m,0,0,f,0.43936224189298456
,,0,1970-01-01T00:00:13Z,m,0,0,f,0.1224409595139043
,,0,1970-01-01T00:00:14Z,m,0,0,f,0.15733684152804783
,,0,1970-01-01T00:00:15Z,m,0,0,f,0.08882282140312904
,,0,1970-01-01T00:00:16Z,m,0,0,f,0.23989257176325227
,,0,1970-01-01T00:00:17Z,m,0,0,f,0.6955232509082638
,,0,1970-01-01T00:00:18Z,m,0,0,f,0.43554475339119303
,,0,1970-01-01T00:00:19Z,m,0,0,f,0.3051713218684253
,,0,1970-01-01T00:00:20Z,m,0,0,f,0.7413025816537797
,,0,1970-01-01T00:00:21Z,m,0,0,f,0.24567297998270615
,,0,1970-01-01T00:00:22Z,m,0,0,f,0.491391504478891
,,0,1970-01-01T00:00:23Z,m,0,0,f,0.13872180750181634
,,0,1970-01-01T00:00:24Z,m,0,0,f,0.06729135892978601
,,0,1970-01-01T00:00:25Z,m,0,0,f,0.2711347220286289
,,0,1970-01-01T00:00:26Z,m,0,0,f,0.5465962906385142
,,0,1970-01-01T00:00:27Z,m,0,0,f,0.1721498986023557
,,0,1970-01-01T00:00:28Z,m,0,0,f,0.928541805026285
,,0,1970-01-01T00:00:29Z,m,0,0,f,0.4390512841392946
,,0,1970-01-01T00:00:30Z,m,0,0,f,0.7891509564074856
,,0,1970-01-01T00:00:31Z,m,0,0,f,0.03752404112396554
,,0,1970-01-01T00:00:32Z,m,0,0,f,0.8731292945164265
,,0,1970-01-01T00:00:33Z,m,0,0,f,0.6590129312109282
,,0,1970-01-01T00:00:34Z,m,0,0,f,0.7298034951937612
,,0,1970-01-01T00:00:35Z,m,0,0,f,0.6880331199538888
,,0,1970-01-01T00:00:36Z,m,0,0,f,0.7884092917020722
,,0,1970-01-01T00:00:37Z,m,0,0,f,0.9071621838398441
,,0,1970-01-01T00:00:38Z,m,0,0,f,0.5029003668295414
,,0,1970-01-01T00:00:39Z,m,0,0,f,0.5545818527629861
,,0,1970-01-01T00:00:40Z,m,0,0,f,0.763728196635538
,,0,1970-01-01T00:00:41Z,m,0,0,f,0.5870046094520823
,,0,1970-01-01T00:00:42Z,m,0,0,f,0.7675553560334312
,,0,1970-01-01T00:00:43Z,m,0,0,f,0.8279726730049255
,,0,1970-01-01T00:00:44Z,m,0,0,f,0.7013474149025897
,,0,1970-01-01T00:00:45Z,m,0,0,f,0.08556981440432106
,,0,1970-01-01T00:00:46Z,m,0,0,f,0.8520957093766447
,,0,1970-01-01T00:00:47Z,m,0,0,f,0.41873957390346783
,,0,1970-01-01T00:00:48Z,m,0,0,f,0.04405459160245573
,,0,1970-01-01T00:00:49Z,m,0,0,f,0.8184927094237151
,,0,1970-01-01T00:00:50Z,m,0,0,f,0.0975526753791771
,,0,1970-01-01T00:00:51Z,m,0,0,f,0.4984015942759995
,,0,1970-01-01T00:00:52Z,m,0,0,f,0.24094630162586889
,,0,1970-01-01T00:00:53Z,m,0,0,f,0.1461722759564162
,,0,1970-01-01T00:00:54Z,m,0,0,f,0.0008451156568219057
,,0,1970-01-01T00:00:55Z,m,0,0,f,0.4633414547017063
,,0,1970-01-01T00:00:56Z,m,0,0,f,0.4539668492775038
,,0,1970-01-01T00:00:57Z,m,0,0,f,0.4868916379116324
,,0,1970-01-01T00:00:58Z,m,0,0,f,0.9566203795860617
,,0,1970-01-01T00:00:59Z,m,0,0,f,0.9599106283927733
,,0,1970-01-01T00:01:00Z,m,0,0,f,0.7293729603954808
,,0,1970-01-01T00:01:01Z,m,0,0,f,0.6455698152977222
,,0,1970-01-01T00:01:02Z,m,0,0,f,0.11441321827059112
,,0,1970-01-01T00:01:03Z,m,0,0,f,0.9955326395256039
,,0,1970-01-01T00:01:04Z,m,0,0,f,0.44266439346958053
,,0,1970-01-01T00:01:05Z,m,0,0,f,0.7183012898949253
,,0,1970-01-01T00:01:06Z,m,0,0,f,0.30706108459030473
,,0,1970-01-01T00:01:07Z,m,0,0,f,0.5034183578538529
,,0,1970-01-01T00:01:08Z,m,0,0,f,0.945541035399725
,,0,1970-01-01T00:01:09Z,m,0,0,f,0.4233995128157775
,,0,1970-01-01T00:01:10Z,m,0,0,f,0.7647066005216012
,,0,1970-01-01T00:01:11Z,m,0,0,f,0.4427721542156412
,,0,1970-01-01T00:01:12Z,m,0,0,f,0.5759588898144714
,,0,1970-01-01T00:01:13Z,m,0,0,f,0.4891738037219912
,,0,1970-01-01T00:01:14Z,m,0,0,f,0.3162573404966396
,,0,1970-01-01T00:01:15Z,m,0,0,f,0.12429098278245032
,,0,1970-01-01T00:01:16Z,m,0,0,f,0.5500314687416078
,,0,1970-01-01T00:01:17Z,m,0,0,f,0.07874290942037632
,,0,1970-01-01T00:01:18Z,m,0,0,f,0.2432131181375912
,,0,1970-01-01T00:01:19Z,m,0,0,f,0.2059157686630176
,,0,1970-01-01T00:01:20Z,m,0,0,f,0.44865547217512164
,,0,1970-01-01T00:01:21Z,m,0,0,f,0.7168101661064027
,,0,1970-01-01T00:01:22Z,m,0,0,f,0.36652553198536764
,,0,1970-01-01T00:01:23Z,m,0,0,f,0.12875338574773973
,,0,1970-01-01T00:01:24Z,m,0,0,f,0.14050907817041347
,,0,1970-01-01T00:01:25Z,m,0,0,f,0.4095172637990756
,,0,1970-01-01T00:01:26Z,m,0,0,f,0.2460700738777719
,,0,1970-01-01T00:01:27Z,m,0,0,f,0.7823912602040078
,,0,1970-01-01T00:01:28Z,m,0,0,f,0.707534534477093
,,0,1970-01-01T00:01:29Z,m,0,0,f,0.6714337668672199
,,0,1970-01-01T00:01:30Z,m,0,0,f,0.6443730852735031
,,0,1970-01-01T00:01:31Z,m,0,0,f,0.8349467641212396
,,0,1970-01-01T00:01:32Z,m,0,0,f,0.7443365385220384
,,0,1970-01-01T00:01:33Z,m,0,0,f,0.778092873581952
,,0,1970-01-01T00:01:34Z,m,0,0,f,0.21451835990529106
,,0,1970-01-01T00:01:35Z,m,0,0,f,0.15132579382756906
,,0,1970-01-01T00:01:36Z,m,0,0,f,0.889690688725347
,,0,1970-01-01T00:01:37Z,m,0,0,f,0.08177608166572663
,,0,1970-01-01T00:01:38Z,m,0,0,f,0.6156947898336163
,,0,1970-01-01T00:01:39Z,m,0,0,f,0.8839098227070676
,,1,1970-01-01T00:00:00Z,m,0,1,f,0.47284307199688513
,,1,1970-01-01T00:00:01Z,m,0,1,f,0.6115110431660992
,,1,1970-01-01T00:00:02Z,m,0,1,f,0.9139676390179812
,,1,1970-01-01T00:00:03Z,m,0,1,f,0.4419580502994864
,,1,1970-01-01T00:00:04Z,m,0,1,f,0.22346720477114235
,,1,1970-01-01T00:00:05Z,m,0,1,f,0.01657253263970824
,,1,1970-01-01T00:00:06Z,m,0,1,f,0.5275526538985256
,,1,1970-01-01T00:00:07Z,m,0,1,f,0.2801453905589357
,,1,1970-01-01T00:00:08Z,m,0,1,f,0.40358058571546174
,,1,1970-01-01T00:00:09Z,m,0,1,f,0.5581225312763497
,,1,1970-01-01T00:00:10Z,m,0,1,f,0.5618381020173508
,,1,1970-01-01T00:00:11Z,m,0,1,f,0.08048303365885615
,,1,1970-01-01T00:00:12Z,m,0,1,f,0.5001751201461243
,,1,1970-01-01T00:00:13Z,m,0,1,f,0.22639175489524663
,,1,1970-01-01T00:00:14Z,m,0,1,f,0.26537476142069744
,,1,1970-01-01T00:00:15Z,m,0,1,f,0.8045352065828273
,,1,1970-01-01T00:00:16Z,m,0,1,f,0.401634967963577
,,1,1970-01-01T00:00:17Z,m,0,1,f,0.9411501472896155
,,1,1970-01-01T00:00:18Z,m,0,1,f,0.2930734491556474
,,1,1970-01-01T00:00:19Z,m,0,1,f,0.18157543568371715
,,1,1970-01-01T00:00:20Z,m,0,1,f,0.9385325130161203
,,1,1970-01-01T00:00:21Z,m,0,1,f,0.17010332650185725
,,1,1970-01-01T00:00:22Z,m,0,1,f,0.04213339793024455
,,1,1970-01-01T00:00:23Z,m,0,1,f,0.5626619227163632
,,1,1970-01-01T00:00:24Z,m,0,1,f,0.6941739177125473
,,1,1970-01-01T00:00:25Z,m,0,1,f,0.5438842736369963
,,1,1970-01-01T00:00:26Z,m,0,1,f,0.6524346931171858
,,1,1970-01-01T00:00:27Z,m,0,1,f,0.062106354006262784
,,1,1970-01-01T00:00:28Z,m,0,1,f,0.6808062354975885
,,1,1970-01-01T00:00:29Z,m,0,1,f,0.4566938577876695
,,1,1970-01-01T00:00:30Z,m,0,1,f,0.15426646385258916
,,1,1970-01-01T00:00:31Z,m,0,1,f,0.7378414694167669
,,1,1970-01-01T00:00:32Z,m,0,1,f,0.35905015546070745
,,1,1970-01-01T00:00:33Z,m,0,1,f,0.25717348995611955
,,1,1970-01-01T00:00:34Z,m,0,1,f,0.8669066045043076
,,1,1970-01-01T00:00:35Z,m,0,1,f,0.7414665987538746
,,1,1970-01-01T00:00:36Z,m,0,1,f,0.7580463272135385
,,1,1970-01-01T00:00:37Z,m,0,1,f,0.223202540983848
,,1,1970-01-01T00:00:38Z,m,0,1,f,0.09675623584194015
,,1,1970-01-01T00:00:39Z,m,0,1,f,0.33037602371875235
,,1,1970-01-01T00:00:40Z,m,0,1,f,0.02419699334564844
,,1,1970-01-01T00:00:41Z,m,0,1,f,0.30660540046813134
,,1,1970-01-01T00:00:42Z,m,0,1,f,0.28087743747358407
,,1,1970-01-01T00:00:43Z,m,0,1,f,0.8125957553254125
,,1,1970-01-01T00:00:44Z,m,0,1,f,0.3996499465775914
,,1,1970-01-01T00:00:45Z,m,0,1,f,0.002859922694346698
,,1,1970-01-01T00:00:46Z,m,0,1,f,0.7743871384683348
,,1,1970-01-01T00:00:47Z,m,0,1,f,0.3428194666142575
,,1,1970-01-01T00:00:48Z,m,0,1,f,0.24529106535786452
,,1,1970-01-01T00:00:49Z,m,0,1,f,0.42074581063787847
,,1,1970-01-01T00:00:50Z,m,0,1,f,0.8230512029974123
,,1,1970-01-01T00:00:51Z,m,0,1,f,0.7612451595826552
,,1,1970-01-01T00:00:52Z,m,0,1,f,0.0025044233308020394
,,1,1970-01-01T00:00:53Z,m,0,1,f,0.8123608833291784
,,1,1970-01-01T00:00:54Z,m,0,1,f,0.094280039506472
,,1,1970-01-01T00:00:55Z,m,0,1,f,0.7414773533860608
,,1,1970-01-01T00:00:56Z,m,0,1,f,0.048248944868655844
,,1,1970-01-01T00:00:57Z,m,0,1,f,0.7876232215876143
,,1,1970-01-01T00:00:58Z,m,0,1,f,0.7708955207540708
,,1,1970-01-01T00:00:59Z,m,0,1,f,0.3210082428062905
,,1,1970-01-01T00:01:00Z,m,0,1,f,0.6199485490487467
,,1,1970-01-01T00:01:01Z,m,0,1,f,0.4526111772487005
,,1,1970-01-01T00:01:02Z,m,0,1,f,0.06993036738408297
,,1,1970-01-01T00:01:03Z,m,0,1,f,0.5391803940621971
,,1,1970-01-01T00:01:04Z,m,0,1,f,0.3786026404218388
,,1,1970-01-01T00:01:05Z,m,0,1,f,0.16987447951514412
,,1,1970-01-01T00:01:06Z,m,0,1,f,0.9622624203254517
,,1,1970-01-01T00:01:07Z,m,0,1,f,0.10609876802280566
,,1,1970-01-01T00:01:08Z,m,0,1,f,0.34039196604520483
,,1,1970-01-01T00:01:09Z,m,0,1,f,0.326997943237989
,,1,1970-01-01T00:01:10Z,m,0,1,f,0.40582069426239586
,,1,1970-01-01T00:01:11Z,m,0,1,f,0.09664389869310906
,,1,1970-01-01T00:01:12Z,m,0,1,f,0.0874716642419619
,,1,1970-01-01T00:01:13Z,m,0,1,f,0.9574787428982809
,,1,1970-01-01T00:01:14Z,m,0,1,f,0.792171281216902
,,1,1970-01-01T00:01:15Z,m,0,1,f,0.8154053514727819
,,1,1970-01-01T00:01:16Z,m,0,1,f,0.9446634309508735
,,1,1970-01-01T00:01:17Z,m,0,1,f,0.7914039734656017
,,1,1970-01-01T00:01:18Z,m,0,1,f,0.5642005948380394
,,1,1970-01-01T00:01:19Z,m,0,1,f,0.9394901508564378
,,1,1970-01-01T00:01:20Z,m,0,1,f,0.09420964672484634
,,1,1970-01-01T00:01:21Z,m,0,1,f,0.8997154088951347
,,1,1970-01-01T00:01:22Z,m,0,1,f,0.8929163087698091
,,1,1970-01-01T00:01:23Z,m,0,1,f,0.14602512562046865
,,1,1970-01-01T00:01:24Z,m,0,1,f,0.061755078411980135
,,1,1970-01-01T00:01:25Z,m,0,1,f,0.050027231315704974
,,1,1970-01-01T00:01:26Z,m,0,1,f,0.06579399435541186
,,1,1970-01-01T00:01:27Z,m,0,1,f,0.5485533330294929
,,1,1970-01-01T00:01:28Z,m,0,1,f,0.08600793471366114
,,1,1970-01-01T00:01:29Z,m,0,1,f,0.0048224932897884395
,,1,1970-01-01T00:01:30Z,m,0,1,f,0.031000679866955753
,,1,1970-01-01T00:01:31Z,m,0,1,f,0.7590758510991269
,,1,1970-01-01T00:01:32Z,m,0,1,f,0.28752964131696107
,,1,1970-01-01T00:01:33Z,m,0,1,f,0.0803113942730073
,,1,1970-01-01T00:01:34Z,m,0,1,f,0.7653660195907919
,,1,1970-01-01T00:01:35Z,m,0,1,f,0.169201547040183
,,1,1970-01-01T00:01:36Z,m,0,1,f,0.2812417370494343
,,1,1970-01-01T00:01:37Z,m,0,1,f,0.5556525309491438
,,1,1970-01-01T00:01:38Z,m,0,1,f,0.21336394958285926
,,1,1970-01-01T00:01:39Z,m,0,1,f,0.843202199200085
,,2,1970-01-01T00:00:00Z,m,1,0,f,0.6745411981120504
,,2,1970-01-01T00:00:01Z,m,1,0,f,0.4341136360856983
,,2,1970-01-01T00:00:02Z,m,1,0,f,0.0779873994184798
,,2,1970-01-01T00:00:03Z,m,1,0,f,0.6045688060594187
,,2,1970-01-01T00:00:04Z,m,1,0,f,0.609806908577383
,,2,1970-01-01T00:00:05Z,m,1,0,f,0.2371373109677929
,,2,1970-01-01T00:00:06Z,m,1,0,f,0.15959047192822226
,,2,1970-01-01T00:00:07Z,m,1,0,f,0.7696930667476671
,,2,1970-01-01T00:00:08Z,m,1,0,f,0.44489788239949923
,,2,1970-01-01T00:00:09Z,m,1,0,f,0.20113730484499945
,,2,1970-01-01T00:00:10Z,m,1,0,f,0.9004310672214374
,,2,1970-01-01T00:00:11Z,m,1,0,f,0.08071979045152104
,,2,1970-01-01T00:00:12Z,m,1,0,f,0.35878401311181407
,,2,1970-01-01T00:00:13Z,m,1,0,f,0.8046013839899406
,,2,1970-01-01T00:00:14Z,m,1,0,f,0.09869242829873062
,,2,1970-01-01T00:00:15Z,m,1,0,f,0.27053244466215826
,,2,1970-01-01T00:00:16Z,m,1,0,f,0.6672055373259661
,,2,1970-01-01T00:00:17Z,m,1,0,f,0.9015798497859395
,,2,1970-01-01T00:00:18Z,m,1,0,f,0.6514438661906353
,,2,1970-01-01T00:00:19Z,m,1,0,f,0.03319201114385362
,,2,1970-01-01T00:00:20Z,m,1,0,f,0.44109087427118215
,,2,1970-01-01T00:00:21Z,m,1,0,f,0.1441063884747634
,,2,1970-01-01T00:00:22Z,m,1,0,f,0.23335939084421864
,,2,1970-01-01T00:00:23Z,m,1,0,f,0.6904277645853616
,,2,1970-01-01T00:00:24Z,m,1,0,f,0.5145930899531316
,,2,1970-01-01T00:00:25Z,m,1,0,f,0.4299752694354613
,,2,1970-01-01T00:00:26Z,m,1,0,f,0.9207494524068397
,,2,1970-01-01T00:00:27Z,m,1,0,f,0.4990764483657634
,,2,1970-01-01T00:00:28Z,m,1,0,f,0.7370053493218158
,,2,1970-01-01T00:00:29Z,m,1,0,f,0.8159190359865772
,,2,1970-01-01T00:00:30Z,m,1,0,f,0.5730300999100897
,,2,1970-01-01T00:00:31Z,m,1,0,f,0.4957548727598841
,,2,1970-01-01T00:00:32Z,m,1,0,f,0.4475722509767004
,,2,1970-01-01T00:00:33Z,m,1,0,f,0.09000105562869058
,,2,1970-01-01T00:00:34Z,m,1,0,f,0.5765896961954948
,,2,1970-01-01T00:00:35Z,m,1,0,f,0.007292186311595296
,,2,1970-01-01T00:00:36Z,m,1,0,f,0.6862338192326899
,,2,1970-01-01T00:00:37Z,m,1,0,f,0.6323091325867545
,,2,1970-01-01T00:00:38Z,m,1,0,f,0.22250144688828086
,,2,1970-01-01T00:00:39Z,m,1,0,f,0.7767158293696542
,,2,1970-01-01T00:00:40Z,m,1,0,f,0.5040765046136644
,,2,1970-01-01T00:00:41Z,m,1,0,f,0.7198824794590694
,,2,1970-01-01T00:00:42Z,m,1,0,f,0.16487220863546403
,,2,1970-01-01T00:00:43Z,m,1,0,f,0.6185190195253291
,,2,1970-01-01T00:00:44Z,m,1,0,f,0.417935209198883
,,2,1970-01-01T00:00:45Z,m,1,0,f,0.143322367253724
,,2,1970-01-01T00:00:46Z,m,1,0,f,0.7110860020844423
,,2,1970-01-01T00:00:47Z,m,1,0,f,0.5190433935276061
,,2,1970-01-01T00:00:48Z,m,1,0,f,0.5947710020498977
,,2,1970-01-01T00:00:49Z,m,1,0,f,0.18632874860445664
,,2,1970-01-01T00:00:50Z,m,1,0,f,0.050671657609869296
,,2,1970-01-01T00:00:51Z,m,1,0,f,0.336667976831678
,,2,1970-01-01T00:00:52Z,m,1,0,f,0.16893598340949662
,,2,1970-01-01T00:00:53Z,m,1,0,f,0.6319794509787114
,,2,1970-01-01T00:00:54Z,m,1,0,f,0.3434433122927547
,,2,1970-01-01T00:00:55Z,m,1,0,f,0.13766344408813833
,,2,1970-01-01T00:00:56Z,m,1,0,f,0.7028890267599247
,,2,1970-01-01T00:00:57Z,m,1,0,f,0.5893915586856076
,,2,1970-01-01T00:00:58Z,m,1,0,f,0.08495375348679511
,,2,1970-01-01T00:00:59Z,m,1,0,f,0.5635570663754376
,,2,1970-01-01T00:01:00Z,m,1,0,f,0.06973804413592974
,,2,1970-01-01T00:01:01Z,m,1,0,f,0.4594087627832006
,,2,1970-01-01T00:01:02Z,m,1,0,f,0.9484143072574632
,,2,1970-01-01T00:01:03Z,m,1,0,f,0.7210862651644585
,,2,1970-01-01T00:01:04Z,m,1,0,f,0.4306492881221061
,,2,1970-01-01T00:01:05Z,m,1,0,f,0.9768511587696722
,,2,1970-01-01T00:01:06Z,m,1,0,f,0.036770411149115535
,,2,1970-01-01T00:01:07Z,m,1,0,f,0.199704171721732
,,2,1970-01-01T00:01:08Z,m,1,0,f,0.044989678879272736
,,2,1970-01-01T00:01:09Z,m,1,0,f,0.4204918747032285
,,2,1970-01-01T00:01:10Z,m,1,0,f,0.7660528673315015
,,2,1970-01-01T00:01:11Z,m,1,0,f,0.07495082447510862
,,2,1970-01-01T00:01:12Z,m,1,0,f,0.979672949703
,,2,1970-01-01T00:01:13Z,m,1,0,f,0.43531431314587743
,,2,1970-01-01T00:01:14Z,m,1,0,f,0.16473009865933294
,,2,1970-01-01T00:01:15Z,m,1,0,f,0.9714924938553514
,,2,1970-01-01T00:01:16Z,m,1,0,f,0.8548205740914873
,,2,1970-01-01T00:01:17Z,m,1,0,f,0.988621458104506
,,2,1970-01-01T00:01:18Z,m,1,0,f,0.42316749552422783
,,2,1970-01-01T00:01:19Z,m,1,0,f,0.5599137447927957
,,2,1970-01-01T00:01:20Z,m,1,0,f,0.7513515954882367
,,2,1970-01-01T00:01:21Z,m,1,0,f,0.07681127373236643
,,2,1970-01-01T00:01:22Z,m,1,0,f,0.04219934813632237
,,2,1970-01-01T00:01:23Z,m,1,0,f,0.27672511415229256
,,2,1970-01-01T00:01:24Z,m,1,0,f,0.6618414211834359
,,2,1970-01-01T00:01:25Z,m,1,0,f,0.04819580958061359
,,2,1970-01-01T00:01:26Z,m,1,0,f,0.8514613397306017
,,2,1970-01-01T00:01:27Z,m,1,0,f,0.654705748814002
,,2,1970-01-01T00:01:28Z,m,1,0,f,0.9967833661484294
,,2,1970-01-01T00:01:29Z,m,1,0,f,0.9631421129969118
,,2,1970-01-01T00:01:30Z,m,1,0,f,0.6286421005881492
,,2,1970-01-01T00:01:31Z,m,1,0,f,0.3783501632738452
,,2,1970-01-01T00:01:32Z,m,1,0,f,0.05114898778086843
,,2,1970-01-01T00:01:33Z,m,1,0,f,0.2473880323048304
,,2,1970-01-01T00:01:34Z,m,1,0,f,0.7842674808782694
,,2,1970-01-01T00:01:35Z,m,1,0,f,0.6130952139646441
,,2,1970-01-01T00:01:36Z,m,1,0,f,0.9762618521418323
,,2,1970-01-01T00:01:37Z,m,1,0,f,0.9219480325346383
,,2,1970-01-01T00:01:38Z,m,1,0,f,0.7986205925631757
,,2,1970-01-01T00:01:39Z,m,1,0,f,0.578541588985068
,,3,1970-01-01T00:00:00Z,m,1,1,f,0.3609497652786835
,,3,1970-01-01T00:00:01Z,m,1,1,f,0.6431495269328852
,,3,1970-01-01T00:00:02Z,m,1,1,f,0.30119517109360755
,,3,1970-01-01T00:00:03Z,m,1,1,f,0.029905756669452933
,,3,1970-01-01T00:00:04Z,m,1,1,f,0.32578997668820153
,,3,1970-01-01T00:00:05Z,m,1,1,f,0.7482046757377168
,,3,1970-01-01T00:00:06Z,m,1,1,f,0.42006674019623874
,,3,1970-01-01T00:00:07Z,m,1,1,f,0.8892383923700209
,,3,1970-01-01T00:00:08Z,m,1,1,f,0.2734890146915862
,,3,1970-01-01T00:00:09Z,m,1,1,f,0.2126705472958595
,,3,1970-01-01T00:00:10Z,m,1,1,f,0.4081541720871348
,,3,1970-01-01T00:00:11Z,m,1,1,f,0.7517886726430452
,,3,1970-01-01T00:00:12Z,m,1,1,f,0.6344255763748975
,,3,1970-01-01T00:00:13Z,m,1,1,f,0.13439033950657941
,,3,1970-01-01T00:00:14Z,m,1,1,f,0.13080770333361982
,,3,1970-01-01T00:00:15Z,m,1,1,f,0.42098106260813917
,,3,1970-01-01T00:00:16Z,m,1,1,f,0.6126625007965338
,,3,1970-01-01T00:00:17Z,m,1,1,f,0.6566130686317417
,,3,1970-01-01T00:00:18Z,m,1,1,f,0.8724405943016941
,,3,1970-01-01T00:00:19Z,m,1,1,f,0.5240118690102152
,,3,1970-01-01T00:00:20Z,m,1,1,f,0.16295208705669978
,,3,1970-01-01T00:00:21Z,m,1,1,f,0.3087465430934554
,,3,1970-01-01T00:00:22Z,m,1,1,f,0.5285274343484349
,,3,1970-01-01T00:00:23Z,m,1,1,f,0.634731960510953
,,3,1970-01-01T00:00:24Z,m,1,1,f,0.21258839107347696
,,3,1970-01-01T00:00:25Z,m,1,1,f,0.418565981182859
,,3,1970-01-01T00:00:26Z,m,1,1,f,0.2537565365571897
,,3,1970-01-01T00:00:27Z,m,1,1,f,0.5464331287426728
,,3,1970-01-01T00:00:28Z,m,1,1,f,0.9960454475764904
,,3,1970-01-01T00:00:29Z,m,1,1,f,0.09275146190386824
,,3,1970-01-01T00:00:30Z,m,1,1,f,0.6976442897720185
,,3,1970-01-01T00:00:31Z,m,1,1,f,0.74713521249196
,,3,1970-01-01T00:00:32Z,m,1,1,f,0.984508958500529
,,3,1970-01-01T00:00:33Z,m,1,1,f,0.735978145078593
,,3,1970-01-01T00:00:34Z,m,1,1,f,0.03272325327489153
,,3,1970-01-01T00:00:35Z,m,1,1,f,0.2789090231376286
,,3,1970-01-01T00:00:36Z,m,1,1,f,0.9009986444969635
,,3,1970-01-01T00:00:37Z,m,1,1,f,0.848311973911401
,,3,1970-01-01T00:00:38Z,m,1,1,f,0.3433130690616337
,,3,1970-01-01T00:00:39Z,m,1,1,f,0.9705860405696857
,,3,1970-01-01T00:00:40Z,m,1,1,f,0.4971554061394775
,,3,1970-01-01T00:00:41Z,m,1,1,f,0.5010737989466268
,,3,1970-01-01T00:00:42Z,m,1,1,f,0.6786336325659156
,,3,1970-01-01T00:00:43Z,m,1,1,f,0.45685893681365386
,,3,1970-01-01T00:00:44Z,m,1,1,f,0.06785712875301617
,,3,1970-01-01T00:00:45Z,m,1,1,f,0.3686928354464234
,,3,1970-01-01T00:00:46Z,m,1,1,f,0.16238519747752908
,,3,1970-01-01T00:00:47Z,m,1,1,f,0.09616346590744834
,,3,1970-01-01T00:00:48Z,m,1,1,f,0.982361090570932
,,3,1970-01-01T00:00:49Z,m,1,1,f,0.24546880258756468
,,3,1970-01-01T00:00:50Z,m,1,1,f,0.4063470659819713
,,3,1970-01-01T00:00:51Z,m,1,1,f,0.02333966735385356
,,3,1970-01-01T00:00:52Z,m,1,1,f,0.7485740576779872
,,3,1970-01-01T00:00:53Z,m,1,1,f,0.6166837184691856
,,3,1970-01-01T00:00:54Z,m,1,1,f,0.05978509722242629
,,3,1970-01-01T00:00:55Z,m,1,1,f,0.8745680789623674
,,3,1970-01-01T00:00:56Z,m,1,1,f,0.7043364028176561
,,3,1970-01-01T00:00:57Z,m,1,1,f,0.5100762819992395
,,3,1970-01-01T00:00:58Z,m,1,1,f,0.16311060736490562
,,3,1970-01-01T00:00:59Z,m,1,1,f,0.8629619678924975
,,3,1970-01-01T00:01:00Z,m,1,1,f,0.10822795841933747
,,3,1970-01-01T00:01:01Z,m,1,1,f,0.009391242035550616
,,3,1970-01-01T00:01:02Z,m,1,1,f,0.8963338627277064
,,3,1970-01-01T00:01:03Z,m,1,1,f,0.2741500937920746
,,3,1970-01-01T00:01:04Z,m,1,1,f,0.8919325188107933
,,3,1970-01-01T00:01:05Z,m,1,1,f,0.6654225234319311
,,3,1970-01-01T00:01:06Z,m,1,1,f,0.02781722451099708
,,3,1970-01-01T00:01:07Z,m,1,1,f,0.1620103430803485
,,3,1970-01-01T00:01:08Z,m,1,1,f,0.4825820756588489
,,3,1970-01-01T00:01:09Z,m,1,1,f,0.6564731088934671
,,3,1970-01-01T00:01:10Z,m,1,1,f,0.5500077260845426
,,3,1970-01-01T00:01:11Z,m,1,1,f,0.40462752766482185
,,3,1970-01-01T00:01:12Z,m,1,1,f,0.8674131498299248
,,3,1970-01-01T00:01:13Z,m,1,1,f,0.8902851603994412
,,3,1970-01-01T00:01:14Z,m,1,1,f,0.1599747356552478
,,3,1970-01-01T00:01:15Z,m,1,1,f,0.4023835778260672
,,3,1970-01-01T00:01:16Z,m,1,1,f,0.8892986579330658
,,3,1970-01-01T00:01:17Z,m,1,1,f,0.05870852811550652
,,3,1970-01-01T00:01:18Z,m,1,1,f,0.08810359195444939
,,3,1970-01-01T00:01:19Z,m,1,1,f,0.5799459169235229
,,3,1970-01-01T00:01:20Z,m,1,1,f,0.675990461828967
,,3,1970-01-01T00:01:21Z,m,1,1,f,0.680028234810394
,,3,1970-01-01T00:01:22Z,m,1,1,f,0.3828707005637953
,,3,1970-01-01T00:01:23Z,m,1,1,f,0.369157111114499
,,3,1970-01-01T00:01:24Z,m,1,1,f,0.12328872455169967
,,3,1970-01-01T00:01:25Z,m,1,1,f,0.43126638642422993
,,3,1970-01-01T00:01:26Z,m,1,1,f,0.24418662053793608
,,3,1970-01-01T00:01:27Z,m,1,1,f,0.22094836458502065
,,3,1970-01-01T00:01:28Z,m,1,1,f,0.10278220106833619
,,3,1970-01-01T00:01:29Z,m,1,1,f,0.7194160988953583
,,3,1970-01-01T00:01:30Z,m,1,1,f,0.9646344422230495
,,3,1970-01-01T00:01:31Z,m,1,1,f,0.462370535565091
,,3,1970-01-01T00:01:32Z,m,1,1,f,0.9386791098643801
,,3,1970-01-01T00:01:33Z,m,1,1,f,0.03801280884674329
,,3,1970-01-01T00:01:34Z,m,1,1,f,0.35603844514090255
,,3,1970-01-01T00:01:35Z,m,1,1,f,0.5083881660913203
,,3,1970-01-01T00:01:36Z,m,1,1,f,0.4326239900843389
,,3,1970-01-01T00:01:37Z,m,1,1,f,0.09453891565081506
,,3,1970-01-01T00:01:38Z,m,1,1,f,0.023503857583366802
,,3,1970-01-01T00:01:39Z,m,1,1,f,0.9492834672803911
"
outData = "
#datatype,string,long,dateTime:RFC3339,string,double
#group,false,false,false,true,false
#default,0,,,,
,result,table,time,_measurement,max
,,0,1970-01-01T00:01:28Z,m,0.9967833661484294
"

// SELECT max(f) FROM m
t_selector = (tables=<-) => tables
    |> range(start: influxql.minTime, stop: influxql.maxTime)
    |> filter(fn: (r) => r._measurement == "m")
    |> filter(fn: (r) => r._field == "f")
    |> group(columns: ["_measurement", "_field"])
    |> max()
    |> keep(columns: ["_time", "_value", "_measurement"])
    |> rename(columns: {_time: "time", _value: "max"})

test _selector = () => ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_selector})
