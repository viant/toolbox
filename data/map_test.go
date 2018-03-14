package data

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"testing"
)

func TestMap_GetValue(t *testing.T) {

	aMap := NewMap()

	{
		subCollection := NewCollection()
		subCollection.Push("item1")
		subCollection.Push("item2")
		aMap.Put("cc", subCollection)

		value, has := aMap.GetValue("cc[0]")
		assert.True(t, has)
		assert.Equal(t, "item1", value)
	}

	{
		metaMap := make(map[string]int)
		metaMap["USER"] = 7
		aMap.Put("meta", metaMap)

		value, ok := aMap.GetValue("meta.USER")
		assert.True(t, ok)
		if !assert.Equal(t, 7, value) {
			return
		}
		aMap.SetValue("meta.USER", toolbox.AsInt(value)+1)
		value, ok = aMap.GetValue("meta.USER")
		assert.True(t, ok)
		if !assert.Equal(t, 8, value) {
			return
		}

	}

	{
		var collection = NewCollection()
		collection.Push("1")
		collection.Push("20")
		collection.Push("30")
		aMap.Put("collection", collection)

		var subMap = NewMap()
		subMap.Put("i", 10)
		subMap.Put("col", collection)
		aMap.Put("a", subMap)
		aMap.Put("b", "123")
		aMap.Put("c", "b")
	}

	{ //test simple get operation
		value, has := aMap.GetValue("c")
		assert.True(t, has)
		assert.Equal(t, "b", value)

	}

	{ //test  get operation
		value, has := aMap.GetValue("a.col")
		assert.True(t, has)
		assert.Equal(t, []interface{}{"1", "20", "30"}, toolbox.AsSlice(value))

	}
	{ //test reference get operation
		value, has := aMap.GetValue("$c")
		assert.True(t, has)
		assert.Equal(t, "123", value)

	}

	{ //test post increment operations
		value, has := aMap.GetValue("a.i++")
		assert.True(t, has)
		assert.Equal(t, 10, value)
		value, has = aMap.GetValue("a.i++")
		assert.True(t, has)
		assert.Equal(t, 11, value)
	}

	{ //test pre increment operations
		value, has := aMap.GetValue("++a.i")
		assert.True(t, has)
		assert.Equal(t, 13, value)
		value, has = aMap.GetValue("++a.i")
		assert.True(t, has)
		assert.Equal(t, 14, value)
	}

	{ //	test shift
		value, has := aMap.GetValue("<-collection")
		assert.True(t, has)
		assert.Equal(t, "1", value)
		value, has = aMap.GetValue("<-collection")
		assert.True(t, has)
		assert.Equal(t, "20", value)

	}
	{ //	test array index

		var aCollection = NewCollection()
		aCollection.Push(map[string]interface{}{
			"k1": 1,
			"K2": 2,
		})
		aCollection.Push(map[string]interface{}{
			"k2": 3,
			"K3": 4,
		})
		aMap.Put("c", aCollection)
		value, has := aMap.GetValue("c[0].k1")
		assert.True(t, has)
		assert.Equal(t, 1, value)

		value, has = aMap.GetValue("c[1].k2")
		assert.True(t, has)
		assert.Equal(t, 3, value)

	}

	{
		subMap := NewMap()
		subCollection := NewCollection()
		subCollection.Push("item1")
		subCollection.Push("item2")
		subMap.Put("c", subCollection)
		aMap.Put("s", subMap)

		value, has := aMap.GetValue("s.c[0]")
		assert.True(t, has)
		assert.Equal(t, "item1", value)
	}

}

func TestMap_SetValue(t *testing.T) {

	aMap := NewMap()

	{ // test simple Set

		_, has := aMap.GetValue("z.a")
		assert.False(t, has)
		aMap.SetValue("z.a", "123")
		value, has := aMap.GetValue("z.a")
		assert.True(t, has)
		assert.Equal(t, "123", value)
	}

	{ // test reference set

		aMap.SetValue("z.b", "111")
		value, has := aMap.GetValue("z.b")
		assert.True(t, has)
		assert.Equal(t, "111", value)

		aMap.SetValue("zzz", "z.b")
		aMap.SetValue("$zzz", "222")
		value, has = aMap.GetValue("z.b")
		assert.True(t, has)
		assert.Equal(t, "222", value)
	}

	{
		//test push
		aMap.SetValue("->a.v", 1)
		aMap.SetValue("->a.v", 2)

		aCollection, has := aMap.GetValue("a.v")
		assert.True(t, has)
		assert.Equal(t, []interface{}{1, 2}, toolbox.AsSlice(aCollection))
	}

}

func Test_Expand(t *testing.T) {

	state := NewMap()
	state.Put("name", "etly")
	build := NewMap()
	state.Put("build", build)
	build.Put("Target", "app")
	build.Put("Args", "-Dmvn.test.skip")

	var text = state.ExpandAsText("a $vv-ee /usr/local/app_${name}v1 $build.Target $abc $build.Args")
	assert.Equal(t, "a $vv-ee /usr/local/app_etlyv1 app $abc -Dmvn.test.skip", text)

}

func Test_ExpandFun(t *testing.T) {

	state := NewMap()
	state.Put("name", "etly")
	build := NewMap()
	state.Put("build", build)
	build.Put("Target", "app")
	build.Put("Args", "-Dmvn.test.skip")

	var text = state.ExpandAsText("a $vv-ee /usr/local/app_${name}v1 $build.Target $abc $build.Args")
	assert.Equal(t, "a $vv-ee /usr/local/app_etlyv1 app $abc -Dmvn.test.skip", text)

}

func Test_Udf(t *testing.T) {

	var test = func(s interface{}, m Map) (interface{}, error) {
		return fmt.Sprintf("%v", s), nil
	}

	var dateOfBirth = func(source interface{}, m Map) (interface{}, error) {
		if !toolbox.IsSlice(source) {
			return nil, fmt.Errorf("expected slice but had: %T %v", source, source)
		}
		return toolbox.NewDateOfBirthrovider().Get(toolbox.NewContext(), toolbox.AsSlice(source)...)
	}

	state := NewMap()
	state.Put("test", test)
	state.Put("name", "endly")
	state.Put("a", "1")
	state.Put("b", "2")
	state.Put("Dob", dateOfBirth)

	{
		var text = "$Dob([11,2,2,\"yyyy\"])"
		expanded := state.Expand(text)
		assert.EqualValues(t, "2007", expanded)

	}
	{
		state.Put("args", []interface{}{11, 2, 2, "yyyy"})

		var text = "$Dob($args)"
		expanded := state.Expand(text)
		assert.EqualValues(t, "2007", expanded)

	}

	{
		var text = "$xyz($name)"
		expanded := state.Expand(text)
		assert.EqualValues(t, "$xyz(endly)", expanded)

	}

	{
		var text = "$xyz(hello $name $abc)"
		expanded := state.Expand(text)
		assert.EqualValues(t, "$xyz(hello endly $abc)", expanded)

	}

	{
		var text = "$test(hello $abc)"
		expanded := state.Expand(text)
		assert.EqualValues(t, "$test(hello $abc)", expanded)
	}

	{
		var text = "$test(hello $name $abc)"
		expanded := state.Expand(text)
		assert.EqualValues(t, "$test(hello endly $abc)", expanded)
	}

	{
		var text = "$test(hello $name)"
		expanded := state.Expand(text)
		assert.EqualValues(t, "hello endly", expanded)
	}

	{
		var text = "zz $a ${b}a"
		expanded := state.Expand(text)
		assert.EqualValues(t, "zz 1 2a", expanded)

	}

}
