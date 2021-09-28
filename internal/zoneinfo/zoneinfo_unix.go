// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || (darwin && !ios) || dragonfly || freebsd || (linux && !android) || netbsd || openbsd || solaris
// +build aix darwin,!ios dragonfly freebsd linux,!android netbsd openbsd solaris

// Parse "zoneinfo" time zone file.
// This is a fairly standard file format used on OS X, Linux, BSD, Sun, and others.
// See tzfile(5), https://en.wikipedia.org/wiki/Zoneinfo,
// and ftp://munnari.oz.au/pub/oldtz/

package zoneinfo

import (
	"runtime"
)

// Many systems use /usr/share/zoneinfo, Solaris 2 has
// /usr/share/lib/zoneinfo, IRIX 6 has /usr/lib/locale/TZ.
var zoneSources = []string{
	"/usr/share/zoneinfo/",
	"/usr/share/lib/zoneinfo/",
	"/usr/lib/locale/TZ/",
	runtime.GOROOT() + "/lib/time/zoneinfo.zip",
}
