package udf

import (
	"github.com/viant/toolbox/data"
	"reflect"
	"strings"
	"testing"
)

func TestSelectMissingNullAndInvalidInputs(t *testing.T) {
	var nilSlice []interface{}
	var nilPointer *[]interface{}
	for _, params := range []interface{}{nil, nilSlice, nilPointer, []interface{}{}, "", "missing[", "missing[]", "missing[name = 'unclosed]"} {
		if _, err := Select(params, data.NewMap()); err == nil {
			t.Fatalf("invalid selector accepted (%T)", params)
		}
	}
	state := data.Map{"null": nil, "false": false, "zero": 0, "array": []interface{}{}, "cookies": map[string]interface{}{"Result": []interface{}{map[string]interface{}{"cookies": nil}}}}
	for _, path := range []string{"missing", "null", "missing/*", "null/*", "missing[name = studio_session]", "cookies.Result[0].cookies[name = studio_session]"} {
		value, err := Select(path, state)
		if err != nil || len(value.([]interface{})) != 0 {
			t.Fatalf("missing/null path did not fail closed: value%T error%v", value, err)
		}
	}
	for _, path := range []string{"false", "zero", "array"} {
		value, err := Select(path, state)
		if err != nil || !reflect.DeepEqual(value, []interface{}{state[path]}) {
			t.Fatalf("literal leaf changed (%T),error%v", value, err)
		}
	}
	value, err := Select("missing[name = studio_session]", nil)
	if err != nil || len(value.([]interface{})) != 0 {
		t.Fatal("nil state must be empty")
	}
}

type selectRowAlias map[string]interface{}

func TestSelectPredicateRowsFailClosedAndPreserveAliases(t *testing.T) {
	valid := map[string]interface{}{"name": "studio_session", "value": "synthetic"}
	for _, invalid := range []interface{}{nil, false, 0, "PRIVATE_PAYLOAD", []interface{}{}, (*map[string]interface{})(nil)} {
		state := data.Map{"rows": []interface{}{valid, invalid}}
		value, err := Select("rows[name = studio_session]", state)
		if err == nil {
			t.Fatalf("invalid row accepted (%T)", invalid)
		}
		if len(value.([]interface{})) != 0 {
			t.Fatal("invalid collection returned partial match")
		}
		if len(err.Error()) > 120 || strings.Contains(err.Error(), "PRIVATE_PAYLOAD") || strings.Contains(err.Error(), "studio_session") {
			t.Fatal("unbounded/value-bearing error")
		}
	}
	for _, invalid := range []interface{}{false, 0, "PRIVATE_PAYLOAD", map[string]interface{}{}} {
		_, err := Select("rows[name = studio_session]", data.Map{"rows": invalid})
		if err == nil || strings.Contains(err.Error(), "PRIVATE_PAYLOAD") {
			t.Fatalf("invalid collection accepted (%T)", invalid)
		}
	}
	alias := selectRowAlias{"name": "studio_session", "value": "$ShouldNotRun()"}
	value, err := Select("rows[name = studio_session]", data.Map{"rows": []interface{}{alias}})
	if err != nil || !reflect.DeepEqual(value, []interface{}{alias}) {
		t.Fatal("map alias changed")
	}
	if alias["value"] != "$ShouldNotRun()" {
		t.Fatal("literal browser string was evaluated")
	}
}

func TestSelectIndexedQuotedPathsAndPassiveLookup(t *testing.T) {
	for _, name := range []string{"a[b]", "a]b["} {
		row := map[string]interface{}{"name": name, "value": "literal"}
		state := data.Map{"cookies": map[string]interface{}{"Result": []interface{}{map[string]interface{}{"cookies": []interface{}{row}}}}}
		value, err := Select("cookies.Result[0].cookies[name = '"+name+"']", state)
		if err != nil || !reflect.DeepEqual(value, []interface{}{row}) {
			t.Fatalf("indexed bracket-containing literal failed: %T error%v", value, err)
		}
	}
	for _, tc := range []struct{ name, path string }{{"a'[b]", `cookies.Result[0].cookies[name = 'a\'[b]']`}, {`a"[b]`, `cookies.Result[0].cookies[name = "a\"[b]"]`}} {
		row := map[string]interface{}{"name": tc.name}
		state := data.Map{"cookies": map[string]interface{}{"Result": []interface{}{map[string]interface{}{"cookies": []interface{}{row}}}}}
		value, err := Select(tc.path, state)
		if err != nil || !reflect.DeepEqual(value, []interface{}{row}) {
			t.Fatal("escaped quoted predicate literal changed")
		}
	}
	for _, path := range []string{`cookies.Result[0].cookies[name = 'a\'[b]']`, `cookies.Result[0].cookies[name = "a\"[b]"]`} {
		index, err := selectFinalBracket(path)
		if err != nil || index != strings.Index(path, "[name") {
			t.Fatalf("quoted bracket scanner failed, error%v", err)
		}
	}
	calls := 0
	state := data.Map{"blocked": func() interface{} { calls++; return []interface{}{map[string]interface{}{"name": "studio_session"}} }}
	value, err := Select("blocked.rows[name = studio_session]", state)
	if err != nil || len(value.([]interface{})) != 0 || calls != 0 {
		t.Fatal("Select invoked a getter while looking up collection")
	}
}

func TestSelectLegacySlashWildcardProjectionAndIndex(t *testing.T) {
	rows := []interface{}{map[string]interface{}{"id": 0, "name": "alpha"}, map[string]interface{}{"id": 1, "name": "beta"}}
	state := data.Map{"root": map[string]interface{}{"rows": rows}}
	for _, path := range []string{"root/rows/*", "root.rows"} {
		value, err := Select(path, state)
		if err != nil {
			t.Fatal(err)
		}
		want := interface{}(rows)
		if path == "root.rows" {
			want = []interface{}{rows}
		}
		if !reflect.DeepEqual(value, want) {
			t.Fatalf("legacy traversal changed, got%T", value)
		}
	}
	value, err := Select([]interface{}{"root/rows[id = 1]", "name:label"}, state)
	if err != nil || !reflect.DeepEqual(value, []interface{}{map[string]interface{}{"label": "beta"}}) {
		t.Fatal("projection changed")
	}
	value, err = Select("root.rows[0]", state)
	if err != nil || !reflect.DeepEqual(value, []interface{}{rows[0]}) {
		t.Fatal("numeric index-only path changed")
	}
}
