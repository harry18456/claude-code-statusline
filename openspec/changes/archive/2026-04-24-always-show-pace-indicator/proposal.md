## Summary

移除 `seven_day` pace indicator 在視窗末段（剩餘時間 < 10%）的抑制行為，改為永遠顯示 pace 箭頭。

## Motivation

目前 `computePaceArrow`（`internal/renderer/renderer.go:334`）在剩餘時間 < 60480 秒（604800 秒的 10%，約 16h 48m）時會直接回傳空字串，pace 箭頭會從 statusline 上整個消失。此行為造成兩個問題：

1. **視覺不連續**：使用者盯著 `7d ▼77%` 看了整個視窗，最後幾小時卻突然消失，容易被誤以為是顯示壞掉。
2. **資訊反而更有價值時被隱藏**：視窗末端的大 deviation 其實是很誠實的訊號 — `▼87%` 搭配倒數 `(1h Xm)` 代表「這週額度用不到一成就要 reset 了，下週可以放開用」或「這週是否 under-utilize」，語意清晰且實用。

抑制邏輯的原始理由（`renderer.go:315-318` 註解）是 `expected_pct` 在視窗末尾逼近 100，讓 deviation 變得「飽和失真」。但實際上 deviation 變大不是雜訊，是真實狀態 — 倒數計時本身已經提供足夠的上下文，使用者能自己判讀這個數字。

## Proposed Solution

- 從 `computePaceArrow` 移除 `renderer.go:333-336` 的 10% 抑制分支。
- 從 `rate-limit-countdown` spec 的「Seven-day usage pace indicator」requirement 中：
  - 刪除「Seven-day near-reset suppression」scenario
  - 從「Seven-day over-pace」、「Seven-day under-pace」、「Seven-day within tolerance」三個 scenarios 的 WHEN 條件移除 `AND remaining window time (resets_at - now) is at least 10% of 604800 seconds`
- 移除 `renderer.go:315-318` 已失效的視窗假設註解（或改寫為說明「pace 永遠顯示」的新註解）。
- 更新 `renderer_test.go` 中驗證「near-reset 抑制」的測試案例，改為驗證該情境下仍會顯示箭頭。

移除後，pace indicator 在 `ResetsAt != 0` 且 `remaining > 0` 的任何時間都會依 deviation 顯示 `▲`/`▼`/`≈`，僅保留兩個真正該抑制的情境：`resets_at` 缺失與已過期。

## Non-Goals

- 不改變 deviation 的數學公式（`expected_pct`、`deviation`、`magnitude` 計算保持原樣）。
- 不改變 5% 容忍門檻（`|deviation| > 5` 才顯示箭頭，否則顯示 `≈`）。
- 不改變 `five_hour` 的顯示（5h 仍不顯示 pace，只有 7d 有）。
- 不調整顏色、符號或 ASCII fallback 對應。
- 不處理「7d 是否為 rolling window」的潛在問題 — 那屬於另一個 change 的範疇，本 change 僅移除抑制邏輯。

## Alternatives Considered

1. **改用淡色 placeholder（`·` 或 `—`）代替箭頭**：視覺連續性保住，但語意變成「pace 已失效」，反而模糊了原本可用的 deviation 資訊。使用者還是看不到真實狀態。
2. **降低閾值（如 5% 或 1h）**：只是把突然消失的時間點往後推，問題本質沒解決。
3. **改成與 24h 前 `used_pct` 比較的短期趨勢**：根本解，但需要額外的 state 持久化（目前 binary 是 stateless），scope 大太多，不適合在這個 change 處理。

## Impact

- Affected specs: `rate-limit-countdown`（modified — 移除一個 scenario、修改三個 scenarios 的 WHEN 條件）
- Affected code:
  - `internal/renderer/renderer.go`（移除 `computePaceArrow` 內第 333-336 行的抑制分支，調整第 315-318 行註解）
  - `internal/renderer/renderer_test.go`（更新 near-reset 相關測試）
