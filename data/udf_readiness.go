package data

import "strings"

// passiveUDFArgumentReady lets the native parser resolve nested variable names
// without executing getters. Unsupported expressions retain the entire legacy
// readiness traversal; never retry an already expanded UDF result.
func (s *Map) passiveUDFArgumentReady(argument string) bool {
	ready := true
	Parse(strings.TrimSpace(argument), func(expression string, udf bool, _ interface{}) (interface{}, bool) {
		if udf || !strings.HasPrefix(expression, "$") {
			ready = false
			return nil, false
		}
		path := expression[1:]
		if strings.HasPrefix(path, "{") && strings.HasSuffix(path, "}") {
			path = path[1 : len(path)-1]
		}
		var current interface{} = *s
		for _, key := range strings.Split(path, ".") {
			if !passiveIdentifier(key) {
				ready = false
				return nil, false
			}
			var found bool
			switch object := current.(type) {
			case Map:
				current, found = object[key]
			case map[string]interface{}:
				current, found = object[key]
			}
			if !found {
				ready = false
				return nil, false
			}
		}
		switch value := current.(type) {
		case string:
			if !passiveIdentifier(value) {
				ready = false
				return nil, false
			}
		case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		default:
			ready = false
			return nil, false
		}
		return current, true
	})
	return ready
}

func passiveIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for _, c := range value {
		if c != '_' && !(c >= 'a' && c <= 'z') && !(c >= 'A' && c <= 'Z') && !(c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}
