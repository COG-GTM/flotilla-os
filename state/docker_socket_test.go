package state

import (
	"testing"

	"github.com/stitchfix/flotilla-os/config"
)

type sliceConfig struct {
	config.Config
	values map[string][]string
}

func (c sliceConfig) GetStringSlice(key string) []string { return c.values[key] }

func TestDockerSocketAllowlist(t *testing.T) {
	if NewDockerSocketAllowlist(nil).Permits("any") {
		t.Errorf("nil config must permit nothing")
	}
	empty := NewDockerSocketAllowlist(sliceConfig{values: map[string][]string{}})
	if empty.Permits("any") || empty.Permits("") {
		t.Errorf("unset allowlist must permit nothing")
	}

	a := NewDockerSocketAllowlist(sliceConfig{values: map[string][]string{
		DockerSocketAllowlistKey: {"def-a, def-b", " def-c ", ""},
	}})
	for _, id := range []string{"def-a", "def-b", "def-c"} {
		if !a.Permits(id) {
			t.Errorf("expected %q to be permitted", id)
		}
	}
	for _, id := range []string{"", "def-d", "def-a, def-b"} {
		if a.Permits(id) {
			t.Errorf("expected %q to be rejected", id)
		}
	}
}
