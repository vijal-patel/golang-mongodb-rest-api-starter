package strutils

import (
	"regexp"
)

// Prevent phishing technique whereby urls are injected into non-url fields like user name etc.
func StringContainsUrl(input string) bool {
	regexPattern := `[(http(s)?)://(www\.)?a-zA-Z0-9@:%._+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_+.~#?&//=]*)`
	urlPattern := regexp.MustCompile(regexPattern) // TODO global init
	matches := urlPattern.FindAllString(input, -1)
	if len(matches) > 0 {
		return true
	}
	return false
}
