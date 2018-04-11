package toolbox

import (
	"github.com/stretchr/testify/assert"
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
  "Datastores": [
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
  "Datastores": [
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
