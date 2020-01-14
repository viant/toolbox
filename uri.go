package toolbox

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
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



//URLStripPath removes path from URL
func URLStripPath(URL string) string {
	protoIndex := strings.Index(URL, "://")
	if protoIndex != -1 {
		pathIndex := strings.Index(string(URL[protoIndex+3:]), "/")
		if pathIndex != -1 {
			return string(URL[:protoIndex+3+pathIndex])
		}
	}
	return URL
}

//URLPathJoin joins URL paths
func URLPathJoin(baseURL, path string) string {
	if path == "" {
		return baseURL
	}
	if strings.HasPrefix(path, "/") {
		return URLStripPath(baseURL) + path
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return baseURL + path
}

//URLBase returns base URL
func URLBase(URL string) string {
	parsedURL, err := url.Parse(URL)
	if err != nil || parsedURL.Path == "" {
		return URL
	}
	pathPosition := strings.Index(URL, parsedURL.Path)
	if pathPosition == -1 {
		return URL
	}
	return string(URL[:pathPosition])
}

//URLSplit returns URL with parent path and resource name
func URLSplit(URL string) (string, string) {
	parsedURL, err := url.Parse(URL)
	if err != nil || parsedURL.Path == "" {
		return URL, ""
	}
	splitPosition := strings.LastIndex(parsedURL.Path, "/")
	if splitPosition == -1 {
		return URL, ""
	}
	return fmt.Sprintf("%v%v", URLBase(URL), string(parsedURL.Path[:splitPosition])), string(parsedURL.Path[splitPosition+1:])
}

//Filename reformat file name
func Filename(filename string) string {
	if strings.Contains(filename, ":/") {
		if parsed, err := url.Parse(filename); err == nil {
			filename = parsed.Path
		}
	}
	var root = make([]string, 0)
	if strings.HasPrefix(filename, "/") {
		root = append(root, "/")
	}

	elements := append(root, strings.Split(filename, "/")...)
	filename = filepath.Join(elements...)
	return filename
}

//OpenFile open file converting path to elements and rebuling path safety with path.Join
func OpenFile(filename string) (*os.File, error) {
	var file = Filename(filename)
	var result, err = os.Open(file)
	return result, err
}
