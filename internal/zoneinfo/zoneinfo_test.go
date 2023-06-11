// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zoneinfo_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/InfluxCommunity/flux/internal/zoneinfo"
)

func init() {
	if zoneinfo.ZoneinfoForTesting() != nil {
		panic(fmt.Errorf("zoneinfo initialized before first LoadLocation"))
	}
}

func TestEnvVarUsage(t *testing.T) {
	zoneinfo.ResetZoneinfoForTesting()

	const testZoneinfo = "foo.zip"
	const env = "ZONEINFO"

	t.Setenv(env, testZoneinfo)

	// Result isn't important, we're testing the side effect of this command
	zoneinfo.LoadLocation("Asia/Jerusalem")
	defer zoneinfo.ResetZoneinfoForTesting()

	if zoneinfo := zoneinfo.ZoneinfoForTesting(); testZoneinfo != *zoneinfo {
		t.Errorf("zoneinfo does not match env variable: got %q want %q", *zoneinfo, testZoneinfo)
	}
}

func TestBadLocationErrMsg(t *testing.T) {
	zoneinfo.ResetZoneinfoForTesting()
	loc := "Asia/SomethingNotExist"
	want := errors.New("unknown time zone " + loc)
	_, err := zoneinfo.LoadLocation(loc)
	if err.Error() != want.Error() {
		t.Errorf("LoadLocation(%q) error = %v; want %v", loc, err, want)
	}
}

func TestLoadLocationValidatesNames(t *testing.T) {
	zoneinfo.ResetZoneinfoForTesting()
	const env = "ZONEINFO"
	t.Setenv(env, "")

	bad := []string{
		"/usr/foo/Foo",
		"\\UNC\foo",
		"..",
		"a..",
	}
	for _, v := range bad {
		_, err := zoneinfo.LoadLocation(v)
		if err != zoneinfo.ErrLocation {
			t.Errorf("LoadLocation(%q) error = %v; want ErrLocation", v, err)
		}
	}
}

func TestVersion3(t *testing.T) {
	zoneinfo.ForceZipFileForTesting(true)
	defer zoneinfo.ForceZipFileForTesting(false)
	_, err := zoneinfo.LoadLocation("Asia/Jerusalem")
	if err != nil {
		t.Fatal(err)
	}
}

// Test that we get the correct results for times before the first
// transition time. To do this we explicitly check early dates in a
// couple of specific timezones.
func TestFirstZone(t *testing.T) {
	zoneinfo.ForceZipFileForTesting(true)
	defer zoneinfo.ForceZipFileForTesting(false)

	var tests = []struct {
		zone  string
		unix  int64
		want1 string
		want2 string
	}{
		{
			"PST8PDT",
			-1633269601,
			"-0800 (PST)",
			"-0700 (PDT)",
		},
		{
			"Pacific/Fakaofo",
			1325242799,
			"-1100 (-11)",
			"+1300 (+13)",
		},
	}

	for _, test := range tests {
		z, err := zoneinfo.LoadLocation(test.zone)
		if err != nil {
			t.Fatal(err)
		}
		name, offset, _, _, _ := z.Lookup(test.unix)
		if got := fmt.Sprintf("%+03d%02d (%s)", offset/3600, offset%3600, name); got != test.want1 {
			t.Errorf("for %s %d got %q want %q", test.zone, test.unix, got, test.want1)
		}
		name, offset, _, _, _ = z.Lookup(test.unix + 1)
		if got := fmt.Sprintf("%+03d%02d (%s)", offset/3600, offset%3600, name); got != test.want2 {
			t.Errorf("for %s %d got %q want %q", test.zone, test.unix, got, test.want2)
		}
	}
}

func TestLocationNames(t *testing.T) {
	if zoneinfo.UTC.String() != "UTC" {
		t.Errorf(`invalid UTC location name: got %q want "UTC"`, zoneinfo.UTC)
	}
}

func TestLoadLocationFromTZData(t *testing.T) {
	zoneinfo.ForceZipFileForTesting(true)
	defer zoneinfo.ForceZipFileForTesting(false)

	const locationName = "Asia/Jerusalem"
	reference, err := zoneinfo.LoadLocation(locationName)
	if err != nil {
		t.Fatal(err)
	}

	tzinfo, err := zoneinfo.LoadTzinfo(locationName, zoneinfo.OrigZoneSources[len(zoneinfo.OrigZoneSources)-1])
	if err != nil {
		t.Fatal(err)
	}
	sample, err := zoneinfo.LoadLocationFromTZData(locationName, tzinfo)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(reference, sample) {
		t.Errorf("return values of LoadLocationFromTZData and LoadLocation don't match")
	}
}

func TestMalformedTZData(t *testing.T) {
	// The goal here is just that malformed tzdata results in an error, not a panic.
	issue29437 := "TZif\x00000000000000000\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x0000"
	_, err := zoneinfo.LoadLocationFromTZData("abc", []byte(issue29437))
	if err == nil {
		t.Error("expected error, got none")
	}
}

