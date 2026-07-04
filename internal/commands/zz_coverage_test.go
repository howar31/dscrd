package commands

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestZZAllCommandsCovered is the coverage guarantee: every runnable leaf
// command in the tree must have been exercised by at least one test in this
// package (via runCmd or markCovered). The zz file name makes it run last.
func TestZZAllCommandsCovered(t *testing.T) {
	root := NewRootCommand("test")
	var missing []string
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		for _, sub := range c.Commands() {
			if sub.Hidden || sub.Name() == "help" || sub.Name() == "completion" {
				continue
			}
			if len(sub.Commands()) == 0 && sub.Runnable() {
				if !coveredCommands[sub.CommandPath()] {
					missing = append(missing, sub.CommandPath())
				}
			}
			walk(sub)
		}
	}
	walk(root)
	if len(missing) > 0 {
		t.Fatalf("commands with no test coverage (add a runCmd/markCovered test for each):\n%s",
			strings.Join(missing, "\n"))
	}
}
