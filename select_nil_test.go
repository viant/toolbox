package toolbox_test

import (
	"github.com/viant/toolbox"
	"testing"
)

func TestIsSliceNilAndTypedKinds(t *testing.T) {
	var slice []interface{}
	var slicePointer *[]interface{}
	var empty interface{}
	for name, tc := range map[string]struct {
		value interface{}
		want  bool
	}{
		"nil": {nil, false}, "nil interface pointer": {&empty, false}, "typed nil slice": {slice, true}, "typed nil slice pointer": {slicePointer, true}, "map": {map[string]interface{}{}, false}, "array": {[1]int{}, false}, "slice": {[]int{0}, true},
	} {
		t.Run(name, func(t *testing.T) {
			if actual := toolbox.IsSlice(tc.value); actual != tc.want {
				t.Fatalf("kind mismatch: got%t want%t", actual, tc.want)
			}
		})
	}
	if toolbox.IsMap(nil) {
		t.Fatal("nil is not a map")
	}
}
