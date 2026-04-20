## 1. Payload 欄位解析

- [x] 1.1 在 `internal/model/payload_test.go` 增加測試覆蓋 "Parse exceeds_200k_tokens field" requirement：present-true / present-false / absent 三種輸入皆能解析為對應布林值（absent 時為 false）
- [x] 1.2 在 `internal/model/payload.go` 的 `Payload` 與 `payloadJSON` struct 新增 `ExceedsTokens200k bool` 欄位（JSON tag `exceeds_200k_tokens`）並於 `ParsePayload` 中完成 "Parse exceeds_200k_tokens field" 的轉寫

## 2. 1M 門檻警示

- [x] 2.1 在 `internal/renderer/renderer_test.go` 增加測試覆蓋 "Context window size label" 修訂行為：(a) 1M + exceeds=false → 灰色，(b) 1M + exceeds=true → 紅色，(c) 200k 情境保持灰色不變
- [x] 2.2 修改 `internal/renderer/renderer.go` 的 `ctxLabel()` 函式，落實 design 決策 "1M label turns red when `exceeds_200k_tokens=true`"：擴充函式簽名接收 `exceeds200k bool`，僅在 size >= 1M 且 exceeds200k=true 時使用 `ansiRed`，其餘情況維持 `ansiGray`
- [x] 2.3 於 `Render()` 呼叫 `ctxLabel()` 處傳入 `p.ExceedsTokens200k`，讓新行為生效

## 3. 子目錄相對路徑顯示

- [x] 3.1 在 `internal/renderer/renderer_test.go` 增加測試覆蓋 "Directory path display on line 2" 所有 scenarios：(a) current==project → base name，(b) current 為 project descendant → `<project_base>/<relative>`（使用正斜線），(c) project 為空 → fallback 至 base name，(d) current == "" 或 "." → 顯示 `.`
- [x] 3.2 在 `internal/renderer/renderer.go` 新增純函式 `directoryDisplay(currentDir, projectDir string) string`，落實 design 決策 "Subdir display: `<project>/<relative>` when distinct"；路徑計算使用 `filepath.Rel` 並將分隔符正規化為 `/`
- [x] 3.3 在 `Render()` Line 2 目錄片段改為呼叫 `directoryDisplay(p.Workspace.CurrentDir, p.Workspace.ProjectDir)`，並於 `Payload` / `payloadJSON` struct 補上 `ProjectDir string` 欄位對應 `workspace.project_dir`

## 4. 7d 使用節奏箭頭

- [x] 4.1 在 `internal/renderer/renderer_test.go` 增加測試覆蓋 "Seven-day usage pace arrow" 所有 scenarios：over-pace、under-pace、within tolerance、near-reset suppression、resets_at absent、ASCII fallback、以及「Five-hour never shows pace arrow」驗證 5h 永不出現箭頭；同時覆蓋 design 決策 "Arrow symbols and colors" 的 `▲`/`▽` 與 ASCII `^`/`v` 對應
- [x] 4.2 在 `internal/renderer/renderer.go` 新增純函式 `computePaceArrow(rl model.RateLimit, now time.Time, opts Options) string`，落實 design 決策 "Use linear expected-usage model for 7d pace"、"Deviation threshold: ±5%"、"Suppress arrow when < 10% window time remains"、"Arrow symbols and colors"；參數 `rl` 傳入已知為 seven_day 的 RateLimit，函式依 `604800` 常數計算 expected_pct，回傳含顏色碼的箭頭字串或空字串
- [x] 4.3 修改 `formatRate()` 使其落實 design 決策 "Show pace arrow only for `seven_day`"：僅當 `label == "7d"` 時呼叫 `computePaceArrow()`，並將回傳字串插入於百分比之後、countdown 之前（e.g., `7d:55%▲ (4d 2h)`）；`5h` 路徑保持不變

## 5. 驗證與整合

- [x] 5.1 執行 `go test ./...` 全部通過
- [x] 5.2 執行 `./dev.sh build` 確認編譯成功並安裝至 `~/.claude/statusline.exe`
- [x] 5.3 以 `./debug.json` 或手動構造 payload 執行 `statusline.exe < debug.json`，視覺確認 (a) subdir 顯示正確，(b) exceeds_200k_tokens=true 時 `1M` 標籤為紅色，(c) `seven_day` pace 箭頭在預期位置與顏色出現

