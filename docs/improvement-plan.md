# Statusline 改善計畫

> 來源：2026-06-13 對整個 codebase 的全面 review（`go build` / `go vet` / `staticcheck` / `go test -race` / `gofmt` 皆通過，以下為工具抓不到的問題）。
> 用途：待辦清單。每項標註 `file:line`、問題、修法方向、優先序與狀態（☐ 待處理 / 🔄 進行中 / ☑ 完成），後續逐項處理。

---

## 進度總表（2026-06-14 全數處理完畢）

| 項 | 狀態 |
|----|------|
| C-1 + L-3（float 容錯 + 消除重複 struct） | ☑ `ed1c3a5` |
| H-1（cache 按 dir 隔離 + 原子寫入） | ☑ `dc97af6` |
| M-2（git dirty 錯誤碼區分） | ☑ `5299569` |
| M-3（rate-limit clamp） | ☑ `334711c` |
| Cache 命中率 feature（⚡ 緊貼 cost） | ☑ `bc8e6c6` |
| M-5（main.go `Printf`→`Print`） | ☑ `2537e2e` |
| M-7（`isCacheStale` ModTime 限制註解） | ☑ `75606ec` |
| M-4 + L-4（路徑 `Clean` + 移除死參數） | ☑ `45a9233` |
| H-2 — debug-tee log `0600` | ☑ `815665a`（註：實測 Windows 忽略 Unix mode、log 仍 `0644`；只在 Linux/macOS 生效） |
| H-2 — symlink / `O_NOFOLLOW` 硬化 | ✗ 不做：Windows 無 `O_NOFOLLOW`、cache 僅存 branch 名（非機密）、風險低 |
| M-6（line2 結尾換行） | ✗ 不改：無結尾換行是 statusline 慣例、現狀運作正常，加 `\n` 風險未知 |
| L-1（gitCmd env 健壯性） | ✗ 不改：review 自述為「非現行 bug」的提醒 |
| M-1（debug-tee `Run` error 被忽略） | ✗ 不做：debug 工具、已實測 `ExitCode()` 對 nil 不 panic |
| **L-2（NerdFont cost 雙空格）** | ⚠ **已調查待決定**：NerdFont `cost` 的 money glyph 遺失成空格（`od -c` 確認），導致 cost 段雙空格且無圖示（隔壁 time 有 ``）。修法是補 ``（nf-fa-money），但需在真 NerdFont 終端驗證 glyph 顯示，故未盲改。 |

> 下方為 2026-06-13 原始逐項清單（個別狀態欄未必逐一更新，**以上表為準**）。

## 一、已知問題

### CRITICAL

| # | 位置 | 問題 | 修法方向 | 狀態 |
|---|------|------|----------|------|
| C-1 | `internal/model/payload.go:42,49,20` | `int64` 欄位（`ContextWindowSize`、`TotalDurationMs`、`ResetsAt`）遇到 float 形式的 JSON 數值（如 `1000000.0`、`1e6`）時，`encoding/json` 會讓**整份 payload 解析失敗**，`main.go:29` 接著只顯示 `─ │ parse error`——連 model、cost 全消失。觸發條件不在本專案掌控（取決於 Claude Code 端如何序列化數字）。 | 把這些欄位改用 `json.Number` 容錯轉換；或改成「部分欄位失敗仍渲染其餘」的策略。對 statusline 而言後者更穩健。 | ☑（change: tolerant-payload-parsing） |

> 驗收附註（C-1，2026-06-13）：容錯涵蓋 `context_window_size` / `total_duration_ms` / `resets_at`（自訂 `tolerantInt64` 型別，string/bool/null → 0 且不整包 fail）。**`cost.total_lines_added` / `total_lines_removed` 仍是 `int`**，理論上有同類「float 形式會整包 fail」風險，但行數為純計數、實務上不會被序列化成 float，故本次未涵蓋；未來若要徹底防禦可一併改用 `tolerantInt64`。

### HIGH

