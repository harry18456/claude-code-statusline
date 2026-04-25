## 1. RED — 新增/更新測試覆蓋零容忍行為

- [x] 1.1 [P] 在 `internal/renderer/renderer_test.go` 新增 3 個測試覆蓋新行為（落實 `Seven-day usage pace indicator` 的 `Seven-day magnitude floor at 1` scenario 與「永遠有方向」語意）：
  - `TestComputePaceArrow_WithinOldToleranceNowOverPace`：elapsed_days=2、expected≈28.57、used=30 → dev≈+1.43 → 期望 `▲1%`（v1.5.0 在此區間顯示 `≈`）
  - `TestComputePaceArrow_WithinOldToleranceNowUnderPace`：elapsed_days=2、expected≈28.57、used=27 → dev≈-1.57 → 期望 `▼2%` 灰色
  - `TestComputePaceArrow_MagnitudeFloorAtOne`：elapsed_days=2、expected≈28.5714、used=29 → dev≈+0.43（`|dev| < 0.5`）→ 期望 `▲1%`，明確斷言不可含 `▲0%`
- [x] 1.2 改寫既有受影響測試以對應新門檻：
  - `TestComputePaceArrow_WithinTolerance`：原 used=30/dev≈+1.43 在新門檻下變 over-pace。重新命名為 `TestComputePaceArrow_ExactMatch`，改用「elapsed_days=0+used=0 → dev=0」場景，期望 `≈`（覆蓋 `Seven-day exact match` scenario）
  - `TestComputePaceArrow_WithinToleranceBoundary`：原迴圈 `dev ∈ {-5, 0, +5}` 全 `≈`。改寫為三個獨立斷言：dev=-5 → `▼5%`、dev=0 → `≈`、dev=+5 → `▲5%`
  - `TestComputePaceArrow_NearResetWithinTolerance`：原 used=98/elapsed_days=7/expected=100/dev=-2 → 期望從 `≈` 改為 `▼2%`
  - `TestComputePaceArrow_AtWindowStart`：原 elapsed=0/used=0/dev=0 → `≈` 仍正確，保留
  - `TestComputePaceArrow_ElapsedDaysCappedAtSeven`：原 used=100 子斷言 dev=0 → `≈` 保留；used=0 子斷言 dev=-100 → `▼100%` 保留
- [x] 1.3 [P] 更新 ASCII 容忍測試：`TestComputePaceArrow_ASCIIWithinTolerance` 原 used=30 在新門檻下變 `^1%`，改名為 `TestComputePaceArrow_ASCIIWithinOldToleranceNowOverPace` 並更新斷言為 `^1%`；新增 `TestComputePaceArrow_ASCIIExactMatch`（dev=0 場景）驗證 `~`
- [x] 1.4 改寫整合測試 `TestRenderSevenDayWithinToleranceApprox`：原 used=43/elapsed_days=3/expected≈42.86/dev≈+0.14 在新門檻+magnitude floor 下變 `▲1%`。重新命名為 `TestRenderSevenDayMagnitudeFloorRendersArrow` 並更新斷言為含 `7d:43% ▲1%`，斷言不可含 `≈`
- [x] 1.5 執行 `go test ./internal/renderer/... -run "TestComputePaceArrow_|TestRenderSevenDay"` 確認 1.x 測試在 v1.5.0 實作下 FAIL（7 個測試在舊實作下 FAIL，覆蓋零容忍 + magnitude floor 全部新行為）

## 2. GREEN — 修改實作（含 inline audit）

- [x] 2.1 修改 `internal/renderer/renderer.go` 的 `computePaceArrow`：
  - 分支條件 `case deviation > 5` → `case deviation > 0`、`case deviation < -5` → `case deviation < 0`
  - 在現有 `magnitude := int(math.Round(math.Abs(deviation)))` 之後加入 `if magnitude == 0 && deviation != 0 { magnitude = 1 }`，落實 `Seven-day magnitude floor at 1` scenario
  - default case（落到 `≈`）保持不變，自然涵蓋 `deviation == 0` 邊界
  - **Inline audit**：檢查 `deviation != 0` 浮點比較的可靠性——`expected_pct = elapsed_days × 100/7`，當 elapsed_days ∈ [1,6] 時 expected 為非整數浮點數，與整數 `used_percentage` 相減幾乎不會剛好為 0；只有 `elapsed_days ∈ {0, 7}` 配上 `used ∈ {0, 100}` 才會精確 dev=0；不需特殊浮點容忍
- [x] 2.2 執行 `go test ./internal/renderer/...` 確認 1.x 測試由 RED 全數轉 GREEN（含補充更新 TestComputePaceArrow_DayBoundaryStepUp 適配 magnitude floor）
- [x] 2.3 執行 `go test ./...` 全部通過（regression check）

## 3. 手動驗證 + 文件 + Commit

- [x] 3.1 `./dev.sh build` 重新安裝；構造 4 個 payload 視覺驗證：
  - `used=10` + `resets_at = now + 6d 16h`（使用者原始 case）→ `7d:10% ▼4%`（v1.5.0 顯示 `≈`）
  - `used=43` + `resets_at = now + 4d 12h`（mid day-3，dev≈+0.14）→ `7d:43% ▲1%`（落實 magnitude floor）
  - `used=80` + `resets_at = now + 3d 12h`（mid day-4，dev≈+22.86）→ `7d:80% ▲23%` 紅色（regression：大偏差行為不變）
  - `five_hour.used=99` + `seven_day` 缺省 → 5h 無 pace indicator（regression）
- [x] 3.2 更新 `README.md` 與 `README.zh-TW.md` 描述 `Seven-day usage pace indicator` 段落：移除「±5% tolerance」/「偏離 ≤ 5% 顯示 ≈」字句，改為「任何非零偏差顯示 `▲<N>%` 或 `▼<N>%`；`≈` 僅在偏差為零時出現（極罕見）」
- [ ] 3.3 撰寫 en-us commit message：`refactor(renderer): drop pace tolerance, always show directional indicator`，body 說明 v1.5.0 → v1.5.1 微調動機（使用者反饋 `≈` 缺乏行動性）與 magnitude floor=1 設計
