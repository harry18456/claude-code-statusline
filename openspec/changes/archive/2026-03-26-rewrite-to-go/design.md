## Context

`statusline.sh` 是一個 365 行的 bash 腳本，透過 Claude Code 的 `statusLine` hook 讀取 stdin JSON 並輸出 ANSI 顏色的狀態列。現有問題：
- 依賴外部工具 `jq`，使用者需要額外安裝
- `stat -f %m` 是 macOS 專屬語法，Linux 上 git 快取失效
- 無法在 Windows 原生執行

目標是重寫為 Go 並以 binary 形式分發，功能完全對等。

## Goals / Non-Goals

**Goals:**
- 功能 100% 對等 `statusline.sh`（輸出格式、顏色邏輯、環境變數行為一致）
- 跨平台支援：macOS（amd64/arm64）、Linux（amd64）、Windows（amd64）
- 使用者零 runtime 依賴（只需要 git 本身）
- 修正 Linux git 快取 bug
- GitHub Actions 自動發布 binary

**Non-Goals:**
- 新增現有 bash 版本沒有的功能
- 支援 32-bit 架構
- 提供套件管理器安裝（brew/apt/choco）

## Decisions

### 使用純標準庫，不引入第三方套件

Go 標準庫已滿足全部需求：
- `encoding/json` — JSON 解析
- `os/exec` — git subprocess
- `os` — 環境變數、temp 路徑
- `fmt` — ANSI 輸出

引入第三方套件（如 `fatih/color`）會增加 module 管理複雜度，對此規模的工具不值得。

### 專案結構：cmd/ + internal/ 分層（Go 慣用結構）

```
cmd/
  statusline/
    main.go          ← 進入點，只負責串接
internal/
  model/
    payload.go       ← JSON struct 定義
  renderer/
    renderer.go      ← ANSI 輸出組裝
    renderer_test.go
  gitcache/
    gitcache.go      ← git branch + dirty-check 快取
    gitcache_test.go
go.mod
go.sum
.github/
  workflows/
    release.yml
reference/           ← 舊版 bash 檔案（暫存，最終刪除）
  statusline.sh
  install.sh
  examples/
  docs/
```

理由：採用 `internal/` 分層讓各模組可以獨立測試（TDD 的前提）。`cmd/` 是 Go CLI 工具的標準慣例。`reference/` 保留舊版供對照，確認行為一致後刪除。

### Git 快取使用 os.TempDir() + 檔案 mtime

```go
cacheFile := filepath.Join(os.TempDir(), "claude-statusline-git-cache")
info, _ := os.Stat(cacheFile)
age := time.Since(info.ModTime())
```

完全對等 bash 版本的 5 秒快取邏輯，且跨平台。

### ANSI 色碼直接輸出，不做 Windows 特判

Windows 10+ 的 Windows Terminal 和 ConEmu 原生支援 ANSI escape codes。Claude Code 本身也在這些環境下運行，不需要額外處理。若使用者的終端不支援，`CLAUDE_STATUSLINE_ASCII=1` 作為退路。

### GitHub Actions 交叉編譯策略

```yaml
strategy:
  matrix:
    include:
      - goos: darwin   goarch: amd64
      - goos: darwin   goarch: arm64
      - goos: linux    goarch: amd64
      - goos: windows  goarch: amd64
```

Go 的交叉編譯只需設定 `GOOS` / `GOARCH` 環境變數，不需要額外工具鏈。

### install.sh 平台偵測邏輯

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
# darwin/linux → 對應 binary
# Windows 用 Git Bash 執行 install.sh 時 uname -s → MINGW/MSYS
```

Windows 使用者需透過 Git Bash 執行 install.sh，或手動下載 `.exe`。

### TDD 開發流程：先寫測試，再實作

每個 `internal/` 子包遵循 RED → GREEN → REFACTOR：
1. 先建立 `*_test.go`，以 bash 版本的行為作為 golden truth 寫 test case
2. 確認測試 FAIL（RED）
3. 實作最小可通過的程式碼（GREEN）
4. 重構（REFACTOR）

測試指令：
```bash
go test ./...          # 全部測試
go test ./internal/... # 只跑 internal 套件
go test -v -run TestRenderer ./internal/renderer/
```

### 舊版檔案移至 reference/ 暫存後刪除

實作前先將現有 bash 相關檔案移到 `reference/` 資料夾，作為行為對照依據。Go 版本通過所有測試、行為確認一致後，整個 `reference/` 目錄刪除。不保留 legacy 版本。

## Risks / Trade-offs

- **Go binary 大小約 3~5MB** vs bash 腳本 15KB → 對使用者而言可接受，但 install.sh 需要 `curl`/`wget` 下載
- **Node.js 啟動 ~100ms 問題已規避** — Go binary 啟動 < 5ms，比原 bash 更快
- **bash 版本並存期間的維護負擔** → 建議保留 `statusline.sh` 但標記為 legacy，待 Go 版本穩定後再移除
- **Windows 上 `~` 路徑展開** → `settings.json` 中的路徑需寫絕對路徑，README 需說明

## Open Questions

- Go binary 的 release asset 命名要跟 bash 版本的 `statusline.sh` 有所區別嗎，還是直接叫 `statusline`？
- 是否要提供 `latest` tag 讓 install.sh 能自動取得最新版？
