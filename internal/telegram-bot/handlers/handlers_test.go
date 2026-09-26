package handlers

import "testing"

func TestPositiveInteger(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int
		ok    bool
	}{
		{name: "one", value: "1", want: 1, ok: true},
		{name: "trim spaces", value: " 42 ", want: 42, ok: true},
		{name: "zero", value: "0"},
		{name: "negative", value: "-2"},
		{name: "fraction with dot", value: "1.5"},
		{name: "fraction with comma", value: "1,5"},
		{name: "plus sign", value: "+3"},
		{name: "empty", value: ""},
		{name: "text", value: "пять"},
		{name: "larger than database integer", value: "2147483648"},
		{name: "overflow", value: "999999999999999999999999999999999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := positiveInteger(tt.value)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("positiveInteger(%q) = (%d, %v), want (%d, %v)", tt.value, got, ok, tt.want, tt.ok)
			}
		})
	}
}
