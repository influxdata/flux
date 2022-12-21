// DO NOT EDIT: This file is autogenerated via the builtin command.
//
// The imports in this file ensures that all the init functions runs and registers
// the builtins for the flux runtime

package stdlib

import (
	_ "github.com/influxdata/flux/stdlib/array"
	_ "github.com/influxdata/flux/stdlib/bitwise"
	_ "github.com/influxdata/flux/stdlib/contrib/RohanSreerama5/naiveBayesClassifier"
	_ "github.com/influxdata/flux/stdlib/contrib/anaisdg/anomalydetection"
	_ "github.com/influxdata/flux/stdlib/contrib/anaisdg/statsmodels"
	_ "github.com/influxdata/flux/stdlib/contrib/bonitoo-io/alerta"
	_ "github.com/influxdata/flux/stdlib/contrib/bonitoo-io/hex"
	_ "github.com/influxdata/flux/stdlib/contrib/bonitoo-io/servicenow"
	_ "github.com/influxdata/flux/stdlib/contrib/bonitoo-io/tickscript"
	_ "github.com/influxdata/flux/stdlib/contrib/bonitoo-io/victorops"
	_ "github.com/influxdata/flux/stdlib/contrib/bonitoo-io/zenoss"
	_ "github.com/influxdata/flux/stdlib/contrib/chobbs/discord"
	_ "github.com/influxdata/flux/stdlib/contrib/jsternberg/influxdb"
	_ "github.com/influxdata/flux/stdlib/contrib/qxip/logql"
	_ "github.com/influxdata/flux/stdlib/contrib/rhajek/bigpanda"
	_ "github.com/influxdata/flux/stdlib/contrib/sranka/opsgenie"
	_ "github.com/influxdata/flux/stdlib/contrib/sranka/sensu"
	_ "github.com/influxdata/flux/stdlib/contrib/sranka/teams"
	_ "github.com/influxdata/flux/stdlib/contrib/sranka/telegram"
	_ "github.com/influxdata/flux/stdlib/contrib/sranka/webexteams"
	_ "github.com/influxdata/flux/stdlib/contrib/tomhollingworth/events"
	_ "github.com/influxdata/flux/stdlib/csv"
	_ "github.com/influxdata/flux/stdlib/date"
	_ "github.com/influxdata/flux/stdlib/date/boundaries"
	_ "github.com/influxdata/flux/stdlib/dict"
	_ "github.com/influxdata/flux/stdlib/experimental"
	_ "github.com/influxdata/flux/stdlib/experimental/aggregate"
	_ "github.com/influxdata/flux/stdlib/experimental/array"
	_ "github.com/influxdata/flux/stdlib/experimental/bigtable"
	_ "github.com/influxdata/flux/stdlib/experimental/bitwise"
	_ "github.com/influxdata/flux/stdlib/experimental/csv"
	_ "github.com/influxdata/flux/stdlib/experimental/date/boundaries"
	_ "github.com/influxdata/flux/stdlib/experimental/dynamic"
	_ "github.com/influxdata/flux/stdlib/experimental/geo"
	_ "github.com/influxdata/flux/stdlib/experimental/http"
	_ "github.com/influxdata/flux/stdlib/experimental/http/requests"
	_ "github.com/influxdata/flux/stdlib/experimental/influxdb"
	_ "github.com/influxdata/flux/stdlib/experimental/iox"
	_ "github.com/influxdata/flux/stdlib/experimental/json"
	_ "github.com/influxdata/flux/stdlib/experimental/mqtt"
	_ "github.com/influxdata/flux/stdlib/experimental/oee"
	_ "github.com/influxdata/flux/stdlib/experimental/polyline"
	_ "github.com/influxdata/flux/stdlib/experimental/prometheus"
	_ "github.com/influxdata/flux/stdlib/experimental/query"
	_ "github.com/influxdata/flux/stdlib/experimental/record"
	_ "github.com/influxdata/flux/stdlib/experimental/table"
	_ "github.com/influxdata/flux/stdlib/experimental/usage"
	_ "github.com/influxdata/flux/stdlib/generate"
	_ "github.com/influxdata/flux/stdlib/http"
	_ "github.com/influxdata/flux/stdlib/http/requests"
	_ "github.com/influxdata/flux/stdlib/influxdata/influxdb"
	_ "github.com/influxdata/flux/stdlib/influxdata/influxdb/monitor"
	_ "github.com/influxdata/flux/stdlib/influxdata/influxdb/sample"
	_ "github.com/influxdata/flux/stdlib/influxdata/influxdb/schema"
	_ "github.com/influxdata/flux/stdlib/influxdata/influxdb/secrets"
	_ "github.com/influxdata/flux/stdlib/influxdata/influxdb/tasks"
	_ "github.com/influxdata/flux/stdlib/influxdata/influxdb/v1"
	_ "github.com/influxdata/flux/stdlib/internal/boolean"
	_ "github.com/influxdata/flux/stdlib/internal/debug"
	_ "github.com/influxdata/flux/stdlib/internal/gen"
	_ "github.com/influxdata/flux/stdlib/internal/influxql"
	_ "github.com/influxdata/flux/stdlib/internal/location"
	_ "github.com/influxdata/flux/stdlib/internal/promql"
	_ "github.com/influxdata/flux/stdlib/internal/testing"
	_ "github.com/influxdata/flux/stdlib/internal/testutil"
	_ "github.com/influxdata/flux/stdlib/interpolate"
	_ "github.com/influxdata/flux/stdlib/join"
	_ "github.com/influxdata/flux/stdlib/json"
	_ "github.com/influxdata/flux/stdlib/kafka"
	_ "github.com/influxdata/flux/stdlib/math"
	_ "github.com/influxdata/flux/stdlib/pagerduty"
	_ "github.com/influxdata/flux/stdlib/planner"
	_ "github.com/influxdata/flux/stdlib/profiler"
	_ "github.com/influxdata/flux/stdlib/pushbullet"
	_ "github.com/influxdata/flux/stdlib/regexp"
	_ "github.com/influxdata/flux/stdlib/runtime"
	_ "github.com/influxdata/flux/stdlib/sampledata"
	_ "github.com/influxdata/flux/stdlib/slack"
	_ "github.com/influxdata/flux/stdlib/socket"
	_ "github.com/influxdata/flux/stdlib/sql"
	_ "github.com/influxdata/flux/stdlib/strings"
	_ "github.com/influxdata/flux/stdlib/system"
	_ "github.com/influxdata/flux/stdlib/testing"
	_ "github.com/influxdata/flux/stdlib/testing/basics"
	_ "github.com/influxdata/flux/stdlib/testing/chronograf"
	_ "github.com/influxdata/flux/stdlib/testing/expect"
	_ "github.com/influxdata/flux/stdlib/testing/influxql"
	_ "github.com/influxdata/flux/stdlib/testing/kapacitor"
	_ "github.com/influxdata/flux/stdlib/testing/pandas"
	_ "github.com/influxdata/flux/stdlib/testing/prometheus"
	_ "github.com/influxdata/flux/stdlib/testing/promql"
	_ "github.com/influxdata/flux/stdlib/testing/usage"
	_ "github.com/influxdata/flux/stdlib/timezone"
	_ "github.com/influxdata/flux/stdlib/types"
	_ "github.com/influxdata/flux/stdlib/universe"
)
