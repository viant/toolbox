package data

import (
	"bytes"
	"github.com/viant/toolbox"
	"log"
	"strings"
	"time"
)

//Map types is an alias type to map[string]interface{} with extended behaviours
type Map map[string]interface{}

//Udf represents a user defined function used to transform data.
type Udf func(interface{}, Map) (interface{}, error)

//Put puts key value into the map.
func (s *Map) Put(key string, value interface{}) {
	(*s)[key] = value
}

//Delete removes the supplied keys, it supports key of path expression with dot i.e. request.method
func (s *Map) Delete(keys ...string) {
	for _, key := range keys {
		if !strings.Contains(key, ".") {
			delete(*s, key)
			continue
		}
		keyParts := strings.Split(key, ".")
		var temp = *s
		for i, part := range keyParts {
			if temp == nil {
				break
			}
			isLasPart := i+1 == len(keyParts)
			if isLasPart {
				delete(temp, part)
			} else if temp[part] != nil && toolbox.IsMap(temp[part]) {
				subMap := toolbox.AsMap(temp[part])
				temp = Map(subMap)
			} else {
				break
			}
		}
	}
}

//Replace replaces supplied key/path with corresponding value
func (s *Map) Replace(key, val string) {
	if !strings.Contains(key, ".") {
		(*s)[key] = val
		return
	}
	keyParts := strings.Split(key, ".")
	var temp = *s
	for i, part := range keyParts {
		if temp == nil {
			break
		}
		isLasPart := i+1 == len(keyParts)
		if isLasPart {
			temp[part] = val
		} else if temp[part] != nil && toolbox.IsMap(temp[part]) {
			subMap := toolbox.AsMap(temp[part])
			temp = Map(subMap)
		} else {
			break
		}
	}

}

//Has returns true if the provided key is present
func (s *Map) Has(key string) bool {
	_, found := (*s)[key]
	return found
}

//Get returns a value for provided key
func (s *Map) Get(key string) interface{} {
	if result, found := (*s)[key]; found {
		return result
	}
	return nil
}

/*
GetValue returns value for provided expression.
The expression uses dot (.) to denote nested data structure.
The following expression as supported
 1) <-key shift
 2) ++key pre increment
 3) key++ post increment
 4) $key reference access
*/
func (s *Map) GetValue(expr string) (interface{}, bool) {
	if expr == "" {
		return nil, false
	}
	if strings.HasPrefix(expr, "{") && strings.HasSuffix(expr, "}") {
		expr = expr[1 : len(expr)-1]
	}

	isShiftOperation := strings.HasPrefix(expr, "<-")
	if isShiftOperation {
		expr = string(expr[2:])
	}

	isPostIncrementOperation := strings.HasSuffix(expr, "++")
	if isPostIncrementOperation {
		expr = string(expr[:len(expr)-2])
	}

	isPreIncrementOperation := strings.HasPrefix(expr, "++")
	if isPreIncrementOperation {
		expr = string(expr[2:])
	}

	isReference := strings.HasPrefix(expr, "$")
	if isReference {
		expr = string(expr[1:])
		expr = s.GetString(expr)
		if expr == "" {
			return nil, false
		}
	}

	state := *s

	if strings.Contains(expr, ".") || strings.HasSuffix(expr, "]") {
		fragments := strings.Split(expr, ".")
		for i, fragment := range fragments {
			var index interface{}
			arrayIndexPosition := strings.Index(fragment, "[")
			if arrayIndexPosition != -1 {
				arrayEndPosition := strings.Index(fragment, "]")
				if arrayEndPosition > arrayIndexPosition && arrayEndPosition < len(fragment) {
					arrayIndex := string(fragment[arrayIndexPosition+1 : arrayEndPosition])
					index = arrayIndex
					fragment = string(fragment[:arrayIndexPosition])
				}
			}
			isLast := i+1 == len(fragments)

			hasKey := state.Has(fragment)
			if !hasKey {
				return nil, false
			}

			var candidate = state.Get(fragment)
			if !isLast && candidate == nil {
				return nil, false
			}

			if index != nil {

				if intIndex, err := toolbox.ToInt(index); err == nil {
					if !toolbox.IsSlice(candidate) {
						return nil, false
					}
					var aSlice = toolbox.AsSlice(candidate)
					if intIndex >= len(aSlice) {
						return nil, false
					}
					if intIndex < len(aSlice) {
						candidate = aSlice[intIndex]
					} else {
						candidate = nil
					}
				} else if textIndex, ok := index.(string); ok {
					if !toolbox.IsMap(candidate) {
						return nil, false
					}
					aMap := toolbox.AsMap(candidate)
					if candidate, ok = aMap[textIndex]; !ok {
						return nil, false
					}
				} else {
					return nil, false
				}

				if isLast {
					return candidate, true
				}
			}

			if isLast {
				expr = fragment
				continue
			}
			if toolbox.IsMap(candidate) {
				newState := toolbox.AsMap(candidate)
				if newState != nil {
					state = newState
				}
			} else {
				value, _ := state.GetValue(fragment)
				if f, ok := value.(func(key string) interface{}); ok {
					return f(fragments[i+1]), true
				}
				return nil, false
			}
		}
	}

	if state.Has(expr) {
		var result = state.Get(expr)
		if isPostIncrementOperation {
			state.Put(expr, toolbox.AsInt(result)+1)
		} else if isPreIncrementOperation {
			result = toolbox.AsInt(result) + 1
			state.Put(expr, result)
		} else if isShiftOperation {

			aCollection := state.GetCollection(expr)
			if len(*aCollection) == 0 {
				return nil, false
			}
			var result = (*aCollection)[0]
			var newCollection = (*aCollection)[1:]
			state.Put(expr, &newCollection)
			return result, true
		}
		if f, ok := result.(func() interface{}); ok {
			return f(), true
		}
		return result, true
	}
	return nil, false
}

