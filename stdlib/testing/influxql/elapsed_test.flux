package influxql_test


import "testing"
import "internal/influxql"

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,0,,,,,,
,result,table,_time,_measurement,t0,_field,_value
,,0,1970-01-01T00:00:00Z,m,0,f,0.19434194999233168
,,0,1970-01-01T01:00:00Z,m,0,f,0.35586976154169886
,,0,1970-01-01T02:00:00Z,m,0,f,0.9008931119054228
,,0,1970-01-01T03:00:00Z,m,0,f,0.6461505985646413
,,0,1970-01-01T04:00:00Z,m,0,f,0.1340222613556339
,,0,1970-01-01T05:00:00Z,m,0,f,0.3050922896043849
,,0,1970-01-01T06:00:00Z,m,0,f,0.16797790004756785
,,0,1970-01-01T07:00:00Z,m,0,f,0.6859900761088404
,,0,1970-01-01T08:00:00Z,m,0,f,0.3813372334346726
,,0,1970-01-01T09:00:00Z,m,0,f,0.37739800802050527
,,0,1970-01-01T10:00:00Z,m,0,f,0.2670215125945959
,,0,1970-01-01T11:00:00Z,m,0,f,0.19857273235709308
,,0,1970-01-01T12:00:00Z,m,0,f,0.7926413090714327
,,0,1970-01-01T13:00:00Z,m,0,f,0.8488436313118317
,,0,1970-01-01T14:00:00Z,m,0,f,0.1960293435787179
,,0,1970-01-01T15:00:00Z,m,0,f,0.27204741679052236
,,0,1970-01-01T16:00:00Z,m,0,f,0.6045056499409555
,,0,1970-01-01T17:00:00Z,m,0,f,0.21508343480255984
,,0,1970-01-01T18:00:00Z,m,0,f,0.2712545253017199
,,0,1970-01-01T19:00:00Z,m,0,f,0.22728191431845607
,,0,1970-01-01T20:00:00Z,m,0,f,0.8232481787306024
,,0,1970-01-01T21:00:00Z,m,0,f,0.9722054606060748
,,0,1970-01-01T22:00:00Z,m,0,f,0.9332942983017809
,,0,1970-01-01T23:00:00Z,m,0,f,0.009704805042322441
,,0,1970-01-02T00:00:00Z,m,0,f,0.4614776151185129
,,0,1970-01-02T01:00:00Z,m,0,f,0.3972854143424396
,,0,1970-01-02T02:00:00Z,m,0,f,0.024157782439736365
,,0,1970-01-02T03:00:00Z,m,0,f,0.7074351703076142
,,0,1970-01-02T04:00:00Z,m,0,f,0.5819899173941508
,,0,1970-01-02T05:00:00Z,m,0,f,0.2974899730817849
,,0,1970-01-02T06:00:00Z,m,0,f,0.3664899570202347
,,0,1970-01-02T07:00:00Z,m,0,f,0.5666625499409519
,,0,1970-01-02T08:00:00Z,m,0,f,0.2592658730352201
,,0,1970-01-02T09:00:00Z,m,0,f,0.6907206550112025
,,0,1970-01-02T10:00:00Z,m,0,f,0.7184801284027215
,,0,1970-01-02T11:00:00Z,m,0,f,0.363103986952813
,,0,1970-01-02T12:00:00Z,m,0,f,0.938825820840304
,,0,1970-01-02T13:00:00Z,m,0,f,0.7034638846507775
,,0,1970-01-02T14:00:00Z,m,0,f,0.5714903231820487
,,0,1970-01-02T15:00:00Z,m,0,f,0.24449047981396105
,,0,1970-01-02T16:00:00Z,m,0,f,0.14165037565843824
,,0,1970-01-02T17:00:00Z,m,0,f,0.05351135846151062
,,0,1970-01-02T18:00:00Z,m,0,f,0.3450781133356193
,,0,1970-01-02T19:00:00Z,m,0,f,0.23254297482426214
,,0,1970-01-02T20:00:00Z,m,0,f,0.15416851272541165
,,0,1970-01-02T21:00:00Z,m,0,f,0.9287113745228632
,,0,1970-01-02T22:00:00Z,m,0,f,0.8464406026410536
,,0,1970-01-02T23:00:00Z,m,0,f,0.7786237155792206
,,0,1970-01-03T00:00:00Z,m,0,f,0.7222630273842695
,,0,1970-01-03T01:00:00Z,m,0,f,0.5702856518144571
,,0,1970-01-03T02:00:00Z,m,0,f,0.4475020612540418
,,0,1970-01-03T03:00:00Z,m,0,f,0.19482413230523188
,,0,1970-01-03T04:00:00Z,m,0,f,0.14555100659831088
,,0,1970-01-03T05:00:00Z,m,0,f,0.3715313467677773
,,0,1970-01-03T06:00:00Z,m,0,f,0.15710124605981904
,,0,1970-01-03T07:00:00Z,m,0,f,0.05115366925369082
,,0,1970-01-03T08:00:00Z,m,0,f,0.49634673580304356
,,0,1970-01-03T09:00:00Z,m,0,f,0.09850492453963475
,,0,1970-01-03T10:00:00Z,m,0,f,0.07088528667647799
,,0,1970-01-03T11:00:00Z,m,0,f,0.9535958852850828
,,0,1970-01-03T12:00:00Z,m,0,f,0.9473123289831784
,,0,1970-01-03T13:00:00Z,m,0,f,0.6321990998686917
,,0,1970-01-03T14:00:00Z,m,0,f,0.5310985616209651
,,0,1970-01-03T15:00:00Z,m,0,f,0.14010236285353878
,,0,1970-01-03T16:00:00Z,m,0,f,0.5143111322693407
,,0,1970-01-03T17:00:00Z,m,0,f,0.1419555013503121
,,0,1970-01-03T18:00:00Z,m,0,f,0.034988171145264535
,,0,1970-01-03T19:00:00Z,m,0,f,0.4646423361131385
,,0,1970-01-03T20:00:00Z,m,0,f,0.7280775859440926
,,0,1970-01-03T21:00:00Z,m,0,f,0.9605223329866902
,,0,1970-01-03T22:00:00Z,m,0,f,0.6294671473626672
,,0,1970-01-03T23:00:00Z,m,0,f,0.09676486946771183
,,0,1970-01-04T00:00:00Z,m,0,f,0.4846624906255957
,,0,1970-01-04T01:00:00Z,m,0,f,0.9000151629241091
,,0,1970-01-04T02:00:00Z,m,0,f,0.8187520581651648
,,0,1970-01-04T03:00:00Z,m,0,f,0.6356479673253379
,,0,1970-01-04T04:00:00Z,m,0,f,0.9172292568869698
,,0,1970-01-04T05:00:00Z,m,0,f,0.25871413585674596
,,0,1970-01-04T06:00:00Z,m,0,f,0.934030201106989
,,0,1970-01-04T07:00:00Z,m,0,f,0.6300301521545785
,,0,1970-01-04T08:00:00Z,m,0,f,0.9898695895471914
,,0,1970-01-04T09:00:00Z,m,0,f,0.6576532850348832
,,0,1970-01-04T10:00:00Z,m,0,f,0.1095953745610317
,,0,1970-01-04T11:00:00Z,m,0,f,0.20714716664645624
,,0,1970-01-04T12:00:00Z,m,0,f,0.49378319061925324
,,0,1970-01-04T13:00:00Z,m,0,f,0.3244630221410883
,,0,1970-01-04T14:00:00Z,m,0,f,0.1425620337332085
,,0,1970-01-04T15:00:00Z,m,0,f,0.37483772088251627
,,0,1970-01-04T16:00:00Z,m,0,f,0.9386123621523778
,,0,1970-01-04T17:00:00Z,m,0,f,0.2944439301474122
,,0,1970-01-04T18:00:00Z,m,0,f,0.8075592894168399
,,0,1970-01-04T19:00:00Z,m,0,f,0.8131183413273094
,,0,1970-01-04T20:00:00Z,m,0,f,0.6056875144431602
,,0,1970-01-04T21:00:00Z,m,0,f,0.5514021237520469
,,0,1970-01-04T22:00:00Z,m,0,f,0.2904517561416824
,,0,1970-01-04T23:00:00Z,m,0,f,0.7773782053605
,,0,1970-01-05T00:00:00Z,m,0,f,0.1390732850129641
,,0,1970-01-05T01:00:00Z,m,0,f,0.36874812027455345
,,0,1970-01-05T02:00:00Z,m,0,f,0.8497133445947114
,,0,1970-01-05T03:00:00Z,m,0,f,0.2842281672817387
,,0,1970-01-05T04:00:00Z,m,0,f,0.5851186942712497
,,0,1970-01-05T05:00:00Z,m,0,f,0.2754694564842422
,,0,1970-01-05T06:00:00Z,m,0,f,0.03545539694267428
,,0,1970-01-05T07:00:00Z,m,0,f,0.4106208929295988
,,0,1970-01-05T08:00:00Z,m,0,f,0.3680257641839746
,,0,1970-01-05T09:00:00Z,m,0,f,0.7484477843640726
,,0,1970-01-05T10:00:00Z,m,0,f,0.2196945379224781
,,0,1970-01-05T11:00:00Z,m,0,f,0.7377409626382783
,,0,1970-01-05T12:00:00Z,m,0,f,0.4340408821652924
,,0,1970-01-05T13:00:00Z,m,0,f,0.04157784831355819
,,0,1970-01-05T14:00:00Z,m,0,f,0.9005324473445669
,,0,1970-01-05T15:00:00Z,m,0,f,0.6243062492954053
,,0,1970-01-05T16:00:00Z,m,0,f,0.4138274722170456
,,0,1970-01-05T17:00:00Z,m,0,f,0.6559961319794279
,,0,1970-01-05T18:00:00Z,m,0,f,0.09452730201881836
,,0,1970-01-05T19:00:00Z,m,0,f,0.35207875464289057
,,0,1970-01-05T20:00:00Z,m,0,f,0.47000290183266497
,,0,1970-01-05T21:00:00Z,m,0,f,0.13384008497720026
,,0,1970-01-05T22:00:00Z,m,0,f,0.2542495300083506
,,0,1970-01-05T23:00:00Z,m,0,f,0.04357411582677676
,,0,1970-01-06T00:00:00Z,m,0,f,0.2730770850239896
,,0,1970-01-06T01:00:00Z,m,0,f,0.07346719069503016
,,0,1970-01-06T02:00:00Z,m,0,f,0.19296870107837727
,,0,1970-01-06T03:00:00Z,m,0,f,0.8550701670111052
,,0,1970-01-06T04:00:00Z,m,0,f,0.9015279993379257
,,0,1970-01-06T05:00:00Z,m,0,f,0.7681329597853651
,,0,1970-01-06T06:00:00Z,m,0,f,0.13458582961527799
,,0,1970-01-06T07:00:00Z,m,0,f,0.5025964032341974
,,0,1970-01-06T08:00:00Z,m,0,f,0.9660611150198847
,,0,1970-01-06T09:00:00Z,m,0,f,0.7406756350132208
,,0,1970-01-06T10:00:00Z,m,0,f,0.48245323402069856
,,0,1970-01-06T11:00:00Z,m,0,f,0.5396866678590079
,,0,1970-01-06T12:00:00Z,m,0,f,0.24056787192459894
,,0,1970-01-06T13:00:00Z,m,0,f,0.5473495899891297
,,0,1970-01-06T14:00:00Z,m,0,f,0.9939487519980328
,,0,1970-01-06T15:00:00Z,m,0,f,0.7718086454038607
,,0,1970-01-06T16:00:00Z,m,0,f,0.3729231862915519
,,0,1970-01-06T17:00:00Z,m,0,f,0.978216628089757
,,0,1970-01-06T18:00:00Z,m,0,f,0.30410501498270626
,,0,1970-01-06T19:00:00Z,m,0,f,0.36293525766110357
,,0,1970-01-06T20:00:00Z,m,0,f,0.45673893698213724
,,0,1970-01-06T21:00:00Z,m,0,f,0.42887470039944864
,,0,1970-01-06T22:00:00Z,m,0,f,0.42264444401794515
,,0,1970-01-06T23:00:00Z,m,0,f,0.3061909271178175
,,0,1970-01-07T00:00:00Z,m,0,f,0.6681291175687905
,,0,1970-01-07T01:00:00Z,m,0,f,0.5494108420781338
,,0,1970-01-07T02:00:00Z,m,0,f,0.31779594303648045
,,0,1970-01-07T03:00:00Z,m,0,f,0.22502703712265368
,,0,1970-01-07T04:00:00Z,m,0,f,0.03498146847868716
,,0,1970-01-07T05:00:00Z,m,0,f,0.16139395876022747
,,0,1970-01-07T06:00:00Z,m,0,f,0.6335318955521227
,,0,1970-01-07T07:00:00Z,m,0,f,0.5854967453622169
,,0,1970-01-07T08:00:00Z,m,0,f,0.43015814365562627
,,0,1970-01-07T09:00:00Z,m,0,f,0.07215482648098204
,,0,1970-01-07T10:00:00Z,m,0,f,0.09348412983453618
,,0,1970-01-07T11:00:00Z,m,0,f,0.9023793546915768
,,0,1970-01-07T12:00:00Z,m,0,f,0.9055451292861832
,,0,1970-01-07T13:00:00Z,m,0,f,0.3280454144164272
,,0,1970-01-07T14:00:00Z,m,0,f,0.05897468763156862
,,0,1970-01-07T15:00:00Z,m,0,f,0.3686339026679373
,,0,1970-01-07T16:00:00Z,m,0,f,0.7547173975990482
,,0,1970-01-07T17:00:00Z,m,0,f,0.457847526142958
,,0,1970-01-07T18:00:00Z,m,0,f,0.5038320054556072
,,0,1970-01-07T19:00:00Z,m,0,f,0.47058145000588336
,,0,1970-01-07T20:00:00Z,m,0,f,0.5333903317331339
,,0,1970-01-07T21:00:00Z,m,0,f,0.1548508614296064
,,0,1970-01-07T22:00:00Z,m,0,f,0.6837681053869291
,,0,1970-01-07T23:00:00Z,m,0,f,0.9081953381867953
,,1,1970-01-01T00:00:00Z,m,1,f,0.15129694889144107
,,1,1970-01-01T01:00:00Z,m,1,f,0.18038761353721244
,,1,1970-01-01T02:00:00Z,m,1,f,0.23198629938985071
,,1,1970-01-01T03:00:00Z,m,1,f,0.4940776062344333
,,1,1970-01-01T04:00:00Z,m,1,f,0.5654050390735228
,,1,1970-01-01T05:00:00Z,m,1,f,0.3788291715942209
,,1,1970-01-01T06:00:00Z,m,1,f,0.39178743939497507
,,1,1970-01-01T07:00:00Z,m,1,f,0.573740997246541
,,1,1970-01-01T08:00:00Z,m,1,f,0.6171205083791419
,,1,1970-01-01T09:00:00Z,m,1,f,0.2562012267655005
,,1,1970-01-01T10:00:00Z,m,1,f,0.41301351982023743
,,1,1970-01-01T11:00:00Z,m,1,f,0.335808747696944
,,1,1970-01-01T12:00:00Z,m,1,f,0.25034171949067086
,,1,1970-01-01T13:00:00Z,m,1,f,0.9866289864317817
,,1,1970-01-01T14:00:00Z,m,1,f,0.42988399575215924
,,1,1970-01-01T15:00:00Z,m,1,f,0.02602624797587471
,,1,1970-01-01T16:00:00Z,m,1,f,0.9926232260423908
,,1,1970-01-01T17:00:00Z,m,1,f,0.9771153046566231
,,1,1970-01-01T18:00:00Z,m,1,f,0.5680196566957276
,,1,1970-01-01T19:00:00Z,m,1,f,0.01952645919207055
,,1,1970-01-01T20:00:00Z,m,1,f,0.3439692491089684
,,1,1970-01-01T21:00:00Z,m,1,f,0.15596143014601407
,,1,1970-01-01T22:00:00Z,m,1,f,0.7986983212658367
,,1,1970-01-01T23:00:00Z,m,1,f,0.31336565203700295
,,1,1970-01-02T00:00:00Z,m,1,f,0.6398281383647288
,,1,1970-01-02T01:00:00Z,m,1,f,0.14018673322595193
,,1,1970-01-02T02:00:00Z,m,1,f,0.2847409792344233
,,1,1970-01-02T03:00:00Z,m,1,f,0.4295460864480138
,,1,1970-01-02T04:00:00Z,m,1,f,0.9674016258565854
,,1,1970-01-02T05:00:00Z,m,1,f,0.108837862280129
,,1,1970-01-02T06:00:00Z,m,1,f,0.47129460971856907
,,1,1970-01-02T07:00:00Z,m,1,f,0.9175708860682784
,,1,1970-01-02T08:00:00Z,m,1,f,0.3383504562747057
,,1,1970-01-02T09:00:00Z,m,1,f,0.7176237840014899
,,1,1970-01-02T10:00:00Z,m,1,f,0.45631599181081023
,,1,1970-01-02T11:00:00Z,m,1,f,0.58210555704762
,,1,1970-01-02T12:00:00Z,m,1,f,0.44833346180841194
,,1,1970-01-02T13:00:00Z,m,1,f,0.847082665931482
,,1,1970-01-02T14:00:00Z,m,1,f,0.1032050849659337
,,1,1970-01-02T15:00:00Z,m,1,f,0.6342038875836871
,,1,1970-01-02T16:00:00Z,m,1,f,0.47157138392000586
,,1,1970-01-02T17:00:00Z,m,1,f,0.5939195811492147
,,1,1970-01-02T18:00:00Z,m,1,f,0.3907003938279841
,,1,1970-01-02T19:00:00Z,m,1,f,0.3737781066004461
,,1,1970-01-02T20:00:00Z,m,1,f,0.6059179847188622
,,1,1970-01-02T21:00:00Z,m,1,f,0.37459130316766875
,,1,1970-01-02T22:00:00Z,m,1,f,0.529020795101784
,,1,1970-01-02T23:00:00Z,m,1,f,0.5797965259387311
,,1,1970-01-03T00:00:00Z,m,1,f,0.4196060336001739
,,1,1970-01-03T01:00:00Z,m,1,f,0.4423826236661577
,,1,1970-01-03T02:00:00Z,m,1,f,0.7562185239602677
,,1,1970-01-03T03:00:00Z,m,1,f,0.29641000596052747
,,1,1970-01-03T04:00:00Z,m,1,f,0.5511866012217823
,,1,1970-01-03T05:00:00Z,m,1,f,0.477231168882557
,,1,1970-01-03T06:00:00Z,m,1,f,0.5783604476492074
,,1,1970-01-03T07:00:00Z,m,1,f,0.6087147255603924
,,1,1970-01-03T08:00:00Z,m,1,f,0.9779728651411874
,,1,1970-01-03T09:00:00Z,m,1,f,0.8559123961968673
,,1,1970-01-03T10:00:00Z,m,1,f,0.039322803759977897
,,1,1970-01-03T11:00:00Z,m,1,f,0.5107877963474311
,,1,1970-01-03T12:00:00Z,m,1,f,0.36939734036661503
,,1,1970-01-03T13:00:00Z,m,1,f,0.24036834333350818
,,1,1970-01-03T14:00:00Z,m,1,f,0.9041140297145132
,,1,1970-01-03T15:00:00Z,m,1,f,0.3088634061697057
,,1,1970-01-03T16:00:00Z,m,1,f,0.3391757217065211
,,1,1970-01-03T17:00:00Z,m,1,f,0.5709032014080667
,,1,1970-01-03T18:00:00Z,m,1,f,0.023692334151288443
,,1,1970-01-03T19:00:00Z,m,1,f,0.9283397254805887
,,1,1970-01-03T20:00:00Z,m,1,f,0.7897301020744532
,,1,1970-01-03T21:00:00Z,m,1,f,0.5499067643037981
,,1,1970-01-03T22:00:00Z,m,1,f,0.20359811467533634
,,1,1970-01-03T23:00:00Z,m,1,f,0.1946255400705282
,,1,1970-01-04T00:00:00Z,m,1,f,0.44702956746887096
,,1,1970-01-04T01:00:00Z,m,1,f,0.44634342940951505
,,1,1970-01-04T02:00:00Z,m,1,f,0.4462164964469759
,,1,1970-01-04T03:00:00Z,m,1,f,0.5245740015591633
,,1,1970-01-04T04:00:00Z,m,1,f,0.29252555227190247
,,1,1970-01-04T05:00:00Z,m,1,f,0.5137169576742285
,,1,1970-01-04T06:00:00Z,m,1,f,0.1624473579380766
,,1,1970-01-04T07:00:00Z,m,1,f,0.30153697909681254
,,1,1970-01-04T08:00:00Z,m,1,f,0.2324327035115191
,,1,1970-01-04T09:00:00Z,m,1,f,0.034393197916253775
,,1,1970-01-04T10:00:00Z,m,1,f,0.4336629996115634
,,1,1970-01-04T11:00:00Z,m,1,f,0.8790573703532555
,,1,1970-01-04T12:00:00Z,m,1,f,0.9016824143089478
,,1,1970-01-04T13:00:00Z,m,1,f,0.34003737969744235
,,1,1970-01-04T14:00:00Z,m,1,f,0.3848952908759773
,,1,1970-01-04T15:00:00Z,m,1,f,0.9951718603202089
,,1,1970-01-04T16:00:00Z,m,1,f,0.8567450174592717
,,1,1970-01-04T17:00:00Z,m,1,f,0.12389207874832112
,,1,1970-01-04T18:00:00Z,m,1,f,0.6712865769046611
,,1,1970-01-04T19:00:00Z,m,1,f,0.46454363710822305
,,1,1970-01-04T20:00:00Z,m,1,f,0.9625945392247928
,,1,1970-01-04T21:00:00Z,m,1,f,0.7535558804101941
,,1,1970-01-04T22:00:00Z,m,1,f,0.744281664085344
,,1,1970-01-04T23:00:00Z,m,1,f,0.6811372884190415
,,1,1970-01-05T00:00:00Z,m,1,f,0.46171144508557443
,,1,1970-01-05T01:00:00Z,m,1,f,0.7701860606472665
,,1,1970-01-05T02:00:00Z,m,1,f,0.25517367370396854
,,1,1970-01-05T03:00:00Z,m,1,f,0.5564394982112523
,,1,1970-01-05T04:00:00Z,m,1,f,0.18256039263141344
,,1,1970-01-05T05:00:00Z,m,1,f,0.08465044152492789
,,1,1970-01-05T06:00:00Z,m,1,f,0.04682876596739505
,,1,1970-01-05T07:00:00Z,m,1,f,0.5116535677666431
,,1,1970-01-05T08:00:00Z,m,1,f,0.26327513076438025
,,1,1970-01-05T09:00:00Z,m,1,f,0.8551637599549397
,,1,1970-01-05T10:00:00Z,m,1,f,0.04908769638903045
,,1,1970-01-05T11:00:00Z,m,1,f,0.6747954667852788
,,1,1970-01-05T12:00:00Z,m,1,f,0.6701210820394512
,,1,1970-01-05T13:00:00Z,m,1,f,0.6698146693971668
,,1,1970-01-05T14:00:00Z,m,1,f,0.32939712697857165
,,1,1970-01-05T15:00:00Z,m,1,f,0.788384711857412
,,1,1970-01-05T16:00:00Z,m,1,f,0.9435078647906675
,,1,1970-01-05T17:00:00Z,m,1,f,0.05526759807741008
,,1,1970-01-05T18:00:00Z,m,1,f,0.3040576381882256
,,1,1970-01-05T19:00:00Z,m,1,f,0.13057573237533082
,,1,1970-01-05T20:00:00Z,m,1,f,0.438829781443743
,,1,1970-01-05T21:00:00Z,m,1,f,0.16639381298657024
,,1,1970-01-05T22:00:00Z,m,1,f,0.17817868556539768
,,1,1970-01-05T23:00:00Z,m,1,f,0.37006948631938175
,,1,1970-01-06T00:00:00Z,m,1,f,0.7711386953356921
,,1,1970-01-06T01:00:00Z,m,1,f,0.37364593618845465
,,1,1970-01-06T02:00:00Z,m,1,f,0.9285996064937719
,,1,1970-01-06T03:00:00Z,m,1,f,0.8685918613936688
,,1,1970-01-06T04:00:00Z,m,1,f,0.049757835180659744
,,1,1970-01-06T05:00:00Z,m,1,f,0.3562051567466768
,,1,1970-01-06T06:00:00Z,m,1,f,0.9028928456702144
,,1,1970-01-06T07:00:00Z,m,1,f,0.45412719022597203
,,1,1970-01-06T08:00:00Z,m,1,f,0.5210991958721604
,,1,1970-01-06T09:00:00Z,m,1,f,0.5013716125947244
,,1,1970-01-06T10:00:00Z,m,1,f,0.7798859934672562
,,1,1970-01-06T11:00:00Z,m,1,f,0.20777334301449937
,,1,1970-01-06T12:00:00Z,m,1,f,0.12979889080684515
,,1,1970-01-06T13:00:00Z,m,1,f,0.6713165183217583
,,1,1970-01-06T14:00:00Z,m,1,f,0.5267649385791876
,,1,1970-01-06T15:00:00Z,m,1,f,0.2766996970172108
,,1,1970-01-06T16:00:00Z,m,1,f,0.837561303602128
,,1,1970-01-06T17:00:00Z,m,1,f,0.10692091027423688
,,1,1970-01-06T18:00:00Z,m,1,f,0.16161417900026617
,,1,1970-01-06T19:00:00Z,m,1,f,0.7596615857389895
,,1,1970-01-06T20:00:00Z,m,1,f,0.9033476318497203
,,1,1970-01-06T21:00:00Z,m,1,f,0.9281794553091864
,,1,1970-01-06T22:00:00Z,m,1,f,0.7691815845690406
,,1,1970-01-06T23:00:00Z,m,1,f,0.5713941284458292
,,1,1970-01-07T00:00:00Z,m,1,f,0.8319045908167892
,,1,1970-01-07T01:00:00Z,m,1,f,0.5839200214729727
,,1,1970-01-07T02:00:00Z,m,1,f,0.5597883274306116
,,1,1970-01-07T03:00:00Z,m,1,f,0.8448107197504592
,,1,1970-01-07T04:00:00Z,m,1,f,0.39141999130543037
,,1,1970-01-07T05:00:00Z,m,1,f,0.3151057211763145
,,1,1970-01-07T06:00:00Z,m,1,f,0.3812489036241129
,,1,1970-01-07T07:00:00Z,m,1,f,0.03893545284960627
,,1,1970-01-07T08:00:00Z,m,1,f,0.513934438417237
,,1,1970-01-07T09:00:00Z,m,1,f,0.07387412770693513
,,1,1970-01-07T10:00:00Z,m,1,f,0.16131994851623296
,,1,1970-01-07T11:00:00Z,m,1,f,0.8524873225734262
,,1,1970-01-07T12:00:00Z,m,1,f,0.7108229805824855
,,1,1970-01-07T13:00:00Z,m,1,f,0.4087372331379091
,,1,1970-01-07T14:00:00Z,m,1,f,0.5408493060971712
,,1,1970-01-07T15:00:00Z,m,1,f,0.8752116934130074
,,1,1970-01-07T16:00:00Z,m,1,f,0.9569196248412628
,,1,1970-01-07T17:00:00Z,m,1,f,0.5206668595695829
,,1,1970-01-07T18:00:00Z,m,1,f,0.012847952493292788
,,1,1970-01-07T19:00:00Z,m,1,f,0.7155605509853933
,,1,1970-01-07T20:00:00Z,m,1,f,0.8293273149090988
,,1,1970-01-07T21:00:00Z,m,1,f,0.38705272903958904
,,1,1970-01-07T22:00:00Z,m,1,f,0.5459991408731746
,,1,1970-01-07T23:00:00Z,m,1,f,0.7066840478612406
"
outData = "
#datatype,string,long,dateTime:RFC3339,string,string,long
#group,false,false,false,true,true,false
#default,0,,,,,
,result,table,time,_measurement,t0,elapsed
,,0,1970-01-01T01:00:00Z,m,0,3600000000000
,,0,1970-01-01T02:00:00Z,m,0,3600000000000
,,0,1970-01-01T03:00:00Z,m,0,3600000000000
,,0,1970-01-01T04:00:00Z,m,0,3600000000000
,,0,1970-01-01T05:00:00Z,m,0,3600000000000
,,0,1970-01-01T06:00:00Z,m,0,3600000000000
,,0,1970-01-01T07:00:00Z,m,0,3600000000000
,,0,1970-01-01T08:00:00Z,m,0,3600000000000
,,0,1970-01-01T09:00:00Z,m,0,3600000000000
,,0,1970-01-01T10:00:00Z,m,0,3600000000000
,,0,1970-01-01T11:00:00Z,m,0,3600000000000
,,0,1970-01-01T12:00:00Z,m,0,3600000000000
,,0,1970-01-01T13:00:00Z,m,0,3600000000000
,,0,1970-01-01T14:00:00Z,m,0,3600000000000
,,0,1970-01-01T15:00:00Z,m,0,3600000000000
,,0,1970-01-01T16:00:00Z,m,0,3600000000000
,,0,1970-01-01T17:00:00Z,m,0,3600000000000
,,0,1970-01-01T18:00:00Z,m,0,3600000000000
,,0,1970-01-01T19:00:00Z,m,0,3600000000000
,,0,1970-01-01T20:00:00Z,m,0,3600000000000
,,0,1970-01-01T21:00:00Z,m,0,3600000000000
,,0,1970-01-01T22:00:00Z,m,0,3600000000000
,,0,1970-01-01T23:00:00Z,m,0,3600000000000
,,0,1970-01-02T00:00:00Z,m,0,3600000000000
,,0,1970-01-02T01:00:00Z,m,0,3600000000000
,,0,1970-01-02T02:00:00Z,m,0,3600000000000
,,0,1970-01-02T03:00:00Z,m,0,3600000000000
,,0,1970-01-02T04:00:00Z,m,0,3600000000000
,,0,1970-01-02T05:00:00Z,m,0,3600000000000
,,0,1970-01-02T06:00:00Z,m,0,3600000000000
,,0,1970-01-02T07:00:00Z,m,0,3600000000000
,,0,1970-01-02T08:00:00Z,m,0,3600000000000
,,0,1970-01-02T09:00:00Z,m,0,3600000000000
,,0,1970-01-02T10:00:00Z,m,0,3600000000000
,,0,1970-01-02T11:00:00Z,m,0,3600000000000
,,0,1970-01-02T12:00:00Z,m,0,3600000000000
,,0,1970-01-02T13:00:00Z,m,0,3600000000000
,,0,1970-01-02T14:00:00Z,m,0,3600000000000
,,0,1970-01-02T15:00:00Z,m,0,3600000000000
,,0,1970-01-02T16:00:00Z,m,0,3600000000000
,,0,1970-01-02T17:00:00Z,m,0,3600000000000
,,0,1970-01-02T18:00:00Z,m,0,3600000000000
,,0,1970-01-02T19:00:00Z,m,0,3600000000000
,,0,1970-01-02T20:00:00Z,m,0,3600000000000
,,0,1970-01-02T21:00:00Z,m,0,3600000000000
,,0,1970-01-02T22:00:00Z,m,0,3600000000000
,,0,1970-01-02T23:00:00Z,m,0,3600000000000
,,0,1970-01-03T00:00:00Z,m,0,3600000000000
,,0,1970-01-03T01:00:00Z,m,0,3600000000000
,,0,1970-01-03T02:00:00Z,m,0,3600000000000
,,0,1970-01-03T03:00:00Z,m,0,3600000000000
,,0,1970-01-03T04:00:00Z,m,0,3600000000000
,,0,1970-01-03T05:00:00Z,m,0,3600000000000
,,0,1970-01-03T06:00:00Z,m,0,3600000000000
,,0,1970-01-03T07:00:00Z,m,0,3600000000000
,,0,1970-01-03T08:00:00Z,m,0,3600000000000
,,0,1970-01-03T09:00:00Z,m,0,3600000000000
,,0,1970-01-03T10:00:00Z,m,0,3600000000000
,,0,1970-01-03T11:00:00Z,m,0,3600000000000
,,0,1970-01-03T12:00:00Z,m,0,3600000000000
,,0,1970-01-03T13:00:00Z,m,0,3600000000000
,,0,1970-01-03T14:00:00Z,m,0,3600000000000
,,0,1970-01-03T15:00:00Z,m,0,3600000000000
,,0,1970-01-03T16:00:00Z,m,0,3600000000000
,,0,1970-01-03T17:00:00Z,m,0,3600000000000
,,0,1970-01-03T18:00:00Z,m,0,3600000000000
,,0,1970-01-03T19:00:00Z,m,0,3600000000000
,,0,1970-01-03T20:00:00Z,m,0,3600000000000
,,0,1970-01-03T21:00:00Z,m,0,3600000000000
,,0,1970-01-03T22:00:00Z,m,0,3600000000000
,,0,1970-01-03T23:00:00Z,m,0,3600000000000
,,0,1970-01-04T00:00:00Z,m,0,3600000000000
,,0,1970-01-04T01:00:00Z,m,0,3600000000000
,,0,1970-01-04T02:00:00Z,m,0,3600000000000
,,0,1970-01-04T03:00:00Z,m,0,3600000000000
,,0,1970-01-04T04:00:00Z,m,0,3600000000000
,,0,1970-01-04T05:00:00Z,m,0,3600000000000
,,0,1970-01-04T06:00:00Z,m,0,3600000000000
,,0,1970-01-04T07:00:00Z,m,0,3600000000000
,,0,1970-01-04T08:00:00Z,m,0,3600000000000
,,0,1970-01-04T09:00:00Z,m,0,3600000000000
,,0,1970-01-04T10:00:00Z,m,0,3600000000000
,,0,1970-01-04T11:00:00Z,m,0,3600000000000
,,0,1970-01-04T12:00:00Z,m,0,3600000000000
,,0,1970-01-04T13:00:00Z,m,0,3600000000000
,,0,1970-01-04T14:00:00Z,m,0,3600000000000
,,0,1970-01-04T15:00:00Z,m,0,3600000000000
,,0,1970-01-04T16:00:00Z,m,0,3600000000000
,,0,1970-01-04T17:00:00Z,m,0,3600000000000
,,0,1970-01-04T18:00:00Z,m,0,3600000000000
,,0,1970-01-04T19:00:00Z,m,0,3600000000000
,,0,1970-01-04T20:00:00Z,m,0,3600000000000
,,0,1970-01-04T21:00:00Z,m,0,3600000000000
,,0,1970-01-04T22:00:00Z,m,0,3600000000000
,,0,1970-01-04T23:00:00Z,m,0,3600000000000
,,0,1970-01-05T00:00:00Z,m,0,3600000000000
,,0,1970-01-05T01:00:00Z,m,0,3600000000000
,,0,1970-01-05T02:00:00Z,m,0,3600000000000
,,0,1970-01-05T03:00:00Z,m,0,3600000000000
,,0,1970-01-05T04:00:00Z,m,0,3600000000000
,,0,1970-01-05T05:00:00Z,m,0,3600000000000
,,0,1970-01-05T06:00:00Z,m,0,3600000000000
,,0,1970-01-05T07:00:00Z,m,0,3600000000000
,,0,1970-01-05T08:00:00Z,m,0,3600000000000
,,0,1970-01-05T09:00:00Z,m,0,3600000000000
,,0,1970-01-05T10:00:00Z,m,0,3600000000000
,,0,1970-01-05T11:00:00Z,m,0,3600000000000
,,0,1970-01-05T12:00:00Z,m,0,3600000000000
,,0,1970-01-05T13:00:00Z,m,0,3600000000000
,,0,1970-01-05T14:00:00Z,m,0,3600000000000
,,0,1970-01-05T15:00:00Z,m,0,3600000000000
,,0,1970-01-05T16:00:00Z,m,0,3600000000000
,,0,1970-01-05T17:00:00Z,m,0,3600000000000
,,0,1970-01-05T18:00:00Z,m,0,3600000000000
,,0,1970-01-05T19:00:00Z,m,0,3600000000000
,,0,1970-01-05T20:00:00Z,m,0,3600000000000
,,0,1970-01-05T21:00:00Z,m,0,3600000000000
,,0,1970-01-05T22:00:00Z,m,0,3600000000000
,,0,1970-01-05T23:00:00Z,m,0,3600000000000
,,0,1970-01-06T00:00:00Z,m,0,3600000000000
,,0,1970-01-06T01:00:00Z,m,0,3600000000000
,,0,1970-01-06T02:00:00Z,m,0,3600000000000
,,0,1970-01-06T03:00:00Z,m,0,3600000000000
,,0,1970-01-06T04:00:00Z,m,0,3600000000000
,,0,1970-01-06T05:00:00Z,m,0,3600000000000
,,0,1970-01-06T06:00:00Z,m,0,3600000000000
,,0,1970-01-06T07:00:00Z,m,0,3600000000000
,,0,1970-01-06T08:00:00Z,m,0,3600000000000
,,0,1970-01-06T09:00:00Z,m,0,3600000000000
,,0,1970-01-06T10:00:00Z,m,0,3600000000000
,,0,1970-01-06T11:00:00Z,m,0,3600000000000
,,0,1970-01-06T12:00:00Z,m,0,3600000000000
,,0,1970-01-06T13:00:00Z,m,0,3600000000000
,,0,1970-01-06T14:00:00Z,m,0,3600000000000
,,0,1970-01-06T15:00:00Z,m,0,3600000000000
,,0,1970-01-06T16:00:00Z,m,0,3600000000000
,,0,1970-01-06T17:00:00Z,m,0,3600000000000
,,0,1970-01-06T18:00:00Z,m,0,3600000000000
,,0,1970-01-06T19:00:00Z,m,0,3600000000000
,,0,1970-01-06T20:00:00Z,m,0,3600000000000
,,0,1970-01-06T21:00:00Z,m,0,3600000000000
,,0,1970-01-06T22:00:00Z,m,0,3600000000000
,,0,1970-01-06T23:00:00Z,m,0,3600000000000
,,0,1970-01-07T00:00:00Z,m,0,3600000000000
,,0,1970-01-07T01:00:00Z,m,0,3600000000000
,,0,1970-01-07T02:00:00Z,m,0,3600000000000
,,0,1970-01-07T03:00:00Z,m,0,3600000000000
,,0,1970-01-07T04:00:00Z,m,0,3600000000000
,,0,1970-01-07T05:00:00Z,m,0,3600000000000
,,0,1970-01-07T06:00:00Z,m,0,3600000000000
,,0,1970-01-07T07:00:00Z,m,0,3600000000000
,,0,1970-01-07T08:00:00Z,m,0,3600000000000
,,0,1970-01-07T09:00:00Z,m,0,3600000000000
,,0,1970-01-07T10:00:00Z,m,0,3600000000000
,,0,1970-01-07T11:00:00Z,m,0,3600000000000
,,0,1970-01-07T12:00:00Z,m,0,3600000000000
,,0,1970-01-07T13:00:00Z,m,0,3600000000000
,,0,1970-01-07T14:00:00Z,m,0,3600000000000
,,0,1970-01-07T15:00:00Z,m,0,3600000000000
,,0,1970-01-07T16:00:00Z,m,0,3600000000000
,,0,1970-01-07T17:00:00Z,m,0,3600000000000
,,0,1970-01-07T18:00:00Z,m,0,3600000000000
,,0,1970-01-07T19:00:00Z,m,0,3600000000000
,,0,1970-01-07T20:00:00Z,m,0,3600000000000
,,0,1970-01-07T21:00:00Z,m,0,3600000000000
,,0,1970-01-07T22:00:00Z,m,0,3600000000000
,,0,1970-01-07T23:00:00Z,m,0,3600000000000
,,1,1970-01-01T01:00:00Z,m,1,3600000000000
,,1,1970-01-01T02:00:00Z,m,1,3600000000000
,,1,1970-01-01T03:00:00Z,m,1,3600000000000
,,1,1970-01-01T04:00:00Z,m,1,3600000000000
,,1,1970-01-01T05:00:00Z,m,1,3600000000000
,,1,1970-01-01T06:00:00Z,m,1,3600000000000
,,1,1970-01-01T07:00:00Z,m,1,3600000000000
,,1,1970-01-01T08:00:00Z,m,1,3600000000000
,,1,1970-01-01T09:00:00Z,m,1,3600000000000
,,1,1970-01-01T10:00:00Z,m,1,3600000000000
,,1,1970-01-01T11:00:00Z,m,1,3600000000000
,,1,1970-01-01T12:00:00Z,m,1,3600000000000
,,1,1970-01-01T13:00:00Z,m,1,3600000000000
,,1,1970-01-01T14:00:00Z,m,1,3600000000000
,,1,1970-01-01T15:00:00Z,m,1,3600000000000
,,1,1970-01-01T16:00:00Z,m,1,3600000000000
,,1,1970-01-01T17:00:00Z,m,1,3600000000000
,,1,1970-01-01T18:00:00Z,m,1,3600000000000
,,1,1970-01-01T19:00:00Z,m,1,3600000000000
,,1,1970-01-01T20:00:00Z,m,1,3600000000000
,,1,1970-01-01T21:00:00Z,m,1,3600000000000
,,1,1970-01-01T22:00:00Z,m,1,3600000000000
,,1,1970-01-01T23:00:00Z,m,1,3600000000000
,,1,1970-01-02T00:00:00Z,m,1,3600000000000
,,1,1970-01-02T01:00:00Z,m,1,3600000000000
,,1,1970-01-02T02:00:00Z,m,1,3600000000000
,,1,1970-01-02T03:00:00Z,m,1,3600000000000
,,1,1970-01-02T04:00:00Z,m,1,3600000000000
,,1,1970-01-02T05:00:00Z,m,1,3600000000000
,,1,1970-01-02T06:00:00Z,m,1,3600000000000
,,1,1970-01-02T07:00:00Z,m,1,3600000000000
,,1,1970-01-02T08:00:00Z,m,1,3600000000000
,,1,1970-01-02T09:00:00Z,m,1,3600000000000
,,1,1970-01-02T10:00:00Z,m,1,3600000000000
,,1,1970-01-02T11:00:00Z,m,1,3600000000000
,,1,1970-01-02T12:00:00Z,m,1,3600000000000
,,1,1970-01-02T13:00:00Z,m,1,3600000000000
,,1,1970-01-02T14:00:00Z,m,1,3600000000000
,,1,1970-01-02T15:00:00Z,m,1,3600000000000
,,1,1970-01-02T16:00:00Z,m,1,3600000000000
,,1,1970-01-02T17:00:00Z,m,1,3600000000000
,,1,1970-01-02T18:00:00Z,m,1,3600000000000
,,1,1970-01-02T19:00:00Z,m,1,3600000000000
,,1,1970-01-02T20:00:00Z,m,1,3600000000000
,,1,1970-01-02T21:00:00Z,m,1,3600000000000
,,1,1970-01-02T22:00:00Z,m,1,3600000000000
,,1,1970-01-02T23:00:00Z,m,1,3600000000000
,,1,1970-01-03T00:00:00Z,m,1,3600000000000
,,1,1970-01-03T01:00:00Z,m,1,3600000000000
,,1,1970-01-03T02:00:00Z,m,1,3600000000000
,,1,1970-01-03T03:00:00Z,m,1,3600000000000
,,1,1970-01-03T04:00:00Z,m,1,3600000000000
,,1,1970-01-03T05:00:00Z,m,1,3600000000000
,,1,1970-01-03T06:00:00Z,m,1,3600000000000
,,1,1970-01-03T07:00:00Z,m,1,3600000000000
,,1,1970-01-03T08:00:00Z,m,1,3600000000000
,,1,1970-01-03T09:00:00Z,m,1,3600000000000
,,1,1970-01-03T10:00:00Z,m,1,3600000000000
,,1,1970-01-03T11:00:00Z,m,1,3600000000000
,,1,1970-01-03T12:00:00Z,m,1,3600000000000
,,1,1970-01-03T13:00:00Z,m,1,3600000000000
,,1,1970-01-03T14:00:00Z,m,1,3600000000000
,,1,1970-01-03T15:00:00Z,m,1,3600000000000
,,1,1970-01-03T16:00:00Z,m,1,3600000000000
,,1,1970-01-03T17:00:00Z,m,1,3600000000000
,,1,1970-01-03T18:00:00Z,m,1,3600000000000
,,1,1970-01-03T19:00:00Z,m,1,3600000000000
,,1,1970-01-03T20:00:00Z,m,1,3600000000000
,,1,1970-01-03T21:00:00Z,m,1,3600000000000
,,1,1970-01-03T22:00:00Z,m,1,3600000000000
,,1,1970-01-03T23:00:00Z,m,1,3600000000000
,,1,1970-01-04T00:00:00Z,m,1,3600000000000
,,1,1970-01-04T01:00:00Z,m,1,3600000000000
,,1,1970-01-04T02:00:00Z,m,1,3600000000000
,,1,1970-01-04T03:00:00Z,m,1,3600000000000
,,1,1970-01-04T04:00:00Z,m,1,3600000000000
,,1,1970-01-04T05:00:00Z,m,1,3600000000000
,,1,1970-01-04T06:00:00Z,m,1,3600000000000
,,1,1970-01-04T07:00:00Z,m,1,3600000000000
,,1,1970-01-04T08:00:00Z,m,1,3600000000000
,,1,1970-01-04T09:00:00Z,m,1,3600000000000
,,1,1970-01-04T10:00:00Z,m,1,3600000000000
,,1,1970-01-04T11:00:00Z,m,1,3600000000000
,,1,1970-01-04T12:00:00Z,m,1,3600000000000
,,1,1970-01-04T13:00:00Z,m,1,3600000000000
,,1,1970-01-04T14:00:00Z,m,1,3600000000000
,,1,1970-01-04T15:00:00Z,m,1,3600000000000
,,1,1970-01-04T16:00:00Z,m,1,3600000000000
,,1,1970-01-04T17:00:00Z,m,1,3600000000000
,,1,1970-01-04T18:00:00Z,m,1,3600000000000
,,1,1970-01-04T19:00:00Z,m,1,3600000000000
,,1,1970-01-04T20:00:00Z,m,1,3600000000000
,,1,1970-01-04T21:00:00Z,m,1,3600000000000
,,1,1970-01-04T22:00:00Z,m,1,3600000000000
,,1,1970-01-04T23:00:00Z,m,1,3600000000000
,,1,1970-01-05T00:00:00Z,m,1,3600000000000
,,1,1970-01-05T01:00:00Z,m,1,3600000000000
,,1,1970-01-05T02:00:00Z,m,1,3600000000000
,,1,1970-01-05T03:00:00Z,m,1,3600000000000
,,1,1970-01-05T04:00:00Z,m,1,3600000000000
,,1,1970-01-05T05:00:00Z,m,1,3600000000000
,,1,1970-01-05T06:00:00Z,m,1,3600000000000
,,1,1970-01-05T07:00:00Z,m,1,3600000000000
,,1,1970-01-05T08:00:00Z,m,1,3600000000000
,,1,1970-01-05T09:00:00Z,m,1,3600000000000
,,1,1970-01-05T10:00:00Z,m,1,3600000000000
,,1,1970-01-05T11:00:00Z,m,1,3600000000000
,,1,1970-01-05T12:00:00Z,m,1,3600000000000
,,1,1970-01-05T13:00:00Z,m,1,3600000000000
,,1,1970-01-05T14:00:00Z,m,1,3600000000000
,,1,1970-01-05T15:00:00Z,m,1,3600000000000
,,1,1970-01-05T16:00:00Z,m,1,3600000000000
,,1,1970-01-05T17:00:00Z,m,1,3600000000000
,,1,1970-01-05T18:00:00Z,m,1,3600000000000
,,1,1970-01-05T19:00:00Z,m,1,3600000000000
,,1,1970-01-05T20:00:00Z,m,1,3600000000000
,,1,1970-01-05T21:00:00Z,m,1,3600000000000
,,1,1970-01-05T22:00:00Z,m,1,3600000000000
,,1,1970-01-05T23:00:00Z,m,1,3600000000000
,,1,1970-01-06T00:00:00Z,m,1,3600000000000
,,1,1970-01-06T01:00:00Z,m,1,3600000000000
,,1,1970-01-06T02:00:00Z,m,1,3600000000000
,,1,1970-01-06T03:00:00Z,m,1,3600000000000
,,1,1970-01-06T04:00:00Z,m,1,3600000000000
,,1,1970-01-06T05:00:00Z,m,1,3600000000000
,,1,1970-01-06T06:00:00Z,m,1,3600000000000
,,1,1970-01-06T07:00:00Z,m,1,3600000000000
,,1,1970-01-06T08:00:00Z,m,1,3600000000000
,,1,1970-01-06T09:00:00Z,m,1,3600000000000
,,1,1970-01-06T10:00:00Z,m,1,3600000000000
,,1,1970-01-06T11:00:00Z,m,1,3600000000000
,,1,1970-01-06T12:00:00Z,m,1,3600000000000
,,1,1970-01-06T13:00:00Z,m,1,3600000000000
,,1,1970-01-06T14:00:00Z,m,1,3600000000000
,,1,1970-01-06T15:00:00Z,m,1,3600000000000
,,1,1970-01-06T16:00:00Z,m,1,3600000000000
,,1,1970-01-06T17:00:00Z,m,1,3600000000000
,,1,1970-01-06T18:00:00Z,m,1,3600000000000
,,1,1970-01-06T19:00:00Z,m,1,3600000000000
,,1,1970-01-06T20:00:00Z,m,1,3600000000000
,,1,1970-01-06T21:00:00Z,m,1,3600000000000
,,1,1970-01-06T22:00:00Z,m,1,3600000000000
,,1,1970-01-06T23:00:00Z,m,1,3600000000000
,,1,1970-01-07T00:00:00Z,m,1,3600000000000
,,1,1970-01-07T01:00:00Z,m,1,3600000000000
,,1,1970-01-07T02:00:00Z,m,1,3600000000000
,,1,1970-01-07T03:00:00Z,m,1,3600000000000
,,1,1970-01-07T04:00:00Z,m,1,3600000000000
,,1,1970-01-07T05:00:00Z,m,1,3600000000000
,,1,1970-01-07T06:00:00Z,m,1,3600000000000
,,1,1970-01-07T07:00:00Z,m,1,3600000000000
,,1,1970-01-07T08:00:00Z,m,1,3600000000000
,,1,1970-01-07T09:00:00Z,m,1,3600000000000
,,1,1970-01-07T10:00:00Z,m,1,3600000000000
,,1,1970-01-07T11:00:00Z,m,1,3600000000000
,,1,1970-01-07T12:00:00Z,m,1,3600000000000
,,1,1970-01-07T13:00:00Z,m,1,3600000000000
,,1,1970-01-07T14:00:00Z,m,1,3600000000000
,,1,1970-01-07T15:00:00Z,m,1,3600000000000
,,1,1970-01-07T16:00:00Z,m,1,3600000000000
,,1,1970-01-07T17:00:00Z,m,1,3600000000000
,,1,1970-01-07T18:00:00Z,m,1,3600000000000
,,1,1970-01-07T19:00:00Z,m,1,3600000000000
,,1,1970-01-07T20:00:00Z,m,1,3600000000000
,,1,1970-01-07T21:00:00Z,m,1,3600000000000
,,1,1970-01-07T22:00:00Z,m,1,3600000000000
,,1,1970-01-07T23:00:00Z,m,1,3600000000000
"

// SELECT elapsed(f) FROM m GROUP BY *
t_elapsed = (tables=<-) => tables
    |> range(start: influxql.minTime, stop: influxql.maxTime)
    |> filter(fn: (r) => r._measurement == "m" and r._field == "f")
    |> elapsed(unit: 1ns)
    |> drop(columns: ["_start", "_stop", "_field", "_value"])
    |> rename(columns: {_time: "time"})

test _elapsed = () => ({
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    fn: t_elapsed,
})
