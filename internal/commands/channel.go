package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// channelTypeNames maps Discord channel type codes to readable names.
var channelTypeNames = map[int]string{
	0:  "text",
	2:  "voice",
	4:  "category",
	5:  "announcement",
	10: "announcement-thread",
	11: "thread",
	12: "private-thread",
	13: "stage",
	15: "forum",
	16: "media",
}

// channelTypeCodes is the reverse mapping used by `channel create --type`.
var channelTypeCodes = map[string]int{
	"text":         0,
	"voice":        2,
	"category":     4,
	"announcement": 5,
	"stage":        13,
	"forum":        15,
	"media":        16,
}

func channelTypeName(code int) string {
	if n, ok := channelTypeNames[code]; ok {
		return n
	}
	return fmt.Sprintf("type-%d", code)
}

func newChannelCommand(g *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "channel", Short: "Manage channels"}
	cmd.AddCommand(
		newChannelListCommand(g),
		newChannelInfoCommand(g),
		newChannelCreateCommand(g),
		newChannelEditCommand(g),
		newChannelDeleteCommand(g),
		newChannelTopicCommand(g),
	)
	return cmd
}

// channelItem is one row of `channel list`.
type channelItem struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Type  string `json:"type"`
	Topic string `json:"topic"`
}

func (c channelItem) Concise() string {
	s := fmt.Sprintf("#%s (%s) — %s", c.Name, c.ID, c.Type)
	if c.Topic != "" {
		s += ": " + c.Topic
	}
	return s
}

// rawChannel is the wire shape shared by channel read paths.
type rawChannel struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  int    `json:"type"`
	Topic string `json:"topic"`
}

func newChannelListCommand(g *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List channels in the server",
		Annotations: map[string]string{
			"discordEndpoint": "GET /guilds/{guild.id}/channels",
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
			raw, err := client.Do("GET", "/guilds/"+gid+"/channels", nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var chans []rawChannel
			if err := json.Unmarshal(raw, &chans); err != nil {
				return err
			}
			items := make([]channelItem, len(chans))
			for i, ch := range chans {
				items[i] = channelItem{Name: ch.Name, ID: ch.ID, Type: channelTypeName(ch.Type), Topic: ch.Topic}
			}
			return emit(cmd, g, items)
		},
	}
}

func newChannelInfoCommand(g *GlobalFlags) *cobra.Command {
	var channel string
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show channel details",
		Annotations: map[string]string{
			"discordEndpoint": "GET /channels/{channel.id}",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(g)
			if err != nil {
				return err
			}
			cid, err := channelID(g, client, channel)
			if err != nil {
				return err
			}
			raw, err := client.Do("GET", "/channels/"+cid, nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var ch rawChannel
			if err := json.Unmarshal(raw, &ch); err != nil {
				return err
			}
			return emit(cmd, g, []channelItem{{Name: ch.Name, ID: ch.ID, Type: channelTypeName(ch.Type), Topic: ch.Topic}})
		},
	}
	cmd.Flags().StringVar(&channel, "channel", "", "channel ID or name")
	cmd.MarkFlagRequired("channel")
	return cmd
}

func newChannelCreateCommand(g *GlobalFlags) *cobra.Command {
	var name, chanType, topic, parent string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a channel",
		Annotations: map[string]string{
			"discordEndpoint": "POST /guilds/{guild.id}/channels",
			"write":           "true",
			"perms":           "MANAGE_CHANNELS",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			code, ok := channelTypeCodes[chanType]
			if !ok {
				return fmt.Errorf("unknown channel type %q (want text|voice|category|announcement|stage|forum|media)", chanType)
			}
			body := map[string]any{"name": name, "type": code}
			if topic != "" {
				body["topic"] = topic
			}
			if parent != "" {
				body["parent_id"] = parent
			}
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			if g.DryRun {
				dryRunPrint(cmd, "POST", "/guilds/"+gid+"/channels", body)
				return nil
			}
			raw, err := client.Do("POST", "/guilds/"+gid+"/channels", body)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var ch rawChannel
			if err := json.Unmarshal(raw, &ch); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created #%s (%s)\n", ch.Name, ch.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "channel name")
	cmd.Flags().StringVar(&chanType, "type", "text", "channel type: text|voice|category|announcement|stage|forum|media")
	cmd.Flags().StringVar(&topic, "topic", "", "channel topic")
	cmd.Flags().StringVar(&parent, "parent", "", "parent category ID")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newChannelEditCommand(g *GlobalFlags) *cobra.Command {
	var channel, name, topic string
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a channel's name or topic",
		Annotations: map[string]string{
			"discordEndpoint": "PATCH /channels/{channel.id}",
			"write":           "true",
			"perms":           "MANAGE_CHANNELS",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{}
			if name != "" {
				body["name"] = name
			}
			if topic != "" {
				body["topic"] = topic
			}
			if len(body) == 0 {
				return fmt.Errorf("nothing to change; pass --name and/or --topic")
			}
			client, cid, err := clientAndChannel(g, channel)
			if err != nil {
				return err
			}
			if g.DryRun {
				dryRunPrint(cmd, "PATCH", "/channels/"+cid, body)
				return nil
			}
			raw, err := client.Do("PATCH", "/channels/"+cid, body)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "edited %s\n", cid)
			return nil
		},
	}
	cmd.Flags().StringVar(&channel, "channel", "", "channel ID or name")
	cmd.Flags().StringVar(&name, "name", "", "new channel name")
	cmd.Flags().StringVar(&topic, "topic", "", "new channel topic")
	cmd.MarkFlagRequired("channel")
	return cmd
}

func newChannelDeleteCommand(g *GlobalFlags) *cobra.Command {
	var channel string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a channel",
		Annotations: map[string]string{
			"discordEndpoint": "DELETE /channels/{channel.id}",
			"write":           "true",
			"perms":           "MANAGE_CHANNELS",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, cid, err := clientAndChannel(g, channel)
			if err != nil {
				return err
			}
			if g.DryRun {
				dryRunPrint(cmd, "DELETE", "/channels/"+cid, nil)
				return nil
			}
			raw, err := client.Do("DELETE", "/channels/"+cid, nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deleted %s\n", cid)
			return nil
		},
	}
	cmd.Flags().StringVar(&channel, "channel", "", "channel ID or name")
	cmd.MarkFlagRequired("channel")
	return cmd
}

func newChannelTopicCommand(g *GlobalFlags) *cobra.Command {
	var channel, topic string
	cmd := &cobra.Command{
		Use:   "topic",
		Short: "Set a channel's topic",
		Annotations: map[string]string{
			"discordEndpoint": "PATCH /channels/{channel.id}",
			"write":           "true",
			"perms":           "MANAGE_CHANNELS",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, cid, err := clientAndChannel(g, channel)
			if err != nil {
				return err
			}
			body := map[string]any{"topic": topic}
			if g.DryRun {
				dryRunPrint(cmd, "PATCH", "/channels/"+cid, body)
				return nil
			}
			raw, err := client.Do("PATCH", "/channels/"+cid, body)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "topic set on %s\n", cid)
			return nil
		},
	}
	cmd.Flags().StringVar(&channel, "channel", "", "channel ID or name")
	cmd.Flags().StringVar(&topic, "topic", "", "new topic")
	cmd.MarkFlagRequired("channel")
	cmd.MarkFlagRequired("topic")
	return cmd
}
