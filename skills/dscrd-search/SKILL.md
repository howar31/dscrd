---
name: dscrd-search
description: "Search messages in a server"
metadata:
  version: 0.1.0
  openclaw:
    category: "productivity"
    requires:
      bins:
        - dscrd
    cliHelp: "dscrd search --help"
---

# dscrd search

Search messages in a server

> **PREREQUISITE:** Read `../dscrd-shared/SKILL.md` for auth, global flags, security rules, and exit codes. If missing, run `dscrd generate-skills`.

| Command | Description |
|---------|-------------|
| `dscrd search messages` | Full-text search over the server's messages |

## dscrd search messages

Full-text search over the server's messages

**Discord API:** `GET /guilds/{guild.id}/messages/search`

**Bot permissions:** `READ_MESSAGE_HISTORY`

**Privileged intents:** `MESSAGE_CONTENT` (enable in Developer Portal -> Bot -> Privileged Gateway Intents)

```bash
dscrd search messages [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--author` | — | — | filter by author user ID |
| `--channel` | — | — | filter by channel ID or name |
| `--content` | — | — | text to search for |
| `--has` | — | — | filter by content kind: link\|embed\|file\|image\|video\|sound\|sticker |
| `--limit` | — | `10` | max results (1-25) |
| `--offset` | — | `0` | pagination offset |
| `--pinned` | — | — | only pinned (or --pinned=false for unpinned) |
| `--sort-by` | — | — | sort key: timestamp\|relevance |
| `--sort-order` | — | — | sort order: asc\|desc |


