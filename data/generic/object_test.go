package dynamic

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestObject_SetValue(t *testing.T) {

	var useCases = []struct {
		description string
		values      map[string]interface{}
	}{
		{
			description:"single value",
			values:map[string]interface{} {
				"K1": "123",
			},
		},
		{
			description:"multi value",
			values:map[string]interface{} {
				"K1": "123",
				"K2": nil,
				"K3": 4.5,
			},
		},


	}

	for _, useCase := range useCases {

		provider := NewProvider()
		object := provider.NewObject()

		for k, v := range useCase.values {
			object.SetValue(k, v)
		}
		for k, v := range useCase.values {
			assert.Equal(t, v, object.GetValue(k), useCase.description)
		}

	}

}


func TestObject_Set(t *testing.T) {

	var useCases = []struct {
		description string
		values      map[string]interface{}
	}{
		{
			description:"single value",
			values:map[string]interface{} {
				"K1": "123",
			},
		},
		{
			description:"multi value",
			values:map[string]interface{} {
				"K1": "123",
				"K2": nil,
				"K3": 4.5,
			},
		},


	}

	for _, useCase := range useCases {
		provider := NewProvider()
		object := provider.NewObject()
		object.Set(useCase.values)
		for k, v := range useCase.values {
			assert.Equal(t, v, object.GetValue(k), useCase.description)
		}

	}

}
