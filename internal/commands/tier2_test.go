package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestEmojiListAndInfo(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /guilds/111111111111111111/emojis", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "777777777777777777", "name": "partyparrot", "animated": true},
		})
	})
	mux.HandleFunc("GET /guilds/111111111111111111/emojis/777777777777777777", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"id": "777777777777777777", "name": "partyparrot", "animated": true})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	out, err := runCmd(t, srv, "emoji", "list", "--guild", "111111111111111111")
	if err != nil || !strings.Contains(out, ":partyparrot: (777777777777777777) animated") {
		t.Fatalf("emoji list: %v (%s)", err, out)
	}
	out, err = runCmd(t, srv, "emoji", "info", "--guild", "111111111111111111", "--id", "777777777777777777")
	if err != nil || !strings.Contains(out, "partyparrot") {
		t.Fatalf("emoji info: %v (%s)", err, out)
	}
}

func TestFileDownload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/attachments/1/2/report.txt" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("attachment content"))
	}))
	defer srv.Close()
	u, _ := url.Parse(srv.URL)
	t.Setenv("DSCRD_CDN_ALLOW_HOST", u.Host)

	dir := t.TempDir()
	outPath := dir + "/saved.txt"
	out, err := runCmd(t, nil, "file", "download", srv.URL+"/attachments/1/2/report.txt", "--out", outPath)
	if err != nil || !strings.Contains(out, "saved") {
		t.Fatalf("download: %v (%s)", err, out)
	}
	data, err := os.ReadFile(outPath)
	if err != nil || string(data) != "attachment content" {
		t.Fatalf("saved file = %q, %v", data, err)
	}
}

func TestFileDownloadRefusesForeignHosts(t *testing.T) {
	if _, err := runCmd(t, nil, "file", "download", "https://example.com/x.txt"); err == nil {
		t.Fatal("non-CDN host must be refused")
	}
}

func TestDMSend(t *testing.T) {
	var sequence []string
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users/@me/channels", func(w http.ResponseWriter, r *http.Request) {
		sequence = append(sequence, "open")
		b, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(b), "444444444444444444") {
			t.Errorf("open body = %s", b)
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "888888888888888888"})
	})
	mux.HandleFunc("POST /channels/888888888888888888/messages", func(w http.ResponseWriter, r *http.Request) {
		sequence = append(sequence, "send")
		json.NewEncoder(w).Encode(map[string]string{"id": "333333333333333338"})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	out, err := runCmd(t, srv, "dm", "send", "--user", "444444444444444444", "--text", "hello there")
	if err != nil || strings.TrimSpace(out) != "sent 333333333333333338 (dm 888888888888888888)" {
		t.Fatalf("dm send: %v (%s)", err, out)
	}
	if strings.Join(sequence, ",") != "open,send" {
		t.Fatalf("sequence = %v", sequence)
	}
}

func TestDMRead(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users/@me/channels", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"id": "888888888888888888"})
	})
	mux.HandleFunc("GET /channels/888888888888888888/messages", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "333333333333333333", "content": "hi bot",
				"timestamp": "2026-07-01T00:00:00+00:00",
				"author":    map[string]any{"username": "alice", "global_name": "Alice"}},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	out, err := runCmd(t, srv, "dm", "read", "--user", "444444444444444444", "--no-resolve")
	if err != nil || !strings.Contains(out, "Alice: hi bot") {
		t.Fatalf("dm read: %v (%s)", err, out)
	}
}

func TestDMSendDryRun(t *testing.T) {
	out, err := runCmd(t, nil, "dm", "send", "--user", "444444444444444444", "--text", "hi", "--dry-run")
	if err != nil || !strings.Contains(out, "[dry-run] DM to 444444444444444444") {
		t.Fatalf("dry-run: %v (%s)", err, out)
	}
}

func TestUserInfoAndMe(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/444444444444444444", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"id": "444444444444444444", "username": "alice", "global_name": "Alice",
		})
	})
	mux.HandleFunc("GET /users/@me", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"id": "999999999999999990", "username": "testbot", "bot": true,
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	out, err := runCmd(t, srv, "user", "info", "--id", "444444444444444444")
	if err != nil || !strings.Contains(out, "Alice (444444444444444444) @alice") {
		t.Fatalf("user info: %v (%s)", err, out)
	}
	out, err = runCmd(t, srv, "user", "me")
	if err != nil || !strings.Contains(out, "testbot (999999999999999990) [bot]") {
		t.Fatalf("user me: %v (%s)", err, out)
	}
}
