package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/howar31/dscrd/internal/skillgen"
)

func TestGenerateSkillsWritesTree(t *testing.T) {
	dir := t.TempDir()
	out, err := runCmd(t, nil, "generate-skills", "--output-dir", dir)
	if err != nil {
		t.Fatalf("generate-skills: %v (%s)", err, out)
	}
	for _, p := range []string{
		"dscrd/SKILL.md",
		"dscrd-shared/SKILL.md",
		"dscrd-msg/SKILL.md",
		"dscrd-channel/SKILL.md",
		"dscrd-search/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
			t.Errorf("missing generated skill %s", p)
		}
	}
	idx, err := os.ReadFile(filepath.Join(dir, "dscrd/SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"dscrd api", "dscrd version", "dscrd-msg"} {
		if !strings.Contains(string(idx), want) {
			t.Errorf("index missing %q", want)
		}
	}
	msg, _ := os.ReadFile(filepath.Join(dir, "dscrd-msg/SKILL.md"))
	for _, want := range []string{
		"**Discord API:** `POST /channels/{channel.id}/messages`",
		"MESSAGE_CONTENT",
		"[!CAUTION]",
	} {
		if !strings.Contains(string(msg), want) {
			t.Errorf("msg skill missing %q", want)
		}
	}
}

func TestGenerateSkillsCleansStaleDirs(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, "dscrd-obsolete")
	if err := os.MkdirAll(stale, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stale, "SKILL.md"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A non-skill dir matching the prefix must survive.
	keep := filepath.Join(dir, "dscrd-notes")
	if err := os.MkdirAll(keep, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := runCmd(t, nil, "generate-skills", "--output-dir", dir); err != nil {
		t.Fatalf("generate-skills: %v", err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Error("stale generated skill dir was not removed")
	}
	if _, err := os.Stat(keep); err != nil {
		t.Error("non-skill dir was wrongly removed")
	}
}

// TestSkillsTreeInSync is the drift guard: the committed skills/ tree must
// match what the current command tree generates.
func TestSkillsTreeInSync(t *testing.T) {
	root := NewRootCommand(readVersionFile(t))
	files, err := skillgen.GenerateAll(root, readVersionFile(t))
	if err != nil {
		t.Fatal(err)
	}
	repoSkills := filepath.Join("..", "..", "skills")
	if _, err := os.Stat(repoSkills); err != nil {
		t.Skip("skills/ not generated yet")
	}
	for rel, want := range files {
		got, err := os.ReadFile(filepath.Join(repoSkills, rel))
		if err != nil {
			t.Errorf("skills/%s missing — run 'go run ./cmd/dscrd generate-skills'", rel)
			continue
		}
		if string(got) != want {
			t.Errorf("skills/%s is stale — run 'go run ./cmd/dscrd generate-skills'", rel)
		}
	}
}

func readVersionFile(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(data))
}
