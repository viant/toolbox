package toolbox

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"time"
)

func TestAtTime_Next(t *testing.T) {

	timeLayout := "2006-01-02 15:04:05"

	var useCases = []struct {
		description string
		at          *AtTime
		baseTime    string
		expectTime  string
	}{

		{
			description: "evey 1/2 - next day",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "*",
				Minute:  "30",
			},
			baseTime:   "2019-01-01 23:33:01",
			expectTime: "2019-01-02 01:30:00",
		},

		{
			description: "evey 1/2, last minute - next day",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "*",
				Minute:  "0,30",
			},
			baseTime:   "2019-01-01 23:59:00",
			expectTime: "2019-01-02 01:00:00",
		},

		{
			description: "evey 1/2, last day of the month - next day",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "*",
				Minute:  "0,30",
			},
			baseTime:   "2019-05-31 23:59:00",
			expectTime: "2019-06-01 01:00:00",
		},

		{
			description: "evey hour - next day",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "0",
				Minute:  "*",
			},
			baseTime:   "2019-01-01 23:01:01",
			expectTime: "2019-01-02 00:02:00",
		},

		{
			description: "evey minute",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "*",
				Minute:  "*",
			},
			baseTime:   "2019-01-01 01:01:01",
			expectTime: "2019-01-01 01:02:00",
		},

		{
			description: "evey 30 minute before",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "*",
				Minute:  "30",
			},
			baseTime:   "2019-01-01 01:01:01",
			expectTime: "2019-01-01 01:30:00",
		},
		{
			description: "evey 30 minute after",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "*",
				Minute:  "30",
			},
			baseTime:   "2019-01-01 01:31:01",
			expectTime: "2019-01-01 02:30:00",
		},
		{
			description: "evey 10, 30 minute, before",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "*",
				Minute:  "10,30",
			},
			baseTime:   "2019-01-01 01:09:01",
			expectTime: "2019-01-01 01:10:00",
		},
		{
			description: "evey 10, 30 minute, after first",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "*",
				Minute:  "10,30",
			},
			baseTime:   "2019-01-01 01:13:01",
			expectTime: "2019-01-01 01:30:00",
		},
		{
			description: "evey 10, 30 minute, after second",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "*",
					Minute:  "10,30",
			},
			baseTime:   "2019-01-01 01:33:01",
			expectTime: "2019-01-01 02:10:00",
		},

		{
			description: "evey *:0 minute",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "*",
				Minute:  "0",
			},
			baseTime:   "2019-01-01 01:59:01",
			expectTime: "2019-01-01 02:00:00",
		},

		{
			description: "evey hour",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "*",
				Minute:  "",
			},
			baseTime:   "2019-01-01 01:33:01",
			expectTime: "2019-01-01 02:00:00",
		},
		{
			description: "at 13 hour, before",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "13",
				Minute:  "",
			},
			baseTime:   "2019-01-01 01:33:01",
			expectTime: "2019-01-01 13:00:00",
		},
		{
			description: "at 13 hour, after",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "13",
				Minute:  "",
			},
			baseTime:   "2019-01-01 15:33:01",
			expectTime: "2019-01-02 13:00:00",
		},
		{
			description: "at midnight",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "0",
				Minute:  "",
			},
			baseTime:   "2019-01-01 23:33:01",
			expectTime: "2019-01-02 00:00:00",
		},

		{
			description: "at midnight weekday 9",
			at: &AtTime{
				WeekDay: "*",
				Hour:    "0",
				Minute:  "",
			},
			baseTime:   "2019-01-06 23:33:01",
			expectTime: "2019-01-07 00:00:00",
		},

		{
			description: "every 0 weekday",
			at: &AtTime{
				WeekDay: "0",
				Hour:    "",
				Minute:  "",
			},
			baseTime:   "2019-01-04 23:33:01",
			expectTime: "2019-01-06 00:00:00",
		},

		{
			description: "every 2 weekday",
			at: &AtTime{
				WeekDay: "2",
				Hour:    "",
				Minute:  "",
			},
			baseTime:   "2019-01-04 23:33:01",
			expectTime: "2019-01-08 00:00:00",
		},

		{
			description: "every 2nd weekday - overlaps with base time",
			at: &AtTime{
				WeekDay: "2",
				Hour:    "",
				Minute:  "",
			},
			baseTime:   "2019-01-08 23:33:01",
			expectTime: "2019-01-15 00:00:00",
		},
		{
			description: "every 5 weekday in the future",
			at: &AtTime{
				WeekDay: "2,5",
				Hour:    "",
				Minute:  "",
			},
			baseTime:   "2019-01-09 23:33:01",
			expectTime: "2019-01-11 00:00:00",
		},
		{
			description: "every 5 weekday in the future tz",
			at: &AtTime{
				WeekDay: "2,5",
				Hour:    "",
				Minute:  "",
				TZ:      "America/Los_Angeles",
			},
			baseTime:   "2019-01-09 23:33:01",
			expectTime: "2019-01-11 00:00:00",
		},
	}

	for _, useCase := range useCases {

		err := useCase.at.Init()
		assert.Nil(t, err)

		var loc *time.Location
		if useCase.at.TZ != "" {
			loc, _ = time.LoadLocation(useCase.at.TZ)
		}
		var baseTime time.Time
		if loc != nil {
			baseTime, err = time.ParseInLocation(timeLayout, useCase.baseTime, loc)
			assert.Nil(t, err, useCase.description)
		} else {
			baseTime, err = time.Parse(timeLayout, useCase.baseTime)
			assert.Nil(t, err, useCase.description)

		}

		var expectTime time.Time
		if loc != nil {
			expectTime, err = time.ParseInLocation(timeLayout, useCase.expectTime, loc)
			assert.Nil(t, err, useCase.description)
		} else {
			expectTime, err = time.Parse(timeLayout, useCase.expectTime)
			assert.Nil(t, err, useCase.description)

		}
		actualTime := useCase.at.Next(baseTime)
		assert.Equal(t, expectTime, actualTime, useCase.description)

		//without tz
		baseTime, err = time.Parse(timeLayout, useCase.baseTime)
		actualTime = useCase.at.Next(baseTime)
		assert.Equal(t, expectTime, actualTime, useCase.description)

	}

}

