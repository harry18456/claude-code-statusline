## Context

專案目前缺乏本地開發快捷工具。`cmd/debug-tee/` 已存在但：（1）exit code 未從 real binary 傳遞，（2）沒有方便把最新 payload 取出到工作目錄的手段。每次要測試都需要手動 `go build` 再複製 binary。

## Goals / Non-Goals

**Goals:**

- 修正 debug-tee exit code 傳遞
- 提供一鍵 build + install 指令
- 提供一鍵取出最新 JSON payload 至 `./debug.json`

**Non-Goals:**

- 不自動切換 `~/.claude/settings.json`
- 不提供 watch mode 或持續重建
- 不修改 release / CI 流程

## Decisions

### dev.sh 作為單一入口腳本

選擇單一 `dev.sh` 而非 Makefile 或多個獨立腳本。

**理由：** 開發者使用 Git Bash（Windows），`make` 需額外安裝；單一腳本用 subcommand 模式（`./dev.sh build`、`./dev.sh last-json`）比多個腳本更好維護，且不需要 PATH 設定。

**替代方案：** Makefile — 拒絕，Windows 環境需額外依賴。

### last-json 寫到呼叫者 CWD

`./dev.sh last-json` 從 `$TEMP/cc-statusline-debug.jsonl` 抽取最後一行，寫到執行指令的 shell 目前目錄下的 `./debug.json`。

**理由：** 開發者在專案根目錄執行指令，debug.json 自然落在那裡，可直接用 `cat debug.json | ./statusline.exe` 重現問題。debug-tee 本身不知道開發者的目錄，分兩步最乾淨。

### Exit code 傳遞使用 ExitCode()

`cmd.Run()` 後取 `cmd.ProcessState.ExitCode()`，以該值呼叫 `os.Exit()`。

**替代方案：** 直接呼叫 `os.Exit(1)` on error — 拒絕，會遮蔽 real binary 的正確 exit code。

## Risks / Trade-offs

- [Risk] `cc-statusline-debug.jsonl` 不存在時 `last-json` 會印出錯誤訊息 → 以清楚的錯誤提示處理，不靜默失敗
- [Trade-off] `dev.sh` 依賴 bash，在 PowerShell 下不能直接執行 → 使用者已使用 Git Bash，可接受
