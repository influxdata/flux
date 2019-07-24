package runtime

import "runtime/debug"

var buildInfo *debug.BuildInfo

func SetBuildInfo(bi *debug.BuildInfo) {
	buildInfo = bi
}

func init() {
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		ok := buildInfo != nil
		return buildInfo, ok
	}
}
