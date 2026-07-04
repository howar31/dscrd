package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func searchServer(t *testing.T, warmupFirst bool) *httptest.Server {
	t.Helper()
	attempts := 0
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/@me/guilds", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{{"id": "111111111111111111", "name": "Test Server"}})
	})
	mux.HandleFunc("GET /guilds/111111111111111111/channels", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{{"id": "222222222222222222", "name": "general", "type": 0}})
	})
	mux.HandleFunc("GET /guilds/111111111111111111/messages/search", func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if warmupFirst && attempts == 1 {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(`{"code":110000,"retry_after":0.01}`))
			return
		}
		q := r.URL.Query()
		if q.Get("content") != "deploy" {
			t.Errorf("content = %q", q.Get("content"))
		}
		if q.Get("author_id") != "444444444444444444" {
			t.Errorf("author_id = %q", q.Get("author_id"))
		}
		if q.Get("channel_id") != "222222222222222222" {
			t.Errorf("channel_id = %q", q.Get("channel_id"))
		}
		json.NewEncoder(w).Encode(map[string]any{
			"total_results": 2,
			"messages": [][]map[string]any{
				{{"id": "333333333333333333", "content": "deploy done",
					"timestamp": "2026-07-01T12:00:00+00:00",
					"author":    map[string]any{"username": "alice", "global_name": "Alice"}}},
				{{"id": "333333333333333334", "content": "deploy started",
					"timestamp": "2026-07-01T11:00:00+00:00",
					"author":    map[string]any{"username": "bob", "global_name": "Bob"}}},
			},
		})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		http.NotFound(w, r)
	})
	return httptest.NewServer(mux)
}

func TestSearchMessages(t *testing.T) {
	srv := searchServer(t, false)
	defer srv.Close()
	out, err := runCmd(t, srv, "search", "messages", "--guild", "Test Server",
		"--content", "deploy", "--author", "444444444444444444", "--channel", "#general", "--no-resolve")
	if err != nil {
		t.Fatalf("search: %v (%s)", err, out)
	}
	for _, want := range []string{"Alice: deploy done", "Bob: deploy started", "total 2"} {
		if !strings.Contains(out, want) {
			t.Errorf("out = %q, missing %q", out, want)
		}
	}
}

func TestSearchMessagesIndexWarmup(t *testing.T) {
	srv := searchServer(t, true)
	defer srv.Close()
	out, err := runCmd(t, srv, "search", "messages", "--guild", "111111111111111111",
		"--content", "deploy", "--author", "444444444444444444", "--channel", "222222222222222222", "--no-resolve")
	if err != nil || !strings.Contains(out, "total 2") {
		t.Fatalf("warm-up retry: %v (%s)", err, out)
	}
}

func TestSearchMessagesEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/messages/search") {
			json.NewEncoder(w).Encode(map[string]any{"total_results": 0, "messages": [][]map[string]any{}})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	out, err := runCmd(t, srv, "search", "messages", "--guild", "111111111111111111", "--content", "nothing")
	if err != nil || !strings.Contains(out, "total 0") {
		t.Fatalf("empty search: %v (%s)", err, out)
	}
}
