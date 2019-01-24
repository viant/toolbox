package toolbox

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

func TestNormalizeKVPairs(t *testing.T) {

	{ //yaml case
		YAML := `- Requests:
    - URL: http://localhost:5000
      Method: GET
      Header:
        aHeader:
          - "v1"
          - "v2"

        someOtherHeader:

          - "CP=RTO"

      Body: "hey there"
      Cookies:
        - Name: aHeader
          Value: a-value
          DYAMLomain: "localhost"
          Expires: "2023-12-16T20:17:38Z"
          RawExpires: Sat, 16 Dec 2023 20:17:38 GMT`

		var data interface{}
		err := yaml.NewDecoder(strings.NewReader(YAML)).Decode(&data)
		assert.Nil(t, err)
		normalized, err := NormalizeKVPairs(data)
		assert.Nil(t, err)
		requests := AsMap(AsSlice(normalized)[0])["Requests"]
		request := AsMap(AsSlice(requests)[0])
		assert.Equal(t, "http://localhost:5000", request["URL"])
		header := AsMap(request["Header"])
		assert.Equal(t, []interface{}{"v1", "v2"}, header["aHeader"])
	}
	{
		JSON := `[
{"Key":"k1", "Value":"v1"},
{"Key":"k2", "Value":"v2"},
{"Key":"k3", "Value":[
	{"Key":"k1", "Value":"v1", "Attr":2}
]}]`

		var data interface{}
		err := json.NewDecoder(strings.NewReader(JSON)).Decode(&data)
		assert.Nil(t, err)
		normalized, err := NormalizeKVPairs(data)
		assert.Nil(t, err)
		aMap := AsMap(normalized)
		assert.Equal(t, "v1", aMap["k1"])
		assert.Equal(t, "v2", aMap["k2"])
		aSlice := AsSlice(aMap["k3"])
		assert.NotNil(t, aSlice)
		anItem := AsMap(aSlice[0])
		assert.Equal(t, "k1", anItem["Key"])
		assert.Equal(t, "v1", anItem["Value"])
		assert.Equal(t, 2.0, anItem["Attr"])
	}

}
