package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func threadServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/@me/guilds", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{{"id": "111111111111111111", "name": "Test Server"}})
	})
	mux.HandleFunc("GET /guilds/111111111111111111/threads/active", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"threads": []map[string]any{
			{"id": "666666666666666666", "name": "bug-hunt", "parent_id": "222222222222222222"},
		}})
	})
	mux.HandleFunc("GET /channels/222222222222222222/threads/archived/public", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"threads": []map[string]any{
			{"id": "666666666666666667", "name": "old-topic", "parent_id": "222222222222222222",
				"thread_metadata": map[string]any{"archived": true}},
		}})
	})
	mux.HandleFunc("POST /channels/222222222222222222/messages/333333333333333333/threads", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"id": "666666666666666668", "name": "from-msg"})
	})
	mux.HandleFunc("POST /channels/222222222222222222/threads", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var body map[string]any
		json.Unmarshal(b, &body)
		if msg, ok := body["message"].(map[string]any); ok {
			if msg["content"] != "post body" {
				t.Errorf("forum message = %v", msg)
			}
		} else if body["type"] != float64(11) {
			t.Errorf("standalone thread type = %v", body["type"])
		}
		json.NewEncoder(w).Encode(map[string]any{"id": "666666666666666669", "name": body["name"]})
	})
	mux.HandleFunc("GET /channels/666666666666666666/messages", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "333333333333333333", "content": "in thread",
				"timestamp": "2026-07-02T10:00:00+00:00",
				"author":    map[string]any{"username": "alice", "global_name": "Alice"}},
		})
	})
	mux.HandleFunc("POST /channels/666666666666666666/messages", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"id": "333333333333333337"})
	})
	mux.HandleFunc("PATCH /channels/666666666666666666", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(b), `"archived":true`) {
			t.Errorf("archive body = %s", b)
		}
		w.Write([]byte(`{"id":"666666666666666666"}`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		http.NotFound(w, r)
	})
	return httptest.NewServer(mux)
}

func TestThreadList(t *testing.T) {
	srv := threadServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "thread", "list", "--guild", "111111111111111111")
	if err != nil || !strings.Contains(out, "bug-hunt (666666666666666666) — in 222222222222222222") {
		t.Fatalf("thread list: %v (%s)", err, out)
	}
}

func TestThreadListArchived(t *testing.T) {
	srv := threadServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "thread", "list", "--archived", "--channel", "222222222222222222")
	if err != nil || !strings.Contains(out, "old-topic") || !strings.Contains(out, "archived") {
		t.Fatalf("archived list: %v (%s)", err, out)
	}
	if _, err := runCmd(t, srv, "thread", "list", "--archived"); err == nil {
		t.Fatal("--archived without --channel must error")
	}
}

func TestThreadCreateFromMessage(t *testing.T) {
	srv := threadServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "thread", "create", "--channel", "222222222222222222",
		"--name", "from-msg", "--from-message", "333333333333333333")
	if err != nil || !strings.Contains(out, "created thread from-msg (666666666666666668)") {
		t.Fatalf("create from message: %v (%s)", err, out)
	}
}

func TestThreadCreateForumPost(t *testing.T) {
	srv := threadServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "thread", "create", "--channel", "222222222222222222",
		"--name", "post-title", "--text", "post body")
	if err != nil || !strings.Contains(out, "created thread post-title") {
		t.Fatalf("forum post: %v (%s)", err, out)
	}
}

func TestThreadCreateStandalone(t *testing.T) {
	srv := threadServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "thread", "create", "--channel", "222222222222222222", "--name", "standalone")
	if err != nil || !strings.Contains(out, "created thread standalone") {
		t.Fatalf("standalone: %v (%s)", err, out)
	}
}

func TestThreadRead(t *testing.T) {
	srv := threadServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "thread", "read", "--thread", "666666666666666666", "--no-resolve")
	if err != nil || !strings.Contains(out, "Alice: in thread") {
		t.Fatalf("thread read: %v (%s)", err, out)
	}
}

func TestThreadReply(t *testing.T) {
	srv := threadServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "thread", "reply", "--thread", "666666666666666666", "--text", "roger")
	if err != nil || !strings.Contains(out, "replied 333333333333333337") {
		t.Fatalf("thread reply: %v (%s)", err, out)
	}
}

func TestThreadArchive(t *testing.T) {
	srv := threadServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "thread", "archive", "--thread", "666666666666666666")
	if err != nil || !strings.Contains(out, "archived 666666666666666666") {
		t.Fatalf("archive: %v (%s)", err, out)
	}
}

func TestThreadArchiveDryRun(t *testing.T) {
	out, err := runCmd(t, nil, "thread", "archive", "--thread", "666666666666666666", "--dry-run")
	if err != nil || !strings.Contains(out, "[dry-run] PATCH /channels/666666666666666666") {
		t.Fatalf("dry-run: %v (%s)", err, out)
	}
}
