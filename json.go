package toolbox

import "strings"

//IsCompleteJSON returns true if supplied represent complete JSON
func IsCompleteJSON(candidate string) bool {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return false
	}
	curlyStart := strings.Count(candidate, "{")
	curlyEnd := strings.Count(candidate, "}")
	squareStart := strings.Count(candidate, "[")
	squareEnd := strings.Count(candidate, "]")
	if !(curlyStart == curlyEnd && squareStart == squareEnd) {
		return false
	}
	var aMap = make(map[string]interface{})
	err := jsonDecoderFactory{}.Create(strings.NewReader(candidate)).Decode(&aMap)
	return err == nil
}

//IsNewLineDelimitedJSON returns true if supplied content is multi line delimited JSON
func IsNewLineDelimitedJSON(candidate string) bool {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return false
	}
	lines := strings.Split(candidate, "\n")
	if len(lines) == 1 {
		return false
	}
	return IsCompleteJSON(lines[0]) && IsCompleteJSON(lines[1])
}
