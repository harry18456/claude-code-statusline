## Why

現有 statusline config 依賴 `CLAUDE_STATUSLINE_*` 環境變數；在 Windows 上不易設定，狀態藏在系統環境中，且和 Claude Code `settings.json` 的 `statusLine.command` 脫節。statusline 本來就是透過 command 啟動，config 應直接寫在 command 參數中。

使用者目前也只能靠 zero-value 自動省略段落，無法主動隱藏高噪音或暫時不需要的段落。

## What Changes

- **BREAKING**: 移除 `CLAUDE_STATUSLINE_ASCII`、`CLAUDE_STATUSLINE_NERDFONT`、`CLAUDE_STATUSLINE_POWERLINE` 作為本專案 config surface；這三個環境變數不再影響輸出。
- 新增 command flags：`--ascii`、`--nerdfont`、`--powerline`，分別對應既有 `renderer.Options` 的 ASCII、Nerd Font、Powerline 模式。
- 保留 `COLORTERM=truecolor|24bit` 偵測，因為它是系統標準能力訊號，不是本專案 config。
- 新增 `--hide <keys>`，以逗號分隔段落 key；預設全顯示，列出的有效段落即使有資料也不顯示。
- CLI config 採容錯渲染：config 問題寫 stderr 警告給 `claude --debug` 使用，但 stdout 仍輸出有效 statusline 並 exit 0。
- Unknown `--hide` key 只忽略該 key；已知 key 照常生效。
- `--ascii` 與 `--nerdfont`/`--powerline` 衝突時 ASCII 優先，作為最安全 fallback。
- 真正 flag 語法錯誤退回可用的安全 Options 並正常渲染；`--version` 仍印版本、exit 0、不讀 stdin、不渲染。
- README.md 與 README.zh-TW.md 明確標示 env var 到 flag 的 breaking upgrade path，並更新 `settings.json` command 範例。

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `go-statusline`: command-line config replaces project-specific env vars, renderer output supports explicit section hiding, and CLI config failures render tolerantly instead of blanking the statusline.
- `rate-limit-countdown`: ASCII-mode scenarios now use `--ascii` instead of the removed project-specific env var trigger.

## Impact

- Affected specs:
  - `openspec/specs/go-statusline/spec.md`
  - `openspec/specs/rate-limit-countdown/spec.md`
- Affected code for apply phase:
  - `cmd/statusline/main.go`
  - `internal/renderer/renderer.go`
  - `internal/renderer/renderer_test.go`
  - `README.md`
  - `README.zh-TW.md`
- No new runtime dependencies; the design uses Go standard library flag parsing.
- Apply-phase verification target: `go build ./...`, `go vet ./...`, `staticcheck ./...`, `go test -race ./...`, `gofmt -l .`, and real binary checks for `--nerdfont`, `--hide effort`, malformed config tolerance, and no-flag default rendering.
