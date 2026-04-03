## Why

本地開發流程缺乏便利工具：編譯並替換 binary 需要手動多步驟操作，而 `cmd/debug-tee` 雖已存在但有 exit code 未傳遞的缺陷，且沒有方便取得最新 JSON payload 的方式。

## What Changes

- 修正 `cmd/debug-tee/main.go`：正確傳遞 real binary 的 exit code
- 新增 `dev.sh`：提供 `build`（編譯 + 替換 `~/.claude/statusline.exe`）與 `last-json`（從 JSONL log 抽取最新一筆 JSON 至 `./debug.json`）兩個指令

## Non-Goals

- 不修改 `cmd/statusline/` 的任何行為
- 不影響 `install.sh`（release 下載流程）
- 不自動切換 `settings.json` 的 `command` 設定（仍由開發者手動切換）

## Capabilities

### New Capabilities

- `dev-workflow`：本地開發輔助腳本與 debug-tee 修正

### Modified Capabilities

（無）

## Impact

- Affected code: `cmd/debug-tee/main.go`、`dev.sh`（新增）
