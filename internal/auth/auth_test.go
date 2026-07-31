package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeKeyring is an in-memory keyringStore for tests.
type fakeKeyring struct {
	data map[string]string
	fail bool
}

func (f *fakeKeyring) key(s, u string) string { return s + "/" + u }
func (f *fakeKeyring) Get(s, u string) (string, error) {
	if f.fail {
		return "", errKeyringUnavailable
	}
	v, ok := f.data[f.key(s, u)]
	if !ok {
		return "", errKeyNotFound
	}
	return v, nil
}
func (f *fakeKeyring) Set(s, u, p string) error {
	if f.fail {
		return errKeyringUnavailable
	}
	f.data[f.key(s, u)] = p
	return nil
}
func (f *fakeKeyring) Delete(s, u string) error {
	if f.fail {
		return errKeyringUnavailable
	}
	delete(f.data, f.key(s, u))
	return nil
}

func useFakeKeyring(t *testing.T, fail bool) *fakeKeyring {
	t.Helper()
	fake := &fakeKeyring{data: map[string]string{}, fail: fail}
	old := activeKeyring
	activeKeyring = fake
	t.Cleanup(func() { activeKeyring = old })
	return fake
}

func TestCryptoRoundTrip(t *testing.T) {
	key, err := newKey()
	if err != nil {
		t.Fatal(err)
	}
	ct, err := encryptValue(key, "secret-token")
	if err != nil {
		t.Fatal(err)
	}
	if !isEncrypted(ct) || strings.Contains(ct, "secret-token") {
		t.Fatalf("ciphertext %q leaks plaintext or lacks prefix", ct)
	}
	pt, err := decryptValue(key, ct)
	if err != nil || pt != "secret-token" {
		t.Fatalf("decrypt = %q, %v", pt, err)
	}
}

func TestCryptoTamperDetected(t *testing.T) {
	key, _ := newKey()
	ct, _ := encryptValue(key, "secret")
	tampered := ct[:len(ct)-2] + "AA"
	if _, err := decryptValue(key, tampered); err == nil {
		t.Fatal("tampered ciphertext must not decrypt")
	}
	otherKey, _ := newKey()
	if _, err := decryptValue(otherKey, ct); err == nil {
		t.Fatal("wrong key must not decrypt")
	}
}

func TestResolveBackendPrecedence(t *testing.T) {
	useFakeKeyring(t, false)

	// Config-recorded backend wins over everything.
	if got := resolveBackend(&Config{KeyBackend: "file"}); got != "file" {
		t.Fatalf("config backend: got %q", got)
	}
	// Env selects when config silent.
	t.Setenv(keyEnvVar, "file")
	if got := resolveBackend(&Config{}); got != "file" {
		t.Fatalf("env backend: got %q", got)
	}
	// Auto probes the keyring.
	t.Setenv(keyEnvVar, "")
	if got := resolveBackend(&Config{}); got != "keyring" {
		t.Fatalf("auto with working keyring: got %q", got)
	}
}

func TestResolveBackendAutoFallsBackToFile(t *testing.T) {
	useFakeKeyring(t, true)
	t.Setenv(keyEnvVar, "")
	if got := resolveBackend(&Config{}); got != "file" {
		t.Fatalf("auto with broken keyring: got %q", got)
	}
}

func TestStoreRoundTripFileBackend(t *testing.T) {
	t.Setenv(keyEnvVar, "file")
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	cfg := &Config{
		Active: "main",
		Profiles: map[string]Profile{
			"main": {Token: "bot-token-123", ApplicationID: "999999999999999999", DefaultGuild: "111111111111111111"},
		},
	}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// On-disk file must be 0600 and must not contain the plaintext token.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("config perms = %o, want 600", perm)
	}
	rawFile, _ := os.ReadFile(path)
	if strings.Contains(string(rawFile), "bot-token-123") {
		t.Fatal("plaintext token found on disk")
	}
	if !strings.Contains(string(rawFile), encPrefix) {
		t.Fatal("token not encrypted on disk")
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	p := loaded.Profiles["main"]
	if p.Token != "bot-token-123" || p.ApplicationID != "999999999999999999" || p.DefaultGuild != "111111111111111111" {
		t.Fatalf("loaded profile = %+v", p)
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil || cfg == nil || cfg.Profiles == nil {
		t.Fatalf("Load missing = %+v, %v", cfg, err)
	}
}

func TestResolveTokenPrecedence(t *testing.T) {
	cfg := &Config{
		Active: "main",
		Profiles: map[string]Profile{
			"main":  {Token: "active-token"},
			"other": {Token: "other-token"},
		},
	}

	// Env wins.
	if tok, _ := ResolveToken(cfg, "other", "env-token"); tok != "env-token" {
		t.Fatalf("env precedence: got %q", tok)
	}
	// Named profile beats active.
	if tok, _ := ResolveToken(cfg, "other", ""); tok != "other-token" {
		t.Fatalf("named profile: got %q", tok)
	}
	// Active profile is the default.
	if tok, _ := ResolveToken(cfg, "", ""); tok != "active-token" {
		t.Fatalf("active profile: got %q", tok)
	}
}

func TestResolveTokenErrors(t *testing.T) {
	if _, err := ResolveToken(&Config{Profiles: map[string]Profile{}}, "", ""); err == nil {
		t.Fatal("no profile selected must error")
	}
	if _, err := ResolveToken(&Config{Active: "gone", Profiles: map[string]Profile{}}, "", ""); err == nil {
		t.Fatal("missing profile must error")
	}
	cfg := &Config{Active: "m", Profiles: map[string]Profile{"m": {Token: encPrefix + "AAAA"}}}
	_, err := ResolveToken(cfg, "", "")
	if err == nil || !strings.Contains(err.Error(), "decrypted") {
		t.Fatalf("undecryptable token: err = %v", err)
	}
	var authErr *AuthError
	if !errorsAs(err, &authErr) || authErr.ExitCode() != 3 {
		t.Fatalf("want AuthError exit 3, got %v", err)
	}
}

// errorsAs avoids importing errors just for one call site.
func errorsAs(err error, target **AuthError) bool {
	e, ok := err.(*AuthError)
	if ok {
		*target = e
	}
	return ok
}

func TestActiveProfile(t *testing.T) {
	cfg := &Config{
		Active: "main",
		Profiles: map[string]Profile{
			"main":  {ApplicationID: "999999999999999999"},
			"other": {ApplicationID: "888888888888888888"},
		},
	}
	p, name, ok := ActiveProfile(cfg, "")
	if !ok || name != "main" || p.ApplicationID != "999999999999999999" {
		t.Fatalf("active: %v %q %+v", ok, name, p)
	}
	p, name, ok = ActiveProfile(cfg, "other")
	if !ok || name != "other" || p.ApplicationID != "888888888888888888" {
		t.Fatalf("named: %v %q %+v", ok, name, p)
	}
	if _, _, ok := ActiveProfile(&Config{Profiles: map[string]Profile{}}, ""); ok {
		t.Fatal("no active profile must report !ok")
	}
}

func TestValidateProfileName(t *testing.T) {
	for _, name := range []string{"default", "Alice", "bot-1", "bot_1", "A"} {
		if err := ValidateProfileName(name); err != nil {
			t.Errorf("ValidateProfileName(%q) = %v, want nil", name, err)
		}
	}
	for _, name := range []string{"", "bad name", "with.dot", "quote\"d", "café", "a/b"} {
		if err := ValidateProfileName(name); err == nil {
			t.Errorf("ValidateProfileName(%q) = nil, want error", name)
		}
	}
}
