## Context

目前 `internal/renderer/renderer.go` 已渲染大部分 payload 欄位，但盤點發現仍有三個具行動性的訊號未顯示：

- `exceeds_200k_tokens` — Claude 1M context 模式跨過 200k 門檻後，input/output 計費倍率提高，目前畫面上與尚未跨過時完全相同。
- `workspace.project_dir` — 目前 Line 2 目錄片段只用 `filepath.Base(current_dir)`，在子目錄工作時會遺失「身處哪個專案的哪個子目錄」的上下文。
- 7 天 rate limit 使用節奏 — 現僅在 `≥80%` 才紅色警告，屬事後；使用者缺乏「按目前速度會否在重置前用完」的預警。

調查 Claude rate limit 行為後已確認：`rate_limits.*.used_percentage` 是 **fixed bucket** 語意（`resets_at` 到達時歸零），故可用線性預期模型安全計算 pace。

## Goals / Non-Goals

**Goals:**

- 讓使用者能在不自覺跨過 1M 貴價區、工作在子目錄失去定位、7 天使用節奏偏快時，**一眼**獲知並採取行動。
- 所有新增顯示都遵循「邊際觸發」原則 — 僅在特定條件下出現，不佔據平時版面。
- 保持既有 ASCII / NerdFont / TrueColor 三層降級策略。

**Non-Goals:**

- 不新增 payload 欄位的累積統計（cache 命中率、API vs total duration 比例等無直接行動性的數字）。
- 不改動 5h rate limit 顯示，包含不套用 pace 箭頭。
- 不改動現有的 `≥80%` rate limit 紅色、`≥90%` context 紅色警告閾值。
- 不變更 `rate_limits` 欄位解析邏輯（新指標純為渲染層衍生）。

## Decisions

### Use linear expected-usage model for 7d pace

比較「實際 used_percentage」與「按時間線性推算的預期使用量」：

```
elapsed        = window_length - (resets_at - now)
expected_pct   = elapsed / window_length × 100
deviation      = used_percentage - expected_pct
```

Window 長度為常數（`seven_day` = 604800s），不需 payload 提供。

**為何選線性模型**：fixed bucket 語意下使用量從 0 開始累積到 100%，均勻使用時應呈線性。任何偏離線性即代表「使用節奏異常」，語意直觀。

**替代方案考慮**：
- 指數/加權模型（愈接近重置時間權重愈高）— 過度工程化，一般使用者不會從中獲得更精確判斷。
- 固定百分比門檻（例如第 3 天超過 60% 才警告）— 缺乏彈性，每週使用時間點不同。

### Show pace arrow only for `seven_day`

`five_hour` 的 window 過短（300 分鐘），人類使用天然爆發式（連續寫碼數十分鐘後閒置），線性預期會產生高頻假警報 — 不提供 pace 箭頭。

### Deviation threshold: ±5%

誤差在 ±5% 內視為「大致符合預期」，顯示中性符號 `≈`（非箭頭）。

**為何是 5%**：
- 小於 3% 會在日常使用中抖動（使用 token 後 `used_percentage` 可能跳 1–2%），訊號噪聲比差。
- 大於 10% 會錯過真正該預警的早期異常（例如第 2 天就已超速 8%）。
- 5% 是平衡點，對應 7 天的 5% ≈ 8.4 小時使用量偏差 — 足夠顯著才提醒。

**為何 ≤5% 顯示 `≈` 而非空白**：使用者回饋指出完全隱藏會讓人無法確定「現在到底是接近預期還是 payload 根本沒送 resets_at」。`≈` 明確傳達「目前大致安全」語意，具行動性（不用改變使用節奏），且避免噪聲（不顯示數字）。

### Suppress arrow when < 10% window time remains

當 `resets_at - now < window_length × 10%`（7d 剩 < 16.8h）時不顯示箭頭。

**理由**：此時 `expected_pct` 已逼近 100%，線性模型的「預期」已失去警示意義（使用者再謹慎也無法在重置前降低 used_percentage）。

### Arrow symbols and colors

箭頭以**實心字符配對**（`▲` / `▼`）呈現，**後方直接附上偏差量**並以 `%` 結尾；前方以單一空格與主要百分比區隔（`7d:55% ▲7%`）。偏差量為 `round(|deviation|)` 整數（四捨五入），避免小數點擠佔版面。

| 狀態 | 符號 | ASCII fallback | 顏色 |
|------|------|---------------|------|
| 超支 > 5% | `▲<N>%` | `^<N>%` | 紅色（ANSI 31）|
| 落後 > 5% | `▼<N>%` | `v<N>%` | 灰色（ANSI 90）|
| ±5% 內（含 0） | `≈` | `~` | 灰色（ANSI 90）|
| 剩 <10% 時間 / `resets_at` 缺 | （無）| （無）| — |

**為何從「純箭頭」改為「箭頭 + 數字」**：初版設計假設方向性足以驅動決策，但使用者實測回饋 `▼` 無法分辨「略慢 4%」與「嚴重落後 30%」，二者對「是否該加速利用」的行動強度完全不同。加上整數偏差量可在 2–3 字元內給出行動強度訊息，仍遠短於「實際/預期」兩值並列（需 6+ 字元）。

**為何用 `≈` 而非重複箭頭符號**：`≈` 在視覺上明顯區別於 `▲`/`▼`，讀者能立即辨認「無偏差」vs「有偏差」；ASCII 的 `~` 同理。

