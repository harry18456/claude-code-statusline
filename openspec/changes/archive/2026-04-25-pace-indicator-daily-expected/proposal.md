## Why

目前 `seven_day` pace indicator 用秒級線性計算 expected_pct（`elapsed / 604800 × 100`）。這導致週期才剛開始幾小時時 expected 很低（例：過了 7 小時 → expected ≈ 4.17%），即使實際 used 為 6%（明顯不到「第 1 天該用 14.3%」的每日配額），仍因 deviation ≈ +1.83% 落在 ±5% 容忍區間而顯示中性 `≈`。這違反使用者「第 N 天就該到 N×14.3%」的直觀心智模型——使用者會覺得「明明落後每日配額卻顯示節奏正常」。

## What Changes

- **BREAKING**：修改 `Seven-day usage pace indicator` requirement 中 `expected_pct` 的計算公式，從秒級線性改成日級線性：
  - 舊：`expected_pct = elapsed / 604800 × 100`
  - 新：`expected_pct = ceil(elapsed / 86400) × (100 / 7)`
- 跳階點對齊 `resets_at` 鐘點（不是日曆午夜）：視窗起點 `= resets_at − 604800`，跳階發生在視窗起點 + N×86400 秒。
- 第 1 天（任意時刻）expected = 14.2857%；第 2 天 = 28.5714%；……；第 7 天 = 100%。
- 容忍門檻 `±5%`、symbol（`▲`/`▼`/`≈` 與 ASCII fallback `^`/`v`/`~`）、顏色、ASCII mode、five_hour 不顯示等其他規格**完全不變**。

## Non-Goals (optional)

不在範圍內：
- 不調整 `±5%` 容忍門檻
- 不引入歷史快取（不做「過去 24h delta」之類的滾動式指標）
- 不改變 5h rate limit 的呈現
- 不改 symbol、不改顏色、不改 ASCII fallback
- 不改變 `resets_at` 缺失或視窗已過期時的抑制行為
- 不修改 countdown 格式

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `rate-limit-countdown`：修改 `Seven-day usage pace indicator` requirement 的 `expected_pct` 公式為日級線性，並調整對應 scenario 的範例與既有的 near-reset 範例。

## Impact

- Affected specs: `rate-limit-countdown`
- Affected code:
  - `internal/renderer/renderer.go`（`computePaceArrow` 內 `expectedPct` 計算）
  - `internal/renderer/renderer_test.go`（既有 over-pace / under-pace / within-tolerance / boundary / near-reset 測試的 expected 值需依新公式重算）
- 不影響 payload struct（`internal/model/payload.go`）、不影響 git/agent/cost 等其他 line 1/line 2 區段
