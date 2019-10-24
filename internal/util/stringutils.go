package util

import "strings"

// TrimString trims string for: ' ', '"', '\n'
func TrimString(val string) string {
	val = strings.TrimSpace(val)
	val = strings.TrimPrefix(val, "\"")
	val = strings.TrimSuffix(val, "\n")
	val = strings.TrimSuffix(val, "\"")
	return val
}