func TestNewDuration(t *testing.T) {

	var useCases = []struct {
		description string
		value       int
		unit        string
		expected    time.Duration
		hasError    bool
	}{
		{
			description: "sec test",
			value:       3,
			unit:        DurationSecond,
			expected:    3 * time.Second,
		},
		{
			description: "min test",
			value:       4,
			unit:        DurationMinute,
			expected:    4 * time.Minute,
		},
		{
			description: "hour test",
			value:       5,
			unit:        DurationHour,
			expected:    5 * time.Hour,
		},
		{
			description: "day test",
			value:       12,
			unit:        DurationDay,
			expected:    12 * time.Hour * 24,
		},
		{
			description: "week test",
			value:       7,
			unit:        DurationWeek,
			expected:    7 * time.Hour * 24 * 7,
		},
		{
			description: "error test",
			value:       4,
			unit:        "abc",
			hasError:    true,
		},
	}

	for _, useCase := range useCases {
		actual, err := NewDuration(useCase.value, useCase.unit)
		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		} else if err != nil {
			assert.Nil(t, err, useCase.description)
			continue
		}
		assert.Equal(t, useCase.expected, actual, useCase.description)

	}

}

func TestIdMatcher_Match(t *testing.T) {
	{
		ts, err := TimeAt("1 sec ahead")
		assert.Nil(t, err)
		assert.EqualValues(t, ts.Unix()-1, time.Now().Unix())
	}
	{ //invalid duration unit
		_, err := TimeAt("1 d ahead")
		assert.NotNil(t, err)
	}
}

