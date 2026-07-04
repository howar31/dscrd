package commands

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVersionSubcommand(t *testing.T) {
	out, err := runCmd(t, nil, "version")
	if err != nil || strings.TrimSpace(out) != "dscrd test" {
		t.Fatalf("version: %v (%s)", err, out)
	}
}

func TestVersionCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/howar31/dscrd/releases/latest" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte(`{"tag_name":"v9.9.9"}`))
	}))
	defer srv.Close()
	t.Setenv("DSCRD_RELEASE_API", srv.URL)

	out, err := runCmd(t, nil, "version", "--check")
	if err != nil || !strings.Contains(out, "latest release: v9.9.9") {
		t.Fatalf("version --check: %v (%s)", err, out)
	}
}

func TestVersionCheckUpToDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"tag_name":"vtest"}`))
	}))
	defer srv.Close()
	t.Setenv("DSCRD_RELEASE_API", srv.URL)

	out, err := runCmd(t, nil, "version", "--check")
	if err != nil || !strings.Contains(out, "up to date") {
		t.Fatalf("version --check: %v (%s)", err, out)
	}
}
