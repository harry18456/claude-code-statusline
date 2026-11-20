## 1. Registry 查詢

- [x] [P] 1.1 為 `Resolve Remote Control bridge address` 寫 `internal/peername` 測試（RED）：`Lookup` 回傳 `Info`；bridge 為合格字串時保留；`null`、缺少、數字、空字串、含空白、含 ESC、129 字元時 `BridgeSessionID` 為空但 `Name` 照常；`session_01-_Ab9` 合格；兩筆相符時 bridge 只取勝出那筆（勝出那筆為 `null` 就沒有 bridge）；驗證 `go test ./internal/peername -run 'TestLookupBridge'` 在實作前失敗。
- [x] 1.2 依 `peername.Lookup 改回傳 Info 結構`、`bridge 地址取自同一筆 registry 紀錄`、`bridgeSessionId 白名單驗證` 實作：`registryEntry` 以寬鬆字串解碼 `bridgeSessionId`，勝出紀錄的值通過白名單（1–128 個 `[A-Za-z0-9_-]`）才放進 `Info.BridgeSessionID`；既有名稱測試改用 `Info.Name`；驗證 `go test -race ./internal/peername` 全綠。

## 2. Renderer 顯示

- [x] [P] 2.1 為 `Session peer name display` 的 bridge 情境與 `Configurable section visibility` 的 `bridge` key 寫 renderer 測試（RED）：預設模式結尾為 `\033[90m⇄ <name> · bridge:<id>\033[0m`；Nerd Font 同預設；ASCII 結尾為 `peer:<name> bridge:<id>` 且無 ANSI；`--hide bridge` 只剩名稱且不含 `bridge:`；`--hide session` 整段消失；名稱為空但有 bridge 時不顯示任何 session 文字；`SectionKeyForName("bridge")` 合法；驗證 `go test ./internal/renderer -run 'TestRenderBridge|TestSectionKeyForNameBridge'` 在實作前失敗。
- [x] 2.2 依 `顯示格式：同一段內以「 · 」接在名稱後` 與 `新增 --hide bridge，只藏 bridge 部分` 實作：`SessionInfo` 新增 `BridgeSessionID`、新增 `SectionBridge`、`formatSessionPeerName` 依模式與隱藏狀態組出 bridge 後綴；驗證 `go test -race ./internal/renderer` 全綠。

## 3. CLI 整合

- [x] 3.1 為 CLI 寫測試（RED）：`resolveSession` 把 `Info` 的兩個欄位都放進 `SessionInfo`；端對端用暫存 `CLAUDE_CONFIG_DIR` registry：bridge 為字串時 line 2 以 `⇄ demo-project-07 · bridge:session_01EXAMPLEb2kQ7mZp4Lx9VdR` 結尾；`--hide bridge` 只剩名稱；`bridgeSessionId` 為 `null` 時只剩名稱；全部情況 stderr 為空、exit 0；驗證 `go test ./cmd/statusline -run 'TestResolveSession|TestMainBridge'` 在實作前失敗。
- [x] 3.2 實作 CLI 整合：`resolveSession` 的 lookup 參數改為回傳 `peername.Info`，並填入 `SessionInfo.BridgeSessionID`；驗證 `go test -race ./cmd/statusline` 全綠。

## 4. 文件

- [x] [P] 4.1 更新 `README.md` 與 `README.zh-TW.md`：定址名稱列加上 RC 連線時的 bridge 地址說明（只有同帳號 session 能用、重新連上 RC 會換新地址、跨機台要用 bridge 地址），hide key 清單加入 `bridge`；驗證 `grep -n "bridge" README.md README.zh-TW.md`。
- [x] [P] 4.2 更新 `CLAUDE.md` 的 line 2 結構、peer name lookup 段落與 `--hide` key 清單；驗證 `grep -n "bridge" CLAUDE.md`。

## 5. 驗收

- [x] 5.1 跑完整測試矩陣：`go build ./...`、`go vet ./...`（windows、linux、darwin）、`staticcheck ./...`、`go test -race ./...`、`gofmt -l .`；驗證：記錄每個命令的輸出。
- [x] 5.2 真 binary 實跑：構造 registry 的字串、`null`、不合格三種 bridge 值；`--ascii`、`--hide bridge`、`--hide session`；本機真實 registry 中開著 RC 的 session 顯示的 bridge 地址與 registry 一致；驗證：記錄命令、stdout、stderr 與 exit code。
- [x] 5.3 審查修正：補 `TestLookupBridgeWinnerReadFirst`（勝出檔排在前面、同分取第一個檔、不從後面目錄借 bridge）與 `TestValidBridgeIDAllowlist`（0–255 每個位元組和獨立正規表示式比對）；兩者以突變測試確認可抓到「取最後讀到的檔」與「黑名單」錯誤實作；spec 補檔案層級規則仍適用；README 補截圖提醒；`TestMainBridgeAddress` 補結尾分隔符檢查；`formatSessionPeerName` 改名 `formatSessionSection` 並簡化；驗證 `go test -race ./...` 全綠、`gofmt -l .` 無輸出。
