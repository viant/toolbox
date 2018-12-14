package data

import (
	"strings"
	"unicode"
)

func ExtractPath(expression string) string {
	var result = ""
	for _, r := range expression {
		aChar := string(r)
		if unicode.IsLetter(r) || unicode.IsDigit(r) || aChar == "[" || aChar == "]" || aChar == "." || aChar == "_" || aChar == "{" || aChar == "}" {
			result += aChar
		}
	}
	return strings.Trim(result, "{}")
}