func TestTzset(t *testing.T) {
	for _, test := range []struct {
		inStr   string
		inStart int64
		inEnd   int64
		inSec   int64
		name    string
		off     int
		start   int64
		end     int64
		isDST   bool
		ok      bool
	}{
		{"", 0, 0, 0, "", 0, 0, 0, false, false},
		{"PST8PDT,M3.2.0,M11.1.0", 0, 0, 2159200800, "PDT", -7 * 60 * 60, 2152173600, 2172733200, true, true},
		{"PST8PDT,M3.2.0,M11.1.0", 0, 0, 2152173599, "PST", -8 * 60 * 60, 2145916800, 2152173600, false, true},
		{"PST8PDT,M3.2.0,M11.1.0", 0, 0, 2152173600, "PDT", -7 * 60 * 60, 2152173600, 2172733200, true, true},
		{"PST8PDT,M3.2.0,M11.1.0", 0, 0, 2152173601, "PDT", -7 * 60 * 60, 2152173600, 2172733200, true, true},
		{"PST8PDT,M3.2.0,M11.1.0", 0, 0, 2172733199, "PDT", -7 * 60 * 60, 2152173600, 2172733200, true, true},
		{"PST8PDT,M3.2.0,M11.1.0", 0, 0, 2172733200, "PST", -8 * 60 * 60, 2172733200, 2177452800, false, true},
		{"PST8PDT,M3.2.0,M11.1.0", 0, 0, 2172733201, "PST", -8 * 60 * 60, 2172733200, 2177452800, false, true},
	} {
		name, off, start, end, isDST, ok := zoneinfo.Tzset(test.inStr, test.inStart, test.inEnd, test.inSec)
		if name != test.name || off != test.off || start != test.start || end != test.end || isDST != test.isDST || ok != test.ok {
			t.Errorf("tzset(%q, %d, %d) = %q, %d, %d, %d, %t, %t, want %q, %d, %d, %d, %t, %t", test.inStr, test.inEnd, test.inSec, name, off, start, end, isDST, ok, test.name, test.off, test.start, test.end, test.isDST, test.ok)
		}
	}
}

func TestTzsetName(t *testing.T) {
	for _, test := range []struct {
		in   string
		name string
		out  string
		ok   bool
	}{
		{"", "", "", false},
		{"X", "", "", false},
		{"PST", "PST", "", true},
		{"PST8PDT", "PST", "8PDT", true},
		{"PST-08", "PST", "-08", true},
		{"<A+B>+08", "A+B", "+08", true},
	} {
		name, out, ok := zoneinfo.TzsetName(test.in)
		if name != test.name || out != test.out || ok != test.ok {
			t.Errorf("tzsetName(%q) = %q, %q, %t, want %q, %q, %t", test.in, name, out, ok, test.name, test.out, test.ok)
		}
	}
}

func TestTzsetOffset(t *testing.T) {
	for _, test := range []struct {
		in  string
		off int
		out string
		ok  bool
	}{
		{"", 0, "", false},
		{"X", 0, "", false},
		{"+", 0, "", false},
		{"+08", 8 * 60 * 60, "", true},
		{"-01:02:03", -1*60*60 - 2*60 - 3, "", true},
		{"01", 1 * 60 * 60, "", true},
		{"100", 100 * 60 * 60, "", true},
		{"1000", 0, "", false},
		{"8PDT", 8 * 60 * 60, "PDT", true},
	} {
		off, out, ok := zoneinfo.TzsetOffset(test.in)
		if off != test.off || out != test.out || ok != test.ok {
			t.Errorf("tzsetName(%q) = %d, %q, %t, want %d, %q, %t", test.in, off, out, ok, test.off, test.out, test.ok)
		}
	}
}

func TestTzsetRule(t *testing.T) {
	for _, test := range []struct {
		in  string
		r   zoneinfo.Rule
		out string
		ok  bool
	}{
		{"", zoneinfo.Rule{}, "", false},
		{"X", zoneinfo.Rule{}, "", false},
		{"J10", zoneinfo.Rule{Kind: zoneinfo.RuleJulian, Day: 10, Time: 2 * 60 * 60}, "", true},
		{"20", zoneinfo.Rule{Kind: zoneinfo.RuleDOY, Day: 20, Time: 2 * 60 * 60}, "", true},
		{"M1.2.3", zoneinfo.Rule{Kind: zoneinfo.RuleMonthWeekDay, Mon: 1, Week: 2, Day: 3, Time: 2 * 60 * 60}, "", true},
		{"30/03:00:00", zoneinfo.Rule{Kind: zoneinfo.RuleDOY, Day: 30, Time: 3 * 60 * 60}, "", true},
		{"M4.5.6/03:00:00", zoneinfo.Rule{Kind: zoneinfo.RuleMonthWeekDay, Mon: 4, Week: 5, Day: 6, Time: 3 * 60 * 60}, "", true},
		{"M4.5.7/03:00:00", zoneinfo.Rule{}, "", false},
		{"M4.5.6/-04", zoneinfo.Rule{Kind: zoneinfo.RuleMonthWeekDay, Mon: 4, Week: 5, Day: 6, Time: -4 * 60 * 60}, "", true},
	} {
		r, out, ok := zoneinfo.TzsetRule(test.in)
		if r != test.r || out != test.out || ok != test.ok {
			t.Errorf("tzsetName(%q) = %#v, %q, %t, want %#v, %q, %t", test.in, r, out, ok, test.r, test.out, test.ok)
		}
	}
}
