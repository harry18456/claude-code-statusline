## 1. Model 層：解析 resets_at timestamp

- [x] 1.1 在 `internal/model/payload.go` 的 `rateLimitRaw` struct 新增 `ResetsAt int64 json:"resets_at"` 欄位（支援 parse resets_at timestamp）
- [x] 1.2 在 `model.RateLimit` struct 新增 `ResetsAt int64` 欄位
- [x] 1.3 更新 `ParsePayload` 中 `RateLimit` 的轉換邏輯，將 `raw.RateLimits.FiveHour.ResetsAt` / `SevenDay.ResetsAt` 複製至對應 struct

## 2. Model 測試

- [x] 2.1 在 `internal/model/payload_test.go` 新增測試：驗證 `resets_at` 存在時正確解析（resets_at present）
- [x] 2.2 新增測試：驗證 `resets_at` 缺席時 `ResetsAt` 為 0（resets_at absent）

## 3. Renderer 層：display rate limit countdown

- [x] 3.1 在 `internal/renderer/renderer.go` 新增 `formatCountdown(resetsAt int64) string` 函式，依照倒數格式規則實作：`ResetsAt is zero`、`Countdown expired`、`Countdown < 60 minutes`、`Countdown >= 60 minutes`（對應 ResetsAt 直接存入 RateLimit struct 設計決策）
- [x] 3.2 修改 `formatRate()` 函式以 Display rate limit countdown：當 `pct >= 80` 且 `rl.ResetsAt != 0` 時，在輸出後附加 `formatCountdown` 結果（對應 rate limit display 的新 scenario）
- [x] 3.3 確認 `pct < 80` 時不顯示倒數（rate limit below threshold）

## 4. Renderer 測試

- [x] 4.1 在 `internal/renderer/renderer_test.go` 新增單元測試：`formatCountdown` 的三種格式（>= 60m、< 60m、expired）
- [x] 4.2 新增整合測試：rate limit >= 80% 且 `resets_at` 存在時，輸出包含倒數字串
- [x] 4.3 新增測試：rate limit < 80% 時無倒數輸出（倒數計算在 renderer 內部執行 設計決策）

## 5. 常駐顯示：移除 80% 門檻

- [x] 5.1 修改 `formatRate()` 函式：移除 `pct >= 80` 的倒數顯示條件，改為只要 `rl.ResetsAt != 0` 即附加倒數（對應 Countdown always shown scenario）
- [x] 5.2 更新 renderer 測試：修改 `TestRenderRateLimitCountdownHiddenBelow80` 為驗證 < 80% 時倒數仍顯示；移除舊的「不顯示」斷言

## 6. 倒數格式：≥ 24h 支援

- [x] 6.1 修改 `formatCountdown()` 函式：在 `>= 60 min` 分支前新增 `>= 24h` 判斷，輸出 `(Xd Yh)` 格式（對應 Countdown >= 24 hours scenario）
- [x] 6.2 在 `internal/renderer/renderer_test.go` 新增單元測試：`formatCountdown` ≥ 24h 格式（e.g. 50 小時 → `(2d 2h)`）
