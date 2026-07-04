package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newGuildCommand(g *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "guild", Short: "Inspect the servers the bot is in"}
	cmd.AddCommand(newGuildListCommand(g), newGuildInfoCommand(g))
	return cmd
}

// guildItem is one row of `guild list`.
type guildItem struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Owner bool   `json:"owner"`
}

func (i guildItem) Concise() string {
	s := fmt.Sprintf("%s (%s)", i.Name, i.ID)
	if i.Owner {
		s += " — owner"
	}
	return s
}

func newGuildListCommand(g *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List servers the bot has been invited to",
		Annotations: map[string]string{
			"discordEndpoint": "GET /users/@me/guilds",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(g)
			if err != nil {
				return err
			}
			var items []guildItem
			after := ""
			for {
				path := "/users/@me/guilds?limit=200"
				if after != "" {
					path += "&after=" + after
				}
				raw, err := client.Do("GET", path, nil)
				if err != nil {
					return err
				}
				if g.Raw {
					fmt.Fprintln(cmd.OutOrStdout(), string(raw))
					return nil
				}
				var page []guildItem
				if err := json.Unmarshal(raw, &page); err != nil {
					return err
				}
				items = append(items, page...)
				if len(page) < 200 {
					break
				}
				after = page[len(page)-1].ID
			}
			return emit(cmd, g, items)
		},
	}
}

func newGuildInfoCommand(g *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show server details (member counts, owner, description)",
		Annotations: map[string]string{
			"discordEndpoint": "GET /guilds/{guild.id}",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(g)
			if err != nil {
				return err
			}
			gid, err := guildID(g, client)
			if err != nil {
				return err
			}
			raw, err := client.Do("GET", "/guilds/"+gid+"?with_counts=true", nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var gd struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				OwnerID     string `json:"owner_id"`
				Description string `json:"description"`
				Members     int    `json:"approximate_member_count"`
				Online      int    `json:"approximate_presence_count"`
			}
			if err := json.Unmarshal(raw, &gd); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s (%s) — members:%d online:%d owner:%s\n",
				gd.Name, gd.ID, gd.Members, gd.Online, gd.OwnerID)
			if gd.Description != "" {
				fmt.Fprintln(cmd.OutOrStdout(), gd.Description)
			}
			return nil
		},
	}
}
