## Context

Claude Code 的跨 session 溝通由兩個工具組成：`ListAgents` 列出本機還在執行的 session，`SendMessage(to: "<name>")` 用名稱傳話。名稱就是地址；`[ref]` 只在同名時用來區分，而且必須剛從列表讀到才有效。

2026-11-20 在 CLI 2.1.329 實測得到以下事實：

- Claude Code 在 `<config>/sessions/<pid>.json` 為每個執行中的 session 寫一份 registry。欄位包含 `sessionId`、`name`、`nameSource`、`kind`、`status`、`updatedAt`、`jobId`、`formerNames`、`messagingSocketPath`。同一目錄還有 `<pid>.<hash>.key` 檔，屬於機密。
- `ListAgents` 顯示的名稱與 registry 的 `name` 完全一致（3 個 session 全部比對相符）。
- 命名規則：互動 session 是 `<cwd 資料夾名>-<2 位數>`（`nameSource: derived`），之後不會被 AI 標題取代。背景 job 先用 job id（session id 前 8 碼），之後自動改成 AI 標題（`nameSource: auto`）。`claude -n <name>` 可以自訂名稱。
- AI 標題不是地址。例如 `my-project-75` 的 AI 標題是「重構設定載入」。payload 的 `session_name` 在 2.1.187 與 2.1.286 的擷取樣本裡都是標題型內容。
- registry 檔可能是空白內容（實測 `39840.json` 全是空白字元），代表讀取時可能遇到半寫入狀態。

目前 renderer 是純函式：不做 IO，由 caller 傳入已解析的 `GitInfo`。payload 解析已有容錯框架（`tolerantInt64`、`OptionalBool` 等）。

## Goals / Non-Goals

**Goals:**

- 在 line 2 最後顯示本 session 的定址名稱，讓使用者能把它告訴別的 session 當作 `SendMessage` 地址。
- 查詢失敗時這一段靜默消失，其他段落與 exit code 不受影響（「statusline 永不消失」原則）。
- 不讀取 `.key` 機密檔。
- `--hide session` 可關閉此段，關閉時完全不碰檔案系統。

**Non-Goals:**

- 不顯示 `session_id` 本身，也不顯示 payload 的 `session_name`。
- 不做快取。每次渲染重新讀取，名稱改變時下一次渲染就會更新。
- 不檢查 registry 中的 pid 是否還活著。
- 不截斷名稱。地址必須完整，截斷會讓使用者拿到錯的地址。
- 不新增 Nerd Font 專屬 glyph。Nerd Font 模式沿用 `⇄`，避免重演 L-2 glyph 遺失問題。
- 不新增 `--hide` 以外的 flag 或環境變數。
- 不顯示 `[ref]`。ref 無法從 session id 算出（已測 md5、sha1、sha256 前綴皆不符），且必須來自最新列表。

## Decisions

### 從 session registry 解析定址名稱，不用 payload session_name

定址名稱只存在 registry，payload 沒有直接提供。payload 的 `session_name` 在樣本中是 AI 標題，對互動 session 而言和地址不同，顯示它會誤導使用者。

替代方案：直接顯示 `session_name`。實作較簡單，但無法保證是地址，已由使用者否決。

### 新增 internal/peername 套件封裝 registry 查詢

registry 讀取是檔案系統 IO，放在獨立套件 `internal/peername`，與 `internal/gitcache` 的定位相同。renderer 維持純函式。

替代方案：放進 renderer。會讓 renderer 開始做 IO，測試要建立暫存目錄，違反現有分層。

### Config 目錄順序：CLAUDE_CONFIG_DIR 再 ~/.claude

Claude Code 用 `CLAUDE_CONFIG_DIR` 覆寫設定目錄。statusline 子程序會繼承 Claude Code 的環境變數（`COLORTERM` 可讀即為證據）。先找 `CLAUDE_CONFIG_DIR`，找不到再找 `<home>/.claude`。因為比對的是完整 session id（UUID），多找一個目錄不會誤配。路徑經 `filepath.Clean` 後相同就只找一次。`CLAUDE_CONFIG_DIR` 只含空白時視為未設定。

