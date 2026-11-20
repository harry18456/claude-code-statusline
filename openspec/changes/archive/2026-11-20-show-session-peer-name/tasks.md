## 1. Payload 解析

- [x] [P] 1.1 為 `Parse session id field` 寫 model 測試（RED）：涵蓋 `session_id` 為字串、缺少、`null`、數字 `123`、物件 `{"id":"x"}`，後四者的 `SessionID` 為空字串且 `ParsePayload` 不回傳錯誤、其他欄位照常解析；驗證 `go test ./internal/model -run TestParsePayload_SessionID` 在實作前失敗。
- [x] 1.2 實作 `session_id 容錯解析`：`model.Payload.SessionID` 以匯出型別 `TolerantString` 對應 JSON 頂層 `session_id`，非字串型別一律視為空字串；驗證 `go test ./internal/model` 全綠（含新測試與既有測試）。

## 2. Session registry 查詢

- [x] [P] 2.1 為 `Resolve session peer name from the session registry` 寫 `internal/peername` 測試（RED）：涵蓋相符項目、`CLAUDE_CONFIG_DIR` 優先、退回 `<home>/.claude`、多筆相符取 `updatedAt` 最大、`updatedAt` 相同或缺少時取檔名最前、控制字元清除（`"peer\nname\u001b[31m"` → `peername[31m`）、空 session id 不讀檔，以及 spec example 表格每一列（目錄不存在、全空白檔、半寫入檔、空名稱、名稱型別錯誤、`.key` 檔、65537 bytes 大檔、大小寫不同）；`ConfigDirs` 測試涵蓋未設定、只含空白、與 home 相同時去重、`home` 為空；驗證 `go test ./internal/peername` 在實作前失敗。
- [x] 2.2 實作 `新增 internal/peername 套件封裝 registry 查詢`：`ConfigDirs(getenv, home)` 依 `Config 目錄順序：CLAUDE_CONFIG_DIR 再 ~/.claude` 回傳去重後的目錄；`Lookup(sessionID, configDirs)` 遵守 `只讀 sessions 目錄下的 .json 一般檔案`、`多筆相符時取 updatedAt 最大者`、`清除名稱中的控制字元`，任何失敗都回傳空字串、不 panic、不寫 stderr；並依 `Windows 以 FILE_SHARE_DELETE 開啟 registry 檔` 實作 `openShared`（讀檔期間 Claude Code 仍可刪除與原地改寫該檔）；驗證 `go test -race ./internal/peername` 全綠（含 `TestOpenSharedAllowsRewriteAndDeleteWhileOpen`）。

## 3. Renderer 顯示

- [x] [P] 3.1 為 `Session peer name display` 與 `Render two-line ANSI output` 的 line 2 順序寫 renderer 測試（RED）：預設模式 line 2 結尾為 `\033[90m⇄ claude-code-statusline-56\033[0m`；Nerd Font 模式使用相同 glyph；ASCII 模式結尾為 `peer:claude-code-statusline-56` 且該段無 ANSI；段落位於 `⚙ code-reviewer` 與 worktree 指示之後且只隔一個分隔符；名稱為空時 line 2 不含 `⇄`、`peer:` 且不以分隔符結尾；`HiddenSections` 含 `session` 時省略；驗證 `go test ./internal/renderer -run TestRenderSessionPeerName` 在實作前失敗。
- [x] 3.2 實作 `Render 簽章新增 SessionInfo 參數` 與 `段落格式：預設 ⇄ <name> 灰色、ASCII peer:<name>`：新增 `renderer.SessionInfo{PeerName string}`、`SectionSession = "session"`（加入合法 hide key）、匯出方法 `Options.Hidden(key)`，`Render(p, git, session, opts)` 在 line 2 最後輸出 session 段落；既有測試透過 `renderWith` helper 傳入空 `SessionInfo`；驗證 `go test ./internal/renderer` 全綠。

## 4. CLI 整合

- [x] 4.1 為 `Configurable section visibility` 的 `session` key 與 `隱藏 session 時跳過 registry 讀取` 寫 `cmd/statusline` 測試（RED）：`--hide session` 不產生 unknown-key 警告；`resolveSession` 在 `session` 隱藏或 `session_id` 為空時不呼叫 lookup，否則以 `ConfigDirs` 結果呼叫一次；端對端用暫存 `CLAUDE_CONFIG_DIR` registry 時 line 2 顯示 registry 名稱，且 payload `session_name` 與 registry 名稱不同時顯示 registry 名稱（`從 session registry 解析定址名稱，不用 payload session_name`）；registry 只有壞檔時 stderr 為空、exit code 為 0；驗證 `go test ./cmd/statusline -run 'TestResolveSession|TestMainSessionPeerName'` 在實作前失敗。
- [x] 4.2 實作 CLI 整合：`runWithGit` 解析 payload 後呼叫 `resolveSession(cfg.Options, p.SessionID, getenv, os.UserHomeDir, peername.Lookup)` 取得 `SessionInfo`，再呼叫新的 `Render`；`--hide session` 時不讀 registry；驗證 `go test ./cmd/statusline` 全綠。

