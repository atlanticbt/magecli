package api

import (
	"encoding/json"
	"testing"
)

func TestParseKeyValue(t *testing.T) {
	tests := []struct {
		input     string
		wantKey   string
		wantValue string
		wantErr   bool
	}{
		{"key=value", "key", "value", false},
		{"name=John Doe", "name", "John Doe", false},
		{"a=b=c", "a", "b=c", false},
		{"noequals", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			k, v, err := parseKeyValue(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if k != tt.wantKey || v != tt.wantValue {
				t.Errorf("got (%q, %q), want (%q, %q)", k, v, tt.wantKey, tt.wantValue)
			}
		})
	}
}

func TestParseHeader(t *testing.T) {
	tests := []struct {
		input     string
		wantKey   string
		wantValue string
		wantErr   bool
	}{
		{"Content-Type: application/json", "Content-Type", "application/json", false},
		{"X-Custom: value:with:colons", "X-Custom", "value:with:colons", false},
		{"no-colon", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			k, v, err := parseHeader(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if k != tt.wantKey || v != tt.wantValue {
				t.Errorf("got (%q, %q), want (%q, %q)", k, v, tt.wantKey, tt.wantValue)
			}
		})
	}
}

func TestInferJSONValue(t *testing.T) {
	tests := []struct {
		input string
		want  any
	}{
		{"42", json.Number("42")},
		{"true", true},
		{"false", false},
		{"null", nil},
		{`"hello"`, "hello"},
		{"plain text", "plain text"},
		{"", ""},
		{`[1,2,3]`, []any{json.Number("1"), json.Number("2"), json.Number("3")}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := inferJSONValue(tt.input)
			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(tt.want)
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("inferJSONValue(%q) = %s, want %s", tt.input, gotJSON, wantJSON)
			}
		})
	}
}
