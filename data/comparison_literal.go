package data

import (
	"strconv"
	"strings"
)

// Quoted comparison operands are literal strings, not variable/UDF expressions.
// UnquoteChar handles both quote forms without introducing another expression parser.
func comparisonQuotedLiteral(expression string) (string, bool, bool) {
	token := strings.TrimSpace(expression)
	if len(token) == 0 || token[0] != '\'' && token[0] != '"' {
		return "", false, true
	}
	quote := token[0]
	if len(token) < 2 || token[len(token)-1] != quote {
		return "", true, false
	}
	remaining := token[1 : len(token)-1]
	var result strings.Builder
	for len(remaining) > 0 {
		value, _, tail, err := strconv.UnquoteChar(remaining, quote)
		if err != nil {
			return "", true, false
		}
		result.WriteRune(value)
		remaining = tail
	}
	return result.String(), true, true
}

func comparisonStringOperand(expression string, handler func(string, bool, interface{}) (interface{}, bool)) interface{} {
	resolved := false
	value := Parse("$"+expression, func(name string, isUDF bool, args interface{}) (interface{}, bool) {
		value, found := handler(name, isUDF, args)
		if found {
			resolved = true
		}
		return value, found
	})
	// Strip only an unresolved parser marker, never a dollar prefix from data.
	if text, ok := value.(string); ok && !resolved && strings.HasPrefix(text, "$") {
		return strings.TrimSpace(text[1:])
	}
	return value
}
