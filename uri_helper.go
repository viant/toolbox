package toolbox

import (
	"net/url"
	"path"
)

//FileSchema file://
var FileSchema = "file://"

//Deprecated start using url.Resource

//ExtractMimeType extracts mime type by extension
func ExtractMimeType(file string) string {
	extension := path.Ext(file)
	if len(extension) > 1 {
		extension = extension[1:]
	}

	if mimeType, ok := FileExtensionMimeType[extension]; ok {
		return mimeType
	}
	return "text/plain"
}

//QueryIntValue returns query value for passed in url's name or default value
func QueryIntValue(u *url.URL, name string, defaultValue int) int {
	value := u.Query().Get(name)
	if value == "" {
		return defaultValue
	}

	result := AsInt(value)
	if result != 0 {
		return result
	}
	return defaultValue
}

//QueryBoolValue returns query value for passed in url's name or default value
func QueryBoolValue(u *url.URL, name string, defaultValue bool) bool {
	value := u.Query().Get(name)
	if value == "" {
		return defaultValue
	}
	return AsBoolean(value)
}

//QueryValue returns query value for passed in url's name or default value
func QueryValue(u *url.URL, name, defaultValue string) string {
	value := u.Query().Get(name)
	if value == "" {
		return defaultValue
	}
	return value
}
