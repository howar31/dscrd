package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

func newSearchCommand(g *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "search", Short: "Search messages in a server"}
	cmd.AddCommand(newSearchMessagesCommand(g))
	return cmd
}

func newSearchMessagesCommand(g *GlobalFlags) *cobra.Command {
	var content, author, channel, has, sortBy, sortOrder string
	var pinned bool
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "messages",
		Short: "Full-text search over the server's messages",
		Annotations: map[string]string{
			"discordEndpoint": "GET /guilds/{guild.id}/messages/search",
			"perms":           "READ_MESSAGE_HISTORY",
			"intents":         "MESSAGE_CONTENT",
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
			if content != "" {
				q.Set("content", content)
			}
			if author != "" {
				q.Set("author_id", author)
			}
			if channel != "" {
				cid, err := channelID(g, client, channel)
				if err != nil {
					return err
				}
				q.Set("channel_id", cid)
			}
			if has != "" {
				q.Set("has", has)
			}
			if cmd.Flags().Changed("pinned") {
				q.Set("pinned", strconv.FormatBool(pinned))
			}
			if limit > 0 {
				q.Set("limit", strconv.Itoa(limit))
			}
			if offset > 0 {
				q.Set("offset", strconv.Itoa(offset))
			}
			if sortBy != "" {
				q.Set("sort_by", sortBy)
			}
			if sortOrder != "" {
				q.Set("sort_order", sortOrder)
			}
			raw, err := client.Do("GET", "/guilds/"+gid+"/messages/search?"+q.Encode(), nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var resp struct {
				TotalResults int            `json:"total_results"`
				Messages     [][]rawMessage `json:"messages"`
			}
			if err := json.Unmarshal(raw, &resp); err != nil {
				return err
			}
			r := newResolver(g, client)
			var items []msgItem
			// Each tuple is a hit plus context; the first element is the hit.
			for _, tuple := range resp.Messages {
				if len(tuple) == 0 {
					continue
				}
				items = append(items, renderMessage(r, tuple[0]))
			}
			if err := emit(cmd, g, items); err != nil {
				return err
			}
			if g.Format == "concise" || g.Format == "" {
				fmt.Fprintf(cmd.OutOrStdout(), "total %d\n", resp.TotalResults)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&content, "content", "", "text to search for")
	cmd.Flags().StringVar(&author, "author", "", "filter by author user ID")
	cmd.Flags().StringVar(&channel, "channel", "", "filter by channel ID or name")
	cmd.Flags().StringVar(&has, "has", "", "filter by content kind: link|embed|file|image|video|sound|sticker")
	cmd.Flags().BoolVar(&pinned, "pinned", false, "only pinned (or --pinned=false for unpinned)")
	cmd.Flags().IntVar(&limit, "limit", 10, "max results (1-25)")
	cmd.Flags().IntVar(&offset, "offset", 0, "pagination offset")
	cmd.Flags().StringVar(&sortBy, "sort-by", "", "sort key: timestamp|relevance")
	cmd.Flags().StringVar(&sortOrder, "sort-order", "", "sort order: asc|desc")
	return cmd
}