func TestTimeDiff(t *testing.T) {
	var useCases = []struct {
		description string
		base        time.Time
		expression  string
		exectedDiff time.Duration
		hasError    bool
	}{
		{
			description: "now test",
			expression:  "now",
			base:        time.Now(),
			exectedDiff: 0,
		},
		{
			description: "tomorrow test",
			expression:  "tomorrow",
			base:        time.Now(),
			exectedDiff: time.Hour * 24,
		},
		{
			description: "yesterday test",
			expression:  "yesterday",
			base:        time.Now(),
			exectedDiff: -time.Hour * 24,
		},
		{
			description: "empty expr error",
			hasError:    true,
		},
		{
			description: "parsing expr error",
			expression:  "a232",
			hasError:    true,
		},
		{
			description: "2 days ago test",
			expression:  "2daysago",
			base:        time.Now(),
			exectedDiff: -time.Hour * 48,
		},

		{
			description: "2 days in the future",
			expression:  "2day in the future",
			base:        time.Now(),
			exectedDiff: time.Hour * 48,
		},

		{
			description: "days in the future",
			expression:  "day InTheFuture",
			base:        time.Now(),
			exectedDiff: time.Hour * 24,
		},

		{
			description: "2 hours before",
			expression:  "2hourbefore",
			base:        time.Now(),
			exectedDiff: -time.Hour * 2,
		},
		{
			description: "2 hours later",
			expression:  "2 hoursLater",
			base:        time.Now(),
			exectedDiff: time.Hour * 2,
		},
		{
			description: "timezone",
			expression:  "nowInUTC",
			base:        time.Now(),
			exectedDiff: 0,
		},

		{
			description: "invalid timezone error",
			expression:  "nowInBAAA",
			base:        time.Now(),
			hasError:    true,
		},

		{
			description: "day in UTC",
			expression:  "2 days ago in UTC",
			base:        time.Now(),
			exectedDiff: -time.Hour * 48,
		},

		{
			description: "day in UTC",
			expression:  "weekAheadInUTC",
			base:        time.Now(),
			exectedDiff: time.Hour * 24 * 7,
		},
	}

	for _, useCase := range useCases {
		actual, err := TimeDiff(useCase.base, useCase.expression)
		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		} else if err != nil {
			assert.Nil(t, err, useCase.description)
			continue
		}
		expected := useCase.base.Add(useCase.exectedDiff)
		assert.EqualValues(t, expected.Unix(), actual.Unix(), useCase.description)

	}

}

func TestDayElapsedInPct(t *testing.T) {

	t0, _ := time.Parse(DateFormatToLayout("yyyy-MM-dd hh:mm:ss"), "2017-01-01 12:00:00")
	elapsedPct := ElapsedDay(t0)
	assert.EqualValues(t, 50, math.Round(100*elapsedPct))

	elapsed, err := ElapsedToday("")
	assert.Nil(t, err)
	assert.True(t, elapsed > 0)

	remaining, err := RemainingToday("")
	assert.Nil(t, err)
	assert.True(t, remaining > 0)
	assert.EqualValues(t, int(remaining+elapsed), 1)

}

func TestTimeWindow_Range(t *testing.T) {

	var useCaes = []struct {
		description   string
		window        *TimeWindow
		expectedCount int
	}{
		{
			description:   "empty window",
			window:        &TimeWindow{},
			expectedCount: 1,
		},
		{
			description: "loopback window",
			window: &TimeWindow{
				TimeFormat: "yyyy-MM-dd HH:mm:ss",
				Loopback:   &Duration{Value: 3, Unit: "sec"},
				EndDate:    "2011-12-01 15:01:01",
				Interval:   &Duration{Value: 1, Unit: "sec"},
			},
			expectedCount: 4,
		},
		{
			description: "default loopback with interval window",
			window: &TimeWindow{
				Loopback: &Duration{Value: 3, Unit: "min"},
				Interval: &Duration{Value: 1, Unit: "min"},
			},
			expectedCount: 4,
		},
		{
			description: "default loopback window",
			window: &TimeWindow{
				Loopback: &Duration{Value: 3, Unit: "min"},
			},
			expectedCount: 2,
		},
		{
			description: "date range window",
			window: &TimeWindow{
				TimeFormat: "yyyy-MM-dd HH:mm:ss",
				StartDate:  "2011-12-01 15:01:01",
				EndDate:    "2011-12-01 15:02:01",
				Interval:   &Duration{Value: 10, Unit: "sec"}},
			expectedCount: 7,
		},
	}

	for _, useCase := range useCaes {
		count := 0
		err := useCase.window.Range(func(time time.Time) (bool, error) {
			count++
			return true, nil
		})
		assert.Nil(t, err, useCase.description)
		assert.Equal(t, useCase.expectedCount, count, useCase.description)
	}

}


func TestN(t *testing.T) {


}