| # | 位置 | 問題 | 修法方向 | 狀態 |
|---|------|------|----------|------|
| H-1 | `internal/gitcache/gitcache.go:15-17,22-40,52` | 所有 session 共用同一個 cache 檔路徑，且 `Get()` 的 cache-hit 路徑**不看 `dir` 參數**——同時開兩個 repo 的 session，5 秒 TTL 內會互相顯示對方的 branch/dirty。另 `writeCache` 用 `os.WriteFile`（truncate+write，非原子），併發下會讀到空檔。 | cache 檔名納入 `dir` 的 hash（一次解決跨專案污染＋大幅降低併發碰撞）；寫入改用 temp 檔 + `os.Rename` 原子覆蓋。 | ☑（change: isolate-gitcache-atomic-writes） |
| H-2 | `internal/gitcache/gitcache.go:15-17,52`、`cmd/debug-tee/main.go:27-29` | temp 檔名固定且可預測（`claude-statusline-git-cache`、`cc-statusline-debug.jsonl`），多人共用機器上有 symlink 預佔風險；`os.WriteFile`/`os.OpenFile` 不帶 `O_NOFOLLOW`。debug-tee 還以 `0644` 寫入完整 payload。 | cache 檔名納入 hash（見 H-1）降低可預測性；考慮 per-user 子目錄（`os.UserCacheDir()`，Windows 上 `os.Getuid()` 回 -1 需另處理）。cache 內容僅 branch 名（非機密），實務風險可下修；debug-tee 寫完整 payload 的部分較需重視。 | ☐ |

### MEDIUM

| # | 位置 | 問題 | 修法方向 | 狀態 |
|---|------|------|----------|------|
| M-2 | `internal/gitcache/gitcache.go:103-105` | `git diff --quiet` 的 exit code `>1` 代表「執行錯誤」（如 index.lock 被佔用、rebase/merge 進行中），目前把任何非 nil error 都當 dirty，會誤顯 `*`。 | 用 `exec.ExitError` 區分：exit 1 = 真 dirty，其他 = 保守視為非 dirty。 | ☑ |
| M-3 | `internal/renderer/renderer.go:368,342` | rate-limit 的 `UsedPercentage` 完全沒 clamp（context window 在 `:396` 有 clamp 作為對照）。病態大值會讓 `int(math.Round(...))` 溢位，輸出 `7d:-9223372036854775808%`。 | 在 `formatRate` / `computePaceArrow` 對 pct 與 magnitude 做 `clamp(..., 0, 100)`（float 層先夾避免 int 溢位）。 | ☑ |
| M-4 | `internal/renderer/renderer.go`（`directoryDisplay` / `resolveProjectRoot`） | 缺 `filepath.Clean`，symlink 或 case-insensitive 檔案系統下 `filepath.Rel` 可能誤判 ancestor，目錄顯示偶發不正確。 | 進入點對 `currentDir` / `projectDir` 做 `filepath.Clean`（必要時 `EvalSymlinks`）。 | ☐ |
| M-5 | `cmd/statusline/main.go:29` | 字串無格式參數卻用 `fmt.Printf`；且 `:52` 的 `fmt.Printf("%s\n%s", line1, line2)` 很容易被未來誤改成 `Printf(line2)` 而引入 format-string bug（branch 含 `%` 會炸）。 | parse-error 行改 `fmt.Print`；輸出行改用不解析格式的 API（如 `fmt.Print(line1, "\n", line2)`）。 | ☐ |
| M-6 | `cmd/statusline/main.go:52` | line2 結尾無 `\n`，與 parse-error 分支（`:29` 有 `\n`）行為不一致；某些終端/管線情境下會與後續輸出黏行。 | 確認宿主需求後補上尾端 `\n`（除非宿主明確要求不能有）。 | ☐ |
| M-7 | `internal/gitcache/gitcache.go:71-77` | `isCacheStale` 用 `time.Since(ModTime())`，系統時鐘回撥（NTP 校正、改時間）時 cache 可能被當成「永遠新鮮」，狀態卡住。 | 檔案型 TTL 的固有限制，嚴重度低；至少在註解標明已知限制。 | ☐ |
| ~~M-1~~ | `cmd/debug-tee/main.go:46-47` | （原 review 判為 nil `ProcessState` 會 panic，**已實測修正**：`ExitCode()` 對 nil 回傳 -1，不會 panic。）真正的問題降級為：`cmd.Run()` 的 error 被忽略，real binary 不存在時靜默以 exit -1 結束、無錯誤訊息。 | 顯式處理 `cmd.Run()` 的 error 並印出診斷訊息。優先序低（debug 工具）。 | ☐ |

### LOW

| # | 位置 | 問題 | 修法方向 | 狀態 |
|---|------|------|----------|------|
| L-1 | `internal/gitcache/gitcache.go:113` | `gitCmd` 每次複製整個 `os.Environ()`；目前 `append` 目標是新建 literal slice，安全。屬健壯性提醒，非現行 bug。 | 維持現狀，留意未來別複用傳入 slice。 | ☐ |
| L-2 | `internal/renderer/renderer.go:73-105,440` | NerdFont 模式 `cost` 圖示含前導空格，`:440` 又是 `sym.cost + "$"`，需確認不會出現雙空格。 | 補一個 NerdFont 模式的 golden test 確認對齊。 | ☐ |
| L-3 | `internal/model/payload.go:35-108` | `Payload` 與 `payloadJSON` 兩個 struct 幾乎完全重複，新增欄位要同步改兩處，易漂移。 | 讓 `RateLimit` 自行實作 `UnmarshalJSON`（指標語意偵測 presence），即可砍掉整個 `payloadJSON` 鏡像 struct。**與 C-1、cache 命中率 feature 一起做最划算**（都動 payload.go）。 | ☑（隨 C-1 完成） |
| L-4 | `internal/renderer/renderer.go:217` | `ctxLabel` 的 `modelName` 參數已是死參數（commit d45fdb7 拿掉規則後殘留）。 | 移除參數，或註解保留原因。 | ☐ |

