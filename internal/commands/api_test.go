package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/users/@me" || r.URL.Query().Get("probe") != "1" {
			t.Errorf("got %s %s", r.Method, r.URL)
		}
		w.Write([]byte(`{"id":"444444444444444444"}`))
	}))
	defer srv.Close()
	out, err := runCmd(t, srv, "api", "GET", "/users/@me", "--query", "probe=1")
	if err != nil || !strings.Contains(out, `"id":"444444444444444444"`) {
		t.Fatalf("api get: %v (%s)", err, out)
	}
}

func TestAPIPostBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var body map[string]any
		if err := json.Unmarshal(b, &body); err != nil || body["content"] != "hi" {
			t.Errorf("body = %s", b)
		}
		w.Write([]byte(`{"id":"333333333333333333"}`))
	}))
	defer srv.Close()
	out, err := runCmd(t, srv, "api", "post", "channels/2/messages", "--body", `{"content":"hi"}`)
	if err != nil || !strings.Contains(out, "333333333333333333") {
		t.Fatalf("api post: %v (%s)", err, out)
	}
}

func TestAPIRejectsBadInput(t *testing.T) {
	if _, err := runCmd(t, nil, "api", "BREW", "/coffee"); err == nil {
		t.Fatal("invalid method must error")
	}
	if _, err := runCmd(t, nil, "api", "POST", "/x", "--body", "{not json"); err == nil {
		t.Fatal("invalid JSON body must error")
	}
	if _, err := runCmd(t, nil, "api", "GET", "/x", "--query", "no-equals"); err == nil {
		t.Fatal("malformed query must error")
	}
}

func TestAPINoContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	out, err := runCmd(t, srv, "api", "DELETE", "/channels/2/messages/3")
	if err != nil || !strings.Contains(out, "ok (no content)") {
		t.Fatalf("api delete: %v (%s)", err, out)
	}
}

func TestAPIDryRun(t *testing.T) {
	out, err := runCmd(t, nil, "api", "POST", "/channels/2/messages", "--body", `{"content":"hi"}`, "--dry-run")
	if err != nil || !strings.Contains(out, "[dry-run] POST /channels/2/messages") {
		t.Fatalf("dry-run: %v (%s)", err, out)
	}
}
