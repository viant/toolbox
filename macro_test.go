package toolbox_test

import (
	"fmt"
	"testing"

	"errors"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
)

func TestMacroExpansion(t *testing.T) {
	valueRegistry := toolbox.NewValueProviderRegistry()
	valueRegistry.Register("abc", TestValueProvider{"Called with %v %v!", nil})
	valueRegistry.Register("xyz", TestValueProvider{"XXXX", nil})
	valueRegistry.Register("klm", TestValueProvider{"Called with %v %v!", errors.New("Test error")})
	evaluator := toolbox.MacroEvaluator{ValueProviderRegistry: valueRegistry, Prefix: "<ds:", Postfix: ">"}
	{
		//simple macro test

		actual, err := evaluator.Expand(nil, "<ds:abc[]>")
		if err != nil {
			t.Errorf("failed expand macro %v", err.Error())
		}
		assert.Equal(t, "Called with %v %v!", actual)
	}

	{
		//simple macro test

		actual, err := evaluator.Expand(nil, "< <ds:abc[]>> <ds:xyz[]>")
		if err != nil {
			t.Errorf("failed expand macro %v", err.Error())
		}
		assert.Equal(t, "< Called with %v %v!> XXXX", actual)
	}

	{
		//simple macro test
		actual, err := evaluator.Expand(nil, "<ds:abc>")
		if err != nil {
			t.Errorf("failed expand macro %v", err.Error())
		}
		assert.Equal(t, "Called with %v %v!", actual)
	}

	{
		//simple macro with arguments

		actual, err := evaluator.Expand(nil, "<ds:abc [1, true]>")
		if err != nil {
			t.Errorf("failed expand macro %v", err.Error())
		}
		assert.Equal(t, "Called with 1 true!", actual)
	}
	{
		//simple macro with arguments

		actual, err := evaluator.Expand(nil, "<ds:abc [1, true]> <ds:abc [2, false]>")
		if err != nil {
			t.Errorf("failed expand macro %v", err.Error())
		}
		assert.Equal(t, "Called with 1 true! Called with 2 false!", actual)
	}

	{
		//embeded macro with arguments

		actual, err := evaluator.Expand(nil, "<ds:abc [1, \"<ds:abc [10,11]>\"]>")
		if err != nil {
			t.Errorf("failed expand macro %v", err.Error())
		}
		assert.Equal(t, "Called with 1 Called with 10 11!!", actual)
	}

	{
		//value provider with error
		_, err := evaluator.Expand(nil, "<ds:abc [1, \"<ds:klm>\"]>")
		assert.NotNil(t, err, "macro argument value provider returns error")
	}

	{
		//value provider with error
		_, err := evaluator.Expand(nil, "<ds:klm>")
		assert.NotNil(t, err, "value provider returns error")
	}

	{
		//simple macro with arguments

		_, err := evaluator.Expand(nil, "<ds:agg>")
		assert.NotNil(t, err)

	}

	{
		//value provider with error

		_, err := evaluator.Expand(nil, "<ds:pos [\"events\"]>")
		assert.NotNil(t, err)

	}
}

type TestValueProvider struct {
	expandeWith string
	err         error
}

func (t TestValueProvider) Init() error {
	return nil
}

func (t TestValueProvider) Get(context toolbox.Context, arguments ...interface{}) (interface{}, error) {
	if len(arguments) > 0 {
		return fmt.Sprintf(t.expandeWith, arguments...), t.err
	}
	return t.expandeWith, t.err
}

func (t TestValueProvider) Destroy() error {
	return nil
}

func TestExpandParameters(t *testing.T) {
	valueRegistry := toolbox.NewValueProviderRegistry()
	valueRegistry.Register("abc", TestValueProvider{"Called with %v %v!", nil})
	valueRegistry.Register("klm", TestValueProvider{"Called with %v %v!", errors.New("Test error")})
	evaluator := toolbox.MacroEvaluator{ValueProviderRegistry: valueRegistry, Prefix: "<ds:", Postfix: ">"}

	{
		aMap := map[string]string{
			"k1": "!<ds:klm>!",
		}
		err := toolbox.ExpandParameters(&evaluator, aMap)
		assert.NotNil(t, err)
	}
	{
		aMap := map[string]string{
			"k1": "!<ds:abc>!",
		}
		err := toolbox.ExpandParameters(&evaluator, aMap)
		assert.Nil(t, err)
		assert.Equal(t, "!Called with %v %v!!", aMap["k1"])
	}
}

func TestExpandValue(t *testing.T) {
	valueRegistry := toolbox.NewValueProviderRegistry()
	valueRegistry.Register("abc", TestValueProvider{"Called with %v %v!", nil})
	valueRegistry.Register("klm", TestValueProvider{"Called with %v %v!", errors.New("Test error")})
	evaluator := toolbox.MacroEvaluator{ValueProviderRegistry: valueRegistry, Prefix: "<ds:", Postfix: ">"}
	{
		expanded, err := toolbox.ExpandValue(&evaluator, "!<ds:abc>!")
		assert.Nil(t, err)
		assert.Equal(t, "!Called with %v %v!!", expanded)
	}
	{
		expanded, err := toolbox.ExpandValue(&evaluator, "!!")
		assert.Nil(t, err)
		assert.Equal(t, "!!", expanded)

	}
	{
		_, err := toolbox.ExpandValue(&evaluator, "<ds:klm>")
		assert.NotNil(t, err)
	}

}
