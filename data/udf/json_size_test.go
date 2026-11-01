package udf

import (
	"github.com/viant/toolbox/data"
	"testing"
)

func TestJSONSizeMeasuresEncodedBytes(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input interface{}
		want  int
	}{
		{"empty list", []interface{}{}, 2},
		{"empty object", map[string]interface{}{}, 2},
		{"null", nil, 4},
		{"number", 120, 3},
		{"unicode", "é", 4},
		{"unicode field", map[string]interface{}{"x": "é"}, 10},
		{"escaped HTML", "<", 8},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JSONSize(tt.input, data.NewMap())
			if err != nil || got != tt.want {
				t.Fatalf("got %v, error %v; want %d bytes", got, err, tt.want)
			}
		})
	}
	if _, err := JSONSize(func() {}, data.NewMap()); err == nil {
		t.Fatal("unsupported value accepted")
	}
}
