## Summary

移除 7d pace indicator 的 ±5% 容忍門檻，永遠以方向性指標 `▲<N>%` / `▼<N>%` 取代中性符號 `≈`。

## Motivation

v1.5.0 將 `expected_pct` 改為日級線性後，使用者實際使用一段時間發現：當 used 落在 ±5% 容忍區間內仍會看到 `≈`（例：第 1 天 used=10、expected=14.29、dev=-4.29 → `≈`），但使用者**真正想要的是永遠的方向性回饋**——「我目前略微落後/超前」比模糊的「節奏正常」更有行動價值。`≈` 視覺上是雜訊：它告訴使用者「不必看了」，但使用者會問「那到底差多少？」需要符號 + 數字才能判斷。

## Proposed Solution

- 將容忍門檻 `±5%` 直接設為 `0`：`deviation > 0` → `▲<N>%`、`deviation < 0` → `▼<N>%`
- `≈` 僅在 `deviation == 0`（理論上極罕見，因 `expected_pct = N × 100/7` 永遠非整數，而 `used_percentage` 觀察到都是整數）時觸發；實務中幾乎不會出現
- `magnitude` 計算保持 `round(abs(deviation))`，但對非零 deviation 取 `max(1, round(|deviation|))` 作下限，避免 `▲0%` / `▼0%` 這種視覺怪異
- ASCII fallback、顏色（紅/灰）、five_hour 抑制、resets_at 缺失/視窗已過期抑制、symbol 形狀均**不變**

## Non-Goals (optional)

- 不改變 `expected_pct` 的日級線性公式（v1.5.0 剛確立的行為）
- 不引入新的容忍/平滑機制（不做 EMA、不做歷史比較）
- 不改變 5h 區段
- 不改 symbol 字元（仍用 `▲` / `▼` / `≈` 與 ASCII fallback `^` / `v` / `~`）
- 不引入小數點顯示（先前討論已暫緩）
- 不修改 README 中對 ASCII fallback 的描述

## Alternatives Considered (optional)

- **保留容忍但改用更小門檻（例如 ±1%）**：仍會有少數 `≈` 出現，使用者抱怨的「為何還有 ≈」依然發生，未根本解決。
- **完全移除 `≈` 字元**：會讓 spec 在數學上 `dev == 0` 的邊界情況沒有定義；保留 `≈` 涵蓋此邊界即可，無額外維護負擔。
- **magnitude 用 `ceil(|dev|)`**：會讓 dev=0.3 顯示 `▲1%`、dev=1.4 顯示 `▲2%`，跟既有 `round` 行為差異大。改用 `max(1, round(|dev|))` 保留既有 round 慣例，僅補上 1% 下限。

## Impact

- Affected specs: `rate-limit-countdown`（修改 `Seven-day usage pace indicator` requirement 的容忍門檻與相關 scenarios）
- Affected code:
  - `internal/renderer/renderer.go`（`computePaceArrow` 的分支條件與 magnitude 下限）
  - `internal/renderer/renderer_test.go`（既有 within-tolerance 測試需重新對應「永遠有方向」的新行為，新增 magnitude 下限測試）
- Documentation: `README.md` 與 `README.zh-TW.md` 描述 7d pace 段落的「±5% 容忍」字句移除
