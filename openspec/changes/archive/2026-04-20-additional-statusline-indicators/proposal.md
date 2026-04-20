## Why

Claude Code 的 statusLine payload 其實還帶了一些目前 renderer 沒用到、但**對使用者決策有實際行動性**的欄位。具體盤點後，有三個訊號值得補上：

1. **進入 1M context 貴價區（`exceeds_200k_tokens=true`）** — 此時 input 變 2×、output 變 1.5× 計費，但目前看不出差異，容易在不自覺中燒錢。
2. **在專案子目錄工作時定位不清** — 目前只顯示 `filepath.Base(current_dir)`，若 `current_dir ≠ project_dir`（例如在 `project/internal/renderer` 工作），畫面只會出現 `renderer`，看不出身處哪個專案的子目錄。
3. **7 天週期使用節奏不明** — 目前只有 `≥80%` 才紅色警告，屬於事後。使用者缺少**預警**指標，無法及早知道「依這個速度會在重置前用完」。

三個指標共通特性：**看到後使用者會做出具體行動**（`/compact`、改變使用節奏、確認工作位置），符合 statusline 「驅動決策」而非「純展示」的原則。

## What Changes

- **1M 門檻警示**：當 `exceeds_200k_tokens=true` 時，將現有 `1M` 灰色標籤改為紅色，提示目前每個 token 正以加倍單價計費。
- **子目錄相對路徑顯示**：Line 2 依序嘗試三層規則決定顯示 project root，讓「在子目錄啟動 Claude Code」亦能看到 `<project>/<relative>` 定位：
  1. 若 `workspace.project_dir` 是 `current_dir` 的嚴格祖先 → 以其為 root
  2. 否則從 `current_dir` 逐層向上尋找 `.git`（dir 或 worktree 的 file） → 該層為 root
  3. 都失敗 → fallback 到 `filepath.Base(current_dir)`（保留現行行為）

  root 判定後，`current_dir == root` 顯示 `<root_base>`，否則顯示 `<root_base>/<relative>`。實作僅用 `os.Stat` 偵測 `.git`，不呼叫 `git` CLI；因此 git 未安裝或非 git 倉庫都能正常 fallback。
- **7d 使用節奏指示**：在 `7d:<pct>%` 後加一個空格再加 pace 指示，依「線性預期使用量」比較實際使用量：
  - 超支 > 5% → 紅色 `▲<N>%`（`N` 為 `round(|deviation|)`，例如 `▲7%`）
  - 落後 > 5% → 灰色 `▼<N>%`（例如 `▼3%`）
  - 誤差 ±5% 內 → 灰色 `≈`（無數字）
  - 剩餘時間 < 10% window 長度或 `resets_at` 缺失時不顯示（線性模型意義衰減）
  - 僅對 `seven_day` 套用；`five_hour` 因窗口過短且使用天然不均勻，**不**顯示 pace
  - ASCII 模式下退回 `^<N>%` / `v<N>%` / `~`

## Non-Goals (optional)

- 不顯示 `cache_read_input_tokens` / `cache_creation_input_tokens` 導出的 cache 命中率 — 無行動性（成本已於 `$` 呈現）。
- 不顯示 `total_api_duration_ms` 與 `total_duration_ms` 的拆解 — 無行動性（tool 執行時間非使用者可控）。
- 不顯示 Claude Code `version`、`output_style.name` — 前者為低頻查詢（透過 `/help` 可得），後者觸發情境過邊緣。
- 不改動 5h rate limit 的顯示方式。
- 不改動既有顏色閾值（`≥80%` 紅色警告維持）。

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `go-statusline`: 擴充 context label 與目錄顯示邏輯 — 加入 `exceeds_200k_tokens` 判斷以及 `workspace.project_dir` 比對。
- `rate-limit-countdown`: 擴充 7 天 rate limit 顯示 — 依線性預期模型計算 pace 箭頭並附加至 `seven_day` 顯示字串。

## Impact

- Affected specs:
  - `openspec/specs/go-statusline/spec.md`（1M 門檻、子目錄路徑）
  - `openspec/specs/rate-limit-countdown/spec.md`（7d pace 箭頭）
- Affected code:
  - `internal/model/payload.go` — 新增 `exceeds_200k_tokens` 欄位解析
  - `internal/renderer/renderer.go` — 更新 `ctxLabel()`、Line 2 目錄片段、`formatRate()` 新增 pace 計算
  - `internal/renderer/renderer_test.go` — 新增上述三條分支的測試
  - `internal/model/payload_test.go` — 新增 `exceeds_200k_tokens` 欄位測試
