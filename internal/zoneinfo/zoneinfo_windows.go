// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zoneinfo

import (
	"runtime"
)

var zoneSources = []string{
	runtime.GOROOT() + "/lib/time/zoneinfo.zip",
}
