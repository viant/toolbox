package toolbox

import (
	"strings"
)

//ExtractURIParameters parses URIs to extract {<param>} defined in templateURI from requestURI, it returns extracted parameters and flag if requestURI matched templateURI
func ExtractURIParameters(templateURI, requestURI string) (map[string]string, bool) {
	var expectingValue, expectingName bool
	var name, value string
	var uriParameters = make(map[string]string)
	maxLength := len(templateURI) + len(requestURI)
	var requestURIIndex, templateURIIndex int

	questionMarkPosition := strings.Index(requestURI, "?")
	if questionMarkPosition != -1 {
		requestURI = string(requestURI[:questionMarkPosition])
	}

	for k := 0; k < maxLength; k++ {
		var requestChar, routingChar string

		if requestURIIndex < len(requestURI) {
			requestChar = requestURI[requestURIIndex : requestURIIndex+1]
		}

		if templateURIIndex < len(templateURI) {
			routingChar = templateURI[templateURIIndex : templateURIIndex+1]
		}
		if (!expectingValue && !expectingName) && requestChar == routingChar && routingChar != "" {
			requestURIIndex++
			templateURIIndex++
			continue
		}

		if routingChar == "}" {
			expectingName = false
			templateURIIndex++
		}

		if expectingValue && requestChar == "/" {
			expectingValue = false
		}

		if expectingName && templateURIIndex < len(templateURI) {
			name += routingChar
			templateURIIndex++
		}

		if routingChar == "{" {
			expectingValue = true
			expectingName = true
			templateURIIndex++

		}

		if expectingValue && requestURIIndex < len(requestURI) {
			value += requestChar
			requestURIIndex++
		}

		if !expectingValue && !expectingName && len(name) > 0 {
			uriParameters[name] = value
			name = ""
			value = ""
		}

	}

	if len(name) > 0 && len(value) > 0 {
		uriParameters[name] = value
	}
	matched := requestURIIndex == len(requestURI) && templateURIIndex == len(templateURI)
	return uriParameters, matched
}


func URLPathJoin(baseURL, path string)  string {
	if strings.HasSuffix(baseURL, "/") {
		return baseURL + path
	}
	return baseURL + "/" + path
}