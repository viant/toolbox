package data

import (
	"github.com/viant/toolbox"
	"strings"
)

// The native parser resolves passive dynamic index/name values before the next
// selector. Unsupported/getter expressions retain the legacy fallback exactly.
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
		if !passiveLookupPath(path) {
			ready = false
			return nil, false
		}
		current, found := s.LookupValue(path)
		if !found {
			ready = false
			return nil, false
		}
		switch value := current.(type) {
		case string:
			if !passiveIdentifier(value) {
				return nil, false
			}
		case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		default:
			if current == nil || toolbox.IsMap(current) || toolbox.IsSlice(current) {
				return nil, false
			}
			ready = false
			return nil, false
		}
		return current, true
	})
	return ready
}

func passiveLookupPath(path string) bool {
	for _, key := range strings.Split(path, ".") {
		if index := strings.IndexByte(key, '['); index >= 0 {
			if !strings.HasSuffix(key, "]") || !passiveIdentifier(key[index+1:len(key)-1]) {
				return false
			}
			key = key[:index]
		}
		if !passiveIdentifier(key) {
			return false
		}
	}
	return true
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