/*
Set value sets value in the map for the supplied expression.
The expression uses dot (.) to denote nested data structure. For instance the following expression key1.subKey1.attr1
Would take or create map for key1, followied bu takeing or creating antother map for subKey1 to set attr1 key with provided value
The following expression as supported
 1) $key reference - the final key is determined from key's content.
 2) ->key push expression to append value at the end of the slice
*/
func (s *Map) SetValue(expr string, value interface{}) {
	if expr == "" {
		return
	}
	if value == nil {
		return
	}
	if strings.Index(expr, "$") != -1 {
		expr = s.ExpandAsText(expr)
	}

	state := *s
	isPushOperation := strings.HasPrefix(expr, "->")
	if isPushOperation {
		expr = string(expr[2:])
	}
	if strings.HasPrefix(expr, "{") && strings.HasSuffix(expr, "}") {
		expr = expr[1 : len(expr)-1]
	}

	if strings.Contains(expr, ".") {
		fragments := strings.Split(expr, ".")
		nodePath := strings.Join(fragments[:len(fragments)-1], ".")
		if node, ok := s.GetValue(nodePath); ok && toolbox.IsMap(node) {
			if _, writable := node.(map[string]interface{}); !writable {
				node = Map(toolbox.AsMap(node))
				s.SetValue(nodePath, node)
			}
			expr = fragments[len(fragments)-1]
			state = Map(toolbox.AsMap(node))
		} else {
			for i, fragment := range fragments {
				isLast := i+1 == len(fragments)
				if isLast {
					expr = fragment
				} else {
					subState := state.GetMap(fragment)
					if subState == nil {
						subState = NewMap()
						state.Put(fragment, subState)
					}
					state = subState
				}
			}
		}
	}

	if isPushOperation {
		collection := state.GetCollection(expr)
		if collection == nil {
			collection = NewCollection()
			state.Put(expr, collection)
		}
		collection.Push(value)
		state.Put(expr, collection)
		return
	}
	state.Put(expr, value)
}

//Apply copies all elements of provided map to this map.
func (s *Map) Apply(source map[string]interface{}) {
	for k, v := range source {
		(*s)[k] = v
	}
}

//GetString returns value for provided key as string.
func (s *Map) GetString(key string) string {
	if result, found := (*s)[key]; found {
		return toolbox.AsString(result)
	}
	return ""
}

//GetInt returns value for provided key as int.
func (s *Map) GetInt(key string) int {
	if result, found := (*s)[key]; found {
		return toolbox.AsInt(result)
	}
	return 0
}

