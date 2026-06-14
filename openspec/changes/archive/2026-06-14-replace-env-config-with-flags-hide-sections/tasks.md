## 1. CLI Config Parsing

- [x] [P] 1.1 為 `Command-line configuration` 建立 main-level parser tests：覆蓋 `--ascii`、`--nerdfont`、`--powerline`、`--hide`、`--version`、default options、removed `CLAUDE_STATUSLINE_*` env vars、`COLORTERM` true color、unknown flag、missing `--hide` value、unknown hide key、positional arg、`--ascii` conflict；驗證 invalid config 會產生 warning 與容錯 options，而不是 exit 2 contract；以 `go test ./cmd/statusline -run TestParseCLIConfig` 驗證。
- [x] 1.2 實作 `Use the Go standard flag package` 與 `Tolerant CLI config never blanks the statusline`：在讀 stdin 前用 `flag.NewFlagSet` 解析 flags，`--version` 印出版本並 exit 0，其他 invalid config 寫 stderr warning、繼續讀 stdin、stdout 輸出有效 statusline、exit 0；以 `go test ./cmd/statusline -run 'Test(ParseCLIConfig|MainToleratesCLIConfig)'` 驗證。
- [x] 1.3 實作 `Preserve COLORTERM as system capability detection` 與 `Preserve current powerline implication`：`COLORTERM=truecolor|24bit` 仍啟用 true color，`--nerdfont` 啟用 Nerd Font 並隱含 Powerline，`--powerline` 只啟用 separator，`--ascii` 衝突時 ASCII 優先並警告；以 `go test ./cmd/statusline -run TestParseCLIConfig` 驗證。

## 2. Renderer Section Visibility

- [x] [P] 2.1 為 `Configurable section visibility` 與 `Render two-line ANSI output` 建立 renderer tests：每個 hide key 都在有資料的 payload 下隱藏對應段落，default hidden set 保持全顯示，hidden adjacent sections 不產生 leading/trailing/duplicate separators；以 `go test ./internal/renderer -run TestRenderHideSections` 驗證。
- [x] 2.2 實作 `Define canonical hide section keys` 與 `Represent hidden sections as renderer options`：`renderer.Options` 接收 canonical hidden set，renderer 不解析 raw `--hide` 字串，unknown hide key 在 CLI parser 階段警告並忽略，已知 key 照常生效；以 `go test ./cmd/statusline -run TestParseCLIConfig` 與 `go test ./internal/renderer -run TestRenderHideSections` 驗證。
- [x] 2.3 實作 `Apply hide at segment assembly boundaries`：只在 `Render` 組段落時 gating，保留 cost/cache/effort/duration/rate/bar/dir formatter 行為；以 `go test ./internal/renderer -run 'Test(RenderHideSections|FormatCacheHit|FormatExecutionMode|FormatRate|ComputePaceArrow)'` 驗證。
- [x] 2.4 覆蓋 `Prompt cache hit-rate display` 與 cost/cache 獨立 hide contract：both visible、cost hidden only、cache hidden only、both hidden 都符合 spec；以 `go test ./internal/renderer -run TestRenderHideCostCache` 驗證。
- [x] 2.5 覆蓋 `Execution mode display` 與 `Seven-day usage pace indicator` 在 hide/ASCII 下的契約：`--hide effort` 不顯 effort，`--hide rate` 不顯 5h/7d rate、countdown、pace indicator，ASCII-derived options 的 pace indicator 仍輸出 `^`/`v`/`~`；以 `go test ./internal/renderer -run 'Test(RenderHideEffort|RenderHideRate|ComputePaceArrow_ASCII)'` 驗證。
- [x] 2.6 覆蓋 `Three-tier color rendering` 與 `Nerd Font and Powerline support`：`--ascii` 導出的 Options 產生 ASCII bar，`--nerdfont` 導出的 Options 產生 Nerd Font output，`--powerline` 導出的 Options 產生 Powerline separators；以 `go test ./cmd/statusline -run TestParseCLIConfig` 與 `go test ./internal/renderer -run 'TestRenderASCIIBar|TestRenderNerdFont|TestRenderPowerline'` 驗證。

## 3. Documentation Migration

- [x] 3.1 更新 README.md 的 breaking migration：移除 `CLAUDE_STATUSLINE_ASCII`、`CLAUDE_STATUSLINE_NERDFONT`、`CLAUDE_STATUSLINE_POWERLINE` 設定教學，改成 `settings.json` command flags 與 `--hide effort,duration,rate` 範例，保留 `COLORTERM` 為 terminal capability detection，並說明 invalid config 只警告且仍渲染；以內容 review 和 `rg -n "CLAUDE_STATUSLINE_(ASCII|NERDFONT|POWERLINE)" README.md` 只命中 breaking note 驗證。
- [x] 3.2 更新 README.zh-TW.md 的同等 breaking migration：繁中說明 env var 已廢棄、flag 對照、hide key 清單、unknown key 會被忽略並警告、statusline 不會因 config typo 消失；以內容 review 和 `rg -n "CLAUDE_STATUSLINE_(ASCII|NERDFONT|POWERLINE)" README.zh-TW.md` 只命中 breaking note 驗證。

## 4. Final Verification

- [x] 4.1 執行完整靜態與測試驗證：`gofmt -l .` 無輸出，`go build ./...`、`go vet ./...`、`staticcheck ./...`、`go test -race ./...` 全部通過。
- [x] 4.2 執行真 binary contract 驗證：建置實際 `statusline` binary，使用代表性 payload 驗證 `--nerdfont` 顯示 Nerd Font、`--hide effort` 不顯 effort、無 flag 預設全顯示、舊 `CLAUDE_STATUSLINE_*` env vars 不改變輸出、unknown hide key 仍輸出 statusline 並警告、`--ascii --nerdfont` 仍輸出 ASCII statusline 並警告、flag syntax error 仍輸出有效 statusline；以記錄的命令輸出摘要完成 review。
