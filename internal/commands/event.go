package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newEventCommand(g *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "event", Short: "Manage scheduled events"}
	cmd.AddCommand(
		newEventListCommand(g),
		newEventCreateCommand(g),
		newEventEditCommand(g),
		newEventDeleteCommand(g),
	)
	return cmd
}

// eventItem is one rendered scheduled event.
type eventItem struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Start string `json:"start"`
	Where string `json:"where"`
}

func (e eventItem) Concise() string {
	s := fmt.Sprintf("%s (%s) — %s", e.Name, e.ID, e.Start)
	if e.Where != "" {
		s += " @ " + e.Where
	}
	return s
}

// rawEvent is the wire shape of a scheduled event.
type rawEvent struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	StartTime      string `json:"scheduled_start_time"`
	ChannelID      string `json:"channel_id"`
	EntityMetadata struct {
		Location string `json:"location"`
	} `json:"entity_metadata"`
}

func renderEvent(e rawEvent) eventItem {
	where := e.EntityMetadata.Location
	if where == "" {
		where = e.ChannelID
	}
	return eventItem{Name: e.Name, ID: e.ID, Start: e.StartTime, Where: where}
}

func newEventListCommand(g *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List scheduled events",
		Annotations: map[string]string{
			"discordEndpoint": "GET /guilds/{guild.id}/scheduled-events",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			raw, err := client.Do("GET", "/guilds/"+gid+"/scheduled-events", nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var events []rawEvent
			if err := json.Unmarshal(raw, &events); err != nil {
				return err
			}
			items := make([]eventItem, len(events))
			for i, e := range events {
				items[i] = renderEvent(e)
			}
			return emit(cmd, g, items)
		},
	}
}

func newEventCreateCommand(g *GlobalFlags) *cobra.Command {
	var name, start, end, location, channel, desc string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a scheduled event (external with --location, or voice with --channel)",
		Annotations: map[string]string{
			"discordEndpoint": "POST /guilds/{guild.id}/scheduled-events",
			"write":           "true",
			"perms":           "MANAGE_EVENTS",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{
				"name":                 name,
				"scheduled_start_time": start,
				"privacy_level":        2, // GUILD_ONLY, the only accepted value
			}
			if desc != "" {
				body["description"] = desc
			}
			switch {
			case location != "":
				if end == "" {
					return fmt.Errorf("--location events require --end")
				}
				body["entity_type"] = 3 // EXTERNAL
				body["entity_metadata"] = map[string]string{"location": location}
				body["scheduled_end_time"] = end
			case channel != "":
				body["entity_type"] = 2 // VOICE
				body["channel_id"] = channel
				if end != "" {
					body["scheduled_end_time"] = end
				}
			default:
				return fmt.Errorf("pass --location (external event) or --channel (voice event)")
			}
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			path := "/guilds/" + gid + "/scheduled-events"
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
			var e rawEvent
			if err := json.Unmarshal(raw, &e); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created event %s (%s)\n", e.Name, e.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "event name")
	cmd.Flags().StringVar(&start, "start", "", "start time (ISO8601, e.g. 2026-08-01T19:00:00+08:00)")
	cmd.Flags().StringVar(&end, "end", "", "end time (ISO8601; required for --location events)")
	cmd.Flags().StringVar(&location, "location", "", "external location text")
	cmd.Flags().StringVar(&channel, "channel", "", "voice channel ID")
	cmd.Flags().StringVar(&desc, "description", "", "event description")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("start")
	return cmd
}

func newEventEditCommand(g *GlobalFlags) *cobra.Command {
	var id, name, start, desc string
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a scheduled event",
		Annotations: map[string]string{
			"discordEndpoint": "PATCH /guilds/{guild.id}/scheduled-events/{event.id}",
			"write":           "true",
			"perms":           "MANAGE_EVENTS",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{}
			if name != "" {
				body["name"] = name
			}
			if start != "" {
				body["scheduled_start_time"] = start
			}
			if desc != "" {
				body["description"] = desc
			}
			if len(body) == 0 {
				return fmt.Errorf("nothing to change; pass --name, --start, and/or --description")
			}
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			path := "/guilds/" + gid + "/scheduled-events/" + id
			if g.DryRun {
				dryRunPrint(cmd, "PATCH", path, body)
				return nil
			}
			if _, err := client.Do("PATCH", path, body); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "edited event %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "event ID")
	cmd.Flags().StringVar(&name, "name", "", "new name")
	cmd.Flags().StringVar(&start, "start", "", "new start time (ISO8601)")
	cmd.Flags().StringVar(&desc, "description", "", "new description")
	cmd.MarkFlagRequired("id")
	return cmd
}

func newEventDeleteCommand(g *GlobalFlags) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a scheduled event",
		Annotations: map[string]string{
			"discordEndpoint": "DELETE /guilds/{guild.id}/scheduled-events/{event.id}",
			"write":           "true",
			"perms":           "MANAGE_EVENTS",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			path := "/guilds/" + gid + "/scheduled-events/" + id
			if g.DryRun {
				dryRunPrint(cmd, "DELETE", path, nil)
				return nil
			}
			if _, err := client.Do("DELETE", path, nil); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deleted event %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "event ID")
	cmd.MarkFlagRequired("id")
	return cmd
}
