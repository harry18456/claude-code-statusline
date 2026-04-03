## 1. debug-tee：修正 exit code propagation

- [x] [P] 1.1 修改 `cmd/debug-tee/main.go`：取得 `cmd.Run()` 的 error，以 Exit code 傳遞使用 ExitCode() 方案讀取 `cmd.ProcessState.ExitCode()` 並呼叫 `os.Exit()`（對應 debug-tee exit code propagation requirement）
- [x] [P] 1.2 commit `cmd/debug-tee/` 至 git

## 2. dev.sh：build 指令

- [x] 2.1 新增 `dev.sh`，實作 `build` subcommand：執行 `go build ./cmd/statusline/`，成功後複製 binary 至 `~/.claude/statusline.exe`（對應 dev.sh build command requirement，採用 dev.sh 作為單一入口腳本設計決策）
- [x] 2.2 `build` 失敗時印出錯誤並以非零 exit code 退出，不替換已安裝 binary

## 3. dev.sh：last-json 指令

- [x] 3.1 在 `dev.sh` 實作 `last-json` subcommand：從 `$TEMP/cc-statusline-debug.jsonl` 抽取最後一行，寫至 `./debug.json`（對應 dev.sh last-json command requirement，採用 last-json 寫到呼叫者 CWD 設計決策）
- [x] 3.2 log 檔不存在時印出明確錯誤訊息並以非零 exit code 退出
- [x] 3.3 實作無法識別或缺少 subcommand 時印出 usage 並以非零 exit code 退出（對應 Unknown subcommand scenario）
