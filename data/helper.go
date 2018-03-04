package data

import (
	"unicode"
)

func ExtractPath(expression string) string {
	var result = ""
	for _, r := range expression {
		aChar := string(r)
		if unicode.IsLetter(r) || unicode.IsDigit(r) || aChar == "[" || aChar == "]" || aChar == "." || aChar == "_" {
			result += aChar
		}
	}
	return result
}
