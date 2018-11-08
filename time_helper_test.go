package toolbox

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

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
	{//invalid duration unit
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
