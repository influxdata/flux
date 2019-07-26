package strings

import "strings"

// replaceAll implements strings.ReplaceAll for go 1.11 and
// earlier. The function was added in go 1.12 and it is an
// alias to calling strings.Replace with an argument of -1.
//
// This function may be removed when InfluxDB 1.x no longer
// uses go 1.11 for its builds.
func replaceAll(s string, old string, new string) string {
	return strings.Replace(s, old, new, -1)
}
