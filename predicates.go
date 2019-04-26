package toolbox

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

//TrueProvider represents a true provider
var TrueProvider = func(input interface{}) bool {
	return true
}

type withinSecPredicate struct {
	baseTime        time.Time
	deltaInSeconds  int
	dateLayout      string
	actual          string
	elapsed         time.Duration
	maxAllowedDelay time.Duration
}

func (p *withinSecPredicate) String() string {
	return fmt.Sprintf("(elapsed: %d, max allowed delay: %d)\n", int(p.elapsed), int(p.maxAllowedDelay))
}

//Apply returns true if passed in time is within deltaInSeconds from baseTime
func (p *withinSecPredicate) Apply(value interface{}) bool {
	timeValue, err := ToTime(value, p.dateLayout)
	if err != nil {
		return false
	}
	elapsed := timeValue.Sub(p.baseTime)
	if elapsed < 0 {
		elapsed *= -1
	}
	var maxAllowedDelay = time.Duration(p.deltaInSeconds) * time.Second
	var passed = maxAllowedDelay >= elapsed
	if !passed {
		p.elapsed = elapsed
		p.maxAllowedDelay = maxAllowedDelay
	}
	return passed
}

func (p *withinSecPredicate) ToString() string {
	return fmt.Sprintf(" %v within %v s", p.baseTime, p.deltaInSeconds)
}

//NewWithinPredicate returns new NewWithinPredicate predicate, it takes base time, delta in second, and dateLayout
func NewWithinPredicate(baseTime time.Time, deltaInSeconds int, dateLayout string) Predicate {
	return &withinSecPredicate{
		baseTime:       baseTime,
		deltaInSeconds: deltaInSeconds,
		dateLayout:     dateLayout,
	}
}

type betweenPredicate struct {
	from float64
	to   float64
}

func (p *betweenPredicate) Apply(value interface{}) bool {
	floatValue := AsFloat(value)
	return floatValue >= p.from && floatValue <= p.to
}

func (p *betweenPredicate) String() string {
	return fmt.Sprintf("x BETWEEN %v AND %v", p.from, p.to)
}

//NewBetweenPredicate creates a new BETWEEN predicate, it takes from, and to.
func NewBetweenPredicate(from, to interface{}) Predicate {
	return &betweenPredicate{
		from: AsFloat(from),
		to:   AsFloat(to),
	}
}

type inPredicate struct {
	predicate Predicate
}

func (p *inPredicate) Apply(value interface{}) bool {
	return p.predicate.Apply(value)
}

//NewInPredicate creates a new IN predicate
func NewInPredicate(values ...interface{}) Predicate {
	converted, kind := DiscoverCollectionValuesAndKind(values)
	switch kind {
	case reflect.Int:
		predicate := inIntPredicate{values: make(map[int]bool)}
		SliceToMap(converted, predicate.values, func(item interface{}) int {
			return AsInt(item)
		}, TrueProvider)
		return &predicate
	case reflect.Float64:
		predicate := inFloatPredicate{values: make(map[float64]bool)}
		SliceToMap(converted, predicate.values, func(item interface{}) float64 {
			return AsFloat(item)
		}, TrueProvider)
		return &predicate
	default:
		predicate := inStringPredicate{values: make(map[string]bool)}
		SliceToMap(converted, predicate.values, func(item interface{}) string {
			return AsString(item)
		}, TrueProvider)
		return &predicate
	}
}

type inFloatPredicate struct {
	values map[float64]bool
}

func (p *inFloatPredicate) Apply(value interface{}) bool {
	candidate := AsFloat(value)
	return p.values[candidate]
}

type inIntPredicate struct {
	values map[int]bool
}

func (p *inIntPredicate) Apply(value interface{}) bool {
	candidate := AsInt(value)
	return p.values[int(candidate)]
}

type inStringPredicate struct {
	values map[string]bool
}

func (p *inStringPredicate) Apply(value interface{}) bool {
	candidate := AsString(value)
	return p.values[candidate]
}

type numericComparablePredicate struct {
	rightOperand float64
	operator     string
}

func (p *numericComparablePredicate) Apply(value interface{}) bool {
	leftOperand := AsFloat(value)
	switch p.operator {
	case ">":
		return leftOperand > p.rightOperand
	case ">=":
		return leftOperand >= p.rightOperand
	case "<":
		return leftOperand < p.rightOperand
	case "<=":
		return leftOperand <= p.rightOperand
	case "=":
		return leftOperand == p.rightOperand
	case "!=":
		return leftOperand != p.rightOperand
	}
	return false
}

type stringComparablePredicate struct {
	rightOperand string
	operator     string
}

func (p *stringComparablePredicate) Apply(value interface{}) bool {
	leftOperand := AsString(value)

	switch p.operator {
	case "=":
		return leftOperand == p.rightOperand
	case "!=":
		return leftOperand != p.rightOperand
	}
	return false
}

//NewComparablePredicate create a new comparable predicate for =, !=, >=, <=
func NewComparablePredicate(operator string, leftOperand interface{}) Predicate {
	if CanConvertToFloat(leftOperand) {
		return &numericComparablePredicate{AsFloat(leftOperand), operator}
	}
	return &stringComparablePredicate{AsString(leftOperand), operator}
}

type nilPredicate struct{}

func (p *nilPredicate) Apply(value interface{}) bool {
	return value == nil || reflect.ValueOf(value).IsNil()
}

//NewNilPredicate returns a new nil predicate
func NewNilPredicate() Predicate {
	return &nilPredicate{}
}

type likePredicate struct {
	matchingFragments []string
}

func (p *likePredicate) Apply(value interface{}) bool {
	textValue := strings.ToLower(AsString(value))
	for _, matchingFragment := range p.matchingFragments {
		matchingIndex := strings.Index(textValue, matchingFragment)
		if matchingIndex == -1 {
			return false
		}
		if matchingIndex < len(textValue) {
			textValue = textValue[matchingIndex:]
		}
	}
	return true
}

//NewLikePredicate create a new like predicate
func NewLikePredicate(matching string) Predicate {
	return &likePredicate{matchingFragments: strings.Split(strings.ToLower(matching), "%")}
}
