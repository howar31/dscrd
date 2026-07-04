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
