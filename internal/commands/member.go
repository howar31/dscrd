package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func newMemberCommand(g *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "member", Short: "Inspect server members"}
	cmd.AddCommand(newMemberListCommand(g), newMemberInfoCommand(g), newMemberSearchCommand(g))
	return cmd
}

// rawMember is the wire shape of a guild member.
type rawMember struct {
	Nick string `json:"nick"`
	User struct {
		ID         string `json:"id"`
		Username   string `json:"username"`
		GlobalName string `json:"global_name"`
	} `json:"user"`
	Roles    []string `json:"roles"`
	JoinedAt string   `json:"joined_at"`
}

// memberItem is one rendered member.
type memberItem struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Roles string `json:"roles"`
}

func (m memberItem) Concise() string {
	s := fmt.Sprintf("%s (%s)", m.Name, m.ID)
	if m.Roles != "" {
		s += " — roles: " + m.Roles
	}
	return s
}

func renderMember(m rawMember) memberItem {
	name := m.Nick
	if name == "" {
		name = m.User.GlobalName
	}
	if name == "" {
		name = m.User.Username
	}
	return memberItem{Name: name, ID: m.User.ID, Roles: strings.Join(m.Roles, ",")}
}

func newMemberListCommand(g *GlobalFlags) *cobra.Command {
	var limit int
	var after string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List server members",
		Annotations: map[string]string{
			"discordEndpoint": "GET /guilds/{guild.id}/members",
			"intents":         "GUILD_MEMBERS",
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
			q := url.Values{}
			q.Set("limit", strconv.Itoa(limit))
			if after != "" {
				q.Set("after", after)
			}
			raw, err := client.Do("GET", "/guilds/"+gid+"/members?"+q.Encode(), nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var members []rawMember
			if err := json.Unmarshal(raw, &members); err != nil {
				return err
			}
			items := make([]memberItem, len(members))
			for i, m := range members {
				items[i] = renderMember(m)
			}
			return emit(cmd, g, items)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "max members (1-1000)")
	cmd.Flags().StringVar(&after, "after", "", "paginate: member IDs after this user ID")
	return cmd
}

func newMemberInfoCommand(g *GlobalFlags) *cobra.Command {
	var user string
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show one member's details",
		Annotations: map[string]string{
			"discordEndpoint": "GET /guilds/{guild.id}/members/{user.id}",
			"intents":         "GUILD_MEMBERS",
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
			raw, err := client.Do("GET", "/guilds/"+gid+"/members/"+user, nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var m rawMember
			if err := json.Unmarshal(raw, &m); err != nil {
				return err
			}
			item := renderMember(m)
			fmt.Fprintf(cmd.OutOrStdout(), "%s joined:%s\n", item.Concise(), m.JoinedAt)
			return nil
		},
	}
	cmd.Flags().StringVar(&user, "user", "", "user ID")
	cmd.MarkFlagRequired("user")
	return cmd
}

func newMemberSearchCommand(g *GlobalFlags) *cobra.Command {
	var query string
	var limit int
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search members by name prefix",
		Annotations: map[string]string{
			"discordEndpoint": "GET /guilds/{guild.id}/members/search",
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
			q := url.Values{}
			q.Set("query", query)
			q.Set("limit", strconv.Itoa(limit))
			raw, err := client.Do("GET", "/guilds/"+gid+"/members/search?"+q.Encode(), nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var members []rawMember
			if err := json.Unmarshal(raw, &members); err != nil {
				return err
			}
			items := make([]memberItem, len(members))
			for i, m := range members {
				items[i] = renderMember(m)
			}
			return emit(cmd, g, items)
		},
	}
	cmd.Flags().StringVar(&query, "query", "", "name prefix to search")
	cmd.Flags().IntVar(&limit, "limit", 25, "max results")
	cmd.MarkFlagRequired("query")
	return cmd
}
