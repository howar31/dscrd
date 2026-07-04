package commands

import "github.com/spf13/cobra"

// GlobalFlags holds flags shared by every command.
type GlobalFlags struct {
	Format    string // concise|json|jsonl|table
	Profile   string
	Guild     string // target guild: snowflake ID or name
	Raw       bool
	DryRun    bool
	NoResolve bool
}

// bindGlobalFlags registers persistent flags on cmd, writing into g.
func bindGlobalFlags(cmd *cobra.Command, g *GlobalFlags) {
	pf := cmd.PersistentFlags()
	pf.StringVar(&g.Format, "format", "concise", "output format: concise|json|jsonl|table")
	pf.StringVar(&g.Profile, "profile", "", "config profile to use")
	pf.StringVar(&g.Guild, "guild", "", "target server: ID or name (default: profile default_guild)")
	pf.BoolVar(&g.Raw, "raw", false, "return raw Discord API response")
	pf.BoolVar(&g.DryRun, "dry-run", false, "validate without calling the API")
	pf.BoolVar(&g.NoResolve, "no-resolve", false, "do not resolve IDs to names")
}
