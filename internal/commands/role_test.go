package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func roleServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /guilds/111111111111111111/roles", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "555555555555555555", "name": "moderator"},
			{"id": "555555555555555556", "name": "member"},
		})
	})
	mux.HandleFunc("POST /guilds/111111111111111111/roles", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var body map[string]any
		json.Unmarshal(b, &body)
		if body["name"] != "helper" {
			t.Errorf("create body = %v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{"id": "555555555555555557", "name": "helper"})
	})
	mux.HandleFunc("PATCH /guilds/111111111111111111/roles/555555555555555555", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":"555555555555555555"}`))
	})
	mux.HandleFunc("DELETE /guilds/111111111111111111/roles/555555555555555555", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("PUT /guilds/111111111111111111/members/444444444444444444/roles/555555555555555555", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("DELETE /guilds/111111111111111111/members/444444444444444444/roles/555555555555555555", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		http.NotFound(w, r)
	})
	return httptest.NewServer(mux)
}

func TestRoleList(t *testing.T) {
	srv := roleServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "role", "list", "--guild", "111111111111111111")
	if err != nil || !strings.Contains(out, "moderator (555555555555555555)") {
		t.Fatalf("role list: %v (%s)", err, out)
	}
}

func TestRoleCreate(t *testing.T) {
	srv := roleServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "role", "create", "--guild", "111111111111111111", "--name", "helper")
	if err != nil || !strings.Contains(out, "created role helper (555555555555555557)") {
		t.Fatalf("role create: %v (%s)", err, out)
	}
}

func TestRoleEdit(t *testing.T) {
	srv := roleServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "role", "edit", "--guild", "111111111111111111",
		"--role", "555555555555555555", "--name", "mods")
	if err != nil || !strings.Contains(out, "edited role 555555555555555555") {
		t.Fatalf("role edit: %v (%s)", err, out)
	}
	if _, err := runCmd(t, srv, "role", "edit", "--guild", "111111111111111111", "--role", "555555555555555555"); err == nil {
		t.Fatal("edit with no changes must error")
	}
}

func TestRoleDelete(t *testing.T) {
	srv := roleServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "role", "delete", "--guild", "111111111111111111", "--role", "555555555555555555")
	if err != nil || !strings.Contains(out, "deleted role 555555555555555555") {
		t.Fatalf("role delete: %v (%s)", err, out)
	}
}

func TestRoleAssignUnassign(t *testing.T) {
	srv := roleServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "role", "assign", "--guild", "111111111111111111",
		"--role", "555555555555555555", "--user", "444444444444444444")
	if err != nil || !strings.Contains(out, "assigned role") {
		t.Fatalf("assign: %v (%s)", err, out)
	}
	out, err = runCmd(t, srv, "role", "unassign", "--guild", "111111111111111111",
		"--role", "555555555555555555", "--user", "444444444444444444")
	if err != nil || !strings.Contains(out, "removed role") {
		t.Fatalf("unassign: %v (%s)", err, out)
	}
}

func TestRoleDeleteDryRun(t *testing.T) {
	// With a snowflake --guild, dry-run must work without any credentials.
	out, err := runCmd(t, nil, "role", "delete", "--guild", "111111111111111111",
		"--role", "555555555555555555", "--dry-run")
	if err != nil || !strings.Contains(out, "[dry-run] DELETE /guilds/111111111111111111/roles/555555555555555555") {
		t.Fatalf("dry-run: %v (%s)", err, out)
	}
}
