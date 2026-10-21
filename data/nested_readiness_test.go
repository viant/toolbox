package data

import (
	"github.com/viant/toolbox"
	"reflect"
	"testing"
)

func TestNestedUDFReadiness(t *testing.T) {
	type useCase struct {
		description, input string
		expected           interface{}
		calls              int
	}
	cases := []useCase{
		{"nested path", "$Count(${ci_ads.CI_CONTACTS.${tag}.member.ID})", 123, 1},
		{"ordinary path", "$Count(${contactID})", 123, 1},
		{"missing inner", "$Count(${ci_ads.CI_CONTACTS.${missing}.member.ID})", "$Count(${ci_ads.CI_CONTACTS.${missing}.member.ID})", 0},
		{"missing outer", "$Count(${absent.${tag}.ID})", "$Count(${absent.case.ID})", 0},
		{"returned expression", "$Text(${contactID})", "$Count(123)", 1},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			calls := 0
			state := Map{"tag": "case", "contactID": 123, "ci_ads": map[string]interface{}{"CI_CONTACTS": map[string]interface{}{"case": map[string]interface{}{"member": map[string]interface{}{"ID": 123}}}}}
			state["Count"] = func(v interface{}, _ Map) (interface{}, error) { calls++; return toolbox.AsInt(v), nil }
			state["Text"] = func(v interface{}, _ Map) (interface{}, error) { calls++; return "$Count(123)", nil }
			got := state.Expand(tc.input)
			if !reflect.DeepEqual(tc.expected, got) || calls != tc.calls {
				t.Fatalf("got %#v calls %d; want %#v calls %d", got, calls, tc.expected, tc.calls)
			}
		})
	}
}

func TestPassiveReadinessDoesNotExecuteGetters(t *testing.T) {
	type useCase struct{ description, input string }
	cases := []useCase{
		{"function leaf", "${rows.${tag}.value}"},
		{"mutating path", "${rows.${tag}.value++}"},
		{"increment", "$counter++"},
		{"shift", "$<-items"},
		{"nested UDF", "$Getter()"},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			calls := 0
			getter := func() interface{} { calls++; return 123 }
			state := Map{"tag": "case", "counter": 1, "items": []interface{}{1, 2}, "Getter": getter, "rows": map[string]interface{}{"case": map[string]interface{}{"value": getter}}}
			if state.passiveUDFArgumentReady(tc.input) {
				t.Fatal("unsupported expression accepted")
			}
			if calls != 0 || state["counter"] != 1 || len(state["items"].([]interface{})) != 2 {
				t.Fatal("passive traversal changed state")
			}
		})
	}
}

func TestNestedReadinessGetterFallback(t *testing.T) {
	getters, calls := 0, 0
	state := Map{"tag": "case", "rows": map[string]interface{}{"case": map[string]interface{}{"value": func() interface{} { getters++; return 123 }}}}
	state["Count"] = func(v interface{}, _ Map) (interface{}, error) { calls++; return toolbox.AsInt(v), nil }
	got := state.Expand("$Count(${rows.${tag}.value})")
	if got != "$Count(123)" || getters != 1 || calls != 0 {
		t.Fatalf("got %#v, getters %d, UDF calls %d", got, getters, calls)
	}
}
