package toolbox_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"net/url"
	"testing"
)

func Test_QueryValue(t *testing.T) {
	u, err := url.Parse("http://localhost/?k1=v1&k2=2&k3=false")
	assert.Nil(t, err)

	assert.Equal(t, "v1", toolbox.QueryValue(u, "k1", "default"))
	assert.Equal(t, "default", toolbox.QueryValue(u, "k10", "default"))

	assert.Equal(t, 2, toolbox.QueryIntValue(u, "k2", 3))
	assert.Equal(t, 3, toolbox.QueryIntValue(u, "k10", 3))

	assert.Equal(t, false, toolbox.QueryBoolValue(u, "k3", true))
	assert.Equal(t, true, toolbox.QueryBoolValue(u, "k10", true))

}
