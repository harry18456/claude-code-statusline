## Context

`computePaceArrow`（`internal/renderer/renderer.go:325-353`）目前以秒級線性比例計算 expected_pct。`elapsed` 用 `resets_at - now` 反推，再除以 `604800` 得到一個連續的百分比。

實作上：
```go
elapsed := sevenDayWindowSeconds - remaining
expectedPct := float64(elapsed) / float64(sevenDayWindowSeconds) * 100
```

使用者回饋指出這個粒度太細：週期才剛開始幾小時時 expected ≈ 4%，導致「第 1 天用 6%」會被當成「微微超前」而顯示中性 `≈`。直觀上，使用者把週期視為 7 個離散的「日配額」（每日 14.2857%），希望進入第 N 天時 expected 直接到 `N × 14.2857%`。

## Goals / Non-Goals

**Goals:**

- expected_pct 改為日級線性，跟使用者「第 N 天該到 N×14.3%」的心智模型對齊
- 跳階點對齊 `resets_at` 鐘點，不依賴日曆午夜（讓非整點 reset 的使用者也能正確使用）
- 維持「外部介面」最小變動：symbol、顏色、容忍門檻、ASCII fallback、five_hour 抑制、resets_at 缺失/過期抑制等規格不變
- 維持單一純函數，不引入狀態、不存歷史、不讀檔

**Non-Goals:**

- 不引入「過去 24h 滾動 delta」這類需要本地快取的方案（雖然能更貼近「今天用了多少」，但需要額外狀態管理，超出本 change 範圍）
- 不調整 `±5%` 容忍門檻
- 不修改 5h 行為
- 不修改 countdown / cost / git / agent 等其他區段
- 不重寫測試以外的呼叫點（`formatRate` 仍呼叫 `computePaceArrow`，呼叫端不變）

## Decisions

### Use ceil for elapsed_days

公式 `elapsed_days = ceil(elapsed / 86400)`，而非 `floor` 或 `round`。

**Rationale:**

- `floor`：第 1 天的 expected = 0%，整天都不會顯示落後 — 違反使用者期待（第 1 天就該對齊每日配額）
- `round`：前半天 expected = 0、後半天 expected = 14.3% — 跳階點落在「第 1 天 12 小時時刻」，語意不直觀
- `ceil`：進入第 1 天的瞬間 expected = 14.3%、進入第 2 天的瞬間 expected = 28.6%……語意上等同「進入第 N 天時，已該用滿 N 天的配額」，跟使用者表述完全對齊

**Edge case — `remaining = 604800`（剛 reset 的瞬間）:**

- elapsed = 0 → ceil(0 / 86400) = 0 → expected_pct = 0%
- 這個瞬間正確：剛重置，期望使用量為 0
- 但下一秒 elapsed = 1 → ceil(1/86400) = 1 → expected_pct = 14.2857%
- 即「重置後第 1 秒就期望 14.3% 已被用掉」，這是 ceil 的代價：跳階點被推到視窗的最開頭

我們接受這個代價，因為：
1. 視窗起點極短時段內 used 通常也是 0，deviation = 0 - 14.3 = -14.3，會顯示 `▼14%`
2. 這正確傳達「你才剛進入新週期、整天的配額還沒開始用」的訊息
3. `floor` 會在第 1 天整天都沒提示，反而更有問題

### Cap elapsed_days at 7

`expectedPct` 理論上應落在 `[0, 100]`。但若 `now > resets_at - 0`（剛好邊界）或 clock skew 造成 `elapsed` 略大於 `604800`，`ceil(elapsed / 86400)` 可能 > 7。

實作上 `remaining <= 0` 已先 return `""`（不顯示指標），所以 `elapsed < 604800` 恆成立，最大 `ceil(elapsed/86400) = 7`。但為了防禦性程式設計，仍夾住 `elapsed_days = min(7, ceil(elapsed/86400))`。

### Tolerance threshold remains ±5%

新公式下，`±5%` 仍是合理門檻：

- 5% 對應每日配額 14.2857% 的約 35%（5/14.2857 ≈ 0.35）— 即「日配額 1/3 的偏差」才會觸發箭頭
- 在「進入新一日的瞬間」expected 跳升 14.3%，若 used 沒同步跳升（自然不會），deviation 會立即從 ~0 變成 ~−14%，跨過 −5% 門檻 → 從 `≈` 變 `▼14%`
- 這個切換正是使用者要的：跨日後立即提示「新一日的配額你還沒用」

### No magnitude formatting changes

`magnitude = round(abs(deviation))` 整數化規則維持不變。輸出格式 `▲<N>%` / `▼<N>%` / `≈` 完全相同。

## Risks / Trade-offs

- **跨日邊界的觀感跳變** → 進入新一日的瞬間 deviation 會跳 14.3%，從 `≈` 切到 `▼14%`（或反向）會讓使用者覺得「前一秒還正常、下一秒突然落後」。Mitigation：這正是新行為的設計意圖（提示「進入新一日」），且 statusline 並非每秒重繪，使用者多半在下一次 prompt 才看到變化，不會體驗到「秒級跳動」。
- **既有測試需要更新** → renderer_test.go 內所有 pace-related 測試的 expected 值是用秒級公式計算的，需依新公式重算。Mitigation：在 tasks.md 列出每個測試案例的新 expected 值；先 RED（測試失敗）再 GREEN。
- **使用者已適應舊行為的可能性** → 已使用此功能的使用者可能習慣舊的 `≈` 出現條件。Mitigation：CHANGELOG / commit message 說明變更原因，並指明「新行為更貼近每日配額直覺」。
- **第 7 天最後幾秒 expected = 100%** → 若使用者剛好沒用滿（例如 used = 95%），deviation = -5（落在容忍邊界）→ `≈`。略低（94%）→ `▼6%`。這是合理行為，不需特殊處理。
