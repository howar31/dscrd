---
name: dscrd-invite
description: "Manage invite links"
metadata:
  version: 0.1.1
  openclaw:
    category: "productivity"
    requires:
      bins:
        - dscrd
    cliHelp: "dscrd invite --help"
---

# dscrd invite

Manage invite links

> **PREREQUISITE:** Read `../dscrd-shared/SKILL.md` for auth, global flags, security rules, and exit codes. If missing, run `dscrd generate-skills`.

| Command | Description |
|---------|-------------|
| `dscrd invite create` | Create an invite for a channel |
| `dscrd invite delete` | Revoke an invite |
| `dscrd invite info` | Show an invite's details |
| `dscrd invite list` | List the server's invites |

## dscrd invite create

Create an invite for a channel

**Discord API:** `POST /channels/{channel.id}/invites`

**Bot permissions:** `CREATE_INSTANT_INVITE`

```bash
dscrd invite create [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--channel` | ✓ | — | channel ID or name |
| `--max-age` | — | `86400` | expiry in seconds (0 = never) |
| `--max-uses` | — | `0` | max uses (0 = unlimited) |

> [!CAUTION]
> Write command — confirm with the user before executing; preview with `--dry-run`.

## dscrd invite delete

Revoke an invite

**Discord API:** `DELETE /invites/{invite.code}`

**Bot permissions:** `MANAGE_CHANNELS`

```bash
dscrd invite delete <code>
```

> [!CAUTION]
> Write command — confirm with the user before executing; preview with `--dry-run`.

## dscrd invite info

Show an invite's details

**Discord API:** `GET /invites/{invite.code}`

```bash
dscrd invite info <code>
```

## dscrd invite list

List the server's invites

**Discord API:** `GET /guilds/{guild.id}/invites`

**Bot permissions:** `MANAGE_GUILD`

```bash
dscrd invite list
```


