package utils

import "testing"

func TestWildcardMatch(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		value   string
		want    bool
	}{
		{name: "exact", pattern: "project", value: "project", want: true},
		{name: "asterisk", pattern: "anchr-*", value: "anchr-mobile", want: true},
		{name: "question mark", pattern: "c?t", value: "cat", want: true},
		{name: "dot literal", pattern: "*.example.com", value: "api.example.com", want: true},
		{name: "slash match", pattern: "team/*", value: "team/backend", want: true},
		{name: "mismatch", pattern: "team/*", value: "org/backend", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WildcardMatch(tt.pattern, tt.value); got != tt.want {
				t.Fatalf("WildcardMatch(%q, %q) = %v, want %v", tt.pattern, tt.value, got, tt.want)
			}
		})
	}
}
