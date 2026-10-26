package udf

import (
	"fmt"
	"github.com/viant/toolbox/data"
	"reflect"
	"strconv"
	"strings"
)

func selectNilValue(value interface{}) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface, reflect.Func, reflect.Chan:
		return reflected.IsNil()
	}
	return false
}

// Locate the final outer bracket without treating brackets inside quoted values
// as a selector. Numeric indexes before the predicate stay in the passive path.
func selectFinalBracket(path string) (int, error) {
	start, depth := -1, 0
	var quote rune
	escaped := false
	for i, char := range path {
		if escaped {
			escaped = false
			continue
		}
		if quote != 0 {
			if char == '\\' {
				escaped = true
			} else if char == quote {
				quote = 0
			}
			continue
		}
		if char == '\'' || char == '"' {
			quote = char
			continue
		}
		switch char {
		case '[':
			if depth == 0 {
				start = i
			}
			depth++
		case ']':
			depth--
			if depth < 0 {
				return -1, fmt.Errorf("Select selector has malformed brackets")
			}
			if depth == 0 && strings.TrimSpace(path[start+1:i]) == "" {
				return -1, fmt.Errorf("Select selector has malformed brackets")
			}
		}
	}
	if depth != 0 || quote != 0 {
		return -1, fmt.Errorf("Select selector has malformed brackets")
	}
	return start, nil
}

func selectLookupValue(node data.Map, path string) (interface{}, bool, error) {
	// Preserve literal map keys before interpreting an authored path.
	if value, exists := node[path]; exists {
		return value, true, nil
	}
	if _, err := selectFinalBracket(path); err != nil {
		return nil, false, err
	}
	value, found := node.LookupValue(path)
	return value, found, nil
}

func selectPredicateBracket(path string, index int) bool {
	if index < 0 || !strings.HasSuffix(path, "]") {
		return false
	}
	_, err := strconv.Atoi(strings.TrimSpace(path[index+1 : len(path)-1]))
	return err != nil
}

func selectTypeError(message string, value interface{}) error {
	name := fmt.Sprintf("%T", value)
	if len(name) > 64 {
		name = name[:64] + "..."
	}
	return fmt.Errorf("%s (%s)", message, name)
}
