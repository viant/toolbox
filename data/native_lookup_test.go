package data_test

import (
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/data"
	"github.com/viant/toolbox/data/udf"
	"reflect"
	"testing"
)

func TestNativeDynamicIndexUDFArguments(t *testing.T) {
	collection := data.Collection{"alpha", "beta"}
	for name, choices := range map[string]interface{}{
		"strings": []string{"alpha", "beta"}, "interfaces": []interface{}{"alpha", "beta"},
		"collection": collection, "collectionPointer": &collection,
	} {
		t.Run(name, func(t *testing.T) {
			state := data.NewMap()
			udf.Register(state)
			state["choices"], state["readerIndex"] = choices, 1
			state["sampleReaders"] = []map[string]interface{}{{"name": "alpha"}, {"name": "omega"}}
			for _, item := range []struct {
				expression string
				want       interface{}
			}{
				{"$Len($choices)", 2}, {"$sampleReaders[$readerIndex].name", "omega"},
				{"$IndexOf($choices, alpha)", 0}, {"$IndexOf($choices, omega)", -1},
				{"$IndexOf($choices, $sampleReaders[$readerIndex].name)", -1},
			} {
				got := state.Expand(item.expression)
				t.Logf("%s -> %#v (%T)", item.expression, got, got)
				if !reflect.DeepEqual(got, item.want) {
					t.Fatalf("%s = %#v (%T), want %#v (%T)", item.expression, got, got, item.want, item.want)
				}
			}
			state["readerIndex"] = 0
			if got := state.Expand("$IndexOf($choices, $sampleReaders[$readerIndex].name)"); got != 0 {
				t.Fatalf("index was frozen: %#v (%T)", got, got)
			}
		})
	}
}

func TestNativeTypedStructPathsPreserveJSONNamesAndValues(t *testing.T) {
	type captured struct {
		Response  map[string]interface{} `json:"response"`
		RequestID string                 `json:"requestId"`
	}
	type response struct {
		Transactions []*captured
		Enabled      bool        `json:"enabled"`
		Count        int         `json:"count"`
		Empty        []string    `json:"empty"`
		Nil          interface{} `json:"nil"`
		Hidden       string      `json:"-"`
	}
	call := &captured{Response: map[string]interface{}{"result": map[string]interface{}{"structuredContent": map[string]interface{}{"sessionId": "child-only", "literal": "$ForeignDSL(keep)"}}}, RequestID: "one"}
	original := &response{Transactions: []*captured{call}, Empty: []string{}, Hidden: "not-public"}
	state := data.Map{"capture": original, "caseMap": map[string]interface{}{"Response": "upper", "response": "lower"}}
	for _, item := range []struct {
		path  string
		want  interface{}
		found bool
	}{
		{"capture.Transactions[0].response.result.structuredContent.sessionId", "child-only", true},
		{"capture.Transactions[0].Response.result.structuredContent.sessionId", "child-only", true},
		{"capture.Transactions[0].requestId", "one", true},
		{"capture.Transactions[0].response.result.structuredContent.literal", "$ForeignDSL(keep)", true},
		{"capture.enabled", false, true}, {"capture.count", 0, true}, {"capture.nil", nil, true},
		{"capture.empty", []string{}, true}, {"capture.Hidden", nil, false}, {"capture.missing", nil, false},
		{"caseMap.Response", "upper", true}, {"caseMap.response", "lower", true},
	} {
		got, found := state.GetValue(item.path)
		if found != item.found || !reflect.DeepEqual(got, item.want) {
			t.Fatalf("%s=%#v (%T),%t want %#v,%t", item.path, got, got, found, item.want, item.found)
		}
	}
	if state["capture"] != original || original.Transactions[0] != call {
		t.Fatal("lookup replaced caller data")
	}
}

func TestNativeTypedReaderUDFArguments(t *testing.T) {
	type reader struct {
		Name string `json:"name"`
	}
	state := data.NewMap()
	udf.Register(state)
	state["choices"] = []string{"alpha", "beta"}
	state["readers"] = []*reader{{Name: "alpha"}, {Name: "omega"}}
	state["index"] = 1
	if got := state.Expand("$IndexOf($choices, $readers[$index].name)"); got != -1 {
		t.Fatalf("typed dynamic argument=%#v (%T), want int(-1)", got, got)
	}
}

