package format

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWrite_Fallback(t *testing.T) {
	var buf bytes.Buffer
	called := false
	err := Write(&buf, Options{}, nil, func() error {
		called = true
		buf.WriteString("table output")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("fallback was not called")
	}
	if buf.String() != "table output" {
		t.Errorf("output = %q", buf.String())
	}
}

func TestWrite_NilFallback(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, Options{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestWrite_JSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "Test"}
	err := Write(&buf, Options{Format: "json"}, data, nil)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["name"] != "Test" {
		t.Errorf("name = %q, want Test", result["name"])
	}
}

func TestWrite_JSON_Indented(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"a": "1"}
	_ = Write(&buf, Options{Format: "json"}, data, nil)
	if !strings.Contains(buf.String(), "  ") {
		t.Error("JSON output should be indented")
	}
}

func TestWrite_YAML(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "Test"}
	err := Write(&buf, Options{Format: "yaml"}, data, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "name: Test") {
		t.Errorf("expected YAML output, got %q", buf.String())
	}
}

func TestWrite_UnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, Options{Format: "xml"}, nil, nil)
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestWrite_JQ(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"items": []any{
		map[string]any{"sku": "A"},
		map[string]any{"sku": "B"},
	}}
	err := Write(&buf, Options{Format: "json", JQ: ".items[].sku"}, data, nil)
	if err != nil {
		t.Fatal(err)
	}
	output := strings.TrimSpace(buf.String())
	// jq with multiple results returns an array
	if !strings.Contains(output, "A") || !strings.Contains(output, "B") {
		t.Errorf("expected SKUs A and B, got %q", output)
	}
}

func TestWrite_JQ_InvalidExpression(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, Options{Format: "json", JQ: ".[invalid"}, map[string]any{}, nil)
	if err == nil {
		t.Error("expected error for invalid jq expression")
	}
}

func TestWrite_Template(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "Widget"}
	err := Write(&buf, Options{Template: "Product: {{.name}}"}, data, nil)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "Product: Widget" {
		t.Errorf("output = %q", buf.String())
	}
}

func TestWrite_Template_Invalid(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, Options{Template: "{{.bad"}, nil, nil)
	if err == nil {
		t.Error("expected error for invalid template")
	}
}

func TestWrite_JQ_SingleResult(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"name": "Widget"}
	err := Write(&buf, Options{Format: "json", JQ: ".name"}, data, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Widget") {
		t.Errorf("expected Widget, got %q", buf.String())
	}
}
