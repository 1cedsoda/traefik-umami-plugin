package traefik_umami_plugin

import (
	"regexp"
)

const insertBeforeRegexPattern = `</body>`

var insertBeforeRegex = regexp.MustCompile(insertBeforeRegexPattern)

func RegexReplaceSingle(bytes []byte, match *regexp.Regexp, replace string) []byte {
	rx := match.FindIndex(bytes)
	if len(rx) == 0 {
		return bytes
	}
	// insert the script before the head tag
	newBytes := append(bytes[:rx[0]], append([]byte(replace), bytes[rx[0]:]...)...)
	return newBytes
}

func InsertAtBodyEnd(bytes []byte, replace string) []byte {
	return RegexReplaceSingle(bytes, insertBeforeRegex, replace)
}
