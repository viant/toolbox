package udf

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox/data"
	"log"
	"reflect"
	"testing"
)

func Test_AsBool(t *testing.T) {
	ok, err := AsBool("true", nil)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
}

func Test_AsFloat(t *testing.T) {
	value, err := AsFloat(0.3, nil)
	assert.Nil(t, err)
	assert.Equal(t, 0.3, value)
}

func Test_AsInt(t *testing.T) {
	value, err := AsInt(4.3, nil)
	assert.Nil(t, err)
	assert.Equal(t, 4, value)
}

func Test_AsMap(t *testing.T) {
	{
		var aMap, err = AsMap(map[string]interface{}{}, nil)
		assert.Nil(t, err)
		assert.NotNil(t, aMap)
	}
	{
		var aMap, err = AsMap("{\"abc\":1}", nil)
		assert.Nil(t, err)
		assert.NotNil(t, aMap)
	}
	{
		var aMap, err = AsMap(`abc: 1`, nil)
		assert.Nil(t, err)
		assert.NotNil(t, aMap)
	}
	{
		_, err := AsMap("{\"abc\":1, \"a}", nil)
		assert.NotNil(t, err)
	}
}

func Test_AsCollection(t *testing.T) {

	{
		var aSlice, err = AsCollection([]interface{}{1}, nil)
		assert.Nil(t, err)
		assert.NotNil(t, aSlice)
	}
	{
		var aSlice, err = AsCollection("[1,2]", nil)
		assert.Nil(t, err)
		assert.NotNil(t, aSlice)
	}
	{
		var aSlice, err = AsCollection(`
- 1
- 2`, nil)
		assert.Nil(t, err)
		assert.NotNil(t, aSlice)
	}

	{
		_, err := AsCollection("[\"a,2]", nil)
		assert.NotNil(t, err)
	}
	{
		var aMap, err = AsData(`abc: 1`, nil)
		assert.Nil(t, err)
		assert.NotNil(t, aMap)
	}
}

func Test_AsData(t *testing.T) {
	{
		var aSlice, err = AsData("[1,2]", nil)
		assert.Nil(t, err)
		assert.NotNil(t, aSlice)

	}
	{
		var aMap, err = AsData("{\"abc\":1}", nil)
		assert.Nil(t, err)
		assert.NotNil(t, aMap)
	}

}

func Test_YamlAsCollection(t *testing.T) {
	var YAML = `- Requests:
    - URL: http://localhost:5000
      Method: GET
      Header:
        aHeader:
          - "myField=a-value; path=/; domain=localhost; Expires=Tue, 19 Jan 2038 03:14:07 GMT;"
        someOtherHeader:
          - "CP=RTO"

      Body: "hey there"
      Cookies:
        - Name: aHeader
          Value: a-value
          DYAMLomain: "localhost"
          Expires: "2023-12-16T20:17:38Z"
          RawExpires: Sat, 16 Dec 2023 20:17:38 GMT`

	expanded, err := AsCollection(YAML, nil)
	if !assert.Nil(t, err) {
		log.Fatal(err)
	}
	assert.Equal(t, reflect.Slice, reflect.TypeOf(expanded).Kind())
}

func Test_YamlAsMap(t *testing.T) {
	YAML := `default: &default
  Name: Jack
person: 
  <<: *default
  Name: Bob`

	expanded, err := AsCollection(YAML, nil)
	assert.Nil(t, err)
	assert.NotNil(t, expanded)
}

func Test_AsString(t *testing.T) {
	aMap := data.NewMap()
	Register(aMap)

	aMap.Put("k0", true)
	expanded := aMap.ExpandAsText(" $AsString(${k0})")
	assert.EqualValues(t, "true", expanded)
	//64 bit int
	aMap.Put("k1", 2323232323223)
	expanded = aMap.ExpandAsText(" $AsString(${k1})")
	assert.EqualValues(t, "2323232323223", expanded)
}
