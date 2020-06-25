package udf

import (
	"fmt"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/data"
	"time"
)

//FormatTime return formatted time, it takes an array of arguments, the first is  time express, or now followed by java style time format, optional timezone and truncate format .
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
	if len(aSlice) > 2 && aSlice[2] != "" {
		timeLocation, err := time.LoadLocation(toolbox.AsString(aSlice[2]))
		if err != nil {
			return nil, err
		}
		timeInLocation := timeValue.In(timeLocation)
		timeValue = &timeInLocation
	}

	if len(aSlice) > 3 {
		switch aSlice[3] {
		case "weekday":
			return timeValue.Weekday(), nil
		default:
			truncFromat := toolbox.DateFormatToLayout(toolbox.AsString(aSlice[3]))
			if ts, err := time.Parse(truncFromat, timeValue.Format(truncFromat));err == nil {
				timeValue = &ts
			}
		}
	}

	return timeValue.Format(timeLayout), nil
}

//Elapsed returns elapsed time
func Elapsed(source interface{}, state data.Map) (interface{}, error) {
	inThePast, err := toolbox.ToTime(source, time.RFC3339)
	if err != nil {
		return nil, err
	}
	elapsed := time.Now().Sub(*inThePast).Truncate(time.Second)

	days := elapsed / (24 * time.Hour)
	hours := int(elapsed.Hours()) % 24
	min := int(elapsed.Minutes()) % 60
	sec := int(elapsed.Seconds()) % 60
	result := ""
	if days > 0 {
		result = fmt.Sprintf("%dd", int(days))
	}
	if result == "" && hours > 0 {
		result += fmt.Sprintf("%dh", hours)
	}
	if result == "" && min > 0 {
		result += fmt.Sprintf("%dm", min)
	}
	result += fmt.Sprintf("%ds", sec)
	return result, nil

}
