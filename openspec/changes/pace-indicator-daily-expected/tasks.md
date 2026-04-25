## 1. RED — 新增失敗測試覆蓋日級行為

實作前，所有新增測試應在現行（秒級）實作下 FAIL。

- [x] 1.1 [P] 在 `internal/renderer/renderer_test.go` 新增以下 5 個測試函式覆蓋 `Seven-day usage pace indicator` 新行為（落實 design `Use ceil for elapsed_days`、`Cap elapsed_days at 7`）：
  - `TestComputePaceArrow_Day1At1Second`：elapsed=1 秒、used=0 → elapsed_days=1、expected≈14.29、dev≈-14.29 → 含 `▼14%`
  - `TestComputePaceArrow_Day1Under6Percent`：使用者原始 case，elapsed=25200 秒（7 小時）、used=6 → 含 `▼8%` 且 ANSI gray（對應 `Seven-day day-1 under-pace example` scenario）
  - `TestComputePaceArrow_DayBoundaryStepUp`：elapsed=86400 vs elapsed=86401 兩次呼叫結果差異反映 expected_pct 跳升 ≈14.29
  - `TestComputePaceArrow_AtWindowStart`：elapsed=0、used=0 → expected_pct=0、deviation=0 → 含 `≈`
  - `TestComputePaceArrow_ElapsedDaysCappedAtSeven`：模擬 elapsed=604801 → elapsed_days clamp 至 7、expected_pct=100
- [x] 1.2 執行 `go test ./internal/renderer/... -run TestComputePaceArrow_` 確認 1.1 新增 5 測試在舊實作下全部 FAIL（實際：3/5 FAIL — Day1At1Second / Day1Under6Percent / DayBoundaryStepUp；AtWindowStart 與 ElapsedDaysCappedAtSeven 在新舊公式邊界等價，作為實作後 regression guard）

## 2. RED — 更新既有測試以對齊新公式

- [x] 2.1 [P] 整理 `TestComputePaceArrow_OverPace` / `UnderPace` / `WithinTolerance` / `WithinToleranceBoundary`：因 elapsed=172800 秒恰為 2 天整，`ceil(172800/86400)=2`、`expected = 2 × (100/7) ≈ 28.5714`，與舊公式恰好相同。期望值不變，但更新註解為「elapsed_days=2 → expected≈28.5714」並把 `WithinToleranceBoundary` 的常數 `100.0 * 172800.0 / 604800.0` 改寫為 `2 * (100.0 / 7.0)`，落實 design `Tolerance threshold remains ±5%` 容忍門檻不變
- [x] 2.2 [P] 改寫 near-reset 三測試以對應新公式（`Cap elapsed_days at 7` + `No magnitude formatting changes`）：
  - `TestComputePaceArrow_NearResetOverPace`：原 used=99/elapsed=544800 在新公式下不再 over-pace。改成 `resetsAt = now + 86400`（remaining=1d）→ elapsed=518400、ceil=6、expected≈85.71；used=95 → dev≈+9.29 → 期望 `▲9%` 紅
  - `TestComputePaceArrow_NearResetUnderPace`：保留 used=12/resetsAt=now+4500，但期望從 `▼87%` 改為 `▼88%`（新公式 expected=100、dev=-88）
  - `TestComputePaceArrow_NearResetWithinTolerance`：原 used=88/elapsed=544800 已不在容忍區。改成 used=98/resetsAt=now+60000 → elapsed_days=7、expected=100、dev=-2 → 仍 `≈`
- [x] 2.3 改寫整合測試 `TestRenderSevenDayOverPaceArrow` 與 `TestRenderSevenDayWithinToleranceApprox`：兩者用 `time.Now().Add(5 * 24 * time.Hour)`，新公式下 elapsed 落在 day boundary（86400 整數倍）邊緣，sub-second 時鐘 drift 會讓 ceil 跨日造成 flaky。改成 `time.Now().Add(4*24*time.Hour + 12*time.Hour)`（落在 day 3 中段）並依新 expected_pct 重算斷言字串

## 3. GREEN — 修改實作

- [x] 3.1 修改 `internal/renderer/renderer.go` 的 `computePaceArrow`，將 `expectedPct := float64(elapsed) / float64(sevenDayWindowSeconds) * 100` 替換為：
  ```go
  elapsedDays := int64(math.Ceil(float64(elapsed) / 86400))
  if elapsedDays > 7 {
      elapsedDays = 7
  }
  expectedPct := float64(elapsedDays) * (100.0 / 7.0)
  ```
  此處同時落實 design `Use ceil for elapsed_days` 與 `Cap elapsed_days at 7` 兩個決策；保留既有 `magnitude := int(math.Round(math.Abs(deviation)))` 與其後分支不動，落實 `No magnitude formatting changes`
- [x] 3.2 執行 `go test ./internal/renderer/...`，1.x 與 2.x 測試應由 RED 全數轉 GREEN
- [x] 3.3 執行 `go test ./...` 全部通過（regression：model / 其他 renderer 測試不受影響）

## 4. 手動驗證

對應 `Seven-day usage pace indicator` requirement 在實際環境的觀感。

- [x] 4.1 `./dev.sh build` 重新安裝後，依 `MEMORY.md` 的「抓取最新 Claude Code statusLine payload」流程取得真實 payload 作為基底
- [x] 4.2 改造 payload `seven_day.used_percentage=6`、`resets_at = <now + 6d 17h>`，`cat debug.json | ~/.claude/statusline.exe` 確認顯示 `7d:6%` + `▼8%`（不再是 `≈`），對應使用者原始 case
- [x] 4.3 改造 payload 進入第 4 天中段 + used=60（dev≈+2.86）確認 `≈`、改 used=80（dev≈+22.86）確認 `▲23%` 紅色，落實 design `Tolerance threshold remains ±5%` 容忍門檻在新公式下仍正確分流
- [x] 4.4 `five_hour.used_percentage=99` + 任意 `resets_at` 視覺確認 5h **不**附加 pace indicator（regression check）

## 5. Audit 與文件

- [x] 5.1 對 `internal/renderer/renderer.go` 的 `computePaceArrow` 變更區塊執行 `/spectra:audit`，重點檢查 `math.Ceil` 浮點 → int64 轉型、elapsed_days clamp 邊界、`100.0/7.0` 浮點精度是否引入 silent failure（inline 三視角審查通過：Scoundrel/Lazy/Confused 無新增 sharp edges；遇異常未來 resets_at 的負 elapsed 處理為既有行為，不在本 change scope）
- [x] 5.2 檢視 `README.md` 與 `README.zh-TW.md` 描述 `Seven-day usage pace indicator` 的段落，把「linear expected usage」改為「daily linear expected usage（每日線性預期）」並補充「跳階對齊 reset 鐘點，非日曆午夜」一句
- [ ] 5.3 撰寫 en-us commit message 並標明 BREAKING：`feat(renderer): switch 7d pace expected_pct to daily granularity`，body 說明使用者體驗變更與 design 連結
