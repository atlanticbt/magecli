package iostreams

import "testing"

func TestSystem_NotNil(t *testing.T) {
	ios := System()
	if ios == nil {
		t.Fatal("System() returned nil")
	}
	if ios.In == nil || ios.Out == nil || ios.ErrOut == nil {
		t.Error("streams should not be nil")
	}
}

func TestNilIOStreams_CanPrompt(t *testing.T) {
	var ios *IOStreams
	if ios.CanPrompt() {
		t.Error("nil IOStreams should not CanPrompt")
	}
}

func TestNilIOStreams_ColorEnabled(t *testing.T) {
	var ios *IOStreams
	if ios.ColorEnabled() {
		t.Error("nil IOStreams should not have color enabled")
	}
}

func TestNilIOStreams_IsStdoutTTY(t *testing.T) {
	var ios *IOStreams
	if ios.IsStdoutTTY() {
		t.Error("nil IOStreams should not be stdout TTY")
	}
}

func TestNilIOStreams_IsStderrTTY(t *testing.T) {
	var ios *IOStreams
	if ios.IsStderrTTY() {
		t.Error("nil IOStreams should not be stderr TTY")
	}
}
