package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// channelServer serves a guild with two channels plus create/edit/delete routes.
func channelServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/users/@me/guilds":
			json.NewEncoder(w).Encode([]map[string]any{{"id": "111111111111111111", "name": "Test Server"}})
		case r.URL.Path == "/guilds/111111111111111111/channels" && r.Method == "GET":
			json.NewEncoder(w).Encode([]map[string]any{
				{"id": "222222222222222222", "name": "general", "type": 0, "topic": "chat"},
				{"id": "222222222222222223", "name": "ideas", "type": 15},
			})
		case r.URL.Path == "/guilds/111111111111111111/channels" && r.Method == "POST":
			b, _ := io.ReadAll(r.Body)
			var body map[string]any
			json.Unmarshal(b, &body)
			if body["name"] != "new-chan" || body["type"] != float64(0) {
				t.Errorf("create body = %v", body)
			}
			json.NewEncoder(w).Encode(map[string]any{"id": "222222222222222224", "name": "new-chan", "type": 0})
		case r.URL.Path == "/channels/222222222222222222" && r.Method == "GET":
			json.NewEncoder(w).Encode(map[string]any{"id": "222222222222222222", "name": "general", "type": 0, "topic": "chat"})
		case r.URL.Path == "/channels/222222222222222222" && r.Method == "PATCH":
			b, _ := io.ReadAll(r.Body)
			w.Header().Set("X-Body", string(b))
			w.Write([]byte(`{"id":"222222222222222222"}`))
		case r.URL.Path == "/channels/222222222222222222" && r.Method == "DELETE":
			w.Write([]byte(`{"id":"222222222222222222"}`))
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestChannelList(t *testing.T) {
	srv := channelServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "channel", "list", "--guild", "111111111111111111")
	if err != nil {
		t.Fatalf("channel list: %v (%s)", err, out)
	}
	want := "#general (222222222222222222) — text: chat\n#ideas (222222222222222223) — forum\n"
	if out != want {
		t.Fatalf("out = %q, want %q", out, want)
	}
}

func TestChannelInfoByName(t *testing.T) {
	srv := channelServer(t)
	defer srv.Close()
	// "#general" resolves through the guild channel list.
	out, err := runCmd(t, srv, "channel", "info", "--guild", "Test Server", "--channel", "#general")
	if err != nil {
		t.Fatalf("channel info: %v (%s)", err, out)
	}
	if !strings.Contains(out, "#general (222222222222222222)") {
		t.Fatalf("out = %q", out)
	}
}

func TestChannelCreate(t *testing.T) {
	srv := channelServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "channel", "create", "--guild", "111111111111111111", "--name", "new-chan")
	if err != nil {
		t.Fatalf("channel create: %v (%s)", err, out)
	}
	if strings.TrimSpace(out) != "created #new-chan (222222222222222224)" {
		t.Fatalf("out = %q", out)
	}
}

func TestChannelCreateRejectsUnknownType(t *testing.T) {
	if _, err := runCmd(t, nil, "channel", "create", "--guild", "1", "--name", "x", "--type", "carrier-pigeon"); err == nil {
		t.Fatal("unknown type must error")
	}
}

func TestChannelEditDryRun(t *testing.T) {
	out, err := runCmd(t, nil, "channel", "edit", "--channel", "222222222222222222", "--topic", "new topic", "--dry-run")
	if err != nil {
		t.Fatalf("dry-run: %v (%s)", err, out)
	}
	if !strings.Contains(out, "[dry-run] PATCH /channels/222222222222222222") || !strings.Contains(out, "new topic") {
		t.Fatalf("out = %q", out)
	}
}

func TestChannelEdit(t *testing.T) {
	srv := channelServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "channel", "edit", "--channel", "222222222222222222", "--name", "renamed")
	if err != nil || !strings.Contains(out, "edited 222222222222222222") {
		t.Fatalf("edit: %v (%s)", err, out)
	}
	if _, err := runCmd(t, srv, "channel", "edit", "--channel", "222222222222222222"); err == nil {
		t.Fatal("edit with no changes must error")
	}
}

func TestChannelDelete(t *testing.T) {
	srv := channelServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "channel", "delete", "--channel", "222222222222222222")
	if err != nil || !strings.Contains(out, "deleted 222222222222222222") {
		t.Fatalf("delete: %v (%s)", err, out)
	}
}

func TestChannelDeleteDryRun(t *testing.T) {
	out, err := runCmd(t, nil, "channel", "delete", "--channel", "222222222222222222", "--dry-run")
	if err != nil || !strings.Contains(out, "[dry-run] DELETE /channels/222222222222222222") {
		t.Fatalf("dry-run: %v (%s)", err, out)
	}
}

func TestChannelTopic(t *testing.T) {
	srv := channelServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "channel", "topic", "--channel", "222222222222222222", "--topic", "hello")
	if err != nil || !strings.Contains(out, "topic set on 222222222222222222") {
		t.Fatalf("topic: %v (%s)", err, out)
	}
}