替代方案：只找其中一個。`CLAUDE_CONFIG_DIR` 下是否一定有 `sessions/` 尚未實測，只找一個風險較高。

### 只讀 sessions 目錄下的 .json 一般檔案

只讀 `<config>/sessions/` 第一層、副檔名為 `.json`、型別為一般檔案（regular file）的項目。`.key` 等其他副檔名、子目錄、symlink 一律跳過。超過 65536 bytes 的檔案跳過，避免異常大檔拖慢渲染。大小限制用 `io.LimitReader` 實作：最多讀 65537 bytes，讀到超過 65536 就丟棄。這樣不需要另外查 metadata，檔案在讀取途中變大也一樣受限。

### 開檔後再確認是一般檔案

目錄掃描與開檔之間，`.json` 項目可能被換成 named pipe 或裝置。讀 pipe 會一直等到對方寫入，statusline 就會卡住。因此開檔後先對 handle 呼叫 `Stat`，不是一般檔案就直接跳過、不讀內容。Windows 測試用 `CreateNamedPipeW` 建一個不寫資料的 pipe，確認 `readEntry` 立即回傳。

殘留風險：Unix 上若換成 FIFO，`open()` 本身就會等待寫入端，發生在 `Stat` 之前。這需要攻擊者能寫入使用者自己的 config 目錄；有這種權限的人可以直接改 `settings.json` 的 hooks，所以不再加 `O_NONBLOCK`。

### Windows 以 FILE_SHARE_DELETE 開啟 registry 檔

Go 1.24 在 Windows 的 `os.Open` 只給 `FILE_SHARE_READ|FILE_SHARE_WRITE`。statusline 讀檔的瞬間，Claude Code 若要刪除自己的 registry 檔會失敗。本機實測（Windows 11、Go 1.24.1）：

| 讀取端開檔方式 | 覆蓋改名（`os.Rename`） | 刪除 | 原地改寫 |
| -------------- | ----------------------- | ---- | -------- |
| `os.Open` | 拒絕 | 拒絕 | 可以 |
| `CreateFile` + `FILE_SHARE_DELETE` | 拒絕 | 可以 | 可以 |

因此 Windows 改用 `syscall.CreateFile` 並加上 `FILE_SHARE_DELETE`（`openShared`），其他平台沿用 `os.Open`。覆蓋改名在兩種方式下都會被拒絕，這是 Windows 舊式 rename 語意的限制；讀取時間只有數微秒，殘留風險可接受。

