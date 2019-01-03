package toolbox

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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