### 測試缺口

- **【對應 C-1】** float 型態數值欄位：完全沒測 `context_window_size:1000000.0`、`total_duration_ms:1e6` 這類合法 JSON。一個 table-driven test 就能釘死回歸。
- **【對應 H-1】** gitcache 併發寫入與跨目錄隔離：`gitcache_test.go` 無併發測試，也沒測「不同 dir 不共用 cache」。
- **【對應 M-2】** git dirty 的錯誤碼區分：沒測 git exit `>1` 時 dirty 應為 false。
- **【對應 M-3】** rate-limit 極端百分比（負數、>100、超大值）是否被 clamp。
- **負數 / 邊界值**：`total_cost_usd`、`total_lines_added`、`used_percentage` 為負時的渲染。
- **ANSI golden test**：目前都用 `strings.Contains` 鬆散斷言，抓不到漏 reset 造成的顏色溢出。建議對 normal/danger 兩情境加完整 byte 比對。
- **`Render` 對 `nil *model.Payload`** 的防禦或 contract 測試。
- **`main.go` / `debug-tee` 無測試**：`isenv`、parse-error fallback 路徑可抽成可測函式。

---

## 二、新功能：Cache 命中率指標

> ✅ **已完成** — commit `bc8e6c6`（change cache-hit-rate-statusline）。定案：⚡<pct>% 緊貼 cost、ASCII cache:<pct>%、三段色（≥80 灰 / 50-79 黃 / <50 紅）、`current_usage` null 或分母 0 不顯示。真 binary 驗證通過。

### 動機

對「按 token 計費」（如 Enterprise）的情境，**prompt cache 是成本最大的槓桿**——cache read 約為標準 input 價的 10%。習慣性開新 session、改 system prompt 導致 cache 全失效時，成本可差近一個數量級。在 statusline 顯示即時命中率能讓工程師對「自己的用法是否省」立刻有感，是這個工具最貼合 Enterprise 價值的延伸。

### 資料來源與現況

Claude Code 的 payload 其實已提供完整 token 細分，但目前 `model.ContextWindow` **只解析了 `used_percentage` 和 `context_window_size`，其餘全忽略**。實際 payload（取自 `debug.json`）：

```json
"context_window": {
  "total_input_tokens": 738,
  "total_output_tokens": 104751,
  "context_window_size": 200000,
  "current_usage": {
    "input_tokens": 1,
    "output_tokens": 374,
    "cache_creation_input_tokens": 1302,
    "cache_read_input_tokens": 144198
  },
  "used_percentage": 73,
  "remaining_percentage": 27
}
```

> 重要語意：`current_usage` 反映的是**最近一次 API 請求**，不是整個 session 累計，會逐輪波動（`/compact` 後第一輪會明顯偏低）。只有 `current_usage` 有 cache 維度的細分，`total_input_tokens` / `total_output_tokens` 沒有，所以命中率只能用 `current_usage` 算，代表「上一輪請求的 cache 效率」。

### 計算方式

命中率定義在 input 側（output 不納入）：

```
分母 = input_tokens + cache_creation_input_tokens + cache_read_input_tokens
分子 = cache_read_input_tokens
hit_rate = 分子 / 分母
```

以上面範例：`144198 / (1 + 1302 + 144198) = 144198 / 145501 ≈ 99.1%`。

### 顯示設計（待定，先列選項）

- 位置：Line 1，接在 rate-limit 後或 cost 附近。
- 圖示：NerdFont/Unicode `⚡99%`；ASCII fallback `cache:99%`。
- 顏色：命中率高是好事——可考慮「低於某門檻才用提示色（灰/紅）」提醒 cache 失效，高命中維持低調色。具體門檻待定。

### 容錯與邊界（呼應 C-1 防禦原則）

- `current_usage` 在 **session 開始前 / `/compact` 後**為 `null` → 這段直接不顯示（回空字串），不可讓整條掛掉。
- 分母為 0（session 剛開始 input 全 0）→ 不顯示，避免除以零。
- 新增欄位採容錯解析，缺欄位 / 型態不符不影響其他段落（與 C-1 同一原則）。

### 實作步驟

