---
name: dscrd-sticker
description: "Inspect the server's stickers"
metadata:
  version: 0.1.0
  openclaw:
    category: "productivity"
    requires:
      bins:
        - dscrd
    cliHelp: "dscrd sticker --help"
---

# dscrd sticker

Inspect the server's stickers

> **PREREQUISITE:** Read `../dscrd-shared/SKILL.md` for auth, global flags, security rules, and exit codes. If missing, run `dscrd generate-skills`.

| Command | Description |
|---------|-------------|
| `dscrd sticker info` | Show one sticker |
| `dscrd sticker list` | List the server's stickers |

## dscrd sticker info

Show one sticker

**Discord API:** `GET /guilds/{guild.id}/stickers/{sticker.id}`

```bash
dscrd sticker info [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--id` | ✓ | — | sticker ID |

## dscrd sticker list

List the server's stickers

**Discord API:** `GET /guilds/{guild.id}/stickers`

```bash
dscrd sticker list
```