//GetFloat returns value for provided key as float64.
func (s *Map) GetFloat(key string) float64 {
	if result, found := (*s)[key]; found {
		return toolbox.AsFloat(result)
	}
	return 0.0
}

//GetBoolean returns value for provided key as boolean.
func (s *Map) GetBoolean(key string) bool {
	if result, found := (*s)[key]; found {
		return toolbox.AsBoolean(result)
	}
	return false
}

//GetCollection returns value for provided key as collection pointer.
func (s *Map) GetCollection(key string) *Collection {
	if result, found := (*s)[key]; found {
		collectionPointer, ok := result.(*Collection)
		if ok {
			return collectionPointer
		}

		aSlice, ok := result.([]interface{})
		collection := Collection(aSlice)
		if ok {
			return &collection
		}
		if !toolbox.IsSlice(result) {
			return nil
		}
		aSlice = toolbox.AsSlice(result)
		collection = Collection(aSlice)
		return &collection
	}
	return nil
}

//GetMap returns value for provided key as  map.
func (s *Map) GetMap(key string) Map {
	if result, found := (*s)[key]; found {
		aMap, ok := result.(Map)
		if ok {
			return aMap
		}
		aMap, ok = result.(map[string]interface{})
		if ok {
			return aMap
		}
		var result = toolbox.AsMap(result)
		(*s)[key] = result
		return result
	}
	return nil
}

//Range iterates every key, value pair of this map, calling supplied callback as long it does return true.
func (s *Map) Range(callback func(k string, v interface{}) (bool, error)) error {
	for k, v := range *s {
		next, err := callback(k, v)
		if err != nil {
			return err
		}
		if !next {
			break
		}
	}
	return nil
}

//Clones create a clone of this map.
func (s *Map) Clone() Map {
	var result = NewMap()
	for key, value := range *s {
		if aMap, casted := value.(Map); casted {
			result[key] = aMap.Clone()
		} else {
			result[key] = value
		}
	}
	return result
}

//Returns a map that can be encoded by json or other decoders.
//Since a map can store a function, it filters out any non marshalable types.
func (s *Map) AsEncodableMap() map[string]interface{} {
	var result = make(map[string]interface{})
	for k, v := range *s {
		if v == nil {
			continue
		}
		result[k] = asEncodableValue(v)
	}
	return result
}

//asEncodableValue returns all non func values or func() literal for function.
func asEncodableValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	value := v
	if toolbox.IsFunc(v) {
		return "func()"
	}
	if toolbox.IsMap(v) {
		var aMap = Map(toolbox.AsMap(v))
		value = aMap.AsEncodableMap()
	} else if toolbox.IsSlice(v) {
		var targetSlice = make([]interface{}, 0)
		var sourceSlice = toolbox.AsSlice(v)
		for _, item := range sourceSlice {
			targetSlice = append(targetSlice, asEncodableValue(item))
		}
		value = targetSlice
	} else if toolbox.IsString(v) || toolbox.IsInt(v) || toolbox.IsFloat(v) {
		value = v
	} else {
		value = toolbox.AsString(v)
	}
	return value
}

func hasGenericKeys(aMap map[string]interface{}) bool {
	for k := range aMap {
		if strings.HasPrefix(k, "$AsInt") || strings.HasPrefix(k, "$AsFloat") || strings.HasPrefix(k, "$AsBool") {
			return true
		}
	}
	return false
}

