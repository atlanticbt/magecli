package update

import (
	"testing"
)

func TestRepository(t *testing.T) {
	if repository != "atlanticbt/magecli" {
		t.Errorf("repository = %q, want atlanticbt/magecli", repository)
	}
}
