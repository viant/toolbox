package toolbox

import "fmt"

//Dump prints passed in data as JSON
func Dump(data interface{}) {
	if text, err := AsJSONText(data); err == nil {
		fmt.Printf("%v\n", text)
		return
	}
}

//DumpIndent prints passed in data as indented JSON
func DumpIndent(data interface{}, removeEmptyKeys bool) error {
	if IsMap(data) || IsStruct(data) {
		var aMap = map[string]interface{}{}
		if err := DefaultConverter.AssignConverted(&aMap, data); err != nil {
			return err
		}
		data = aMap
		if removeEmptyKeys {
			data = DeleteEmptyKeys(aMap)
		}
	}

	text, err := AsIndentJSONText(data)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", text)
	return nil
}
