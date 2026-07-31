# dscrd — Agent-facing Discord CLI 設計文件

> 狀態：已上線（v0.2.0，2026-07-31）。發布通路：GitHub Releases、npm `@howar31/dscrd`、Homebrew（`brew install howar31/tap/dscrd`）、`go install`。本文件為原始設計規格；現行架構的權威文件是 SPEC.md。

## 1. 定位

`dscrd` 是一個給 AI agent（也適用於人類）在終端機操作 Discord 的 CLI 工具：

- **Bot token only**：Discord ToS 明文禁止 user token 自動化（self-bot，違者永久停權個人帳號），因此本工具僅支援 bot token。定位是「Discord 伺服器操作 CLI」，bot 的可見範圍 = 被邀請進入的伺服器。
- **Stateless one-shot**:一指令一 API 呼叫，純 REST（`https://discord.com/api/v10`），不開 websocket、無常駐進程。
- **Token-efficient output**:預設輸出為精簡單行格式，為 LLM agent 的 context 成本最佳化；完整 JSON 為 opt-in。

### 非目標（明確排除）

- User token 自動化（ToS 禁止）。
- Presence / 線上狀態讀寫（無 REST endpoint，需 gateway）。
- 跨 guild / 全域搜尋、讀取使用者個人 DM 收件匣（bot 權限模型不允許）。
- 常駐 bot / 事件監聽（那是 bot framework 的職責，不是 CLI）。

## 2. 認證

- **Token 型式**：`Authorization: Bot <token>`，從 Discord Developer Portal 複製，無需 OAuth flow。
- **儲存**：`~/.config/dscrd/config.toml`（mode 0600），TOML profile 制，可管理多個 bot。
- **加密**：token 以 AES-256-GCM 加密存放。加密金鑰後端由 `DSCRD_KEYRING_BACKEND` 控制：
  - `auto`（預設）— 有 OS keyring 就用（macOS Keychain / Linux Secret Service / Windows Credential Manager），否則落到檔案。
  - `keyring` — 強制 OS keyring。
  - `file` — 金鑰存 `~/.config/dscrd/.encryption_key`（mode 0600）。
- **Token 解析優先序**（高到低）：`DSCRD_TOKEN` env → `--profile <name>` flag → `DSCRD_PROFILE` env → config 內 active profile。
- **Profile 欄位**：token（加密）、application_id（產 invite URL 用）、default_guild（選填）。
- **Profile 名稱**：寫入時限 `[A-Za-z0-9_-]+`（載入不驗，舊 config 不受影響）。`set-token` 對已存在的 profile 需 `--force`，因為 token 存後讀不回來、誤蓋無法復原。
- **子指令**：
  - `auth set-token` — 貼上 token（隱藏輸入）。
  - `auth status` — 列出 profiles、對 Discord 驗證 active token（`GET /users/@me`）。
  - `auth test` — 明確的 live token 檢查。
  - `auth switch <name>` / `auth logout` — 切換 / 移除 profile。
  - `auth rename <old> <new> [--force]` — 改名 profile;若改的是 active,`active` 一併跟著改。
  - `auth invite-url [--permissions <bits|preset>]` — 產生把 bot 邀進伺服器的 OAuth2 連結。
- **保證**：token 絕不印出、不落 log；client 端任何錯誤訊息都不得含 token。

## 3. 指令面

分三個 tier 逐步實作；`api` escape hatch 自 v1 起即可觸達所有官方公開 endpoint，功能完整性不受包裝進度限制。

### Tier 1（v1 核心）

