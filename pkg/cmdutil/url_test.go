package cmdutil

import "testing"

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"full https URL", "https://store.example.com", "https://store.example.com", false},
		{"http URL", "http://store.example.com", "http://store.example.com", false},
		{"bare hostname", "store.example.com", "https://store.example.com", false},
		{"trailing slash", "https://store.example.com/", "https://store.example.com", false},
		{"with path", "https://store.example.com/magento", "https://store.example.com/magento", false},
		{"trailing slash on path", "https://store.example.com/magento/", "https://store.example.com/magento", false},
		{"strips query", "https://store.example.com?foo=bar", "https://store.example.com", false},
		{"strips fragment", "https://store.example.com#section", "https://store.example.com", false},
		{"empty", "", "", true},
		{"whitespace", "  ", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeBaseURL(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("NormalizeBaseURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHostKeyFromURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"standard", "https://store.example.com", "store.example.com", false},
		{"with port", "https://store.example.com:8080", "store.example.com:8080", false},
		{"with path", "https://store.example.com/magento", "store.example.com", false},
		{"no host", "not-a-url", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HostKeyFromURL(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("HostKeyFromURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
