package toolbox_test

import (
	"testing"
	"github.com/viant/toolbox"
	"github.com/stretchr/testify/assert"
	"time"
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
		_, err := provider.Get(nil, "ABC", "1")
		assert.NotNil(t, err, "NOT IMPLEMENTED")
	}
}


func TestNewCurrentTimeProvider(t *testing.T) {
	provider := toolbox.NewCurrentTimeProvider()
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