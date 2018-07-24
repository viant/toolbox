package toolbox

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

//MacroEvaluator represents a macro expression evaluator, macros has the following format   macro prefix [macro parameter] macro postfix
type MacroEvaluator struct {
	Prefix, Postfix       string
	ValueProviderRegistry ValueProviderRegistry
}

//HasMacro checks if candidate has a macro fragment
func (e *MacroEvaluator) HasMacro(candidate string) bool {
	prefix, postfix := e.Prefix, e.Postfix
	prefixPosition := strings.Index(candidate, prefix)
	if prefixPosition == -1 {
		return false
	}
	postfixPosition := strings.Index(string(candidate[prefixPosition:]), postfix)
	return postfixPosition != -1
}

func (e *MacroEvaluator) expandArguments(context Context, arguments *[]interface{}) error {
	//expanded macros within the macro
	for i, argument := range *arguments {
		if IsString(argument) {
			if argumentAsText, ok := argument.(string); ok {
				if e.HasMacro(argumentAsText) {
					expanded, err := e.Expand(context, argumentAsText)
					if err != nil {
						return fmt.Errorf("failed to expand argument: " + argumentAsText + " due to:\n\t" + err.Error())
					}
					(*arguments)[i] = expanded
				}
			}
		}
	}
	return nil
}

func (e *MacroEvaluator) decodeArguments(context Context, decodedArguments string, macro string) ([]interface{}, error) {
	var arguments = make([]interface{}, 0)
	if len(decodedArguments) > 0 {
		decodedArguments = strings.Replace(decodedArguments, `\"`, `"`, len(decodedArguments))
		decoder := json.NewDecoder(strings.NewReader(decodedArguments))
		err := decoder.Decode(&arguments)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to process macro arguments: " + decodedArguments + " due to:\n\t" + err.Error())
		}
		err = e.expandArguments(context, &arguments)
		if err != nil {
			return nil, err
		}
	}
	return arguments, nil
}

func (e *MacroEvaluator) extractMacro(input string) (success bool, macro, macroName, macroArguments string) {
	prefix, postfix := e.Prefix, e.Postfix
	var isInQuotes, argumentCount, previousChar, expectArguements, argumentStartPosition, argumentEndPosition = false, 0, "", false, 0, 0
	prefixPosition := strings.Index(input, prefix)
	if prefixPosition == -1 {
		return false, "", "", ""
	}
	for i := prefixPosition + len(prefix); i < len(input); i++ {
		aChar := input[i : i+1]
		if i > 0 {
			previousChar = input[i-1 : i]
		}

		if strings.ContainsAny(aChar, " \b\n[") {
			expectArguements = true
		}
		if aChar == "\"" && previousChar != "\\" {
			isInQuotes = !isInQuotes
		}
		if !isInQuotes && aChar == "[" && previousChar != "\\" {
			if argumentCount == 0 {
				argumentStartPosition = i
			}
			argumentCount++
		}
		if !isInQuotes && aChar == "]" && previousChar != "\\" {
			argumentEndPosition = i
			argumentCount--
		}
		macro = macro + aChar
		if argumentCount == 0 {
			if aChar == postfix {
				break
			}
			if !expectArguements {
				macroName = macroName + aChar
			}
		}
	}
	if argumentStartPosition > 0 && argumentStartPosition < argumentEndPosition {
		macroArguments = input[argumentStartPosition : argumentEndPosition+1]
	}

	return true, prefix + macro, macroName, macroArguments
}

//Expand expands passed in input, it returns expanded value of any type or error
func (e *MacroEvaluator) Expand(context Context, input string) (interface{}, error) {
	success, macro, macroName, macroArguments := e.extractMacro(input)
	if !success {
		return input, nil
	}
	valueProviderRegistry := e.ValueProviderRegistry
	if !valueProviderRegistry.Contains(macroName) {
		return nil, fmt.Errorf("failed to lookup macro: '%v' while processing: %v", macroName, input)
	}
	arguments, err := e.decodeArguments(context, macroArguments, macro)
	if err != nil {
		return nil, fmt.Errorf("failed expand macro: %v due to %v", macro, err.Error())
	}
	valueProvider := valueProviderRegistry.Get(macroName)
	value, err := valueProvider.Get(context, arguments...)
	if err != nil {
		return nil, err
	}
	if len(macro) == len(input) {
		return value, nil
	}
	expandedMacro := fmt.Sprintf("%v", value)
	result := strings.Replace(input, macro, expandedMacro, 1)
	if e.HasMacro(result) {
		return e.Expand(context, result)
	}
	return result, nil
}

//NewMacroEvaluator returns a new macro evaluator
func NewMacroEvaluator(prefix, postfix string, registry ValueProviderRegistry) *MacroEvaluator {
	return &MacroEvaluator{
		Prefix:                prefix,
		Postfix:               postfix,
		ValueProviderRegistry: registry,
	}
}

//ExpandParameters expands passed in parameters as strings
func ExpandParameters(macroEvaluator *MacroEvaluator, parameters map[string]string) error {
	for key := range parameters {
		value := parameters[key]
		if macroEvaluator.HasMacro(value) {
			textValue, err := macroEvaluator.Expand(nil, AsString(value))
			if err != nil {
				return err
			}
			parameters[key] = AsString(textValue)
		}
	}
	return nil
}

//ExpandValue expands passed in value, it returns expanded string value or error
func ExpandValue(macroEvaluator *MacroEvaluator, value string) (string, error) {
	if macroEvaluator.HasMacro(value) {
		expanded, err := macroEvaluator.Expand(nil, value)
		if err != nil {
			return "", err
		}
		return AsString(expanded), nil
	}
	return value, nil
}
