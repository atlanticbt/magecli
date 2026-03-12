package root

import (
	"bytes"
	"io"
	"testing"

	"github.com/atlanticbt/magecli/internal/config"
	"github.com/atlanticbt/magecli/pkg/cmdutil"
	"github.com/atlanticbt/magecli/pkg/iostreams"
)

func testFactory() *cmdutil.Factory {
	return &cmdutil.Factory{
		AppVersion:     "test",
		ExecutableName: "magecli",
		IOStreams: &iostreams.IOStreams{
			In:     io.NopCloser(&bytes.Buffer{}),
			Out:    &bytes.Buffer{},
			ErrOut: &bytes.Buffer{},
		},
		Config: func() (*config.Config, error) {
			return &config.Config{
				Version:  1,
				Contexts: make(map[string]*config.Context),
				Hosts:    make(map[string]*config.Host),
			}, nil
		},
	}
}

func TestNewCmdRoot(t *testing.T) {
	f := testFactory()
	cmd, err := NewCmdRoot(f)
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Use != "magecli" {
		t.Errorf("Use = %q, want magecli", cmd.Use)
	}
}

func TestRootHasSubcommands(t *testing.T) {
	f := testFactory()
	cmd, _ := NewCmdRoot(f)

	expected := []string{"auth", "context", "product", "category", "attribute", "inventory", "store", "config", "promo", "cms", "api"}
	commands := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		commands[sub.Name()] = true
	}
	for _, name := range expected {
		if !commands[name] {
			t.Errorf("missing subcommand %q", name)
		}
	}
}

func TestRootGlobalFlags(t *testing.T) {
	f := testFactory()
	cmd, _ := NewCmdRoot(f)

	flags := []string{"context", "store-code", "json", "yaml", "jq", "template"}
	for _, name := range flags {
		if cmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("missing global flag --%s", name)
		}
	}
}

func TestRootHelpRuns(t *testing.T) {
	f := testFactory()
	cmd, _ := NewCmdRoot(f)

	// Running with no args should print help without error
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Errorf("help should not error: %v", err)
	}
}