//Expand expands provided value of any type with dollar sign expression/
func (s *Map) Expand(source interface{}) interface{} {
	switch value := source.(type) {
	case bool, []byte, int, uint, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64, time.Time:
		return source
	case *[]byte:
		return s.expandExpressions(string(*value))
	case *string:
		if value == nil {
			return ""
		}
		return s.expandExpressions(*value)
	case string:
		return s.expandExpressions(value)
	case map[string]interface{}:
		if hasGenericKeys(value) {
			result := make(map[interface{}]interface{})
			for k, v := range value {
				var key = s.Expand(k)
				result[key] = s.Expand(v)
			}
			return result
		}
		resultMap := make(map[string]interface{})
		for k, v := range value {
			var key = s.ExpandAsText(k)
			var expanded = s.Expand(v)

			if key == "..." {
				if expanded != nil && toolbox.IsMap(expanded) {
					for key, value := range toolbox.AsMap(expanded) {
						resultMap[key] = value
					}
					continue
				}
			}
			resultMap[key] = expanded
		}
		return resultMap
	case map[interface{}]interface{}:
		var resultMap = make(map[interface{}]interface{})
		for k, v := range value {
			var key = s.Expand(k)
			var expanded = s.Expand(v)
			if key == "..." {
				if expanded != nil && toolbox.IsMap(expanded) {
					for key, value := range toolbox.AsMap(expanded) {
						resultMap[key] = value
					}
					continue
				}
			}
			resultMap[key] = expanded
		}
		return resultMap
	case []interface{}:
		var resultSlice = make([]interface{}, len(value))
		for i, value := range value {
			resultSlice[i] = s.Expand(value)
		}
		return resultSlice
	default:

		if source == nil {
			return nil
		}

		if toolbox.IsMap(source) {
			switch aMap := value.(type) {
			case map[string]interface{}:
				return s.Expand(aMap)
			case map[interface{}]interface{}:
				return s.Expand(aMap)
			default:
				return s.Expand(toolbox.AsMap(value))
			}

		} else if toolbox.IsSlice(source) {
			return s.Expand(toolbox.AsSlice(value))
		} else if toolbox.IsStruct(value) {
			aMap := toolbox.AsMap(value)
			return s.Expand(aMap)
		} else if value != nil {
			return s.Expand(toolbox.AsString(value))
		}
	}
	return source
}

//ExpandAsText expands all matching expressions starting with dollar sign ($)
func (s *Map) ExpandAsText(text string) string {
	result := s.expandExpressions(text)
	if toolbox.IsSlice(result) || toolbox.IsMap(result) {
		buf := new(bytes.Buffer)
		err := toolbox.NewJSONEncoderFactory().Create(buf).Encode(result)
		if err == nil {
			return buf.String()
		}
	}
	if text, ok := result.(string); ok || result == nil {
		return text
	}
	return toolbox.AsString(result)
}

func (s *Map) evaluateUDF(candidate interface{}, argument interface{}) (interface{}, bool) {
	var canExpandAll = true

	if toolbox.IsString(argument) {
		var expandable = strings.TrimSpace(toolbox.AsString(argument))
		Parse(expandable, func(expression string, udf bool, argument interface{}) (interface{}, bool) {
			if _, has := s.GetValue(string(expression[1:])); !has {
				canExpandAll = false
			}
			return nil, false
		})
	}

	if !canExpandAll {
		return nil, false
	}
	udf, ok := candidate.(func(interface{}, Map) (interface{}, error))
	if !ok {
		return nil, false
	}

	expandedArgument := s.expandArgumentsExpressions(argument)
	if toolbox.IsString(expandedArgument) {
		expandedText := toolbox.AsString(expandedArgument)
		if toolbox.IsStructuredJSON(expandedText) {
			evaluated, err := toolbox.JSONToInterface(expandedText)
			if err != nil {
				return nil, false
			}
			expandedArgument = evaluated
		}
	}
	evaluated, err := udf(expandedArgument, *s)
	if err == nil {
		return evaluated, true
	}
	log.Printf("failed to evaluate %v, %v", candidate, err)
	return nil, false
}

func (s *Map) hasCycle(source interface{}, ownerVariable string) bool {
	switch value := source.(type) {
	case string:
		return strings.Contains(value, ownerVariable)
	case Map:
		for k, v := range value {
			if s.hasCycle(k, ownerVariable) || s.hasCycle(v, ownerVariable) {
				return true
			}
		}

	case map[string]interface{}:
		for k, v := range value {
			if s.hasCycle(k, ownerVariable) || s.hasCycle(v, ownerVariable) {
				return true
			}
		}

	case []interface{}:
		for _, v := range value {
			if s.hasCycle(v, ownerVariable) {
				return true
			}
		}
	case Collection:
		for _, v := range value {
			if s.hasCycle(v, ownerVariable) {
				return true
			}
		}
	}
	return false
}

