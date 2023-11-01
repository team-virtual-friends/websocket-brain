package foundation

import (
	"regexp"
	"strings"
	"unicode"
)

func TrimPunctuation(s string) string {
	// Define a function to check if a character is punctuation
	isPunctuation := func(r rune) bool {
		return unicode.IsPunct(r)
	}

	// Trim punctuation characters from the left and right of the string
	trimmed := strings.TrimFunc(s, isPunctuation)

	return trimmed
}

func RemoveNewlines(input string) string {
	re := regexp.MustCompile(`[\r\n]+`)
	return re.ReplaceAllString(input, "")
}
