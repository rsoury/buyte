package util

import "strings"

func Rjust(str string, n int, fill string) string {
	if n > len(str) {
		places := n - len(str)
		return strings.Repeat(fill, places) + str
	}
	return str
}
