package join_test

import (
	"context"
	"testing"

	arrowmem "github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/join"
	"github.com/influxdata/flux/values"
)

var (
	leftID  execute.DatasetID
	rightID execute.DatasetID
)

func init() {
	leftID = executetest.RandomDatasetID()
	rightID = executetest.RandomDatasetID()
}

func TestMergeJoin(t *testing.T) {
	testCases := []struct {
		name         string
		on           []join.ColumnPair
		as           string
		method       string
		left, right  []table.Chunk
		wantTables   []table.Chunk
		wantErrLeft  error
		wantErrRight error
	}{
		{
			name: "exclude groupkey column",
			wantErrRight: errors.New(
				codes.Invalid,
				"join cannot modify group key: output record has a missing or invalid value for column 'group:uint'",
			),
			method: "inner",
			on: []join.ColumnPair{
				{Left: "label", Right: "id"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => ({_time: l._time, lv: l._value, rv: r._value, label: l.label})`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 1.2, "label": "a", "group": uint64(1)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(1), "id": "a", "group": uint64(1)},
				},
			),
		},
		{
			name: "bad column name in on parameter",
			wantErrRight: errors.New(
				codes.Invalid,
				"cannot set join columns in right table stream: table is missing column 'name'",
			),
			method: "inner",
			on: []join.ColumnPair{
				{Left: "label", Right: "name"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => ({_time: l._time, lv: l._value, rv: r._value, label: l.label, group: l.group})`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
				},
			),
			wantTables: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "lv", Type: flux.TFloat},
					{Label: "rv", Type: flux.TInt},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(1)},
				},
			),
		},
		{
			name:   "inner one group one chunk",
			method: "inner",
			on: []join.ColumnPair{
				{Left: "label", Right: "id"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => ({_time: l._time, lv: l._value, rv: r._value, label: l.label, group: l.group})`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(1)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(1)},
				},
			),
			wantTables: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "lv", Type: flux.TFloat},
					{Label: "rv", Type: flux.TInt},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(1)},
				},
			),
		},
		{
			name:   "left one group one chunk",
			method: "left",
			on: []join.ColumnPair{
				{Left: "label", Right: "id"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => ({_time: l._time, lv: l._value, rv: r._value, label: l.label, group: l.group})`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(1)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(1)},
				},
			),
			wantTables: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "lv", Type: flux.TFloat},
					{Label: "rv", Type: flux.TInt},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 9.0, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 1.9, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 2.8, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 3.7, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(1)},
				},
			),
		},
		{
			name:   "right one group one chunk",
			method: "right",
			on: []join.ColumnPair{
				{Left: "label", Right: "id"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => ({_time: r._time, lv: l._value, rv: r._value, label: r.id, group: r.group})`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(1)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(1)},
				},
			),
			wantTables: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "lv", Type: flux.TFloat},
					{Label: "rv", Type: flux.TInt},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": values.Null, "rv": int64(5), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": values.Null, "rv": int64(6), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": values.Null, "rv": int64(7), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": values.Null, "rv": int64(8), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(1)},
				},
			),
		},
		{
			name:   "full one group one chunk",
			method: "full",
			on: []join.ColumnPair{
				{Left: "label", Right: "id"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => {
        label = if exists l.label then l.label else r.id
        time = if exists l._time then l._time else r._time

        return {
            label: label,
            lv: l._value,
            group: l.group,
            rv: r._value,
            _time: time,
        }
			}`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(1)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(1)},
				},
			),
			wantTables: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "lv", Type: flux.TFloat},
					{Label: "rv", Type: flux.TInt},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 9.0, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 1.9, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 2.8, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 3.7, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": values.Null, "rv": int64(5), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": values.Null, "rv": int64(6), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": values.Null, "rv": int64(7), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": values.Null, "rv": int64(8), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(1)},
				},
			),
		},
		{
			name:   "inner one group multi chunk",
			method: "inner",
			on: []join.ColumnPair{
				{Left: "label", Right: "id"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => ({_time: l._time, lv: l._value, rv: r._value, label: l.label, group: l.group})`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(4), "_value": 0.1, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 0.2, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 0.3, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(5), "_value": 0.4, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 0.1, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 0.2, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 0.3, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 0.4, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 0.1, "label": "f", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 0.2, "label": "f", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 0.3, "label": "f", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 0.4, "label": "f", "group": uint64(1)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(4), "_value": int64(13), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(14), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(15), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(5), "_value": int64(16), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(17), "id": "f", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(18), "id": "f", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(19), "id": "f", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(20), "id": "f", "group": uint64(1)},
				},
			),
			wantTables: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "lv", Type: flux.TFloat},
					{Label: "rv", Type: flux.TInt},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(13), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(14), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(15), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.1, "rv": int64(12), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.1, "rv": int64(13), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.1, "rv": int64(14), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.1, "rv": int64(15), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.2, "rv": int64(12), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.2, "rv": int64(13), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.2, "rv": int64(14), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.2, "rv": int64(15), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.3, "rv": int64(12), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.3, "rv": int64(13), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.3, "rv": int64(14), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.3, "rv": int64(15), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(5), "lv": 0.4, "rv": int64(16), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 0.1, "rv": int64(17), "label": "f", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 0.2, "rv": int64(18), "label": "f", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 0.3, "rv": int64(19), "label": "f", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.4, "rv": int64(20), "label": "f", "group": uint64(1)},
				},
			),
		},
		{
			name:   "left one group multi chunk",
			method: "left",
			on: []join.ColumnPair{
				{Left: "label", Right: "id"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => ({_time: l._time, lv: l._value, rv: r._value, label: l.label, group: l.group})`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 0.1, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 0.2, "label": "e", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(2), "_value": 0.25, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 0.3, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 0.4, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 0.1, "label": "f", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 0.2, "label": "f", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 0.3, "label": "f", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 0.4, "label": "f", "group": uint64(1)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(17), "id": "f", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(18), "id": "f", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(19), "id": "f", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(20), "id": "f", "group": uint64(1)},
				},
			),
			wantTables: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "lv", Type: flux.TFloat},
					{Label: "rv", Type: flux.TInt},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 9.0, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 1.9, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 2.8, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 3.7, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 0.1, "rv": values.Null, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 0.2, "rv": values.Null, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 0.25, "rv": values.Null, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 0.3, "rv": values.Null, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.4, "rv": values.Null, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 0.1, "rv": int64(17), "label": "f", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 0.2, "rv": int64(18), "label": "f", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 0.3, "rv": int64(19), "label": "f", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.4, "rv": int64(20), "label": "f", "group": uint64(1)},
				},
			),
		},
		{
			name:   "right one group multi chunk",
			method: "right",
			on: []join.ColumnPair{
				{Left: "label", Right: "id"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => ({_time: r._time, lv: l._value, rv: r._value, label: r.id, group: r.group})`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 0.1, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 0.2, "label": "e", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(2), "_value": 0.25, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 0.3, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 0.4, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 0.1, "label": "f", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 0.2, "label": "f", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 0.3, "label": "f", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 0.4, "label": "f", "group": uint64(1)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(2), "_value": int64(66), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(67), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(17), "id": "f", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(18), "id": "f", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(19), "id": "f", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(20), "id": "f", "group": uint64(1)},
				},
			),
			wantTables: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "lv", Type: flux.TFloat},
					{Label: "rv", Type: flux.TInt},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": values.Null, "rv": int64(5), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": values.Null, "rv": int64(6), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": values.Null, "rv": int64(66), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": values.Null, "rv": int64(67), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": values.Null, "rv": int64(7), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": values.Null, "rv": int64(8), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 0.1, "rv": int64(17), "label": "f", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 0.2, "rv": int64(18), "label": "f", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 0.3, "rv": int64(19), "label": "f", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.4, "rv": int64(20), "label": "f", "group": uint64(1)},
				},
			),
		},
		{
			name:   "full one group multi chunk",
			method: "full",
			on: []join.ColumnPair{
				{Left: "label", Right: "id"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => {
        label = if exists l.label then l.label else r.id
        time = if exists l._time then l._time else r._time

        return {
            label: label,
            lv: l._value,
            group: l.group,
            rv: r._value,
            _time: time,
        }
			}`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 0.1, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 0.2, "label": "e", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(2), "_value": 0.25, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 0.3, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 0.4, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 0.1, "label": "f", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 0.2, "label": "f", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 0.3, "label": "f", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 0.4, "label": "f", "group": uint64(1)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(2), "_value": int64(66), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(67), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(17), "id": "f", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(18), "id": "f", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(19), "id": "f", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(20), "id": "f", "group": uint64(1)},
				},
			),
			wantTables: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "lv", Type: flux.TFloat},
					{Label: "rv", Type: flux.TInt},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 9.0, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 1.9, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 2.8, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 3.7, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": values.Null, "rv": int64(5), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": values.Null, "rv": int64(6), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": values.Null, "rv": int64(66), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": values.Null, "rv": int64(67), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": values.Null, "rv": int64(7), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": values.Null, "rv": int64(8), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 0.1, "rv": values.Null, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 0.2, "rv": values.Null, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 0.25, "rv": values.Null, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 0.3, "rv": values.Null, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.4, "rv": values.Null, "label": "e", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 0.1, "rv": int64(17), "label": "f", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 0.2, "rv": int64(18), "label": "f", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 0.3, "rv": int64(19), "label": "f", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 0.4, "rv": int64(20), "label": "f", "group": uint64(1)},
				},
			),
		},
		{
			name:   "inner multi group multi chunk",
			method: "inner",
			on: []join.ColumnPair{
				{Left: "label", Right: "id"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => ({_time: l._time, lv: l._value, rv: r._value, label: l.label, group: l.group})`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(2)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(2)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(2)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(2)},
				},
			),
			wantTables: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "lv", Type: flux.TFloat},
					{Label: "rv", Type: flux.TInt},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(2)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(2)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(2)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(2)},
				},
			),
		},
		{
			name:   "left multi group multi chunk",
			method: "left",
			on: []join.ColumnPair{
				{Left: "label", Right: "id"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => ({_time: l._time, lv: l._value, rv: r._value, label: l.label, group: l.group})`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(2)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(2)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(2)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(2)},
				},
			),
			wantTables: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "lv", Type: flux.TFloat},
					{Label: "rv", Type: flux.TInt},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 9.0, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 1.9, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 2.8, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 3.7, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(1), "lv": 9.0, "rv": values.Null, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(2), "lv": 1.9, "rv": values.Null, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(3), "lv": 2.8, "rv": values.Null, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(4), "lv": 3.7, "rv": values.Null, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(2)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(2)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(2)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(2)},
				},
			),
		},
		{
			name:   "right multi group multi chunk",
			method: "right",
			on: []join.ColumnPair{
				{Left: "label", Right: "id"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => ({_time: r._time, lv: l._value, rv: r._value, label: r.id, group: r.group})`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(2)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(2)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(2)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(2)},
				},
			),
			wantTables: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "lv", Type: flux.TFloat},
					{Label: "rv", Type: flux.TInt},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": values.Null, "rv": int64(5), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": values.Null, "rv": int64(6), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": values.Null, "rv": int64(7), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": values.Null, "rv": int64(8), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(1), "lv": values.Null, "rv": int64(5), "label": "c", "group": uint64(2)},
					{"_time": execute.Time(2), "lv": values.Null, "rv": int64(6), "label": "c", "group": uint64(2)},
					{"_time": execute.Time(3), "lv": values.Null, "rv": int64(7), "label": "c", "group": uint64(2)},
					{"_time": execute.Time(4), "lv": values.Null, "rv": int64(8), "label": "c", "group": uint64(2)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(2)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(2)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(2)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(2)},
				},
			),
		},
		{
			name:   "full multi group multi chunk",
			method: "full",
			on: []join.ColumnPair{
				{Left: "label", Right: "id"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => {
        label = if exists l.label then l.label else r.id
        time = if exists l._time then l._time else r._time

        return {
            label: label,
            lv: l._value,
            group: l.group,
            rv: r._value,
            _time: time,
        }
			}`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": 3.4, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": 5.6, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": 7.8, "label": "a", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": 9.0, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": 1.9, "label": "b", "group": uint64(2)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": 2.8, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": 3.7, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": 4.6, "label": "d", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": 5.5, "label": "d", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": 1.3, "label": "d", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": 2.4, "label": "d", "group": uint64(2)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": int64(2), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": int64(3), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": int64(4), "id": "a", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": int64(5), "id": "c", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": int64(6), "id": "c", "group": uint64(2)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(3), "_value": int64(7), "id": "c", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": int64(8), "id": "c", "group": uint64(2)},
					{"_time": execute.Time(1), "_value": int64(9), "id": "d", "group": uint64(2)},
					{"_time": execute.Time(2), "_value": int64(10), "id": "d", "group": uint64(2)},
					{"_time": execute.Time(3), "_value": int64(11), "id": "d", "group": uint64(2)},
					{"_time": execute.Time(4), "_value": int64(12), "id": "d", "group": uint64(2)},
				},
			),
			wantTables: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "lv", Type: flux.TFloat},
					{Label: "rv", Type: flux.TInt},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 9.0, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 1.9, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 2.8, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 3.7, "rv": values.Null, "label": "b", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": values.Null, "rv": int64(5), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": values.Null, "rv": int64(6), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": values.Null, "rv": int64(7), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": values.Null, "rv": int64(8), "label": "c", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(1)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(2), "lv": 3.4, "rv": int64(2), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(3), "lv": 5.6, "rv": int64(3), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(4), "lv": 7.8, "rv": int64(4), "label": "a", "group": uint64(2)},
					{"_time": execute.Time(1), "lv": 9.0, "rv": values.Null, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(2), "lv": 1.9, "rv": values.Null, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(3), "lv": 2.8, "rv": values.Null, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(4), "lv": 3.7, "rv": values.Null, "label": "b", "group": uint64(2)},
					{"_time": execute.Time(1), "lv": values.Null, "rv": int64(5), "label": "c", "group": uint64(2)},
					{"_time": execute.Time(2), "lv": values.Null, "rv": int64(6), "label": "c", "group": uint64(2)},
					{"_time": execute.Time(3), "lv": values.Null, "rv": int64(7), "label": "c", "group": uint64(2)},
					{"_time": execute.Time(4), "lv": values.Null, "rv": int64(8), "label": "c", "group": uint64(2)},
					{"_time": execute.Time(1), "lv": 4.6, "rv": int64(9), "label": "d", "group": uint64(2)},
					{"_time": execute.Time(2), "lv": 5.5, "rv": int64(10), "label": "d", "group": uint64(2)},
					{"_time": execute.Time(3), "lv": 1.3, "rv": int64(11), "label": "d", "group": uint64(2)},
					{"_time": execute.Time(4), "lv": 2.4, "rv": int64(12), "label": "d", "group": uint64(2)},
				},
			),
		},
		{
			name:   "cross product single row chunks",
			method: "full",
			on: []join.ColumnPair{
				{Left: "label", Right: "id"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => ({_time: l._time, lv: l._value, rv: r._value, label: l.label, group: l.group})`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 2.1, "label": "a", "group": uint64(1)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(2), "id": "a", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(3), "id": "a", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(4), "id": "a", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(5), "id": "a", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(6), "id": "a", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(7), "id": "a", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(8), "id": "a", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(9), "id": "a", "group": uint64(1)},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(10), "id": "a", "group": uint64(1)},
				},
			),
			wantTables: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "lv", Type: flux.TFloat},
					{Label: "rv", Type: flux.TInt},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(1), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(2), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(3), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(4), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(5), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(6), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(7), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(8), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(9), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 1.2, "rv": int64(10), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 2.1, "rv": int64(1), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 2.1, "rv": int64(2), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 2.1, "rv": int64(3), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 2.1, "rv": int64(4), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 2.1, "rv": int64(5), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 2.1, "rv": int64(6), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 2.1, "rv": int64(7), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 2.1, "rv": int64(8), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 2.1, "rv": int64(9), "label": "a", "group": uint64(1)},
					{"_time": execute.Time(1), "lv": 2.1, "rv": int64(10), "label": "a", "group": uint64(1)},
				},
			),
		},
		{
			name: "change group key column",
			wantErrRight: errors.New(
				codes.Invalid,
				"join cannot modify group key: output record has a missing or invalid value for column 'group:uint'",
			),
			method: "inner",
			on: []join.ColumnPair{
				{Left: "label", Right: "id"},
				{Left: "_time", Right: "_time"},
			},
			as: `(l, r) => ({_time: l._time, lv: l._value, rv: r._value, label: l.label, group: 3})`,
			left: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "label", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": 1.2, "label": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": 1.2, "label": "a", "group": uint64(1)},
				},
			),
			right: constructChunks(
				[]flux.ColMeta{{Label: "group", Type: flux.TUInt}},
				[]flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "id", Type: flux.TString},
					{Label: "group", Type: flux.TUInt},
				},
				[]map[string]interface{}{
					{"_time": execute.Time(1), "_value": int64(1), "id": "a", "group": uint64(1)},
					{"_time": execute.Time(2), "_value": int64(1), "id": "a", "group": uint64(1)},
				},
			),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn, err := fnFromSrc(tc.as)
			if err != nil {
				t.Errorf("got unexpected error: %s", err)
			}
			spec := join.SortMergeJoinProcedureSpec{
				On:     tc.on,
				As:     *fn,
				Method: tc.method,
			}
			id := executetest.RandomDatasetID()
			checked := arrowmem.NewCheckedAllocator(memory.DefaultAllocator)
			mem := memory.NewResourceAllocator(checked)

			defer checked.AssertSize(t, 0)

			mjt, err := join.NewMergeJoinTransformation(
				context.Background(),
				id,
				&spec,
				leftID,
				rightID,
				mem,
			)
			if err != nil {
				t.Errorf("got unexpected error: %s", err)
			}

			dataset := mjt.Dataset()
			store := executetest.NewDataStore()
			dataset.AddTransformation(store)
			tr := execute.NewTransformationFromTransport(mjt)

			leftDataset := execute.NewTransportDataset(leftID, mem)
			leftDataset.AddTransformation(tr)

			rightDataset := execute.NewTransportDataset(rightID, mem)
			rightDataset.AddTransformation(tr)

			process := func(chunks []table.Chunk, d *execute.TransportDataset) error {
				for _, chunk := range chunks {
					err = d.Process(chunk)
					if err != nil {
						return err
					}
				}
				return nil
			}

			err = process(tc.left, leftDataset)
			if err != nil {
				if tc.wantErrLeft != nil {
					if tc.wantErrLeft.Error() != err.Error() {
						t.Fatalf("expected error: '%s' - got: '%s'", tc.wantErrLeft, err)
					}
				} else {
					t.Fatalf("got unexpected error: '%s'", err)
				}
			} else {
				if tc.wantErrLeft != nil {
					t.Fatalf("expected error: '%s' - got: none", tc.wantErrLeft)
				}
			}
			tr.Finish(leftID, nil)

			err = process(tc.right, rightDataset)
			if err != nil {
				if tc.wantErrRight != nil {
					if tc.wantErrRight.Error() != err.Error() {
						t.Fatalf("expected error: %s - got: %s", tc.wantErrRight, err)
					}
					return
				} else {
					t.Fatalf("got unexpected error: %s", err)
				}
			} else {
				if tc.wantErrRight != nil {
					t.Fatalf("expected error `%s` - got: none", tc.wantErrRight)
				}
			}
			tr.Finish(rightID, nil)

			for _, tbl := range tc.wantTables {
				wantBuf := tbl.Buffer()
				gotTbl, err := store.Table(wantBuf.Key())
				if err != nil {
					t.Errorf("got unexpected error: %s", err)
				}
				want := table.Stringify(table.FromBuffer(&wantBuf))
				got := table.Stringify(gotTbl)
				if !cmp.Equal(want, got) {
					t.Errorf("table chunks differ, -want/+got:\n%v",
						cmp.Diff(want, got))
				}
			}
		})
	}
}

func fnFromSrc(src string) (*interpreter.ResolvedFunction, error) {
	pkg, err := runtime.AnalyzeSource(context.Background(), src)
	if err != nil {
		return nil, err
	}

	exprStmt, ok := pkg.Files[0].Body[0].(*semantic.ExpressionStatement)
	if !ok {
		return nil, errors.New(codes.Inherit, "not an expression statement")
	}
	fnNode, ok := exprStmt.Expression.(*semantic.FunctionExpression)
	if !ok {
		return nil, errors.New(codes.Inherit, "bad function expression")
	}

	return &interpreter.ResolvedFunction{
		Fn:    fnNode,
		Scope: values.NewScope(),
	}, nil
}

func constructChunks(keyCols, cols []flux.ColMeta, tables ...[]map[string]interface{}) []table.Chunk {
	chunks := make([]table.Chunk, len(tables))
	for i, table := range tables {
		keyVals := make([]values.Value, len(keyCols))
		for j, col := range keyCols {
			v := table[0][col.Label]
			keyVals[j] = values.New(v)
		}
		groupKey := execute.NewGroupKey(keyCols, keyVals)
		c := constructChunk(groupKey, cols, table)
		chunks[i] = c
	}
	return chunks
}

func constructChunk(groupKey flux.GroupKey, cols []flux.ColMeta, rows []map[string]interface{}) table.Chunk {
	b := execute.NewChunkBuilder(cols, len(rows), memory.DefaultAllocator)
	for _, row := range rows {
		vrow := make(map[string]values.Value)
		for k, v := range row {
			val := values.New(v)
			vrow[k] = val
		}
		record := values.NewObjectWithValues(vrow)
		_ = b.AppendRecord(record)
	}
	return b.Build(groupKey)
}
