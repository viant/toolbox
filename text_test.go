package toolbox

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func TestIsASCIIText(t *testing.T) {

	var useCases = []struct {
		Description string
		Candidate   string
		Expected    bool
	}{
		{
			Description: "basic text",
			Candidate:   `abc`,
			Expected:    true,
		},
		{
			Description: "JSON object like text",
			Candidate:   `{"k1"}`,
			Expected:    true,
		},
		{
			Description: "JSON array like text",
			Candidate:   `["$k1"]`,
			Expected:    true,
		},
		{
			Description: "bin data",
			Candidate:   "\u0000",
			Expected:    false,
		},
		{
			Description: "JSON  text",
			Candidate: `{
  "RepositoryDatastore":"db1",
  "Db": [
    {
      "Name": "db1",
      "Config": {
        "PoolSize": 3,
        "MaxPoolSize": 5,
        "DriverName": "mysql",
        "Descriptor": "[username]:[password]@tcp(127.0.0.1:3306)/db1?parseTime=true",
        "Credentials": "$mysqlCredentials"
      }
    }
  ]
}
`,
			Expected: true,
		},
	}

	for _, useCase := range useCases {
		assert.EqualValues(t, useCase.Expected, IsASCIIText(useCase.Candidate), useCase.Description)
	}
}

func TestIsPrintText(t *testing.T) {
	var useCases = []struct {
		Description string
		Candidate   string
		Expected    bool
	}{
		{
			Description: "basic text",
			Candidate:   `abc`,
			Expected:    true,
		},
		{
			Description: "JSON object like text",
			Candidate:   `{"k1"}`,
			Expected:    true,
		},
		{
			Description: "JSON array like text",
			Candidate:   `["$k1"]`,
			Expected:    true,
		},
		{
			Description: "bin data",
			Candidate:   "\u0000",
			Expected:    false,
		},
		{
			Description: "JSON  text",
			Candidate: `{
  "RepositoryDatastore":"db1",
  "Db": [
    {
      "Name": "db1",
      "Config": {
        "PoolSize": 3,
        "MaxPoolSize": 5,
        "DriverName": "mysql",
        "Descriptor": "[username]:[password]@tcp(127.0.0.1:3306)/db1?parseTime=true",
        "Credentials": "mysql"
      }
    }
  ]
}
`,
			Expected: true,
		},
	}

	for _, useCase := range useCases {
		assert.EqualValues(t, useCase.Expected, IsPrintText(useCase.Candidate), useCase.Description)
	}
}

func TestTerminatedSplitN(t *testing.T) {
	var data = make([]byte, 0)
	for i := 0; i < 9; i++ {
		data = append(data, []byte(fmt.Sprintf("%v %v\n", strings.Repeat("x", 32), i))...)
	}
	text := string(data)

	useCases := []struct {
		description           string
		fragmentCount         int
		expectedFragmentSizes []int
	}{
		{
			description:           "one fragment case",
			fragmentCount:         1,
			expectedFragmentSizes: []int{len(data)},
		},
		{
			description:           "two fragments case",
			fragmentCount:         2,
			expectedFragmentSizes: []int{175, 140},
		},
		{
			description:           "3 fragments case",
			fragmentCount:         3,
			expectedFragmentSizes: []int{140, 140, 35},
		},
		{
			description:           "7 fragments case",
			fragmentCount:         7,
			expectedFragmentSizes: []int{70, 70, 70, 70, 35},
		},
		{
			description:           "10 fragments case", //no more fragments then lines, so only 9 fragments here
			fragmentCount:         10,
			expectedFragmentSizes: []int{35, 35, 35, 35, 35, 35, 35, 35, 35},
		},
	}

	for _, useCase := range useCases {
		fragments := TerminatedSplitN(text, useCase.fragmentCount, "\n")
		var actualFragmentSizes = make([]int, len(fragments))
		for i, fragment := range fragments {
			actualFragmentSizes[i] = len(fragment)
		}
		assert.EqualValues(t, useCase.expectedFragmentSizes, actualFragmentSizes, useCase.description)
	}
}

type testWriter struct {
	*bytes.Buffer
	data *[]string
}

