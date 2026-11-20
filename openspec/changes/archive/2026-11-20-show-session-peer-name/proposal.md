## Why

Claude Code 的跨 session 溝通（`ListAgents` 找人、`SendMessage(to: "<name>")` 傳話）用 session 的**定址名稱**當地址，不是 session id，也不是 AI 標題。使用者希望不同 session 能互相找到對方，但 statusline 目前沒有顯示這個名稱。

payload 的 `session_name` 不能直接拿來用：已擷取的樣本裡它是 AI 標題，和定址名稱不同。例如互動 session `my-project-75` 的 AI 標題是「重構設定載入」，使用者若照標題傳訊會找不到人。定址名稱只存在 Claude Code 的 session registry（`<config>/sessions/<pid>.json` 的 `name`），`ListAgents` 顯示的名稱與它完全一致（2026-11-20 於 CLI 2.1.329 實測）。

## What Changes

- 解析 payload 頂層 `session_id`。欄位缺少、為 null 或型別錯誤時視為無值，不影響其他欄位的解析。
- 新增 session registry 查詢：用 `session_id` 到 `<config>/sessions/*.json` 找 `sessionId` 相同的那一筆，取其 `name`。`<config>` 依序為 `CLAUDE_CONFIG_DIR`（有設定時）與 `~/.claude`。
- 查詢一律容錯：目錄不存在、檔案空白或半寫入、JSON 格式錯誤、找不到相符項目，都視為「沒有名稱」。不輸出警告，也不影響其他段落。
- Line 2 最後新增 session 段落：預設與 Nerd Font 模式顯示 `⇄ <name>`（灰色），ASCII 模式顯示 `peer:<name>`。沒有名稱時整段省略，不留分隔符。
- `--hide` 新增 key `session`。隱藏時不讀取 registry。
- 更新 `README.md`、`README.zh-TW.md`、`CLAUDE.md` 的 line 2 結構說明與 hide key 清單。

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `go-statusline`: 解析 `session_id`、從 session registry 解析定址名稱並顯示在 line 2 最後；`--hide` 新增 `session` key。

## Impact

- Affected specs: `openspec/specs/go-statusline/spec.md`
- Affected code:
  - `internal/model/payload.go`、`internal/model/payload_test.go`
  - `internal/peername/peername.go`、`internal/peername/peername_test.go`（新增）
  - `internal/renderer/renderer.go`、`internal/renderer/renderer_test.go`
  - `cmd/statusline/main.go`、`cmd/statusline/main_test.go`
  - `README.md`、`README.zh-TW.md`、`CLAUDE.md`
- 依賴：只使用 Go 標準函式庫。
- 執行成本：每次渲染多讀 `<config>/sessions/` 下的小 JSON 檔（實測 6 個、每個小於 1 KB）。`session` 被隱藏或 payload 沒有 `session_id` 時不做任何檔案讀取。
- 風險：session registry 是 Claude Code 未公開的內部格式，改版可能變動。格式變動時這一段會自動不顯示，其他輸出不受影響。
