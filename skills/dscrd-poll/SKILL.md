---
name: dscrd-poll
description: "Create and read polls"
metadata:
  version: 0.1.1
  openclaw:
    category: "productivity"
    requires:
      bins:
        - dscrd
    cliHelp: "dscrd poll --help"
---

# dscrd poll

Create and read polls

> **PREREQUISITE:** Read `../dscrd-shared/SKILL.md` for auth, global flags, security rules, and exit codes. If missing, run `dscrd generate-skills`.

| Command | Description |
|---------|-------------|
| `dscrd poll create` | Post a poll to a channel |
| `dscrd poll end` | End a poll immediately |
| `dscrd poll results` | List voters for one poll answer |

## dscrd poll create

Post a poll to a channel

**Discord API:** `POST /channels/{channel.id}/messages`

**Bot permissions:** `SEND_MESSAGES`

```bash
dscrd poll create [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--answer` | — | `[]` | poll answer (repeat 2-10 times) |
| `--channel` | ✓ | — | channel ID or name |
| `--duration` | — | `24` | poll duration in hours |
| `--multi` | — | — | allow selecting multiple answers |
| `--question` | ✓ | — | poll question |

> [!CAUTION]
> Write command — confirm with the user before executing; preview with `--dry-run`.

## dscrd poll end

End a poll immediately

**Discord API:** `POST /channels/{channel.id}/polls/{message.id}/expire`

```bash
dscrd poll end [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--channel` | ✓ | — | channel ID or name |
| `--message` | ✓ | — | poll message ID |

> [!CAUTION]
> Write command — confirm with the user before executing; preview with `--dry-run`.

## dscrd poll results

List voters for one poll answer

**Discord API:** `GET /channels/{channel.id}/polls/{message.id}/answers/{answer.id}`

```bash
dscrd poll results [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--answer` | ✓ | — | answer ID (1-based position) |
| `--channel` | ✓ | — | channel ID or name |
| `--message` | ✓ | — | poll message ID |


