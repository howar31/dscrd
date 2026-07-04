---
name: dscrd-role
description: "Manage roles"
metadata:
  version: 0.1.1
  openclaw:
    category: "productivity"
    requires:
      bins:
        - dscrd
    cliHelp: "dscrd role --help"
---

# dscrd role

Manage roles

> **PREREQUISITE:** Read `../dscrd-shared/SKILL.md` for auth, global flags, security rules, and exit codes. If missing, run `dscrd generate-skills`.

| Command | Description |
|---------|-------------|
| `dscrd role assign` | Give a member a role |
| `dscrd role create` | Create a role |
| `dscrd role delete` | Delete a role |
| `dscrd role edit` | Edit a role |
| `dscrd role list` | List roles in the server |
| `dscrd role unassign` | Remove a role from a member |

## dscrd role assign

Give a member a role

**Discord API:** `PUT /guilds/{guild.id}/members/{user.id}/roles/{role.id}`

**Bot permissions:** `MANAGE_ROLES`

```bash
dscrd role assign [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--role` | ✓ | — | role ID |
| `--user` | ✓ | — | user ID |

> [!CAUTION]
> Write command — confirm with the user before executing; preview with `--dry-run`.

## dscrd role create

Create a role

**Discord API:** `POST /guilds/{guild.id}/roles`

**Bot permissions:** `MANAGE_ROLES`

```bash
dscrd role create [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--color` | — | `0` | RGB color as integer |
| `--hoist` | — | — | display separately in the member list |
| `--mentionable` | — | — | allow anyone to mention |
| `--name` | ✓ | — | role name |

> [!CAUTION]
> Write command — confirm with the user before executing; preview with `--dry-run`.

## dscrd role delete

Delete a role

**Discord API:** `DELETE /guilds/{guild.id}/roles/{role.id}`

**Bot permissions:** `MANAGE_ROLES`

```bash
dscrd role delete [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--role` | ✓ | — | role ID |

> [!CAUTION]
> Write command — confirm with the user before executing; preview with `--dry-run`.

## dscrd role edit

Edit a role

**Discord API:** `PATCH /guilds/{guild.id}/roles/{role.id}`

**Bot permissions:** `MANAGE_ROLES`

```bash
dscrd role edit [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--color` | — | `0` | new RGB color as integer |
| `--name` | — | — | new name |
| `--role` | ✓ | — | role ID |

> [!CAUTION]
> Write command — confirm with the user before executing; preview with `--dry-run`.

## dscrd role list

List roles in the server

**Discord API:** `GET /guilds/{guild.id}/roles`

```bash
dscrd role list
```

## dscrd role unassign

Remove a role from a member

**Discord API:** `DELETE /guilds/{guild.id}/members/{user.id}/roles/{role.id}`

**Bot permissions:** `MANAGE_ROLES`

```bash
dscrd role unassign [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--role` | ✓ | — | role ID |
| `--user` | ✓ | — | user ID |

> [!CAUTION]
> Write command — confirm with the user before executing; preview with `--dry-run`.


