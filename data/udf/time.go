package udf

import (
	"encoding/json"
	"fmt"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/data"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// FormatTime returns formatted time. Arguments are a time expression or value,
// Java-style output format, optional timezone, optional truncation format, and
// optional calendar offset object with integer years, months and days fields.
// Calendar offsets use time.AddDate after timezone conversion and before the
// existing format/truncation handling, including Go's end-of-month normalization.
// In the five-argument form, an empty fourth argument means no truncation.
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
	var timeValue *time.Time
	timeValue, err = toolbox.TimeAt(timeText)
	var timeLayout string
	if err != nil {
		timeFormat := toolbox.AsString(aSlice[1])
		timeLayout = toolbox.DateFormatToLayout(timeFormat)
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
	if len(aSlice) > 4 {
		offset, err := calendarOffset(aSlice[4])
		if err != nil {
			return nil, fmt.Errorf("FormatTime calendar offset: %w", err)
		}
		shifted := timeValue.AddDate(offset.years, offset.months, offset.days)
		timeValue = &shifted
	}

	if len(aSlice) > 1 {
		formatArg := toolbox.AsString(aSlice[1])
		switch strings.Trim(formatArg, `" ,`) {
		case "secondsOfDay":
			return timeValue.Hour()*60*60 + timeValue.Minute()*60 + timeValue.Second(), nil
		}
		if timeLayout == "" {
			timeFormat := toolbox.AsString(aSlice[1])
			timeLayout = toolbox.DateFormatToLayout(timeFormat)
		}
	}

	if len(aSlice) > 3 && !(len(aSlice) > 4 && aSlice[3] == "") {
		switch aSlice[3] {
		case "weekday":
			return timeValue.Weekday(), nil
		default:
			truncFromat := toolbox.DateFormatToLayout(toolbox.AsString(aSlice[3]))
			if ts, err := time.Parse(truncFromat, timeValue.Format(truncFromat)); err == nil {
				timeValue = &ts
			}
		}
	}

	return timeValue.Format(timeLayout), nil
}

type timeCalendarOffset struct {
	years, months, days int
}

func calendarOffset(source interface{}) (timeCalendarOffset, error) {
	result := timeCalendarOffset{}
	if !toolbox.IsMap(source) {
		return result, fmt.Errorf("expected calendar offset object, had %T", source)
	}
	for key, value := range toolbox.AsMap(source) {
		var target *int
		switch key {
		case "years":
			target = &result.years
		case "months":
			target = &result.months
		case "days":
			target = &result.days
		default:
			return result, fmt.Errorf("unknown calendar offset field %q", key)
		}
		var text string
		if number, ok := value.(json.Number); ok {
			text = string(number)
		} else if value != nil {
			v := reflect.ValueOf(value)
			switch v.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				text = strconv.FormatInt(v.Int(), 10)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				text = strconv.FormatUint(v.Uint(), 10)
			case reflect.Float32, reflect.Float64:
				// JSON numbers arrive as float64. Formatting without decimals
				// only after the integrality check avoids silently rounding offsets.
				number := v.Float()
				if number == float64(int(number)) {
					text = strconv.FormatFloat(number, 'f', 0, 64)
				}
			}
		}
		number, err := strconv.ParseInt(text, 10, strconv.IntSize)
		if err != nil {
			return result, fmt.Errorf("%s must be an integer within int range: %w", key, err)
		}
		*target = int(number)
	}
	return result, nil
}

// Elapsed returns elapsed time
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
