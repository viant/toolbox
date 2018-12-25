package data

import (
	"bytes"
	"github.com/viant/toolbox"
	"math"
	"strings"
)

const (
	eofToken     = -1
	invalidToken = iota
	beforeVarToken
	varToken
	incToken
	decrementToken
	shiftToken
	enclosedVarToken
	callToken
	idToken
	arrayIndexToken
	unmatchedToken
	keyIndexToken
	whitespace
	groupingToken
	operatorTojeb
	doubleQuoteEnclosedToken
	comaToken
)

var matchers = map[int]toolbox.Matcher{
	beforeVarToken:           toolbox.NewTerminatorMatcher("$"),
	varToken:                 toolbox.NewCharactersMatcher("$"),
	comaToken:                toolbox.NewTerminatorMatcher(","),
	idToken:                  toolbox.NewCustomIdMatcher("_"),
	incToken:                 toolbox.NewKeywordsMatcher(true, "++"),
	decrementToken:           toolbox.NewKeywordsMatcher(true, "--"),
	shiftToken:               toolbox.NewKeywordsMatcher(true, "<-"),
	arrayIndexToken:          toolbox.NewBodyMatcher("[", "]"),
	callToken:                toolbox.NewBodyMatcher("(", ")"),
	enclosedVarToken:         toolbox.NewBodyMatcher("{", "}"),
	doubleQuoteEnclosedToken: toolbox.NewBodyMatcher(`"`, `"`),
	keyIndexToken:            toolbox.NewCustomIdMatcher("."),
	unmatchedToken:           toolbox.NewRemainingSequenceMatcher(),
	groupingToken:            toolbox.NewBodyMatcher("(", ")"),
	operatorTojeb:            toolbox.NewTerminatorMatcher("+", "-", "*", "/", "^", "%"),
	whitespace:               toolbox.NewCharactersMatcher(" \t\n\r"),
}

//Parse parses expression
func Parse(expression string, handler func(expression string, isUDF bool, argument interface{}) (interface{}, bool)) interface{} {
	tokenizer := toolbox.NewTokenizer(expression, invalidToken, eofToken, matchers)
	var value interface{}
	var result = fragments{}
	var ok bool
	done := false
	for tokenizer.Index < len(expression) && !done {
		match := tokenizer.Nexts(beforeVarToken, varToken, unmatchedToken, eofToken)
		switch match.Token {
		case unmatchedToken:
			result.Append(match.Matched)
			done = true
			continue
		case eofToken:
			break
		case beforeVarToken:
			result.Append(match.Matched)
			continue

		case varToken:
			variable := "$"
			match = tokenizer.Nexts(idToken, enclosedVarToken, incToken, decrementToken, shiftToken)
			switch match.Token {
			case eofToken:
				result.Append(variable)
				continue
			case enclosedVarToken:

				expanded := expandEnclosed(match.Matched, handler)
				if toolbox.IsFloat(expanded) || toolbox.IsInt(expanded) {
					value = expanded
					result.Append(value)
					continue
				}
				expandedText := toolbox.AsString(expanded)
				if strings.HasSuffix(expandedText, ")") {
					value = Parse("$"+expandedText, handler)
					if textValue, ok := value.(string); ok && textValue == "$"+expandedText {
						value = "${" + expandedText + "}"
					}
					result.Append(value)
					continue
				}

				variable := "${" + expandedText + "}"
				if value, ok = handler(variable, false, ""); !ok {
					value = variable
				}
				result.Append(value)
				continue

			case incToken, decrementToken, shiftToken:
				variable += match.Matched
				match = tokenizer.Nexts(idToken) //enclosedVarToken, idToken ?
				if match.Token != idToken {
					result.Append(variable)
					continue
				}
				fallthrough

			case idToken:

				variable += match.Matched
				variable = expandVariable(tokenizer, variable, handler)
				match = tokenizer.Nexts(callToken, incToken, decrementToken, beforeVarToken, unmatchedToken, eofToken)
				switch match.Token {

				case callToken:
					arguments := string(match.Matched[1 : len(match.Matched)-1])
					if value, ok = handler(variable, true, arguments); !ok {
						value = variable + match.Matched
					}
					result.Append(value)
					continue
				case incToken, decrementToken:
					variable += match.Matched
					match.Matched = ""
					fallthrough

				case beforeVarToken, unmatchedToken, eofToken, invalidToken:
					if value, ok = handler(variable, false, ""); !ok {
						value = variable
					}
					result.Append(value)
					result.Append(match.Matched)
					continue
				}

			default:
				result.Append(variable)
			}
		}
	}
	return result.Get()
}

func expandVariable(tokenizer *toolbox.Tokenizer, variable string, handler func(expression string, isUDF bool, argument interface{}) (interface{}, bool)) string {
	match := tokenizer.Nexts(keyIndexToken, arrayIndexToken)
	switch match.Token {
	case keyIndexToken:
		variable = expandSubKey(variable, match, tokenizer, handler)
	case arrayIndexToken:
		variable = expandIndex(variable, match, handler, tokenizer)
	}
	return variable
}

