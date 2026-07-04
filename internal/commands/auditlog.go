package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

// auditActionNames maps common audit-log action type codes to readable names.
var auditActionNames = map[int]string{
	1:   "guild-update",
	10:  "channel-create",
	11:  "channel-update",
	12:  "channel-delete",
	20:  "member-kick",
	22:  "member-ban",
	23:  "member-unban",
	24:  "member-update",
	25:  "member-role-update",
	30:  "role-create",
	31:  "role-update",
	32:  "role-delete",
	40:  "invite-create",
	42:  "invite-delete",
	50:  "webhook-create",
	52:  "webhook-delete",
	60:  "emoji-create",
	62:  "emoji-delete",
	72:  "message-delete",
	74:  "message-pin",
	110: "thread-create",
	112: "thread-delete",
}

func auditActionName(code int) string {
	if n, ok := auditActionNames[code]; ok {
		return n
	}
	return fmt.Sprintf("action-%d", code)
}

func newAuditLogCommand(g *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "audit-log", Short: "Read the server audit log"}
	cmd.AddCommand(newAuditLogReadCommand(g))
	return cmd
}

// auditItem is one rendered audit-log entry.
type auditItem struct {
	Action string `json:"action"`
	UserID string `json:"user_id"`
	Target string `json:"target_id"`
	Reason string `json:"reason"`
	ID     string `json:"id"`
}

func (a auditItem) Concise() string {
	s := fmt.Sprintf("%s by %s", a.Action, a.UserID)
	if a.Target != "" {
		s += " on " + a.Target
	}
	if a.Reason != "" {
		s += " — " + a.Reason
	}
	return s + " (" + a.ID + ")"
}

func newAuditLogReadCommand(g *GlobalFlags) *cobra.Command {
	var action, user string
	var limit int
	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read recent audit-log entries",
		Annotations: map[string]string{
			"discordEndpoint": "GET /guilds/{guild.id}/audit-logs",
			"perms":           "VIEW_AUDIT_LOG",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			q := url.Values{}
			q.Set("limit", strconv.Itoa(limit))
			if action != "" {
				q.Set("action_type", action)
			}
			if user != "" {
				q.Set("user_id", user)
			}
			raw, err := client.Do("GET", "/guilds/"+gid+"/audit-logs?"+q.Encode(), nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var resp struct {
				Entries []struct {
					ID         string `json:"id"`
					ActionType int    `json:"action_type"`
					UserID     string `json:"user_id"`
					TargetID   string `json:"target_id"`
					Reason     string `json:"reason"`
				} `json:"audit_log_entries"`
			}
			if err := json.Unmarshal(raw, &resp); err != nil {
				return err
			}
			items := make([]auditItem, len(resp.Entries))
			for i, e := range resp.Entries {
				items[i] = auditItem{
					Action: auditActionName(e.ActionType), UserID: e.UserID,
					Target: e.TargetID, Reason: e.Reason, ID: e.ID,
				}
			}
			return emit(cmd, g, items)
		},
	}
	cmd.Flags().StringVar(&action, "action", "", "filter by numeric action type")
	cmd.Flags().StringVar(&user, "user", "", "filter by acting user ID")
	cmd.Flags().IntVar(&limit, "limit", 25, "max entries (1-100)")
	return cmd
}
