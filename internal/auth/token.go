package auth

import "fmt"

// ResolveToken picks the active bot token. Precedence: envToken, then the
// named profile (or cfg.Active if profileName is empty).
func ResolveToken(cfg *Config, profileName, envToken string) (string, error) {
	if envToken != "" {
		return envToken, nil
	}
	name := profileName
	if name == "" {
		name = cfg.Active
	}
	if name == "" {
		return "", &AuthError{Reason: "no profile selected; run 'dscrd auth set-token' or set DSCRD_TOKEN"}
	}
	p, ok := cfg.Profiles[name]
	if !ok {
		return "", &AuthError{Reason: fmt.Sprintf("profile %q not found", name)}
	}
	if p.Token == "" {
		return "", &AuthError{Reason: fmt.Sprintf("profile %q has no token", name)}
	}
	// A still-encrypted value means Load could not decrypt it (key unavailable).
	if isEncrypted(p.Token) {
		return "", &AuthError{Reason: fmt.Sprintf("profile %q token could not be decrypted (encryption key unavailable)", name)}
	}
	return p.Token, nil
}

// ActiveProfile returns the profile the given flag/env selection resolves to,
// without touching the token. Used by commands that need ApplicationID or
// DefaultGuild alongside the token.
func ActiveProfile(cfg *Config, profileName string) (Profile, string, bool) {
	name := profileName
	if name == "" {
		name = cfg.Active
	}
	if name == "" {
		return Profile{}, "", false
	}
	p, ok := cfg.Profiles[name]
	return p, name, ok
}
