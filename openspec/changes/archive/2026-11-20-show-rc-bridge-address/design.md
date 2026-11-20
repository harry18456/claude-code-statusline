## Context

`show-session-peer-name`（v1.6.0）讓 line 2 顯示本機定址名稱 `⇄ <name>`，但這個名稱只在同一台機器有效。2026-11-20 的跨機台討論與 RC 開關實驗確認了以下事實：

- 別台機器只能用 `bridge:<bridgeSessionId>` 送達；本機名稱（例如 `my-project-5c`）在別台回報「No agent named 'my-project-5c' is reachable」。雙方的 `bridgeSessionId` 與收到的 `from` 屬性逐字一致。
- `bridgeSessionId` 存在本機 registry `<config>/sessions/<pid>.json`。RC 連線時是字串；關掉 RC 約 2 秒內變成 `null`（欄位保留）；重新連上會拿到新值。
- statusline payload 完全沒有 RC 或 bridge 欄位：同一 session 開、關、再開 RC，payload 欄位集合不變。
- 沒開 RC 的本機 session 可以用 bridge 地址傳給本機開著 RC 的 session（單向，對方無法回覆）。
- 實測的值形如 `session_` 加上 24 個英數字。

## Goals / Non-Goals

**Goals:**

- RC 連線時，session 段落同時顯示本機名稱與 bridge 地址；RC 關閉時只顯示名稱。
- 使用者可以用 `--hide bridge` 只關掉較長的 bridge 部分。
- bridge 值格式異常時靜默略過，不影響名稱，也不影響 statusline 其他段落。

**Non-Goals:**

- 不顯示 RC 對話標題（會變，也和 payload `session_name` 不同）。
- 不偵測其他機器上的 session，也不呼叫任何網路服務。
- 不改變名稱的比對規則與勝出規則。
- 不新增 `--hide bridge` 以外的 flag。

## Decisions

### bridge 地址取自同一筆 registry 紀錄

bridge 值與名稱來自同一筆勝出的紀錄（`sessionId` 相符、`updatedAt` 最大）。不從其他相符紀錄補 bridge 值，避免把已關閉 RC 的舊 bridge 地址和新名稱湊在一起。`--hide session` 時完全不讀 registry 的既有行為不變。

替代方案：另外掃描所有相符紀錄找「任何一個有 bridge 的」。可能拿到程序當掉後殘留的舊值，否決。

### bridgeSessionId 白名單驗證

`bridgeSessionId` 會原樣印在終端機上，也會被使用者複製當作 `SendMessage` 的地址。只接受 1 到 128 個 ASCII 英數字、`_`、`-` 組成的字串；其他一律視為沒有 bridge 地址。白名單比照名稱的控制字元清除更嚴格，因為這是地址格式，不需要容納任意文字。解碼使用寬鬆字串型別：`null`、數字、物件不會讓整筆 JSON 解碼失敗，所以名稱照常比對。

### 顯示格式：同一段內以「 · 」接在名稱後

預設與 Nerd Font 模式為 `⇄ <name> · bridge:<id>`，整段同一個灰色區間；ASCII 模式為 `peer:<name> bridge:<id>`（U+00B7 不是 ASCII，所以改用空白）。bridge 是同一個 session 的另一個地址，放在同一段比拆成獨立段落清楚，也不會多一個分隔符。`bridge:` 前綴照原樣顯示，因為這就是 `SendMessage` 要填的完整地址。

### 新增 --hide bridge，只藏 bridge 部分

`bridge` 是新的合法 hide key，只移除 ` · bridge:<id>` 這段；`session` 仍然移除整段並跳過 registry 讀取。名稱沒有時，即使有 bridge 也不顯示，維持「段落以名稱為主」的規則。

### peername.Lookup 改回傳 Info 結構

`Lookup` 從回傳 `string` 改成回傳 `peername.Info{Name, BridgeSessionID string}`，一次讀檔取得兩個值。`renderer.SessionInfo` 新增 `BridgeSessionID`，由 caller 傳入已驗證的值；renderer 仍然不做 IO、也不做驗證。

## Implementation Contract

**Behavior**

- 本機 session 連著 RC：line 2 結尾為 `⇄ <name> · bridge:<id>`（ASCII `peer:<name> bridge:<id>`）。
- RC 關閉或值不合格：line 2 結尾為 `⇄ <name>`，和 v1.6.0 完全相同。
- `--hide bridge`：只有名稱。`--hide session`：整段消失且不讀 registry。

**Interface / data shape**

- `peername.Info{Name string; BridgeSessionID string}`；`peername.Lookup(sessionID string, configDirs []string) Info`。名稱為空時 `BridgeSessionID` 也必為空。
- registry 讀取欄位新增 `bridgeSessionId`（寬鬆字串解碼，再做白名單驗證）。
- `renderer.SessionInfo{PeerName string; BridgeSessionID string}`；`renderer.SectionBridge SectionKey = "bridge"`。
- `cmd/statusline` 的 `resolveSession` 的 lookup 參數型別改為 `func(sessionID string, configDirs []string) peername.Info`。

**Failure modes**

- `bridgeSessionId` 缺少、`null`、非字串、空字串、超過 128 字元、含白名單外字元：視為沒有 bridge 地址，名稱照常，stderr 不輸出，exit code 不變。

**Acceptance criteria**

- `go build ./...`、`go vet ./...`（windows、linux、darwin）、`staticcheck ./...`、`go test -race ./...` 通過，`gofmt -l .` 無輸出。
- 單元測試覆蓋 spec 的每個 scenario 與 example 表格列。
- 真 binary：構造 registry 的字串、`null`、不合格值三種情況；`--ascii`；`--hide bridge`；`--hide session`；以及本機真實 registry 中開著 RC 的 session，顯示的 bridge 地址與 registry 一致。

**Scope boundaries**

- In scope：`internal/peername`、`internal/renderer`、`cmd/statusline` 的 bridge 支援、`--hide bridge`、README（中英）與 CLAUDE.md、對應測試。
- Out of scope：RC 標題顯示、網路查詢、其他機器的 registry、名稱比對規則調整、CI 與 release 流程。

## Risks / Trade-offs

- [RC 連線時 line 2 變長約 40 字元] → 提供 `--hide bridge`；README 說明。
- [每次重新連上 RC，bridge 地址都會換] → statusline 每次刷新重讀 registry，畫面永遠是當下的值；README 提醒舊地址會失效。
- [registry 是未公開格式，`bridgeSessionId` 可能改名或改格式] → 白名單驗證失敗時只會少掉 bridge 部分，名稱與其他段落不受影響。
- [bridge 地址是你帳號內 session 的識別碼，會顯示在畫面上] → 只有同帳號、經過 Claude Code 驗證的 session 能用它送訊息；README 提醒截圖分享時注意。
