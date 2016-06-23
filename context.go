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
	"reflect"
)

//Context represents type safe map.
type Context interface {

	//GetRequired returns a value for a target type of panic if it does not exist
	GetRequired(targetType interface{}) interface{}

	//GetRequired  returns a value for a target type
	GetOptional(targetType interface{}) interface{}

	//Put puts target type value to the context, or panic if value exists,  is nil or incompatible with target type
	Put(targetType interface{}, value interface{})

	//Replace repaces value in the context
	Replace(targetType interface{}, value interface{})

	//Remove removes value from the context
	Remove(targetType interface{}) interface{}

	//Contains chekcs if a value of a terget type is in contet
	Contains(targetType interface{}) bool
}

type contextImpl struct {
	context map[string]interface{}
}

func (c *contextImpl) getReflectType(targetType interface{}) reflect.Type {
	var reflectType reflect.Type
	var ok bool
	reflectType, ok = targetType.(reflect.Type)
	if !ok {
		reflectType = reflect.TypeOf(targetType)
	}
	return reflectType
}

func (c *contextImpl) getKey(targetType interface{}) string {
	var reflectType = c.getReflectType(targetType)
	return reflectType.String()
}

func (c *contextImpl) GetRequired(targetType interface{}) interface{} {
	if !c.Contains(targetType) {
		key := c.getKey(targetType)
		panic("Failed to lookup key:" + key)
	}
	return c.GetOptional(targetType)
}

func (c *contextImpl) GetOptional(targetType interface{}) interface{} {
	key := c.getKey(targetType)
	if result, ok := c.context[key]; ok {
		return result
	}
	return nil
}

func (c *contextImpl) Put(targetType interface{}, value interface{}) {
	if c.Contains(targetType) {
		key := c.getKey(targetType)
		panic("Failed to put key - already exist: " + key)
	}
	c.Replace(targetType, value)
}

func (c *contextImpl) Replace(targetType interface{}, value interface{}) {
	key := c.getKey(targetType)
	targetReflectType := c.getReflectType(targetType)
	valueReflectType := reflect.TypeOf(value)
	if valueReflectType == targetReflectType {
		c.context[key] = value
		return
	}
	if valueReflectType.AssignableTo(targetReflectType) {
		panic(fmt.Sprintf("value of type %v is not assignable to %v", valueReflectType, targetReflectType))
	}
	if targetReflectType.Kind() == reflect.Ptr {
		converted := reflect.ValueOf(value).Elem().Convert(targetReflectType.Elem())
		convertedPointer := reflect.New(targetReflectType.Elem())
		convertedPointer.Elem().Set(converted)
		value = convertedPointer.Interface()

	} else {
		value = reflect.ValueOf(value).Convert(targetReflectType).Interface()
	}
	c.context[key] = value
}

func (c *contextImpl) Remove(targetType interface{}) interface{} {
	key := c.getKey(targetType)
	result := c.GetOptional(targetType)
	delete(c.context, key)
	return result
}

func (c *contextImpl) Contains(targetType interface{}) bool {
	key := c.getKey(targetType)
	if _, ok := c.context[key]; ok {
		return true
	}
	return false
}

//NewContext creates a new context
func NewContext() Context {
	var result Context = &contextImpl{context: make(map[string]interface{})}
	return result
}
