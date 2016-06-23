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

// Package toolbox - uri utilities
package toolbox

//ExtractURIParameters parses URIs to extract {<param>} defined in templateURI from requestURI, it returns extracted parameters and flag if requestURI matched templateURI
func ExtractURIParameters(templateURI, requestURI string) (map[string]string, bool) {
	var expectingValue, expectingName bool
	var name, value string
	var uriParameters = make(map[string]string)
	maxLength := len(templateURI) + len(requestURI)
	var requestURIIndex, templateURIIndex int
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
			requestURIIndex++
			if requestChar == routingChar {
				templateURIIndex++
			}
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

		if !expectingValue && !expectingName {
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
