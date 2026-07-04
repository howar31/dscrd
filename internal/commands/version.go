package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// releaseAPIBase is GitHub's API root; DSCRD_RELEASE_API overrides it in tests.
func releaseAPIBase() string {
	if v := os.Getenv("DSCRD_RELEASE_API"); v != "" {
		return v
	}
	return "https://api.github.com"
}

func newVersionCommand(g *GlobalFlags, version string) *cobra.Command {
	var check bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the dscrd version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "dscrd %s\n", version)
			if !check {
				return nil
			}
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get(releaseAPIBase() + "/repos/howar31/dscrd/releases/latest")
			if err != nil {
				return fmt.Errorf("release check failed: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("release check failed: HTTP %d", resp.StatusCode)
			}
			var rel struct {
				TagName string `json:"tag_name"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
				return err
			}
			latest := rel.TagName
			if latest == "v"+version || latest == version {
				fmt.Fprintln(cmd.OutOrStdout(), "up to date")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "latest release: %s\n", latest)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "check GitHub for the latest release")
	return cmd
}
