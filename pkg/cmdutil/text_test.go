package cmdutil

import "testing"

func TestValidateLimit(t *testing.T) {
	tests := []struct {
		limit   int
		wantErr bool
	}{
		{1, false},
		{20, false},
		{10000, false},
		{0, true},
		{-5, true},
		{10001, true},
	}
	for _, tt := range tests {
		err := ValidateLimit(tt.limit)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateLimit(%d) error = %v, wantErr %v", tt.limit, err, tt.wantErr)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},                      // max <= 3: no room for an ellipsis
		{"héllo wörld éxtra", 10, "héllo w..."}, // rune-safe: no mid-rune split
		{"héllo", 10, "héllo"},                  // multibyte but fits by rune count
		{"日本語のテキスト", 5, "日本..."},                // CJK runes counted, not bytes
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := Truncate(tt.input, tt.max); got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
			}
		})
	}
}
