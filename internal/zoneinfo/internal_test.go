// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zoneinfo

var OrigZoneSources = zoneSources

func forceZipFileForTesting(zipOnly bool) {
	zoneSources = make([]string, len(OrigZoneSources))
	copy(zoneSources, OrigZoneSources)
	if zipOnly {
		zoneSources = zoneSources[len(zoneSources)-1:]
	}
}

func (l *Location) Lookup(sec int64) (name string, offset int, start, end int64, isDST bool) {
	return l.lookup(sec)
}