`CreateFile` 失敗時改用 `os.Open` 重試。直接呼叫 `CreateFile` 會少掉 `os.Open` 內建的長路徑處理（加上 `\\?\` 前綴）。系統未開啟長路徑支援、且路徑超過 260 字元時，沒有重試的話每個檔案都會開檔失敗，名稱會靜默消失。

### 多筆相符時取 updatedAt 最大者

同一 session id 可能出現在多個檔案（例如程序當掉留下舊檔，之後被 resume）。取 `updatedAt` 數值最大者，它代表最近活動的那個程序。`updatedAt` 缺少或相同時保留檔名順序中最早的一筆（`os.ReadDir` 依檔名排序），結果可重現。

### updatedAt 寬鬆解析

`updatedAt` 只用來在多筆相符時決勝負，不該決定一筆紀錄能不能用。registry 是未公開格式，Claude Code 哪天把它改成 ISO 時間字串也不奇怪。如果用嚴格的 `float64` 解碼，型別一變，整個 JSON 解碼就失敗，所有紀錄都會被丟掉，這一段永遠消失。因此 `updatedAt` 改用私有型別 `lenientFloat`：數字照常解析，字串、null、物件都當成 0，不回傳錯誤。`sessionId` 與 `name` 仍維持嚴格字串，型別不符就不算相符。

### 清除名稱中的控制字元

名稱會直接印到終端機。換行會破壞兩行輸出協定，ESC 會注入終端控制序列。bidi 覆寫字元（例如 U+202E）會讓畫面上的地址看起來和實際不同，零寬字元則讓人無法照著畫面打出地址。因此移除 Unicode 類別 Cc、Cf、Zl、Zp 的所有字元，再去除前後空白；順序是先移除再 trim，ESC 加空白開頭的名稱也能正確去掉空白。清除後為空字串視為不相符。

代價：Cf 包含 ZWJ（U+200D），emoji 組合序列會被拆成多個獨立 emoji。名稱是給人看、給人打字用的，接受這個代價。

### Render 簽章新增 SessionInfo 參數

新增 `renderer.SessionInfo{PeerName string}`，`Render` 簽章改為 `Render(p, git, session, opts)`。與 `GitInfo` 相同，由 caller 傳入 payload 以外解析到的資料。測試透過既有的 `renderWith` helper 呼叫，只需改 helper 一處。renderer 會原樣輸出 `PeerName`，所以 caller 必須傳入已清除過的名稱；`SessionInfo` 的註解寫明這個前提，`peername.Lookup` 回傳的名稱已清除。

替代方案：把名稱塞進 `model.Payload`。payload 代表 Claude Code 傳來的 JSON，混入衍生資料會模糊型別意義。

### 隱藏 session 時跳過 registry 讀取

`cmd/statusline` 在解析 payload 後，只有在 `session` 未被隱藏且 `session_id` 非空時才呼叫 `peername.Lookup`。需要 `renderer.Options` 提供可從套件外查詢隱藏狀態的方法。

### 段落格式：預設 ⇄ <name> 灰色、ASCII peer:<name>

- 預設與 Nerd Font 模式：`\033[90m⇄ <name>\033[0m`。使用者已核准此草圖。
- ASCII 模式：`peer:<name>`，不帶 ANSI 色碼，與 `cache:`、`effort:` 這類 ASCII 標籤段落一致。`peer` 對應 `ListAgents` 的「Peer sessions」用詞。
- 名稱本身照原樣輸出（只做控制字元清除），與 branch、dir 等使用者資料的處理一致。

### session_id 容錯解析

新增匯出型別 `model.TolerantString`（底層型別為 string）：JSON 字串照常解析，`null`、數字、物件等其他型別一律視為空字串且不回傳錯誤。避免 `session_id` 型別異常讓整份 payload 解析失敗、statusline 退回 parse error 行。沿用既有 `OptionalBool` 這類「欄位用自訂型別承載容錯」的慣例，不為 `Payload` 加自訂 `UnmarshalJSON`。

## Implementation Contract

**Behavior**

- 在 Claude Code 中，line 2 最後出現 `⇄ <name>`，`<name>` 與 `ListAgents` 對本 session 顯示的名稱相同。
- 背景 job 被自動改名後，下一次渲染顯示新名稱。
- 找不到名稱時，line 2 與加入此功能前完全相同。

**Interface / data shape**

- `model.Payload` 新增 `SessionID model.TolerantString`，對應 JSON 頂層 `session_id`，呼叫端以 `string(p.SessionID)` 取值；解析規則見「session_id 容錯解析」。
- `peername.ConfigDirs(getenv func(string) string, home string) []string`：回傳候選 config 目錄（順序與去重規則見上）。`home` 為空字串時不加入 `<home>/.claude`。
- `peername.Lookup(sessionID string, configDirs []string) string`：回傳清除後的定址名稱，失敗回傳空字串，永不回傳 error、永不 panic。
- 讀取的 registry 欄位：`sessionId`（字串）、`name`（字串）、`updatedAt`（數字）。其他欄位一律忽略。`sessionId` 或 `name` 型別不符時該檔視為不相符；`updatedAt` 不是數字時當成 0，仍可相符。
- `renderer.SessionInfo{PeerName string}`；`renderer.Render(p *model.Payload, git GitInfo, session SessionInfo, opts Options) (line1, line2 string)`。
- `renderer.SectionSession SectionKey = "session"`，`--hide session` 為合法 key。
- `renderer.Options` 新增匯出方法 `Hidden(key SectionKey) bool`。

**Failure modes**

- registry 目錄不存在、讀檔失敗、開檔後不是一般檔案、空白或半寫入檔案、JSON 錯誤、`sessionId` 或 `name` 型別錯誤、名稱為空：一律靜默，結果為「沒有名稱」。
- 查詢不寫 stderr、不改 exit code、不影響其他段落。

**Acceptance criteria**

- `go build ./...`、`go vet ./...`、`go test -race ./...` 通過，`gofmt -l .` 無輸出；`staticcheck ./...` 若本機可用也要通過。
- 單元測試覆蓋 spec 的每個 scenario 與 example 表格列。
- 真 binary 實跑：
  1. 用暫存 `CLAUDE_CONFIG_DIR` 放構造的 registry，`cat payload | statusline.exe`，line 2 結尾為 `⇄ <name>`。
  2. 加 `--ascii`，line 2 結尾為 `peer:<name>`。
  3. 加 `--hide session`，名稱消失且無多餘分隔符。
  4. registry 只有壞檔、空白檔、`.key` 檔時，名稱不出現，stderr 為空，exit code 為 0。
  5. 用本機真實 registry 與本 session 的 session id 實跑，顯示的名稱與 `ListAgents` 對本 session 的名稱一致。

**Scope boundaries**

- In scope：payload `session_id` 解析、`internal/peername` 查詢、line 2 session 段落、`--hide session`、README（英文與繁中）與 CLAUDE.md 文件更新、對應測試。
- Out of scope：快取、pid 存活檢查、顯示 session id 或 session_name、名稱截斷、Nerd Font 專屬 glyph、其他新 flag 或環境變數、修改 `debug-tee`。

## Risks / Trade-offs

- [registry 是 Claude Code 未公開格式，改版可能變動] → 任何解析失敗都讓本段靜默消失，其他輸出不受影響；README 註明此段依賴內部格式。
- [`CLAUDE_CONFIG_DIR` 下是否有 `sessions/` 尚未實測] → 找不到時退回 `<home>/.claude`；以完整 UUID 比對，不會誤配。
- [程序當掉留下同 session id 的舊檔] → 取 `updatedAt` 最大者。
- [每次渲染都讀檔] → 實測只有約 6 個小於 1 KB 的檔案；有 65536 bytes 大小上限；隱藏時不讀。
- [讀檔瞬間擋住 Claude Code 改寫 registry] → Windows 以 `FILE_SHARE_DELETE` 開檔，刪除與原地改寫不受影響；覆蓋改名仍會在數微秒的讀取窗口內被拒絕。
- [`CLAUDE_CONFIG_DIR` 為相對路徑或網路路徑（UNC）] → 接受。相對路徑依 statusline 的工作目錄解析；UNC 路徑代表每次渲染都會存取網路磁碟。兩者都是 Claude Code 本身也在使用的設定，不另外限制。
- [sessions 目錄沒有檔案數量上限] → 接受。目錄由 Claude Code 維護，實測只有個位數檔案；設上限反而可能漏掉真正的那一筆。
- [專案沒有 CI test job；symlink 測試在 Windows 預設權限下會 SKIP] → 本 change 不擴大範圍去加 CI。symlink 規則仍由 Unix 與開啟開發者模式的 Windows 驗證；pipe 與裝置由「開檔後再確認是一般檔案」另外把關。
- [誤讀 `.key` 機密檔] → 只接受 `.json` 副檔名，並有測試確認 `.key` 內容即使相符也不會被採用。
- [互動 session 的名稱 `<資料夾名>-NN` 與 dir 段落視覺重複，line 2 變長] → 可用 `--hide dir` 或 `--hide session` 調整；地址必須完整，不做縮寫。
- [AI 標題可能含 CJK 等寬字元，line 2 變長] → 同上，屬於可接受的代價。
