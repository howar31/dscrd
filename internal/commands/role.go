package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newRoleCommand(g *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "role", Short: "Manage roles"}
	cmd.AddCommand(
		newRoleListCommand(g),
		newRoleCreateCommand(g),
		newRoleEditCommand(g),
		newRoleDeleteCommand(g),
		newRoleAssignCommand(g),
		newRoleUnassignCommand(g),
	)
	return cmd
}

// roleItem is one rendered role.
type roleItem struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Members string `json:"members"`
}

func (r roleItem) Concise() string { return fmt.Sprintf("%s (%s)", r.Name, r.ID) }

func newRoleListCommand(g *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List roles in the server",
		Annotations: map[string]string{
			"discordEndpoint": "GET /guilds/{guild.id}/roles",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			raw, err := client.Do("GET", "/guilds/"+gid+"/roles", nil)
			if err != nil {
				return err
			}
			if g.Raw {
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var roles []roleItem
			if err := json.Unmarshal(raw, &roles); err != nil {
				return err
			}
			return emit(cmd, g, roles)
		},
	}
}

func newRoleCreateCommand(g *GlobalFlags) *cobra.Command {
	var name string
	var color int
	var hoist, mentionable bool
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a role",
		Annotations: map[string]string{
			"discordEndpoint": "POST /guilds/{guild.id}/roles",
			"write":           "true",
			"perms":           "MANAGE_ROLES",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{"name": name, "hoist": hoist, "mentionable": mentionable}
			if color != 0 {
				body["color"] = color
			}
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			path := "/guilds/" + gid + "/roles"
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
			var role roleItem
			if err := json.Unmarshal(raw, &role); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created role %s (%s)\n", role.Name, role.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "role name")
	cmd.Flags().IntVar(&color, "color", 0, "RGB color as integer")
	cmd.Flags().BoolVar(&hoist, "hoist", false, "display separately in the member list")
	cmd.Flags().BoolVar(&mentionable, "mentionable", false, "allow anyone to mention")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newRoleEditCommand(g *GlobalFlags) *cobra.Command {
	var role, name string
	var color int
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a role",
		Annotations: map[string]string{
			"discordEndpoint": "PATCH /guilds/{guild.id}/roles/{role.id}",
			"write":           "true",
			"perms":           "MANAGE_ROLES",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{}
			if name != "" {
				body["name"] = name
			}
			if cmd.Flags().Changed("color") {
				body["color"] = color
			}
			if len(body) == 0 {
				return fmt.Errorf("nothing to change; pass --name and/or --color")
			}
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			path := "/guilds/" + gid + "/roles/" + role
			if g.DryRun {
				dryRunPrint(cmd, "PATCH", path, body)
				return nil
			}
			if _, err := client.Do("PATCH", path, body); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "edited role %s\n", role)
			return nil
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "role ID")
	cmd.Flags().StringVar(&name, "name", "", "new name")
	cmd.Flags().IntVar(&color, "color", 0, "new RGB color as integer")
	cmd.MarkFlagRequired("role")
	return cmd
}

func newRoleDeleteCommand(g *GlobalFlags) *cobra.Command {
	var role string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a role",
		Annotations: map[string]string{
			"discordEndpoint": "DELETE /guilds/{guild.id}/roles/{role.id}",
			"write":           "true",
			"perms":           "MANAGE_ROLES",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			path := "/guilds/" + gid + "/roles/" + role
			if g.DryRun {
				dryRunPrint(cmd, "DELETE", path, nil)
				return nil
			}
			if _, err := client.Do("DELETE", path, nil); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deleted role %s\n", role)
			return nil
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "role ID")
	cmd.MarkFlagRequired("role")
	return cmd
}

func newRoleAssignCommand(g *GlobalFlags) *cobra.Command {
	var role, user string
	cmd := &cobra.Command{
		Use:   "assign",
		Short: "Give a member a role",
		Annotations: map[string]string{
			"discordEndpoint": "PUT /guilds/{guild.id}/members/{user.id}/roles/{role.id}",
			"write":           "true",
			"perms":           "MANAGE_ROLES",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			path := "/guilds/" + gid + "/members/" + user + "/roles/" + role
			if g.DryRun {
				dryRunPrint(cmd, "PUT", path, nil)
				return nil
			}
			if _, err := client.Do("PUT", path, nil); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "assigned role %s to %s\n", role, user)
			return nil
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "role ID")
	cmd.Flags().StringVar(&user, "user", "", "user ID")
	cmd.MarkFlagRequired("role")
	cmd.MarkFlagRequired("user")
	return cmd
}

func newRoleUnassignCommand(g *GlobalFlags) *cobra.Command {
	var role, user string
	cmd := &cobra.Command{
		Use:   "unassign",
		Short: "Remove a role from a member",
		Annotations: map[string]string{
			"discordEndpoint": "DELETE /guilds/{guild.id}/members/{user.id}/roles/{role.id}",
			"write":           "true",
			"perms":           "MANAGE_ROLES",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, gid, err := clientAndGuild(g)
			if err != nil {
				return err
			}
			path := "/guilds/" + gid + "/members/" + user + "/roles/" + role
			if g.DryRun {
				dryRunPrint(cmd, "DELETE", path, nil)
				return nil
			}
			if _, err := client.Do("DELETE", path, nil); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed role %s from %s\n", role, user)
			return nil
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "role ID")
	cmd.Flags().StringVar(&user, "user", "", "user ID")
	cmd.MarkFlagRequired("role")
	cmd.MarkFlagRequired("user")
	return cmd
}
