package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// meServer serves GET /users/@me with a fixture bot user.
func meServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/@me" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "444444444444444444", "username": "testbot"})
	}))
}

func TestAuthSetTokenAndStatusRoundTrip(t *testing.T) {
	srv := meServer(t)
	defer srv.Close()

	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	t.Setenv("DSCRD_CONFIG", cfgPath)
	t.Setenv("DSCRD_KEYRING_BACKEND", "file")
	t.Setenv("DSCRD_PROFILE", "")
	t.Setenv("DSCRD_TOKEN", "")
	t.Setenv("DSCRD_API_BASE", srv.URL)

	// set-token writes an encrypted profile.
	out, err := runCmdSharedEnv(t, "auth", "set-token",
		"--token", "secret-bot-token",
		"--application-id", "999999999999999999",
		"--default-guild", "111111111111111111")
	if err != nil {
		t.Fatalf("set-token: %v (%s)", err, out)
	}
	if !strings.Contains(out, `profile "default" saved`) {
		t.Fatalf("set-token out = %q", out)
	}
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "secret-bot-token") {
		t.Fatal("token stored in plaintext")
	}
	if !strings.Contains(string(raw), "enc:v1:") {
		t.Fatal("token not encrypted on disk")
	}

	// status lists the profile and live-checks the token.
	out, err = runCmdSharedEnv(t, "auth", "status")
	if err != nil {
		t.Fatalf("status: %v (%s)", err, out)
	}
	for _, want := range []string{"* default token:set", "app:999999999999999999", "token ok: testbot"} {
		if !strings.Contains(out, want) {
			t.Errorf("status out = %q, missing %q", out, want)
		}
	}
}

func TestAuthTest(t *testing.T) {
	srv := meServer(t)
	defer srv.Close()
	out, err := runCmd(t, srv, "auth", "test")
	if err != nil {
		t.Fatalf("auth test: %v (%s)", err, out)
	}
	if strings.TrimSpace(out) != "ok testbot (444444444444444444)" {
		t.Fatalf("auth test out = %q", out)
	}
}

func TestAuthSwitchAndLogout(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DSCRD_CONFIG", dir+"/config.toml")
	t.Setenv("DSCRD_KEYRING_BACKEND", "file")
	t.Setenv("DSCRD_PROFILE", "")
	t.Setenv("DSCRD_TOKEN", "")

	for _, name := range []string{"one", "two"} {
		if out, err := runCmdSharedEnv(t, "auth", "set-token", "--token", "tok-"+name, "--name", name); err != nil {
			t.Fatalf("set-token %s: %v (%s)", name, err, out)
		}
	}

	out, err := runCmdSharedEnv(t, "auth", "switch", "two")
	if err != nil || !strings.Contains(out, "active profile: two") {
		t.Fatalf("switch: %v (%s)", err, out)
	}
	if _, err := runCmdSharedEnv(t, "auth", "switch", "missing"); err == nil {
		t.Fatal("switch to missing profile must error")
	}

	out, err = runCmdSharedEnv(t, "auth", "logout", "--name", "one")
	if err != nil || !strings.Contains(out, `removed profile "one"`) {
		t.Fatalf("logout: %v (%s)", err, out)
	}
	// Removing the active profile clears active.
	out, err = runCmdSharedEnv(t, "auth", "logout")
	if err != nil || !strings.Contains(out, `removed profile "two"`) {
		t.Fatalf("logout active: %v (%s)", err, out)
	}
	if _, err := runCmdSharedEnv(t, "auth", "logout"); err == nil {
		t.Fatal("logout with nothing left must error")
	}
}

// authEnv isolates a config file and points the API base at srv (or at a dead
// address when srv is nil) so `auth status` never reaches the real API.
func authEnv(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	cfgPath := t.TempDir() + "/config.toml"
	t.Setenv("DSCRD_CONFIG", cfgPath)
	t.Setenv("DSCRD_KEYRING_BACKEND", "file")
	t.Setenv("DSCRD_PROFILE", "")
	t.Setenv("DSCRD_TOKEN", "")
	if srv != nil {
		t.Setenv("DSCRD_API_BASE", srv.URL)
	} else {
		t.Setenv("DSCRD_API_BASE", "http://127.0.0.1:0")
	}
	return cfgPath
}

