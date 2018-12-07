package toolbox

import (
	"fmt"
	"strings"
	"time"
)

const (
	Now       = "now"
	Tomorrow  = "tomorrow"
	Yesterday = "yesterday"

	//TimeAtTwoHoursAgo   = "2hoursAgo"
	//TimeAtHourAhead     = "hourAhead"
	//TimeAtTwoHoursAhead = "2hoursAhead"

	DurationWeek            = "week"
	DurationDay             = "day"
	DurationHour            = "hour"
	DurationMinute          = "minute"
	DurationMinuteAbbr      = "min"
	DurationSecond          = "second"
	DurationSecondAbbr      = "sec"
	DurationMillisecond     = "millisecond"
	DurationMillisecondAbbr = "ms"
	DurationMicrosecond     = "microsecond"
	DurationMicrosecondAbbr = "us"
	DurationNanosecond      = "nanosecond"
	DurationNanosecondAbbr  = "ns"
)

//NewDuration returns a durationToken for supplied value and time unit, 3, "second"
func NewDuration(value int, unit string) (time.Duration, error) {
	var duration time.Duration
	switch unit {
	case DurationWeek:
		duration = time.Hour * 24 * 7
	case DurationDay:
		duration = time.Hour * 24
	case DurationHour:
		duration = time.Hour
	case DurationMinute, DurationMinuteAbbr:
		duration = time.Minute
	case DurationSecond, DurationSecondAbbr:
		duration = time.Second
	case DurationMillisecond, DurationMillisecondAbbr:
		duration = time.Millisecond
	case DurationMicrosecond, DurationMicrosecondAbbr:
		duration = time.Microsecond
	case DurationNanosecond, DurationNanosecondAbbr:
		duration = time.Nanosecond
	default:
		return 0, fmt.Errorf("unsupported unit: %v", unit)
	}
	return time.Duration(value) * duration, nil
}

const (
	eofToken     = -1
	invalidToken = iota
	timeValueToken
	nowToken
	yesterdayToken
	tomorrowToken
	whitespacesToken
	durationToken
	inTimezoneToken
	durationPluralToken
	positiveModifierToken
	negativeModifierToken
	timezoneToken
)

var timeAtExpressionMatchers = map[int]Matcher{
	timeValueToken:        NewIntMatcher(),
	whitespacesToken:      CharactersMatcher{" \n\t"},
	durationToken:         NewKeywordsMatcher(false, DurationWeek, DurationDay, DurationHour, DurationMinute, DurationMinuteAbbr, DurationSecond, DurationSecondAbbr, DurationMillisecond, DurationMillisecondAbbr, DurationMicrosecond, DurationMicrosecondAbbr, DurationNanosecond, DurationNanosecondAbbr),
	durationPluralToken:   NewKeywordsMatcher(false, "s"),
	nowToken:              NewKeywordsMatcher(false, Now),
	yesterdayToken:        NewKeywordsMatcher(false, Yesterday),
	tomorrowToken:         NewKeywordsMatcher(false, Tomorrow),
	positiveModifierToken: NewKeywordsMatcher(false, "onward", "ahead", "after", "later", "in the future", "inthefuture"),
	negativeModifierToken: NewKeywordsMatcher(false, "past", "ago", "before", "earlier", "in the past", "inthepast"),
	inTimezoneToken:       NewKeywordsMatcher(false, "in"),
	timezoneToken:         NewRemainingSequenceMatcher(),
	eofToken:              &EOFMatcher{},
}

//TimeAt returns time of now supplied offsetExpression, this function uses TimeDiff
func TimeAt(offsetExpression string) (*time.Time, error) {
	return TimeDiff(time.Now(), offsetExpression)
}

//TimeDiff returns time for supplied base time and literal, the supported literal now, yesterday, tomorrow, or the following template:
// 	- [timeValueToken]  durationToken past_or_future_modifier [IN tz]
// where time modifier can be any of the following:  "onward", "ahead", "after", "later", or "past", "ago", "before", "earlier", "in the future", "in the past") )
func TimeDiff(base time.Time, expression string) (*time.Time, error) {
	if expression == "" {
		return nil, fmt.Errorf("expression was empty")
	}
	var delta time.Duration
	var isNegative = false

	tokenizer := NewTokenizer(expression, invalidToken, eofToken, timeAtExpressionMatchers)
	var val = 1
	var isTimeExtracted = false
	token, err := ExpectToken(tokenizer, "", timeValueToken, nowToken, yesterdayToken, tomorrowToken)
	if err == nil {
		switch token.Token {
		case timeValueToken:
			val, _ = ToInt(token.Matched)
		case yesterdayToken:
			isNegative = true
			fallthrough
		case tomorrowToken:
			delta, _ = NewDuration(1, DurationDay)
			fallthrough
		case nowToken:
			isTimeExtracted = true
		}
	}

	if !isTimeExtracted {
		token, err = ExpectTokenOptionallyFollowedBy(tokenizer, whitespacesToken, "expected time unit", durationToken)
		if err != nil {
			return nil, err
		}
		delta, _ = NewDuration(val, strings.ToLower(token.Matched))
		ExpectToken(tokenizer, "", durationPluralToken)
		token, err = ExpectTokenOptionallyFollowedBy(tokenizer, whitespacesToken, "expected time modifier", positiveModifierToken, negativeModifierToken)
		if err != nil {
			return nil, err
		}
		if token.Token == negativeModifierToken {
			isNegative = true
		}
	}

	if token, err = ExpectTokenOptionallyFollowedBy(tokenizer, whitespacesToken, "expected in", inTimezoneToken); err == nil {
		token, err = ExpectToken(tokenizer, "epected timezone", timezoneToken)
		if err != nil {
			return nil, err
		}
		tz := strings.TrimSpace(token.Matched)
		tzLocation, err := time.LoadLocation(tz)
		if err != nil {
			return nil, fmt.Errorf("failed to load timezone tzLocation: %v, %v", tz, err)
		}
		base = base.In(tzLocation)
	}
	token, err = ExpectToken(tokenizer, "expected eofToken", eofToken)
	if err != nil {
		return nil, err
	}
	if isNegative {
		delta *= -1
	}
	base = base.Add(delta)
	return &base, nil
}

//ElapsedToday returns elapsed today time percent, it takes optionally timezone
func ElapsedToday(tz string) (float64, error) {
	if tz != "" {
		tz = "In" + tz
	}
	now, err := TimeAt("now" + tz)
	if err != nil {
		return 0, err
	}
	return ElapsedDay(*now), nil
}

//ElapsedDay returns elapsed pct for passed in day (second elapsed that day over 24 hours)
func ElapsedDay(dateTime time.Time) float64 {
	elapsedToday := time.Duration(dateTime.Hour())*time.Hour + time.Duration(dateTime.Minute())*time.Minute + time.Duration(dateTime.Second()) + time.Second
	elapsedTodayPct := float64(elapsedToday) / float64((24 * time.Hour))
	return elapsedTodayPct
}

//RemainingToday returns remaining today time percent, it takes optionally timezone
func RemainingToday(tz string) (float64, error) {
	elapsedToday, err := ElapsedToday(tz)
	if err != nil {
		return 0, err
	}
	return 1.0 - elapsedToday, nil
}