| 群組 | 子指令 | 說明 |
|---|---|---|
| `auth` | set-token, status, test, switch, logout, invite-url | 憑證管理 |
| `guild` | list, info | bot 所在伺服器（`GET /users/@me/guilds`） |
| `channel` | list, info, create, edit, delete, topic | 頻道管理（含 category、forum） |
| `msg` | read, send, edit, delete, react, unreact, reactions, pin, unpin, pins, permalink | 訊息讀寫；`send --reply-to <msgID>` 回覆；`send --file <path>` 附件上傳 |
| `thread` | list, read, reply, create, archive | 討論串（Discord thread 本質是 channel）；forum post 即 thread + 首則訊息 |
| `search` | messages | Guild 訊息搜尋（`GET /guilds/{id}/messages/search`，2025-08 起開放 bot；需 MESSAGE_CONTENT intent；處理 index 未熱的 202 + retry_after / error 110000 重試） |
| `api` | `dscrd api <METHOD> <path> [--body <json>]` | Escape hatch：任意官方 endpoint |
| `version` | (default), --check | 版本資訊 |

### Tier 2

| 群組 | 子指令 | 說明 |
|---|---|---|
| `member` | list, info, search | 伺服器成員（需 GUILD_MEMBERS intent） |
| `role` | list, info, create, edit, delete, assign, unassign | 身分組 |
| `emoji` | list, info | 自訂 emoji |
| `file` | download | 附件下載（CDN URL）；上傳走 `msg send --file` |
| `dm` | send, read | 以 bot 身分開 DM channel 並收發（限共同伺服器成員；bot 無法列出既有 DM 清單） |
| `user` | info, me | 使用者資訊 |

### Tier 3

| 群組 | 子指令 | 說明 |
|---|---|---|
| `invite` | list, create, delete, info | 邀請連結 |
| `webhook` | list, create, delete, execute | Webhook 管理 |
| `audit-log` | read | 稽核紀錄 |
| `event` | list, create, edit, delete | 排程活動（scheduled events） |
| `sticker` | list, info | 貼圖 |
| `poll` | create, results, end | 投票 |
| `automod` | list, info, create, edit, delete | 自動審核規則 |

### 全域 flags

- `--format concise|json|jsonl|table`（預設 concise）
- `--guild <name|id>` — 目標伺服器；未指定時用 profile 的 default_guild
- `--profile <name>` — 指定 config profile
- `--raw` — 回傳 Discord 原始 JSON
- `--dry-run` — 不打 API，印出即將執行的呼叫
- `--no-resolve` — 不做 ID→名稱解析
- `--limit` / `--before` / `--after` — 讀取類指令的分頁控制

## 4. 架構

Go 1.25+、Cobra、單一靜態 binary。REST client 手寫（不依賴 discordgo 等 framework）——Discord REST 端點面小且穩定，手寫換來對新 endpoint 的即時支援與最小依賴。

```
cmd/dscrd/main.go        # 進入點：建 root command、執行、錯誤→exit code 映射
internal/
├── api/                 # REST client：base URL、auth header、429 retry、錯誤映射、multipart 上傳
├── auth/                # 憑證加密儲存、profile 管理、token 解析
├── commands/            # 指令樹：一群一檔 + 對應 _test.go
├── output/              # concise/json/jsonl/table 四種 emitter
├── resolve/             # snowflake ID→名稱快取（~/.config/dscrd/cache/resolve.json）
└── skillgen/            # 從 Cobra 指令樹自動產生 agent skills
```

### 指令 annotations

每個 Cobra 指令掛 metadata，驅動 skillgen 與行為：

- `discordEndpoint` — 對應的 API 路由（如 `POST /channels/{id}/messages`）
- `write` — `"true"` 表示寫入操作（skill 內產生 CAUTION 提示、`--dry-run` 支援）
- `requiredPerms` — 所需 Discord permissions（如 `SEND_MESSAGES`）
- `intents` — 所需 privileged intents（如 `MESSAGE_CONTENT`）

### skillgen

`go run ./cmd/dscrd generate-skills` 從指令樹 + 模板產生：

- `skills/dscrd/SKILL.md` — 索引（語法、指令群目錄）
- `skills/dscrd-shared/SKILL.md` — 全域 flags、安全規則、exit codes
- `skills/dscrd-<group>/SKILL.md` — 每群一份 verb reference

