package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func memberServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	member := map[string]any{
		"nick": "",
		"user": map[string]any{
			"id": "444444444444444444", "username": "alice", "global_name": "Alice",
		},
		"roles":     []string{"555555555555555555"},
		"joined_at": "2026-01-01T00:00:00+00:00",
	}
	mux.HandleFunc("GET /guilds/111111111111111111/members", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") == "" {
			t.Error("missing limit")
		}
		json.NewEncoder(w).Encode([]any{member})
	})
	mux.HandleFunc("GET /guilds/111111111111111111/members/444444444444444444", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(member)
	})
	mux.HandleFunc("GET /guilds/111111111111111111/members/search", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query") != "ali" {
			t.Errorf("query = %q", r.URL.Query().Get("query"))
		}
		json.NewEncoder(w).Encode([]any{member})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		http.NotFound(w, r)
	})
	return httptest.NewServer(mux)
}

func TestMemberList(t *testing.T) {
	srv := memberServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "member", "list", "--guild", "111111111111111111")
	if err != nil || !strings.Contains(out, "Alice (444444444444444444) — roles: 555555555555555555") {
		t.Fatalf("member list: %v (%s)", err, out)
	}
}

func TestMemberInfo(t *testing.T) {
	srv := memberServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "member", "info", "--guild", "111111111111111111", "--user", "444444444444444444")
	if err != nil || !strings.Contains(out, "Alice (444444444444444444)") || !strings.Contains(out, "joined:2026-01-01") {
		t.Fatalf("member info: %v (%s)", err, out)
	}
}

func TestMemberSearch(t *testing.T) {
	srv := memberServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "member", "search", "--guild", "111111111111111111", "--query", "ali")
	if err != nil || !strings.Contains(out, "Alice") {
		t.Fatalf("member search: %v (%s)", err, out)
	}
}
