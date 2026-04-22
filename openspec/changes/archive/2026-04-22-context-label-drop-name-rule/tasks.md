## 1. 測試先行（Red）

- [x] 1.1 [P] 更新 `internal/renderer/renderer_test.go` 中依賴舊規則的斷言：`TestRenderStartup`（及任何使用 `jsonStartup` 的測試）— `display_name="Opus 4.6 (1M context)"` 配 `context_window_size=1000000` 的情境，原本預期沒有 `1M` label，改為**預期包含** `1M` label
- [x] 1.2 [P] 新增 `TestCtxLabel_ModelNameContainsContextStillShows1M` 直接測 `ctxLabel()`：`size=1_000_000`、`modelName="Opus 4.7 (1M context)"`、`exceeds200k=false` → 回傳值必須包含 `1M` 與 `ansiGray`
- [x] 1.3 [P] 新增 `TestCtxLabel_ModelNameContainsContextRedWhenExceeds`：`size=1_000_000`、`modelName="Opus 4.6 (1M context)"`、`exceeds200k=true` → 回傳值必須包含 `1M` 與 `ansiRed`
- [x] 1.4 執行 `go test ./internal/renderer/` 確認 RED（新增兩條+既有斷言失敗）

## 2. 實作（Green）

- [x] 2.1 修改 `internal/renderer/renderer.go` 的 `ctxLabel()`：刪除開頭 `if strings.Contains(nameLower, "context") { return "" }` 那三行，以及 `nameLower := strings.ToLower(modelName)` 若不再被使用也一併移除；函式簽名保持 `(size int64, modelName string, exceeds200k bool) string` 以減少呼叫端改動
- [x] 2.2 若 `modelName` 參數完全不再被使用，保留參數以維持呼叫端簽名，或改以 `_ string` 忽略；依 Go 慣例選可讀性高的方式（建議保留具名以便日後若再需要模型名時能直接用）
- [x] 2.3 執行 `go test ./...` 全部通過

## 3. 文件同步

- [x] 3.1 [P] 更新 `README.md` Line 1 表格「Context size」一列：移除 "Shown only when not already in the model name" 字樣，改為描述純由 `context_window_size` 驅動；紅色條件（`exceeds_200k_tokens=true`）保留
- [x] 3.2 [P] 更新 `README.zh-TW.md` 第一行表格「視窗大小」一列：移除「僅在模型名稱未包含此資訊時顯示」字樣，對齊英文版敘述

## 4. 驗證與歸檔

- [x] 4.1 `./dev.sh build` 編譯安裝
- [x] 4.2 手動構造 payload（`display_name="Opus 4.7 (1M context)"`, `context_window_size=1000000`, `exceeds_200k_tokens=false`）執行 `statusline.exe`，視覺確認畫面出現 `Opus 4.7 (1M context) ... 1M`（即使重複）
- [x] 4.3 手動構造第二份 payload（`display_name="Opus 4.7"`, `context_window_size=1000000`）執行 `statusline.exe`，視覺確認畫面出現 `Opus 4.7 ... 1M`，與第一個情境對齊
- [x] 4.4 執行 `spectra archive context-label-drop-name-rule` 讓 spec delta 套進 `openspec/specs/go-statusline/spec.md`（兩處舊的 "Context window size label" requirement 都會被新版覆蓋）