func TestNativePassiveLookupNeverMutatesOrCallsGetters(t *testing.T) {
	calls := 0
	state := data.Map{"counter": 1, "items": []interface{}{1}, "getter": func(string) interface{} { calls++; return "side-effect" }, "getter0": func() interface{} { calls++; return "side-effect" }, "nil": nil}
	for _, path := range []string{"counter++", "++counter", "<-items", "getter.value"} {
		if value, found := state.LookupValue(path); found {
			t.Fatalf("passive path %s unexpectedly evaluated: %#v", path, value)
		}
	}
	if _, found := state.LookupValue("getter0"); !found {
		t.Fatal("function leaf lost presence")
	}
	if calls != 0 || state["counter"] != 1 {
		t.Fatal("passive path had side effects")
	}
	if value, found := state.LookupValue("nil"); !found || value != nil {
		t.Fatal("nil leaf lost presence")
	}
}

func TestNativeLookupArrayBoundsAreNotFound(t *testing.T) {
	state := data.Map{"rows": []interface{}{"first"}}
	for _, path := range []string{"rows[-1]", "rows[1]"} {
		if value, found := state.LookupValue(path); found {
			t.Fatalf("passive %s unexpectedly resolved %#v", path, value)
		}
		if value, found := state.GetValue(path); found {
			t.Fatalf("ordinary %s unexpectedly resolved %#v", path, value)
		}
	}
}

func TestNativePassiveLookupWithNilIntermediateMap(t *testing.T) {
	var nilMap *map[string]interface{}
	var nilTypedMap *map[string]string
	present := map[string]string{"value": "nested"}
	state := data.Map{
		"leaf":    "root decoy",
		"present": &present,
		"rows":    []data.Map{{"leaf": "array"}},
	}
	for _, absent := range []interface{}{map[string]interface{}(nil), nilMap, nilTypedMap, &nilTypedMap} {
		state["node"] = absent
		for _, path := range []string{"node.leaf", "node[leaf]", "node.inner.leaf"} {
			if value, found := state.LookupValue(path); found {
				t.Fatalf("%s resolved through absent map: %#v", path, value)
			}
		}
		if value, found := state.LookupValue("node"); !found || !reflect.DeepEqual(value, absent) {
			t.Fatalf("present nil leaf lost: %#v, %t", value, found)
		}
	}
	for path, want := range map[string]string{"present.value": "nested", "present[value]": "nested", "rows[0].leaf": "array"} {
		if got, found := state.LookupValue(path); !found || got != want {
			t.Fatalf("%s = %#v, %t; want %q", path, got, found, want)
		}
	}
}

func TestToMapStillCopiesPointerToNamedMap(t *testing.T) {
	type namedMap map[string]interface{}
	source := namedMap{"leaf": "original"}
	converted, err := toolbox.ToMap(&source)
	if err != nil {
		t.Fatal(err)
	}
	converted["leaf"] = "changed"
	if source["leaf"] != "original" {
		t.Fatal("conversion of a pointer to a named map now aliases the caller")
	}
}

func TestNativeLengthNilAndMissingRemainDistinct(t *testing.T) {
	for name, source := range map[string]interface{}{"nil": nil, "emptySlice": []string{}, "emptyMap": map[string]interface{}{}, "emptyText": ""} {
		value, err := udf.Length(source, nil)
		if err != nil || value != 0 {
			t.Fatalf("%s Length -> %#v (%T), err=%v", name, value, value, err)
		}
	}
	state := data.NewMap()
	udf.Register(state)
	state["optional"] = nil
	if value := state.Expand("$Len($optional)"); value != 0 {
		t.Fatalf("present nil Len -> %#v (%T)", value, value)
	}
	if value := state.Expand("$Len($missing)"); value != "$Len($missing)" {
		t.Fatalf("missing Len must remain unresolved, got %#v (%T)", value, value)
	}
}
