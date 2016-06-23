/*
 *
 *
 * Copyright 2012-2016 Viant.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 *  use this file except in compliance with the License. You may obtain a copy of
 *  the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 *  License for the specific language governing permissions and limitations under
 *  the License.
 *
 */

// Package toolbox - function utilities
package toolbox

import (
	"fmt"
	"reflect"
)

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

//BuildFunctionParameters builds function parameters provided in the parameterValues.
// Parameters value will be converted if needed to expected by the function signature type. It returns function parameters , or error
func BuildFunctionParameters(function interface{}, parameters []string, parameterValues map[string]interface{}) ([]interface{}, error) {
	AssertKind(function, reflect.Func, "function")
	functionValue := reflect.ValueOf(function)
	funcSignature := GetFuncSignature(function)
	actualMethodSignatureLength := len(funcSignature)
	converter := Converter{}
	if actualMethodSignatureLength != len(parameters) {
		return nil, fmt.Errorf("Invalid number of parameters wanted: [%T],  had: %v", function, 0)
	}
	var functionParameters = make([]interface{}, 0)
	for i, name := range parameters {
		parameterValue := parameterValues[name]
		reflectValue := reflect.ValueOf(parameterValue)
		if reflectValue.Kind() == reflect.Slice && funcSignature[i].Kind() != reflectValue.Kind() {
			return nil, fmt.Errorf("Incompatible types expected: %v, but had %v", funcSignature[i].Kind(), reflectValue.Kind())
		}
		if reflectValue.Type() != funcSignature[i] {
			newValuePointer := reflect.New(funcSignature[i])
			err := converter.AssignConverted(newValuePointer.Interface(), parameterValue)
			if err != nil {
				return nil, fmt.Errorf("Failed to assign convert %v to %v due to %v", parameterValue, newValuePointer.Interface(), err)
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
