package toolbox_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"strings"
	"testing"
	"time"
)

func TestTimeFormat(t *testing.T) {

	{
		timeLayout := toolbox.DateFormatToLayout("yyyy/MM/dd/hh")
		fmt.Printf("%s\n", timeLayout)
	}

		{
		timeLayout := toolbox.DateFormatToLayout("yyyy-MM-dd HH:mm:ss z")
		timeValue, err := time.Parse(timeLayout, "2018-01-15 08:02:23 UTC")
		assert.Nil(t, err)
		assert.EqualValues(t, 23, timeValue.Second())
	}

	{
		timeLayout := toolbox.DateFormatToLayout("yyyy-MM-dd HH:mm:ss")
		timeValue, err := time.Parse(timeLayout, "2016-03-01 03:10:11")
		assert.Nil(t, err)
		assert.EqualValues(t, 11, timeValue.Second())
	}

	{

		dateLayout := toolbox.DateFormatToLayout("yyyy-MM-dd HH:mm:ss.SSSZ")
		timeValue, err := time.Parse(dateLayout, "2022-11-10 10:32:28.984-08")
		assert.Nil(t, err)

		assert.Equal(t, int64(1668105148), timeValue.Unix())
	}

	{
		dateLaout := toolbox.DateFormatToLayout("yyyy-MM-dd HH:mm:ss.SSSZ")
		timeValue, err := time.Parse(dateLaout, "2022-11-10 10:32:28.984-08")
		assert.Nil(t, err)
		assert.Equal(t, int64(1668105148), timeValue.Unix())
	}

	{
		dateLaout := toolbox.DateFormatToLayout("yyyy-MM-dd HH:mm:ss.SSS")
		timeValue, err := time.Parse(dateLaout, "2017-11-04 22:29:33.363")
		assert.Nil(t, err)

		assert.Equal(t, 2017, timeValue.Year())
		assert.Equal(t, time.Month(11), timeValue.Month())
		assert.Equal(t, 4, timeValue.Day())

		assert.Equal(t, int64(1509834573), timeValue.Unix())
	}

	{
		dateLaout := toolbox.DateFormatToLayout("dd/MM/yyyy hh:mm:ss")
		timeValue, err := time.Parse(dateLaout, "22/02/2016 12:32:01")
		assert.Nil(t, err)
		assert.Equal(t, int64(1456144321), timeValue.Unix())
	}

	{
		dateLaout := toolbox.DateFormatToLayout("yyyyMMdd hh:mm:ss")
		timeValue, err := time.Parse(dateLaout, "20160222 12:32:01")
		assert.Nil(t, err)
		assert.Equal(t, int64(1456144321), timeValue.Unix())
	}

	{
		dateLaout := toolbox.DateFormatToLayout("yyyy-MM-dd hh:mm:ss z")
		timeValue, err := time.Parse(dateLaout, "2016-02-22 12:32:01 UTC")
		assert.Nil(t, err)
		assert.Equal(t, int64(1456144321), timeValue.Unix())
	}

	{
		dateLaout := toolbox.DateFormatToLayout("yyyy-MM-dd hh:mm:ss z")
		timeValue, err := time.Parse(dateLaout, "2016-02-22 12:32:01 UTC")
		assert.Nil(t, err)
		assert.Equal(t, int64(1456144321), timeValue.Unix())
	}

	{
		dateLaout := toolbox.DateFormatToLayout("yyyy-MM-dd hh:mm:ss z")
		timeValue, err := time.Parse(dateLaout, "2016-02-22 12:32:01 UTC")
		assert.Nil(t, err)
		assert.Equal(t, int64(1456144321), timeValue.Unix())
	}

	{
		dateLaout := toolbox.DateFormatToLayout("yyyy-MM-dd HH:mm:ss z")
		timeValue, err := time.Parse(dateLaout, "2016-06-02 21:46:19 UTC")
		assert.Nil(t, err)
		assert.Equal(t, int64(1464903979), timeValue.Unix())
	}

}

func TestGetTimeLayout(t *testing.T) {
	{
		settings := map[string]string{
			toolbox.DateFormatKeyword: "yyyy-MM-dd HH:mm:ss z",
		}
		assert.Equal(t, "2006-01-02 15:04:05 MST", toolbox.GetTimeLayout(settings))
		assert.True(t, toolbox.HasTimeLayout(settings))
	}
	{
		settings := map[string]string{
			toolbox.DateLayoutKeyword: "2006-1-02 15:04:05 MST",
		}
		assert.Equal(t, "2006-1-02 15:04:05 MST", toolbox.GetTimeLayout(settings))
		assert.True(t, toolbox.HasTimeLayout(settings))
	}
	{
		settings := map[string]string{}
		assert.False(t, toolbox.HasTimeLayout(settings))

	}
}

func TestTimestampToString(t *testing.T) {

	{
		date := toolbox.TimestampToString("yyyy-MM-dd HH:mm:ss z", int64(0), 1480435743722684356)
		assert.True(t, strings.Contains(date, "2016-11"))
	}
	{

		date := toolbox.TimestampToString("yyyyMMddhh", int64(0), 1489512277722684356)
		assert.True(t, strings.Contains(date, "201703"))
	}

}
