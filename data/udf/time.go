package udf

import (
	"fmt"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/data"
	"time"
)

//FormatTime return formatted time, it takes an array of two arguments, the first id time, or now followed by java style time format.
func FormatTime(source interface{}, state data.Map) (interface{}, error) {
	if !toolbox.IsSlice(source) {
		return nil, fmt.Errorf("unable to run FormatTime: expected %T, but had: %T", []interface{}{}, source)
	}
	aSlice := toolbox.AsSlice(source)
	if len(aSlice) < 2 {
		return nil, fmt.Errorf("unable to run FormatTime, expected 2 parameters, but had: %v", len(aSlice))
	}
	var err error
	var timeText = toolbox.AsString(aSlice[0])
	var timeFormat = toolbox.AsString(aSlice[1])
	var timeLayout = toolbox.DateFormatToLayout(timeFormat)
	var timeValue *time.Time
	timeValue, err = toolbox.TimeAt(timeText)
	if err != nil {
		timeValue, err = toolbox.ToTime(aSlice[0], timeLayout)
	}
	if err != nil {
		return nil, err
	}
	if len(aSlice) > 2 {
		timeLocation, err := time.LoadLocation(toolbox.AsString(aSlice[2]))
		if err != nil {
			return nil, err
		}
		timeInLocation := timeValue.In(timeLocation)
		timeValue = &timeInLocation
	}
	return timeValue.Format(timeLayout), nil
}