func expandIndex(variable string, match *toolbox.Token, handler func(expression string, isUDF bool, argument interface{}) (interface{}, bool), tokenizer *toolbox.Tokenizer) string {
	variable += toolbox.AsString(Parse(match.Matched, handler))
	match = tokenizer.Nexts(arrayIndexToken, keyIndexToken)
	switch match.Token {
	case keyIndexToken:
		variable = expandSubKey(variable, match, tokenizer, handler)
	case arrayIndexToken:
		variable += toolbox.AsString(Parse(match.Matched, handler))
	}
	return variable
}

func expandSubKey(variable string, match *toolbox.Token, tokenizer *toolbox.Tokenizer, handler func(expression string, isUDF bool, argument interface{}) (interface{}, bool)) string {
	variable += match.Matched
	match = tokenizer.Nexts(idToken, enclosedVarToken, arrayIndexToken)
	switch match.Token {
	case idToken:
		variable += match.Matched
		variable = expandVariable(tokenizer, variable, handler)
	case enclosedVarToken:
		expanded := expandEnclosed(match.Matched, handler)
		variable += toolbox.AsString(expanded)
	case arrayIndexToken:
		variable = expandIndex(variable, match, handler, tokenizer)
	}
	return variable
}

func expandEnclosed(expr string, handler func(expression string, isUDF bool, argument interface{}) (interface{}, bool)) interface{} {
	if strings.HasPrefix(expr, "{") && strings.HasSuffix(expr, "}") {
		expr = string(expr[1 : len(expr)-1])

	}
	tokenizer := toolbox.NewTokenizer(expr, invalidToken, eofToken, matchers)
	match, err := toolbox.ExpectTokenOptionallyFollowedBy(tokenizer, whitespace, "expected operatorTojeb", groupingToken, operatorTojeb)
	if err != nil {
		return Parse(expr, handler)
	}
	switch match.Token {
	case groupingToken:
		groupExpr := string(match.Matched[1 : len(match.Matched)-1])
		result := expandEnclosed(groupExpr, handler)
		if !(toolbox.IsInt(result) || toolbox.IsFloat(result)) {
			return Parse(expr, handler)
		}
		expandedGroup := toolbox.AsString(result) + string(expr[tokenizer.Index:])
		return expandEnclosed(expandedGroup, handler)
	case operatorTojeb:
		leftOperand, leftOk := tryNumericOperand(match.Matched, handler).(float64)
		operator := string(expr[tokenizer.Index : tokenizer.Index+1])
		rightOperand, rightOk := tryNumericOperand(string(expr[tokenizer.Index+1:]), handler).(float64)
		if !leftOk || !rightOk {
			return Parse(expr, handler)
		}
		var floatResult float64
		switch operator {
		case "+":
			floatResult = leftOperand + rightOperand
		case "-":
			floatResult = leftOperand - rightOperand
		case "/":
			if rightOperand == 0 { //division by zero issue
				return Parse(expr, handler)
			}
			floatResult = leftOperand / rightOperand
		case "*":
			floatResult = leftOperand * rightOperand
		case "^":
			floatResult = math.Pow(leftOperand, rightOperand)
		case "%":
			floatResult = float64(int(leftOperand) % int(rightOperand))
		default:
			return Parse(expr, handler)
		}
		intResult := int(floatResult)
		if floatResult == float64(intResult) {
			return intResult
		}
		return floatResult
	}
	return Parse(expr, handler)
}

func tryNumericOperand(expression string, handler func(expression string, isUDF bool, argument interface{}) (interface{}, bool)) interface{} {
	expression = strings.TrimSpace(expression)
	if result, err := toolbox.ToFloat(expression); err == nil {
		return result
	}
	left := expandEnclosed(expression, handler)
	if result, err := toolbox.ToFloat(left); err == nil {
		return result
	}

	left = Parse("$"+expression, handler)
	if result, err := toolbox.ToFloat(left); err == nil {
		return result
	}
	return expression
}

func asExpandedText(source interface{}) string {
	if source != nil && (toolbox.IsSlice(source) || toolbox.IsMap(source)) {
		buf := new(bytes.Buffer)
		err := toolbox.NewJSONEncoderFactory().Create(buf).Encode(source)
		if err == nil {
			return buf.String()
		}
	}
	return toolbox.AsString(source)
}

type fragments []interface{}

func (f *fragments) Append(item interface{}) {
	if text, ok := item.(string); ok {
		if text == "" {
			return
		}
	}
	*f = append(*f, item)
}

func (f fragments) Get() interface{} {
	count := len(f)
	if count == 0 {
		return ""
	}
	var emptyCount = 0
	var result interface{}
	for _, item := range f {
		if text, ok := item.(string); ok && strings.TrimSpace(text) == "" {
			emptyCount++
		} else {
			result = item
		}
	}
	if emptyCount == count-1 {
		return result
	}
	var textResult = ""
	for _, item := range f {
		textResult += asExpandedText(item)
	}
	return textResult
}