//expandExpressions will check provided text with any expression starting with dollar sign ($) to substitute it with key in the map if it is present.
//The result can be an expanded text or type of key referenced by the expression.
func (s *Map) expandExpressions(text string) interface{} {
	if strings.Index(text, "$") == -1 {
		return text
	}
	var expandVariable = func(expression string, isUDF bool, argument interface{}) (interface{}, bool) {
		value, hasExpValue := s.GetValue(string(expression[1:]))
		if hasExpValue {
			if value != expression && s.hasCycle(value, expression) {
				log.Printf("detected data cycle on %v in value: %v", expression, value)
				return expression, true
			}
			if isUDF {
				if evaluated, ok := s.evaluateUDF(value, argument); ok {
					return evaluated, true
				}
			} else {
				if value != nil && (toolbox.IsMap(value) || toolbox.IsSlice(value)) {
					return s.Expand(value), true
				}
				if text, ok := value.(string); ok {
					return text, true
				}
				if value != nil {
					return toolbox.DereferenceValue(value), true
				}
				return value, true
			}
		}

		if isUDF {
			expandedArgument := s.expandArgumentsExpressions(argument)
			_, isByteArray := expandedArgument.([]byte)
			if !toolbox.IsMap(expandedArgument) && !toolbox.IsSlice(expandedArgument) || isByteArray {
				argument = toolbox.AsString(expandedArgument)
			}
			return expression + "(" + toolbox.AsString(argument) + ")", true
		}
		return expression, true
	}

	return Parse(text, expandVariable)
}

//expandExpressions will check provided text with any expression starting with dollar sign ($) to substitute it with key in the map if it is present.
//The result can be an expanded text or type of key referenced by the expression.
func (s *Map) expandArgumentsExpressions(argument interface{}) interface{} {
	if argument == nil || !toolbox.IsString(argument) {
		return argument
	}

	argumentLiteral, ok := argument.(string)

	if ok {
		if toolbox.IsStructuredJSON(argumentLiteral) {
			return s.expandExpressions(argumentLiteral)
		}
	}

	var expandVariable = func(expression string, isUDF bool, argument interface{}) (interface{}, bool) {
		value, hasExpValue := s.GetValue(string(expression[1:]))
		if hasExpValue {
			if value != expression && s.hasCycle(value, expression) {
				log.Printf("detected data cycle on %v in value: %v", expression, value)
				return expression, true
			}
			if isUDF {
				if evaluated, ok := s.evaluateUDF(value, argument); ok {
					return evaluated, true
				}
			} else {
				if value != nil && (toolbox.IsMap(value) || toolbox.IsSlice(value)) {
					return s.Expand(value), true
				}
				if text, ok := value.(string); ok {
					return text, true
				}
				if value != nil {
					return toolbox.DereferenceValue(value), true
				}
				return value, true
			}
		}
		if isUDF {
			expandedArgument := s.expandArgumentsExpressions(argument)
			_, isByteArray := expandedArgument.([]byte)
			if !toolbox.IsMap(expandedArgument) && !toolbox.IsSlice(expandedArgument) || isByteArray {
				argument = toolbox.AsString(expandedArgument)
			}
			expression = expression + "(" + toolbox.AsString(argument) + ")"
			return expression, true
		}
		return expression, true
	}

	tokenizer := toolbox.NewTokenizer(argumentLiteral, invalidToken, eofToken, matchers)
	var result = make([]interface{}, 0)
	for tokenizer.Index < len(argumentLiteral) {
		match, err := toolbox.ExpectTokenOptionallyFollowedBy(tokenizer, whitespace, "expected argument", doubleQuoteEnclosedToken, comaToken, unmatchedToken, eofToken)
		if err != nil {
			return Parse(argumentLiteral, expandVariable)
		}
		switch match.Token {
		case doubleQuoteEnclosedToken:
			result = append(result, strings.Trim(match.Matched, `"`))
		case comaToken:
			result = append(result, match.Matched)
			tokenizer.Index++
		case unmatchedToken:
			result = append(result, match.Matched)
		}
	}

	for i, arg := range result {
		textArg, ok := arg.(string)
		if !ok {
			continue
		}
		textArg = strings.Trim(textArg, "'")
		result[i] = Parse(textArg, expandVariable)

	}
	if len(result) == 1 {
		return result[0]
	}
	return result
}

//NewMap creates a new instance of a map.
func NewMap() Map {
	return make(map[string]interface{})
}
