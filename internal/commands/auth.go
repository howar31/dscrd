package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/howar31/dscrd/internal/auth"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Invite-URL permission presets (Discord permission bit flags).
const (
	permViewChannel        = 1 << 10
	permReadMessageHistory = 1 << 16
	permAddReactions       = 1 << 6
	permSendMessages       = 1 << 11
	permEmbedLinks         = 1 << 14
	permAttachFiles        = 1 << 15
	permCreatePublicThread = 1 << 35
	permSendInThreads      = 1 << 38
	permAdministrator      = 1 << 3

	presetRead  = permViewChannel | permReadMessageHistory
	presetWrite = presetRead | permAddReactions | permSendMessages |
		permEmbedLinks | permAttachFiles | permCreatePublicThread | permSendInThreads
	presetAdmin = permAdministrator
)

func newAuthCommand(g *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "Manage bot credentials"}
	cmd.AddCommand(
		newAuthSetTokenCommand(g),
		newAuthStatusCommand(g),
		newAuthTestCommand(g),
		newAuthSwitchCommand(g),
		newAuthLogoutCommand(g),
		newAuthInviteURLCommand(g),
	)
	return cmd
}

func newAuthSetTokenCommand(g *GlobalFlags) *cobra.Command {
	var token, name, appID, defaultGuild string
	cmd := &cobra.Command{
		Use:   "set-token",
		Short: "Store a bot token (prompted securely when --token is omitted)",
		Annotations: map[string]string{
			"write": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" {
				if !term.IsTerminal(int(os.Stdin.Fd())) {
					return fmt.Errorf("no TTY for secure prompt; pass --token")
				}
				fmt.Fprint(cmd.OutOrStdout(), "Bot token: ")
				b, err := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Fprintln(cmd.OutOrStdout())
				if err != nil {
					return err
				}
				token = string(b)
			}
			if token == "" {
				return fmt.Errorf("empty token")
			}
			path, err := auth.ConfigPath()
			if err != nil {
				return err
			}
			cfg, err := auth.Load(path)
			if err != nil {
				return err
			}
			p := cfg.Profiles[name]
			p.Token = token
			if appID != "" {
				p.ApplicationID = appID
			}
			if defaultGuild != "" {
				p.DefaultGuild = defaultGuild
			}
			cfg.Profiles[name] = p
			if cfg.Active == "" {
				cfg.Active = name
			}
			if err := auth.Save(path, cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "profile %q saved (active: %s)\n", name, cfg.Active)
			return nil
		},
	}
	cmd.Flags().StringVar(&token, "token", "", "bot token (omit to be prompted)")
	cmd.Flags().StringVar(&name, "name", "default", "profile name")
	cmd.Flags().StringVar(&appID, "application-id", "", "application ID (for invite-url)")
	cmd.Flags().StringVar(&defaultGuild, "default-guild", "", "default guild ID or name")
	return cmd
}

// profileItem is one row of `auth status`.
type profileItem struct {
	Name         string `json:"name"`
	Active       bool   `json:"active"`
	Token        string `json:"token"` // state only: set|empty|locked — never the value
	AppID        string `json:"application_id"`
	DefaultGuild string `json:"default_guild"`
}

func (p profileItem) Concise() string {
	marker := " "
	if p.Active {
		marker = "*"
	}
	s := fmt.Sprintf("%s %s token:%s", marker, p.Name, p.Token)
	if p.AppID != "" {
		s += " app:" + p.AppID
	}
	if p.DefaultGuild != "" {
		s += " guild:" + p.DefaultGuild
	}
	return s
}

func newAuthStatusCommand(g *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "List profiles and verify the active token",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := auth.ConfigPath()
			if err != nil {
				return err
			}
			cfg, err := auth.Load(path)
			if err != nil {
				return err
			}
			items := make([]profileItem, 0, len(cfg.Profiles))
			for name, p := range cfg.Profiles {
				state := "set"
				if p.Token == "" {
					state = "empty"
				} else if auth.IsEncrypted(p.Token) {
					state = "locked"
				}
				items = append(items, profileItem{
					Name: name, Active: name == cfg.Active, Token: state,
					AppID: p.ApplicationID, DefaultGuild: p.DefaultGuild,
				})
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no profiles; run 'dscrd auth set-token'")
				return nil
			}
			if err := emit(cmd, g, items); err != nil {
				return err
			}
			// Live check of the active token, informational only.
			if client, err := buildClient(g); err == nil {
				if raw, err := client.Do("GET", "/users/@me", nil); err == nil {
					var me struct {
						ID       string `json:"id"`
						Username string `json:"username"`
					}
					if json.Unmarshal(raw, &me) == nil {
						fmt.Fprintf(cmd.OutOrStdout(), "token ok: %s (%s)\n", me.Username, me.ID)
					}
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "token check failed: %v\n", err)
				}
			}
			return nil
		},
	}
}

func newAuthTestCommand(g *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Verify the active token against Discord",
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
			var me struct {
				ID       string `json:"id"`
				Username string `json:"username"`
			}
			if err := json.Unmarshal(raw, &me); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ok %s (%s)\n", me.Username, me.ID)
			return nil
		},
	}
}

func newAuthSwitchCommand(g *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "switch <name>",
		Short: "Switch the active profile",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			"write": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := auth.ConfigPath()
			if err != nil {
				return err
			}
			cfg, err := auth.Load(path)
			if err != nil {
				return err
			}
			if _, ok := cfg.Profiles[args[0]]; !ok {
				return fmt.Errorf("profile %q not found", args[0])
			}
			cfg.Active = args[0]
			if err := auth.Save(path, cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "active profile: %s\n", args[0])
			return nil
		},
	}
}

func newAuthLogoutCommand(g *GlobalFlags) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove a stored profile",
		Annotations: map[string]string{
			"write": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := auth.ConfigPath()
			if err != nil {
				return err
			}
			cfg, err := auth.Load(path)
			if err != nil {
				return err
			}
			target := name
			if target == "" {
				target = cfg.Active
			}
			if target == "" {
				return fmt.Errorf("no profile to remove; pass --name")
			}
			if _, ok := cfg.Profiles[target]; !ok {
				return fmt.Errorf("profile %q not found", target)
			}
			delete(cfg.Profiles, target)
			if cfg.Active == target {
				cfg.Active = ""
			}
			if err := auth.Save(path, cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed profile %q\n", target)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "profile to remove (default: active)")
	return cmd
}

func newAuthInviteURLCommand(g *GlobalFlags) *cobra.Command {
	var perms string
	cmd := &cobra.Command{
		Use:   "invite-url",
		Short: "Generate the OAuth2 URL that invites the bot to a server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			p, name, ok := auth.ActiveProfile(cfg, profileName(g))
			if !ok || p.ApplicationID == "" {
				return fmt.Errorf("profile %q has no application_id; set it via 'dscrd auth set-token --application-id <id>'", name)
			}
			var bits uint64
			switch perms {
			case "read":
				bits = presetRead
			case "write":
				bits = presetWrite
			case "admin":
				bits = presetAdmin
			default:
				bits, err = strconv.ParseUint(perms, 10, 64)
				if err != nil {
					return fmt.Errorf("--permissions must be read|write|admin or a permission bit set: %w", err)
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"https://discord.com/oauth2/authorize?client_id=%s&scope=bot&permissions=%d\n",
				p.ApplicationID, bits)
			return nil
		},
	}
	cmd.Flags().StringVar(&perms, "permissions", "write", "permission preset (read|write|admin) or raw bit set")
	return cmd
}
