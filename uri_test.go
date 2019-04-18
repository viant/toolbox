package toolbox_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/url"
	"testing"
)

func TestExtractURIParameters(t *testing.T) {

	{
		parameters, matched := toolbox.ExtractURIParameters("/v1/path/{app}/{version}/", "/v1/path/app/1.0/?v=12")
		assert.True(t, matched)
		if !matched {
			t.FailNow()
		}
		assert.Equal(t, 2, len(parameters))
		assert.Equal(t, "app", parameters["app"])
		assert.Equal(t, "1.0", parameters["version"])
	}

	{
		parameters, matched := toolbox.ExtractURIParameters("/v1/path/{ids}/{sub}/a/{name}", "/v1/path/1,2,3,4,5/subpath/a/abc")
		assert.True(t, matched)
		assert.Equal(t, 3, len(parameters))
		assert.Equal(t, "1,2,3,4,5", parameters["ids"])
		assert.Equal(t, "subpath", parameters["sub"])
		assert.Equal(t, "abc", parameters["name"])
	}

	{
		parameters, matched := toolbox.ExtractURIParameters("/v1/path/{ids}", "/v1/path/this-is-test")
		assert.True(t, matched)
		assert.Equal(t, 1, len(parameters))
		assert.Equal(t, "this-is-test", parameters["ids"])
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

func TestURLBase(t *testing.T) {

	URL := "http://github.com/abc"
	baseURL := toolbox.URLBase(URL)
	assert.Equal(t, "http://github.com", baseURL)

}

func TestURLSplit(t *testing.T) {

	{
		URL := "http://github.com/abc/trter/rds"
		parentURL, resource := toolbox.URLSplit(URL)
		assert.Equal(t, "http://github.com/abc/trter", parentURL)
		assert.Equal(t, "rds", resource)
	}

}

func TestURLStripPath(t *testing.T) {
	{
		URL := "http://github.com/abc"
		assert.EqualValues(t, "http://github.com", toolbox.URLStripPath(URL))
	}
	{
		URL := "http://github.com"
		assert.EqualValues(t, "http://github.com", toolbox.URLStripPath(URL))
	}
}

func TestURL_Rename(t *testing.T) {
	{
		URL := "http://github.com/abc/"
		resource := url.NewResource(URL)
		resource.Rename("/tmp/abc")
		assert.Equal(t, "http://github.com//tmp/abc", resource.URL)

	}

}

func TestURLPathJoin(t *testing.T) {

	{
		URL := "http://github.com/abc"
		assert.EqualValues(t, "http://github.com/abc/path/a.txt", toolbox.URLPathJoin(URL, "path/a.txt"))
	}
	{
		URL := "http://github.com/abc/"
		assert.EqualValues(t, "http://github.com/abc/path/a.txt", toolbox.URLPathJoin(URL, "path/a.txt"))
	}
	{
		URL := "http://github.com/abc/"
		assert.EqualValues(t, "http://github.com/a.txt", toolbox.URLPathJoin(URL, "/a.txt"))
	}
}
