package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// validMethods are the HTTP verbs the api escape hatch accepts.
var validMethods = map[string]bool{
	"GET": true, "POST": true, "PATCH": true, "PUT": true, "DELETE": true,
}

func newAPICommand(g *GlobalFlags) *cobra.Command {
	var body, bodyFile string
	var query []string
	cmd := &cobra.Command{
		Use:   "api <METHOD> <path>",
		Short: "Call any Discord REST endpoint (escape hatch)",
		Long: `Call any Discord REST endpoint directly, e.g.:
  dscrd api GET /users/@me
  dscrd api POST /channels/{id}/messages --body '{"content":"hi"}'
Output is always the raw JSON response.`,
		Args: cobra.ExactArgs(2),
		Annotations: map[string]string{
			"write": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(args[0])
			if !validMethods[method] {
				return fmt.Errorf("invalid method %q (want GET|POST|PATCH|PUT|DELETE)", args[0])
			}
			path := args[1]
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
			if len(query) > 0 {
				q := url.Values{}
				for _, kv := range query {
					k, v, ok := strings.Cut(kv, "=")
					if !ok {
						return fmt.Errorf("--query %q must be key=value", kv)
					}
					q.Add(k, v)
				}
				sep := "?"
				if strings.Contains(path, "?") {
					sep = "&"
				}
				path += sep + q.Encode()
			}

			var payload any
			raw := body
			if bodyFile != "" {
				data, err := os.ReadFile(bodyFile)
				if err != nil {
					return err
				}
				raw = string(data)
			}
			if raw != "" {
				if err := json.Unmarshal([]byte(raw), &payload); err != nil {
					return fmt.Errorf("--body is not valid JSON: %w", err)
				}
			}

			if g.DryRun {
				dryRunPrint(cmd, method, path, payload)
				return nil
			}
			client, err := buildClient(g)
			if err != nil {
				return err
			}
			resp, err := client.Do(method, path, payload)
			if err != nil {
				return err
			}
			if len(resp) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "ok (no content)")
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(resp))
			return nil
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "JSON request body")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "path to JSON body file")
	cmd.Flags().StringArrayVar(&query, "query", nil, "query parameter key=value (repeatable)")
	return cmd
}
