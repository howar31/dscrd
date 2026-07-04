package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newStickerCommand(g *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "sticker", Short: "Inspect the server's stickers"}
	cmd.AddCommand(newStickerListCommand(g), newStickerInfoCommand(g))
	return cmd
}

// stickerItem is one rendered sticker.
type stickerItem struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	Description string `json:"description"`
}

func (s stickerItem) Concise() string {
	out := fmt.Sprintf("%s (%s)", s.Name, s.ID)
	if s.Description != "" {
		out += " — " + s.Description
	}
	return out
}

func newStickerListCommand(g *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the server's stickers",
		Annotations: map[string]string{
			"discordEndpoint": "GET /guilds/{guild.id}/stickers",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			raw, err := client.Do("GET", "/guilds/"+gid+"/stickers", nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var stickers []stickerItem
			if err := json.Unmarshal(raw, &stickers); err != nil {
				return err
			}
			return emit(cmd, g, stickers)
		},
	}
}

func newStickerInfoCommand(g *GlobalFlags) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show one sticker",
		Annotations: map[string]string{
			"discordEndpoint": "GET /guilds/{guild.id}/stickers/{sticker.id}",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			raw, err := client.Do("GET", "/guilds/"+gid+"/stickers/"+id, nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var s stickerItem
			if err := json.Unmarshal(raw, &s); err != nil {
				return err
			}
			return emit(cmd, g, []stickerItem{s})
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "sticker ID")
	cmd.MarkFlagRequired("id")
	return cmd
}