## 5. 文件

- [x] [P] 5.1 更新 `README.md` 與 `README.zh-TW.md`：line 2 範例加入 session 段落、line 2 表格新增 session 列（說明它是 `ListAgents` / `SendMessage` 的定址名稱、來源是 Claude Code 內部 session registry、找不到時隱藏、ASCII 為 `peer:<name>`）、`--hide` key 清單加入 `session`；驗證：人工比對兩份 README 對應段落一致，`grep -n "session" README.md README.zh-TW.md` 可找到新內容。
- [x] [P] 5.2 更新 `CLAUDE.md` 的 Output structure line 2 與 `--hide` key 清單，加入 session 段落與 `session` key；驗證 `grep -n "session" CLAUDE.md` 可找到新內容。

## 6. 驗收

- [x] 6.1 跑完整測試矩陣：`go build ./...`、`go vet ./...`、`go test -race ./...` 通過，`gofmt -l .` 無輸出，`staticcheck ./...` 若本機可用也通過；驗證：記錄每個命令的實際輸出。
- [x] 6.2 用真 binary 實跑 design 的 5 個驗收情境：構造 registry 顯示 `⇄ <name>`、`--ascii` 顯示 `peer:<name>`、`--hide session` 無名稱且無多餘分隔符、壞檔／空白檔／`.key` 檔時無名稱且 stderr 空且 exit 0、真實 registry 搭配本 session id 顯示的名稱與 `ListAgents` 一致；驗證：記錄每個情境的命令、stdout、stderr 與 exit code。

## 7. Review 修正

- [x] 7.1 依 `updatedAt 寬鬆解析`：`updatedAt` 為字串、null 或物件時當成 0 且仍可相符，數字較大者照常勝出；驗證 `go test ./internal/peername -run TestLookupLenientUpdatedAt` 通過（修正前 string 與 object 兩案失敗）。
- [x] 7.2 擴充 `清除名稱中的控制字元`：移除 Unicode 類別 Cc、Cf、Zl、Zp（bidi 覆寫與隔離、零寬字元、U+2028、U+2029），且先移除再 trim；測試名稱以 `rune` 常數與 `json.Marshal` 產生，原始碼不含隱形字元；驗證 `go test ./internal/peername -run TestLookupSanitizesName` 通過（修正前 4 案失敗）。
- [x] 7.3 依 `開檔後再確認是一般檔案`：`readEntry` 在讀取前對 handle 做 `Stat`，pipe 與裝置直接跳過；Windows `openShared` 在 `CreateFile` 失敗時改用 `os.Open`（長路徑）；驗證 `TestReadEntrySkipsNamedPipe`（修正前 3 秒逾時失敗）與 `TestOpenSharedLongPath` 通過。
- [x] 7.4 補強測試與文件一致性：優先順序測試讓後面目錄的 `updatedAt` 較大仍預期前面目錄勝出；測試改用含字母的假 session id；renderer 測試改用 `\033[90m` 字面值與固定預期字串，補 line 2 完整順序、隱藏 line 1 時 session 仍顯示、全隱藏 line 2 時 line 1 不受影響；CLI 測試同時攔截 `os.Stderr` 與 `log` 輸出；移除多餘的 `renderWithSession` 與 `found`；`SessionInfo` 註解寫明 caller 須傳入已清除的名稱；`open_windows.go` 註解改正（覆蓋改名仍會被拒）；驗證 `go test -race ./...` 全綠且 `gofmt -l .` 無輸出。
- [x] 7.5 重跑驗收：`go build ./...`、`go vet ./...`（windows、linux、darwin）、`staticcheck ./...`、`go test -race ./...`、`gofmt -l .`，以及真 binary 的 5 個情境加上「名稱含 bidi 覆寫字元」情境；驗證：記錄每個命令與情境的輸出。
- [x] 7.6 複審 LOW 修正：新增 `TestOpenSharedFallsBackToOSOpen`（以目錄讓 `CreateFile` 失敗、證明 `os.Open` 退回路徑在任何長路徑設定下都會執行；突變測試確認拿掉退回時測試失敗）；`captureProcessStderr` 在呼叫前先開 goroutine 讀 pipe，並用 `t.Cleanup` 還原 `os.Stderr` 與 logger；驗證 `go test -race ./...` 全綠。
