// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zoneinfo

import (
	"sync"
)

func ZoneinfoForTesting() *string {
	return zoneinfo
}

func ResetZoneinfoForTesting() {
	zoneinfo = nil
	zoneinfoOnce = sync.Once{}
}

var (
	ForceZipFileForTesting = forceZipFileForTesting
	ErrLocation            = errLocation
	LoadTzinfo             = loadTzinfo
	Tzset                  = tzset
	TzsetName              = tzsetName
	TzsetOffset            = tzsetOffset
)

type RuleKind int

const (
	RuleJulian       = RuleKind(ruleJulian)
	RuleDOY          = RuleKind(ruleDOY)
	RuleMonthWeekDay = RuleKind(ruleMonthWeekDay)
)

type Rule struct {
	Kind RuleKind
	Day  int
	Week int
	Mon  int
	Time int
}

func TzsetRule(s string) (Rule, string, bool) {
	r, rs, ok := tzsetRule(s)
	rr := Rule{
		Kind: RuleKind(r.kind),
		Day:  r.day,
		Week: r.week,
		Mon:  r.mon,
		Time: r.time,
	}
	return rr, rs, ok
}
