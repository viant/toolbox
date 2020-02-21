package udf

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/data"
	"strings"
	"testing"
	"time"
)

func Test_FormatTime(t *testing.T) {

	{
		value, err := FormatTime([]interface{}{"now", "yyyy"}, nil)
		assert.Nil(t, err)
		now := time.Now()
		assert.Equal(t, now.Year(), toolbox.AsInt(value))
	}
	{
		value, err := FormatTime([]interface{}{"2015-02-11", "yyyy"}, nil)
		assert.Nil(t, err)
		assert.Equal(t, 2015, toolbox.AsInt(value))
	}
	{
		_, err := FormatTime([]interface{}{"2015-02-11"}, nil)
		assert.NotNil(t, err)
	}
	{
		_, err := FormatTime([]interface{}{"201/02/11 2", "y-d"}, nil)
		assert.NotNil(t, err)
	}
	{
		_, err := FormatTime("a", nil)
		assert.NotNil(t, err)
	}

	{
		value, err := FormatTime([]interface{}{"now", "yyyy", "UTC"}, nil)
		assert.Nil(t, err)
		now := time.Now()
		assert.Equal(t, now.Year(), toolbox.AsInt(value))
	}
	{
		aMap := data.NewMap()
		aMap.Put("ts", "2015-02-11")
		Register(aMap)
		expanded := aMap.ExpandAsText(`$FormatTime($ts, "yyyy")`)
		assert.Equal(t, "2015", expanded)
	}

	{
		value, err := FormatTime([]interface{}{"now", "yyyy-MM-dd HH:mm:ss", "", "yyyy-MM"}, nil)
		assert.Nil(t, err)
		now := time.Now()
		assert.True(t, strings.HasPrefix(toolbox.AsString(value), toolbox.AsString(now.Year())))
	}

}

func TestElapsed(t *testing.T) {
	{
		value, err := Elapsed(time.Now().Add(-time.Hour).Format(time.RFC3339), nil)
		assert.Nil(t, err)
		assert.Equal(t, "1h0s", value)
	}
}
