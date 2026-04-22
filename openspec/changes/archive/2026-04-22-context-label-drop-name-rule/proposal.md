## Why

Claude Code 的 `/model` 選單現在把同一個 1M-capacity 模型包成兩個選項（例如 "Opus 4.7 (1M context)" 與 "Opus 4.7"），但 payload 兩者都吐 `context_window_size: 1_000_000`。現行 `ctxLabel()` 有一條「模型名含 'context' 就隱藏 label」的去重規則，導致兩個選項的畫面不一致：

- `Opus 4.7 (1M context)` → 名字擋下 label → 畫面無 `1M` 提示
- `Opus 4.7` → 規則不觸發 → 畫面出現 ` 1M`

這規則原本是為了避免早期「名字含 1M」時重複顯示（例：`Opus 4.6 (1M context) 1M`），但 Claude Code 的 naming 已脫鉤，規則產生的「一致性」反而變成「不一致」。與其猜測哪種命名代表哪種 capacity，**直接信任 payload 的 `context_window_size` 並一律依它顯示**更可預測。

## What Changes

- **BREAKING**（視覺層）：移除 `ctxLabel()` 中「modelName 含 'context' 就回傳空字串」的那一支。`1M` / `200k` label 僅由 `context_window_size` 決定。
- 結果：當 display_name 為 `Opus 4.7 (1M context)` 時，畫面將變為 `Opus 4.7 (1M context) 1M`（帶冗餘，但語意明確且與其他選項對齊）。
- 更新 `internal/renderer/renderer_test.go` 中依賴舊規則的測試。
- 更新 `openspec/specs/go-statusline/spec.md` 中兩處 "Context window size label" requirement：刪去 "model name does not contain 'context' or 'Context'" 條件。
- 更新 README（英 / 繁）的 Line 1 表格，移除「僅在模型名稱未包含此資訊時顯示」敘述。

## Non-Goals (optional)

- 不調整 `200k` / `1M` 本身的顯示顏色或閾值（已在前一 change 敲定）。
- 不嘗試猜測 Option 1 vs Option 4 的實際 API 差異（Claude Code 內部行為，不在 statusline 層處理）。
- 不改 `exceeds_200k_tokens` 紅色警示邏輯。

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `go-statusline`: "Context window size label" requirement 移除模型名過濾條件，改為純粹依 `context_window_size` 決定 label。

## Impact

- Affected specs:
  - `openspec/specs/go-statusline/spec.md`（兩處 "Context window size label" requirement）
- Affected code:
  - `internal/renderer/renderer.go` — `ctxLabel()` 移除名字檢查那段
  - `internal/renderer/renderer_test.go` — 更新既有測試：`jsonStartup` 的斷言（display_name 含 "1M context" 原本預期沒 label，新行為需預期有 label）、任何依賴舊規則的 scenario
  - `internal/model/payload_test.go` — 若有對應斷言需同步
- Affected docs:
  - `README.md` / `README.zh-TW.md` — 「視窗大小」列移除「僅在模型名稱未包含此資訊時顯示」字樣
