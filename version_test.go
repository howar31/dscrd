package dscrd

import (
	"strings"
	"testing"
)

func TestVersionNonEmpty(t *testing.T) {
	if Version == "" {
		t.Fatal("Version must not be empty")
	}
}

func TestVersionTrimmed(t *testing.T) {
	if strings.TrimSpace(Version) != Version {
		t.Fatalf("Version %q contains surrounding whitespace", Version)
	}
}
