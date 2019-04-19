package toolbox

import (
	"fmt"
	"reflect"
	"strings"
)

//GetFunction returns function for provided owner and name, or error
func GetFunction(owner interface{}, name string) (interface{}, error) {
	if owner == nil {
		return nil, fmt.Errorf("failed to lookup %v on %T, owner was nil", name, owner)
	}
	var ownerType = reflect.TypeOf(owner)
	var method, has = ownerType.MethodByName(name)
	if !has {
		var available = make([]string, 0)
		for i := 0; i < ownerType.NumMethod(); i++ {
			available = append(available, ownerType.Method(i).Name)
		}
		return nil, fmt.Errorf("failed to lookup %T.%v, available:[%v]", owner, name, strings.Join(available, ","))
	}
	return reflect.ValueOf(owner).MethodByName(method.Name).Interface(), nil
}

//CallFunction calls passed in function with provided parameters,it returns a function result.
func CallFunction(function interface{}, parameters ...interface{}) []interface{} {
	AssertKind(function, reflect.Func, "function")
	var functionParameters = make([]reflect.Value, 0)
	ProcessSlice(parameters, func(item interface{}) bool {
		functionParameters = append(functionParameters, reflect.ValueOf(item))
		return true
	})

	functionValue := reflect.ValueOf(function)
	var resultValues = functionValue.Call(functionParameters)
	var result = make([]interface{}, len(resultValues))
	for i, resultValue := range resultValues {
		result[i] = resultValue.Interface()
	}
	return result
}

//AsCompatibleFunctionParameters takes incompatible function parameters and converts then into provided function signature compatible
func AsCompatibleFunctionParameters(function interface{}, parameters []interface{}) ([]interface{}, error) {
	return AsFunctionParameters(function, parameters, map[string]interface{}{})
}

//AsFunctionParameters takes incompatible function parameters and converts then into provided function signature compatible
func AsFunctionParameters(function interface{}, parameters []interface{}, parametersKV map[string]interface{}) ([]interface{}, error) {
	AssertKind(function, reflect.Func, "function")
	functionValue := reflect.ValueOf(function)
	funcSignature := GetFuncSignature(function)
	actualMethodSignatureLength := len(funcSignature)
	converter := Converter{}
	if actualMethodSignatureLength != len(parameters) {
		return nil, fmt.Errorf("invalid number of parameters wanted: [%T],  had: %v", function, len(parameters))
	}
	var functionParameters = make([]interface{}, 0)
	for i, parameterValue := range parameters {
		isStruct := IsStruct(funcSignature[i])
		if isStruct && parameterValue == nil {
			parameterValue = make(map[string]interface{})
		}
		reflectValue := reflect.ValueOf(parameterValue)
		if !isStruct {
			if parameterValue == nil {
				return nil, fmt.Errorf("parameter[%v] was empty", i)
			}
			if reflectValue.Kind() == reflect.Slice && funcSignature[i].Kind() != reflectValue.Kind() {
				return nil, fmt.Errorf("incompatible types expected: %v, but had %v", funcSignature[i].Kind(), reflectValue.Kind())
			} else if !reflectValue.IsValid() {
				if funcSignature[i].Kind() == reflect.Slice {
					parameterValue = reflect.New(funcSignature[i]).Interface()
					reflectValue = reflect.ValueOf(parameterValue)
				}
			}
		}
		if reflectValue.Type() != funcSignature[i] {
			newValuePointer := reflect.New(funcSignature[i])
			var err error
			if IsStruct(funcSignature[i]) && !(IsStruct(parameterValue) || IsMap(parameterValue)) {
				err = converter.AssignConverted(newValuePointer.Interface(), parametersKV)
			} else {
				err = converter.AssignConverted(newValuePointer.Interface(), parameterValue)
			}
			if err != nil {
				return nil, fmt.Errorf("failed to assign convert %v to %v due to %v", parametersKV, newValuePointer.Interface(), err)
			}
			reflectValue = newValuePointer.Elem()
		}
		if functionValue.Type().IsVariadic() && funcSignature[i].Kind() == reflect.Slice && i+1 == len(funcSignature) {
			ProcessSlice(reflectValue.Interface(), func(item interface{}) bool {
				functionParameters = append(functionParameters, item)
				return true
			})
		} else {
			functionParameters = append(functionParameters, reflectValue.Interface())
		}
	}
	return functionParameters, nil
}

//BuildFunctionParameters builds function parameters provided in the parameterValues.
// Parameters value will be converted if needed to expected by the function signature type. It returns function parameters , or error
func BuildFunctionParameters(function interface{}, parameters []string, parameterValues map[string]interface{}) ([]interface{}, error) {
	var functionParameters = make([]interface{}, 0)
	for _, name := range parameters {
		functionParameters = append(functionParameters, parameterValues[name])
	}
	return AsFunctionParameters(function, functionParameters, parameterValues)
}

//GetFuncSignature returns a function signature
func GetFuncSignature(function interface{}) []reflect.Type {
	AssertKind(function, reflect.Func, "function")
	functionValue := reflect.ValueOf(function)
	var result = make([]reflect.Type, 0)
	functionType := functionValue.Type()
	for i := 0; i < functionType.NumIn(); i++ {
		result = append(result, functionType.In(i))
	}
	return result
}
