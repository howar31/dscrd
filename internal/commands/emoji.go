package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newEmojiCommand(g *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "emoji", Short: "Inspect custom emoji"}
	cmd.AddCommand(newEmojiListCommand(g), newEmojiInfoCommand(g))
	return cmd
}

// emojiItem is one rendered custom emoji.
type emojiItem struct {
	Name     string `json:"name"`
	ID       string `json:"id"`
	Animated bool   `json:"animated"`
}

func (e emojiItem) Concise() string {
	s := fmt.Sprintf(":%s: (%s)", e.Name, e.ID)
	if e.Animated {
		s += " animated"
	}
	return s
}

func newEmojiListCommand(g *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the server's custom emoji",
		Annotations: map[string]string{
			"discordEndpoint": "GET /guilds/{guild.id}/emojis",
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
			raw, err := client.Do("GET", "/guilds/"+gid+"/emojis", nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var emojis []emojiItem
			if err := json.Unmarshal(raw, &emojis); err != nil {
				return err
			}
			return emit(cmd, g, emojis)
		},
	}
}

func newEmojiInfoCommand(g *GlobalFlags) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show one custom emoji",
		Annotations: map[string]string{
			"discordEndpoint": "GET /guilds/{guild.id}/emojis/{emoji.id}",
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
			raw, err := client.Do("GET", "/guilds/"+gid+"/emojis/"+id, nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var e emojiItem
			if err := json.Unmarshal(raw, &e); err != nil {
				return err
			}
			return emit(cmd, g, []emojiItem{e})
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "emoji ID")
	cmd.MarkFlagRequired("id")
	return cmd
}
