## 1. 舊版檔案遷移至 reference/

- [x] 1.1 依「舊版檔案移至 reference/ 暫存後刪除」決策，建立 `reference/` 目錄並移入 `statusline.sh`、`examples/`、`docs/slides.js`
- [x] 1.2 移動舊版 `install.sh` 至 `reference/install.sh`（新版 install.sh 稍後另建）

## 2. 初始化 Go 專案

- [x] 2.1 建立 `go.mod`（module 名稱 `claude-code-statusline`）；依「使用純標準庫，不引入第三方套件」決策，確認不加入任何第三方依賴
- [x] 2.2 依「專案結構：cmd/ + internal/ 分層（Go 慣用結構）」決策，建立目錄骨架：`cmd/statusline/`、`internal/model/`、`internal/renderer/`、`internal/gitcache/`

## 3. model 套件（JSON 解析）

- [x] 3.1 依「TDD 開發流程：先寫測試，再實作」決策，先建立 `internal/model/payload_test.go`，涵蓋正常 JSON、缺欄位、空輸入等情境
- [x] 3.2 實作「Read JSON from stdin」：建立 `internal/model/payload.go`，定義對應 Claude Code JSON payload 的 Go struct，通過測試

## 4. renderer 套件（ANSI 輸出）

- [x] 4.1 先建立 `internal/renderer/renderer_test.go`，以 `reference/statusline.sh` 的輸出作為 golden truth，涵蓋所有情境（normal、warning、danger、startup、agent、worktree、ascii、nerdfont）
- [x] 4.2 實作「Three-tier color rendering」：依 `COLORTERM` 切換 true color / ANSI / ASCII 三種進度條渲染；依「ANSI 色碼直接輸出，不做 Windows 特判」決策，Windows 不需要特殊處理
- [x] 4.3 實作「Render two-line ANSI output」Line 1：model、progress bar、cost、duration、rate limits
- [x] 4.4 實作「Render two-line ANSI output」Line 2：branch、lines added/removed、dirname、agent/worktree
- [x] 4.5 實作「Zero-value hiding」：所有零值 section 正確省略
- [x] 4.6 實作「Nerd Font and Powerline support」：依 `CLAUDE_STATUSLINE_NERDFONT` / `CLAUDE_STATUSLINE_POWERLINE` 切換符號集與分隔符
- [x] 4.7 實作「Cost color thresholds」：$0.00 灰色、$0.01–$4.99 黃色、>=$5 黃色、>=$10 紅色（$0.00 仍顯示但 dimmed）
- [x] 4.8 實作「Rate limit display」：>= 80% 紅色，< 80% 灰色，absent 省略
- [x] 4.9 實作「Context window size label」：model name 不含 context 時顯示 1M / 200k
- [x] 4.10 實作「Agent and Worktree indicator」：worktree.name 優先於 agent.name
- [x] 4.11 實作「Warning symbol at high context usage」：ctx >= 90% 時在百分比後加紅色 ⚠（ASCII 模式用 `!`）
- [x] 4.12 實作「Progress bar dimensions and clamping」：進度條固定 10 格，pct 輸入 clamp 到 [0, 100]
- [x] 4.13 實作「Section text colors」：◆ 用 Anthropic purple RGB(114,102,234)、model 用 cyan、branch 用 gray、dir 用 blue
- [x] 4.14 實作「Lines added and removed colors」：+N 綠色、-N 紅色，格式 `+N/-N`
- [x] 4.15 實作「Duration sub-minute suppression」：dur_ms > 0 但換算後 < 1 秒仍省略 duration section

## 5. gitcache 套件

- [x] 5.1 先建立 `internal/gitcache/gitcache_test.go`，涵蓋 cache hit、cache miss、非 git 目錄等情境
- [x] 5.2 實作「Git branch with dirty-check caching」：依「Git 快取使用 os.TempDir() + 檔案 mtime」決策，用 `os.TempDir()` 確定 cache 路徑，以 `os.Stat().ModTime()` 計算 cache age（修正 `stat -f %m` Linux bug）
- [x] 5.3 實作 cache miss 邏輯：執行 `git branch --show-current` 與 `git diff --quiet`，結果寫入 cache 檔；加 `-c core.useBuiltinFSMonitor=false` 實作「Git subprocess noise suppression」
- [x] 5.4 實作「Branch name fallback to short SHA」：`git branch --show-current` 回傳空值時改用 `git rev-parse --short HEAD`

## 6. 進入點

- [x] 6.1 建立 `cmd/statusline/main.go`：串接 model / renderer / gitcache，從 stdin 讀取 JSON，輸出兩行 ANSI 文字；parse error 時輸出 fallback 並 exit 0

## 7. GitHub Actions 發布流程

- [x] [P] 7.1 建立 `.github/workflows/release.yml`：依「GitHub Actions 交叉編譯策略」決策設定 matrix；實作「Cross-compile for all target platforms」，以 `GOOS`/`GOARCH` 交叉編譯四個平台
- [x] [P] 7.2 實作「Triggered by git tag push」：workflow trigger 設為 `push: tags: ['v*']`
- [x] [P] 7.3 實作「Binary naming convention」：輸出 `statusline-<os>-<arch>`，Windows 加 `.exe`
- [x] [P] 7.4 實作「Attach binaries to GitHub Release」：使用 `softprops/action-gh-release` 上傳所有 binary

## 8. install.sh 更新

- [x] [P] 8.1 重新建立 `install.sh`，實作「Install script detects platform and downloads correct binary」：依「install.sh 平台偵測邏輯」決策，加入 `uname -s` / `uname -m` 偵測
- [x] [P] 8.2 依偵測結果從 GitHub Releases 下載對應 binary，放至 `~/.claude/statusline` 並 `chmod +x`
- [x] [P] 8.3 處理「Unsupported platform」：偵測到非 Darwin/Linux 時印出錯誤並引導手動下載

## 10. 版本號注入（ldflags）

- [x] 10.1 在 `cmd/statusline/main.go` 加入 `var version = "dev"` 變數（build time 預設值）
- [x] 10.2 更新 `.github/workflows/release.yml`：build 步驟加入 `-X main.version=${{ github.ref_name }}` 使 tag 名稱在 release build 時自動注入
- [x] [P] 10.3 commit 所有變更（包含 Go 重寫、README、CLAUDE.md、install.sh、release.yml 等）
- [x] [P] 10.4 push 至 remote，確認 GitHub Actions 可正常觸發

## 9. 驗證與收尾

- [x] 9.1 執行 `go test ./...` 確認全部測試通過；執行 `go build ./cmd/statusline/` 確認編譯成功
- [x] 9.2 以 `examples/test-mock.sh` 的 JSON fixtures（從 reference/ 取）對照 Go binary 輸出，確認行為與 bash 版本一致
- [x] [P] 9.3 確認行為一致後，刪除 `reference/` 目錄（完成「舊版檔案移至 reference/ 暫存後刪除」）
- [x] [P] 9.4 更新 `README.md` 與 `README.zh-TW.md`：安裝步驟改為下載 binary，移除 `jq` prerequisite，加入 Windows 說明
- [x] [P] 9.5 更新 `CLAUDE.md`：反映 Go 專案結構與 build / test 指令
