package format

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCase_To(t *testing.T) {
	var useCases = []struct {
		description string
		caseFrom    Case
		caseTo      Case
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
		{
			description: "upper camel to underscore",
			input:       "ClientID",
			caseFrom:    CaseUpperCamel,
			caseTo:      CaseUpperUnderscore,
			expect:      "CLIENT_ID",
		},
	}

	for _, useCase := range useCases {
		actual := useCase.caseFrom.Format(useCase.input, useCase.caseTo)
		assert.Equal(t, useCase.expect, actual, useCase.description)
	}

}
