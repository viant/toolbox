package toolbox

import "fmt"

//Dump prints passed in data as JSON
func Dump(data interface{}) {
	if text, err := AsJSONText(data);err ==nil {
		fmt.Printf("%v\n", text)
		return
	}
}