func TestAuthRename(t *testing.T) {
	srv := meServer(t)
	defer srv.Close()
	authEnv(t, srv)

	for _, name := range []string{"one", "two"} {
		if out, err := runCmdSharedEnv(t, "auth", "set-token", "--token", "tok-"+name, "--name", name); err != nil {
			t.Fatalf("set-token %s: %v (%s)", name, err, out)
		}
	}
	if out, err := runCmdSharedEnv(t, "auth", "switch", "two"); err != nil {
		t.Fatalf("switch: %v (%s)", err, out)
	}

	// Renaming the active profile carries `active` with it.
	out, err := runCmdSharedEnv(t, "auth", "rename", "two", "two-new")
	if err != nil || !strings.Contains(out, `renamed profile "two" -> "two-new" (active: two-new)`) {
		t.Fatalf("rename active: %v (%s)", err, out)
	}
	// The stored token survives the rename and still resolves.
	if out, err := runCmdSharedEnv(t, "auth", "test", "--profile", "two-new"); err != nil {
		t.Fatalf("auth test after rename: %v (%s)", err, out)
	}

	// Renaming a non-active profile leaves `active` alone.
	out, err = runCmdSharedEnv(t, "auth", "rename", "one", "one-new")
	if err != nil || !strings.Contains(out, "(active: two-new)") {
		t.Fatalf("rename non-active: %v (%s)", err, out)
	}

	for _, tc := range []struct {
		name string
		args []string
	}{
		{"missing source", []string{"auth", "rename", "nope", "whatever"}},
		{"same name", []string{"auth", "rename", "one-new", "one-new"}},
		{"target taken", []string{"auth", "rename", "one-new", "two-new"}},
		{"invalid name", []string{"auth", "rename", "one-new", "bad name"}},
	} {
		if out, err := runCmdSharedEnv(t, tc.args...); err == nil {
			t.Errorf("%s must error, got %q", tc.name, out)
		}
	}

	// --force overwrites the target.
	out, err = runCmdSharedEnv(t, "auth", "rename", "one-new", "two-new", "--force")
	if err != nil || !strings.Contains(out, `renamed profile "one-new" -> "two-new"`) {
		t.Fatalf("rename --force: %v (%s)", err, out)
	}
	out, err = runCmdSharedEnv(t, "auth", "status")
	if err != nil {
		t.Fatalf("status: %v (%s)", err, out)
	}
	if strings.Contains(out, "one-new") {
		t.Errorf("source profile still present after rename: %q", out)
	}
}

func TestAuthSetTokenOverwriteProtection(t *testing.T) {
	authEnv(t, nil)

	if out, err := runCmdSharedEnv(t, "auth", "set-token", "--token", "tok", "--name", "one"); err != nil {
		t.Fatalf("set-token: %v (%s)", err, out)
	}
	if out, err := runCmdSharedEnv(t, "auth", "set-token", "--token", "other", "--name", "one"); err == nil {
		t.Fatalf("overwriting an existing profile must error, got %q", out)
	}
	if out, err := runCmdSharedEnv(t, "auth", "set-token", "--token", "other", "--name", "one", "--force"); err != nil {
		t.Fatalf("set-token --force: %v (%s)", err, out)
	}
	if out, err := runCmdSharedEnv(t, "auth", "set-token", "--token", "tok", "--name", "bad name"); err == nil {
		t.Fatalf("invalid profile name must error, got %q", out)
	}
}

func TestAuthSetTokenDefaultNameHint(t *testing.T) {
	authEnv(t, nil)

	// The first profile carries no hint: there is nothing to disambiguate yet.
	out, err := runCmdSharedEnv(t, "auth", "set-token", "--token", "tok", "--name", "one")
	if err != nil {
		t.Fatalf("set-token: %v (%s)", err, out)
	}
	if strings.Contains(out, "note: saved as") {
		t.Errorf("named profile must not be hinted: %q", out)
	}

	// Falling back to "default" while other profiles exist warns.
	out, err = runCmdSharedEnv(t, "auth", "set-token", "--token", "tok2")
	if err != nil {
		t.Fatalf("set-token default: %v (%s)", err, out)
	}
	if !strings.Contains(out, `note: saved as "default"`) {
		t.Errorf("unnamed profile must be hinted: %q", out)
	}
}

