package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGuildList(t *testing.T) {
	srv := guildListServer(t)
	defer srv.Close()

	out, err := runCmd(t, srv, "guild", "list")
	if err != nil {
		t.Fatalf("guild list: %v (%s)", err, out)
	}
	want := "Test Server (111111111111111111) — owner\nOther Server (111111111111111112)\n"
	if out != want {
		t.Fatalf("out = %q, want %q", out, want)
	}
}

func TestGuildListPagination(t *testing.T) {
	pages := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pages++
		after := r.URL.Query().Get("after")
		if after == "" {
			// Full page of 200 forces a second fetch.
			page := make([]map[string]any, 200)
			for i := range page {
				page[i] = map[string]any{"id": "100000000000000000", "name": "Bulk"}
			}
			page[199]["id"] = "100000000000000199"
			json.NewEncoder(w).Encode(page)
			return
		}
		if after != "100000000000000199" {
			t.Errorf("after = %q", after)
		}
		json.NewEncoder(w).Encode([]map[string]any{{"id": "111111111111111111", "name": "Tail"}})
	}))
	defer srv.Close()

	out, err := runCmd(t, srv, "guild", "list")
	if err != nil {
		t.Fatalf("guild list: %v", err)
	}
	if pages != 2 || !strings.Contains(out, "Tail") {
		t.Fatalf("pages = %d out = %q", pages, out)
	}
}

func TestGuildInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/@me/guilds":
			json.NewEncoder(w).Encode([]map[string]any{{"id": "111111111111111111", "name": "Test Server"}})
		case "/guilds/111111111111111111":
			if r.URL.Query().Get("with_counts") != "true" {
				t.Error("missing with_counts=true")
			}
			json.NewEncoder(w).Encode(map[string]any{
				"id": "111111111111111111", "name": "Test Server", "owner_id": "444444444444444444",
				"approximate_member_count": 42, "approximate_presence_count": 7,
				"description": "fixture server",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	// Resolve by name to exercise the name->ID path end to end.
	out, err := runCmd(t, srv, "guild", "info", "--guild", "Test Server")
	if err != nil {
		t.Fatalf("guild info: %v (%s)", err, out)
	}
	for _, want := range []string{"Test Server (111111111111111111)", "members:42", "online:7", "fixture server"} {
		if !strings.Contains(out, want) {
			t.Errorf("out = %q, missing %q", out, want)
		}
	}
}