CI 有 drift guard：skills 目錄與指令樹不同步即 fail。

## 5. 輸出設計

- **concise（預設）**，單行一項：
  - 訊息：`User: text [MM-DD HH:MM] (msgID)` — Discord 的回覆/reaction/pin 都靠 message ID，必須帶出。
  - 頻道：`#name (id) — type, topic 摘要`
  - 成員：`name (id) — roles 摘要`
  - 寫入確認：單行（如 `sent <msgID>`）。
- **ID 解析**：channel/user/role/guild snowflake 預設解析為名稱，快取於 `~/.config/dscrd/cache/resolve.json`，查詢失敗時退回顯示 ID。`--no-resolve` 跳過。
- **讀取預設保守**：`msg read` 預設 `--limit 5`，agent 需要更多再明確要求。
- **多行輸入**：`--text-file <path>` 或 `-`（stdin），避免 shell quoting 問題。

## 6. 錯誤處理

- **Exit codes**：`0` ok、`3` auth 錯誤、`4` not found、`5` rate-limited、`1` 其他。
- **429**：讀取 `retry_after` 自動重試，最多 3 次；仍失敗則 exit 5。
- **Search index 未熱**：`202` + `retry_after`（error code `110000`）自動等待重試。
- **Intent 提示**：權限類 403/50001 錯誤時，若該指令標注了 privileged intent，錯誤訊息附上「需在 Developer Portal → Bot → Privileged Gateway Intents 開啟 MESSAGE_CONTENT」等具體指引——讓 agent 拿到錯誤就知道怎麼修，不用再查文件。
- **DM 失敗**：`50007 Cannot send messages to this user` 附說明（需共同伺服器、對方未關閉成員私訊）。

## 7. 測試策略

**硬性要求：所有指令都必須有測試覆蓋，以自動化手段保證，不靠自律。**

1. **單元測試**：每個指令群一個 `_test.go`，以 `httptest` mock Discord API，table-driven 覆蓋：正常路徑、API 錯誤映射、輸出格式（concise/json）、`--dry-run`。
2. **覆蓋保證 meta-test**：一個測試走訪整棵 Cobra 指令樹，比對「已註冊 leaf 指令集合」與「測試註冊表（每個測試向其登記所覆蓋的指令）」，有任何 leaf 指令未被登記即 fail。新增指令而沒寫測試 → CI 直接紅燈。
3. **API client 測試**：429 retry、202 search retry、錯誤映射、multipart 上傳。
4. **auth 測試**：加密 round-trip、token 優先序、profile 切換。
5. **Live smoke test**（手動/選跑）：`scripts/smoke.sh` 對真實測試伺服器跑 read → send → react → search → delete 全流程，release 前執行。

CI（GitHub Actions）：`go vet` + `go test ./...`（含 meta-test）+ skills drift guard。

## 8. 發佈

- **GitHub Releases**：goreleaser 產 darwin/linux/windows 多平台 binary。
- **npm**：`@howar31/dscrd`（prebuilt binary + SHA256 驗證的 wrapper package）。
- **go install** 直接可用。
- **License**：MIT。
- **Skills 安裝**：`npx skills add <repo-url>` 或手動複製 `skills/` 目錄。

## 9. 實作階段

1. **Phase 1 — 骨架**：repo 初始化、api client（含 429 retry）、auth（加密儲存 + profile）、output、exit codes、meta-test 框架。
2. **Phase 2 — Tier 1 指令**：guild / channel / msg / thread / search / api，各含測試。
3. **Phase 3 — skillgen**：skills 產生器 + CI drift guard。
4. **Phase 4 — Tier 2 指令**：member / role / emoji / file / dm / user。
5. **Phase 5 — 發佈管線**：goreleaser、npm wrapper、README、smoke test。
6. **Phase 6 — Tier 3 指令**：invite / webhook / audit-log / event / sticker / poll / automod。

每個 Phase 完成的定義：指令可用 + 測試綠 + meta-test 通過。
