package commands

import (
	"bytes"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/howar31/dscrd/internal/api"
)

// testAPIClient returns an api.Client pointed at srv.
func testAPIClient(srv *httptest.Server) *api.Client {
	c := api.New("test-token")
	c.BaseURL = srv.URL
	return c
}

// coveredCommands records every leaf command exercised through runCmd or
// markCovered; zz_coverage_test.go fails the build when a leaf is missing.
var coveredCommands = map[string]bool{}

// runCmd executes a fresh root command with args against srv (when non-nil),
// records coverage for the resolved leaf command, and returns combined output.
// The environment is fully isolated: temp config, file keyring, no profile.
func runCmd(t *testing.T, srv *httptest.Server, args ...string) (string, error) {
	t.Helper()
	t.Setenv("DSCRD_CONFIG", filepath.Join(t.TempDir(), "config.toml"))
	t.Setenv("DSCRD_KEYRING_BACKEND", "file")
	t.Setenv("DSCRD_PROFILE", "")
	if srv != nil {
		t.Setenv("DSCRD_API_BASE", srv.URL)
		t.Setenv("DSCRD_TOKEN", "test-token")
	} else {
		t.Setenv("DSCRD_API_BASE", "http://127.0.0.1:0")
		t.Setenv("DSCRD_TOKEN", "")
	}

	return runCmdSharedEnv(t, args...)
}

// runCmdSharedEnv executes args against whatever environment the caller has
// already set (config path, API base), still recording leaf coverage. Use it
// for multi-invocation tests that share one config file.
func runCmdSharedEnv(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := NewRootCommand("test")
	if target, _, err := root.Find(args); err == nil && target != nil && target.Runnable() {
		coveredCommands[target.CommandPath()] = true
	}
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

// markCovered registers coverage for a leaf a test exercises without going
// through runCmd (e.g. pure flag-registration checks).
func markCovered(t *testing.T, path string) {
	t.Helper()
	coveredCommands[path] = true
}
