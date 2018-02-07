package data

import (
  "bytes"
  "fmt"
  "github.com/viant/toolbox"
  "log"
  "strings"
  "time"
  "unicode"
)

const (
  disableUDFKey       = "__$__disableUDF"
  expectVariableStart = iota
  expectVariableName
  expectElementPosition
  expectVariableNameEnclosureEnd
)

//Map types is an alias type to map[string]interface{} with extended behaviours
type Map map[string]interface{}

//Udf represents a user defined function used to transform data.
type Udf func(interface{}, Map) (interface{}, error)

//Put puts key value into the map.
func (s *Map) Put(key string, value interface{}) {
  (*s)[key] = value
}

//Delete removes the supplied keys
func (s *Map) Delete(keys ...string) {
  for _, key := range keys {
    delete(*s, key)
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
  if string(expr[0:1]) == "{" {
    expr = expr[1 : len(expr)-1]
  }

  if strings.Contains(expr, ".") || strings.HasSuffix(expr, "]") {
    fragments := strings.Split(expr, ".")
    for i, fragment := range fragments {
      var index *int
      arrayIndexPosition := strings.Index(fragment, "[")
      if arrayIndexPosition != -1 {
        arrayEndPosition := strings.Index(fragment, "]")
        arrayIndex := toolbox.AsInt(string(fragment[arrayIndexPosition+1 : arrayEndPosition]))
        index = &arrayIndex
        fragment = string(fragment[:arrayIndexPosition])
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

        if !toolbox.IsSlice(candidate) {
          return nil, false
        }
        var aSlice = toolbox.AsSlice(candidate)
        if *index >= len(aSlice) {
          return nil, false
        }
        if (*index) < len(aSlice) {
          candidate = aSlice[*index]
        } else {
          candidate = nil
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
  state := *s
  isReference := strings.HasPrefix(expr, "$")
  if isReference {
    expr = string(expr[1:])
    expr = s.GetString(expr)
    s.Put(expr, value)
  }

  isPushOperation := strings.HasPrefix(expr, "->")
  if isPushOperation {
    expr = string(expr[2:])
  }
  if string(expr[0:1]) == "{" {
    expr = expr[1 : len(expr)-1]
  }
  if strings.Contains(expr, ".") {
    fragments := strings.Split(expr, ".")
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
    collectionPoiner, ok := result.(*Collection)
    if ok {
      return collectionPoiner
    }
    aSlice, ok := result.([]interface{})
    collection := Collection(aSlice)
    if ok {
      return &collection
    }
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
    result[k] = encodableValue(v)
  }
  return result
}

func encodableValue(v interface{}) interface{} {
  if v == nil {
    return nil
  }
  var value interface{} = v

  if toolbox.IsFunc(v) {
    return "func()"
  }
  if toolbox.IsMap(v) {
    var aMap = Map(toolbox.AsMap(v))
    value = aMap.AsEncodableMap()
  } else if toolbox.IsSlice(v) {
    var aSlice = make([]interface{}, 0)
    for _, item := range toolbox.AsSlice(aSlice) {
      aSlice = append(aSlice, encodableValue(item))
    }
    value = aSlice
  } else if toolbox.IsString(v) || toolbox.IsInt(v) || toolbox.IsFloat(v) {
    value = v
  } else {
    value = toolbox.AsString(v)
  }
  return value
}

//Expand expands provided value of any type with dollar sign expression/
func (s *Map) Expand(source interface{}) interface{} {
  switch value := source.(type) {
  case bool, []byte, int, uint, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64, time.Time:
    return source
  case string:
    var has bool
    udf, value, suffix := s.getUdfIfDefined(value)
    var sourceValue interface{} = value
    if strings.HasPrefix(value, "$") {
      sourceValue, has = s.GetValue(string(value[1:]))
      if !has && udf != nil { //variable is not present in the context, thus delay udf execution
        return source
      }
      //you do not want to double evaluate case if there is UDF, just get value for it
      if !has || udf == nil {
        sourceValue = s.expandExpressions(value)
        has = true
      }
    }

    if udf != nil {
      transformed, err := udf(sourceValue, *s)
      if err != nil {
        log.Printf("failed to run udf: %v, %v\n", value, err)
        return source
      }
      if suffix != "" {
        transformed = toolbox.AsString(transformed) + suffix
      }
      return transformed
    } else if has {
      return sourceValue
    }
    return s.ExpandAsText(value)

  case map[string]interface{}:
    var resultMap = make(map[string]interface{})
    for k, v := range value {
      var expanded = s.Expand(v)
      resultMap[s.ExpandAsText(k)] = expanded
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
      return s.Expand(toolbox.AsMap(value))
    } else if toolbox.IsSlice(source) {
      return s.Expand(toolbox.AsSlice(value))
    } else {
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
  return toolbox.AsString(result)
}

//expandExpressions will check provided text with any expression starting with dollar sign ($) to substitute it with key in the map if it is present.
//The result can be an expanded text or type of key referenced by the expression.
func (s *Map) expandExpressions(text string) interface{} {
  if strings.Index(text, "$") == -1 {
    return text
  }
  var expandVariable = func(variableName string) interface{} {
    value, has := s.GetValue(string(variableName[1:]))
    if has {
      if value != nil && (toolbox.IsMap(value) || toolbox.IsSlice(value)) {
        return s.Expand(value)
      }
      return value
    }
    return variableName
  }
  var variableName = ""
  var parsingState = expectVariableStart
  var result = ""
  var expectIndexEnd = false
  for i, rune := range text {
    aChar := string(text[i : i+1])
    var isLast = i+1 == len(text)
    switch parsingState {
    case expectVariableStart:
      if aChar == "$" {
        variableName += aChar
        if i+1 < len(text) {
          nextChar := string(text[i+1 : i+2])
          if nextChar == "{" {
            parsingState = expectVariableNameEnclosureEnd
            continue

          }
        }
        parsingState = expectVariableName
        continue
      }
      result += aChar

    case expectVariableNameEnclosureEnd:
      variableName += aChar
      if aChar != "}" {
        continue
      }
      var expanded = expandVariable(variableName)
      if isLast && result == "" {
        return expanded
      }
      result += asExpandedText(expanded)
      variableName = ""
      parsingState = expectVariableStart

    case expectVariableName:

      if unicode.IsLetter(rune) || unicode.IsDigit(rune) || aChar == "[" || (expectIndexEnd && aChar == "]") || aChar == "." || aChar == "_" || aChar == "+" || aChar == "<" || aChar == "-" {
        if aChar == "[" {
          expectIndexEnd = true
        } else if aChar == "]" {
          expectIndexEnd = false
        }
        variableName += aChar
        continue
      }
      var expanded = expandVariable(variableName)
      if isLast && result == "" {
        return expanded
      }
      result += asExpandedText(expanded)
      result += aChar
      variableName = ""
      parsingState = expectVariableStart

    }
  }
  if len(variableName) > 0 {
    var expanded = expandVariable(variableName)
    if result == "" {
      return expanded
    }
    result += asExpandedText(expanded)
  }
  return result
}

func (s *Map) getUdfIfDefined(expression string) (func(interface{}, Map) (interface{}, error), string, string) {
  if !strings.HasPrefix(expression, "!") {
    return nil, expression, ""
  }

  startArgumentPosition := strings.Index(expression, "(")
  endArgumentPosition := strings.LastIndex(expression, ")")
  if startArgumentPosition != -1 && endArgumentPosition > startArgumentPosition {
    udfName := string(expression[1:startArgumentPosition])
    var has bool
    udfCandidate, has := s.GetValue(udfName)
    if !has {
      return nil, expression, ""
    }
    var udf, ok = udfCandidate.(func(interface{}, Map) (interface{}, error))
    if !ok {
      var errorMessage = fmt.Sprintf("Invalid UDF signature expacted %T but had %v", udf, udfCandidate)
      panic(errorMessage)
    }

    value := string(expression[startArgumentPosition+1 : endArgumentPosition])
    remaining := ""
    if !strings.HasSuffix(expression, ")") {
      remaining = expression[endArgumentPosition+1:]
    }
    return udf, value, remaining
  }
  return nil, expression, ""
}

func asExpandedText(source interface{}) string {
  if toolbox.IsSlice(source) || toolbox.IsMap(source) {
    buf := new(bytes.Buffer)
    err := toolbox.NewJSONEncoderFactory().Create(buf).Encode(source)
    if err == nil {
      return buf.String()
    }
  }
  return toolbox.AsString(source)
}

//NewMap creates a new instance of a map.
func NewMap() Map {
  return make(map[string]interface{})
}
