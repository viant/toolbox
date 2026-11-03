package data

import (
	"bytes"
	"encoding/json"
	"reflect"
	"runtime"
	"testing"
)

func TestEncodableMapCopiesBinaryAndKeepsJSONBinaryContract(t *testing.T) {
	for _, binary := range [][]byte{nil, {}, []byte("\x89PNG\r\n\x1a\n${literal}\x00\xff")} {
		source := Map{"binary": binary, "nested": map[string]interface{}{"screenshot": binary}, "list": []interface{}{binary}}
		snapshot := source.AsEncodableMap()
		actual, ok := snapshot["binary"].([]byte)
		if !ok {
			t.Fatalf("binary snapshot type=%T, want []byte", snapshot["binary"])
		}
		if !reflect.DeepEqual(binary, actual) {
			t.Fatal("binary value changed")
		}
		expectedJSON, err := json.Marshal(struct {
			Binary []byte `json:"binary"`
		}{binary})
		if err != nil {
			t.Fatal(err)
		}
		actualJSON, err := json.Marshal(map[string]interface{}{"binary": actual})
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(expectedJSON, actualJSON) {
			t.Fatal("standard JSON binary representation changed")
		}
		var decoded struct {
			Binary []byte `json:"binary"`
		}
		if err = json.Unmarshal(actualJSON, &decoded); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(binary, decoded.Binary) {
			t.Fatal("JSON roundtrip changed binary")
		}
		nested := snapshot["nested"].(map[string]interface{})["screenshot"].([]byte)
		listed := snapshot["list"].([]interface{})[0].([]byte)
		if len(binary) > 0 {
			want := append([]byte(nil), binary...)
			binary[0] ^= 1
			if !bytes.Equal(want, actual) || !bytes.Equal(want, nested) || !bytes.Equal(want, listed) {
				t.Fatal("snapshot aliases source bytes")
			}
			actual[0] ^= 2
			if !bytes.Equal(want, nested) || !bytes.Equal(want, listed) {
				t.Fatal("snapshot paths alias binary storage")
			}
		}
	}
}

func TestEncodableMapNonBinaryBehaviorUnchanged(t *testing.T) {
	source := Map{UDFKey: map[string]interface{}{"hidden": 1}, "nil": nil, "function": func() {}, "number": 17, "text": "literal", "bool": false, "list": []interface{}{1, "two", nil, func() {}}, "map": map[string]interface{}{"nil": nil, "value": 3}, "empty": []interface{}{}}
	got := source.AsEncodableMap()
	want := map[string]interface{}{"function": "func()", "number": 17, "text": "literal", "bool": "false", "list": []interface{}{1, "two", nil, "func()"}, "map": map[string]interface{}{"value": 3}, "empty": []interface{}{}}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("nonbinary behavior changed: %#v", got)
	}
	if len(source) != 8+1 {
		t.Fatal("source map changed")
	}
}

var binarySnapshotSink map[string]interface{}

func TestEncodableBinaryAllocationsStayBounded(t *testing.T) {
	source := Map{"binary": bytes.Repeat([]byte{0, 1, 2, 255}, 4096)}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	binarySnapshotSink = source.AsEncodableMap()
	runtime.ReadMemStats(&after)
	t.Logf("16KiB snapshot allocated bytes: %d", after.TotalAlloc-before.TotalAlloc)
	if after.TotalAlloc-before.TotalAlloc > 64*1024 {
		t.Errorf("binary snapshot allocated %d bytes; want at most64KiB", after.TotalAlloc-before.TotalAlloc)
	}
	allocations := testing.AllocsPerRun(3, func() { binarySnapshotSink = source.AsEncodableMap() })
	t.Logf("16KiB snapshot allocations: %.0f", allocations)
	if allocations > 16 {
		t.Fatalf("binary snapshot allocated %.0f objects; want bounded binary copy", allocations)
	}
	if _, ok := binarySnapshotSink["binary"].([]byte); !ok {
		t.Fatal("snapshot did not preserve binary type")
	}
}
