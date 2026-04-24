## 1. 測試先行（TDD Red）

- [x] 1.1 在 `internal/renderer/renderer_test.go` 新增測試案例：驗證 Seven-day usage pace indicator 在 `remaining < 60480` 秒時仍會依 deviation 顯示 `▲<N>%`、`▼<N>%` 或 `≈`（涵蓋 near-reset over-pace、near-reset under-pace、near-reset within-tolerance 三種子情境）。此時執行 `go test ./internal/renderer/...` 應看到新增測試 FAIL。
- [x] 1.2 在 `internal/renderer/renderer_test.go` 定位既有驗證「near-reset 抑制」的測試（預期 arrow 為空字串的 case），改為驗證在同樣時間條件下會輸出正確的 pace indicator，或以 1.1 新增的案例取代並刪除舊案例。

## 2. 實作（TDD Green）

- [x] 2.1 在 `internal/renderer/renderer.go` 的 `computePaceArrow` 函式移除第 333-336 行 `if remaining*10 < sevenDayWindowSeconds { return "" }` 抑制分支，使 Seven-day usage pace indicator 在剩餘時間小於視窗 10% 時仍會計算並輸出。
- [x] 2.2 更新 `internal/renderer/renderer.go` 第 315-318 行的註解：刪除「fixed-bucket window 假設」與「pace arrows become stale signals」等已失效敘述，改為說明 `sevenDayWindowSeconds` 僅用於計算 `expected_pct` 與 deviation，pace indicator 會在整個視窗期間持續顯示。
- [x] 2.3 執行 `go test ./internal/renderer/...` 確認 1.1、1.2 的測試全部通過。

## 3. 整體驗證

- [x] 3.1 [P] 執行 `go test ./...` 確認全專案測試通過（Seven-day usage pace indicator 以及其他 renderer、model、gitcache 測試均綠）。
- [x] 3.2 [P] 執行 `./dev.sh build` 確認專案可成功編譯並安裝到 `~/.claude/statusline.exe`。
- [x] 3.3 使用 `debug.json` 的既有 payload 執行 `cat debug.json | ~/.claude/statusline.exe` 手動驗證輸出：在 `seven_day.resets_at` 接近 now 的情境下仍可看到 `▲<N>%` / `▼<N>%` / `≈` 其中之一，不會再整段消失。
