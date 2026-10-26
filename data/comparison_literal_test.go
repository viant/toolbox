package data_test

import (
	"github.com/viant/toolbox/data"
	"testing"
)

func TestQuotedComparisonPreservesLiteralDataAndDoesNotRunUDF(t *testing.T) {
	state := data.NewMap()
	calls := 0
	state["Probe"] = func(interface{}, data.Map) (interface{}, error) { calls++; return "changed", nil }
	for _, tc := range []struct{ name, expression string }{
		{"a[b]", `${name = 'a[b]'}`}, {"a[b]", `${name = "a[b]"}`},
		{"a'[b]", `${name = 'a\'[b]'}`}, {`a"[b]`, `${name = "a\"[b]"}`},
		{"$Probe()", `${name = '$Probe()'}`}, {"$Probe()", `${name = "$Probe()"}`},
		{"$name", `${name = '$name'}`},
	} {
		state["name"] = tc.name
		actual := state.Expand(tc.expression)
		if actual != true {
			t.Fatalf("literal comparison changed: %#v (%T)", actual, actual)
		}
	}
	if calls != 0 {
		t.Fatal("quoted literal/data executed a UDF")
	}
	state["name"] = "different"
	if actual := state.Expand(`${name = '$Probe()'}`); actual != false {
		t.Fatalf("nonmatching quoted literal: %#v", actual)
	}
	if calls != 0 {
		t.Fatal("nonmatching literal executed a UDF")
	}
}

func TestQuotedComparisonKeepsUnquotedNumericBooleanAndStringSemantics(t *testing.T) {
	state := data.Map{"zero": 0, "one": 1, "enabled": true, "disabled": false, "name": "alpha"}
	for expression, want := range map[string]bool{
		`${zero = 0}`: true, `${one > 0}`: true, `${one < 0}`: false,
		`${enabled & disabled}`: false, `${enabled | disabled}`: true,
		`${name = alpha}`: true, `${name = other}`: false,
	} {
		if actual := state.Expand(expression); actual != want {
			t.Fatalf("legacy comparison changed: %s -> %#v (%T)", expression, actual, actual)
		}
	}
}
