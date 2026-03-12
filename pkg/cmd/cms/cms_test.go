package cms

import "testing"

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
		{"abcd", 3, "..."},  // edge case: truncates to just ellipsis when max == 3
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if tt.max >= 3 {
				if got != tt.want {
					t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
				}
			}
			// For any max, result should not exceed max length
			if len(got) > tt.max {
				t.Errorf("truncate(%q, %d) length = %d, exceeds max", tt.input, tt.max, len(got))
			}
		})
	}
}
