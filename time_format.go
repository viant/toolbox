package toolbox

import (
	"strings"
	"time"
)

//DateFormatKeyword constant 'dateFormat' key
var DateFormatKeyword = "dateFormat"

//DateLayoutKeyword constant 'dateLayout' key
var DateLayoutKeyword = "dateLayout"

//DateFormatToLayout converts java date format https://docs.oracle.com/javase/6/docs/api/java/text/SimpleDateFormat.html#rfc822timezone into go date layout
func DateFormatToLayout(dateFormat string) string {

	dateFormat = strings.Replace(dateFormat, "ddd", "_2", 1)
	dateFormat = strings.Replace(dateFormat, "dd", "02", 1)
	dateFormat = strings.Replace(dateFormat, "d", "2", 1)

	dateFormat = strings.Replace(dateFormat, "HH", "15", 1)

	dateFormat = strings.Replace(dateFormat, "hh", "03", 1)
	dateFormat = strings.Replace(dateFormat, "h", "3", 1)

	dateFormat = strings.Replace(dateFormat, "mm", "04", 1)
	dateFormat = strings.Replace(dateFormat, "m", "4", 1)

	dateFormat = strings.Replace(dateFormat, "ss", "05", 1)
	dateFormat = strings.Replace(dateFormat, "s", "5", 1)

	dateFormat = strings.Replace(dateFormat, "yyyy", "2006", 1)
	dateFormat = strings.Replace(dateFormat, "yy", "06", 1)
	dateFormat = strings.Replace(dateFormat, "y", "06", 1)

	dateFormat = strings.Replace(dateFormat, "SSS", "000", 1)

	dateFormat = strings.Replace(dateFormat, "a", "pm", 1)
	dateFormat = strings.Replace(dateFormat, "aa", "PM", 1)

	dateFormat = strings.Replace(dateFormat, "MMMM", "January", 1)
	dateFormat = strings.Replace(dateFormat, "MMM", "Jan", 1)
	dateFormat = strings.Replace(dateFormat, "MM", "01", 1)
	dateFormat = strings.Replace(dateFormat, "M", "1", 1)

	dateFormat = strings.Replace(dateFormat, "ZZ", "-0700", 1)

	dateFormat = strings.Replace(dateFormat, "Z", "-07", 1)

	dateFormat = strings.Replace(dateFormat, "zz:zz", "Z07:00", 1)
	dateFormat = strings.Replace(dateFormat, "zzzz", "Z0700", 1)
	dateFormat = strings.Replace(dateFormat, "z", "MST", 1)

	dateFormat = strings.Replace(dateFormat, "EEEE", "Monday", 1)
	dateFormat = strings.Replace(dateFormat, "E", "Mon", 1)

	return dateFormat
}

//GetTimeLayout returns time laout from passed in map, first it check if DateLayoutKeyword is defined is so it returns it, otherwise it check DateFormatKeyword and if exists converts it to  dateLayout
//If neithers keys exists it panics, please use HasTimeLayout to avoid panic
func GetTimeLayout(input interface{}) string {
	switch settings := input.(type) {
	case map[string]string:
		if value, found := settings[DateLayoutKeyword]; found {
			return value
		}

		if value, found := settings[DateFormatKeyword]; found {
			return DateFormatToLayout(value)
		}

	case map[string]interface{}:
		if value, found := settings[DateLayoutKeyword]; found {
			return AsString(value)
		}

		if value, found := settings[DateFormatKeyword]; found {
			return DateFormatToLayout(AsString(value))
		}

	}
	return ""
}

//HasTimeLayout checks if dateLayout can be taken from the passed in setting map
func HasTimeLayout(input interface{}) bool {
	switch settings := input.(type) {
	case map[string]string:
		if _, found := settings[DateLayoutKeyword]; found {
			return true
		}
		if _, found := settings[DateFormatKeyword]; found {
			return true
		}

	case map[string]interface{}:
		if _, found := settings[DateLayoutKeyword]; found {
			return true
		}
		if _, found := settings[DateFormatKeyword]; found {
			return true
		}

	}
	return false
}

//TimestampToString formats timestamp to passed in java style date format
func TimestampToString(dateFormat string, unixTimestamp, unixNanoTimestamp int64) string {
	t := time.Unix(unixTimestamp, unixNanoTimestamp)
	dateLayout := DateFormatToLayout(dateFormat)
	return t.Format(dateLayout)
}
