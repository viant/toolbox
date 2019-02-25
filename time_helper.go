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

//Duration represents duration
type Duration struct {
	Value int
	Unit  string
}

//Duration return durations
func (d Duration) Duration() (time.Duration, error) {
	return NewDuration(d.Value, d.Unit)
}

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

//TimeWindow represents a time window
type TimeWindow struct {
	Loopback   *Duration
	StartDate  string
	startTime  *time.Time
	EndDate    string
	endTime    *time.Time
	TimeLayout string
	TimeFormat string
	Interval   *Duration
}

//Range iterates with interval step between start and window end.
func (w *TimeWindow) Range(handler func(time time.Time) (bool, error)) error {
	start, err := w.StartTime()
	if err != nil {
		return err
	}

	end, err := w.EndTime()
	if err != nil {
		return err
	}
	if w.Interval == nil && w.Loopback != nil {
		w.Interval = w.Loopback
	}

	if w.Interval == nil {
		_, err = handler(*end)
		return err
	}
	interval, err := w.Interval.Duration()
	if err != nil {
		return err
	}
	for ts := *start; ts.Before(*end) || ts.Equal(*end); ts = ts.Add(interval) {
		if ok, err := handler(ts); err != nil || !ok {
			return err
		}
	}
	return err
}

//Layout return time layout
func (w *TimeWindow) Layout() string {
	if w.TimeLayout != "" {
		return w.TimeLayout
	}
	if w.TimeFormat != "" {
		w.TimeLayout = DateFormatToLayout(w.TimeFormat)
	}
	if w.TimeLayout == "" {
		w.TimeLayout = time.RFC3339
	}
	return w.TimeLayout
}

//StartTime returns time window start time
func (w *TimeWindow) StartTime() (*time.Time, error) {
	if w.StartDate != "" {
		if w.startTime != nil {
			return w.startTime, nil
		}
		timeLayout := w.Layout()
		startTime, err := time.Parse(timeLayout, w.StartDate)
		if err != nil {
			return nil, err
		}
		w.startTime = &startTime
		return w.startTime, nil
	}
	endDate, err := w.EndTime()
	if err != nil {
		return nil, err
	}
	if w.Loopback == nil || w.Loopback.Value == 0 {
		return endDate, nil
	}
	loopback, err := w.Loopback.Duration()
	if err != nil {
		return nil, err
	}
	startTime := endDate.Add(-loopback)
	return &startTime, nil
}

//EndTime returns time window end time
func (w *TimeWindow) EndTime() (*time.Time, error) {
	if w.EndDate != "" {
		if w.endTime != nil {
			return w.endTime, nil
		}
		timeLayout := w.Layout()
		endTime, err := time.Parse(timeLayout, w.EndDate)
		if err != nil {
			return nil, err
		}
		w.endTime = &endTime
		return w.endTime, nil
	}
	now := time.Now()
	return &now, nil
}
