package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoForeignBranding keeps the project's text self-contained: no source,
// doc, or generated file may reference other chat platforms or their tooling.
// The tokens are assembled at runtime so this file does not trip its own scan.
func TestNoForeignBranding(t *testing.T) {
	banned := []string{"s" + "lk", "sla" + "ck"}
	repoRoot := filepath.Join("..", "..")
	exts := map[string]bool{".go": true, ".md": true, ".tmpl": true, ".yml": true, ".yaml": true, ".json": true, ".sh": true, ".js": true}

	err := filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "dist", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if !exts[filepath.Ext(path)] {
			return nil
		}
		if filepath.Base(path) == "branding_test.go" {
			return nil // the scanner itself
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lower := strings.ToLower(string(data))
		for _, token := range banned {
			if strings.Contains(lower, token) {
				t.Errorf("%s contains banned token %q", path, token)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
