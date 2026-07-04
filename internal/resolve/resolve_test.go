package resolve

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeLookup struct {
	calls int
	fail  bool
}

func (f *fakeLookup) Name(kind, id string) (string, error) {
	f.calls++
	if f.fail {
		return "", errors.New("lookup failed")
	}
	return kind + "-name-" + id, nil
}

func TestResolveCachesLookups(t *testing.T) {
	path := filepath.Join(t.TempDir(), "resolve.json")
	l := &fakeLookup{}
	r := New(path, l)

	got := r.Resolve("user", "444444444444444444")
	if got != "user-name-444444444444444444" {
		t.Fatalf("Resolve = %q", got)
	}
	r.Resolve("user", "444444444444444444")
	if l.calls != 1 {
		t.Fatalf("lookup calls = %d, want 1 (cache hit)", l.calls)
	}

	// Cache must be persisted and readable by a fresh resolver.
	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), "user:444444444444444444") {
		t.Fatalf("cache file = %s, %v", data, err)
	}
	l2 := &fakeLookup{}
	r2 := New(path, l2)
	if r2.Resolve("user", "444444444444444444") != "user-name-444444444444444444" || l2.calls != 0 {
		t.Fatal("fresh resolver must hit the persisted cache")
	}
}

func TestResolveDegradesToID(t *testing.T) {
	r := New(filepath.Join(t.TempDir(), "resolve.json"), &fakeLookup{fail: true})
	if got := r.Resolve("channel", "222222222222222222"); got != "222222222222222222" {
		t.Fatalf("Resolve on failure = %q, want the id back", got)
	}
}

func TestKindsAreNamespaced(t *testing.T) {
	r := New(filepath.Join(t.TempDir(), "resolve.json"), &fakeLookup{})
	u := r.Resolve("user", "111111111111111111")
	c := r.Resolve("channel", "111111111111111111")
	if u == c {
		t.Fatalf("kinds must not collide: %q vs %q", u, c)
	}
}

func TestIsSnowflake(t *testing.T) {
	cases := map[string]bool{
		"222222222222222222":    true,
		"12345678901234567":     true,
		"1234567890123456":      false, // too short
		"123456789012345678901": false, // too long
		"22222222222222222x":    false, // non-digit
		"general":               false,
		"":                      false,
	}
	for in, want := range cases {
		if got := IsSnowflake(in); got != want {
			t.Errorf("IsSnowflake(%q) = %v, want %v", in, got, want)
		}
	}
}
