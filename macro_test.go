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
package toolbox_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
)

func TestMacroExpansion(t *testing.T) {
	valueRegistry := toolbox.NewValueProviderRegistry()
	valueRegistry.Register("abc", TestValueProvider{"Called with %v %v!"})

	evaluator := toolbox.MacroEvaluator{ValueProviderRegistry: valueRegistry, Prefix: "<ds:", Postfix: ">"}
	{
		//simple macro test

		actual, err := evaluator.Expand(nil, "<ds:abc[]>")
		if err != nil {
			t.Errorf("Failed expand macro %v", err.Error())
		}
		assert.Equal(t, "Called with %v %v!", actual)
	}

	{
		//simple macro test
		actual, err := evaluator.Expand(nil, "<ds:abc>")
		if err != nil {
			t.Errorf("Failed expand macro %v", err.Error())
		}
		assert.Equal(t, "Called with %v %v!", actual)
	}

	{
		//simple macro with arguments

		actual, err := evaluator.Expand(nil, "<ds:abc [1, true]>")
		if err != nil {
			t.Errorf("Failed expand macro %v", err.Error())
		}
		assert.Equal(t, "Called with 1 true!", actual)
	}
	{
		//simple macro with arguments

		actual, err := evaluator.Expand(nil, "<ds:abc [1, true]> <ds:abc [2, false]>")
		if err != nil {
			t.Errorf("Failed expand macro %v", err.Error())
		}
		assert.Equal(t, "Called with 1 true! Called with 2 false!", actual)
	}

	{
		//embeded macro with arguments

		actual, err := evaluator.Expand(nil, "<ds:abc [1, \"<ds:abc [10,11]>\"]>")
		if err != nil {
			t.Errorf("Failed expand macro %v", err.Error())
		}
		assert.Equal(t, "Called with 1 Called with 10 11!!", actual)
	}

	{
		//simple macro with arguments

		_, err := evaluator.Expand(nil, "<ds:agg>")
		assert.NotNil(t, err)

	}
}
type TestValueProvider struct {
	expandeWith string
}

func (this TestValueProvider) Init() error {
	return nil
}

func (this TestValueProvider) Get(context toolbox.Context, arguments ...interface{}) (interface{}, error) {
	if len(arguments) > 0 {
		return fmt.Sprintf(this.expandeWith, arguments...), nil
	}
	return this.expandeWith, nil
}

func (this TestValueProvider) Destroy() error {
	return nil
}


func TestExpandParameters(t *testing.T) {
	valueRegistry := toolbox.NewValueProviderRegistry()
	valueRegistry.Register("abc", TestValueProvider{"Called with %v %v!"})
	evaluator := toolbox.MacroEvaluator{ValueProviderRegistry: valueRegistry, Prefix: "<ds:", Postfix: ">"}
	aMap := map[string]string {
		"k1": "!<ds:abc>!",
	}

	err := toolbox.ExpandParameters(&evaluator, aMap)
	assert.Nil(t, err)
	assert.Equal(t, "!Called with %v %v!!", aMap["k1"])
}

func TestExpandValue(t *testing.T) {
	valueRegistry := toolbox.NewValueProviderRegistry()
	valueRegistry.Register("abc", TestValueProvider{"Called with %v %v!"})
	evaluator := toolbox.MacroEvaluator{ValueProviderRegistry: valueRegistry, Prefix: "<ds:", Postfix: ">"}
	expanded, err:= toolbox.ExpandValue(&evaluator, "!<ds:abc>!")
	assert.Nil(t, err)
	assert.Equal(t, "!Called with %v %v!!", expanded)
}