func (t *testWriter) Close() error {
	*t.data = append(*t.data, t.String())
	return nil
}

func newTestWriter(data *[]string) io.WriteCloser {
	return &testWriter{
		data:   data,
		Buffer: new(bytes.Buffer),
	}
}

func TestSplitTextStream(t *testing.T) {

	var data = make([]byte, 0)
	for i := 0; i < 9; i++ {
		data = append(data, []byte(fmt.Sprintf("%v %v\n", strings.Repeat("x", 2), i))...)
	}
	text := string(data)

	useCases := []struct {
		description string
		elements    int
		expect      []string
	}{
		{
			description: "no more then 4 lines case",
			elements:    4,
			expect: []string{
				"xx 0\nxx 1\nxx 2\nxx 3", "xx 4\nxx 5\nxx 6\nxx 7", "xx 8",
			},
		},
		{
			description: "3 elements each",
			elements:    3,
			expect: []string{
				"xx 0\nxx 1\nxx 2", "xx 3\nxx 4\nxx 5", "xx 6\nxx 7\nxx 8",
			},
		},
		{
			description: "9 elements",
			elements:    1,
			expect: []string{
				"xx 0", "xx 1", "xx 2", "xx 3", "xx 4", "xx 5", "xx 6", "xx 7", "xx 8",
			},
		},
		{
			description: "9 elements",
			elements:    0,
			expect: []string{
				"xx 0", "xx 1", "xx 2", "xx 3", "xx 4", "xx 5", "xx 6", "xx 7", "xx 8",
			},
		},
		{
			description: "1 elements",
			elements:    10,
			expect: []string{
				"xx 0\nxx 1\nxx 2\nxx 3\nxx 4\nxx 5\nxx 6\nxx 7\nxx 8",
			},
		},
		{
			description: "1 elements",
			elements:    9,
			expect: []string{
				"xx 0\nxx 1\nxx 2\nxx 3\nxx 4\nxx 5\nxx 6\nxx 7\nxx 8",
			},
		},
	}

	for _, useCase := range useCases {
		var data = make([]string, 0)
		err := SplitTextStream(strings.NewReader(text), func() io.WriteCloser { return newTestWriter(&data) }, useCase.elements)
		assert.Nil(t, err)
		assert.EqualValues(t, useCase.expect, data, useCase.description)
	}
}

func Test_CaseFormat(t *testing.T) {
	var useCases = []struct {
		description string
		caseFrom    int
		caseTo      int
		input       string
		expect      string
	}{
		{
			description: "camel to uppercase",
			input:       "thisIsMyTest",
			caseFrom:    CaseLowerCamel,
			caseTo:      CaseUpper,
			expect:      "THISISMYTEST",
		},
		{
			description: "camel to lower underscore",
			input:       "thisIsMyTest",
			caseFrom:    CaseLowerCamel,
			caseTo:      CaseLowerUnderscore,
			expect:      "this_is_my_test",
		},
		{
			description: "camel to upper underscore",
			input:       "thisIsMyTest",
			caseFrom:    CaseLowerCamel,
			caseTo:      CaseUpperUnderscore,
			expect:      "THIS_IS_MY_TEST",
		},
		{
			description: "lower underscore to upper camel",
			input:       "this_is_my_test",
			caseFrom:    CaseLowerUnderscore,
			caseTo:      CaseUpperCamel,
			expect:      "ThisIsMyTest",
		},
		{
			description: "upper underscore to lower camel",
			input:       "THIS_IS_MY_TEST",
			caseFrom:    CaseUpperUnderscore,
			caseTo:      CaseLowerCamel,
			expect:      "thisIsMyTest",
		},

		{
			description: "upper camel to lower camel",
			input:       "ThisIsMyTest",
			caseFrom:    CaseUpperCamel,
			caseTo:      CaseLowerCamel,
			expect:      "thisIsMyTest",
		},
	}

	for _, useCase := range useCases {
		actual := ToCaseFormat(useCase.input, useCase.caseFrom, useCase.caseTo)
		assert.Equal(t, useCase.expect, actual, useCase.description)
	}

}
