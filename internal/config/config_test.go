package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func tempConfig(t *testing.T) *Config {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	return &Config{
		Version:  1,
		Contexts: make(map[string]*Context),
		Hosts:    make(map[string]*Host),
		path:     path,
	}
}

func TestConfig_SetAndGetContext(t *testing.T) {
	cfg := tempConfig(t)
	ctx := &Context{Host: "example.com", StoreCode: "default"}
	cfg.SetContext("prod", ctx)

	got, err := cfg.Context("prod")
	if err != nil {
		t.Fatal(err)
	}
	if got.Host != "example.com" {
		t.Errorf("Host = %q, want example.com", got.Host)
	}
}

func TestConfig_ContextNotFound(t *testing.T) {
	cfg := tempConfig(t)
	_, err := cfg.Context("missing")
	if err != ErrContextNotFound {
		t.Errorf("err = %v, want ErrContextNotFound", err)
	}
}

func TestConfig_DeleteContext(t *testing.T) {
	cfg := tempConfig(t)
	cfg.SetContext("test", &Context{Host: "h"})
	cfg.ActiveContext = "test"
	cfg.DeleteContext("test")

	_, err := cfg.Context("test")
	if err != ErrContextNotFound {
		t.Error("context should be deleted")
	}
	if cfg.ActiveContext != "" {
		t.Error("active context should be cleared when deleted")
	}
}

func TestConfig_SetActiveContext(t *testing.T) {
	cfg := tempConfig(t)
	cfg.SetContext("prod", &Context{Host: "h"})
	if err := cfg.SetActiveContext("prod"); err != nil {
		t.Fatal(err)
	}
	if cfg.ActiveContext != "prod" {
		t.Errorf("ActiveContext = %q, want prod", cfg.ActiveContext)
	}
}

func TestConfig_SetActiveContext_NotFound(t *testing.T) {
	cfg := tempConfig(t)
	err := cfg.SetActiveContext("missing")
	if err != ErrContextNotFound {
		t.Errorf("err = %v, want ErrContextNotFound", err)
	}
}

func TestConfig_SetActiveContext_Empty(t *testing.T) {
	cfg := tempConfig(t)
	cfg.ActiveContext = "something"
	if err := cfg.SetActiveContext(""); err != nil {
		t.Fatal(err)
	}
	if cfg.ActiveContext != "" {
		t.Error("should clear active context")
	}
}

func TestConfig_SetAndGetHost(t *testing.T) {
	cfg := tempConfig(t)
	h := &Host{BaseURL: "https://example.com", StoreCode: "default"}
	cfg.SetHost("example.com", h)

	got, err := cfg.Host("example.com")
	if err != nil {
		t.Fatal(err)
	}
	if got.BaseURL != "https://example.com" {
		t.Errorf("BaseURL = %q", got.BaseURL)
	}
}

func TestConfig_HostNotFound(t *testing.T) {
	cfg := tempConfig(t)
	_, err := cfg.Host("missing")
	if err != ErrHostNotFound {
		t.Errorf("err = %v, want ErrHostNotFound", err)
	}
}

func TestConfig_DeleteHost(t *testing.T) {
	cfg := tempConfig(t)
	cfg.SetHost("h", &Host{BaseURL: "https://h.com"})
	cfg.DeleteHost("h")
	_, err := cfg.Host("h")
	if err != ErrHostNotFound {
		t.Error("host should be deleted")
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	cfg := tempConfig(t)
	cfg.SetHost("store.example.com", &Host{BaseURL: "https://store.example.com", Token: "secret-token"})
	cfg.SetContext("prod", &Context{Host: "store.example.com", StoreCode: "default"})
	cfg.ActiveContext = "prod"

	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	// Verify token is NOT written to disk
	data, err := os.ReadFile(cfg.path)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	hosts := raw["hosts"].(map[string]any)
	host := hosts["store.example.com"].(map[string]any)
	if token, ok := host["token"]; ok && token != "" {
		t.Errorf("token should not be written to disk, got %q", token)
	}
}

func TestConfig_SaveCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.yml")
	cfg := &Config{
		Version:  1,
		Contexts: make(map[string]*Context),
		Hosts:    make(map[string]*Host),
		path:     path,
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save should create directory: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("config file not created")
	}
}

func TestConfig_AllowWrites(t *testing.T) {
	cfg := tempConfig(t)
	cfg.SetContext("dev", &Context{Host: "dev.example.com", AllowWrites: true})

	ctx, err := cfg.Context("dev")
	if err != nil {
		t.Fatal(err)
	}
	if !ctx.AllowWrites {
		t.Error("AllowWrites should be true")
	}
}

func TestConfig_AllowWrites_DefaultFalse(t *testing.T) {
	cfg := tempConfig(t)
	cfg.SetContext("prod", &Context{Host: "prod.example.com"})

	ctx, _ := cfg.Context("prod")
	if ctx.AllowWrites {
		t.Error("AllowWrites should default to false")
	}
}

func TestHost_MarshalYAML_StripToken(t *testing.T) {
	h := &Host{BaseURL: "https://example.com", Token: "secret", StoreCode: "default"}
	out, err := h.MarshalYAML()
	if err != nil {
		t.Fatal(err)
	}
	data, _ := yaml.Marshal(out)
	if string(data) != "" {
		var check map[string]any
		yaml.Unmarshal(data, &check)
		if token, ok := check["token"]; ok && token != "" {
			t.Error("MarshalYAML should strip token")
		}
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	t.Setenv("MAGECLI_CONFIG_DIR", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != 1 {
		t.Errorf("Version = %d, want 1", cfg.Version)
	}
	if cfg.Contexts == nil {
		t.Error("Contexts should be initialized")
	}
}