### 1M label turns red when `exceeds_200k_tokens=true`

擴充既有 `ctxLabel()`：當 `context_window_size >= 1_000_000` **且** `exceeds_200k_tokens=true` 時，將 `1M` 字串從 `ansiGray` 改為 `ansiRed`。標籤文字本身不變（仍是 `1M`），僅色彩變化。

**為何僅變色而非加標記（如 `1M!` 或 `1M×2`）**：
- 1M 標籤已具識別性，變色足以傳達「進入計費門檻」語意。
- 保持寬度不變，避免 Line 1 抖動。

### Subdir display: `<project>/<relative>` when distinct

Line 2 目錄片段以 **3 層 fallback** 決定 project root：

```
resolveProjectRoot(currentDir, payloadProjectDir) →
  1. if payloadProjectDir ≠ "" and is_strict_ancestor(payloadProjectDir, currentDir):
         return payloadProjectDir
  2. walk upward from currentDir, stopping at the first directory whose child
     ".git" (file OR directory) exists → return that directory
  3. return ""   // no root detected
```

顯示邏輯：

```
root := resolveProjectRoot(current_dir, payload_project_dir)
if root == "":
    display = filepath.Base(current_dir)
elif current_dir == root:
    display = filepath.Base(root)
else:
    display = filepath.Base(root) + "/" + relative(root, current_dir)  // forward slashes
```

**為何走這三層順序**：
- 第 1 層尊重 Claude Code 未來若送出正確 `project_dir`（例如以 workspace 根為準）時能立即生效，不需改碼。
- 第 2 層處理**目前 Claude Code 的實際行為**：在 subfolder 啟動時，payload 的 `project_dir` 會與 `current_dir` 同步成 subfolder，第 1 層失效；此時走 git 偵測才能正確抓到真正的 repo root。
- 第 3 層守住「非 git 倉庫 / 沒裝 git / 跨磁碟符號連結」情境，不讓路徑顯示壞掉。

**為何不用 `git` CLI**：
- `.git` 偵測只需 `os.Stat`，純檔案系統操作。
- 避免 fork/exec 開銷；statusline 每輪對話都要渲染，延遲敏感。
- 使用者環境可能沒裝 git（雖然罕見）仍能運作。
- `.git` 為檔案的情境（git worktree、submodule）也能正確識別 — `os.Stat` 不區分 file/dir。

**為何不快取 root**：
- 偵測成本 = O(depth) 次 `os.Stat`，典型 ≤ 10 層、總耗時微秒級。
- root 在 session 存活期內不變，但既有 gitcache（5 秒 TTL）是為了 git status 的 fork cost，root 偵測無此負擔。
- 若觀察到效能問題再加快取即可（low-risk 後續優化）。

**為何不直接顯示完整路徑**：`project_dir` 本身可能深入 `D:/side_project/foo/bar`，畫面會過長。只保留 `<project name>/<relative>` 足以定位。

**Submodule：第一個 `.git` 勝出，不穿透到 parent**

在 submodule 內（例如 `/app/vendor/lib/src`）向上搜尋會先命中 `/app/vendor/lib/.git`（submodule 會把 `.git` 寫成 file），`/app/.git` 雖在更上層但**不會**被使用。理由：

- 既有 `gitcache` 透過 `git rev-parse` 取得的 branch 已是 **submodule 的 HEAD**，不是 parent 的；若 root 走到 parent 會造成「parent repo 路徑 + submodule branch」語意錯配。
- worktree 的 `.git` 也是 file（內容為 `gitdir: /path/to/main/repo/.git/worktrees/<name>`），與 submodule 結構相似但語意相反（worktree 是獨立 root，而 submodule 使用者可能希望穿透）。要可靠區分需解析 `.git` 檔案內容並判斷 `gitdir:` 是否指向某 repo 的 `modules/` 子路徑，複雜度高、錯誤率不低。
- 與 VS Code、多數 shell prompt 工具、`git status` 行為一致（皆以第一個 `.git` 為 repo boundary）。
- nested submodule 也有一致行為（最內層 submodule 為 root）。

取捨：使用者若把 parent repo 視為「專案」，在 submodule 工作時會看到 `lib/src` 而非 `app/vendor/lib/src`。接受此取捨以換取邏輯簡單與與既有 branch 顯示語意一致。

## Risks / Trade-offs

- **[風險] Rate limit fixed bucket 假設若未來 Claude 改為 rolling window** → pace 箭頭會變假訊號。**緩解**：在 renderer 注解標記此假設；若觀察到 `used_percentage` 行為不符 fixed bucket（例如 `resets_at` 到達時未歸零），快速切到 feature flag 關閉 pace 顯示。
- **[風險] `workspace.project_dir` 在某些使用情境可能為空字串** → 退回 base name 路徑。**緩解**：`renderer.go` 在 `project_dir == ""` 或等於 `current_dir` 時均走現行邏輯。
- **[取捨] 新欄位雖多但頻率低** → Line 1/2 平時版面不變，但 1M 貴價區、subdir、pace 三者同時觸發時（罕見），Line 會比過去更擠。評估後認為同時觸發的情境（且使用者恰好需要讀取）足夠稀少，不設定互斥優先級。
- **[取捨] `exceeds_200k_tokens` 為 payload 布林，無歷史感** → 若使用者剛跨過門檻，僅看到 `1M` 突然變紅，需自行連結「剛才哪輪對話跨過」。可接受 — 重點是「現在開始是貴的」，非歷史追蹤。
