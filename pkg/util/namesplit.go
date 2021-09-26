package util

import "strings"

func Namesplit(name string) (string, string) {
	if name == "" {
		return "", ""
	}
	split := strings.Fields(name)
	givenName := split[0]
	familyName := strings.Join(split[1:], " ")

	return givenName, familyName
}
