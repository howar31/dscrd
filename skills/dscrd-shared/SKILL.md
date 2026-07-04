---
name: dscrd-shared
description: "dscrd CLI: shared reference for authentication, global flags, and security rules."
metadata:
  version: 0.1.1
  openclaw:
    category: "productivity"
    requires:
      bins:
        - dscrd
---

# dscrd — Shared Reference

PREREQUISITE for every `dscrd-<group>` skill. Covers discovery, global flags, and the rules that apply to all commands.

## Discovering Commands

- `dscrd --help` — all command groups.
- `dscrd <group> --help` — a group's verbs.
- `dscrd <group> <verb> --help` — a verb's flags.

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | — | validate without calling the API |
| `--format` | `concise` | output format: concise\|json\|jsonl\|table |
| `--guild` | — | target server: ID or name (default: profile default_guild) |
| `--no-resolve` | — | do not resolve IDs to names |
| `--profile` | — | config profile to use |
| `--raw` | — | return raw Discord API response |

## Targeting a Server

Most commands operate on one guild (server). Pass `--guild <id|name>` or set a default once: `dscrd auth set-token --default-guild <id>`. Channels accept IDs or names (`--channel "#general"`).

## Security Rules

- Never print or log bot token strings.
- Confirm with the user before any write/destructive command; preview with `--dry-run`.
- Write verbs honor `--raw` (return raw Discord JSON) and `--dry-run` (print the about-to-fire call and return without hitting the API).
- The bot acts under its own identity, never as the user's personal account.

## Exit Codes

`0` ok · `3` auth · `4` not found · `5` rate-limited · `1` other.

## Shell Tips

- `dscrd api` bodies are JSON; single-quote them so the shell keeps the inner double quotes (`--body '{"content":"hi"}'`).
- Shell `"\n"` is literal — for multi-line text use `--text-file` (`-` for stdin).
- Error messages include a `hint:` line when Discord-side setup (permissions, intents, invites) is the likely cause — follow it.
