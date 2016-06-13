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
package toolbox

import (
	"fmt"
	"os"
	"time"
)

//ValueProvider represents a value provider
type ValueProvider interface {

	//Get returns a value for passed in context and arguments. Context can be used to manage state.
	Get(context Context, arguments ...interface{}) (interface{}, error)
}

//ValueProviderRegistry registry of value providers
type ValueProviderRegistry interface {
	Register(name string, valueProvider ValueProvider)

	Contains(name string) bool

	Names() []string

	Get(name string) ValueProvider
}

type valueProviderRegistryImpl struct {
	registry map[string](ValueProvider)
}

func (r valueProviderRegistryImpl) Register(name string, valueProvider ValueProvider) {
	r.registry[name] = valueProvider
}

func (r valueProviderRegistryImpl) Contains(name string) bool {
	_, ok := r.registry[name]
	return ok
}

func (r valueProviderRegistryImpl) Get(name string) ValueProvider {
	if result, ok := r.registry[name]; ok {
		return result
	}
	panic(fmt.Sprintf("Failed to lookup name: %v", name))
}

func (r valueProviderRegistryImpl) Names() []string {
	return MapKeysToStringSlice(&r.registry)
}

//NewValueProviderRegistry create new NewValueProviderRegistry
func NewValueProviderRegistry() ValueProviderRegistry {
	var result ValueProviderRegistry = &valueProviderRegistryImpl{
		registry: make(map[string]ValueProvider),
	}
	return result
}

type envValueProvider struct{}

func (p envValueProvider) Get(context Context, arguments ...interface{}) (interface{}, error) {
	key := arguments[0].(string)
	value, found := os.LookupEnv(key)
	if found {
		return value, nil
	}
	return nil, fmt.Errorf("Failed to lookup %v in env", key)
}

//NewEnvValueProvider returns a provider that returns a value of env variables.
func NewEnvValueProvider() ValueProvider {
	var result ValueProvider = &envValueProvider{}
	return result
}

type castedValueProvider struct{}

func (p castedValueProvider) Get(context Context, arguments ...interface{}) (interface{}, error) {
	key := arguments[0].(string)
	if len(arguments) == 2 {
		return nil, fmt.Errorf("Failed to cast to %v due to invalud number of arguments", key)
	}

	switch key {
	case "time":
		if len(arguments) != 3 {
			return nil, fmt.Errorf("Failed to cast to time due to invalud number of arguments expected 2, but had %v", len(arguments)-1)
		}
		castedTime, err := ParseTime(AsString(arguments[1]), AsString(arguments[2]))
		if err != nil {
			return nil, fmt.Errorf("Failed to cast to time %v due to %v", AsString(arguments[1]), err)
		}
		return castedTime, nil
	case "int":
		return AsInt(arguments[1]), nil
	case "float":
		return AsFloat(arguments[1]), nil
	case "bool":
		return AsBoolean(arguments[1]), nil
	case "string":
		return AsString(arguments[1]), nil

	}
	return nil, fmt.Errorf("Failed to cast to %v - unsupported type", key)
}

//NewCastedValueProvider return a provider that return casted value type
func NewCastedValueProvider() ValueProvider {
	var result ValueProvider = &castedValueProvider{}
	return result
}

type currentTimeProvider struct{}

func (p currentTimeProvider) Get(context Context, arguments ...interface{}) (interface{}, error) {
	return time.Now(), nil
}

//NewCurrentTimeProvider returns a provder that returns time.Now()
func NewCurrentTimeProvider() ValueProvider {
	var result ValueProvider = &currentTimeProvider{}
	return result
}

type nilValueProvider struct{}

func (p nilValueProvider) Get(context Context, arguments ...interface{}) (interface{}, error) {
	return nil, nil
}

//NewNilValueProvider returns a provider that returns a nil
func NewNilValueProvider() ValueProvider {
	var result ValueProvider = &nilValueProvider{}
	return result
}