func TestAuthWriteCommandsHonorDryRun(t *testing.T) {
	cfgPath := authEnv(t, nil)

	// set-token --dry-run writes nothing and needs no token.
	out, err := runCmdSharedEnv(t, "auth", "set-token", "--name", "one", "--dry-run")
	if err != nil || !strings.Contains(out, `[dry-run] would save profile "one" (new)`) {
		t.Fatalf("set-token --dry-run: %v (%s)", err, out)
	}
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		t.Fatalf("set-token --dry-run wrote the config file")
	}

	for _, name := range []string{"one", "two"} {
		if out, err := runCmdSharedEnv(t, "auth", "set-token", "--token", "tok-"+name, "--name", name); err != nil {
			t.Fatalf("set-token %s: %v (%s)", name, err, out)
		}
	}
	before, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"switch", []string{"auth", "switch", "two", "--dry-run"}, `[dry-run] would set active profile to "two"`},
		{"logout", []string{"auth", "logout", "--name", "two", "--dry-run"}, `[dry-run] would remove profile "two"`},
		{"logout active", []string{"auth", "logout", "--dry-run"}, "active would become unset"},
		{"rename", []string{"auth", "rename", "two", "two-new", "--dry-run"}, `would rename profile "two" -> "two-new" (active unchanged)`},
		{"rename active", []string{"auth", "rename", "one", "one-new", "--dry-run"}, "active follows the rename"},
	} {
		out, err := runCmdSharedEnv(t, tc.args...)
		if err != nil || !strings.Contains(out, tc.want) {
			t.Errorf("%s --dry-run: %v (%s), want %q", tc.name, err, out, tc.want)
		}
	}

	after, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("a --dry-run command mutated the config file")
	}
}

func TestAuthStatusIsSorted(t *testing.T) {
	authEnv(t, nil)

	for _, name := range []string{"zeta", "alpha", "mid"} {
		if out, err := runCmdSharedEnv(t, "auth", "set-token", "--token", "tok", "--name", name); err != nil {
			t.Fatalf("set-token %s: %v (%s)", name, err, out)
		}
	}
	out, err := runCmdSharedEnv(t, "auth", "status")
	if err != nil {
		t.Fatalf("status: %v (%s)", err, out)
	}
	alpha, mid, zeta := strings.Index(out, "alpha"), strings.Index(out, "mid"), strings.Index(out, "zeta")
	if alpha < 0 || mid < 0 || zeta < 0 || !(alpha < mid && mid < zeta) {
		t.Fatalf("status not sorted by name: %q", out)
	}
}

func TestAuthSetTokenRequiresTokenWithoutTTY(t *testing.T) {
	if _, err := runCmd(t, nil, "auth", "set-token"); err == nil {
		t.Fatal("set-token without --token and without TTY must error")
	}
}

func TestAuthInviteURL(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DSCRD_CONFIG", dir+"/config.toml")
	t.Setenv("DSCRD_KEYRING_BACKEND", "file")
	t.Setenv("DSCRD_PROFILE", "")
	t.Setenv("DSCRD_TOKEN", "")

	if out, err := runCmdSharedEnv(t, "auth", "set-token", "--token", "tok", "--application-id", "999999999999999999"); err != nil {
		t.Fatalf("set-token: %v (%s)", err, out)
	}

	out, err := runCmdSharedEnv(t, "auth", "invite-url", "--permissions", "read")
	if err != nil {
		t.Fatalf("invite-url: %v (%s)", err, out)
	}
	want := "https://discord.com/oauth2/authorize?client_id=999999999999999999&scope=bot&permissions=66560"
	if strings.TrimSpace(out) != want {
		t.Fatalf("invite-url = %q, want %q", out, want)
	}

	// Raw bits accepted; junk rejected.
	if out, err := runCmdSharedEnv(t, "auth", "invite-url", "--permissions", "8"); err != nil || !strings.Contains(out, "permissions=8") {
		t.Fatalf("raw bits: %v (%s)", err, out)
	}
	if _, err := runCmdSharedEnv(t, "auth", "invite-url", "--permissions", "everything"); err == nil {
		t.Fatal("junk permission preset must error")
	}
}

func TestAuthInviteURLWithoutAppID(t *testing.T) {
	if _, err := runCmd(t, nil, "auth", "invite-url"); err == nil {
		t.Fatal("invite-url without application_id must error")
	}
}
