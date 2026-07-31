---
name: dscrd-member
description: "Inspect server members"
metadata:
  version: 0.2.0
  openclaw:
    category: "productivity"
    requires:
      bins:
        - dscrd
    cliHelp: "dscrd member --help"
---

# dscrd member

Inspect server members

> **PREREQUISITE:** Read `../dscrd-shared/SKILL.md` for auth, global flags, security rules, and exit codes. If missing, run `dscrd generate-skills`.

| Command | Description |
|---------|-------------|
| `dscrd member info` | Show one member's details |
| `dscrd member list` | List server members |
| `dscrd member search` | Search members by name prefix |

## dscrd member info

Show one member's details

**Discord API:** `GET /guilds/{guild.id}/members/{user.id}`

**Privileged intents:** `GUILD_MEMBERS` (enable in Developer Portal -> Bot -> Privileged Gateway Intents)

```bash
dscrd member info [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--user` | ✓ | — | user ID |

## dscrd member list

List server members

**Discord API:** `GET /guilds/{guild.id}/members`

**Privileged intents:** `GUILD_MEMBERS` (enable in Developer Portal -> Bot -> Privileged Gateway Intents)

```bash
dscrd member list [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--after` | — | — | paginate: member IDs after this user ID |
| `--limit` | — | `25` | max members (1-1000) |

## dscrd member search

Search members by name prefix

**Discord API:** `GET /guilds/{guild.id}/members/search`

```bash
dscrd member search [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--limit` | — | `25` | max results |
| `--query` | ✓ | — | name prefix to search |


