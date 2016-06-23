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

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
)

func TestExtractURIParameters(t *testing.T) {
	{
		parameters, matched := toolbox.ExtractURIParameters("/v1/path/{ids}/{sub}/a/{name}", "/v1/path/1,2,3,4,5/subpath/a/abc")
		assert.True(t, matched)
		assert.Equal(t, 3, len(parameters))
		assert.Equal(t, "1,2,3,4,5", parameters["ids"])
		assert.Equal(t, "subpath", parameters["sub"])
		assert.Equal(t, "abc", parameters["name"])
	}
	{
		_, matched := toolbox.ExtractURIParameters("/v2/path/{ids}/{sub}/a/{name}", "/v1/path/1,2,3,4,5/subpath/a/abc")
		assert.False(t, matched)

	}
	{
		_, matched := toolbox.ExtractURIParameters("/v1/path1/{ids}/{sub}/a/{name}", "/v1/path/1,2,3,4,5/subpath/a/abc")
		assert.False(t, matched)

	}

	{
		_, matched := toolbox.ExtractURIParameters("/v1/path1/{ids}/{sub}/a/{name}", "/v1/path/1,2,3,4,5/subpath/a/abc")
		assert.False(t, matched)
	}

	{
		_, matched := toolbox.ExtractURIParameters("/v1/path/{ids}", "/v1/path/1")
		assert.True(t, matched)
	}

	{
		_, matched := toolbox.ExtractURIParameters("/v1/reverse/", "/v1/reverse/")
		assert.True(t, matched)
	}

	{
		parameters, matched := toolbox.ExtractURIParameters("/v1/path/{ids}/{sub}/a/{name}", "/v1/path/1,2,3,4,5/subpath/a/abcwrwr")
		assert.True(t, matched)
		assert.Equal(t, 3, len(parameters))
		assert.Equal(t, "1,2,3,4,5", parameters["ids"])
		assert.Equal(t, "subpath", parameters["sub"])
		assert.Equal(t, "abcwrwr", parameters["name"])
	}
}
