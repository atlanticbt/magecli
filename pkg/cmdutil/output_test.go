package cmdutil

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newOutputTestCmd(t *testing.T, jsonOut bool) *cobra.Command {
	t.Helper()
	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("yaml", false, "")
	root.PersistentFlags().String("jq", "", "")
	root.PersistentFlags().String("template", "", "")
	if jsonOut {
		if err := root.PersistentFlags().Set("json", "true"); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestValidateFields(t *testing.T) {
	tests := []struct {
		name    string
		fields  string
		jsonOut bool
		wantErr string
	}{
		{"empty fields, table output", "", false, ""},
		{"fields with json", "sku,name", true, ""},
		{"fields with table output", "sku,name", false, "--fields requires"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFields(newOutputTestCmd(t, tt.jsonOut), tt.fields)
			checkErr(t, err, tt.wantErr)
		})
	}
}

func TestValidateListFields(t *testing.T) {
	tests := []struct {
		name    string
		fields  string
		wantErr string
	}{
		{"bare list", "sku,name,price", ""},
		{"nested sub-projection is fine", "increment_id,items[sku,qty_ordered]", ""},
		// A pre-wrapped list gets wrapped again into items[items[...]],
		// which Magento answers with silently empty objects.
		{"pre-wrapped items", "items[sku,name,price],total_count", "wrapped in items[...] automatically"},
		{"pre-wrapped with whitespace", "  items[sku]", "wrapped in items[...] automatically"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateListFields(newOutputTestCmd(t, true), tt.fields)
			checkErr(t, err, tt.wantErr)
		})
	}
}

func checkErr(t *testing.T, err error, want string) {
	t.Helper()
	if want == "" {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error %q does not contain %q", err.Error(), want)
	}
}