## 6. Git-root fallback for project root

- [x] 6.1 在 `internal/renderer/renderer_test.go` 增加測試覆蓋 "Directory path display on line 2" 的新情境（使用 `t.TempDir()` 建構暫存目錄結構，不依賴真實 git）：(a) payload `project_dir` 為 ancestor → 沿用現行行為，(b) `project_dir` 等於 `current_dir` 且上層存在 `.git` 目錄 → 以 `.git` 所在層為 root 顯示 `<root_base>/<relative>`，(c) `.git` 以 file 形式存在（worktree 結構，寫入最小內容）→ 行為同 (b)，(d) submodule 情境（current 下層有 `.git` file，更上層另有 `.git` dir）→ 第一個 `.git` 勝出，(e) 向上走到 filesystem root 仍找不到 `.git` → fallback 到 `filepath.Base(current_dir)`
- [x] 6.2 在 `internal/renderer/renderer.go` 新增純函式 `resolveProjectRoot(currentDir, payloadProjectDir string) string`，落實 design 決策 "Subdir display: `<project>/<relative>` when distinct" 的三層 fallback：(1) payload `project_dir` 為 strict ancestor 時回傳之，(2) 否則從 `current_dir` 逐層向上用 `os.Stat` 檢查 `.git`（file 或 dir 皆算），第一個命中回傳該層，(3) 走到 `filepath.Dir(p) == p` 仍未命中回傳空字串；並將 `directoryDisplay()` 重寫成呼叫 `resolveProjectRoot()` 後依 root 為空 / 等於 current / 為 current 祖先三種情境決定輸出
- [x] 6.3 以手動構造 payload（`current_dir` = `internal/renderer`、`project_dir` = `internal/renderer`）在 repo 內實際執行 `statusline.exe` 驗證 line 2 出現 `claude-code-statusline/internal/renderer`，確認新 fallback 在「subfolder 啟動」情境實際生效

## 7. Pace 指示顯示偏差量與中性符號

- [x] 7.1 在 `internal/renderer/renderer_test.go` 更新既有 pace 測試並新增覆蓋：(a) over-pace 偏差 7% → 回傳字串包含 `▲7%` 且為紅色，(b) under-pace 偏差 3% → 回傳字串包含 `▼3%` 且為灰色，(c) 偏差 ±5% 內（含 0、±1、±5 邊界）→ 回傳字串為 ` ≈` 灰色，(d) near-reset (<10% window) 仍為空字串，(e) `resets_at` 缺失仍為空字串，(f) ASCII 模式：over/under 為 `^7%`/`v3%` 無顏色碼、within-tolerance 為 ` ~`、near-reset 仍空；同步更新 `TestFormatRate_SevenDayOverPace` / `TestFormatRate_SevenDayUnderPace` / `TestFormatRate_SevenDayWithinTolerance` 等整合測試以反映新格式
- [x] 7.2 修改 `internal/renderer/renderer.go` 的 `computePaceArrow()` 函式，落實 spec "Seven-day usage pace indicator" 新規格：(a) 計算 `magnitude := int(math.Round(math.Abs(deviation)))`，(b) `deviation > 5` 回傳 `<red>▲<N>%<reset>`（ASCII 模式回傳 `^<N>%` 無顏色），(c) `deviation < -5` 回傳 `<gray>▼<N>%<reset>`（ASCII 回傳 `v<N>%`），(d) 偏差 ±5% 內回傳 `<gray>≈<reset>`（ASCII 回傳 `~`），(e) near-reset 或 resets_at=0 仍回傳空字串；更新 `formatRate()` 呼叫端已加空格前綴邏輯不變
- [x] 7.3 執行 `go test ./...` 全部通過後，以 `./dev.sh build` 安裝，以手動構造 payload（`seven_day.used_percentage` 分別為明顯超支/落後/接近預期三組值並搭配合理 `resets_at`）執行 `statusline.exe` 視覺確認 `▲<N>%` 紅色、`▼<N>%` 灰色、`≈` 灰色三種形態正確顯示
