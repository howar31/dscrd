package commands

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// allowedCDNHosts are the only hosts `file download` will fetch from.
var allowedCDNHosts = map[string]bool{
	"cdn.discordapp.com":   true,
	"media.discordapp.net": true,
}

func newFileCommand(g *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "file", Short: "Download message attachments"}
	cmd.AddCommand(newFileDownloadCommand(g))
	return cmd
}

func newFileDownloadCommand(g *GlobalFlags) *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "download <attachment-url>",
		Short: "Download an attachment from the Discord CDN",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u, err := url.Parse(args[0])
			if err != nil {
				return err
			}
			// DSCRD_CDN_ALLOW_HOST admits one extra host (tests).
			if !allowedCDNHosts[u.Host] && u.Host != os.Getenv("DSCRD_CDN_ALLOW_HOST") {
				return fmt.Errorf("refusing to download from %q (only Discord CDN hosts)", u.Host)
			}
			target := out
			if target == "" {
				target = path.Base(u.Path)
			}
			if target == "" || target == "/" || target == "." {
				return fmt.Errorf("cannot derive a file name from %q; pass --out", args[0])
			}
			if g.DryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] GET %s -> %s\n", args[0], target)
				return nil
			}
			client := &http.Client{Timeout: 5 * time.Minute}
			resp, err := client.Get(args[0])
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
			}
			f, err := os.Create(target)
			if err != nil {
				return err
			}
			n, err := io.Copy(f, resp.Body)
			if cerr := f.Close(); err == nil {
				err = cerr
			}
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "saved %s (%d bytes)\n", strings.TrimPrefix(target, "./"), n)
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "output path (default: attachment file name)")
	return cmd
}
