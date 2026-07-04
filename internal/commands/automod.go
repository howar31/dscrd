package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newAutomodCommand(g *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "automod", Short: "Manage auto-moderation rules"}
	cmd.AddCommand(
		newAutomodListCommand(g),
		newAutomodInfoCommand(g),
		newAutomodCreateCommand(g),
		newAutomodDeleteCommand(g),
	)
	return cmd
}

// automodItem is one rendered auto-moderation rule.
type automodItem struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

func (a automodItem) Concise() string {
	state := "disabled"
	if a.Enabled {
		state = "enabled"
	}
	return fmt.Sprintf("%s (%s) — %s", a.Name, a.ID, state)
}

func newAutomodListCommand(g *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List auto-moderation rules",
		Annotations: map[string]string{
			"discordEndpoint": "GET /guilds/{guild.id}/auto-moderation/rules",
			"perms":           "MANAGE_GUILD",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			raw, err := client.Do("GET", "/guilds/"+gid+"/auto-moderation/rules", nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var rules []automodItem
			if err := json.Unmarshal(raw, &rules); err != nil {
				return err
			}
			return emit(cmd, g, rules)
		},
	}
}

func newAutomodInfoCommand(g *GlobalFlags) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show one auto-moderation rule (raw JSON)",
		Annotations: map[string]string{
			"discordEndpoint": "GET /guilds/{guild.id}/auto-moderation/rules/{rule.id}",
			"perms":           "MANAGE_GUILD",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			raw, err := client.Do("GET", "/guilds/"+gid+"/auto-moderation/rules/"+id, nil)
			if err != nil {
				return err
			}
			// Rule structure is deeply nested; raw JSON is the useful form.
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "rule ID")
	cmd.MarkFlagRequired("id")
	return cmd
}

func newAutomodCreateCommand(g *GlobalFlags) *cobra.Command {
	var name string
	var keywords []string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a keyword-blocking rule",
		Annotations: map[string]string{
			"discordEndpoint": "POST /guilds/{guild.id}/auto-moderation/rules",
			"write":           "true",
			"perms":           "MANAGE_GUILD",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(keywords) == 0 {
				return fmt.Errorf("pass at least one --keyword")
			}
			body := map[string]any{
				"name":             name,
				"event_type":       1, // MESSAGE_SEND
				"trigger_type":     1, // KEYWORD
				"trigger_metadata": map[string]any{"keyword_filter": keywords},
				"actions":          []map[string]any{{"type": 1}}, // BLOCK_MESSAGE
				"enabled":          true,
			}
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			path := "/guilds/" + gid + "/auto-moderation/rules"
			if g.DryRun {
				dryRunPrint(cmd, "POST", path, body)
				return nil
			}
			raw, err := client.Do("POST", path, body)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var rule automodItem
			if err := json.Unmarshal(raw, &rule); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created rule %s (%s)\n", rule.Name, rule.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "rule name")
	cmd.Flags().StringArrayVar(&keywords, "keyword", nil, "keyword to block (repeatable, * wildcards allowed)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newAutomodDeleteCommand(g *GlobalFlags) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an auto-moderation rule",
		Annotations: map[string]string{
			"discordEndpoint": "DELETE /guilds/{guild.id}/auto-moderation/rules/{rule.id}",
			"write":           "true",
			"perms":           "MANAGE_GUILD",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			path := "/guilds/" + gid + "/auto-moderation/rules/" + id
			if g.DryRun {
				dryRunPrint(cmd, "DELETE", path, nil)
				return nil
			}
			if _, err := client.Do("DELETE", path, nil); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deleted rule %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "rule ID")
	cmd.MarkFlagRequired("id")
	return cmd
}