1. `model.ContextWindow` 新增 `CurrentUsage` struct（`input_tokens` / `output_tokens` / `cache_creation_input_tokens` / `cache_read_input_tokens`），用 presence 偵測（指標或 `*struct`）區分 `null`。
2. renderer 新增 `formatCacheHit(...)`，含上述容錯與邊界。
3. Line 1 組裝加入該段（NerdFont / ASCII / 顏色分支）。
4. 測試：命中率計算、`null` 時不顯示、分母為 0 不顯示、ASCII/NerdFont 輸出。

---

## 三、建議處理順序

| 順序 | 項目 | 理由 |
|------|------|------|
| 1 | **C-1**（float 容錯） | 影響「整條 statusline 是否消失」，最高優先；同時建立 payload 容錯框架，後續欄位都受惠。 |
| 2 | **H-1**（cache 按 dir 隔離 + 原子寫入） | 多 session 日常會實際踩到的正確性問題。 |
| 3 | **M-3 + M-2**（rate-limit clamp、git dirty 錯誤碼） | 成本極低的正確性修復，可一併處理。 |
| 4 | **Cache 命中率 feature** | 新功能；資料現成（與 C-1 同改 `payload.go`，可順帶做 L-3 消除重複 struct）。最貼合 Enterprise 價值。 |
| 5 | 其餘 MEDIUM / LOW + 測試補強 | M-4 / M-5 / M-6 / M-7 / L-* 與對應測試。 |

> 每項建議走 `/spectra:propose` 開 change，再 `/spectra:apply` 實作；完成後回來把對應狀態改 ☑。

---

## 四、新功能 backlog：執行模式顯示 + config 改 flag

> 來源：2026-06-14 brainstorming（payload 新欄位 + env var 不方便的反思）。狀態：☐ 待 propose。

### 目標 1 — 顯示執行模式（effort / thinking / fast_mode）✅ 已完成（commit `2a44a65`：⚙level + on 後綴 T/F、漸強色 scale、真 binary 驗證）

最新 Claude Code payload（`2.1.177`+）新增欄位（4/23 的 `2.1.114` 還沒有）：
- `effort.level`（low/medium/high/xhigh/max）：反映 `/effort` 即時設定，直接影響 thinking token 用量 → **成本**。模型不支援 effort 時此欄位 absent。
- `thinking.enabled`（bool）：extended thinking 開關。
- `fast_mode`（bool）：Opus fast 模式。

在 statusline 顯示這組「執行模式 = 火力 = 成本」指標。`model.ContextWindow` 所在的 payload struct 需新增解析這三欄位（effort 用指標 / presence 偵測 absent，沿用 `tolerantInt64` 等容錯框架）。顯示位置 / 格式 / 顏色待 propose 定（effort 高訊號優先；thinking / fast 訊號弱、可搭配成一個「模式」小區）。

### 目標 2 — config 機制 env var → 命令列 flag + 全段落可控

**動機**：env var 設定不方便（Windows 麻煩、藏在系統看不見、改了要重啟、跟 `settings.json` 脫節）。statusLine 本來就在 `settings.json` 用 command 定義，config 應寫成 command 參數、一眼看到。

- 現有 env var（`CLAUDE_STATUSLINE_ASCII` / `_NERDFONT` / `_POWERLINE`）→ flag（`--ascii` / `--nerdfont` / `--powerline`）。
- **決定 B：直接廢棄 env var、只留 flag**（breaking change，`README.md` / `README.zh-TW.md` 要標明升級方式）。`COLORTERM` 是系統標準變數、非本專案 config，保留。
- 新增 `--hide <keys>`（逗號分隔）控制**所有段落**顯示 / 隱藏，例：`--hide effort,duration,rate_limits`。預設全顯示，使用者列出要藏的。
- 每個 line 1 / line 2 段落需有 key（model / bar / cost / cache / effort / duration / rate / branch / lines / dir / agent …）。
- `settings.json` 範例：`"command": "statusline.exe --nerdfont --hide effort,duration"`

**影響**：`cmd/statusline/main.go` arg 解析、`renderer.Options`（改從 flag 來）、`renderer` 每段掛 key 判斷、兩份 README 升級說明、移除現有 env var 解析與測試。

> 兩個目標建議**一個 change 一起做**（目標 1 的新段落天生納入目標 2 的 hide 機制）。走 Spectra propose。

---

## 附錄：驗證指令

```bash
go build ./...                          # 編譯
go vet ./...                            # 靜態分析
staticcheck ./...                       # 進階靜態分析
go test -race ./...                     # 含資料競爭偵測
gofmt -l .                              # 格式檢查（無輸出 = 通過）
./dev.sh last-json                      # 抓最新 Claude payload → debug.json
```
