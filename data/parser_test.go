package data

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"sort"
	"strings"
	"testing"
)

func IndexOf(source interface{}, state Map) (interface{}, error) {
	if !toolbox.IsSlice(source) {
		return nil, fmt.Errorf("expected arguments but had: %T", source)
	}
	args := toolbox.AsSlice(source)
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments but had: %v", len(args))
	}

	collection := toolbox.AsSlice(args[0])
	for i, candidate := range collection {
		if candidate == args[1] || toolbox.AsString(candidate) == toolbox.AsString(args[1]) {
			return i, nil
		}
	}
	return -1, nil
}

func TestParseExpression(t *testing.T) {
	var useCases = []struct {
		description string
		aMap        Map
		expression  string
		expected    interface{}
	}{
		{
			description: "simple variable",
			aMap: Map(map[string]interface{}{
				"k1": 123,
			}),
			expression: "$k1",
			expected:   123,
		},
		{
			description: "simple enclosed variable",
			aMap: Map(map[string]interface{}{
				"k1": 123,
			}),
			expression: "${k1}",
			expected:   123,
		},

		{
			description: "simple embedding",
			aMap: Map(map[string]interface{}{
				"k1": 123,
			}),
			expression: "abc $k1 xyz",
			expected:   "abc 123 xyz",
		},

		{
			description: "simple embedding",
			aMap: Map(map[string]interface{}{
				"k1": 123,
			}),
			expression: "abc $k1/xyz",
			expected:   "abc 123/xyz",
		},

		{
			description: "double embedding",
			aMap: Map(map[string]interface{}{
				"k1": 123,
				"k2": 88,
			}),
			expression: "abc $k1 xyz $k2 ",
			expected:   "abc 123 xyz 88 ",
		},

		{
			description: "enclosing ",
			aMap: Map(map[string]interface{}{
				"k1": 123,
				"k2": 88,
			}),
			expression: "abc ${k1} xyz $k2 ",
			expected:   "abc 123 xyz 88 ",
		},

		{
			description: "enclosing and partialy unexpanded",
			aMap: Map(map[string]interface{}{
				"k1": 123,
				"k2": 88,
			}),
			expression: " $z1 abc ${k1} xyz $k2 ",
			expected:   " $z1 abc 123 xyz 88 ",
		},

		{
			description: "sub key access",
			aMap: Map(map[string]interface{}{
				"k2": map[string]interface{}{
					"z": 111,
					"x": 333,
				},
			}),
			expression: "abc ${k2.z} xyz $k2.x/ ",
			expected:   "abc 111 xyz 333/ ",
		},

		{
			description: "slice & nested access",
			aMap: Map(map[string]interface{}{
				"array": []interface{}{
					map[string]interface{}{
						"z": 111,
						"x": map[string]interface{}{
							"k": 444,
						},
						"y": []interface{}{"a", "b"},
					},
				},
			}),
			expression: "abc $array[0].z $array[0].y[0]* !${array[0].x.k}#$array[0].x.k",
			expected:   "abc 111 a* !444#444",
		},

		{
			description: "slice with index variable",
			aMap: Map(map[string]interface{}{
				"i": 1,
				"array": []interface{}{
					111, 222, 333,
				},
			}),
			expression: "$array[$i]",
			expected:   222,
		},
		{
			description: "slice with index variable",
			aMap: Map(map[string]interface{}{
				"i": 2,
				"array": []interface{}{
					111, 222, 333,
				},
			}),
			expression: "$array[${i}]",
			expected:   333,
		},
		{
			description: "slice with index variable",
			aMap: Map(map[string]interface{}{
				"i": 2,
				"array": []interface{}{
					111, 222, 333,
				},
			}),
			expression: "${array[${i}]}",
			expected:   333,
		},
		{
			description: "variable func",
			aMap: Map(map[string]interface{}{
				"f": func(key interface{}, state Map) (interface{}, error) {
					return "test " + toolbox.AsString(key), nil
				},
			}),
			expression: "$f(123)",
			expected:   "test 123",
		},
		{
			description: "variable func",
			aMap: Map(map[string]interface{}{
				"f": func(key interface{}, state Map) (interface{}, error) {
					return "test " + toolbox.AsString(key), nil
				},
			}),
			expression: "a $f(123) b",
			expected:   "a test 123 b",
		},
		{
			description: "variable func",
			aMap: Map(map[string]interface{}{
				"f": func(key interface{}, state Map) (interface{}, error) {
					return "test " + toolbox.AsString(key), nil
				},
			}),
			expression: "a ${f(123)} b",
			expected:   "a test 123 b",
		},

		{
			description: "variable func with unexpanded variables",
			aMap: Map(map[string]interface{}{
				"f": func(key interface{}, state Map) (interface{}, error) {
					return "test " + toolbox.AsString(key), nil
				},
			}),
			expression: "${a()} ${f(123)} $b()",
			expected:   "${a()} test 123 $b()",
		},

		{
			description: "variable func with slice arguments",
			aMap: Map(map[string]interface{}{
				"f": func(args interface{}, state Map) (interface{}, error) {

					aSlice := toolbox.AsSlice(args)
					textSlice := []string{}
					for _, item := range aSlice {
						textSlice = append(textSlice, toolbox.AsString(item))
					}
					return strings.Join(textSlice, ":"), nil
				},
			}),
			expression: `! $f(["a", "b", "c"]) !`,
			expected:   "! a:b:c !",
		},
		{
			description: "variable func with aMap arguments",
			aMap: Map(map[string]interface{}{
				"f": func(args interface{}, state Map) (interface{}, error) {
					aMap := toolbox.AsMap(args)
					aSlice := []string{}
					for k, v := range aMap {
						aSlice = append(aSlice, toolbox.AsString(fmt.Sprintf("%v->%v", k, v)))
					}
					sort.Strings(aSlice)
					return strings.Join(aSlice, ":"), nil
				},
			}),
			expression: `! $f({"a":1, "b":2, "c":3}) !`,
			expected:   "! a->1:b->2:c->3 !",
		},

		{
			description: "slice element shift",
			aMap: Map(map[string]interface{}{
				"s": []interface{}{3, 2, 1},
			}),
			expression: `! $<-s ${<-s} !`,
			expected:   "! 3 2 !",
		},
		{
			description: "element inc",
			aMap: Map(map[string]interface{}{
				"i": 2,
				"j": 5,
			}),
			expression: `!${i++}/${i}/${++i}!`,
			expected:   "!2/3/4!",
		},

		{
			description: "basic arithmetic",
			aMap: Map(map[string]interface{}{
				"i": 1,
				"j": 2,
				"k": 0.4,
			}),
			expression: `${(i + j) / 2}`,
			expected:   1.5,
		},
		{
			description: "enclosed basic arithmetic",
			aMap: Map(map[string]interface{}{
				"i": 1,
				"j": 2,
				"k": 0.4,
			}),
			expression: `z${(i + j) / 2}z`,
			expected:   "z1.5z",
		},
		{
			description: "multi arithmetic",
			aMap: Map(map[string]interface{}{
				"i": 1,
				"j": 2,
				"k": 0.4,
			}),
			expression: `${10 + 1 - 2}`,
			expected:   9,
		},
		{
			description: "sub attribute arithmetic",
			aMap: Map(map[string]interface{}{
				"i": 1,
				"j": 2,
				"k": map[string]interface{}{
					"z": 0.4,
				},
				"s": []interface{}{10},
			}),
			expression: `${k.z * s[0]}`,
			expected:   4,
		},
		{
			description: "unexpanded ",
			aMap: Map(map[string]interface{}{
				"index": 1,
			}),
			expression: `${index}*`,
			expected:   "1*",
		},

		{
			description: "unexpanded ",
			aMap: Map(map[string]interface{}{
				"i": 1,
			}),
			expression: `[]Orders,`,
			expected:   "[]Orders,",
		},

		{
			description: "unexpanded tags",
			aMap: Map(map[string]interface{}{
				"tag":   "Root",
				"tagId": "Root",
			}),
			expression: `[]Orders,Id,Name,LineItems,SubTotal`,
			expected:   "[]Orders,Id,Name,LineItems,SubTotal",
		},
		{
			description: "unexpanded dolar",
			aMap: Map(map[string]interface{}{
				"tag": "Root",
			}),
			expression: `$`,
			expected:   "$",
		},
		{
			description: "unexpanded enclosed dolar",
			aMap: Map(map[string]interface{}{
				"tag": "Root",
			}),
			expression: `a/$/z`,
			expected:   "a/$/z",
		},
		{
			description: "udf with text argument",
			aMap: Map(map[string]interface{}{
				"r": func(key interface{}, state Map) (interface{}, error) {
					return true, nil
				},
			}),
			expression: `$r(use_cases/001_event_processing_use_case/skip.txt):true`,
			expected:   "true:true",
		},

		{
			description: "int conversion",
			aMap: Map(map[string]interface{}{
				"AsInt": func(source interface{}, state Map) (interface{}, error) {
					return toolbox.AsInt(source), nil
				},
			}),
			expression: `z $AsInt(3434)`,
			expected:   "z 3434",
		},

		{
			description: "int conversion",
			aMap: Map(map[string]interface{}{
				"dailyCap":   100,
				"overallCap": 2,
				"AsFloat": func(source interface{}, state Map) (interface{}, error) {
					return toolbox.AsFloat(source), nil
				},
			}),
			expression: `{
		  "DAILY_CAP": "$AsFloat($dailyCap)"
		}`,
			expected: "{\n\t\t  \"DAILY_CAP\": \"100\"\n\t\t}",
		},

		{
			description: "post increment",
			aMap: map[string]interface{}{
				"i": 0,
				"z": 3,
			},
			expression: "$i++ $i  $z++ $z",
			expected:   "0 1  3 4",
		},

		{
			description: "pre increment",
			aMap: map[string]interface{}{
				"i": 10,
				"z": 20,
			},
			expression: "$++i $i  $++z $z",
			expected:   "11 11  21 21",
		},

		{
			description: "arguments as text glitch",
			aMap: map[string]interface{}{
				"f": func(source interface{}, state Map) (interface{}, error) {
					return source, nil
				},
			},
			expression: "#$f(554257_popularmechanics.com)#",
			expected:   "#554257_popularmechanics.com#",
		},

		{
			description: "embedded UDF expression",
			aMap: map[string]interface{}{
				"IndexOf":    IndexOf,
				"collection": []interface{}{"abc", "xtz"},
				"key":        "abc",
			},
			expression: `$IndexOf($collection, $key)`,
			expected:   0,
		},
		{
			description: "embedded UDF expression with literal",
			aMap: map[string]interface{}{
				"IndexOf":    IndexOf,
				"collection": []interface{}{"abc", "xtz"},
				"key":        "abc",
			},
			expression: `$IndexOf($collection, xtz)`,
			expected:   1,
		},

		{
			description: "multi udf neating",
			aMap: map[string]interface{}{
				"IndexOf":    IndexOf,
				"collection": []interface{}{"abc", "xtz"},
				"key":        "abc",
			},
			expression: `$IndexOf($collection, xtz)`,
			expected:   1,
		},
		{
			description: "unresolved expression",
			aMap: map[string]interface{}{
				"IndexOf":    IndexOf,
				"collection": []interface{}{"abc", "xtz"},
				"key":        "abc",
			},
			expression: `${$appPath}/hello/main.zip`,
			expected:   `${$appPath}/hello/main.zip`,
		},
		{
			description: "resolved  expression",
			aMap: map[string]interface{}{
				"appPath": "/abc/a",
			},
			expression: `${appPath}/hello/main.zip`,
			expected:   `/abc/a/hello/main.zip`,
		},

		{
			description: "byte extraction",
			aMap: map[string]interface{}{
				"Payload": []byte{
					34,
					72,
					101,
					108,
					108,
					111,
					32,
					87,
					111,
					114,
					108,
					100,
					34,
				},
				"AsString": func(source interface{}, state Map) (interface{}, error) {
					return toolbox.AsString(source), nil
				},
			},
			expression: `$AsString($Payload)`,
			expected:   `"Hello World"`,
		},
	}

	//$Join($AsCollection($Cat($env.APP_HOME/app-config/schema/go/3.json)), “,”)

	for _, useCase := range useCases {
		var expandHandler = func(expression string, isUDF bool, argument interface{}) (interface{}, bool) {
			result, has := useCase.aMap.GetValue(string(expression[1:]))
			if isUDF {
				if udf, ok := result.(func(interface{}, Map) (interface{}, error)); ok {
					expandedArgs := useCase.aMap.expandArgumentsExpressions(argument)
					if toolbox.IsString(expandedArgs) && toolbox.IsStructuredJSON(toolbox.AsString(expandedArgs)) {
						if evaluated, err := toolbox.JSONToInterface(toolbox.AsString(expandedArgs)); err == nil {
							expandedArgs = evaluated
						}
					}
					result, err := udf(expandedArgs, nil)
					return result, err == nil
				}
			}
			return result, has
		}
		actual := Parse(useCase.expression, expandHandler)
		if !assert.Equal(t, useCase.expected, actual, useCase.description) {
			fmt.Printf("!%v!\n", actual)
		}
	}
}
