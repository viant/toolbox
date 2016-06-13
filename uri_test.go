package toolbox_test

import (
	"testing"
	"github.com/viant/toolbox"
	"github.com/stretchr/testify/assert"
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
