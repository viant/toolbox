package toolbox_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
)

func TestNewEnvValueProvider(t *testing.T) {
	provider := toolbox.NewEnvValueProvider()
	{
		_, err := provider.Get(nil, "USER")
		assert.Nil(t, err)
	}
	{
		_, err := provider.Get(nil, "_blahblah")
		assert.NotNil(t, err)
	}

}

func TestNewCastedValueProvider(t *testing.T) {
	provider := toolbox.NewCastedValueProvider()

	for _, source := range []interface{}{2, "2"} {
		value, err := provider.Get(nil, "int", source)
		assert.Nil(t, err)
		assert.Equal(t, 2, value)
	}

	{
		_, err := provider.Get(nil, "int")
		assert.NotNil(t, err, "Invalid number of parameters")
	}

	for _, source := range []interface{}{2, "2", 2.0} {
		value, err := provider.Get(nil, "float", source)
		assert.Nil(t, err)
		assert.Equal(t, 2.0, value)
	}

	for _, source := range []interface{}{true, "true", 1} {
		value, err := provider.Get(nil, "bool", source)
		assert.Nil(t, err)
		assert.Equal(t, true, value)
	}
	for _, source := range []interface{}{1, "1"} {
		value, err := provider.Get(nil, "string", source)
		assert.Nil(t, err)
		assert.Equal(t, "1", value)
	}

	{
		value, err := provider.Get(nil, "time", "2016-02-22 12:32:01 UTC", toolbox.DateFormatToLayout("yyyy-MM-dd hh:mm:ss z"))
		assert.Nil(t, err)
		timeValue := value.(time.Time)
		assert.Equal(t, int64(1456144321), timeValue.Unix())
	}

	{
		_, err := provider.Get(nil, "time", "2016/02-22 12:32:01 UTC", toolbox.DateFormatToLayout("yyyy-MM-dd hh:mm:ss"))
		assert.NotNil(t, err, "invalid format")
	}

	{
		_, err := provider.Get(nil, "time", "2016-02-22 12:32:01 UTC", toolbox.DateFormatToLayout("yyyy-MM-dd hh:mm:ss z"), "1")
		assert.NotNil(t, err, "to many parameters")
	}

	{
		_, err := provider.Get(nil, "ABC", "1")
		assert.NotNil(t, err, "NOT IMPLEMENTED")
	}
}

func TestNewWeekdayProvider(t *testing.T) {
	provider := toolbox.NewWeekdayProvider()
	value, err := provider.Get(toolbox.NewContext(), nil)
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, toolbox.AsInt(value), int(time.Now().Weekday()))
}

func TestNewCurrentTimeProvider(t *testing.T) {
	provider := toolbox.NewCurrentTimeProvider()
	value, err := provider.Get(nil)
	assert.Nil(t, err)
	assert.NotNil(t, value)
}

func TestNewCurrentDateProvider(t *testing.T) {
	provider := toolbox.NewCurrentDateProvider()
	value, err := provider.Get(nil)
	assert.Nil(t, err)
	assert.NotNil(t, value)
}

func TestNewNilProvider(t *testing.T) {
	provider := toolbox.NewNilValueProvider()
	value, err := provider.Get(nil)
	assert.Nil(t, err)
	assert.Nil(t, value)
}

func TestNewValueProviderRegistry(t *testing.T) {
	registry := toolbox.NewValueProviderRegistry()
	assert.False(t, registry.Contains("a"))
	registry.Register("a", toolbox.NewNilValueProvider())
	assert.True(t, registry.Contains("a"))
	provider := registry.Get("a")
	assert.NotNil(t, provider)
	assert.Equal(t, 1, len(registry.Names()))
}

func TestNewDictionaryProviderRegistry(t *testing.T) {

	var dictionary toolbox.MapDictionary = make(map[string]interface{})
	var key toolbox.MapDictionary

	dictionary["k1"] = "123"
	dictionary["k2"] = "xyz"

	provider := toolbox.NewDictionaryProvider(&key)
	context := toolbox.NewContext()
	context.Put(&key, &dictionary)

	{
		value, err := provider.Get(context, "k1")
		assert.Nil(t, err)
		assert.Equal(t, "123", value)

	}
	{
		value, err := provider.Get(context, "k2")
		assert.Nil(t, err)
		assert.Equal(t, "xyz", value)

	}

	{
		value, err := provider.Get(context, "k13", "true")
		assert.NotNil(t, err)
		assert.Nil(t, value)

	}

}

func Test_NewNewTimeProvider(t *testing.T) {

	var now = time.Now()
	provider := toolbox.NewTimeDiffProvider()

	{
		result, err := provider.Get(nil, "now", 1, "day")
		assert.Nil(t, err)

		var timeResult = toolbox.AsTime(result, "")
		in23Hours := now.Add(23 * time.Hour)
		in25Hours := now.Add(25 * time.Hour)
		assert.True(t, timeResult.After(in23Hours))
		assert.True(t, timeResult.Before(in25Hours))
	}

	{
		result, err := provider.Get(nil, "now", 1, "hour", "timestamp")
		assert.Nil(t, err)

		var timeResult = toolbox.AsInt(result)
		in59Mins := int(now.Add(59*time.Minute).Unix() * 1000)
		in61Mins := int(now.Add(61*time.Minute).Unix() * 1000)
		assert.True(t, in59Mins < timeResult)
		assert.True(t, timeResult < in61Mins)
	}

	{
		result, err := provider.Get(nil, "now", 1, "week", "unix")
		assert.Nil(t, err)

		var timeResult = toolbox.AsInt(result)
		in6Days := int(now.Add(6 * 24 * time.Hour).Unix())
		in8Days := int(now.Add(8 * 24 * time.Hour).Unix())
		assert.True(t, in6Days < timeResult)
		assert.True(t, timeResult < in8Days)
	}
	{
		result, err := provider.Get(nil, "now", 1, "hour", "h")
		assert.Nil(t, err)
		assert.Equal(t, time.Now().Hour()%12+1, toolbox.AsInt(result))
	}
}

func Test_NewDateOfBirthValueProvider(t *testing.T) {
	//provider := toolbox.NewDateOfBirthrovider()

	//{
	//	result, err := provider.Get(toolbox.NewContext(), 3, 6, 3)
	//	assert.Nil(t, err)
	//	assert.EqualValues(t, "2016-06-03", toolbox.AsString(result))
	//}
	//
	//{
	//	result, err := provider.Get(toolbox.NewContext(), 3, 6, 3, "yyyy-MM-dd")
	//	assert.Nil(t, err)
	//	assert.EqualValues(t, "2016-06-03", toolbox.AsString(result))
	//}
	//
	//{
	//	result, err := provider.Get(toolbox.NewContext(), 3, 6, 3, "yyyy")
	//	assert.Nil(t, err)
	//	assert.EqualValues(t, "2016", toolbox.AsString(result))
	//}
	//
	//{
	//	result, err := provider.Get(toolbox.NewContext(), 3, 9, 2, "yyyy-MM")
	//	assert.Nil(t, err)
	//	assert.EqualValues(t, "2016-09", toolbox.AsString(result))
	//}
	//
	//{
	//	result, err := provider.Get(toolbox.NewContext(), 5, 12, 25, "-MM-dd")
	//	assert.Nil(t, err)
	//	assert.EqualValues(t, "-12-25", toolbox.AsString(result))
	//}
	//
	//{
	//	_, err := provider.Get(toolbox.NewContext())
	//	assert.NotNil(t, err)
	//
	//}

}
