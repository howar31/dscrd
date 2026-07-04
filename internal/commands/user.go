package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newUserCommand(g *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "user", Short: "Inspect users"}
	cmd.AddCommand(newUserInfoCommand(g), newUserMeCommand(g))
	return cmd
}

func printUser(cmd *cobra.Command, raw []byte) error {
	var u struct {
		ID         string `json:"id"`
		Username   string `json:"username"`
		GlobalName string `json:"global_name"`
		Bot        bool   `json:"bot"`
	}
	if err := json.Unmarshal(raw, &u); err != nil {
		return err
	}
	name := u.GlobalName
	if name == "" {
		name = u.Username
	}
	s := fmt.Sprintf("%s (%s)", name, u.ID)
	if u.GlobalName != "" && u.Username != "" && u.GlobalName != u.Username {
		s += " @" + u.Username
	}
	if u.Bot {
		s += " [bot]"
	}
	fmt.Fprintln(cmd.OutOrStdout(), s)
	return nil
}

func newUserInfoCommand(g *GlobalFlags) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show a user's profile",
		Annotations: map[string]string{
			"discordEndpoint": "GET /users/{user.id}",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(g)
			if err != nil {
				return err
			}
			raw, err := client.Do("GET", "/users/"+id, nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			return printUser(cmd, raw)
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "user ID")
	cmd.MarkFlagRequired("id")
	return cmd
}

func newUserMeCommand(g *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "me",
		Short: "Show the bot's own user",
		Annotations: map[string]string{
			"discordEndpoint": "GET /users/@me",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(g)
			if err != nil {
				return err
			}
			raw, err := client.Do("GET", "/users/@me", nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			return printUser(cmd, raw)
		},
	}
}
