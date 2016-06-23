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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
)

func TestTimeFormat(t *testing.T) {

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
		dateLaout := toolbox.DateFormatToLayout("yyyy-MM-dd HH:mm:ss z")
		timeValue, err := time.Parse(dateLaout, "2016-06-02 21:46:19 UTC")
		assert.Nil(t, err)
		assert.Equal(t, int64(1464903979), timeValue.Unix())
	}

}

func TestGetTimeLayout(t *testing.T) {
	{
		settings := map[string]string{
			toolbox.DateFormatKeyword:"yyyy-MM-dd HH:mm:ss z",
		}
		assert.Equal(t, "2006-1-02 15:04:05 MST", toolbox.GetTimeLayout(settings))
		assert.True(t, toolbox.HasTimeLayout(settings))
	}
	{
		settings := map[string]string{
			toolbox.DateLayoutKeyword:"2006-1-02 15:04:05 MST",
		}
		assert.Equal(t, "2006-1-02 15:04:05 MST", toolbox.GetTimeLayout(settings))
		assert.True(t, toolbox.HasTimeLayout(settings))
	}
	{
		settings := map[string]string{

		}
		assert.False(t, toolbox.HasTimeLayout(settings))

	}
}
