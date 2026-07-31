---
name: dscrd-user
description: "Inspect users"
metadata:
  version: 0.2.0
  openclaw:
    category: "productivity"
    requires:
      bins:
        - dscrd
    cliHelp: "dscrd user --help"
---

# dscrd user

Inspect users

> **PREREQUISITE:** Read `../dscrd-shared/SKILL.md` for auth, global flags, security rules, and exit codes. If missing, run `dscrd generate-skills`.

| Command | Description |
|---------|-------------|
| `dscrd user info` | Show a user's profile |
| `dscrd user me` | Show the bot's own user |

## dscrd user info

Show a user's profile

**Discord API:** `GET /users/{user.id}`

```bash
dscrd user info [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--id` | ✓ | — | user ID |

## dscrd user me

Show the bot's own user

**Discord API:** `GET /users/@me`

```bash
dscrd user me
```


