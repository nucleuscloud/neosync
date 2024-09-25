package gotypeutil

import "strings"

func CaseInsensitiveContains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
