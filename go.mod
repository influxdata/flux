module github.com/influxdata/flux

go 1.18

require (
	cloud.google.com/go v0.110.0
	cloud.google.com/go/bigtable v1.10.1
	github.com/Azure/go-autorest/autorest v0.11.9
	github.com/Azure/go-autorest/autorest/adal v0.9.13
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.3
	github.com/DATA-DOG/go-sqlmock v1.4.1
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/SAP/go-hdb v0.14.1
	github.com/andreyvit/diff v0.0.0-20170406064948-c7f18ee00883
	github.com/apache/arrow/go/v7 v7.0.1
	github.com/benbjohnson/immutable v0.3.0
	github.com/bonitoo-io/go-sql-bigquery v0.3.4-1.4.0
	github.com/c-bata/go-prompt v0.2.2
	github.com/cespare/xxhash/v2 v2.2.0
	github.com/dave/jennifer v1.2.0
	github.com/denisenkom/go-mssqldb v0.10.0
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/fatih/color v1.13.0
	github.com/foxcpp/go-mockdns v0.0.0-20201212160233-ede2f9158d15
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/golang/geo v0.0.0-20190916061304-5b978397cfec
	github.com/google/flatbuffers v22.9.30-0.20221019131441-5792623df42e+incompatible
	github.com/google/go-cmp v0.5.9
	github.com/influxdata/gosnowflake v1.6.9
	github.com/influxdata/influxdb-client-go/v2 v2.3.1-0.20210518120617-5d1fff431040
	github.com/influxdata/line-protocol v0.0.0-20200327222509-2487e7298839
	github.com/influxdata/line-protocol/v2 v2.2.1
	github.com/influxdata/pkg-config v0.2.11
	github.com/influxdata/tdigest v0.0.2-0.20210216194612-fc98d27c9e8b
	github.com/lib/pq v1.0.0
	github.com/mattn/go-runewidth v0.0.3 // indirect
	github.com/mattn/go-sqlite3 v1.14.18
	github.com/mattn/go-tty v0.0.0-20180907095812-13ff1204f104 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/pkg/term v0.0.0-20180730021639-bffc007b7fd5 // indirect
	github.com/prometheus/client_model v0.1.0
	github.com/prometheus/common v0.7.0
	github.com/segmentio/kafka-go v0.1.0
	github.com/sergi/go-diff v1.0.0 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/stretchr/testify v1.8.1
	github.com/uber/athenadriver v1.1.4
	github.com/uber/jaeger-client-go v2.28.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/vertica/vertica-sql-go v1.1.1
	go.uber.org/zap v1.16.0
	golang.org/x/exp v0.0.0-20220827204233-334a2380cb91
	golang.org/x/net v0.17.0
	golang.org/x/tools v0.6.0
	gonum.org/v1/gonum v0.11.0
	google.golang.org/api v0.114.0
	google.golang.org/grpc v1.56.3
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require github.com/apache/arrow/go/arrow v0.0.0-20211112161151-bc219186db40 // indirect

require github.com/influxdata/influxdb-iox-client-go v1.0.0-beta.1

require (
	cloud.google.com/go/bigquery v1.50.0 // indirect
	cloud.google.com/go/compute v1.19.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/iam v0.13.0 // indirect
	cloud.google.com/go/longrunning v0.4.1 // indirect
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/azure-storage-blob-go v0.14.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.2 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/apache/arrow/go/v11 v11.0.0 // indirect
	github.com/apache/thrift v0.16.0 // indirect
	github.com/aws/aws-sdk-go v1.34.0 // indirect
	github.com/aws/aws-sdk-go-v2 v1.11.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.6.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.5.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.5.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.9.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.19.0 // indirect
	github.com/aws/smithy-go v1.9.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deepmap/oapi-codegen v1.6.0 // indirect
	github.com/dimchansky/utfbom v1.1.0 // indirect
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible // indirect
	github.com/gabriel-vasile/mimetype v1.4.0 // indirect
	github.com/goccy/go-json v0.9.11 // indirect
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.7.1 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/asmfmt v1.3.2 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-ieproxy v0.0.1 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/miekg/dns v1.1.25 // indirect
	github.com/minio/asm2plan9s v0.0.0-20200509001527-cdd76441f9d8 // indirect
	github.com/minio/c2goasm v0.0.0-20190812172519-36a3d3bbc4f3 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/uber-go/tally v3.3.15+incompatible // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/oauth2 v0.7.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)
