## Why

跨機台的 session 用本機名稱找不到。2026-11-20 的實驗證實，別台機器只能用 `bridge:<bridgeSessionId>` 送達。這個 ID 只存在本機 session registry，而且只有 session 連上 Remote Control（RC）時才有值：關掉 RC 約 2 秒內變成 `null`，重新連上會換成新值。statusline 目前只顯示本機名稱，使用者沒辦法從畫面上得知跨機台要用的地址。

## What Changes

- session registry 查詢多讀同一筆紀錄的 `bridgeSessionId`。值必須是 1 到 128 個字元、只含 ASCII 英數字、`_`、`-` 的字串才採用；`null`、其他型別或不合格字元都視為沒有 bridge 地址，而且不影響名稱比對。
- RC 連線時，line 2 的 session 段落在名稱後面加上 bridge 地址：預設與 Nerd Font 模式為 `⇄ <name> · bridge:<id>`（同一段灰色），ASCII 模式為 `peer:<name> bridge:<id>`。
- 新增 `--hide bridge`：只隱藏 bridge 部分，名稱照常顯示。`--hide session` 仍然隱藏整段，並跳過 registry 讀取。
- 更新 `README.md`、`README.zh-TW.md`、`CLAUDE.md`。

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `go-statusline`: session 段落在 RC 連線時加上 bridge 地址；`--hide` 新增 `bridge` key。

## Impact

- Affected specs: `openspec/specs/go-statusline/spec.md`
- Affected code:
  - `internal/peername/peername.go`、`internal/peername/peername_test.go`
  - `internal/renderer/renderer.go`、`internal/renderer/renderer_test.go`
  - `cmd/statusline/main.go`、`cmd/statusline/session_test.go`
  - `README.md`、`README.zh-TW.md`、`CLAUDE.md`
- RC 連線時 line 2 會變長（bridge 地址約 40 字元）；可以用 `--hide bridge` 關掉。
- 每次重新連上 RC，地址都會改變。statusline 每次刷新都重讀 registry，所以永遠顯示當下的地址。
